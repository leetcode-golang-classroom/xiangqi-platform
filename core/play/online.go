package play

import (
	"github.com/yuanyu90221/xiangqi-platform/core/board"
	"github.com/yuanyu90221/xiangqi-platform/core/rules"
	"github.com/yuanyu90221/xiangqi-platform/player"
)

// OnlineController 為權威伺服器對戰的客戶端驅動器。它與離線 Controller 滿足同一個 Driver
// 介面，但走法權威性不同：
//
//   - 本地（人類）走法**只上送**伺服器（tr.Send），**不立即套用**到本地盤面；
//   - 本地盤面只依伺服器確認的走法串流（tr.Incoming，紅黑兩色皆然）推進。
//
// 因伺服器把 move_applied 廣播給雙方（含走子者本人），自己那步會以「回聲」形式回到 Incoming，
// 屆時才套用。此設計避免回聲被誤當對手走法、或本地與伺服器狀態分歧。
type OnlineController struct {
	s          *Session
	humanColor board.Color
	human      *player.Human
	tr         player.MoveTransport

	armed       bool              // 是否已向人類武裝取步
	waitingEcho bool              // 已上送本地走法、等待伺服器確認回聲
	localCh     <-chan board.Move // 當前回合人類的取步通道
}

// 編譯期保證：OnlineController 滿足 Driver。
var _ Driver = (*OnlineController)(nil)

// NewOnlineController 以紅黑名稱、人類執方與走法傳輸建立線上驅動器，回傳驅動器與人類玩家
// （供 GUI 餵入點擊、讀取選取高亮）。開局盤面為標準開局，與伺服器一致。
func NewOnlineController(red, black string, humanColor board.Color, tr player.MoveTransport) (*OnlineController, *player.Human) {
	h := player.NewHuman("你")
	return &OnlineController{
		s:          NewSession(red, black),
		humanColor: humanColor,
		human:      h,
		tr:         tr,
	}, h
}

// Session 回傳底層對局狀態。
func (o *OnlineController) Session() *Session { return o.s }

// Current 回傳當前盤面。
func (o *OnlineController) Current() *rules.Game { return o.s.Current() }

// Turn 回傳當前輪走方。
func (o *OnlineController) Turn() board.Color { return o.s.Turn() }

// Outcome 回傳對局結果。
func (o *OnlineController) Outcome() rules.Result { return o.s.Outcome() }

// CurrentInteractive 僅在「己方回合、未等待伺服器確認、且對局未結束」時回傳人類玩家。
func (o *OnlineController) CurrentInteractive() (player.Interactive, bool) {
	if o.s.Outcome().Over || o.waitingEcho || o.s.Turn() != o.humanColor {
		return nil, false
	}
	return o.human, true
}

// Thinking 回報是否正等待非本地輸入：對手回合，或已上送本地走法等待伺服器確認。
func (o *OnlineController) Thinking() bool {
	if o.s.Outcome().Over {
		return false
	}
	return o.waitingEcho || o.s.Turn() != o.humanColor
}

// Step 推進一格：先排空並套用所有伺服器確認的走法，再於己方回合武裝人類並把其產出的走法
// 上送伺服器（不本地套用）。回傳本次是否套用了確認走法。應每幀呼叫。
func (o *OnlineController) Step() bool {
	applied := o.drainConfirmed()

	if o.s.Outcome().Over {
		o.armed = false
		return applied
	}

	if o.s.Turn() == o.humanColor && !o.waitingEcho {
		if !o.armed {
			o.localCh = o.human.RequestMove(o.s.Current())
			o.armed = true
		}
		select {
		case m := <-o.localCh:
			o.armed = false
			o.waitingEcho = true
			_ = o.tr.Send(m) // 只上送，等伺服器回聲才套用
		default:
		}
	} else {
		o.armed = false
	}
	return applied
}

// drainConfirmed 非阻塞排空伺服器確認的走法串流，逐一套用到本地 Session；
// 套用任一筆即清除等待回聲旗標。回傳本次是否套用了走法。
func (o *OnlineController) drainConfirmed() bool {
	applied := false
	for {
		select {
		case m, ok := <-o.tr.Incoming():
			if !ok {
				return applied // 連線關閉
			}
			if err := o.s.Play(m); err == nil {
				applied = true
				o.waitingEcho = false
			}
		default:
			return applied
		}
	}
}
