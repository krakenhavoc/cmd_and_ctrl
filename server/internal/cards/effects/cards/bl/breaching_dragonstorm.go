package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breaching Dragonstorm — Enchantment {4}{R} (Tarkir: Dragonstorm, #4463):
//
//	"When this enchantment enters, exile cards from the top of your library
//	 until you exile a nonland card. You may cast it without paying its
//	 mana cost if that spell's mana value is 8 or less. If you don't, put
//	 that card into your hand.
//	 When a Dragon you control enters, return this enchantment to its
//	 owner's hand."
//
// The enters trigger is the exile-until-nonland run (UntilNonland) followed
// by Beseech the Mirror's offer, Game.CastExiledFreeOrToHandForEffect with a
// cap of 8: the free cast is made during the trigger's resolution (CR
// 608.2g), "that spell's mana value" is asked of the face being cast (a
// modal double-faced card or an Adventure whose castable face fits is
// offered), and a card over the cap, or one the player declines, goes into
// its owner's hand. The lands exiled on the way stay in exile.
//
// The Dragon trigger is the cycle's shared one (dragonstorm.go): any
// Dragon entering under your control, a token or a changeling included,
// returns the enchantment if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5370ee80-0354-454f-a16e-ec133e432ecb",
		Name:         "Breaching Dragonstorm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Breaching Dragonstorm — exile until a nonland card; cast it free if its mana value is 8 or less, or put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.MillToZoneKeyedForEffect(item.Controller, 0, game.ZoneExile, effects.UntilNonland, ctx.ItemTail(breachingDragonstormExiled))
				}),
			effects.WhenADragonYouControlEntersReturnThis("Breaching Dragonstorm"),
		},
	})
}

// breachingDragonstormExiled offers the nonland card the run stopped on.
// An on-disk key.
var breachingDragonstormExiled = effects.CardTail("breaching-dragonstorm/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsLand() {
			continue
		}
		return ctx.Game.CastExiledFreeOrToHandForEffect(ctx.Controller(), ctx.Source(), id, 8)
	}
	return nil
})
