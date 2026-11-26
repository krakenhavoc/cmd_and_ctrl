package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b348_cards_test.go — Khans of Tarkir, batch 348 (S82 older sets,
// #5554): the life gain and loss cards. Every test plays the real catalog
// entry through the engine.

const (
	oldB348Sage       = "02c741d0-99d3-48e5-9846-0fea5a0a29fe"
	oldB348Frenzy     = "0842cc1d-de1e-4fb7-a54b-78a4c8f40d3e"
	oldB348Butcher    = "1ff93e62-6ae0-4438-a88e-ae6fd7d91e41"
	oldB348Dreadmaw   = "8c8c9f16-a3d0-49f2-a95b-dfef756de0b3"
	oldB348Parapet    = "9df1c48b-52a8-47d7-8315-272fecf1a8a0"
	oldB348Flayer     = "df621822-c811-40fa-99c1-ea9920bcd2e2"
	oldB348RushBattle = "e203a201-2567-4684-ae9f-fd4e91c6dc37"
)

func TestOldB348EveryCardIsRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Sage of the Inward Eye": oldB348Sage, "Death Frenzy": oldB348Frenzy, "Butcher of the Horde": oldB348Butcher,
		"Kheru Dreadmaw": oldB348Dreadmaw, "Archers' Parapet": oldB348Parapet, "Sultai Flayer": oldB348Flayer,
		"Rush of Battle": oldB348RushBattle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not in the catalog", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %d, want Full with none", name, spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB348Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	return g, me, opp
}

func TestOldB348SageGivesLifelinkOnNoncreatureSpellsOnly(t *testing.T) {
	g, me, opp := oldB348Table(t)
	effectstest.B12Push(g, me.ID, "Sage of the Inward Eye", "Creature — Djinn Wizard", oldB348Sage, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Bear Spell", "Creature — Bear", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Fatal("a creature spell gave lifelink")
	}
	effectstest.CastFromHandForTest(t, g, me, "Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("a noncreature spell did not give lifelink to my creature")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature gained lifelink")
	}
}

func TestOldB348DeathFrenzyShrinksAllAndGainsLifePerDeathThisTurn(t *testing.T) {
	g, me, opp := oldB348Table(t)
	small1 := effectstest.B12Creature(g, me.ID, "Small One", "Creature — Bear", 2, 2)
	small2 := effectstest.B12Creature(g, opp.ID, "Small Two", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 5, 5)
	life := me.Life
	effectstest.CastFromHandForTest(t, g, me, "Death Frenzy", "Sorcery", "", oldB348Frenzy, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small1) || g.Battlefield.Contains(small2) {
		t.Fatal("the 2/2s survived -2/-2")
	}
	if p, tough := effectstest.MkmB1PT(t, g, big); p != 3 || tough != 3 {
		t.Errorf("the 5/5 is %d/%d, want 3/3", p, tough)
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 (two creatures died)", life, me.Life)
	}
	effectstest.KillCreature(t, g, opp.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d after a third death this turn", me.Life, life+3)
	}
}

func TestOldB348ButcherOfTheHordeSacrificesAnotherCreatureForAChosenKeyword(t *testing.T) {
	for idx, kw := range []string{"vigilance", "lifelink", "haste"} {
		g, me, _ := oldB348Table(t)
		id := effectstest.B12Push(g, me.ID, "Butcher of the Horde", "Creature — Demon", oldB348Butcher, 5, 4)
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		if idx == 0 {
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
				t.Fatal("sacrificed itself for \"another creature\"")
			}
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, idx)
		if g.Battlefield.Contains(fodder) {
			t.Error("the sacrificed creature is still on the battlefield")
		}
		for _, other := range []string{"vigilance", "lifelink", "haste"} {
			if got := effectstest.HasEffectiveKeyword(t, g, id, other); got != (other == kw) {
				t.Errorf("choice %s: has %s = %v", kw, other, got)
			}
		}
	}
}

func TestOldB348KheruDreadmawGainsLifeEqualToTheSacrificedToughness(t *testing.T) {
	g, me, _ := oldB348Table(t)
	id := effectstest.B12Push(g, me.ID, "Kheru Dreadmaw", "Creature — Zombie Crocodile", oldB348Dreadmaw, 4, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Wall", 0, 3)
	effectstest.AddCountersForTest(g, fodder, game.CounterPlusOne, 2)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("G", "C")...) })
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}}); err == nil {
		t.Fatal("sacrificed itself for \"another creature\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (a 0/3 with two counters)", life, me.Life)
	}
}

func TestOldB348ArchersParapetDrainsEachOpponentForOne(t *testing.T) {
	g, me, _ := oldB348Table(t)
	id := effectstest.B12Push(g, me.ID, "Archers' Parapet", "Creature — Wall", oldB348Parapet, 0, 5)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("B", "C")...) })
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
	if !effectstest.TappedOnBattlefield(t, g, id) {
		t.Error("the Parapet did not tap")
	}
}

func TestOldB348SultaiFlayerGainsFourForABigCreatureYouControl(t *testing.T) {
	g, me, opp := oldB348Table(t)
	effectstest.B12Push(g, me.ID, "Sultai Flayer", "Creature — Snake Shaman", oldB348Flayer, 3, 4)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 4, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life changed by %d for a toughness-3 or an opponent's creature", me.Life-life)
	}
	effectstest.KillCreature(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4", life, me.Life)
	}
}

func TestOldB348RushOfBattlePumpsMineAndGivesWarriorsLifelink(t *testing.T) {
	g, me, opp := oldB348Table(t)
	warrior := effectstest.B12Creature(g, me.ID, "My Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Rush of Battle", "Sorcery", "", oldB348RushBattle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{warrior, bear} {
		if p, tough := effectstest.MkmB1PT(t, g, id); p != 4 || tough != 3 {
			t.Errorf("my creature is %d/%d, want 4/3", p, tough)
		}
	}
	if p, tough := effectstest.MkmB1PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, warrior, "lifelink") {
		t.Error("my Warrior has no lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") || effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("a non-Warrior or an opponent's Warrior gained lifelink")
	}
}
