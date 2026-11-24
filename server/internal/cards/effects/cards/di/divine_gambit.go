package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Divine Gambit — Sorcery {W}{W} (Strixhaven Mystical Archive, #5503):
//
//	"Exile target artifact, creature, or enchantment an opponent
//	 controls. That player may put a permanent card from their hand
//	 onto the battlefield."
//
// "That player" is the target's controller as the spell resolves, read
// before the exile. The put runs from the exile's continuation, whether
// or not the permanent actually reached exile (a commander that took the
// command zone still left): the second sentence is not an "if you do".
// It is that player's own "may": they pick any permanent card in their
// hand, or none, and it enters under their control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25fcf381-6f0a-49ef-9c41-f202be7d8368",
		Name:         "Divine Gambit",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, creature, or enchantment an opponent controls",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()), effects.OpponentControls()),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			player, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			return effects.ExileTarget{Target: id, ThenTail: effects.WithCarry(ctx.ItemTail(divineGambitExiled), player)}.Apply(ctx)
		},
	})
}

// divineGambitExiled is the second sentence: that player may put a
// permanent card from their hand onto the battlefield.
var divineGambitExiled = effects.CardTail("divine-gambit/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TailCarry0(o)
	return effects.PutFromHandOntoBattlefield{
		Player:   player,
		Optional: true,
		Label:    "Divine Gambit — you may put a permanent card from your hand onto the battlefield",
	}.Apply(ctx)
})
