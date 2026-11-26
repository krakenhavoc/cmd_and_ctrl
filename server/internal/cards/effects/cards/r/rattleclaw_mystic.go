package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rattleclaw Mystic — Khans of Tarkir (S82 OLD b345, #5554):
//
//	Creature — Human Shaman {1}{G}, 2/1: {T}: Add {G}, {U}, or {R}. Morph {2}.
//	When this creature is turned face up, add {G}{U}{R}.

func init() {
	effects.Register(effects.Spec{
		OracleID:         "84c9d8b8-8dd9-41fc-a2b6-af1ffb9090f9",
		Name:             "Rattleclaw Mystic",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{2}")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U|R}",
			Label:    "{T}: Add {G}, {U}, or {R}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp("Rattleclaw Mystic — add {G}{U}{R}", effects.Do(effects.AddMana{Produced: "{G}{U}{R}"})),
		},
	})
}
