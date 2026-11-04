package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b12_cards_test.go — Teenage Mutant Ninja Turtles, batch 12 (#4457):
// the sneak cards (CR 702.190). Each test plays the real catalog entry and
// asserts what its text says, the sneak cast included.

const (
	tmtB12MichelangeloOracle = "1939679b-ea2e-4a75-b3ce-d7af68437484"
	tmtB12NewGenOracle       = "e7ae3ee7-824f-4dcb-8f28-7de99e4d8c3e"
	tmtB12OrokuOracle        = "d14f620c-0fbd-485c-9fe4-34949e742147"
	tmtB12RaphTechOracle     = "f58a5f8b-254c-4206-bdbe-55a23798ed2d"
	tmtB12NightwatcherOracle = "8c1e5b25-ee5a-4ea9-a715-08dd1a4f51c4"
	tmtB12SharkOracle        = "dabd0867-ee3f-40cb-8314-b4a3966f8b24"
	tmtB12ShredTechOracle    = "d92714c0-4e22-47f9-a4d4-405894e734d0"
	tmtB12UnrelentingOracle  = "59d96fc0-6580-43c2-8c6a-dcf983644561"
	tmtB12SplinterTechOracle = "532ee9c3-391c-44db-8ed5-a1d28fd7f876"
	tmtB12SplinterOracle     = "41fbdaa2-7a7a-44d5-a8f4-77624350eedd"
	tmtB12RoninTechOracle    = "6a51bb21-0c01-4ad8-99d7-17685db90b69"
	tmtB12KunoichiOracle     = "0f35cf85-6783-4d69-b5e6-a81f7c1f21e7"
)

func TestTmtB12EveryCardIsRegisteredFullWithItsSneakCost(t *testing.T) {
	for oracle, want := range map[string]struct{ name, sneak string }{
		tmtB12MichelangeloOracle: {"Michelangelo, Improviser", "{2}{G}{G}"},
		tmtB12NewGenOracle:       {"New Generation's Technique", "{2}{G}"},
		tmtB12OrokuOracle:        {"Oroku Saki, Shredder Rising", "{1}{B}"},
		tmtB12RaphTechOracle:     {"Raphael's Technique", "{2}{R}"},
		tmtB12NightwatcherOracle: {"Raphael, the Nightwatcher", "{1}{R}{R}"},
		tmtB12SharkOracle:        {"Shark Shredder, Killer Clone", "{3}{B}{B}"},
		tmtB12ShredTechOracle:    {"Shredder's Technique", "{B}"},
		tmtB12UnrelentingOracle:  {"Shredder, Unrelenting", "{3}{B}"},
		tmtB12SplinterTechOracle: {"Splinter's Technique", "{1}{B}"},
		tmtB12SplinterOracle:     {"Splinter, Hamato Yoshi", "{B}"},
		tmtB12RoninTechOracle:    {"The Last Ronin's Technique", "{1}{W}"},
		tmtB12KunoichiOracle:     {"Turncoat Kunoichi", "{2}{W}{B}"},
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want.name)
			continue
		}
		if spec.Name != want.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", want.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		alt := game.AlternativeCostByKey(oracle, "sneak")
		if alt == nil || !alt.Sneak || alt.ManaCost != want.sneak || alt.ReturnToHand == nil {
			t.Errorf("%s: sneak offer %+v, want Sneak %s returning an unblocked attacker", want.name, alt, want.sneak)
		}
	}
}

// tmtB12Board swings my 1/1 at seat 1 and finishes blocks with none, so the
// cursor is in my declare blockers step with the attacker unblocked.
func tmtB12Board(t *testing.T, g *game.Game) (attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	attacker = b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("left the declare blockers step (at %s)", g.Turn.Step)
	}
	return attacker, me, opp
}

// tmtB12SneakCast puts a card in my hand, floats `mana` and casts it for its
// sneak cost, returning `attacker`; the returned creature must be in hand.
func tmtB12SneakCast(t *testing.T, g *game.Game, attacker uuid.UUID, c game.Card, mana string, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	floatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{attacker}, Targets: targets,
	}); err != nil {
		t.Fatalf("sneak cast of %s: %v", c.Name, err)
	}
	if !me.Hand.Contains(attacker) {
		t.Fatalf("%s: the unblocked attacker was not returned to hand as the cost", c.Name)
	}
	return c.InstanceID
}

// tmtB12Sneaked checks a permanent sneaked in is tapped and attacking opp.
func tmtB12Sneaked(t *testing.T, g *game.Game, id, defender uuid.UUID) {
	t.Helper()
	c := tmtB1Card(t, g, id)
	if !c.Tapped || c.AttackingTarget != defender {
		t.Errorf("%s: tapped=%v attacking %s, want tapped and attacking %s", c.Name, c.Tapped, c.AttackingTarget, defender)
	}
}

// --- Michelangelo, Improviser -------------------------------------------

func TestTmtB12MichelangeloPutsACreatureAndALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mikey := b27Push(g, me.ID, "Michelangelo, Improviser", "Legendary Creature — Mutant Ninja Turtle", tmtB12MichelangeloOracle, "{3}{G}", 4, 4, "G")
	bearA := handCardForTest(me, "Bear A", "Creature — Bear", "")
	bearB := handCardForTest(me, "Bear B", "Creature — Bear", "")
	forest := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	dealCombatDamageToPlayer(g, mikey, opp.ID, 4)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put offer after Michelangelo's combat damage")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bearA, bearB}); err == nil {
		t.Fatal("two creature cards were accepted")
	}
	answerChooseCards(t, g, me.ID, bearA, forest)
	if !g.Battlefield.Contains(bearA) || !g.Battlefield.Contains(forest) || g.Battlefield.Contains(bearB) {
		t.Error("want the creature and the land on the battlefield, the second creature still in hand")
	}
}

func TestTmtB12MichelangeloSneaksInAttacking(t *testing.T) {
	g := newCatalogGame(t)
	attacker, _, opp := tmtB12Board(t, g)
	mikey := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Michelangelo, Improviser", TypeLine: "Legendary Creature — Mutant Ninja Turtle", ManaCost: "{3}{G}",
		OracleID: tmtB12MichelangeloOracle, Power: 4, Toughness: 4,
	}, "{G}{G}{G}{G}")
	passPriorityAroundTable(t, g)
	tmtB12Sneaked(t, g, mikey, opp.ID)
}

// --- New Generation's Technique -------------------------------------------

func TestTmtB12NewGenerationsTechniqueSneakedFetchesTwoBasicsTapped(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me, _ := tmtB12Board(t, g)
	g.WithWriteLock(func() {
		seedSearchLibrary(me,
			game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"},
			game.Card{Name: "Temple", TypeLine: "Land"})
	})
	tmtB12SneakCast(t, g, attacker, game.Card{Name: "New Generation's Technique", TypeLine: "Sorcery", ManaCost: "{3}{G}", OracleID: tmtB12NewGenOracle}, "{G}{G}{G}")
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Temple") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest", "Swamp")
	for _, name := range []string{"Forest", "Swamp"} {
		ids := tmtB1Named(g, name)
		if len(ids) != 1 {
			t.Fatalf("%d %s on the battlefield, want 1", len(ids), name)
		}
		if !tmtB1Card(t, g, ids[0]).Tapped {
			t.Errorf("%s entered untapped", name)
		}
	}
}

// --- Oroku Saki, Shredder Rising --------------------------------------------

func TestTmtB12OrokuSakiDrawsAndLosesOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saki := b27Push(g, me.ID, "Oroku Saki, Shredder Rising", "Legendary Creature — Human Ninja", tmtB12OrokuOracle, "{2}{B}", 3, 1, "B")
	seedLibrary(me, "One", "Two")
	hand, life := me.Hand.Size(), me.Life
	dealCombatDamageToPlayer(g, saki, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
}

// Sneaked in, he is unblocked and his damage draws a card.
func TestTmtB12OrokuSakiSneakedInDealsDamageAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me, opp := tmtB12Board(t, g)
	seedLibrary(me, "One", "Two")
	saki := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Oroku Saki, Shredder Rising", TypeLine: "Legendary Creature — Human Ninja", ManaCost: "{2}{B}",
		OracleID: tmtB12OrokuOracle, Power: 3, Toughness: 1,
	}, "{B}{B}")
	passPriorityAroundTable(t, g)
	tmtB12Sneaked(t, g, saki, opp.ID)
	hand, life, oppLife := me.Hand.Size(), me.Life, opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("defender life %d, want %d", opp.Life, oppLife-3)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
}

// --- Raphael's Technique ------------------------------------------------------

func TestTmtB12RaphaelsTechniqueEachPlayerMayWheel(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me, opp := tmtB12Board(t, g)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			seedLibrary(p, "L1", "L2", "L3", "L4", "L5", "L6", "L7", "L8")
		}
	})
	tmtB12SneakCast(t, g, attacker, game.Card{Name: "Raphael's Technique", TypeLine: "Instant", ManaCost: "{4}{R}{R}", OracleID: tmtB12RaphTechOracle}, "{R}{R}{R}")
	oppHand := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		oppHand[i] = p.Hand.Size()
	}
	passPriorityAroundTable(t, g)
	// APNAP: I am the active player and am asked first. My yes changes
	// nothing until every player has answered (CR 101.4).
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("the active player is not asked first")
	}
	answerConfirmFor(t, g, me, true)
	if !me.Hand.Contains(attacker) {
		t.Fatal("my hand was discarded before the other players answered")
	}
	for i := 0; i < len(g.Seats); i++ {
		for _, p := range g.Seats {
			if latestConfirmFor(g, p.ID) != nil {
				answerConfirmFor(t, g, p, p.ID == me.ID)
			}
		}
	}
	if me.Hand.Size() != 7 || me.Hand.Contains(attacker) {
		t.Errorf("my hand %d (rat kept %v), want a fresh seven", me.Hand.Size(), me.Hand.Contains(attacker))
	}
	if !me.Graveyard.Contains(attacker) {
		t.Error("the returned rat was not discarded with my hand")
	}
	if opp.Hand.Size() != oppHand[1] {
		t.Errorf("the opponent who declined has %d cards, want %d", opp.Hand.Size(), oppHand[1])
	}
}

// --- Raphael, the Nightwatcher ------------------------------------------------

func TestTmtB12NightwatcherGivesAttackersDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raph := b27Push(g, me.ID, "Raphael, the Nightwatcher", "Legendary Creature — Mutant Ninja Turtle", tmtB12NightwatcherOracle, "{2}{R}{R}", 2, 3, "R")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if effectiveAbilitiesContain(t, g, bear, "double strike") {
		t.Fatal("double strike before attacking")
	}
	declareAttack(t, g, opp.ID, bear)
	assertKeywords(t, g, bear, "double strike")
	for _, id := range []uuid.UUID{raph, home, theirs} {
		if effectiveAbilitiesContain(t, g, id, "double strike") {
			t.Errorf("%s has double strike without attacking", tmtB1Card(t, g, id).Name)
		}
	}
}

func TestTmtB12NightwatcherSneakedInHasDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	attacker, _, opp := tmtB12Board(t, g)
	raph := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Raphael, the Nightwatcher", TypeLine: "Legendary Creature — Mutant Ninja Turtle", ManaCost: "{2}{R}{R}",
		OracleID: tmtB12NightwatcherOracle, Power: 2, Toughness: 3,
	}, "{R}{R}{R}")
	passPriorityAroundTable(t, g)
	tmtB12Sneaked(t, g, raph, opp.ID)
	assertKeywords(t, g, raph, "double strike")
}

// --- Shark Shredder, Killer Clone ---------------------------------------------

func TestTmtB12SharkShredderReanimatesFromTheDamagedPlayersGraveyardAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	shark := b27Push(g, me.ID, "Shark Shredder, Killer Clone", "Legendary Creature — Shark Octopus Ninja", tmtB12SharkOracle, "{2}{B}{B}", 4, 4, "B")
	var theirs, elsewhere, mine uuid.UUID
	g.WithWriteLock(func() {
		theirs = pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
		elsewhere = pushGraveyardCardTyped(other, "Other Dead Bear", "Creature — Bear")
		mine = pushGraveyardCardTyped(me, "My Dead Bear", "Creature — Bear")
	})
	declareAttack(t, g, opp.ID, shark)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	advanceTo(t, g, game.StepFirstStrikeDamage)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) == nil && triggerOnStack(g, shark) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Shark Shredder's trigger")
	}
	for _, bad := range []uuid.UUID{elsewhere, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("a card outside the damaged player's graveyard (%s) was accepted", bad)
		}
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: theirs}, true)
	c := tmtB1Card(t, g, theirs)
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %s owner %s, want under my control and still theirs", c.Controller, c.Owner)
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
	}
}

func TestTmtB12SharkShredderSneaksIn(t *testing.T) {
	g := newCatalogGame(t)
	attacker, _, opp := tmtB12Board(t, g)
	shark := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Shark Shredder, Killer Clone", TypeLine: "Legendary Creature — Shark Octopus Ninja", ManaCost: "{2}{B}{B}",
		OracleID: tmtB12SharkOracle, Power: 4, Toughness: 4,
	}, "{B}{B}{B}{B}{B}")
	passPriorityAroundTable(t, g)
	tmtB12Sneaked(t, g, shark, opp.ID)
	assertKeywords(t, g, shark, "first strike")
}

// --- Shredder's Technique -----------------------------------------------------

func TestTmtB12ShreddersTechniqueSneakedDestroysAnEnchantmentForTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	attacker, me, _ := tmtB12Board(t, g)
	life := me.Life
	tmtB12SneakCast(t, g, attacker, game.Card{Name: "Shredder's Technique", TypeLine: "Sorcery", ManaCost: "{2}{B}", OracleID: tmtB12ShredTechOracle},
		"{B}", game.TargetRef{Kind: game.TargetCard, ID: aura})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the enchantment was not destroyed")
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestTmtB12ShreddersTechniqueDestroysACreatureForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	castCatalogSpell(t, g, "Shredder's Technique", "Sorcery", tmtB12ShredTechOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not destroyed")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: no enchantment was destroyed", me.Life, life)
	}
}

func TestTmtB12ShreddersTechniqueChargesNothingWhenTheEnchantmentSurvives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sturdy Enchantment", TypeLine: "Enchantment",
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	life := me.Life
	castCatalogSpell(t, g, "Shredder's Technique", "Sorcery", tmtB12ShredTechOracle, cardRefs(aura))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the indestructible enchantment was destroyed")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d: nothing was destroyed", me.Life, life)
	}
}

// --- Shredder, Unrelenting ----------------------------------------------------

func TestTmtB12ShredderUnrelentingSneakedInGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ally := b12Creature(g, me.ID, "Ally", "Creature — Human Ninja", 2, 2)
	attacker, _, opp := tmtB12Board(t, g)
	shredder := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Shredder, Unrelenting", TypeLine: "Legendary Creature — Human Ninja", ManaCost: "{4}{B}",
		OracleID: tmtB12UnrelentingOracle, Power: 6, Toughness: 4,
	}, "{B}{B}{B}{B}")
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ally}, true)
	tmtB12Sneaked(t, g, shredder, opp.ID)
	assertKeywords(t, g, ally, "deathtouch")
	assertKeywords(t, g, shredder, "deathtouch")
	if n := attackEventsFor(g, shredder); n != 0 {
		t.Errorf("%d attack events for a sneaked Shredder, want 0", n)
	}
}

func TestTmtB12ShredderUnrelentingAttackTriggerTargetsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := b27Push(g, me.ID, "Shredder, Unrelenting", "Legendary Creature — Human Ninja", tmtB12UnrelentingOracle, "{4}{B}", 6, 4, "B")
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, shredder)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	for _, bad := range []uuid.UUID{shredder, theirs} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as another creature I control", tmtB1Card(t, g, bad).Name)
		}
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ally}, true)
	assertKeywords(t, g, ally, "deathtouch")
	endTurn(t, g)
	if effectiveAbilitiesContain(t, g, ally, "deathtouch") {
		t.Error("deathtouch outlasted the turn")
	}
}

// --- Splinter's Technique -----------------------------------------------------

func TestTmtB12SplintersTechniqueSneakedTutorsAnyCard(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me, _ := tmtB12Board(t, g)
	g.WithWriteLock(func() {
		seedSearchLibrary(me, game.Card{Name: "Wanted", TypeLine: "Artifact"}, game.Card{Name: "Other", TypeLine: "Sorcery"})
	})
	tmtB12SneakCast(t, g, attacker, game.Card{Name: "Splinter's Technique", TypeLine: "Sorcery", ManaCost: "{3}{B}", OracleID: tmtB12SplinterTechOracle}, "{B}{B}")
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Wanted")
	found := false
	for _, c := range me.Hand.Cards {
		found = found || c.Name == "Wanted"
	}
	if !found {
		t.Error("the searched card is not in hand")
	}
}

// --- Splinter, Hamato Yoshi ---------------------------------------------------

func TestTmtB12SplinterPumpsOtherNinjasYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	splinter := b27Push(g, me.ID, "Splinter, Hamato Yoshi", "Legendary Creature — Mutant Ninja Rat", tmtB12SplinterOracle, "{1}{B}", 1, 3, "B")
	ninja := b12Creature(g, me.ID, "Foot Soldier", "Creature — Human Ninja", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirNinja := b12Creature(g, opp.ID, "Their Ninja", "Creature — Human Ninja", 2, 2)
	for id, want := range map[uuid.UUID]int{splinter: 1, ninja: 3, bear: 2, theirNinja: 2} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("%s power %d, want %d", tmtB1Card(t, g, id).Name, got, want)
		}
	}
	if got := effectiveToughness(t, g, ninja); got != 3 {
		t.Errorf("Ninja toughness %d, want 3", got)
	}
	assertKeywords(t, g, splinter, "menace")
}

func TestTmtB12SplinterSneaksIn(t *testing.T) {
	g := newCatalogGame(t)
	attacker, _, opp := tmtB12Board(t, g)
	splinter := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Splinter, Hamato Yoshi", TypeLine: "Legendary Creature — Mutant Ninja Rat", ManaCost: "{1}{B}",
		OracleID: tmtB12SplinterOracle, Power: 1, Toughness: 3,
	}, "{B}")
	passPriorityAroundTable(t, g)
	tmtB12Sneaked(t, g, splinter, opp.ID)
}

// --- The Last Ronin's Technique -----------------------------------------------

func TestTmtB12LastRoninsTechniqueCastNormallyMakesThreeIdleSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "The Last Ronin's Technique", "Instant", tmtB12RoninTechOracle, nil)
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Ninja Turtle Spirit")
	if len(ids) != 3 {
		t.Fatalf("%d spirits, want 3", len(ids))
	}
	for _, id := range ids {
		c := tmtB1Card(t, g, id)
		if c.Tapped || c.AttackingTarget != uuid.Nil || c.Controller != me.ID {
			t.Errorf("spirit tapped=%v attacking %s controller %s, want untapped, not attacking, mine", c.Tapped, c.AttackingTarget, c.Controller)
		}
		if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" || !c.HasSubtype("Turtle") || !c.HasSubtype("Spirit") || !c.HasSubtype("Ninja") {
			t.Errorf("spirit %+v, want a 1/1 white Ninja Turtle Spirit", c)
		}
	}
}

func TestTmtB12LastRoninsTechniqueSneakedSpiritsEnterTappedAndAttacking(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me, opp := tmtB12Board(t, g)
	tmtB12SneakCast(t, g, attacker, game.Card{Name: "The Last Ronin's Technique", TypeLine: "Instant", ManaCost: "{3}{W}", OracleID: tmtB12RoninTechOracle}, "{W}{W}")
	passAndSettleAttackSplit(t, g, me.ID)
	ids := tmtB1Named(g, "Ninja Turtle Spirit")
	if len(ids) != 3 {
		t.Fatalf("%d spirits, want 3", len(ids))
	}
	for _, id := range ids {
		tmtB12Sneaked(t, g, id, opp.ID)
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("defender life %d, want %d after three unblocked spirits", opp.Life, life-3)
	}
}

// --- Turncoat Kunoichi --------------------------------------------------------

func TestTmtB12TurncoatKunoichiCastNormallyExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	kunoichi := castCatalogSpell(t, g, "Turncoat Kunoichi", "Creature — Mutant Ninja Fox", tmtB12KunoichiOracle, nil)
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if !g.Exile.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	b27Kill(g, kunoichi)
	g.RunStateChecksForTest()
	assistCardSettle(t, g, me.ID, nil, true)
	if len(tmtB1Named(g, "Their Bear")) != 1 {
		t.Error("the creature did not return (as a new object) when the Kunoichi left")
	}
}

func TestTmtB12TurncoatKunoichiSneakedExilesForGood(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	attacker, me, _ := tmtB12Board(t, g)
	kunoichi := tmtB12SneakCast(t, g, attacker, game.Card{
		Name: "Turncoat Kunoichi", TypeLine: "Creature — Mutant Ninja Fox", ManaCost: "{2}{W}",
		OracleID: tmtB12KunoichiOracle, Power: 3, Toughness: 2,
	}, "{W}{W}{B}{B}")
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	tmtB12Sneaked(t, g, kunoichi, opp.ID)
	if !g.Exile.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	b27Kill(g, kunoichi)
	g.RunStateChecksForTest()
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Exile.Contains(victim) || len(tmtB1Named(g, "Their Bear")) != 0 {
		t.Error("the creature came back: a sneaked Kunoichi exiles it for good")
	}
}
