package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assassin's Trophy — Instant {B}{G}:
//
//	"Destroy target permanent an opponent controls. Its controller
//	may search their library for a basic land card, put it onto the
//	battlefield, then shuffle."
//
// Two mana, any permanent, no restriction — the drawback is the
// replacement land, and it is a real one in Commander where the
// victim is usually already at the mana they need.
//
// The land arrives UNTAPPED (the printed text does not say tapped),
// which is what separates this from Beast Within's 3/3: a Trophy on
// turn three is nearly free for the opponent.
//
// SANDBOX SIMPLIFICATION — the "may" is not offered. The opponent
// always searches, because a yes/no prompt on someone else's
// resolution needs the PayUnless-style deferred-choice plumbing and
// this effect has no such hook. The forced search is the WEAKER
// direction for the caster (it hands the victim a land they might
// have declined), never the stronger one. It also means a victim
// who would rather keep their library unshuffled — the Brainstorm /
// tutor-on-top case — cannot decline the shuffle.
//
// The victim's search picks the first basic in library order, the
// same deterministic pick every other SearchLibrary card in the
// catalog makes.
func init() {
	Register(Spec{
		OracleID: "ac10d218-f9a6-4058-9cda-a15ca1b0b7b5",
		Name:     "Assassin's Trophy",
		Targets:  TargetPermanent("target permanent an opponent controls", OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:    victim,
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
			}.Apply(ctx)
		},
	})
}
