package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// angelic_intervention_test.go — #5650. Angelic Intervention through the
// engine: the six-answer prompt, protection from colorless against a
// colourless source (targeting, damage, blocking), the colour answer,
// the counter only for a creature, the target restriction and the
// restore round trip.

const angelicInterventionOracle = "3d036f68-5f3a-4182-88f0-cdd685817dd0"

func castAngelicIntervention(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Angelic Intervention", "Instant", angelicInterventionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestAngelicInterventionOffersColorlessAndTheFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.AiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)

	c := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if c == nil {
		t.Fatal("no prompt after Angelic Intervention resolved")
	}
	want := []string{"W", "U", "B", "R", "G", "C"}
	if len(c.ColorOptions) != len(want) {
		t.Fatalf("options = %v, want %v", c.ColorOptions, want)
	}
	for i, o := range want {
		if c.ColorOptions[i] != o {
			t.Fatalf("options = %v, want %v", c.ColorOptions, want)
		}
	}
	if c.ColorPurpose != game.ColorForProtection {
		t.Errorf("purpose = %q, want protect", c.ColorPurpose)
	}
	// Nothing is granted and no counter lands before the answer.
	if got := effectstest.ProtectionsOn(t, g, bear); len(got) != 0 {
		t.Errorf("protections before the answer = %q", got)
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 0 {
		t.Errorf("+1/+1 counters before the answer = %d, want 0", n)
	}
}

func TestAngelicInterventionColorlessStopsColourlessSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.AiBear(g, me.ID)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stone Golem", TypeLine: "Artifact Creature — Golem",
		ManaCost: "{5}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	ogre := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gray Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	castAngelicIntervention(t, g, bear)
	effectstest.AnswerColor(t, g, me.ID, "C")

	if got := effectstest.ProtectionsOn(t, g, bear); len(got) != 1 || got[0] != "colorless" {
		t.Fatalf("protections = %q, want [colorless]", got)
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1 (it is a creature)", n)
	}

	card := func(id uuid.UUID) *game.Card {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				return &g.Battlefield.Cards[i]
			}
		}
		t.Fatalf("card %v is not on the battlefield", id)
		return nil
	}
	// Blocking, CR 702.16f.
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(card(bear), card(golem)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a colourless Golem blocking: reason = %q, want protection", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(bear), card(ogre)); !r.Legal() {
			t.Errorf("a red Ogre blocking: reason = %q, want legal", r.Reason)
		}
	})

	// Targeting, CR 702.16b: a colourless spell may not, a red one may.
	colourless := game.NewCard("Scour from Existence", opp.ID)
	colourless.TypeLine = "Instant"
	colourless.ManaCost = "{7}"
	red := game.NewCard("Shock", opp.ID)
	red.TypeLine = "Instant"
	red.ManaCost = "{R}"
	red.Colors = []string{"R"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(colourless)
		g.Stack.PushTop(red)
	})
	g.ReadSnapshot(func() {
		c := card(bear)
		if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, effectstest.FindStackCard(g, colourless.InstanceID))) {
			t.Error("a colourless spell may not target a permanent with protection from colorless")
		}
		if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, effectstest.FindStackCard(g, red.InstanceID))) {
			t.Error("a red spell may target it")
		}
	})

	// Damage, CR 702.16e: the colourless 2 is prevented, the red 1 dealt.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(colourless.InstanceID, bear, 2); err != nil {
			t.Fatalf("colourless damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(red.InstanceID, bear, 1); err != nil {
			t.Fatalf("red damage: %v", err)
		}
	})
	if marked := card(bear).DamageMarked; marked != 1 {
		t.Errorf("damage marked = %d, want 1", marked)
	}

	// Until end of turn (CR 514.2).
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.ProtectionsOn(t, g, bear); len(got) != 0 {
		t.Errorf("protections after the turn ended = %q, want none", got)
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("the counter stays: %d", n)
	}
}

func TestAngelicInterventionAColourIsStillAChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.AiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if got := effectstest.ProtectionsOn(t, g, bear); len(got) != 1 || got[0] != "red" {
		t.Fatalf("protections = %q, want [red]", got)
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
}

func TestAngelicInterventionGivesAPlaneswalkerNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		ManaCost: "{2}{W}", Colors: []string{"W"}, Counters: map[string]int{"loyalty": 3},
		Owner: me.ID, Controller: me.ID,
	})
	castAngelicIntervention(t, g, walker)
	effectstest.AnswerColor(t, g, me.ID, "C")
	if got := effectstest.ProtectionsOn(t, g, walker); len(got) != 1 || got[0] != "colorless" {
		t.Fatalf("protections = %q, want [colorless]", got)
	}
	if n := effectstest.PlusOneCounters(g, walker); n != 0 {
		t.Errorf("+1/+1 counters on a planeswalker = %d, want 0", n)
	}
}

func TestAngelicInterventionTargetsOnlyYourCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.AiBear(g, opp.ID)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mind Stone", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	for name, id := range map[string]uuid.UUID{"an opponent's creature": theirs, "your noncreature artifact": rock} {
		spell := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: spell, Name: "Angelic Intervention", TypeLine: "Instant",
			OracleID: angelicInterventionOracle, Owner: me.ID, Controller: me.ID,
		})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		err := g.CastSpell(me.ID, spell, game.CastSpellParams{
			Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		})
		if err == nil {
			t.Errorf("%s was accepted as a target", name)
		}
	}
}

func TestAngelicInterventionPromptIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.AiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)

	g = effectstest.RestoreAtKeyedPrompt(t, g, "then/angelic-intervention-protection-then-counter")
	c := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if c == nil || len(c.ColorOptions) != 6 || c.ColorOptions[5] != game.ColorlessChoice {
		t.Fatalf("the restored prompt lost the colourless answer: %+v", c)
	}
	effectstest.AnswerColor(t, g, g.Seats[g.Turn.ActiveSeat].ID, "C")
	if got := effectstest.ProtectionsOn(t, g, bear); !effectstest.HasString(got, "colorless") {
		t.Errorf("the restored answer granted no protection from colorless: %q", got)
	}
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters after the restored answer = %d, want 1", n)
	}
}
