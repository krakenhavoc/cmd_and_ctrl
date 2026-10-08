package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// --- #2680 part 1: putting a card from hand onto the battlefield -------

// putChoice is the prompt effects.PutFromHandOntoBattlefield queues:
// a choose_cards over seat 0's own hand that says where the pick goes.
func putChoice(id, dest string, opts ...protocol.CardView) protocol.PendingChoiceView {
	c := handChoice(id, "Uro — you may put a land card from your hand onto the battlefield", 0, 1, opts...)
	c.ChooseDestination = dest
	return c
}

func karoo(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := land(id, controller, opts...)
	c.Name, c.TypeLine = "Simic Growth Chamber", "Land"
	c.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: "{G}{U}"}}
	return c
}

// Seq 234 of review game 1: Uro's "you may put a land card from your
// hand onto the battlefield", priced as a discard of the land, so the
// bot chose nothing. A land put onto the battlefield is a land drop
// the bot did not spend.
func TestThePutFromHandPutsTheLand(t *testing.T) {
	const choiceID = "choice-put"
	chamber := karoo(cardID(1), 0)
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	for _, tc := range []struct {
		name, dest string
		bf         []protocol.CardView
	}{
		{"early, lands wanted", "battlefield", []protocol.CardView{land(cardID(30), 0), land(cardID(31), 0), land(cardID(32), 0)}},
		{"late, past RampWantCap", "battlefield", lands(8)},
		{"enters tapped", "battlefield_tapped", []protocol.CardView{land(cardID(30), 0)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := newView([]protocol.PlayerView{newSeat(0, withHand(chamber, bear)), newSeat(1)},
				withBattlefield(tc.bf...), withChoice(putChoice(choiceID, tc.dest, chamber)))
			in := input(0, v,
				choiceMove(t, 0, choiceID, "choose nothing", ids()),
				choiceMove(t, 0, choiceID, "put the Chamber", ids(cardID(1))),
			)
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != "put the Chamber" {
				t.Fatalf("chose %q, want the land put onto the battlefield", got)
			}
			off := heuristic.DefaultConfig()
			off.PricePutsFromHand = false
			if got := chose(t, in, decide(t, heuristic.NewWithConfig(off), in)); got != "choose nothing" {
				t.Fatalf("PricePutsFromHand off chose %q, want the old discard price's choose nothing", got)
			}
		})
	}
}

// Among two lands, the one the hand's colours want is put (#2677's
// landColorFit), whatever the enumerator's order.
func TestThePutFromHandPutsTheLandTheHandWants(t *testing.T) {
	const choiceID = "choice-put"
	mountain := land(cardID(1), 0)
	forest := land(cardID(2), 0)
	forest.Name, forest.TypeLine = "Forest", "Basic Land — Forest"
	forest.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: "{G}"}}
	elf := creature(cardID(3), 0, "Elf", 1, 1)
	elf.ManaCost = "{G}{G}"
	v := newView([]protocol.PlayerView{newSeat(0, withHand(mountain, forest, elf)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)), withChoice(putChoice(choiceID, "battlefield", mountain, forest)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "put the Mountain", ids(cardID(1))),
		choiceMove(t, 0, choiceID, "put the Forest", ids(cardID(2))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "put the Forest" {
		t.Fatalf("chose %q, want the Forest the Elf needs", got)
	}
}

// A choose_cards over the bot's own hand with no destination is still
// a discard: the loot's half keeps the old price (#798).
func TestAHandPickWithNoDestinationIsStillADiscard(t *testing.T) {
	const choiceID = "choice-discard"
	mountain := land(cardID(1), 0)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(mountain)), newSeat(1)},
		withBattlefield(sixLands()...),
		withChoice(handChoice(choiceID, "Discard up to one card", 0, 1, mountain)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "discard the Mountain", ids(cardID(1))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose nothing" {
		t.Fatalf("chose %q, want to keep the land", got)
	}
}

// --- #2680 part 2: own_permanents picks ----------------------------------

func ownPermanentsChoice(id string, lo, hi int, opts ...protocol.CardView) protocol.PendingChoiceView {
	return protocol.PendingChoiceView{
		ID: id, Kind: "own_permanents", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Reason: "Simic Growth Chamber — return a land you control to its owner's hand",
		Options: opts, ChooseMin: lo, ChooseMax: hi, Count: hi,
	}
}

// Seq 238 of review game 1: the karoo's return scored every answer 0,
// "unrecognised choice", and the enumerator's order chose. Return the
// tapped basic: not the karoo, which makes two mana, and not the
// untapped land, whose mana is still to spend this turn. The tapped
// basic is offered last here, so a test that passed on order would fail.
func TestTheKarooReturnsATappedLand(t *testing.T) {
	const choiceID = "choice-karoo"
	chamber := karoo(cardID(1), 0, tapped())
	orchard := land(cardID(2), 0)
	orchard.Name, orchard.TypeLine = "Exotic Orchard", "Land"
	orchard.ManaAbilities = []protocol.ManaAbilityView{{Index: 0, TapCost: true}}
	island := land(cardID(3), 0, tapped())
	island.Name = "Island"
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(chamber, orchard, island),
		withChoice(ownPermanentsChoice(choiceID, 1, 1, chamber, orchard, island)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "return the Chamber", ids(cardID(1))),
		choiceMove(t, 0, choiceID, "return the Orchard", ids(cardID(2))),
		choiceMove(t, 0, choiceID, "return the Island", ids(cardID(3))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "return the Island" {
		t.Fatalf("chose %q, want the tapped Island", got)
	}
	off := heuristic.DefaultConfig()
	off.PriceOwnPermanentPicks = false
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(off), in)); got != "return the Chamber" {
		t.Fatalf("PriceOwnPermanentPicks off chose %q, want the enumerator's first", got)
	}
}

// "Any number" (Scapeshift) has no sign the rule can read: naming more
// lands is the point of the card. It keeps the enumerator's order.
func TestAnAnyNumberOwnPermanentsPickKeepsTheEnumeratorsOrder(t *testing.T) {
	const choiceID = "choice-scapeshift"
	a, b := land(cardID(1), 0), land(cardID(2), 0)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(a, b), withChoice(ownPermanentsChoice(choiceID, 0, 2, a, b)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "both", ids(cardID(1), cardID(2))),
		choiceMove(t, 0, choiceID, "none", ids()),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "both" {
		t.Fatalf("chose %q, want the enumerator's first", got)
	}
}

// --- #2678: an extra land drop -------------------------------------------

func oracle(id string) protocol.CardView {
	c := creature(id, 0, "Oracle of Mul Daya", 2, 2)
	c.TypeLine, c.ManaCost = "Creature — Elf Shaman", "{3}{G}"
	c.Purpose = &protocol.PurposeView{ExtraLandDrops: 1}
	return c
}

// Seq 292 of review game 1: turn 6, the land drop used, a land in hand
// and Oracle of Mul Daya castable. Oracle was priced as a 2/2. With the
// extra drop priced it is worth one more land this turn on top.
func TestAnExtraLandDropIsPriced(t *testing.T) {
	or := oracle(cardID(1))
	grove := land(cardID(2), 0)
	used := func(p *protocol.PlayerView) { p.LandDropsPerTurn, p.LandsPlayedThisTurn = 1, 1 }
	run := func(hand []protocol.CardView, opts ...seatOpt) (on, off float64) {
		t.Helper()
		v := newView([]protocol.PlayerView{newSeat(0, append([]seatOpt{withHand(hand...)}, opts...)...), newSeat(1)},
			withBattlefield(lands(6)...), withTurn(6, 0, "precombat_main"))
		in := input(0, v, passMove(0), castMove(t, 0, or.InstanceID, "Cast Oracle"))
		base := heuristic.DefaultConfig()
		base.PriceExtraLandDrops = false
		return rankValue(t, heuristic.New(), in, "Cast Oracle"), rankValue(t, heuristic.NewWithConfig(base), in, "Cast Oracle")
	}
	cfg := heuristic.DefaultConfig()

	// The drop is used and a land waits: one more land this turn, plus
	// the recurring value (six sources, under RampWantCap).
	on, off := run([]protocol.CardView{or, grove}, used)
	if want := off + cfg.Weights.ManaSource + cfg.ExtraLandDropRecurring; on < want-1e-9 {
		t.Errorf("drop used, land in hand: Oracle %+.2f, want at least %+.2f (off %+.2f)", on, want, off)
	}
	// No land in hand: only the recurring value.
	on, off = run([]protocol.CardView{or}, used)
	if !nearly(on, off+cfg.ExtraLandDropRecurring) {
		t.Errorf("no land in hand: Oracle %+.2f, want %+.2f", on, off+cfg.ExtraLandDropRecurring)
	}
	// The drop unused and one land in hand: the normal drop plays it.
	on, off = run([]protocol.CardView{or, grove}, func(p *protocol.PlayerView) { p.LandDropsPerTurn = 1 })
	if !nearly(on, off+cfg.ExtraLandDropRecurring) {
		t.Errorf("drop unused: Oracle %+.2f, want %+.2f", on, off+cfg.ExtraLandDropRecurring)
	}
}
