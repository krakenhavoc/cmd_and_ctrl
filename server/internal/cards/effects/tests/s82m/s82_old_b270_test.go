package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b270_test.go — Conspiracy: Take the Crown, S82 older-sets batch
// 270 (#5540): damage and melee. Each test plays the real catalog entry.

const (
	oldB270HurlyOracle      = "2149f9f2-adcf-47d3-9564-326b5c0ced32"
	oldB270HailOracle       = "37429f88-e29b-4b1b-90fd-23a04d703c0a"
	oldB270DevilsOracle     = "62110234-690b-4498-8076-26ba910f9f92"
	oldB270BurnAwayOracle   = "65dcad6b-9a87-4313-bc71-682e1de262d0"
	oldB270SpecterOracle    = "733cfc16-9a16-41e2-b36e-3d7a546c67f0"
	oldB270RuffiansOracle   = "3ed4ee5c-91da-41a3-894c-6181496ec955"
	oldB270WingsOracle      = "73c0a809-fb0e-4773-a385-a6e32feb03cd"
	oldB270FangOracle       = "ae8dffde-e59f-4f4f-8e7f-e7dc34f62e92"
	oldB270LiberatorOracle  = "d91dd07a-93e1-4f88-b5f1-faa56df70ed0"
	oldB270SoulcallerOracle = "dc25bdde-1c73-4c61-99d8-b772aae278cd"
)

func TestOldB270EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB270HurlyOracle: "Hurly-Burly", oldB270HailOracle: "Hail of Arrows", oldB270DevilsOracle: "Gang of Devils",
		oldB270BurnAwayOracle: "Burn Away", oldB270SpecterOracle: "Guul Draz Specter", oldB270RuffiansOracle: "Grenzo's Ruffians",
		oldB270WingsOracle: "Wings of the Guard", oldB270FangOracle: "Fang of the Pack",
		oldB270LiberatorOracle: "Menagerie Liberator", oldB270SoulcallerOracle: "Custodi Soulcaller",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB270Enter puts a catalog creature into owner's hand and moves it onto
// the battlefield through the real zone move, then clears its summoning
// sickness so it can attack this turn.
func oldB270Enter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := effectstest.PlayerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func oldB270Wall(g *game.Game, owner uuid.UUID, name string, typeLine string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, typeLine, 0, 30)
}

// --- Hurly-Burly ----------------------------------------------------------

func TestOldB270HurlyBurlyHitsOnlyTheChosenKind(t *testing.T) {
	for _, flying := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 0, 30)
		air := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Air", TypeLine: "Creature — Bird",
			Power: 0, Toughness: 30, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"}})
		mode := 0
		if flying {
			mode = 1
		}
		effectstest.CastCatalogSpellWithModes(t, g, "Hurly-Burly", "Sorcery", oldB270HurlyOracle, []int{mode})
		effectstest.PassPriorityAroundTable(t, g)
		gotGround, gotAir := effectstest.DamageMarkedOn(g, ground), effectstest.DamageMarkedOn(g, air)
		wantGround, wantAir := 1, 0
		if flying {
			wantGround, wantAir = 0, 1
		}
		if gotGround != wantGround || gotAir != wantAir {
			t.Errorf("flying mode=%v: ground took %d (want %d), flier took %d (want %d)", flying, gotGround, wantGround, gotAir, wantAir)
		}
	}
}

// --- Hail of Arrows -------------------------------------------------------

func TestOldB270HailOfArrowsDividesXAmongAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := oldB270Wall(g, me.ID, "Attacker A", "Creature — Wall")
	b := oldB270Wall(g, me.ID, "Attacker B", "Creature — Wall")
	idle := oldB270Wall(g, me.ID, "Idle", "Creature — Wall")
	effectstest.Pr7bAttack(t, g, opp.ID, a, b)
	id := effectstest.HandCardFull(me, "Hail of Arrows", "Instant", "{X}{W}", oldB270HailOracle, []string{"W"})
	// A creature that is not attacking is not a legal target.
	effectstest.B10AddMana(me, "W", "W", "W", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(idle),
		Distribution: map[uuid.UUID]int{idle: 3}}); err == nil {
		t.Fatal("Hail of Arrows targeted a creature that is not attacking")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.CardRefs(a, b),
		Distribution: map[uuid.UUID]int{a: 1, b: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gotA, gotB := effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b); gotA != 1 || gotB != 2 {
		t.Errorf("damage %d/%d, want 1/2", gotA, gotB)
	}
}

// --- Gang of Devils -------------------------------------------------------

func TestOldB270GangOfDevilsDividesThreeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	devils := oldB270Enter(t, g, me.ID, "Gang of Devils", "Creature — Devil", oldB270DevilsOracle, 3, 3)
	a := oldB270Wall(g, opp.ID, "Wall A", "Creature — Wall")
	b := oldB270Wall(g, opp.ID, "Wall B", "Creature — Wall")
	effectstest.KillCreature(t, g, me.ID, devils)
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if gotA, gotB := effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b); gotA != 1 || gotB != 2 {
		t.Errorf("damage %d/%d, want 1/2", gotA, gotB)
	}
}

// --- Burn Away ------------------------------------------------------------

func oldB270BurnAway(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Burn Away", "Instant", oldB270BurnAwayOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB270BurnAwayExilesTheControllersGraveyardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.P1g2Graveyard(opp, "Creature — Bear", 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 6)
	oldB270BurnAway(t, g, victim)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the 3/6 survived 6 damage")
	}
	if opp.Graveyard.Size() != 0 {
		t.Errorf("their graveyard holds %d cards after the creature died, want 0 (the creature itself is exiled too)", opp.Graveyard.Size())
	}
}

func TestOldB270BurnAwayLeavesTheGraveyardWhileTheCreatureLives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.P1g2Graveyard(opp, "Creature — Bear", 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 7)
	oldB270BurnAway(t, g, victim)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the 3/7 died to 6 damage")
	}
	if opp.Graveyard.Size() != 3 {
		t.Errorf("their graveyard holds %d cards, want 3", opp.Graveyard.Size())
	}
	// It dies later this turn: the delayed trigger still fires.
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("their graveyard holds %d cards after it died later this turn, want 0", opp.Graveyard.Size())
	}
}

func TestOldB270BurnAwayRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Burn Away", "Instant", oldB270BurnAwayOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("Burn Away targeted a player")
	}
}

// --- Guul Draz Specter ----------------------------------------------------

func TestOldB270GuulDrazSpecterGetsPlusThreeWhileAnOpponentHasNoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	specter := oldB270Enter(t, g, me.ID, "Guul Draz Specter", "Creature — Specter", oldB270SpecterOracle, 2, 2)
	for opp.Hand.Size() == 0 {
		t.Fatal("seat 2 has no hand to start with")
	}
	if got := effectstest.EffectivePower(t, g, specter); got != 2 {
		t.Fatalf("power %d with every opponent holding cards, want 2", got)
	}
	for _, c := range append([]game.Card(nil), opp.Hand.Cards...) {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, c.InstanceID); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if got := effectstest.EffectivePower(t, g, specter); got != 5 {
		t.Errorf("power %d with an opponent's hand empty, want 5", got)
	}
}

func TestOldB270GuulDrazSpecterMakesThePlayerItHitDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	specter := oldB270Enter(t, g, me.ID, "Guul Draz Specter", "Creature — Specter", oldB270SpecterOracle, 2, 2)
	before := opp.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, specter, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := before - opp.Hand.Size(); got != 1 {
		t.Errorf("discarded %d cards, want 1", got)
	}
}

// --- Melee ----------------------------------------------------------------

func oldB270PowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.EffectivePower(t, g, id)
}

func TestOldB270MeleeCreaturesPumpPerOpponentAttacked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Wings of the Guard", "Creature — Bird", oldB270WingsOracle, 1, 1},
		{"Menagerie Liberator", "Creature — Human Warrior", oldB270LiberatorOracle, 3, 2},
		{"Grenzo's Ruffians", "Creature — Goblin", oldB270RuffiansOracle, 2, 2},
	} {
		g := effectstest.NewCatalogGame(t)
		me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
		c := oldB270Enter(t, g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
		other := oldB270Wall(g, me.ID, "Other", "Creature — Wall")
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(c, o1.ID); err != nil {
			t.Fatal(err)
		}
		if err := g.DeclareAttacker(other, o2.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := oldB270PowerOf(t, g, c); got != tc.power+2 {
			t.Errorf("%s: power %d after attacking two opponents, want %d", tc.name, got, tc.power+2)
		}
	}
}

func TestOldB270GrenzosRuffiansSharesCombatDamageWithTheOtherOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	c := oldB270Enter(t, g, me.ID, "Grenzo's Ruffians", "Creature — Goblin", oldB270RuffiansOracle, 2, 2)
	l1, l2, l3, mine := o1.Life, o2.Life, o3.Life, me.Life
	effectstest.Pr7bAttack(t, g, o1.ID, c)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Melee: one opponent attacked, so a 3/3.
	if got := l1 - o1.Life; got != 3 {
		t.Errorf("the defender took %d, want 3", got)
	}
	if got := l2 - o2.Life; got != 3 {
		t.Errorf("a second opponent took %d, want 3", got)
	}
	if got := l3 - o3.Life; got != 3 {
		t.Errorf("a third opponent took %d, want 3", got)
	}
	if me.Life != mine {
		t.Errorf("its controller lost %d life", mine-me.Life)
	}
}

// --- Fang of the Pack -----------------------------------------------------

func TestOldB270FangOfThePackGrantsMeleeToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	fang := oldB270Enter(t, g, me.ID, "Fang of the Pack", "Creature — Wolf", oldB270FangOracle, 5, 3)
	other := oldB270Enter(t, g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(other)); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(other, o1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(fang, o2.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB270PowerOf(t, g, other); got != 4 {
		t.Errorf("the granted creature has power %d after attacking two opponents, want 4 (2 + melee 2)", got)
	}
	if got := oldB270PowerOf(t, g, fang); got != 7 {
		t.Errorf("Fang has power %d, want 7 (5 + melee 2)", got)
	}
}

// --- Custodi Soulcaller ---------------------------------------------------

func TestOldB270CustodiSoulcallerReturnsACreatureUpToTheOpponentsAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	c := oldB270Enter(t, g, me.ID, "Custodi Soulcaller", "Creature — Human Cleric", oldB270SoulcallerOracle, 1, 2)
	other := oldB270Wall(g, me.ID, "Other", "Creature — Wall")
	small := effectstest.B38Graveyard(g, me.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	big := effectstest.B38Graveyard(g, me.ID, "Big", "Creature — Bear", "{2}{G}{G}", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(c, o1.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(other, o2.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	// Two opponents attacked: X = 2, so only the mana value 2 creature is a
	// legal target (it is the only one, so the trigger targets it itself).
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(big)); err == nil {
			t.Error("a creature card with mana value 4 was accepted with X = 2")
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(small)); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	if effectstest.TriggerOrderPrompt(g) != nil {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g) // melee and the return
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the mana value 2 creature did not come back")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the mana value 4 creature came back")
	}
}
