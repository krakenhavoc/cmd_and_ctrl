package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avatar of Woe — Creature — Avatar {6}{B}{B}, 6/5 (Conspiracy: Take the
// Crown, S82 older sets batch 273, #5540):
//
//	"If there are ten or more creature cards total in all graveyards, this
//	 spell costs {6} less to cast.
//	 Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 {T}: Destroy target creature. It can't be regenerated."
//
// A conditional self cost reduction (CR 601.2f) read as the spell is cast:
// every creature card in every graveyard (Nighthowler's count) is ten or
// more, and the reduction spends against the generic {6} only, leaving
// {B}{B}. Fear is the engine's keyword (CR 702.36); the tap ability is
// Tsabo Tavoc's, with no restriction on the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2da5aa33-3794-4f83-936c-9d8640833d86",
		Name:            "Avatar of Woe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(6, "This spell costs {6} less to cast if there are ten or more creature cards total in all graveyards.",
				func(q game.CostQuery) bool { return effects.B41CreatureCardsInAllGraveyards(q.Game) >= 10 }),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature. It can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
			},
		}},
	})
}
