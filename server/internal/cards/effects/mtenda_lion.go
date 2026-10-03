package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mtenda Lion — Creature — Cat {G}, 2/1:
//
//	"Whenever this creature attacks, defending player may pay {U}. If that player does, prevent all combat damage that would be dealt by this creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the defending player is the one
// the Lion attacks (or who controls the planeswalker or protects the
// battle it attacks), asked as the trigger resolves. The payment is held
// in the declare attackers step (MayPay's InThisStep, Hellkite Charger's
// anchor), so the answer always comes before combat damage. Paid, the
// Lion's combat damage this turn is prevented, the Lion being the
// shield's source while it is the same object (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "af029853-cfb0-403b-af51-141ba02ae2e4",
		Name:         "Mtenda Lion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			}, "Mtenda Lion — defending player may pay {U} to prevent its combat damage",
				func(g *game.Game, item *game.StackItem) error {
					defender := g.DefendingPlayerForAttackerForEffect(item.SourceCardID)
					if g.PlayerByIDForEffect(defender) == nil {
						return nil
					}
					return MayPay{
						Chooser:    defender,
						Cost:       "{U}",
						Question:   "Mtenda Lion — pay {U} to prevent all combat damage it would deal this turn?",
						InThisStep: true,
						OnPay:      shieldAgainstThisCombatDamage,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
