package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Retrieval — Sorcery {3}{U} (Innistrad Remastered):
//
//	"Return target instant or sorcery card from your graveyard to your hand.
//	 Flashback {2}{R}"
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
