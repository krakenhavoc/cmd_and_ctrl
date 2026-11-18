package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_dmu_b14_helpers.go — shared bodies for the DMU batch 14 cards.

// UpToOneTargetPermanent is "up to one target <permanent>" with the label
// the card prints.
func UpToOneTargetPermanent(label string, preds ...CardPredicate) *game.TargetSpec {
	return TargetPermanent(label, preds...).WithCount(0, 1)
}

// yotiaDamageBody is Yotia Declares War's reflexive trigger: "this Saga
// deals that much damage to target creature or planeswalker", the amount
// the artifacts tapped, carried as Params.Amount.
var yotiaDamageBody = game.ReflexiveBody("yotia-declares-war/damage", breechesBlastEffect,
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))
	}))
