package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Downdraft — {2}{G} Enchantment (Weatherlight, S82 OLD batch 771, #5629):
//
//	"{G}: Target creature loses flying until end of turn.
//	 Sacrifice this enchantment: It deals 2 damage to each creature with flying."
//
// The first ability removes flying from the target until end of turn
// (Emerald Charm's mode), if it is still legal as the ability resolves (CR
// 608.2b). The second sacrifices the enchantment as a cost, so it is gone while
// the ability is on the stack; the Downdraft then deals 2 damage to each creature
// with flying as it resolves, as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2508190-5aca-4511-adc8-2072cd228119",
		Name:         "Downdraft",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{G}: Target creature loses flying until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
				Cost:    effects.ManaCost("{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.UntilEndOfTurn(ctx, t.ID, nil, "Downdraft — loses flying until end of turn", game.RemoveKeywordsMod("flying"))
					}
					return nil
				},
			},
			{
				Label: "Sacrifice this enchantment: It deals 2 damage to each creature with flying.",
				Purpose: game.Purpose{
					Answers: game.AnswerRemove,
					Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true},
				},
				Cost: effects.SacrificeThis(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.HasKeyword("flying")), 2)
				},
			},
		},
	})
}
