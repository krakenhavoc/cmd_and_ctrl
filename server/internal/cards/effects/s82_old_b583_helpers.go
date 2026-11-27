package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b583_helpers.go — shared bodies for the Saviors of Kamigawa cards of
// S82 OLD batch 583 (#5601).

// B583YouHaveMoreCardsInHandThanEachOpponent is "you have more cards in hand
// than each opponent": strictly more than every opponent still in the game.
func B583YouHaveMoreCardsInHandThanEachOpponent(g *game.Game, you uuid.UUID) bool {
	me := g.PlayerByIDForEffect(you)
	if me == nil || me.Hand == nil {
		return false
	}
	mine := me.Hand.Size()
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if p.Hand != nil && p.Hand.Size() >= mine {
			return false
		}
	}
	return true
}

// B583YouHaveCardsInHand returns "you have N or more cards in hand".
func B583YouHaveCardsInHand(n int) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, you uuid.UUID) bool {
		p := g.PlayerByIDForEffect(you)
		return p != nil && p.Hand != nil && p.Hand.Size() >= n
	}
}

// B583SelfWhileYou is the AppliesTo of "as long as <a condition on your hand>,
// this creature …": the source itself while the test holds for its controller.
func B583SelfWhileYou(test func(g *game.Game, you uuid.UUID) bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && test(g, source.Controller)
	}
}

// B583SelfPumpWhileYou is "as long as <a condition on your hand>, this
// creature gets +p/+t": a layer 7c modification read live. DependsOnHandSize
// drops the layer cache when a hand changes size.
func B583SelfPumpWhileYou(test func(g *game.Game, you uuid.UUID) bool, power, toughness int, label string) game.StaticAbility {
	return game.StaticAbility{
		Label:             label,
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		DependsOnHandSize: true,
		AppliesTo:         B583SelfWhileYou(test),
		Apply:             PumpBy(power, toughness),
	}
}

// B583SelfKeywordWhileYou is "as long as <a condition on your hand>, this
// creature has <keyword>".
func B583SelfKeywordWhileYou(test func(g *game.Game, you uuid.UUID) bool, keyword, label string) game.StaticAbility {
	ab := KeywordGrant(B583SelfWhileYou(test), keyword)
	ab.Label = label
	ab.DependsOnHandSize = true
	return ab
}

// B583SelfGrantWhileYou grants the card's own bundle `key` to this creature
// while the test holds for its controller.
func B583SelfGrantWhileYou(test func(g *game.Game, you uuid.UUID) bool, key, label string) game.StaticAbility {
	ab := GrantAbilities(B583SelfWhileYou(test), key)
	ab.Label = label
	ab.DependsOnHandSize = true
	return ab
}

// B583TargetCreatureDamagesItself is "Target creature deals damage to itself
// equal to its power": the creature is its own damage source, and its power is
// read as the spell resolves.
func B583TargetCreatureDamagesItself(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	return DealDamage{Source: t.ID, Target: t.ID, Amount: c.CurrentPower()}.Apply(ctx)
}

// B583WhenThisDealsCombatDamageToACreatureTapFreeze is the shared first
// ability of Matsu-Tribe Birdstalker and Kashi-Tribe Elite: "Whenever this
// creature deals combat damage to a creature, tap that creature and it doesn't
// untap during its controller's next untap step."
func B583WhenThisDealsCombatDamageToACreatureTapFreeze(label string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: ThisDealtCombatDamageToACreature,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, ev.Target)
			return item
		},
		Effect: OldB166TapAndFreezeDamagedCreature(label),
	}
}
