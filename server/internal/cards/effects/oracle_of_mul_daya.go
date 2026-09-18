package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oracle of Mul Daya — Creature — Elf Shaman {3}{G}, 2/2:
//
//	"You may play an additional land on each of your turns.
//	 Play with the top card of your library revealed.
//	 You may play lands from the top of your library."
//
// Three static abilities, three engine seams, and not one line of
// resolution code — which is the measure of ADR 0066 and #500 both
// having landed the derivation the right way round.
//
//   - "An additional land on each of your turns" is #500's
//     AdditionalLandPlays, summed over the permanents you control, so
//     two Oracles give two extra drops and one dying does not strand
//     the other's.
//   - "Play with the top card of your library REVEALED" is the public
//     strength of CR 401.5's visibility, as against Bolas's Citadel's
//     private "you may look". It is the only thing in the engine that
//     puts a card from an opponent's library on the wire, and the
//     per-viewer filter keeps it to exactly that one card.
//   - "You may play LANDS from the top of your library" — a filtered
//     standing permission. Playing a land is not casting (CR 305.1),
//     so the filter is what stops this being a free Future Sight; a
//     spell on top is visible to the table and unplayable.
//
// The land still spends a land drop. CR 305.2 is about playing a land,
// not about where the land was, so the extra drop from the first line
// is what makes the second and third lines worth anything — and that
// composition is the whole card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "6c5b02eb-7829-436a-8555-fea200e4b67f",
		Name:                "Oracle of Mul Daya",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		LibraryTopVisible:   game.LibraryTopRevealed,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Oracle of Mul Daya)"),
		},
	})
}
