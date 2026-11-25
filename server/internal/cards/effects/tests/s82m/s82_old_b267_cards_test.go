package testss82m

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b267_cards_test.go — S82 OLD batch 267 (Conspiracy: Take the Crown,
// #5540): Summoner's Bond, Hold the Perimeter, Echoing Boon, Natural Unity,
// Adriana's Valor. Every test starts a real game with the conspiracy in the
// sideboard, so it goes through the command-zone start and the hidden agenda
// prompt (ADR 0227).

const (
	oldB267Bond      = "07b6326a-b73d-422b-bcb5-c508aa3fc558"
	oldB267Perimeter = "20f7f33d-0c49-49f7-80fa-aed86e33a528"
	oldB267Boon      = "22e89c6d-2304-4def-8a74-e53a6271e9b9"
	oldB267Unity     = "3deb19f6-c1a3-4bca-a6ae-df5d0b42a66e"
	oldB267Valor     = "512a9ab9-ddd4-4872-b4f4-dee56f9f5575"
)

func TestOldB267EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB267Bond: "Summoner's Bond", oldB267Perimeter: "Hold the Perimeter", oldB267Boon: "Echoing Boon",
		oldB267Unity: "Natural Unity", oldB267Valor: "Adriana's Valor",
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

func oldB267Card(name, oracle string) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Conspiracy", OracleID: oracle, IsSideboard: true}
}

func oldB267Table(t *testing.T, side game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := game.NewGame()
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
	return g, g.Seats[0]
}

// oldB267Name answers the open hidden agenda prompt.
func oldB267Name(t *testing.T, g *game.Game, me *game.Player, name string) {
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

func oldB267Begin(t *testing.T, g *game.Game) {
	t.Helper()
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
}

func oldB267Reveal(t *testing.T, g *game.Game, me *game.Player, name string) {
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

func oldB267Named(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	c := effectstest.ApaCreature(name, "", p, tough)
	c.TypeLine = typeLine
	return effectstest.ApaPush(g, owner, owner, c)
}

func oldB267Counters(g *game.Game, id uuid.UUID) int {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func TestOldB267HoldThePerimeterMakesTokensAtEachPlayersFirstUpkeepOnly(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Hold the Perimeter", oldB267Perimeter))
	if g.PendingDealSeat != nil {
		t.Fatal("a face-up conspiracy held up the deal")
	}
	oldB267Begin(t, g)
	// Walk through two full rounds of turns.
	for i := 0; i < 400 && g.Turn.Seq <= 8; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Turn.Seq <= 8 {
		t.Fatalf("stopped on turn %d", g.Turn.Seq)
	}
	soldiers, goblins := 0, 0
	for _, c := range g.Battlefield.Cards {
		switch c.Name {
		case "Soldier":
			if c.Controller != me.ID {
				t.Errorf("a Soldier under %s", c.Controller)
			}
			soldiers++
		case "Goblin":
			if c.Controller == me.ID {
				t.Error("the conspiracy's owner got a Goblin")
			}
			goblins++
		}
	}
	if soldiers != 1 || goblins != 3 {
		t.Errorf("%d Soldiers and %d Goblins, want 1 and 3", soldiers, goblins)
	}
}

// oldB267BoltSpell casts Lightning Bolt at the target from my hand, paying
// {R}, and returns the spell's card id.
func oldB267BoltSpell(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatalf("cast Lightning Bolt: %v", err)
	}
	return id
}

func oldB267Damage(g *game.Game, id uuid.UUID) int {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.DamageMarked
}

func TestOldB267EchoingBoonCopiesASpellAtTheNamedCreatureOnlyFaceUp(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Echoing Boon", oldB267Boon))
	oldB267Name(t, g, me, "Stone Wall")
	oldB267Begin(t, g)
	wall := oldB267Named(g, me.ID, "Stone Wall", "Creature — Wall", 0, 10)
	other := oldB267Named(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 10)

	// Face down: nothing happens (CR 315.5).
	oldB267BoltSpell(t, g, me, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil || oldB267Damage(g, wall) != 3 {
		t.Fatalf("a face-down Echoing Boon acted (damage %d)", oldB267Damage(g, wall))
	}

	oldB267Reveal(t, g, me, "Echoing Boon")
	// A creature with another name: no trigger.
	oldB267BoltSpell(t, g, me, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("Echoing Boon triggered for a creature with another name")
	}

	// The named creature: the copy hits it again (3 more) once new targets
	// are kept.
	oldB267BoltSpell(t, g, me, wall)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4 && effectstest.CopyPromptFor(g, me.ID) != nil; i++ {
		p := effectstest.CopyPromptFor(g, me.ID)
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: wall}); err != nil {
			t.Fatalf("keeping the copy's target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := oldB267Damage(g, wall); got != 9 {
		t.Errorf("Stone Wall has %d damage, want 9 (three Bolts and a copy: 3 + 3 + 3)", got)
	}
}

func TestOldB267NaturalUnityLetsTheNamedCreaturePayForACounter(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Natural Unity", oldB267Unity))
	oldB267Name(t, g, me, "Grizzly Bears")
	oldB267Begin(t, g)
	named := oldB267Named(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	other := oldB267Named(g, me.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	theirs := oldB267Named(g, g.Seats[1].ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	// Face down it grants nothing.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a face-down Natural Unity granted its trigger")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	oldB267NextMyTurn(t, g, me)

	oldB267Reveal(t, g, me, "Natural Unity")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the named creature's beginning-of-combat trigger did not ask for {G}")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("another creature was offered the counter")
	}
	if oldB267Counters(g, named) != 1 || oldB267Counters(g, other) != 0 || oldB267Counters(g, theirs) != 0 {
		t.Errorf("counters: named %d, other %d, theirs %d; want 1, 0, 0",
			oldB267Counters(g, named), oldB267Counters(g, other), oldB267Counters(g, theirs))
	}
}

// oldB267NextMyTurn plays steps until seat 0 is the active player in a new turn.
func oldB267NextMyTurn(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	start := g.Turn.Seq
	for i := 0; i < 400 && !(g.Turn.Seq > start && g.Seats[g.Turn.ActiveSeat].ID == me.ID && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Turn.Seq == start {
		t.Fatal("never reached my next turn")
	}
}

func TestOldB267AdrianasValorGivesIndestructibleForW(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Adriana's Valor", oldB267Valor))
	oldB267Name(t, g, me, "Grizzly Bears")
	oldB267Begin(t, g)
	named := oldB267Named(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	oldB267Reveal(t, g, me, "Adriana's Valor")
	opp := g.Seats[1].ID
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(named, opp); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, named, "indestructible") {
		t.Error("the attacker did not gain indestructible")
	}
}

func TestOldB267AdrianasValorDoesNothingWhenDeclinedOrForAnotherName(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Adriana's Valor", oldB267Valor))
	oldB267Name(t, g, me, "Grizzly Bears")
	oldB267Begin(t, g)
	named := oldB267Named(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	other := oldB267Named(g, me.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	oldB267Reveal(t, g, me, "Adriana's Valor")
	opp := g.Seats[1].ID
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.FillPoolColored(me, "W", 1)
	for _, id := range []uuid.UUID{named, other} {
		if err := g.DeclareAttacker(id, opp); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a second prompt was asked: the other creature has another name")
	}
	for _, id := range []uuid.UUID{named, other} {
		if effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("creature %s is indestructible", id)
		}
	}
}

func TestOldB267SummonersBondFetchesTheOtherNamedCreature(t *testing.T) {
	g, me := oldB267Table(t, oldB267Card("Summoner's Bond", oldB267Bond))
	oldB267Name(t, g, me, "Grizzly Bears")
	oldB267Name(t, g, me, "Goblin Guide")
	oldB267Begin(t, g)
	oldB267Reveal(t, g, me, "Summoner's Bond")
	guide := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Goblin Guide", TypeLine: "Creature — Goblin Scout"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		game.Card{Name: "Goblin Guide Tribute", TypeLine: "Instant"})
	bears := effectstest.HandSpell(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, bears, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Grizzly Bears: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, guide[0])
	if !me.Hand.Contains(guide[0]) {
		t.Error("Goblin Guide did not reach the hand")
	}
	if me.Hand.Contains(guide[1]) {
		t.Error("the Grizzly Bears (the name just cast) was fetched")
	}
}
