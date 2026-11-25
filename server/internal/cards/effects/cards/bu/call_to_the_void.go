package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call to the Void — Sorcery {4}{B} (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Each player secretly chooses a creature they control and a creature
//	 they don't control. Then those choices are revealed. Destroy each
//	 creature chosen this way."
//
// A secret choice round (#5948, ADR 0180 amendment 2026-11-24) over every
// creature on the battlefield as the spell resolves: two asks per player,
// in turn order from the caster, the first offering only the creatures
// that player controls and the second only the ones they don't. A player
// with nothing to choose for an ask (no creature of their own, or none of
// anyone else's) chooses nothing for it. Nobody sees another player's
// choice until the last one is made; then every choice is revealed at
// once and each creature chosen (once or more) is destroyed, all at the
// same time. Nothing is targeted.
//
// A bot gives up its least valuable creature and chooses the best one it
// doesn't control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2db2c28-3ddc-41ce-a0c2-ddc555dea133",
		Name:         "Call to the Void",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			var asks []game.VoteAsk
			if len(ids) > 0 {
				players := append([]uuid.UUID{ctx.Controller()}, effects.OpponentsInTurnOrder(ctx.Game, ctx.Controller())...)
				for _, p := range players {
					asks = append(asks,
						game.VoteAsk{Voter: p, Relation: game.VoteAskControls, Question: "Call to the Void — secretly choose a creature you control"},
						game.VoteAsk{Voter: p, Relation: game.VoteAskDoesNotControl, Question: "Call to the Void — secretly choose a creature you don't control"},
					)
				}
			}
			return effects.SecretChoices{
				Question: "Call to the Void — secretly choose a creature",
				Options:  effects.VoteForPermanents(ctx.Game, ids),
				Asks:     asks,
				Then:     callToTheVoidChosen,
			}.Apply(ctx)
		},
	})
}

var callToTheVoidChosen = effects.VoteResultThen("secret/call-to-the-void", func(ctx *effects.Context, r game.VoteResult) error {
	var doomed []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, b := range r.Ballots {
		id := effects.VotedPermanent(r, b.Option)
		if id == uuid.Nil || seen[id] || !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		seen[id] = true
		doomed = append(doomed, id)
	}
	if len(doomed) == 0 {
		return nil
	}
	ctx.Game.DestroyPermanentsForEffect(doomed)
	return nil
})
