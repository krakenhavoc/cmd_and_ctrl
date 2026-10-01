package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mizzium Meddler — Creature — Vedalken Wizard {2}{U}, 1/4:
//
//	"Flash
//	 When this creature enters, you may change a target of target
//	 spell or ability to this creature."
//
// Spellskite's enters-trigger cousin, on the same pinned retarget
// (#1743, ChangeTargets.ToSource): a target of the spell or ability
// changes to this creature only if this creature is a legal target
// for that slot, judged for the item's controller, and only while
// this creature is still on the battlefield.
//
// "You may" is asked at resolution, where the printed text puts it:
// the trigger targets the spell or ability when it goes on the stack
// (CR 603.3d), and as it resolves the controller is offered the
// current targets that could become this creature — click one to
// change it, or decline with nothing. "Change A target" over an item
// with several is the same choice of which one Spellskite makes.
//
// DECLARED CAVEAT, Spellskite's: two instances of "target" that chose
// the same object are one thing to click, and picking it changes the
// earlier of them.
func init() {
	Register(Spec{
		OracleID:     "48a909b6-e6ee-4148-8b50-b35f11bc065f",
		Name:         "Mizzium Meddler",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"If a spell or ability targets the same thing more than once, you can't choose which of those targets changes to Mizzium Meddler — the first one does.",
		},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetSpellOrAbility("target spell or ability"),
			Key:       "Mizzium Meddler — change a target to this creature",
			Effect:    changeATargetToThisCreature("Mizzium Meddler — choose a target to change to Mizzium Meddler, or none", true),
		}},
	})
}
