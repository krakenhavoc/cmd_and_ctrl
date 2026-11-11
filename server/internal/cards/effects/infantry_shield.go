package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infantry Shield — Artifact — Equipment {2}{R} (Tarkir: Dragonstorm
// Commander):
//
//	"Equipped creature has menace and mobilize X, where X is its power.
//	 (Whenever it attacks, create X tapped and attacking 1/1 red Warrior
//	 creature tokens. Sacrifice them at the beginning of the next end step.)
//	 Equip {2}"
//
// Menace is a layer-6 keyword grant. Mobilize is a granted triggered ability
// (ADR 0093, Black Mage's Rod's shape): it belongs to the equipped creature
// and goes with the Equipment, and the creature is its source, so the
// Warriors are made under its controller's control. The trigger is the
// shared MobilizeX (Avenger of the Fallen's counted X), counted as the attack
// trigger resolves from the equipped creature's power (CR 608.2h: its last
// known power if it has left), so a pump in response counts.
//
// No simplification.
func init() {
	const grantKey = "infantry-shield/mobilize-x"
	Register(Spec{
		OracleID:     "71c68002-1489-4858-a8fb-1f575641983e",
		Name:         "Infantry Shield",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{MobilizeX("This creature", infantryShieldX)},
			Text:      "Mobilize X, where X is this creature's power.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("menace"),
			GrantAbilitiesToAttached(grantKey),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// infantryShieldX is the equipped creature's power now, or as it last
// existed once it has left.
func infantryShieldX(g *game.Game, item *game.StackItem) int {
	return sourcePowerNow(NewContext(g, item))
}
