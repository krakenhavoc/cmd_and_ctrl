package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b741_test.go — Exodus counters (S82 OLD batch 741, #5626).

const (
	oldB741HatcherOracle   = "726fe7a4-49b3-4b67-9d8c-6b334d0fd915"
	oldB741RogueOracle     = "90ec956f-b6de-4214-9b5a-895280e542c1"
	oldB741WorkhorseOracle = "bb4ab425-3da7-455d-b217-3d585d476113"
	oldB741CannibalOracle  = "c175ab0e-7085-41ec-83e4-38fc85c338fa"
	oldB741SquadronOracle  = "f1482ca4-610b-48fe-ac23-4526637bb72a"
	oldB741MaggotsOracle   = "fcaa5a54-eb73-4d78-96a3-826d1902f642"
)

func TestOldB741CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB741HatcherOracle:   "Spike Hatcher",
		oldB741RogueOracle:     "Spike Rogue",
		oldB741WorkhorseOracle: "Workhorse",
		oldB741CannibalOracle:  "Spike Cannibal",
		oldB741SquadronOracle:  "Thopter Squadron",
		oldB741MaggotsOracle:   "Mind Maggots",
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

func TestOldB741TheSpikesEnterWithTheirCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		counters               int
	}{
		{"Spike Hatcher", "Creature — Spike", oldB741HatcherOracle, 6},
		{"Spike Rogue", "Creature — Spike", oldB741RogueOracle, 2},
		{"Workhorse", "Artifact Creature — Horse", oldB741WorkhorseOracle, 4},
		{"Spike Cannibal", "Creature — Spike", oldB741CannibalOracle, 1},
		{"Thopter Squadron", "Artifact Creature — Thopter", oldB741SquadronOracle, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.counters {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.counters)
			}
			if p, tough := effectstest.PtOf(t, g, id); p != tc.counters || tough != tc.counters {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.counters, tc.counters)
			}
		})
	}
}

func TestOldB741SpikeHatcherRegeneratesForACounterAndMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Hatcher", "Creature — Spike", oldB741HatcherOracle, 6)
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("regenerated with an empty mana pool")
	}
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 5 {
		t.Errorf("the Hatcher has %d counters after paying one, want 5", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spike); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(spike) {
		t.Fatal("the shielded Hatcher was destroyed")
	}
	if !effectstest.AlliTapped(t, g, spike) {
		t.Error("regeneration did not tap it")
	}
}

func TestOldB741SpikeHatcherMovesACounterToTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, spike := effectstest.SthB2Spike(t, g, "Spike Hatcher", "Creature — Spike", oldB741HatcherOracle, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, spike, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, spike, game.CounterPlusOne); got != 5 {
		t.Errorf("the Hatcher has %d counters, want 5", got)
	}
}

func TestOldB741SpikeRogueDrawsACounterFromAnyOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, rogue := effectstest.SthB2Spike(t, g, "Spike Rogue", "Creature — Spike", oldB741RogueOracle, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rogue, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{theirs}, CounterCounts: []int{1}}); err == nil {
		t.Fatal("removed a counter from an opponent's creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, rogue, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters, want 0", got)
	}
	if got := effectstest.CountersOn(g, rogue, game.CounterPlusOne); got != 3 {
		t.Errorf("the Rogue has %d counters, want 3", got)
	}
	// Its own counter is a legal payment too.
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, rogue, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{rogue}, CounterCounts: []int{1}}); err != nil {
		t.Fatalf("activate with its own counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rogue, game.CounterPlusOne); got != 3 {
		t.Errorf("the Rogue has %d counters after moving one onto itself, want 3", got)
	}
}

func TestOldB741WorkhorseMakesManaFromItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, horse := effectstest.SthB2Spike(t, g, "Workhorse", "Artifact Creature — Horse", oldB741WorkhorseOracle, 1)
	if err := g.ActivateManaAbility(me.ID, horse, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.CountersOn(g, horse, game.CounterPlusOne); got != 0 {
		t.Errorf("the Workhorse has %d counters, want 0", got)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "C" {
		t.Errorf("pool %v, want one {C}", me.ManaPool)
	}
}

func TestOldB741WorkhorseWithNoCounterMakesNoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, horse := effectstest.SthB2Spike(t, g, "Workhorse", "Artifact Creature — Horse", oldB741WorkhorseOracle, 0)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, horse).Power = 1
		effectstest.FindBattlefieldCardForTest(g, horse).Toughness = 1
	})
	if err := g.ActivateManaAbility(me.ID, horse, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("made mana with no counter to remove")
	}
}

func TestOldB741SpikeCannibalTakesEveryCounterFromEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	clean := effectstest.B12Creature(g, opp.ID, "Clean", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 3)
	})
	cannibal := effectstest.CastCatalogSpell(t, g, "Spike Cannibal", "Creature — Spike", oldB741CannibalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, cannibal, game.CounterPlusOne); got != 6 {
		t.Errorf("the Cannibal has %d counters, want 6 (its own plus five)", got)
	}
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "clean": clean} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 0 {
			t.Errorf("%s still has %d +1/+1 counters", name, got)
		}
	}
}

func TestOldB741ThopterSquadronMakesAThopterOnlyAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, squad := effectstest.SthB2Spike(t, g, "Thopter Squadron", "Artifact Creature — Thopter", oldB741SquadronOracle, 3)
	effectstest.DcMana(t, g, me.ID, "{C}")
	before := len(g.Battlefield.Cards)
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, squad, game.CounterPlusOne); got != 2 {
		t.Errorf("the Squadron has %d counters, want 2", got)
	}
	if got := len(g.Battlefield.Cards); got != before+1 {
		t.Fatalf("battlefield has %d cards, want %d", got, before+1)
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := g.Battlefield.Cards[i]; c.Name == "Thopter" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil || token.Power != 1 || token.Toughness != 1 || !effectstest.HasKeyword(token.Keywords, "flying") {
		t.Fatalf("no 1/1 flying Thopter token: %+v", token)
	}
	// Not at sorcery speed: the stack is non-empty while another ability waits.
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second activate: %v", err)
	}
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a sorcery-speed ability with an ability on the stack")
	}
}

func TestOldB741ThopterSquadronSacrificesAnotherThopterForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, squad := effectstest.SthB2Spike(t, g, "Thopter Squadron", "Artifact Creature — Thopter", oldB741SquadronOracle, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	thopter := effectstest.B12Creature(g, me.ID, "Thopter", "Artifact Creature — Thopter", 1, 1)
	effectstest.DcMana(t, g, me.ID, "{C}")
	for _, bad := range []uuid.UUID{squad, bear} {
		if err := g.ActivateCatalogAbility(me.ID, squad, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bad}}); err == nil {
			t.Errorf("sacrificed %v, which is not another Thopter", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, squad, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{thopter}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thopter) {
		t.Error("the Thopter was not sacrificed")
	}
	if got := effectstest.CountersOn(g, squad, game.CounterPlusOne); got != 4 {
		t.Errorf("the Squadron has %d counters, want 4", got)
	}
}

func oldB741MaggotsHand(g *game.Game, me *game.Player) (c1, c2, spell uuid.UUID) {
	c1 = effectstest.PutInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	c2 = effectstest.PutInHand(me, game.Card{Name: "Elk", TypeLine: "Creature — Elk", Owner: me.ID, Controller: me.ID})
	spell = effectstest.PutInHand(me, game.Card{Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	return
}

func TestOldB741MindMaggotsGrowsTwoCountersForEachCreatureDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c1, c2, spell := oldB741MaggotsHand(g, me)
	maggots := effectstest.CastCatalogSpell(t, g, "Mind Maggots", "Creature — Insect", oldB741MaggotsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no discard prompt")
	}
	for _, id := range prompt.ChooseCards {
		if id == spell {
			t.Error("a noncreature card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, c1, c2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, maggots, game.CounterPlusOne); got != 4 {
		t.Errorf("the Maggots have %d counters, want 4", got)
	}
	if !effectstest.HandHasID(me, spell) || effectstest.HandHasID(me, c1) || effectstest.HandHasID(me, c2) {
		t.Error("the wrong cards were discarded")
	}
}

func TestOldB741MindMaggotsMayDiscardNothingAndAsksNothingWithoutCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c1, _, _ := oldB741MaggotsHand(g, me)
	maggots := effectstest.CastCatalogSpell(t, g, "Mind Maggots", "Creature — Insect", oldB741MaggotsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, maggots, game.CounterPlusOne); got != 0 {
		t.Errorf("the Maggots have %d counters, want 0", got)
	}
	if !effectstest.HandHasID(me, c1) {
		t.Error("a creature was discarded though none was chosen")
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PutInHand(me2, game.Card{Name: "Bolt", TypeLine: "Instant", Owner: me2.ID, Controller: me2.ID})
	effectstest.CastCatalogSpell(t, g2, "Mind Maggots", "Creature — Insect", oldB741MaggotsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, me2.ID) != nil {
		t.Error("asked to discard with no creature card in hand")
	}
}
