package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thassa's Oracle — Creature — Merfolk Wizard {U}{U}, 1/3:
//
//	"When this creature enters, look at the top X cards of your
//	 library, where X is your devotion to blue. Put up to one of them
//	 on top of your library and the rest on the bottom of your library
//	 in a random order. If X is greater than or equal to the number of
//	 cards in your library, you win the game."
//
// ADR 0057's resolution-check win (#749, CR 104.2b). The ETB trigger
// resolves in three beats, in printed order:
//
//  1. X is the controller's devotion to blue as the trigger resolves
//     (CR 700.5) — the Oracle's own {U}{U} counts, since it is on the
//     battlefield.
//  2. They look at the top X and pick up to one to keep (a reveal-pick
//     prompt with a floor of zero); every card they didn't keep goes to
//     the bottom in a random order, off the game's keyed RNG. Keeping
//     one card puts it on top, because everything else it was looked
//     at with has gone underneath.
//  3. If X is at least the number of cards now in the library, they win
//     the game — immediately, during the resolution. With an empty
//     library X of 0 wins, which is the combo.
//
// The win reads the library AFTER the arrangement, as printed; the
// arrangement moves nothing in or out of it, so the count is the same
// either way. An opponent's Platinum Angel prevents the win and the
// game goes on; nothing is drawn, so nothing loses either.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1de1b591-a73f-4974-b507-8c63e07a0868",
		Name:         "Thassa's Oracle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Key:    "Thassa's Oracle — look at the top X cards, then win if X is at least your library",
			Effect: thassasOracleResolve,
		}},
	})
}

// thassasOracleResolve is the trigger's effect. It captures nothing:
// the controller comes off the item, and X is fixed once, at the start
// of the resolution, and carried into the continuation by value.
func thassasOracleResolve(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	x := devotionTo(g, me, "U")
	looked := g.LookAtTopOfPlayersLibraryForEffect(me, me, x)
	return RevealPick{
		Player:   me,
		Owner:    me,
		Question: "Thassa's Oracle — put up to one of these on top of your library; the rest go to the bottom in a random order",
		Cards:    looked,
		Min:      0,
		Max:      1,
		Then: func(ctx *Context, _, left []uuid.UUID) error {
			if err := ctx.Game.PutOnBottomInRandomOrderForEffect(me, game.ZoneLibrary, left); err != nil {
				return err
			}
			p := ctx.Game.PlayerByIDForEffect(me)
			if p == nil || x < p.Library.Size() {
				return nil
			}
			return WinTheGame{Player: me}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
