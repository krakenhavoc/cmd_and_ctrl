package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sazh's Chocobo — Creature — Bird {G}, 0/1 (EDHREC rank 3052):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1
//	 counter on this creature."
//
// The one-drop landfall body. Tireless Provisioner's condition
// (b13LandYouControlEntered — a land entering under the
// controller's control, played, fetched or reanimated alike); the
// counter goes on the Chocobo if it is still on the battlefield
// when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2b082b4-5d9c-4125-b353-d1cf4f334720",
		Name:         "Sazh's Chocobo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Sazh's Chocobo — landfall: put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
