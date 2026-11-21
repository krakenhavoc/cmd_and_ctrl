package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bismuth Mindrender — Creature — Eldrazi {3}{B}, 4/3:
//
//	"Devoid (This card has no color.)
//	 Menace
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles cards from the top of their library until they exile a
//	 nonland card. You may cast that card by paying life equal to the
//	 spell's mana value rather than paying its mana cost."
//
// Black Widow, Super Spy's exile-until-nonland on the damaged player's
// library. The cast is offered during the resolution (CR 608.2g, the
// resolution_cast prompt Bilbo uses): any timing, cast only, paying life
// equal to the spell's mana value as the alternative cost (CR 118.9, 119.4,
// 202.3: Bolas's Citadel's offer), and a player without the life is not
// offered it. Declining leaves the card in exile, where it stays; the
// permission does not outlast the resolution it was offered in.
func init() {
	const name = "Bismuth Mindrender"
	Register(Spec{
		OracleID:        "12b859f8-36f3-43cd-b537-6e07857b247e",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(name+" — that player exiles cards from the top of their library until they exile a nonland card; you may cast it by paying life equal to its mana value",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					return MillToZone{
						Player:   item.Trigger.Event.Target,
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.ItemTail(bismuthMindrenderExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// bismuthMindrenderExiled offers the cast of the nonland card the exile
// ended on. An on-disk key.
var bismuthMindrenderExiled = CardTail("bismuth-mindrender/after-exile", func(ctx *Context, o game.TailOutcome) error {
	var hits []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if ok && !c.IsLand() && z != nil && z.Kind == game.ZoneExile {
			hits = append(hits, id)
		}
	}
	if len(hits) == 0 {
		return nil
	}
	perm := game.CastPermission{
		Player:               ctx.Controller(),
		Zone:                 game.ZoneExile,
		Duration:             ctx.Game.UntilEndOfTurnDuration(),
		AltCostKey:           "bismuth_mindrender",
		LifeEqualToManaValue: true,
		Timing:               game.TimingFlash,
		CastOnly:             true,
		LapseOnPass:          game.LapseStaysInExile,
		Source:               ctx.Source(),
		SourceName:           "Bismuth Mindrender",
		Label:                "Pay life equal to its mana value (Bismuth Mindrender)",
	}
	ctx.Game.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Reason:       "Bismuth Mindrender",
		Cards:        hits,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordPaidCast,
		AcceptLabel:  "Cast it, paying life",
		DeclineLabel: "Leave it in exile",
	})
	return nil
})
