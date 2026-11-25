package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// surge_keyword_test.go — surge (CR 702.117a, #6255): "You may pay
// [cost] rather than pay this spell's mana cost as you cast this spell if
// you or one of your teammates has cast another spell this turn."
// Through the real announce: a strict cast with the surge cost floated.

const (
	testSurgeSpellOracle    = "test-6255-surge-spell"
	testSurgeCreatureOracle = "test-6255-surge-creature"
	testSurgeFillerOracle   = "test-6255-filler"
)

// registerSurgeCards registers Crush of Tentacles' and Reckless
// Bushwhacker's shapes on test oracles, plus a free instant to cast
// first.
func registerSurgeCards(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:         testSurgeSpellOracle,
		Name:             "Test Surge Spell",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{U}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !effects.SurgeCostWasPaid(ctx) {
				return nil
			}
			return effects.GainLife{Amount: 4}.Apply(ctx)
		},
	})
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:         testSurgeCreatureOracle,
		Name:             "Test Surge Creature",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.CastWithSurge), "Test Surge Creature — you gain 3 life",
				effects.Do(effects.GainLife{Amount: 3})),
		},
	})
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     testSurgeFillerOracle,
		Name:         "Test Filler",
		Completeness: effects.CompletenessFull,
		OnResolve:    func(*game.StackItem, *effects.Context) error { return nil },
	})
}

func castSurgeTest(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, pool, alt string) error {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, pool)
	return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{AlternativeCost: alt, Strict: true})
}

func castFiller(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	if err := castSurgeTest(t, g, me, "Test Filler", "Instant", "", testSurgeFillerOracle, "", ""); err != nil {
		t.Fatalf("cast the first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// The first spell of the turn has no other spell before it: surge is
// refused. After one spell, it is offered and the cheaper cost pays.
func TestSurgeNeedsAnotherSpellCastThisTurn(t *testing.T) {
	registerSurgeCards(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := castSurgeTest(t, g, me, "Test Surge Spell", "Sorcery", "{3}{U}{U}", testSurgeSpellOracle, "U", effects.SurgeAltCostKey); err == nil {
		t.Fatal("surge was accepted before any other spell was cast this turn")
	}
	castFiller(t, g, me)
	life := me.Life
	if err := castSurgeTest(t, g, me, "Test Surge Spell", "Sorcery", "{3}{U}{U}", testSurgeSpellOracle, "U", effects.SurgeAltCostKey); err != nil {
		t.Fatalf("surge after another spell this turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("life = %d, want %d: \"if this spell's surge cost was paid\" did not hold", me.Life, life+4)
	}
}

// Another PLAYER's spell is not "you or one of your teammates" (CR
// 102.3; a C2 table has no teams).
func TestSurgeIgnoresAnOpponentsSpell(t *testing.T) {
	registerSurgeCards(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		}
		g.SpellsCastThisTurn[opp.ID] = game.CastTally{Total: 1}
	})
	if err := castSurgeTest(t, g, me, "Test Surge Spell", "Sorcery", "{3}{U}{U}", testSurgeSpellOracle, "U", effects.SurgeAltCostKey); err == nil {
		t.Fatal("surge was accepted on an opponent's spell")
	}
}

// Cast for its mana cost, the spell's "if surge was paid" does nothing.
func TestSurgeSpellCastNormallyDoesNotCountAsSurged(t *testing.T) {
	registerSurgeCards(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castFiller(t, g, me)
	life := me.Life
	if err := castSurgeTest(t, g, me, "Test Surge Spell", "Sorcery", "{3}{U}{U}", testSurgeSpellOracle, "CCCUU", ""); err != nil {
		t.Fatalf("cast for the mana cost: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d -> %d after a cast that was not surged", life, me.Life)
	}
}

// Reckless Bushwhacker's shape: "When this creature enters, if its surge
// cost was paid, …" reads the permanent's cast record.
func TestSurgedCreatureEntersWithItsSurgeTrigger(t *testing.T) {
	for _, surged := range []bool{true, false} {
		t.Run(map[bool]string{true: "surged", false: "mana cost"}[surged], func(t *testing.T) {
			registerSurgeCards(t)
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castFiller(t, g, me)
			life := me.Life
			alt, pool := "", "CCR"
			if surged {
				alt, pool = effects.SurgeAltCostKey, "R"
			}
			if err := castSurgeTest(t, g, me, "Test Surge Creature", "Creature — Goblin", "{2}{R}", testSurgeCreatureOracle, pool, alt); err != nil {
				t.Fatalf("surged=%v: cast: %v", surged, err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := life
			if surged {
				want += 3
			}
			if me.Life != want {
				t.Fatalf("surged=%v: life = %d, want %d", surged, me.Life, want)
			}
		})
	}
}
