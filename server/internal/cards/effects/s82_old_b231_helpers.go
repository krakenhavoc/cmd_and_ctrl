package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b231_helpers.go — the cast restriction Ashiok's Erasure and Ixalan's
// Binding share (S82 older sets, batch 231).

// OpponentsCantCastSpellsNamedLikeTheExiledCard is "Your opponents can't cast
// spells with the same name as the exiled card": a static of the permanent,
// read live off the cards exiled with it (CR 607.2a, the link Card.ExiledWith
// names this object by), so each opponent of its controller is barred from a
// spell whose name is an exiled card's, and it lifts the moment the permanent
// is gone.
func OpponentsCantCastSpellsNamedLikeTheExiledCard() game.CastRestriction {
	return game.CastRestriction{
		Label: "Your opponents can't cast spells with the same name as the exiled card.",
		Forbids: func(q game.CastQuery) bool {
			if q.Game == nil || q.Controller == q.Source.Controller {
				return false
			}
			for _, c := range q.Game.ExiledWithForEffect(game.PermissionCardRef{ID: q.Source.InstanceID, Epoch: q.Source.ObjectEpoch}) {
				if c.Name != "" && (game.PermissionFilter{Named: c.Name}).Matches(q.Card) {
					return true
				}
			}
			return false
		},
	}
}
