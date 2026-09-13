package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steel Overseer — Artifact Creature — Construct {2}, 1/1 (EDHREC
// rank 808):
//
//	"{T}: Put a +1/+1 counter on each artifact creature you control."
//
// The artifact-aggro lord. One tap ability (summoning sickness
// applies, CR 302.1), whose effect walks the battlefield when it
// resolves — so an artifact creature that enters in response gets a
// counter, as printed. The Overseer counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "986ae327-f433-4c58-93dc-afc544b9bfcb",
		Name:         "Steel Overseer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: Put a +1/+1 counter on each artifact creature you control.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != item.Controller || !c.IsCreature() || !c.IsArtifact() {
						continue
					}
					if err := (AddCounter{Target: c.InstanceID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
