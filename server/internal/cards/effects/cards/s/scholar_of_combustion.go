package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scholar of Combustion — Creature — Human Wizard {3}{R}, 3/2 (Foundations
// Jumpstart, S82 BNS):
//
//	"When this creature enters, exile up to one target instant or sorcery
//	 card from your graveyard. You may cast that card until the end of your
//	 next turn. (You still pay its costs. Timing rules still apply.)"
//
// The card is targeted in your graveyard as the trigger goes on the stack
// ("up to one", so none is fine) and exiled as it resolves only if it is
// still an instant or sorcery card there. Its owner is you, who may cast it
// from exile until the end of your next turn (Cryptic Pursuit's seat-turn
// window, ADR 0063), paying its costs.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Scholar of Combustion — exile up to one target instant or sorcery card from your graveyard; you may cast it until the end of your next turn",
		scholarOfCombustionExile)
	enters.Targets = effects.TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
		effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:     "ebae61bf-6e8c-4cc9-8071-b63e781f6530",
		Name:         "Scholar of Combustion",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

func scholarOfCombustionExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
		Player:   item.Controller,
		CastOnly: true,
		Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
	})
}
