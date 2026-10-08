package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Sleep Paralysis — Enchantment Creature — Nightmare {5}{U}, 6/6:
//
//	"Flying
//	 Eerie — Whenever this creature or another enchantment you control
//	 enters and whenever you fully unlock a Room, tap up to one target
//	 creature and put a stun counter on it.
//	 Stun counters can't be removed from permanents your opponents control."
//
// The eerie half is the shared Eerie trigger (a creature that is an
// enchantment is "an enchantment you control", itself included). The
// lock is Spec.CounterRemovalLocks (#1824, ADR 0058's 2026-10-08
// amendment): a prohibition read at the one counter-removal choke
// point, so an opponent's stunned permanent keeps its counter through
// every untap step, and a removal effect or a counter-removal cost
// can't take it either. Your own permanents' stun counters come off
// as usual. The lock lifts the moment this leaves the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3a1f149-febe-4ff6-9e34-8b82ef21c2ee",
		Name:            "Fear of Sleep Paralysis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(Eerie("Fear of Sleep Paralysis — tap and stun up to one target creature (eerie)", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				if err := (TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
					return err
				}
				return (AddCounter{Target: ts[0].ID, Kind: game.CounterStun, N: 1}).Apply(ctx)
			}), TargetCreature("up to one target creature").WithCount(0, 1)),
		},
		CounterRemovalLocks: []game.CounterRemovalLock{{
			Counter: game.CounterStun,
			Locks: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller != source.Controller
			},
			Label: "Stun counters can't be removed from permanents your opponents control.",
		}},
	})
}
