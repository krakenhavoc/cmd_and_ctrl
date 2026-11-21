package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch26_test.go — card-level coverage for the card-coverage
// roadmap's batch 26 (#388, `edhrec_rank` 2738–2838): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b26-prefixed.

const (
	B26TraumatizeOracle            = "e2ea7d01-6564-4a7f-b935-49a5e3978dac"
	B26DuskLegionDuelistOracle     = "30367a59-a812-41d4-a451-80d9071dd4a4"
	B26AerialExtortionistOracle    = "7e7d7f4d-2c22-4803-ba5e-5eb653c25b5e"
	B26BiteDownOracle              = "623903de-3c04-4745-9af3-d7ec9fb2574d"
	B26TheEarthKingOracle          = "d5770b0f-4493-42e9-a2d7-0e74a28da4ba"
	B26AetherSpellbombOracle       = "4b033a0a-c1ae-44d7-9662-72cbbfda024b"
	B26NightshadeDryadOracle       = "f8263274-8cb4-4eab-b541-c7655b555830"
	B26PeerlessRecyclingOracle     = "938c03fc-8adf-4c7a-8ae1-eca8401f7a83"
	B26ExpelTheInterlopersOracle   = "744b7e93-c893-4770-a3bb-6a45b7f8e4f7"
	B26ChampionOfThePerishedOracle = "f3654fbd-16a5-4953-84ac-534e8421032f"
	B26AetherspoutsOracle          = "48369aec-a991-4bef-8554-01c84302b063"
	B26CybermanPatrolOracle        = "1e51fab7-3ca5-4fbb-a1e9-b39c842895e8"
	B26HeapedHarvestOracle         = "bbfc5011-a9b7-442d-a443-974a5a64de46"
	B26DwarvenMineOracle           = "74ed0bd3-ac31-41a4-8220-d8e7c8c1c437"
	B26ThrashingBrontodonOracle    = "60bc63dc-ac9f-4a2f-aef5-c90d0aa31553"
	B26ElvenAmbushOracle           = "a2373025-20c1-4416-91f7-e51f68dbd146"
	B26UlvenwaldHydraOracle        = "40b85f70-78b6-427a-9fb9-c3a72b8528ed"
	B26CarmenTestOracle            = "84c24e62-26bd-4ca4-b6b9-d31e2065f13b"
	B26SylvanAnthemOracle          = "5ab5eefd-47bf-4cdf-bc29-b517e4f6adc0"
	B26RaiseThePastOracle          = "a69a24d0-ca58-4a44-8af1-a3bd1608d2f9"
	B26FellOracle                  = "8f968470-ce00-43a3-95f5-2e3fe987be19"
	B26ZurTheEnchanterOracle       = "d7950018-d744-48a8-81aa-0d8384703f48"
	B26PriceOfProgressOracle       = "e9da499c-fa43-4e94-8395-5c030ff39502"
	B26GraypeltRefugeOracle        = "60b36821-0fad-423c-98c4-f64d991719f3"
	B26GenesisChamberOracle        = "150ab025-5cc3-4468-a724-bbba8838445d"
	B26ChromaticStarOracle         = "fedbd40b-e3a5-449c-a8a3-b42e9da191a9"
	B26InspiringLeaderOracle       = "d46dcc75-55ff-4226-a392-a49755d269d2"
	B26ValkyrieHarbingerOracle     = "18a5e71e-eb75-4ac7-bedd-2220aaa24f78"
)

// b26DeclinePickTarget answers an open "up to one" pick_target prompt
// with nothing.
func B26DeclinePickTarget(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	p := LatestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTargets(p.ID, chooser, nil); err != nil {
		t.Fatalf("ResolvePickTargets (decline): %v", err)
	}
}

// b26AnswerScryAllTop answers a player's open scry prompt by keeping
// every card on top in the offered order.
func B26AnswerScryAllTop(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := ScryChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no scry prompt for %s", chooser)
	}
	if err := g.ResolveScry(c.ID, chooser, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
}

// b26AnswerDamageAssignments answers every open CR 510.1c assignment
// prompt by putting the attacker's whole power on its first declared
// blocker, and returns how many it answered. Since #730 an unanswered
// prompt gates the table, so a test that double-blocks and then passes
// priority has to drain these.
func B26AnswerDamageAssignments(t *testing.T, g *game.Game) int {
	t.Helper()
	n := 0
	for i := 0; i < 16; i++ {
		var c *game.PendingChoice
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceDamageAssignment && p.DamageAssignment != nil {
				c = p
				break
			}
		}
		if c == nil {
			return n
		}
		entries := make([]game.DamageAssignmentEntry, 0, len(c.DamageAssignment.BlockerIDs))
		for j, b := range c.DamageAssignment.BlockerIDs {
			amount := 0
			if j == 0 {
				amount = c.DamageAssignment.AttackerPower
			}
			entries = append(entries, game.DamageAssignmentEntry{BlockerID: b, Amount: amount})
		}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, entries, 0); err != nil {
			t.Fatalf("ResolveDamageAssignment: %v", err)
		}
		n++
	}
	t.Fatal("damage-assignment prompts did not drain")
	return n
}
