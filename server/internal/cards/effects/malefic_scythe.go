package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malefic Scythe — Artifact — Equipment {1}{B} (Jumpstart, #4892):
//
//	"This Equipment enters with a soul counter on it.
//	 Equipped creature gets +1/+1 for each soul counter on this Equipment.
//	 Whenever equipped creature dies, put a soul counter on this Equipment.
//	 Equip {1}"
//
// The entry counter is a self-replacement (CR 614.1c). The bonus is read
// live from the Equipment's counters on every layer pass. The dies
// trigger is checked against the attachment as the creature died, and
// the counter lands on the Scythe even though it is no longer attached
// by then; if the Scythe has left the battlefield, nothing happens
// (CR 400.7).
//
// No simplification.
func init() {
	const soul = "soul"
	Register(Spec{
		OracleID:     "0b4925d7-e2bb-4b74-a538-80fe759799ad",
		Name:         "Malefic Scythe",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(soul, 1, "Malefic Scythe: enters with a soul counter"),
		},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int {
				return source.Counters[soul]
			}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return equippedCreatureDied(ev, source)
			}, "Malefic Scythe — put a soul counter on this Equipment",
				Do(CounterOnThis{Kind: soul, N: 1})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}
