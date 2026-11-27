package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crown of Convergence — Artifact {2} (Ravnica: City of Guilds, S82 OLD batch
// 575, #5600):
//
//	"Play with the top card of your library revealed.
//	 As long as the top card of your library is a creature card, creatures you
//	 control that share a color with that card get +1/+1.
//	 {G}{W}: Put the top card of your library on the bottom of your library."
//
// The first line is the engine's revealed-top flag (LibraryTopRevealed). The
// anthem is a layer 7c static that reads the top card as the layers are
// computed (DependsOnLibraryTop makes the table recompute when the top card
// changes with no event): it applies only while the top card is a creature
// card, to each creature its controller has that shares a color with it
// (colorless shares none). The activated ability moves the top card to the
// bottom of its owner's library, which is the top card changing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "316c265b-67e2-48b3-aac8-07a898dbb20f",
		Name:              "Crown of Convergence",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		Static: []game.StaticAbility{{
			Label:               "As long as the top card of your library is a creature card, creatures you control that share a color with that card get +1/+1.",
			Layer:               game.Layer7PT,
			SubLayer:            game.SubLayer7C_Modify,
			DependsOnLibraryTop: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !target.IsCreature() || target.Controller != source.Controller {
					return false
				}
				top, ok := effects.OldB483TopCardOfYourLibrary(g, source.Controller)
				return ok && top.IsCreature() && effects.SharesAColorNow(*target, top)
			},
			Apply: effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}: Put the top card of your library on the bottom of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				p := g.PlayerByIDForEffect(ctx.Controller())
				if p == nil || p.Library == nil {
					return nil
				}
				top, err := p.Library.Top()
				if err != nil {
					return nil
				}
				return g.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, []uuid.UUID{top.InstanceID})
			},
		}},
	})
}
