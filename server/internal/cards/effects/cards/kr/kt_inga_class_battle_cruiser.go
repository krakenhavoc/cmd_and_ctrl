package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// K't'inga-Class Battle Cruiser — Artifact — Spacecraft {3}, 5/5 (Star
// Trek Commander, #4974):
//
//	"Whenever this Spacecraft enters or attacks, one or two target
//	 Klingons and/or Warriors you control gain flying until end of turn.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 4+.)
//	 4+ | Flying"
//
// One printed ability with two conditions (WhenThisEntersOrAttacks); it
// can attack only once it is a creature at four charge counters. The one
// or two targets are Klingon and/or Warrior creatures you control, read
// through the layered view; each still legal as it resolves gains flying.
// With none to target the trigger is removed (CR 603.3d). Station is the
// ordinary keyword (station.go).
//
// No simplification.
func init() {
	const label = "K't'inga-Class Battle Cruiser — one or two target Klingons and/or Warriors you control gain flying"
	enters := effects.WhenThisEntersOrAttacks(label, func(g *game.Game, item *game.StackItem) error {
		return effects.GrantEachLegalTargetUntilEOT(effects.NewContext(g, item), "flying", label)
	})
	enters.Targets = effects.TargetCreature("one or two target Klingons and/or Warriors you control",
		effects.YouControl(), effects.Or(effects.HasSubtype("Klingon"), effects.HasSubtype("Warrior"))).WithCount(1, 2)
	effects.Register(effects.Spec{
		OracleID:     "8ea8ee52-30e1-415c-af52-66b209f05f0a",
		Name:         "K't'inga-Class Battle Cruiser",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(4, 5, 5),
			effects.ThresholdKeywords(4, "flying")),
		Triggered: []game.TriggeredAbility{enters},
	})
}
