package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wreck and Rebuild — Sorcery {1}{R}{G} (Outlaws of Thunder Junction Commander,
// #5368):
//
//	"Choose one —
//	 • Destroy target artifact or enchantment.
//	 • Mill five cards, then you may put a land card from your graveyard onto
//	   the battlefield tapped.
//	 Flashback {3}{R}{G}"
//
// A two-bullet modal sorcery (Abrade's shape). The mill bullet picks after the
// mill (CR 608.2c), so a land just milled is a legal pick, from the whole
// graveyard, optionally; it enters tapped through the CR 614 entry pipeline.
// Flashback is the ordinary alternative cost: cast from the graveyard, exiled
// as it leaves the stack, and either bullet may be chosen again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7e10c141-dfb0-409b-b8c9-cdfe93b42a06",
		Name:             "Wreck and Rebuild",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove, Lands: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{R}{G}")},
		Modes: ChooseOne(
			Mode("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment()))),
			Mode("Mill five cards, then you may put a land card from your graveyard onto the battlefield tapped.", nil),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return Do(MillCards{N: 5}, Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Wreck and Rebuild — you may put a land card from your graveyard onto the battlefield tapped",
					Match:    game.Card.IsLand,
					Max:      1,
					Tapped:   true,
				}.Apply(ctx)
			}))(ctx.Game, item)
		},
	})
}
