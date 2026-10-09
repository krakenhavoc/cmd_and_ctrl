package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plan for All Outcomes — Enchantment {3}{U} (Reality Fracture):
//
//	"When this enchantment enters, the owner of up to one other target
//	 nonland permanent puts it on their choice of the top or bottom of
//	 their library.
//	 Whenever you cast your first noncreature spell each turn, empower
//	 Jace 1."
//
// ADR 0139 proof card: the keyword action from a cast trigger. The
// enters trigger is the ADR 0088 top-or-bottom placement with the
// OWNER as the chooser, Aetherspouts' shape for one permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fd2cac5-cd29-4a47-9e89-9c6be39a0b67",
		Name:         "Plan for All Outcomes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Plan for All Outcomes — its owner puts up to one other target nonland permanent on the top or bottom of their library",
					planForAllOutcomesTuck),
				Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
			On(game.EventCast, YouCastYourFirstNoncreatureSpellThisTurn,
				"Plan for All Outcomes — empower Jace 1", Do(EmpowerJace{N: 1})),
		},
	})
}

// planForAllOutcomesTuck asks the target's owner top or bottom. A
// target that left in response is skipped (CR 608.2b).
func planForAllOutcomesTuck(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || c.Owner == uuid.Nil {
			continue
		}
		return PutInLibraryInAnyOrder{
			Chooser:   c.Owner,
			Cards:     []uuid.UUID{t.ID},
			From:      game.ZoneBattlefield,
			Placement: game.LibraryPlaceTopOrBottom,
			Label:     "Plan for All Outcomes — put it on the top or bottom of your library",
		}.Apply(ctx)
	}
	return nil
}
