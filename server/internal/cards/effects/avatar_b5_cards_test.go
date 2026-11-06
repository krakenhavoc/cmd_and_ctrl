package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b5_cards_test.go — Avatar: The Last Airbender batch 5 (#4459).

const (
	avB5IguanaParrotOracle     = "edd3a3ea-6bde-408c-8b16-85ed3cc312b2"
	avB5SubmersibleOracle      = "97369152-3b85-4cf1-aeb9-b3a8f1c37611"
	avB5QuenchOracle           = "cb184995-700f-4900-994a-4ce7fd12f942"
	avB5KataraOracle           = "55adde23-c818-40fe-9bba-123de93568a5"
	avB5KnowledgeSeekerOracle  = "bdc2d980-ca14-4040-a9a1-786d724dc65c"
	avB5LostDaysOracle         = "3bc47eae-cf6f-4b5e-96e1-94a4f307f474"
	avB5MasterPakkuOracle      = "b1727c07-c829-4d64-a1c5-8b53c743888a"
	avB5MechanistOracle        = "bad27b43-7716-4ce1-82ba-2a6d437363e7"
	avB5NorthPolePatrolOracle  = "a1230e78-0bfa-489e-a1b9-3c227bceb6fb"
	avB5OctopusFormOracle      = "6fec368d-374d-4318-861b-ff478ba51a3c"
	avB5MechanistTypeLine      = "Legendary Creature — Human Artificer Ally"
	avB5NorthPolePatrolTypeLne = "Creature — Human Soldier Ally"
)

func TestAvatarB5CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB5IguanaParrotOracle:    "Iguana Parrot",
		avB5SubmersibleOracle:     "Invasion Submersible",
		avB5QuenchOracle:          "It'll Quench Ya!",
		avB5KataraOracle:          "Katara, Bending Prodigy",
		avB5KnowledgeSeekerOracle: "Knowledge Seeker",
		avB5LostDaysOracle:        "Lost Days",
		avB5MasterPakkuOracle:     "Master Pakku",
		avB5MechanistOracle:       "The Mechanist, Aerial Artisan",
		avB5NorthPolePatrolOracle: "North Pole Patrol",
		avB5OctopusFormOracle:     "Octopus Form",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is not Full", name)
		}
	}
}

func avB5HasType(t *testing.T, g *game.Game, id uuid.UUID, typ string) bool {
	t.Helper()
	for _, s := range layeredCard(t, g, id).Effective().Types {
		if strings.EqualFold(s, typ) {
			return true
		}
	}
	return false
}

func avB5HasSubtype(t *testing.T, g *game.Game, id uuid.UUID, sub string) bool {
	t.Helper()
	for _, s := range effectiveSubtypes(t, g, id) {
		if strings.EqualFold(s, sub) {
			return true
		}
	}
	return false
}

func avB5Clues(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && IsToken(c) && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Iguana Parrot -------------------------------------------------------------

func TestAvatarB5IguanaParrotFliesWatchesAndHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	parrot := pushCatalogPermanent(g, me.ID, "Iguana Parrot", "Creature — Lizard Bird Pirate", avB5IguanaParrotOracle, false)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, parrot, kw) {
			t.Errorf("Iguana Parrot has no %s", kw)
		}
	}
	base := effectivePower(t, g, parrot)
	kataraCastInstant(t, g, me)
	if got := effectivePower(t, g, parrot); got != base+1 {
		t.Errorf("power %d after a noncreature spell, want %d (prowess)", got, base+1)
	}
}

// --- Invasion Submersible ------------------------------------------------------

func TestAvatarB5InvasionSubmersibleBouncesThenWaterbendsIntoACreatureOnce(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	sub := castCatalogSpell(t, g, "Invasion Submersible", "Artifact — Vehicle", avB5SubmersibleOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Fatal("the targeted creature was not returned")
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("the bounced creature is not in its owner's hand")
	}
	if avB5HasType(t, g, sub, "Creature") {
		t.Fatal("the Vehicle is a creature before its exhaust")
	}

	helpers := []uuid.UUID{
		pushVanillaCreature(g, me.ID, "Ally", 1, 1),
		pushVanillaCreature(g, me.ID, "Ally", 1, 1),
		pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false),
	}
	if err := g.ActivateCatalogAbility(me.ID, sub, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	}); err != nil {
		t.Fatalf("Exhaust — Waterbend {3} with three taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB5HasType(t, g, sub, "Creature") || !avB5HasType(t, g, sub, "Artifact") {
		t.Error("the Vehicle did not become an artifact creature")
	}
	if got := countersOn(g, sub, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if got := effectivePowerWithCounters(t, g, sub); got != 3 {
		t.Errorf("power %d, want 3", got)
	}

	// The animation has no duration: still a creature next turn.
	advanceToNextSeatsTurn(t, g)
	if !avB5HasType(t, g, sub, "Creature") {
		t.Error("the animation ended; it has no printed duration")
	}

	// Exhaust: once only.
	advanceToMainOf(t, g, avB5SeatIndex(g, me.ID))
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, sub, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second exhaust activation was allowed")
	}
}

func avB5SeatIndex(g *game.Game, id uuid.UUID) int {
	for i, s := range g.Seats {
		if s.ID == id {
			return i
		}
	}
	return -1
}

func TestAvatarB5InvasionSubmersibleWaterbendCannotBePaidShort(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sub := pushCatalogPermanent(g, me.ID, "Invasion Submersible", "Artifact — Vehicle", avB5SubmersibleOracle, false)
	helpers := []uuid.UUID{pushVanillaCreature(g, me.ID, "Ally", 1, 1), pushVanillaCreature(g, me.ID, "Ally", 1, 1)}
	if err := g.ActivateCatalogAbility(me.ID, sub, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	}); err == nil {
		t.Fatal("two taps and an empty pool paid Waterbend {3}")
	}
	for _, id := range helpers {
		if isTapped(g, id) {
			t.Error("a refused activation tapped a waterbender")
		}
	}
}

// --- It'll Quench Ya! ----------------------------------------------------------

func TestAvatarB5ItllQuenchYaTaxesAnySpell(t *testing.T) {
	for _, pay := range []bool{false, true} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			toMainForCost(t, g)
			victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			castFromSeat(t, g, them, "It'll Quench Ya!", "Instant — Lesson", avB5QuenchOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passUntilTaxed(t, g, me.ID)
			if !g.Stack.Contains(victim) {
				t.Fatal("the creature spell left the stack before the tax was answered")
			}
			if pay {
				me.ManaPool.AddMana(game.ManaToken{Color: "C"})
				me.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
			answerPayUnless(t, g, me.ID, pay)
			if g.Stack.Contains(victim) == !pay {
				t.Errorf("paid=%v: the spell is still on the stack = %v", pay, g.Stack.Contains(victim))
			}
		})
	}
}

// --- Katara, Bending Prodigy ---------------------------------------------------

func TestAvatarB5KataraGrowsAtYourEndStepOnlyIfTapped(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		name := "untapped"
		if tapped {
			name = "tapped"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			katara := pushCatalogPermanent(g, me.ID, "Katara, Bending Prodigy", "Legendary Creature — Human Warrior Ally", avB5KataraOracle, false)
			if tapped {
				g.WithWriteLock(func() {
					for i := range g.Battlefield.Cards {
						if g.Battlefield.Cards[i].InstanceID == katara {
							g.Battlefield.Cards[i].Tapped = true
						}
					}
				})
			}
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			want := 0
			if tapped {
				want = 1
			}
			if got := countersOn(g, katara, game.CounterPlusOne); got != want {
				t.Errorf("%d +1/+1 counters, want %d", got, want)
			}
		})
	}
}

// The intervening "if" is checked again on resolution (CR 603.4): untapped
// in response, she gets nothing.
func TestAvatarB5KataraRechecksTappedOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	katara := pushCatalogPermanent(g, me.ID, "Katara, Bending Prodigy", "Legendary Creature — Human Warrior Ally", avB5KataraOracle, false)
	setTapped := func(v bool) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == katara {
					g.Battlefield.Cards[i].Tapped = v
				}
			}
		})
	}
	setTapped(true)
	advanceToEndStep(t, g)
	if len(g.StackMeta) == 0 {
		t.Fatal("the end-step trigger did not go on the stack")
	}
	setTapped(false)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, katara, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after she untapped in response, want 0", got)
	}
}

func TestAvatarB5KataraWaterbendsSixToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	katara := pushCatalogPermanent(g, me.ID, "Katara, Bending Prodigy", "Legendary Creature — Human Warrior Ally", avB5KataraOracle, true)
	helpers := []uuid.UUID{katara}
	for i := 0; i < 4; i++ {
		helpers = append(helpers, pushVanillaCreature(g, me.ID, "Ally", 1, 1))
	}
	helpers = append(helpers, pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false))

	// Five taps and an empty pool is short.
	if err := g.ActivateCatalogAbility(me.ID, katara, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers[:5], Strict: true,
	}); err == nil {
		t.Fatal("five taps and no mana paid Waterbend {6}")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, katara, 0, game.ActivateAbilityParams{
		WaterbendIDs: helpers, Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {6} with six taps (Katara herself, summoning sick, among them): %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	// Tapped by her own waterbend, she grows at the end step.
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, katara, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after waterbending with herself, want 1", got)
	}
}

// --- Knowledge Seeker ----------------------------------------------------------

func TestAvatarB5KnowledgeSeekerGrowsOnTheSecondDraw(t *testing.T) {
	hobA5SecondDrawGrows(t, "Knowledge Seeker", "Creature — Fox Spirit", avB5KnowledgeSeekerOracle)
	g := newCatalogGame(t)
	seeker := pushCatalogPermanent(g, g.Seats[0].ID, "Knowledge Seeker", "Creature — Fox Spirit", avB5KnowledgeSeekerOracle, false)
	if !hasKeywordOnBattlefield(t, g, seeker, "vigilance") {
		t.Error("Knowledge Seeker has no vigilance")
	}
}

func TestAvatarB5KnowledgeSeekerLeavesAClueWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seeker := pushCatalogPermanent(g, me.ID, "Knowledge Seeker", "Creature — Fox Spirit", avB5KnowledgeSeekerOracle, false)
	killCreature(t, g, me.ID, seeker)
	passPriorityAroundTable(t, g)
	if got := len(avB5Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Lost Days -----------------------------------------------------------------

func TestAvatarB5LostDaysTucksByTheOwnersChoiceAndMakesAClue(t *testing.T) {
	for _, bury := range []bool{false, true} {
		name := "second from the top"
		if bury {
			name = "on the bottom"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			lib := seedLibrary(opp, "Top", "Next")
			ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
			castCatalogSpell(t, g, "Lost Days", "Instant — Lesson", avB5LostDaysOracle, cardRefs(ench))
			passPriorityAroundTable(t, g)
			if putInLibraryChoiceFor(g, me.ID) != nil {
				t.Fatal("the caster was asked; the owner chooses")
			}
			ask := putInLibraryChoiceFor(g, opp.ID)
			if ask == nil || ask.LibraryPlacement != game.LibraryPlaceTopOrBottom || ask.LibraryTopDepth != 2 {
				t.Fatalf("the owner is asked, top lane at depth 2: %+v", ask)
			}
			if n := len(avB5Clues(g, me.ID)); n != 0 {
				t.Fatalf("the Clue came before the owner's answer (%d)", n)
			}
			var top, bottom []uuid.UUID
			if bury {
				bottom = []uuid.UUID{ench}
			} else {
				top = []uuid.UUID{ench}
			}
			if err := g.ResolvePutInLibrary(ask.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if onBattlefield(g, ench) {
				t.Fatal("the enchantment is still on the battlefield")
			}
			if bury {
				if opp.Library.Cards[0].InstanceID != ench {
					t.Error("not on the bottom of its owner's library")
				}
			} else if got := libraryTopIDs(opp, 3); !sameIDs(got, []uuid.UUID{lib[0], ench, lib[1]}) {
				t.Errorf("top three %v, want [Top Their-Enchantment Next]", got)
			}
			if n := len(avB5Clues(g, me.ID)); n != 1 {
				t.Errorf("%d Clues for the caster, want 1", n)
			}
			if n := len(avB5Clues(g, opp.ID)); n != 0 {
				t.Errorf("the owner got %d Clues", n)
			}
		})
	}
}

func TestAvatarB5LostDaysRefusesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	if err := castCatalogSpellErr(t, g, "Lost Days", "Instant — Lesson", avB5LostDaysOracle, cardRefs(relic)); err == nil {
		t.Error("Lost Days targeted a noncreature, nonenchantment permanent")
	}
}

// --- Master Pakku --------------------------------------------------------------

func TestAvatarB5MasterPakkuMillsTargetPlayerForYourLessons(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pakku := pushCatalogPermanent(g, me.ID, "Master Pakku", "Legendary Creature — Human Advisor Ally", avB5MasterPakkuOracle, false)
	if !hasKeywordOnBattlefield(t, g, pakku, "prowess") {
		t.Error("Master Pakku has no prowess")
	}
	for _, c := range []game.Card{
		{Name: "Lesson A", TypeLine: "Instant — Lesson"},
		{Name: "Lesson B", TypeLine: "Sorcery — Lesson"},
		{Name: "Not a Lesson", TypeLine: "Instant"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	// A Lesson in the TARGET's graveyard does not count.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Lesson", TypeLine: "Instant — Lesson", Owner: opp.ID, Controller: opp.ID})
	seedLibrary(opp, "1", "2", "3", "4")
	oppGrave := len(opp.Graveyard.Cards)
	oppLib := opp.Library.Size()

	declareAttack(t, g, opp.ID, pakku)
	lockInAttacks(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2 (two Lessons in your graveyard)", got)
	}
	if got := len(opp.Graveyard.Cards) - oppGrave; got != 2 {
		t.Errorf("%d cards reached the target's graveyard, want 2", got)
	}
}

// --- The Mechanist, Aerial Artisan ---------------------------------------------

func TestAvatarB5MechanistMakesClueAndAnimatesIt(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mech := pushCatalogPermanent(g, me.ID, "The Mechanist, Aerial Artisan", avB5MechanistTypeLine, avB5MechanistOracle, false)
	kataraCastInstant(t, g, me)
	clues := avB5Clues(g, me.ID)
	if len(clues) != 1 {
		t.Fatalf("%d Clues after a noncreature spell, want 1", len(clues))
	}
	// A creature spell makes none.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := len(avB5Clues(g, me.ID)); got != 1 {
		t.Fatalf("%d Clues after a creature spell, want still 1", got)
	}

	clue := clues[0]
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: cardRefs(clue), Strict: true}); err != nil {
		t.Fatalf("{T}: animate the Clue: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !avB5HasType(t, g, clue, "Creature") || !avB5HasType(t, g, clue, "Artifact") {
		t.Fatal("the Clue is not an artifact creature")
	}
	if !avB5HasSubtype(t, g, clue, "Construct") || !avB5HasSubtype(t, g, clue, "Clue") {
		t.Errorf("subtypes %v, want Construct and the Clue's own Clue", effectiveSubtypes(t, g, clue))
	}
	if p, tt := effectivePower(t, g, clue), effectiveToughness(t, g, clue); p != 3 || tt != 1 {
		t.Errorf("%d/%d, want 3/1", p, tt)
	}
	if !hasEffectiveKeyword(t, g, clue, "flying") {
		t.Error("the Construct has no flying")
	}

	advanceToNextSeatsTurn(t, g)
	if avB5HasType(t, g, clue, "Creature") {
		t.Error("the Clue is still a creature after end of turn")
	}
}

// "Construct artifact creature" replaces the creature types an artifact
// creature token already had (CR 205.1b).
func TestAvatarB5MechanistReplacesExistingCreatureTypes(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mech := pushCatalogPermanent(g, me.ID, "The Mechanist, Aerial Artisan", avB5MechanistTypeLine, avB5MechanistOracle, false)
	var thopter uuid.UUID
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Thopter", TypeLine: "Token Artifact Creature — Thopter", Power: 1, Toughness: 1}, 1)
	})
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			thopter = c.InstanceID
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: cardRefs(thopter), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if avB5HasSubtype(t, g, thopter, "Thopter") || !avB5HasSubtype(t, g, thopter, "Construct") {
		t.Errorf("subtypes %v, want Construct only", effectiveSubtypes(t, g, thopter))
	}
	if p, tt := effectivePower(t, g, thopter), effectiveToughness(t, g, thopter); p != 3 || tt != 1 {
		t.Errorf("%d/%d, want 3/1", p, tt)
	}
}

func TestAvatarB5MechanistRefusesANontokenArtifact(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mech := pushCatalogPermanent(g, me.ID, "The Mechanist, Aerial Artisan", avB5MechanistTypeLine, avB5MechanistOracle, false)
	relic := pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: cardRefs(relic), Strict: true}); err == nil {
		t.Error("a nontoken artifact was a legal target")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, ClueToken(), 1)
	})
	theirs := avB5Clues(g, opp.ID)
	if len(theirs) != 1 {
		t.Fatalf("seeded %d opposing Clues", len(theirs))
	}
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs[0]), Strict: true}); err == nil {
		t.Error("an opponent's artifact token was a legal target")
	}
}

// --- North Pole Patrol ---------------------------------------------------------

func TestAvatarB5NorthPolePatrolUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	patrol := pushCatalogPermanent(g, me.ID, "North Pole Patrol", avB5NorthPolePatrolTypeLne, avB5NorthPolePatrolOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, patrol, 0, game.ActivateAbilityParams{Targets: cardRefs(patrol), Strict: true}); err == nil {
		t.Error("the Patrol untapped itself; it says another")
	}
	if err := g.ActivateCatalogAbility(me.ID, patrol, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("untap the Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("the Bear is still tapped")
	}
	if !isTapped(g, patrol) {
		t.Error("the Patrol did not tap")
	}
}

func TestAvatarB5NorthPolePatrolWaterbendsToTapAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	patrol := pushCatalogPermanent(g, me.ID, "North Pole Patrol", avB5NorthPolePatrolTypeLne, avB5NorthPolePatrolOracle, false)
	theirs := pushVanillaCreature(g, opp.ID, "Their Ogre", 3, 3)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	helpers := []uuid.UUID{
		pushVanillaCreature(g, me.ID, "Ally", 1, 1),
		pushVanillaCreature(g, me.ID, "Ally", 1, 1),
		pushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false),
	}
	if err := g.ActivateCatalogAbility(me.ID, patrol, 1, game.ActivateAbilityParams{
		Targets: cardRefs(mine), WaterbendIDs: helpers, Strict: true,
	}); err == nil {
		t.Fatal("a creature you control was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, patrol, 1, game.ActivateAbilityParams{
		Targets: cardRefs(theirs), WaterbendIDs: append([]uuid.UUID{patrol}, helpers[:2]...), Strict: true,
	}); err == nil {
		t.Fatal("the Patrol paid both its {T} and a waterbend tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, patrol, 1, game.ActivateAbilityParams{
		Targets: cardRefs(theirs), WaterbendIDs: helpers, Strict: true,
	}); err != nil {
		t.Fatalf("Waterbend {3}, {T} with three taps: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, theirs) {
		t.Error("the opposing creature is untapped")
	}
	if !isTapped(g, patrol) {
		t.Error("the Patrol did not tap for {T}")
	}
}

// --- Octopus Form --------------------------------------------------------------

func TestAvatarB5OctopusFormPumpsShieldsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Octopus Form", "Instant — Lesson", avB5OctopusFormOracle, cardRefs(theirs)); err == nil {
		t.Error("Octopus Form targeted a creature you don't control")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	castCatalogSpell(t, g, "Octopus Form", "Instant — Lesson", avB5OctopusFormOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("%d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("no hexproof")
	}
	if isTapped(g, bear) {
		t.Error("the creature was not untapped")
	}
	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "hexproof") || effectivePower(t, g, bear) != 2 {
		t.Error("the pump or the hexproof outlived the turn")
	}
}
