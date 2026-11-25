package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b149_cards_test.go — S82 older sets, Throne of Eldraine batch 149
// (#5516): Tournament Grounds; Flaxen Intruder // Welcome Home; Rowan,
// Fearless Sparkmage.

const (
	eldB149GroundsOracle  = "a3ed5d7f-c554-4d46-86e5-7bb8df2693f6"
	eldB149IntruderOracle = "bacedc99-46d9-4757-8a27-8df77d7c2f02"
	eldB149RowanOracle    = "cd83f886-c651-4ad7-82ca-2446c60c8f94"
)

func TestEldB149CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eldB149GroundsOracle:         "Tournament Grounds",
		eldB149IntruderOracle:        "Flaxen Intruder",
		eldB149IntruderOracle + "#1": "Welcome Home",
		eldB149RowanOracle:           "Rowan, Fearless Sparkmage",
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

func TestEldB149TournamentGroundsRestrictsItsColoredMana(t *testing.T) {
	spec, _ := effects.Lookup(eldB149GroundsOracle)
	if len(spec.ManaAbilities) != 2 {
		t.Fatalf("mana abilities %+v, want {C} and a restricted {R}/{W}/{B}", spec.ManaAbilities)
	}
	if len(spec.ManaAbilities[0].Restrictions) != 0 {
		t.Errorf("the {C} ability is restricted: %v", spec.ManaAbilities[0].Restrictions)
	}
	got := spec.ManaAbilities[1]
	if got.Produced != "{R|W|B}" {
		t.Errorf("produces %q, want {R|W|B}", got.Produced)
	}
	want := []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Knight|Equipment")}
	if len(got.Restrictions) != 2 || got.Restrictions[0] != want[0] || got.Restrictions[1] != want[1] {
		t.Errorf("restrictions %v, want %v", got.Restrictions, want)
	}
}

func eldB149Intruder(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   eldB149IntruderOracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Flaxen Intruder", TypeLine: "Creature — Human Berserker", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 2},
			{Name: "Welcome Home", TypeLine: "Sorcery — Adventure", ManaCost: "{5}{G}{G}", Colors: []string{"G"}},
		},
	}
	c.SetFace(0)
	return c
}

func TestEldB149WelcomeHomeMakesThreeBearsThenTheIntruderIsCastable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := eldB149Intruder(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Welcome Home: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bears := 0
	for _, b := range g.Battlefield.Cards {
		if b.Controller == me.ID && effects.IsToken(b) && b.Name == "Bear" && b.Power == 2 && b.Toughness == 2 {
			bears++
		}
	}
	if bears != 3 {
		t.Fatalf("%d Bear tokens, want 3", bears)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Fatal("the resolved Adventure is not in exile (CR 715.3d)")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast Flaxen Intruder from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatal("Flaxen Intruder did not enter")
	}
}

func eldB149Attack(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	intruder := effectstest.B12Push(g, me.ID, "Flaxen Intruder", "Creature — Human Berserker", eldB149IntruderOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, intruder, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	return intruder
}

func TestEldB149FlaxenIntruderSacrificesToDestroyAnArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	intruder := eldB149Attack(t, g, me, opp)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(intruder) {
		t.Fatal("the Intruder was not sacrificed")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(ench) || !g.Battlefield.Contains(land) {
		t.Error("something other than the chosen target was destroyed")
	}
}

func TestEldB149FlaxenIntruderDeclinedKeepsItAndDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	intruder := eldB149Attack(t, g, me, opp)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(intruder) || !g.Battlefield.Contains(rock) {
		t.Error("declining the sacrifice still cost the Intruder or the artifact")
	}
}

func TestEldB149RowanPlusOnePumpsAndGivesFirstStrike(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Rowan, Fearless Sparkmage", eldB149RowanOracle, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike") {
		t.Error("the bear lacks first strike")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestEldB149RowanPlusOneWithNoTargetStillGainsLoyalty(t *testing.T) {
	g, seat, me, _ := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Rowan, Fearless Sparkmage", eldB149RowanOracle, 5)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1 with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, pw); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestEldB149RowanMinusTwoDamagesTwoCreaturesAndStopsThemBlocking(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Rowan, Fearless Sparkmage", eldB149RowanOracle, 5)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b, c)}); err == nil {
		t.Fatal("three targets were accepted for up to two")
	}
	if err := g.ActivateCatalogAbility(me.ID, pw, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
			t.Errorf("%s can still block", id)
		}
	}
	if effectstest.AuraRestrictions(t, g, c)&game.CantBlock != 0 {
		t.Error("an untargeted creature can't block")
	}
	if got := effectstest.LoyaltyOf(g, pw); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.DamageMarkedOn(g, id); got != 1 {
			t.Errorf("target %s has %d damage, want 1", id, got)
		}
	}
	if got := effectstest.DamageMarkedOn(g, c); got != 0 {
		t.Errorf("an untargeted creature has %d damage", got)
	}
}

func TestEldB149RowanMinusNineTakesEveryCreatureUntapsAndGivesHaste(t *testing.T) {
	g, seat, me, opp := effectstest.S51ActivePlayer(t)
	pw := effectstest.PushWalkerForTest(g, me.ID, "Rowan, Fearless Sparkmage", eldB149RowanOracle, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, pw, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("-9: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{theirs, mine} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("creature %s is not under my control", id)
		}
		if c.Tapped {
			t.Errorf("creature %s is still tapped", id)
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
			t.Errorf("creature %s lacks haste", id)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, land); c.Controller != opp.ID {
		t.Error("a land changed controllers")
	}
	if g.Battlefield.Contains(pw) {
		t.Error("Rowan at 0 loyalty should have left the battlefield")
	}
}
