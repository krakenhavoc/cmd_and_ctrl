package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Aberration — Creature — Elf Mutant {5}{G}, 4/5 (Scourge):
//
//	"{T}: Add {G}{G}{G}.
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Canopy Tactician's {T}: Add {G}{G}{G} mana ability (CR 605.1a), which a
// creature can't use the turn it arrives unless it has haste, and the
// shared typecycling ability (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d25b831-c1de-4f79-bbb3-1582eda96452",
		Name:         "Elvish Aberration",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
