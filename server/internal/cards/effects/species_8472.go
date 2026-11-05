package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Species 8472 — {4}{G}{G} Creature — Insect Beast 7/5 (Star Trek, #4973):
//
//	"Strange new worlds — When this creature enters, you gain life equal
//	 to the number of differently named lands you control.
//	 This creature can't be blocked by artifact creatures.
//	 This creature can't be the target of abilities from artifact sources
//	 your opponents control."
//
// Strange new worlds is an ability word (CR 207.2c); X is counted as the
// trigger resolves, each land name once (socB23DifferentlyNamedLands).
// The block rule is Fen Hauler's. The targeting clause refuses abilities
// only, not spells, and only an opponent's
// (trekB22AbilitiesFromOpponentsArtifactSourcesCantTargetThis); it is
// asked at announcement (CR 601.2c, 602.2b) and again on resolution (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "645a5619-d0b2-4153-aae0-ca2adc6f5e3d",
		Name:         "Species 8472",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Species 8472 — strange new worlds: gain life equal to the number of differently named lands you control",
				Do(Instruction(func(ctx *Context) error {
					return GainLife{Player: ctx.Controller(), Amount: socB23DifferentlyNamedLands(ctx)}.Apply(ctx)
				}))),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), Artifact()), "artifact creatures"),
		},
		TargetingRestrictions: []game.TargetingRestriction{
			trekB22AbilitiesFromOpponentsArtifactSourcesCantTargetThis(
				"This creature can't be the target of abilities from artifact sources your opponents control."),
		},
	})
}
