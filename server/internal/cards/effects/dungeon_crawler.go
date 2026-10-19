package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dungeon Crawler — Creature — Zombie {B}, 2/1:
//
//	"This creature enters tapped.
//	 Whenever you complete a dungeon, you may return this card from your
//	 graveyard to your hand."
//
// The second ability works only from the graveyard (InGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26b53524-9ec5-4ec9-a1e2-73e78959481b",
		Name:         "Dungeon Crawler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			InGraveyard(Optional(WheneverYouCompleteADungeon("Dungeon Crawler — return this card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
				}), "Return Dungeon Crawler from your graveyard to your hand?")),
		},
	})
}
