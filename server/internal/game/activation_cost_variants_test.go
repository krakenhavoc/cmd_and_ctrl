package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// activation_cost_variants_test.go — #6208: three variants of an
// activation cost (CR 602.2b, CR 118.3), each a reading of a component
// the engine already had.
//
//	RevealCardsCost.SharesColor      "Reveal two cards from your hand
//	                                  that share a color" (Illuminated
//	                                  Folio)
//	AbilityCost.SacrificeAll         "Sacrifice all lands you control"
//	                                  (Tomb of Urami)
//	TapOthersCost.OpponentControls   "Untap a tapped land an opponent
//	                                  controls" (Benthic Explorers)
//
// The cards themselves are built by their own OLD batches; what is
// pinned here is the engine: the one candidate walk, the one payment,
// the validator and the payer of each clause.

// onlyStackItemForTest is the one item on the stack, or nil when the
// stack does not hold exactly one. Caller must hold g.mu.
func onlyStackItemForTest(t *testing.T, g *Game) *StackItem {
	t.Helper()
	if len(g.StackMeta) != 1 {
		return nil
	}
	for _, item := range g.StackMeta {
		return item
	}
	return nil
}

// --- the shared-colour reveal -----------------------------------------

// revealSharesColorCost is Illuminated Folio's clause: two cards from
// your hand that share a colour.
func revealSharesColorCost() *RevealCardsCost {
	return &RevealCardsCost{
		N:           2,
		Label:       "two cards from your hand that share a color",
		SharesColor: true,
	}
}

// pushHandCardWithColors seats a card of the given colours in `p`'s
// hand. The colours are stamped rather than derived from a mana cost,
// which is what EffectiveColors reads off a card in a hand (CR 202.2).
func pushHandCardWithColors(p *Player, name string, colors ...string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Creature — Bear"
	c.Colors = colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// CR 118.3 / CR 202.2: the walk offers only the cards that could
// actually pay — a card with no partner of its colour cannot, and a
// colourless card shares no colour at all (CR 105.2c), so it never can.
func TestRevealSharesColorOffersOnlyCardsWithAPartner(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Illuminated Folio", "Artifact", false)
	blueA := pushHandCardWithColors(me, "Blue A", "U")
	blueB := pushHandCardWithColors(me, "Blue B", "U")
	lonelyRed := pushHandCardWithColors(me, "Lonely Red", "R")
	colorless := pushHandCardWithColors(me, "Colorless")
	// A gold card shares black with nothing here, but shares blue with
	// the two blue cards, so it IS a candidate.
	goldUB := pushHandCardWithColors(me, "Gold", "U", "B")

	g.mu.RLock()
	opts := g.RevealCardsOptionsForEffect(me.ID, src, revealSharesColorCost())
	plain := g.RevealCardsOptionsForEffect(me.ID, src, &RevealCardsCost{N: 2, Label: "two cards"})
	g.mu.RUnlock()

	want := map[uuid.UUID]bool{blueA: true, blueB: true, goldUB: true}
	if len(opts) != len(want) {
		t.Fatalf("options = %v, want the three cards that share blue", opts)
	}
	for _, id := range opts {
		if !want[id] {
			t.Errorf("a card with no sharing partner was offered: %v", id)
		}
	}
	// The same hand with no sharing clause offers every card in it: the
	// narrowing is the clause's, not the walk's. (The opening hand the
	// table was dealt is colourless, so it is in `plain` and not in
	// `opts`.)
	if len(plain) != len(me.Hand.Cards) {
		t.Errorf("the plain clause offered %d of %d cards in hand", len(plain), len(me.Hand.Cards))
	}
	_, _ = lonelyRed, colorless
}

// The headline refusal the issue asks for: two cards with no shared
// colour do not pay, and the refusal names the rule.
func TestRevealSharesColorRefusesTwoCardsThatShareNone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Illuminated Folio", "Artifact", false)
	blue := pushHandCardWithColors(me, "Blue", "U")
	blueTwo := pushHandCardWithColors(me, "Blue Two", "U")
	red := pushHandCardWithColors(me, "Red", "R")
	pushHandCardWithColors(me, "Red Two", "R")
	colorless := pushHandCardWithColors(me, "Colorless")

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		// Both cards are candidates — each has a partner of its own
		// colour — but not each other's, which is what the set rule
		// catches and what names the rule in the refusal.
		{"two colours that share nothing", []uuid.UUID{blue, red}, ErrRevealedColorMismatch},
		// A colourless card shares no colour with anything (CR 105.2c),
		// so the walk never offers it at all and it is refused as any
		// non-candidate is.
		{"a colourless card shares no colour", []uuid.UUID{blue, colorless}, ErrInvalidParam},
		{"one card named twice", []uuid.UUID{blue, blue}, ErrInvalidParam},
		{"only one card named", []uuid.UUID{blue}, ErrInvalidParam},
		{"nothing named", nil, ErrInvalidParam},
	}
	for _, tc := range cases {
		g.mu.Lock()
		err := g.validateRevealCardsCostLocked(me.ID, src, revealSharesColorCost(), tc.ids, 0)
		g.mu.Unlock()
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}

	g.mu.Lock()
	ok := g.validateRevealCardsCostLocked(me.ID, src, revealSharesColorCost(), []uuid.UUID{blue, blueTwo}, 0)
	g.mu.Unlock()
	if ok != nil {
		t.Fatalf("two blue cards were refused: %v", ok)
	}
}

// A gold card pays with either half of its colours: the clause asks for
// one colour in common, not for the same colours (CR 202.2).
func TestRevealSharesColorAcceptsAGoldCardWithEitherHalf(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Illuminated Folio", "Artifact", false)
	gold := pushHandCardWithColors(me, "Gold", "U", "B")
	blue := pushHandCardWithColors(me, "Blue", "U")
	black := pushHandCardWithColors(me, "Black", "B")
	// Each of the two mono-coloured cards has a partner of its own
	// colour, so both are candidates and the pair reaches the set rule.
	pushHandCardWithColors(me, "Blue Two", "U")
	pushHandCardWithColors(me, "Black Two", "B")

	for _, pair := range [][]uuid.UUID{{gold, blue}, {gold, black}} {
		g.mu.Lock()
		err := g.validateRevealCardsCostLocked(me.ID, src, revealSharesColorCost(), pair, 0)
		g.mu.Unlock()
		if err != nil {
			t.Errorf("%v was refused: %v", pair, err)
		}
	}
	// But blue and black still share nothing with each other.
	g.mu.Lock()
	err := g.validateRevealCardsCostLocked(me.ID, src, revealSharesColorCost(), []uuid.UUID{blue, black}, 0)
	g.mu.Unlock()
	if !errors.Is(err, ErrRevealedColorMismatch) {
		t.Errorf("blue and black: err = %v, want ErrRevealedColorMismatch", err)
	}
}

// The enumerator's one payment has to be a LEGAL one: the first two of
// the walk can each have a partner without being each other's.
func TestRevealSharesColorPaymentIsAPairThatShares(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Illuminated Folio", "Artifact", false)
	// Hand order: red, blue, red, blue. Each card has a partner, but the
	// first two do not share a colour.
	pushHandCardWithColors(me, "Red A", "R")
	pushHandCardWithColors(me, "Blue A", "U")
	pushHandCardWithColors(me, "Red B", "R")
	pushHandCardWithColors(me, "Blue B", "U")

	g.mu.RLock()
	pay := g.RevealCardsPaymentForEffect(me.ID, src, revealSharesColorCost())
	g.mu.RUnlock()
	if len(pay) != 2 {
		t.Fatalf("payment = %v, want two cards", pay)
	}
	g.mu.Lock()
	err := g.validateRevealCardsCostLocked(me.ID, src, revealSharesColorCost(), pay, 0)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("the engine's own payment was refused: %v", err)
	}

	// A hand with no sharing pair pays nothing, so the ability is not
	// offered at all (CR 118.3).
	g2 := newActiveGame(t)
	me2 := g2.Seats[0]
	src2 := pushTapOthersPermanent(g2, me2, "Illuminated Folio", "Artifact", false)
	pushHandCardWithColors(me2, "Red", "R")
	pushHandCardWithColors(me2, "Blue", "U")
	g2.mu.RLock()
	none := g2.RevealCardsPaymentForEffect(me2.ID, src2, revealSharesColorCost())
	g2.mu.RUnlock()
	if none != nil {
		t.Errorf("a hand with no sharing pair paid %v", none)
	}
}

// End to end on an activated ability: the pair is revealed (CR 701.20a)
// and nothing leaves the hand (CR 701.20b).
func TestRevealSharesColorOnAnActivatedAbility(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{RevealCards: revealSharesColorCost()})
	blue := pushHandCardWithColors(me, "Blue", "U")
	blueTwo := pushHandCardWithColors(me, "Blue Two", "U")
	red := pushHandCardWithColors(me, "Red", "R")
	pushHandCardWithColors(me, "Red Two", "R")
	hand := len(me.Hand.Cards)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{RevealIDs: []uuid.UUID{blue, red}}); !errors.Is(err, ErrRevealedColorMismatch) {
		t.Fatalf("a non-sharing pair: err = %v, want ErrRevealedColorMismatch", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{RevealIDs: []uuid.UUID{blue, blueTwo}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := len(me.Hand.Cards); n != hand {
		t.Errorf("hand has %d cards, want %d — revealing moves nothing (CR 701.20b)", n, hand)
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// --- the ability's sacrifice-all clause --------------------------------

// pushTypedAbilitySource is pushTapOthersAbilitySource with a type line
// of its own: Tomb of Urami's own ability is printed on a LAND, and the
// clause takes it with the rest.
func pushTypedAbilitySource(g *Game, owner *Player, name, typeLine string, cost AbilityCost) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = typeLine
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: name + ": mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// sacrificeAllLandsSpec is Tomb of Urami's clause: all lands you
// control, with the open "any number" bounds the flag needs.
func sacrificeAllLandsSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "all lands you control",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.HasCardType("land")
		},
		Min: 0, Max: 0,
	}
}

// The headline: every land goes, including the source's own land, and
// the activator names nothing.
func TestSacrificeAllOnAnAbilityTakesEveryMatchingPermanent(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTypedAbilitySource(g, me, "Tomb of Urami", "Legendary Land", AbilityCost{SacrificeOther: sacrificeAllLandsSpec(), SacrificeAll: true})
	mine := pushTapOthersPermanent(g, me, "Swamp", "Land — Swamp", false)
	alsoMine := pushTapOthersPermanent(g, me, "Island", "Land — Island", true)
	bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)
	theirs := pushTapOthersPermanent(g, opp, "Their Swamp", "Land — Swamp", false)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no sacrifice_ids: %v", err)
	}
	g.mu.RLock()
	for _, id := range []uuid.UUID{src, mine, alsoMine} {
		if findBattlefieldCard(g, id) != nil {
			t.Errorf("a land the clause matches is still on the battlefield: %v", id)
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if findBattlefieldCard(g, id) == nil {
			t.Errorf("the clause took something it does not match: %v", id)
		}
	}
	item := onlyStackItemForTest(t, g)
	paid := item != nil && item.Paid.Sacrificed == 3
	g.mu.RUnlock()
	if !paid {
		t.Error("PaidCost.Sacrificed does not record the three lands")
	}
}

// `sacrifice_ids` may be empty or exactly the set; anything else is a
// client that is confused about which ability it is firing, because the
// activator chooses nothing (CR 601.2b via CR 602.2b).
func TestSacrificeAllOnAnAbilityRefusesAnyOtherList(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushTapOthersPermanent(g, me, "Tomb of Urami", "Legendary Land", false)
	one := pushTapOthersPermanent(g, me, "Swamp", "Land — Swamp", false)
	two := pushTapOthersPermanent(g, me, "Island", "Land — Island", false)
	bear := pushTapOthersPermanent(g, me, "Bear", "Creature — Bear", false)

	g.mu.RLock()
	all := g.SacrificeAllCandidatesForEffect(me.ID, sacrificeAllLandsSpec())
	g.mu.RUnlock()
	if len(all) != 3 {
		t.Fatalf("candidates = %v, want the source and the two lands", all)
	}

	cases := []struct {
		name string
		ids  []uuid.UUID
		ok   bool
	}{
		{"empty: the engine fills it", nil, true},
		{"exactly the set", []uuid.UUID{src, two, one}, true},
		{"a subset", []uuid.UUID{one}, false},
		{"a land plus something the clause does not match", []uuid.UUID{src, one, bear}, false},
	}
	for _, tc := range cases {
		g.mu.Lock()
		got, err := g.sacrificeAllPaymentLocked(me.ID, sacrificeAllLandsSpec(), tc.ids)
		g.mu.Unlock()
		if tc.ok {
			if err != nil {
				t.Errorf("%s: err = %v, want nil", tc.name, err)
			} else if len(got) != 3 {
				t.Errorf("%s: payment = %v, want all three", tc.name, got)
			}
			continue
		}
		if !errors.Is(err, ErrSacrificeAllMismatch) {
			t.Errorf("%s: err = %v, want ErrSacrificeAllMismatch", tc.name, err)
		}
	}
}

// CR 118.3: with no matching permanent the cost needs nothing, so it is
// paid in full by sacrificing nothing.
func TestSacrificeAllOnAnAbilityIsPaidByAnEmptyBoard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := pushTypedAbilitySource(g, me, "Urami Rock", "Artifact", AbilityCost{SacrificeOther: sacrificeAllLandsSpec(), SacrificeAll: true})

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with no land on the board: %v", err)
	}
	g.mu.RLock()
	item := onlyStackItemForTest(t, g)
	zero := item != nil && item.Paid.Sacrificed == 0
	g.mu.RUnlock()
	if !zero {
		t.Error("an empty payment was recorded as something")
	}
	passBothForTest(g)
	if counterOf(g, src, "effect-ran") != 1 {
		t.Error("the ability did not resolve")
	}
}

// --- the untap cost turned over to an opponent's board -----------------

// untapOpponentLandCost is Benthic Explorers' clause: one tapped land an
// opponent controls.
func untapOpponentLandCost() *TapOthersCost {
	return &TapOthersCost{
		Count: 1,
		Filter: &TargetSpec{
			Mode:  "permanent",
			Label: "a tapped land an opponent controls",
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
				return c.HasCardType("land")
			},
			Min: 1, Max: 1,
		},
		Label:            "a tapped land an opponent controls",
		Untap:            true,
		OpponentControls: true,
	}
}

// The walk offers an opponent's TAPPED land and nothing else: not your
// own land (the clause inverts "you control"), not their untapped one
// (CR 701.26b), not their creature.
func TestUntapOpponentOffersOnlyTheirTappedPermanents(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", false)
	theirTapped := pushTapOthersPermanent(g, opp, "Their Island", "Land — Island", true)
	theirUntapped := pushTapOthersPermanent(g, opp, "Their Forest", "Land — Forest", false)
	theirBear := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", true)
	myTapped := pushTapOthersPermanent(g, me, "My Island", "Land — Island", true)

	g.mu.RLock()
	opts := g.TapOthersOptionsForEffect(me.ID, src, untapOpponentLandCost())
	payable := g.TapOthersPayable(me.ID, src, untapOpponentLandCost())
	g.mu.RUnlock()
	if len(opts) != 1 || opts[0] != theirTapped {
		t.Fatalf("options = %v, want only the opponent's tapped land", opts)
	}
	if !payable {
		t.Error("the cost is not payable with their tapped land on the board")
	}
	_, _, _ = theirUntapped, theirBear, myTapped
}

// The refusals. "An opponent controls" is the engine's own clause, so
// your own land is ErrCardCallerMismatch exactly as an opponent's is for
// the ordinary form, and an untapped permanent is ErrNotTapped.
func TestUntapOpponentValidationRefusesAndUntapsNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", false)
	theirTapped := pushTapOthersPermanent(g, opp, "Their Island", "Land — Island", true)
	theirUntapped := pushTapOthersPermanent(g, opp, "Their Forest", "Land — Forest", false)
	theirBear := pushTapOthersPermanent(g, opp, "Their Bear", "Creature — Bear", true)
	myTapped := pushTapOthersPermanent(g, me, "My Island", "Land — Island", true)

	cases := []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"nothing named", nil, ErrInvalidParam},
		{"my own tapped land", []uuid.UUID{myTapped}, ErrCardCallerMismatch},
		{"their untapped land", []uuid.UUID{theirUntapped}, ErrNotTapped},
		{"their tapped creature", []uuid.UUID{theirBear}, ErrIllegalTarget},
		{"a card that does not exist", []uuid.UUID{uuid.New()}, ErrCardNotFound},
	}
	for _, tc := range cases {
		g.mu.Lock()
		err := g.validateTapOthersCostLocked(me.ID, src, untapOpponentLandCost(), tc.ids)
		g.mu.Unlock()
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}

	g.mu.Lock()
	ok := g.validateTapOthersCostLocked(me.ID, src, untapOpponentLandCost(), []uuid.UUID{theirTapped})
	g.mu.Unlock()
	if ok != nil {
		t.Fatalf("their tapped land was refused: %v", ok)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, id := range []uuid.UUID{theirTapped, theirBear, myTapped} {
		if c := findBattlefieldCard(g, id); c == nil || !c.Tapped {
			t.Errorf("a refused validation untapped %v", id)
		}
	}
}

// CR 614.17b: a permanent that can't become untapped can't pay — and a
// stun counter is not that rule, because removing the counter instead
// still pays the cost (CR 122.1d, CR 118.11).
func TestUntapOpponentCannotPayWithAPermanentThatCannotUntap(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", false)
	stunned := pushTapOthersPermanent(g, opp, "Stunned Island", "Land — Island", true)

	g.WithWriteLock(func() { _ = g.applyCounterLocked(stunned, CounterStun, 1) })

	g.mu.Lock()
	opts := g.TapOthersOptionsForEffect(me.ID, src, untapOpponentLandCost())
	err := g.validateTapOthersCostLocked(me.ID, src, untapOpponentLandCost(), []uuid.UUID{stunned})
	g.mu.Unlock()
	if len(opts) != 1 || err != nil {
		t.Fatalf("a stunned land: options = %v, err = %v — a stun counter still pays", opts, err)
	}

	g.mu.Lock()
	g.payUntapOthersCostLocked([]uuid.UUID{stunned}, uuid.Nil)
	c := findBattlefieldCard(g, stunned)
	stillTapped, counters := c != nil && c.Tapped, 0
	if c != nil {
		counters = c.Counters[CounterStun]
	}
	g.mu.Unlock()
	if !stillTapped || counters != 0 {
		t.Errorf("tapped = %v, stun counters = %d — the untap should have removed the counter instead (CR 122.1d)", stillTapped, counters)
	}
}

// Untapping to pay a cost does not target (CR 601.2h, CR 602.2b), so an
// opponent's hexproof land still pays.
func TestUntapOpponentIgnoresHexproof(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", false)
	hexproof := pushTapOthersPermanent(g, opp, "Hexproof Island", "Land — Island", true)

	g.mu.Lock()
	if c := findBattlefieldCard(g, hexproof); c != nil {
		c.Keywords = []string{"hexproof"}
	}
	err := g.validateTapOthersCostLocked(me.ID, src, untapOpponentLandCost(), []uuid.UUID{hexproof})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("an opponent's hexproof land was refused: %v", err)
	}

	g.mu.RLock()
	c := findBattlefieldCard(g, hexproof)
	targetable := c != nil && CanBeTargetedBy(c, ZoneBattlefield, SourceChooser(me.ID))
	g.mu.RUnlock()
	if targetable {
		t.Fatal("test setup: the land should be untargetable, or this test proves nothing")
	}
}

// benthicExplorersAbility is Benthic Explorers' mana ability: "{T},
// Untap a tapped land an opponent controls: Add one mana of any type
// that land could produce." The produced mana is read off
// PaidCost.UntappedOthers, the record the untap payment writes.
func benthicExplorersAbility() []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:   true,
		TapOthers: untapOpponentLandCost(),
		ProducedForPaid: func(g *Game, _, _ uuid.UUID, paid PaidCost) ManaOutput {
			if len(paid.UntappedOthers) != 1 {
				return ManaOutput{}
			}
			land, ok := g.LookupCardForEffect(paid.UntappedOthers[0].ID)
			if !ok {
				return ManaOutput{}
			}
			types := g.ProducibleManaLocked(land)
			if len(types) == 0 {
				return ManaOutput{}
			}
			return ManaOutput{Amount: ManaAmount{Options: types, N: 1, OneColor: true}}
		},
		DerivesFromOtherSources: true,
		Label:                   "{T}, Untap a tapped land an opponent controls: Add one mana of any type that land could produce",
	}}
}

// End to end, and the acceptance the issue asks for: the opponent's
// Island is untapped and {U} lands in the pool (CR 106.7, CR 605.3b).
func TestUntapOpponentOnAManaAbilityAddsWhatThatLandCouldProduce(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushIntrinsicPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", benthicExplorersAbility(), nil)
	island := pushTapOthersPermanent(g, opp, "Island", "Land — Island", true)
	g.mu.Lock()
	if c := findBattlefieldCard(g, src); c != nil {
		c.Controller = me.ID
	}
	g.mu.Unlock()

	if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{TapIDs: []uuid.UUID{island}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("U"); got != 1 {
		t.Errorf("pool = %v, want one {U} — the type the untapped Island could produce", me.ManaPool)
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	if c := findBattlefieldCard(g, island); c == nil || c.Tapped {
		t.Error("the opponent's Island was not untapped by the payment")
	}
	if c := findBattlefieldCard(g, src); c == nil || !c.Tapped {
		t.Error("the {T} half of the cost was not paid")
	}
}

// The mana half's refusals, and that none of them pays anything: the
// same validate-all-then-pay discipline (ADR 0020 §3, CR 601.2h).
func TestUntapOpponentManaAbilityRefusesAndPaysNothing(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushIntrinsicPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", benthicExplorersAbility(), nil)
	mine := pushTapOthersPermanent(g, me, "My Island", "Land — Island", true)
	theirUntapped := pushTapOthersPermanent(g, opp, "Their Forest", "Land — Forest", false)
	g.mu.Lock()
	if c := findBattlefieldCard(g, src); c != nil {
		c.Controller = me.ID
	}
	g.mu.Unlock()

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
	}{
		{"nothing named", nil},
		{"my own tapped land", []uuid.UUID{mine}},
		{"their untapped land", []uuid.UUID{theirUntapped}},
	} {
		if err := g.ActivateManaAbility(me.ID, src, 0, ManaAbilityParams{TapIDs: tc.ids}); err == nil {
			t.Errorf("%s: the activation was accepted", tc.name)
		}
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if c := findBattlefieldCard(g, src); c == nil || c.Tapped {
		t.Error("a refused activation tapped the source")
	}
	if c := findBattlefieldCard(g, mine); c == nil || !c.Tapped {
		t.Error("a refused activation untapped my land")
	}
}

// The record the mana output reads: the untap form writes
// UntappedOthers and leaves TappedOthers alone, because nothing was
// tapped (#6208, #5669).
func TestUntapOpponentRecordsWhatItUntappedAndNotWhatItTapped(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTapOthersAbilitySource(g, me, AbilityCost{TapOthers: untapOpponentLandCost()})
	island := pushTapOthersPermanent(g, opp, "Island", "Land — Island", true)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{TapIDs: []uuid.UUID{island}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	item := onlyStackItemForTest(t, g)
	if item == nil {
		t.Fatalf("StackMeta has %d items, want 1", len(g.StackMeta))
	}
	paid := item.Paid
	if len(paid.UntappedOthers) != 1 || paid.UntappedOthers[0].ID != island {
		t.Errorf("UntappedOthers = %v, want the one land the cost untapped", paid.UntappedOthers)
	}
	if len(paid.TappedOthers) != 0 {
		t.Errorf("TappedOthers = %v, want none — nothing was tapped", paid.TappedOthers)
	}
	if c := findBattlefieldCard(g, island); c == nil || c.Tapped {
		t.Error("the opponent's land was not untapped")
	}
}

// The seam's quietest trap: every reader that asks what a mana ability
// WOULD add — the view's up-front colour pick, ManaAbilityAddsNoMana,
// CR 106.7's "could produce" — asks it before any cost is paid. For a
// clause whose output is the untapped permanent's own, that answer has
// to stand in the payment (CR 106.7 ignores whether the cost could be
// paid), or the ability would read as adding nothing and never be
// offered at all.
func TestUntapOpponentManaAbilityIsOfferedBeforeThePaymentIsNamed(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushIntrinsicPermanent(g, me, "Benthic Explorers", "Creature — Merfolk", benthicExplorersAbility(), nil)
	g.mu.Lock()
	if c := findBattlefieldCard(g, src); c != nil {
		c.Controller = me.ID
	}
	g.mu.Unlock()
	ab := benthicExplorersAbility()[0]

	// With nothing to untap the cost cannot be paid and the ability adds
	// nothing: CR 118.3 keeps it off the menu either way.
	g.mu.RLock()
	bare := ManaAbilityAddsNoMana(g, me.ID, src, ab)
	g.mu.RUnlock()
	if !bare {
		t.Error("with no opponent's tapped land the ability was read as adding mana")
	}

	island := pushTapOthersPermanent(g, opp, "Island", "Land — Island", true)
	g.mu.RLock()
	offered := !ManaAbilityAddsNoMana(g, me.ID, src, ab)
	colors := ManaAbilityColorOptions(g, me.ID, src, ab)
	g.mu.RUnlock()
	if !offered {
		t.Error("with their tapped Island on the board the ability was read as adding nothing")
	}
	// And no colour is published up front: which colours the slot offers
	// is decided by WHICH land the activator untaps, so the slot asks its
	// own mana_pick.
	if colors != nil {
		t.Errorf("color options = %v, want none until the land is named", colors)
	}
	_ = island
}
