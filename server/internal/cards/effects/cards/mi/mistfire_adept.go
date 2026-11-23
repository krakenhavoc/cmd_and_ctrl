package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistfire Adept — Creature — Human Monk {3}{U}, 3/3 (S82 GAP, #5948):
//
//	"Prowess
//	 Whenever you cast a noncreature spell, target creature gains flying
//	 until end of turn."
func init() {
	const label = "Mistfire Adept — target creature gains flying until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "794255b1-d94d-4de4-ac0a-87ef743f9102",
		Name:            "Mistfire Adept",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: item.Targets[0].ID, Keywords: []string{"flying"}, Label: label}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
