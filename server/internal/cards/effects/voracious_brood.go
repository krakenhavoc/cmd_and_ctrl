package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Brood — Creature — Alien Insect {2}{G}, 1/1 (Marvel Super Heroes
// Commander, #4455):
//
//	"This creature enters with a +1/+1 counter on it for each creature card
//	 in your graveyard.
//	 Whenever one or more creature cards are put into your graveyard from
//	 anywhere, put that many +1/+1 counters on this creature."
//
// The entry counters are Golgari Grave-Troll's CR 614 self-entry replacement,
// counted as the Brood enters. The second clause is the first "from anywhere"
// batch trigger (#4721, graveyard_arrival.go): a discard, a mill, a spell
// resolving or countered, a surveil or a creature dying each put a card in
// the graveyard, and every arrival in one event batch (CR 603.2c, one
// resolution or one turn-based action) is one occurrence. The first creature
// card of the batch fires the ability (OncePerBatch) and "that many" is read
// off the batch's events as the ability resolves, so a creature card that
// left the graveyard in response still counts: it was put there. A token is
// not a card (CR 108.2b), so a dying creature token adds nothing, and a
// card is judged by what it is in the graveyard (CR 603.6c), so a Vehicle
// that was a creature as it died is no creature card.
//
// "Your graveyard" is the one the Brood's controller owns; a creature card
// that goes to an opponent's graveyard does nothing.
//
// No simplification.
func init() {
	const label = "Voracious Brood — put that many +1/+1 counters on it"
	Register(Spec{
		OracleID:     "aeee8308-76aa-4ed5-b7b2-04fb9e29c676",
		Name:         "Voracious Brood",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted("+1/+1", func(g *game.Game, src *game.Card) int {
				return b11CreatureCardsInGraveyard(g, src.Controller)
			}, "Voracious Brood: enters with a +1/+1 counter per creature card in your graveyard"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(graveyardArrivalKinds, creatureCardsToYourGraveyard.When(), label,
				Do(Instruction(voraciousBroodCounters)))),
		},
	})
}

// voraciousBroodCounters is "that many": the creature cards the triggering
// batch put into the controller's graveyard.
func voraciousBroodCounters(ctx *Context) error {
	item := ctx.Item
	if item.Trigger == nil {
		return nil
	}
	n := len(creatureCardsToYourGraveyard.InBatch(ctx.Game, item.Trigger.Event.Batch, item.Controller))
	if n == 0 {
		return nil
	}
	return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
