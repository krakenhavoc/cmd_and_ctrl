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
// The second ability is a `TriggerSuppressor` (#1735), the doubler's
// sibling at the same harvest point: `SuppressesEntering(nil)` (any
// permanent) narrowed by `OfOpponentsPermanents`. It shipped as a
// caveat until the engine could stop a trigger rather than only add
// one. Because the harvest asks the suppressors before the doublers,
// two players who each control an Elesh Norn get no enters triggers at
// all: each player's ability is stopped by the other player's Elesh
// Norn, so there is nothing left for their own Elesh Norn to double.
// An opponent's evoked creature is not sacrificed, because evoke's
// sacrifice is that creature's own enters trigger (CR 702.74a). Both
// abilities apply to permanents entering at the same time as Elesh
// Norn, including Elesh Norn itself, because an enters trigger is
// judged on the board after the event (CR 603.10).
func init() {
	Register(Spec{
		OracleID:           "5ade11c0-41dd-4b6a-9f5b-c5903a3a0d7f",
		Name:               "Elesh Norn, Mother of Machines",
		Completeness:       CompletenessFull,
		PrintedKeywords:    []string{"vigilance"},
		TriggerDoublers:    []game.TriggerDoubler{eleshNornMotherOfMachinesDoubler()},
		TriggerSuppressors: []game.TriggerSuppressor{eleshNornMotherOfMachinesSuppressor()},
	})
}

func eleshNornMotherOfMachinesDoubler() game.TriggerDoubler {
	d := DoublesEntering(nil)
	d.Label = "Elesh Norn, Mother of Machines"
	return d
}

func eleshNornMotherOfMachinesSuppressor() game.TriggerSuppressor {
	s := OfOpponentsPermanents(SuppressesEntering(nil))
	s.Label = "Elesh Norn, Mother of Machines"
	return s
}
