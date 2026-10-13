package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pick Up the Pace — Enchantment {1}{R}:
//
//	"Whenever a creature you control that entered this turn attacks,
//	 exile the top card of your library.
//	 You may play cards exiled with this enchantment as long as you
//	 attacked this turn."
//
// Two things make the second sentence a static ability of the
// ENCHANTMENT rather than a grant that outlives it, and both are on the
// permission (#3192, ADR 0066's second 2026-10-13 amendment):
//
//   - "As long as you attacked this turn" is the conditional exile play
//     game.CondAttackedThisTurn: dormant, not lost, on a turn nothing
//     attacked.
//   - "Cards exiled with this enchantment" stops being true when the
//     enchantment leaves, so the permission is bound to it with a
//     for-as-long-as-it-remains-on-the-battlefield window
//     (CR 611.2b, game.WhileSourceOnBattlefield, the source-bound
//     duration the engine already had): once the enchantment has left,
//     or when it was already gone as the trigger resolved, the cards it
//     exiled are not playable, and a new Pick Up the Pace is a new
//     object with no claim on them (CR 400.7).
//
// The attacker "entered this turn" is read when the trigger fires, from
// the per-turn entry tally, and fires once per such attacker.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "873e739c-edf4-46a7-9024-5caaf87f507e",
		Name:         "Pick Up the Pace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && g.EnteredThisTurn(ev.CardID)
			}, "Pick Up the Pace — exile the top card of your library", func(g *game.Game, item *game.StackItem) error {
				window := forAsLongAsOnBattlefieldWindow(g, item.SourceCardID)
				return ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
					Window:    &window,
					Condition: attackedCondition(game.CondAttackedThisTurn),
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
