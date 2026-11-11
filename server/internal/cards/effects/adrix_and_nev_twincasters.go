package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adrix and Nev, Twincasters — Legendary Creature — Merfolk Wizard
// {2}{G}{U}, 2/2 (Murders at Karlov Manor Commander, #5368):
//
//	"Ward {2}
//	 If one or more tokens would be created under your control, twice
//	 that many of those tokens are created instead."
//
// Parallel Lives on a body: the CR 701.7b creation event's counts double
// (TokensDoubled, #762), for tokens created under its controller's
// control whoever's effect makes them. Beside another doubler the
// affected player orders the two (CR 616.1), and either order is x4.
// Ward {2} is the shared CR 702.21a trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "12a6cad9-eb42-43bd-9e68-aaf862cd83db",
		Name:         "Adrix and Nev, Twincasters",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Adrix and Nev, Twincasters — ward {2}"),
		},
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Adrix and Nev, Twincasters: double tokens"),
		},
	})
}
