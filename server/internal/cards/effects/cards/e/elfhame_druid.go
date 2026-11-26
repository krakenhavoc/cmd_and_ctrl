package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elfhame Druid — Creature — Elf Druid {1}{G}, 0/2 (Dominaria, S82 OLD,
// #5530):
//
//	"{T}: Add {G}.
//	 {T}: Add {G}{G}. Spend this mana only to cast kicked spells."
//
// Two mana abilities (CR 605.1a). The second adds two {G} that may only pay for
// casting a spell whose kicker was paid (ManaRestrictKickedSpell, CR 702.33a, as
// Throne of Makindi's). Summoning sickness gates both taps (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "519fccdb-fde4-42cb-b2ca-35fc36cf984a",
		Name:         "Elfhame Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "{T}: Add {G}",
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{G}{G}",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictKickedSpell},
				Label:        "{T}: Add {G}{G}. Spend this mana only to cast kicked spells",
			},
		},
	})
}
