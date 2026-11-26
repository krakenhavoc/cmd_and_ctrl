package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b515_helpers.go — shared bodies for S82 OLD batch 515 (Shards of
// Alara, #5588).

// OldB515DamageEqualToTargetsHand is "[this] deals damage to target player
// equal to the number of cards in that player's hand" (Vicious Shadows): the
// hand is counted as the ability resolves (CR 608.2b). A target that is no
// longer legal, or an empty hand, deals nothing.
func OldB515DamageEqualToTargetsHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: p.Hand.Size()}.Apply(ctx)
	}
	return nil
}

// OldB515YouControlForestAndPlains is the "you control a Forest and a Plains"
// half of Qasali Ambusher's offer: a Forest and a Plains among the
// permanents `controller` controls, read as land types (CR 205.3i).
func OldB515YouControlForestAndPlains(g *game.Game, controller uuid.UUID) bool {
	forest, plains := false, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		forest = forest || c.HasSubtypeWord("Forest")
		plains = plains || c.HasSubtypeWord("Plains")
	}
	return forest && plains
}

// OldB515ACreatureIsAttackingYou is "a creature is attacking you": a creature
// attacking the player as a player, not one of their planeswalkers (CR 506.3).
func OldB515ACreatureIsAttackingYou(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget == you {
			return true
		}
	}
	return false
}
