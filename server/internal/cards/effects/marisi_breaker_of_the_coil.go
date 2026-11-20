package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marisi, Breaker of the Coil — Legendary Creature — Cat Warrior
// {1}{R}{G}{W}, 5/4 (Commander 2019, #5517):
//
//	"Your opponents can't cast spells during combat.
//	 Whenever a creature you control deals combat damage to a player, goad
//	 each creature that player controls. (Until your next turn, those
//	 creatures attack each combat if able and attack a player other than
//	 you if able.)"
//
// The cast restriction is Megatron, Tyrant's: every step from the beginning
// of combat to the end of combat, on any player's turn. The trigger is
// Toski's condition (combat damage to a player by a creature you control)
// with Komainu Battle Armor's body: the player is the one the damage was
// dealt to, "each creature that player controls" is read as the trigger
// resolves, and the goad is Marisi's controller's (CR 701.15). Each
// creature that connects triggers it once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17475cf0-661d-4361-827b-1e4d638089f1",
		Name:         "Marisi, Breaker of the Coil",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Marisi — your opponents can't cast spells during combat.",
			Forbids: func(q game.CastQuery) bool {
				return q.Source.Controller != q.Controller && inCombatStep(q.Game.Turn.Step)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Marisi, Breaker of the Coil — goad each creature that player controls", komainuBattleArmorGoad),
		},
	})
}
