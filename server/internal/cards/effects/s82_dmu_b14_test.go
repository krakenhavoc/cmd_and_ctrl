package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	citizensArrestOracle     = "6f554418-9ee6-4a03-b81e-f93d5c70873e"
	temporaryLockdownOracle  = "b21c99b3-b975-4428-9c9a-326ded2fa8dc"
	historiansBoonOracle     = "ce1e6ae2-78c1-4bc6-bb73-bb025dd67a90"
	combatResearchOracle     = "fbfd7f38-be9a-49bf-8769-79e1e7e8dfdc"
	weatherseedTreatyOracle  = "7d723a6f-854f-43df-9401-a4eb0c0f639f"
	yotiaDeclaresWarOracle   = "4ce785b9-aee0-4ac6-ab20-ee20d0551be4"
	braidsFrightfulOracle    = "4567a42a-aa1c-4ca2-b5e2-b09acac3f276"
	theWorldSpellOracle      = "f5186d3f-52b2-43ef-be11-d183ab765663"
	foundingThirdPathOracle  = "0eff910a-712d-4652-8b3b-9eddf63eb891"
	dmuB14EnchantmentTypeLin = "Enchantment"
)

func dmuB14Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return ""
	}
	return z.Kind
}

// Citizen's Arrest exiles an opponent's creature or planeswalker until it
// leaves; its own controller's permanents and other types are illegal.
func TestCitizensArrestExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	artifact := pushPermanentForTest(g, opp.ID, "Their Rock", "", "Artifact")

	id := castCatalogSpell(t, g, "Citizen's Arrest", dmuB14EnchantmentTypeLin, citizensArrestOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{mine, artifact} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("accepted an illegal target %s", bad)
		}
	}
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if dmuB14Zone(g, theirs) != game.ZoneExile {
		t.Fatalf("target is in %q, want exile", dmuB14Zone(g, theirs))
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	g.RunStateChecksForTest()
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == opp.ID {
			back = true
		}
	}
	if !back {
		t.Fatal("the creature did not return when Citizen's Arrest left")
	}
}

// Temporary Lockdown exiles every nonland permanent with mana value 2 or
// less, tokens and its controller's own included, and they come back when
// it leaves.
func TestTemporaryLockdownExilesSmallPermanentsUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	push := func(owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner})
		return id
	}
	small := push(opp.ID, "Small Bear", "Creature — Bear", "{1}{G}")
	mine := push(me.ID, "My Mox", "Artifact", "{0}")
	big := push(opp.ID, "Big Bear", "Creature — Bear", "{2}{G}")
	land := push(opp.ID, "Test Land", "Land", "")
	token := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: token, Name: "Soldier", TypeLine: "Token Creature — Soldier", Owner: opp.ID, Controller: opp.ID})

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Temporary Lockdown", TypeLine: dmuB14EnchantmentTypeLin, OracleID: temporaryLockdownOracle,
		ManaCost: "{1}{W}{W}", Owner: me.ID, Controller: me.ID})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		name string
	}{{small, "Small Bear"}, {mine, "My Mox"}, {token, "Soldier"}} {
		if g.Battlefield.Contains(c.id) {
			t.Errorf("%s was not exiled", c.name)
		}
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(land) || !g.Battlefield.Contains(id) {
		t.Fatal("a mana value 3 permanent, a land or the enchantment itself was exiled")
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	g.RunStateChecksForTest()
	if countBattlefieldByName(g, "Small Bear") != 1 || countBattlefieldByName(g, "My Mox") != 1 {
		t.Fatal("the exiled permanents did not return when Temporary Lockdown left")
	}
	if countBattlefieldByName(g, "Soldier") != 0 {
		t.Fatal("an exiled token came back")
	}
}

// Historian's Boon: a Soldier for itself and for each other nontoken
// enchantment you control that enters, none for a token or an opponent's.
func TestHistoriansBoonMakesSoldiersForEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	castCatalogSpell(t, g, "Historian's Boon", dmuB14EnchantmentTypeLin, historiansBoonOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Soldier"); n != 1 {
		t.Fatalf("Soldiers after Boon entered = %d, want 1", n)
	}
	castCatalogSpell(t, g, "Test Enchantment", dmuB14EnchantmentTypeLin, "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Soldier"); n != 2 {
		t.Fatalf("Soldiers after another enchantment = %d, want 2", n)
	}
	// A creature spell is no enchantment.
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Soldier"); n != 2 {
		t.Fatalf("Soldiers after a creature = %d, want 2", n)
	}
	_ = me
}

// Historian's Boon: the final chapter of a Saga you control triggers an
// Angel; an earlier chapter does not.
func TestHistoriansBoonAngelOnFinalChapter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushPermanentForTest(g, me.ID, "Historian's Boon", historiansBoonOracle, dmuB14EnchantmentTypeLin)
	bear := pushVanillaCreature(g, me.ID, "Target Bear", 1, 1)
	saga := castReadAheadSaga(t, g, "The Weatherseed Treaty", weatherseedTreatyOracle)
	chooseChapter(t, g, 2)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Angel"); n != 0 {
		t.Fatalf("an Angel came from chapter II (%d)", n)
	}
	for i := 0; i < 400; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			break
		}
	}
	if got := chaptersFired(g, saga); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("chapters fired = %v, want [2 3] (step %v seat %d lore %d)", got, g.Turn.Step, g.Turn.ActiveSeat, loreCountersOn(g, saga))
	}
	for i := 0; i < 4; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		if latestPickTarget(g, me.ID) != nil {
			pickTriggerTarget(t, g, me.ID, bear)
		}
		passPriorityAroundTable(t, g)
	}
	if n := countBattlefieldByName(g, "Angel"); n != 1 {
		t.Fatalf("Angels after the final chapter = %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && (c.Power != 4 || c.Toughness != 4 || !c.IsToken()) {
			t.Fatalf("Angel is %d/%d token=%v", c.Power, c.Toughness, c.IsToken())
		}
	}
}

// An opponent's Saga reaching its final chapter gives nothing.
func TestHistoriansBoonIgnoresAnOpponentsSaga(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%4]
	pushPermanentForTest(g, opp.ID, "Historian's Boon", historiansBoonOracle, dmuB14EnchantmentTypeLin)
	castReadAheadSaga(t, g, "The Elder Dragon War", elderDragonWarOracle)
	chooseChapter(t, g, 3)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Angel"); n != 0 {
		t.Fatalf("an opponent's Boon made %d Angels for my Saga", n)
	}
}

// Combat Research: the Aura draws on combat damage to a player; a
// legendary creature also gets +1/+1 and ward {1}.
func TestCombatResearchDrawsAndBuffsLegends(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	plain := pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	legend := pushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Test", 2, 2)

	castCatalogSpell(t, g, "Combat Research", "Enchantment — Aura", combatResearchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: plain}})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, plain); p != 2 {
		t.Fatalf("a non-legendary creature got a bonus: power %d", p)
	}
	handBefore := me.Hand.Size()
	attackWith(t, g, opp.ID, plain)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("cards drawn from combat damage = %d, want 1", got)
	}

	castCatalogSpell(t, g, "Combat Research", "Enchantment — Aura", combatResearchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, legend); p != 3 {
		t.Fatalf("legendary creature power %d, want 3", p)
	}
}
