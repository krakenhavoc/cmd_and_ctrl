package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade of the Bloodchief — Artifact — Equipment {1} (Lost Caverns of Ixalan
// Commander):
//
//	"Whenever a creature dies, put a +1/+1 counter on equipped creature. If
//	 equipped creature is a Vampire, put two +1/+1 counters on it instead.
//	 Equip {1}"
//
// Any creature's death triggers it (diedCreature, which reads the type the
// creature last had, CR 603.10a), the equipped creature's own included. The
// Equipment's host and its Vampire type are read as the trigger resolves, so
// moving the Blade in response changes where the counters go; an unattached
// Blade, or a host that has left the battlefield, gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8985e545-07c3-45f9-9058-fe3c1f50de04",
		Name:         "Blade of the Bloodchief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := diedCreature(ev, g)
				return died
			}, "Blade of the Bloodchief — put a +1/+1 counter on equipped creature (two if it's a Vampire)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					host, ok := cmdr2B21EquippedCreature(g, ctx.Source())
					if !ok {
						return nil
					}
					n := 1
					if cmdr2B21HasSubtype(host, "Vampire") {
						n = 2
					}
					return AddCounter{Target: host.InstanceID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
