package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// return_transformed_attached.go — "return this card to the battlefield
// transformed under your control attached to target …" (#5677; Radiant
// Grace, Vengeful Strangler).

// ReturnThisCardTransformedAttachedToTarget is the resolution of a
// trigger that reads "return this card [it] to the battlefield
// transformed under your control attached to target <host>". Give the
// trigger its target with Targeting; the pick is item.Targets[0].
//
// "This card" is the source as the graveyard object it became when it
// left the battlefield (CR 400.7): one that left the graveyard while the
// trigger waited, even to come back, is not returned. The back face is
// an Aura the effect attaches, so its controller is not asked what it
// enchants (CR 303.4f), and a target the Aura can't legally enchant
// leaves the card in the graveyard (CR 303.4i). An illegal target on
// resolution removes the whole ability (CR 608.2b) before this runs.
func ReturnThisCardTransformedAttachedToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || !sourceCardInGraveyardAfterLeaving(ctx) {
		return nil
	}
	return ReturnFromGraveyard{
		Target:      item.SourceCardID,
		Dest:        game.ZoneBattlefield,
		Controller:  item.Controller,
		Transformed: true,
		AttachedTo:  item.Targets[0],
	}.Apply(ctx)
}
