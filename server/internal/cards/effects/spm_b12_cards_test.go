package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b12_cards_test.go — Marvel's Spider-Man, batch 12 (#4460). Each test
// plays the real catalog entry and asserts what its text says.

const (
	spmB12GreenGoblinOracle  = "d1f1e48f-f1f4-497f-9112-79159d2f443b"
	spmB12GrendelOracle      = "a93f0735-9854-487d-b646-d55e2bee00c0"
	spmB12LethalOracle       = "49a7c053-81e9-47b7-af5c-c62cd0389ed4"
	spmB12LylaOracle         = "b7fb0141-4fbd-4269-a2e9-92f75ec95271"
	spmB12ProwlerOracle      = "8a56c2a7-7131-4495-aa27-52f48e5f1fb6"
	spmB12PumpkinOracle      = "b0695221-a61a-41e4-8195-3eeaaab2d969"
	spmB12ClassmateOracle    = "bf3ff6ff-eb0e-4d0a-99cb-c377622ae640"
	spmB12SensationalOracle  = "8c489e39-b9ff-400f-a855-3cddf7e112e6"
	spmB12MiguelOracle       = "c0280d66-699f-4139-828b-a0f79ae2d20d"
	spmB12MilesOracle        = "229ecaa1-b3c9-44a1-8e7a-857ea5b8b0df"
	spmB12GreenGoblinTypes   = "Legendary Creature — Goblin Human Villain"
	spmB12SpiderHeroTypeLine = "Legendary Creature — Spider Human Hero"
)

func TestSpmB12EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB12GreenGoblinOracle: "Green Goblin, Nemesis",
		spmB12GrendelOracle:     "Grendel, Spawn of Knull",
		spmB12LethalOracle:      "Lethal Protection",
		spmB12LylaOracle:        "Lyla, Holographic Assistant",
		spmB12ProwlerOracle:     "Prowler, Misguided Mentor",
		spmB12PumpkinOracle:     "Pumpkin Bombs",
		spmB12ClassmateOracle:   "Rampaging Classmate",
		spmB12SensationalOracle: "Sensational Spider-Man",
		spmB12MiguelOracle:      "Spider-Man 2099, Miguel O'Hara",
		spmB12MilesOracle:       "Spider-Man, Miles Morales",
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

// spmB12PickIfAsked answers an open pick_target prompt with a card, if one
// is open, then settles the stack.
func spmB12PickIfAsked(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, chooser) != nil {
		pickCard(t, g, chooser, target)
		passPriorityAroundTable(t, g)
	}
}

func spmB12Counters(g *game.Game, id uuid.UUID, kind string) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return -1
	}
	return c.Counters[kind]
}

func spmB12Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Grendel, Spawn of Knull ---

func TestSpmB12GrendelHasFlyingAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Grendel, Spawn of Knull", "Legendary Creature — Symbiote Dragon Villain", spmB12GrendelOracle, 3, 3)
	assertKeywords(t, g, id, "flying", "deathtouch")
}

// --- Rampaging Classmate ---

func TestSpmB12ClassmateGetsPlusOnePerOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cm := b12Push(g, me.ID, "Rampaging Classmate", "Creature — Lizard Berserker", spmB12ClassmateOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, cm, a, b)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, cm); got != 5 {
		t.Errorf("Classmate power %d, want 3 + 2 other attackers", got)
	}
	if got := hobA4Power(t, g, idle); got != 0 {
		t.Errorf("a creature that did not attack changed power: %d", got)
	}
}

func TestSpmB12ClassmateAttackingAloneStaysThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cm := b12Push(g, me.ID, "Rampaging Classmate", "Creature — Lizard Berserker", spmB12ClassmateOracle, 3, 2)
	declareAttack(t, g, opp.ID, cm)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, cm); got != 3 {
		t.Errorf("Classmate power %d, want 3", got)
	}
}

// --- Lyla, Holographic Assistant ---

func TestSpmB12LylaPutsACounterOnTargetCreatureForEachCardYouDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lyla, Holographic Assistant", "Legendary Artifact Creature — Illusion Advisor", spmB12LylaOracle, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B", "C")
	hobA5Draw(t, g, me, &game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if got := spmB12Counters(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("their creature has %d +1/+1 counters, want 1 (any creature is a legal target)", got)
	}
	hobA5Draw(t, g, me, &game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if got := spmB12Counters(g, theirs, game.CounterPlusOne); got != 2 {
		t.Errorf("after a second draw: %d counters, want 2 (one trigger per card)", got)
	}
}

func TestSpmB12LylaIgnoresAnOpponentsDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lyla := b12Push(g, me.ID, "Lyla, Holographic Assistant", "Legendary Artifact Creature — Illusion Advisor", spmB12LylaOracle, 2, 2)
	seedLibrary(opp, "A", "B")
	hobA5Draw(t, g, opp, nil)
	if got := spmB12Counters(g, lyla, game.CounterPlusOne); got != 0 {
		t.Errorf("Lyla got %d counters from an opponent's draw", got)
	}
}

// --- Prowler, Misguided Mentor ---

func TestSpmB12ProwlerCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	for _, tc := range []struct {
		power   int
		refused bool
	}{{2, true}, {3, false}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		prowler := b12Push(g, me.ID, "Prowler, Misguided Mentor", "Legendary Creature — Human Rogue Villain", spmB12ProwlerOracle, 3, 3)
		blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", tc.power, 4)
		got := msc13BlockRefused(t, g, opp.ID, blocker, prowler)[prowler]
		if got != tc.refused {
			t.Errorf("blocker with power %d refused = %v, want %v", tc.power, got, tc.refused)
		}
	}
}

func TestSpmB12ProwlerCombatDamagePutsACounterOnAnotherCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	prowler := b12Push(g, me.ID, "Prowler, Misguided Mentor", "Legendary Creature — Human Rogue Villain", spmB12ProwlerOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	attackWith(t, g, opp.ID, prowler)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: prowler}); err == nil {
			t.Error("Prowler was allowed to target itself")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
			t.Error("Prowler was allowed to target an opponent's creature")
		}
	}
	spmB12PickIfAsked(t, g, me.ID, bear)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-3)
	}
	if got := spmB12Counters(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters, want 1", got)
	}
	if got := spmB12Counters(g, prowler, game.CounterPlusOne); got != 0 {
		t.Errorf("Prowler put a counter on itself (%d)", got)
	}
}

// --- Spider-Man, Miles Morales ---

func TestSpmB12MilesEntersPutsCountersOnEachOtherCreatureAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Ox", "Creature — Ox", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	miles, err := spmB4Cast(t, g, "Spider-Man, Miles Morales", spmB12SpiderHeroTypeLine, spmB12MilesOracle, "", 5, 5, 0)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := spmB12Counters(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("creature %s has %d counters, want 1", id, got)
		}
		if !hasKeywordOnBattlefield(t, g, id, "trample") {
			t.Errorf("creature %s did not gain trample", id)
		}
	}
	if got := spmB12Counters(g, miles, game.CounterPlusOne); got != 0 {
		t.Errorf("Miles put a counter on himself (%d)", got)
	}
	if got := spmB12Counters(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if hasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
	assertKeywords(t, g, miles, "vigilance", "trample")
}

func TestSpmB12MilesAttackingDoesItAgainAndTrampleEnds(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	miles := b12Push(g, me.ID, "Spider-Man, Miles Morales", spmB12SpiderHeroTypeLine, spmB12MilesOracle, 5, 5)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, miles)
	passPriorityAroundTable(t, g)
	if got := spmB12Counters(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("the other creature has %d counters after Miles attacks, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, a, "trample") {
		t.Error("the other creature did not gain trample")
	}
	advanceToEndStep(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	passPriorityAroundTable(t, g)
	advanceToMainOf(t, g, 1)
	if hasKeywordOnBattlefield(t, g, a, "trample") {
		t.Error("trample outlasted the turn")
	}
	if got := spmB12Counters(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter did not stay (%d)", got)
	}
}

// --- Spider-Man 2099, Miguel O'Hara ---

func TestSpmB12MiguelEntersBouncesUpToOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := opp.Hand.Size()
	if _, err := spmB4Cast(t, g, "Spider-Man 2099, Miguel O'Hara", spmB12SpiderHeroTypeLine, spmB12MiguelOracle, "", 4, 4, 0); err != nil {
		t.Fatalf("cast: %v", err)
	}
	spmB12PickIfAsked(t, g, me.ID, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Error("their creature is still on the battlefield")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("their hand %d -> %d, want the creature back in it", hand, opp.Hand.Size())
	}
}

func TestSpmB12MiguelDrawsOnceForCreaturesDealingCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	miguel := b12Push(g, me.ID, "Spider-Man 2099, Miguel O'Hara", spmB12SpiderHeroTypeLine, spmB12MiguelOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, miguel, bear)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card for two creatures connecting", hand, me.Hand.Size())
	}
}

func TestSpmB12MiguelDoesNotDrawWhenOnlyAnOpponentsCreatureConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	b12Push(g, me.ID, "Spider-Man 2099, Miguel O'Hara", spmB12SpiderHeroTypeLine, spmB12MiguelOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	attackWith(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: an opponent's creature hitting me drew a card", hand, me.Hand.Size())
	}
}

// --- Green Goblin, Nemesis ---

func TestSpmB12GreenGoblinDiscardingANonlandCardGrowsAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gg := b12Push(g, me.ID, "Green Goblin, Nemesis", spmB12GreenGoblinTypes, spmB12GreenGoblinOracle, 3, 3)
	assertKeywords(t, g, gg, "flying")
	other := b12Creature(g, me.ID, "Hobgoblin", "Creature — Goblin Rogue", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := spmB12Hand(me, "Shock", "Instant")
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	answerDiscard(t, g, me.ID, spell)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
			t.Error("a non-Goblin was a legal target")
		}
	}
	spmB12PickIfAsked(t, g, me.ID, other)
	if got := spmB12Counters(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the other Goblin has %d counters, want 1", got)
	}
	if got := spmB12Counters(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("a non-Goblin got a counter")
	}
}

func TestSpmB12GreenGoblinDiscardingALandMakesATappedTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gg := b12Push(g, me.ID, "Green Goblin, Nemesis", spmB12GreenGoblinTypes, spmB12GreenGoblinOracle, 3, 3)
	land := spmB12Hand(me, "Forest", "Basic Land — Forest")
	before := onBattlefieldNamed(g, "Treasure")
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	answerDiscard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if got := onBattlefieldNamed(g, "Treasure"); got != before+1 {
		t.Fatalf("Treasures %d -> %d, want one more", before, got)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Treasure" && c.Controller == me.ID && !c.Tapped {
			t.Error("the Treasure entered untapped")
		}
	}
	if got := spmB12Counters(g, gg, game.CounterPlusOne); got != 0 {
		t.Errorf("a land discard grew the Goblin (%d)", got)
	}
}

func TestSpmB12GreenGoblinIgnoresAnOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Green Goblin, Nemesis", spmB12GreenGoblinTypes, spmB12GreenGoblinOracle, 3, 3)
	land := spmB12Hand(opp, "Forest", "Basic Land — Forest")
	before := onBattlefieldNamed(g, "Treasure")
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(opp.ID, 1) })
	answerDiscard(t, g, opp.ID, land)
	passPriorityAroundTable(t, g)
	if got := onBattlefieldNamed(g, "Treasure"); got != before {
		t.Errorf("an opponent's discard made a Treasure")
	}
}

// --- Lethal Protection ---

func spmB12CastLethal(t *testing.T, g *game.Game, victim, back uuid.UUID, strict bool) error {
	t.Helper()
	me := g.Seats[0]
	id := spmB12Hand(me, "Lethal Protection", "Sorcery")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].OracleID = spmB12LethalOracle
			}
		}
	})
	advanceToMain(t, g)
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: victim, Slot: 0}}
	if back != uuid.Nil {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: back, Slot: 1})
	}
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: refs, Strict: strict})
}

func TestSpmB12LethalProtectionDestroysAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	back := pushGraveyardPermanent(me, "Fallen", "Creature — Elf", "{1}{G}")
	if err := spmB12CastLethal(t, g, victim, back, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size() // after the spell left the hand
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target creature survived")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the creature card back", hand, me.Hand.Size())
	}
	if me.Graveyard.Contains(back) {
		t.Error("the card is still in the graveyard")
	}
}

func TestSpmB12LethalProtectionReturnIsOptional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pushGraveyardPermanent(me, "Fallen", "Creature — Elf", "{1}{G}")
	if err := spmB12CastLethal(t, g, victim, uuid.Nil, true); err != nil {
		t.Fatalf("cast with only the first target: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target creature survived")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d with no return target", hand, me.Hand.Size())
	}
}

func TestSpmB12LethalProtectionRefusesANonCreatureCardInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := pushGraveyardPermanent(me, "Rock", "Artifact", "{1}")
	if err := spmB12CastLethal(t, g, victim, rock, true); err == nil {
		t.Error("returned a noncreature card")
	}
}

func TestSpmB12LethalProtectionStillReturnsWhenTheVictimIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	back := pushGraveyardPermanent(me, "Fallen", "Creature — Elf", "{1}{G}")
	if err := spmB12CastLethal(t, g, victim, back, true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d: the return should still happen (CR 608.2b)", hand, me.Hand.Size())
	}
}

// --- Sensational Spider-Man ---

func spmB12ChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			if err := g.ResolveChooseCards(c.ID, chooser, ids); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			return
		}
	}
	t.Fatalf("no choose_cards prompt for %s", chooser)
}

func spmB12HasChooseCards(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return true
		}
	}
	return false
}

func spmB12Attack(t *testing.T, g *game.Game, theirs uuid.UUID) (uuid.UUID, *game.Player, *game.Player) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sp := b12Push(g, me.ID, "Sensational Spider-Man", spmB12SpiderHeroTypeLine, spmB12SensationalOracle, 3, 3)
	declareAttack(t, g, opp.ID, sp)
	if theirs != uuid.Nil {
		spmB12PickIfAsked(t, g, me.ID, theirs)
	}
	return sp, me, opp
}

func TestSpmB12SensationalTapsAndStunsThenYouMayDrawByRemovingStun(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == other {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterStun: 2}
			}
		}
	})
	seedLibrary(g.Seats[0], "A", "B", "C", "D", "E")
	_, me, _ := spmB12Attack(t, g, theirs)
	hand := me.Hand.Size()
	if !spmB12HasChooseCards(g, me.ID) {
		t.Fatal("the removal was not offered")
	}
	// Take the new stun counter, then both on the other creature; a fourth round is not offered.
	spmB12ChooseCards(t, g, me.ID, theirs)
	spmB12ChooseCards(t, g, me.ID, other)
	spmB12ChooseCards(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if spmB12HasChooseCards(g, me.ID) {
		t.Fatal("a fourth removal was offered")
	}
	c, _ := g.LookupCardForEffect(theirs)
	if !c.Tapped {
		t.Error("the target was not tapped")
	}
	if got := spmB12Counters(g, theirs, game.CounterStun); got != 0 {
		t.Errorf("target stun counters %d, want 0 after removing it", got)
	}
	if got := spmB12Counters(g, other, game.CounterStun); got != 0 {
		t.Errorf("other stun counters %d, want 0", got)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want three cards drawn", hand, me.Hand.Size())
	}
}

func TestSpmB12SensationalDecliningKeepsTheStunAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(g.Seats[0], "A", "B")
	_, me, _ := spmB12Attack(t, g, theirs)
	hand := me.Hand.Size()
	spmB12ChooseCards(t, g, me.ID) // none
	passPriorityAroundTable(t, g)
	if got := spmB12Counters(g, theirs, game.CounterStun); got != 1 {
		t.Errorf("stun counters %d, want 1", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
}

func TestSpmB12SensationalDrawsOnlyForWhatWasRemoved(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	seedLibrary(g.Seats[0], "A", "B", "C")
	_, me, _ := spmB12Attack(t, g, theirs)
	hand := me.Hand.Size()
	spmB12ChooseCards(t, g, me.ID, theirs) // remove the only counter
	passPriorityAroundTable(t, g)
	if spmB12HasChooseCards(g, me.ID) {
		t.Fatal("offered a round with no stun counter on the board")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, me.Hand.Size())
	}
}

// --- Pumpkin Bombs ---

func TestSpmB12PumpkinBombsDrawsFuseDamagesAndGivesAwayTheArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bombs := pushCatalogPermanent(g, me.ID, "Pumpkin Bombs", "Artifact", spmB12PumpkinOracle, false)
	d1 := spmB12Hand(me, "One", "Instant")
	d2 := spmB12Hand(me, "Two", "Instant")
	seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	life := opp.Life
	err := g.ActivateCatalogAbility(me.ID, bombs, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-2+3 {
		t.Errorf("hand %d -> %d, want -2 +3", hand, me.Hand.Size())
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d (one fuse counter)", opp.Life, life-1)
	}
	c, ok := g.LookupCardForEffect(bombs)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("the opponent did not gain control: %+v", c.Controller)
	}
	if c.Counters["fuse"] != 1 {
		t.Errorf("fuse counters %d, want 1", c.Counters["fuse"])
	}
	if !c.Tapped {
		t.Error("the artifact was not tapped for its cost")
	}
}

func TestSpmB12PumpkinBombsDamageIsTheTotalFuseCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bombs := pushCatalogPermanent(g, me.ID, "Pumpkin Bombs", "Artifact", spmB12PumpkinOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bombs {
				g.Battlefield.Cards[i].Counters = map[string]int{"fuse": 2}
			}
		}
	})
	d1 := spmB12Hand(me, "One", "Instant")
	d2 := spmB12Hand(me, "Two", "Instant")
	seedLibrary(me, "A", "B", "C", "D")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bombs, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d (three fuse counters)", opp.Life, life-3)
	}
}

func TestSpmB12PumpkinBombsRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bombs := pushCatalogPermanent(g, me.ID, "Pumpkin Bombs", "Artifact", spmB12PumpkinOracle, false)
	d1 := spmB12Hand(me, "One", "Instant")
	d2 := spmB12Hand(me, "Two", "Instant")
	oppTarget := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, bombs, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1}, Targets: oppTarget,
	}); err == nil {
		t.Error("activated with one discard")
	}
	if err := g.ActivateCatalogAbility(me.ID, bombs, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Error("targeted its own controller")
	}
}
