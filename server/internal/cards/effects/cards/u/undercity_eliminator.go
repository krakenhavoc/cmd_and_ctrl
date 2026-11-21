package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Eliminator — Creature — Gorgon Assassin {3}{B}{B}, 3/3
// (Murders at Karlov Manor, #4469):
//
//	"When this creature enters, you may sacrifice an artifact or
//	 creature. When you do, exile target creature an opponent controls."
//
// The enters trigger asks the "you may sacrifice" as it resolves (In the
// Pale Moonlight's maySacrificeOne): the Eliminator itself and any Clue
// are legal, and nothing to sacrifice asks nothing. Only once a
// permanent has really been sacrificed (CR 701.21a) does the CR 603.12
// reflexive trigger exist; its target is chosen as it goes on the stack,
// not when the enters trigger was announced, and every player may respond
// to it (ruling). With no legal target it is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93cb9c7e-d6d3-4b1e-8bcf-2a4a1d95a549",
		Name:         "Undercity Eliminator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Undercity Eliminator — you may sacrifice an artifact or creature; when you do, exile target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Undercity Eliminator — sacrifice an artifact or creature to exile target creature an opponent controls?",
						effects.Or(effects.Artifact(), effects.Creature()), false, undercityEliminatorSacrificed)
				}),
		},
	})
}

// undercityEliminatorSacrificed is "when you do": the reflexive trigger,
// made only after a sacrifice. Keyed: an on-disk identity, never renamed
// or reused.
var undercityEliminatorSacrificed = effects.RegisterSacrificedOne("prompt-run/undercity-eliminator-sacrificed", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Undercity Eliminator — exile target creature an opponent controls", undercityEliminatorExileBody).Apply(ctx)
})

// undercityEliminatorExileBody exiles the reflexive trigger's target.
var undercityEliminatorExileBody = game.ReflexiveBody("undercity-eliminator/exile",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		target := effects.FirstLegalBattlefieldTarget(ctx)
		if target == uuid.Nil {
			return nil
		}
		return effects.ExileTarget{Target: target}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
