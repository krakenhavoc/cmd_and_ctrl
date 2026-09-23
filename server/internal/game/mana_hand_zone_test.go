package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mana_hand_zone_test.go — #1228: ManaAbilityShape.Zones and
// ManaAbilityShape.ExileSelf, the CR 113.6 dimension and the
// exile-this cost on the OTHER activation entry point.
//
// "Exile this card from your hand: Add {R}" is CR 605.1a's definition
// of a mana ability (it could add mana, it is not a loyalty ability,
// it targets nothing) printed on a card that is not a permanent. Three
// rules meet on it and each one has a test below: CR 113.6 (the
// ability functions where it says it functions, and nowhere else),
// CR 108.4 (off the battlefield the OWNER is "you") and CR 601.2h /
// 602.2b (paying a cost is one indivisible step, so the exit cannot
// pause on a prompt).

// spiritGuideAbility is the printed Spirit Guide ability for a
// `produced` mana string.
func spiritGuideAbility(produced string) []ManaAbilityShape {
	return []ManaAbilityShape{{
		Zones:     []ZoneKind{ZoneHand},
		ExileSelf: true,
		Produced:  produced,
		Label:     "Exile this card from your hand: Add " + produced,
	}}
}

// seedSpiritGuide puts a Spirit Guide in p's hand and returns its ID.
func seedSpiritGuide(p *Player, name, produced string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Creature — Ape Spirit"
	c.ManaCost = "{2}{R}"
	c.Power, c.Toughness = 2, 2
	c.ManaAbilities = spiritGuideAbility(produced)
	p.Hand.PushTop(c)
	return c.InstanceID
}

// The headline: the card leaves the hand for exile and the mana lands
// in the pool, in one step.
func TestManaAbilityFromHandExilesTheCardAndProducesMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	if err := g.ActivateManaAbility(me.ID, guide, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if me.Hand.Contains(guide) {
		t.Error("the Spirit Guide is still in hand — its own cost exiles it")
	}
	exiledCard(t, g, guide) // fails the test if it is not there
	if got := len(me.ManaPool); got != 1 {
		t.Fatalf("mana pool has %d, want 1", got)
	}
	if c := me.ManaPool[0].Color; c != "R" {
		t.Errorf("produced %q, want R", c)
	}
}

// CR 113.6 in the direction that is easy to get wrong: a declared zone
// is not an ADDITION to the battlefield. A Spirit Guide that somehow
// reached the battlefield is a 2/2 Ape with no abilities.
func TestManaAbilityFromHandIsRefusedFromTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	onBoard := pushIntrinsicPermanent(g, me, "Simian Spirit Guide",
		"Creature — Ape Spirit", spiritGuideAbility("{R}"), nil)

	err := g.ActivateManaAbility(me.ID, onBoard, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrActivationZoneNotAllowed) {
		t.Fatalf("activate from the battlefield: %v, want ErrActivationZoneNotAllowed", err)
	}
	if len(me.ManaPool) != 0 {
		t.Error("the refused activation still minted mana")
	}
}

// And the opposite direction, which is the half that was free before
// this issue and has to stay free: an ordinary "{T}: Add {G}" declares
// nothing, so it functions from the battlefield alone and a Forest in
// hand is not a mana source.
func TestBattlefieldManaAbilityIsRefusedFromTheHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := seedHandCard(me, "Forest", "", "Basic Land — Forest", "").InstanceID

	err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrActivationZoneNotAllowed) && !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("activate a Forest out of hand: %v, want a zone refusal", err)
	}
	if len(me.ManaPool) != 0 {
		t.Error("a Forest in hand produced mana")
	}
	if !me.Hand.Contains(forest) {
		t.Error("the refused activation moved the card")
	}
}

// CR 108.4: a card outside the battlefield and the stack has no
// controller, so its OWNER is the "you" of its printed text — and
// nobody else's.
func TestManaAbilityFromHandIsOnlyItsOwners(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	err := g.ActivateManaAbility(them.ID, guide, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("an opponent activating my hand card: %v, want ErrCardCallerMismatch", err)
	}
	if !me.Hand.Contains(guide) {
		t.Error("the refused activation exiled the card anyway")
	}
	if len(them.ManaPool) != 0 {
		t.Error("the opponent got mana out of my hand")
	}
}

// #1212 / #1216: the mana carries a snapshot of what its source WAS,
// taken before the cost moved it. A Spirit Guide is a creature card,
// so the token records a creature source — the whole point of the
// snapshot being taken early is that by now the card is in exile.
func TestManaAbilityFromHandStampsTheSourceKinds(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	if err := g.ActivateManaAbility(me.ID, guide, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("mana pool has %d, want 1", len(me.ManaPool))
	}
	kinds := me.ManaPool[0].SourceKinds
	if !kinds.Has(ManaSourceCreature) {
		t.Errorf("SourceKinds = %b, want the creature bit — a Spirit Guide is a creature card", kinds)
	}
	if kinds.Has(ManaSourceLand) {
		t.Errorf("SourceKinds = %b claims a land source", kinds)
	}
	if me.ManaPool[0].Source != guide {
		t.Error("the token does not name the card that made it")
	}
}

// CR 601.2h / 602.2b: a cost is one indivisible step, so the exit
// settles itself. A COMMANDER spent as a Spirit Guide's cost goes to
// exile and asks nobody — the same posture the cost discard takes
// (#660) and the reason payExileSelfCostLocked sets MustSettleNow.
func TestManaAbilityFromHandNeverPausesOnTheCommanderQuestion(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seatCommander(t, me.Hand, me)
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].ManaAbilities = spiritGuideAbility("{G}")
		}
	}

	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Fatalf("%d pending choices after a cost exile — CR 601.2h does not pause", n)
	}
	// Fails the test if the commander is anywhere else: a cost cannot
	// take CR 903.9's "may".
	exiledCard(t, g, id)
	if len(me.ManaPool) != 1 {
		t.Error("the mana did not arrive")
	}
}

// The activation emits the same event a battlefield one does, so
// "whenever you activate an ability" watchers see one kind of event
// whatever zone it came out of (#1184, CR 605.1a).
func TestManaAbilityFromHandEmitsTheActivationEvent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Elvish Spirit Guide", "{G}")

	before := len(g.Events)
	if err := g.ActivateManaAbility(me.ID, guide, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	n := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventManaAbilityActivated && ev.CardID == guide {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d EventManaAbilityActivated, want exactly one", n)
	}
}

// The zone predicate itself, in both directions and for the default.
func TestManaAbilityZonesDefaultsToTheBattlefield(t *testing.T) {
	plain := ManaAbilityShape{TapCost: true, Produced: "{C}"}
	if !ManaAbilityFunctionsFromZone(plain, ZoneBattlefield) {
		t.Error("an undeclared mana ability does not function from the battlefield")
	}
	if ManaAbilityFunctionsFromZone(plain, ZoneHand) {
		t.Error("an undeclared mana ability functions from a hand")
	}
	guide := spiritGuideAbility("{R}")[0]
	if !ManaAbilityFunctionsFromZone(guide, ZoneHand) {
		t.Error("a hand-declared mana ability does not function from a hand")
	}
	if ManaAbilityFunctionsFromZone(guide, ZoneBattlefield) {
		t.Error("a hand-declared mana ability also functions from the battlefield")
	}
}

// The boot-time rules the catalog is held to, asked of the predicate
// they are written against. effects.Register turns each of these into
// a panic; here they are just strings.
func TestManaAbilityNeedsPermanentSourceNamesTheComponent(t *testing.T) {
	cases := []struct {
		name string
		ab   ManaAbilityShape
		want bool
	}{
		{"tap", ManaAbilityShape{TapCost: true}, true},
		{"sacrifice this", ManaAbilityShape{SacrificeCost: true}, true},
		{"add a counter", ManaAbilityShape{AddCounter: &CounterAddCost{Counter: "charge", N: 1}}, true},
		{"remove from this", ManaAbilityShape{RemoveCounters: &CounterRemovalCost{Counter: "charge", N: 1}}, true},
		{"exile this", ManaAbilityShape{ExileSelf: true}, false},
	}
	for _, tc := range cases {
		got := ManaAbilityNeedsPermanentSource(tc.ab) != ""
		if got != tc.want {
			t.Errorf("%s: needs a permanent = %v, want %v", tc.name, got, tc.want)
		}
	}
	if ManaAbilityZoneUnsupported(ZoneHand) != "" {
		t.Error("the hand is the one supported non-battlefield mana zone and was refused")
	}
	if ManaAbilityZoneUnsupported(ZoneGraveyard) == "" {
		t.Error("the graveyard is not walked for mana abilities and was accepted")
	}
}
