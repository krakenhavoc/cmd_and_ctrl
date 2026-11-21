package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldrazi Monument — Artifact {5}:
//
//	"Creatures you control get +1/+1 and have flying and indestructible.
//	 At the beginning of your upkeep, sacrifice a creature. If you can't,
//	 sacrifice this artifact."
//
// The three statics are one lord's over every creature its controller
// controls (TribeFilter with no tribe names every creature). The upkeep
// trigger is Doomgape's keyed sacrifice: its controller picks a creature to
// sacrifice, and when they controlled none to pick (CR 701.21a, "if you
// can't") the Monument itself is sacrificed, if it is still on the
// battlefield. A creature sacrificed to the first half spares the Monument,
// whatever it was; indestructible does not stop a sacrifice.
func init() {
	const name = "Eldrazi Monument"
	yours := TribeFilter{YoursOnly: true}
	Register(Spec{
		OracleID:     "c7ee1b21-d02e-49ef-95d6-d6dbf2f886ae",
		Name:         name,
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(yours, 1, 1),
			TribalKeywordGrant(yours, "flying"),
			TribalKeywordGrant(yours, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep(name+" — sacrifice a creature; if you can't, sacrifice this artifact",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, item.Controller,
						sacrificeSpec("a creature", Creature()),
						name+" — sacrifice a creature", 1,
						eldraziMonumentSacrificed, []uuid.UUID{item.Controller})
				}),
		},
	})
}

// eldraziMonumentSacrificed sacrifices the Monument when its controller
// sacrificed no creature. An on-disk key.
var eldraziMonumentSacrificed = game.RegisterPromptRunThen("prompt-run/eldrazi-monument-sacrificed", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	me := r.Carry[0]
	if len(r.Sacrifices().By(me)) > 0 {
		return nil
	}
	if z := g.FindCardZoneForEffect(r.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return SacrificePermanent{Target: r.Source}.Apply(keyedContext(g, me, r.Source))
})
