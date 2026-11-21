package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// La Forge, Perceptive Engineer — Legendary Creature — Human Artificer
// {1}{U}, 1/4 (Star Trek Commander, #4974):
//
//	"Vigilance
//	 {T}: Add {U}. Spend this mana only to cast an artifact spell."
//
// Castle Doom's restricted mana (cast + Artifact tag, CR 106.6): it pays
// for an artifact spell and for nothing else, an ability's cost included.
// The tap is a creature's {T}, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "338619f2-bb33-4988-bafe-7460f86acef8",
		Name:            "La Forge, Perceptive Engineer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{U}",
			Label:        "{T}: Add {U}. Spend this mana only to cast an artifact spell.",
			Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
		}},
	})
}
