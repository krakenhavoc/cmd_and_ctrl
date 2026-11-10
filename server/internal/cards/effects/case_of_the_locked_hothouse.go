package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Locked Hothouse — Enchantment — Case {3}{G} (Murders at
// Karlov Manor, #4469):
//
//	"You may play an additional land on each of your turns.
//	 To solve — You control seven or more lands. (If unsolved, solve at the
//	 beginning of your end step.)
//	 Solved — You may look at the top card of your library any time, and
//	 you may play lands and cast creature and enchantment spells from the
//	 top of your library."
//
// The extra land drop is Exploration's (Spec.AdditionalLandPlays, so two
// Cases give two, the printed ruling). The solve condition counts lands
// at the end step and again on resolution (CR 603.4).
//
// The solved line is Augur of Autumn's: a private look at the top card and
// standing permissions to play lands and cast creature and enchantment
// spells from it, all gated on the Case being solved. The look is gated
// too (Spec.LibraryTopVisibleWhen, b41's seam), so an unsolved Case shows
// nothing and offers nothing; "creature and enchantment spells" is two
// permissions, one per card type, so an enchantment creature is castable
// by either. Ordinary timing applies (CR 401.5, 305.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "2380fd04-6d19-4f5a-8a9e-c17d9ae09fb6",
		Name:                  "Case of the Locked Hothouse",
		Completeness:          CompletenessFull,
		AdditionalLandPlays:   1,
		Purpose:               game.Purpose{ExtraLandDrops: 1},
		LibraryTopVisible:     game.LibraryTopOwner,
		LibraryTopVisibleWhen: Solved(),
		Triggered: []game.TriggeredAbility{
			ToSolve("Case of the Locked Hothouse — to solve: you control seven or more lands",
				func(g *game.Game, controller, _ uuid.UUID) bool {
					return len(permanentsControlledByMatching(g, controller, Land())) >= 7
				}),
		},
		GatedCastPermissions: []game.CastPermissionGate{
			{
				Permission: PlayFromTopOfYourLibrary(game.PermissionFilter{LandsOnly: true},
					"Play a land from the top of your library (Case of the Locked Hothouse)"),
				ActiveWhen: Solved(),
			},
			{
				Permission: PlayFromTopOfYourLibrary(game.PermissionFilter{CreatureOnly: true},
					"Cast a creature spell from the top of your library (Case of the Locked Hothouse)"),
				ActiveWhen: Solved(),
			},
			{
				Permission: PlayFromTopOfYourLibrary(game.PermissionFilter{CardType: "enchantment"},
					"Cast an enchantment spell from the top of your library (Case of the Locked Hothouse)"),
				ActiveWhen: Solved(),
			},
		},
	})
}
