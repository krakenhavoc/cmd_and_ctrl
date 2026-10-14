package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
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

// #3309: an Aura that declares what it does to its host is priced by it.
func TestAuraPurposePricesTheHost(t *testing.T) {
	aura := func(name, purpose string) protocol.CardView {
		c := creature(cardID(1), 0, name, 0, 0)
		c.TypeLine = "Enchantment — Aura"
		c.ManaCost = "{1}{G}"
		c.Purpose = &protocol.PurposeView{Aura: purpose}
		return c
	}
	mine := creature(cardID(2), 0, "My Bear", 2, 2)
	theirs := creature(cardID(3), 1, "Their Bear", 2, 2)

	castOn := func(t *testing.T, card protocol.CardView, host protocol.CardView, label string) legal.Move {
		m := castMove(t, 0, card.InstanceID, label)
		m.Params = mustJSON(t, map[string]any{"instance_id": card.InstanceID, "from_zone": "hand",
			"targets": []map[string]string{cardTarget(host.InstanceID)}})
		return m
	}

	for _, tc := range []struct {
		purpose, want string
	}{
		{"benefit", "on mine"},
		{"harm", "on theirs"},
	} {
		t.Run("cast "+tc.purpose, func(t *testing.T) {
			card := aura("Test Aura", tc.purpose)
			onMine := castOn(t, card, mine, "on mine")
			onTheirs := castOn(t, card, theirs, "on theirs")
			v := newView([]protocol.PlayerView{newSeat(0, withHand(card)), newSeat(1)},
				withBattlefield(mine, theirs), withTurn(9, 0, "precombat_main"))
			in := input(0, v, passMove(0), onMine, onTheirs)
			pol := heuristic.NewWithConfig(heuristic.DefaultConfig())
			mv, tv := rankValue(t, pol, in, onMine.Label), rankValue(t, pol, in, onTheirs.Label)
			if (tc.want == "on mine") != (mv > tv) {
				t.Errorf("%s Aura: on mine %.3f, on theirs %.3f; want %s ahead", tc.purpose, mv, tv, tc.want)
			}
			// The baseline prices the host as a target to hit.
			base := heuristic.NewWithConfig(heuristic.BaselineConfig())
			if rankValue(t, base, in, onTheirs.Label) <= rankValue(t, base, in, onMine.Label) {
				t.Errorf("baseline: want the opponent's creature priced above the bot's own")
			}
		})
	}

	for _, tc := range []struct {
		purpose, want string
	}{
		{"benefit", "enchant mine"},
		{"harm", "enchant theirs"},
	} {
		t.Run("host "+tc.purpose, func(t *testing.T) {
			prompt := protocol.PendingChoiceView{
				ID: "aura-host-2", Kind: "aura_host", Chooser: seatID(0).String(), AuraPurpose: tc.purpose,
			}
			v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
				withBattlefield(mine, theirs), withChoice(prompt))
			in := input(0, v,
				choiceMove(t, 0, prompt.ID, "enchant mine", map[string]any{"target": cardTarget(mine.InstanceID)}),
				choiceMove(t, 0, prompt.ID, "enchant theirs", map[string]any{"target": cardTarget(theirs.InstanceID)}),
			)
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
				t.Fatalf("%s Aura chose %q, want %q", tc.purpose, got, tc.want)
			}
		})
	}
}
