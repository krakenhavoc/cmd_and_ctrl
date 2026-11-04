package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visionary's Dance — Sorcery {5}{U}{R} (Secrets of Strixhaven, #4456):
//
//	"Create two 3/3 blue and red Elemental creature tokens with flying.
//	 {2}, Discard this card: Look at the top two cards of your library. Put one
//	 of them into your hand and the other into your graveyard."
//
// The spell is two Elementals made together. The second ability is a hand-zone
// activated ability (Zones is the hand, and DiscardSelf is part of the cost,
// as for cycling and channel): the discard is paid with the ability, so the
// card is already in the graveyard while the ability is on the stack. Its
// resolution is Stargaze's look-and-take with a count of one: a private look
// at the top two, a mandatory pick of one for the hand, and the other put into
// the graveyard (takeRestGraveyard), not discarded or milled. A library of
// fewer than two cards is looked at whole, and a library of one puts that card
// into the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9f35490-462a-4c8f-9d8b-d8abca6596f7",
		Name:         "Visionary's Dance",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("3/3 blue and red Elemental with flying"), N: 2}.Apply(ctx)
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Discard this card: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
			Cost:    game.AbilityCost{Mana: "{2}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				player := ctx.Controller()
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
					Max:    1,
					Label:  "Visionary's Dance — put one of them into your hand and the other into your graveyard",
					Key:    takeRestGraveyard,
				}.Apply(ctx)
			},
		}},
	})
}
