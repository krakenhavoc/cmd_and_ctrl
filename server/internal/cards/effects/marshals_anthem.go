package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marshal's Anthem — Enchantment {2}{W}{W} (Outlaws of Thunder Junction
// Commander):
//
//	"Multikicker {1}{W}
//	 Creatures you control get +1/+1.
//	 When this enchantment enters, return up to X target creature cards from
//	 your graveyard to the battlefield, where X is the number of times this
//	 enchantment was kicked."
//
// The anthem is Glorious Anthem's layer 7c static. The multikicker count is
// read off the record the resolution stamped onto the permanent
// (game.CardKickedTimes, CR 702.33d), which is what sets how many targets
// the enters trigger takes (TargetsFrom): up to X, so none at all when it
// was not kicked or was put onto the battlefield without being cast. Each
// still-legal pick returns in one simultaneous entry. The multikicker cap
// of 20 is the engine's, not the card's.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Marshal's Anthem — return up to X target creature cards from your graveyard to the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			return returnLegalGraveyardTargetsToBattlefield(NewContext(g, item))
		})
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		n := game.CardKickedTimes(*source)
		if n <= 0 {
			return nil
		}
		return TargetCardInGraveyard("up to X target creature cards from your graveyard, where X is the number of times this enchantment was kicked",
			YouOwn(), Creature()).WithCount(0, n)
	}
	Register(Spec{
		OracleID:      "41fb2b0e-d0a0-4e47-a6f3-fc9860ade65a",
		Name:          "Marshal's Anthem",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Multikicker("{1}{W}", 20)},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Creatures you control get +1/+1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{enters},
	})
}
