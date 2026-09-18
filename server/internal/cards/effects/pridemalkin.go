package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pridemalkin — 2/1 Creature — Cat for {2}{G} (EDHREC rank 4366):
//
//	"When this creature enters, put a +1/+1 counter on target
//	 creature you control.
//	 Each creature you control with a +1/+1 counter on it has
//	 trample."
//
// A counters deck's cheap anthem: the ETB pays for itself and the
// static turns every counter already on the board into damage that
// gets through. Roadmap batch 42 (#449), "no new machinery".
//
// The two halves are different machinery and that is the point of the
// card file. The ETB is a CR 603 trigger with its own target, chosen
// as the ability goes on the stack (CR 603.3d) and re-checked when it
// resolves. The static is a CR 613 continuous effect at Layer 6,
// recomputed from scratch on every counter change — so a creature
// that gains its first counter gains trample in the same beat, and
// one whose counters are removed loses it again with no bookkeeping.
//
// The static reads COUNTERS, which the layer engine does not write.
// A +1/+1 counter is a physical marker on the permanent (CR 122), so
// the predicate reads the live card rather than the layered
// characteristic; the power and toughness the counter grants are the
// layer pass's business and this ability does not care about them.
//
// "Each creature YOU CONTROL", so an opponent's countered-up creature
// gets nothing. Pridemalkin itself is included — it has no counter of
// its own on arrival, but the ETB may target it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9672b63-415a-448b-a3da-140df63a0f0c",
		Name:         "Pridemalkin",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b42CreatureYouControlWithAPlusOneCounter, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pridemalkin — +1/+1 counter on target creature you control",
					plusOneCounterOnChosen),
				PermanentYouControl("target creature you control", Creature())),
		},
	})
}
