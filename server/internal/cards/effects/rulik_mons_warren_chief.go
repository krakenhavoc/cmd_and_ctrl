package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rulik Mons, Warren Chief — Legendary Creature — Goblin {1}{R}{G}{G},
// 3/3 (Dominaria United):
//
//	"Menace
//	 Whenever Rulik Mons attacks, look at the top card of your library.
//	 If it's a land card, you may put it onto the battlefield tapped. If
//	 you didn't put a card onto the battlefield this way, create a 1/1
//	 red Goblin creature token."
//
// A private look at the top card (CR 701.20), the land put as Elvish
// Rejuvenator's optional entry, and the Goblin as the put's tail: it is
// made whenever nothing entered, whether the card was a nonland, the
// land was declined, a replacement kept it off the battlefield or the
// library was empty. A card that was not put stays on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "741dc9c3-1998-4f4c-9c07-6f34fea2bae4",
		Name:            "Rulik Mons, Warren Chief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Lands: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rulik Mons — look at the top card of your library; if it's a land card, you may put it onto the battlefield tapped; if you didn't put a card onto the battlefield this way, create a 1/1 red Goblin creature token",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Rulik Mons — you may put the land card onto the battlefield tapped",
						ThenTail: rulikMonsGoblinIfNoneEntered,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// rulikMonsGoblinIfNoneEntered is "If you didn't put a card onto the
// battlefield this way, create a 1/1 red Goblin creature token", keyed.
// An on-disk key.
var rulikMonsGoblinIfNoneEntered = PutFromLibraryTail("rulik-mons/goblin-if-none-entered",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 {
			return nil
		}
		return CreateToken{Controller: res.Player, Template: TokenCard("1/1 red Goblin"), N: 1}.Apply(ctx)
	})
