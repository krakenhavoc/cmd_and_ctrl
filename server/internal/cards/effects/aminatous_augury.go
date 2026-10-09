package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aminatou's Augury — Sorcery {6}{U}{U} (#2167):
//
//	"Exile the top eight cards of your library. You may put a land card
//	 from among them onto the battlefield. Until end of turn, for each
//	 nonland card type, you may cast a spell of that type from among the
//	 exiled cards without paying its mana cost."
//
// The exile is MillToZone{To: exile} (an exile, not a mill), and its
// continuation is handed the cards that actually reached exile. Over
// those it grants ONE stored permission (ADR 0066): free ("{0}", which
// also locks X at 0, ruling 2018-07-13), cast only, until end of turn,
// with a per-type budget of the eight nonland card types (#2167,
// CastPermission.PerType). Each spell cast through it spends one of its
// types, the caster's choice when it has two ("you could cast one
// artifact creature as your artifact card and another artifact creature
// as your creature card"). Then, if any of them is a land, the caster
// may put one onto the battlefield — before any of the free casts can
// be made, because the spell is still resolving.
//
// The other rulings (2018-07-13) fall out:
//
//   - "You may [put a land onto the battlefield] even if you've already
//     played a land this turn": it is not a land play.
//   - "If you exile a land with another card type, you can't play it
//     later in the turn": the permission is cast-only and a land can
//     never spend a nonland type.
//   - "You must follow the normal timing permissions": TimingNormal.
//   - "If you cast a card 'without paying its mana cost,' you can't
//     choose to cast it for any alternative costs": the free price is
//     the permission's own and names no claimable offer.
//   - "Casting an exiled card causes it to leave exile. You can't cast
//     it multiple times": the permission names each card OBJECT, and one
//     that has moved is a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9160997-0305-47f5-9a2f-77588d167da0",
		Name:         "Aminatou's Augury",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := item.Controller
			return MillToZone{
				Player: you,
				N:      8,
				To:     game.ZoneExile,
				Then: func(ctx *Context, exiled []uuid.UUID) error {
					return aminatousAuguryPayoff(ctx.Game, item, you, exiled)
				},
			}.Apply(ctx)
		},
	})
}

// aminatousAuguryName labels the permission, the prompt and the log.
const aminatousAuguryName = "Aminatou's Augury"

// aminatousAuguryPayoff grants the free casts over the cards that
// reached exile and offers one of their lands.
func aminatousAuguryPayoff(g *game.Game, item *game.StackItem, you uuid.UUID, exiled []uuid.UUID) error {
	var cards []game.Card
	var lands []uuid.UUID
	for _, id := range exiled {
		z := g.FindCardZoneForEffect(id)
		c, ok := g.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
		if !c.FaceDown && c.IsLand() {
			lands = append(lands, id)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     you,
		Zone:       game.ZoneExile,
		Cost:       "{0}",
		CastOnly:   true,
		PerType:    game.NonlandPermissionTypes,
		Source:     item.SourceCardID,
		SourceName: aminatousAuguryName,
		Label:      aminatousAuguryName + " — cast a spell of each nonland card type without paying its mana cost",
	}, cards)
	if len(lands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     item.SourceCardID,
		Question:   aminatousAuguryName + " — you may put a land card from among them onto the battlefield",
		Cards:      lands,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ReturnFromExileTogether{Targets: picked, Controller: you}.Apply(NewContext(g, item))
		},
	})
	return nil
}
