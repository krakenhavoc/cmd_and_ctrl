package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael, Fiendish Savior — Legendary Creature — Devil Noble {3}{B}{R},
// 4/4 (Commander Legends: Battle for Baldur's Gate, #5490):
//
//	"Flying
//	 Other Demons, Devils, Imps, and Tieflings you control get +1/+1 and
//	 have lifelink.
//	 At the beginning of each end step, if a creature card was put into
//	 your graveyard from anywhere this turn, create a 1/1 red Devil
//	 creature token with "When this token dies, it deals 1 damage to any
//	 target.""
//
// Flying is printed. The lord half is TribalAnthem and TribalKeywordGrant
// over the four types, Raphael himself left out. "Each end step" is every
// player's, so the trigger does not read the event's actor (Joined
// Researchers' shape). The "if" is an intervening if (CR 603.4), asked as
// the end step begins and again as the ability resolves, against
// Macabre Reconstruction's read: a nontoken creature card you own that
// arrived in your graveyard this turn from any zone, so a discarded or
// milled creature card counts as well as one that died. The Devil is the
// catalog's printed Devil (Spiked Corridor's).
//
// No simplification.
func init() {
	fiends := TribeFilter{Tribes: []string{"Demon", "Devil", "Imp", "Tiefling"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "23bd166b-d766-4d03-8016-6708481d2195",
		Name:            "Raphael, Fiendish Savior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(fiends, 1, 1),
			TribalKeywordGrant(fiends, "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureCardPutIntoYourGraveyardThisTurn(g, source.Controller)
			}, "Raphael, Fiendish Savior — create a 1/1 red Devil token",
				func(g *game.Game, item *game.StackItem) error {
					if !creatureCardPutIntoYourGraveyardThisTurn(g, item.Controller) {
						return nil
					}
					return CreateToken{Template: devilToken(), N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
