package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disturbing Conversion — Enchantment — Aura {1}{U} (March of the Machine,
// S82 MOM batch 13):
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, each player mills two cards.
//	 Enchanted creature gets -X/-0, where X is the number of cards in its
//	 controller's graveyard."
//
// The mill is every player's, in APNAP order (CR 101.4). The -X/-0 is a layer 7c modify read live on the host's
// CONTROLLER's graveyard (not the Aura's), so it moves as that graveyard
// fills or empties and follows the creature if it changes hands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "adde6737-516c-4d6f-9363-388ad37d4114",
		Name:            "Disturbing Conversion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets -X/-0, where X is the number of cards in its controller's graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				if p := g.PlayerByIDForEffect(target.Controller); p != nil && p.Graveyard != nil {
					c.Power -= len(p.Graveyard.Cards)
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disturbing Conversion — each player mills two cards", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range apnapPlayers(g) {
					if err := (MillCards{Player: id, N: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
