package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drossclaw — Artifact — Equipment {1}{B} (Modern Horizons 3, #5948):
//
//	"Living weapon (When this Equipment enters, create a 0/0 black Phyrexian
//	 Germ creature token, then attach this to it.)
//	 Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, each opponent loses 1 life.
//	 Equip {2}"
//
// The shared LivingWeapon trigger, the flat +1/+1 and one attack trigger on
// the creature the Equipment is attached to (attachedCreatureAttacked): each
// attack by it drains every opponent for 1, the life lost as the trigger
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cda1ac77-3986-4600-9835-b0c0063ecc2f",
		Name:         "Drossclaw",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			LivingWeapon("Drossclaw"),
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Drossclaw — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
		},
		Static:    []game.StaticAbility{PumpAttached(1, 1)},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
