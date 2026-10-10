package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oona's Prowler — Creature — Faerie Rogue {1}{B}, 3/1:
//
//	"Flying
//	 Discard a card: This creature gets -2/-0 until end of turn. Any
//	 player may activate this ability."
//
// Flying is a printed keyword (CR 702.9). The activation is an
// any-player row (CR 602.2, 602.1b) whose whole cost is a card from the
// ACTIVATOR's hand (CR 602.1a, 701.9a), so an opponent about to be hit
// can discard to blunt it, and the Prowler itself gets -2/-0 until end
// of turn (CR 611.2c, 514.2). Its power can go below 0; it then deals
// no combat damage.
//
// No Purpose: the bot never pays a card to shrink a creature it does
// not control (ADR 0106 owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f87252e-3218-4759-accc-a4f186fa8857",
		Name:            "Oona's Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:     "Discard a card: This creature gets -2/-0 until end of turn. Any player may activate this ability.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      DiscardACard(),
			AnyPlayer: true,
			Effect:    thisGetsUntilEndOfTurn(-2, 0, "Oona's Prowler — -2/-0 until end of turn"),
		}},
	})
}
