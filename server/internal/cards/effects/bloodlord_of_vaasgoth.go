package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodlord of Vaasgoth — Creature — Vampire Warrior {3}{B}{B}, 3/3:
//
//	"Bloodthirst 3
//	 Flying
//	 Whenever you cast a Vampire creature spell, it gains bloodthirst 3."
//
// Bloodthirst is the shared replacement (bloodthirst.go). The cast trigger
// marks the spell on the stack as it resolves, the way Wildgrowth
// Archaic's cast trigger marks "enters with counters", with a mark that
// carries bloodthirst's own condition: the three +1/+1 counters are put
// on only if an opponent was dealt damage this turn, read as the creature
// enters (CR 702.54a). A spell that already has bloodthirst gets this
// instance as well as its own (CR 702.54c), Bloodlord's own included.
// If the trigger is countered or the spell has left the stack nothing is
// marked, and a copy of the spell is another object with no mark.
//
// No simplification.
func init() {
	const label = "Bloodlord of Vaasgoth — that Vampire spell gains bloodthirst 3"
	Register(Spec{
		OracleID:        "d8560ee9-8cbb-440b-a132-e102a1eb3e89",
		Name:            "Bloodlord of Vaasgoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{Bloodthirst(3, "Bloodlord of Vaasgoth")},
		Triggered: []game.TriggeredAbility{
			rprB27BloodthirstOnCast("Bloodlord of Vaasgoth", And(Creature(), HasSubtype("Vampire")), label),
		},
	})
}
