package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
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

// #3161: a reanimation target is priced as the permanent it becomes,
// whoever owns the graveyard.
func TestReanimationTargetIsPricedByCreatureValue(t *testing.T) {
	spell := creature(cardID(1), 0, "Reanimate", 0, 0)
	spell.TypeLine = "Sorcery"
	spell.ManaCost = "{B}"
	spell.Purpose = &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{Slot: 0, Reanimates: true}}}

	// An expensive creature of the bot's own is discounted by cardValue;
	// the opponent's smaller one used to be worth OwnPermanentTarget.
	mineBig := creature(cardID(2), 0, "My Titan", 8, 8)
	mineBig.ManaCost = "{8}{G}{G}"
	theirs := creature(cardID(3), 1, "Their Dragon", 6, 6)
	theirs.ManaCost = "{4}{R}{R}"
	small := creature(cardID(4), 0, "My Squire", 1, 1)
	small.ManaCost = "{W}"

	castAt := func(t *testing.T, target protocol.CardView, label string) legal.Move {
		m := castMove(t, 0, spell.InstanceID, label)
		m.Params = mustJSON(t, map[string]any{"instance_id": spell.InstanceID, "from_zone": "hand",
			"targets": []map[string]string{cardTarget(target.InstanceID)}})
		return m
	}
	in := func(t *testing.T, moves ...legal.Move) aiseat.Input {
		v := newView([]protocol.PlayerView{
			newSeat(0, withHand(spell), withGraveyardCards(mineBig, small)),
			newSeat(1, withGraveyardCards(theirs)),
		}, withTurn(9, 0, "precombat_main"))
		return input(0, v, append([]legal.Move{passMove(0)}, moves...)...)
	}

	t.Run("an opponent's creature beats a small one of the bot's own", func(t *testing.T) {
		onTheirs, onSmall := castAt(t, theirs, "theirs"), castAt(t, small, "small")
		i := in(t, onTheirs, onSmall)
		pol := heuristic.NewWithConfig(heuristic.DefaultConfig())
		if a, b := rankValue(t, pol, i, "theirs"), rankValue(t, pol, i, "small"); a <= b {
			t.Errorf("their dragon %.3f, my squire %.3f; want the dragon ahead", a, b)
		}
		// Baseline: the opponent's graveyard was worth OwnPermanentTarget.
		base := heuristic.NewWithConfig(heuristic.BaselineConfig())
		if a, b := rankValue(t, base, i, "theirs"), rankValue(t, base, i, "small"); a > b {
			t.Errorf("baseline: dragon %.3f above squire %.3f; the old price should rank it below", a, b)
		}
	})

	t.Run("an expensive creature is not discounted", func(t *testing.T) {
		onBig, onTheirs := castAt(t, mineBig, "big"), castAt(t, theirs, "theirs")
		i := in(t, onBig, onTheirs)
		pol := heuristic.NewWithConfig(heuristic.DefaultConfig())
		if a, b := rankValue(t, pol, i, "big"), rankValue(t, pol, i, "theirs"); a <= b {
			t.Errorf("my 8/8 %.3f, their 6/6 %.3f; want the bigger creature ahead", a, b)
		}
	})
}
