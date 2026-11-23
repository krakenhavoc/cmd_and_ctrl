package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Regenerations Restored — Enchantment {W}{U} (Doctor Who):
//
//	"Vanishing 12 (This enchantment enters with twelve time counters on
//	 it. At the beginning of your upkeep, remove a time counter from it.
//	 When the last is removed, sacrifice it.)
//	 Whenever one or more time counters are removed from this enchantment,
//	 scry 1 and you gain 1 life. Then if this enchantment has no time
//	 counters on it, exile it. When you do, take an extra turn after this
//	 one."
//
// Vanishing is cmdr2B17Vanishing*. The second ability fires on a removal
// (EventCounterPlaced whose total fell, one trigger per removal event, CR
// 603.2c via the engine's batching). If the enchantment is gone or still
// has counters when the "then if" is checked, nothing is exiled and no
// reflexive trigger exists; if the vanishing sacrifice resolved first, the
// enchantment is no longer on the battlefield and cannot be exiled, so
// there is no extra turn (the controller orders the two triggers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1e691f3-a3e9-4cfa-a761-d7e020364dfa",
		Name:         "Regenerations Restored",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1, LifeGain: 1, ExtraTurns: 1},
		Replacements: []game.ReplacementEffect{effects.Cmdr2B17VanishingEntry(12)},
		Triggered: append(effects.Cmdr2B17VanishingTriggers(),
			effects.On(game.EventCounterPlaced, regenerationsRestoredRemoved,
				"Regenerations Restored — scry 1 and gain 1 life; then if it has no time counters, exile it, and when you do take an extra turn",
				regenerationsRestoredEffect)),
	})
}

// regenerationsRestoredRemoved is "one or more time counters are removed
// from this enchantment": a time-counter event on the source whose total
// fell.
func regenerationsRestoredRemoved(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCounterPlaced && ev.Target == source.InstanceID &&
		ev.Label == game.CounterTime && ev.Amount < g.CounterCountBeforeForEffect(ev)
}

func regenerationsRestoredEffect(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := effects.Do(effects.Scry{N: 1}, effects.GainLife{Amount: 1})(g, item); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !effects.OnBattlefield(g, item.SourceCardID) || c.Counters[game.CounterTime] > 0 {
		return nil
	}
	if err := (effects.ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	if effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.WhenYouDo("Regenerations Restored — take an extra turn after this one", regenerationsRestoredExtraTurnBody).Apply(ctx)
}

// The reflexive body. An on-disk identity.
var regenerationsRestoredExtraTurnBody = game.SimpleDelayedBody("regenerations-restored/extra-turn", effects.YouTakeAnExtraTurnEffect)
