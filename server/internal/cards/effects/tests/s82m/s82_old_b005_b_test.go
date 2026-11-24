package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b005_b_test.go — Secret Lair Drop: Planeswalkers, batch 5 (#5462):
// Vraska, Golgari Queen; Nicol Bolas, Planeswalker; The Royal Scions;
// Chandra, Fire Artisan.

const (
	s82OldB5VraskaOracle  = "131096a3-2d53-4ec2-96d3-fcb58e698a13"
	s82OldB5BolasOracle   = "05e9b55e-6329-48ec-b7d7-24c6b9692244"
	s82OldB5ScionsOracle  = "4ecde8d1-e4ec-4bd9-8a80-49885b032557"
	s82OldB5ChandraOracle = "2867d539-6a63-4fae-aae2-a7fdd4dd9f28"
)

func TestS82OldB005BCardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OldB5VraskaOracle: "Vraska, Golgari Queen", s82OldB5BolasOracle: "Nicol Bolas, Planeswalker",
		s82OldB5ScionsOracle: "The Royal Scions", s82OldB5ChandraOracle: "Chandra, Fire Artisan",
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

func s82OldB5Permanent(g *game.Game, owner uuid.UUID, name, typeLine, cost string, p, tough int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: p, Toughness: tough, Owner: owner, Controller: owner})
	return id
}

func s82OldB5Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

// --- Vraska, Golgari Queen ------------------------------------------------

func TestS82OldB005BVraskaPlusTwoSacrificesForLifeAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear was not sacrificed")
	}
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d; want +1 and +1", life, me.Life, hand, me.Hand.Size())
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82OldB005BVraskaPlusTwoDeclinedGivesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the Bear was sacrificed after declining")
	}
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d→%d, hand %d→%d after declining", life, me.Life, hand, me.Hand.Size())
	}
}

func TestS82OldB005BVraskaPlusTwoNeverSacrificesHerself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 4)
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Vraska is the only permanent: there is nothing to sacrifice.
	if _, ok := effectstest.BattlefieldCard(g, vraska); !ok {
		t.Fatal("Vraska sacrificed herself")
	}
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d→%d, hand %d→%d with nothing sacrificed", life, me.Life, hand, me.Hand.Size())
	}
}

func TestS82OldB005BVraskaMinusThreeDestroysACheapNonlandPermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 4)
	cheap := s82OldB5Permanent(g, opp.ID, "Cheap", "Artifact", "{3}", 0, 0)
	dear := s82OldB5Permanent(g, opp.ID, "Dear", "Creature — Bear", "{2}{G}{G}", 4, 4)
	land := s82OldB5Permanent(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	for _, bad := range []uuid.UUID{dear, land} {
		if err := g.ActivateCatalogAbility(me.ID, vraska, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bad)}); err == nil {
			t.Fatalf("activated against an illegal target %v", bad)
		}
	}
	effectstest.S82Activate(t, g, me.ID, vraska, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(cheap)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, cheap); ok {
		t.Error("the mana value 3 permanent survived")
	}
	if got := effectstest.LoyaltyOf(g, vraska); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestS82OldB005BVraskaEmblemMakesAPlayerDamagedByYourCreatureLoseTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 9)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatal("no emblem")
	}
	deal := func(src uuid.UUID, target uuid.UUID, combat bool) {
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: src, Target: target, Amount: 2, Combat: combat})
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	deal(mine, opp.ID, false) // noncombat: nothing
	deal(theirs, me.ID, true) // their creature: nothing
	if opp.Eliminated || me.Eliminated {
		t.Fatal("a player lost without combat damage from one of my creatures")
	}
	deal(mine, opp.ID, true)
	if !opp.Eliminated {
		t.Error("the damaged player did not lose the game")
	}
	if me.Eliminated {
		t.Error("the emblem's owner lost")
	}
}

// --- Nicol Bolas, Planeswalker ---------------------------------------------

func TestS82OldB005BBolasPlusThreeDestroysANoncreaturePermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bolas := effectstest.S82Walker(g, me.ID, "Nicol Bolas, Planeswalker", s82OldB5BolasOracle, 5)
	rock := s82OldB5Permanent(g, opp.ID, "Rock", "Artifact", "{5}", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bolas, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)}); err == nil {
		t.Fatal("+3 targeted a creature")
	}
	effectstest.S82Activate(t, g, me.ID, bolas, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	if got := effectstest.LoyaltyOf(g, bolas); got != 8 {
		t.Errorf("loyalty %d, want 8", got)
	}
}

func TestS82OldB005BBolasMinusTwoStealsACreaturePermanently(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bolas := effectstest.S82Walker(g, me.ID, "Nicol Bolas, Planeswalker", s82OldB5BolasOracle, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, bolas, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Errorf("controller %v, want me", got)
	}
	// Still mine after the turn ends.
	for i := 0; i < 40 && g.Seats[g.Turn.ActiveSeat].ID == me.ID; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Errorf("control ended with the turn: controller %v", got)
	}
}

func TestS82OldB005BBolasUltimateDamagesDiscardsSevenThenSacrificesSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bolas := effectstest.S82Walker(g, me.ID, "Nicol Bolas, Planeswalker", s82OldB5BolasOracle, 9)
	var perms []uuid.UUID
	for i := 0; i < 8; i++ {
		perms = append(perms, s82OldB5Permanent(g, opp.ID, "Rock", "Artifact", "{1}", 0, 0))
	}
	effectstest.AdvanceToMain(t, g)
	for opp.Hand.Size() < 9 {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	life := opp.Life
	effectstest.S82Activate(t, g, me.ID, bolas, 2, game.ActivateAbilityParams{Targets: s82OldB5Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-7 {
		t.Errorf("life %d, want %d", opp.Life, life-7)
	}
	var hand []uuid.UUID
	for i := 0; i < 7; i++ {
		hand = append(hand, opp.Hand.Cards[i].InstanceID)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand...)
	if got := opp.Hand.Size(); got != 2 {
		t.Fatalf("hand %d after discarding seven of nine, want 2", got)
	}
	for _, id := range perms[:7] {
		effectstest.AnswerSacrifice(t, g, opp.ID, id)
	}
	effectstest.PassPriorityAroundTable(t, g)
	left := 0
	for _, id := range perms {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			left++
		}
	}
	if left != 1 {
		t.Errorf("%d of the eight permanents left, want 1", left)
	}
}

func TestS82OldB005BBolasUltimateOnAPlaneswalkerHitsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bolas := effectstest.S82Walker(g, me.ID, "Nicol Bolas, Planeswalker", s82OldB5BolasOracle, 9)
	walker := effectstest.S82Walker(g, opp.ID, "Other Walker", "other-walker", 3)
	effectstest.AdvanceToMain(t, g)
	for opp.Hand.Size() < 7 {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.S82Activate(t, g, me.ID, bolas, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	var hand []uuid.UUID
	for _, c := range opp.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand[:7]...)
	if opp.Hand.Size() != len(hand)-7 {
		t.Errorf("the walker's controller kept %d cards, want %d", opp.Hand.Size(), len(hand)-7)
	}
	// The walker, at 0 loyalty from the damage, is the one permanent left to
	// sacrifice for "sacrifices seven permanents".
	if got := effectstest.LoyaltyOf(g, walker); got != 0 {
		t.Errorf("walker loyalty %d, want 0 after seven damage", got)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, walker); ok {
		t.Error("the planeswalker survived seven damage")
	}
}

// --- The Royal Scions -----------------------------------------------------

func TestS82OldB005BScionsPlusOneLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scions := effectstest.S82Walker(g, me.ID, "The Royal Scions", s82OldB5ScionsOracle, 5)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, scions, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after drawing then discarding", me.Hand.Size(), hand)
	}
	if got := effectstest.LoyaltyOf(g, scions); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82OldB005BScionsSecondPlusOnePumpsAndGrants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scions := effectstest.S82Walker(g, me.ID, "The Royal Scions", s82OldB5ScionsOracle, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, scions, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ApaLive(g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("Bear is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("missing first strike or trample")
	}
}

func TestS82OldB005BScionsUltimateDrawsFourThenDealsDamageForTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	scions := effectstest.S82Walker(g, me.ID, "The Royal Scions", s82OldB5ScionsOracle, 8)
	effectstest.AdvanceToMain(t, g)
	hand, life := me.Hand.Size(), opp.Life
	effectstest.S82Activate(t, g, me.ID, scions, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+4)
	}
	if opp.Life != life-(hand+4) {
		t.Errorf("opponent life %d, want %d (damage equal to a hand of %d)", opp.Life, life-(hand+4), hand+4)
	}
	if _, ok := effectstest.BattlefieldCard(g, scions); ok {
		t.Error("The Royal Scions should have died at 0 loyalty")
	}
}

// --- Chandra, Fire Artisan --------------------------------------------------

// s82OldB5ChandraWith puts Chandra onto the battlefield and gives her her loyalty
// the way a placement does, so the event log knows how many counters she had.
func s82OldB5ChandraWith(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Chandra, Fire Artisan", TypeLine: "Legendary Planeswalker — Chandra",
		OracleID: s82OldB5ChandraOracle, Owner: owner, Controller: owner})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		_ = g.AddCounterForEffect(id, game.CounterLoyalty, loyalty)
	})
	g.RunStateChecksForTest()
	return id
}

func TestS82OldB005BChandraPlusOneExilesTheTopCardAndMayPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	chandra := s82OldB5ChandraWith(g, me.ID, 4)
	top := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	effectstest.AdvanceToMain(t, g)
	lib := me.Library.Size()
	effectstest.S82Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 {
		t.Fatalf("library %d, want %d", me.Library.Size(), lib-1)
	}
	if !g.CastPermissionOnCardByIDForEffect(top).Granted() {
		t.Error("the exiled card is not playable this turn")
	}
	if got := effectstest.LoyaltyOf(g, chandra); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB005BChandraMinusSevenExilesSevenAndTriggersSevenDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	chandra := s82OldB5ChandraWith(g, me.ID, 7)
	effectstest.AdvanceToMain(t, g)
	lib, life := me.Library.Size(), opp.Life
	effectstest.S82Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-7 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib-7)
	}
	if opp.Life != life-7 {
		t.Errorf("opponent life %d, want %d (7 loyalty removed as the cost)", opp.Life, life-7)
	}
}

func TestS82OldB005BChandraTriggersOnDamageThatRemovesLoyaltyOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	chandra := s82OldB5ChandraWith(g, me.ID, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, -2) })
	g.RunStateChecksForTest()
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d for two loyalty counters removed", opp.Life, life-2)
	}
}
