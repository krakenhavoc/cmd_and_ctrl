package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elesh Norn, Mother of Machines — Legendary Creature — Phyrexian
// Praetor {4}{W}, 4/7:
//
//	"Vigilance
//	 If a permanent entering causes a triggered ability of a permanent
//	 you control to trigger, that ability triggers an additional time.
//	 Permanents entering don't cause abilities of permanents your
//	 opponents control to trigger."
//
// Vigilance is PrintedKeywords. The first ability is Panharmonicon's
// own `TriggerDoubler` (`DoublesEntering`) widened from "an artifact
// or creature entering" to "A PERMANENT entering" — a nil filter,
// since `DoublesEntering`'s own controller check already restricts it
// to "an ability of a permanent you control" and there is nothing
// left on the clause to narrow.
//
// # The second ability is a declared gap
//
// "Permanents entering don't cause abilities of permanents your
// opponents control to trigger" is a SUPPRESSION, not a doubling —
// the engine's `TriggerDoubler` can only ADD instances of a trigger
// that would otherwise fire (`Applies` returning true contributes one
// more copy); there is no vocabulary anywhere in the harvester for
// removing a trigger that would otherwise fire at all. Building it
// would mean teaching the trigger harvester a whole new kind of
// catalog hook, which is engine machinery this card-slice PR does not
// add.
//
// Leaving it out is the weaker-than-printed direction the #259 rule
// wants: an opponent's own ETB payoffs still fire normally against a
// permanent entering under Elesh Norn's controller's side (or
// anyone's), which only helps the OPPONENT, never the controller —
// Elesh Norn's own doubling clause is unaffected and fully live.
func init() {
	Register(Spec{
		OracleID:        "5ade11c0-41dd-4b6a-9f5b-c5903a3a0d7f",
		Name:            "Elesh Norn, Mother of Machines",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"vigilance"},
		Caveats: []string{
			"Permanents entering still cause your opponents' own triggered abilities to trigger normally — Elesh Norn doesn't suppress those.",
		},
		TriggerDoublers: []game.TriggerDoubler{eleshNornMotherOfMachinesDoubler()},
	})
}

func eleshNornMotherOfMachinesDoubler() game.TriggerDoubler {
	d := DoublesEntering(nil)
	d.Label = "Elesh Norn, Mother of Machines"
	return d
}
