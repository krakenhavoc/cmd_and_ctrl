package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ixidor, Reality Sculptor — Legendary Creature — Human Wizard {3}{U}{U},
// 3/4 (The List):
//
//	"Face-down creatures get +1/+1.
//	 {2}{U}: Turn target face-down creature face up."
//
// The anthem is a layer 7c static over every face-down creature, whoever
// controls it (a face-down permanent is a 2/2 creature from layer 1, so it
// becomes 3/3). The activation turns the target up as an effect, paying
// nothing (CR 708.8, TurnFaceUp; Hauntwoods Shrieker's seam); a target that
// is no longer face down, or that the rules refuse, is simply left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90b98aa7-c8d0-4853-887c-164f8160c3f3",
		Name:         "Ixidor, Reality Sculptor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Face-down creatures get +1/+1.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.FaceDownIsPermanent() && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Turn target face-down creature face up.",
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCreature("target face-down creature", FaceDown()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return TurnFaceUp{Targets: []uuid.UUID{t.ID}}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
