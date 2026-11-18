package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul of Emancipation — Creature — Avatar {4}{G}{W}{U}, 5/7 (Streets of New
// Capenna, #5491):
//
//	"When this creature enters, destroy up to three other target nonland
//	 permanents. For each of those permanents, its controller creates a 3/3
//	 white Angel creature token with flying."
//
// The permanents are destroyed together (CR 700.4). Each one's controller is
// read before the destroy (CR 608.2h), since the permanent is in a graveyard
// afterwards; a target that was an illegal target by resolution is not one of
// "those permanents" (CR 608.2b). A permanent that is indestructible or is
// regenerated still counts: the token follows the instruction, not the
// destruction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9666fd2-3ebf-4f57-a51d-a80e625f39fa",
		Name:         "Soul of Emancipation",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Soul of Emancipation — destroy up to three other target nonland permanents; each one's controller creates a 3/3 white Angel with flying",
					soulOfEmancipationResolve),
				Another(TargetPermanent("up to three other target nonland permanents", Nonland()).WithCount(0, 3))),
		},
	})
}

func soulOfEmancipationResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	var controllers []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || slices.Contains(ids, t.ID) {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		ids = append(ids, t.ID)
		controllers = append(controllers, c.Controller)
	}
	if len(ids) == 0 {
		return nil
	}
	g.DestroyPermanentsForEffect(ids)
	for _, who := range controllers {
		if err := (CreateToken{Controller: who, Template: TokenCard("3/3 white Angel with flying"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
