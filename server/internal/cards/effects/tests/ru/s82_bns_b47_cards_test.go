package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b47_cards_test.go — S82, bonus sheets and child sets batch 47,
// Foundations Jumpstart instants (#5929).

const (
	bnsB47FieryFall   = "f014b1a1-20ef-475c-bf83-f1a82e68ab97"
	bnsB47Prophecy    = "b5e51171-2e36-4b09-b43d-e5d437a19305"
	bnsB47Groundswell = "ddbacb74-1f98-4607-a92e-d14973b9d0ef"
	bnsB47Timing      = "102bbedf-9f84-433a-bd01-e1a4c05fbedb"
	bnsB47Infuriate   = "d23119ca-73d1-4a48-a8d0-dab8919d891f"
	bnsB47Appetite    = "53570a4e-a750-4776-bc52-0025cd510a35"
	bnsB47Rampage     = "d1a665f1-b06f-4775-bcb1-c57a476fa651"
	bnsB47Hope        = "5b6f655b-1518-4ab1-8765-3f549819ef3b"
	bnsB47Masses      = "f73416a8-40e5-4791-948f-147ef2221ee7"
	bnsB47OpenFire    = "8b218023-50b8-4d9d-b17b-9647479d3675"
	bnsB47Outflank    = "c3436d70-230c-4bda-b7da-935a270f135f"
)

func TestBNSB47CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB47FieryFall: "Fiery Fall", bnsB47Prophecy: "Fire Prophecy", bnsB47Groundswell: "Groundswell",
		bnsB47Timing: "Impeccable Timing", bnsB47Infuriate: "Infuriate", bnsB47Appetite: "Insatiable Appetite",
		bnsB47Rampage: "Invigorated Rampage", bnsB47Hope: "Light of Hope", bnsB47Masses: "Might of the Masses",
		bnsB47OpenFire: "Open Fire", bnsB47Outflank: "Outflank",
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

// bnsB47Cast casts an instant from the active seat's hand in the current
// step (combat included), strictly.
func bnsB47Cast(t *testing.T, g *game.Game, name, oracle string, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

func bnsB47Pool(t *testing.T, g *game.Game, mana string) {
	t.Helper()
	effectstest.AddPool(t, g, g.Seats[g.Turn.ActiveSeat].ID, mana)
}

// bnsB47Combat declares attacker (mine) at the opponent and locks it in.
func bnsB47Combat(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
}

func bnsB47PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func TestBNSB47FieryFallDealsFiveAndHasBasicLandcycling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B16Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	small := effectstest.B16Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	_ = small
	effectstest.CastCatalogSpell(t, g, "Fiery Fall", "Instant", bnsB47FieryFall, effectstest.CardRefs(big))
	_ = me
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(big) {
		t.Error("the 5/5 took 5 and lived")
	}
	spec, _ := effects.Lookup(bnsB47FieryFall)
	if len(spec.Activated) != 1 {
		t.Errorf("%d activated abilities, want basic landcycling", len(spec.Activated))
	}
}

func TestBNSB47FieryFallRefusesAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	bnsB47Pool(t, g, "{R}{R}{R}{R}{R}{R}")
	if err := bnsB47Cast(t, g, "Fiery Fall", bnsB47FieryFall, nil, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}); err == nil {
		t.Fatal("a player was a legal target")
	}
}

func TestBNSB47FireProphecyDamagesThenMayRummage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.SeedLibrary(me, "Drawn")
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Fire Prophecy", "Instant", bnsB47Prophecy, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, keep)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear took 3 and lived")
	}
	drew := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Drawn" {
			drew = true
		}
		if c.InstanceID == keep {
			t.Error("the tucked card is still in hand")
		}
	}
	if !drew {
		t.Error("no card drawn after the tuck")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != keep {
		t.Error("the tucked card is not on the bottom")
	}
}

func TestBNSB47FireProphecyDeclinedRummageDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.SeedLibrary(me, "Drawn")
	me.Hand.Cards = me.Hand.Cards[:0]
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Fire Prophecy", "Instant", bnsB47Prophecy, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand)
	}
}

func TestBNSB47GroundswellPumpsTwoThenFourWithLandfall(t *testing.T) {
	for _, landfall := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if landfall {
			g.WithWriteLock(func() { g.TurnTally.Players[me.ID] = game.PlayerTurnTally{LandsEntered: 1} })
		}
		effectstest.CastCatalogSpell(t, g, "Groundswell", "Instant", bnsB47Groundswell, effectstest.CardRefs(bear))
		effectstest.PassPriorityAroundTable(t, g)
		want := 4
		if landfall {
			want = 6
		}
		if p, tough := bnsB47PT(t, g, bear); p != want || tough != want {
			t.Errorf("landfall=%v: %d/%d, want %d/%d (instead, not on top of +2/+2)", landfall, p, tough, want, want)
		}
	}
}

func TestBNSB47ImpeccableTimingHitsOnlyAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 3)
	idle := effectstest.B16Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 3)
	bnsB47Combat(t, g, mine)
	bnsB47Pool(t, g, "{W}{W}{W}{W}")
	if err := bnsB47Cast(t, g, "Impeccable Timing", bnsB47Timing, nil, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("a creature not in combat was a legal target")
	}
	if err := bnsB47Cast(t, g, "Impeccable Timing", bnsB47Timing, nil, effectstest.CardRefs(mine)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mine) {
		t.Error("the attacker took 3 and lived")
	}
}

func TestBNSB47InfuriateGivesThreeTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B16Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Infuriate", "Instant", bnsB47Infuriate, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := bnsB47PT(t, g, bear); p != 5 || tough != 4 {
		t.Errorf("%d/%d, want 5/4", p, tough)
	}
}

func bnsB47RunAppetite(t *testing.T, food bool, sacrifice bool) (int, int, bool) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	var f uuid.UUID
	if food {
		f = effectstest.PushToken(g, me.ID, effects.FoodToken())
	}
	effectstest.CastCatalogSpell(t, g, "Insatiable Appetite", "Instant", bnsB47Appetite, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if food {
		if sacrifice {
			effectstest.AnswerOwnPermanents(t, g, me.ID, f)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	p, tough := bnsB47PT(t, g, bear)
	stillThere := food && g.Battlefield.Contains(f)
	return p - 2, tough - 2, stillThere
}

func TestBNSB47InsatiableAppetiteSacrificeGivesFivePlusFive(t *testing.T) {
	if p, tough, kept := bnsB47RunAppetite(t, true, true); p != 5 || tough != 5 || kept {
		t.Errorf("sacrificed: +%d/+%d, Food kept %v, want +5/+5 and the Food gone", p, tough, kept)
	}
}

func TestBNSB47InsatiableAppetiteDeclinedOrNoFoodGivesThree(t *testing.T) {
	if p, tough, kept := bnsB47RunAppetite(t, true, false); p != 3 || tough != 3 || !kept {
		t.Errorf("declined: +%d/+%d, Food kept %v, want +3/+3 and the Food kept", p, tough, kept)
	}
	if p, tough, _ := bnsB47RunAppetite(t, false, false); p != 3 || tough != 3 {
		t.Errorf("no Food: +%d/+%d, want +3/+3", p, tough)
	}
}

func TestBNSB47InvigoratedRampageBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B16Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Invigorated Rampage", "Instant", bnsB47Rampage, []int{0}, effectstest.CardRefs(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := bnsB47PT(t, g, a); p != 6 || !effectstest.HasKeywordOnBattlefield(t, g, a, "trample") {
		t.Errorf("mode one: power %d / trample %v, want 6 and trample", p, effectstest.HasKeywordOnBattlefield(t, g, a, "trample"))
	}
	g2 := effectstest.NewCatalogGame(t)
	c := effectstest.B16Creature(g2, g2.Seats[0].ID, "C", "Creature — Bear", 2, 2)
	d := effectstest.B16Creature(g2, g2.Seats[1].ID, "D", "Creature — Bear", 2, 2)
	_ = b
	effectstest.CastModal(t, g2, "Invigorated Rampage", "Instant", bnsB47Rampage, []int{1}, effectstest.CardRefs(c, d))
	effectstest.PassPriorityAroundTable(t, g2)
	for _, id := range []uuid.UUID{c, d} {
		if p, _ := bnsB47PT(t, g2, id); p != 4 || !effectstest.HasKeywordOnBattlefield(t, g2, id, "trample") {
			t.Errorf("mode two: power %d, want 4 and trample", p)
		}
	}
}

func TestBNSB47LightOfHopeThreeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	effectstest.CastModal(t, g, "Light of Hope", "Instant", bnsB47Hope, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("life %d, want %d", me.Life, before+4)
	}

	g2 := effectstest.NewCatalogGame(t)
	ench := effectstest.B12Permanent(g2, g2.Seats[1].ID, "Test Aura", "Enchantment")
	effectstest.CastModal(t, g2, "Light of Hope", "Instant", bnsB47Hope, []int{1}, effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}

	g3 := effectstest.NewCatalogGame(t)
	bear := effectstest.B16Creature(g3, g3.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g3, "Light of Hope", "Instant", bnsB47Hope, []int{2}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g3)
	if n := effectstest.PlusOneCounters(g3, bear); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

func TestBNSB47MightOfTheMassesCountsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Creature(g, me.ID, "Two", "Creature — Bear", 1, 1)
	effectstest.B16Creature(g, me.ID, "Three", "Creature — Bear", 1, 1)
	effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Might of the Masses", "Instant", bnsB47Masses, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := bnsB47PT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("%d/%d, want 5/5 (+3 for three creatures you control)", p, tough)
	}
}

func TestBNSB47OpenFireHitsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Open Fire", "Instant", bnsB47OpenFire, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life %d, want %d", opp.Life, before-3)
	}
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Open Fire", "Instant", bnsB47OpenFire, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear took 3 and lived")
	}
}

func TestBNSB47OutflankDamageEqualsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 3)
	effectstest.B16Creature(g, me.ID, "Two", "Creature — Bear", 1, 1)
	effectstest.B16Creature(g, me.ID, "Three", "Creature — Bear", 1, 1)
	idle := effectstest.B16Creature(g, opp.ID, "Idle", "Creature — Bear", 1, 1)
	bnsB47Combat(t, g, atk)
	bnsB47Pool(t, g, "{W}")
	if err := bnsB47Cast(t, g, "Outflank", bnsB47Outflank, nil, effectstest.CardRefs(idle)); err == nil {
		t.Fatal("a creature not in combat was a legal target")
	}
	if err := bnsB47Cast(t, g, "Outflank", bnsB47Outflank, nil, effectstest.CardRefs(atk)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(atk) {
		t.Error("the 1/3 attacker took 3 (three creatures) and lived")
	}
}
