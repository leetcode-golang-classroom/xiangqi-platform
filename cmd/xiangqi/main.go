//go:build ebiten

// Command xiangqi 是單機雙人對弈的 Ebiten 棋盤渲染層。
//
// 本檔以建構標籤 `ebiten` 隔離：預設建構/測試（含無頭 CI）不會編譯它，
// 故圖形相依不影響 `go test ./...` 與 `pants test ::`。執行：
//
//	go run -tags ebiten ./cmd/xiangqi
//
// 人機對戰：可自由選邊——鍵盤 1 執紅、2 執黑（己方永遠在畫面下方，棋盤自動翻轉）。
// 互動：滑鼠左鍵點選棋子→點合法落點走子；U 悔棋、R 認輸、N 同陣營新局、
// S 匯出棋譜（存成 xiangqi-record-v1 JSON）、Q 結束遊戲。AI 那一方的搜尋於背景 goroutine 進行
// （讀取不可變盤面快照），完成後於主迴圈套用，避免畫面凍結。
// 對局協調委由純邏輯控制器 core/play.Controller。
//
// 棋子以正確中文字（帥/將…）繪製：嵌入 CJK 字型（Droid Sans Fallback）並以
// ebiten/v2/text/v2 渲染。狀態列等 ASCII 文字沿用內建點陣字型。
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/notation"
	"github.com/yuanyu90221/xiangqi-platform/core/play"
	"github.com/yuanyu90221/xiangqi-platform/core/record"
	"github.com/yuanyu90221/xiangqi-platform/core/storage"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

//go:embed assets/DroidSansFallbackFull.ttf
var cjkFontTTF []byte

// 字型：pieceFace 棋子中文字、uiFace 介面中文字（皆嵌入 CJK 字型）；
// latinFace 為 ASCII／符號專用（Go Regular，字形較 CJK fallback 清晰）。
var (
	pieceFace *text.GoTextFace
	uiFace    *text.GoTextFace
	latinFace *text.GoXFace
)

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(cjkFontTTF))
	if err != nil {
		log.Fatalf("載入棋子字型失敗: %v", err)
	}
	pieceFace = &text.GoTextFace{Source: src, Size: 34}
	uiFace = &text.GoTextFace{Source: src, Size: 16}

	ft, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("載入英文字型失敗: %v", err)
	}
	face, err := opentype.NewFace(ft, &opentype.FaceOptions{Size: 16, DPI: 72, Hinting: 0})
	if err != nil {
		log.Fatalf("建立英文字型失敗: %v", err)
	}
	latinFace = text.NewGoXFace(face)
}

const (
	margin  = 48
	cell    = 60
	radius  = 26
	headerH = 88                              // 頂部資訊／操作列高度（棋盤下移騰出此區）
	winW    = margin*2 + cell*(board.Files-1) // 9 條縱線
	winH    = headerH + margin*2 + cell*(board.Ranks-1)
)

var (
	colBg      = color.RGBA{0xf0, 0xd9, 0xa8, 0xff} // 棋盤底色
	colLine    = color.RGBA{0x5a, 0x3a, 0x1a, 0xff}
	colRed     = color.RGBA{0xc0, 0x30, 0x20, 0xff}
	colBlack   = color.RGBA{0x20, 0x20, 0x20, 0xff}
	colDisc    = color.RGBA{0xf7, 0xec, 0xd0, 0xff}
	colSelect  = color.RGBA{0x1e, 0x88, 0xe5, 0xff}
	colTarget  = color.RGBA{0x2e, 0x7d, 0x32, 0xff}
	colBanner  = color.RGBA{0x00, 0x00, 0x00, 0xc0}
	colWhite   = color.RGBA{0xff, 0xff, 0xff, 0xff}
	colMenuBg  = color.RGBA{0x1c, 0x1c, 0x1c, 0xff} // 選單背景（深色，與棋盤明顯區隔）
	colMenuBar = color.RGBA{0x37, 0x47, 0x4f, 0xff}
	colMenuFg  = color.RGBA{0xf5, 0xf5, 0xf5, 0xff} // 未選列文字（高對比白）
	colHeader  = color.RGBA{0xe8, 0xcf, 0x98, 0xff} // 頂部資訊列底色（較棋盤略深）
)

// pieceGlyph：FEN kind → 中文棋子字（紅/黑各一套），供日後嵌入 CJK 字型時使用。
var pieceGlyph = map[bool]map[byte]rune{
	true:  {'k': '帥', 'a': '仕', 'b': '相', 'n': '傌', 'r': '俥', 'c': '炮', 'p': '兵'},
	false: {'k': '將', 'a': '士', 'b': '象', 'n': '馬', 'r': '車', 'c': '砲', 'p': '卒'},
}

func abs(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}

// Game 為 Ebiten 遊戲：持有對局協調器並繪製之。可自由選邊（執紅/執黑），
// 並支援對局結束自動存譜與由 records/ 載入復盤。
type Game struct {
	c          *play.Controller
	store      *storage.FileStore
	humanColor board.Color // 人類執方
	difficulty int         // AI 難度（player.Easy/Medium/Hard）
	flip       bool        // 棋盤翻轉：人類執黑時讓己方在下方
	status     string      // 暫態訊息（如存檔路徑）
	savedOver  bool        // 本局結束時是否已自動存譜

	// 載入選單 / 復盤模式
	menu        bool
	menuEntries []storage.Entry
	menuIdx     int
	replay      bool
	replayer    *record.Replayer
	loadedID    string
}

func newGameState(store *storage.FileStore) *Game {
	g := &Game{store: store, difficulty: player.Medium}
	g.start(board.Red) // 預設執紅、普通難度
	return g
}

// start 以指定人類執方開新局，並依執方決定棋盤翻轉（己方永遠在下方）。
// AI 強度取自 g.difficulty。
func (g *Game) start(humanColor board.Color) {
	if g.difficulty < player.Easy {
		g.difficulty = player.Medium
	}
	ai := player.NewAI(g.difficulty)
	// 棋譜中以難度標明 AI 一方（如「電腦（普通）」），人類一方為「玩家」。
	redName, blackName := "玩家", ai.Name()
	if humanColor == board.Black {
		redName, blackName = ai.Name(), "玩家"
	}
	c, _ := play.VsComputer(redName, blackName, humanColor, ai)
	g.c = c
	g.humanColor = humanColor
	g.flip = humanColor == board.Black
	g.status = ""
	g.savedOver = false
	g.menu = false
	g.replay = false
	g.replayer = nil
}

// screenOf 回傳某棋格中心的螢幕座標。翻轉時 file/rank 同時鏡射（180° 旋轉），
// 使人類執方永遠位於畫面下方。
func (g *Game) screenOf(sq board.Square) (float32, float32) {
	file, rank := sq.File(), sq.Rank()
	if g.flip {
		file = board.Files - 1 - file
		rank = board.Ranks - 1 - rank
	}
	x := margin + file*cell
	y := headerH + margin + (board.Ranks-1-rank)*cell
	return float32(x), float32(y)
}

// squareAt 將螢幕座標轉成最接近的棋格（含翻轉）；超出容差回傳 InvalidSquare。
func (g *Game) squareAt(px, py int) board.Square {
	dfile := (px - margin + cell/2) / cell
	dtop := (py - headerH - margin + cell/2) / cell // 由上往下第幾列
	if dfile < 0 || dfile >= board.Files || dtop < 0 || dtop >= board.Ranks {
		return board.InvalidSquare
	}
	file, rank := dfile, board.Ranks-1-dtop
	if g.flip {
		file = board.Files - 1 - file
		rank = board.Ranks - 1 - rank
	}
	sq := board.MakeSquare(file, rank)
	cx, cy := g.screenOf(sq)
	if abs(float32(px)-cx) > cell/2 || abs(float32(py)-cy) > cell/2 {
		return board.InvalidSquare
	}
	return sq
}

func (g *Game) Update() error {
	// 全域結束遊戲：任何模式下按 Q 皆優雅關閉視窗。
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}
	if g.menu {
		return g.updateMenu()
	}
	if g.replay {
		return g.updateReplay()
	}

	// 統一迴圈：向當前 Player 請求一步、完成即套用。AI 的背景搜尋與套用皆在
	// Controller.Step 內非同步處理，本層不需自行管理 goroutine。
	g.c.Step()

	// 對局結束 → 自動存譜一次，確保棋譜必有產出。
	over := g.c.Outcome().Over
	if over && !g.savedOver {
		g.savedOver = true
		g.save(true)
	}

	// 全域可用鍵：開新局、選邊、載入、存譜（終局後仍可用）。
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyN):
		g.start(g.humanColor) // 同陣營重新開局
	case inpututil.IsKeyJustPressed(ebiten.Key1):
		g.start(board.Red) // 執紅（先手）
	case inpututil.IsKeyJustPressed(ebiten.Key2):
		g.start(board.Black) // 執黑（後手）
	case inpututil.IsKeyJustPressed(ebiten.KeyD):
		g.cycleDifficulty() // 切換難度（並以同陣營開新局）
	case inpututil.IsKeyJustPressed(ebiten.KeyS):
		g.save(false)
	case inpututil.IsKeyJustPressed(ebiten.KeyL):
		g.openMenu()
	}

	// 終局後鎖定對局操作：不接受悔棋／認輸／點擊走子。
	if over {
		return nil
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyU):
		g.c.Undo()
	case inpututil.IsKeyJustPressed(ebiten.KeyR):
		g.c.Resign()
	}

	// 人類回合：把左鍵點擊餵給當前互動式玩家。
	if iv, ok := g.c.CurrentInteractive(); ok {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if sq := g.squareAt(ebiten.CursorPosition()); sq != board.InvalidSquare {
				iv.Tap(sq)
			}
		}
	}
	return nil
}

// updateMenu 處理載入選單輸入：↑/↓ 選擇、Enter 載入、Esc 取消。
func (g *Game) updateMenu() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		if g.menuIdx < len(g.menuEntries)-1 {
			g.menuIdx++
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		if g.menuIdx > 0 {
			g.menuIdx--
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
		g.loadSelected()
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		g.menu = false
		g.status = ""
	}
	return nil
}

// updateReplay 處理復盤模式輸入：←/→ 逐手、L 回清單、Esc 返回對局。
func (g *Game) updateReplay() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		g.replayer.Next()
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		g.replayer.Prev()
	case inpututil.IsKeyJustPressed(ebiten.KeyL):
		g.openMenu()
	case inpututil.IsKeyJustPressed(ebiten.KeyEscape):
		g.replay = false
		g.status = ""
	}
	return nil
}

// save 將當前棋譜匯出為 xiangqi-record-v1 JSON（存於 records/）。auto 為對局結束自動存譜。
func (g *Game) save(auto bool) {
	id := "game-" + time.Now().Format("20060102-150405")
	if err := g.store.Save(id, g.c.Session().Record()); err != nil {
		g.status = "存檔失敗: " + err.Error()
		return
	}
	if auto {
		g.status = "棋局結束，已自動存譜 records/" + id + ".json"
	} else {
		g.status = "已匯出 records/" + id + ".json"
	}
}

// openMenu 由 records/ 載入棋譜清單並進入載入選單（預設選最新一份）。
func (g *Game) openMenu() {
	entries, err := g.store.List()
	if err != nil || len(entries) == 0 {
		g.status = "尚無棋譜可載入（先對局或按 S 存譜）"
		g.menu = false
		return
	}
	g.menuEntries = entries
	g.menuIdx = len(entries) - 1 // 最新一份
	g.menu = true
	g.replay = false
	g.status = ""
}

// loadSelected 載入選單中目前選定的棋譜，進入復盤模式。
func (g *Game) loadSelected() {
	e := g.menuEntries[g.menuIdx]
	rec, err := g.store.Load(e.ID)
	if err != nil {
		g.status = "載入失敗: " + err.Error()
		return
	}
	rp, err := record.NewReplayer(rec)
	if err != nil {
		g.status = "復盤失敗: " + err.Error()
		return
	}
	g.replayer = rp
	g.loadedID = e.ID
	g.menu = false
	g.replay = true
	g.status = ""
}

// boardFEN 回傳當前要繪製的盤面 FEN（復盤時為游標盤面，否則為對局當前盤面）。
func (g *Game) boardFEN() string {
	if g.replay && g.replayer != nil {
		return g.replayer.Current().ToFEN()
	}
	return g.c.Current().ToFEN()
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.menu {
		g.drawMenu(screen)
		return
	}
	screen.Fill(colBg)
	g.drawGrid(screen)
	g.drawPieces(screen)
	g.drawHints(screen)
	g.drawStatus(screen)
	g.drawBanner(screen)
}

func (g *Game) drawGrid(screen *ebiten.Image) {
	// 縱線（9 條）：邊線（a、i）貫通；內線於楚河漢界（rank4↔rank5）斷開。
	for f := 0; f < board.Files; f++ {
		if f == 0 || f == board.Files-1 {
			x0, y0 := g.screenOf(board.MakeSquare(f, board.Ranks-1))
			x1, y1 := g.screenOf(board.MakeSquare(f, 0))
			vector.StrokeLine(screen, x0, y0, x1, y1, 1.5, colLine, true)
			continue
		}
		// 上半（黑側 rank9→rank5）
		xt0, yt0 := g.screenOf(board.MakeSquare(f, board.Ranks-1))
		xt1, yt1 := g.screenOf(board.MakeSquare(f, 5))
		vector.StrokeLine(screen, xt0, yt0, xt1, yt1, 1.5, colLine, true)
		// 下半（紅側 rank4→rank0）
		xb0, yb0 := g.screenOf(board.MakeSquare(f, 4))
		xb1, yb1 := g.screenOf(board.MakeSquare(f, 0))
		vector.StrokeLine(screen, xb0, yb0, xb1, yb1, 1.5, colLine, true)
	}
	// 橫線（10 條）貫通
	for r := 0; r < board.Ranks; r++ {
		x0, y0 := g.screenOf(board.MakeSquare(0, r))
		x1, y1 := g.screenOf(board.MakeSquare(board.Files-1, r))
		vector.StrokeLine(screen, x0, y0, x1, y1, 1.5, colLine, true)
	}
	// 楚河漢界（河界文字）
	g.drawRiverText(screen)
	// 九宮對角（上下兩宮）
	g.drawPalace(screen, 7, 9) // 黑宮
	g.drawPalace(screen, 0, 2) // 紅宮
}

// drawRiverText 於河界中央標示「楚河　漢界」。
func (g *Game) drawRiverText(screen *ebiten.Image) {
	// 河界中心 y：rank4 與 rank5 之間。
	_, y4 := g.screenOf(board.MakeSquare(0, 4))
	_, y5 := g.screenOf(board.MakeSquare(0, 5))
	cy := float64(y4+y5) / 2
	xL, _ := g.screenOf(board.MakeSquare(2, 4))
	xR, _ := g.screenOf(board.MakeSquare(6, 4))
	centerText(screen, "楚河", float64(xL), cy, colLine)
	centerText(screen, "漢界", float64(xR), cy, colLine)
}

func (g *Game) drawPalace(screen *ebiten.Image, rLo, rHi int) {
	x0, y0 := g.screenOf(board.MakeSquare(3, rHi))
	x1, y1 := g.screenOf(board.MakeSquare(5, rLo))
	vector.StrokeLine(screen, x0, y0, x1, y1, 1.5, colLine, true)
	x2, y2 := g.screenOf(board.MakeSquare(5, rHi))
	x3, y3 := g.screenOf(board.MakeSquare(3, rLo))
	vector.StrokeLine(screen, x2, y2, x3, y3, 1.5, colLine, true)
}

func (g *Game) drawPieces(screen *ebiten.Image) {
	b, _, _, _, err := notation.ParseFEN(g.boardFEN())
	if err != nil {
		return
	}
	for i := 0; i < board.NumSquares; i++ {
		sq := board.Square(i)
		p := b.Get(sq)
		if p.IsEmpty() {
			continue
		}
		cx, cy := g.screenOf(sq)
		isRed := p.Color() == board.Red
		outline := colBlack
		if isRed {
			outline = colRed
		}
		vector.DrawFilledCircle(screen, cx, cy, radius, colDisc, true)
		vector.StrokeCircle(screen, cx, cy, radius, 2, outline, true)

		glyph := pieceGlyph[isRed][p.Kind()]
		op := &text.DrawOptions{}
		op.GeoM.Translate(float64(cx), float64(cy))
		op.PrimaryAlign = text.AlignCenter
		op.SecondaryAlign = text.AlignCenter
		op.ColorScale.ScaleWithColor(outline)
		text.Draw(screen, string(glyph), pieceFace, op)
	}
}

func (g *Game) drawHints(screen *ebiten.Image) {
	if g.replay {
		return
	}
	iv, ok := g.c.CurrentInteractive()
	if !ok {
		return
	}
	if sel, ok := iv.Selected(); ok {
		cx, cy := g.screenOf(sel)
		vector.StrokeCircle(screen, cx, cy, radius+3, 3, colSelect, true)
	}
	for _, t := range iv.Targets() {
		cx, cy := g.screenOf(t)
		vector.DrawFilledCircle(screen, cx, cy, 6, colTarget, true)
	}
}

func colorName(c board.Color) string {
	if c == board.Red {
		return "紅"
	}
	return "黑"
}

// difficultyName 將難度值轉中文標籤。
func difficultyName(d int) string {
	switch d {
	case player.Easy:
		return "簡單"
	case player.Hard:
		return "困難"
	default:
		return "普通"
	}
}

// cycleDifficulty 依序切換難度（簡單→普通→困難→簡單），並以同陣營重新開局套用。
func (g *Game) cycleDifficulty() {
	switch g.difficulty {
	case player.Easy:
		g.difficulty = player.Medium
	case player.Medium:
		g.difficulty = player.Hard
	default:
		g.difficulty = player.Easy
	}
	g.start(g.humanColor)
	g.status = "難度：" + difficultyName(g.difficulty)
}

func (g *Game) drawStatus(screen *ebiten.Image) {
	// 頂部資訊列背景（與棋盤明顯區隔）。所有中文皆以 CJK 字型繪製，避免亂碼。
	vector.DrawFilledRect(screen, 0, 0, float32(winW), headerH, colHeader, false)

	if g.replay {
		drawText(screen, fmt.Sprintf("復盤　%d / %d　棋譜：%s", g.replayer.Index(), g.replayer.Len()-1, g.loadedID), margin, 10, colLine)
		drawText(screen, "左/右鍵：逐手　　L：載入清單　　Esc：返回對局　　Q：結束", margin, 34, colLine)
		if g.status != "" {
			drawText(screen, g.status, margin, 58, colRed)
		}
		return
	}

	out := g.c.Outcome()
	var status string
	statusCol := colLine
	switch {
	case out.Over:
		status = "棋局結束"
		statusCol = colRed
	case g.c.Thinking():
		status = "電腦思考中…"
	case g.c.Turn() == g.humanColor:
		status = "輪到：你（" + colorName(g.humanColor) + "）"
	default:
		status = "輪到：電腦（" + colorName(g.humanColor.Opposite()) + "）"
	}
	drawText(screen, status, margin, 10, statusCol)
	drawText(screen, fmt.Sprintf("你執%s　難度：%s　手數：%d", colorName(g.humanColor), difficultyName(g.difficulty), len(g.c.Session().Record().Moves)), margin, 34, colLine)
	drawText(screen, "1 執紅  2 執黑  D 難度  N 新局  U 悔棋  R 認輸  S 存譜  L 載入  Q 結束", margin, 58, colLine)
	if g.status != "" {
		drawText(screen, g.status, margin, 72, colRed)
	}
}

// resultZh 將棋譜結果字串轉中文。
func resultZh(r string) string {
	switch r {
	case "red_win":
		return "紅勝"
	case "black_win":
		return "黑勝"
	case "draw":
		return "和局"
	case "":
		return "進行中"
	default:
		return r
	}
}

// drawText 於 (x,y) 左上對齊繪製混合文字：ASCII／符號以 latinFace（較清晰）、
// 中文等以 uiFace 繪製，逐段切換並累加寬度，避免 CJK fallback 字型的英文字形問題。
func drawText(screen *ebiten.Image, s string, x, y float64, clr color.Color) {
	runes := []rune(s)
	cx := x
	for i := 0; i < len(runes); {
		ascii := runes[i] < 0x80
		j := i + 1
		for j < len(runes) && (runes[j] < 0x80) == ascii {
			j++
		}
		seg := string(runes[i:j])
		var face text.Face = uiFace
		if ascii {
			face = latinFace
		}
		op := &text.DrawOptions{}
		op.GeoM.Translate(cx, y)
		op.ColorScale.ScaleWithColor(clr)
		text.Draw(screen, seg, face, op)
		w, _ := text.Measure(seg, face, 0)
		cx += w
		i = j
	}
}

// drawMenu 繪製獨立的載入選單畫面（不繪製棋盤）：標題列 + 棋譜清單 + 操作列。
func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(colMenuBg)

	// 標題列
	vector.DrawFilledRect(screen, 0, 0, float32(winW), 40, colMenuBar, false)
	drawText(screen, "載入棋譜", margin, 12, colWhite)
	drawText(screen, fmt.Sprintf("%d/%d", g.menuIdx+1, len(g.menuEntries)),
		float64(winW)-float64(margin)-40, 12, colWhite)

	// 清單（可捲動視窗）
	const top, rowH = 52, 26
	footerY := winH - 28
	maxVisible := (footerY - top) / rowH
	start := 0
	if g.menuIdx >= maxVisible {
		start = g.menuIdx - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(g.menuEntries) {
		end = len(g.menuEntries)
	}
	for i := start; i < end; i++ {
		e := g.menuEntries[i]
		y := top + (i-start)*rowH
		txtCol := colMenuFg
		if i == g.menuIdx {
			vector.DrawFilledRect(screen, 8, float32(y-3), float32(winW-16), rowH, colSelect, false)
			txtCol = colWhite
		}
		drawText(screen, fmt.Sprintf("%2d.", i+1), margin, float64(y), txtCol)
		drawText(screen, resultZh(e.Result), float64(margin)+34, float64(y), txtCol)
		drawText(screen, e.Red+" vs "+e.Black, float64(margin)+90, float64(y), txtCol)
		drawText(screen, e.ID, float64(margin)+200, float64(y), txtCol)
	}

	// 操作列
	vector.DrawFilledRect(screen, 0, float32(footerY-4), float32(winW), 32, colMenuBar, false)
	drawText(screen, "上/下鍵：選擇    Enter：載入    Esc：取消    Q：結束", margin, float64(footerY), colWhite)
}

// reasonZh 將結果原因轉中文。
func reasonZh(r string) string {
	switch r {
	case "checkmate":
		return "將死"
	case "stalemate":
		return "困斃"
	case "resign":
		return "認輸"
	case "repetition_draw":
		return "重複和局"
	case "natural_limit":
		return "自然限著和"
	case "perpetual_check":
		return "長將判負"
	default:
		return r
	}
}

// centerText 以嵌入 CJK 字型於 (x,y) 置中繪製文字。
func centerText(screen *ebiten.Image, s string, x, y float64, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, s, pieceFace, op)
}

// drawBanner 於對局結束時，在畫面中央顯示「棋局結束」與勝方。
func (g *Game) drawBanner(screen *ebiten.Image) {
	if g.replay {
		return
	}
	out := g.c.Outcome()
	if !out.Over {
		return
	}
	cx, cy := float64(winW)/2, float64(headerH+winH)/2 // 棋盤區中央
	vector.DrawFilledRect(screen, 0, float32(cy-60), float32(winW), 120, colBanner, false)

	var winLine string
	switch out.Winner {
	case "red":
		winLine = "紅方勝"
	case "black":
		winLine = "黑方勝"
	default:
		winLine = "和　局"
	}
	centerText(screen, "棋局結束", cx, cy-26, colWhite)
	centerText(screen, winLine+"（"+reasonZh(out.Reason)+"）", cx, cy+8, colWhite)
	drawText(screen, "棋盤已鎖定　按 N 重新開始　L 載入棋譜　Q 結束遊戲",
		cx-180, cy+34, colWhite)
}

func (g *Game) Layout(int, int) (int, int) { return winW, winH }

func main() {
	store, err := storage.NewFileStore("records")
	if err != nil {
		log.Fatalf("初始化棋譜儲存失敗: %v", err)
	}
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("中國象棋 — 人機對戰")
	if err := ebiten.RunGame(newGameState(store)); err != nil {
		log.Fatal(err)
	}
}
