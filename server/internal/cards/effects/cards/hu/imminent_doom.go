package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imminent Doom — Enchantment {2}{R} (Hour of Devastation, #5534):
//
//	"This enchantment enters with a doom counter on it.
//	 Whenever you cast a spell with mana value equal to the number of doom
//	 counters on this enchantment, this enchantment deals that much damage to
//	 any target. Then put a doom counter on this enchantment."
//
// The counter is a self-replacement on entry (CR 614.1c), so it is there for the
// first spell cast. The trigger watches your casts and compares the spell's
// mana value (X counted as announced, CR 202.3e) with the doom counters as it
// is cast; the target is chosen as the trigger goes on the stack (CR 603.3d).
// "That much" is read as the ability resolves, from the enchantment as it is
// (or as it last was, CR 608.2h), and the damage is dealt before the counter is
// added; a target that has gone makes the whole ability do nothing (CR 608.2b),
// so no counter is added then. A source that has left cannot take the counter.
//
// No simplification.
func init() {
	const doom = "doom"
	effects.Register(effects.Spec{
		OracleID:     "6ded956d-30ed-49ca-8653-cbaaae676b0c",
		Name:         "Imminent Doom",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(doom, 1, "Imminent Doom: enters with a doom counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !effects.ByYou(ev, source, game.Characteristic{}, g) {
						return false
					}
					mv, ok := g.LastKnownSpellManaValueForEffect(ev.CardID)
					return ok && mv == source.Counters[doom]
				}, "Imminent Doom — it deals that much damage to any target, then put a doom counter on it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if err := effects.SourceDealsComputedDamageToEachLegalTarget(func(c *effects.Context) int {
							info, ok := c.SourcePermanent()
							if !ok {
								return 0
							}
							return info.Counters[doom]
						})(g, item); err != nil {
							return err
						}
						if len(ctx.LegalTargets()) == 0 {
							return nil
						}
						return effects.PutACounterOnThis(doom)(g, item)
					}),
				effects.TargetAny()),
		},
	})
}
