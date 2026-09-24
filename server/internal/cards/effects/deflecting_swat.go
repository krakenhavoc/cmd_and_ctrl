package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deflecting Swat — Instant {2}{R}:
//
//	"If you control a commander, you may cast this spell without
//	 paying its mana cost.
//	 You may choose new targets for target spell or ability."
//
// The red half of the Commander Legends free-spell cycle, and the
// card the "Stack-item retarget" seam row was opened for (#294). It
// is the reason the row exists and the reason the row was undercounted:
// "you may choose new targets" and "change the target … with a single
// target" are printed across a whole family of red and blue instants,
// and none of them could ship.
//
// Two halves, and the engine has both since #1196:
//
//   - The free cast is the same AlternativeCost.Condition shape
//     Fierce Guardianship and Flawless Maneuver already use
//     (FreeIfYouControlCommander, #988/#1004). "You control a
//     commander" is a permanent you CONTROL, not one you own: a
//     commander in the command zone does not switch it on, and one
//     you have stolen does.
//   - The redirect is CR 115.7c — "choose new targets", the shape
//     where EVERY slot may move and any number may be left alone,
//     which is why this card can rescue a Blasphemous-Act-sized
//     multi-target spell and Bolt Bend cannot.
//
// The prompt goes to the SWAT's controller while legality is judged
// for the redirected spell: a Swat on an opponent's "destroy target
// creature you control" can only move it to another creature THAT
// OPPONENT controls, and protection is tested against the spell being
// redirected rather than against the Swat.
//
// The "OR ABILITY" half landed with #1211 and the caveat this card
// shipped with is gone. It cost one constructor: `TargetSpellOrAbility`
// enumerates the spell cards in the stack zone and the ability items in
// StackMeta into one legal set, and `RetargetStackItemForEffect` was
// written over StackMeta from the start — so redirecting an opponent's
// triggered ability is the same call the spell half already made. A
// Swat now answers the Bolt and the Deathrite.
//
// A spell or ability with no targets is still a legal thing to point
// this at (nothing in the printed text says otherwise) and redirecting
// it does nothing, which is what ChangeTargets absorbs.
func init() {
	Register(Spec{
		OracleID:     "ae120613-97d6-4393-b39d-c3e6c076f5d6",
		Name:         "Deflecting Swat",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrAbility("target spell or ability"),
		AlternativeCosts: []game.AlternativeCost{
			FreeIfYouControlCommander("Cast without paying its mana cost (you control a commander)"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ChangeTargets{
				StackID:  item.Targets[0].ID,
				Policy:   game.RetargetChooseNew,
				Optional: true,
				Reason:   "Deflecting Swat — choose a new target",
			}.Apply(ctx)
		},
	})
}
