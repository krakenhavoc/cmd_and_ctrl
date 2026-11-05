package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Don & Leo, Problem Solvers — Legendary Creature — Mutant Ninja Turtle
// {3}{W/U}{W/U}, 4/6 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Vigilance
//	 At the beginning of your end step, exile up to one target artifact
//	 you control and up to one target creature you control. Then return
//	 them to the battlefield under their owners' control."
//
// Two target clauses, each "up to one", flickered as one exile and one
// return (Brago, King Eternal's body), so the two come back together and
// each sees the other enter (CR 603.6a). An artifact creature may be
// named by both clauses (CR 115.3, one object for each "target"); it is
// exiled and returned once. A token does not come back (CR 111.8), and
// each card returns under its owner's control as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4ae71c8-ea1b-4fa4-9e47-cf00ab60c5ee",
		Name:            "Don & Leo, Problem Solvers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("Don & Leo, Problem Solvers — exile up to one artifact and up to one creature you control, then return them",
					donAndLeoFlicker),
				Clauses(
					TargetPermanent("up to one target artifact you control", Artifact(), YouControl()).WithCount(0, 1),
					TargetPermanent("up to one target creature you control", Creature(), YouControl()).WithCount(0, 1),
				)),
		},
	})
}

// donAndLeoFlicker exiles the still-legal targets as one event, each once,
// and returns what reached exile under its owner's control.
func donAndLeoFlicker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		ids = append(ids, t.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(bragoReturn))
}
