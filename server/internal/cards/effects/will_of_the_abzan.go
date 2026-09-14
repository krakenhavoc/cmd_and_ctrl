package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Will of the Abzan — Sorcery {3}{B} (EDHREC rank 2605):
//
//	"Choose one. If you control a commander as you cast this spell,
//	 you may choose both instead.
//	 • Any number of target opponents each sacrifice a creature with
//	   the greatest power among creatures that player controls and
//	   lose 3 life.
//	 • Return target creature card from your graveyard to the
//	   battlefield."
//
// An edict that takes the biggest creature, or a reanimation, and
// with a commander out, both. The first mode targets any number of
// opponents; each still-legal one chooses a creature with the
// greatest power among their own (GreatestPowerYouControl evaluated
// for the chooser through the sacrifice prompt, so ties are theirs
// to break — b24PlayerSacrificesGreatestPowerCreatureAndLosesLife)
// and loses 3 life, in announce order. The second returns the chosen
// creature card under its owner's control, which is the caster's:
// the clause says "your graveyard".
//
// SANDBOX GAP, weaker than printed — the Will of the Mardu posture:
// the "choose both" rider is not offered. Each mode carries its own
// target, and a modal spec whose maximum is above one may carry a
// target on at most one option (per-mode target slots are the open
// multi-target work), so the card is "choose one" whether or not a
// commander is on the board. Never stronger: the wider choice is
// simply absent.
func init() {
	Register(Spec{
		OracleID:     "1ae29791-aa7c-4050-bf72-dd0f739b11b8",
		Name:         "Will of the Abzan",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Choosing both modes when you control a commander isn't implemented — you always choose one."},
		Modes: ChooseOne(
			Mode("Any number of target opponents each sacrifice a creature with the greatest power among creatures that player controls and lose 3 life.",
				TargetPlayer("any number of target opponents", Opponent()).WithCount(1, 0)),
			Mode("Return target creature card from your graveyard to the battlefield.",
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := b24PlayerSacrificesGreatestPowerCreatureAndLosesLife(ctx.Game, item, t.ID, 3); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
