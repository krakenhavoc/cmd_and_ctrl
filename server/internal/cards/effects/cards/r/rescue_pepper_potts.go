package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rescue, Pepper Potts — Legendary Artifact Creature — Human Hero {1}{U},
// 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Flying
//	 When Rescue enters, return up to one other target artifact or creature
//	 you control to its owner's hand. If it was an artifact, put a +1/+1
//	 counter on Rescue."
//
// An enters trigger with an "up to one other" clause (Repulsor Bots'
// Another(..).WithCount shape). Whether the returned permanent was an
// artifact is read before the bounce, from the permanent as it stood on the
// battlefield. The counter goes on Rescue only while it is still the
// permanent that triggered (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "54917578-9493-499e-b200-51d5cf7978ff",
		Name:            "Rescue, Pepper Potts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Rescue — return up to one other target artifact or creature you control to its owner's hand; a +1/+1 counter on Rescue if it was an artifact",
				rescuePepperPottsBounce),
				effects.Another(effects.TargetPermanent("up to one other target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()).WithCount(0, 1))),
		},
	})
}

func rescuePepperPottsBounce(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	wasArtifact := c.IsArtifact()
	if err := effects.B35BounceChosenPermanents(ctx); err != nil {
		return err
	}
	if !wasArtifact || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return effects.PutACounterOnThis(game.CounterPlusOne)(g, item)
}
