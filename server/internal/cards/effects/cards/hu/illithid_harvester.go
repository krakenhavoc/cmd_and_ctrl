package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illithid Harvester // Plant Tadpoles — an Adventure creature (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	Illithid Harvester {4}{U}, Creature — Horror, 4/4
//	  "Ceremorphosis — When this creature enters, turn any number of target
//	   tapped nontoken creatures face down. They're 2/2 Horror creatures."
//	Plant Tadpoles {X}{U}{U}, Sorcery — Adventure
//	  "Tap X target creatures. They don't untap during their controllers'
//	   next untap steps."
//
// Ceremorphosis turns each still-legal target face down (Cyber Conversion's
// shape: TurnFaceDown with a listed body, CR 708.2), the listing being a 2/2
// Horror creature with nothing else (CR 708.2 replaces the default body, so
// it is nameless and textless). "Any number" is a 0/0 clause; "tapped" and
// "nontoken" are checked as the trigger goes on the stack and again as it
// resolves, so a target that untapped in response is simply not turned over
// (CR 608.2b). Plant Tadpoles is Rooftop Nuisance's TapAndFreeze over X
// targets (TargetsCountedByX), one hold per target.
//
// No simplification.
func init() {
	const oracleID = "299ffd3a-29cf-4667-b191-dd30d09f799c"
	const label = "Illithid Harvester — turn any number of target tapped nontoken creatures face down; they're 2/2 Horror creatures"
	nontoken := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !effects.IsToken(c) }
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Illithid Harvester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							ids = append(ids, t.ID)
						}
					}
					return effects.TurnFaceDown{
						Source:  item.SourceCardID,
						Targets: ids,
						Listed: &game.FaceDownListing{
							Types:     []string{"Creature"},
							Subtypes:  []string{"Horror"},
							Power:     2,
							Toughness: 2,
						},
					}.Apply(ctx)
				}),
				effects.TargetCreature("any number of target tapped nontoken creatures",
					effects.TappedPermanent(), nontoken).WithCount(0, 0)),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Plant Tadpoles",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetsCountedByX(effects.TargetCreature("X target creatures")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return effects.TapAndFreeze{Targets: ids, Label: "Plant Tadpoles — they don't untap during their controllers' next untap steps"}.Apply(ctx)
		},
	})
}
