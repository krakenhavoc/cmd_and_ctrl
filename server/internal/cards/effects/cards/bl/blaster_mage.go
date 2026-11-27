package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blaster Mage — Creature — Human Spellshaper {2}{R}, 2/2 (Mercadian Masques,
// S82 OLD, #5619):
//
//	"{R}, {T}, Discard a card: Destroy target Wall."
//
// The discard is a cost, and the {T} is the Mage's own, so it is summoning-sick
// the turn it arrives (CR 302.6). A Wall is any creature (or permanent) with the
// Wall subtype, and the target is checked again as the ability resolves
// (CR 608.2b). There is no regeneration clause, so a regeneration shield saves it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5b35b54-77d9-4cb0-86a1-efc8e49aea09",
		Name:         "Blaster Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: Destroy target Wall.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPermanent("target Wall", effects.HasSubtype("Wall")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
