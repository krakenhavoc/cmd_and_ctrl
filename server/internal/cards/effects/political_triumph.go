package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Political Triumph — {W} Enchantment — Plan (Marvel Super Heroes, #4455):
//
//	"Whenever a creature you control enters, scry 1 and put a plan counter
//	 on this enchantment.
//	 When the fourth plan counter is put on this enchantment, sacrifice it,
//	 draw a card, and put a +1/+1 counter on each creature you control."
//
// Construct a Cosmic Cube's plan shape. The enter trigger scries and then
// adds the counter in one resolution. The payoff is a second trigger
// watching the plan-counter total cross four from below (Claim the Kingdom);
// its three verbs are one sentence, so the draw and the counters happen even
// when the enchantment is already gone, and the sacrifice only when it is
// still the same object (CR 400.7). The set of creatures is read after the
// draw and the sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bbceb51-1fb8-44d7-8afc-b4746d292262",
		Name:         "Political Triumph",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, politicalTriumphCreatureEntered,
				"Political Triumph — scry 1 and put a plan counter on this enchantment", politicalTriumphScryAndCount),
			On(game.EventCounterPlaced, mshB2PlanCounterReached(4),
				"Political Triumph — sacrifice it, draw a card, and put a +1/+1 counter on each creature you control",
				politicalTriumphPayoff),
		},
	})
}

// politicalTriumphCreatureEntered is "a creature you control enters".
func politicalTriumphCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature()
}

// politicalTriumphScryAndCount scries 1, then adds the plan counter.
func politicalTriumphScryAndCount(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (Scry{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(ctx)
}

// politicalTriumphPayoff is the fourth-counter payoff.
func politicalTriumphPayoff(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if onBattlefield(g, item.SourceCardID) && !ctx.isNewSourceObjectAsThis(item.SourceCardID) {
		if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return b11PutCounterOnEachCreatureYouControl(g, item)
}
