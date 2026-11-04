package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doc Ock's Tentacles — {1} Artifact — Equipment (Marvel's Spider-Man
// promos, #4460):
//
//	"Whenever a creature you control with mana value 5 or greater
//	 enters, you may attach this Equipment to it.
//	 Equipped creature gets +4/+4.
//	 Equip {5}"
//
// Hero's Blade's shape with the legendary test swapped for mana value
// (CR 202.3), read off the creature as it entered. The attach is an
// optional trigger, free and at any speed; it does nothing if the
// creature or the Tentacles left before it resolves (AttachSourceForEffect).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5a18764-eb29-4bd8-a2e9-0ce55f4f6c88",
		Name:         "Doc Ock's Tentacles",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(4, 4)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.ManaValue() >= 5
			},
			Key: "Doc Ock's Tentacles — attach to the creature with mana value 5 or greater",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Doc Ock's Tentacles — attach to it?"},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{5}"),
		},
	})
}
