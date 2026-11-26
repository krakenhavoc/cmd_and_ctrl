package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b481_cards_test.go — Rise of the Eldrazi, batch 481 (S82, #5581).
// Each test plays the real catalog entry through the engine.

const (
	oldB481EnatuOracle    = "1cb3f27c-2670-43b8-9305-8adf2e7bd3ae"
	oldB481NearDeathOracl = "2afb027f-3c02-46cf-97a1-af0e946ee477"
	oldB481JaddiOracle    = "9c8027b7-b823-4214-936c-91b7a8986228"
	oldB481DestinyOracle  = "ba7be3e9-d240-4fed-943d-a405409b13b5"
	oldB481OmenOracle     = "dc1711cf-c62b-436b-ba35-56bef8ef14d0"
	oldB481CurseOracle    = "e0d78210-ba85-439f-bf1f-92ef80b6030d"
)

func TestOldB481EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB481EnatuOracle: "Enatu Golem", oldB481NearDeathOracl: "Near-Death Experience",
		oldB481JaddiOracle: "Jaddi Lifestrider", oldB481DestinyOracle: "Living Destiny",
		oldB481OmenOracle: "Baneful Omen", oldB481CurseOracle: "Curse of Wizardry",
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

func TestOldB481EnatuGolemGainsFourWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.B12Push(g, me.ID, "Enatu Golem", "Artifact Creature — Golem", oldB481EnatuOracle, 3, 5)
	life := me.Life
	effectstest.B27Kill(g, golem)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestOldB481NearDeathExperienceWinsOnlyAtExactlyOneLife(t *testing.T) {
	for _, tc := range []struct {
		life int
		win  bool
	}{{1, true}, {2, false}, {40, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Near-Death Experience", "Enchantment", oldB481NearDeathOracl, 0, 0)
		me.Life = tc.life
		effectstest.EndTurn(t, g)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		won := g.Outcome != nil && g.Outcome.Winner == me.ID
		if won != tc.win {
			t.Errorf("life %d: won %v, want %v", tc.life, won, tc.win)
		}
	}
}

func TestOldB481NearDeathExperienceChecksAgainOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Near-Death Experience", "Enchantment", oldB481NearDeathOracl, 0, 0)
	me.Life = 1
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	me.Life = 5 // gained in response, before the trigger resolves
	effectstest.PassPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Error("won although the life was no longer exactly 1 as the trigger resolved")
	}
}

func TestOldB481JaddiLifestriderGainsTwoForEachCreatureTapped(t *testing.T) {
	for _, tc := range []struct {
		pick int
		want int
	}{{2, 4}, {0, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		a := effectstest.B12Push(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
		b := effectstest.B12Push(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
		life := me.Life
		effectstest.EnterFromHand(t, g, me.ID, "Jaddi Lifestrider", "Creature — Elemental", oldB481JaddiOracle)
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if ask == nil {
			t.Fatal("no choose_cards prompt for the tap")
		}
		picks := []uuid.UUID{a, b}[:tc.pick]
		if err := g.ResolveChooseCards(ask.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+tc.want {
			t.Errorf("tapped %d: life %d, want %d", tc.pick, me.Life, life+tc.want)
		}
		if tc.pick == 2 && (!effectstest.TappedForTest(t, g, a) || !effectstest.TappedForTest(t, g, b)) {
			t.Error("the chosen creatures were not tapped")
		}
	}
}

func TestOldB481LivingDestinyGainsTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	reveal := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: reveal, Name: "Revealed", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CCCG")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Living Destiny", TypeLine: "Instant", ManaCost: "{3}{G}", OracleID: oldB481DestinyOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast without naming a creature card to reveal")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, RevealIDs: []uuid.UUID{reveal}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (mana value 3)", me.Life, life+3)
	}
	if !me.Hand.Contains(reveal) {
		t.Error("the revealed card left the hand")
	}
}

func TestOldB481LivingDestinyNeedsACreatureCardInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCG")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Living Destiny", TypeLine: "Instant", ManaCost: "{3}{G}", OracleID: oldB481DestinyOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast with no creature card in hand")
	}
}

func TestOldB481BanefulOmenDrainsForTheRevealedManaValue(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Baneful Omen", "Enchantment", oldB481OmenOracle, 0, 0)
		top := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Creature — Bear", ManaCost: "{3}{B}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		effectstest.AdvanceToEndStep(t, g)
		oppLife, life := opp.Life, me.Life
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := oppLife
		if yes {
			want -= 4
		}
		if opp.Life != want || me.Life != life {
			t.Errorf("yes=%v: opponent %d (want %d), me %d (want %d)", yes, opp.Life, want, me.Life, life)
		}
		if !me.Library.Contains(top) {
			t.Error("the revealed card left the library")
		}
	}
}

func TestOldB481CurseOfWizardryDrainsTheCasterOfTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Curse of Wizardry", "Enchantment", oldB481CurseOracle, "R")
	effectstest.AdvanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	effectstest.CastFromHandForTest(t, g, me, "Red Spell", "Sorcery", "{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("the red caster is at %d, want %d", me.Life, life-1)
	}
	if opp.Life != oppLife {
		t.Errorf("the opponent lost life (%d, want %d)", opp.Life, oppLife)
	}
	life = me.Life
	effectstest.CastFromHandForTest(t, g, me, "Green Spell", "Sorcery", "{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a green spell cost life: %d, want %d", me.Life, life)
	}
}
