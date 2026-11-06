package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anafenza, Unyielding Lineage — Legendary Creature — Spirit Soldier
// {2}{W}, 2/2 (Tarkir: Dragonstorm, #4463):
//
//	"Flash
//	 First strike
//	 Whenever another nontoken creature you control dies, Anafenza endures
//	 2. (Put two +1/+1 counters on it or create a 2/2 white Spirit
//	 creature token.)"
//
// The dies condition is Sifter of Skulls': it looks back in time (CR
// 603.10a), so "you control" is the controller as the creature died,
// "another" excludes Anafenza and "nontoken" excludes a token. Each
// qualifying creature is its own trigger. Endure is endure_n.go (CR
// 701.63): the controller chooses the counters or the 2/2 Spirit, and an
// Anafenza that has left the battlefield by the time the trigger resolves
// can only make the Spirit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "659bc01c-fa60-4d00-bdad-5bca1d2a978f",
		Name:            "Anafenza, Unyielding Lineage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "first strike"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, sifterOfSkullsCondition,
				"Anafenza, Unyielding Lineage — Anafenza endures 2",
				func(g *game.Game, item *game.StackItem) error {
					return EndureN(NewContext(g, item), item.SourceCardID, 2, "Anafenza, Unyielding Lineage")
				}),
				game.Purpose{DeathPayoff: true}),
		},
	})
}
