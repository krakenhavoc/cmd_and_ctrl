package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidetsugu and Kairi — Legendary Creature — Ogre Demon Dragon
// {2}{U}{U}{B}, 5/4:
//
//	"Flying
//	 When Hidetsugu and Kairi enters, draw three cards, then put two cards
//	 from your hand on top of your library in any order.
//	 When Hidetsugu and Kairi dies, exile the top card of your library.
//	 Target opponent loses life equal to its mana value. If it's an instant
//	 or sorcery card, you may cast it without paying its mana cost."
//
// The enters ability is Brainstorm's draw-then-put-back (a registered
// DrawThen, PutFromHandOnTopInAnyOrder).
//
// The dies ability exiles the top card through the keyed exile (a
// replacement on the exile is honoured), and its tail, with the trigger's
// target in hand, does the rest in printed order: the target opponent loses
// life equal to the exiled card's mana value (a loss, CR 119.3; a land is mana
// value 0, so nothing is lost), then, for an instant or sorcery card, the
// free cast is offered as part of the resolution (CR 608.2g), the exiled
// card staying in exile if it is declined.
//
// No simplification.
var hidetsuguPutBack = game.RegisterDrawThen("hidetsugu-and-kairi-put-back", func(g *game.Game, d game.DrawThen) error {
	ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
	return PutFromHandOnTopInAnyOrder{
		Player: d.Player,
		N:      d.N,
		Label:  "Hidetsugu and Kairi — put two cards from your hand on top of your library",
	}.Apply(ctx)
})

var hidetsuguExiled = CardTail("tail/s82-mom-b21-hidetsugu-exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if mv := card.ManaValue(); mv > 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -mv); err != nil {
				return err
			}
		}
		break
	}
	if !card.IsInstant() && !card.IsSorcery() {
		return nil
	}
	const name = "Hidetsugu and Kairi"
	perm := ctx.Game.FreeCastFromExilePermission(ctx.Controller(), ctx.Source(), name)
	return exileWithPermissionThenOffer(ctx.Game, card.InstanceID, ctx.Source(), perm, name)
})

func init() {
	Register(Spec{
		OracleID:        "18591d9e-c9f5-4ae8-aee1-8580bc8fa600",
		Name:            "Hidetsugu and Kairi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hidetsugu and Kairi — draw three cards, then put two cards from your hand on top of your library in any order",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
						Ref: hidetsuguPutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
					})
				}),
			Targeting(
				WhenThisDies("Hidetsugu and Kairi — exile the top card of your library; target opponent loses life equal to its mana value; if it's an instant or sorcery card, you may cast it free",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 1, ctx.ItemTail(hidetsuguExiled))
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
