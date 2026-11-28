package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b685_test.go — Mercadian Masques Auras, batch 685 (#5619). Every card is
// the real catalog entry played through the engine.

const (
	m685FlamingSword  = "05c62f91-2a5b-4cba-8e66-78fb370ea409"
	m685Insubord      = "0d3e843e-eeb4-421e-b55d-ceb21f51a080"
	m685Ferocity      = "1b1a0f35-3126-4b88-90aa-c56e67850a9d"
	m685TigerClaws    = "2ef8ebc9-4f95-42f8-86e6-85eff0b8f021"
	m685Inviolability = "3f336198-2425-447a-88b4-9be21ef7e90a"
	m685FalseDemise   = "50ffa9f0-fb3b-4de4-b463-6d7ee8b5e93a"
	m685SoulChannel   = "7041b3d5-c61a-4b94-811c-c5eb9f1ba664"
	m685CaveSense     = "7d2fcdac-c8d5-43ca-81bc-130744ed8d3d"
)

func TestS82OldB685EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m685FlamingSword: "Flaming Sword", m685Insubord: "Insubordination", m685Ferocity: "Ferocity",
		m685TigerClaws: "Tiger Claws", m685Inviolability: "Inviolability", m685FalseDemise: "False Demise",
		m685SoulChannel: "Soul Channeling", m685CaveSense: "Cave Sense",
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

// m685Cast casts an Aura from the active player's hand onto host, strictly paid from a full pool, and
// lets it resolve.
func m685Cast(t *testing.T, g *game.Game, name, cost, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "{2}{G}{G}{B}{B}{U}{R}{W}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Enchantment — Aura", ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func m685Attached(g *game.Game, aura, host uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == aura {
			return c.AttachedTo.ID == host
		}
	}
	return false
}

func TestS82OldB685FlamingSwordAndTigerClawsPumpAndGrantAKeyword(t *testing.T) {
	for _, tc := range []struct {
		name, cost, oracle, keyword string
		p, th                       int
	}{
		{"Flaming Sword", "{1}{R}", m685FlamingSword, "first strike", 3, 2},
		{"Tiger Claws", "{2}{G}", m685TigerClaws, "trample", 3, 3},
		{"Cave Sense", "{1}{R}", m685CaveSense, "mountainwalk", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			bear := effectstest.SeedBear(g, g.Seats[1].ID)
			aura := m685Cast(t, g, tc.name, tc.cost, tc.oracle, bear)
			if !m685Attached(g, aura, bear) {
				t.Fatal("the Aura did not attach")
			}
			if p, th := effectstest.PtOf(t, g, bear); p != tc.p || th != tc.th {
				t.Errorf("enchanted Bear is %d/%d, want %d/%d", p, th, tc.p, tc.th)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the enchanted creature has no %s", tc.keyword)
			}
			g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(aura) })
			if p, th := effectstest.PtOf(t, g, bear); p != 2 || th != 2 {
				t.Errorf("after the Aura left the Bear is %d/%d, want 2/2", p, th)
			}
		})
	}
}

func TestS82OldB685InviolabilityPreventsAllDamageToTheEnchantedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	other := effectstest.SeedBear(g, opp.ID)
	m685Cast(t, g, "Inviolability", "{1}{W}", m685Inviolability, bear)
	effectstest.DealDamageTo(t, g, uuid.New(), bear, 5)
	effectstest.DealDamageTo(t, g, uuid.New(), other, 2)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("the enchanted Bear died to damage")
	}
	if g.Battlefield.Contains(other) {
		t.Error("an unenchanted Bear survived 2 damage")
	}
}

func TestS82OldB685FalseDemiseReturnsTheCreatureUnderTheAurasController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	m685Cast(t, g, "False Demise", "{2}{U}", m685FalseDemise, bear)
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Bear" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("the Bear did not return to the battlefield")
	}
	if back.Controller != me.ID {
		t.Errorf("the Bear returned under %v, want the Aura's controller", back.Controller)
	}
}

func TestS82OldB685SoulChannelingPaysTwoLifeToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	aura := m685Cast(t, g, "Soul Channeling", "{2}{B}", m685SoulChannel, bear)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("the Aura's controller is at %d, want %d", me.Life, life-2)
	}
	if n := effectstest.RegenShieldsOn(t, g, bear); n != 1 {
		t.Fatalf("the enchanted creature has %d regeneration shields, want 1", n)
	}
	effectstest.DestroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regenerated Bear was destroyed")
	}
}

func TestS82OldB685InsubordinationDamagesTheControllerUnlessTheCreatureAttacked(t *testing.T) {
	// The creature's controller is the opponent (seat 1): its end step comes on their turn.
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.SeedBear(g, opp.ID)
	m685Cast(t, g, "Insubordination", "{B}{B}", m685Insubord, bear)
	life := opp.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the creature's controller lost life in the Aura controller's end step: %d", opp.Life)
	}
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("a creature that did not attack: its controller is at %d, want %d", opp.Life, life-2)
	}
}

func TestS82OldB685InsubordinationSparesAControllerWhoseCreatureAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Aura's controller enchants their own creature, which attacks.
	bear := effectstest.SeedBear(g, me.ID)
	m685Cast(t, g, "Insubordination", "{B}{B}", m685Insubord, bear)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the controller of an attacking creature took damage: %d, want %d", me.Life, life)
	}
}

func TestS82OldB685FerocityAsksOnBothBlockingAndBeingBlocked(t *testing.T) {
	for _, tc := range []struct {
		name         string
		enchantMine  bool
		wantBlocker  bool
		answerYes    bool
		wantCounters int
	}{
		{"becomes blocked, yes", true, false, true, 1},
		{"becomes blocked, no", true, false, false, 0},
		{"blocks, yes", false, true, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := effectstest.SeedBear(g, me.ID)
			blocker := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
			host := attacker
			if tc.wantBlocker {
				host = blocker
			}
			m685Cast(t, g, "Ferocity", "{1}{G}", m685Ferocity, host)
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(blocker, attacker); err != nil {
				t.Fatalf("DeclareBlocker: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.answerYes)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, host, game.CounterPlusOne); n != tc.wantCounters {
				t.Errorf("%d +1/+1 counters on the enchanted creature, want %d", n, tc.wantCounters)
			}
		})
	}
}
