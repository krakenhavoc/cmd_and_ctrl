package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b233_test.go — Commander Legends: Battle for Baldur's Gate
// sorceries (GAP batch 233, #5948), played through the real catalog entries.

const (
	gapB233RecruitmentDrive = "d37e37d7-ce66-42cb-862e-fc5c414cfeb6"
	gapB233SculptedSunburst = "aee0388f-544e-4c28-a34f-29cc0a97d424"
	gapB233StickTogether    = "4abda0ad-0a2b-4bd9-81c7-56329e62311d"
)

func gapB233Cast(t *testing.T, g *game.Game, id uuid.UUID, name, oracle string) {
	t.Helper()
	me := g.Seats[0]
	effectstest.D5681Spell(g, me, id, name, "Sorcery", "", oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

func gapB233Answer(t *testing.T, g *game.Game, seat uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, seat)
	if c == nil {
		t.Fatalf("seat %v was not asked: %+v", seat, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, seat, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB233CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB233RecruitmentDrive: "Recruitment Drive", gapB233SculptedSunburst: "Sculpted Sunburst",
		gapB233StickTogether: "Stick Together",
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

func TestGapB233RecruitmentDriveTable(t *testing.T) {
	for _, tc := range []struct {
		natural int
		name    string
		count   int
	}{{9, "Soldier", 2}, {10, "Knight", 2}, {19, "Knight", 2}, {20, "Knight", 3}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		gapB233Cast(t, g, id, "Recruitment Drive", gapB233RecruitmentDrive)
		effectstest.PassPriorityAroundTable(t, g)
		toks := effectstest.AvB34Tokens(g, g.Seats[0].ID, tc.name)
		if len(toks) != tc.count {
			t.Errorf("natural %d: %d %s tokens, want %d", tc.natural, len(toks), tc.name, tc.count)
			continue
		}
		wantPT := 2
		if tc.name == "Soldier" {
			wantPT = 1
		}
		for _, tk := range toks {
			if tk.Power != wantPT || tk.Toughness != wantPT {
				t.Errorf("natural %d: token is %d/%d, want %d/%d", tc.natural, tk.Power, tk.Toughness, wantPT, wantPT)
			}
		}
	}
}

func TestGapB233StickTogetherKeepsAPartyAndSacrificesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	cleric := effectstest.B12Push(g, me.ID, "Cleric", "Creature — Human Cleric", "", 1, 1)
	// A Cleric Wizard may stand for either type, so with the Cleric it is a party of two.
	both := effectstest.B12Push(g, me.ID, "Both", "Creature — Human Cleric Wizard", "", 1, 1)
	spare := effectstest.B12Push(g, me.ID, "Spare Cleric", "Creature — Human Cleric", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	rogue := effectstest.B12Push(g, opp.ID, "Rogue", "Creature — Human Rogue", "", 1, 1)
	oppBear := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	gapB233Cast(t, g, uuid.New(), "Stick Together", gapB233StickTogether)
	effectstest.PassPriorityAroundTable(t, g)

	// Two Clerics are not a party: both would stand for the same type.
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("I was not asked: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("%d candidates, want my three party-type creatures", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{cleric, spare}); err == nil {
		t.Fatal("two Clerics were accepted as a party")
	}
	gapB233Answer(t, g, me.ID, cleric, both)
	gapB233Answer(t, g, opp.ID, rogue)

	for name, id := range map[string]uuid.UUID{"the spare Cleric": spare, "my Bear": bear, "their Bear": oppBear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"my Cleric": cleric, "my Cleric Wizard": both, "their Rogue": rogue, "my land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed", name)
		}
	}
}

func TestGapB233StickTogetherSacrificesEverythingWithNoPartyType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[0]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gapB233Cast(t, g, uuid.New(), "Stick Together", gapB233StickTogether)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a creature with no party type survived")
	}
}

func TestGapB233SculptedSunburstExilesEveryCreatureNobodyChose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Push(g, me.ID, "Big", "Creature — Bear", "", 4, 4)
	small := effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 1, 1)
	strong := effectstest.B12Push(g, opp.ID, "Strong", "Creature — Bear", "", 5, 5)
	mid := effectstest.B12Push(g, opp.ID, "Mid", "Creature — Bear", "", 3, 3)
	weak := effectstest.B12Push(g, opp.ID, "Weak", "Creature — Bear", "", 4, 4)
	loner := effectstest.B12Push(g, g.Seats[2].ID, "Loner", "Creature — Bear", "", 6, 6)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	gapB233Cast(t, g, uuid.New(), "Sculpted Sunburst", gapB233SculptedSunburst)
	effectstest.PassPriorityAroundTable(t, g)
	gapB233Answer(t, g, me.ID, big)

	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if c == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 2 {
		t.Errorf("%d candidates, want the two with power 4 or less", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{strong}); err == nil {
		t.Fatal("a creature with greater power was accepted")
	}
	gapB233Answer(t, g, opp.ID, weak)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, g.Seats[2].ID) != nil {
		t.Error("an opponent with no eligible creature was asked")
	}
	for name, id := range map[string]uuid.UUID{"my other creature": small, "a bigger creature": strong, "an unchosen creature": mid, "the loner": loner} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was not exiled", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"my chosen creature": big, "their chosen creature": weak, "a land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was exiled", name)
		}
	}
}

func TestGapB233SculptedSunburstWithNoCreatureOfMineExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	theirs := effectstest.B12Push(g, g.Seats[1].ID, "Bear", "Creature — Bear", "", 2, 2)
	gapB233Cast(t, g, uuid.New(), "Sculpted Sunburst", gapB233SculptedSunburst)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("a creature was exiled though I chose none")
	}
}
