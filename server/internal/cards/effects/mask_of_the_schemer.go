package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of the Schemer — Artifact — Equipment {2}{U} (Streets of New
// Capenna Commander, #5491):
//
//	"Whenever equipped creature deals combat damage to a player, it
//	 connives X, where X is the amount of damage it dealt to that player.
//	 (Draw X cards, then discard X cards. Put a +1/+1 counter on that
//	 creature for each nonland card discarded this way.)
//	 Equip {2}"
//
// The trigger is the Swords' condition (attachedCreatureDealtCombatDamage
// ToPlayer); X is the amount on the triggering damage event (the damage
// actually dealt after prevention), and the equipped creature connives X
// through ConniveX.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "584f9c45-9737-47bb-8ba4-a7cd9b0b2c1f",
		Name:         "Mask of the Schemer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			},
			Key:    "Mask of the Schemer — equipped creature connives X",
			Effect: theEquippedCreatureConnivesForDamage,
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
