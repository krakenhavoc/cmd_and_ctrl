package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:      "8d4e0866-d8f5-4eb6-a0fd-3fa9d4b9cf4a",
		Name:          "Waste Management",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{3}{B}"), effects.TargetPlayer("target player"))},
		Targets:       effects.UpToNCardsFromASingleGraveyard(2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !ctx.WasKicked() {
				return effects.ExileTargetCardsThen(ctx, wasteManagementRoguesTail)
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
			return effects.ExileCardsThen(ctx, ids, wasteManagementRoguesTail)
		},
	})
}

// wasteManagementRogues is "create a 2/2 black Rogue creature token for
// each creature card exiled this way".
func wasteManagementRogues(ctx *effects.Context, exiled []uuid.UUID, wasCreature map[uuid.UUID]bool) error {
	n := effects.CreatureCardsAmong(exiled, wasCreature)
	if n == 0 {
		return nil
	}
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 black Rogue"), N: n}.Apply(ctx)
}

// wasteManagementRoguesTail is wasteManagementRogues after the exile
// (#4253 slice 9).
var wasteManagementRoguesTail = effects.GraveyardExileTail("waste-management/rogue-per-creature", wasteManagementRogues)
