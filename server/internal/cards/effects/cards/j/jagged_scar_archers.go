package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jagged-Scar Archers — Creature — Elf Archer {1}{G}{G}, */* (Kaldheim
// Commander, #5506):
//
//	"Jagged-Scar Archers's power and toughness are each equal to the number
//	 of Elves you control.
//	 {T}: This creature deals damage equal to its power to target creature
//	 with flying."
//
// A characteristic-defining ability (layer 7a), itself counted. The damage
// is dealt by the Archers, its power read as the ability resolves (its
// last-known power if it has left, CR 608.2h); a target that no longer has
// flying is illegal and takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7139c70-dfcc-41ae-9d0d-b6e8c07a063d",
		Name:         "Jagged-Scar Archers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Jagged-Scar Archers's power and toughness are each equal to the number of Elves you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB074ElvesYouControl(g, source.Controller, source)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals damage equal to its power to target creature with flying.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: effects.SourcePowerNow(ctx)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
