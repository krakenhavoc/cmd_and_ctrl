package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gideons_test.go — #2046: the Gideons that become creatures while
// staying planeswalkers, and the engine seam under them (ADR 0032
// amendment of 2026-10-07; the state-based half is in
// game/planeswalker_creature_test.go).

const (
	gideonAllyOracle       = "51eae7ff-fed9-4afe-8053-7690379449dd"
	gideonParagonOracle    = "4cbce730-7b70-41bd-b463-bfec46afe3b0"
	gideonOathswornOracle  = "19a81aa7-823b-43fa-abc2-b2700a122bc1"
	gideonBlackbladeOracle = "813c19f5-3580-488d-9eee-c7a563def532"
)

// pushGideon seats a Gideon planeswalker (subtype Gideon, as printed)
// with `loyalty` counters. No battlefield-entry event is emitted, so it
// is not summoning sick.
func pushGideon(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Legendary Planeswalker — Gideon",
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

// unpreventableDamage deals n damage that can't be prevented to target
// from `src`, then runs the state-based actions.
func unpreventableDamage(t *testing.T, g *game.Game, src, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(src, nil, target, n, game.DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
}

// pushStampedGideon is pushGideon through the battlefield-entry event,
// which stamps the entry timestamp a static ability's layer effect is
// ordered by. Gideon Blackblade's statics need it.
func pushStampedGideon(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	id := pushGideon(g, owner, name, oracle, loyalty)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	return id
}

// wantCreaturePlaneswalker fails unless id is both on the battlefield.
func wantCreaturePlaneswalker(t *testing.T, g *game.Game, id uuid.UUID, power, toughness int) game.Card {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Gideon left the battlefield")
	}
	if !c.IsCreature() || !c.IsPlaneswalker() {
		t.Fatalf("types = %v, want a creature that is still a planeswalker", c.Effective().Types)
	}
	if eff := c.Effective(); eff.Power != power || eff.Toughness != toughness {
		t.Errorf("P/T = %d/%d, want %d/%d", eff.Power, eff.Toughness, power, toughness)
	}
	return c
}

// --- the seam, through a real card -----------------------------------

// The headline of #2046 and Luke's note on Gideon of the Trials: with
// his shield up ordinary damage is prevented, but damage that can't be
// prevented (CR 615.12) both removes loyalty (CR 120.3c) and is marked
// (CR 120.3e), and when it takes the last loyalty counter he is put into
// the graveyard however indestructible he is (CR 704.5i).
func TestGideonOfTheTrialsUnpreventableDamageRemovesLoyaltyAndMarks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	gideon := a109p6Walker(t, g, me, "Gideon of the Trials", gideonOfTheTrialsOracle, 3)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})

	unpreventableDamage(t, g, bear, gideon, 2)
	c := wantCreaturePlaneswalker(t, g, gideon, 4, 4)
	if got := c.Counters[game.CounterLoyalty]; got != 1 {
		t.Errorf("loyalty after 2 unpreventable damage = %d, want 1 (CR 120.3c)", got)
	}
	if c.DamageMarked != 2 {
		t.Errorf("damage marked = %d, want 2 (CR 120.3e)", c.DamageMarked)
	}

	unpreventableDamage(t, g, bear, gideon, 1)
	if g.Battlefield.Contains(gideon) {
		t.Error("an indestructible Gideon with no loyalty counters stayed on the battlefield (CR 704.5i)")
	}
}

// --- Gideon, Ally of Zendikar ----------------------------------------

func TestGideonAllyOfZendikarPlusOneMakesHimAShieldedFiveFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Ally of Zendikar", gideonAllyOracle, 4)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{})

	c := wantCreaturePlaneswalker(t, g, gideon, 5, 5)
	for _, st := range []string{"Human", "Soldier", "Ally"} {
		if !c.HasSubtype(st) {
			t.Errorf("missing subtype %s: %v", st, c.Effective().Subtypes)
		}
	}
	if !game.HasKeyword(&c, "indestructible") {
		t.Error("no indestructible")
	}
	// +1 took him to 5 loyalty; ordinary damage is prevented whole.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, gideon, 3) })
	if got := loyaltyCount(g, gideon); got != 5 {
		t.Errorf("loyalty after prevented damage = %d, want 5", got)
	}
}

func TestGideonAllyOfZendikarZeroMakesAKnightAlly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Ally of Zendikar", gideonAllyOracle, 4)
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})

	knight := findBattlefieldByName(g, "Knight Ally")
	if knight == uuid.Nil {
		t.Fatal("no Knight Ally token")
	}
	c, _ := battlefieldCard(g, knight)
	if c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "W" || !c.HasSubtype("Knight") || !c.HasSubtype("Ally") {
		t.Errorf("token = %d/%d %v %v, want a 2/2 white Knight Ally", c.Power, c.Toughness, c.Colors, c.Effective().Subtypes)
	}
}

func TestGideonAllyOfZendikarEmblemIsAnAnthemForItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Ally of Zendikar", gideonAllyOracle, 4)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Activate(t, g, me.ID, gideon, 2, game.ActivateAbilityParams{})

	if me.Emblems == nil || len(me.Emblems.Cards) != 1 {
		t.Fatal("the −4 made no emblem")
	}
	if g.Battlefield.Contains(gideon) {
		t.Error("Gideon at 0 loyalty is still on the battlefield")
	}
	g.RunStateChecksForTest()
	if p, tt := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my Bear = %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("their Bear = %d/%d, want 2/2 (the anthem is for the emblem's owner)", p, tt)
	}
}

// --- Gideon, Martial Paragon -----------------------------------------

func TestGideonMartialParagonPlusTwoUntapsAndPumpsMyCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Martial Paragon", gideonParagonOracle, 5)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Tap(g, mine)
	b16Tap(g, theirs)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{})

	if b16Tapped(t, g, mine) {
		t.Error("my creature is still tapped")
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
	if p, tt := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("my Bear = %d/%d, want 3/3", p, tt)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their Bear power = %d, want 2", p)
	}
	if got := loyaltyCount(g, gideon); got != 7 {
		t.Errorf("loyalty = %d, want 7", got)
	}
}

func TestGideonMartialParagonZeroMakesHimAShieldedFiveFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Martial Paragon", gideonParagonOracle, 5)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})

	c := wantCreaturePlaneswalker(t, g, gideon, 5, 5)
	if !c.HasSubtype("Human") || !c.HasSubtype("Soldier") || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("subtypes %v abilities %v, want Human Soldier with indestructible", c.Effective().Subtypes, c.Effective().Abilities)
	}
	unpreventableDamage(t, g, bear, gideon, 5)
	if g.Battlefield.Contains(gideon) {
		t.Error("5 unpreventable damage at 5 loyalty left Gideon on the battlefield")
	}
}

func TestGideonMartialParagonMinusTenPumpsMineAndTapsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, Martial Paragon", gideonParagonOracle, 10)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Activate(t, g, me.ID, gideon, 2, game.ActivateAbilityParams{})

	if p, tt := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("my Bear = %d/%d, want 4/4", p, tt)
	}
	if !b16Tapped(t, g, theirs) {
		t.Error("their creature is untapped")
	}
	if b16Tapped(t, g, mine) {
		t.Error("my creature was tapped")
	}
}

// --- Gideon, the Oathsworn -------------------------------------------

func TestGideonTheOathswornPlusTwoIsAWhiteSoldierWithoutIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, the Oathsworn", gideonOathswornOracle, 4)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{})

	c := wantCreaturePlaneswalker(t, g, gideon, 5, 5)
	if !c.HasSubtype("Soldier") || c.HasSubtype("Human") {
		t.Errorf("subtypes = %v, want Soldier only", c.Effective().Subtypes)
	}
	if cols := c.Effective().Colors; len(cols) != 1 || cols[0] != "W" {
		t.Errorf("colors = %v, want white", cols)
	}
	if game.HasKeyword(&c, "indestructible") {
		t.Error("the Oathsworn has no indestructible")
	}
	// Prevented damage does nothing; unpreventable lethal damage kills.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, gideon, 9) })
	if got := loyaltyCount(g, gideon); got != 6 {
		t.Errorf("loyalty after prevented damage = %d, want 6", got)
	}
	unpreventableDamage(t, g, bear, gideon, 5)
	if g.Battlefield.Contains(gideon) {
		t.Error("5 unpreventable damage on a non-indestructible 5/5 left him on the battlefield (CR 704.5g)")
	}
}

func TestGideonTheOathswornMinusNineExilesHimAndTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, the Oathsworn", gideonOathswornOracle, 11)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	theirLand := pushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})

	if g.Battlefield.Contains(theirs) {
		t.Error("their creature was not exiled")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirLand) {
		t.Error("my creature or their land was exiled")
	}
	if g.Battlefield.Contains(gideon) {
		t.Error("Gideon was not exiled")
	}
	if g.Seats[0].Graveyard.Contains(gideon) {
		t.Error("Gideon went to the graveyard rather than exile")
	}
}

// Paid down from exactly 9, he has no loyalty counters when the ability
// resolves, so CR 704.5i has already put him into the graveyard: he is a
// new object there (CR 400.7) and "exile Gideon" finds nothing. The
// creatures are exiled all the same.
func TestGideonTheOathswornMinusNineFromNineLeavesHimInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, the Oathsworn", gideonOathswornOracle, 9)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Activate(t, g, me.ID, gideon, 1, game.ActivateAbilityParams{})

	if g.Battlefield.Contains(theirs) {
		t.Error("their creature was not exiled")
	}
	if !me.Graveyard.Contains(gideon) {
		t.Error("Gideon, paid down to 0 loyalty, is not in the graveyard")
	}
}

func TestGideonTheOathswornCountersGoOnEachNonGideonAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	pushGideon(g, me.ID, "Gideon, the Oathsworn", gideonOathswornOracle, 4)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	idle := pushVanillaCreature(g, me.ID, "Bear C", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if c, _ := battlefieldCard(g, id); c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("attacker %s has %d +1/+1 counters, want 1", c.Name, c.Counters[game.CounterPlusOne])
		}
	}
	if c, _ := battlefieldCard(g, idle); c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a creature that stayed home got %d counters", c.Counters[game.CounterPlusOne])
	}
}

// One non-Gideon creature is not "two or more", and an animated Gideon
// attacking beside it is a Gideon, so it does not make a second.
func TestGideonTheOathswornNeedsTwoNonGideonAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushGideon(g, me.ID, "Gideon, the Oathsworn", gideonOathswornOracle, 4)
	b16Activate(t, g, me.ID, gideon, 0, game.ActivateAbilityParams{})
	bear := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, gideon} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", id, err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bear, gideon} {
		if c, _ := battlefieldCard(g, id); c.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("%s got %d +1/+1 counters from one non-Gideon attacker, want 0", c.Name, c.Counters[game.CounterPlusOne])
		}
	}
}

// --- Gideon Blackblade -----------------------------------------------

func TestGideonBlackbladeIsACreatureOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	toMain(t, g)
	gideon := pushStampedGideon(g, me.ID, "Gideon Blackblade", gideonBlackbladeOracle, 4)
	g.RunStateChecksForTest()

	c := wantCreaturePlaneswalker(t, g, gideon, 4, 4)
	if !c.HasSubtype("Human") || !c.HasSubtype("Soldier") || !c.HasSubtype("Gideon") || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("on my turn: subtypes %v abilities %v, want a Human Soldier Gideon with indestructible",
			c.Effective().Subtypes, c.Effective().Abilities)
	}

	for i := 0; i < 60 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("never reached another player's turn")
	}
	g.RunStateChecksForTest()
	c, ok := battlefieldCard(g, gideon)
	if !ok {
		t.Fatal("Gideon left the battlefield")
	}
	if c.IsCreature() || !c.IsPlaneswalker() || game.HasKeyword(&c, "indestructible") {
		t.Errorf("on an opponent's turn: types %v abilities %v, want a plain planeswalker",
			c.Effective().Types, c.Effective().Abilities)
	}
}

// "Prevent all damage that would be dealt to Gideon Blackblade during
// your turn": on his controller's turn ordinary damage does nothing; on
// an opponent's turn it removes loyalty like any planeswalker's.
func TestGideonBlackbladeDamageIsPreventedOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushStampedGideon(g, me.ID, "Gideon Blackblade", gideonBlackbladeOracle, 4)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	g.RunStateChecksForTest()

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, gideon, 3) })
	if got := loyaltyCount(g, gideon); got != 4 {
		t.Errorf("loyalty after damage on my turn = %d, want 4 (prevented)", got)
	}

	// Unpreventable damage on my turn still lands (CR 615.12): loyalty
	// off and damage marked, both rules in play.
	unpreventableDamage(t, g, bear, gideon, 3)
	c := wantCreaturePlaneswalker(t, g, gideon, 4, 4)
	if got := c.Counters[game.CounterLoyalty]; got != 1 || c.DamageMarked != 3 {
		t.Errorf("after 3 unpreventable damage: loyalty %d, marked %d; want 1 and 3", got, c.DamageMarked)
	}

	for i := 0; i < 60 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, gideon, 1) })
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(gideon) {
		t.Error("1 damage on an opponent's turn at 1 loyalty did not remove him")
	}
}

func TestGideonBlackbladePlusOneGrantsTheChosenKeyword(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := a109p6Seats(g)
	toMain(t, g)
	gideon := pushStampedGideon(g, me.ID, "Gideon Blackblade", gideonBlackbladeOracle, 4)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, gideon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no keyword pick was offered")
	}
	if n := len(latestOptionPickFor(g, me.ID).PickOptions); n != 3 {
		t.Fatalf("pick has %d options, want 3", n)
	}
	answerOptionPick(t, g, me.ID, 1)
	c, _ := battlefieldCard(g, bear)
	if !game.HasKeyword(&c, "lifelink") || game.HasKeyword(&c, "vigilance") || game.HasKeyword(&c, "indestructible") {
		t.Errorf("abilities = %v, want lifelink only", c.Effective().Abilities)
	}
}

func TestGideonBlackbladeMinusSixExilesANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := a109p6Seats(g)
	toMain(t, g)
	gideon := pushStampedGideon(g, me.ID, "Gideon Blackblade", gideonBlackbladeOracle, 6)
	bear := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	if err := g.ActivateCatalogAbility(me.ID, gideon, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the target is still on the battlefield")
	}
	// A land is not a legal target.
	g2 := newCatalogGame(t)
	me2, opp2 := a109p6Seats(g2)
	toMain(t, g2)
	walker := pushGideon(g2, me2.ID, "Gideon Blackblade", gideonBlackbladeOracle, 6)
	land := pushCatalogPermanent(g2, opp2.ID, "Forest", "Basic Land — Forest", "", false)
	if err := g2.ActivateCatalogAbility(me2.ID, walker, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Error("the −6 accepted a land as its target")
	}
	if got := loyaltyCount(g2, walker); got != 6 {
		t.Errorf("a refused activation spent loyalty: %d, want 6", got)
	}
}
