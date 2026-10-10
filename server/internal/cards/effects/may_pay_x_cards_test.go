package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_pay_x_cards_test.go — #2727: "you may pay {X}{R}" as a trigger
// resolves (ADR 0129's amendment of 2026-10-09).

const (
	tilonallisSummonerOracle = "830d3450-18e7-4bac-b330-3906ef5e4514"
	flameblastDragonOracle   = "3180ee4f-66d7-4b73-bc10-9ee0c790ec18"
)

func TestMayPayXCardsAreFull(t *testing.T) {
	for _, oid := range []string{tilonallisSummonerOracle, flameblastDragonOracle} {
		s, ok := Lookup(oid)
		if !ok {
			t.Fatalf("%s is not in the catalog", oid)
		}
		if s.Completeness != CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v; want Full with none", s.Name, s.Completeness, s.Caveats)
		}
	}
	if s, _ := Lookup(tilonallisSummonerOracle); !hasKeyword(s.PrintedKeywords, game.KeywordAscend) {
		t.Error("Tilonalli's Summoner does not declare ascend")
	}
}

// summonerAttacks puts Tilonalli's Summoner on the active seat's
// battlefield, fills its pool with `generic` colourless and `red` red
// mana, attacks the next seat with it and resolves the trigger.
func summonerAttacks(t *testing.T, g *game.Game, generic, red int) (me, opp *game.Player) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	summoner := pushDiesCreatureForTest(g, me.ID, "Tilonalli's Summoner", tilonallisSummonerOracle, "Creature — Human Shaman", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(summoner, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	fillPool(me, generic)
	fillPoolColored(me, "R", red)
	passPriorityAroundTable(t, g)
	return me, opp
}

// answerNumber answers the open number prompt for `p` with n.
func answerNumber(t *testing.T, g *game.Game, p uuid.UUID, n int) *game.PayAmountPrompt {
	t.Helper()
	c := payAmountPrompt(g, p)
	if c == nil {
		t.Fatalf("no number prompt for %s", p)
	}
	pa := *c.PayAmount
	if err := g.ResolvePayAmount(c.ID, p, n); err != nil {
		t.Fatalf("answer %d: %v", n, err)
	}
	return &pa
}

func elementals(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

// X is chosen from 0 to what the pool can pay, {X}{R} is paid with X
// settled, and X tapped Elementals enter attacking, split among the
// opponents as their controller chose. With no city's blessing they are
// exiled at the next end step.
func TestTilonallisSummonerPaysXAndSplitsTheAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := summonerAttacks(t, g, 3, 1)
	pa := answerNumber(t, g, me.ID, 3)
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
	answerPayUnless(t, g, me.ID, true)
	if n := len(me.ManaPool); n != 0 {
		t.Errorf("%d mana left in the pool, want 0", n)
	}
	// Three opponents: asked about the attacked one first, then the next.
	first := answerNumber(t, g, me.ID, 2)
	if first.Max != 3 || first.Goal != 3 {
		t.Errorf("first split prompt = %+v, want 0..3 with goal 3", first)
	}
	second := answerNumber(t, g, me.ID, 0)
	if second.Max != 1 || second.Goal != 0 {
		t.Errorf("second split prompt = %+v, want 0..1 with goal 0", second)
	}
	if payAmountPrompt(g, me.ID) != nil {
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
	toks := elementals(g, me.ID)
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
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(elementals(g, me.ID)); n != 0 {
		t.Errorf("%d Elementals left after the end step without the city's blessing", n)
	}
}

// With the city's blessing as the end step trigger resolves, the tokens
// stay.
func TestTilonallisSummonerKeepsTheTokensWithTheCitysBlessing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := summonerAttacks(t, g, 2, 1)
	answerNumber(t, g, me.ID, 2)
	answerPayUnless(t, g, me.ID, true)
	answerNumber(t, g, me.ID, 2)
	if n := len(elementals(g, me.ID)); n != 2 {
		t.Fatalf("%d Elementals, want 2", n)
	}
	grantBlessing(g, me)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(elementals(g, me.ID)); n != 2 {
		t.Errorf("%d Elementals after the end step with the city's blessing, want 2", n)
	}
}

// "Don't pay" makes nothing and spends nothing; a player with no {R} is
// not asked at all.
func TestTilonallisSummonerDeclinedOrUnpayable(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := summonerAttacks(t, g, 2, 1)
		answerNumber(t, g, me.ID, 2)
		answerPayUnless(t, g, me.ID, false)
		if n := len(elementals(g, me.ID)); n != 0 {
			t.Errorf("%d Elementals after declining", n)
		}
		if n := len(me.ManaPool); n != 3 {
			t.Errorf("pool %d after declining, want 3", n)
		}
	})
	t.Run("no red", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := summonerAttacks(t, g, 4, 0)
		if payAmountPrompt(g, me.ID) != nil || hasPayUnlessFor(g, me.ID) {
			t.Error("asked to pay {X}{R} with no red mana")
		}
	})
}

// Flameblast Dragon pays {X}{R} for X damage to the target chosen as the
// trigger went on the stack; a bot's goal is the target's lethal damage.
func TestFlameblastDragonDealsXDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dragon := pushDiesCreatureForTest(g, me.ID, "Flameblast Dragon", flameblastDragonOracle, "Creature — Dragon", 5, 5)
	victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	answerTriggerTargets(t, g, me.ID, victim)
	fillPool(me, 5)
	fillPoolColored(me, "R", 1)
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 3)
	if pa.Max != 5 || pa.Goal != 3 || pa.Unit != game.PayAmountDamage {
		t.Fatalf("X prompt = %+v, want 0..5 with goal 3 in damage", pa)
	}
	if c := answerPayUnless(t, g, me.ID, true); c.PayCost != "{3}{R}" {
		t.Errorf("payment = %s, want {3}{R}", c.PayCost)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the 3/3 survived 3 damage")
	}
	if n := len(me.ManaPool); n != 2 {
		t.Errorf("pool %d, want 2 left", n)
	}
}

// The ceiling counts what the auto-tapper can produce, not only the
// pool: five untapped lands with one Mountain pay up to {4}{R}, and
// paying taps them.
func TestTilonallisSummonerCeilingCountsUntappedLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lands := []uuid.UUID{pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")}
	for i := 0; i < 4; i++ {
		lands = append(lands, pushLandFor(g, me.ID, "Island", "Basic Land — Island"))
	}
	summonerAttacks(t, g, 0, 0)
	pa := answerNumber(t, g, me.ID, 4)
	if pa.Max != 4 {
		t.Fatalf("X ceiling %d, want 4", pa.Max)
	}
	answerPayUnless(t, g, me.ID, true)
	for _, id := range lands {
		if !isTapped(g, id) {
			t.Errorf("land %s untapped after paying {4}{R}", id)
		}
	}
	answerNumber(t, g, me.ID, 4)
	if n := len(elementals(g, me.ID)); n != 4 {
		t.Errorf("%d Elementals, want 4", n)
	}
}
