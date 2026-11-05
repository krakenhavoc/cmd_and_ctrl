package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tokka & Rahzar, Unsupervised — {2}{R} Legendary Creature — Turtle Wolf
// Mutant 3/3 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"First strike
//	 Whenever another nontoken creature you control leaves the
//	 battlefield, put a +1/+1 counter on Tokka & Rahzar and create a
//	 Treasure token. This ability triggers only once each turn."
//
// The Ozolith's leaves-the-battlefield condition (aCreatureYouControlLeft:
// a creature its controller controlled as it left, read from last-known
// information, CR 603.10a), narrowed to another object and to a nontoken
// one. "Only once each turn" is the per-turn trigger record Exemplar of
// Light and Ant-Man read (b11TriggeredThisTurn): once the trigger has been
// put on the stack this turn no later departure triggers it, even if the
// first is countered. The counter lands only while Tokka & Rahzar is still
// on the battlefield (CounterOnThis); the Treasure is made regardless.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8d857d79-bdb2-4ca3-abd6-7db8f6b3bcfa",
		Name:            "Tokka & Rahzar, Unsupervised",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, tokkaAndRahzarAnotherNontokenLeft, tokkaAndRahzarLabel,
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

const tokkaAndRahzarLabel = "Tokka & Rahzar — put a +1/+1 counter on Tokka & Rahzar and create a Treasure token"

// tokkaAndRahzarAnotherNontokenLeft is the trigger condition: another
// nontoken creature its controller controlled left the battlefield, and the
// ability has not yet triggered this turn.
func tokkaAndRahzarAnotherNontokenLeft(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
	if source == nil || ev.CardID == source.InstanceID || !aCreatureYouControlLeft(ev, source, ch, g) {
		return false
	}
	left, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || left.IsToken() {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, tokkaAndRahzarLabel)
}
