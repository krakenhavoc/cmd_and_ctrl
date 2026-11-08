package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectacular Pileup — Sorcery {3}{W}{W} (Aetherdrift, #4464):
//
//	"All creatures and Vehicles lose indestructible until end of turn,
//	 then destroy all creatures and Vehicles.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Smite the Deathless's loss of indestructible (a layer-6 removal with
// its own timestamp, CR 613.1f, ending in cleanup) over every creature
// and Vehicle on the battlefield as the spell resolves (CR 611.2c), then
// Wrath of God's DestroyAllMatching over the same class. Regeneration is
// not barred, so a shield still saves its permanent (CR 701.19). Cycling
// works from the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f7479e6-b2bd-46b1-96ad-34deb43565c1",
		Name:         "Spectacular Pileup",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			match := Or(Creature(), HasSubtype("Vehicle"))
			for _, c := range MatchingBattlefield(ctx, match) {
				if err := g2LosesIndestructibleUntilEOT(ctx, c.InstanceID); err != nil {
					return err
				}
			}
			return DestroyAllMatching{Match: match}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
