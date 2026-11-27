package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wings of Velis Vel — Kindred Instant — Shapeshifter {1}{U} (Lorwyn, S82
// OLD b548, #5593):
//
//	"Changeling (This card is every creature type.)
//	 Until end of turn, target creature has base power and toughness 4/4,
//	 gains all creature types, and gains flying."
//
// Three until-end-of-turn effects on the target, locked at resolution (CR
// 611.2c) and ending in the cleanup step (CR 514.2): the base 4/4 (layer 7b,
// Eidolon of Astral Winds' pump), the every-creature-type fact (layer 4, as
// Shields of Velis Vel gives it) and flying (layer 6). Counters and other
// pumps still apply on top of the base. A target that is illegal by resolution
// stops the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0bb2b3cc-d3e6-40e9-a5da-d589e8e8f0bd",
		Name:            "Wings of Velis Vel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Targets:         effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			const label = "Wings of Velis Vel — base power and toughness 4/4, all creature types and flying until end of turn"
			if err := effects.UntilEndOfTurn(ctx, id, nil, label, game.SetBasePTMods(4, 4)...); err != nil {
				return err
			}
			if err := (effects.GrantAllCreatureTypesUntilEOT{Target: id, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label}.Apply(ctx)
		},
	})
}
