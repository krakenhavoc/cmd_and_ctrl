package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evershrike — {3}{W/B}{W/B} Creature — Elemental Spirit, 2/2:
//
//	"Flying
//	 This creature gets +2/+2 for each Aura attached to it.
//	 {X}{W/B}{W/B}: Return this card from your graveyard to the battlefield. You
//	 may put an Aura card with mana value X or less from your hand onto the
//	 battlefield attached to it. If you don't, exile this creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9f148f1f-53c5-477d-ba91-a2415db65adf",
		Name:            "Evershrike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		XMatters:        true,
		Static: []game.StaticAbility{{
			Label:    "Gets +2/+2 for each Aura attached to it.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B18AurasAttachedTo(g, target.InstanceID)
				c.Power += 2 * n
				c.Toughness += 2 * n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{W/B}{W/B}: Return this card from your graveyard to the battlefield. You may put an Aura card with mana value X or less from your hand onto the battlefield attached to it. If you don't, exile this creature.",
			Cost:    effects.ManaCost("{X}{W/B}{W/B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.S82OldB524EvershrikeReturn(g, item, ctx.X())
			},
		}},
	})
}
