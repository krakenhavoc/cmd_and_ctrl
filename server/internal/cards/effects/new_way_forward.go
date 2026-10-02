package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// New Way Forward — Instant {2}{U}{R}{W}:
//
//	"The next time a source of your choice would deal damage to you this turn, prevent that damage. When damage is prevented this way, New Way Forward deals that much damage to that source's controller and you draw that many cards."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8, 609.7a). "When
// damage is prevented this way" is a reflexive trigger (CR 603.12): the
// shield's additional effect (CR 615.5) puts it on the stack with the
// amount prevented and the source's controller as they were when the
// damage would have been dealt. It deals that much damage from New Way
// Forward as it last existed, and you draw that many cards.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "429368e6-3de1-4a44-a062-86cbbb73e243",
		Name:         "New Way Forward",
		Completeness: CompletenessFull,
		OnResolve:    nextDamageShieldSpell(PreventNextDamageFromChosenSource(ShieldYou).WithThen(newWayForwardPreventedBody)),
	})
}

var (
	// The shield's follow-up: nothing was prevented, nothing triggers.
	newWayForwardPreventedBody = game.DelayedBody("new-way-forward/prevented", newWayForwardPrevented)
	// The reflexive trigger itself, untargeted.
	newWayForwardStrikeBody = game.DelayedBody("new-way-forward/strike-back", newWayForwardStrike)
)

func newWayForwardPrevented(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	t := WhenYouDo("New Way Forward — deal the prevented damage to its source's controller and draw that many", newWayForwardStrikeBody)
	t.Params = game.EffectParams{Amount: p.Amount, Player: p.Player}
	return t.Apply(NewContext(g, item))
}

func newWayForwardStrike(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	if p.Player != uuid.Nil {
		if err := g.DealDamageToPlayerForEffect(item.SourceCardID, p.Player, p.Amount); err != nil {
			return err
		}
	}
	return DrawCards{Player: item.Controller, N: p.Amount}.Apply(NewContext(g, item))
}
