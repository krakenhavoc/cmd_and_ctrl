package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belbe's Armor — Artifact {3} (Nemesis, S82 OLD batch 679, #5618):
//
//	"{X}, {T}: Target creature gets -X/+X until end of turn."
//
// X is announced with the activation (CR 602.2b) and read as the ability
// resolves; the target is re-checked then (CR 608.2b). With X = 0 nothing
// changes. The {T} is the Armor's own, so summoning sickness never applies
// (it is not a creature).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1650d85f-4c65-429d-8952-f68b62935a55",
		Name:         "Belbe's Armor",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}: Target creature gets -X/+X until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := ctx.X()
				return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: x, Label: "Belbe's Armor — -X/+X"}.Apply(ctx)
			},
		}},
	})
}
