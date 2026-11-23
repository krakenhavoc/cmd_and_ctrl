package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Salvation — Sorcery {X}{X}{B}:
//
//	"Target player creates X 2/2 black Zombie creature tokens, then up
//	 to one target creature gets -1/-1 until end of turn for each Zombie
//	 that player controls."
//
// Two target clauses: a player, then up to one creature. The tokens are
// made for the targeted player, then the Zombies THAT player controls are
// counted (the ones they already had included) as the second half
// resolves. The creature target is optional, so the tokens are made
// without it, and with a player target that is gone the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5c2b700-da52-4436-b234-e32ba7759e17",
		Name:         "Dark Salvation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.TargetCreature("up to one target creature").WithCount(0, 1),
		),
		OnResolve: effects.ResolveSteps(effects.Instruction(darkSalvationResolve)),
	})
}

func darkSalvationResolve(ctx *effects.Context) error {
	player, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if x := ctx.X(); x > 0 {
		if err := (effects.CreateToken{Controller: player, Template: effects.BlackZombieToken(), N: x}).Apply(ctx); err != nil {
			return err
		}
	}
	n := len(effects.S82RprB31ZombiesYouControl(ctx, player))
	if n == 0 {
		return nil
	}
	for _, t := range ctx.ClauseTargets(1) {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: -n, Toughness: -n, Label: "Dark Salvation — -1/-1 for each Zombie"}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
