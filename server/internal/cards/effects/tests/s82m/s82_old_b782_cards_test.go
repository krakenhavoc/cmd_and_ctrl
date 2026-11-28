package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b782_cards_test.go — Portal, batch 782 (S82, #5630). Each test plays
// the real catalog entry through the engine.

const (
	oldB782SoulShredOracle    = "2894ef5e-738b-4ece-b143-6662d9453295"
	oldB782FireDragonOracle   = "349d8ef9-e07a-416f-a5b1-2c2be6bb322d"
	oldB782CloakOracle        = "4527c622-cebe-41eb-9178-4895cccffe99"
	oldB782ForkedOracle       = "66107cfd-4bdb-4266-a650-940743555ea4"
	oldB782VampiricOracle     = "b701378f-681b-4840-bd11-1a97c2a675f2"
	oldB782HarshJusticeOracle = "fe4bee2c-f03f-44a4-94a4-55a06bcd0ad8"
)

func TestOldB782EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB782SoulShredOracle: "Soul Shred", oldB782FireDragonOracle: "Fire Dragon",
		oldB782CloakOracle: "Burning Cloak", oldB782ForkedOracle: "Forked Lightning",
		oldB782VampiricOracle: "Vampiric Touch", oldB782HarshJusticeOracle: "Harsh Justice",
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

func TestOldB782SoulShredDamagesANonblackCreatureAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	blackOne := effectstest.SfPush(g, opp.ID, "Black One", "Creature — Test", 1, []string{"B"})
	redOne := effectstest.SfPush(g, opp.ID, "Red One", "Creature — Test", 1, []string{"R"})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Soul Shred", "Sorcery", oldB782SoulShredOracle, effectstest.CardTargets(blackOne)); err == nil {
		t.Fatal("Soul Shred targeted a black creature")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Soul Shred", "Sorcery", oldB782SoulShredOracle, effectstest.CardTargets(redOne))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, redOne).DamageMarked; got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestOldB782VampiricTouchDamagesAnOpponentAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	player := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}} }
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Vampiric Touch", "Sorcery", oldB782VampiricOracle, player(me.ID)); err == nil {
		t.Fatal("Vampiric Touch targeted its own caster")
	}
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Vampiric Touch", "Sorcery", oldB782VampiricOracle, player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Errorf("opponent at %d (want %d), caster at %d (want %d)", opp.Life, oppLife-2, me.Life, life+2)
	}
}

func TestOldB782BurningCloakPumpsThenDamagesTheSameCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Burning Cloak", "Sorcery", oldB782CloakOracle, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tt != 3 {
		t.Errorf("creature is %d/%d, want 4/3", p, tt)
	}
	if got := effectstest.FindCardAnywhere(t, g, bear).DamageMarked; got != 2 {
		t.Errorf("creature has %d damage, want 2", got)
	}
}

func TestOldB782ForkedLightningDividesFourAmongUpToThreeCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 5)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 5)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 5)
	d := effectstest.B12Creature(g, opp.ID, "D", "Creature — Bear", 2, 5)
	if err := effectstest.D1658CastErr(t, g, "Forked Lightning", "Sorcery", oldB782ForkedOracle, game.CastSpellParams{Strict: true,
		Targets: effectstest.CardTargets(a, b, c, d), Distribution: map[uuid.UUID]int{a: 1, b: 1, c: 1, d: 1}}); err == nil {
		t.Fatal("Forked Lightning took four targets")
	}
	if err := effectstest.D1658CastErr(t, g, "Forked Lightning", "Sorcery", oldB782ForkedOracle, game.CastSpellParams{Strict: true,
		Targets: effectstest.CardTargets(a, b), Distribution: map[uuid.UUID]int{a: 3, b: 3}}); err == nil {
		t.Fatal("Forked Lightning accepted a division that is not 4")
	}
	effectstest.D1658Cast(t, g, "Forked Lightning", "Sorcery", oldB782ForkedOracle, game.CastSpellParams{Strict: true,
		Targets: effectstest.CardTargets(a, b, c), Distribution: map[uuid.UUID]int{a: 2, b: 1, c: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 2, b: 1, c: 1, d: 0} {
		if got := effectstest.FindCardAnywhere(t, g, id).DamageMarked; got != want {
			t.Errorf("creature has %d damage, want %d", got, want)
		}
	}
}

func TestOldB782FireDragonDealsDamageEqualToYourMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 7)
	effectstest.CastHoldCreature(t, g, "Fire Dragon", "Creature — Dragon", oldB782FireDragonOracle, 6, 6)
	effectstest.AnswerPickTargetIfOpen(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 3 {
		t.Errorf("the Ogre has %d damage, want 3 (only your Mountains count)", got)
	}
}

func TestOldB782HarshJusticeReflectsCombatDamageToTheAttackersController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	cast := func(p *game.Player) error {
		effectstest.FloatForTest(g, p, "CCW")
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Harsh Justice", TypeLine: "Instant", ManaCost: "{2}{W}",
				OracleID: oldB782HarshJusticeOracle, Owner: p.ID, Controller: p.ID})
		})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast(opp); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, atk)
	if err := cast(g.Seats[2]); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	if err := cast(opp); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	lost := effectstest.SfCombat(t, g)
	if lost[opp.ID] != 3 {
		t.Errorf("the defender lost %d life, want 3", lost[opp.ID])
	}
	if lost[me.ID] != 3 {
		t.Errorf("the attacker's controller lost %d life, want 3", lost[me.ID])
	}
}
