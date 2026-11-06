package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earthen Ally — Creature — Human Soldier Ally {G}, 0/2:
//
//	"This creature gets +1/+0 for each color among Allies you control.
//	 {2}{W}{U}{B}{R}{G}: Earthbend 5."
//
// The pump is Faeburrow Elder's self-static (layer 7c) counting the
// colours among Allies its controller controls, itself included, so it
// is at most +5/+0 (the ruling). The activated ability earthbends 5 at
// instant speed: no timing restriction is printed.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "913bca28-faf9-42e6-a8c0-469cd9bc0208",
		Name:         "Earthen Ally",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "This creature gets +1/+0 for each color among Allies you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += avatarB22ColorsAmongSubtypeYouControl(g, source.Controller, source.TextSubtype("Ally"))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}{U}{B}{R}{G}: Earthbend 5.",
			Cost:    ManaCost("{2}{W}{U}{B}{R}{G}"),
			Targets: EarthbendTargets(),
			Effect:  avatarB22Earthbend(5),
		}},
	})
}
