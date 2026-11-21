package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terisian Mindbreaker — Artifact Creature — Juggernaut {7}, 6/4 (The
// Brothers' War, #5482):
//
//	"Whenever this creature attacks, defending player mills half their
//	 library, rounded up.
//	 Unearth {1}{U}{U}{U} (…)"
//
// The defending player is read off the attack as it is declared (attacking a
// planeswalker or battle mills the player behind it, as Agate-Blade Assassin's
// drain does); the half is read from their library as the ability resolves,
// rounded up (s82Bro13MillHalfRoundedUp). Unearth is the shared keyword
// (unearth.go).
//
// No simplification.
func init() {
	const label = "Terisian Mindbreaker — defending player mills half their library, rounded up"
	effects.Register(effects.Spec{
		OracleID:     "d16fcb45-7fe7-4640-a325-badb64cadd62",
		Name:         "Terisian Mindbreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: effects.ThisAttacked,
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = effects.B17DefendingPlayer(g, ev)
				return item
			},
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				defender := ctx.Item.Params.Player
				if defender == uuid.Nil || !effects.DefendingPlayerStillIn(ctx.Game, defender) {
					return nil
				}
				return effects.S82Bro13MillHalfRoundedUp(ctx, defender)
			})),
		}},
		Activated: []effects.ActivatedAbility{effects.Unearth("{1}{U}{U}{U}")},
	})
}
