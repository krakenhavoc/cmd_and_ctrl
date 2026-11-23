package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thassa, God of the Sea — Legendary Enchantment Creature — God
// {2}{U}, 5/5 (EDHREC rank 731):
//
//	"Indestructible
//	 As long as your devotion to blue is less than five, Thassa isn't
//	 a creature.
//	 At the beginning of your upkeep, scry 1.
//	 {1}{U}: Target creature you control can't be blocked this turn."
//
// The same shape as Thassa, Deep-Dwelling and Heliod, Sun-Crowned:
// Indestructible rides PrintedKeywords, the devotion-gated "isn't a
// creature" clause is the shared Layer 4 self-static reading
// devotionTo blue and calling notACreature. The upkeep trigger is a
// plain AtYourUpkeep scry; the activated ability is Rogue's Passage's
// can't-be-blocked restriction, narrowed to a creature the controller
// controls (the printed clause, unlike Rogue's Passage's "any
// creature").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69e5df2f-be1f-4608-a90f-3e2f51e2fea4",
		Name:            "Thassa, God of the Sea",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Label: "Isn't a creature unless your devotion to blue is five or more.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					effects.DevotionTo(g, source.Controller, source.TextColor("U")) < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.NotACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Thassa, God of the Sea — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Target creature you control can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.RestrictUntilEOT{
					Target:       id,
					Restrictions: game.CantBeBlocked,
					Label:        "Thassa, God of the Sea — can't be blocked",
				}.Apply(ctx)
			},
		}},
	})
}
