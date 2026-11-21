package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost Isle Calling — Enchantment {1}{U} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Whenever you scry, put a verse counter on this enchantment.
//	 {4}{U}{U}, Exile this enchantment: Draw a card for each verse counter
//	 on this enchantment. If it had seven or more verse counters on it,
//	 take an extra turn after this one. Activate only as a sorcery."
//
// The counter trigger is Galadriel of Lothlórien's scry watcher: one verse
// counter per scry, however many cards the scry looked at (CR 701.22). The
// ability exiles the enchantment as a cost, so the count is read at
// resolution as the enchantment last existed on the battlefield
// (CR 608.2h, ctx.SourcePermanent). "If it had seven or more" reads the
// same count. Sorcery speed is the activation timing (CR 307.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0efbf29c-09a0-4472-83ad-8d692c0232a5",
		Name:         "Lost Isle Calling",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventScry, effects.ByYou, "Lost Isle Calling — put a verse counter on this enchantment",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AddCounter{Target: item.SourceCardID, Kind: lostIsleVerse, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}{U}{U}, Exile this enchantment: Draw a card for each verse counter on this enchantment. If it had seven or more verse counters on it, take an extra turn after this one. Activate only as a sorcery.",
			Purpose:      game.Purpose{ExtraTurns: 1},
			Cost:         effects.Plus(effects.ManaCost("{4}{U}{U}"), effects.ExileThis()),
			SorcerySpeed: true,
			Effect:       lostIsleCallingPayoff,
		}},
	})
}

// lostIsleVerse is the verse counter's name.
const lostIsleVerse = "verse"

func lostIsleCallingPayoff(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := info.Counters[lostIsleVerse]
	if n > 0 {
		if err := (effects.DrawCards{N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n >= 7 {
		return effects.TakeExtraTurn{}.Apply(ctx)
	}
	return nil
}
