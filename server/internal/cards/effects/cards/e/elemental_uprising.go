package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elemental Uprising — Instant {1}{G} (Jumpstart, #5510):
//
//	"Target land you control becomes a 4/4 Elemental creature with haste until
//	 end of turn. It's still a land. It must be blocked this turn if able."
//
// Avalanche Caller's animation (the land keeps its types and abilities and
// gains Creature and Elemental) with King Harald's Revenge's "must be
// blocked this turn if able".
//
// No simplification.
func init() {
	const label = "Elemental Uprising — the land becomes a 4/4 Elemental creature with haste until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "d9d3f6fc-5087-451a-94b1-ffe1a5d30f5e",
		Name:         "Elemental Uprising",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerAnimate},
		Targets:      effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			land := effects.FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			if err := (effects.BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, Subtypes: []string{"Elemental"}, SetPower: 4, SetToughness: 4, Label: label}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: land, Keywords: []string{"haste"}, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return effects.BlockRequirementUntilEOT{Target: land, Kind: game.BlockRequirementMustBeBlocked,
				Label: "Elemental Uprising — must be blocked this turn if able"}.Apply(ctx)
		},
	})
}
