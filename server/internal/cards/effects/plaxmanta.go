package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plaxmanta — Creature — Beast {1}{U}, 2/2 (Dissension, #5598):
//
//	"Flash
//	 When this creature enters, creatures you control gain shroud until end
//	 of turn. (They can't be the targets of spells or abilities.)
//	 When this creature enters, sacrifice it unless {G} was spent to cast it."
//
// The shroud is Blinding Fog's grant to the creatures you control as the
// trigger resolves (CR 611.2c).
// The sacrifice is disSacrificeUnlessSpent: it reads what paid for the
// spell this permanent was (CR 400.7d), so a creature put onto the
// battlefield without being cast is sacrificed too. The auto-tapper is
// asked to spend the colour (WantsColors), an ordering and never a filter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36780340-824d-420b-b41b-182f88d944de",
		Name:            "Plaxmanta",
		Completeness:    CompletenessFull,
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"G"},
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Plaxmanta — creatures you control gain shroud until end of turn", Do(GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"shroud"},
				Label:    "Plaxmanta — shroud",
			})),
			disSacrificeUnlessSpent("Plaxmanta", "G"),
		},
	})
}
