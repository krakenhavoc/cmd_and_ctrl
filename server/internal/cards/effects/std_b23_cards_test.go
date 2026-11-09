package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b23_cards_test.go — Standard-set batch 23 (Duskmourn, #4466) and
// Sawblade Skinripper (batch 26). Each test plays the real catalog entry and
// asserts what its text says.

const (
	stdB23FanaticOracle    = "878e5e0d-beb9-4f8d-a640-62140c907e52"
	stdB23MeathookOracle   = "68957dca-df5c-4051-af22-407cb7e47200"
	stdB23MiasmaOracle     = "b1df6475-e1e8-4ef3-a7b2-919d5f121c57"
	stdB23SticktwisterOID  = "33c3a81a-6cea-48c9-b966-67084d96e74f"
	stdB23EgotistOracle    = "e6377683-1c36-4786-bd8d-c62ff2d25767"
	stdB23InfectionOracle  = "659a259d-f75e-4af3-b568-c63225e647de"
	stdB23BeastieOracle    = "7143d82a-931a-4918-bb14-37bc5b9a4d99"
	stdB23RipperOracle     = "0cbe2a00-4303-4ade-98b5-4172d97f15fb"
	stdB23PercussionOracle = "3b74c8ac-9c19-45b0-935f-8630b8da4543"
	stdB23DiversionOracle  = "7130d063-fc1b-408b-97f2-14cb6b55d9cd"
	stdB23SkinripperOracle = "96fd3ee0-84fb-439b-9e2d-b55afe6850a7"
)

func TestStdB23EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB23FanaticOracle:    "Fanatic of the Harrowing",
		stdB23MeathookOracle:   "Meathook Massacre II",
		stdB23MiasmaOracle:     "Miasma Demon",
		stdB23SticktwisterOID:  "Osseous Sticktwister",
		stdB23EgotistOracle:    "Popular Egotist",
		stdB23InfectionOracle:  "Sporogenic Infection",
		stdB23BeastieOracle:    "Bedhead Beastie",
		stdB23RipperOracle:     "Boilerbilges Ripper",
		stdB23PercussionOracle: "Clockwork Percussionist",
		stdB23DiversionOracle:  "Diversion Specialist",
		stdB23SkinripperOracle: "Sawblade Skinripper",
	} {
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

// stdB23Seats is the active player and the three opponents in turn order.
func stdB23Seats(g *game.Game) (me *game.Player, opps []*game.Player) {
	start := g.Turn.ActiveSeat
	me = g.Seats[start]
	for k := 1; k < len(g.Seats); k++ {
		opps = append(opps, g.Seats[(start+k)%len(g.Seats)])
	}
	return me, opps
}

// stdB23EmptyHands empties every hand.
func stdB23EmptyHands(g *game.Game) {
	for _, p := range g.Seats {
		p.Hand.Cards = p.Hand.Cards[:0]
	}
}

// stdB23Mana adds mana to a pool.
func stdB23Mana(t *testing.T, g *game.Game, p *game.Player, cost string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, cost); err != nil {
		t.Fatalf("AddManaForEffect %s: %v", cost, err)
	}
}

// --- Bedhead Beastie ----------------------------------------------------------

func TestStdB23BedheadBeastieMountaincyclingFetchesAMountainAndItHasMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := uuid.New()
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	id, _ := cycleFromHand(t, g, "Bedhead Beastie", "Creature — Beast", stdB23BeastieOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Beastie is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	pick := searchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if hasID(pick.SearchCards, forest) || !hasID(pick.SearchCards, mountain) {
		t.Errorf("the search offers %v, want only the Mountain", pick.SearchCards)
	}
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, []uuid.UUID{mountain}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mountain) {
		t.Error("the Mountain did not come to hand")
	}
	beastie := pushCatalogPermanent(g, me.ID, "Bedhead Beastie", "Creature — Beast", stdB23BeastieOracle, false)
	if !finB3Has(s58p6View(t, g, beastie), "menace") {
		t.Error("the Beastie lacks menace")
	}
}

// --- Fanatic of the Harrowing --------------------------------------------------

func TestStdB23FanaticEveryPlayerDiscardsAndYouDrawIfYouDid(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	mine := handCardForTest(me, "Spare", "Sorcery", "")
	theirs := handCardForTest(opps[0], "Their Spare", "Sorcery", "")
	library := me.Library.Size()
	castHoldCreature(t, g, "Fanatic of the Harrowing", "Creature — Human Cleric", stdB23FanaticOracle, 2, 2)
	answerDiscard(t, g, me.ID, mine)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opps[0].ID) == nil && discardChoiceFor(g, opps[0].ID) == nil {
		t.Fatal("the opponent was not asked to discard")
	}
	answerDiscard(t, g, opps[0].ID, theirs)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) || !opps[0].Graveyard.Contains(theirs) {
		t.Error("both players should have discarded")
	}
	if me.Hand.Size() != 1 || me.Library.Size() != library-1 {
		t.Errorf("hand %d, library %d → %d, want one card drawn", me.Hand.Size(), library, me.Library.Size())
	}
	if opps[0].Hand.Size() != 0 {
		t.Error("the opponent drew; only you do")
	}
}

func TestStdB23FanaticDoesNotDrawWhenYouHadNothingToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	theirs := handCardForTest(opps[1], "Their Spare", "Sorcery", "")
	castHoldCreature(t, g, "Fanatic of the Harrowing", "Creature — Human Cleric", stdB23FanaticOracle, 2, 2)
	answerDiscard(t, g, opps[1].ID, theirs)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0: you discarded no card this way", me.Hand.Size())
	}
	if !opps[1].Graveyard.Contains(theirs) {
		t.Error("the opponent should still discard")
	}
}

// --- Miasma Demon ----------------------------------------------------------------

func TestStdB23MiasmaDemonDiscardsAnyNumberThenShrinksThatManyTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	a := handCardForTest(me, "A", "Sorcery", "")
	b := handCardForTest(me, "B", "Sorcery", "")
	keep := handCardForTest(me, "Keep", "Sorcery", "")
	x := b12Creature(g, opps[0].ID, "X", "Creature — Bear", 2, 2)
	y := b12Creature(g, opps[1].ID, "Y", "Creature — Bear", 2, 2)
	z := b12Creature(g, opps[2].ID, "Z", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, 5, 4)
	answerDiscard(t, g, me.ID, a, b)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) || !me.Hand.Contains(keep) {
		t.Fatal("exactly the two chosen cards should be discarded")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{mshA4Ref(x), mshA4Ref(y), mshA4Ref(z)}); err == nil {
		t.Fatal("three targets were accepted for two cards discarded")
	}
	pickTargetRefs(t, g, me.ID, nil, mshA4Ref(x), mshA4Ref(y))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
		t.Error("the two 2/2s targeted should have died to -2/-2")
	}
	if !g.Battlefield.Contains(z) {
		t.Error("the untargeted creature died")
	}
}

func TestStdB23MiasmaDemonDiscardingNothingMakesNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	a := handCardForTest(me, "A", "Sorcery", "")
	x := b12Creature(g, opps[0].ID, "X", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, 5, 4)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("discarding nothing still asked for targets")
	}
	if !me.Hand.Contains(a) || !g.Battlefield.Contains(x) {
		t.Error("nothing should have happened")
	}
}

func TestStdB23MiasmaDemonHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := pushCatalogPermanent(g, me.ID, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, false)
	if !finB3Has(s58p6View(t, g, demon), "flying") {
		t.Error("Miasma Demon lacks flying")
	}
}

// --- Boilerbilges Ripper -----------------------------------------------------------

func TestStdB23BoilerbilgesRipperSacrificesACreatureOrEnchantmentForTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := b12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	aura := b12Permanent(g, me.ID, "My Ward", "Enchantment")
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := b12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 2, 2)
	ripper := castHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	for name, id := range map[string]uuid.UUID{"an artifact": rock, "a land": land, "the Ripper itself": ripper} {
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{id}); err == nil {
			t.Fatalf("sacrificed %s", name)
		}
	}
	if !hasID(c.ChooseCards, aura) || !hasID(c.ChooseCards, fodder) {
		t.Errorf("the prompt offers %v, want the enchantment and the creature", c.ChooseCards)
	}
	answerOwnPermanents(t, g, me.ID, aura)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the sacrifice happens before the reflexive trigger targets")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage did not kill the 2/2")
	}
}

func TestStdB23BoilerbilgesRipperTwoDamageCanHitAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	before := opps[2].Life
	castHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	answerOwnPermanents(t, g, me.ID, fodder)
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(latestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opps[2].ID}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if opps[2].Life != before-2 {
		t.Errorf("life %d → %d, want -2", before, opps[2].Life)
	}
}

func TestStdB23BoilerbilgesRipperDeclinedDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := b12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a declined sacrifice still asked for a target")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("nothing should have happened")
	}
}

// --- Clockwork Percussionist ---------------------------------------------------------

func TestStdB23ClockworkPercussionistDiesExilesTheTopCardPlayableUntilYourNextTurnEnds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	loot := topCard(me, "Loot", "Sorcery", "{R}")
	mon := pushCatalogPermanent(g, me.ID, "Clockwork Percussionist", "Artifact Creature — Monkey Toy", stdB23PercussionOracle, false)
	if !finB3Has(s58p6View(t, g, mon), "haste") {
		t.Error("the Percussionist lacks haste")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mon, game.DestroyOptions{}) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatal("the top card was not exiled when it died")
	}
	perm := exiledPermission(g, loot)
	if !permissionLive(g, perm, me.ID) {
		t.Fatal("the exiled card is not playable")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if permissionLive(g, perm, opp.ID) {
		t.Error("an opponent may play it")
	}
	// It outlasts this turn: still live in the next player's turn and at
	// the start of mine.
	seat := g.Turn.ActiveSeat
	for i := 0; i < 80 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if !permissionLive(g, exiledPermission(g, loot), me.ID) {
		t.Error("the window closed with this turn; it lasts until the end of your next turn")
	}
}

// --- Diversion Specialist ------------------------------------------------------------

func TestStdB23DiversionSpecialistSacrificesAnotherCreatureOrEnchantmentToImpulse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spec := pushCatalogPermanent(g, me.ID, "Diversion Specialist", "Creature — Human Warrior", stdB23DiversionOracle, false)
	if !finB3Has(s58p6View(t, g, spec), "menace") {
		t.Error("the Specialist lacks menace")
	}
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	aura := b12Permanent(g, me.ID, "Ward", "Enchantment")
	loot := topCard(me, "Loot", "Sorcery", "{R}")
	stdB23Mana(t, g, me, "{C}{C}{C}")
	for name, id := range map[string]uuid.UUID{"itself": spec, "an artifact": rock} {
		if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("sacrificed %s to Diversion Specialist", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aura}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(aura) {
		t.Fatal("the enchantment was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatal("the top card was not exiled")
	}
	perm := exiledPermission(g, loot)
	if !permissionLive(g, perm, me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 80 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if permissionLive(g, exiledPermission(g, loot), me.ID) {
		t.Error("the window outlasted the turn")
	}
}

func TestStdB23DiversionSpecialistSacrificesAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spec := pushCatalogPermanent(g, me.ID, "Diversion Specialist", "Creature — Human Warrior", stdB23DiversionOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	stdB23Mana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(spec) {
		t.Error("the other creature, not the Specialist, should be the sacrifice")
	}
}

// --- Popular Egotist -----------------------------------------------------------------

func TestStdB23PopularEgotistSacrificeGivesIndestructibleAndTapsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ego := pushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	aura := b12Permanent(g, me.ID, "Ward", "Enchantment")
	stdB23Mana(t, g, me, "{B}{B}{B}{B}")
	for name, id := range map[string]uuid.UUID{"itself": ego, "an artifact": rock} {
		if err := g.ActivateCatalogAbility(me.ID, ego, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("sacrificed %s to the Egotist's ability", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, ego, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aura}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sacrifice is a cost, so the Egotist's own trigger goes first.
	b04WaitForPick(t, g, me.ID)
	_, opps := stdB23Seats(g)
	myLife, theirLife := me.Life, opps[0].Life
	pickPlayer(t, g, me.ID, opps[0].ID)
	passPriorityAroundTable(t, g)
	if opps[0].Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d → %d, them %d → %d, want +1 and -1", myLife, me.Life, theirLife, opps[0].Life)
	}
	if !hasEffectiveKeyword(t, g, ego, "indestructible") {
		t.Error("the Egotist did not gain indestructible")
	}
	if !dskRT2Tapped(g, ego) {
		t.Error("the Egotist was not tapped")
	}
}

func TestStdB23PopularEgotistDrainsForEveryPermanentYouSacrificeAnyWay(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	treasure := b12Permanent(g, me.ID, "Treasure", "Artifact — Treasure")
	myLife, theirLife := me.Life, opps[1].Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opps[1].ID)
	passPriorityAroundTable(t, g)
	if opps[1].Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("a sacrificed Treasure: life me %d → %d, them %d → %d", myLife, me.Life, theirLife, opps[1].Life)
	}
}

func TestStdB23PopularEgotistDoesNotDrainForAnOpponentsSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	theirs := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	myLife := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || me.Life != myLife {
		t.Error("an opponent's sacrifice triggered the Egotist")
	}
}

func TestStdB23PopularEgotistSacrificingItselfStillTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	ego := pushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	theirLife := opps[0].Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(ego) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opps[0].ID)
	passPriorityAroundTable(t, g)
	if opps[0].Life != theirLife-1 {
		t.Errorf("opponent life %d → %d, want -1", theirLife, opps[0].Life)
	}
}

// --- Sawblade Skinripper -------------------------------------------------------------

func TestStdB23SkinripperSacrificeAnotherCreatureOrEnchantmentForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	rip := b12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	if !hasEffectiveKeyword(t, g, rip, "menace") {
		t.Error("the Skinripper lacks menace")
	}
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	aura := b12Permanent(g, me.ID, "Ward", "Enchantment")
	stdB23Mana(t, g, me, "{C}{C}{C}{C}")
	for name, id := range map[string]uuid.UUID{"itself": rip, "an artifact": rock} {
		if err := g.ActivateCatalogAbility(me.ID, rip, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("sacrificed %s to the Skinripper's ability", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, rip, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aura}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(aura) {
		t.Fatal("the enchantment was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, rip) != 1 {
		t.Errorf("counters %d, want 1", plusOnes(g, rip))
	}
}

func TestStdB23SkinripperEndStepDealsDamageForEachPermanentSacrificedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	one := b12Permanent(g, me.ID, "Rock", "Artifact")
	two := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	theirs := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(one)
		_ = g.SacrificePermanentForEffect(two)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	before := opps[1].Life
	advanceToEndStep(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opps[1].ID)
	passPriorityAroundTable(t, g)
	if opps[1].Life != before-2 {
		t.Errorf("life %d → %d, want -2 for the two permanents you sacrificed (not the opponent's)", before, opps[1].Life)
	}
}

func TestStdB23SkinripperCanTargetACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := b12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	advanceToEndStep(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("1 damage did not kill the 1/1")
	}
}

func TestStdB23SkinripperDoesNotTriggerWithoutASacrificeThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	theirs := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	lives := lifeOfOpponents(g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the end step trigger went on the stack without a permanent sacrificed by you")
	}
	for i, l := range lifeOfOpponents(g) {
		if l != lives[i] {
			t.Errorf("opponent %d lost life", i)
		}
	}
}

// --- Sporogenic Infection ------------------------------------------------------------

func TestStdB23SporogenicInfectionTargetPlayerSacrificesAnotherCreatureThenDamageDestroysTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := b12Creature(g, opps[0].ID, "Host", "Creature — Bear", 3, 3)
	other := b12Creature(g, opps[0].ID, "Other", "Creature — Bear", 2, 2)
	third := b12Creature(g, opps[0].ID, "Third", "Creature — Bear", 2, 2)
	infection := castCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opps[0].ID)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opps[0].ID)
	if c == nil {
		t.Fatalf("the targeted player was not asked to sacrifice: %+v", g.PendingChoices)
	}
	cands := c.SacrificeOptions
	if len(cands) == 0 {
		cands = c.ChooseCards
	}
	if hasID(cands, host) || !hasID(cands, other) || !hasID(cands, third) {
		t.Errorf("sacrifice offers %v, want the other two creatures and never the enchanted one", cands)
	}
	answerSacrifice(t, g, opps[0].ID, other)
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(host) || !g.Battlefield.Contains(third) {
		t.Fatal("exactly the chosen creature should be sacrificed")
	}
	if inf, ok := battlefieldCard(g, infection); !ok || !inf.IsAttachedTo(host) {
		t.Fatal("the Aura is not attached to the host")
	}
	// Damage to the enchanted creature destroys it, however small.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, host, 1) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived being dealt damage")
	}
	if !g.Battlefield.Contains(third) {
		t.Error("damage to the host destroyed another creature")
	}
}

func TestStdB23SporogenicInfectionCanTargetItsControllerWithTheHostSpared(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := b12Creature(g, opps[1].ID, "Host", "Creature — Bear", 3, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, me.ID)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(mine) || !g.Battlefield.Contains(host) {
		t.Error("the controller, targeting themselves, should sacrifice their own creature")
	}
}

func TestStdB23SporogenicInfectionTargetPlayerWithOnlyTheHostSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := b12Creature(g, opps[0].ID, "Host", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opps[0].ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was sacrificed")
	}
}

// --- Osseous Sticktwister --------------------------------------------------------------

func stdB23Sticktwister(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := b12Push(g, me.ID, "Osseous Sticktwister", "Artifact Creature — Scarecrow", stdB23SticktwisterOID, 2, 2)
	if !hasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("the Sticktwister lacks lifelink")
	}
	return id
}

func TestStdB23SticktwisterDeliriumOpponentsChooseThenTheRestTakeItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	stdB23EmptyHands(g)
	stdB23Sticktwister(t, g, me)
	dskRT2Delirium(g, me, 4)
	rock := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	h0 := handCardForTest(opps[0], "H0", "Sorcery", "")
	h1 := handCardForTest(opps[1], "H1", "Sorcery", "")
	lives := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	// Turn order: the first opponent decides first.
	if c := latestOptionPickFor(g, opps[0].ID); c == nil {
		t.Fatalf("the first opponent was not asked: %+v", g.PendingChoices)
	} else if len(c.PickOptions) != 3 {
		t.Fatalf("options %+v, want neither, sacrifice and discard", c.PickOptions)
	}
	if latestOptionPickFor(g, opps[1].ID) != nil {
		t.Fatal("the second opponent was asked before the first answered")
	}
	// Option 1: sacrifice a nonland permanent.
	answerOptionPick(t, g, opps[0].ID, 1)
	// The second opponent controls no nonland permanent: neither or discard.
	c := latestOptionPickFor(g, opps[1].ID)
	if c == nil {
		t.Fatal("the second opponent was not asked")
	}
	if len(c.PickOptions) != 2 {
		t.Fatalf("options %+v, want neither and discard only", c.PickOptions)
	}
	answerOptionPick(t, g, opps[1].ID, 1)
	// The third has neither a permanent nor a card: not asked, hit.
	answerDiscard(t, g, opps[1].ID, h1)
	answerSacrifice(t, g, opps[0].ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the first opponent did not sacrifice")
	}
	if !opps[1].Graveyard.Contains(h1) || !opps[0].Hand.Contains(h0) {
		t.Error("the second opponent should have discarded and the first kept their card")
	}
	if opps[0].Life != lives[0] || opps[1].Life != lives[1] {
		t.Errorf("an opponent who sacrificed or discarded took damage: %d %d", opps[0].Life, opps[1].Life)
	}
	if opps[2].Life != lives[2]-2 {
		t.Errorf("the third opponent's life %d → %d, want -2", lives[2], opps[2].Life)
	}
}

func TestStdB23SticktwisterAnOpponentWhoDeclinesTakesTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	stdB23EmptyHands(g)
	stdB23Sticktwister(t, g, me)
	dskRT2Delirium(g, me, 4)
	rock := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	h0 := handCardForTest(opps[0], "H0", "Sorcery", "")
	before := opps[0].Life
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opps[0].ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || !opps[0].Hand.Contains(h0) {
		t.Error("declining should sacrifice and discard nothing")
	}
	if opps[0].Life != before-2 {
		t.Errorf("life %d → %d, want -2", before, opps[0].Life)
	}
	for _, p := range opps[1:] {
		if p.Life != before-2 {
			t.Errorf("opponent with nothing to give kept %d life", p.Life)
		}
	}
}

func TestStdB23SticktwisterNeedsDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	advanceToMain(t, g)
	stdB23Sticktwister(t, g, me)
	dskRT2Delirium(g, me, 3)
	lives := lifeOfOpponents(g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, opps[0].ID) != nil {
		t.Error("the trigger fired with three card types")
	}
	for i, l := range lifeOfOpponents(g) {
		if l != lives[i] {
			t.Errorf("opponent %d lost life without delirium", i)
		}
	}
}

// --- Meathook Massacre II --------------------------------------------------------------

func stdB23Massacre(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return dcSeed(g, me.ID, "Meathook Massacre II", "Legendary Enchantment", stdB23MeathookOracle, 0, 0)
}

func TestStdB23MeathookMassacreIIEachPlayerSacrificesXCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	mine := []uuid.UUID{
		b12Creature(g, me.ID, "Mine 1", "Creature — Bear", 1, 1),
		b12Creature(g, me.ID, "Mine 2", "Creature — Bear", 1, 1),
		b12Creature(g, me.ID, "Mine 3", "Creature — Bear", 1, 1),
	}
	theirs := []uuid.UUID{
		b12Creature(g, opps[0].ID, "Their 1", "Creature — Bear", 1, 1),
		b12Creature(g, opps[0].ID, "Their 2", "Creature — Bear", 1, 1),
	}
	rock := b12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	castXSpell(t, g, "Meathook Massacre II", "Legendary Enchantment", stdB23MeathookOracle, "{X}{X}{B}{B}{B}{B}", 2, nil)
	passPriorityAroundTable(t, g)
	// Two creatures each: I choose two of my three; the opponent with two
	// sacrifices both; nobody else has any.
	answerSacrifice(t, g, me.ID, mine[0], mine[1])
	passPriorityAroundTable(t, g)
	alive := 0
	for _, id := range mine {
		if g.Battlefield.Contains(id) {
			alive++
		}
	}
	if alive != 1 {
		t.Errorf("%d of my creatures remain, want 1 (X = 2)", alive)
	}
	for _, id := range theirs {
		if g.Battlefield.Contains(id) {
			t.Error("an opponent with two creatures kept one")
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature permanent was sacrificed")
	}
}

func TestStdB23MeathookMassacreIIEnteringUncastSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := stdB23Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stdB23Massacre(t, g, me)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("an uncast Massacre (X = 0) sacrificed a creature")
	}
}

func TestStdB23MeathookMassacreIIYouMayPayLifeToReturnYourCreatureWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 paid", life, me.Life)
	}
	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if c.Counters[game.CounterFinality] != 1 || c.Controller != me.ID {
		t.Errorf("returned with finality %d under %s, want one under me", c.Counters[game.CounterFinality], c.Controller)
	}
	// A finality counter exiles it instead of letting it die again.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature with a finality counter did not go to exile")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an exiled creature raised another dies trigger")
	}
}

func TestStdB23MeathookMassacreIIDecliningToPayLeavesYourCreatureDead(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) || me.Life != life {
		t.Error("declining should leave the creature in the graveyard and the life untouched")
	}
}

func TestStdB23MeathookMassacreIIAControllerWithTooLittleLifeIsNotAsked(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	me.Life = 2
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a player who can't pay 3 life was asked to")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature returned without the life")
	}
}

func TestStdB23MeathookMassacreIIAnOpponentWhoDoesNotPayGivesYouTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opps[0].Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("I was asked about an opponent's creature")
	}
	answerPayUnless(t, g, opps[0].ID, false)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if c.Controller != me.ID || c.Owner != opps[0].ID {
		t.Errorf("returned under %s owned by %s, want my control, their ownership", c.Controller, c.Owner)
	}
	if c.Counters[game.CounterFinality] != 1 {
		t.Errorf("finality counters %d, want 1", c.Counters[game.CounterFinality])
	}
	if opps[0].Life != life {
		t.Error("they paid nothing")
	}
}

func TestStdB23MeathookMassacreIIAnOpponentWhoPaysKeepsTheCreatureInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opps[0].Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opps[0].ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opps[0].Graveyard.Contains(bear) {
		t.Error("the creature should stay in its owner's graveyard")
	}
	if opps[0].Life != life-3 {
		t.Errorf("life %d → %d, want 3 paid", life, opps[0].Life)
	}
}

func TestStdB23MeathookMassacreIIAnOpponentWhoCannotPayGivesYouTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := b12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	opps[0].Life = 2
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opps[0].ID, true)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, bear); !ok || c.Controller != me.ID {
		t.Error("an opponent who can't pay 3 life doesn't pay, so the creature is mine")
	}
	if opps[0].Life != 2 {
		t.Errorf("life %d, want it untouched", opps[0].Life)
	}
}

func TestStdB23MeathookMassacreIITokensNeverReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: opps[0].ID, Controller: opps[0].ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, opps[0].ID) {
		answerPayUnless(t, g, opps[0].ID, false)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("a token returned from the graveyard")
	}
}
