package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Poppet Stitcher // Poppet Factory — transforming double-faced card.
//
// Front, Creature — Human Wizard {2}{U}, 2/3:
//
//	"Whenever you cast an instant or sorcery spell, create a 2/2 black
//	 Zombie creature token with decayed.
//	 At the beginning of your upkeep, if you control three or more
//	 creature tokens, you may transform this creature."
//
// Back, Artifact:
//
//	"Creature tokens you control lose all abilities and have base power
//	 and toughness 3/3.
//	 At the beginning of your upkeep, you may transform this artifact."
//
// The front's upkeep trigger has an intervening "if" (CR 603.4): the
// token count is checked as the trigger would fire and again as it
// resolves. Both faces turn over in place (TransformThis, CR 701.27),
// so the permanent stays the same object.
//
// The Factory's static is two layer effects from one sentence: the
// ability loss in layer 6 (LosesAllAbilitiesAndHas with no grant, so
// decayed tokens lose decayed) and the base 3/3 in layer 7b. A token's
// counters and pumps still apply on top (7c, 7d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     PoppetStitcherOracle,
		Name:         "Poppet Stitcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Poppet Stitcher — create a 2/2 Zombie with decayed",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: 1})),
			effects.Optional(effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && creatureTokensControlledBy(g, source.Controller) >= 3
			}, "Poppet Stitcher — you may transform it", func(g *game.Game, item *game.StackItem) error {
				if creatureTokensControlledBy(g, item.Controller) < 3 {
					return nil
				}
				return effects.TransformThis{}.Apply(effects.NewContext(g, item))
			}), "Poppet Stitcher — transform it into Poppet Factory?"),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     PoppetStitcherOracle + "#1",
		Name:         "Poppet Factory",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.LosesAllAbilitiesAndHas(effects.CreatureTokensYouControl),
			{
				Label:     "Creature tokens you control have base power and toughness 3/3.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: effects.CreatureTokensYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 3
					c.Toughness = 3
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Poppet Factory — you may transform it", func(g *game.Game, item *game.StackItem) error {
				return effects.TransformThis{}.Apply(effects.NewContext(g, item))
			}), "Poppet Factory — transform it into Poppet Stitcher?"),
		},
	})
}

const PoppetStitcherOracle = "bc2d0702-b140-4b78-8a8f-3a6d19073b59"

// creatureTokensControlledBy counts the creature tokens `player`
// controls.
func creatureTokensControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsToken() && c.IsCreature() {
			n++
		}
	}
	return n
}
