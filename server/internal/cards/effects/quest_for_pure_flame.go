package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quest for Pure Flame — Enchantment {R}:
//
//	"Whenever a source you control deals damage to an opponent, you may
//	 put a quest counter on this enchantment.
//	 Remove four quest counters from this enchantment and sacrifice it:
//	 If any source you control would deal damage to a permanent or
//	 player this turn, it deals double that damage to that permanent or
//	 player instead."
//
// The counter trigger reads one damage event: a source you control
// dealing damage to an opponent (a player other than you). Each source
// hitting an opponent is its own event and its own "may", as printed;
// damage to an opponent's permanents does not count.
//
// The activated ability pays its counters and the sacrifice at announce
// (#625), then makes ADR 0108 §3's multiplier over your sources for the
// rest of the turn — Insult's, without the "can't be prevented".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71220cc2-5f3d-4c97-ae66-05b1b79adef1",
		Name:         "Quest for Pure Flame",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Amount <= 0 || ev.Target == source.Controller || g.PlayerByIDForEffect(ev.Target) == nil {
					return false
				}
				controller, ok := b34DamageSourceController(ev, g)
				return ok && controller == source.Controller
			}, "Quest for Pure Flame — put a quest counter", func(g *game.Game, item *game.StackItem) error {
				return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
			}), "Quest for Pure Flame — put a quest counter on it?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove four quest counters from this enchantment and sacrifice it: If any source you control would deal damage to a permanent or player this turn, it deals double that damage to that permanent or player instead.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(RemoveCountersFromThis("quest", 4), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MultiplyDamage{Factor: 2, Sources: game.DamageSourcesYours,
					Label: "Quest for Pure Flame — your sources deal double damage"}.Apply(NewContext(g, item))
			},
		}},
	})
}
