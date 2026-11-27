package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b424_helpers.go — S82 OLD batch 424 (Return to Ravnica, #5566): the
// overload pump spells' subject and Skymark Roc's target clause.

// OldB424Subject is who an overload pump spell acts on, decided as it resolves
// (CR 611.2c). Cast for its overload cost the spell names no target and acts on
// every creature `side` passes, the set locked now (match, with a nil id); cast
// normally it acts on its one target if that is still legal (id, with a nil
// match). ok is false when there is nothing to act on, which is a spell that
// does nothing (CR 608.2b).
func OldB424Subject(ctx *Context, side CardPredicate) (id uuid.UUID, match CardPredicate, ok bool) {
	if ctx.PaidAltCost("overload") {
		return uuid.Nil, And(Creature(), side), true
	}
	id, ok = B16FirstLegalTargetCard(ctx)
	return id, nil, ok
}

// OldB424TargetSmallCreatureDefendingPlayerControls is "target creature
// defending player controls with toughness 2 or less" for an attack trigger: the
// player the attacking creature is attacking (CR 802.2a), as
// TargetCreatureDefendingPlayerControls reads it, with the toughness bound.
func OldB424TargetSmallCreatureDefendingPlayerControls(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature defending player controls with toughness 2 or less",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == defender && c.CurrentToughness() <= 2
		},
		Min: 1, Max: 1,
	}
}
