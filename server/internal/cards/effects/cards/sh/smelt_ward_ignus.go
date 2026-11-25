package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smelt-Ward Ignus — Creature — Elemental {1}{R}, 2/1 (Ravnica Allegiance):
//
//	"{2}{R}, Sacrifice this creature: Gain control of target creature with
//	 power 3 or less until end of turn. Untap that creature. It gains haste
//	 until end of turn. Activate only as a sorcery."
//
// Bloody Betrayal's steal, untap and haste, aimed at a creature with power 3
// or less, behind a sorcery-speed sacrifice cost.
//
// No simplification.
func init() {
	const label = "Smelt-Ward Ignus — gain control of the creature until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "494b642c-b2b5-40a0-b741-dca1aa9bdb59",
		Name:         "Smelt-Ward Ignus",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{R}, Sacrifice this creature: Gain control of target creature with power 3 or less until end of turn. Untap that creature. It gains haste until end of turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 || targets[0].Kind != game.TargetCard {
					return nil
				}
				target := targets[0].ID
				if err := (effects.GainControl{Target: target, Duration: effects.DurationUntilEndOfTurn(ctx), Label: label}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"}, Label: "Smelt-Ward Ignus — haste until end of turn"}.Apply(ctx)
			},
		}},
	})
}
