package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// discard_hand_cost_test.go — #1600: "Discard your hand" as a cost
// (DiscardCost.Hand), on both owners of the discard component, and
// "Activate only as an instant" on a mana ability
// (Game.InstantWindowOpenForEffect).
//
// The mana-ability source is Lion's Eye Diamond's shape built by hand:
// sacrifice itself, discard the hand, add three mana, only as an
// instant. The real card is proved end to end in
// cards/effects/discard_hand_cards_test.go.

func handClause() *DiscardCost { return &DiscardCost{Hand: true, Label: "your hand"} }

// instantOnly is effects.OnlyAsAnInstant written out, so this package
// tests the window without importing the catalog.
func instantOnly(g *Game, controller, _ uuid.UUID) bool {
	return g.InstantWindowOpenForEffect(controller)
}

// pushDiamond seats a Lion's Eye Diamond-shaped artifact for `owner`.
func pushDiamond(g *Game, owner *Player, produced string) uuid.UUID {
	c := NewCard("Test Diamond", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ManaAbilities = []ManaAbilityShape{{
		SacrificeCost: true,
		DiscardCards:  handClause(),
		Produced:      produced,
		Label:         "Discard your hand, Sacrifice this artifact: Add " + produced + ". Activate only as an instant.",
		Condition:     instantOnly,
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func discardEventsFor(g *Game, from int) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventDiscardCard {
			out[ev.CardID]++
		}
	}
	return out
}

// The headline: every card in hand goes, once each, through the one
// discard door; the source is sacrificed; the mana lands. No ids are
// sent — there is nothing to choose.
func TestDiscardYourHandManaCostDiscardsEveryCard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiamond(g, me, "{R}{R}{R}")
	hand := handIDs(me, me.Hand.Size())
	if len(hand) == 0 {
		t.Fatal("the opening hand is empty; the test needs a full one")
	}

	before := len(g.Events)
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := me.Hand.Size(); n != 0 {
		t.Errorf("%d cards left in hand, want 0", n)
	}
	seen := discardEventsFor(g, before)
	for _, id := range hand {
		if !me.Graveyard.Contains(id) {
			t.Errorf("hand card %s did not reach the graveyard", id)
		}
		if seen[id] != 1 {
			t.Errorf("hand card %s: %d EventDiscardCard, want exactly one", id, seen[id])
		}
	}
	if findBattlefieldCard(g, src) != nil {
		t.Error("the source is still on the battlefield; the sacrifice was not paid")
	}
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("pool has %d mana, want 3", got)
	}
}

// CR 118.3: discarding every card of an empty hand is paying the cost
// in full. An empty hand is the commonest way the card is played.
func TestDiscardYourHandManaCostIsPaidByAnEmptyHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	src := pushDiamond(g, me, "{G}{G}{G}")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("an empty hand must pay \"Discard your hand\": %v", err)
	}
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("pool has %d mana, want 3", got)
	}
	if findBattlefieldCard(g, src) != nil {
		t.Error("the source was not sacrificed")
	}
}

// Ids for a hand clause are refused, not ignored — a client that names
// cards is confused about which cost it is paying — and nothing is
// paid: the hand, the source and the pool are as they were.
func TestDiscardYourHandRefusesNamedCards(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiamond(g, me, "{R}{R}{R}")
	hand := handIDs(me, me.Hand.Size())

	err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{DiscardIDs: hand[:1]})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if got := me.Hand.Size(); got != len(hand) {
		t.Errorf("hand has %d cards, want %d — a refused activation discarded", got, len(hand))
	}
	if findBattlefieldCard(g, src) == nil {
		t.Error("a refused activation sacrificed the source")
	}
	if len(me.ManaPool) != 0 {
		t.Error("a refused activation minted mana")
	}
}

// Validate every component before paying any: a Diamond Lion-shaped
// source that is already tapped refuses with the whole hand still in
// it, and a source that has already been sacrificed (the second click
// on a Diamond that is gone) costs nothing either.
func TestDiscardYourHandIsNotPaidWhenAnotherComponentFails(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	lion := pushDiamond(g, me, "{W}{W}{W}")
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, lion)
		c.ManaAbilities[0].TapCost = true
		c.Tapped = true
	})
	size := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, lion, 0, ManaAbilityParams{}); !errors.Is(err, ErrAlreadyTapped) {
		t.Fatalf("tapped source: err = %v, want ErrAlreadyTapped", err)
	}
	if me.Hand.Size() != size {
		t.Errorf("hand has %d cards after a refused activation, want %d", me.Hand.Size(), size)
	}

	diamond := pushDiamond(g, me, "{B}{B}{B}")
	if err := g.ActivateManaAbility(me.ID, diamond, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	seed := seedHandCard(me, "Drawn Later", "", "Instant", "{U}").InstanceID
	// The Diamond is in the graveyard now, where its ability does not
	// function (CR 113.6).
	if err := g.ActivateManaAbility(me.ID, diamond, 0, ManaAbilityParams{}); !errors.Is(err, ErrActivationZoneNotAllowed) {
		t.Fatalf("source already gone: err = %v, want ErrActivationZoneNotAllowed", err)
	}
	if !me.Hand.Contains(seed) {
		t.Error("an activation of a source that is gone discarded a card")
	}
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("pool has %d mana, want the first activation's 3 and nothing more", got)
	}
}

// CR 605.3b / CR 603.3: the mana ability resolves at once, and what its
// cost triggered waits for the next time a player would receive
// priority — so "whenever a player discards a card" goes on the stack
// AFTER the mana is in the pool, once per card discarded.
func TestDiscardYourHandTriggersGoOnTheStackAfterTheMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const watcher = "test-discard-watcher"
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(),
		Name:       "Discard Watcher",
		OracleID:   watcher,
		TypeLine:   "Enchantment",
		Owner:      g.Seats[1].ID,
		Controller: g.Seats[1].ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != watcher {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventDiscardCard},
			AppliesTo: func(ev Event, _ *Card, _ Characteristic, _ *Game) bool {
				return ev.Actor == me.ID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Discard Watcher — a card was discarded",
					func(*Game, *StackItem) error { return nil })
			},
		}}
	})
	src := pushDiamond(g, me, "{R}{R}{R}")
	n := me.Hand.Size()

	// At the moment the last mana lands, every discard trigger is
	// still waiting (PendingTriggers) and none is on the stack.
	probe := &manaThenTriggersProbe{}
	g.RegisterListener(probe)
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if probe.manaEvents != 3 {
		t.Fatalf("%d mana_added events, want 3", probe.manaEvents)
	}
	if probe.onStackAtLastMana != 0 {
		t.Errorf("%d triggered items were already on the stack when the mana landed, want 0", probe.onStackAtLastMana)
	}
	if probe.pendingAtLastMana != n {
		t.Errorf("%d discard triggers were waiting when the mana landed, want one per card (%d)", probe.pendingAtLastMana, n)
	}
	// …and on the way out of the activation they are put on the stack
	// (CR 117.5), with the mana still in the pool.
	if got := triggeredOnStack(g); got != n {
		t.Errorf("%d discard triggers on the stack after the activation, want %d", got, n)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("%d triggers still pending", len(g.PendingTriggers))
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool has %d mana, want 3", len(me.ManaPool))
	}
}

type manaThenTriggersProbe struct {
	manaEvents        int
	onStackAtLastMana int
	pendingAtLastMana int
}

func (p *manaThenTriggersProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventManaAdded {
		return
	}
	p.manaEvents++
	p.onStackAtLastMana = triggeredOnStack(g)
	p.pendingAtLastMana = len(g.PendingTriggers)
}

// CR 702.35a: a madness card discarded to a "Discard your hand" cost is
// exiled and offered, exactly as one discarded to a one-card cost is —
// the clause uses the one discard door and the replacement reads no
// cause.
func TestDiscardYourHandIsSeenByMadness(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withMadnessCard(t, "{R}")
	src := pushDiamond(g, me, "{R}{R}{R}")
	id := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if inGraveyard(me, id) {
		t.Fatal("the madness card went to the graveyard; CR 702.35a exiles it instead")
	}
	if exiledCardByIDLocked(g, id) == nil {
		t.Fatal("the madness card is not in exile")
	}
	settleStack(t, g)
	if madnessOffer(g) == nil {
		t.Error("no madness cast was offered after a discard-your-hand cost")
	}
}

// "Activate only as an instant": refused while the activator does not
// hold priority, while they owe a prompt (the pay-unless tax the
// classic illegal play would pay), and while anyone's blocking prompt
// is open — each with nothing paid. Allowed again once they hold
// priority with nothing open, even with a spell of theirs on the stack.
func TestActivateOnlyAsAnInstantWindow(t *testing.T) {
	g := newActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]
	src := pushDiamond(g, me, "{U}{U}{U}")
	size := me.Hand.Size()

	refused := func(why string) {
		t.Helper()
		err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{})
		if !errors.Is(err, ErrConditionNotMet) {
			t.Fatalf("%s: err = %v, want ErrConditionNotMet", why, err)
		}
		if me.Hand.Size() != size || findBattlefieldCard(g, src) == nil || len(me.ManaPool) != 0 {
			t.Fatalf("%s: a refused activation paid something", why)
		}
	}

	// Priority with the other seat.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if g.InstantWindowOpenForEffect(me.ID) {
		t.Fatal("the window is open for a seat that does not hold priority")
	}
	refused("not holding priority")
	if !g.InstantWindowOpenForEffect(other.ID) {
		t.Error("the window is shut for the seat that holds priority")
	}
	// Back to me: the other seat passes and the step moves on; walk
	// until I hold priority again.
	for i := 0; i < 8 && !g.InstantWindowOpenForEffect(me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if !g.InstantWindowOpenForEffect(me.ID) {
		t.Fatal("never got priority back")
	}

	// A prompt I owe: Mana Leak's tax, mid-resolution.
	owed := &PendingChoice{ID: uuid.New(), Kind: PendingChoicePayUnless, Chooser: me.ID, Reason: "Mana Leak"}
	g.WithWriteLock(func() { g.PendingChoices = append(g.PendingChoices, owed) })
	refused("owing a pay-unless prompt")
	g.WithWriteLock(func() { g.PendingChoices = nil })

	// Somebody else's blocking prompt: a resolution still under way.
	blocking := &PendingChoice{ID: uuid.New(), Kind: PendingChoiceSacrifice, Chooser: other.ID, Reason: "Fleshbag Marauder"}
	g.WithWriteLock(func() { g.PendingChoices = append(g.PendingChoices, blocking) })
	if !g.ChoicePromptBlocksTable(blocking) {
		t.Fatal("test premise: a sacrifice prompt blocks the table")
	}
	refused("another player's blocking prompt open")
	g.WithWriteLock(func() { g.PendingChoices = nil })

	// Nothing open, my priority: open.
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("with priority and nothing open: %v", err)
	}
}

// Split second (CR 702.61b) stops abilities that aren't mana abilities;
// this one is, so the window it reads is not shut by it.
func TestActivateOnlyAsAnInstantIgnoresSplitSecond(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushDiamond(g, me, "{R}{R}{R}")
	g.WithWriteLock(func() { g.SplitSecondActive = true })
	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("split second refused a mana ability: %v", err)
	}
}

// The auto-tapper never plans a "Discard your hand" source — not even
// with an empty hand, where nothing is thrown away and no card has to
// be chosen. The planner runs in the middle of a cast, which "Activate
// only as an instant" forbids, and it cannot read that out of a
// Condition closure: refusing every discard component is what keeps
// Lion's Eye Diamond out of a plan.
func TestAutoTapNeverPlansADiscardYourHandSource(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	src := pushDiamond(g, me, "{R}{R}{R}")

	g.mu.RLock()
	card := *findBattlefieldCard(g, src)
	picked := g.autoTapAbilityFor(me.ID, card, ManaAbilitiesForCard(card))
	g.mu.RUnlock()
	if picked != nil {
		t.Error("the auto-tapper picked a discard-your-hand mana ability")
	}
	cost, err := ParseCost("{R}")
	if err != nil {
		t.Fatal(err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("a plan for {R} spent %v; the Diamond is the only source and must not be one", plan)
	}
	if findBattlefieldCard(g, src) == nil {
		t.Error("planning cracked the Diamond")
	}
}

// The CR 602 owner (Null Brooch, Slate of Ancestry): the hand is
// discarded at announce and recorded on the payment, and an empty hand
// pays it. The ability itself is not instant-restricted; the CR 602
// path's own timing applies.
func TestDiscardYourHandOnAnActivatedAbility(t *testing.T) {
	for _, empty := range []bool{false, true} {
		name := "full hand"
		if empty {
			name = "empty hand"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			if empty {
				me.Hand.Cards = nil
			}
			hand := handIDs(me, me.Hand.Size())
			c := NewCard("Test Slate", me.ID)
			c.TypeLine = "Artifact"
			c.Controller = me.ID
			c.ActivatedAbilities = []ActivatedAbilityShape{{
				Label: "Discard your hand: Draw nothing.",
				Cost:  AbilityCost{DiscardCards: handClause()},
			}}
			g.Battlefield.PushTop(c)

			before := len(g.Events)
			if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if me.Hand.Size() != 0 {
				t.Errorf("%d cards left in hand", me.Hand.Size())
			}
			seen := discardEventsFor(g, before)
			for _, id := range hand {
				if seen[id] != 1 || !me.Graveyard.Contains(id) {
					t.Errorf("hand card %s: %d discard events, in graveyard %v", id, seen[id], me.Graveyard.Contains(id))
				}
			}
			item := onlyStackItem(t, g)
			if len(item.Paid.Discarded) != len(hand) {
				t.Errorf("Paid.Discarded has %d cards, want %d", len(item.Paid.Discarded), len(hand))
			}
		})
	}
}

// The CR 602 path refuses ids for a hand clause too, with nothing paid.
func TestDiscardYourHandOnAnActivatedAbilityRefusesNamedCards(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hand := handIDs(me, me.Hand.Size())
	c := NewCard("Test Slate", me.ID)
	c.TypeLine = "Artifact"
	c.Controller = me.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "{T}, Discard your hand: Draw nothing.",
		Cost:  AbilityCost{Tap: true, DiscardCards: handClause()},
	}}
	g.Battlefield.PushTop(c)

	err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{DiscardIDs: hand[:1]})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if me.Hand.Size() != len(hand) || findBattlefieldCard(g, c.InstanceID).Tapped {
		t.Error("a refused activation paid part of its cost")
	}
}

// No options are offered for a hand clause: the view stamps none and
// the client opens no picker.
func TestDiscardYourHandOffersNoOptions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var opts []uuid.UUID
	g.ReadSnapshot(func() {
		opts = g.DiscardCostOptionsForEffect(me.ID, uuid.Nil, handClause())
	})
	if len(opts) != 0 {
		t.Errorf("a hand clause offered %d options, want none", len(opts))
	}
	if handClause().Matches(me.Hand.Cards[0]) {
		t.Error("a hand clause matched a named card")
	}
}
