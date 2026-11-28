package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noxious Field — Enchantment — Aura {1}{B}{B} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant land
//	 Enchanted land has '{T}: This land deals 1 damage to each creature and
//	 each player.'"
//
// A granted activated ability on a land (Squirrel Nest's shape, ADR 0093)
// with Pyrohemia's damage-to-everything body, the land as the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c75ec3c6-2786-4386-b758-ac03610f2a99",
		Name:         "Noxious Field",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Grants: []effects.AbilityGrant{{
			Key: "noxious-field/sweep",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This land deals 1 damage to each creature and each player.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
				Cost:    effects.TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B23DamageEachCreatureAndEachPlayer(effects.NewContext(g, item), 1)
				},
			}},
			Text: "{T}: This land deals 1 damage to each creature and each player.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached("noxious-field/sweep")},
	})
}
