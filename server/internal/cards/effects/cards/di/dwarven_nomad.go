package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Nomad — Creature — Dwarf Nomad {2}{R}, 1/1 (Mirage, S82 OLD b808):
//
//	"{T}: Target creature with power 2 or less can't be blocked this turn."
//
// Access Tunnel's evasion body with a power cap of 2, read as the target is
// chosen and again as the ability resolves (CR 608.2b); the Nomad may target
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db130ff4-7525-421f-b7e7-200921ecbdec",
		Name:         "Dwarven Nomad",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 2 or less", effects.PowerLE(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Dwarven Nomad — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
