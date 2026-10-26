package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unlawful Entry — Instant {1}{U}:
//
//	"Target creature gets +1/+0 and gains flying until end of turn.
//	 Whenever you put an art sticker on a creature, you may return this
//	 card from your graveyard to your hand."
//
// The second ability works only from the graveyard (CR 113.6k), where "you"
// is the card's owner (CR 108.4). A sticker is put on an object by
// game.PutStickerForEffect, which announces EventStickerPlaced with the
// sticker's kind and the object it went on; this watches for an art sticker
// on a creature (as it stands when the sticker goes on). The return is a
// "you may", asked as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "898ee071-dd37-4a02-895d-a41b568d7eef",
		Name:         "Unlawful Entry",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 1, Label: "Unlawful Entry — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Unlawful Entry — flying"}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(Optional(
				On(game.EventStickerPlaced, youPutAStickerOfKindOnACreature(game.StickerArt),
					"Unlawful Entry — return this card from your graveyard to your hand",
					returnThisFromYourGraveyardToHand),
				"Return Unlawful Entry from your graveyard to your hand?")),
		},
	})
}

// youPutAStickerOfKindOnACreature is the When of "whenever you put a <kind>
// sticker on a creature": the card's controller (its owner, in a graveyard)
// put a sticker of that kind on an object that is a creature.
func youPutAStickerOfKindOnACreature(kind game.StickerKind) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Actor != source.Controller || !strings.HasPrefix(ev.Label, string(kind)+":") {
			return false
		}
		// A creature is a permanent: a creature CARD in a graveyard that
		// got a sticker (Scampire) is not "a creature".
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && onBattlefield(g, ev.CardID) && c.IsCreature()
	}
}

// returnThisFromYourGraveyardToHand is "return this card from your
// graveyard to your hand": nothing if the card has moved on.
func returnThisFromYourGraveyardToHand(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
