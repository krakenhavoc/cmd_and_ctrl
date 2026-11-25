package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dalakos, Crafter of Wonders — Legendary Creature — Merfolk Artificer
// {1}{U}{R}, 2/4 (Theros Beyond Death, #5514):
//
//	"{T}: Add {C}{C}. Spend this mana only to cast artifact spells or
//	 activate abilities of artifacts.
//	 Equipped creatures you control have flying and haste."
//
// Vedalken Engineer's restricted mana ability, producing two colourless
// (CR 106.6), and Firion's grant to each of the controller's equipped
// creatures: flying and haste are layer-6 grants read on every recompute, so
// an Equipment that leaves takes them with it. Summoning sickness applies to
// the {T} (CR 302.6).
//
// No simplification.
func init() {
	equipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && effects.IsEquippedCreature(g, *target)
	}
	effects.Register(effects.Spec{
		OracleID:     "211814a5-195d-41c6-b96c-4c26b2cbb33c",
		Name:         "Dalakos, Crafter of Wonders",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Static: []game.StaticAbility{
			effects.KeywordGrant(equipped, "flying"),
			effects.KeywordGrant(equipped, "haste"),
		},
	})
}
