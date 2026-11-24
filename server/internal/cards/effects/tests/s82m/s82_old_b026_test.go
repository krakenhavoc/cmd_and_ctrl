package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB26ChaniOID     = "029f91b8-f5da-4ebb-835d-b6e42cffceaa"
	oldB26MuadDibOID   = "61a6317a-6abb-4899-a7e4-36035042f8e4"
	oldB26BlazeOID     = "74b08a70-b0bb-4340-98a0-b1d5b7c9d2cc"
	oldB26RoilOID      = "7e0d2965-bbcd-4214-b1b6-c0d2601ce078"
	oldB26OmnathOID    = "cd133d30-51ff-4114-a7d7-029345f0f0d7"
	oldB26DeathrenOID  = "05ac6e5c-b773-45ed-9c74-3cd4b05bc287"
	oldB26TruthOID     = "7a8baaf9-e21f-41a0-9f15-80764f6e5e68"
	oldB26ShovelOID    = "9c298c7b-9c63-4a66-a0fc-977491b623c7"
	oldB26WarPeaceOID  = "b10ad415-0aae-4d70-82e6-a4055a40cbe4"
	oldB26SunforgerOID = "d1421070-a3cc-4af3-bf91-0f97372f4161"
	oldB26Bear         = "Creature — Bear"
	oldB26Equipment    = "Artifact — Equipment"
)

func TestOldB26CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB26ChaniOID: "Chani, Blade of Sietch Tabr", oldB26MuadDibOID: "Muad'Dib, Prophet Foretold",
		oldB26BlazeOID: "Searing Blaze", oldB26RoilOID: "Roil Elemental",
		oldB26OmnathOID: "Omnath, Locus of Creation", oldB26DeathrenOID: "Deathrender",
		oldB26TruthOID: "Sword of Truth and Justice", oldB26ShovelOID: "Gisa's Favorite Shovel",
		oldB26WarPeaceOID: "Sword of War and Peace", oldB26SunforgerOID: "Sunforger",
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

func oldB26Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

// oldB26PlayLand plays a land of the given type line from the active
// player's hand through the real play path, so landfall sees it.
func oldB26PlayLand(t *testing.T, g *game.Game, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Land", TypeLine: typeLine, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play land: %v", err)
	}
	return id
}

func oldB26Target(id uuid.UUID) *game.TargetRef {
	return &game.TargetRef{Kind: game.TargetCard, ID: id}
}

func oldB26Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s left the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

// --- Chani, Blade of Sietch Tabr ------------------------------------------------

func TestOldB26ChaniHasDoubleStrikeAndAnExtraLandDrop(t *testing.T) {
	spec, _ := effects.Lookup(oldB26ChaniOID)
	if spec.AdditionalLandPlays != 1 {
		t.Errorf("additional land plays = %d, want 1", spec.AdditionalLandPlays)
	}
	g := effectstest.NewCatalogGame(t)
	chani := oldB26Push(g, g.Seats[0].ID, "Chani, Blade of Sietch Tabr", "Legendary Creature — Human Warrior", oldB26ChaniOID, 3, 3)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.HasEffectiveKeyword(t, g, chani, "double strike") {
		t.Error("Chani has no double strike")
	}
}

func TestOldB26ChaniLandfallPutsACounterAndADesertGivesKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB26Push(g, me.ID, "Chani, Blade of Sietch Tabr", "Legendary Creature — Human Warrior", oldB26ChaniOID, 3, 3)
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	oldB26PlayLand(t, g, "Land — Forest")
	effectstest.AssistCardSettle(t, g, me.ID, oldB26Target(bear), true)
	if got := oldB26Counters(t, g, bear); got != 1 {
		t.Fatalf("counters = %d, want 1", got)
	}
	for _, kw := range []string{"vigilance", "hexproof", "haste"} {
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a Forest granted %s", kw)
		}
	}
	oldB26PlayLand(t, g, "Land — Desert")
	effectstest.AssistCardSettle(t, g, me.ID, oldB26Target(bear), true)
	if got := oldB26Counters(t, g, bear); got != 2 {
		t.Fatalf("counters = %d, want 2", got)
	}
	for _, kw := range []string{"vigilance", "hexproof", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("a Desert did not grant %s", kw)
		}
	}
}

// --- Muad'Dib, Prophet Foretold --------------------------------------------------

func TestOldB26MuadDibLandfallMillIsOptional(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		oldB26Push(g, me.ID, "Muad'Dib, Prophet Foretold", "Legendary Creature — Human Noble Warrior", oldB26MuadDibOID, 3, 3)
		before := me.Graveyard.Size()
		oldB26PlayLand(t, g, "Land — Forest")
		effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
		want := 0
		if yes {
			want = 1
		}
		if got := me.Graveyard.Size() - before; got != want {
			t.Errorf("answer %v: milled %d, want %d", yes, got, want)
		}
	}
}

func TestOldB26MuadDibPlaysTheTopCardOnlyWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB26Push(g, me.ID, "Muad'Dib, Prophet Foretold", "Legendary Creature — Human Noble Warrior", oldB26MuadDibOID, 3, 3)
	effectstest.AdvanceToMain(t, g)
	me.LandDropsPerTurn += 2
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Land", TypeLine: "Land — Forest", Owner: me.ID, Controller: me.ID})
	for _, tl := range []string{"Creature — Bear", "Instant", "Artifact"} {
		effectstest.PushGraveyardPermanent(me, "Filler", tl, "")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("played the top card with three card types in the graveyard")
	}
	effectstest.PushGraveyardPermanent(me, "Old Land", "Land", "")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("with four card types in the graveyard the top card should be playable: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Error("the land played from the top of the library is not on the battlefield")
	}
}

// --- Searing Blaze --------------------------------------------------------------

func oldB26Blaze(t *testing.T, landfall bool) (g *game.Game, opp *game.Player, creature uuid.UUID, startLife int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creature = oldB26Push(g, opp.ID, "Their Bear", oldB26Bear, "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if landfall {
		oldB26PlayLand(t, g, "Land — Mountain")
	}
	startLife = effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Searing Blaze", "Instant", oldB26BlazeOID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: creature},
	})
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	return g, opp, creature, startLife
}

func TestOldB26SearingBlazeDealsOneAndOneWithoutLandfall(t *testing.T) {
	g, opp, creature, start := oldB26Blaze(t, false)
	if got := start - effectstest.LifeOf(g, opp.ID); got != 1 {
		t.Errorf("the player lost %d, want 1", got)
	}
	if !g.Battlefield.Contains(creature) {
		t.Error("a 2/2 died to 1 damage")
	}
}

func TestOldB26SearingBlazeDealsThreeAndThreeWithLandfall(t *testing.T) {
	g, opp, creature, start := oldB26Blaze(t, true)
	if got := start - effectstest.LifeOf(g, opp.ID); got != 3 {
		t.Errorf("the player lost %d, want 3", got)
	}
	if g.Battlefield.Contains(creature) {
		t.Error("the creature survived 3 damage")
	}
}

func TestOldB26SearingBlazeSecondTargetMustBelongToTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	theirs := oldB26Push(g, other.ID, "Other Bear", oldB26Bear, "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Searing Blaze", "Instant", oldB26BlazeOID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: theirs},
	})
	if err == nil {
		t.Error("a creature another player controls was accepted as the second target")
	}
}

// --- Roil Elemental -------------------------------------------------------------

func TestOldB26RoilElementalStealsForAsLongAsYouControlIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	roil := oldB26Push(g, me.ID, "Roil Elemental", "Creature — Elemental", oldB26RoilOID, 3, 2)
	theirs := oldB26Push(g, opp.ID, "Their Bear", oldB26Bear, "", 2, 2)
	oldB26PlayLand(t, g, "Land — Island")
	effectstest.AssistCardSettle(t, g, me.ID, oldB26Target(theirs), true)
	if got := effectstest.ControllerOf(t, g, theirs); got != me.ID {
		t.Fatalf("the Bear is controlled by %s, want me", got)
	}
	effectstest.KillCreature(t, g, me.ID, roil)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Error("the Bear did not go home when the Elemental left")
	}
}

func TestOldB26RoilElementalStealIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB26Push(g, me.ID, "Roil Elemental", "Creature — Elemental", oldB26RoilOID, 3, 2)
	theirs := oldB26Push(g, opp.ID, "Their Bear", oldB26Bear, "", 2, 2)
	oldB26PlayLand(t, g, "Land — Island")
	effectstest.AssistCardSettle(t, g, me.ID, oldB26Target(theirs), false)
	if got := effectstest.ControllerOf(t, g, theirs); got != opp.ID {
		t.Error("declining still took the Bear")
	}
}

// --- Omnath, Locus of Creation --------------------------------------------------

func TestOldB26OmnathDrawsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Omnath, Locus of Creation", "Legendary Creature — Elemental", oldB26OmnathOID, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	// The cast removes one card from hand; the trigger draws one.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestOldB26OmnathLandfallDoesADifferentThingEachTime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB26Push(g, me.ID, "Omnath, Locus of Creation", "Legendary Creature — Elemental", oldB26OmnathOID, 4, 4)
	theirWalker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test", Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterLoyalty: 6},
	})
	myWalker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Walker", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 6},
	})
	life := effectstest.LifeOf(g, me.ID)
	oppLife := effectstest.LifeOf(g, opp.ID)

	oldB26PlayLand(t, g, "Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID) - life; got != 4 {
		t.Errorf("first resolution: gained %d, want 4", got)
	}
	if me.ManaPool.Total() != 0 || effectstest.LifeOf(g, opp.ID) != oppLife {
		t.Error("the first resolution did more than gain life")
	}

	oldB26PlayLand(t, g, "Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []string{"R", "G", "W", "U"} {
		if me.ManaPool.CountColor(c) != 1 {
			t.Errorf("second resolution: %d %s mana, want 1", me.ManaPool.CountColor(c), c)
		}
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != 4 {
		t.Errorf("second resolution changed life: %d", got)
	}

	oldB26PlayLand(t, g, "Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - effectstest.LifeOf(g, opp.ID); got != 4 {
		t.Errorf("third resolution: the opponent lost %d, want 4", got)
	}
	if got := effectstest.LoyaltyOf(g, theirWalker); got != 2 {
		t.Errorf("their planeswalker has %d loyalty, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, myWalker); got != 6 {
		t.Errorf("my planeswalker has %d loyalty, want 6", got)
	}

	oldB26PlayLand(t, g, "Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - effectstest.LifeOf(g, opp.ID); got != 4 {
		t.Errorf("a fourth resolution did something: the opponent has lost %d", got)
	}
}

// --- Deathrender ----------------------------------------------------------------

func oldB26Equip(t *testing.T, g *game.Game, me *game.Player, sword uuid.UUID, mana []string, creature uuid.UUID, idx int) {
	t.Helper()
	effectstest.B06AddMana(me, mana...)
	effectstest.B16Activate(t, g, me.ID, sword, idx, game.ActivateAbilityParams{Targets: effectstest.CardRefs(creature)})
}

func TestOldB26DeathrenderPutsACreatureFromHandAndAttaches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	sword := oldB26Push(g, me.ID, "Deathrender", oldB26Equipment, oldB26DeathrenOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, sword, []string{"C", "C"}, bear, 0)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+2: power %d, want 4", got)
	}
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Bear", TypeLine: oldB26Bear, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("Deathrender offered no creature card from hand")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{inHand}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(inHand) {
		t.Fatal("the creature was not put onto the battlefield")
	}
	if !effectstest.AttachedTo(t, g, sword, inHand) {
		t.Error("Deathrender was not attached to the new creature")
	}
}

func TestOldB26DeathrenderMayPutNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	sword := oldB26Push(g, me.ID, "Deathrender", oldB26Equipment, oldB26DeathrenOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, sword, []string{"C", "C"}, bear, 0)
	inHand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Bear", TypeLine: oldB26Bear, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("Deathrender offered no creature card from hand")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(inHand) || !me.Hand.Contains(inHand) {
		t.Error("declining still put the creature onto the battlefield")
	}
}

// --- Sword of Truth and Justice -------------------------------------------------

func TestOldB26SwordOfTruthAndJusticeCountersThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	sword := oldB26Push(g, me.ID, "Sword of Truth and Justice", oldB26Equipment, oldB26TruthOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, sword, []string{"C", "C"}, bear, 0)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+2: power %d, want 4", got)
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
			t.Fatalf("choosing the creature: %v", err)
		}
	}
	c := effectstest.ProliferatePrompt(g)
	if c == nil {
		t.Fatal("no proliferate prompt after the counter")
	}
	if got := oldB26Counters(t, g, bear); got != 1 {
		t.Fatalf("counters before proliferate = %d, want 1", got)
	}
	effectstest.AnswerProliferate(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB26Counters(t, g, bear); got != 2 {
		t.Errorf("counters after proliferate = %d, want 2", got)
	}
}

// --- Sword of War and Peace -----------------------------------------------------

func TestOldB26SwordOfWarAndPeaceDamagesByTheirHandAndGainsByYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	sword := oldB26Push(g, me.ID, "Sword of War and Peace", oldB26Equipment, oldB26WarPeaceOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, sword, []string{"C", "C"}, bear, 0)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "protection from red") || !effectstest.HasEffectiveKeyword(t, g, bear, "protection from white") {
		t.Error("the equipped creature lacks protection from red and from white")
	}
	theirHand := opp.Hand.Size()
	myHand := me.Hand.Size()
	life, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - effectstest.LifeOf(g, opp.ID); got != theirHand {
		t.Errorf("the opponent lost %d, want their hand size %d", got, theirHand)
	}
	if got := effectstest.LifeOf(g, me.ID) - life; got != myHand {
		t.Errorf("I gained %d, want my hand size %d", got, myHand)
	}
}

// --- Gisa's Favorite Shovel -----------------------------------------------------

func oldB26Walkers(g *game.Game, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == "Walker" && effects.IsToken(c) {
				n++
			}
		}
	})
	return n
}

func TestOldB26ShovelMakesTheDefenderSacrificeAndYouGetAWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	shovel := oldB26Push(g, me.ID, "Gisa's Favorite Shovel", "Legendary Artifact — Equipment", oldB26ShovelOID, 0, 0)
	victim := oldB26Push(g, opp.ID, "Their Bear", oldB26Bear, "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, shovel, []string{"C", "C", "C", "C"}, bear, 0)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("+2/+0: power %d, want 4", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the equipped creature has no menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSacrifice, opp.ID); c != nil {
		effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the defending player's creature was not sacrificed")
	}
	if got := oldB26Walkers(g, me.ID); got != 1 {
		t.Errorf("Walker tokens = %d, want 1", got)
	}
}

func TestOldB26ShovelMakesNoWalkerWhenTheDefenderHasNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	shovel := oldB26Push(g, me.ID, "Gisa's Favorite Shovel", "Legendary Artifact — Equipment", oldB26ShovelOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	oldB26Equip(t, g, me, shovel, []string{"C", "C", "C", "C"}, bear, 0)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB26Walkers(g, me.ID); got != 0 {
		t.Errorf("Walker tokens = %d, want 0 when nothing was sacrificed", got)
	}
}

// --- Sunforger ------------------------------------------------------------------

func oldB26SunforgerTable(t *testing.T) (g *game.Game, me *game.Player, sunforger, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	bear = oldB26Push(g, me.ID, "Bear", oldB26Bear, "", 2, 2)
	sunforger = oldB26Push(g, me.ID, "Sunforger", oldB26Equipment, oldB26SunforgerOID, 0, 0)
	effectstest.AdvanceToMain(t, g)
	return g, me, sunforger, bear
}

func oldB26Library(me *game.Player, name, typeLine, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB26SunforgerGivesPlusFourAndRequiresBeingAttached(t *testing.T) {
	g, me, sunforger, bear := oldB26SunforgerTable(t)
	oldB26Equip(t, g, me, sunforger, []string{"C", "C", "C"}, bear, 1)
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("+4/+0: power %d, want 6", got)
	}
	// Unattached, the ability can't be activated: its cost unattaches it.
	effectstest.B06AddMana(me, "R", "W")
	loose := oldB26Push(g, me.ID, "Loose Sunforger", oldB26Equipment, oldB26SunforgerOID, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, loose, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an unattached Sunforger activated its ability")
	}
}

func TestOldB26SunforgerUnattachesSearchesAndCastsFree(t *testing.T) {
	g, me, sunforger, bear := oldB26SunforgerTable(t)
	oldB26Equip(t, g, me, sunforger, []string{"C", "C", "C"}, bear, 1)
	_ = oldB26Library(me, "Blue Instant", "Instant", "{U}", "U")
	_ = oldB26Library(me, "Big Red Instant", "Instant", "{3}{R}{R}", "R")
	_ = oldB26Library(me, "Red Sorcery", "Sorcery", "{R}", "R")
	bolt := oldB26Library(me, "Test Bolt", "Instant", "{R}", "R")
	effectstest.B06AddMana(me, "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, sunforger, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating: %v", err)
	}
	if effectstest.AttachedTo(t, g, sunforger, bear) {
		t.Error("the cost did not unattach the Equipment")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d after unattaching, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, bolt)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the found card was not offered as a free cast")
	}
	if effectstest.KeywordCastCard(prompt) != bolt {
		t.Errorf("offered %s, want the found instant", effectstest.KeywordCastCard(prompt))
	}

	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting the offered card: %v", err)
	}
	if !g.Stack.Contains(bolt) && !me.Graveyard.Contains(bolt) {
		t.Error("the instant was not cast")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the cast cost mana: pool %d", me.ManaPool.Total())
	}
}

func TestOldB26SunforgerFindingNothingOrDecliningStillShuffles(t *testing.T) {
	g, me, sunforger, bear := oldB26SunforgerTable(t)
	oldB26Equip(t, g, me, sunforger, []string{"C", "C", "C"}, bear, 1)
	bolt := oldB26Library(me, "Test Bolt", "Instant", "{R}", "R")
	effectstest.B06AddMana(me, "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, sunforger, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activating: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, bolt)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(bolt) {
		t.Error("the declined instant left the library")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Error("the declined instant could still be cast for free later")
	}
}
