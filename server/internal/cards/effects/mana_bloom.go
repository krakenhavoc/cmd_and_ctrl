package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Bloom — Enchantment {X}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"This enchantment enters with X charge counters on it.
//	 Remove a charge counter from this enchantment: Add one mana of any
//	 color. Activate only once each turn.
//	 At the beginning of your upkeep, if this enchantment has no charge
//	 counters on it, return it to its owner's hand."
//
// The mana ability's cost removes one charge counter and it may be
// activated once each turn (OncePerTurn, Ramos's shape). The upkeep
// trigger's "if" is an intervening if (CR 603.4): checked as the upkeep
// begins, before anyone has priority to remove a last counter (the
// ruling), and again as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "7c699180-d18e-4599-acd1-27b951d82136",
		Name:                       "Mana Bloom",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterCharge)},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				RemoveCounters: RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
			},
			Produced:    "{W|U|B|R|G}",
			Label:       "Remove a charge counter from this enchantment: Add one mana of any color",
			OncePerTurn: true,
		}},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{Name: "this enchantment has no charge counters on it",
				Test: func(ctx *Context) bool {
					info, ok := ctx.SourcePermanent()
					return ok && !info.Left && info.Counters[game.CounterCharge] == 0
				}},
				On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return ByYou(ev, source, lki, g) && source.Counters[game.CounterCharge] == 0
				}, "Mana Bloom — return it to its owner's hand (no charge counters)",
					func(g *game.Game, item *game.StackItem) error {
						return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
					})),
		},
	})
}
