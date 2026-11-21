package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rogues' Gallery — Sorcery {2}{B} (Streets of New Capenna):
//
//	"For each color, return up to one target creature card of that color
//	 from your graveyard to your hand."
//
// Five clauses, one per colour, each "up to one" (CR 601.2c: every
// "target" word is its own choice, so a multicoloured card may fill two
// clauses). The clauses are independent, so one that has become illegal
// does not stop the others (CR 608.2b); a card chosen for two colours is
// returned once.
//
// No simplification.
func init() {
	clause := func(word, letter string) *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to one target "+word+" creature card from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.OfColor(letter)).WithCount(0, 1)
	}
	effects.Register(effects.Spec{
		OracleID:     "7ea1b0e8-a577-47e2-b4cb-a6ca973742d0",
		Name:         "Rogues' Gallery",
		Completeness: effects.CompletenessFull,
		Targets: clause("white", "W").Then(
			clause("blue", "U"),
			clause("black", "B"),
			clause("red", "R"),
			clause("green", "G"),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			seen := map[uuid.UUID]bool{}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || seen[t.ID] {
					continue
				}
				seen[t.ID] = true
				if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
