package testss82m

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 361 (Conspiracy, #5557): Sentinel Dispatch, Double Stroke,
// Secrets of Paradise, each started from a real sideboard (ADR 0227).

const (
	oldB361Dispatch = "7df49476-7b6c-4b5d-bede-463a7d2e0a58"
	oldB361Stroke   = "e25ab152-e166-44df-8f92-668bab0e9619"
	oldB361Secrets  = "f309fe9f-0732-4b81-a342-a394225cbbb1"
)

func TestOldB361EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB361Dispatch: "Sentinel Dispatch", oldB361Stroke: "Double Stroke", oldB361Secrets: "Secrets of Paradise",
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

func oldB361Table(t *testing.T, name, oracle string) (*game.Game, *game.Player) {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if i == 0 {
			deck = append(deck, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Conspiracy", OracleID: oracle, IsSideboard: true})
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g, g.Seats[0]
}

func oldB361Name(t *testing.T, g *game.Game, me *game.Player, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName && c.Chooser == me.ID {
			if err := g.ResolveCardNameChoice(c.ID, me.ID, name); err != nil {
				t.Fatalf("ResolveCardNameChoice(%q): %v", name, err)
			}
			return
		}
	}
	t.Fatal("no hidden agenda prompt is open")
}

func oldB361Begin(t *testing.T, g *game.Game) {
	t.Helper()
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

func oldB361Reveal(t *testing.T, g *game.Game, me *game.Player, name string) {
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

func TestOldB361SentinelDispatchMakesOneConstructAtTheGamesFirstUpkeepOnly(t *testing.T) {
	g, me := oldB361Table(t, "Sentinel Dispatch", oldB361Dispatch)
	oldB361Begin(t, g)
	for i := 0; i < 400 && g.Turn.Seq <= 8; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Turn.Seq <= 8 {
		t.Fatalf("stopped on turn %d", g.Turn.Seq)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" {
			n++
			if c.Controller != me.ID || !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 {
				t.Errorf("Construct under %s: %+v", c.Controller, c)
			}
			if !game.HasKeyword(&c, "defender") {
				t.Error("the Construct lacks defender")
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Constructs after two rounds, want exactly 1", n)
	}
}

func TestOldB361SecretsOfParadiseGrantsManaOnlyToNamedCreaturesFaceUp(t *testing.T) {
	g, me := oldB361Table(t, "Secrets of Paradise", oldB361Secrets)
	oldB361Name(t, g, me, "Stone Wall")
	oldB361Begin(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	wall := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Stone Wall", "", 0, 4))
	bears := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Grizzly Bears", "", 2, 2))
	theirs := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, effectstest.ApaCreature("Stone Wall", "", 0, 4))
	if n := effectstest.GrantedManaRowsOf(g, wall); n != 0 {
		t.Fatalf("a face-down Secrets of Paradise granted %d mana rows", n)
	}
	oldB361Reveal(t, g, me, "Secrets of Paradise")
	if n := effectstest.GrantedManaRowsOf(g, wall); n != 1 {
		t.Errorf("the named creature has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, bears); n != 0 {
		t.Errorf("a creature with another name has %d granted rows", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's creature with the name has %d granted rows", n)
	}
}

func TestOldB361DoubleStrokeCopiesOnlyANamedSpellFaceUp(t *testing.T) {
	g, me := oldB361Table(t, "Double Stroke", oldB361Stroke)
	oldB361Name(t, g, me, "Lightning Bolt")
	oldB361Begin(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	wall := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Stone Wall", "", 0, 40))
	damage := func() int {
		c := effectstest.FindBattlefieldCardForTest(g, wall)
		if c == nil {
			return -1
		}
		return c.DamageMarked
	}
	bolt := func(name string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: "{R}",
			OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
		effectstest.FillPoolColored(me, "R", 1)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
	}
	// Face down: no trigger (CR 315.5b).
	bolt("Lightning Bolt")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || damage() != 3 {
		t.Fatalf("a face-down Double Stroke acted (damage %d)", damage())
	}
	oldB361Reveal(t, g, me, "Double Stroke")
	// A spell with another name: no trigger.
	bolt("Shock")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("Double Stroke triggered for a spell with another name")
	}
	base := damage()
	// The chosen name: the copy hits again once its targets are kept.
	bolt("Lightning Bolt")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && effectstest.CopyPromptFor(g, me.ID) != nil; i++ {
		p := effectstest.CopyPromptFor(g, me.ID)
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: wall}); err != nil {
			t.Fatalf("keeping the copy's target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := damage(); got != base+6 {
		t.Errorf("Stone Wall has %d damage, want %d (the Bolt and its copy)", got, base+6)
	}
}
