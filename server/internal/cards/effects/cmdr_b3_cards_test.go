package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b3_cards_test.go — the Commander deck cards of batches 3 and 4
// (#5368, Foundations Commander).

const (
	cb3ThrabenWatcherOracle     = "53cd9d9f-c1a7-404e-9934-4542415e3431"
	cb3WojekInvestigatorOracle  = "ab41243d-b178-4ebe-a7ac-bea37427be99"
	cb3BroodstarOracle          = "428cc1cb-a491-4cc4-9436-28ee8acc37d6"
	cb3FallFromFavorOracle      = "e89e1fae-bac3-4cc0-a233-24b9e08a5290"
	cb3LaunchMishapOracle       = "5d4ed768-0a37-4604-8121-43920ed9fb4a"
	cb3ResearchThiefOracle      = "ba5055ce-6809-47a4-bee6-fb97a5b60f8a"
	cb3ShimmerDragonOracle      = "de1316f2-1cca-4ac4-aead-0fd8b4036d88"
	cb3TamiyosLogbookOracle     = "d9356199-4ce3-44fa-9463-90fc134088f9"
	cb3BadMoonOracle            = "fc5d3341-cbce-49e5-93cc-8add92479dca"
	cb3EternalTaskmasterOracle  = "a7c000a8-1864-4bb2-80fa-a61d74dbd966"
	cb4GravebornMuseOracle      = "d2a65ff4-ab82-42f7-9d90-feb29132121e"
	cb4JosuVessOracle           = "974a46f9-aa84-4b34-bee5-c635166e5841"
	cb4LilianasMasteryOracle    = "bd104c7e-311e-4b03-98d3-5f20f3a99d26"
	cb4LilianasReaverOracle     = "97332d4e-a55c-4e4f-9040-bf426a7c8e39"
	cb4MarchesasDecreeOracle    = "e4ab5ee4-4683-4219-9e15-31975e69efbb"
	cb4MireTritonOracle         = "4fe6f784-0290-4d4a-85ae-0b66522d31e9"
	cb4MoanOfTheUnhallowedOracl = "ce87e5ed-2562-4563-9a5b-bd53c04ec4a5"
	cb4MutilateOracle           = "698e79ec-c2bc-4973-a095-f3c57b50e39c"
	cb4NecroticHexOracle        = "a6e720a2-9819-40ec-a811-7812898b531a"
	cb4RazorlashOracle          = "580996d4-db9a-4b57-a0e1-a2dbb524121a"
)

func TestCmdrB3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cb3ThrabenWatcherOracle:     "Thraben Watcher",
		cb3WojekInvestigatorOracle:  "Wojek Investigator",
		cb3BroodstarOracle:          "Broodstar",
		cb3FallFromFavorOracle:      "Fall from Favor",
		cb3LaunchMishapOracle:       "Launch Mishap",
		cb3ResearchThiefOracle:      "Research Thief",
		cb3ShimmerDragonOracle:      "Shimmer Dragon",
		cb3TamiyosLogbookOracle:     "Tamiyo's Logbook",
		cb3BadMoonOracle:            "Bad Moon",
		cb3EternalTaskmasterOracle:  "Eternal Taskmaster",
		cb4GravebornMuseOracle:      "Graveborn Muse",
		cb4JosuVessOracle:           "Josu Vess, Lich Knight",
		cb4LilianasMasteryOracle:    "Liliana's Mastery",
		cb4LilianasReaverOracle:     "Liliana's Reaver",
		cb4MarchesasDecreeOracle:    "Marchesa's Decree",
		cb4MireTritonOracle:         "Mire Triton",
		cb4MoanOfTheUnhallowedOracl: "Moan of the Unhallowed",
		cb4MutilateOracle:           "Mutilate",
		cb4NecroticHexOracle:        "Necrotic Hex",
		cb4RazorlashOracle:          "Razorlash Transmogrant",
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

// cb3Hand puts a catalog card in the active seat's hand with a printed cost.
func cb3Hand(g *game.Game, name, typeLine, mana, oracle string) uuid.UUID {
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: mana,
		OracleID: oracle, Owner: active.ID, Controller: active.ID})
	return id
}

// cb3Fund replaces the pool with the given mana.
func cb3Fund(g *game.Game, p *game.Player, colors ...string) {
	g.WithWriteLock(func() {
		p.ManaPool = nil
		for _, c := range colors {
			p.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func cb3Bf(g *game.Game, id uuid.UUID) (c game.Card, ok bool) {
	g.ReadSnapshot(func() {
		for _, x := range g.Battlefield.Cards {
			if x.InstanceID == id {
				c, ok = x, true
			}
		}
	})
	return c, ok
}

func cb3Tapped(g *game.Game, id uuid.UUID) bool {
	c, ok := cb3Bf(g, id)
	return ok && c.Tapped
}

// --- Thraben Watcher ----------------------------------------------------

func TestCmdrB3ThrabenWatcherPumpsOtherNontokenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	watcher := b12Push(g, me.ID, "Thraben Watcher", "Creature — Angel", cb3ThrabenWatcherOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier",
		TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("my bear is %d/%d, want 3/3", p, tg)
	}
	assertKeywords(t, g, bear, "vigilance")
	if p := effectivePower(t, g, watcher); p != 2 {
		t.Errorf("the Watcher pumps itself: power %d", p)
	}
	assertKeywords(t, g, watcher, "flying", "vigilance")
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is pumped: %d", p)
	}
	if p := effectivePower(t, g, token); p != 1 {
		t.Errorf("a token is pumped: %d", p)
	}
	if kws := effectiveAbilities(t, g, token); containsString(kws, "vigilance") {
		t.Errorf("a token has vigilance: %v", kws)
	}
}

// --- Bad Moon -----------------------------------------------------------

func TestCmdrB3BadMoonPumpsEveryBlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Bad Moon", "Enchantment", cb3BadMoonOracle, 0, 0)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Black Bear",
		TypeLine: "Creature — Bear", Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Black Bear",
		TypeLine: "Creature — Bear", Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
		TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tg != 3 {
			t.Errorf("a black creature is %d/%d, want 3/3", p, tg)
		}
	}
	if p := effectivePower(t, g, green); p != 2 {
		t.Errorf("a green creature is pumped: %d", p)
	}
}

// --- Liliana's Mastery --------------------------------------------------

func TestCmdrB4LilianasMasteryMakesTwoPumpedZombies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	zombie := b12Creature(g, me.ID, "Old Zombie", "Creature — Zombie", 2, 2)
	castHoldCreature(t, g, "Liliana's Mastery", "Enchantment", cb4LilianasMasteryOracle, 0, 0)
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 2 {
		t.Fatalf("Zombie tokens: %d, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && IsToken(c) {
			if p := effectivePower(t, g, c.InstanceID); p != 3 {
				t.Errorf("a Zombie token has power %d, want 3", p)
			}
		}
	}
	if p := effectivePower(t, g, zombie); p != 3 {
		t.Errorf("an existing Zombie has power %d, want 3", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Zombie is pumped: %d", p)
	}
}

// --- Mire Triton --------------------------------------------------------

func TestCmdrB4MireTritonMillsTwoAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	life, yard, lib := me.Life, me.Graveyard.Size(), me.Library.Size()
	id := castHoldCreature(t, g, "Mire Triton", "Creature — Zombie Merfolk", cb4MireTritonOracle, 2, 1)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if got := me.Graveyard.Size() - yard; got != 2 {
		t.Errorf("%d cards milled, want 2", got)
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("library fell by %d, want 2", got)
	}
	assertKeywords(t, g, id, "deathtouch")
}

// --- Moan of the Unhallowed ---------------------------------------------

func TestCmdrB4MoanMakesTwoZombiesAndHasFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	moan := cb3Hand(g, "Moan of the Unhallowed", "Sorcery", "{2}{B}{B}", cb4MoanOfTheUnhallowedOracl)
	advanceToMain(t, g)
	cb3Fund(g, me, "B", "B", "C", "C")
	if err := g.CastSpell(me.ID, moan, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 2 {
		t.Fatalf("Zombie tokens after the cast: %d, want 2", n)
	}
	if !me.Graveyard.Contains(moan) {
		t.Fatal("Moan is not in the graveyard")
	}
	// Flashback needs {5}{B}{B}: six mana is refused, seven pays.
	cb3Fund(g, me, "B", "B", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, moan, game.CastSpellParams{Strict: true, AlternativeCost: "flashback", FromZone: "graveyard"}); err == nil {
		t.Fatal("flashback paid with six mana was accepted")
	}
	cb3Fund(g, me, "B", "B", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, moan, game.CastSpellParams{Strict: true, AlternativeCost: "flashback", FromZone: "graveyard"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 4 {
		t.Errorf("Zombie tokens after flashback: %d, want 4", n)
	}
	if me.Graveyard.Contains(moan) {
		t.Error("a flashed-back Moan is still in the graveyard")
	}
}

// --- Josu Vess ----------------------------------------------------------

func TestCmdrB4JosuVessKickedMakesEightKnights(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		josu := cb3Hand(g, "Josu Vess, Lich Knight", "Legendary Creature — Zombie Knight", "{2}{B}{B}", cb4JosuVessOracle)
		advanceToMain(t, g)
		cb3Fund(g, me, "B", "B", "C", "C", "C", "C", "C", "C", "C", "C", "B")
		params := game.CastSpellParams{Strict: true}
		if kicked {
			params.OptionalCosts = []int{0}
		}
		if err := g.CastSpell(me.ID, josu, params); err != nil {
			t.Fatalf("cast (kicked %v): %v", kicked, err)
		}
		passPriorityAroundTable(t, g)
		n, _ := countTokensNamed(g, me.ID, "Zombie Knight")
		want := 0
		if kicked {
			want = 8
		}
		if n != want {
			t.Errorf("kicked %v: %d Zombie Knights, want %d", kicked, n, want)
		}
		if kicked {
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Zombie Knight" && IsToken(c) {
					assertKeywords(t, g, c.InstanceID, "menace")
					if c.Power != 2 || c.Toughness != 2 {
						t.Errorf("a Knight is %d/%d", c.Power, c.Toughness)
					}
					break
				}
			}
		}
	}
}

// --- Graveborn Muse -----------------------------------------------------

func TestCmdrB4GravebornMuseDrawsAndLosesPerZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Graveborn Muse", "Creature — Zombie Spirit", cb4GravebornMuseOracle, 3, 3)
	b12Creature(g, me.ID, "Zombie A", "Creature — Zombie", 2, 2)
	b12Creature(g, me.ID, "Zombie B", "Creature — Zombie", 2, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Their Zombie", "Creature — Zombie", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (the Muse and two Zombies)", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

// --- Wojek Investigator -------------------------------------------------

func TestCmdrB3WojekInvestigatorInvestigatesPerOpponentWithMoreCards(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b12Push(g, me.ID, "Wojek Investigator", "Creature — Angel Detective", cb3WojekInvestigatorOracle, 2, 4)
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		for i := 0; i < 3; i++ {
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
		for _, o := range []*game.Player{o1, o2, o3} {
			o.Hand.Cards = nil
		}
		for i := 0; i < 5; i++ { // more than me
			o1.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Sorcery", Owner: o1.ID, Controller: o1.ID})
		}
		for i := 0; i < 4; i++ { // more than me even after my draw? no: 4 > 4 is false
			o2.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Sorcery", Owner: o2.ID, Controller: o2.ID})
		}
		for i := 0; i < 9; i++ {
			o3.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Sorcery", Owner: o3.ID, Controller: o3.ID})
		}
	})
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// Upkeep comes before the draw step: I hold 3, so o1 (5), o2 (4) and
	// o3 (9) each have more.
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 3 {
		t.Errorf("Clues: %d, want 3", n)
	}
}

func TestCmdrB3WojekInvestigatorNoOpponentWithMoreCardsMakesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Wojek Investigator", "Creature — Angel Detective", cb3WojekInvestigatorOracle, 2, 4)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 0 {
		t.Errorf("Clues: %d, want 0", n)
	}
}

// --- Liliana's Reaver ---------------------------------------------------

func TestCmdrB4LilianasReaverDiscardsAndMakesATappedZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := b12Push(g, me.ID, "Liliana's Reaver", "Creature — Zombie", cb4LilianasReaverOracle, 4, 3)
	assertKeywords(t, g, reaver, "deathtouch")
	hand := opp.Hand.Size()
	attackWith(t, g, opp.ID, reaver)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := hand - opp.Hand.Size(); got != 1 {
		t.Errorf("the defender discarded %d cards, want 1", got)
	}
	if n, tapped := countTokensNamed(g, me.ID, "Zombie"); n != 1 || tapped != 1 {
		t.Errorf("Zombie tokens %d (tapped %d), want one tapped", n, tapped)
	}
}

func TestCmdrB4LilianasReaverMakesTheZombieEvenWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := b12Push(g, me.ID, "Liliana's Reaver", "Creature — Zombie", cb4LilianasReaverOracle, 4, 3)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	attackWith(t, g, opp.ID, reaver)
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Zombie"); n != 1 || tapped != 1 {
		t.Errorf("Zombie tokens %d (tapped %d), want one tapped", n, tapped)
	}
}

// --- Marchesa's Decree --------------------------------------------------

func TestCmdrB4MarchesasDecreeMakesYouMonarchAndDrainsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	castHoldCreature(t, g, "Marchesa's Decree", "Enchantment", cb4MarchesasDecreeOracle, 0, 0)
	if g.Monarch != me.ID {
		t.Fatalf("the monarch is %s, want the caster", g.Monarch)
	}
	a := b12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Attacker C", "Creature — Bear", 2, 2)
	life, thirdLife := opp.Life, third.Life
	advanceToDeclareAttackersOf(t, g, 1)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(c, third.ID); err != nil {
		t.Fatalf("DeclareAttacker at a third player: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the attackers' controller lost %d life, want 2 (one per creature that attacked me)", got)
	}
	if third.Life != thirdLife {
		t.Errorf("the third player's life changed: %d", third.Life-thirdLife)
	}
}

func TestCmdrB4MarchesasDecreeTriggersForAttacksOnYourPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Marchesa's Decree", "Enchantment", cb4MarchesasDecreeOracle, 0, 0)
	walker := p1g2Planeswalker(g, me.ID, 5)
	a := b12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	life := opp.Life
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(a, walker); err != nil {
		t.Fatalf("DeclareAttacker at the walker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("lost %d life, want 1", got)
	}
}

// --- Fall from Favor ----------------------------------------------------

func TestCmdrB3FallFromFavorTapsMakesMonarchAndLocksUntapUnlessMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fall from Favor", "Enchantment — Aura", cb3FallFromFavorOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if !cb3Tapped(g, theirs) {
		t.Fatal("the enchanted creature was not tapped")
	}
	if g.Monarch != me.ID {
		t.Fatalf("the monarch is %s, want the Aura's controller", g.Monarch)
	}
	// Its controller's untap step: not the monarch, so it stays tapped.
	advanceToUpkeepOf(t, g, 1)
	if !cb3Tapped(g, theirs) {
		t.Fatal("the creature untapped while its controller is not the monarch")
	}
	// Their controller takes the crown; the next untap step frees it.
	g.WithWriteLock(func() { g.Monarch = opp.ID })
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOf(t, g, 1)
	if cb3Tapped(g, theirs) {
		t.Error("the creature stayed tapped although its controller is the monarch")
	}
}

func TestCmdrB3FallFromFavorRefusesANonCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Fall from Favor", "Enchantment — Aura", cb3FallFromFavorOracle, b16TargetCard(rock)); err == nil {
		t.Error("an Aura that enchants a creature was cast on an artifact")
	}
	_ = me
}

// --- Launch Mishap ------------------------------------------------------

func TestCmdrB3LaunchMishapCountersACreatureSpellAndMakesAThopter(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	interrupter := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	spell := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: spell, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	mishap := uuid.New()
	interrupter.Hand.PushTop(game.Card{InstanceID: mishap, Name: "Launch Mishap", TypeLine: "Instant",
		OracleID: cb3LaunchMishapOracle, ManaCost: "{2}{U}", Owner: interrupter.ID, Controller: interrupter.ID})
	if err := g.CastSpell(interrupter.ID, mishap, game.CastSpellParams{Targets: b16TargetCard(spell)}); err != nil {
		t.Fatalf("cast Launch Mishap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) || !caster.Graveyard.Contains(spell) {
		t.Error("the creature spell was not countered")
	}
	if n, _ := countTokensNamed(g, interrupter.ID, "Thopter"); n != 1 {
		t.Errorf("the counterspell's caster has %d Thopters, want 1", n)
	}
	if n, _ := countTokensNamed(g, caster.ID, "Thopter"); n != 0 {
		t.Errorf("the countered spell's caster has %d Thopters", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && IsToken(c) {
			assertKeywords(t, g, c.InstanceID, "flying")
			if !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || len(c.Colors) != 0 {
				t.Errorf("the Thopter is wrong: %+v", c)
			}
		}
	}
}

func TestCmdrB3LaunchMishapRefusesANoncreatureNonPlaneswalkerSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	interrupter := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	spell := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: spell, Name: "Divination", TypeLine: "Sorcery", Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	mishap := uuid.New()
	interrupter.Hand.PushTop(game.Card{InstanceID: mishap, Name: "Launch Mishap", TypeLine: "Instant",
		OracleID: cb3LaunchMishapOracle, Owner: interrupter.ID, Controller: interrupter.ID})
	if err := g.CastSpell(interrupter.ID, mishap, game.CastSpellParams{Strict: true, Targets: b16TargetCard(spell)}); err == nil {
		t.Error("Launch Mishap targeted a sorcery spell")
	}
}

// --- Research Thief -----------------------------------------------------

func TestCmdrB3ResearchThiefDrawsForArtifactCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Research Thief", "Artifact Creature — Moonfolk Wizard", cb3ResearchThiefOracle, 3, 3)
	robot := b12Creature(g, me.ID, "Robot", "Artifact Creature — Construct", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Fatalf("a nonartifact creature connecting drew %d", got)
	}
	// The next turn of mine: the artifact creature connects.
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	hand = me.Hand.Size()
	attackWith(t, g, opp.ID, robot)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got < 1 {
		t.Errorf("an artifact creature connecting drew %d, want 1", got)
	}
}

func TestCmdrB3ResearchThiefIsAnArtifactCreatureWithFlashAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thief := b12Push(g, me.ID, "Research Thief", "Artifact Creature — Moonfolk Wizard", cb3ResearchThiefOracle, 3, 3)
	assertKeywords(t, g, thief, "flash", "flying")
}

// --- Shimmer Dragon -----------------------------------------------------

func TestCmdrB3ShimmerDragonHexproofNeedsFourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := b12Push(g, me.ID, "Shimmer Dragon", "Creature — Dragon", cb3ShimmerDragonOracle, 5, 6)
	assertKeywords(t, g, dragon, "flying")
	for i := 0; i < 3; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	if containsString(effectiveAbilities(t, g, dragon), "hexproof") {
		t.Fatal("hexproof with three artifacts")
	}
	b12Permanent(g, me.ID, "Fourth Rock", "Artifact")
	assertKeywords(t, g, dragon, "hexproof")
}

func TestCmdrB3ShimmerDragonTapsTwoArtifactsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	dragon := b12Push(g, me.ID, "Shimmer Dragon", "Creature — Dragon", cb3ShimmerDragonOracle, 5, 6)
	a := b12Permanent(g, me.ID, "Rock A", "Artifact")
	b := b12Permanent(g, me.ID, "Rock B", "Artifact")
	hand := me.Hand.Size()
	if err := dskRT2Activate(g, me.ID, dragon, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated tapping one artifact")
	}
	if err := dskRT2Activate(g, me.ID, dragon, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if !cb3Tapped(g, a) || !cb3Tapped(g, b) {
		t.Error("the artifacts were not tapped")
	}
	if cb3Tapped(g, dragon) {
		t.Error("the Dragon was tapped")
	}
	// Both are tapped now: they cannot pay again.
	if err := dskRT2Activate(g, me.ID, dragon, 0, "", game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Error("tapped artifacts paid the cost")
	}
}

// --- Tamiyo's Logbook ---------------------------------------------------

func TestCmdrB3TamiyosLogbookCostsOneLessPerOtherArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	book := b12Push(g, me.ID, "Tamiyo's Logbook", "Artifact — Book", cb3TamiyosLogbookOracle, 0, 0)
	b12Permanent(g, me.ID, "Rock A", "Artifact")
	b12Permanent(g, me.ID, "Rock B", "Artifact")
	hand := me.Hand.Size()
	// {5}{U} less two: {3}{U}. Three mana is not enough.
	cb3Fund(g, me, "U", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for three mana")
	}
	cb3Fund(g, me, "U", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {3}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestCmdrB3TamiyosLogbookStillNeedsBlueWithManyArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	book := b12Push(g, me.ID, "Tamiyo's Logbook", "Artifact — Book", cb3TamiyosLogbookOracle, 0, 0)
	for i := 0; i < 8; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	cb3Fund(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without blue mana")
	}
	cb3Fund(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {U}: %v", err)
	}
}

// --- Broodstar ----------------------------------------------------------

func TestCmdrB3BroodstarIsAsBigAsYourArtifactsAndCostsLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	b12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	star := cb3Hand(g, "Broodstar", "Creature — Beast", "{8}{U}{U}", cb3BroodstarOracle)
	advanceToMain(t, g)
	// {8}{U}{U} less three: {5}{U}{U}. Six mana is not enough.
	cb3Fund(g, me, "U", "U", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, star, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for six mana")
	}
	cb3Fund(g, me, "U", "U", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, star, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {5}{U}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, star), effectiveToughness(t, g, star); p != 3 || tg != 3 {
		t.Errorf("Broodstar is %d/%d, want 3/3 (three artifacts; the opponent's does not count)", p, tg)
	}
	assertKeywords(t, g, star, "flying")
	b12Permanent(g, me.ID, "Another Rock", "Artifact")
	if p := effectivePower(t, g, star); p != 4 {
		t.Errorf("Broodstar is %d/*, want 4 after a fourth artifact", p)
	}
}

// --- Eternal Taskmaster -------------------------------------------------

func TestCmdrB3EternalTaskmasterEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Eternal Taskmaster", "Creature — Zombie", cb3EternalTaskmasterOracle, 2, 3)
	if !cb3Tapped(g, id) {
		t.Error("Eternal Taskmaster entered untapped")
	}
}

func cb3TaskmasterAttack(t *testing.T) (g *game.Game, me *game.Player, dead uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	taskmaster := b12Push(g, me.ID, "Eternal Taskmaster", "Creature — Zombie", cb3EternalTaskmasterOracle, 2, 3)
	dead = gyCard(me, "Dead Bear", "Creature — Bear", "", "")
	gyCard(me, "Dead Spell", "Sorcery", "", "")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(taskmaster, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		b40PickCard(t, g, me.ID, dead)
	}
	passPriorityAroundTable(t, g)
	return g, me, dead
}

func TestCmdrB3EternalTaskmasterPaysToReturnACreatureCard(t *testing.T) {
	g, me, dead := cb3TaskmasterAttack(t)
	b06AddMana(me, "B", "C", "C")
	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the attack does not ask about {2}{B}")
	}
	if ask.PayCost != "{2}{B}" {
		t.Errorf("the optional cost is %s", ask.PayCost)
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !me.Hand.Contains(dead) {
		t.Error("paying returns the creature card to hand")
	}
}

func TestCmdrB3EternalTaskmasterDecliningReturnsNothing(t *testing.T) {
	g, me, dead := cb3TaskmasterAttack(t)
	b06AddMana(me, "B", "C", "C")
	ask := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("the attack does not ask about {2}{B}")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if me.Hand.Contains(dead) || !me.Graveyard.Contains(dead) {
		t.Error("declining moved the card")
	}
}

// --- Mutilate -----------------------------------------------------------

func TestCmdrB4MutilateShrinksEveryCreatureBySwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	awbSwamps(g, me.ID, 3)
	awbSwamps(g, opp.ID, 5) // not mine: they do not count
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	castCatalogSpell(t, g, "Mutilate", "Sorcery", cb4MutilateOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("a 2/3 survived -3/-3")
	}
	if p, tg := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tg != 2 {
		t.Errorf("the 5/5 is %d/%d, want 2/2", p, tg)
	}
	// The count was locked as it resolved: a Swamp later changes nothing.
	awbSwamps(g, me.ID, 2)
	if p := effectivePower(t, g, big); p != 2 {
		t.Errorf("later Swamps changed the shrink: power %d", p)
	}
}

func TestCmdrB4MutilateWithNoSwampsDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Mutilate", "Sorcery", cb4MutilateOracle, nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
}

// --- Necrotic Hex -------------------------------------------------------

func TestCmdrB4NecroticHexEachPlayerSacrificesSixAndYouMakeSixTappedZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	var mine, theirs1, theirs2 []uuid.UUID
	for i := 0; i < 3; i++ {
		mine = append(mine, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	}
	for i := 0; i < 8; i++ {
		theirs1 = append(theirs1, b12Creature(g, o1.ID, "Theirs", "Creature — Bear", 2, 2))
	}
	for i := 0; i < 6; i++ {
		theirs2 = append(theirs2, b12Creature(g, o2.ID, "Theirs", "Creature — Bear", 2, 2))
	}
	land := b12Permanent(g, o3.ID, "Their Land", "Basic Land — Swamp")
	castCatalogSpell(t, g, "Necrotic Hex", "Sorcery", cb4NecroticHexOracle, nil)
	passPriorityAroundTable(t, g)
	for round := 0; round < 12; round++ {
		asked := false
		for _, p := range g.Seats {
			if c := sacrificeChoiceFor(g, p.ID); c != nil {
				asked = true
				opts := sacrificeOptions(c)
				if len(opts) < 6 {
					t.Fatalf("%s offered %d creatures", p.Name, len(opts))
				}
				answerSacrifice(t, g, p.ID, opts[:6]...)
			}
		}
		passPriorityAroundTable(t, g)
		if !asked {
			break
		}
	}
	survivors := func(ids []uuid.UUID) int {
		n := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if got := survivors(mine); got != 0 {
		t.Errorf("%d of my 3 creatures survived (fewer than six: all go)", got)
	}
	if got := survivors(theirs1); got != 2 {
		t.Errorf("%d of the 8 survived, want 2", got)
	}
	if got := survivors(theirs2); got != 0 {
		t.Errorf("%d of the 6 survived, want 0", got)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a land was sacrificed")
	}
	if n, tapped := countTokensNamed(g, me.ID, "Zombie"); n != 6 || tapped != 6 {
		t.Errorf("Zombie tokens %d (tapped %d), want six tapped", n, tapped)
	}
	if n, _ := countTokensNamed(g, o1.ID, "Zombie"); n != 0 {
		t.Errorf("an opponent got %d Zombies", n)
	}
}

// --- Razorlash Transmogrant ---------------------------------------------

func TestCmdrB4RazorlashCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	razor := b12Push(g, opp.ID, "Razorlash Transmogrant", "Artifact Creature — Zombie", cb4RazorlashOracle, 3, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(razor, attacker); err == nil {
		t.Error("Razorlash Transmogrant blocked")
	}
}

func TestCmdrB4RazorlashReturnsFromTheGraveyardWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	razor := pushCatalogGraveyardCard(me, "Razorlash Transmogrant", "Artifact Creature — Zombie", cb4RazorlashOracle, 3, 1)
	if err := dskRT2Activate(g, me.ID, razor, 0, "{B}{B}{C}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for five mana")
	}
	if err := dskRT2Activate(g, me.ID, razor, 0, "{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate for {4}{B}{B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(razor) {
		t.Fatal("Razorlash did not return")
	}
	if got := countersOn(g, razor, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if c, ok := cb3Bf(g, razor); !ok || c.CurrentPower() != 4 {
		t.Errorf("current power %d, want 4 (3 and the counter)", c.CurrentPower())
	}
}

func TestCmdrB4RazorlashCostsFourLessAgainstFourNonbasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	razor := pushCatalogGraveyardCard(me, "Razorlash Transmogrant", "Artifact Creature — Zombie", cb4RazorlashOracle, 3, 1)
	for i := 0; i < 3; i++ {
		b12Permanent(g, opp.ID, "Nonbasic", "Land")
	}
	b12Permanent(g, opp.ID, "Basic", "Basic Land — Swamp")
	if err := dskRT2Activate(g, me.ID, razor, 0, "{B}{B}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for {B}{B} against three nonbasic lands")
	}
	b12Permanent(g, opp.ID, "Fourth Nonbasic", "Land")
	cb3Fund(g, me)
	if err := dskRT2Activate(g, me.ID, razor, 0, "{B}{B}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate for {B}{B} against four nonbasic lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(razor) {
		t.Error("Razorlash did not return")
	}
}
