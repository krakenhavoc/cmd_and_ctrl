package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lost in the Maze — Enchantment {X}{U}{U} (Murders at Karlov Manor,
// #4469):
//
//	"Flash
//	 When this enchantment enters, tap X target creatures. Put a stun
//	 counter on each of those creatures you don't control. (If a
//	 permanent with a stun counter would become untapped, remove one from
//	 it instead.)
//	 Tapped creatures you control have hexproof."
//
// The enters trigger is counted by the X the enchantment was cast with:
// a CountFromX clause on a XFromCast row (game.xCountedTriggerTargets), so
// exactly X targets are chosen as the trigger goes on the stack (CR
// 601.2c), and with X = 0 there are none. Every target still legal as it
// resolves is tapped, then each of those not controlled by the
// enchantment's controller gets a stun counter. The hexproof is a layer 6
// grant keyed on Tapped, Saryth's shape: it follows a creature tapping and
// untapping.
//
// No simplification.
func init() {
	targets := TargetCreature("X target creatures")
	targets.CountFromX = true
	Register(Spec{
		OracleID:        "3a024389-91a1-460c-b4e8-3f212d2b8400",
		Name:            "Lost in the Maze",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		XMatters:        true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Lost in the Maze — tap X target creatures; put a stun counter on each of those you don't control",
			XFromCast: true,
			Targets:   targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
						continue
					}
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(t.ID); ok && c.Controller != item.Controller {
						if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			},
		}},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.Tapped && target.IsCreature()
			}, "hexproof"),
		},
	})
}
