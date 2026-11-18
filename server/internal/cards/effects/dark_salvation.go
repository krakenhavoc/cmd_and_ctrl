package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f5c2b700-da52-4436-b234-e32ba7759e17",
		Name:         "Dark Salvation",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets: Clauses(
			TargetPlayer("target player"),
			TargetCreature("up to one target creature").WithCount(0, 1),
		),
		OnResolve: ResolveSteps(Instruction(darkSalvationResolve)),
	})
}

func darkSalvationResolve(ctx *Context) error {
	player, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	if x := ctx.X(); x > 0 {
		if err := (CreateToken{Controller: player, Template: BlackZombieToken(), N: x}).Apply(ctx); err != nil {
			return err
		}
	}
	n := len(s82RprB31ZombiesYouControl(ctx, player))
	if n == 0 {
		return nil
	}
	for _, t := range ctx.ClauseTargets(1) {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: -n, Toughness: -n, Label: "Dark Salvation — -1/-1 for each Zombie"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
