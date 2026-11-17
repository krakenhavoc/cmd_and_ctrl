package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight-Errant of Eos — Creature — Human Knight {4}{W} 4/4 (March of the
// Machine):
//
//	"Convoke
//	 When this creature enters, look at the top six cards of your library.
//	 You may reveal up to two creature cards with mana value X or less from
//	 among them, where X is the number of creatures that convoked this
//	 creature. Put the revealed cards into your hand, then shuffle."
//
// X is the creatures the cast tapped (CastProvenance.Convoked), counted as
// the trigger is built and carried on the item, so it is exact if the
// Knight has left. The pick is Thunderous Debut's reveal-and-take, with the
// library shuffled after, whatever was taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bd20d79-0861-494a-a8f2-f6ddb5088004",
		Name:         "Knight-Errant of Eos",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Tutors: 2},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       knightErrantOfEosLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, knightErrantOfEosLabel)
				item.Params.Amount = len(source.ConvokedBy())
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				look := g.LookAtTopOfLibraryForEffect(item.Controller, 6)
				return TakeFromLibraryToHand{
					Player:   item.Controller,
					Cards:    look,
					Match:    And(Creature(), ManaValueLE(item.Params.Amount)),
					Max:      2,
					Optional: true,
					Label:    knightErrantOfEosLabel,
					Key:      takeRevealThenShuffle,
				}.Apply(ctx)
			},
		}},
	})
}

const knightErrantOfEosLabel = "Knight-Errant of Eos — look at the top six, you may reveal up to two creature cards with mana value X or less and put them into your hand, then shuffle"
