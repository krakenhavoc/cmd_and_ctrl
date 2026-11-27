package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pollen Lullaby — Instant {1}{W} (Commander 2011, #5574):
//
//	"Prevent all combat damage that would be dealt this turn. Clash with an
//	 opponent. If you win, creatures that player controls don't untap during
//	 the player's next untap step."
//
// The prevention is a Fog (CR 615) put in place first, as printed; the clash
// is the controller's choice of an opponent (CR 701.30b) and the rest waits
// for it. "That player" is the opponent the clash was with, read off the
// clash's won event (Actor the winner, Target the loser). Their creatures
// then skip the untap in that player's next untap step (a creature they gain
// afterwards stays tapped too, CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f64d70d-ea38-4419-be91-8b68aab3401e",
		Name:         "Pollen Lullaby",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.PreventAllCombatDamageThisTurn{Label: "Pollen Lullaby — prevent all combat damage this turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.ClashWithAnOpponent{Then: ctx.Tail(pollenLullabyClashed)}.Apply(ctx)
		},
	})
}

// pollenLullabyClashed is "If you win, creatures that player controls don't
// untap during the player's next untap step". An on-disk key.
var pollenLullabyClashed = effects.CardTail("then/pollen-lullaby-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) {
		return nil
	}
	events := ctx.Game.EventsThisTurn()
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Kind != game.EventClashWon || ev.Source != ctx.Source() || ev.Actor != ctx.Controller() || ev.Target == uuid.Nil {
			continue
		}
		return effects.PermanentsDontUntapNextStep{Player: ev.Target, Types: []string{"creature"},
			Label: "Pollen Lullaby — creatures don't untap during that player's next untap step"}.Apply(ctx)
	}
	return nil
})
