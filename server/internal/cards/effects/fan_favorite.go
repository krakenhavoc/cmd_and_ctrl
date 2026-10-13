package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fan Favorite — Creature — Human Rogue {3}{B}, 2/2:
//
//	"Assist (Another player can pay up to {3} of this spell's cost.)
//	 {2}: This creature gets +1/+1 until end of turn. Any player may
//	 activate this ability."
//
// Assist (CR 702.132a, ADR 0150 §4) is the printed keyword: the caster
// may name another player on cast_spell's `assist_player`, and the cast
// waits for that player to pay any amount of the generic mana in the
// spell's total cost, {3} plus commander tax and increases, less
// reductions. The engine reads it off the card (game.HasAssist), so
// nothing else is declared for it here.
//
// The pump is an any-player row (CR 602.2, 602.1b): whoever activates
// it pays the {2} out of their own pool (CR 602.1a) and Fan Favorite
// itself gets +1/+1 until end of turn (CR 611.2c, 514.2). No Purpose:
// the bot never pumps a creature it does not control (ADR 0106 owner
// decision 2).
func init() {
	Register(Spec{
		OracleID:        "9991c9f2-cb6e-4cfe-a3b7-650212454195",
		Name:            "Fan Favorite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Activated: []ActivatedAbility{{
			Label:     "{2}: This creature gets +1/+1 until end of turn. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{2}"),
			AnyPlayer: true,
			Effect:    thisGetsUntilEndOfTurn(1, 1, "Fan Favorite — +1/+1 until end of turn"),
		}},
	})
}
