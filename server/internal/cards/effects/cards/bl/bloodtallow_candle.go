package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodtallow Candle — Artifact {1} (Dominaria, S82 OLD, #5530):
//
//	"{6}, {T}, Sacrifice this artifact: Target creature gets -5/-5 until end
//	 of turn."
//
// The mana, the tap and the sacrifice are paid on announcement (CR 602.2b);
// the -5/-5 is the target's as the ability resolves, if it is still legal
// (CR 608.2b). The Candle being gone changes nothing: the effect does not
// depend on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da9b3edc-bc2d-4ff5-b614-3626f504c4e9",
		Name:         "Bloodtallow Candle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this artifact: Target creature gets -5/-5 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.BoostUntilEOT{
						Target: t.ID, Power: -5, Toughness: -5,
						Label: "Bloodtallow Candle — -5/-5 until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
