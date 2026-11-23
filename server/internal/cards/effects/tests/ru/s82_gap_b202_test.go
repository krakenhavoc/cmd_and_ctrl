package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b202_test.go — cards missing from sets since AFR, batch 202
// (#5948). Each test plays the real catalog entry and asserts what its text
// says.

const (
	gapB202Silumgar  = "44581d4f-746b-43b8-a34d-16dc383ab70e"
	gapB202Amalgam   = "8499b3ec-9a61-47a4-a880-e8230b41c81b"
	gapB202Symbiont  = "34439fb2-0272-4624-aa4c-9ea439b60489"
	gapB202Gargoyle  = "681ca1eb-2ef5-4d60-b834-58bd9b5a47b8"
	gapB202Liege     = "caf89756-fb78-4e58-a730-001393ffa0a8"
	gapB202Shoat     = "d385c2e1-a62e-43ad-80ef-a6178deac82c"
	gapB202Webweaver = "c8cf9387-54df-45b1-ad8d-9de57ed8916b"
)

func TestGapB202EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB202Silumgar:  "Dragonlord Silumgar",
		gapB202Amalgam:   "Prized Amalgam",
		gapB202Symbiont:  "Psychic Symbiont",
		gapB202Gargoyle:  "Tower Gargoyle",
		gapB202Liege:     "Deathbringer Liege",
		gapB202Shoat:     "Brindle Shoat",
		gapB202Webweaver: "Webweaver Changeling",
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

// --- Dragonlord Silumgar -----------------------------------------------------

func TestGapB202SilumgarStealsUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	silumgar, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Dragonlord Silumgar", TypeLine: "Legendary Creature — Elder Dragon",
		OracleID: gapB202Silumgar, Power: 3, Toughness: 5}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "deathtouch"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, silumgar, kw) {
			t.Errorf("Silumgar does not have %s", kw)
		}
	}
	if c := effectstest.B12Card(t, g, bear); c.Controller != me.ID {
		t.Fatal("Silumgar did not gain control of the Bear")
	}
	effectstest.KillCreature(t, g, me.ID, silumgar)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, bear); c.Controller != opp.ID {
		t.Error("the Bear did not go home when Silumgar left")
	}
}

// --- Psychic Symbiont --------------------------------------------------------

func TestGapB202PsychicSymbiontMakesTheOpponentDiscardAndDrawsYouACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	pitch := effectstest.PushCatalogHandCard(opp, "Pitch", "Sorcery", "")
	hand := me.Hand.Size()
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Psychic Symbiont", TypeLine: "Creature — Nightmare Horror", OracleID: gapB202Symbiont, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || !opp.Hand.Contains(keep) {
		t.Error("the chosen card was not the one discarded")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("my hand %d -> %d, want one drawn card (the Symbiont left the hand as it was cast)", hand, me.Hand.Size())
	}
}

// --- Tower Gargoyle ----------------------------------------------------------

func TestGapB202TowerGargoyleFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Tower Gargoyle", "Artifact Creature — Gargoyle", gapB202Gargoyle, 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Tower Gargoyle does not have flying")
	}
}

// --- Brindle Shoat -----------------------------------------------------------

func TestGapB202BrindleShoatLeavesABoarWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shoat := effectstest.B12Push(g, me.ID, "Brindle Shoat", "Creature — Boar", gapB202Shoat, 1, 1)
	effectstest.KillCreature(t, g, me.ID, shoat)
	effectstest.PassPriorityAroundTable(t, g)
	n, boar := effectstest.B14Tokens(g, me.ID, "Boar")
	if n != 1 {
		t.Fatalf("%d Boar tokens, want 1", n)
	}
	if p, tough := effectstest.PtOf(t, g, boar.InstanceID); p != 3 || tough != 3 {
		t.Errorf("the Boar is %d/%d, want 3/3", p, tough)
	}
	if len(boar.Colors) != 1 || boar.Colors[0] != "G" {
		t.Errorf("the Boar's colours are %v, want green", boar.Colors)
	}
}

// --- Webweaver Changeling ----------------------------------------------------

func TestGapB202WebweaverChangelingGainsFiveLifeWithThreeCreatureCards(t *testing.T) {
	for _, tc := range []struct {
		creatures int
		wantGain  int
	}{{2, 0}, {3, 5}, {4, 5}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < tc.creatures; i++ {
			effectstest.PushGraveyardCardTyped(me, "Dead", "Creature — Bear")
		}
		// A noncreature card never counts.
		effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
		life := me.Life
		id := effectstest.Msc17Play(t, g, me, game.Card{Name: "Webweaver Changeling", TypeLine: "Creature — Shapeshifter", OracleID: gapB202Webweaver, Power: 3, Toughness: 5})
		if me.Life != life+tc.wantGain {
			t.Errorf("%d creature cards: life %d -> %d, want %+d", tc.creatures, life, me.Life, tc.wantGain)
		}
		for _, kw := range []string{"changeling", "reach"} {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("Webweaver Changeling does not have %s", kw)
			}
		}
	}
}

// --- Deathbringer Liege ------------------------------------------------------

func TestGapB202DeathbringerLiegeGivesWhiteAndBlackCreaturesPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Deathbringer Liege", "Creature — Horror", gapB202Liege, 3, 4)
	mk := func(owner uuid.UUID, colors ...string) uuid.UUID {
		id := effectstest.B12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Colors = colors
				}
			}
			g.BumpLayerVersionForTest()
		})
		return id
	}
	w, b, both, red, theirs := mk(me.ID, "W"), mk(me.ID, "B"), mk(me.ID, "W", "B"), mk(me.ID, "R"), mk(opp.ID, "W")
	for _, want := range []struct {
		id  uuid.UUID
		who string
		p   int
		t   int
	}{
		{liege, "the Liege", 3, 4}, {w, "white", 3, 3}, {b, "black", 3, 3}, {both, "white and black", 4, 4},
		{red, "red", 2, 2}, {theirs, "an opponent's white creature", 2, 2},
	} {
		if p, tough := effectstest.PtOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", want.who, p, tough, want.p, want.t)
		}
	}
}

func TestGapB202DeathbringerLiegeTapsOnAWhiteSpellAndDestroysTappedOnABlackOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deathbringer Liege", "Creature — Horror", gapB202Liege, 3, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	// A black spell with the victim untapped: the trigger destroys nothing.
	effectstest.B14CastAt(t, g, me, "Black Spell", "Sorcery", "", []string{"B"}, game.CastSpellParams{})
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	answerLiegeMay(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, victim) {
		t.Fatal("an untapped creature was destroyed")
	}

	// A white spell taps it.
	effectstest.B14CastAt(t, g, me, "White Spell", "Sorcery", "", []string{"W"}, game.CastSpellParams{})
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	answerLiegeMay(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, victim); !c.Tapped {
		t.Fatal("the white spell's trigger did not tap the creature")
	}

	// A black spell now destroys it, unless declined.
	effectstest.B14CastAt(t, g, me, "Black Spell", "Sorcery", "", []string{"B"}, game.CastSpellParams{})
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	answerLiegeMay(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the tapped creature was not destroyed")
	}
}

func answerLiegeMay(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	if effectstest.LatestConfirmFor(g, who) != nil {
		effectstest.AnswerMayChoice(t, g, who, yes)
		return
	}
	if prompt := effectstest.AwaitTriggerPrompt(t, g, who); prompt != nil {
		if err := g.ResolveTriggerPrompt(prompt.ID, who, yes); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
}

// --- Prized Amalgam ----------------------------------------------------------

func TestGapB202PrizedAmalgamIgnoresACreatureEnteringFromTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	amalgam := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: amalgam, Name: "Prized Amalgam", TypeLine: "Creature — Zombie", OracleID: gapB202Amalgam,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})

	// A creature entering from the hand does not trigger it.
	effectstest.EnterFromHand(t, g, me.ID, "Hand Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, amalgam) {
		t.Fatal("the Amalgam returned for a creature that entered from the hand")
	}
}

func TestGapB202PrizedAmalgamReturnsAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	amalgam := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: amalgam, Name: "Prized Amalgam", TypeLine: "Creature — Zombie", OracleID: gapB202Amalgam,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	other := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")

	// A creature entering from your graveyard triggers it.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, other); err != nil {
		t.Fatalf("reanimate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, amalgam) {
		t.Fatal("the Amalgam returned before the end step")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, amalgam) {
		t.Fatal("the Amalgam did not return at the beginning of the next end step")
	}
	if c := effectstest.B12Card(t, g, amalgam); !c.Tapped {
		t.Error("the Amalgam did not return tapped")
	}
}
