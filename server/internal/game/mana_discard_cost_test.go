package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mana_discard_cost_test.go — #1213: ManaAbilityShape.DiscardCards,
// the #660 discard component with a second owner.
//
// Skirge Familiar's "Discard a card: Add {B}" is the card, and the
// point of the component is that nothing about the discard changes
// when the cost is on a mana ability: it goes through the ONE discard
// door, so EventDiscardCard fires once per card, the CR 614 window
// runs over the exit and madness sees it — and, because a mana ability
// resolves immediately with no stack (CR 605.3b), the move settles
// without a prompt.

// pushDiscardManaSource seats a creature with a mana ability whose
// only cost is a discard.
func pushDiscardManaSource(g *Game, owner *Player, cost *DiscardCost, produced string) uuid.UUID {
	c := NewCard("Discard Familiar", owner.ID)
	c.TypeLine = "Creature — Imp"
	c.Controller = owner.ID
	c.Power, c.Toughness = 3, 2
	c.ManaAbilities = []ManaAbilityShape{{
		DiscardCards: cost,
		Produced:     produced,
		Label:        "Discard a card: Add " + produced,
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func anyCardDiscardCost() *DiscardCost {
	return &DiscardCost{N: 1, Label: "a card"}
}

// The headline: the named card leaves the hand, the mana lands in the
// pool, and the two happen in one step.
func TestManaAbilityDiscardCostPaysAndProduces(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
	pitch := seedHandCard(me, "Pitch Me", "", "Instant", "{3}{R}").InstanceID
	keep := seedHandCard(me, "Keep Me", "", "Instant", "{1}{U}").InstanceID

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the named card is still in hand")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card did not reach the graveyard")
	}
	if !me.Hand.Contains(keep) {
		t.Error("a card nobody named was discarded")
	}
	if got := len(me.ManaPool); got != 1 {
		t.Errorf("mana pool has %d, want 1 — the ability produced its {B}", got)
	}
}

// One EventDiscardCard, exactly — the whole reason the component pays
// through discardCardsLocked rather than moving the card itself.
func TestManaAbilityDiscardCostFiresTheDiscardEventOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
	pitch := seedHandCard(me, "Pitch Me", "", "Instant", "{3}{R}").InstanceID

	before := len(g.Events)
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	n := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventDiscardCard && ev.CardID == pitch {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d EventDiscardCard for the pitched card, want exactly one", n)
	}
}

// CR 702.35a: a madness card discarded to pay a MANA ability's cost is
// exiled rather than binned, and the offer goes up. The replacement
// reads no cause at all, which is exactly why it needed no change for
// this surface — the component just had to use the one door.
func TestManaAbilityDiscardCostIsSeenByMadness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: []uuid.UUID{id}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if inGraveyard(me, id) {
		t.Fatal("the madness card went to the graveyard; CR 702.35a exiles it instead")
	}
	if exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the madness card is not in exile — the cost discard did not reach the CR 614 window")
	}
	settleStack(t, g)
	if madnessOffer(g) == nil {
		t.Error("no madness cast was offered after a mana-ability cost discard")
	}
}

// Every way the payment is refused, and that none of them taps, moves
// or mints anything.
func TestManaAbilityDiscardCostRejectsWhatCannotPay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
	mine := seedHandCard(me, "Mine", "", "Instant", "{R}").InstanceID

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"two named for a one-card clause", []uuid.UUID{mine, mine}, ErrInvalidParam},
		{"a card that is not in hand", []uuid.UUID{uuid.New()}, ErrCardNotFound},
	}
	for _, tc := range cases {
		err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: tc.ids})
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if !me.Hand.Contains(mine) {
		t.Error("a refused activation discarded a card")
	}
	if got := len(me.ManaPool); got != 0 {
		t.Errorf("a refused activation minted %d mana", got)
	}
}

// IDs for a mana ability with no discard component are refused rather
// than ignored, exactly as an unexpected sacrifice_ids is.
func TestManaAbilityDiscardCostRefusesIDsForAnAbilityWithoutOne(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiscardManaSource(g, me, nil, "{B}")
	card := seedHandCard(me, "Mine", "", "Instant", "{R}").InstanceID

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: []uuid.UUID{card}}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(card) {
		t.Error("a refused activation discarded the card anyway")
	}
}

// The AUTO-TAPPER never plans a discard-cost source: which card to
// pitch is a decision, and the planner makes none. Same bar the life
// cost and the tap-others cost already fail.
func TestAutoTapNeverPlansADiscardCostManaAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
	// Give it a {T} too, so the only reason to decline is the
	// discard: autoTapAbilityFor skips a tapless ability anyway.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == src {
				g.Battlefield.Cards[i].ManaAbilities[0].TapCost = true
			}
		}
	})
	seedHandCard(me, "Pitch Me", "", "Instant", "{3}{R}")

	g.mu.RLock()
	card := *findBattlefieldCard(g, src)
	picked := g.autoTapAbilityFor(me.ID, card, ManaAbilitiesForCard(card))
	g.mu.RUnlock()
	if picked != nil {
		t.Error("the auto-tapper planned a source whose cost is a discard")
	}
}
