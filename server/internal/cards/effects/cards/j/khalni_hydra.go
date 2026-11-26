package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Khalni Hydra — {G}{G}{G}{G}{G}{G}{G}{G} Creature — Hydra 8/8 (Rise of the Eldrazi, S82 OLD #5581):
//
//	This spell costs {G} less to cast for each green creature you control.
//	Trample
//
// The reduction is a self cost modifier whose unit is {G} (CR 118.7b, 118.7c: it eats green first, then generic mana), counted over the green creatures the caster controls as the spell is priced; the spell itself is on the stack and never counted.

func init() {
	reduction := effects.CostsLessEach(khalniHydraGreenCreaturesYouControl,
		"This spell costs {G} less to cast for each green creature you control.")
	unit, err := game.ParseCost("{G}")
	if err != nil {
		panic(err)
	}
	reduction.Unit = &unit
	effects.Register(effects.Spec{
		OracleID:          "d4649dd8-9a30-4f7e-8b9d-8edee8b2dbd7",
		Name:              "Khalni Hydra",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"trample"},
		SelfCostModifiers: []game.CostModifier{reduction},
	})
}

// khalniHydraGreenCreaturesYouControl is "for each green creature you control"
// as a cost count (CR 601.2f), read through the layers.
func khalniHydraGreenCreaturesYouControl(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.BattlefieldCardsForEffect() {
		if c.Controller == q.Controller && c.IsCreature() && c.HasColor("G") {
			n++
		}
	}
	return n
}
