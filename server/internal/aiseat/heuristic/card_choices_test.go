package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// card_choices_test.go pins #2677 (a land search takes the colour the
// hand needs) and #2691 (a discard gives up the card furthest from
// castable). Each test fails with its knob off.

// manaLand is a land of the given name and type line that taps for
// `produced`.
func manaLand(id string, controller int, name, typeLine, produced string) protocol.CardView {
	c := land(id, controller)
	c.Name, c.TypeLine = name, typeLine
	c.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: produced}}
	return c
}

// searchFor builds a search_library window over `options`, with one
// answer per option labelled by its name, in the order given.
func searchFor(t *testing.T, seat protocol.PlayerView, bf []protocol.CardView, options ...protocol.CardView) aiseat.Input {
	t.Helper()
	const choiceID = "fetch"
	v := newView([]protocol.PlayerView{seat, newSeat(1)},
		withBattlefield(bf...), withTurn(1, 0, "precombat_main"),
		withChoice(protocol.PendingChoiceView{
			ID: choiceID, Kind: "search_library", Chooser: seatID(0).String(),
			Reason: "Fetchland — choose a land to put onto the battlefield", Options: options,
		}))
	in := input(0, v)
	for i := range options {
		in.Moves = append(in.Moves, choiceMove(t, 0, choiceID, "take "+options[i].Name,
			map[string]any{"card_ids": []string{options[i].InstanceID}}))
	}
	return in
}

// Review game 8a9f18d7, seq 40: turn 1, Polluted Delta resolves into
// Island, Breeding Pool, Tropical Island. Every land in the hand needs
// G and nothing on the battlefield makes it. The bot took the Island.
func TestALandSearchTakesTheColourTheHandNeeds(t *testing.T) {
	hand := withHand(
		sorcery(cardID(1), 0, "Rampant Growth", "{1}{G}", nil),
		sorcery(cardID(2), 0, "Explosive Vegetation", "{3}{G}", nil),
		creature(cardID(3), 0, "Oracle of Mul Daya", 2, 2, func(c *protocol.CardView) { c.ManaCost = "{3}{G}" }),
	)
	in := searchFor(t, newSeat(0, hand), nil,
		manaLand(cardID(10), 0, "Island", "Basic Land — Island", "{U}"),
		manaLand(cardID(11), 0, "Breeding Pool", "Land — Forest Island", "{G|U}"),
		manaLand(cardID(12), 0, "Tropical Island", "Land — Forest Island", "{G|U}"),
		manaLand(cardID(13), 0, "Island", "Basic Land — Island", "{U}"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take Breeding Pool" && got != "take Tropical Island" {
		t.Errorf("chose %q, want a land that makes G (#2677)", got)
	}
	off := heuristic.DefaultConfig()
	off.LandColorNeed = 0
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(off), in)); got != "take Island" {
		t.Errorf("LandColorNeed off chose %q, want the first land offered", got)
	}
}

// When two lands meet the hand equally, the one that makes more colours
// is taken: a Forest and a Breeding Pool both give a G-hungry hand its
// first G, and the Pool also makes U.
func TestALandSearchPrefersADualOverABasic(t *testing.T) {
	hand := withHand(sorcery(cardID(1), 0, "Rampant Growth", "{1}{G}", nil))
	in := searchFor(t, newSeat(0, hand), nil,
		manaLand(cardID(10), 0, "Forest", "Basic Land — Forest", "{G}"),
		manaLand(cardID(11), 0, "Breeding Pool", "Land — Forest Island", "{G|U}"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take Breeding Pool" {
		t.Errorf("chose %q, want the dual", got)
	}
	off := heuristic.DefaultConfig()
	off.LandColorNeed = 0
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(off), in)); got != "take Forest" {
		t.Errorf("LandColorNeed off chose %q, want the first land offered", got)
	}
}

// A colour the bot already makes is worth less than one it lacks: with
// two Islands out, a Forest beats a third Island for a hand that needs
// both colours.
func TestALandSearchTakesTheMissingColourOverAnotherOfOneItHas(t *testing.T) {
	hand := withHand(
		sorcery(cardID(1), 0, "Rampant Growth", "{1}{G}", nil),
		spell(cardID(2), 0, "Counterspell", "{U}{U}"),
	)
	bf := []protocol.CardView{
		manaLand(cardID(20), 0, "Island", "Basic Land — Island", "{U}"),
		manaLand(cardID(21), 0, "Island", "Basic Land — Island", "{U}"),
	}
	in := searchFor(t, newSeat(0, hand), bf,
		manaLand(cardID(10), 0, "Island", "Basic Land — Island", "{U}"),
		manaLand(cardID(11), 0, "Forest", "Basic Land — Forest", "{G}"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take Forest" {
		t.Errorf("chose %q, want the Forest", got)
	}
}

// Review game 06e98afa, seq 156: turn 4 cleanup on two lands, after two
// missed drops. The bot discarded Counterspell, the one card it could
// nearly cast, and kept Angrath's Marauders, five lands away.
func TestTheCleanupDiscardGivesUpTheCardFurthestFromCastable(t *testing.T) {
	big := func(cost string) func(*protocol.CardView) {
		return func(c *protocol.CardView) { c.ManaCost = cost }
	}
	cards := []protocol.CardView{
		spell(cardID(1), 0, "Counterspell", "{U}{U}"),
		creature(cardID(2), 0, "Angrath's Marauders", 6, 6, big("{5}{R}{R}")),
		creature(cardID(3), 0, "Goldspan Dragon", 4, 4, big("{3}{R}{R}"), keywords("flying", "haste")),
		creature(cardID(4), 0, "Solphim", 2, 5, big("{2}{R}{R}")),
		creature(cardID(5), 0, "Bear A", 2, 2),
		creature(cardID(6), 0, "Bear B", 2, 2),
		creature(cardID(7), 0, "Bear C", 2, 2),
		creature(cardID(8), 0, "Bear D", 2, 2),
	}
	bf := []protocol.CardView{
		manaLand(cardID(20), 0, "Island", "Basic Land — Island", "{U}"),
		manaLand(cardID(21), 0, "Mountain", "Basic Land — Mountain", "{R}"),
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(cards...)), newSeat(1)},
		withBattlefield(bf...), withTurn(4, 0, "cleanup"))
	in := input(0, v)
	for i := range cards {
		in.Moves = append(in.Moves, discardMove(t, 0, "discard "+cards[i].Name, cards[i].InstanceID))
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard Angrath's Marauders" {
		t.Errorf("chose %q, want the Marauders: five mana short (#2691)", got)
	}
	off := heuristic.DefaultConfig()
	off.DiscardByDistance = false
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(off), in)); got != "discard Counterspell" {
		t.Errorf("DiscardByDistance off chose %q, want the Counterspell (the pre-#2691 price)", got)
	}
}

// A land is worth what it brings the hand closer to castable. Early, on
// two lands with a three- and a four-drop in hand, the Mountain is next
// turn's drop and is kept over a castable two-mana spell; late, with
// eight lands out and everything castable, it is the card to pitch.
func TestTheDiscardKeepsTheLandTheHandNeeds(t *testing.T) {
	mountains := func(n, from int) []protocol.CardView {
		var out []protocol.CardView
		for i := 0; i < n; i++ {
			out = append(out, manaLand(cardID(from+i), 0, "Mountain", "Basic Land — Mountain", "{R}"))
		}
		return out
	}
	cards := []protocol.CardView{
		manaLand(cardID(1), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		spell(cardID(2), 0, "Two Mana Spell", "{1}{R}"),
		creature(cardID(3), 0, "Three Drop", 3, 3, func(c *protocol.CardView) { c.ManaCost = "{2}{R}" }),
		creature(cardID(4), 0, "Four Drop", 4, 4, func(c *protocol.CardView) { c.ManaCost = "{3}{R}" }),
	}
	for _, tc := range []struct {
		name  string
		lands int
		want  string
	}{
		{"early", 2, "discard Two Mana Spell"},
		{"late", 8, "discard Mountain"},
	} {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(cards...)), newSeat(1)},
			withBattlefield(mountains(tc.lands, 20)...), withTurn(4, 0, "cleanup"))
		in := input(0, v,
			discardMove(t, 0, "discard Mountain", cardID(1)),
			discardMove(t, 0, "discard Two Mana Spell", cardID(2)),
		)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("%s: chose %q, want %q", tc.name, got, tc.want)
		}
	}
}

// Below RampWantCap sources a land is a future land drop even when the
// hand is castable without it: with four lands out and two Mountains in
// hand, the bot pitches the two-mana spell, not the second Mountain.
func TestTheDiscardKeepsASpareLandWhileShortOfSevenMana(t *testing.T) {
	var bf []protocol.CardView
	for i := 0; i < 4; i++ {
		bf = append(bf, manaLand(cardID(20+i), 0, "Mountain", "Basic Land — Mountain", "{R}"))
	}
	cards := []protocol.CardView{
		manaLand(cardID(1), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		manaLand(cardID(2), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		spell(cardID(3), 0, "Two Mana Spell", "{1}{R}"),
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(cards...)), newSeat(1)},
		withBattlefield(bf...), withTurn(5, 0, "cleanup"))
	in := input(0, v,
		discardMove(t, 0, "discard Mountain", cardID(1)),
		discardMove(t, 0, "discard Two Mana Spell", cardID(3)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard Two Mana Spell" {
		t.Errorf("chose %q, want the spell", got)
	}
	noFloor := heuristic.DefaultConfig()
	noFloor.DiscardLandFloor = 0
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(noFloor), in)); got != "discard Mountain" {
		t.Errorf("DiscardLandFloor off chose %q, want the spare Mountain", got)
	}
}

// Lands in hand are future land drops: a 5-drop with three lands out and
// two in hand is not short, and a cheaper card the hand cannot colour is.
func TestTheDiscardCountsLandsInHandAndColours(t *testing.T) {
	cards := []protocol.CardView{
		creature(cardID(1), 0, "Five Drop", 4, 4, func(c *protocol.CardView) { c.ManaCost = "{4}{R}" }),
		spell(cardID(2), 0, "Off Colour", "{1}{B}{B}"),
		manaLand(cardID(3), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		manaLand(cardID(4), 0, "Mountain", "Basic Land — Mountain", "{R}"),
	}
	bf := []protocol.CardView{
		manaLand(cardID(20), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		manaLand(cardID(21), 0, "Mountain", "Basic Land — Mountain", "{R}"),
		manaLand(cardID(22), 0, "Mountain", "Basic Land — Mountain", "{R}"),
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(cards...)), newSeat(1)},
		withBattlefield(bf...), withTurn(4, 0, "cleanup"))
	in := input(0, v,
		discardMove(t, 0, "discard the five-drop", cardID(1)),
		discardMove(t, 0, "discard the off-colour spell", cardID(2)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard the off-colour spell" {
		t.Errorf("chose %q, want the spell no source can colour", got)
	}
}
