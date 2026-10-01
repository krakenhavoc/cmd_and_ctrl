package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waste Management — Instant {2}{B}:
//
//	"Kicker {3}{B} (You may pay an additional {3}{B} as you cast this
//	 spell.)
//	 Exile up to two target cards from a single graveyard. If this
//	 spell was kicked, instead exile target player's graveyard. Create
//	 a 2/2 black Rogue creature token for each creature card exiled
//	 this way."
//
// #1807, ADR 0106 §5. "Instead" is a different TARGET clause, chosen
// when the spell is cast (CR 601.2c: "a spell may require alternative
// targets only if an alternative or additional cost was chosen for
// it"), so the kicker rewrites the clause through WhenPaid, as
// Bloodchief's Thirst's does: unkicked, up to two cards from one
// graveyard; kicked, one target player.
//
// The Rogues count the creature cards that actually reached exile, on
// either branch (the 2022-04-29 ruling), read before the move.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8d4e0866-d8f5-4eb6-a0fd-3fa9d4b9cf4a",
		Name:          "Waste Management",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{3}{B}"), TargetPlayer("target player"))},
		Targets:       upToNCardsFromASingleGraveyard(2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if !ctx.WasKicked() {
				return exileTargetCardsThen(ctx, wasteManagementRogues)
			}
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			p := ctx.PlayerByID(t.ID)
			if p == nil || p.Graveyard == nil {
				return nil
			}
			ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
			return exileCardsThen(ctx, ids, wasteManagementRogues)
		},
	})
}

// wasteManagementRogues is "create a 2/2 black Rogue creature token for
// each creature card exiled this way".
func wasteManagementRogues(ctx *Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
	n := creatureCardsAmong(exiled, wasCreature)
	if n == 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Rogue"), N: n}.Apply(ctx)
}
