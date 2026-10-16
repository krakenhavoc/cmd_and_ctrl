package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_exile.go — #3192, ADR 0066's second 2026-10-13
// amendment: the card-side vocabulary for an exiled card the holder may
// play only while something is true.
//
// The engine half is game.CastPermission.Condition (a registered key
// with plain parameters, read by the one liveness test). The cards write
// it through ExileTopWithPermission's Condition, Window and Filter
// fields; the helpers here only build the three pieces of data that
// recur.

// attackedCondition builds a condition on what the holder attacked with
// this turn: "during any turn you attacked with …".
func attackedCondition(key game.PermissionConditionKey) *game.PermissionCondition {
	return &game.PermissionCondition{Key: key}
}

// attackedWithObject is "during any turn you attacked with this
// creature": the condition names the source OBJECT (id and CR 400.7
// epoch) as the trigger put it on the stack, so a creature that has
// since left and come back is not it.
func attackedWithObject(item *game.StackItem) *game.PermissionCondition {
	return &game.PermissionCondition{
		Key:         game.CondAttackedWithObject,
		ObjectID:    item.SourceObject.ID,
		ObjectEpoch: item.SourceObject.Epoch,
	}
}

// forAsLongAsYouControlWindow is "for as long as you control <source>"
// (CR 611.2b) as a permission's duration. The engine already carries it
// (game.Duration ForAsLongAs + WhileYouControlSource, read by the same
// liveness test), so no condition is involved. A source that is already
// gone, or not yours, gets a window that has never started: the cards
// that print it still exile, and the exiled card is simply never
// playable (CR 611.2b: an effect whose condition is false as it would
// begin does nothing).
func forAsLongAsYouControlWindow(g *game.Game, source, player uuid.UUID) game.Duration {
	if d, ok := g.ForAsLongAsYouControlDuration(source, player); ok {
		return d
	}
	return game.Duration{
		Kind: game.ForAsLongAs, Condition: game.WhileYouControlSource,
		Player: player, Source: source, SourceEnteredAt: -1,
	}
}

// exileTopPlayableWhileYouControlThis is "exile the top card of your
// library. You may play that card for as long as you control <this>"
// (Lightning, Security Sergeant; Gwen Stacy): one card, the controller's
// own library, with forAsLongAsYouControlWindow as the window.
func exileTopPlayableWhileYouControlThis(g *game.Game, item *game.StackItem) error {
	window := forAsLongAsYouControlWindow(g, item.SourceCardID, item.Controller)
	return ExileTopWithPermission{
		From: item.Controller, GrantTo: item.Controller, N: 1,
		Window: &window,
	}.Apply(NewContext(g, item))
}

// forAsLongAsOnBattlefieldWindow is "for as long as <source> remains on
// the battlefield" / "as long as this Saga remains", with the same
// never-started fallback.
func forAsLongAsOnBattlefieldWindow(g *game.Game, source uuid.UUID) game.Duration {
	if d, ok := g.ForAsLongAsOnBattlefieldDuration(source); ok {
		return d
	}
	return game.Duration{
		Kind: game.ForAsLongAs, Condition: game.WhileSourceOnBattlefield,
		Source: source, SourceEnteredAt: -1,
	}
}

// defendingPlayerOfAttack is the "defending player" of an attack
// declaration event: the player attacked, or the controller of the
// planeswalker or battle attacked.
func defendingPlayerOfAttack(g *game.Game, ev game.Event) uuid.UUID {
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return ev.Target
	}
	if c, ok := g.LookupCardForEffect(ev.Target); ok {
		return c.Controller
	}
	return uuid.Nil
}

// attackersYouControlOtherThan counts the creatures `controller` has
// declared as attackers in the declaration under way, not counting
// `except`: the "and at least two other creatures attack" of battalion.
// Every attacker is staged before the declaration is announced, so the
// source's own EventAttack already sees the whole group.
func attackersYouControlOtherThan(g *game.Game, controller, except uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.InstanceID != except && c.AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n
}
