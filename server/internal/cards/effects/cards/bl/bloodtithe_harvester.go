package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodtithe Harvester — Creature — Vampire {B}{R}, 3/2 (Innistrad:
// Crimson Vow, #5495):
//
//	"When this creature enters, create a Blood token.
//	 {T}, Sacrifice this creature: Target creature gets -X/-X until end of
//	 turn, where X is twice the number of Blood tokens you control.
//	 Activate only as a sorcery."
//
// X is counted as the ability resolves (CR 608.2h: a count of what you
// control is read then, not at activation), twice the Blood tokens its
// controller has by then. With no Blood token it is -0/-0.
//
// No simplification.
func init() {
	const label = "Bloodtithe Harvester — target creature gets -X/-X until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "419a8ab2-cff6-4811-a559-8791c9bb3d8e",
		Name:         "Bloodtithe Harvester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bloodtithe Harvester — create a Blood token",
				effects.Do(effects.CreateToken{Template: effects.BloodToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Sacrifice this creature: Target creature gets -X/-X until end of turn, where X is twice the number of Blood tokens you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.TargetCreature("target creature"),
			Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := 2 * effects.VowB04BloodTokensControlledBy(g, item.Controller)
				return effects.BoostUntilEOT{Target: id, Power: -x, Toughness: -x, Label: label}.Apply(ctx)
			},
		}},
	})
}
