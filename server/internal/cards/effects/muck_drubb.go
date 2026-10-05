package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muck Drubb — Creature — Beast {3}{B}{B}, 4/4:
//
//	"Flash
//	 When this creature enters, change the target of target spell that
//	 targets only a single creature to this creature.
//	 Madness {2}{B}"
//
// Mizzium Meddler's pinned retarget (ChangeTargets.ToSource) with a
// mandatory trigger and a different clause: the spell must have one
// target slot and it must name a creature on the battlefield. The
// change is made only if Muck Drubb is a legal target for that slot
// (hexproof, a clause that does not take it) — otherwise nothing
// changes, exactly as the engine's gate runs it for Spellskite.
func init() {
	Register(Spec{
		OracleID:        "032a5616-bbaf-4659-86c4-43edf29b9788",
		Name:            "Muck Drubb",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Madness:         "{2}{B}",
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetSpell("target spell that targets only a single creature",
				TargetsOnlyACreature()),
			Key:    "Muck Drubb — change the target to this creature",
			Effect: changeATargetToThisCreature("Muck Drubb — change the target to Muck Drubb", false),
		}},
	})
}
