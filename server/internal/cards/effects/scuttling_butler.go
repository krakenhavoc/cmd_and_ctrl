package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scuttling Butler — Artifact Creature — Construct {3}, 4/1 (Streets of
// New Capenna Commander, #5491):
//
//	"At the beginning of combat on your turn, if you control two or more
//	 multicolored permanents, this creature gains double strike until end
//	 of turn."
//
// An intervening "if" (CR 603.4): the count is made as combat begins and
// again as the trigger resolves, on the layered colours, so a permanent
// that stops being multicolored in between turns the grant off. The Butler
// itself is colourless and never counts. The grant is on the Butler as it
// is when the trigger resolves; if it has left the battlefield nothing is
// granted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b51d6e5-a4cd-4f2f-9762-6ad2cfd8b084",
		Name:         "Scuttling Butler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(snc17TwoMulticoloredPermanents, AtBeginningOfYourCombat(
				"Scuttling Butler — gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"double strike"},
						Label:    "Scuttling Butler — double strike",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
