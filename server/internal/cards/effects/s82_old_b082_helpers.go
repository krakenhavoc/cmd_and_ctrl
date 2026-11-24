package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b082_helpers.go — helpers for OLD batch 082 (#5507), Commander
// Legends encore creatures.

// S82OldB082TargetAttackerOfTheSameTarget is the target clause of "another
// target creature attacking the same player or planeswalker": a creature, other
// than the trigger's source, whose attack target is the one the source was
// declared against (the event's Target).
func S82OldB082TargetAttackerOfTheSameTarget(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	attacked := tc.Event.Target
	if attacked == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "another target creature attacking the same player or planeswalker",
		Zones:         []game.ZoneKind{game.ZoneBattlefield},
		ExcludeSource: true,
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.AttackingTarget == attacked
		},
		Min: 1, Max: 1,
	}
}

// S82OldB082ThisGetsPowerPerLandDefendingPlayerControls is "it gets +1/+0 until
// end of turn for each land defending player controls", the defending player
// being the one the attack was declared against (CR 508.5), counted as the
// trigger resolves.
func S82OldB082ThisGetsPowerPerLandDefendingPlayerControls(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
		if defender == uuid.Nil {
			return nil
		}
		n := CountControlled(g, defender, func(c game.Card) bool { return c.HasCardType("land") })
		if n == 0 {
			return nil
		}
		return ThisGetsUntilEndOfTurn(n, 0, label)(g, item)
	}
}

// S82OldB082DamageAttackedPlayerPerArtifact is "it deals damage to the player
// or planeswalker it's attacking equal to the number of artifacts you control",
// counted as the trigger resolves. The engine has no planeswalker defenders, so
// the attacked object is the player.
func S82OldB082DamageAttackedPlayerPerArtifact(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(defender) == nil {
		return nil
	}
	n := CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasCardType("artifact") })
	if n == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: defender, Amount: n}.Apply(NewContext(g, item))
}
