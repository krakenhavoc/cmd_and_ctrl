package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_by_opponent_test.go is ADR 0153 §5 (#3168): a discard an
// opponent forces is priced by what the card does when it is discarded
// that way.

// balothInHand is an Obstinate Baloth as the server projects it into its
// owner's hand: `discard_by_opponent` declared by the catalog.
func balothInHand(id string) protocol.CardView {
	c := creature(id, 0, "Obstinate Baloth", 4, 4)
	c.ManaCost = "{2}{G}{G}"
	c.DiscardByOpponent = "battlefield"
	return c
}

// metrognomeInHand is Metrognome with its trigger's declared payoff.
func metrognomeInHand(id string) protocol.CardView {
	c := spell(id, 0, "Metrognome", "{4}")
	c.TypeLine = "Artifact"
	c.DiscardByOpponent = "trigger"
	c.AbilityRows = []protocol.AbilityRowView{{
		Kind: "triggered", Label: "discarded: four Gnomes",
		Purpose: &protocol.PurposeView{OnOpponentDiscard: &protocol.DiscardPayoffView{Any: true, Tokens: 4}},
	}}
	return c
}

// mindRotPosition is the owner's position: a Mind Rot is resolving, and
// the bot holds `held` and a land, so it must pick one to pitch. `by` is
// the seat that cast it. As in a real game, the resolving spell has left
// the stack and sits in its caster's graveyard when the prompt opens, so
// the prompt's source is that card.
func mindRotPosition(t *testing.T, by int, held protocol.CardView) (protocol.GameView, []legal.Move) {
	t.Helper()
	const choiceID = "choice-discard"
	forest := land(cardID(2), 0)
	rot := spell(cardID(9), by, "Mind Rot", "{2}{B}")
	ch := handChoice(choiceID, "Mind Rot — discard a card", 1, 1, held, forest)
	ch.Source = rot.InstanceID
	seats := []protocol.PlayerView{newSeat(0, withHand(held, forest)), newSeat(1)}
	if by == 0 {
		seats[0] = newSeat(0, withHand(held, forest), withGraveyard(rot))
	} else {
		seats[1] = newSeat(1, withGraveyard(rot))
	}
	v := newView(seats, withBattlefield(manaLands(8, 0, 100)...), withChoice(ch))
	moves := []legal.Move{
		choiceMove(t, 0, choiceID, "pitch the held card", map[string]any{"card_ids": []string{held.InstanceID}}),
		choiceMove(t, 0, choiceID, "pitch the land", map[string]any{"card_ids": []string{forest.InstanceID}}),
	}
	return v, moves
}

// With a Baloth and a land in hand, an opponent's Mind Rot: the bot
// pitches the Baloth, which comes back as a 4/4.
func TestTheBotPitchesABalothToAnOpponentsMindRot(t *testing.T) {
	v, moves := mindRotPosition(t, 1, balothInHand(cardID(1)))
	in := input(0, v, moves...)

	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the held card" {
		t.Fatalf("chose %q, want the Baloth pitched", got)
	}
}

// The same hand, but the bot's own spell is making it discard: the Baloth
// goes to the graveyard, so it is not pitched.
func TestTheBotKeepsABalothWhenItIsItsOwnDiscard(t *testing.T) {
	v, moves := mindRotPosition(t, 0, balothInHand(cardID(1)))
	in := input(0, v, moves...)

	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the land" {
		t.Fatalf("chose %q, want the land pitched when the discard is its own", got)
	}
}

// A card with no declaration is priced as before even for an opponent's
// discard, so the same Baloth without the declaration is kept.
func TestTheBotPricesAnUndeclaredCardAsBefore(t *testing.T) {
	baloth := balothInHand(cardID(1))
	baloth.DiscardByOpponent = ""
	v, moves := mindRotPosition(t, 1, baloth)
	in := input(0, v, moves...)

	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pitch the land" {
		t.Fatalf("chose %q, want the land pitched", got)
	}
}

// A trigger card is cheaper by what its trigger pays, in the units of a
// discard payoff: four tokens at TokenWeight each.
func TestATriggerCardIsCheaperByWhatItPays(t *testing.T) {
	forced, moves := mindRotPosition(t, 1, metrognomeInHand(cardID(1)))
	own, _ := mindRotPosition(t, 0, metrognomeInHand(cardID(1)))
	pol := heuristic.New()
	inForced, inOwn := input(0, forced, moves...), input(0, own, moves...)

	gap := rankValue(t, pol, inForced, moves[0].Label) - rankValue(t, pol, inOwn, moves[0].Label)
	if want := 4 * heuristic.DefaultConfig().TokenWeight; !nearly(gap, want) {
		t.Errorf("pitching Metrognome is worth %.3f more to an opponent's Mind Rot, want four tokens' %.3f", gap, want)
	}
}
