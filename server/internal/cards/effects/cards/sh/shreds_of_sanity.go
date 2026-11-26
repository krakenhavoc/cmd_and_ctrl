package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shreds of Sanity — Sorcery {2}{R} (Eldritch Moon, S82 OLD, #5541):
//
//	"Return up to one target instant card and up to one target sorcery card
//	 from your graveyard to your hand, then discard a card. Exile Shreds of
//	 Sanity."
//
// Two "up to one" clauses (CR 601.2c) from your graveyard, returned in
// announce order, then the controller chooses a card to discard (so one just
// returned may be it), then the spell exiles itself (Experimental Overload's
// last step), never reaching the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6eaa343d-0097-4f05-80ab-deb016d4e15e",
		Name:         "Shreds of Sanity",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Discards: 1},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("up to one target instant card from your graveyard", effects.YouOwn(), effects.Instant()).WithCount(0, 1),
			effects.TargetCardInGraveyard("up to one target sorcery card from your graveyard", effects.YouOwn(), effects.Sorcery()).WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error { return effects.DiscardOne(ctx.Game, ctx.Item, ctx.Controller()) }),
			effects.Instruction(func(ctx *effects.Context) error { return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID) }),
		),
	})
}
