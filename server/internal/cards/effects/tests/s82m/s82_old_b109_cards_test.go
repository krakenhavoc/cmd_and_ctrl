package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b109_cards_test.go — Core Set 2021, batch 109 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	o109Sporeweb     = "2e7c3881-6fa6-4fd2-ae86-9e3a0fcecdbd"
	o109Necromentia  = "92acbbeb-f82b-413b-b243-f7a012f5400c"
	o109Replicator   = "f379fa0b-e9fc-4307-a574-a0c95e5e7f72"
	o109TurnToSlag   = "4339f621-8502-4433-a37a-654591ca1b43"
	o109Transmogrify = "4c1c3a0a-5782-4b7d-8b9c-61eea913166d"
	o109Discontinuit = "5cfd73ab-8509-4e4b-9704-68d8dad3b301"
	o109Wurm         = "99ed6e18-7f14-4ea8-baee-ad6c081388ff"
	o109Haze         = "b24d1540-ed8b-4df3-abfe-1b8492a58388"
	o109Kaervek      = "b338d879-189d-4b74-9d63-42920fdb9855"
	o109Sanctum      = "1818b0b1-0104-4673-9be8-acfc773ef87d"
	o109Riddleform   = "cc261487-060f-47e6-ac5e-a204254058b3"
	o109Bear         = "Creature — Bear"
)

func TestS82OldB109EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		o109Sporeweb, o109Necromentia, o109Replicator, o109TurnToSlag, o109Transmogrify, o109Discontinuit,
		o109Wurm, o109Haze, o109Kaervek, o109Sanctum, o109Riddleform,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o109Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	p.Strict = true
	return id, g.CastSpell(me.ID, id, p)
}

func o109MustCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := o109Cast(t, g, name, typeLine, cost, oracle, p)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func o109Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o109Pay(t *testing.T, g *game.Game, spec string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, g.Seats[g.Turn.ActiveSeat], spec)
}

// o109Perm puts a permanent with the given oracle id and P/T onto the battlefield.
func o109Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func o109CardOn(g *game.Game, id uuid.UUID) (game.Card, bool) { return effectstest.AangCardOnBF(g, id) }

// --- Sporeweb Weaver ---------------------------------------------------------------------

func TestS82OldB109SporewebWeaverGainsLifeAndMakesASaprolingWhenDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weaver := o109Perm(g, me.ID, "Sporeweb Weaver", "Creature — Spider", o109Sporeweb, 1, 4)
	if !effectstest.HasKeywordOf(g, weaver, "reach") || !effectstest.HasKeywordOf(g, weaver, game.HexproofFromColor("U")) {
		t.Error("the Weaver lacks reach or hexproof from blue")
	}
	life := me.Life
	effectstest.DealDamageTo(t, g, weaver, weaver, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
	effectstest.DealDamageTo(t, g, weaver, weaver, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); n != 2 {
		t.Errorf("%d Saprolings after a second hit, want 2", n)
	}
}

// --- Chrome Replicator -------------------------------------------------------------------

func TestS82OldB109ChromeReplicatorMakesAConstructOnlyWithTwoSameNamed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o109Perm(g, me.ID, "Bear", o109Bear, "", 2, 2)
	o109Perm(g, me.ID, "Bear", o109Bear, "", 2, 2)
	effectstest.EnterFromHand(t, g, me.ID, "Chrome Replicator", "Artifact Creature — Construct", o109Replicator)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Construct", me.ID); n != 1 {
		t.Errorf("%d Constructs with two Bears, want 1", n)
	}
}

func TestS82OldB109ChromeReplicatorMakesNothingWithDistinctNamesOrTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o109Perm(g, me.ID, "Bear", o109Bear, "", 2, 2)
	o109Perm(g, me.ID, "Wolf", o109Bear, "", 2, 2)
	// Two same-named tokens and two same-named lands do not count.
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		effectstest.AangPushLand(g, me.ID, "Forest", false)
	}
	effectstest.EnterFromHand(t, g, me.ID, "Chrome Replicator", "Artifact Creature — Construct", o109Replicator)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Construct", me.ID); n != 0 {
		t.Errorf("%d Constructs, want 0", n)
	}
}

func TestS82OldB109ChromeReplicatorCountsItselfWithAnotherCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o109Perm(g, me.ID, "Chrome Replicator", "Artifact Creature — Construct", o109Replicator, 4, 4)
	effectstest.EnterFromHand(t, g, me.ID, "Chrome Replicator", "Artifact Creature — Construct", o109Replicator)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Construct", me.ID); n != 1 {
		t.Errorf("%d Constructs, want 1", n)
	}
}

// --- Turn to Slag ------------------------------------------------------------------------

func TestS82OldB109TurnToSlagDamagesAndDestroysAttachedEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	tank := o109Perm(g, opp.ID, "Tank", o109Bear, "", 6, 6)
	other := o109Perm(g, opp.ID, "Other", o109Bear, "", 2, 2)
	mk := func(host uuid.UUID, name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Equipment",
			Owner: opp.ID, Controller: opp.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	}
	e1, e2, e3 := mk(tank, "Sword"), mk(tank, "Axe"), mk(other, "Spear")
	o109Pay(t, g, "RRRRR")
	o109MustCast(t, g, "Turn to Slag", "Sorcery", "{3}{R}{R}", o109TurnToSlag, game.CastSpellParams{Targets: o109Target(tank)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := o109CardOn(g, tank); !ok || c.DamageMarked != 5 {
		t.Errorf("the Tank should survive with 5 damage, got %+v (on battlefield %v)", c.DamageMarked, ok)
	}
	for _, id := range []uuid.UUID{e1, e2} {
		if _, ok := o109CardOn(g, id); ok {
			t.Error("an Equipment attached to the target survived")
		}
	}
	if _, ok := o109CardOn(g, e3); !ok {
		t.Error("an Equipment attached to another creature was destroyed")
	}
}

func TestS82OldB109TurnToSlagRefusesANoncreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	o109Pay(t, g, "RRRRR")
	if _, err := o109Cast(t, g, "Turn to Slag", "Sorcery", "{3}{R}{R}", o109TurnToSlag, game.CastSpellParams{Targets: o109Target(rock)}); err == nil {
		t.Error("Turn to Slag targeted an artifact")
	}
	g.WithWriteLock(func() { g.Seats[g.Turn.ActiveSeat].ManaPool.EmptyPool() })
	bear := o109Perm(g, g.Seats[1].ID, "Bear", o109Bear, "", 2, 2)
	if _, err := o109Cast(t, g, "Turn to Slag", "Sorcery", "{3}{R}{R}", o109TurnToSlag, game.CastSpellParams{Targets: o109Target(bear)}); err == nil {
		t.Error("Turn to Slag was cast with no mana")
	}
}

// --- Transmogrify ------------------------------------------------------------------------

func TestS82OldB109TransmogrifyReplacesTheCreatureWithTheNextOneInTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	victim := o109Perm(g, opp.ID, "Victim", o109Bear, "", 2, 2)
	var land, hit uuid.UUID
	g.WithWriteLock(func() {
		hit, land = uuid.New(), uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: hit, Name: "Hit Dragon", TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
		opp.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	})
	o109Pay(t, g, "RRRR")
	o109MustCast(t, g, "Transmogrify", "Sorcery", "{3}{R}", o109Transmogrify, game.CastSpellParams{Targets: o109Target(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := o109CardOn(g, victim); ok {
		t.Error("the target was not exiled")
	}
	c, ok := o109CardOn(g, hit)
	if !ok || c.Controller != opp.ID {
		t.Errorf("the revealed creature is not on the battlefield under its owner (%v)", ok)
	}
	if _, ok := o109CardOn(g, land); ok {
		t.Error("the revealed land entered the battlefield")
	}
	if z := g.FindCardZoneForEffect(land); z == nil || z.Kind != game.ZoneLibrary {
		t.Error("the revealed land is not back in the library")
	}
}

func TestS82OldB109TransmogrifyRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	o109Pay(t, g, "RRRR")
	if _, err := o109Cast(t, g, "Transmogrify", "Sorcery", "{3}{R}", o109Transmogrify, game.CastSpellParams{Targets: o109Target(land)}); err == nil {
		t.Error("Transmogrify targeted a land")
	}
}

// --- Discontinuity -----------------------------------------------------------------------

func TestS82OldB109DiscontinuityCostsLessOnYourTurnAndEndsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	o109Pay(t, g, "UU")
	turn := g.Turn.Seq
	o109MustCast(t, g, "Discontinuity", "Instant", "{3}{U}{U}{U}", o109Discontinuit, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardIfOwed(t, g)
	if g.Turn.Seq != turn+1 {
		t.Errorf("the turn did not end: turn %d, want %d", g.Turn.Seq, turn+1)
	}
}

func TestS82OldB109DiscontinuityCostsFullPriceOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Discontinuity", TypeLine: "Instant", ManaCost: "{3}{U}{U}{U}",
			OracleID: o109Discontinuit, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Discontinuity was cast for {1}{U} on another player's turn")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "UUUUUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Discontinuity could not be cast for its full cost: %v", err)
	}
}

// --- Predatory Wurm ----------------------------------------------------------------------

func TestS82OldB109PredatoryWurmGrowsWithYourGarruk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	wurm := o109Perm(g, me.ID, "Predatory Wurm", "Creature — Wurm", o109Wurm, 4, 4)
	if p, tough := effectstest.PtOf(t, g, wurm); p != 4 || tough != 4 {
		t.Errorf("alone it is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, wurm, "vigilance") {
		t.Error("the Wurm lacks vigilance")
	}
	o109Perm(g, opp.ID, "Garruk, Theirs", "Legendary Planeswalker — Garruk", "", 0, 0)
	if p, _ := effectstest.PtOf(t, g, wurm); p != 4 {
		t.Errorf("an opponent's Garruk pumped it to %d", p)
	}
	o109Perm(g, me.ID, "Chandra, Mine", "Legendary Planeswalker — Chandra", "", 0, 0)
	if p, _ := effectstest.PtOf(t, g, wurm); p != 4 {
		t.Errorf("a Chandra pumped it to %d", p)
	}
	o109Perm(g, me.ID, "Garruk, Mine", "Legendary Planeswalker — Garruk", "", 0, 0)
	if p, tough := effectstest.PtOf(t, g, wurm); p != 6 || tough != 6 {
		t.Errorf("with your Garruk it is %d/%d, want 6/6", p, tough)
	}
}

// --- Pestilent Haze ----------------------------------------------------------------------

func TestS82OldB109PestilentHazeShrinksEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mine := o109Perm(g, me.ID, "Mine", o109Bear, "", 2, 2)
	theirs := o109Perm(g, opp.ID, "Theirs", o109Bear, "", 4, 4)
	o109Pay(t, g, "BBB")
	o109MustCast(t, g, "Pestilent Haze", "Sorcery", "{1}{B}{B}", o109Haze, game.CastSpellParams{Modes: []int{0}})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := o109CardOn(g, mine); ok {
		t.Error("a 2/2 survived -2/-2")
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the 4/4 is %d/%d, want 2/2", p, tough)
	}
}

func TestS82OldB109PestilentHazeRemovesTwoLoyaltyFromEachPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	bear := o109Perm(g, me.ID, "Bear", o109Bear, "", 2, 2)
	mk := func(owner uuid.UUID, name string, loyalty int) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name,
			TypeLine: "Legendary Planeswalker — Test", Counters: map[string]int{game.CounterLoyalty: loyalty},
			Owner: owner, Controller: owner})
	}
	big, small, mine := mk(opp.ID, "Big", 5), mk(opp.ID, "Small", 1), mk(me.ID, "Mine", 3)
	o109Pay(t, g, "BBB")
	o109MustCast(t, g, "Pestilent Haze", "Sorcery", "{1}{B}{B}", o109Haze, game.CastSpellParams{Modes: []int{1}})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := o109CardOn(g, big); !ok || c.Counters[game.CounterLoyalty] != 3 {
		t.Errorf("Big has %d loyalty (%v), want 3", c.Counters[game.CounterLoyalty], ok)
	}
	if c, ok := o109CardOn(g, mine); !ok || c.Counters[game.CounterLoyalty] != 1 {
		t.Errorf("Mine has %d loyalty (%v), want 1", c.Counters[game.CounterLoyalty], ok)
	}
	if _, ok := o109CardOn(g, small); ok {
		t.Error("a 1-loyalty planeswalker survived")
	}
	if _, ok := o109CardOn(g, bear); !ok {
		t.Error("the creature mode ran too")
	}
}

// --- Kaervek, the Spiteful ---------------------------------------------------------------

func TestS82OldB109KaervekShrinksEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	k := o109Perm(g, me.ID, "Kaervek, the Spiteful", "Legendary Creature — Human Warlock", o109Kaervek, 3, 2)
	mine := o109Perm(g, me.ID, "Mine", o109Bear, "", 3, 3)
	theirs := o109Perm(g, opp.ID, "Theirs", o109Bear, "", 3, 3)
	fragile := o109Perm(g, opp.ID, "Fragile", o109Bear, "", 1, 1)
	if p, tough := effectstest.PtOf(t, g, k); p != 3 || tough != 2 {
		t.Errorf("Kaervek is %d/%d, want 3/2", p, tough)
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tough := effectstest.PtOf(t, g, id); p != 2 || tough != 2 {
			t.Errorf("a 3/3 is %d/%d, want 2/2", p, tough)
		}
	}
	g.RunStateChecksForTest()
	if _, ok := o109CardOn(g, fragile); ok {
		t.Error("a 1/1 survived Kaervek")
	}
}

// --- Sanctum of Tranquil Light -----------------------------------------------------------

func TestS82OldB109SanctumOfTranquilLightCostsLessForEachShrine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	sanctum := effectstest.PushCatalogPermanent(g, me.ID, "Sanctum of Tranquil Light", "Legendary Enchantment — Shrine", o109Sanctum, false)
	victim := o109Perm(g, opp.ID, "Victim", o109Bear, "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	act := func() error {
		return g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Strict: true, Targets: o109Target(victim)})
	}
	// One Shrine (itself): {4}{W}.
	effectstest.FloatForTest(g, me, "WWWW")
	if err := act(); err == nil {
		t.Fatal("activated for {3}{W} with one Shrine")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "WWWWW")
	if err := act(); err != nil {
		t.Fatalf("could not activate for {4}{W} with one Shrine: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := o109CardOn(g, victim); !c.Tapped {
		t.Error("the target was not tapped")
	}
	// A second Shrine, an opponent's Shrine not counting: {3}{W}.
	o109Perm(g, me.ID, "Other Shrine", "Enchantment — Shrine", "", 0, 0)
	o109Perm(g, opp.ID, "Their Shrine", "Enchantment — Shrine", "", 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "WWW")
	if err := act(); err == nil {
		t.Fatal("activated for {2}{W} with two Shrines")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "WWWW")
	if err := act(); err != nil {
		t.Fatalf("could not activate for {3}{W} with two Shrines: %v", err)
	}
}

// --- Riddleform --------------------------------------------------------------------------

func TestS82OldB109RiddleformBecomesASphinxOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rf := effectstest.PushCatalogPermanent(g, me.ID, "Riddleform", "Enchantment", o109Riddleform, false)
	// A creature spell does not trigger it.
	o109Pay(t, g, "GG")
	o109MustCast(t, g, "Bear Cub", "Creature — Bear", "{1}{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AssistCardTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a creature spell asked about Riddleform")
	}
	effectstest.FloatForTest(g, me, "R")
	o109MustCast(t, g, "Shock", "Instant", "{R}", effectstest.LightningBoltOracle, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AssistCardTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("Riddleform did not ask")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, rf); p != 3 || tough != 3 {
		t.Errorf("Riddleform is %d/%d, want 3/3", p, tough)
	}
	types := effectstest.EffectiveTypes(t, g, rf)
	if !effectstest.HasKeyword(types, "Creature") && !effectstest.HasKeyword(types, "creature") {
		t.Errorf("Riddleform's types %v lack Creature", types)
	}
	if !effectstest.HasKeywordOf(g, rf, "flying") {
		t.Error("Riddleform lacks flying")
	}
}

func TestS82OldB109RiddleformScriesForTwoU(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rf := effectstest.PushCatalogPermanent(g, me.ID, "Riddleform", "Enchantment", o109Riddleform, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, rf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {1}{U}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "UUU")
	if err := g.ActivateCatalogAbility(me.ID, rf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

// --- Necromentia -------------------------------------------------------------------------

func TestS82OldB109NecromentiaExilesEveryCopyAndMakesZombiesForTheHandOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mk := func() game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID}
	}
	var hand, yard, lib [2]uuid.UUID
	var other uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			h, y, l := mk(), mk(), mk()
			hand[i], yard[i], lib[i] = h.InstanceID, y.InstanceID, l.InstanceID
			opp.Hand.PushTop(h)
			opp.Graveyard.PushTop(y)
			opp.Library.PushTop(l)
		}
		other = uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	o109Pay(t, g, "BBB")
	o109MustCast(t, g, "Necromentia", "Sorcery", "{1}{B}{B}", o109Necromentia,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Forest"); err == nil {
		t.Fatal("a basic land card's name was accepted")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "lightning bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{hand[0], hand[1], yard[0], yard[1], lib[0], lib[1]} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("a Lightning Bolt is not in exile (zone %v)", z)
		}
	}
	if z := g.FindCardZoneForEffect(other); z == nil || z.Kind != game.ZoneHand {
		t.Error("the Shock left the hand")
	}
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", opp.ID); n != 2 {
		t.Errorf("the opponent has %d Zombies, want 2 (one per card exiled from their hand)", n)
	}
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 0 {
		t.Errorf("the caster has %d Zombies, want 0", n)
	}
}

func TestS82OldB109NecromentiaRefusesYourselfAndAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o109Pay(t, g, "BBB")
	if _, err := o109Cast(t, g, "Necromentia", "Sorcery", "{1}{B}{B}", o109Necromentia,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Error("Necromentia targeted its own controller")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if _, err := o109Cast(t, g, "Necromentia", "Sorcery", "{1}{B}{B}", o109Necromentia,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Error("Necromentia was cast with no mana")
	}
}
