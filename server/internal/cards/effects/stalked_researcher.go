package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stalked Researcher — Creature — Human Wizard {1}{U}, 3/3:
//
//	"Defender
//	 Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, this creature can attack this turn as
//	 though it didn't have defender."
//
// Eerie is one ability with two conditions (Eerie, rooms.go). Defender
// is the only thing that stops a creature attacking, so "can attack as
// though it didn't have defender" is written as losing the keyword until
// end of turn (RemoveKeywordsMod). The one place that differs from the
// printed wording is a card that cares whether the creature has
// defender while it attacks; none in the catalog does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ec3c334-8a53-46d8-8cfb-c647c3a1ef74",
		Name:            "Stalked Researcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			Eerie("Stalked Researcher — can attack this turn as though it didn't have defender (eerie)",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return untilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
						"Stalked Researcher — can attack as though it didn't have defender",
						game.RemoveKeywordsMod("defender"))
				}),
		},
	})
}
