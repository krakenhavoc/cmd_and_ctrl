package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ferocity of the Hunt — Enchantment — Aura {1}{B/G} (Reality Fracture):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+0 and has deathtouch.
//	 When enchanted creature dies, return that card to the battlefield
//	 tapped under its owner's control."
//
// The dies trigger is Gift of Immortality's first half: the creature
// card comes back as a new object, tapped, under its OWNER's control.
// A token has no card to return, so it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "747a1b36-eb4f-44e6-9e5a-bb1a0259a471",
		Name:            "Ferocity of the Hunt",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 0),
			GrantToAttached("deathtouch"),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Ferocity of the Hunt — return that card to the battlefield tapped", func(g *game.Game, item *game.StackItem) error {
				_, err := g.ReturnToBattlefieldForEffect(item.Trigger.Event.CardID, uuid.Nil, true)
				if errors.Is(err, game.ErrCardNotFound) {
					return nil
				}
				return err
			}),
		},
	})
}
