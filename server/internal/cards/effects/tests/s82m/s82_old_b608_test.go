package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b608_test.go — Eighth Edition, batch 608 (#5607): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB608WoodenOracle  = "0bd8917c-bec4-4603-bc3f-8e0c2afae56a"
	oldB608SanctOracle   = "2fc7740a-dfa4-4279-8822-3f7c93e6b685"
	oldB608ReversalOracl = "70e0b676-61a2-4dcf-8f61-d9281467ed43"
	oldB608SalveOracle   = "8da8644c-75a1-4fe9-8e94-900d948d631c"
	oldB608CupOracle     = "8e2017c3-057d-4e30-bd60-f486fddbc6ca"
	oldB608StreamOracle  = "9eb2912d-2130-49f2-9529-b58fa5a97a15"
	oldB608OathOracle    = "ad66a99b-c7e7-4b2c-8525-299d3609d1df"
	oldB608ColossusOracl = "caab555c-7852-418a-bdc3-f908af7e0932"
)

func TestOldB608CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB608WoodenOracle: "Wooden Sphere", oldB608SanctOracle: "Sanctimony", oldB608ReversalOracl: "Blessed Reversal",
		oldB608SalveOracle: "Healing Salve", oldB608CupOracle: "Ivory Cup", oldB608StreamOracle: "Stream of Life",
		oldB608OathOracle: "Peach Garden Oath", oldB608ColossusOracl: "Phyrexian Colossus",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Wooden Sphere and Ivory Cup ---

func TestOldB608SphereAndCupGainLifeOnlyWhenTheColoredCastIsPaidFor(t *testing.T) {
	for _, c := range []struct {
		name, oracle, color, other string
	}{
		{"Wooden Sphere", oldB608WoodenOracle, "G", "W"},
		{"Ivory Cup", oldB608CupOracle, "W", "G"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, c.name, "Artifact", c.oracle, false)
		effectstest.AdvanceToMain(t, g)
		before := me.Life
		cast := func(p *game.Player, color string) {
			t.Helper()
			id := effectstest.HandCardFull(p, "Spell", "Instant", "{1}", "", []string{color})
			if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("%s: cast: %v", c.name, err)
			}
		}
		// A spell of another color: no question.
		cast(me, c.other)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, me.ID) != nil {
			t.Fatalf("%s: a spell of the wrong color asked for a payment", c.name)
		}
		// My own spell of its color, paid for.
		cast(me, c.color)
		effectstest.B06AddMana(me, "R")
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d after paying, want %d", c.name, me.Life, before+1)
		}
		// An opponent's spell of its color, declined.
		g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
		effectstest.AdvanceToMain(t, g)
		cast(opp, c.color)
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d after declining, want %d", c.name, me.Life, before+1)
		}
	}
}

// --- Sanctimony ---

func TestOldB608SanctimonyGainsLifeWhenAnOpponentTapsAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sanctimony", "Enchantment", oldB608SanctOracle, false)
	effectstest.AdvanceToMain(t, g)
	mountain := effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	mine := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	before := me.Life
	for _, land := range []struct {
		id   uuid.UUID
		who  *game.Player
		gain bool
	}{{forest, opp, false}, {mine, me, false}, {mountain, opp, true}} {
		effectstest.TapForMana(t, g, land.who.ID, land.id)
		effectstest.PassPriorityAroundTable(t, g)
		if land.gain {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
		} else if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
			t.Fatalf("a land tapped by %s that is no opposing Mountain triggered Sanctimony", land.who.ID)
		}
	}
	if me.Life != before+1 {
		t.Errorf("life %d, want %d after one opposing Mountain", me.Life, before+1)
	}
}

// --- Blessed Reversal ---

func TestOldB608BlessedReversalCountsTheCreaturesAttackingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attackers := []uuid.UUID{
		effectstest.B12Creature(g, opp.ID, "Attacker A", "Creature — Bear", 2, 2),
		effectstest.B12Creature(g, opp.ID, "Attacker B", "Creature — Bear", 2, 2),
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	effectstest.Pr7bAttack(t, g, me.ID, attackers...)
	before := me.Life
	id := effectstest.HandCardFull(me, "Blessed Reversal", "Instant", "{1}{W}", oldB608ReversalOracl, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life %d, want %d for two attackers", me.Life, before+6)
	}
}

// --- Healing Salve ---

func TestOldB608HealingSalveGainsThreeLifeOrShieldsThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB608SalveOracle, []int{0}, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before+3 {
		t.Errorf("life %d, want %d from the first mode", opp.Life, before+3)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	effectstest.CastCatalogSpellWithModesTargets(t, g, oldB608SalveOracle, []int{1}, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.DealDamageTo(t, g, src, bear, 4)
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("%d damage marked, want 1 after 3 of 4 were prevented", got)
	}
}

// --- Stream of Life ---

func TestOldB608StreamOfLifeGainsXLifeForTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	id := effectstest.HandCardFull(me, "Stream of Life", "Sorcery", "{X}{G}", oldB608StreamOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 4, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before+4 {
		t.Errorf("life %d, want %d", opp.Life, before+4)
	}
}

// --- Peach Garden Oath ---

func TestOldB608PeachGardenOathGainsTwoPerCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	}
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	before := me.Life
	id := effectstest.HandCardFull(me, "Peach Garden Oath", "Sorcery", "{W}", oldB608OathOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+6 {
		t.Errorf("life %d, want %d for three creatures", me.Life, before+6)
	}
}

// --- Phyrexian Colossus ---

func TestOldB608PhyrexianColossusUntapsForEightLifeAndStaysTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	colossus := effectstest.B12Push(g, me.ID, "Phyrexian Colossus", "Artifact Creature — Phyrexian Golem", oldB608ColossusOracl, 8, 8)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(colossus) })
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.IsTapped(g, colossus) {
		t.Fatal("the Colossus untapped during its controller's untap step")
	}
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pay 8 life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-8 {
		t.Errorf("life %d, want %d", me.Life, before-8)
	}
	if effectstest.IsTapped(g, colossus) {
		t.Error("the Colossus is still tapped")
	}
	g.WithWriteLock(func() { me.Life = 5 })
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(colossus) })
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("paid 8 life with only 5")
	}
}

func TestOldB608PhyrexianColossusNeedsThreeBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.B12Push(g, me.ID, "Phyrexian Colossus", "Artifact Creature — Phyrexian Golem", oldB608ColossusOracl, 8, 8)
	var blockers []uuid.UUID
	for i := 0; i < 3; i++ {
		blockers = append(blockers, effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 1))
	}
	effectstest.BrAttack(t, g, colossus)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blockers[0], Attacker: colossus}, {Blocker: blockers[1], Attacker: colossus},
	}); err == nil {
		t.Error("two blockers were accepted for a creature that needs three")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blockers[0], Attacker: colossus}, {Blocker: blockers[1], Attacker: colossus}, {Blocker: blockers[2], Attacker: colossus},
	}); err != nil {
		t.Errorf("three blockers are legal: %v", err)
	}
}
