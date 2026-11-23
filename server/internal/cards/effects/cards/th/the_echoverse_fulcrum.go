package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Echoverse Fulcrum — Legendary Artifact {2} (Reality Fracture):
//
//	"When The Echoverse Fulcrum enters, draw a card, then discard a
//	 card.
//	 {5}, {T}, Exile The Echoverse Fulcrum: Destroy all creatures.
//	 Activate only as a sorcery."
//
// The loot is the usual draw-then-discard in printed order. The wipe is
// Wrath of God's body (creatures can regenerate; the card does not say
// otherwise), paid by exiling the artifact itself at announce.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c04197ad-d2f3-499f-b863-bed0c869519c",
		Name:         "The Echoverse Fulcrum",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Echoverse Fulcrum — draw a card, then discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}, {T}, Exile The Echoverse Fulcrum: Destroy all creatures. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.WrathDestroyAllCreatures(item, effects.NewContext(g, item))
			},
		}},
	})
}
