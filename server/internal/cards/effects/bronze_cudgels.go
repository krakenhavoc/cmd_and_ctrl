package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bronze Cudgels — Artifact — Equipment {1} (Kamigawa: Neon Dynasty,
// #5493):
//
//	"{2}: Until end of turn, equipped creature gets +X/+0, where X is the
//	 number of times this ability has resolved this turn.
//	 Equip {1}"
//
// The count is Belladonna Took's: the per-turn tally of this Equipment's
// resolutions of this ability, which includes the one resolving
// (Game.ResolvedThisTurn), per object (CR 400.7). Each resolution gives
// the creature equipped at that moment a separate +X/+0 until end of turn
// (CR 611.2a), so the second resolution adds +2 on top of the first's +1.
// With nothing equipped when it resolves the ability does nothing, but it
// still counts as having resolved.
//
// No simplification.
const bronzeCudgelsLabel = "{2}: Until end of turn, equipped creature gets +X/+0, where X is the number of times this ability has resolved this turn."

func init() {
	Register(Spec{
		OracleID:     "82ebd4fa-9db1-4723-8305-8b8a9ce958e2",
		Name:         "Bronze Cudgels",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   bronzeCudgelsLabel,
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					host := attachedHostFor(g, item.SourceCardID)
					if host == nil {
						return nil
					}
					x := g.ResolvedThisTurn(item.SourceCardID, bronzeCudgelsLabel)
					return BoostUntilEOT{Target: host.InstanceID, Power: x, Label: "Bronze Cudgels — +X/+0"}.Apply(NewContext(g, item))
				},
			},
			EquipAbility("{1}"),
		},
	})
}
