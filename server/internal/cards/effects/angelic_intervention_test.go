package effects

import (
	"testing"

	"github.com/google/uuid"

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
	castCatalogSpell(t, g, "Angelic Intervention", "Instant", angelicInterventionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func aiBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2,
		Owner: owner, Controller: owner,
	})
}

func TestAngelicInterventionOffersColorlessAndTheFiveColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := aiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)

	c := pendingOfKind(g, game.PendingChoiceColor)
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
	if got := protectionsOn(t, g, bear); len(got) != 0 {
		t.Errorf("protections before the answer = %q", got)
	}
	if n := plusOneCounters(g, bear); n != 0 {
		t.Errorf("+1/+1 counters before the answer = %d, want 0", n)
	}
}

func TestAngelicInterventionColorlessStopsColourlessSources(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := aiBear(g, me.ID)
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stone Golem", TypeLine: "Artifact Creature — Golem",
		ManaCost: "{5}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	ogre := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gray Ogre", TypeLine: "Creature — Ogre",
		ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	castAngelicIntervention(t, g, bear)
	answerColor(t, g, me.ID, "C")

	if got := protectionsOn(t, g, bear); len(got) != 1 || got[0] != "colorless" {
		t.Fatalf("protections = %q, want [colorless]", got)
	}
	if n := plusOneCounters(g, bear); n != 1 {
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
		if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, findStackCard(g, colourless.InstanceID))) {
			t.Error("a colourless spell may not target a permanent with protection from colorless")
		}
		if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, findStackCard(g, red.InstanceID))) {
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
	advancePastCleanupForTest(t, g)
	if got := protectionsOn(t, g, bear); len(got) != 0 {
		t.Errorf("protections after the turn ended = %q, want none", got)
	}
	if n := plusOneCounters(g, bear); n != 1 {
		t.Errorf("the counter stays: %d", n)
	}
}

func TestAngelicInterventionAColourIsStillAChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := aiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)
	answerColor(t, g, me.ID, "R")
	if got := protectionsOn(t, g, bear); len(got) != 1 || got[0] != "red" {
		t.Fatalf("protections = %q, want [red]", got)
	}
	if n := plusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
}

func TestAngelicInterventionGivesAPlaneswalkerNoCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		ManaCost: "{2}{W}", Colors: []string{"W"}, Counters: map[string]int{"loyalty": 3},
		Owner: me.ID, Controller: me.ID,
	})
	castAngelicIntervention(t, g, walker)
	answerColor(t, g, me.ID, "C")
	if got := protectionsOn(t, g, walker); len(got) != 1 || got[0] != "colorless" {
		t.Fatalf("protections = %q, want [colorless]", got)
	}
	if n := plusOneCounters(g, walker); n != 0 {
		t.Errorf("+1/+1 counters on a planeswalker = %d, want 0", n)
	}
}

func TestAngelicInterventionTargetsOnlyYourCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := aiBear(g, opp.ID)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := aiBear(g, me.ID)
	castAngelicIntervention(t, g, bear)

	g = restoreAtKeyedPrompt(t, g, "then/angelic-intervention-protection-then-counter")
	c := pendingOfKind(g, game.PendingChoiceColor)
	if c == nil || len(c.ColorOptions) != 6 || c.ColorOptions[5] != game.ColorlessChoice {
		t.Fatalf("the restored prompt lost the colourless answer: %+v", c)
	}
	answerColor(t, g, g.Seats[g.Turn.ActiveSeat].ID, "C")
	if got := protectionsOn(t, g, bear); !hasString(got, "colorless") {
		t.Errorf("the restored answer granted no protection from colorless: %q", got)
	}
	if n := plusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters after the restored answer = %d, want 1", n)
	}
}
