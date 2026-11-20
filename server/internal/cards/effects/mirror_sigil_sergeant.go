package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirror-Sigil Sergeant — Creature — Rhino Soldier {5}{W}, 4/4
// (Conflux, #5587):
//
//	"Trample
//	 At the beginning of your upkeep, if you control a blue permanent,
//	 you may create a token that's a copy of this creature."
//
// An intervening-if upkeep trigger (CR 603.4) with a "you may" (CR
// 603.5), whose body is Scute Swarm's token copy of the source. A token
// copy has the same ability, so each copy triggers on later upkeeps
// too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c021c808-dc58-4386-8758-68b858badbd5",
		Name:            "Mirror-Sigil Sergeant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Optional(conB2AtYourUpkeepIfYouControlA("U", "blue",
				"Mirror-Sigil Sergeant — create a token that's a copy of it",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
				"Mirror-Sigil Sergeant — create a token that's a copy of it?"),
		},
	})
}
