package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boon of the Spirit Realm — Enchantment {3}{W}{W} (EDHREC rank
// 4475):
//
//	"Constellation — Whenever this enchantment or another enchantment
//	 you control enters, put a blessing counter on this enchantment.
//	 Creatures you control get +1/+1 for each blessing counter on this
//	 enchantment."
//
// A growing anthem for the enchantress deck. It arrives already at
// +1/+1 — its own entry is a constellation trigger, which is what
// "this enchantment OR another" means and is the half a card file
// would drop — and every enchantment after it is another point.
//
// Two abilities that are really one engine:
//
//   - The constellation trigger watches EventETB for an enchantment
//     its controller controls, the Boon included. It uses the stack,
//     so a response window opens, and the counter is skipped if the
//     Boon has left by the time it resolves (nothing to put it on).
//   - The anthem is a layer 7c static whose size is READ LIVE off the
//     Boon's blessing counters on every recompute. That is what makes
//     it grow: no snapshot, no bookkeeping, and removing the
//     enchantment removes the whole bonus at once.
//
// The counter kind is "blessing" — a custom counter, not a +1/+1
// counter — so proliferate grows the anthem, Hardened Scales does
// nothing to it, and a Solemnity shuts the card off entirely, all as
// printed.
//
// "Creatures you control" has no "other", and the Boon is not a
// creature anyway, so there is no self-exclusion to write. An
// enchantment CREATURE entering both triggers the counter and takes
// the anthem.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7731217e-cdbb-4c38-a9c1-083c96f51931",
		Name:         "Boon of the Spirit Realm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsEnchantment()
			}, "Boon of the Spirit Realm — a blessing counter (constellation)", func(g *game.Game, item *game.StackItem) error {
				if !b09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "blessing", N: 1}.Apply(NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					source.Counters["blessing"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters["blessing"]
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
