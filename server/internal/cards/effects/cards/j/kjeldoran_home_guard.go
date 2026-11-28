package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kjeldoran Home Guard — Creature — Human Soldier {3}{W} 1/6 (Alliances, S82
// OLD batch 819):
//
//	"At end of combat, if this creature attacked or blocked this combat, put a
//	 -0/-1 counter on this creature and create a 0/1 white Deserter creature
//	 token."
//
// An intervening "if" (CR 603.4): rechecked as the ability resolves. The
// counter goes on first, then the token is made whether or not the counter
// could be.
func init() {
	const label = "Kjeldoran Home Guard — put a -0/-1 counter on it and create a 0/1 white Deserter creature token"
	effects.Register(effects.Spec{
		OracleID:     "5f7ae611-a78d-4535-9d76-755001bda34c",
		Name:         "Kjeldoran Home Guard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "this creature attacked or blocked this combat",
				Test: func(ctx *effects.Context) bool {
					ref, ok := ctx.SourceRef()
					return ok && effects.OldB819WasInCombat(ctx.Game, ref)
				},
			}, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEndCombat, false),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.OldB819WasInCombat(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
				}),
				label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsStillThisPermanent(g, item) {
						if err := (effects.AddCounter{Target: item.SourceCardID, Kind: "-0/-1", N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CreateToken{Template: game.Card{
						Name: "Deserter", TypeLine: "Token Creature — Deserter", Colors: []string{"W"}, Power: 0, Toughness: 1,
					}, N: 1}.Apply(ctx)
				})),
		},
	})
}
