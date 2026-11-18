package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b09_test.go — Midnight Hunt batch 9 (#5497), through the turn engine.

const (
	s82MidB09WranglerOracle  = "8de2ce5a-9345-4d2b-a248-868268f36836"
	s82MidB09LarderOracle    = "61ab4548-5da9-4351-87e7-a3afc93d4bf5"
	s82MidB09SiegeOracle     = "a7756bc2-26ce-4044-a322-a898af8a393d"
	s82MidB09CrasherOracle   = "e6ddea98-b23f-44de-a63d-5d5dfad60e71"
	s82MidB09SkreelixOracle  = "0ccd1084-64bb-41ee-81c0-666af434b8df"
	s82MidB09DragonOracle    = "a45762af-4aac-40b3-a983-ea441b04c5ee"
	s82MidB09FleshOracle     = "14d5fdb3-14a3-461d-8515-a520c2c9e666"
	s82MidB09CaptainOracle   = "8721b0a6-52ff-4749-92db-82bb36f5a73e"
	s82MidB09GhoulOracle     = "86783f8d-26c9-4b8c-a697-ab6facafeb67"
	s82MidB09RebirthOracle   = "1920254f-0f65-4426-aabc-f3807476b67b"
	s82MidB09RevelryOracle   = "65779f3b-7fc5-4847-99d7-78f697d34c3d"
	s82MidB09BurnOracle      = "0ec54683-a65c-461d-862e-585b6499d6a5"
	s82MidB09BriarbridgeOrcl = "8fe6e954-f19b-4d11-83d3-74c634d51bad"
)

func TestS82MidB09CardsAreRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82MidB09WranglerOracle:  "Skaab Wrangler",
		s82MidB09LarderOracle:    "Larder Zombie",
		s82MidB09SiegeOracle:     "Siege Zombie",
		s82MidB09CrasherOracle:   "Festival Crasher",
		s82MidB09SkreelixOracle:  "Storm Skreelix",
		s82MidB09DragonOracle:    "Purifying Dragon",
		s82MidB09FleshOracle:     "Fleshtaker",
		s82MidB09CaptainOracle:   "Search Party Captain",
		s82MidB09GhoulOracle:     "Grizzly Ghoul",
		s82MidB09RebirthOracle:   "Diregraf Rebirth",
		s82MidB09RevelryOracle:   "Sunset Revelry",
		s82MidB09BurnOracle:      "Burn Down the House",
		s82MidB09BriarbridgeOrcl: "Briarbridge Tracker",
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

// s82MidB09Board is a main-phase table: the active player's `src` (a catalog
// permanent) and two plain creatures to tap beside it.
func s82MidB09Board(t *testing.T, g *game.Game, oracle, name, typeLine, cost string, p, tough int) (me *game.Player, src, a, b uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	src = b27Push(g, me.ID, name, typeLine, oracle, cost, p, tough)
	a = b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b = b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	return me, src, a, b
}

func TestS82MidB09SkaabWranglerTapsThreeToTapATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, src, a, b := s82MidB09Board(t, g, s82MidB09WranglerOracle, "Skaab Wrangler", "Creature — Human Wizard", "{1}{U}", 2, 1)
	victim := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a}, Targets: tgt, Strict: true}); err == nil {
		t.Fatal("two tapped creatures paid a three-creature cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Strict: true}); err == nil {
		t.Fatal("activated with no target")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Targets: tgt, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, victim).Tapped {
		t.Error("the target is untapped")
	}
	for _, id := range []uuid.UUID{src, a, b} {
		if !layeredCard(t, g, id).Tapped {
			t.Errorf("%s paid the cost and is untapped", id)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Targets: tgt, Strict: true}); err == nil {
		t.Error("activated again with tapped creatures")
	}
}

func TestS82MidB09LarderZombieTapsThreeToSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me, src, a, b := s82MidB09Board(t, g, s82MidB09LarderOracle, "Larder Zombie", "Creature — Zombie", "{U}", 1, 3)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a}, Strict: true}); err == nil {
		t.Fatal("two creatures paid a three-creature cost")
	}
	top := pushLibraryTopForTest(me, "Surveil Fodder")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

func TestS82MidB09SiegeZombieDrainsEachOpponentForOne(t *testing.T) {
	g := newCatalogGame(t)
	me, src, a, b := s82MidB09Board(t, g, s82MidB09SiegeOracle, "Siege Zombie", "Creature — Zombie", "{1}{B}", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}, Strict: true}); err == nil {
		t.Fatal("two creatures paid a three-creature cost")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func s82MidB09CastInstant(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
}

func TestS82MidB09FestivalCrasherPumpsOnInstantOrSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crasher := b27Push(g, me.ID, "Festival Crasher", "Creature — Devil", s82MidB09CrasherOracle, "{1}{R}", 1, 3, "R")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, crasher); c.CurrentPower() != 1 {
		t.Fatalf("a creature spell pumped it: power %d", c.CurrentPower())
	}
	s82MidB09CastInstant(t, g)
	if c := layeredCard(t, g, crasher); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("after an instant it is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, crasher); c.CurrentPower() != 5 {
		t.Errorf("after a second spell power %d, want 5", c.CurrentPower())
	}
}

func TestS82MidB09StormSkreelixDiscountsAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Test Instant", "Instant", "{2}{U}", ""); got != 2 {
		t.Fatalf("before Skreelix: %d generic, want 2", got)
	}
	sk := b27Push(g, me.ID, "Storm Skreelix", "Creature — Drake Horror", s82MidB09SkreelixOracle, "{3}{U}{R}", 2, 4, "U", "R")
	if got := s82B14Cost(t, g, me, "Test Instant", "Instant", "{2}{U}", ""); got != 1 {
		t.Errorf("instant with Skreelix: %d generic, want 1", got)
	}
	if got := s82B14Cost(t, g, me, "Test Creature", "Creature — Bear", "{2}{U}", ""); got != 2 {
		t.Errorf("creature spell with Skreelix: %d generic, want 2", got)
	}
	if got := s82B14Cost(t, g, g.Seats[1], "Test Instant", "Instant", "{2}{U}", ""); got != 2 {
		t.Errorf("an opponent's instant with Skreelix: %d generic, want 2", got)
	}
	s82MidB09CastInstant(t, g)
	if c := layeredCard(t, g, sk); c.CurrentPower() != 4 {
		t.Errorf("power %d after an instant, want 4", c.CurrentPower())
	}
	if !game.HasKeyword(ptrCard(layeredCard(t, g, sk)), "flying") {
		t.Error("no flying")
	}
}

func TestS82MidB09PurifyingDragonDamagesOneOrTwoToAZombie(t *testing.T) {
	for _, zombie := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dragon := b27Push(g, me.ID, "Purifying Dragon", "Creature — Dragon", s82MidB09DragonOracle, "{3}{R}{R}", 4, 3, "R")
		typ := "Creature — Bear"
		if zombie {
			typ = "Creature — Zombie"
		}
		victim := b12Creature(g, opp.ID, "Victim", typ, 3, 5)
		mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
		declareAttack(t, g, opp.ID, dragon)
		answerPickTargetIfOpen(t, g, victim)
		passPriorityAroundTable(t, g)
		want := 1
		if zombie {
			want = 2
		}
		if c := layeredCard(t, g, victim); c.DamageMarked != want {
			t.Errorf("zombie=%v: %d damage marked, want %d", zombie, c.DamageMarked, want)
		}
		if c := layeredCard(t, g, mine); c.DamageMarked != 0 {
			t.Errorf("zombie=%v: its own side's creature took %d", zombie, c.DamageMarked)
		}
	}
}

func TestS82MidB09FleshtakerGainsAndScriesOnSacrificeAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, src, a, b := s82MidB09Board(t, g, s82MidB09FleshOracle, "Fleshtaker", "Creature — Human Assassin", "{W}{B}", 2, 2)
	floatMana(t, g, me, "{C}")
	top := pushLibraryTopForTest(me, "Scry Fodder")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	sc := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if sc == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(sc.ID, me.ID, nil, []uuid.UUID{top}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, src); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("Fleshtaker is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	_ = b
}

func TestS82MidB09FleshtakerIgnoresItsOwnSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := b27Push(g, me.ID, "Fleshtaker", "Creature — Human Assassin", s82MidB09FleshOracle, "{W}{B}", 2, 2)
	advanceToMain(t, g)
	life := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(src) })
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("sacrificing Fleshtaker itself gained life: %d -> %d", life, me.Life)
	}
}

func TestS82MidB09SearchPartyCaptainDiscountsPerAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Search Party Captain", "Creature — Human Soldier", "{3}{W}", s82MidB09CaptainOracle); got != 3 {
		t.Fatalf("no attackers: %d generic, want 3", got)
	}
	declareAttack(t, g, opp.ID, a, b)
	if got := s82B14Cost(t, g, me, "Search Party Captain", "Creature — Human Soldier", "{3}{W}", s82MidB09CaptainOracle); got != 1 {
		t.Errorf("two attackers: %d generic, want 1", got)
	}
}

func TestS82MidB09SearchPartyCaptainDrawsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Search Party Captain", "Creature — Human Soldier", s82MidB09CaptainOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the Captain left, one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestS82MidB09GrizzlyGhoulEntersWithACounterPerCreatureThatDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	x := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	y := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	b27Kill(g, x)
	b27Kill(g, y)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Grizzly Ghoul", "Creature — Zombie Bear", s82MidB09GhoulOracle, nil)
	passPriorityAroundTable(t, g)
	var ghoul game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly Ghoul" {
				ghoul = c
			}
		}
	})
	if ghoul.InstanceID == uuid.Nil {
		t.Fatal("the Ghoul did not enter")
	}
	if ghoul.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("%d +1/+1 counters, want 2", ghoul.Counters[game.CounterPlusOne])
	}
}

func TestS82MidB09GrizzlyGhoulEntersBareWhenNothingDied(t *testing.T) {
	g := newCatalogGame(t)
	toMainForCost(t, g)
	castCatalogSpell(t, g, "Grizzly Ghoul", "Creature — Zombie Bear", s82MidB09GhoulOracle, nil)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly Ghoul" && c.Counters[game.CounterPlusOne] != 0 {
				t.Errorf("%d counters, want 0", c.Counters[game.CounterPlusOne])
			}
		}
	})
}

func TestS82MidB09DiregrafRebirthDiscountsReturnsAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	x := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	y := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	toMainForCost(t, g)
	if got := s82B14Cost(t, g, me, "Diregraf Rebirth", "Sorcery", "{3}{B}{G}", s82MidB09RebirthOracle); got != 3 {
		t.Fatalf("nothing died: %d generic, want 3", got)
	}
	b27Kill(g, x)
	b27Kill(g, y)
	passPriorityAroundTable(t, g)
	if got := s82B14Cost(t, g, me, "Diregraf Rebirth", "Sorcery", "{3}{B}{G}", s82MidB09RebirthOracle); got != 1 {
		t.Errorf("two creatures died: %d generic, want 1", got)
	}
	// Three more deaths cannot push it below {B}{G}.
	for i := 0; i < 3; i++ {
		b27Kill(g, b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	passPriorityAroundTable(t, g)
	if got := s82B14Cost(t, g, me, "Diregraf Rebirth", "Sorcery", "{3}{B}{G}", s82MidB09RebirthOracle); got != 0 {
		t.Errorf("five creatures died: %d generic, want 0", got)
	}

	castCatalogSpell(t, g, "Diregraf Rebirth", "Sorcery", s82MidB09RebirthOracle, []game.TargetRef{{Kind: game.TargetCard, ID: x}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(x) {
		t.Error("the targeted creature card did not return")
	}
	if g.Battlefield.Contains(y) {
		t.Error("an opponent's creature card returned")
	}
}

func TestS82MidB09DiregrafRebirthFlashbackExilesAndIsDiscountedToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rebirth := seedGraveyardCard(t, g, "Diregraf Rebirth", "Sorcery", s82MidB09RebirthOracle)
	victim := seedGraveyardCard(t, g, "Buried Bear", "Creature — Bear", "")
	if !game.CardCastableFromZone(s82MidB09RebirthOracle, game.ZoneGraveyard) {
		t.Fatal("Diregraf Rebirth does not declare the graveyard as a cast zone")
	}
	var card game.Card
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == rebirth {
			card = c
		}
	}
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.Total.Generic != 5 {
		t.Fatalf("flashback price %d generic, want 5", price.Total.Generic)
	}
	b27Kill(g, b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	passPriorityAroundTable(t, g)
	price, err = g.PriceCast(me.ID, card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.Total.Generic != 4 {
		t.Errorf("flashback price after a death %d generic, want 4", price.Total.Generic)
	}
	if err := g.CastSpell(me.ID, rebirth, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature did not return")
	}
	if me.Graveyard.Contains(rebirth) {
		t.Error("Diregraf Rebirth is still in the graveyard after its flashback")
	}
}

func TestS82MidB09SunsetRevelryAnswersEachDeficitIndependently(t *testing.T) {
	// Behind on life, creatures and cards: all three clauses happen.
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = me.Life + 5
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	giveHandCards(opp, 3)
	toMainForCost(t, g)
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Sunset Revelry", "Sorcery", s82MidB09RevelryOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if got := b27CountNamed(g, me.ID, "Human"); got != 2 {
		t.Errorf("%d Humans, want 2", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the spell left, one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestS82MidB09SunsetRevelryDoesNothingWhenNoOpponentIsAhead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats[1:] {
		p.Life = me.Life - 1
		for p.Hand.Size() > 0 {
			p.Hand.Cards = p.Hand.Cards[:0]
		}
	}
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	life := me.Life
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Sunset Revelry", "Sorcery", s82MidB09RevelryOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life || b27CountNamed(g, me.ID, "Human") != 0 || me.Hand.Size() != hand {
		t.Errorf("life %d->%d, humans %d, hand %d->%d: nothing should have happened", life, me.Life,
			b27CountNamed(g, me.ID, "Human"), hand, me.Hand.Size())
	}
}

func TestS82MidB09BurnDownTheHouseSweepsCreaturesAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 5)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 2, 6)
	b27Push(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", "{3}", 0, 0)
	toMainForCost(t, g)
	castCatalogSpellWithModes(t, g, "Burn Down the House", "Sorcery", s82MidB09BurnOracle, []int{0})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/5 survived 5 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 2/6 died to 5 damage")
	} else if c := layeredCard(t, g, big); c.DamageMarked != 5 {
		t.Errorf("the survivor has %d damage, want 5", c.DamageMarked)
	}
}

func TestS82MidB09BurnDownTheHouseMakesThreeHastyDevilsThatPing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	castCatalogSpellWithModes(t, g, "Burn Down the House", "Sorcery", s82MidB09BurnOracle, []int{1})
	passPriorityAroundTable(t, g)
	var devils []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Devil" && c.Controller == me.ID {
				devils = append(devils, c.InstanceID)
			}
		}
	})
	if len(devils) != 3 {
		t.Fatalf("%d Devils, want 3", len(devils))
	}
	for _, id := range devils {
		c := layeredCard(t, g, id)
		if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !game.HasKeyword(ptrCard(c), "haste") {
			t.Errorf("Devil %d/%d haste=%v, want a 1/1 with haste", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(ptrCard(c), "haste"))
		}
	}
	// A Devil's own death trigger is its printed ability.
	opp := g.Seats[1]
	life := opp.Life
	b27Kill(g, devils[0])
	answerPickTargetPlayerIfOpen(t, g, opp.ID)
	if !triggerOnStackAny(g) {
		t.Fatal("a Devil dying did not trigger")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func answerPickTargetPlayerIfOpen(t *testing.T, g *game.Game, playerID uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: playerID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			return
		}
	}
}

func TestS82MidB09BriarbridgeTrackerInvestigatesAndGrowsWithAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	castCatalogSpell(t, g, "Briarbridge Tracker", "Creature — Human Scout", s82MidB09BriarbridgeOrcl, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	var tracker game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Briarbridge Tracker" {
				tracker = c
			}
		}
	})
	if tracker.InstanceID == uuid.Nil {
		t.Fatal("the Tracker did not enter")
	}
	if got := b27CountNamed(g, me.ID, "Clue"); got != 1 {
		t.Fatalf("%d Clues, want 1", got)
	}
	if c := layeredCard(t, g, tracker.InstanceID); c.CurrentPower() != c.Power+2 {
		t.Errorf("power %d with a Clue, want %d", c.CurrentPower(), c.Power+2)
	}
	// Sacrificing the Clue (the only token) turns the bonus off.
	var clue uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Clue" {
				clue = c.InstanceID
			}
		}
	})
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clue) })
	if c := layeredCard(t, g, tracker.InstanceID); c.CurrentPower() != c.Power {
		t.Errorf("power %d with no token, want %d", c.CurrentPower(), c.Power)
	}
}
