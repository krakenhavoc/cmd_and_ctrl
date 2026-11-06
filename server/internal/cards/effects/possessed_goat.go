package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Possessed Goat — Creature — Goat {W}, 1/1 (Duskmourn: House of Horror,
// #4466):
//
//	"{3}, Discard a card: Put three +1/+1 counters on this creature and it
//	 becomes a black Demon in addition to its other colors and types.
//	 Activate only once."
//
// "Activate only once" is the OnlyOnce bit (#5221, CR 602.5b): once per
// object, so a Goat that leaves and comes back may activate it again
// (CR 400.7). It is not an exhaust ability, so nothing that watches
// exhaust abilities sees it.
//
// The change prints no duration, so it lasts until the end of the game
// (CR 611.2a), pinned to this object: a Goat that leaves the battlefield
// is a new object and leaves the change behind (CR 400.7). A Goat that
// left in response gets neither the counters nor the change.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "821f4ff3-ebc7-4f47-ac35-d00093b6d35b",
		Name:         "Possessed Goat",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{3}, Discard a card: Put three +1/+1 counters on this creature and it becomes a black Demon " +
				"in addition to its other colors and types. Activate only once.",
			Purpose:  game.Purpose{Answers: game.AnswerPump},
			Cost:     Plus(ManaCost("{3}"), DiscardACard()),
			OnlyOnce: true,
			Effect:   possessedGoatBecomesADemon,
		}},
	})
}

// possessedGoatBecomesADemon is the ability's effect: the counters, then
// the colour and type change, on this object only.
func possessedGoatBecomesADemon(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if err := (AddCounter{Target: self, Kind: game.CounterPlusOne, N: 3}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.AddColorsMod(ctx.TextColors("B")...),
			game.AddSubtypesMod(ctx.TextSubtypes("Demon")...),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Possessed Goat — a black Demon in addition to its other colors and types",
	}.Apply(ctx)
}
