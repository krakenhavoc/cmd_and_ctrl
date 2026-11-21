package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neglected Heirloom // Ashmouth Blade — Artifact — Equipment {1} // Artifact —
// Equipment (Innistrad Remastered, #5948):
//
//	Neglected Heirloom: "Equipped creature gets +1/+1.
//	 When equipped creature transforms, transform this Equipment.
//	 Equip {1}"
//	Ashmouth Blade: "Equipped creature gets +3/+3 and has first strike.
//	 Equip {3}"
//
// A double-faced Equipment, the back face its own catalog row (oracle ID plus
// "#1"). The front's trigger watches the creature it is attached to turning
// over (EventTransform, CR 701.28) and turns the Equipment over too; it stays
// attached (CR 701.3c: a transformed Equipment is still an Equipment). The back
// has no such trigger, and no way to turn back.
//
// No simplification.
func init() {
	const oracle = "1374498c-58fd-40b9-b63b-f87fedc42e58"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Neglected Heirloom",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventTransform, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventTransform && source.IsAttachedTo(ev.CardID)
			}, "Neglected Heirloom — transform this Equipment", Do(TransformThis{})),
		},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Ashmouth Blade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("first strike"),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
