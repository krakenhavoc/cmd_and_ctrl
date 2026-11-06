package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade of the Swarm — Creature — Insect Assassin {3}{B}, 3/1 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, choose one —
//	 • Put two +1/+1 counters on this creature.
//	 • Put target exiled card with warp on the bottom of its owner's
//	   library."
//
// Black Panther, Vanguard's modal entry trigger: the bullet is chosen
// as the trigger goes on the stack (CR 603.3c), and the second is not
// offered when no exiled card has warp. The counters go on the Blade
// only while it is still the object that entered (CounterOnThis). The
// second bullet is the answer to a warped creature waiting in exile to
// be cast again: any exiled card with warp, whoever owns it and however
// it got there, goes to the bottom of its owner's library
// (PutIntoLibrary, the tuck route).
//
// Simplification: "with warp" is read off the card's catalog entry
// (WithWarp), because the deck importer keeps no record of warp on a
// card the catalog does not cover. An exiled card with warp that is not
// in the catalog can't be chosen. Every card that can actually be warp
// cast here is in the catalog, so a warped creature in exile is always
// a legal pick.
func init() {
	trigger := WhenThisEnters("Blade of the Swarm — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Put two +1/+1 counters on this creature.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CounterOnThis{Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}),
		ModeDoing("Put target exiled card with warp on the bottom of its owner's library.",
			TargetCardInExile("target exiled card with warp", WithWarp()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return PutIntoLibrary{Card: t.ID, ToBottom: true}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "33042d12-f08b-4b3e-b106-c424ef868c13",
		Name:         "Blade of the Swarm",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The second mode can't choose an exiled card with warp that isn't automated on this site.",
		},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
