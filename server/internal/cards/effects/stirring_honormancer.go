package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stirring Honormancer — Creature — Rhino Bard {2}{W}{W/B}{B}, 4/5 (Secrets
// of Strixhaven, #4456):
//
//	"When this creature enters, look at the top X cards of your library,
//	 where X is the number of creatures you control. Put one of those cards
//	 into your hand and the rest into your graveyard."
//
// Wail of the Forgotten's look-and-take: a private look, a mandatory pick of
// exactly one card (the prompt's floor and ceiling are both 1), and the rest
// into the graveyard (takeRestGraveyard, not a mill). X is counted as the
// ability resolves, so the Honormancer itself counts while it is still on the
// battlefield; a library of fewer than X cards is looked at whole.
//
// No simplification.
func init() {
	const label = "Stirring Honormancer — look at the top X cards, put one into your hand and the rest into your graveyard"
	Register(Spec{
		OracleID:     "3bcf54f8-fcd7-4ec0-b011-92b367411cdb",
		Name:         "Stirring Honormancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				x := g.CountControlledMatchingForEffect(player, game.PermanentQuery{Types: []string{"creature"}})
				if x <= 0 {
					return nil
				}
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, x),
					Max:    1,
					Label:  label,
					Key:    takeRestGraveyard,
				}.Apply(ctx)
			}),
		},
	})
}
