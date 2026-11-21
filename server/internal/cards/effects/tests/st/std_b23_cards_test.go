package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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

// --- Bedhead Beastie ----------------------------------------------------------

func TestStdB23BedheadBeastieMountaincyclingFetchesAMountainAndItHasMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := uuid.New()
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	id, _ := effectstest.CycleFromHand(t, g, "Bedhead Beastie", "Creature — Beast", stdB23BeastieOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Beastie is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.HasID(pick.SearchCards, forest) || !effectstest.HasID(pick.SearchCards, mountain) {
		t.Errorf("the search offers %v, want only the Mountain", pick.SearchCards)
	}
	if err := g.ResolveSearchLibrary(pick.ID, me.ID, []uuid.UUID{mountain}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mountain) {
		t.Error("the Mountain did not come to hand")
	}
	beastie := effectstest.PushCatalogPermanent(g, me.ID, "Bedhead Beastie", "Creature — Beast", stdB23BeastieOracle, false)
	if !effectstest.FinB3Has(effectstest.S58p6View(t, g, beastie), "menace") {
		t.Error("the Beastie lacks menace")
	}
}

// --- Fanatic of the Harrowing --------------------------------------------------

func TestStdB23FanaticEveryPlayerDiscardsAndYouDrawIfYouDid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	mine := effectstest.HandCardForTest(me, "Spare", "Sorcery", "")
	theirs := effectstest.HandCardForTest(opps[0], "Their Spare", "Sorcery", "")
	library := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "Fanatic of the Harrowing", "Creature — Human Cleric", stdB23FanaticOracle, 2, 2)
	effectstest.AnswerDiscard(t, g, me.ID, mine)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opps[0].ID) == nil && effectstest.DiscardChoiceFor(g, opps[0].ID) == nil {
		t.Fatal("the opponent was not asked to discard")
	}
	effectstest.AnswerDiscard(t, g, opps[0].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	theirs := effectstest.HandCardForTest(opps[1], "Their Spare", "Sorcery", "")
	effectstest.CastHoldCreature(t, g, "Fanatic of the Harrowing", "Creature — Human Cleric", stdB23FanaticOracle, 2, 2)
	effectstest.AnswerDiscard(t, g, opps[1].ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0: you discarded no card this way", me.Hand.Size())
	}
	if !opps[1].Graveyard.Contains(theirs) {
		t.Error("the opponent should still discard")
	}
}

// --- Miasma Demon ----------------------------------------------------------------

func TestStdB23MiasmaDemonDiscardsAnyNumberThenShrinksThatManyTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	a := effectstest.HandCardForTest(me, "A", "Sorcery", "")
	b := effectstest.HandCardForTest(me, "B", "Sorcery", "")
	keep := effectstest.HandCardForTest(me, "Keep", "Sorcery", "")
	x := effectstest.B12Creature(g, opps[0].ID, "X", "Creature — Bear", 2, 2)
	y := effectstest.B12Creature(g, opps[1].ID, "Y", "Creature — Bear", 2, 2)
	z := effectstest.B12Creature(g, opps[2].ID, "Z", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, 5, 4)
	effectstest.AnswerDiscard(t, g, me.ID, a, b)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) || !me.Hand.Contains(keep) {
		t.Fatal("exactly the two chosen cards should be discarded")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{effectstest.MshA4Ref(x), effectstest.MshA4Ref(y), effectstest.MshA4Ref(z)}); err == nil {
		t.Fatal("three targets were accepted for two cards discarded")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.MshA4Ref(x), effectstest.MshA4Ref(y))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
		t.Error("the two 2/2s targeted should have died to -2/-2")
	}
	if !g.Battlefield.Contains(z) {
		t.Error("the untargeted creature died")
	}
}

func TestStdB23MiasmaDemonDiscardingNothingMakesNoTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23EmptyHands(g)
	a := effectstest.HandCardForTest(me, "A", "Sorcery", "")
	x := effectstest.B12Creature(g, opps[0].ID, "X", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, 5, 4)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("discarding nothing still asked for targets")
	}
	if !me.Hand.Contains(a) || !g.Battlefield.Contains(x) {
		t.Error("nothing should have happened")
	}
}

func TestStdB23MiasmaDemonHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	demon := effectstest.PushCatalogPermanent(g, me.ID, "Miasma Demon", "Creature — Demon", stdB23MiasmaOracle, false)
	if !effectstest.FinB3Has(effectstest.S58p6View(t, g, demon), "flying") {
		t.Error("Miasma Demon lacks flying")
	}
}

// --- Boilerbilges Ripper -----------------------------------------------------------

func TestStdB23BoilerbilgesRipperSacrificesACreatureOrEnchantmentForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Mountain", "Basic Land — Mountain")
	aura := effectstest.B12Permanent(g, me.ID, "My Ward", "Enchantment")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 2, 2)
	ripper := effectstest.CastHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	for name, id := range map[string]uuid.UUID{"an artifact": rock, "a land": land, "the Ripper itself": ripper} {
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{id}); err == nil {
			t.Fatalf("sacrificed %s", name)
		}
	}
	if !effectstest.HasID(c.ChooseCards, aura) || !effectstest.HasID(c.ChooseCards, fodder) {
		t.Errorf("the prompt offers %v, want the enchantment and the creature", c.ChooseCards)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, aura)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(aura) {
		t.Fatal("the sacrifice happens before the reflexive trigger targets")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage did not kill the 2/2")
	}
}

func TestStdB23BoilerbilgesRipperTwoDamageCanHitAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	before := opps[2].Life
	effectstest.CastHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opps[2].ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opps[2].Life != before-2 {
		t.Errorf("life %d → %d, want -2", before, opps[2].Life)
	}
}

func TestStdB23BoilerbilgesRipperDeclinedDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Boilerbilges Ripper", "Creature — Human Assassin", stdB23RipperOracle, 4, 4)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a declined sacrifice still asked for a target")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("nothing should have happened")
	}
}

// --- Clockwork Percussionist ---------------------------------------------------------

func TestStdB23ClockworkPercussionistDiesExilesTheTopCardPlayableUntilYourNextTurnEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	loot := effectstest.TopCard(me, "Loot", "Sorcery", "{R}")
	mon := effectstest.PushCatalogPermanent(g, me.ID, "Clockwork Percussionist", "Artifact Creature — Monkey Toy", stdB23PercussionOracle, false)
	if !effectstest.FinB3Has(effectstest.S58p6View(t, g, mon), "haste") {
		t.Error("the Percussionist lacks haste")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mon, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatal("the top card was not exiled when it died")
	}
	perm := effectstest.ExiledPermission(g, loot)
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Fatal("the exiled card is not playable")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if effectstest.PermissionLive(g, perm, opp.ID) {
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
	if !effectstest.PermissionLive(g, effectstest.ExiledPermission(g, loot), me.ID) {
		t.Error("the window closed with this turn; it lasts until the end of your next turn")
	}
}

// --- Diversion Specialist ------------------------------------------------------------

func TestStdB23DiversionSpecialistSacrificesAnotherCreatureOrEnchantmentToImpulse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spec := effectstest.PushCatalogPermanent(g, me.ID, "Diversion Specialist", "Creature — Human Warrior", stdB23DiversionOracle, false)
	if !effectstest.FinB3Has(effectstest.S58p6View(t, g, spec), "menace") {
		t.Error("the Specialist lacks menace")
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	loot := effectstest.TopCard(me, "Loot", "Sorcery", "{R}")
	effectstest.StdB23Mana(t, g, me, "{C}{C}{C}")
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
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatal("the top card was not exiled")
	}
	perm := effectstest.ExiledPermission(g, loot)
	if !effectstest.PermissionLive(g, perm, me.ID) {
		t.Error("the exiled card is not playable this turn")
	}
	seat := g.Turn.ActiveSeat
	for i := 0; i < 80 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.PermissionLive(g, effectstest.ExiledPermission(g, loot), me.ID) {
		t.Error("the window outlasted the turn")
	}
}

func TestStdB23DiversionSpecialistSacrificesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spec := effectstest.PushCatalogPermanent(g, me.ID, "Diversion Specialist", "Creature — Human Warrior", stdB23DiversionOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.StdB23Mana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, spec, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(spec) {
		t.Error("the other creature, not the Specialist, should be the sacrifice")
	}
}

// --- Popular Egotist -----------------------------------------------------------------

func TestStdB23PopularEgotistSacrificeGivesIndestructibleAndTapsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	ego := effectstest.PushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	effectstest.StdB23Mana(t, g, me, "{B}{B}{B}{B}")
	for name, id := range map[string]uuid.UUID{"itself": ego, "an artifact": rock} {
		if err := g.ActivateCatalogAbility(me.ID, ego, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
			t.Fatalf("sacrificed %s to the Egotist's ability", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, ego, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{aura}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The sacrifice is a cost, so the Egotist's own trigger goes first.
	effectstest.B04WaitForPick(t, g, me.ID)
	_, opps := stdB23Seats(g)
	myLife, theirLife := me.Life, opps[0].Life
	effectstest.PickPlayer(t, g, me.ID, opps[0].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opps[0].Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d → %d, them %d → %d, want +1 and -1", myLife, me.Life, theirLife, opps[0].Life)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ego, "indestructible") {
		t.Error("the Egotist did not gain indestructible")
	}
	if !effectstest.DskRT2Tapped(g, ego) {
		t.Error("the Egotist was not tapped")
	}
}

func TestStdB23PopularEgotistDrainsForEveryPermanentYouSacrificeAnyWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	treasure := effectstest.B12Permanent(g, me.ID, "Treasure", "Artifact — Treasure")
	myLife, theirLife := me.Life, opps[1].Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(treasure) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opps[1].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opps[1].Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("a sacrificed Treasure: life me %d → %d, them %d → %d", myLife, me.Life, theirLife, opps[1].Life)
	}
}

func TestStdB23PopularEgotistDoesNotDrainForAnOpponentsSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	theirs := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	myLife := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || me.Life != myLife {
		t.Error("an opponent's sacrifice triggered the Egotist")
	}
}

func TestStdB23PopularEgotistSacrificingItselfStillTriggers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	ego := effectstest.PushCatalogPermanent(g, me.ID, "Popular Egotist", "Creature — Human Rogue", stdB23EgotistOracle, false)
	theirLife := opps[0].Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(ego) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opps[0].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opps[0].Life != theirLife-1 {
		t.Errorf("opponent life %d → %d, want -1", theirLife, opps[0].Life)
	}
}

// --- Sawblade Skinripper -------------------------------------------------------------

func TestStdB23SkinripperSacrificeAnotherCreatureOrEnchantmentForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	rip := effectstest.B12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, rip, "menace") {
		t.Error("the Skinripper lacks menace")
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	aura := effectstest.B12Permanent(g, me.ID, "Ward", "Enchantment")
	effectstest.StdB23Mana(t, g, me, "{C}{C}{C}{C}")
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOnes(g, rip) != 1 {
		t.Errorf("counters %d, want 1", effectstest.PlusOnes(g, rip))
	}
}

func TestStdB23SkinripperEndStepDealsDamageForEachPermanentSacrificedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	one := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	two := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(one)
		_ = g.SacrificePermanentForEffect(two)
		_ = g.SacrificePermanentForEffect(theirs)
	})
	before := opps[1].Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opps[1].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opps[1].Life != before-2 {
		t.Errorf("life %d → %d, want -2 for the two permanents you sacrificed (not the opponent's)", before, opps[1].Life)
	}
}

func TestStdB23SkinripperCanTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opps[0].ID, "Victim", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.AdvanceToEndStep(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("1 damage did not kill the 1/1")
	}
}

func TestStdB23SkinripperDoesNotTriggerWithoutASacrificeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sawblade Skinripper", "Creature — Human Assassin", stdB23SkinripperOracle, 3, 2)
	theirs := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	lives := effectstest.LifeOfOpponents(g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the end step trigger went on the stack without a permanent sacrificed by you")
	}
	for i, l := range effectstest.LifeOfOpponents(g) {
		if l != lives[i] {
			t.Errorf("opponent %d lost life", i)
		}
	}
}

// --- Sporogenic Infection ------------------------------------------------------------

func TestStdB23SporogenicInfectionTargetPlayerSacrificesAnotherCreatureThenDamageDestroysTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := effectstest.B12Creature(g, opps[0].ID, "Host", "Creature — Bear", 3, 3)
	other := effectstest.B12Creature(g, opps[0].ID, "Other", "Creature — Bear", 2, 2)
	third := effectstest.B12Creature(g, opps[0].ID, "Third", "Creature — Bear", 2, 2)
	infection := effectstest.CastCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opps[0].ID)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opps[0].ID)
	if c == nil {
		t.Fatalf("the targeted player was not asked to sacrifice: %+v", g.PendingChoices)
	}
	cands := c.SacrificeOptions
	if len(cands) == 0 {
		cands = c.ChooseCards
	}
	if effectstest.HasID(cands, host) || !effectstest.HasID(cands, other) || !effectstest.HasID(cands, third) {
		t.Errorf("sacrifice offers %v, want the other two creatures and never the enchanted one", cands)
	}
	effectstest.AnswerSacrifice(t, g, opps[0].ID, other)
	if g.Battlefield.Contains(other) || !g.Battlefield.Contains(host) || !g.Battlefield.Contains(third) {
		t.Fatal("exactly the chosen creature should be sacrificed")
	}
	if inf, ok := effectstest.BattlefieldCard(g, infection); !ok || !inf.IsAttachedTo(host) {
		t.Fatal("the Aura is not attached to the host")
	}
	// Damage to the enchanted creature destroys it, however small.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, host, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived being dealt damage")
	}
	if !g.Battlefield.Contains(third) {
		t.Error("damage to the host destroyed another creature")
	}
}

func TestStdB23SporogenicInfectionCanTargetItsControllerWithTheHostSpared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := effectstest.B12Creature(g, opps[1].ID, "Host", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(mine) || !g.Battlefield.Contains(host) {
		t.Error("the controller, targeting themselves, should sacrifice their own creature")
	}
}

func TestStdB23SporogenicInfectionTargetPlayerWithOnlyTheHostSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	host := effectstest.B12Creature(g, opps[0].ID, "Host", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Sporogenic Infection", "Enchantment — Aura", stdB23InfectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opps[0].ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) {
		t.Error("the enchanted creature was sacrificed")
	}
}

// --- Osseous Sticktwister --------------------------------------------------------------

func stdB23Sticktwister(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.B12Push(g, me.ID, "Osseous Sticktwister", "Artifact Creature — Scarecrow", stdB23SticktwisterOID, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
		t.Error("the Sticktwister lacks lifelink")
	}
	return id
}

func TestStdB23SticktwisterDeliriumOpponentsChooseThenTheRestTakeItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	stdB23EmptyHands(g)
	stdB23Sticktwister(t, g, me)
	effectstest.DskRT2Delirium(g, me, 4)
	rock := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	h0 := effectstest.HandCardForTest(opps[0], "H0", "Sorcery", "")
	h1 := effectstest.HandCardForTest(opps[1], "H1", "Sorcery", "")
	lives := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Turn order: the first opponent decides first.
	if c := effectstest.LatestOptionPickFor(g, opps[0].ID); c == nil {
		t.Fatalf("the first opponent was not asked: %+v", g.PendingChoices)
	} else if len(c.PickOptions) != 3 {
		t.Fatalf("options %+v, want neither, sacrifice and discard", c.PickOptions)
	}
	if effectstest.LatestOptionPickFor(g, opps[1].ID) != nil {
		t.Fatal("the second opponent was asked before the first answered")
	}
	// Option 1: sacrifice a nonland permanent.
	effectstest.AnswerOptionPick(t, g, opps[0].ID, 1)
	// The second opponent controls no nonland permanent: neither or discard.
	c := effectstest.LatestOptionPickFor(g, opps[1].ID)
	if c == nil {
		t.Fatal("the second opponent was not asked")
	}
	if len(c.PickOptions) != 2 {
		t.Fatalf("options %+v, want neither and discard only", c.PickOptions)
	}
	effectstest.AnswerOptionPick(t, g, opps[1].ID, 1)
	// The third has neither a permanent nor a card: not asked, hit.
	effectstest.AnswerDiscard(t, g, opps[1].ID, h1)
	effectstest.AnswerSacrifice(t, g, opps[0].ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	stdB23EmptyHands(g)
	stdB23Sticktwister(t, g, me)
	effectstest.DskRT2Delirium(g, me, 4)
	rock := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	h0 := effectstest.HandCardForTest(opps[0], "H0", "Sorcery", "")
	before := opps[0].Life
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opps[0].ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	effectstest.AdvanceToMain(t, g)
	stdB23Sticktwister(t, g, me)
	effectstest.DskRT2Delirium(g, me, 3)
	lives := effectstest.LifeOfOpponents(g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, opps[0].ID) != nil {
		t.Error("the trigger fired with three card types")
	}
	for i, l := range effectstest.LifeOfOpponents(g) {
		if l != lives[i] {
			t.Errorf("opponent %d lost life without delirium", i)
		}
	}
}

// --- Meathook Massacre II --------------------------------------------------------------

func stdB23Massacre(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.DcSeed(g, me.ID, "Meathook Massacre II", "Legendary Enchantment", stdB23MeathookOracle, 0, 0)
}

func TestStdB23MeathookMassacreIIEachPlayerSacrificesXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	mine := []uuid.UUID{
		effectstest.B12Creature(g, me.ID, "Mine 1", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Mine 2", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, me.ID, "Mine 3", "Creature — Bear", 1, 1),
	}
	theirs := []uuid.UUID{
		effectstest.B12Creature(g, opps[0].ID, "Their 1", "Creature — Bear", 1, 1),
		effectstest.B12Creature(g, opps[0].ID, "Their 2", "Creature — Bear", 1, 1),
	}
	rock := effectstest.B12Permanent(g, opps[0].ID, "Their Rock", "Artifact")
	effectstest.CastXSpell(t, g, "Meathook Massacre II", "Legendary Enchantment", stdB23MeathookOracle, "{X}{X}{B}{B}{B}{B}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two creatures each: I choose two of my three; the opponent with two
	// sacrifices both; nobody else has any.
	effectstest.AnswerSacrifice(t, g, me.ID, mine[0], mine[1])
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, _ := stdB23Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stdB23Massacre(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("an uncast Massacre (X = 0) sacrificed a creature")
	}
}

func TestStdB23MeathookMassacreIIYouMayPayLifeToReturnYourCreatureWithAFinalityCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 paid", life, me.Life)
	}
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if c.Counters[game.CounterFinality] != 1 || c.Controller != me.ID {
		t.Errorf("returned with finality %d under %s, want one under me", c.Counters[game.CounterFinality], c.Controller)
	}
	// A finality counter exiles it instead of letting it die again.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature with a finality counter did not go to exile")
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an exiled creature raised another dies trigger")
	}
}

func TestStdB23MeathookMassacreIIDecliningToPayLeavesYourCreatureDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) || me.Life != life {
		t.Error("declining should leave the creature in the graveyard and the life untouched")
	}
}

func TestStdB23MeathookMassacreIIAControllerWithTooLittleLifeIsNotAsked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	me.Life = 2
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a player who can't pay 3 life was asked to")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature returned without the life")
	}
}

func TestStdB23MeathookMassacreIIAnOpponentWhoDoesNotPayGivesYouTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opps[0].Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("I was asked about an opponent's creature")
	}
	effectstest.AnswerPayUnless(t, g, opps[0].ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, bear)
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
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opps[0].Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opps[0].ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opps[0].Graveyard.Contains(bear) {
		t.Error("the creature should stay in its owner's graveyard")
	}
	if opps[0].Life != life-3 {
		t.Errorf("life %d → %d, want 3 paid", life, opps[0].Life)
	}
}

func TestStdB23MeathookMassacreIIAnOpponentWhoCannotPayGivesYouTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	bear := effectstest.B12Creature(g, opps[0].ID, "Their Bear", "Creature — Bear", 2, 2)
	opps[0].Life = 2
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opps[0].ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.Controller != me.ID {
		t.Error("an opponent who can't pay 3 life doesn't pay, so the creature is mine")
	}
	if opps[0].Life != 2 {
		t.Errorf("life %d, want it untouched", opps[0].Life)
	}
}

func TestStdB23MeathookMassacreIITokensNeverReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := stdB23Seats(g)
	stdB23Massacre(t, g, me)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: opps[0].ID, Controller: opps[0].ID,
	})
	// The token dies to a resolving spell, so the CR 704.5d sweep runs at
	// the boundary after it, before the dies trigger resolves. (A bare
	// destroy outside any resolution reaches no such boundary; since
	// #5793 the trigger's question holds the table, so nothing else would
	// run the sweep before the decline.)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardRefs(token))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opps[0].ID) {
		effectstest.AnswerPayUnless(t, g, opps[0].ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("a token returned from the graveyard")
	}
}
