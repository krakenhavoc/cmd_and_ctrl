package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zaxara, the Exemplary — Legendary Creature — Nightmare Hydra
// {1}{B}{G}{U}, 2/3:
//
//	"Deathtouch
//	 {T}: Add two mana of any one color.
//	 Whenever you cast a spell with {X} in its mana cost, create a 0/0
//	 green Hydra creature token, then put X +1/+1 counters on it."
//
// The mana ability is Alchemist's Talent's two-of-one-colour add. "{X} in
// its mana cost" is the spell's own printed cost (Geometer's Arthropod's
// matcher); X is the value announced for that spell, read as the trigger
// resolves. With X of 0 the Hydra stays a 0/0 and dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a9946ea0-bf0e-4c9c-a3c6-21337a5029ab",
		Name:            "Zaxara, the Exemplary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "Add two mana of any one color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.SocB26YouCastASpellWithXInItsManaCost,
				"Zaxara, the Exemplary — create a 0/0 green Hydra creature token, then put X +1/+1 counters on it",
				effects.Do(effects.RprB19HydraWithXCountersSteps()...)),
		},
	})
}
