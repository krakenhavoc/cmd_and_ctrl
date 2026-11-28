package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Might — Enchantment — Aura {U} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Enchant land you control
//	 Cumulative upkeep {1}{U}
//	 Enchanted land has "{T}: Target creature gets +2/+2 until end of turn.""
//
// Squirrel Nest's granted activated ability (ADR 0093) on the enchanted land, whose
// controller activates it, taps the land for it, and chooses the creature: the boost
// is a layer 7c modification that ends at cleanup (CR 514.2).
//
// No simplification.
const mysticMightGrant = "mystic-might/pump"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "967eca77-d273-4fc6-97d1-df7d8cc6ff59",
		Name:         "Mystic Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Mystic Might — cumulative upkeep {1}{U}", "{1}{U}")},
		Grants: []effects.AbilityGrant{{
			Key: mysticMightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gets +2/+2 until end of turn",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BoostTheTargetUntilEOT(2, 2, "Mystic Might — +2/+2 until end of turn")(item, effects.NewContext(g, item))
				},
			}},
			Text: "{T}: Target creature gets +2/+2 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(mysticMightGrant)},
	})
}
