package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chalice of Life // Chalice of Death — Artifact {3} // Artifact (Innistrad
// Remastered, #5948):
//
//	Chalice of Life: "{T}: You gain 1 life. Then if you have at least 10 life
//	 more than your starting life total, transform this artifact."
//	Chalice of Death: "{T}: Target player loses 5 life."
//
// A double-faced artifact, the back face its own catalog row (oracle ID plus
// "#1"). The front gains the life, then asks the controller's life against the
// table's starting life total plus ten as the ability resolves, and turns the
// Chalice over (TransformThis, which does nothing if it has already turned over
// since the ability was activated, CR 701.28f). The back's target is a player,
// chosen as it is activated.
//
// No simplification.
func init() {
	const oracle = "28df85d6-e13a-4178-a361-7015afd316ca"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Chalice of Life",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You gain 1 life. Then if you have at least 10 life more than your starting life total, transform this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeKeyedForEffect(item.SourceCardID, item.Controller, 1,
					effects.NewContext(g, item).ItemTail(gapB07ChaliceOfLifeCheck))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Chalice of Death",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target player loses 5 life.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -5)
			},
		}},
	})
}

// gapB07ChaliceOfLifeCheck is "then if you have at least 10 life more than your
// starting life total, transform this artifact", read once the life gain has
// settled (a replacement may have changed it, or stopped it). An on-disk key.
var gapB07ChaliceOfLifeCheck = effects.CardTail("then/gap-b07-chalice-of-life-check", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Life < ctx.Game.Settings.StartingLife+10 {
		return nil
	}
	return effects.TransformThis{}.Apply(ctx)
})
