package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brazen Cannonade — Enchantment {3}{R}:
//
//	"Whenever an attacking creature you control dies, this enchantment
//	 deals 2 damage to each opponent.
//	 Raid — At the beginning of each of your postcombat main phases, if
//	 you attacked this turn, exile the top card of your library. Until end
//	 of combat on your next turn, you may play that card."
//
// The first trigger watches each death of a creature that was attacking
// under its controller's control as it died (last-known information, CR
// 603.10a), and deals 2 damage to each opponent from this enchantment.
//
// Raid is an intervening "if" (CR 603.4): the trigger fires only at a
// postcombat main phase of a turn in which you attacked, and checks again
// as it resolves. "Each of your postcombat main phases" fires again in an
// added postcombat main phase. The exiled card may be played (a land
// included) until the end of combat on your next turn, the #6093 duration
// (game.UntilEndOfCombatOnYourNextTurn): through the rest of this turn,
// the other players' turns, and the beginning, first main phase and combat
// of your next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9212658-8406-489f-80d6-7af42df8956c",
		Name:         "Brazen Cannonade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedWhileAttacking(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller
			}, "Brazen Cannonade — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
			effects.InterveningIfYou("you attacked this turn", effects.B18AttackedThisTurn,
				effects.AtYourPostcombatMain("Brazen Cannonade — exile the top card of your library; you may play it until end of combat on your next turn",
					brazenCannonadeExile)),
		},
	})
}

// brazenCannonadeExile is the raid trigger's body.
func brazenCannonadeExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	d := effects.DurationUntilEndOfCombatOnYourNextTurn(ctx, item.Controller)
	return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, Window: &d}.Apply(ctx)
}
