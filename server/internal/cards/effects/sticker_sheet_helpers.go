package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sticker_sheet_helpers.go — pieces the ability stickers of the Unfinity
// sticker sheets share (ADR 0208, #4307). Append-only.

// tapLegalCardTargets taps each target still legal as the ability
// resolves (CR 608.2b).
func tapLegalCardTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetCardIDs(ctx) {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// stickerFirstLegalTarget is the first target card still legal, or false.
func stickerFirstLegalTarget(ctx *Context) (uuid.UUID, bool) {
	return b16FirstLegalTargetCard(ctx)
}

// CastTargetFromYourGraveyardThisTurn is "you may cast target <card> from
// your graveyard this turn" (Elemental Time Flamingo's ability sticker):
// a permission over that one card, until end of turn, for the ability's
// controller, paying the card's mana cost and every other cost as for any
// cast (CR 601.2). It is a stored per-card permission (game.CastPermission,
// ADR 0066), so the cast path, the view and the bot read one answer, and a
// card that has left the graveyard by resolution is simply not castable
// (CR 608.2b). Unlike flashback it exiles nothing.
type CastTargetFromYourGraveyardThisTurn struct {
	Target uuid.UUID
	// Label names the permission in the client's cost picker.
	Label string
}

// Apply writes the permission.
func (e CastTargetFromYourGraveyardThisTurn) Apply(ctx *Context) error {
	if e.Target == uuid.Nil {
		return nil
	}
	label := e.Label
	if label == "" {
		label = "Cast from your graveyard this turn"
	}
	ctx.Game.GrantCastPermissionOverCardForEffect(e.Target, game.CastPermission{
		Player: ctx.Controller(),
		Zone:   game.ZoneGraveyard,
		Source: ctx.Source(),
		Label:  label,
	})
	return nil
}

// supertypesTypesAndSubtypesOf is "the number of supertypes, card types,
// and subtypes it has" (Unique Charmed Pants): the distinct entries of the
// three lists as the object has them now (CR 205), a changeling counting
// every creature type (CR 702.73a).
func supertypesTypesAndSubtypesOf(c game.Card) int {
	ch := c.Effective()
	seen := map[string]bool{}
	for _, s := range ch.Supertypes {
		seen["super:"+s] = true
	}
	for _, t := range ch.Types {
		seen["type:"+t] = true
	}
	for _, s := range ch.Subtypes {
		seen["sub:"+s] = true
	}
	n := len(seen)
	if ch.AllCreatureTypes {
		for _, t := range game.AllCreatureTypes {
			if !seen["sub:"+t] {
				n++
			}
		}
	}
	return n
}
