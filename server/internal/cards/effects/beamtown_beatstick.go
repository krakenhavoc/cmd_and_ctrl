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
// The trigger fires on combat damage to a player OR a battle (#3172):
// damage to a battle is the same EventDealDamage with a battle card as its
// target, so the condition accepts either.
func init() {
	Register(Spec{
		OracleID:     "035e33ab-adfe-4831-bdbc-52f4ddc7ffe8",
		Name:         "Beamtown Beatstick",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attachedCreatureDealtCombatDamageToPlayerOrBattle(ev, source, g)
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
