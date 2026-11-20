package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seedguide Ash — Creature — Treefolk Druid {4}{G}, 4/4 (Lorwyn, #5593):
//
//	"When this creature dies, you may search your library for up to three
//	 Forest cards, put them onto the battlefield tapped, then shuffle."
//
// The "you may" is asked as the trigger resolves (CR 603.5). A Forest card
// is any land card with the Forest subtype, a nonbasic one included, and a
// search for a stated quality may find fewer than three, or none (CR
// 701.23b). The lands enter tapped together; nothing is revealed, as the
// card doesn't say so (CR 701.23e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ff007cb-c994-4643-a6d4-c58e030514e1",
		Name:         "Seedguide Ash",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisDies("Seedguide Ash — search for up to three Forest cards, onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsLandWithSubtype("forest"),
						Dest:          game.ZoneBattlefield,
						Limit:         3,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Seedguide Ash — up to three Forest cards, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}), "Seedguide Ash — search your library for up to three Forest cards?"),
		},
	})
}
