package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abominable Treefolk — Snow Creature — Treefolk {2}{G}{U}, */* (Modern
// Horizons):
//
//	"Trample
//	 Abominable Treefolk's power and toughness are each equal to the number of
//	 snow permanents you control.
//	 When this creature enters, tap target creature an opponent controls. That
//	 creature doesn't untap during its controller's next untap step."
//
// The first static is a characteristic-defining ability (CR 604.3, layer 7a):
// it counts the snow permanents its controller controls as the layers are
// applied, the Treefolk itself included because it is snow. The trigger is
// Spire Patrol's tap-and-freeze.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d341ed4-4f9b-40df-9bd9-3bfeaf0fe595",
		Name:            "Abominable Treefolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Abominable Treefolk's power and toughness are each equal to the number of snow permanents you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CountPermanentsYouControl(g, source, func(p *game.Card) bool { return p.HasSupertype("snow") })
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Abominable Treefolk — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "Abominable Treefolk — doesn't untap during its controller's next untap step")
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
