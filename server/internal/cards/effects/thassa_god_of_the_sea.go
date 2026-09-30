package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "69e5df2f-be1f-4608-a90f-3e2f51e2fea4",
		Name:            "Thassa, God of the Sea",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					devotionTo(g, source.Controller, "U") < 5
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				notACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Thassa, God of the Sea — scry 1", func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Target creature you control can't be blocked this turn.",
			Cost:    ManaCost("{1}{U}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{
					Target:       id,
					Restrictions: game.CantBeBlocked,
					Label:        "Thassa, God of the Sea — can't be blocked",
				}.Apply(ctx)
			},
		}},
	})
}
