package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alluring Siren — Creature — Siren {1}{U}, 1/1 (Magic 2012, S82 older sets
// batch 451, #5573):
//
//	"{T}: Target creature an opponent controls attacks you this turn if
//	 able."
//
// A CR 508.1d requirement on the one creature, "attacks that player if able",
// until end of turn (game.AttackPlayerIfAbleMod, as Knight Rampager's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f358544-7319-4f8e-9d60-a410c4feb774",
		Name:         "Alluring Siren",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature an opponent controls attacks you this turn if able.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AttackPlayerIfAbleMod(item.Controller)},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Alluring Siren — attacks you this turn if able",
				}.Apply(ctx)
			},
		}},
	})
}
