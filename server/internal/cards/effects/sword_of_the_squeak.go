package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sword of the Squeak — Artifact — Equipment {2} (Bloomburrow Commander):
//
//	"Equipped creature gets +1/+1 for each creature you control with base
//	 power or toughness 1.
//	 Whenever a Hamster, Mouse, Rat, or Squirrel you control enters, you may
//	 attach this Equipment to that creature.
//	 Equip {2}"
//
// The pump counts on every layer recompute (PumpAttachedPer) the creatures its
// controller controls whose base power or toughness is 1 (CR 208.4b: the values
// after layer 7b, so a counter or an anthem neither adds to nor removes from
// the count). The equipped creature counts itself when its own base is 1.
//
// The attach is Thornbite Staff's: a "you may" trigger that attaches to the
// creature that entered, here only for one its controller controls. The four
// words are creature types of the Equipment's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ede28a7-57ea-47d6-a5fb-b0f061da2689",
		Name:         "Sword of the Squeak",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return cmdr2B21BasePTCount(g, source.Controller, func(p, t int) bool { return p == 1 || t == 1 })
			}),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventETB {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !c.IsCreature() || c.Controller != source.Controller {
					return false
				}
				for _, w := range []string{"Hamster", "Mouse", "Rat", "Squirrel"} {
					if c.HasSubtype(source.TextSubtype(w)) {
						return true
					}
				}
				return false
			},
			Key: "Sword of the Squeak — attach to the creature",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return attachSourceTo(g, item, item.Trigger.Event.CardID)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Sword of the Squeak — attach it to the creature?"},
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
