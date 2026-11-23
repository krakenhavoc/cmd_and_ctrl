package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kirk, Commanding Officer — Legendary Creature — Human Officer {1}{R}{W},
// 3/3 (Star Trek, #4973):
//
//	"Whenever you attack while creatures you control have total power 8
//	 or greater, exile the top card of your library. You may play that
//	 card this turn.
//	 {2}, Sacrifice Kirk: Destroy target artifact or enchantment."
//
// The attack trigger is Hologram Janeway's: once per declaration however
// many creatures attack (OncePerBatch), and "while" is a condition of the
// trigger event, read as attackers are declared over every creature you
// control (creaturesYouControlTotalPowerAtLeast), not an intervening "if"
// (CR 603.4 needs the word "if"). Its body is Breeches, Eager Pillager's
// impulse bullet: play, not cast, so a land may be played from exile, and
// the grant lasts this turn.
//
// The second ability is Cathar Commando's: a mana-plus-sacrifice-self
// cost paid at announce and a real target clause, destroyed through the
// ordinary destroy path.
//
// No simplification.
func init() {
	totalPowerEight := effects.CreaturesYouControlTotalPowerAtLeast(8)
	effects.Register(effects.Spec{
		OracleID:     "e85ad29a-ebd6-4f81-af99-556deaedbfe0",
		Name:         "Kirk, Commanding Officer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && totalPowerEight(g, source.Controller)
			}, "Kirk, Commanding Officer — exile the top card of your library; you may play it this turn",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTopCardYouMayPlayThisTurn(ctx.Item, ctx, 0) })))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice Kirk: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
