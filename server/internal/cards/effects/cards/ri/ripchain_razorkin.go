package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ripchain Razorkin — Creature — Human Berserker {3}{R}, 5/3 (Duskmourn:
// House of Horror, #4466):
//
//	"Reach
//	 {2}{R}, Sacrifice a land: Draw a card."
//
// Dal R'El, Augment Orphan's activated ability at a different price: the
// land you control is sacrificed as part of the cost (CR 602.2b), so with
// no land the ability can't be activated. No {T} in the cost, so it works
// the turn the Razorkin arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce000c0b-db42-4569-855f-f4eae0431c09",
		Name:            "Ripchain Razorkin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice a land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(1, "a land", effects.Land())),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
