package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chainer, Nightmare Adept — Legendary Creature — Human Minion {2}{B}{R},
// 3/2 (Modern Horizons 2, #5501):
//
//	"Discard a card: You may cast a creature spell from your graveyard this
//	 turn. Activate only once each turn.
//	 Whenever a nontoken creature you control enters, if you didn't cast it
//	 from your hand, it gains haste until your next turn."
//
// The activated ability opens one standing, end-of-turn cast permission over
// your graveyard for a creature spell (Yawgmoth's Will's shape, narrowed to
// creatures and to one cast), paying the spell's own cost; once each turn is
// a per-object tally (CR 602.5b). The trigger is judged on the entry: a
// creature that was reanimated, put in by an effect or cast from anywhere
// but its controller's hand qualifies, a creature cast from hand does not
// (the permanent's cast provenance, CR 400.7d). Those facts cannot change
// afterwards, so the "if" is judged as the trigger goes on the stack
// (CR 603.4). Chainer itself is a nontoken creature you control.
//
// No simplification.
func init() {
	const label = "Discard a card: You may cast a creature spell from your graveyard this turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "14f6fcb0-bed4-4a38-b7df-1e438d81bedb",
		Name:         "Chainer, Nightmare Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.DiscardACard(),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.GrantCastPermissionForEffect(game.CastPermission{
					Player:    item.Controller,
					Zone:      game.ZoneGraveyard,
					Scope:     game.ScopeStanding,
					Filter:    game.PermissionFilter{CreatureOnly: true},
					CastsLeft: 1,
					Source:    item.SourceCardID,
					Label:     "Chainer, Nightmare Adept — cast a creature spell from your graveyard",
				})
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.OldB032NontokenCreatureNotCastFromHand,
				"Chainer, Nightmare Adept — it gains haste until your next turn", effects.OldB032EnteredCreatureGainsHaste),
		},
	})
}
