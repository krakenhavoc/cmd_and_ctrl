package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Robo-Piñata — Artifact Creature — Clown Robot {2}{W}, 2/1:
//
//	"When Robo-Piñata dies, choose one —
//	 • You get {TK}{TK}.
//	 • You may put a sticker on a nonland permanent you own."
//
// A modal dies trigger (CR 700.2b): the mode is chosen as it is put on the
// stack. The tickets are a player counter (CR 107.17); the sticker is the
// engine's keyed questions (ADR 0208 §4), a bullet whose step the mode
// walk can park behind them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ce17fde-71ec-482a-b9f9-78681bc894d8",
		Name:         "Robo-Piñata",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventLTB},
			AppliesTo: ThisDied,
			Key:       "Robo-Piñata — choose one",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeSteps("You get {TK}{TK}.", nil, GetTickets{N: 2}),
				ModeSteps("You may put a sticker on a nonland permanent you own.", nil, PutASticker{}),
			),
		}},
	})
}
