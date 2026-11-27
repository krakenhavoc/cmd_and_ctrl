package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coffin Puppets — Creature — Zombie {3}{B}{B}, 3/3 (Prophecy, S82 OLD batch
// 676, #5617):
//
//	"Sacrifice two lands: Return this card from your graveyard to the
//	 battlefield. Activate only during your upkeep and only if you control a
//	 Swamp."
//
// Firemane Angel's graveyard return, gated on the controller's own upkeep
// (DuringYourUpkeep) and, with it, on controlling a Swamp (ControlsAtLeast over
// the Swamp subtype, Mox Jasper's shape), both read at activation (CR 602.5b).
// The cost is two lands of the player's choice, sacrificed as it is activated
// (the Swamp may be one of them, since the condition is read before the cost is
// paid). If the card has left the graveyard by the time the ability resolves,
// it does nothing (CR 608.2b); it returns under its owner's control (CR 400.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e49771ab-cabb-4a5f-aef3-a0ac8147e6bf",
		Name:         "Coffin Puppets",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two lands: Return this card from your graveyard to the battlefield. Activate only during your upkeep and only if you control a Swamp.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("two lands", effects.Land()).WithCount(2, 2)},
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.AllConditions(effects.DuringYourUpkeep(),
				effects.ControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord("Swamp") })),
			Effect: effects.ReturnThisFromYourGraveyardToTheBattlefield,
		}},
	})
}
