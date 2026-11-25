package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// hardCoverGrant is the bundle the Aura gives the enchanted creature. An
// on-disk identity: never renamed.
const hardCoverGrant = "hard-cover/loot"

// Hard Cover — Enchantment — Aura {U} (Core Set 2020, #5518):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2 and has "{T}: Draw a card, then discard a
//	 card.""
//
// A layer 7c pump and a layer 6 bundle (ADR 0093): the granted loot has a {T}
// in its cost, so it obeys summoning sickness (CR 302.6) for the creature it is
// on, and its controller activates it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8613d71b-4ac9-490d-b883-aefe03bcebce",
		Name:         "Hard Cover",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: hardCoverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
				},
			}},
			Text: "{T}: Draw a card, then discard a card.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			effects.GrantAbilitiesToAttached(hardCoverGrant),
		},
	})
}
