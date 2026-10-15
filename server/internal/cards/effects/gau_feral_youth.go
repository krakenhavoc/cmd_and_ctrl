package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gau, Feral Youth — Legendary Creature — Human Berserker {1}{R}, 2/2
// (deck request #3655):
//
//	"Rage — Whenever Gau attacks, put a +1/+1 counter on it.
//	 At the beginning of each end step, if a card left your graveyard
//	 this turn, Gau deals damage equal to its power to each opponent."
//
// The end-step ability has an intervening "if" (CR 603.4), checked when
// the step begins and again as the trigger resolves. "A card left your
// graveyard" reads the turn's event log for a zone change out of a
// graveyard from a card its controller owns, so a card cast from the
// graveyard, returned, exiled or shuffled away all count. The damage is
// the power Gau has as it resolves (its last-known power if it left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e96d12b-8381-4706-8089-3521aa2b5422",
		Name:         "Gau, Feral Youth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Gau, Feral Youth — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			{
				Watches: []game.EventKind{game.EventStepBegan},
				AppliesTo: AllOf(
					StepBegan(game.StepEnd, false),
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return cardLeftYourGraveyardThisTurn(g, source.Controller)
					},
				),
				Key: "Gau, Feral Youth — damage equal to its power to each opponent",
				Effect: func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the intervening if again on resolution.
					if !cardLeftYourGraveyardThisTurn(g, item.Controller) {
						return nil
					}
					power := 0
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						power = c.CurrentPower()
					}
					if power <= 0 {
						return nil
					}
					return damageToEachOpponent(g, item, power)
				},
			},
		},
	})
}

// cardLeftYourGraveyardThisTurn is "a card left your graveyard this
// turn": any zone change this turn out of a graveyard, of a card the
// player owns.
func cardLeftYourGraveyardThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventZoneMove || ev.OldZone != game.ZoneGraveyard {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.Owner == player {
			return true
		}
	}
	return false
}
