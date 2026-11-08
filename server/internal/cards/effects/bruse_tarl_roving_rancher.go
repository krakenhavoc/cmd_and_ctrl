package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bruse Tarl, Roving Rancher — Legendary Creature — Human Warrior
// {2}{R}{W}, 4/3 (Outlaws of Thunder Junction, #4468):
//
//	"Oxen you control have double strike.
//	 Whenever Bruse Tarl enters or attacks, exile the top card of your
//	 library. If it's a land card, create a 2/2 white Ox creature token.
//	 Otherwise, you may cast it until the end of your next turn."
//
// The double strike is a layer 6 lord grant over your Oxen, the word read
// through Bruse's text (TribeFilter, CR 612.2). The trigger is Sun
// Titan's enters-or-attacks shape. The exile is the keyed impulse exile
// (Game.ExileTopWithPermissionKeyedForEffect, Bonehoard Dracosaur's
// form), so the land check reads the card that actually landed, after
// any commander prompt is answered. The permission is cast-only, until
// the end of your next turn (a seat-turn window, ADR 0063): a land can't
// be cast (CR 305.1), so an exiled land is never playable and gives the
// Ox instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c22589e1-1721-43a3-897e-990af6ebcffa",
		Name:         "Bruse Tarl, Roving Rancher",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Ox"}, YoursOnly: true}, "double strike"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Bruse Tarl, Roving Rancher — exile the top card of your library; a land makes a 2/2 Ox, anything else you may cast until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					perm := game.CastPermission{
						CastOnly: true,
						Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
					}
					return g.ExileTopWithPermissionKeyedForEffect(item.Controller, item.Controller, 1, perm,
						ctx.ItemTail(bruseTarlOxIfLand))
				}),
		},
	})
}

// bruseTarlOxIfLand is "If it's a land card, create a 2/2 white Ox
// creature token", told the card the exile landed. An on-disk key: never
// renamed, never reused.
var bruseTarlOxIfLand = CardTail("bruse-tarl/ox-if-land", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !c.IsLand() {
			continue
		}
		if err := (CreateToken{Template: TokenCard("2/2 white Ox"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
