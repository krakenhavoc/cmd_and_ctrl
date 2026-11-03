package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b30_helpers.go — shared pieces of Secrets of Strixhaven batch 30
// (#4456).

// socB30ToughnessAtLeast is the cost modifier's condition "creatures you
// control have total toughness N or greater" (Orysa, Tide Choreographer),
// read as the price is worked out (CR 601.2f). The spell being cast is on the
// stack, not the battlefield, so it counts for nothing here; the sum is
// current toughness (layers plus counters), and a negative toughness adds as
// itself.
func socB30ToughnessAtLeast(n int) func(q game.CostQuery) bool {
	return func(q game.CostQuery) bool {
		if q.Game == nil || q.Game.Battlefield == nil {
			return false
		}
		total := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller == q.Controller && c.IsCreature() {
				total += c.CurrentToughness()
			}
		}
		return total >= n
	}
}

// socB30RevealUntilToHandRestBottom is "reveal cards from the top of your
// library until you reveal a card matching pred. Put that card into your hand
// and the rest on the bottom of your library in a random order" (Page, Loose
// Leaf's Grandeur). The whole run is revealed; with no hit the whole library
// goes under in a random order and nothing reaches the hand.
func socB30RevealUntilToHandRestBottom(label string, pred func(game.Card) bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := ctx.Controller()
		run, hit := revealUntil(ctx, player, pred, label+" — reveal")
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  run,
			Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
			All:    true,
			Label:  label + " — put the revealed card into your hand",
			Key:    takeRestBottomRandom,
		}.Apply(ctx)
	}
}

// socB30SpellYouDontOwn is "a spell you don't own": the cast spell's
// owner is not the caster (Nita, Forum Conciliator). Ownership is the card's
// printed owner, so a spell cast from an opponent's graveyard or exile counts.
func socB30SpellYouDontOwn() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner != caster }
}
