package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tectonic Hellion — Creature — Hellion {5}{R}{R}, 8/5 (Commander 2019,
// #5517):
//
//	"Haste
//	 Whenever this creature attacks, each player who controls the most
//	 lands sacrifices two lands of their choice."
//
// The lands are counted as the trigger resolves (CR 608.2h), and every
// player tied for the most is in the set, the Hellion's controller
// included. Each of them is asked for two lands (a seat listed twice owes
// two, Ember Swallower's edict), in APNAP order, and everything chosen is
// sacrificed together. A player with fewer than two lands sacrifices what
// they have (CR 701.21a). Not targeted, so a hexproof land is as
// sacrificeable as any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e67b8eb2-7b39-4fdb-9ee2-686d1d781cfa",
		Name:            "Tectonic Hellion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Tectonic Hellion — each player who controls the most lands sacrifices two lands",
				c19TectonicHellionEdict),
		},
	})
}

func c19TectonicHellionEdict(g *game.Game, item *game.StackItem) error {
	lands := map[uuid.UUID]int{}
	most := -1
	for _, seat := range effects.SeatsFromActive(g) {
		n := len(effects.PermanentsControlledByMatching(g, seat, effects.Land()))
		lands[seat] = n
		if n > most {
			most = n
		}
	}
	var asks []uuid.UUID
	for _, seat := range effects.SeatsFromActive(g) {
		if lands[seat] == most && most > 0 {
			asks = append(asks, seat, seat)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	g.PlayersSacrificeForEffect(item.SourceCardID, asks,
		effects.SacrificeSpec("a land", effects.Land()), "Sacrifice a land")
	return nil
}
