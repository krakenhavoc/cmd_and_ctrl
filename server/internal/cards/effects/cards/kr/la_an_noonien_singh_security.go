package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// La'An Noonien-Singh, Security — Legendary Creature — Human Soldier
// {2}{R}{R}, 4/3 (Star Trek, #4973):
//
//	"Whenever you attack while creatures you control have total power 8 or
//	 greater, target creature can't block this turn."
//
// "Whenever you attack" is once per attack declaration however many
// creatures attack (OncePerBatch, The Chief Warg's shape), and an attack
// at a planeswalker or a battle counts: the text names no player (CR
// 508.3e is about "attacks [another player]"). The total is the current
// power of every creature you control, attacking or not, read through the
// layers as attackers are declared. It is a condition of the trigger, not
// an intervening "if", so it is not checked again on resolution. The
// creature chosen gains "can't block" until the turn ends.
//
// No simplification.
func init() {
	const label = "La'An Noonien-Singh — target creature can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "ed3b80eb-8915-4428-89b8-90f4278f2733",
		Name:         "La'An Noonien-Singh, Security",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && effects.B43TotalPowerControlled(g, source.Controller) >= 8
			}, label, effects.TargetCreatureCantBlockThisTurn(label))), effects.TargetCreature("target creature")),
		},
	})
}
