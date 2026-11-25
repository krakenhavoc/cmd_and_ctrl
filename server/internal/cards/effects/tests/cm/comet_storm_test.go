package testscm

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const (
	cometStormOracle          = "9c2b262b-53c0-41a1-8f1a-0c336453d4bf"
	strengthOfTheTajuruOracle = "132401fb-6754-4190-9c85-1e57e7f63ed6"
)

// kickCountHand puts the card in the active player's hand in a main phase.
func kickCountHand(t *testing.T, g *game.Game, name, oracle, cost string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return me, id
}

// kickCountBig is a 2/5 creature, which survives 2 damage.
func kickCountBig(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test",
		Power: 2, Toughness: 5, Owner: owner, Controller: owner})
	return id
}

func kickCountRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func kickCountToughnessLeft(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.DamageMarked
		}
	}
	return -1
}

// Comet Storm (#6150, CR 601.2c, 702.33c-d): X=2 kicked twice takes
// exactly three different targets and deals 2 to each; two targets or
// four are refused, as is the same target twice.
func TestCometStormKickedTwiceTakesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := kickCountHand(t, g, "Comet Storm", cometStormOracle, "")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := kickCountBig(g, opp.ID, "A")
	b := kickCountBig(g, opp.ID, "B")
	c := kickCountBig(g, opp.ID, "C")
	d := kickCountBig(g, opp.ID, "D")
	kicked := []int{0, 0}
	for _, bad := range [][]game.TargetRef{
		kickCountRefs(a, b),
		kickCountRefs(a, b, c, d),
		kickCountRefs(a, a, b),
	} {
		err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: kicked, Targets: bad})
		if err == nil {
			t.Fatalf("kicked twice with %d targets: %v, want a refusal", len(bad), err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: kicked,
		Targets: kickCountRefs(a, b, c)}); err != nil {
		t.Fatalf("kicked twice with three targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{a, b, c} {
		if got := kickCountToughnessLeft(g, x); got != 2 {
			t.Errorf("target %s has %d damage, want 2", x, got)
		}
	}
	if got := kickCountToughnessLeft(g, d); got != 0 {
		t.Errorf("the untargeted creature has %d damage", got)
	}
}

// Unkicked, Comet Storm takes one target, and two are refused.
func TestCometStormUnkickedTakesOneTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := kickCountHand(t, g, "Comet Storm", cometStormOracle, "")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("unkicked with two targets was accepted")
	}
	before := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("unkicked with one target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Fatalf("the opponent lost %d life, want 3", got)
	}
}

// The move enumerator offers Comet Storm only with one target plus one
// for each kick it announces, and offers a kicked cast.
func TestCometStormEnumeratedTargetsMatchTheKicks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := kickCountHand(t, g, "Comet Storm", cometStormOracle, "{X}{R}{R}")
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "A")
	effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "B")
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}{R}")
	kickedMoves := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != id {
			continue
		}
		var p struct {
			OptionalCosts []int             `json:"optional_costs"`
			Targets       []json.RawMessage `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if len(p.Targets) != 1+len(p.OptionalCosts) {
			t.Fatalf("a move kicks %d times and names %d targets", len(p.OptionalCosts), len(p.Targets))
		}
		if len(p.OptionalCosts) > 0 {
			kickedMoves++
		}
	}
	if kickedMoves == 0 {
		t.Fatal("the enumerator offers no kicked Comet Storm")
	}
}

// Strength of the Tajuru: X=2 kicked once takes exactly two creatures and
// puts two +1/+1 counters on each.
func TestStrengthOfTheTajuruKickedOncePumpsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := kickCountHand(t, g, "Strength of the Tajuru", strengthOfTheTajuruOracle, "")
	a := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "A")
	b := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: []int{0},
		Targets: kickCountRefs(a)}); err == nil {
		t.Fatal("kicked once with one target was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: []int{0},
		Targets: kickCountRefs(a, b)}); err != nil {
		t.Fatalf("kicked once with two targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{a, b} {
		c, ok := effectstest.BattlefieldCard(g, x)
		if !ok || c.Counters[game.CounterPlusOne] != 2 {
			t.Errorf("creature %s has %v, want two +1/+1 counters", x, c.Counters)
		}
	}
}
