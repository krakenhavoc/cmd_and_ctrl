package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Servant — 3/2 Creature — Zombie Warlock for {3}{B} (EDHREC
// rank 4011):
//
//	"Menace
//	 When this creature enters, search your library for a card with
//	 mana value less than or equal to your devotion to black, reveal
//	 it, put it into your hand, then shuffle. You lose 3 life."
//
// A four-mana body that tutors for ANY card — no type restriction at
// all — with the ceiling set by how black the board is. In a
// mono-black deck the Servant routinely finds a five- or six-drop;
// in a two-colour deck it finds a Sol Ring.
//
// It is in the batch because the search's ceiling is computed at
// resolution from the board, which is a shape no other tutor in the
// catalog has: every other one takes its predicate from a fixed type
// or from X.
//
// # Devotion is counted when the trigger resolves
//
// "Your devotion to black" is the number of {B} symbols in the mana
// costs of permanents its controller controls (CR 700.5). Three
// consequences the card file has to get right:
//
//   - The Servant itself counts. It is on the battlefield by the time
//     its own enters-trigger resolves, and its cost has one {B}, so
//     the floor is 1 — the search always has something it could find.
//   - A hybrid {B/R} pip counts toward devotion to black (and to
//     red), because the symbol includes {B}. That falls out of
//     reading the parsed cost's options rather than the colour of the
//     card.
//   - Tokens contribute nothing: a token has no mana cost.
//
// A killed-in-response Servant lowers the number by one, which can be
// the difference between finding the four-drop and not.
//
// # The life loss is not conditional, and it comes after
//
// "You lose 3 life" is a separate sentence with no "if you do": the
// controller pays it whether or not the search found anything, and
// whether or not they chose to find. Loss, not damage — prevention
// and protection do nothing about it.
//
// It rides the search's continuation rather than the line below it,
// because a search that has to ask only QUEUES its prompt: a
// statement written after the Apply would run before the player had
// answered. The continuation runs in every case, including the case
// where there was nothing to find and no prompt was asked.
//
// Menace is the printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1caa7fa2-a881-4a65-a87e-12a974520c83",
		Name:            "Grim Servant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Grim Servant — tutor for a card within your devotion to black, then lose 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					limit := devotionTo(g, item.Controller, "B")
					player, source := item.Controller, ctx.Source()
					return SearchLibrary{
						Player:    player,
						Predicate: func(c game.Card) bool { return c.ManaValue() <= limit },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Grim Servant — a card with mana value at most your devotion to black, revealed, to hand",
						Then: func(g *game.Game, _ []uuid.UUID) error {
							return g.ChangePlayerLifeForEffect(source, player, -3)
						},
					}.Apply(ctx)
				}),
		},
	})
}
