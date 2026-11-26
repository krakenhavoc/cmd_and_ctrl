package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undercity Uprising — Sorcery {2}{B}{G} (Guilds of Ravnica, #5525):
//
//	"Creatures you control gain deathtouch until end of turn. Then target
//	 creature you control fights target creature you don't control. (Each
//	 deals damage equal to its power to the other.)"
//
// The deathtouch goes to the creatures you control as the spell resolves
// (CR 611.2c), so the fight is the second instruction and sees it; the fight
// happens only while both targets are still legal (CR 701.14b, 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53e540b5-8338-4b21-a61a-c5b7fe652a6e",
		Name:         "Undercity Uprising",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"deathtouch"},
				Label:    "Undercity Uprising — deathtouch until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.FightTargetClauses(ctx)
		},
	})
}
