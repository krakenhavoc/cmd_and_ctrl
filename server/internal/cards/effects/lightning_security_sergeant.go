package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning, Security Sergeant — Legendary Creature — Human Soldier
// {2}{R}, 2/3:
//
//	"Menace
//	 Whenever Lightning deals combat damage to a player, exile the top
//	 card of your library. You may play that card for as long as you
//	 control Lightning."
//
// "For as long as you control Lightning" is a source-bound duration the
// permission already carries (game.ForAsLongAs + WhileYouControlSource,
// CR 611.2b), read by the same liveness test as every other grant, so
// this card needs no condition. It names the Lightning object (CR
// 400.7) and ends for good when she leaves the battlefield or changes
// hands. If she is already gone as the trigger resolves the card is
// exiled all the same and is never playable.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1626880-0fa3-441a-9978-b293fc87c418",
		Name:            "Lightning, Security Sergeant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Lightning, Security Sergeant — exile the top card of your library", func(g *game.Game, item *game.StackItem) error {
				window := forAsLongAsYouControlWindow(g, item.SourceCardID, item.Controller)
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					Window: &window,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
