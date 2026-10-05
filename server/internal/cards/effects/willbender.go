package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Willbender — 1/2 Human Wizard for {1}{U}:
//
//	"Morph {1}{U}. When this creature is turned face up, change the
//	 target of target spell or ability with a single target."
//
// The card morph was printed for, and all of it works. The morph is
// whole: cast face down for {3} it is a nameless 2/2 neither opponent
// can price, and {1}{U} at any time they have to respond to — in
// response to the removal spell pointed at it, which is the point —
// turns it face up (CR 708.6, CR 702.37c). Willbender is also the
// printed reason the turn-face-up EVENT exists at all (CR 708.8,
// ADR 0082 decision 7): a permanent turning over has to be observable
// or this trigger has nothing to watch.
//
// The trigger is "change the target of target spell or ability with a
// single target" — CR 115.7b over the same gate Bolt Bend and
// Misdirection run (#1830). The clause is TargetSpellOrAbility with
// ItemHasASingleTarget, so a two-target spell or ability is never
// offered, and an ability item is reachable. The trigger targets when it
// goes on the stack (CR 603.3d) and is dropped if nothing qualifies;
// at resolution the controller picks the new target over the original
// announce's legality check (hexproof, protection, the clause's own
// restriction), and with no other legal target the target stays. It is
// mandatory — Willbender prints no "may".
func init() {
	Register(Spec{
		OracleID:     "0aae277e-e58e-4115-b5fd-0459451e17ec",
		Name:         "Willbender",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Morph("{1}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisIsTurnedFaceUp("Willbender — change the target of target spell or ability with a single target",
				changeTheTargetEffect("Willbender — change the target")),
				TargetSpellOrAbility("target spell or ability with a single target", ItemHasASingleTarget())),
		},
	})
}
