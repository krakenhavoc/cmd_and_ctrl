package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// s82_old_b445_test.go — Innistrad creatures with keywords, S82 OLD batch 445
// (#5572), played through the real catalog entries.

const (
	oldB445BatOracle      = "209de43d-0cdd-4075-9f41-88e3d2ff90df"
	oldB445OverseerOracle = "ec22386d-42fd-40f9-b800-a043e80faabe"
)

var oldB445W0 = effectstest.IwCard{Oracle: "abcfcc71-77c0-46e7-93c3-726967946b12", Front: "Grizzled Outcasts", FrontType: "Creature — Human Werewolf", Cost: "{4}{G}", FrontText: effectstest.IwFrontText, Fp: "4", Ft: "4",
	Back: "Krallenhorde Wantons", BackText: effectstest.IwBackText, Bp: "7", Bt: "7", Colors: []string{"G"}}

var oldB445W1 = effectstest.IwCard{Oracle: "ba57df46-11c9-4f28-8818-175fb52e67bc", Front: "Reckless Waif", FrontType: "Creature — Human Rogue Werewolf", Cost: "{R}", FrontText: effectstest.IwFrontText, Fp: "1", Ft: "1",
	Back: "Merciless Predator", BackText: effectstest.IwBackText, Bp: "3", Bt: "2", Colors: []string{"R"}}

var oldB445W2 = effectstest.IwCard{Oracle: "bbcc2d2d-b29c-410f-bee7-5ef1609a5b32", Front: "Tormented Pariah", FrontType: "Creature — Human Warrior Werewolf", Cost: "{3}{R}", FrontText: effectstest.IwFrontText, Fp: "3", Ft: "2",
	Back: "Rampaging Werewolf", BackText: effectstest.IwBackText, Bp: "6", Bt: "4", Colors: []string{"R"}}

var oldB445W3 = effectstest.IwCard{Oracle: "c05c5375-d0d7-4cf2-92e6-76f86cd0febf", Front: "Gatstaf Shepherd", FrontType: "Creature — Human Werewolf", Cost: "{1}{G}", FrontText: effectstest.IwFrontText, Fp: "2", Ft: "2",
	Back: "Gatstaf Howler", BackText: "Intimidate\n" + effectstest.IwBackText, Bp: "3", Bt: "3", Colors: []string{"G"}}

var oldB445W4 = effectstest.IwCard{Oracle: "c7b1aaff-e2c3-427c-8478-1fd053f7ffb2", Front: "Village Ironsmith", FrontType: "Creature — Human Werewolf", Cost: "{1}{R}", FrontText: "First strike\n" + effectstest.IwFrontText, Fp: "1", Ft: "1",
	Back: "Ironfang", BackText: "First strike\n" + effectstest.IwBackText, Bp: "3", Bt: "1", Colors: []string{"R"}}

func TestOldB445EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB445BatOracle: "Screeching Bat", oldB445BatOracle + "#1": "Stalking Vampire", oldB445OverseerOracle: "Angelic Overseer",
		oldB445W0.Oracle: "Grizzled Outcasts", oldB445W0.Oracle + "#1": "Krallenhorde Wantons",
		oldB445W1.Oracle: "Reckless Waif", oldB445W1.Oracle + "#1": "Merciless Predator",
		oldB445W2.Oracle: "Tormented Pariah", oldB445W2.Oracle + "#1": "Rampaging Werewolf",
		oldB445W3.Oracle: "Gatstaf Shepherd", oldB445W3.Oracle + "#1": "Gatstaf Howler",
		oldB445W4.Oracle: "Village Ironsmith", oldB445W4.Oracle + "#1": "Ironfang",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB445WerewolvesTurnOverBothWays(t *testing.T) {
	for _, w := range []effectstest.IwCard{oldB445W0, oldB445W1, oldB445W2, oldB445W3, oldB445W4} {
		t.Run(w.Front, func(t *testing.T) {
			g, me := effectstest.IwTable(t)
			id := effectstest.IwPut(t, g, w, me)
			effectstest.IwNextUpkeep(t, g, map[int]int{1: 1})
			if got := effectstest.IwFace(g, id); got != 0 {
				t.Fatalf("%s turned over after a turn with a spell, face %d", w.Front, got)
			}
			effectstest.IwNextUpkeep(t, g, map[int]int{})
			if got := effectstest.IwFace(g, id); got != 1 {
				t.Fatalf("%s did not turn into %s after a turn with no spells, face %d", w.Front, w.Back, got)
			}
			effectstest.IwNextUpkeep(t, g, map[int]int{0: 1, 1: 1})
			if got := effectstest.IwFace(g, id); got != 1 {
				t.Fatalf("%s turned back although no player cast two spells, face %d", w.Back, got)
			}
			effectstest.IwNextUpkeep(t, g, map[int]int{2: 2})
			if got := effectstest.IwFace(g, id); got != 0 {
				t.Fatalf("%s did not turn back after a player cast two spells, face %d", w.Back, got)
			}
		})
	}
}

func TestOldB445WerewolfKeywordsFollowTheFace(t *testing.T) {
	g, me := effectstest.IwTable(t)
	howler := effectstest.IwPut(t, g, oldB445W3, me)
	smith := effectstest.IwPut(t, g, oldB445W4, me)
	if effectstest.HasEffectiveKeyword(t, g, howler, "intimidate") {
		t.Error("Gatstaf Shepherd has intimidate")
	}
	if !effectstest.HasEffectiveKeyword(t, g, smith, "first strike") {
		t.Error("Village Ironsmith lacks first strike")
	}
	effectstest.IwTransform(t, g, howler)
	effectstest.IwTransform(t, g, smith)
	if !effectstest.HasEffectiveKeyword(t, g, howler, "intimidate") {
		t.Error("Gatstaf Howler lacks intimidate")
	}
	if !effectstest.HasEffectiveKeyword(t, g, smith, "first strike") {
		t.Error("Ironfang lacks first strike")
	}
}

var oldB445Bat = effectstest.IwCard{Oracle: oldB445BatOracle, Front: "Screeching Bat", FrontType: "Creature — Bat", Cost: "{2}{B}",
	FrontText: "Flying\nAt the beginning of your upkeep, you may pay {2}{B}{B}. If you do, transform this creature.", Fp: "2", Ft: "2",
	Back: "Stalking Vampire", BackText: "At the beginning of your upkeep, you may pay {2}{B}{B}. If you do, transform this creature.", Bp: "5", Bt: "5",
	Colors: []string{"B"}, Keywords: []string{"Flying"}}

func TestOldB445ScreechingBatTransformsOnlyWhenPaidAtYourUpkeep(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB445Bat, me)
	// Declined: nothing happens.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.IwFace(g, id); got != 0 {
		t.Fatalf("Screeching Bat turned over without payment, face %d", got)
	}
	// It does not trigger on another player's upkeep.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the bat asked for payment on an opponent's upkeep")
	}
	// Paid: it transforms.
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.IwFace(g, id); got != 1 {
		t.Fatalf("Screeching Bat did not turn into Stalking Vampire after payment, face %d", got)
	}
}

func TestOldB445AngelicOverseerIsProtectedWhileYouControlAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.PushCatalogPermanent(g, me.ID, "Angelic Overseer", "Creature — Angel", oldB445OverseerOracle, false)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if effectstest.HasEffectiveKeyword(t, g, angel, kw) {
			t.Errorf("Angelic Overseer has %s with no Human", kw)
		}
	}
	effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, angel, "hexproof") {
		t.Error("an opponent's Human turned the protection on")
	}
	effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human", 1, 1)
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, angel, kw) {
			t.Errorf("Angelic Overseer lacks %s with a Human", kw)
		}
	}
}
