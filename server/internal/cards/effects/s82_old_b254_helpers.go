package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b254_helpers.go — shared vocabulary for S82 OLD batch 254 (Aether
// Revolt, #5536).

// S82OldB254EtheriumCellToken is "a colorless artifact token named Etherium
// Cell with '{T}, Sacrifice this token: Add one mana of any color.'"
// (Tezzeret the Schemer). The slug is an on-disk identity: never renamed,
// never reused.
func S82OldB254EtheriumCellToken() game.Card {
	return TokenFromCatalog(printedS82OldB254EtheriumCellToken)
}

func printedS82OldB254EtheriumCellToken() TokenTemplate {
	return TokenTemplate{
		Slug: "etherium-cell",
		Card: game.Card{
			Name:     "Etherium Cell",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      "{W|U|B|R|G}",
			Label:         "{T}, Sacrifice this token: Add one mana of any color",
		}},
		Text: "{T}, Sacrifice this token: Add one mana of any color.",
	}
}

// DealtCombatDamageBySource narrows a player clause to "target player who
// was dealt combat damage by this [source] this turn" (Hope of Ghirapur,
// game.TargetSpec.DealtCombatDamageBySource).
func DealtCombatDamageBySource(spec *game.TargetSpec) *game.TargetSpec {
	spec.DealtCombatDamageBySource = true
	return spec
}
