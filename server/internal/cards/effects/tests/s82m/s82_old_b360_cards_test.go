package testss82m

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b360_cards_test.go — S82 OLD batch 360 (#5557, Conspiracy): Muzzio's
// Preparations, Secret Summoning, Unexpected Potential and Power Play, on the
// conspiracy seam (#6102, ADR 0227). Each test plays the real catalog entry
// through the engine.

const (
	oldB360Muzzio     = "0e505699-8388-4862-93e0-059cef91cdd4"
	oldB360Summoning  = "1342ee23-c3a0-4578-833a-e30a9fc9141c"
	oldB360Unexpected = "1acbbf61-edf7-4f67-968b-6832c2dd54d1"
	oldB360PowerPlay  = "15710c57-369a-442b-b7ea-175f3d9d164d"
)

func TestOldB360EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB360Muzzio: "Muzzio's Preparations", oldB360Summoning: "Secret Summoning",
		oldB360Unexpected: "Unexpected Potential", oldB360PowerPlay: "Power Play",
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

func oldB360Card(name, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Conspiracy", OracleID: oracle, IsSideboard: true}
}

// oldB360Table starts a four-seat game in which seat i has side[i] in its
// sideboard. The deal may be waiting on a hidden agenda.
func oldB360Table(t *testing.T, side map[int][]game.Card) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		deck = append(deck, side[i]...)
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func oldB360Name(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == chooser {
			if err := g.ResolveCardNameChoice(c.ID, chooser, name); err != nil {
				t.Fatalf("ResolveCardNameChoice: %v", err)
			}
			return
		}
	}
	t.Fatal("no hidden agenda prompt is open")
}

func oldB360Begin(t *testing.T, g *game.Game) {
	t.Helper()
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
}

func oldB360Reveal(t *testing.T, g *game.Game, p *game.Player, name string) {
	t.Helper()
	for _, c := range g.ConspiraciesOf(p.ID) {
		if c.Name == name {
			if err := g.PerformSpecialAction(p.ID, c.InstanceID, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err != nil {
				t.Fatalf("reveal_agenda: %v", err)
			}
			return
		}
	}
	t.Fatalf("%s is not in the command zone", name)
}

func oldB360CastBears(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	id := effectstest.HandSpell(p, name, "Creature — Bear", "{1}{G}")
	effectstest.FillPoolColored(p, "G", 2)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Muzzio's Preparations -----------------------------------------------------

func TestOldB360MuzziosPreparationsAddsACounterOnlyFaceUpAndOnlyForTheChosenName(t *testing.T) {
	g := oldB360Table(t, map[int][]game.Card{0: {oldB360Card("Muzzio's Preparations", oldB360Muzzio)}})
	me := g.Seats[0]
	oldB360Name(t, g, me.ID, "Grizzly Bears")
	oldB360Begin(t, g)
	early := oldB360CastBears(t, g, me, "Grizzly Bears")
	if c, _ := effectstest.BattlefieldCard(g, early); c.Counters[game.CounterPlusOne] != 0 {
		t.Fatal("a face-down Muzzio's Preparations added a counter (CR 315.5b)")
	}
	oldB360Reveal(t, g, me, "Muzzio's Preparations")
	other := oldB360CastBears(t, g, me, "Runeclaw Bear")
	if c, _ := effectstest.BattlefieldCard(g, other); c.Counters[game.CounterPlusOne] != 0 {
		t.Error("a creature with another name got a counter")
	}
	named := oldB360CastBears(t, g, me, "Grizzly Bears")
	if c, _ := effectstest.BattlefieldCard(g, named); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("a creature with the chosen name has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

// --- Secret Summoning ----------------------------------------------------------

func TestOldB360SecretSummoningFetchesEveryCopyWhenTheNamedCreatureEnters(t *testing.T) {
	g := oldB360Table(t, map[int][]game.Card{0: {oldB360Card("Secret Summoning", oldB360Summoning)}})
	me := g.Seats[0]
	oldB360Name(t, g, me.ID, "Grizzly Bears")
	oldB360Begin(t, g)
	var copies [2]uuid.UUID
	g.WithWriteLock(func() {
		for i := range copies {
			c := game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
			copies[i] = c.InstanceID
			me.Library.PushTop(c)
		}
	})
	oldB360CastBears(t, g, me, "Grizzly Bears")
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a face-down Secret Summoning triggered")
	}
	oldB360Reveal(t, g, me, "Secret Summoning")
	oldB360CastBears(t, g, me, "Runeclaw Bear")
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a creature with another name triggered Secret Summoning")
	}
	oldB360CastBears(t, g, me, "Grizzly Bears")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, copies[0], copies[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(copies[0]) || !me.Hand.Contains(copies[1]) {
		t.Error("the library copies did not go to the hand")
	}
}

// --- Unexpected Potential ------------------------------------------------------

func TestOldB360UnexpectedPotentialLetsAnyColorPayOnlyForTheChosenName(t *testing.T) {
	g := oldB360Table(t, map[int][]game.Card{0: {oldB360Card("Unexpected Potential", oldB360Unexpected)}})
	me := g.Seats[0]
	oldB360Name(t, g, me.ID, "Lightning Bolt")
	oldB360Begin(t, g)
	bolt := effectstest.HandSpell(me, "Lightning Bolt", "Instant", "{R}")
	shock := effectstest.HandSpell(me, "Shock", "Instant", "{R}")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}), "a face-down Unexpected Potential")
	oldB360Reveal(t, g, me, "Unexpected Potential")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, shock, game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(g.Seats[1].ID)}), "a spell with another name")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(g.Seats[1].ID)}); err != nil {
		t.Fatalf("Lightning Bolt paid with green mana under a face-up Unexpected Potential: %v", err)
	}
}

// --- Power Play ----------------------------------------------------------------

func TestOldB360PowerPlayMakesItsOwnerTheStartingPlayer(t *testing.T) {
	g := oldB360Table(t, map[int][]game.Card{2: {oldB360Card("Power Play", oldB360PowerPlay)}})
	if g.StartingSeat != 2 {
		t.Errorf("the starting seat is %d, want 2 (the Power Play's owner)", g.StartingSeat)
	}
	if g.Turn.ActiveSeat != 2 {
		t.Errorf("the first turn is seat %d's, want 2", g.Turn.ActiveSeat)
	}
}

func TestOldB360PowerPlayWithSeveralOwnersPicksOneOfThem(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3, 4, 5, 6} {
		g := game.NewGame()
		for i := 0; i < 4; i++ {
			deck := make([]game.Card, 20)
			for j := range deck {
				deck[j] = game.NewCard("basic-filler", uuid.Nil)
			}
			if i == 1 || i == 3 {
				deck = append(deck, oldB360Card("Power Play", oldB360PowerPlay))
			}
			if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
				t.Fatalf("AddPlayer: %v", err)
			}
		}
		if err := g.Start(rand.New(rand.NewPCG(seed, 22))); err != nil {
			t.Fatalf("Start: %v", err)
		}
		if g.StartingSeat != 1 && g.StartingSeat != 3 {
			t.Errorf("seed %d: the starting seat is %d, want one of the two Power Play owners", seed, g.StartingSeat)
		}
	}
}

func TestOldB360AHiddenAgendaTableWithoutPowerPlayKeepsTheChosenSeat(t *testing.T) {
	g := oldB360Table(t, nil)
	if g.StartingSeat < 0 || g.StartingSeat > 3 {
		t.Errorf("the starting seat is %d", g.StartingSeat)
	}
}
