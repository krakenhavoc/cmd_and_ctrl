package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const (
	oldB551Balloonist = "4ca5d2c7-621b-4d5e-9bf4-ff3ae271ba44"
	oldB551Oakgnarl   = "5cafdf77-5500-4471-a848-89863eea2a35"
	oldB551Cloudcrown = "63c52071-e3a7-4236-9dcb-e5ea68240dc2"
	oldB551WarrenElf  = "64368768-77b9-4f78-adb7-951ca71dfb28"
	oldB551Hillcomber = "a56b3434-7478-4704-b04b-92f32a1bc462"
	oldB551Plover     = "c4a99095-38cc-4fd3-b3da-c72a6ccceb25"
	oldB551Moonglove  = "d4870995-f434-4985-a921-1dc9556cee41"
	oldB551Buffoon    = "da404c9c-3fb0-418c-947f-f42d6aff6899"
)

func TestOldB551CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB551Balloonist: "Kinsbaile Balloonist", oldB551Oakgnarl: "Oakgnarl Warrior", oldB551Cloudcrown: "Cloudcrown Oak",
		oldB551WarrenElf: "Warren-Scourge Elf", oldB551Hillcomber: "Hillcomber Giant", oldB551Plover: "Plover Knights",
		oldB551Moonglove: "Moonglove Winnower", oldB551Buffoon: "Nath's Buffoon",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestOldB551KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Oakgnarl Warrior", "Creature — Treefolk Warrior", oldB551Oakgnarl, []string{"vigilance", "trample"}},
		{"Cloudcrown Oak", "Creature — Treefolk Warrior", oldB551Cloudcrown, []string{"reach"}},
		{"Warren-Scourge Elf", "Creature — Elf Warrior", oldB551WarrenElf, []string{"protection from Goblins"}},
		{"Hillcomber Giant", "Creature — Giant Scout", oldB551Hillcomber, []string{"mountainwalk"}},
		{"Plover Knights", "Creature — Kithkin Knight", oldB551Plover, []string{"flying", "first strike"}},
		{"Moonglove Winnower", "Creature — Elf Rogue", oldB551Moonglove, []string{"deathtouch"}},
		{"Nath's Buffoon", "Creature — Goblin Rogue", oldB551Buffoon, []string{"protection from Elves"}},
		{"Kinsbaile Balloonist", "Creature — Kithkin Soldier", oldB551Balloonist, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		effectstest.AssertKeywords(t, g, id, tc.keywords...)
	}
}

// Protection from Elves keeps an Elf's damage off the Buffoon and an Elf from
// blocking it, and says nothing about other creatures.
func TestOldB551ProtectionFromAnElfTypeStopsOnlyThatType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	buffoon := effectstest.B12Push(g, me.ID, "Nath's Buffoon", "Creature — Goblin Rogue", oldB551Buffoon, 1, 1)
	elf := effectstest.B12Push(g, opp.ID, "Elf", "Creature — Elf Warrior", "", 2, 2)
	bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DealDamageTo(t, g, elf, buffoon, 2)
	if !g.Battlefield.Contains(buffoon) {
		t.Error("an Elf's damage destroyed the Buffoon")
	}
	effectstest.DealDamageTo(t, g, bear, buffoon, 2)
	if g.Battlefield.Contains(buffoon) {
		t.Error("a Bear's damage did not destroy the Buffoon")
	}
}

func TestOldB551KinsbaileBalloonistGivesAnyCreatureFlyingWhenItAttacks(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		balloonist := effectstest.PushDiesCreatureForTest(g, me.ID, "Kinsbaile Balloonist", oldB551Balloonist, "Creature — Kithkin Soldier", 2, 2)
		bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, balloonist)
		prompt := effectstest.LatestPickTarget(g, me.ID)
		if prompt == nil {
			t.Fatal("no target prompt for the attack trigger")
		}
		if !effectstest.HasID(prompt.PickTargetCards, bear) || !effectstest.HasID(prompt.PickTargetCards, balloonist) {
			t.Error("any creature should be a legal target, the Balloonist and a non-attacker included")
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasKeywordOf(g, bear, "flying"); got != yes {
			t.Errorf("accept %v: the Bear has flying = %v", yes, got)
		}
	}
}
