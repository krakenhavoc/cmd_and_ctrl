package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b135_helpers.go — Theros Beyond Death batch 135 (S82 OLD, #5514).

// devotionBonusSources maps the oracle id of a permanent that prints "Your
// devotion to each color and each combination of colors is increased by N" to N.
// Filled by a card file's init through RegisterDevotionBonus (Altar of the
// Pantheon).
var devotionBonusSources = map[string]int{}

// RegisterDevotionBonus declares that a permanent with this oracle id, while it
// is on the battlefield with its abilities, increases its controller's devotion
// to every colour and every combination of colours by n (CR 700.5).
func RegisterDevotionBonus(oracleID string, n int) {
	devotionBonusSources[oracleID] = n
}

// DevotionBonusFor is the amount the permanents `controller` controls add to
// their devotion to any colour or combination of colours: each source adds its
// number once, however many colours the devotion is to, and counts only while it
// has its abilities (CR 613.1f) and is not face down (CR 708.2a).
func DevotionBonusFor(g *game.Game, controller uuid.UUID) int {
	if len(devotionBonusSources) == 0 {
		return 0
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.FaceDownIsPermanent() || c.HasLostAllAbilities() {
			continue
		}
		n += devotionBonusSources[c.OracleID]
	}
	return n
}

// ControlsAGodDemigodOrLegendaryEnchantment is Altar of the Pantheon's rider
// condition: "if you control a God, a Demigod, or a legendary enchantment". The
// creature types are words of the source's text (CR 612.2), read through it.
func ControlsAGodDemigodOrLegendaryEnchantment(g *game.Game, controller, source uuid.UUID) bool {
	god, demigod := "God", "Demigod"
	if src, ok := g.LookupCardForEffect(source); ok {
		god, demigod = src.TextSubtype("God"), src.TextSubtype("Demigod")
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		if c.HasSubtype(god) || c.HasSubtype(demigod) || (c.IsLegendary() && c.HasCardType("enchantment")) {
			return true
		}
	}
	return false
}
