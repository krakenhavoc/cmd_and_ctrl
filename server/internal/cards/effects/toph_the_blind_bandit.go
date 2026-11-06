package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Toph, the Blind Bandit — Legendary Creature — Human Warrior Ally {2}{G},
// */3 (Avatar: The Last Airbender, #4459):
//
//	"When Toph enters, earthbend 2. (Target land you control becomes a
//	 0/0 creature with haste that's still a land. Put two +1/+1 counters
//	 on it. When it dies or is exiled, return it to the battlefield
//	 tapped.)
//	 Toph's power is equal to the number of +1/+1 counters on lands you
//	 control."
//
// The enters trigger is the earthbend keyword action with its own target
// clause (EarthbendTargets), chosen as the trigger goes on the stack.
//
// The power is a layer 7a characteristic-defining ability (CR 604.3,
// 613.4a), which works in every zone (the ruling): it counts the +1/+1
// counters on the permanents its controller controls that are lands
// through the layers, so a land earthbent by any card counts, and a
// creature that is not a land does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ce20f43-9357-448f-bb40-7315feb08433",
		Name:         "Toph, the Blind Bandit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Toph, the Blind Bandit — earthbend 2", avB16Earthbend(2)), EarthbendTargets()),
		},
		Static: []game.StaticAbility{{
			Label:     "Toph's power is equal to the number of +1/+1 counters on lands you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = avB16PlusOneCountersOnLandsYouControl(g, source.Controller)
			},
		}},
	})
}
