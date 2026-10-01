package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gyre Sage — Creature — Elf Druid {1}{G}, 1/2:
//
//	"Evolve (Whenever a creature you control enters, if that creature
//	 has greater power or toughness than this creature, put a +1/+1
//	 counter on this creature.)
//	 {T}: Add {G} for each +1/+1 counter on this creature."
//
// The count is read as the mana ability resolves, which is the instant
// it is activated (CR 605.3b), so it is the counters the Sage has right
// then. With none it taps for nothing, which is the printed card.
// Evolve is the engine's keyword trigger (game/evolve.go, #1805).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3aa16cf-079b-4737-9ffd-7bfdffef0cb2",
		Name:            "Gyre Sage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordEvolve},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: gyreSageProduced,
			Label:        "Add {G} for each +1/+1 counter on this creature",
		}},
	})
}

// gyreSageProduced is one {G} per +1/+1 counter on the ability's own
// source.
func gyreSageProduced(g *game.Game, _, source uuid.UUID) string {
	c, ok := g.LookupCardForEffect(source)
	if !ok {
		return ""
	}
	return strings.Repeat("{G}", c.Counters[game.CounterPlusOne])
}
