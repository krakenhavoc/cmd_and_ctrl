package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b08_helpers.go — shared shapes for the GAP2 batch 8 cards
// (#5948), over the #6092 seams.

// ProcessCardExiledWithThis is "Put a <label> exiled with <this> into its
// owner's graveyard" as a CR 602 activated ability's cost (Shelob, Dread
// Weaver, #6092): the processing component (ProcessCards) read against the
// cards exiled WITH the paying object (CR 607.2a), whoever owns them
// (game.ExileCost.ExiledWithSource). The activator names the card on
// exile_ids; it goes to its owner's graveyard as the cost is paid.
func ProcessCardExiledWithThis(label string, match func(game.Card) bool) game.AbilityCost {
	c := ProcessCards(1, label, match)
	c.ExileCards.ExiledWithSource = true
	return c
}

// OwnedByNthOpponent passes for a card owned by the n-th seat (1-based)
// after the chooser's, in turn order: the graveyard twin of
// ControlledByNthOpponent, since a card in a graveyard is in its OWNER's
// graveyard (CR 400.3). Eliminated seats keep their place.
func OwnedByNthOpponent(n int) CardPredicate {
	return func(g *game.Game, chooser uuid.UUID, c game.Card) bool {
		seats := len(g.Seats)
		if n < 1 || n >= seats {
			return false
		}
		at := -1
		for i, p := range g.Seats {
			if p != nil && p.ID == chooser {
				at = i
				break
			}
		}
		if at < 0 {
			return false
		}
		p := g.Seats[(at+n)%seats]
		return p != nil && c.Owner == p.ID
	}
}

// PerOpponentGraveyardClauses is "for each opponent, up to one target
// <noun> from that player's graveyard" (Tasha, the Witch Queen): one clause
// per possible opponent, each "up to one" and bound to that opponent's
// graveyard by turn order (PerOpponentClauses' shape), so the CR 608.2b
// re-check asks whether the card is still in THAT player's graveyard.
func PerOpponentGraveyardClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	var out []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		out = append(out, TargetCardInGraveyard("up to one target "+noun+" from the graveyard of the "+NthOpponentWords[n]+" opponent after you in turn order",
			append([]CardPredicate{OwnedByNthOpponent(n)}, what...)...).WithCount(0, 1))
	}
	return Clauses(out[0], out[1:]...)
}

// MarkTriggeringSpellExilesLinked marks the spell a cast trigger fired on
// to be exiled, as it resolves, instead of going to a graveyard, linked to
// the trigger's source object (CR 607.2a; Game.MarkSpellExilesLinkedForEffect):
// "exile it instead of putting it into a graveyard as it resolves" (Rod of
// Absorption, River Song's Diary). A source with no object reference links
// nothing.
func MarkTriggeringSpellExilesLinked(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := ctx.Trigger().Event.CardID
	ref, ok := ctx.SourceRef()
	if !ok || spell == uuid.Nil {
		return nil
	}
	g.MarkSpellExilesLinkedForEffect(spell, game.PermissionCardRef(ref))
	return nil
}
