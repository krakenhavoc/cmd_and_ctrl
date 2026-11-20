package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 3, 3)
	artifact := effectstest.PushPermanentForTest(g, opp.ID, "Their Rock", "", "Artifact")

	id := effectstest.CastCatalogSpell(t, g, "Citizen's Arrest", dmuB14EnchantmentTypeLin, citizensArrestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{mine, artifact} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("accepted an illegal target %s", bad)
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	if effectstest.CountBattlefieldByName(g, "Small Bear") != 1 || effectstest.CountBattlefieldByName(g, "My Mox") != 1 {
		t.Fatal("the exiled permanents did not return when Temporary Lockdown left")
	}
	if effectstest.CountBattlefieldByName(g, "Soldier") != 0 {
		t.Fatal("an exiled token came back")
	}
}

// Historian's Boon: a Soldier for itself and for each other nontoken
// enchantment you control that enters, none for a token or an opponent's.
func TestHistoriansBoonMakesSoldiersForEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.CastCatalogSpell(t, g, "Historian's Boon", dmuB14EnchantmentTypeLin, historiansBoonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Soldier"); n != 1 {
		t.Fatalf("Soldiers after Boon entered = %d, want 1", n)
	}
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", dmuB14EnchantmentTypeLin, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Soldier"); n != 2 {
		t.Fatalf("Soldiers after another enchantment = %d, want 2", n)
	}
	// A creature spell is no enchantment.
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Soldier"); n != 2 {
		t.Fatalf("Soldiers after a creature = %d, want 2", n)
	}
	_ = me
}

// Historian's Boon: the final chapter of a Saga you control triggers an
// Angel; an earlier chapter does not.
func TestHistoriansBoonAngelOnFinalChapter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushPermanentForTest(g, me.ID, "Historian's Boon", historiansBoonOracle, dmuB14EnchantmentTypeLin)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Target Bear", 1, 1)
	saga := effectstest.CastReadAheadSaga(t, g, "The Weatherseed Treaty", weatherseedTreatyOracle)
	effectstest.ChooseChapter(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Angel"); n != 0 {
		t.Fatalf("an Angel came from chapter II (%d)", n)
	}
	for i := 0; i < 400; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepPrecombatMain && g.Turn.ActiveSeat == seat {
			break
		}
	}
	if got := effectstest.ChaptersFired(g, saga); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("chapters fired = %v, want [2 3] (step %v seat %d lore %d)", got, g.Turn.Step, g.Turn.ActiveSeat, effectstest.LoreCountersOn(g, saga))
	}
	for i := 0; i < 4; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickTriggerTarget(t, g, me.ID, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountBattlefieldByName(g, "Angel"); n != 1 {
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
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%4]
	effectstest.PushPermanentForTest(g, opp.ID, "Historian's Boon", historiansBoonOracle, dmuB14EnchantmentTypeLin)
	effectstest.CastReadAheadSaga(t, g, "The Elder Dragon War", effectstest.ElderDragonWarOracle)
	effectstest.ChooseChapter(t, g, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Angel"); n != 0 {
		t.Fatalf("an opponent's Boon made %d Angels for my Saga", n)
	}
}

// Combat Research: the Aura draws on combat damage to a player; a
// legendary creature also gets +1/+1 and ward {1}.
func TestCombatResearchDrawsAndBuffsLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	legend := effectstest.PushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Test", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Combat Research", "Enchantment — Aura", combatResearchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: plain}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, plain); p != 2 {
		t.Fatalf("a non-legendary creature got a bonus: power %d", p)
	}
	handBefore := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("cards drawn from combat damage = %d, want 1", got)
	}

	effectstest.CastCatalogSpell(t, g, "Combat Research", "Enchantment — Aura", combatResearchOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: legend}})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, legend); p != 3 {
		t.Fatalf("legendary creature power %d, want 3", p)
	}
}
