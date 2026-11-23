package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// madness_cards_test.go — the three cards #657 ships: the two madness
// cards and the enabler that discards as a COST.
//
// The keyword's own contract is pinned in game/madness_test.go. What
// is proved here is that a card file declaring one string gets both
// halves of it, and that the two discard roads a real game takes —
// an effect's instruction (Faithless Looting) and an activation cost
// (Olivia's Dragoon) — both end in a castable card in exile.

const (
	FieryTemperOracle     = "f07bd49d-8e71-4d56-be2a-638514011318"
	BigGameHunterOracle   = "ab55834f-c935-4773-89c6-bec9712284eb"
	OliviasDragoonOracle  = "a5019399-91a8-4233-b16f-399718c4be9c"
	FaithlessLootingOracl = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
)

// madnessAbilitiesFor counts the two halves the keyword must grow
// from a card's `Madness` string: the discard replacement and the
// exile trigger that watches a discard.
func MadnessAbilitiesFor(oracle string) (replacements, triggers int) {
	for _, r := range game.CatalogReplacements(oracle) {
		for _, kind := range r.Watches {
			if kind == game.EventDiscardCard {
				replacements++
			}
		}
	}
	for _, tr := range game.CatalogTriggers(oracle) {
		if !game.TriggerWatchesFromZone(tr, game.ZoneExile) {
			continue
		}
		for _, kind := range tr.Watches {
			if kind == game.EventDiscardCard {
				triggers++
			}
		}
	}
	return replacements, triggers
}
