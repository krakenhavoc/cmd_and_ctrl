package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kaitoBaneOracle = "fa43077b-5d8b-4fb1-8dda-2a9185e6715e"

func pushKaito(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Kaito, Bane of Nightmares", TypeLine: "Legendary Planeswalker — Kaito",
		OracleID: kaitoBaneOracle, Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	g.RunStateChecksForTest()
	return id
}

func TestKaitoBaneIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(kaitoBaneOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Kaito: %+v %v", spec, ok)
	}
}

func TestKaitoIsAHexproof3x4NinjaCreatureOnlyOnYourTurnWithLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	k := pushKaito(g, me.ID, 4)
	c, _ := battlefieldCard(g, k)
	eff := c.Effective()
	if !c.IsCreature() || c.IsPlaneswalker() || !c.HasSubtype("Ninja") || eff.Power != 3 || eff.Toughness != 4 {
		t.Fatalf("on my turn: types %v subtypes %v %d/%d, want a 3/4 Ninja creature that is not a planeswalker", eff.Types, eff.Subtypes, eff.Power, eff.Toughness)
	}
	if !game.HasKeyword(&c, "hexproof") {
		t.Error("no hexproof on my turn")
	}
	// No loyalty counters: still a planeswalker.
	k2 := pushKaito(g, me.ID, 0)
	c2, _ := battlefieldCard(g, k2)
	if c2.IsCreature() {
		t.Error("a Kaito with no loyalty counters is a creature")
	}
	// Opponent's turn: a planeswalker.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	c3, _ := battlefieldCard(g, k)
	if c3.IsCreature() || !c3.IsPlaneswalker() {
		t.Errorf("on an opponent's turn: types %v, want a planeswalker", c3.Effective().Types)
	}
}

func TestKaitoPlusOneEmblemPumpsNinjas(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	k := pushKaito(g, me.ID, 4)
	ninja := b12Creature(g, me.ID, "Ninja", "Creature — Human Ninja", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, k, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	n, _ := battlefieldCard(g, ninja)
	b, _ := battlefieldCard(g, bear)
	kc, _ := battlefieldCard(g, k)
	if e := n.Effective(); e.Power != 3 || e.Toughness != 3 {
		t.Errorf("Ninja %d/%d, want 3/3", e.Power, e.Toughness)
	}
	if e := b.Effective(); e.Power != 2 {
		t.Errorf("Bear %d/%d, want unchanged", e.Power, e.Toughness)
	}
	if e := kc.Effective(); e.Power != 4 || e.Toughness != 5 {
		t.Errorf("Kaito %d/%d, want 4/5 (he is a Ninja)", e.Power, e.Toughness)
	}
}

func TestKaitoMinusTwoTapsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	k := pushKaito(g, me.ID, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, k, 3, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	passPriorityAroundTable(t, g)
	b, _ := battlefieldCard(g, bear)
	if !b.Tapped || counterOn(g, bear, game.CounterStun) != 2 {
		t.Errorf("tapped=%v stun=%d, want tapped with two stun counters", b.Tapped, counterOn(g, bear, game.CounterStun))
	}
}

func TestKaitoZeroSurveilsThenDrawsPerOpponentWhoLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	k := pushKaito(g, me.ID, 4)
	ids := seedLibrary(me, "A", "B", "C", "D", "E")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -1) })
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[2].ID, -2) })
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, k, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("0: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (two opponents lost life)", got)
	}
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card is not in the graveyard")
	}
}

func TestKaitoNinjutsuEntersTappedAndAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	kaito := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: kaito, Name: "Kaito, Bane of Nightmares", TypeLine: "Legendary Planeswalker — Kaito", OracleID: kaitoBaneOracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, kaito, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, kaito)
	if !ok || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("Kaito on battlefield=%v tapped=%v attacking=%v, want tapped and attacking the defender", ok, c.Tapped, c.AttackingTarget)
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	if got := life - opp.Life; got != 3 {
		t.Errorf("defender lost %d, want 3 from the 3/4 Ninja", got)
	}
}
