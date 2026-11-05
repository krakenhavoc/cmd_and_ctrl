package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// 70,000 Light-Years from Home — Sorcery {1}{W} (Star Trek, #4973):
//
//	"Exile target creature or Spacecraft. Its controller may search
//	 their library for a basic land card, reveal it, put it into their
//	 hand, then shuffle."
//
// "Spacecraft" is the artifact subtype, so a Spacecraft below its
// station threshold (not yet a creature) is still a legal target.
//
// "Its controller" is read before the exile: the permanent's controller
// as it last existed on the battlefield (CR 608.2h), which is not always
// its owner. The search is the victim's "may" (Assassin's Trophy's
// shape), with the land going to the hand instead of the battlefield.
// It happens whether or not the exile did (a commander that went to the
// command zone): the second sentence does not depend on the first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9356aee1-3d2b-4d17-8863-663e1ddf56f3",
		Name:         "70,000 Light-Years from Home",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or Spacecraft", Or(Creature(), HasSubtype("Spacecraft"))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:    victim,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Optional:  true,
				Reason:    "70,000 Light-Years from Home — you may search for a basic land card",
			}.Apply(ctx)
		},
	})
}
