package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrewd Negotiation — Sorcery {4}{U} (Kaladesh, #5539):
//
//	"Exchange control of target artifact you control and target artifact or
//	 creature you don't control."
//
// Trade the Helm's exchange over two clauses: the first must be an artifact of
// yours, the second an artifact or a creature of another player's. The exchange
// is all or nothing (CR 701.12a): if either target is illegal at resolution,
// nothing changes hands. It has no duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d36fea45-d8aa-469a-bcfc-5c0e4a0e807b",
		Name:         "Shrewd Negotiation",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
			effects.TargetPermanent("target artifact or creature you don't control", effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, okMine := ctx.ClauseTarget(0)
			theirs, okTheirs := ctx.ClauseTarget(1)
			if !okMine || !okTheirs {
				return nil
			}
			return effects.ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Shrewd Negotiation — exchange control"}.Apply(ctx)
		},
	})
}
