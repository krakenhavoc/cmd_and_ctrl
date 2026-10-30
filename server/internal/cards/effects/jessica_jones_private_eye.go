package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jessica Jones, Private Eye — Legendary Creature — Human Detective Hero
// {2}{R}, 2/3 (Marvel Super Heroes, #4455):
//
//	"{T}, Put a stun counter on Jessica Jones: Exile the top X cards of
//	 your library, where X is Jessica Jones's power. You may play those
//	 cards this turn. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// The stun counter is a cost (AddCounterToThis), put on as she taps, so
// she stays tapped through the next untap step. X is her power as the
// ability resolves, or her last known power if she has left (CR 608.2h).
// The cards are exiled with the this-turn play permission (a land may be
// played as the land drop), the impulse-draw shape Daredevil uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6677cd5a-0785-499b-ba99-f6bdcfb196d5",
		Name:         "Jessica Jones, Private Eye",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Put a stun counter on Jessica Jones: Exile the top X cards of your library, where X is Jessica Jones's power. You may play those cards this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), AddCounterToThis(game.CounterStun, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := sourcePowerNow(NewContext(g, item))
				if x <= 0 {
					return nil
				}
				_, err := b12ImpulseExileForTurn(g, item, x)
				return err
			},
		}},
	})
}
