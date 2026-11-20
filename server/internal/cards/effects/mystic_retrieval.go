package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Retrieval — Sorcery {3}{U} (Dark Ascension, #5571):
//
//	"Return target instant or sorcery card from your graveyard to
//	 your hand.
//	 Flashback {2}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Relearn's return. Cast with flashback, the card is on the stack, not in
// the graveyard, so it can't target itself (CR 601.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b7cfbfa1-dd38-47ff-9138-fa6b247afc11",
		Name:             "Mystic Retrieval",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
		OnResolve:        returnTargetFromGraveyardToHand,
	})
}
