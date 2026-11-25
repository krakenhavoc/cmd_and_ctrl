package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nalathni Dragon — Creature — Dragon {2}{R}{R}, 1/1 (S82 GAP2 batch 13,
// #5948, #6101):
//
//	"Flying; banding
//	 {R}: This creature gets +1/+0 until end of turn. If this ability has
//	 been activated four or more times this turn, sacrifice this creature
//	 at the beginning of the next end step."
//
// Banding is the engine's (game/banding.go, ADR 0224). The pump is Dragon
// Whelp's, rider and all (SacrificeAtNextEndStepIfActivatedRider): the
// count is the per-object, per-ability, per-turn activation record, read as
// the ability resolves, so the fourth activation (and any later one)
// schedules the sacrifice, bound to this object.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +1/+0 until end of turn. If this ability has been activated four or more times this turn, sacrifice this creature at the beginning of the next end step."
	rider := effects.SacrificeAtNextEndStepIfActivatedRider(4, label, "Nalathni Dragon — sacrifice this creature")
	effects.Register(effects.Spec{
		OracleID:        "0887dc95-3e75-4f72-bf05-4067d3785376",
		Name:            "Nalathni Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordBanding},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Nalathni Dragon — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return rider(g, item.Controller, ctx.Source())
			},
		}},
	})
}
