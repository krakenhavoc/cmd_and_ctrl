package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overburden — Enchantment {1}{U} (Prophecy, S82 OLD batch 674, #5617):
//
//	"Whenever a player puts a nontoken creature onto the battlefield, that
//	 player returns a land they control to its owner's hand."
//
// Genesis Chamber's watch over every nontoken creature that enters, whoever's
// it is (a cast, a reanimation, a flicker's return). "That player" is the one
// who put it there, the entry's PutBy (Thelon's Chant's chantPutter), not
// necessarily its controller. As the trigger resolves that player picks a land
// they control, which returns to its owner's hand; a player with no land has
// nothing to return. The land may be the one just played, and a land that is
// a creature is itself a nontoken creature that triggers it.
//
// No simplification.
func init() {
	const label = "Overburden — that player returns a land they control to its owner's hand"
	effects.Register(effects.Spec{
		OracleID:     "8104e454-77f5-4d38-894e-41ed0d74c17c",
		Name:         "Overburden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B26NontokenCreatureEntered(ev, g)
				return ok
			}, label, func(g *game.Game, item *game.StackItem) error {
				player := effects.S82OldB674EnteringCreaturePutter(g, item)
				if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
					return nil
				}
				return effects.S82OldB674ReturnALandYouControl(player,
					"Overburden — return a land you control to its owner's hand", overburdenReturned).Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// overburdenReturned returns the chosen land to its owner's hand, keyed. An
// on-disk key: never renamed, never reused.
var overburdenReturned = effects.RegisterPermanentsPick("prompt-run/s82-old-b674-overburden-return-land", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
