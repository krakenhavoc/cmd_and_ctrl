package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoving Match — Instant {2}{U} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Until end of turn, all creatures gain "{T}: Tap target creature.""
//
// Banishing Knack's grant (ADR 0093) over every creature on the battlefield
// as the spell resolves (CR 611.2c): each creature's controller may tap it to
// tap a creature.
//
// No simplification.
const shovingMatchTap = "shoving-match/tap"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ef07ae5-3b04-45c8-a080-852b84da4fef",
		Name:         "Shoving Match",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: shovingMatchTap,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Tap target creature.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			}},
			Text: "{T}: Tap target creature.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantAbilitiesFor{
				Match: effects.Creature(),
				Keys:  []string{shovingMatchTap},
				Label: "Shoving Match — all creatures can tap a creature until end of turn",
			}.Apply(ctx)
		},
	})
}
