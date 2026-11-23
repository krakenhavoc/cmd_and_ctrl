package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2b51d6e5-a4cd-4f2f-9762-6ad2cfd8b084",
		Name:         "Scuttling Butler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.Snc17TwoMulticoloredPermanents, effects.AtBeginningOfYourCombat(
				"Scuttling Butler — gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"double strike"},
						Label:    "Scuttling Butler — double strike",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
