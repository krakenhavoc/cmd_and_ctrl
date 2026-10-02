package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radha, Heart of Keld — Legendary Creature — Elf Warrior {1}{R}{G},
// 3/3:
//
//	"During your turn, Radha has first strike.
//	 You may look at the top card of your library any time, and you may
//	 play lands from the top of your library.
//	 {4}{R}{G}: Radha gets +X/+X until end of turn, where X is the
//	 number of lands you control."
//
// First strike during your turn is Razorkin Needlehead's conditional
// layer 6 grant. The look is private (LibraryTopOwner) and the
// permission is lands only, the Oracle of Mul Daya shape without the
// extra drop. The pump reads the land count when the ability resolves
// (X is not chosen on activation) and goes through BoostUntilEOT on the
// source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "927fb139-0493-48be-8cc4-2c7d854d6a55",
		Name:              "Radha, Heart of Keld",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Radha, Heart of Keld)"),
		},
		Static: []game.StaticAbility{firstStrikeDuringYourTurn()},
		Activated: []ActivatedAbility{{
			Label: "{4}{R}{G}: Radha gets +X/+X until end of turn, where X is the number of lands you control",
			Cost:  ManaCost("{4}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := b10LandsControlled(g, item.Controller)
				return BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     x,
					Toughness: x,
					Label:     "Radha, Heart of Keld — +X/+X",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
