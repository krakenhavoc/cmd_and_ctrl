package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ice Flan — Creature — Elemental Ooze {4}{U}{U}, 5/4 (Final Fantasy,
// #4462):
//
//	"When this creature enters, tap target artifact or creature an
//	 opponent controls. Put a stun counter on it.
//	 Islandcycling {2}"
//
// Summon: Shiva's tap-and-stun over an artifact-or-creature target an
// opponent controls. Islandcycling is Angel of the Ruins's Plainscycling
// with the Island subtype.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b7120a6-700b-4c61-bea4-a835f0cc0bcb",
		Name:         "Ice Flan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Ice Flan — tap target artifact or creature an opponent controls and put a stun counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
				}), TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
		Activated: []ActivatedAbility{
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
		},
	})
}
