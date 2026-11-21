package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worf, Defiant Commander — Legendary Creature — Klingon Warrior
// Officer {2}{R}{W}{B}, 3/4 (Star Trek Commander, #4974):
//
//	"When Worf enters, create a 2/1 red Klingon Warrior creature token
//	 with haste.
//	 Battlefront — Whenever you attack, exile the top X cards of your
//	 library, where X is the number of opponents you attacked this
//	 turn. Until the end of your next turn, you may play those cards."
//
// "Battlefront" is an ability word. "Whenever you attack" is one
// trigger per declare-attackers batch (Bard, Heir of Girion's shape).
// X is read as the trigger resolves and counts every opponent you
// attacked this turn (OpponentsYouAttackedThisTurn): an attack on a
// planeswalker or battle is not an attack on its player, and a second
// combat adds the opponents it attacks to those already counted. The
// play permission is ADR 0063's "until the end of your next turn".
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7fcdbab-491f-40d9-be40-409af0d6c3f5",
		Name:         "Worf, Defiant Commander",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Worf, Defiant Commander — create a 2/1 red Klingon Warrior with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 red Klingon Warrior with haste"), N: 1})),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Worf, Defiant Commander — exile the top X cards; you may play them until the end of your next turn",
				worfDefiantCommanderBattlefront)),
		},
	})
}

// worfDefiantCommanderBattlefront exiles the top X cards of the
// controller's library with a play permission lasting until the end of
// their next turn, X the opponents they attacked this turn.
func worfDefiantCommanderBattlefront(g *game.Game, item *game.StackItem) error {
	n := effects.OpponentsYouAttackedThisTurn(g, item.Controller)
	if n <= 0 {
		return nil
	}
	_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, n, game.CastPermission{
		Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
	})
	return err
}
