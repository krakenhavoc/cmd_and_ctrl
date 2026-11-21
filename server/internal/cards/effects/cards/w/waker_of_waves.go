package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waker of Waves — Creature — Whale {5}{U}{U}, 7/7 (Foundations Jumpstart,
// S82, #5929):
//
//	"Creatures your opponents control get -1/-0.
//	 {1}{U}, Discard this card: Look at the top two cards of your library.
//	 Put one of them into your hand and the other into your graveyard."
//
// The static is Elesh Norn's opponents-only layer 7c modifier with power
// only. The hand-zone ability is Action News Crew's discard-self cost
// with Orazca Puzzle-Door's look-at-two body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a19a959-ea83-404b-89f7-e5dbf4cb7995",
		Name:         "Waker of Waves",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Creatures your opponents control get -1/-0.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Discard this card: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
			Cost:    game.AbilityCost{Mana: "{1}{U}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
					Max:    1,
					Label:  "Waker of Waves — put one of them into your hand and the other into your graveyard",
					Key:    effects.TakeRestGraveyard,
				}.Apply(ctx)
			},
		}},
	})
}
