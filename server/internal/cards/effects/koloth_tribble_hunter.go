package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Koloth, Tribble Hunter — Legendary Creature — Klingon Warrior Officer
// {3}{R}, 3/3 (Star Trek Commander, #4974):
//
//	"At the beginning of your end step, each player creates two 1/1
//	 white Tribble creature tokens with "This token can't block."
//	 All Tribbles get +1/+0 and are goaded. (They attack each combat if
//	 able and attack a player other than you if able.)"
//
// The end-step trigger is Party Dude's "each player creates", in turn
// order from the active player; each player controls their own two
// Tribbles (the printed ttrc Tribble, shared with Tribble Cluster).
//
// The second line is two statics over every Tribble creature on the
// battlefield, whoever controls it: a layer-7c +1/+0 (a lord that names
// no controller, Lord of Atlantis's reach) and a STATIC goad by Koloth's
// controller (GoadAttached's shape, CR 701.15b), which lasts exactly as
// long as Koloth is on the battlefield and has no "until your next turn"
// end. Your own Tribbles are goaded by you too, so they attack each
// combat if able. A Tribble still can't block.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01b7ced9-5dcb-45ce-bcda-ffdc1afefb2a",
		Name:         "Koloth, Tribble Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Koloth, Tribble Hunter — each player creates two Tribbles", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, p := range seatsFromActive(g) {
					if err := (CreateToken{Controller: p, Template: TribbleToken(), N: 2}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Tribble"}}, 1, 0),
			trekB37TribblesAreGoaded(),
		},
	})
}
