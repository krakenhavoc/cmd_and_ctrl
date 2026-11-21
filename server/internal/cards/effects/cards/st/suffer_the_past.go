package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suffer the Past — Instant {X}{B} (Modern Horizons 3 Commander, #5948):
//
//	"Exile X target cards from target player's graveyard. For each card
//	 exiled this way, that player loses 1 life and you gain 1 life."
//
// Two target clauses (Quandrix Command's shape): the player, then exactly X
// cards in THAT player's graveyard (FromTargetPlayersGraveyard), X being the
// value announced with the spell (CR 601.2c). Each card is re-judged as the
// spell resolves (CR 608.2b): one that has left the graveyard is skipped and
// the rest are exiled. The drain is one point per card that actually reached
// exile (the continuation is told them, so a card an owner's replacement sent
// elsewhere does not count): the player loses that much life and you gain
// that much. The life is lost and gained only while the player is still a
// legal target (CR 608.2b); a player who left the game has no life to lose.
//
// No simplification.
func init() {
	cards := effects.TargetCardInGraveyard("X target cards from that player's graveyard")
	cards.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "edf46ef0-159d-4f0b-ae14-14e0c8446382",
		Name:         "Suffer the Past",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(cards, 0),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetCardsThen(ctx, sufferThePastDrain)
		},
	})
}

// sufferThePastDrain is "for each card exiled this way, that player loses 1
// life and you gain 1 life", keyed (#4253 slice 9).
var sufferThePastDrain = effects.GraveyardExileTail("graveyard-exile/suffer-the-past-drain", func(ctx *effects.Context, exiled []uuid.UUID, _ map[uuid.UUID]bool) error {
	n := len(exiled)
	if n == 0 || ctx.Item == nil {
		return nil
	}
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -n); err != nil {
		return err
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
})
