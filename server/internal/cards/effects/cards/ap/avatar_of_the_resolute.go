package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of the Resolute — Creature — Avatar {G}{G}, 3/2 (Dragons of Tarkir):
//
//	"Reach, trample
//	 This creature enters with a +1/+1 counter on it for each other creature you
//	 control with a +1/+1 counter on it."
//
// The count is taken as it enters, as a self-replacement (CR 614.1c), over the
// other creatures its controller controls that carry a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "15fd66db-f9dd-40ab-92bc-9e0575bd7489",
		Name:            "Avatar of the Resolute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, avatarOfTheResoluteCount,
				"Avatar of the Resolute enters with a +1/+1 counter on it for each other creature you control with a +1/+1 counter on it"),
		},
	})
}

func avatarOfTheResoluteCount(g *game.Game, src *game.Card) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == src.Controller && c.InstanceID != src.InstanceID && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
			n++
		}
	}
	return n
}
