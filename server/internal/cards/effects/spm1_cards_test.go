package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm1_cards_test.go — Marvel's Spider-Man, batch 1 (#4460). Each test plays
// the real catalog entry and asserts what its text says.

const (
	spm1AcrobaticsOracle = "0466f23c-d07c-4015-82d5-eab4b226f1ff"
	spm1RabbleOracle     = "cf6dc79e-840c-4d47-8f94-c8623528acd2"
	spm1BagelOracle      = "1389c9fa-fec2-4212-a7aa-6dfb318dc202"
	spm1BeetleOracle     = "87972de1-c1d7-408e-8d56-995590f639b8"
	spm1CarapaceOracle   = "7d27c583-469c-42d7-a547-8c65654e9d95"
	spm1BlackCatOracle   = "f5cdb413-2b72-4dc6-b93d-2532ac2ae568"
	spm1CarnageOracle    = "c0efafc6-8ffb-4908-ae64-3f391722370a"
	spm1CrowdOracle      = "e11839d5-99fe-4cc8-802e-be669ad274ee"
	spm1CosmicOracle     = "0334a2f8-a74a-44c9-ba64-d9b7fabe160c"
	spm1BugleOracle      = "483e0c6c-8131-486c-b482-cc3396c9786b"
)

func TestSpm1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spm1AcrobaticsOracle: "Amazing Acrobatics",
		spm1RabbleOracle:     "Angry Rabble",
		spm1BagelOracle:      "Bagel and Schmear",
		spm1BeetleOracle:     "Beetle, Legacy Criminal",
		spm1CarapaceOracle:   "Biorganic Carapace",
		spm1BlackCatOracle:   "Black Cat, Cunning Thief",
		spm1CarnageOracle:    "Carnage, Crimson Chaos",
		spm1CrowdOracle:      "Cheering Crowd",
		spm1CosmicOracle:     "Cosmic Spider-Man",
		spm1BugleOracle:      "Daily Bugle Building",
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

// --- Amazing Acrobatics ---

// spm1OpponentBolt has the opponent cast a Lightning Bolt at me and returns
// its stack id.
func spm1OpponentBolt(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("opponent's Bolt: %v", err)
	}
	return id
}

func TestSpm1AcrobaticsCountersTheSpellAndTapsTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	bolt := spm1OpponentBolt(t, g)
	life := me.Life
	b02cCastModalSpell(t, g, "Amazing Acrobatics", "Instant", spm1AcrobaticsOracle, []int{0, 1},
		[]game.TargetRef{p1g3ModeCard(bolt, 0), p1g3ModeCard(a, 1), p1g3ModeCard(b, 1)})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Errorf("Bolt countered = %v, life %d (was %d)", opp.Graveyard.Contains(bolt), me.Life, life)
	}
	if !isTapped(g, a) || !isTapped(g, b) || isTapped(g, c) {
		t.Errorf("tapped: A %v, B %v, C %v; want true, true, false", isTapped(g, a), isTapped(g, b), isTapped(g, c))
	}
}

func TestSpm1AcrobaticsEitherBulletAloneAndTheLimits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Amazing Acrobatics", TypeLine: "Instant",
		OracleID: spm1AcrobaticsOracle, Owner: me.ID, Controller: me.ID})
	// Three creatures is more than "one or two".
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{p1g3ModeCard(a, 0), p1g3ModeCard(b, 0), p1g3ModeCard(c, 0)}})
	if err == nil {
		t.Fatal("tapping three creatures was accepted")
	}
	// No bullet at all is refused too.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{}}); err == nil {
		t.Fatal("a cast with no bullet chosen was accepted")
	}
	// The tap bullet alone, on one creature.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{p1g3ModeCard(a, 0)}}); err != nil {
		t.Fatalf("tap one creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, a) || isTapped(g, b) {
		t.Errorf("tapped: A %v, B %v; want true, false", isTapped(g, a), isTapped(g, b))
	}
}

// --- Angry Rabble ---

func TestSpm1RabbleDamagesEachOpponentOnlyForBigSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rabble := b27Push(g, me.ID, "Angry Rabble", "Creature — Human Citizen", spm1RabbleOracle, "{1}{R}", 2, 2, "R")
	assertKeywords(t, g, rabble, "trample")
	life, oppLife := me.Life, opp.Life
	mscM2CastArtifact(t, g, "{3}", "{C}{C}{C}")
	if opp.Life != oppLife {
		t.Fatalf("a mana value 3 spell hurt the opponent: %d", opp.Life)
	}
	mscM2CastArtifact(t, g, "{4}", "{C}{C}{C}{C}")
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d after a mana value 4 spell, want %d", opp.Life, oppLife-1)
	}
	if me.Life != life {
		t.Errorf("my life changed: %d", me.Life)
	}
}

func TestSpm1RabbleGrowsAtSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rabble := b27Push(g, me.ID, "Angry Rabble", "Creature — Human Citizen", spm1RabbleOracle, "{1}{R}", 2, 2, "R")
	advanceToMain(t, g)
	floatMana(t, g, me, "{R}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rabble, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, rabble); got != 2 {
		t.Fatalf("counters %d, want 2", got)
	}
	// With the stack busy it is not a sorcery-speed moment.
	spm1OpponentBolt(t, g)
	floatMana(t, g, me, "{R}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rabble, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the pump was activated with a spell on the stack")
	}
}

// --- Bagel and Schmear ---

func TestSpm1BagelShareCountersAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bagel := pushCatalogPermanent(g, me.ID, "Bagel and Schmear", "Artifact — Food", spm1BagelOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	floatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, bagel, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("Share: %v", err)
	}
	if g.Battlefield.Contains(bagel) {
		t.Error("the Bagel was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 || me.Hand.Size()-hand != 1 {
		t.Errorf("counters %d, cards drawn %d; want 1, 1", plusOnes(g, bear), me.Hand.Size()-hand)
	}
}

func TestSpm1BagelShareWithNoTargetStillDrawsAndNoshGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bagel := pushCatalogPermanent(g, me.ID, "Bagel and Schmear", "Artifact — Food", spm1BagelOracle, false)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	floatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, bagel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Share with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}

	nosh := pushCatalogPermanent(g, me.ID, "Bagel and Schmear", "Artifact — Food", spm1BagelOracle, false)
	life, hand := me.Life, me.Hand.Size()
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, nosh, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Nosh: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 || me.Hand.Size()-hand != 1 {
		t.Errorf("life %+d, drew %d; want +3, 1", me.Life-life, me.Hand.Size()-hand)
	}
}

func TestSpm1BagelShareIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bagel := pushCatalogPermanent(g, me.ID, "Bagel and Schmear", "Artifact — Food", spm1BagelOracle, false)
	spm1OpponentBolt(t, g)
	floatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, bagel, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("Share was activated with a spell on the stack")
	}
}

// --- Beetle, Legacy Criminal ---

func TestSpm1BeetleExilesItselfFromTheGraveyardToGrowAndFly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	beetle := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: beetle, Name: "Beetle, Legacy Criminal", OracleID: spm1BeetleOracle,
		TypeLine: "Legendary Creature — Human Rogue Villain", ManaCost: "{3}{U}", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{U}{C}")
	if err := g.ActivateCatalogAbility(me.ID, beetle, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if got := b12ZoneOf(g, beetle); got != game.ZoneExile {
		t.Fatalf("Beetle is in %q, want exile as a cost", got)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 || !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Errorf("counters %d, flying %v; want 1, true", plusOnes(g, bear), hasEffectiveKeyword(t, g, bear, "flying"))
	}
}

func TestSpm1BeetleIsAFlyerAndOnlyActivatesFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beetle := b27Push(g, me.ID, "Beetle, Legacy Criminal", "Legendary Creature — Human Rogue Villain", spm1BeetleOracle, "{3}{U}", 3, 3, "U")
	assertKeywords(t, g, beetle, "flying")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{U}{C}")
	if err := g.ActivateCatalogAbility(me.ID, beetle, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Biorganic Carapace ---

func TestSpm1CarapaceAttachesPumpsAndDrawsPerModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, grown).Counters = map[string]int{game.CounterPlusOne: 1}
		findBattlefieldCardForTest(g, theirs).Counters = map[string]int{game.CounterPlusOne: 1}
	})

	id := castCatalogSpell(t, g, "Biorganic Carapace", "Artifact — Equipment", spm1CarapaceOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, id); host.ID != bear {
		t.Fatalf("the Carapace did not attach on entry: %+v", host)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	if got := effectivePower(t, g, plain); got != 2 {
		t.Errorf("unequipped power %d, want 2", got)
	}

	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear, plain)
	passPriorityAroundTable(t, g)
	// Modified creatures I control: the equipped bear and the countered one,
	// not the plain one and not the opponent's countered creature.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (equipped bear + countered bear)", got)
	}
}

func TestSpm1CarapaceDrawsNothingWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	castCatalogSpell(t, g, "Biorganic Carapace", "Artifact — Equipment", spm1CarapaceOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d off a blocked attacker, want 0", got)
	}
}

// --- Black Cat, Cunning Thief ---

// spm1BlackCatTable casts Black Cat with the opponent's library stacked with
// `n` cards (top card last pushed) and points the trigger at the opponent.
func spm1BlackCatTable(t *testing.T, g *game.Game, n int) (me, opp *game.Player, lib []uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	opp.Library.Cards = nil
	for i := 1; i <= n; i++ {
		c := game.NewCard("Library Card "+string(rune('A'+i-1)), opp.ID)
		c.TypeLine = "Creature — Test"
		c.ManaCost = "{1}"
		opp.Library.PushTop(c)
		lib = append([]uuid.UUID{c.InstanceID}, lib...)
	}
	castHoldCreature(t, g, "Black Cat, Cunning Thief", "Legendary Creature — Human Rogue Villain", spm1BlackCatOracle, 2, 3)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	return me, opp, lib
}

func TestSpm1BlackCatExilesTwoOfTheTopNineFaceDownAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, lib := spm1BlackCatTable(t, g, 12)
	// lib[0] is the top card. Nine are looked at; the tenth stays on top.
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("want a choose-exactly-two prompt, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Fatal("one card was accepted for an exile-two prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[0], lib[9]}); err == nil {
		t.Fatal("a card below the top nine was accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[2], lib[5]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	for _, id := range []uuid.UUID{lib[2], lib[5]} {
		perm := e2Permission(g, id)
		if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
			t.Errorf("card %s permission %+v, want an any-type play grant to Black Cat's controller", id, perm)
		}
		e2AssertFaceDownFor(t, g, id, me.ID)
		if got := b12ZoneOf(g, id); got != game.ZoneExile {
			t.Errorf("card %s is in %q, want exile", id, got)
		}
	}
	if len(opp.Library.Cards) != 10 {
		t.Fatalf("library has %d cards, want 12 - 2 = 10", len(opp.Library.Cards))
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; top != lib[9] {
		t.Errorf("the top of the library is %s, want the tenth card (not looked at)", top)
	}
	// The seven others from the nine are the bottom seven, in some order.
	bottom := map[uuid.UUID]bool{}
	for _, c := range opp.Library.Cards[:7] {
		bottom[c.InstanceID] = true
	}
	for _, i := range []int{0, 1, 3, 4, 6, 7, 8} {
		if !bottom[lib[i]] {
			t.Errorf("looked-at card %d is not among the bottom seven", i)
		}
	}
}

func TestSpm1BlackCatSmallLibraryExilesWhatThereIs(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, lib := spm1BlackCatTable(t, g, 1)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("a one-card library asked for a choice")
	}
	if got := b12ZoneOf(g, lib[0]); got != game.ZoneExile {
		t.Errorf("the only card is in %q, want exile", got)
	}
	if len(opp.Library.Cards) != 0 {
		t.Errorf("library has %d cards, want 0", len(opp.Library.Cards))
	}
}

func TestSpm1BlackCatCardsAreCastableWithAnyTypeOfMana(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, lib := spm1BlackCatTable(t, g, 3)
	// Three cards or fewer than nine: still a choice of two.
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choose prompt for a three-card library")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{lib[0], lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := b12ZoneOf(g, lib[2]); got != game.ZoneLibrary || len(opp.Library.Cards) != 1 {
		t.Errorf("the unpicked card is in %q (library %d), want the library", got, len(opp.Library.Cards))
	}
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, lib[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the stolen {1} card with green mana: %v", err)
	}
	// The other player has no permission over the second card.
	if err := g.CastSpell(opp.ID, lib[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the opponent cast a card exiled by Black Cat")
	}
}

// --- Carnage, Crimson Chaos ---

func TestSpm1CarnageReturnsACheapCreatureWithTheTwoGrantedAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cheap := uuid.New()
	pricey := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Dead Giant", TypeLine: "Creature — Giant", ManaCost: "{4}",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	castHoldCreature(t, g, "Carnage, Crimson Chaos", "Legendary Creature — Symbiote Villain", spm1CarnageOracle, 4, 3)
	ch := latestPickTarget(g, me.ID)
	if ch == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if err := g.ResolvePickTarget(ch.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: pricey}); err == nil {
		t.Fatal("a mana value 4 creature card was accepted")
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || !me.Graveyard.Contains(pricey) {
		t.Fatalf("bear on the battlefield = %v, giant still in the graveyard = %v", g.Battlefield.Contains(cheap), me.Graveyard.Contains(pricey))
	}

	// It must attack each combat if able, and is sacrificed when it connects.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, cheap).SummonedThisTurn = false })
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.PassPriority()
	if err == nil || !strings.Contains(err.Error(), "must attack") {
		t.Fatalf("declaring no attackers: err = %v, want the must-attack refusal", err)
	}
	if err := g.DeclareAttacker(cheap, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the returned creature survived dealing combat damage to a player")
	}
	if !me.Graveyard.Contains(cheap) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

func TestSpm1CarnageIsATrampler(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	carnage := b27Push(g, me.ID, "Carnage, Crimson Chaos", "Legendary Creature — Symbiote Villain", spm1CarnageOracle, "{2}{B}{R}", 4, 3, "B", "R")
	assertKeywords(t, g, carnage, "trample")
	if spec, _ := Lookup(spm1CarnageOracle); len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].ManaCost != "{B}{R}" {
		t.Errorf("alternative costs %+v, want mayhem {B}{R}", spec.AlternativeCosts)
	}
}

func TestSpm1CarnageCastsFromTheGraveyardForMayhemOnlyAfterADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	carnage := mayhemHandCard(g, me, "Carnage, Crimson Chaos", "Legendary Creature — Symbiote Villain", "{2}{B}{R}", spm1CarnageOracle)
	mayhemDiscard(t, g, me, carnage)
	floatForTest(g, me, "B")
	if err := g.CastSpell(me.ID, carnage, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err == nil {
		t.Fatal("mayhem {B}{R} paid with one mana")
	}
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, carnage, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "mayhem"}); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(carnage) {
		t.Error("Carnage did not resolve from the graveyard")
	}
}

// --- Cheering Crowd ---

func TestSpm1CrowdGrowsAndMakesManaForTheActivePlayerOnEveryonesTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crowd := b27Push(g, me.ID, "Cheering Crowd", "Creature — Human Citizen", spm1CrowdOracle, "{1}{R/G}", 2, 2, "R", "G")
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, crowd); got != 1 {
		t.Fatalf("counters %d, want 1", got)
	}
	if got := mcPoolCount(me, "C"); got != 1 {
		t.Errorf("I have %d {C}, want 1", got)
	}
}

func TestSpm1CrowdDeclinedGivesNothingAndCountsEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crowd := b27Push(g, me.ID, "Cheering Crowd", "Creature — Human Citizen", spm1CrowdOracle, "{1}{R/G}", 2, 2, "R", "G")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, crowd).Counters = map[string]int{game.CounterPlusOne: 2, "charge": 1}
	})
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if plusOnes(g, crowd) != 2 || mcPoolCount(me, "C") != 0 {
		t.Fatalf("declined: counters %d, {C} %d; want 2, 0", plusOnes(g, crowd), mcPoolCount(me, "C"))
	}
}

func TestSpm1CrowdMakesAllCountersWorthOfManaForAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	crowd := b27Push(g, opp.ID, "Cheering Crowd", "Creature — Human Citizen", spm1CrowdOracle, "{1}{R/G}", 2, 2, "R", "G")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, crowd).Counters = map[string]int{game.CounterPlusOne: 2, "charge": 1}
	})
	// The Crowd belongs to the opponent; it is MY first main phase.
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, crowd); got != 3 {
		t.Fatalf("counters %d, want 3", got)
	}
	if got := mcPoolCount(me, "C"); got != 4 {
		t.Errorf("I have %d {C}, want 4 (three +1/+1 and a charge counter)", got)
	}
	if got := mcPoolCount(opp, "C"); got != 0 {
		t.Errorf("the Crowd's controller got %d {C}, want 0", got)
	}
}

// --- Cosmic Spider-Man ---

func TestSpm1CosmicGivesOtherSpidersEverythingAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cosmic := b27Push(g, me.ID, "Cosmic Spider-Man", "Legendary Creature — Spider Human Hero", spm1CosmicOracle, "{W}{U}{B}{R}{G}", 5, 5, "W", "U", "B", "R", "G")
	assertKeywords(t, g, cosmic, "flying", "first strike", "trample", "lifelink", "haste")
	spider := b12Creature(g, me.ID, "Little Spider", "Creature — Spider", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Spider", "Creature — Spider", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "first strike", "trample", "lifelink", "haste"} {
		if !hasEffectiveKeyword(t, g, spider, kw) {
			t.Errorf("my other Spider lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) || hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a creature that is not my Spider has %s", kw)
		}
	}
}

func TestSpm1CosmicDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	b27Push(g, opp.ID, "Cosmic Spider-Man", "Legendary Creature — Spider Human Hero", spm1CosmicOracle, "{W}{U}{B}{R}{G}", 5, 5, "W", "U", "B", "R", "G")
	spider := b12Creature(g, opp.ID, "Their Spider", "Creature — Spider", 1, 1)
	_ = me
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, spider, "flying") {
		t.Error("an opponent's Spider got flying on my turn's combat")
	}
}

// --- Daily Bugle Building ---

func TestSpm1BugleMakesColorlessAndFiltersForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Daily Bugle Building", "Land", spm1BugleOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if mcPoolCount(me, "C") != 1 {
		t.Fatalf("pool %+v, want {C}", me.ManaPool.Units())
	}
	land2 := pushCatalogPermanent(g, me.ID, "Daily Bugle Building", "Land", spm1BugleOracle, false)
	if err := g.ActivateManaAbility(me.ID, land2, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1},{T}: any color: %v", err)
	}
	if mcPoolCount(me, "U") != 1 || mcPoolCount(me, "C") != 0 {
		t.Errorf("pool %+v, want one {U} for the {C}", me.ManaPool.Units())
	}
}

func TestSpm1BugleSmearCampaignGivesALegendaryCreatureMenaceAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Daily Bugle Building", "Land", spm1BugleOracle, false)
	hero := b12Creature(g, me.ID, "Hero", "Legendary Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a nonlegendary creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}},
	}); err != nil {
		t.Fatalf("Smear Campaign: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, hero, "menace") {
		t.Error("the legendary creature did not gain menace")
	}
	land2 := pushCatalogPermanent(g, me.ID, "Daily Bugle Building", "Land", spm1BugleOracle, false)
	spm1OpponentBolt(t, g)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, land2, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hero}},
	}); err == nil {
		t.Error("Smear Campaign was activated with a spell on the stack")
	}
}
