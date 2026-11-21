package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// SeleniasCurseToken is Selenia, the Cursed Heart's token (S82 GAP batch
// 3, #5948): a legendary black Aura Curse enchantment named Selenia's
// Curse with enchant player and "If enchanted player would lose life,
// they lose twice that much life instead."
func SeleniasCurseToken() game.Card { return tokenFromCatalog(printedSeleniasCurseToken) }

func printedSeleniasCurseToken() tokenTemplate {
	return tokenTemplate{
		Slug: "selenias-curse",
		Card: game.Card{
			Name:     "Selenia's Curse",
			TypeLine: "Token Legendary Enchantment — Aura Curse",
			Colors:   []string{"B"},
		},
		Replacements: []game.ReplacementEffect{
			enchantedPlayerLosesTwiceThatMuchLife("Selenia's Curse: enchanted player loses twice that much life"),
		},
		Text: "Enchant player\nIf enchanted player would lose life, they lose twice that much life instead.",
	}
}
