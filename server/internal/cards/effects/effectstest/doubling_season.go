package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// doubling_season_test.go — the S17 sprint exit-criterion test.
//
// > Doubling Season + Hardened Scales in play → add a +1/+1 counter
// > → prompt for order → 4 counters land per chosen order.
//
// Both permutations exercised explicitly:
//   - [HS, DS] → (1 + 1) * 2 = 4
//   - [DS, HS] → (1 * 2) + 1 = 3
//
// Coverage also asserts the single-replacement fast path (Doubling
// Season alone → 2 counters, no prompt) and the predicate gates
// (opponent's creature, foreign counter name).

const (
	DoublingSeasonOracle     = "01546b7d-a233-4176-8843-d732074dc5b6"
	HardenedScalesOracle     = "a1f3da21-af6d-450e-bf0b-985d158418e6"
	BranchingEvolutionOracle = "28fe909b-06e0-424c-9f75-c824a25f5865"
)

// seedCreature pushes a 2/2 creature onto the battlefield via the
// timestamp helper (fires EventZoneMove so the S16 listener bumps
// layerVersion + stamps EnteredBattlefieldAt).
func SeedCreature(g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      controller,
		Controller: controller,
	})
}

// seedReplacementPermanent pushes a catalog card onto the
// battlefield so its Replacements become active. OracleID gates
// the CatalogReplacements lookup.
func SeedReplacementPermanent(g *game.Game, oracleID, name string, controller uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Enchantment",
		OracleID:   oracleID,
		Owner:      controller,
		Controller: controller,
	})
}

// countersOn returns the +1/+1 count on the given card. Zero when
// the counter map is nil or absent.
func CountersOn(g *game.Game, cardID uuid.UUID, name string) int {
	var n int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == cardID {
				n = c.Counters[name]
				return
			}
		}
	})
	return n
}

// replacementIDsForSources maps a prompt's ReplacementEffectIDs
// back to (hsID, dsID) by reading each ID's metadata and matching
// on source card. Avoids hard-coding the engine's ID-minting
// scheme in the test.
func ReplacementIDsForSources(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, hsCardID, dsCardID uuid.UUID) (hsEff, dsEff game.ReplacementEffectID) {
	t.Helper()
	return ReplacementIDForSource(t, g, ids, hsCardID), ReplacementIDForSource(t, g, ids, dsCardID)
}

// replacementIDForSource is the one-card half of the above: the
// prompt entry contributed by that permanent. Fatal when the prompt
// has none, which is always a test bug.
func ReplacementIDForSource(t *testing.T, g *game.Game, ids []game.ReplacementEffectID, cardID uuid.UUID) game.ReplacementEffectID {
	t.Helper()
	for _, id := range ids {
		if _, srcCardID := g.ReplacementOptionMetaForEffect(id); srcCardID == cardID {
			return id
		}
	}
	t.Fatalf("no prompt entry for source card %s", cardID)
	return 0
}
