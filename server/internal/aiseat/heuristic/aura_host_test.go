package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// aura_host_test.go — #3158, CR 303.4f: what an Aura put onto the
// battlefield without being cast enchants (aura_host.go).

func withGraveyardCards(cards ...protocol.CardView) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Graveyard = protocol.ZoneView{Kind: "graveyard", Owner: p.ID, Count: len(cards), Cards: cards}
	}
}

func TestAuraHostChoosesSensibly(t *testing.T) {
	const choiceID = "aura-host-1"
	prompt := protocol.PendingChoiceView{
		ID: choiceID, Kind: "aura_host", Chooser: seatID(0).String(), Reason: "Choose what it enchants",
	}

	// Animate Dead: a card in a graveyard is a creature coming back
	// under the bot's control, so the best creature wins, whoever owns it.
	t.Run("a graveyard host is the best creature in any graveyard", func(t *testing.T) {
		small := creature(cardID(1), 0, "My Squire", 1, 1)
		big := creature(cardID(2), 1, "Their Dragon", 6, 6)
		v := newView([]protocol.PlayerView{
			newSeat(0, withGraveyardCards(small)),
			newSeat(1, withGraveyardCards(big)),
		}, withChoice(prompt))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "enchant my squire", map[string]any{"target": cardTarget(cardID(1))}),
			choiceMove(t, 0, choiceID, "enchant their dragon", map[string]any{"target": cardTarget(cardID(2))}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "enchant their dragon" {
			t.Fatalf("chose %q", got)
		}
	})

	// Pacifism: read like an Aura spell's target, a harmful Aura goes on
	// an opponent's creature rather than the bot's own.
	t.Run("a battlefield host is an opponent's creature", func(t *testing.T) {
		mine := creature(cardID(10), 0, "Mine", 3, 3)
		theirs := creature(cardID(20), 1, "Theirs", 3, 3)
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(mine, theirs), withChoice(prompt))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "enchant mine", map[string]any{"target": cardTarget(cardID(10))}),
			choiceMove(t, 0, choiceID, "enchant theirs", map[string]any{"target": cardTarget(cardID(20))}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "enchant theirs" {
			t.Fatalf("chose %q", got)
		}
	})

	// "Enchant creature you control": only the bot's own are offered, and
	// the one worth most carries it.
	t.Run("among its own creatures, the best one", func(t *testing.T) {
		small := creature(cardID(30), 0, "Small", 1, 1)
		big := creature(cardID(31), 0, "Big", 5, 5)
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(small, big), withChoice(prompt))
		in := input(0, v,
			choiceMove(t, 0, choiceID, "enchant small", map[string]any{"target": cardTarget(cardID(30))}),
			choiceMove(t, 0, choiceID, "enchant big", map[string]any{"target": cardTarget(cardID(31))}),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != "enchant big" {
			t.Fatalf("chose %q", got)
		}
	})
}
