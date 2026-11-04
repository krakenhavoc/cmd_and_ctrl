package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm14_cards_test.go — Marvel's Spider-Man family, batch 14 (#4460). Each
// test plays the real catalog entry and asserts what its text says.

const (
	spm14PenanceStareOracle  = "e21a1ad5-bf31-425f-92c2-7818bdb1cef8"
	spm14RiseNetherOracle    = "cc8c0234-a355-435d-b3de-8a6fc024c628"
	spm14TempleKhonshuOracle = "34916c4b-fc1b-4864-bb3f-968977fd533d"
	spm14DarkholdOracle      = "3981970f-b332-49b7-9698-3296830ac399"
	spm14AlibouOracle        = "e5e8e116-10fe-48b0-b3d8-6edb39bd5f90"
	spm14AnthemOracle        = "f1d8e9a6-1903-4be5-8609-1009a463f393"
	spm14BedlamOracle        = "c28d4caf-9b7c-40b1-9cbe-275ca9dae06a"
	spm14BlackPantherOracle  = "48deff27-b421-430b-a484-55b4a593010a"
	spm14ConcertedOracle     = "fd4236e5-6a2c-45ec-afb3-cd4fc516fbf1"
	spm14TrueNameOracle      = "112322ad-8f66-4cd4-98a1-f425d61a69ce"
	spm14BurrentonOracle     = "70bb275b-3458-4690-a50f-b231fbf0bccb"
	spm14SandSquidOracle     = "40578245-2cb8-4a6a-aeff-2252719c2d1c"
)

func TestSpm14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spm14PenanceStareOracle:  "Penance Stare",
		spm14RiseNetherOracle:    "Rise from the Nether",
		spm14TempleKhonshuOracle: "Temple of Khonshu",
		spm14DarkholdOracle:      "The Darkhold",
		spm14AlibouOracle:        "Alibou, Ancient Witness",
		spm14AnthemOracle:        "Anthem of Champions",
		spm14BedlamOracle:        "Bedlam",
		spm14BlackPantherOracle:  "Black Panther, Wakandan King",
		spm14ConcertedOracle:     "Concerted Effort",
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

func spm14Hand(me *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: me.ID, Controller: me.ID,
	})
	return id
}

func spm14Costed(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Penance Stare ---

func TestSpm14PenanceStareDestroysAnAttackerWithManaValueThreeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	cheap := spm14Costed(g, opp.ID, "Cheap Attacker", "{1}{G}{G}", 3, 3)
	declareAttack(t, g, me.ID, cheap)
	spell := spm14Hand(me, "Penance Stare", "Instant", spm14PenanceStareOracle, "{W}")
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(cheap), Strict: true}); err != nil {
		t.Fatalf("cast on an attacking mana value 3 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the attacking creature survived Penance Stare")
	}
}

func TestSpm14PenanceStareRefusesTooExpensiveAndNonAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
	dear := spm14Costed(g, opp.ID, "Dear Attacker", "{2}{G}{G}", 4, 4)
	idle := spm14Costed(g, opp.ID, "Idle Bear", "{1}{G}", 2, 2)
	declareAttack(t, g, me.ID, dear)
	spell := spm14Hand(me, "Penance Stare", "Instant", spm14PenanceStareOracle, "{W}")
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(dear), Strict: true}); err == nil {
		t.Error("targeted an attacking creature with mana value 4")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(idle), Strict: true}); err == nil {
		t.Error("targeted a creature that is not attacking")
	}
}

// --- Rise from the Nether ---

func TestSpm14RiseFromTheNetherMakesATokenPerNoncreatureNonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	push := func(name, typeLine string) {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	}
	push("Bolt", "Instant")
	push("Divination", "Sorcery")
	push("Old Wall", "Enchantment")
	push("Gadget", "Artifact")
	push("Gizmo Bear", "Artifact Creature — Construct") // a creature card: no
	push("Forest", "Basic Land — Forest")               // a land card: no
	push("Cave", "Land Artifact")                       // a land card: no
	spell := spm14Hand(me, "Rise from the Nether", "Sorcery — Arcane", spm14RiseNetherOracle, "{6}{R}")
	floatMana(t, g, me, "{R}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Mindless-One" {
				n++
				if c.Power != 3 || c.Toughness != 1 {
					t.Errorf("token is %d/%d, want 3/1", c.Power, c.Toughness)
				}
			}
		}
	})
	if n != 4 {
		t.Errorf("%d Mindless-One tokens, want 4 (Bolt, Divination, Old Wall, Gadget)", n)
	}
}

func TestSpm14RiseFromTheNetherWithNothingToCountMakesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spell := spm14Hand(me, "Rise from the Nether", "Sorcery — Arcane", spm14RiseNetherOracle, "{6}{R}")
	floatMana(t, g, me, "{R}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Mindless-One" {
				t.Error("a token was made with an empty graveyard")
			}
		}
	})
}

func TestSpm14RiseFromTheNetherRefusesTooLittleMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spell := spm14Hand(me, "Rise from the Nether", "Sorcery — Arcane", spm14RiseNetherOracle, "{6}{R}")
	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast {6}{R} with three mana")
	}
}

// --- Temple of Khonshu ---

func spm14PlayTemple(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return playLandFromHand(t, g, "Temple of Khonshu", spm14TempleKhonshuOracle)
}

func TestSpm14TempleEntersTappedWithoutADimension(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Some Enchantment", "Enchantment")
	id := spm14PlayTemple(t, g, me)
	got, ok := aangCardOnBF(g, id)
	if !ok || !got.Tapped {
		t.Errorf("Temple entered untapped with no Dimension (found %v, tapped %v)", ok, got.Tapped)
	}
}

func TestSpm14TempleEntersUntappedWithADimension(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "A Dimension", "Enchantment — Dimension")
	id := spm14PlayTemple(t, g, me)
	got, ok := aangCardOnBF(g, id)
	if !ok || got.Tapped {
		t.Errorf("Temple entered tapped though you control a Dimension (found %v, tapped %v)", ok, got.Tapped)
	}
}

func TestSpm14TempleIgnoresAnOpponentsDimension(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, opp.ID, "Their Dimension", "Enchantment — Dimension")
	id := spm14PlayTemple(t, g, me)
	if got, _ := aangCardOnBF(g, id); !got.Tapped {
		t.Error("an opponent's Dimension kept the Temple untapped")
	}
}

func TestSpm14TempleMakesWhiteOrBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "A Dimension", "Enchantment — Dimension")
	id := spm14PlayTemple(t, g, me)
	spec, ok := Lookup(spm14TempleKhonshuOracle)
	if !ok || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|B}" {
		t.Fatalf("mana abilities %+v", spec.ManaAbilities)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
}

// --- The Darkhold ---

func spm14Darkhold(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "The Darkhold", "Legendary Artifact — Book", spm14DarkholdOracle, false)
}

func TestSpm14DarkholdDrawsPerSinCounterThenLosesLifePerCardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six")
	dark := spm14Darkhold(g, me)
	hand, life := me.Hand.Size(), me.Life

	// First activation: one counter, one card, then life = cards in hand.
	b16Activate(t, g, me.ID, dark, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := spmB4Live(t, g, dark).Counters["sin"]; got != 1 {
		t.Fatalf("sin counters %d, want 1", got)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-(hand+1) {
		t.Fatalf("after 1st: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+1, me.Life, life-(hand+1))
	}

	// Untap and activate again: two counters, two cards.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dark {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	hand, life = me.Hand.Size(), me.Life
	b16Activate(t, g, me.ID, dark, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := spmB4Live(t, g, dark).Counters["sin"]; got != 2 {
		t.Fatalf("sin counters %d, want 2", got)
	}
	if me.Hand.Size() != hand+2 || me.Life != life-(hand+2) {
		t.Fatalf("after 2nd: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+2, me.Life, life-(hand+2))
	}
}

func TestSpm14DarkholdSecondAbilityLevelsEveryoneToTheLowest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	dark := spm14Darkhold(g, me)
	me.Life = 31
	g.Seats[1].Life = 12
	g.Seats[2].Life = 27
	g.Seats[3].Life = 40
	floatMana(t, g, me, "{B}{B}{C}{C}{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, dark, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	for i, s := range g.Seats {
		if s.Life != 12 {
			t.Errorf("seat %d life %d, want 12", i, s.Life)
		}
	}
}

func TestSpm14DarkholdSecondAbilityIsSorceryOnlyAndNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	dark := spm14Darkhold(g, me)
	floatMana(t, g, me, "{B}{B}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, dark, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated {6}{B}{B} with seven mana")
	}
	// Not at sorcery speed: during combat, with enough mana.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	dark2 := spm14Darkhold(g2, me2)
	advanceTo(t, g2, game.StepBeginCombat)
	floatMana(t, g2, me2, "{B}{B}{C}{C}{C}{C}{C}{C}")
	if err := g2.ActivateCatalogAbility(me2.ID, dark2, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated the second ability outside a main phase")
	}
}

// --- Anthem of Champions ---

func TestSpm14AnthemPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	pushCatalogPermanent(g, me.ID, "Anthem of Champions", "Enchantment", spm14AnthemOracle, false)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tg)
	}
	if p, tg := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tg != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tg)
	}
}

// --- Bedlam ---

func spm14BlockAttempt(t *testing.T, withBedlam bool) error {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocker := b12Creature(g, me.ID, "Blocker", "Creature — Bear", 2, 2)
	if withBedlam {
		pushCatalogPermanent(g, opp.ID, "Bedlam", "Enchantment", spm14BedlamOracle, false)
	}
	advanceToNextSeatsTurn(t, g)
	raider := b12Creature(g, opp.ID, "Raider", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, raider)
}

func TestSpm14BedlamStopsEveryBlock(t *testing.T) {
	if err := spm14BlockAttempt(t, false); err != nil {
		t.Fatalf("control: the block is legal without Bedlam: %v", err)
	}
	if err := spm14BlockAttempt(t, true); err == nil {
		t.Error("a creature blocked while Bedlam was on the battlefield")
	}
}

// --- Black Panther, Wakandan King ---

func spm14PanthersLand(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
}

func spm14Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return spmB4Live(t, g, id).Counters[game.CounterPlusOne]
}

func TestSpm14BlackPantherPutsACounterOnALandWhenCreaturesEnter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := spm14PanthersLand(g, me)
	// Black Panther itself entering.
	if _, err := spmB4Cast(t, g, "Black Panther, Wakandan King", "Legendary Creature — Human Noble Hero", spm14BlackPantherOracle, "", 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	if got := spm14Counters(t, g, land); got != 1 {
		t.Fatalf("after Black Panther entered: %d counters, want 1", got)
	}
	// Another creature you control.
	if _, err := spmB4Cast(t, g, "Bear", "Creature — Bear", "", "", 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	if got := spm14Counters(t, g, land); got != 2 {
		t.Fatalf("after a second creature entered: %d counters, want 2", got)
	}
	// An opponent's creature does nothing.
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if got := spm14Counters(t, g, land); got != 2 {
		t.Errorf("an opponent's creature added a counter: %d", got)
	}
}

func TestSpm14BlackPantherMineVibraniumMovesCountersGainsLifeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := spm14PanthersLand(g, me)
	bp, err := spmB4Cast(t, g, "Black Panther, Wakandan King", "Legendary Creature — Human Noble Hero", spm14BlackPantherOracle, "", 2, 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(land, game.CounterPlusOne, 2); err != nil {
			t.Fatal(err)
		}
	})
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	passPriorityAroundTable(t, g)
	seedLibrary(me, "One", "Two")
	hand, life := me.Hand.Size(), me.Life
	if got := spm14Counters(t, g, land); got != 3 {
		t.Fatalf("setup: land has %d counters, want 3", got)
	}

	floatMana(t, g, me, "{C}{C}{C}")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	}}
	if err := g.ActivateCatalogAbility(me.ID, bp, 0, params); err != nil {
		t.Fatalf("Mine Vibranium: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := spm14Counters(t, g, land); got != 0 {
		t.Errorf("the land still has %d counters", got)
	}
	if got := spm14Counters(t, g, victim); got != 3 {
		t.Errorf("the creature has %d counters, want 3", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestSpm14BlackPantherMineVibraniumWithNoCountersGainsAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := spm14PanthersLand(g, me)
	bp := b12Push(g, me.ID, "Black Panther, Wakandan King", "Legendary Creature — Human Noble Hero", spm14BlackPantherOracle, 2, 2)
	answerPickTargetIfOpen(t, g, land)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(land, game.CounterPlusOne, -1); err != nil {
			t.Fatal(err)
		}
	})
	if got := spm14Counters(t, g, land); got != 0 {
		t.Fatalf("setup: land has %d counters", got)
	}
	seedLibrary(me, "One", "Two")
	hand, life := me.Hand.Size(), me.Life
	floatMana(t, g, me, "{C}{C}{C}")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Slot: 0},
		{Kind: game.TargetCard, ID: bp, Slot: 1},
	}}
	if err := g.ActivateCatalogAbility(me.ID, bp, 0, params); err != nil {
		t.Fatalf("Mine Vibranium: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d -> %d, hand %d -> %d; nothing moved so nothing happens", life, me.Life, hand, me.Hand.Size())
	}
}

func TestSpm14BlackPantherMineVibraniumRefusesAnOpponentsLandAndTooLittleMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bp := b12Push(g, me.ID, "Black Panther, Wakandan King", "Legendary Creature — Human Noble Hero", spm14BlackPantherOracle, 2, 2)
	theirLand := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	params := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: theirLand, Slot: 0},
		{Kind: game.TargetCard, ID: bp, Slot: 1},
	}}
	if err := g.ActivateCatalogAbility(me.ID, bp, 0, params); err == nil {
		t.Error("targeted an opponent's land")
	}
	myLand := spm14PanthersLand(g, me)
	answerPickTargetIfOpen(t, g, myLand)
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	bp2 := b12Push(g2, me2.ID, "Black Panther, Wakandan King", "Legendary Creature — Human Noble Hero", spm14BlackPantherOracle, 2, 2)
	land2 := spm14PanthersLand(g2, me2)
	floatMana(t, g2, me2, "{C}{C}")
	params2 := game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: land2, Slot: 0},
		{Kind: game.TargetCard, ID: bp2, Slot: 1},
	}}
	if err := g2.ActivateCatalogAbility(me2.ID, bp2, 0, params2); err == nil {
		t.Error("activated {3} with two mana")
	}
}

func answerPickTargetIfOpen(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			answerPickTarget(t, g, target)
			return
		}
	}
}

// --- Alibou, Ancient Witness ---

func spm14Alibou(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Alibou, Ancient Witness", "Legendary Artifact Creature — Golem", spm14AlibouOracle, 4, 5)
}

func TestSpm14AlibouGivesOtherArtifactCreaturesYouControlHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alibou := spm14Alibou(g, me)
	gizmo := b12Creature(g, me.ID, "Gizmo", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Gizmo", "Artifact Creature — Construct", 1, 1)
	treasure := b12Permanent(g, me.ID, "Trinket", "Artifact")
	assertKeywords(t, g, gizmo, "haste")
	for name, id := range map[string]uuid.UUID{"Alibou itself": alibou, "a non-artifact creature": bear, "an opponent's artifact creature": theirs, "a noncreature artifact": treasure} {
		if containsString(effectiveAbilities(t, g, id), "haste") {
			t.Errorf("%s has haste", name)
		}
	}
}

func TestSpm14AlibouDamagesAndScriesForTappedArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	alibou := spm14Alibou(g, me)
	gizmo := b12Creature(g, me.ID, "Gizmo", "Artifact Creature — Construct", 1, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Trinket", TypeLine: "Artifact", Tapped: true, Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Idle Trinket", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	seedLibrary(me, "One", "Two", "Three", "Four", "Five")
	life := opp.Life
	// Alibou and Gizmo attack together: one trigger. Tapped artifacts are
	// the Trinket, Alibou and Gizmo: X = 3.
	declareAttack(t, g, opp.ID, alibou, gizmo)
	answerPickTargetPlayer(t, g, opp.ID)
	answerScryIfAsked(t, g, me)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d (one trigger, X = 3)", opp.Life, life-3)
	}
}

func TestSpm14AlibouIgnoresNonArtifactAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	spm14Alibou(g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a non-artifact attacker triggered Alibou: life %d -> %d", life, opp.Life)
	}
}

// --- Concerted Effort ---

func spm14Upkeep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepUpkeep)
	passPriorityAroundTable(t, g)
}

func TestSpm14ConcertedEffortSharesKeywordsAtEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Concerted Effort", "Enchantment", spm14ConcertedOracle, false)
	angel := b12Push(g, me.ID, "Serra Angel", "Creature — Angel", serraAngelOracle, 4, 4)
	dread := b12Push(g, me.ID, "Colossal Dreadmaw", "Creature — Dinosaur", colossalDreadmawOracle, 6, 6)
	swamp := b12Push(g, me.ID, "Sand Squid", "Creature — Squid", spm14SandSquidOracle, 6, 6)
	burr := b12Push(g, me.ID, "Burrenton Forge-Tender", "Creature — Kithkin Wizard", spm14BurrentonOracle, 1, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertNoKeywords := func(id uuid.UUID, kws ...string) {
		t.Helper()
		for _, kw := range kws {
			if containsString(effectiveAbilities(t, g, id), kw) {
				t.Errorf("%s has %q", id, kw)
			}
		}
	}
	assertNoKeywords(bear, "flying", "vigilance", "trample", "islandwalk", "protection from red")

	spm14Upkeep(t, g) // an opponent's upkeep still triggers it
	for _, id := range []uuid.UUID{angel, dread, swamp, burr, bear} {
		assertKeywords(t, g, id, "flying", "vigilance", "trample", "islandwalk", "protection from red")
	}
	assertNoKeywords(bear, "fear", "first strike", "double strike")
	assertNoKeywords(theirs, "flying", "vigilance", "trample", "islandwalk", "protection from red")
}

func TestSpm14ConcertedEffortGivesNothingWhenNoCreatureHasTheKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Concerted Effort", "Enchantment", spm14ConcertedOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	spm14Upkeep(t, g)
	for _, kw := range []string{"flying", "fear", "first strike", "double strike", "trample", "vigilance", "islandwalk"} {
		for _, id := range []uuid.UUID{bear, other} {
			if containsString(effectiveAbilities(t, g, id), kw) {
				t.Errorf("a creature gained %q with nobody having it", kw)
			}
		}
	}
}

func TestSpm14ConcertedEffortSharesProtectionFromTheChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Concerted Effort", "Enchantment", spm14ConcertedOracle, false)
	nemesis := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "True-Name Nemesis", TypeLine: "Creature — Human Wizard", OracleID: spm14TrueNameOracle,
		Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID, ChosenPlayer: opp.ID,
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spm14Upkeep(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != bear {
				continue
			}
			card := c
			protected := false
			for _, q := range game.ProtectionQualities(&card) {
				if q.Kind == game.ProtectionQualityPlayer && q.Player == opp.ID {
					protected = true
				}
			}
			if !protected {
				t.Errorf("the Bear has no protection from the opponent the Nemesis chose: %v", card.Effective().Abilities)
			}
		}
	})
	_ = nemesis
}
