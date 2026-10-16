package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evendo, Waking Haven — Land — Planet (EDHREC rank 906):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Planet. Station only as a sorcery.)
//	 12+ | {G}, {T}: Add {G} for each creature you control."
//
// Adagia's shape (a land with station, CR 721.1) with The Eternity
// Elevator's gated mana ability: the 12+ line is a mana ability whose
// Condition reads the charge counters, because a ManaAbility carries
// no designation gate. Below twelve counters it is listed and greyed
// rather than absent, which changes nothing a player can do with it.
// The amount is Gaea's Cradle's, counted as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83161d59-2520-4741-9328-e2a4a8b5d5bc",
		Name:         "Evendo, Waking Haven",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:           ManaAbilityCost{Tap: true, Mana: "{G}"},
				Condition:      evendoAt12,
				ProducedAmount: ProducedPerPermanent("G", MatchCreature),
				Label:          "12+ — {G}, {T}: Add {G} for each creature you control",
			},
		},
		Activated: []ActivatedAbility{Station()},
	})
}

// evendoAt12 is the "12+" threshold.
func evendoAt12(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.Counters[game.CounterCharge] >= 12
}
