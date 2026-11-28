package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b784_cards_test.go — Portal, batch 784 (S82, #5630). Each test plays
// the real catalog entry through the engine.

const (
	oldB784DawnOracle      = "54ea46ea-7c83-44a9-85b0-eff9745c6ffa"
	oldB784FruitionOracle  = "5f3ef680-90f4-487f-b44b-db25e29e57ce"
	oldB784StarlightOracle = "6237896e-f033-41b2-9943-670f1becb582"
	oldB784StandOracle     = "05e61628-cdae-4664-b866-be438fb3a6ba"
	oldB784ChargeOracle    = "1f51fea2-ce19-44ef-a330-65cc0fafcd64"
	oldB784SeasOracle      = "7421b711-81a4-4042-b024-f55bf9ff203f"
	oldB784FlowOracle      = "91c0a76e-3992-437f-b85a-97b0b4adbb84"
	oldB784ThingOracle     = "9542bd7f-bb99-4b59-a4e3-b88ed9f798bf"
)

func TestOldB784EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB784DawnOracle: "Renewing Dawn", oldB784FruitionOracle: "Fruition", oldB784StarlightOracle: "Starlight",
		oldB784StandOracle: "Defiant Stand", oldB784ChargeOracle: "Valorous Charge", oldB784SeasOracle: "Boiling Seas",
		oldB784FlowOracle: "Lava Flow", oldB784ThingOracle: "Thing from the Deep",
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

func oldB784Players(g *game.Game) (*game.Player, *game.Player) {
	me := g.Seats[g.Turn.ActiveSeat]
	return me, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB784PlayerTarget(p *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: p.ID}}
}

func TestOldB784RenewingDawnGainsTwoPerOpponentsMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB784Players(g)
	effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Renewing Dawn", "Sorcery", oldB784DawnOracle, oldB784PlayerTarget(me)); err == nil {
		t.Fatal("Renewing Dawn targeted its own caster")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Renewing Dawn", "Sorcery", oldB784DawnOracle, oldB784PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d (two Mountains, not mine)", me.Life, life+4)
	}
}

func TestOldB784StarlightGainsThreePerBlackCreatureTheOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB784Players(g)
	effectstest.SfPush(g, opp.ID, "Black A", "Creature — Test", 1, []string{"B"})
	effectstest.SfPush(g, opp.ID, "Black B", "Creature — Test", 1, []string{"B"})
	effectstest.SfPush(g, opp.ID, "Red", "Creature — Test", 1, []string{"R"})
	effectstest.SfPush(g, me.ID, "Mine", "Creature — Test", 1, []string{"B"})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Starlight", "Sorcery", oldB784StarlightOracle, oldB784PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d (two black creatures)", me.Life, life+6)
	}
}

func TestOldB784FruitionCountsEveryForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB784Players(g)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, opp.ID, "Plains", "Basic Land — Plains")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Fruition", "Sorcery", oldB784FruitionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (three Forests)", me.Life, life+3)
	}
}

func TestOldB784DefiantStandPumpsAndUntapsAfterYouWereAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	blocker := effectstest.B12Creature(g, opp.ID, "Defender", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, blocker).Tapped = true })
	cast := func(p *game.Player, target uuid.UUID) error {
		effectstest.FloatForTest(g, p, "CW")
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Defiant Stand", TypeLine: "Instant", ManaCost: "{1}{W}",
				OracleID: oldB784StandOracle, Owner: p.ID, Controller: p.ID})
		})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(target)})
	}
	if err := cast(opp, blocker); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, atk)
	if err := cast(g.Seats[2], blocker); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	if err := cast(opp, blocker); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, blocker), effectstest.EffectiveToughness(t, g, blocker); p != 3 || tt != 5 {
		t.Errorf("the creature is %d/%d, want 3/5", p, tt)
	}
	if effectstest.BattlefieldCardPtr(t, g, blocker).Tapped {
		t.Error("the creature was not untapped")
	}
}

func TestOldB784ValorousChargePumpsEveryWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB784Players(g)
	mine := effectstest.SfPush(g, me.ID, "My White", "Creature — Test", 1, []string{"W"})
	theirs := effectstest.SfPush(g, opp.ID, "Their White", "Creature — Test", 1, []string{"W"})
	red := effectstest.SfPush(g, opp.ID, "Red", "Creature — Test", 1, []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Valorous Charge", "Sorcery", oldB784ChargeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.SfPush(g, me.ID, "Late White", "Creature — Test", 1, []string{"W"})
	for id, want := range map[uuid.UUID]int{mine: 3, theirs: 3, red: 1, late: 1} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
}

func TestOldB784BoilingSeasDestroysEveryIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB784Players(g)
	a := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	b := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	other := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Boiling Seas", "Sorcery", oldB784SeasOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !g.Battlefield.Contains(other) {
		t.Error("the Islands were not all destroyed, or the Swamp was")
	}
}

func TestOldB784LavaFlowDestroysACreatureOrALandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB784Players(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	art := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lava Flow", "Sorcery", oldB784FlowOracle, effectstest.CardTargets(art)); err == nil {
		t.Fatal("Lava Flow targeted an artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Lava Flow", "Sorcery", oldB784FlowOracle, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lava Flow", "Sorcery", oldB784FlowOracle, effectstest.CardTargets(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) || !g.Battlefield.Contains(art) {
		t.Error("the creature and land were not both destroyed, or the artifact was")
	}
}

func oldB784ThingAttacks(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.B12Push(g, me.ID, "Thing from the Deep", "Creature — Leviathan", oldB784ThingOracle, 9, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB784ThingFromTheDeepSacrificesAnIslandToKeepAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	swamp := effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	id := oldB784ThingAttacks(t, g, me, opp)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerChooseCards(t, g, me.ID, island)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || g.Battlefield.Contains(island) || !g.Battlefield.Contains(swamp) {
		t.Error("the Island was not sacrificed in place of the Leviathan, or the Swamp was")
	}
}

func TestOldB784ThingFromTheDeepIsSacrificedWhenYouChooseSoOrHaveNoIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	id := oldB784ThingAttacks(t, g, me, opp)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !g.Battlefield.Contains(island) {
		t.Error("the Leviathan stayed, or the Island was sacrificed anyway")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	id2 := oldB784ThingAttacks(t, g2, me2, opp2)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(id2) {
		t.Error("the Leviathan stayed with no Island to sacrifice")
	}
}
