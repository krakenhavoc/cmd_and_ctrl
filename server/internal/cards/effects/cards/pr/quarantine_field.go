package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quarantine Field — Enchantment {X}{X}{W}{W} (Battle for Zendikar):
//
//	"This enchantment enters with X isolation counters on it.
//	 When this enchantment enters, for each isolation counter on it, exile up
//	 to one target nonland permanent an opponent controls until this
//	 enchantment leaves the battlefield."
//
// The targets are read as the ability is put on the stack, one up-to-one
// clause per isolation counter on the enchantment as it entered.
//
// No simplification.
func init() {
	const name = "Quarantine Field"
	enters := effects.WhenThisEnters(name+" — for each isolation counter on it, exile up to one target nonland permanent an opponent controls until this enchantment leaves the battlefield",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.ExileUntil{Target: t.ID, ThisLeaves: true,
					Label: name + " — the exiled card returns when it leaves the battlefield"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.Harmful(effects.TargetPermanent("up to one target nonland permanent an opponent controls for each isolation counter",
			effects.And(effects.Nonland(), effects.OpponentControls())).WithCount(0, source.Counters[effects.S82OldB312IsolationCounter]))
	}
	effects.Register(effects.Spec{
		OracleID:                   "f8b07b1a-f9fb-427d-a5c8-b65a01f7b079",
		Name:                       name,
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(effects.S82OldB312IsolationCounter)},
		Triggered:                  []game.TriggeredAbility{enters},
	})
}
