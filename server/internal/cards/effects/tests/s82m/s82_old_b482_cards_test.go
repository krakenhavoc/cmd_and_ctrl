package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b482_cards_test.go — Rise of the Eldrazi, batch 482 (S82, #5581).
// Each test plays the real catalog entry through the engine.

const (
	oldB482DrakeOracle     = "3aca5718-9151-4a5b-b8a3-bfac7abf65ef"
	oldB482TeacherOracle   = "be118a2d-a27e-4174-86db-d9a2d304da2f"
	oldB482SpellbladeOracl = "e87813a3-893d-4416-9263-2ea2e3140490"
	oldB482GorgerOracle    = "ec5dc165-c760-4060-99fb-ec50f87374cb"
	oldB482VialOracle      = "efbc9b9c-2b8c-4121-b232-350c80f12a2a"
	oldB482HeroesOracle    = "fcfb96d7-9b3b-4f60-8714-eebe4d9c002b"
	oldB482EchoMageOracle  = "fb98f6b7-5986-4c5d-98fc-e5c4106f48bf" // a real leveler (Level up {1}{U})
)

func oldB482AddLevels(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLevel, n) })
}

// oldB482EmptyHand clears the opening hand the test table deals.
func oldB482EmptyHand(p *game.Player) {
	for p.Hand.Size() > 0 {
		_, _ = p.Hand.PopTop()
	}
}

func TestOldB482EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB482DrakeOracle: "Champion's Drake", oldB482TeacherOracle: "Venerated Teacher",
		oldB482SpellbladeOracl: "Surrakar Spellblade", oldB482GorgerOracle: "Thought Gorger",
		oldB482VialOracle: "Angelheart Vial", oldB482HeroesOracle: "Time of Heroes",
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

func TestOldB482ChampionsDrakeFliesAndGrowsWithALevelThreeCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.B12Push(g, me.ID, "Champion's Drake", "Creature — Drake", oldB482DrakeOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("no flying")
	}
	mage := effectstest.B12Push(g, me.ID, "Echo Mage", "Creature — Human Wizard", oldB482EchoMageOracle, 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Echo Mage", "Creature — Human Wizard", oldB482EchoMageOracle, 1, 1)
	oldB482AddLevels(g, theirs, 3)
	oldB482AddLevels(g, mage, 2)
	if p, tough := effectstest.PtOf(t, g, drake); p != 1 || tough != 1 {
		t.Errorf("Drake is %d/%d with only level 2 (and an opponent's level 3), want 1/1", p, tough)
	}
	oldB482AddLevels(g, mage, 1)
	if p, tough := effectstest.PtOf(t, g, drake); p != 4 || tough != 4 {
		t.Errorf("Drake is %d/%d with a level 3 creature, want 4/4", p, tough)
	}
}

func TestOldB482TimeOfHeroesPumpsOnlyYourLeveledCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Time of Heroes", "Enchantment", oldB482HeroesOracle, 0, 0)
	leveled := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	plain := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	oldB482AddLevels(g, leveled, 1)
	oldB482AddLevels(g, theirs, 1)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"leveled", leveled, 4}, {"plain", plain, 2}, {"opponent's leveled", theirs, 2}} {
		if p, tough := effectstest.PtOf(t, g, tc.id); p != tc.want || tough != tc.want {
			t.Errorf("%s bear is %d/%d, want %d/%d", tc.name, p, tough, tc.want, tc.want)
		}
	}
}

func TestOldB482VeneratedTeacherLevelsOnlyLevelUpCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Echo Mage", "Creature — Human Wizard", oldB482EchoMageOracle, 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Echo Mage", "Creature — Human Wizard", oldB482EchoMageOracle, 1, 1)
	effectstest.EnterFromHand(t, g, me.ID, "Venerated Teacher", "Creature — Human Wizard", oldB482TeacherOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, mage, game.CounterLevel); n != 2 {
		t.Errorf("your leveler has %d level counters, want 2", n)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterLevel); n != 0 {
		t.Errorf("a creature without level up got %d level counters", n)
	}
	if n := effectstest.CountersOn(g, theirs, game.CounterLevel); n != 0 {
		t.Errorf("an opponent's leveler got %d level counters", n)
	}
}

func TestOldB482SurrakarSpellbladeChargesOnInstantsAndDrawsForTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.B12Push(g, me.ID, "Surrakar Spellblade", "Creature — Surrakar", oldB482SpellbladeOracl, 2, 1)
	effectstest.AdvanceToMain(t, g)
	for _, yes := range []bool{true, true, false} {
		effectstest.CastFromHandForTest(t, g, me, "Bolt", "Instant", "{R}", "", game.CastSpellParams{})
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.CountersOn(g, blade, game.CounterCharge); n != 2 {
		t.Fatalf("charge counters %d, want 2 (two yes, one no)", n)
	}
	// A creature spell is not an instant or sorcery.
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, blade, game.CounterCharge); n != 2 {
		t.Errorf("a creature spell charged it: %d", n)
	}
	hand := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, blade, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (drew X=2)", got, hand+2)
	}
}

func TestOldB482ThoughtGorgerGrowsDiscardsTheHandAndDrawsOnLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB482EmptyHand(me)
	for i := 0; i < 3; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	gorger := effectstest.EnterFromHand(t, g, me.ID, "Thought Gorger", "Creature — Horror", oldB482GorgerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, gorger, game.CounterPlusOne); n != 3 {
		t.Errorf("+1/+1 counters %d, want 3 (three other cards in hand)", n)
	}
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0 after discarding it", me.Hand.Size())
	}
	effectstest.B27Kill(g, gorger)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Errorf("drew %d cards on leaving, want 3 (its counters as it last existed)", me.Hand.Size())
	}
}

func TestOldB482ThoughtGorgerWithAnEmptyHandDiscardsNothingAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB482EmptyHand(me)
	gorger := effectstest.EnterFromHand(t, g, me.ID, "Thought Gorger", "Creature — Horror", oldB482GorgerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, gorger, game.CounterPlusOne); n != 0 {
		t.Errorf("+1/+1 counters %d, want 0", n)
	}
	effectstest.B27Kill(g, gorger)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("drew %d cards with no counters", me.Hand.Size())
	}
}

func TestOldB482AngelheartVialStoresDamageAndSpendsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vial := effectstest.B12Push(g, me.ID, "Angelheart Vial", "Artifact", oldB482VialOracle, 0, 0)
	effectstest.GptB1DamagePlayer(g, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, vial, game.CounterCharge); n != 5 {
		t.Fatalf("charge counters %d, want 5 (damage dealt)", n)
	}
	effectstest.GptB1DamagePlayer(g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, vial, game.CounterCharge); n != 5 {
		t.Errorf("declined, yet charge counters are %d, want 5", n)
	}
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lib", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	life, hand := me.Life, me.Hand.Size()
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	effectstest.FloatForTest(g, me, "C")
	effectstest.ApaActivate(t, g, me, vial, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d (want %d), hand %d (want %d)", me.Life, life+2, me.Hand.Size(), hand+1)
	}
	if n := effectstest.CountersOn(g, vial, game.CounterCharge); n != 1 {
		t.Errorf("charge counters %d, want 1 after removing four", n)
	}
	// One counter is not enough.
	effectstest.Untap(g, vial)
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with one charge counter")
	}
}
