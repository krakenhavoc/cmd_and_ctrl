package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Root-Kin Ally — Creature — Elemental Warrior {4}{G}{G}, 3/3 (Ravnica: City of
// Guilds, S82 OLD batch 575, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you tap
//	 while casting this spell pays for {1} or one mana of that creature's
//	 color.)
//	 Tap two untapped creatures you control: This creature gets +2/+2 until end
//	 of turn."
//
// Convoke is the engine's keyword (Rosemane Centaur's). The cost is a
// tap-others component, not the {T} symbol, so the Ally may tap itself as one
// of the two and a creature that arrived this turn may pay (CR 302.6).
//
// No simplification.
func init() {
	const tapTwo = "two untapped creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "99d0876f-174b-4cfe-aa52-f3d5434f6cba",
		Name:         "Root-Kin Ally",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap two untapped creatures you control: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Creature()),
				Label:  tapTwo,
			}},
			Effect: effects.ThisGetsUntilEndOfTurn(2, 2, "Root-Kin Ally — +2/+2 until end of turn"),
		}},
	})
}
