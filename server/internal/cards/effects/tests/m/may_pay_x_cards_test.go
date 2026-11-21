package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMayPayXCardsAreFull(t *testing.T) {
	for _, oid := range []string{effectstest.TilonallisSummonerOracle, effectstest.FlameblastDragonOracle} {
		s, ok := effects.Lookup(oid)
		if !ok {
			t.Fatalf("%s is not in the catalog", oid)
		}
		if s.Completeness != effects.CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v; want Full with none", s.Name, s.Completeness, s.Caveats)
		}
	}
	if s, _ := effects.Lookup(effectstest.TilonallisSummonerOracle); !effectstest.HasKeyword(s.PrintedKeywords, game.KeywordAscend) {
		t.Error("Tilonalli's Summoner does not declare ascend")
	}
}

// X is chosen from 0 to what the pool can pay, {X}{R} is paid with X
// settled, and X tapped Elementals enter attacking, split among the
// opponents as their controller chose. With no city's blessing they are
// exiled at the next end step.
func TestTilonallisSummonerPaysXAndSplitsTheAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SummonerAttacks(t, g, 3, 1)
	pa := effectstest.AnswerNumber(t, g, me.ID, 3)
	if pa.ResourceOrEnergy() != game.PayResourceNone || pa.Min != 0 || pa.Max != 3 || pa.Goal != 3 {
		t.Fatalf("X prompt = %+v, want 0..3 with goal 3", pa)
	}
	var pay *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			pay = c
		}
	}
	if pay == nil || pay.PayCost != "{3}{R}" {
		t.Fatalf("payment prompt = %+v, want {3}{R}", pay)
	}
	if pay.OwedInStep.Step != game.StepDeclareAttackers {
		t.Errorf("payment owed in %+v, want the declare attackers step", pay.OwedInStep)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left in the pool, want 0", n)
	}
	// Three opponents: asked about the attacked one first, then the next.
	first := effectstest.AnswerNumber(t, g, me.ID, 2)
	if first.Max != 3 || first.Goal != 3 {
		t.Errorf("first split prompt = %+v, want 0..3 with goal 3", first)
	}
	second := effectstest.AnswerNumber(t, g, me.ID, 0)
	if second.Max != 1 || second.Goal != 0 {
		t.Errorf("second split prompt = %+v, want 0..1 with goal 0", second)
	}
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Fatal("a third split prompt: the last opponent should take the rest")
	}
	// The attacked opponent comes first, then the others in seat order.
	var others []uuid.UUID
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != opp.ID {
			others = append(others, s.ID)
		}
	}
	third := others[len(others)-1]
	toks := effectstest.Elementals(g, me.ID)
	if len(toks) != 3 {
		t.Fatalf("%d Elementals, want 3", len(toks))
	}
	at := map[uuid.UUID]int{}
	for _, c := range toks {
		if !c.Tapped {
			t.Errorf("Elemental %s is untapped", c.InstanceID)
		}
		if c.Power != 1 || c.Toughness != 1 {
			t.Errorf("Elemental is %d/%d, want 1/1", c.Power, c.Toughness)
		}
		at[c.AttackingTarget]++
	}
	if at[opp.ID] != 2 || at[third] != 1 {
		t.Errorf("attacking %v, want 2 at the attacked opponent and 1 at the last", at)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.Elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals left after the end step without the city's blessing", n)
	}
}

// With the city's blessing as the end step trigger resolves, the tokens
// stay.
func TestTilonallisSummonerKeepsTheTokensWithTheCitysBlessing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SummonerAttacks(t, g, 2, 1)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	if n := len(effectstest.Elementals(g, me.ID)); n != 2 {
		t.Fatalf("%d Elementals, want 2", n)
	}
	effectstest.GrantBlessing(g, me)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.Elementals(g, me.ID)); n != 2 {
		t.Errorf("%d Elementals after the end step with the city's blessing, want 2", n)
	}
}

// "Don't pay" makes nothing and spends nothing; a player with no {R} is
// not asked at all.
func TestTilonallisSummonerDeclinedOrUnpayable(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SummonerAttacks(t, g, 2, 1)
		effectstest.AnswerNumber(t, g, me.ID, 2)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		if n := len(effectstest.Elementals(g, me.ID)); n != 0 {
			t.Errorf("%d Elementals after declining", n)
		}
		if n := me.ManaPool.Total(); n != 3 {
			t.Errorf("pool %d after declining, want 3", n)
		}
	})
	t.Run("no red", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.SummonerAttacks(t, g, 4, 0)
		if effectstest.PayAmountPrompt(g, me.ID) != nil || effectstest.HasPayUnlessFor(g, me.ID) {
			t.Error("asked to pay {X}{R} with no red mana")
		}
	})
}

// Flameblast Dragon pays {X}{R} for X damage to the target chosen as the
// trigger went on the stack; a bot's goal is the target's lethal damage.
func TestFlameblastDragonDealsXDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := effectstest.PushDiesCreatureForTest(g, me.ID, "Flameblast Dragon", effectstest.FlameblastDragonOracle, "Creature — Dragon", 5, 5)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.FillPool(me, 5)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 3)
	if pa.Max != 5 || pa.Goal != 3 || pa.Unit != game.PayAmountDamage {
		t.Fatalf("X prompt = %+v, want 0..5 with goal 3 in damage", pa)
	}
	if c := effectstest.AnswerPayUnless(t, g, me.ID, true); c.PayCost != "{3}{R}" {
		t.Errorf("payment = %s, want {3}{R}", c.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the 3/3 survived 3 damage")
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool %d, want 2 left", n)
	}
}

// The ceiling counts what the auto-tapper can produce, not only the
// pool: five untapped lands with one Mountain pay up to {4}{R}, and
// paying taps them.
func TestTilonallisSummonerCeilingCountsUntappedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lands := []uuid.UUID{effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")}
	for i := 0; i < 4; i++ {
		lands = append(lands, effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island"))
	}
	effectstest.SummonerAttacks(t, g, 0, 0)
	pa := effectstest.AnswerNumber(t, g, me.ID, 4)
	if pa.Max != 4 {
		t.Fatalf("X ceiling %d, want 4", pa.Max)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	for _, id := range lands {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("land %s untapped after paying {4}{R}", id)
		}
	}
	effectstest.AnswerNumber(t, g, me.ID, 4)
	if n := len(effectstest.Elementals(g, me.ID)); n != 4 {
		t.Errorf("%d Elementals, want 4", n)
	}
}
