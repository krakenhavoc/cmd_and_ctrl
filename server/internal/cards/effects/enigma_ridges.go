package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enigma Ridges — Plane — Echoir (#4870):
//
//	"When you planeswalk to Enigma Ridges, each player who controls
//	 fewer lands than the player who controls the most lands searches
//	 their library for a number of basic land cards less than or equal
//	 to the difference, reveals them, puts them into their hand, then
//	 shuffles.
//	 Whenever chaos ensues, draw a card, then you may put a land card
//	 from your hand onto the battlefield."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// land counts and each player's difference are fixed as the ability resolves,
// before anyone searches (the searches only add cards to hands); each such
// player then searches, in turn order starting with the active player. The
// chaos ability (CR 311.7) draws, then offers the planar controller a
// land card from their hand (it enters untapped, as put onto the battlefield).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93b3acc1-2d41-4f5f-9225-7a30ea2d8158",
		Name:         "Enigma Ridges",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("Enigma Ridges — each player who controls fewer lands than the player who controls the most searches for that many basic land cards",
				enigmaRidgesSearch),
			WheneverChaosEnsues("Enigma Ridges — draw a card, then you may put a land card from your hand onto the battlefield",
				Do(DrawCards{N: 1}, MayPutALandFromHand("Enigma Ridges"))),
		},
	})
}

func enigmaRidgesSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := apnapPlayers(g)
	counts := map[uuid.UUID]int{}
	most := 0
	for _, id := range players {
		counts[id] = len(MatchingBattlefield(ctx, And(Land(), ControlledBy(id))))
		most = max(most, counts[id])
	}
	for _, id := range players {
		diff := most - counts[id]
		if diff <= 0 {
			continue
		}
		if err := (SearchLibrary{Player: id, Predicate: IsBasicLand, Dest: game.ZoneHand, Limit: diff, Reveal: true, Shuffle: true,
			Source: item.SourceCardID, Reason: "Enigma Ridges — search for basic land cards"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
