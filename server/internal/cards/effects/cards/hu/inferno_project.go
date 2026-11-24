package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inferno Project — Creature — Elemental {6}{R}, 0/0 (Commander 2021,
// #5504):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it, where X is the
//	 total mana value of instant and sorcery cards in your graveyard."
//
// The counters are a CR 614.1c self-replacement counted as it enters (Bone
// Devourer's shape) over the instant and sorcery cards in its controller's
// graveyard, each at its printed mana value (an X is 0, CR 107.3m, a card in
// a graveyard having no announced X). With none it is a 0/0 and dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "36d641d4-d3f5-4394-81d7-3209955a1628",
		Name:            "Inferno Project",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				p := g.PlayerByIDForEffect(src.Controller)
				if p == nil || p.Graveyard == nil {
					return 0
				}
				total := 0
				for _, c := range p.Graveyard.Cards {
					if c.IsInstant() || c.IsSorcery() {
						total += c.ManaValue()
					}
				}
				return total
			}, "Inferno Project enters with X +1/+1 counters on it, X the total mana value of instant and sorcery cards in your graveyard"),
		},
	})
}
