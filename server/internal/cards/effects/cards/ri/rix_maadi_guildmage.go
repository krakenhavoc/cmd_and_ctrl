package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rix Maadi Guildmage — Creature — Human Shaman {B}{R}, 2/2 (Return to Ravnica,
// #5566):
//
//	"{B}{R}: Target blocking creature gets -1/-1 until end of turn.
//	 {B}{R}: Target player who lost life this turn loses 1 life."
//
// Both clauses are judged when chosen and again on resolution (CR 608.2b): a
// creature that stopped blocking, or a player who somehow has not lost life, is
// no longer a legal target. "Lost life" reads the turn's tally, so damage counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb88ada9-d3bf-4f0a-817a-c41f0ea9cc78",
		Name:         "Rix Maadi Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}{R}: Target blocking creature gets -1/-1 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
				Cost:    effects.ManaCost("{B}{R}"),
				Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Rix Maadi Guildmage — -1/-1 until end of turn"}.Apply(ctx)
				},
			},
			{
				Label:   "{B}{R}: Target player who lost life this turn loses 1 life.",
				Cost:    effects.ManaCost("{B}{R}"),
				Targets: effects.TargetPlayer("target player who lost life this turn", effects.OldB423PlayerLostLifeThisTurn()),
				Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}),
			},
		},
	})
}
