package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iron Hills Stalwart — Creature — Dwarf Warrior {4}{R}, 4/5 (The Hobbit,
// HOB):
//
//	"Reach, trample
//	 When this creature enters, attach target Equipment you control to up
//	 to one target creature you control."
//
// Brass Squire's two clauses (AttachClauseTargets), the second one "up to
// one": the trigger needs an Equipment you control to go on the stack, and
// with no creature chosen it attaches nothing (CR 701.3b). Each clause is
// re-checked at resolution (CR 608.2b). Control of the Equipment does not
// change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd4ab393-9538-4230-b2a1-bc77099481c9",
		Name:            "Iron Hills Stalwart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Iron Hills Stalwart — attach target Equipment you control to up to one target creature you control", AttachClauseTargets),
				Clauses(
					TargetPermanent("target Equipment you control", HasSubtype("Equipment"), YouControl()),
					UpToOneTargetCreature("up to one target creature you control", YouControl()),
				)),
		},
	})
}
