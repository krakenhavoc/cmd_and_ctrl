package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zar Ojanen, Scion of Efrava — Legendary Creature — Cat Warrior
// {3}{G}{W}, 4/4 (Dominaria United):
//
//	"Domain — Whenever Zar Ojanen becomes tapped, put a +1/+1 counter on
//	 each creature you control with toughness less than the number of
//	 basic land types among lands you control."
//
// "Becomes tapped" is the tap event or an attack that taps it. The
// domain count (DomainCount, #5666) and each creature's toughness are
// read as the ability resolves, and the set is fixed before any counter
// goes on (CR 608.2): a counter that raises one creature's toughness
// does not pull in or push out another.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1402c792-6e2f-4a85-8a0c-8db3c8ce5f1f",
		Name:         "Zar Ojanen, Scion of Efrava",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			dmuB11BecameTapped("Zar Ojanen — put a +1/+1 counter on each creature you control with toughness less than the number of basic land types among lands you control",
				dmuB11CountersOnSmallCreatures),
		},
	})
}
