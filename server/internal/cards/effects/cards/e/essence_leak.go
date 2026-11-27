package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Leak — Enchantment — Aura {U} (Invasion, S82 older sets batch 661, #5616):
//
//	"Enchant permanent
//	 As long as enchanted permanent is red or green, it has "At the beginning
//	 of your upkeep, sacrifice this permanent unless you pay its mana cost.""
//
// A layer 6 grant of a catalog bundle (ADR 0093), Kataki's shape, to the
// enchanted permanent while it is red or green; the grant follows the
// attachment and the color is re-read every pass. The permanent's controller
// is the "you" of the upkeep trigger. The cost is the permanent's own mana
// cost, read as the trigger resolves; a permanent with no mana cost can't pay
// and is sacrificed (CR 202.1a).
//
// No simplification.
func init() {
	const key = "essence-leak/upkeep-sacrifice"
	effects.Register(effects.Spec{
		OracleID:     "cb02daf6-9013-4ccc-bf1f-773b757ac468",
		Name:         "Essence Leak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key:       key,
			Triggered: []game.TriggeredAbility{effects.OldB661UpkeepSacrificeUnlessYouPayItsManaCost()},
			Text:      "At the beginning of your upkeep, sacrifice this permanent unless you pay its mana cost.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !effects.AttachedToSource(target, g, source) {
				return false
			}
			for _, col := range target.EffectiveColors() {
				if col == "R" || col == "G" {
					return true
				}
			}
			return false
		}, key)},
	})
}
