package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Masked Bandits — Creature — Raccoon Rogue {3}{B}{R}{G}, 5/5 (Streets of
// New Capenna):
//
//	"Vigilance
//	 Menace
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {B},
//	 {R}, or {G}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The hand ability is the cycle's shared one (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := effects.SncB19ExileForLand("Masked Bandits", "masked-bandits/tap-brg", "{B|R|G}", "{B}, {R}, or {G}")
	effects.Register(effects.Spec{
		OracleID:        "84ac3096-098a-477a-8907-fc30ac15f00f",
		Name:            "Masked Bandits",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Grants:          []effects.AbilityGrant{grant},
		Activated:       []effects.ActivatedAbility{ability},
	})
}
