package testss82m

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b269_cards_test.go — S82 OLD batch 269 (Conspiracy: Take the Crown,
// #5540): Hired Heist and Hymn of the Wilds. Each test starts a real game with
// the conspiracy in the sideboard (ADR 0227).

const (
	oldB269Heist = "b0d84455-dc35-45d3-901e-8d28363142f4"
	oldB269Hymn  = "dfa62414-48c6-4e5c-8609-1ce507b41505"
)

func oldB269Table(t *testing.T, name, oracle string) (*game.Game, *game.Player) {
	t.Helper()
	g := game.NewGame()
	side := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Conspiracy", OracleID: oracle, IsSideboard: true}
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if i == 0 {
			deck = append(deck, side)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	me := g.Seats[0]
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == me.ID {
			if err := g.ResolveCardNameChoice(c.ID, me.ID, "Grizzly Bears"); err != nil {
				t.Fatalf("ResolveCardNameChoice: %v", err)
			}
			break
		}
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	return g, me
}

func TestOldB269EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{oldB269Heist: "Hired Heist", oldB269Hymn: "Hymn of the Wilds"} {
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

func oldB269Bear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	c := effectstest.ApaCreature(name, "", 2, 2)
	c.TypeLine = "Creature — Bear"
	return effectstest.ApaPush(g, owner, owner, c)
}

func oldB269Attack(t *testing.T, g *game.Game, me *game.Player, ids ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range ids {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB269HiredHeistDrawsForThePaidNamedCreatureOnly(t *testing.T) {
	g, me := oldB269Table(t, "Hired Heist", oldB269Heist)
	named := oldB269Bear(g, me.ID, "Grizzly Bears")
	other := oldB269Bear(g, me.ID, "Runeclaw Bear")
	for _, c := range g.ConspiraciesOf(me.ID) {
		if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err != nil {
			t.Fatalf("reveal_agenda: %v", err)
		}
	}
	before := len(me.Hand.Cards)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{named, other} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the creature with another name offered a draw")
	}
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d, want %d (one card for the named creature)", got, before+1)
	}
}

func TestOldB269HiredHeistFaceDownDoesNothing(t *testing.T) {
	g, me := oldB269Table(t, "Hired Heist", oldB269Heist)
	named := oldB269Bear(g, me.ID, "Grizzly Bears")
	before := len(me.Hand.Cards)
	oldB269Attack(t, g, me, named)
	if effectstest.HasPayUnlessFor(g, me.ID) || len(me.Hand.Cards) != before {
		t.Fatal("a face-down Hired Heist triggered")
	}
}

func TestOldB269HymnOfTheWildsDiscountsTheFirstCreatureSpellAndBarsInstants(t *testing.T) {
	g, me := oldB269Table(t, "Hymn of the Wilds", oldB269Hymn)
	first := effectstest.HandSpell(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	second := effectstest.HandSpell(me, "Runeclaw Bear", "Creature — Bear", "{1}{G}")
	bolt := effectstest.HandSpell(me, "Shock", "Instant", "{R}")
	sorcery := effectstest.HandSpell(me, "Divination", "Sorcery", "{2}{U}")

	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an instant was cast under Hymn of the Wilds")
	}
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery was cast under Hymn of the Wilds")
	}
	// The first creature spell costs {1} less: {G} is enough. A pool with only
	// {G} can pay it, and cannot pay the second one.
	me.ManaPool = game.ManaPool{}
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the first creature spell for {G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true}), "the second creature spell")
}
