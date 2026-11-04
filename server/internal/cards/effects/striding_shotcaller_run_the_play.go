package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Striding Shotcaller // Run the Play — Creature — Troll Druid {G}{U}, 0/4,
// with the prepare spell Run the Play, Sorcery {X}{G}{U} (Secrets of
// Strixhaven Commander, #4834):
//
//	"Reach
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, this creature becomes prepared. (While it's prepared, you may
//	 cast a copy of its spell. Doing so unprepares it.)"
//
//	Run the Play: "Put a +1/+1 counter on each of up to X target creatures.
//	 Those creatures gain flying until end of turn. Draw a card."
//
// A preparation card (CR 722): the permanent under the bare oracle ID and
// its prepare spell under "#1". The trigger is one occurrence per player
// hit in a damage step (CR 603.2c, OncePerBatchPerPlayer), and a creature
// that is already prepared does not become prepared again.
//
// "Up to X target creatures" is CountFromX with UpToX (March of Swirling
// Mist's clause, which The Crowd Goes Wild also calls): the announced X is the ceiling. The counters and the
// flying go to the targets that are still legal as the spell resolves
// (CR 608.2b); "those creatures" is that set, and the card is drawn even
// with no target, X = 0 included, so the spell does not declare XMatters:
// at X = 0 it still draws.
//
// No simplification.
func init() {
	const oracle = "10389ff7-2ea4-4413-90cc-0e3ca268c64d"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Striding Shotcaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return combatDamageToPlayerBy(ev, source.Controller, g)
				},
				"Striding Shotcaller — this creature becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				})),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Run the Play",
		Completeness: CompletenessFull,
		Targets:      marchOfSwirlingMistTargets(),
		OnResolve:    runThePlayResolve,
	})
}

// runThePlayResolve puts a +1/+1 counter on each still-legal target, has
// those creatures gain flying until end of turn, then draws a card.
func runThePlayResolve(item *game.StackItem, ctx *Context) error {
	for _, id := range legalTargetCards(item, ctx.Game) {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range legalTargetCards(item, ctx.Game) {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"flying"},
			Label:    "Run the Play — flying until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}
