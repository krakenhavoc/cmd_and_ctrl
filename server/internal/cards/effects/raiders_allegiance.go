package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raiders' Allegiance — Plane — Ixalan (Unknown Event, #4870):
//
//	"Raid — At the beginning of your second main phase, for each player
//	 dealt combat damage this turn, you get a point counter and create a
//	 2/2 black Pirate creature token with menace.
//	 Whenever you planeswalk away from here, each player with the most
//	 point counters or tied for the most point counters creates three
//	 Treasure tokens and draws three cards. Then remove all point counters
//	 from each player.
//	 Whenever Chaos ensues, proliferate for each Pirate you control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6), so
// "you" is the planar controller. Point counters are named player counters.
// The raid count is the number of players dealt combat damage by anything
// this turn. The planeswalk-away ability looks back at the plane that has just
// left (CR 901.11b); a table where nobody has a point counter ties everyone
// at zero, and each of them is "tied for the most". The chaos ability
// proliferates once per Pirate, each asked after the one before it is
// answered (the proliferate tail).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33ed1120-82b1-4936-b210-2f0323a9401c",
		Name:         "Raiders' Allegiance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourSecondMainPhase("Raid — Raiders' Allegiance: for each player dealt combat damage this turn, you get a point counter and create a 2/2 black Pirate with menace",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, p := range g.Seats {
						if p != nil && !p.Eliminated && len(g.CreaturesThatDealtCombatDamageToThisTurn(p.ID)) > 0 {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					if err := g.AddPlayerCounterForEffect(item.Controller, planesB13PointCounter, n); err != nil {
						return err
					}
					return CreateToken{Template: TokenCard("2/2 black Pirate with menace"), N: n}.Apply(NewContext(g, item))
				}),
			WhenYouPlaneswalkAwayFromThis("Raiders' Allegiance — each player with the most point counters creates three Treasures and draws three cards; remove all point counters",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					players := b13PlayersFrom(g, item.Controller)
					most := 0
					for _, id := range players {
						if n := g.PlayerByIDForEffect(id).Counters[planesB13PointCounter]; n > most {
							most = n
						}
					}
					var winners []uuid.UUID
					for _, id := range players {
						if g.PlayerByIDForEffect(id).Counters[planesB13PointCounter] == most {
							winners = append(winners, id)
						}
					}
					for _, id := range winners {
						if err := (CreateToken{Controller: id, Template: TreasureToken(), N: 3}).Apply(ctx); err != nil {
							return err
						}
						if err := (DrawCards{Player: id, N: 3}).Apply(ctx); err != nil {
							return err
						}
					}
					for _, id := range players {
						if n := g.PlayerByIDForEffect(id).Counters[planesB13PointCounter]; n > 0 {
							if err := g.AddPlayerCounterForEffect(id, planesB13PointCounter, -n); err != nil {
								return err
							}
						}
					}
					return nil
				}),
			WheneverChaosEnsues("Raiders' Allegiance — proliferate for each Pirate you control",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					pirate := NewContext(g, item).TextSubtype("Pirate")
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype(pirate) {
							n++
						}
					}
					return proliferateTimesKeyed(g, item.Controller, item.SourceCardID, n)
				}),
		},
	})
}

const planesB13PointCounter = "point"
