package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creakwood Safewright — Creature — Elf Warrior {1}{B}, 5/5 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with three -1/-1 counters on it.
//	 At the beginning of your end step, if there is an Elf card in your
//	 graveyard and this creature has a -1/-1 counter on it, remove a -1/-1
//	 counter from this creature."
//
// The counters are an entry replacement (Loch Mare). Both halves of the
// "if" are an intervening if (CR 603.4): checked as the end step begins,
// so with no Elf card in the graveyard or no -1/-1 counter left nothing
// triggers, and checked again as the ability resolves, so an Elf card
// exiled from the graveyard in response stops it. The removal is
// Bristlebane Battler's. "Elf" is a word of the permanent's text, read
// through it (CR 612.2, Lys Alana Dignitary), and an Elf card is judged by
// its subtypes in the graveyard, so a changeling card counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9b7bbea-ea20-400b-a5b7-4d6993d8777b",
		Name:         "Creakwood Safewright",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 3, "Creakwood Safewright: enters with three -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creakwoodSafewrightCondition(g, source)
			}), "Creakwood Safewright — remove a -1/-1 counter from this creature", func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !onBattlefield(g, item.SourceCardID) || !creakwoodSafewrightCondition(g, &src) {
					return nil
				}
				return removeAMinusOneCounterFromThis(g, item)
			}),
		},
	})
}

// creakwoodSafewrightCondition is the intervening if: an Elf card in the
// source's controller's graveyard, and a -1/-1 counter on the source.
func creakwoodSafewrightCondition(g *game.Game, source *game.Card) bool {
	if source.Counters[game.CounterMinusOne] <= 0 {
		return false
	}
	elf := source.TextSubtype("Elf")
	return GraveyardAtLeast(1, func(c game.Card) bool { return c.HasSubtype(elf) })(g, source.Controller, source.InstanceID)
}
