package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydroelectric Specimen // Hydroelectric Laboratory — the FRONT
// face, Creature — Weird {2}{U}, 1/4:
//
//	"Flash
//	 When this creature enters, you may change the target of target
//	 instant or sorcery spell with a single target to this creature."
//
// The land back (pay 3 life or enter tapped; {T}: Add {U}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey). Flash rides PrintedKeywords.
//
// The enters trigger waited on a retarget whose new target is a FIXED
// object rather than a chooser's pick: ChangeTargets alone opens a
// picker over every legal new target, which would let the controller
// point the spell anywhere — stronger than printed (#259). #1743 added
// the pinned variant (ChangeTargets.ToSource, ADR 0019's 2026-10-01
// amendment): the spell's one target changes to this creature only if
// this creature is a legal target for it, judged for the spell's
// controller, and only while this creature is still on the
// battlefield. Otherwise nothing changes (CR 115.7a).
//
// "You may" is asked at resolution, where the printed text puts it:
// the trigger targets the spell when it goes on the stack (CR 603.3d),
// and the prompt that opens as it resolves offers the spell's current
// target, answered by clicking it or declined with nothing.
func init() {
	Register(Spec{
		OracleID:        "573151f0-00d4-4a8a-8a09-745c5f376532",
		Name:            "Hydroelectric Specimen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetSpell("target instant or sorcery spell with a single target",
				Or(Instant(), Sorcery()), HasASingleTarget()),
			Key:    "Hydroelectric Specimen — change the target to this creature",
			Effect: changeATargetToThisCreature("Hydroelectric Specimen — change the target to Hydroelectric Specimen?", true),
		}},
	})
}
