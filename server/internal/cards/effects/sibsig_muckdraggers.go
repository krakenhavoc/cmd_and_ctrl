package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sibsig Muckdraggers — Creature — Zombie 3/5 {8}{B}:
//
//	"Delve. When this creature enters, return target creature card from
//	 your graveyard to your hand."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// The enters trigger is mandatory ("return", not "you may") and is
// removed if the graveyard holds no creature card (CR 603.3d). A card
// delved away to cast it is in exile and cannot be returned, which is
// the printed interaction. No simplification.
func init() {
	Register(Spec{
		OracleID:     "423c1079-bdba-42fc-8732-59cb83ebafc3",
		Name:         "Sibsig Muckdraggers",
		Completeness: CompletenessFull,
		Delve:        true,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sibsig Muckdraggers — return target creature card to your hand", returnTargetCardToHand),
				TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
			),
		},
	})
}
