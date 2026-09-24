package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teferi_temporal_archmage_test.go — #1275's proof cards. The −10 is
// the point: an emblem whose whole text is a per-player activation
// timing statement, read from the command zone by the same
// ActivationTimingOpenLocked the battlefield statics go through. The
// +1 and −1 are exercised because the card registers them.

const (
	teferiTemporalArchmageOracle = "07d0b06b-80cb-4518-92c9-84ea87a7e08a"
	teferisTalentOracle          = "efc23669-1bbb-487f-94b3-cc07f3cd35f6"
	teferiHeroOracleForTiming    = "f2f165b6-ef0a-42ad-9352-ba68be8248b0"
)

// With the emblem, a loyalty ability of ANY planeswalker its owner
// controls activates on an opponent's end step — and CR 606.3's other
// half still holds: once per turn per planeswalker, each walker its
// own once.
func TestTeferiTemporalArchmageEmblemOpensLoyaltyOnAnOpponentsTurnOncePerWalker(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	archmage := pushCatalogWalker(g, me.ID, "Teferi, Temporal Archmage", teferiTemporalArchmageOracle, 12)
	hero := pushCatalogWalker(g, me.ID, "Teferi, Hero of Dominaria", teferiHeroOracleForTiming, 4)

	b16Activate(t, g, me.ID, archmage, 2, game.ActivateAbilityParams{})
	emblems := emblemsOfPlayer(g, me.ID)
	if len(emblems) != 1 || emblems[0].Label != "Teferi, Temporal Archmage emblem" {
		t.Fatalf("emblems after the −10: %+v", emblems)
	}
	if got := loyaltyCount(g, archmage); got != 2 {
		t.Fatalf("loyalty after −10 on a 12-loyalty Archmage: got %d, want 2", got)
	}

	advanceToNextSeatsTurn(t, g)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("setup: the turn did not move to another seat")
	}
	advanceTo(t, g, game.StepEnd)

	// The Archmage's +1 on somebody else's end step: the emblem is the
	// only thing that can open it (CR 606.3 names the active player).
	if err := g.ActivateCatalogAbility(me.ID, archmage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1 on an opponent's end step under the emblem: %v", err)
	}
	// CR 606.3: once per turn per planeswalker. The emblem opens the
	// window; it does not lift the count.
	if err := g.ActivateCatalogAbility(me.ID, archmage, 1, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
		t.Errorf("a second loyalty activation of the same walker: got %v, want ErrLoyaltyAlreadyActivated", err)
	}
	// "Planeswalkers you control" — every one of them, each its own
	// once, and not only the Teferi that made the emblem.
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{}); err != nil {
		t.Errorf("another walker's +1 under the emblem: %v", err)
	}
}

// Without the emblem the same activation is refused for timing, and an
// emblem an OPPONENT has does not open it: the statement is about
// planeswalkers ITS OWNER controls.
func TestTeferiTemporalArchmageLoyaltyRefusedWithoutYourEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hero := pushCatalogWalker(g, me.ID, "Teferi, Hero of Dominaria", teferiHeroOracleForTiming, 4)
	advanceToNextSeatsTurn(t, g)
	opp := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepEnd)

	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("a loyalty ability on an opponent's end step with no emblem: got %v, want ErrSorcerySpeedRequired", err)
	}

	source := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: source, Name: "Teferi, Temporal Archmage",
		TypeLine: "Legendary Planeswalker — Teferi", OracleID: teferiTemporalArchmageOracle,
		Owner: opp.ID, Controller: opp.ID,
	})
	var err error
	g.WithWriteLock(func() { err = g.CreateEmblemForEffect(opp.ID, source) })
	if err != nil {
		t.Fatalf("CreateEmblemForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hero, 0, game.ActivateAbilityParams{}); err != game.ErrSorcerySpeedRequired {
		t.Errorf("an opponent's emblem opened my loyalty ability: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// +1: look at two, one to hand, the other to the bottom. A rest of one
// card has no order, so no put_in_library prompt is raised.
func TestTeferiTemporalArchmagePlusOneTakesOneAndBottomsTheOther(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	archmage := pushCatalogWalker(g, me.ID, "Teferi, Temporal Archmage", teferiTemporalArchmageOracle, 5)
	seeded := seedLibrary(me, "Look1", "Look2", "Filler")
	look1, look2 := seeded[0], seeded[1]
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, archmage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	passPriorityAroundTable(t, g)
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choose-cards prompt for the +1's pick")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{look1}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(look1) || me.Hand.Size() != handBefore+1 {
		t.Fatal("the chosen card did not reach the hand")
	}
	if order := putInLibraryChoiceFor(g, me.ID); order != nil {
		t.Errorf("a pile of one was put up for ordering: %+v", order)
	}
	if got := libraryBottomIDs(me, 1); !sameIDs(got, []uuid.UUID{look2}) {
		t.Errorf("the bottom card is %v, want Look2", got)
	}
	if got := loyaltyCount(g, archmage); got != 6 {
		t.Errorf("loyalty after +1: got %d, want 6", got)
	}
}

// −1: up to four target permanents of anybody's untap.
func TestTeferiTemporalArchmageMinusOneUntapsUpToFour(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	archmage := pushCatalogWalker(g, me.ID, "Teferi, Temporal Archmage", teferiTemporalArchmageOracle, 5)
	a := b12Permanent(g, me.ID, "My Land", "Land")
	b := b12Permanent(g, me.ID, "My Rock", "Artifact")
	c := b12Permanent(g, opp.ID, "Their Land", "Land")
	left := b12Permanent(g, me.ID, "Untargeted Land", "Land")
	tapForTest(t, g, a, b, c, left)

	b16Activate(t, g, me.ID, archmage, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a},
			{Kind: game.TargetCard, ID: b},
			{Kind: game.TargetCard, ID: c},
		},
	})
	for _, id := range []uuid.UUID{a, b, c} {
		if b16Tapped(t, g, id) {
			t.Errorf("target %s is still tapped", id)
		}
	}
	if !b16Tapped(t, g, left) {
		t.Error("a permanent that was not targeted untapped")
	}
	if got := loyaltyCount(g, archmage); got != 4 {
		t.Errorf("loyalty after −1: got %d, want 4", got)
	}
}

// --- the EmblemSpec slot's guard (#1275) -----------------------------

// An emblem's activation timing statement is checked by the same guard
// a permanent's is: one that says nothing, has no label, or covers
// nothing is refused at boot. And an emblem whose only ability is a
// timing statement is an emblem with an ability.
func TestEmblemActivationTimingsAreGuardedLikeAPermanents(t *testing.T) {
	covers := func(game.ActivationQuery) bool { return true }
	for _, tc := range []struct {
		name string
		t    game.ActivationTiming
		want string
	}{
		{"says nothing", game.ActivationTiming{Label: "x", Covers: covers}, "says nothing"},
		{"no label", game.ActivationTiming{Timing: game.TimingFlash, Covers: covers}, "no printed Label"},
		{"covers nothing", game.ActivationTiming{Label: "x", Timing: game.TimingFlash}, "covers nothing"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() {
				Register(Spec{
					OracleID: "emblem-timing-guard-" + tc.name,
					Name:     "Emblem Timing Guard",
					Emblem: &EmblemSpec{
						Label:             "Emblem Timing Guard emblem",
						Text:              "x",
						ActivationTimings: []game.ActivationTiming{tc.t},
					},
				})
			})
		})
	}
}

// --- Teferi's Talent ------------------------------------------------

// "Whenever you draw a card, put a loyalty counter on enchanted
// planeswalker" — once per card, yours only.
func TestTeferisTalentAddsLoyaltyWhenYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	walker := pushCatalogWalker(g, me.ID, "Teferi, Hero of Dominaria", teferiHeroOracleForTiming, 4)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Teferi's Talent", TypeLine: "Enchantment — Aura",
		OracleID: teferisTalentOracle, Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: walker},
	})

	for range 2 {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 6 {
		t.Errorf("loyalty after two draws: got %d, want 6", got)
	}

	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, walker); got != 6 {
		t.Errorf("an opponent's draw moved the loyalty: got %d, want 6", got)
	}
}

// The granted −12 is the seam this card still waits on, and the
// catalog says so rather than shipping a walker with an ability the
// engine cannot give it.
func TestTeferisTalentDeclaresTheMissingGrant(t *testing.T) {
	spec, ok := Lookup(teferisTalentOracle)
	if !ok {
		t.Fatal("Teferi's Talent is not registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("completeness %v, caveats %v — want one caveat naming the −12", spec.Completeness, spec.Caveats)
	}
	if spec.Emblem != nil {
		t.Error("Teferi's Talent declares an emblem nothing can create")
	}
}
