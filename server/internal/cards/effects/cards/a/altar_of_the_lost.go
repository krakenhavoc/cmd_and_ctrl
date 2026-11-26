package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Altar of the Lost — Artifact {3}:
//
//	"This artifact enters tapped.
//	 {T}: Add two mana in any combination of colors. Spend this mana only to
//	 cast spells with flashback from a graveyard."
//
// The restriction is two tags that must both hold (game.ManaRestrictSpellFlashback:
// the spell has flashback; game.ManaRestrictCastFromGraveyard: it is cast from
// a graveyard). It never pays an activation or a spell cast from a hand.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "171a3ac4-0c9c-46ac-b273-089d566cca05",
		Name:         "Altar of the Lost",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     effects.AnyCombinationOfColors(2),
			Label:        "Add two mana in any combination of colors. Spend this mana only to cast spells with flashback from a graveyard",
			Restrictions: []string{game.ManaRestrictSpellFlashback, game.ManaRestrictCastFromGraveyard},
		}},
	})
}
