package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Practiced Scrollsmith — Creature — Dwarf Cleric {R}{R/W}{W}, 3/2 (Secrets of
// Strixhaven, #4456):
//
//	"First strike
//	 When this creature enters, exile target noncreature, nonland card from
//	 your graveyard. Until the end of your next turn, you may cast that card."
//
// The enters trigger targets (so with no legal card it is never put on the
// stack), and a target that is illegal on resolution does nothing (CR 608.2b).
// The exile is Victor Mancha's: the card goes to exile and the controller is
// granted a cast permission over it for the window "until the end of your next
// turn" (a seat-turn duration, ADR 0063: it ends at the cleanup step of your
// next turn, CR 514.2, so this turn's remainder counts too). It is a cast only
// (CR 601.2): the card is paid for at its printed cost and cast at its own
// timing. The card stays in exile if the window passes unused. First strike is
// a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a02f9f42-4c3a-4a97-8667-68ba4617f193",
		Name:            "Practiced Scrollsmith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Practiced Scrollsmith — exile target noncreature, nonland card from your graveyard; until the end of your next turn, you may cast that card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
							Player:     item.Controller,
							Duration:   g.UntilEndOfYourNextTurnDuration(item.Controller),
							CastOnly:   true,
							Source:     item.SourceCardID,
							SourceName: "Practiced Scrollsmith",
							Label:      "Practiced Scrollsmith — cast this card",
						})
					}),
				TargetCardInGraveyard("target noncreature, nonland card from your graveyard",
					YouOwn(), Noncreature(), Nonland())),
		},
	})
}
