package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beamtown Beatstick — Artifact — Equipment for {R} (EDHREC rank
// 2566):
//
//	"Equipped creature gets +1/+0 and has menace.
//	 Whenever equipped creature deals combat damage to a player or
//	 battle, create a Treasure token.
//	 Equip {2}"
//
// A one-mana Equipment that grants its own evasion: menace is the
// keyword that makes the combat-damage trigger actually fire, so the
// two halves of the card are one plan.
//
// Menace is honoured by the block-legality check (a creature with it
// can't be blocked except by two or more), so the "+1/+0 and gets
// through" really is what happens.
//
// ONE SIMPLIFICATION, strictly weaker: the trigger fires on damage to
// a PLAYER only, not to a battle. The engine has battles (S27) but
// EventDealDamage's combat branch reaches players; damage to a battle
// does not route through the condition every other equipped-creature
// trigger in the catalog shares. Attacking a battle with the Beatstick
// makes no Treasure, which is a card that does less than printed.
func init() {
	Register(Spec{
		OracleID:     "035e33ab-adfe-4831-bdbc-52f4ddc7ffe8",
		Name:         "Beamtown Beatstick",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Combat damage to a battle makes no Treasure; only damage to a player does."},
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Beamtown Beatstick — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			})),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
