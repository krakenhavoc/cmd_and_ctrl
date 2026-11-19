package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB50CountersPlacedOnThis is "whenever one or more +1/+1 counters are put
// on this permanent": one trigger per placement event that added a counter.
func gapB50CountersPlacedOnThis(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Target == source.InstanceID && b33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
}

// gapB50AttacksAnOpponent is "whenever this creature attacks an opponent": the
// declared attack's defender is a player other than the controller, not a
// planeswalker or battle.
func gapB50AttacksAnOpponent(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ThisAttacked(ev, source, lki, g) &&
		g.ClassifyAttackTargetForEffect(ev.Target) == game.AttackTargetPlayer &&
		ev.Target != source.Controller
}

// gapB50KaaliaPick puts the chosen card from the hand onto the battlefield
// tapped and attacking the opponent Kaalia attacked, and only that opponent
// (the carry's first id). An on-disk key: never renamed, never reused.
var gapB50KaaliaPick = game.RegisterChooseCardsThen("choose-cards/kaalia-of-the-vast-put",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 || r.Asked == nil || len(r.Carry) == 0 {
			return nil
		}
		card := r.Picked[0]
		if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneHand {
			return nil
		}
		attacking := uuid.Nil
		for _, ref := range g.AttackTargetsForEffect(r.Asked.Controller) {
			if ref.ID == r.Carry[0] {
				attacking = ref.ID
			}
		}
		return g.PutFromHandOntoBattlefieldKeyedForEffect(card, game.HandEntryOptions{
			Controller: r.Asked.Controller,
			Tapped:     true,
			Attacking:  attacking,
		}, game.KeyedTail{})
	})

// gapB50SacrificedAngel reports whether the permanent a cost sacrificed was
// an Angel as it last existed on the battlefield.
func gapB50SacrificedAngel(info game.PermanentInfo) bool {
	return slices.Contains(info.Characteristic.Subtypes, "Angel")
}
