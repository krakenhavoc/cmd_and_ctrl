package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Port Razer — Creature — Orc Pirate {3}{R}{R}, 4/4 (Lost Caverns of
// Ixalan Commander, #5368):
//
//	"Whenever this creature deals combat damage to a player, untap each
//	 creature you control. After this phase, there is an additional
//	 combat phase.
//	 This creature can't attack a player it has already attacked this
//	 turn."
//
// Bloodthirster's pair with a wider untap: the trigger untaps every
// creature its controller controls as it resolves (Aurelia's
// untapAllYouControlThenExtraCombat) and adds a combat phase after the
// one in progress, with no main phase between. The
// restriction is the CR 508.1c clause Bloodthirster prints
// (CantAttackAPlayerItAlreadyAttackedThisTurn): it reads the turn's
// attack record for this object, so a Port Razer that left and returned
// is a new object (CR 400.7) with a clean slate, and planeswalkers and
// battles are not players and are never refused by it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "653e4b68-3184-4352-ac91-127f7e6d7c12",
		Name:         "Port Razer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CantAttackAPlayerItAlreadyAttackedThisTurn()},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Port Razer — untap each creature you control, additional combat phase", untapAllYouControlThenExtraCombat),
		},
	})
}
