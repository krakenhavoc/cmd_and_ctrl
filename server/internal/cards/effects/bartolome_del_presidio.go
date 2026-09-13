package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bartolomé del Presidio — Legendary Creature — Vampire Knight
// {W}{B}, 2/1 (EDHREC rank 1780):
//
//	"Sacrifice another creature or artifact: Put a +1/+1 counter on
//	 Bartolomé del Presidio."
//
// The free sac outlet. One activated ability whose whole cost is the
// sacrifice — Warren Soultrader's shape without the life — over "a
// creature or artifact" that isn't Bartolomé himself. "Another" is
// read by name (b03NotNamed), because the sacrifice clause is built
// once at init with no instance to compare against; the only thing
// that excludes is a second Bartolomé, which the legend rule already
// forbids, and a token copy of him, which is weaker than printed and
// declared.
//
// Sandbox simplification, declared: a token copy of Bartolomé cannot
// be fed to his own ability.
func init() {
	Register(Spec{
		OracleID:     "f47e4c56-0a0b-422d-bf3d-7a20ee289f15",
		Name:         "Bartolomé del Presidio",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A token copy of Bartolomé del Presidio can't be sacrificed to his own ability."},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice another creature or artifact: Put a +1/+1 counter on Bartolomé del Presidio.",
			Cost: game.AbilityCost{
				SacrificeOther: sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()), b03NotNamed("Bartolomé del Presidio")),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !b15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
