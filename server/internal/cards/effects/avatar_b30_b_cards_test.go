package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_b_cards_test.go — Avatar: The Last Airbender Eternal, batch
// 31's group-b slice (#5075).

const (
	avB30bStormOracle   = "6f8ca795-d6fd-4e3e-911c-c621a942acbb"
	avB30bZukoOracle    = "33db6dc8-2c88-4a7f-8d4d-3bc9fb6544b2"
	avB30bAttendOracle  = "787e6cc6-d97a-4229-97d1-ec2377d138d3"
	avB30bTeaOracle     = "9ed4e48f-91bd-45f9-9f31-a2705b330713"
	avB30bWhistleOracle = "47e77918-13ec-47ce-9256-5a3b5bf8d315"
	avB30bBoscoOracle   = "a213c650-06cc-4ea8-86f8-78ccfe103b38"
	avB30bFeastOracle   = "41f53bb5-8a62-4e6b-b9c3-5aa530a036bc"
	avB30bCrackedOracle = "d5314006-afe0-4cdc-88f0-ff1f7d85fba9"
	avB30bCoatingOracle = "c23f2546-bf9e-4390-b6f7-c616920e04ca"
	avB30bArmorOracle   = "93d2638b-0c19-4fba-a0ad-84b6a9784eb3"
)

func TestAvatarB30bCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB30bStormOracle:   "Storm of Memories",
		avB30bZukoOracle:    "Zuko, Firebending Master",
		avB30bAttendOracle:  "Animal Attendant",
		avB30bTeaOracle:     "The Art of Tea",
		avB30bWhistleOracle: "Bison Whistle",
		avB30bBoscoOracle:   "Bosco, Just a Bear",
		avB30bFeastOracle:   "Bumi's Feast Lecture",
		avB30bCrackedOracle: "Cracked Earth Technique",
		avB30bCoatingOracle: "Creeping Crystal Coating",
		avB30bArmorOracle:   "Crystalline Armor",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB30bGraveyard puts a card into p's graveyard and returns its ID.
func avB30bGraveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// avB30bInExile reports whether id is in some exile zone.
func avB30bInExile(g *game.Game, id uuid.UUID) bool {
	inExile := false
	g.ReadSnapshot(func() {
		z := g.FindCardZoneForEffect(id)
		inExile = z != nil && z.Kind == game.ZoneExile
	})
	return inExile
}

// avB30bFoods is how many Foods p controls.
func avB30bFoods(g *game.Game, p uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		n = countControlled(g, p, func(c game.Card) bool { return c.HasSubtype("Food") })
	})
	return n
}

// --- Storm of Memories ---------------------------------------------------------

// The only instant or sorcery with mana value 3 or less is exiled and
// offered free inside the resolution; cast, it is exiled again instead of
// going to the graveyard. The mana value 5 sorcery and the creature card
// are never candidates.
func TestAvatarB30StormOfMemoriesCastsTheExiledCardFreeAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	shock := avB30bGraveyard(g, me, game.Card{Name: "Small Instant", TypeLine: "Instant", ManaCost: "{1}{R}"})
	big := avB30bGraveyard(g, me, game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}"})
	bear := avB30bGraveyard(g, me, game.Card{Name: "Bear Card", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})

	castCatalogSpell(t, g, "Storm of Memories", "Sorcery", avB30bStormOracle, nil)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, shock)
	if !avB30bInExile(g, shock) {
		t.Fatal("the mana value 2 instant was not exiled")
	}
	if !me.Graveyard.Contains(big) || !me.Graveyard.Contains(bear) {
		t.Fatal("an ineligible card left the graveyard")
	}
	if err := castThroughResolutionCast(t, g, me.ID, shock, game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	castInsideTheResolution(t, g, shock)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(shock) || !avB30bInExile(g, shock) {
		t.Error("the free spell went to the graveyard instead of exile")
	}
}

// Storm count 1: the copy and the original each exile one card at random
// from the eligible three (the filler instant that made the count is one
// of them); declined, both stay in exile and the third stays put.
func TestAvatarB30StormOfMemoriesEachCopyExilesItsOwnCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := avB30bGraveyard(g, me, game.Card{Name: "Instant A", TypeLine: "Instant", ManaCost: "{U}"})
	b := avB30bGraveyard(g, me, game.Card{Name: "Sorcery B", TypeLine: "Sorcery", ManaCost: "{2}{B}"})

	filler := stormFiller(t, g)
	castCatalogSpell(t, g, "Storm of Memories", "Sorcery", avB30bStormOracle, nil)
	offers := 0
	for i := 0; i < 16 && !stackFullyEmpty(g); i++ {
		if resolutionCastOfferFor(g, me.ID) != nil {
			offers++
			declineResolutionCast(t, g, me.ID)
			continue
		}
		passPriorityAroundTable(t, g)
	}
	if offers != 2 {
		t.Errorf("%d free-cast offers, want 2 (the copy and the original)", offers)
	}
	exiled, left := 0, 0
	for _, id := range []uuid.UUID{a, b, filler} {
		if avB30bInExile(g, id) {
			exiled++
		}
		if me.Graveyard.Contains(id) {
			left++
		}
	}
	if exiled != 2 || left != 1 {
		t.Errorf("%d exiled and %d left in the graveyard, want 2 and 1 (one card per resolution)", exiled, left)
	}
}

// --- Zuko, Firebending Master ---------------------------------------------------

// A spell cast during combat is an experience counter; one cast in a main
// phase is not. Attacking with two counters adds {R}{R} that lasts until
// end of combat.
func TestAvatarB30ZukoCountsCombatSpellsAndFirebendsPerCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zuko := avB11Creature(g, me.ID, "Zuko, Firebending Master", avB30bZukoOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, zuko, "first strike") {
		t.Error("Zuko has no first strike")
	}
	advanceToMain(t, g)
	kataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 0 {
		t.Fatalf("experience = %d after a main-phase spell, want 0", got)
	}
	if err := g.AddPlayerCounterForEffect(me.ID, game.CounterExperience, 2); err != nil {
		t.Fatalf("seed experience: %v", err)
	}
	declareAttack(t, g, opp.ID, zuko)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
	kataraCastInstant(t, g, me)
	if got := me.Counters[game.CounterExperience]; got != 3 {
		t.Errorf("experience = %d after a combat spell, want 3", got)
	}
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

// With no experience counters, firebending adds nothing.
func TestAvatarB30ZukoWithNoExperienceAddsNoMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zuko := avB11Creature(g, me.ID, "Zuko, Firebending Master", avB30bZukoOracle, 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, zuko)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{})
}

// --- Animal Attendant ----------------------------------------------------------

func TestAvatarB30AnimalAttendantCountersOnlyANonHumanCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		counters       int
	}{
		{"a Bear", "Creature — Bear", 1},
		{"a Human", "Creature — Human Soldier", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			attendant := seedPermanentWithOracle(g, me.ID, "Animal Attendant", "Creature — Human Citizen", avB30bAttendOracle)
			activateManaFor(t, g, me.ID, attendant, 0, game.ManaAbilityParams{Colors: []string{"G"}})
			creature := castFromHandForTest(t, g, me, "Test Creature", tc.typeLine, "{G}", "", strict)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, creature, "+1/+1"); got != tc.counters {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.counters)
			}
		})
	}
}

// --- The Art of Tea --------------------------------------------------------------

func TestAvatarB30ArtOfTeaCountersYourCreatureAndMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "The Art of Tea", "Instant — Lesson", avB30bTeaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("The Art of Tea targeted an opponent's creature")
	}
	castCatalogSpell(t, g, "The Art of Tea", "Instant — Lesson", avB30bTeaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := avB30bFoods(g, me.ID); got != 1 {
		t.Errorf("Foods = %d, want 1", got)
	}
	// Up to one: with no target it still makes the Food.
	castCatalogSpell(t, g, "The Art of Tea", "Instant — Lesson", avB30bTeaOracle, nil)
	passPriorityAroundTable(t, g)
	if got := avB30bFoods(g, me.ID); got != 2 {
		t.Errorf("Foods = %d after an untargeted cast, want 2", got)
	}
}

// --- Bison Whistle ----------------------------------------------------------------

func avB30bActivateWhistle(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	whistle := pushCatalogPermanent(g, me.ID, "Bison Whistle", "Artifact", avB30bWhistleOracle, false)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, whistle, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Bison Whistle: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestAvatarB30BisonWhistlePutsABisonOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bison := plTop(me, "Sky Bison", "Creature — Bison", "{4}{W}")
	avB30bActivateWhistle(t, g, me)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("Bison Whistle did not offer the Bison")
	}
	answerChooseCards(t, g, me.ID, bison)
	if !g.Battlefield.Contains(bison) {
		t.Error("the Bison did not enter the battlefield")
	}
}

// A Bison creature declined to the battlefield may still go to hand.
func TestAvatarB30BisonWhistleDeclinedBisonCreatureMayGoToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bison := plTop(me, "Sky Bison", "Creature — Bison", "{4}{W}")
	avB30bActivateWhistle(t, g, me)
	answerChooseCards(t, g, me.ID)
	answerMayChoice(t, g, me.ID, true)
	if !me.Hand.Contains(bison) {
		t.Error("the declined Bison creature should be offered to hand")
	}
}

func TestAvatarB30BisonWhistleCreatureToHandOrLeftOnTop(t *testing.T) {
	for _, take := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		bear := plTop(me, "Bear Card", "Creature — Bear", "{1}{G}")
		avB30bActivateWhistle(t, g, me)
		if chooseCardsChoiceFor(g, me.ID) != nil {
			t.Fatal("a non-Bison creature is not offered to the battlefield")
		}
		answerMayChoice(t, g, me.ID, take)
		if take && !me.Hand.Contains(bear) {
			t.Error("accepted: the creature goes to hand")
		}
		if top, err := me.Library.Top(); !take && (err != nil || top.InstanceID != bear) {
			t.Error("declined: the creature stays on top")
		}
		if me.Graveyard.Contains(bear) {
			t.Error("a creature card is never offered to the graveyard")
		}
	}
}

func TestAvatarB30BisonWhistleNoncreatureMayGoToTheGraveyard(t *testing.T) {
	for _, bin := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
		avB30bActivateWhistle(t, g, me)
		answerMayChoice(t, g, me.ID, bin)
		if bin != me.Graveyard.Contains(bolt) {
			t.Errorf("answer %v: in graveyard = %v", bin, me.Graveyard.Contains(bolt))
		}
		if me.Hand.Contains(bolt) {
			t.Error("a noncreature card never goes to hand")
		}
	}
}

// --- Bosco, Just a Bear ---------------------------------------------------------

func TestAvatarB30BoscoMakesAFoodPerLegendAndEatsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Some Legend", "Legendary Creature — Human", 1, 1)
	b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	bosco := castCatalogSpell(t, g, "Bosco, Just a Bear", "Legendary Creature — Bear", avB30bBoscoOracle, nil)
	passPriorityAroundTable(t, g)
	if got := avB30bFoods(g, me.ID); got != 2 {
		t.Fatalf("Foods = %d, want 2 (Bosco and the other legend)", got)
	}
	food := socB6Named(g, "Food")[0]
	if err := g.ActivateCatalogAbility(me.ID, bosco, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}}); err == nil {
		t.Fatal("activated with no mana")
	}
	b06AddMana(me, "G", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bosco, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bosco, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, bosco, "trample") {
		t.Error("Bosco did not gain trample")
	}
}

// --- Bumi's Feast Lecture -----------------------------------------------------------

// One Food already, one made: earthbend 4.
func TestAvatarB30BumisFeastLectureEarthbendsTwicePerFoodAfterMakingOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, FoodToken(), 1); err != nil {
			t.Fatalf("food: %v", err)
		}
	})
	forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Bumi's Feast Lecture", "Sorcery — Lesson", avB30bFeastOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	passPriorityAroundTable(t, g)
	if got := avB30bFoods(g, me.ID); got != 2 {
		t.Errorf("Foods = %d, want 2", got)
	}
	if got := countersOn(g, forest, "+1/+1"); got != 4 {
		t.Errorf("the earthbent land has %d +1/+1 counters, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, forest, "haste") {
		t.Error("the earthbent land has no haste")
	}
}

// --- Cracked Earth Technique ---------------------------------------------------------

func TestAvatarB30CrackedEarthTechniqueEarthbendsTwiceAndGainsThree(t *testing.T) {
	for _, same := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		forest := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
		other := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
		second := other
		if same {
			second = forest
		}
		life := me.Life
		castFromHandForTest(t, g, me, "Cracked Earth Technique", "Sorcery — Lesson", "{4}{G}", avB30bCrackedOracle,
			game.CastSpellParams{TargetsStamped: true, Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: forest, Slot: 0},
				{Kind: game.TargetCard, ID: second, Slot: 1},
			}})
		passPriorityAroundTable(t, g)
		if same {
			if got := countersOn(g, forest, "+1/+1"); got != 6 {
				t.Errorf("one land earthbent twice has %d counters, want 6", got)
			}
		} else {
			if a, b := countersOn(g, forest, "+1/+1"), countersOn(g, other, "+1/+1"); a != 3 || b != 3 {
				t.Errorf("two lands have %d and %d counters, want 3 and 3", a, b)
			}
		}
		if me.Life != life+3 {
			t.Errorf("life %d → %d, want +3", life, me.Life)
		}
	}
}

// --- Creeping Crystal Coating -------------------------------------------------------

func TestAvatarB30CreepingCrystalCoatingToughensAndFeedsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	coating := castCatalogSpell(t, g, "Creeping Crystal Coating", auraTypeLine, avB30bCoatingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, coating); host.ID != bear {
		t.Fatalf("attached to %+v, want the bear", host)
	}
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tg != 5 {
		t.Fatalf("enchanted bear is %d/%d, want 2/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, coating, "flash") {
		t.Error("Creeping Crystal Coating has no flash")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).SummonedThisTurn = false })
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := avB30bFoods(g, me.ID); got != 1 {
		t.Errorf("Foods = %d after the attack, want 1", got)
	}
}

// --- Crystalline Armor ---------------------------------------------------------------

func TestAvatarB30CrystallineArmorPumpsPerLandYouControlAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	seedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	pushAuraOnLand(t, g, me.ID, "Crystalline Armor", avB30bArmorOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tg != 4 {
		t.Fatalf("enchanted bear is %d/%d with two lands, want 4/4", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the enchanted creature has no trample")
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID})
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("power %d with three lands, want 5", p)
	}
}
