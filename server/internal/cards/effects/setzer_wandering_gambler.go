package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Setzer, Wandering Gambler — Legendary Creature — Human Rogue Pilot
// {1}{B}{R}, 2/2 (deck request #3655):
//
//	"When Setzer enters, create The Blackjack, a legendary 3/3 colorless
//	 Vehicle artifact token with flying and crew 2.
//	 Whenever a Vehicle you control deals combat damage to a player, flip
//	 a coin.
//	 Whenever you win a coin flip, create two tapped Treasure tokens."
//
// The Blackjack is a token template with its own crew ability
// (BlackjackToken, #3767). The Vehicle trigger fires once per Vehicle
// that connects, whoever crewed it and whichever Vehicle it is, and its
// flip is a called one (FlipACoin): Setzer's third ability cares whether
// you win. That third ability sees every coin you win, from any source —
// Krark, the Thumbless or Tavern Scoundrel included — and one won coin
// is one trigger (WheneverYouWinACoinFlip).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acdd5082-7b3c-4fd3-941b-d19dd72698ec",
		Name:         "Setzer, Wandering Gambler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Setzer, Wandering Gambler — create The Blackjack", Do(CreateToken{Template: BlackjackToken(), N: 1})),
			On(game.EventDealDamage, setzerVehicleDealtCombatDamageToAPlayer,
				"Setzer, Wandering Gambler — flip a coin",
				Do(FlipACoin{Question: "Setzer, Wandering Gambler — call the coin flip"})),
			WheneverYouWinACoinFlip("Setzer, Wandering Gambler — create two tapped Treasure tokens", setzerTappedTreasures),
		},
	})
}

// setzerVehicleDealtCombatDamageToAPlayer is "Whenever a Vehicle you
// control deals combat damage to a player": the damage source is a
// creature the trigger's controller controls (combatDamageToPlayerBy) and
// a Vehicle as it deals the damage.
func setzerVehicleDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.HasSubtype("Vehicle")
}

// setzerTappedTreasures creates the two tapped Treasures.
func setzerTappedTreasures(g *game.Game, item *game.StackItem) error {
	if item.Controller == uuid.Nil {
		return nil
	}
	_, err := g.CreateTokensForEffect(item.Controller, TreasureToken(), 2, game.TokenEntryOptions{Tapped: true})
	return err
}
