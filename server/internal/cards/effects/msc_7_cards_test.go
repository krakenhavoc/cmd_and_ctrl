package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_7_cards_test.go — Marvel Super Heroes Commander, seventh batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc7SpectrumOracle    = "fad80e75-8d08-4a0b-a05e-14f4c96947bf"
	msc7StrangeOracle     = "43c31a03-2bbd-49a6-8b78-c5f3cab04a07"
	msc7GravestoneOracle  = "cbf5bd0a-20ef-4dce-9f70-13339d438ff7"
	msc7EverettOracle     = "e017e323-e297-40de-bc77-dd49b610cf5a"
	msc7RedwingOracle     = "0f7c7035-e395-4da5-8536-9f192d03e3ba"
	msc7TorresOracle      = "46399b09-1b5c-4c28-9134-cc2ed573b6aa"
	msc7FallToEarthOracle = "223db021-979f-4312-8cb5-169ae4bf17bc"
	msc7BounceOracle      = "71d09e00-e526-4899-beb3-0ee6d9b87f44"
	msc7ThroneOracle      = "53129de9-4809-4c8c-9d11-37369899e70a"
	msc7FirebirdOracle    = "3d917331-9f49-4d0c-bb2d-4853e3ff6ca1"
)

func TestMsc7EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc7SpectrumOracle:    "Doctor Spectrum",
		msc7StrangeOracle:     "Doctor Strange, Surgeon",
		msc7GravestoneOracle:  "Eerie Gravestone",
		msc7EverettOracle:     "Everett K. Ross, Hapless Attaché",
		msc7RedwingOracle:     "Falcon and Redwing",
		msc7TorresOracle:      "Falcon, Joaquin Torres",
		msc7FallToEarthOracle: "Fall to Earth",
		msc7BounceOracle:      "Fantastic Bounce",
		msc7ThroneOracle:      "Fight for the Throne",
		msc7FirebirdOracle:    "Firebird, Blazing Ranger",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// msc7PickMode answers the open mode_pick prompt with one bullet.
func msc7PickMode(t *testing.T, g *game.Game, chooser uuid.UUID, bullet int) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, chooser)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{bullet}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", bullet, err)
	}
}

func msc7Spectrum(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castHoldCreature(t, g, "Doctor Spectrum", "Legendary Creature — Human Hero", msc7SpectrumOracle, 3, 3)
}

func TestMsc7DoctorSpectrumFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Doctor Spectrum", "Legendary Creature — Human Hero", msc7SpectrumOracle, 3, 3)
	assertKeywords(t, g, id, "flying")
}

func TestMsc7DoctorSpectrumMakesAZeroFourWall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc7Spectrum(t, g)
	msc7PickMode(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Wall"); got != 1 {
		t.Fatalf("Walls on the battlefield: %d, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Wall" {
			continue
		}
		if !c.IsToken() {
			t.Error("the Wall is not a token")
		}
		if p, tg := effectivePower(t, g, c.InstanceID), effectiveToughness(t, g, c.InstanceID); p != 0 || tg != 4 {
			t.Errorf("the Wall is %d/%d, want 0/4", p, tg)
		}
		assertKeywords(t, g, c.InstanceID, "defender")
	}
}

func TestMsc7DoctorSpectrumCountersEachOtherHeroOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	hero := b12Creature(g, me.ID, "Hero Pal", "Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	spectrum := msc7Spectrum(t, g)
	msc7PickMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"another Hero I control", hero, 1},
		{"a non-Hero", bear, 0},
		{"an opponent's Hero", theirs, 0},
		{"Doctor Spectrum itself", spectrum, 0},
	} {
		if got := miCounters(t, g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.name, got, c.want)
		}
	}
}

func TestMsc7DoctorSpectrumDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ench := b12Permanent(g, opp.ID, "Their Pact", "Enchantment")
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	msc7Spectrum(t, g)
	msc7PickMode(t, g, me.ID, 2)
	pickCard(t, g, me.ID, ench)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("a non-enchantment was destroyed")
	}
}

func TestMsc7DoctorStrangeDoublesYourLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	strange := b12Push(g, me.ID, "Doctor Strange, Surgeon", "Legendary Creature — Human Doctor Hero", msc7StrangeOracle, 2, 5)
	assertKeywords(t, g, strange, "lifelink")
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(strange, me.ID, 3)
		_ = g.ChangePlayerLifeForEffect(strange, opp.ID, 3)
	})
	if me.Life != life+6 {
		t.Errorf("my life %d -> %d, want +6", life, me.Life)
	}
	if opp.Life != oppLife+3 {
		t.Errorf("an opponent's life %d -> %d, want +3", oppLife, opp.Life)
	}
}

func TestMsc7DoctorStrangePumpsAtCombatOnlyTenAboveStarting(t *testing.T) {
	for _, tc := range []struct {
		life int
		pump bool
	}{{game.StartingLife + 10, true}, {game.StartingLife + 9, false}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b12Push(g, me.ID, "Doctor Strange, Surgeon", "Legendary Creature — Human Doctor Hero", msc7StrangeOracle, 2, 5)
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		me.Life = tc.life

		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		wantP, wantT := 2, 2
		if tc.pump {
			wantP, wantT = 4, 4
		}
		if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != wantP || tg != wantT {
			t.Errorf("life %d: my bear is %d/%d, want %d/%d", tc.life, p, tg, wantP, wantT)
		}
		if got := hasEffectiveKeyword(t, g, bear, "vigilance"); got != tc.pump {
			t.Errorf("life %d: my bear has vigilance %v, want %v", tc.life, got, tc.pump)
		}
		if p := effectivePower(t, g, theirs); p != 2 {
			t.Errorf("life %d: their bear has power %d, want 2", tc.life, p)
		}
	}
}

func TestMsc7EerieGravestoneDrawsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two")
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Eerie Gravestone", "Artifact", msc7GravestoneOracle, nil)
	passPriorityAroundTable(t, g)
	// The cast takes the card from the hand, and the trigger draws one.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size %d, want %d", got, before+1)
	}
}

func msc7ActivateGravestone(t *testing.T, g *game.Game, me *game.Player) (stone, creature, other uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	stone = b12Push(g, me.ID, "Eerie Gravestone", "Artifact", msc7GravestoneOracle, 0, 0)
	plTop(me, "Deep Card", "Instant", "{U}")
	other = plTop(me, "Milled Land", "Basic Land — Forest", "")
	creature = plTop(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	plTop(me, "Milled Spell", "Sorcery", "{R}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	return stone, creature, other
}

func TestMsc7EerieGravestoneMillsFourAndTakesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yard := me.Graveyard.Size()
	stone, creature, land := msc7ActivateGravestone(t, g, me)
	if g.Battlefield.Contains(stone) || !me.Graveyard.Contains(stone) {
		t.Error("the Gravestone was not sacrificed")
	}
	if got := me.Graveyard.Size() - yard; got != 5 {
		t.Errorf("graveyard grew by %d, want 5 (four milled and the Gravestone)", got)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature was offered from among the milled cards")
	}
	if len(pick.ChooseCards) != 1 || !hasID(pick.ChooseCards, creature) {
		t.Errorf("offered %v, want only the milled creature", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, creature)
	if !miInHand(me, creature) {
		t.Error("the chosen creature did not reach the hand")
	}
	if miInHand(me, land) {
		t.Error("a land reached the hand")
	}
}

func TestMsc7EerieGravestoneTakingACreatureIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, creature, _ := msc7ActivateGravestone(t, g, me)
	answerChooseCards(t, g, me.ID)
	if miInHand(me, creature) || !me.Graveyard.Contains(creature) {
		t.Error("the creature left the graveyard although the player declined")
	}
}

func TestMsc7FalconAndRedwingMakesThatManyBirdsThenGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	falcon := b12Push(g, me.ID, "Falcon and Redwing", "Legendary Creature — Human Bird Hero", msc7RedwingOracle, 3, 2)
	assertKeywords(t, g, falcon, "flying")
	attackWith(t, g, opp.ID, falcon)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Bird"); got != 3 {
		t.Errorf("Bird tokens: %d, want 3 (the damage dealt)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" {
			if p, tg := effectivePower(t, g, c.InstanceID), effectiveToughness(t, g, c.InstanceID); p != 1 || tg != 1 {
				t.Errorf("a Bird is %d/%d, want 1/1", p, tg)
			}
			assertKeywords(t, g, c.InstanceID, "flying")
		}
	}
	if got := miCounters(t, g, falcon, game.CounterPlusOne); got != 1 {
		t.Errorf("Falcon and Redwing has %d +1/+1 counters, want 1", got)
	}
}

func TestMsc7FalconTorresBattalionGrowsAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	falcon := b12Push(g, me.ID, "Falcon, Joaquin Torres", "Legendary Creature — Bird Human Hero", msc7TorresOracle, 2, 2)
	assertKeywords(t, g, falcon, "flying", "lifelink")
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, falcon, a, b)
	passPriorityAroundTable(t, g)
	if got := miCounters(t, g, falcon, game.CounterPlusOne); got != 1 {
		t.Errorf("Falcon has %d +1/+1 counters, want 1", got)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the battalion trigger did not scry")
	}
}

func TestMsc7FalconTorresNeedsTwoOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	falcon := b12Push(g, me.ID, "Falcon, Joaquin Torres", "Legendary Creature — Bird Human Hero", msc7TorresOracle, 2, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, falcon, a)
	passPriorityAroundTable(t, g)
	if got := miCounters(t, g, falcon, game.CounterPlusOne); got != 0 {
		t.Errorf("Falcon has %d counters with one other attacker, want 0", got)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("battalion scried with only one other attacker")
	}
}

func TestMsc7FallToEarthExilesAndEveryPlayerGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	castCatalogSpell(t, g, "Fall to Earth", "Instant", msc7FallToEarthOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	for i, p := range g.Seats {
		if p.Life != lives[i]+3 {
			t.Errorf("seat %d life %d -> %d, want +3", i, lives[i], p.Life)
		}
	}
}

func TestMsc7FallToEarthBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Command Tower", "Land"),
	)
	id, _ := cycleFromHand(t, g, "Fall to Earth", "Instant", msc7FallToEarthOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if searchOptionNamed(g, c, "Command Tower") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land did not reach the hand")
	}
}

func TestMsc7FantasticBounceCostsTwoLessAtATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := pushPermanent(g, opp.ID, game.Card{Name: "Tapped Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	untapped := pushPermanent(g, opp.ID, game.Card{Name: "Ready Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	tappedRock := pushPermanent(g, opp.ID, game.Card{Name: "Tapped Rock", TypeLine: "Artifact", Tapped: true})
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	for _, tc := range []struct {
		name string
		to   []game.TargetRef
		want int
	}{
		{"a tapped creature", at(tapped), 2},
		{"an untapped creature", at(untapped), 4},
		{"a tapped noncreature", at(tappedRock), 4},
		{"no target", nil, 4},
	} {
		if got := selfPriced(t, g, me, msc7BounceOracle, "Sorcery", "{3}{U}", game.ZoneHand, tc.to).ManaValue(); got != tc.want {
			t.Errorf("Fantastic Bounce at %s: %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestMsc7FantasticBounceReturnsTheNonlandAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedLibrary(me, "One", "Two")
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Fantastic Bounce", "Sorcery", msc7BounceOracle, b16TargetCard(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !opp.Hand.Contains(relic) {
		t.Error("the permanent did not return to its owner's hand")
	}
	// The cast takes the card from the hand, and the spell draws one.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size %d, want %d", got, before+1)
	}
}

func TestMsc7FantasticBounceRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Fantastic Bounce", "Sorcery", msc7BounceOracle, b16TargetCard(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

func TestMsc7FirebirdPumpsOtherAttackersByItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := b12Push(g, me.ID, "Firebird, Blazing Ranger", "Legendary Creature — Human Ranger Hero", msc7FirebirdOracle, 2, 3)
	assertKeywords(t, g, bird, "flying")
	attacker := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bird, attacker)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, attacker); p != 4 {
		t.Errorf("the other attacker has power %d, want 4", p)
	}
	if p := effectivePower(t, g, bird); p != 2 {
		t.Errorf("Firebird has power %d, want 2 (it pumps only the others)", p)
	}
	if p := effectivePower(t, g, home); p != 2 {
		t.Errorf("a creature that stayed home has power %d, want 2", p)
	}
}

func msc7Everett(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Everett K. Ross, Hapless Attaché", "Legendary Creature — Human Advisor", msc7EverettOracle, 1, 3)
}

func TestMsc7EverettBuffsOnlyYourCommanderCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	everett := msc7Everett(g, me.ID)
	commander := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Commander", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID, IsCommander: true,
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tg := effectivePower(t, g, commander), effectiveToughness(t, g, commander); p != 4 || tg != 4 {
		t.Errorf("my commander is %d/%d, want 4/4", p, tg)
	}
	if !hasEffectiveKeyword(t, g, commander, "lifelink") {
		t.Error("my commander lacks lifelink")
	}
	for name, id := range map[string]uuid.UUID{"their commander": theirs, "a non-commander": bear, "Everett": everett} {
		if p := effectivePower(t, g, id); p == 4 || (name == "Everett" && p != 1) {
			t.Errorf("%s got the bonus (power %d)", name, p)
		}
		if name != "Everett" && hasEffectiveKeyword(t, g, id, "lifelink") {
			t.Errorf("%s has lifelink", name)
		}
	}
}

func TestMsc7EverettDrawsWhenAnOpponentAttacksYouWithTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	msc7Everett(g, me.ID)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	declareAttack(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestMsc7EverettDoesNotDrawForOneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	msc7Everett(g, me.ID)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	declareAttack(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d cards for a lone attacker, want 0", got)
	}
}

func TestMsc7EverettDoesNotDrawWhenTheAttackIsAtAnotherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[1], g.Seats[0], g.Seats[2]
	msc7Everett(g, me.ID)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	declareAttack(t, g, third.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d cards for an attack at someone else, want 0", got)
	}
}

func msc7Commander(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner, IsCommander: true,
	})
}

func msc7CastThrone(t *testing.T, g *game.Game, mine, theirs uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Fight for the Throne", "Instant", msc7ThroneOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
}

func TestMsc7FightForTheThroneCountersFightsAndCrownsYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc7Commander(g, me.ID)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 3)
	msc7CastThrone(t, g, mine, theirs)
	if got := miCounters(t, g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my creature has %d +1/+1 counters, want 1", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Fatal("their creature survived a 3-power fight")
	}
	if g.Monarch != me.ID {
		t.Error("the controller did not become the monarch")
	}
}

func TestMsc7FightForTheThroneNeedsYourCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 3)
	msc7CastThrone(t, g, mine, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("their creature survived a 3-power fight")
	}
	if g.Monarch == me.ID {
		t.Error("became the monarch without controlling a commander")
	}
}

func TestMsc7FightForTheThroneWatchesTheCreatureAllTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc7Commander(g, me.ID)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	msc7CastThrone(t, g, mine, theirs)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("their creature died in the fight")
	}
	if g.Monarch == me.ID {
		t.Fatal("became the monarch before the creature died")
	}
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Error("the creature died this turn with my commander out, and I am not the monarch")
	}
}
