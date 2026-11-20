package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormscape Battlemage — Creature — Metathran Wizard {2}{U}, 2/2:
//
//	"Kicker {W} and/or {2}{B} (You may pay an additional {W} and/or
//	 {2}{B} as you cast this spell.)
//	 When this creature enters, if it was kicked with its {W} kicker,
//	 you gain 3 life.
//	 When this creature enters, if it was kicked with its {2}{B}
//	 kicker, destroy target nonblack creature. That creature can't be
//	 regenerated."
//
// Thornscape Battlemage's shape (#2153). The destruction ignores
// regeneration shields (CR 701.19c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "38ee748d-adcd-41df-9b23-d2a34829784c",
		Name:          "Stormscape Battlemage",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: effects.Kickers("{W}", "{2}{B}"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{W}")),
				"Stormscape Battlemage — kicked with {W}, you gain 3 life",
				effects.Do(effects.GainLife{Amount: 3})),
			effects.Targeting(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{2}{B}")),
				"Stormscape Battlemage — kicked with {2}{B}, destroy target nonblack creature",
				func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.DestroyTarget{Target: item.Targets[0].ID, CantBeRegenerated: true}.Apply(effects.NewContext(g, item))
				}), effects.TargetCreature("target nonblack creature", effects.NonBlack())),
		},
	})
}
