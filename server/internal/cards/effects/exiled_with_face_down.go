package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exiled_with_face_down.go — "exile it face down ... you may play cards
// exiled with this" (CR 607.2a, CR 406.3, ADR 0091's 2026-10-15
// amendment and ADR 0066's, #3787). Urianger Augurelt and Kayla's Music
// Box are the two cards on it.
//
// Two linked abilities, one pair of primitives:
//
//	ExileTopFaceDownLinked  the look-and-exile. The card lands
//	                        FaceDownHidden (its source's controller may
//	                        look at it, nobody else may) and carries
//	                        Card.ExiledWith naming the source object.
//	PlayCardsExiledWith     the payoff: a STORED STANDING cast
//	                        permission over exile, narrowed by
//	                        PermissionFilter.ExiledWithSource, for the
//	                        rest of the turn, optionally with "spells you
//	                        cast this way cost {N} less".
//
// The link is an OBJECT reference {instance, CR 400.7 epoch}, captured
// from the activation (HiddenRefOfActivation), so an ability still on the
// stack when its source leaves reads what that object exiled, and a
// source that left and came back has exiled nothing (CR 607.2a reads the
// exile, not the permanent).

// ExileTopFaceDownLinked is "look at the top card of your library. You
// may exile it face down" (Urianger Augurelt's Draw Arcanum) and, with
// Required, "look at the top card of your library, then exile it face
// down" (Kayla's Music Box): CR 701.20 "look at", so only the player
// becomes a knower; then CR 406.3's face-down exile, linked to Source.
//
// Optional is a prompt with a floor of zero over the one card; Required
// asks nothing. A library with no card in it does nothing and is not an
// error (looking is not drawing, so no CR 704.5b loss either).
//
// The continuation captures only scalars and the object reference, so an
// undo replays it against the restored game.
type ExileTopFaceDownLinked struct {
	// Source is the permanent OBJECT whose ability this is.
	Source game.PermissionCardRef
	// Player looks and exiles. Zero means the resolving item's controller.
	Player uuid.UUID
	// Name labels the prompt ("Urianger Augurelt").
	Name string
	// Required drops the "you may": the card is exiled without asking.
	Required bool
}

func (e ExileTopFaceDownLinked) Apply(ctx *Context) error {
	player := e.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return exileTopFaceDownLinked(ctx.Game, player, e.Source, e.Name, e.Required)
}

func exileTopFaceDownLinked(g *game.Game, player uuid.UUID, source game.PermissionCardRef, name string, required bool) error {
	looked := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(looked) == 0 {
		return nil
	}
	exile := func(g *game.Game, picked []uuid.UUID) error {
		if len(picked) == 0 {
			return nil
		}
		_, err := g.ExileHiddenLinkedForEffect(source, player, picked[0])
		return err
	}
	if required {
		return exile(g, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source.ID,
		Question: name + " — exile it face down?",
		Cards:    looked,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneLibrary,
		Then:     exile,
	})
	return nil
}

// PlayCardsExiledWith is "until end of turn, you may play cards exiled
// with <this>" (CR 607.2a), with "spells you cast this way cost {N} less
// to cast" when CostLess is set.
//
// A STORED STANDING permission, not a ScopeCards one over the cards
// there are now: the clause is about every card exiled with the source
// for the rest of the turn, so a card exiled after the activation (an
// untap effect, a second activation) is playable too. Its filter is
// ExiledWithSource with the source object written into it directly —
// the catalog derivation that fills it for a battlefield permanent's
// own standing permission (stampStandingPermissionLocked) does not run
// for a stored one. Play, not cast: a land card is playable (CR 305.1),
// and CastPermission.CastOnly is left unset.
//
// The Duration is the zero value, "until end of turn" stamped to the
// turn of the grant. No ZoneOwner, because exile is shared. Nothing
// exiled with the source means the permission opens nothing, which is
// the printed behaviour and not an error.
type PlayCardsExiledWith struct {
	// Source is the permission-granting object, as the activation
	// captured it (HiddenRefOfActivation).
	Source game.PermissionCardRef
	// Player may play them. Zero means the resolving item's controller.
	Player uuid.UUID
	// CostLess is "cost {N} less": generic mana taken off a spell cast
	// through this permission.
	CostLess int
	// OwnedByPlayer is Kayla's Music Box's "cards YOU OWN exiled with
	// ...": a card exiled with the source whose owner is someone else is
	// not covered.
	OwnedByPlayer bool
	Label         string
}

func (p PlayCardsExiledWith) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	ctx.Game.GrantCastPermissionForEffect(game.CastPermission{
		Player: player,
		Zone:   game.ZoneExile,
		Scope:  game.ScopeStanding,
		Filter: game.PermissionFilter{
			ExiledWithSource: true,
			ExiledWith:       p.Source,
			OwnedByHolder:    p.OwnedByPlayer,
		},
		CostLess: p.CostLess,
		Source:   p.Source.ID,
		Label:    p.Label,
	})
	return nil
}
