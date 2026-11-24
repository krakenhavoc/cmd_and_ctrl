package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lullmage's Familiar — Creature — Beast {1}{G}{U}, 2/4 (Zendikar Rising,
// #5508):
//
//	"{T}: Add {G} or {U}.
//	 Whenever you cast a kicked spell, you gain 2 life."
//
// A mana ability for one of two colours (CR 605.1a, summoning sickness
// gates the tap) and a cast trigger over a kicked spell
// (S82OldB097YouCastAKickedSpell, CR 702.33a). The life is gained on the
// trigger's resolution, whatever happens to the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a56fe0b-f8a9-499c-bce1-cec6cb3b66f0",
		Name:         "Lullmage's Familiar",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB097YouCastAKickedSpell,
				"Lullmage's Familiar — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
