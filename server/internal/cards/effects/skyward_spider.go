package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyward Spider — Creature — Spider Human Hero {W/U}{W/U}, 2/2
// (Marvel's Spider-Man):
//
//	"Ward {2}
//	 This creature has flying as long as it's modified. (Equipment, Auras
//	 you control, and counters are modifications.)"
//
// Ward is the shared triggered ward. The flying grant is Obstinate
// Gargoyle's: a layer 6 keyword grant on this permanent while CR 700.9's
// "modified" holds (b07IsModified), so a counter, an Equipment or an Aura
// its controller controls gives it flying and removing the last one takes it
// away again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6ba76fa-441c-4a74-bfea-8ae2a2c9e391",
		Name:         "Skyward Spider",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && b07IsModified(g, *target)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Skyward Spider — ward {2}"),
		},
	})
}
