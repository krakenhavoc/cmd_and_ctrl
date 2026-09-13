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
// S22 — the "may" is real. The search is Optional, so the VICTIM
// gets the prompt (they are searching their own library, so they are
// the chooser) and may decline both the land and the shuffle. That
// second half matters more than it looks: a victim who has just set
// up a Brainstorm or a tutor-on-top keeps their stack by saying no.
// Until the search chooser existed, the engine forced the search,
// which erred toward the WEAKER direction for the caster but took
// the decision away from the player whose card it is.
func init() {
	Register(Spec{
		OracleID:     "ac10d218-f9a6-4058-9cda-a15ca1b0b7b5",
		Name:         "Assassin's Trophy",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent an opponent controls", OpponentControls()),
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
				Optional:  true,
				Reason:    "Assassin's Trophy — you may search for a basic land",
			}.Apply(ctx)
		},
	})
}
