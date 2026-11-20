package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lavakin Brawler — Creature — Elemental Warrior {3}{R}, 2/4 (Core Set 2020, #5518):
//
//	"Whenever this creature attacks, it gets +1/+0 until end of turn for each
//	 Elemental you control."
//
// The count is read as the trigger resolves and counts the Brawler itself
// (m20ElementalsYouControl).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa4ad622-5021-44fe-92af-0de115059eb6",
		Name:         "Lavakin Brawler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Lavakin Brawler — +1/+0 until end of turn for each Elemental you control",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Target: item.SourceCardID, Power: m20ElementalsYouControl(g, item.Controller),
						Label: "Lavakin Brawler — +1/+0 for each Elemental you control"}.Apply(NewContext(g, item))
				}),
		},
	})
}
