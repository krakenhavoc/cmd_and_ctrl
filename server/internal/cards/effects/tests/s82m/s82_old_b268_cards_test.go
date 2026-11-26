package testss82m

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b268_cards_test.go — S82 OLD batch 268 (Conspiracy: Take the Crown,
// #5540): Incendiary Dissent and Assemble the Rank and Vile. Each test starts a
// real game with the conspiracy in the sideboard (ADR 0227).

const (
	oldB268Dissent  = "70caaa64-9623-4d05-9843-18460fea963b"
	oldB268Assemble = "8c912b1f-4584-4945-9931-f6850e5ebfa4"
)

func TestOldB268EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB268Dissent: "Incendiary Dissent", oldB268Assemble: "Assemble the Rank and Vile",
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

func oldB268Table(t *testing.T, name, oracle string) (*game.Game, *game.Player) {
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

func oldB268Reveal(t *testing.T, g *game.Game, me *game.Player, name string) {
	t.Helper()
	for _, c := range g.ConspiraciesOf(me.ID) {
		if c.Name == name {
			if err := g.PerformSpecialAction(me.ID, c.InstanceID, game.SpecialActionRevealAgenda, game.SpecialActionParams{}); err != nil {
				t.Fatalf("reveal_agenda: %v", err)
			}
			return
		}
	}
	t.Fatalf("%s is not in the command zone", name)
}

func oldB268Bear(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	c := effectstest.ApaCreature(name, "", 2, 2)
	c.TypeLine = "Creature — Bear"
	return effectstest.ApaPush(g, owner, owner, c)
}

func TestOldB268IncendiaryDissentGivesFirebreathingToTheNamedCreatureOnly(t *testing.T) {
	g, me := oldB268Table(t, "Incendiary Dissent", oldB268Dissent)
	named := oldB268Bear(g, me.ID, "Grizzly Bears")
	other := oldB268Bear(g, me.ID, "Runeclaw Bear")
	theirs := oldB268Bear(g, g.Seats[1].ID, "Grizzly Bears")

	// Face down: no ability to activate (CR 315.5).
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, named, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a face-down Incendiary Dissent granted an ability")
	}
	oldB268Reveal(t, g, me, "Incendiary Dissent")
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a creature with another name was given the ability")
	}
	if err := g.ActivateCatalogAbility(g.Seats[1].ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("an opponent's creature with the chosen name was given the ability")
	}
	effectstest.ApaActivate(t, g, me, named, 0, game.ActivateAbilityParams{})
	if got := effectstest.EffectivePower(t, g, named); got != 3 {
		t.Errorf("power %d after {R}, want 3", got)
	}
}

func TestOldB268AssembleTheRankAndVilePaysBForATappedZombieWhenTheNamedCreatureDies(t *testing.T) {
	g, me := oldB268Table(t, "Assemble the Rank and Vile", oldB268Assemble)
	oldB268Reveal(t, g, me, "Assemble the Rank and Vile")
	named := oldB268Bear(g, me.ID, "Grizzly Bears")
	other := oldB268Bear(g, me.ID, "Runeclaw Bear")

	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature with another name offered a Zombie")
	}

	effectstest.KillCreature(t, g, me.ID, named)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID {
			n++
			if !c.Tapped {
				t.Error("the Zombie did not enter tapped")
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Zombies, want 1", n)
	}
}

func TestOldB268AssembleTheRankAndVileDeclinedMakesNothing(t *testing.T) {
	g, me := oldB268Table(t, "Assemble the Rank and Vile", oldB268Assemble)
	oldB268Reveal(t, g, me, "Assemble the Rank and Vile")
	named := oldB268Bear(g, me.ID, "Grizzly Bears")
	effectstest.KillCreature(t, g, me.ID, named)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" {
			t.Fatal("a Zombie was created without paying")
		}
	}
}
