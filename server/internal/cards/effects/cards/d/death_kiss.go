package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Kiss — Creature — Beholder {5}{R}, 5/5 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Whenever a creature an opponent controls attacks one of your opponents, double
//	 its power until end of turn.
//	 {X}{X}{R}: Monstrosity X. (If this creature isn't monstrous, put X +1/+1
//	 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, goad up to X target creatures your
//	 opponents control."
//
// The first ability fires per attacker: a creature a player other than Death Kiss's
// controller controls, declared as an attacker against a player (CR 508.3a) who is
// also an opponent of Death Kiss's controller. The doubling is the shared "double
// its power" (+X/+0 where X is its power as the trigger resolves, CR 701.10b),
// pinned to that attacker. Monstrosity X is the shared constructor over {X}{X}{R};
// the becomes-monstrous trigger sizes its target clause by the announced X
// (Vitality Hunter's shape, CR 701.37c) and goads each target still legal as it
// resolves; X of 0 leaves the trigger with no targets.
//
// No simplification.
func init() {
	becomes := effects.WhenBecomesMonstrous("Death Kiss — goad up to X target creatures your opponents control", effects.GalvanizingTenorGoad)
	becomes.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		x := tc.Amount()
		if x <= 0 {
			return nil
		}
		return effects.TargetCreature("up to X target creatures your opponents control", effects.OpponentControls()).WithCount(0, x)
	}
	effects.Register(effects.Spec{
		OracleID:     "42410e81-f4a6-4501-82e4-6e745934976c",
		Name:         "Death Kiss",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Activated:    []effects.ActivatedAbility{effects.MonstrosityX(effects.ManaCost("{X}{X}{R}"))},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
					return false
				}
				p := effects.AttackedPlayer(g, ev)
				if p == uuid.Nil || p == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller != source.Controller
			}, "Death Kiss — double that creature's power until end of turn", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return effects.DoublePowerUntilEOT(effects.NewContext(g, item), item.Trigger.Event.CardID, "Death Kiss — double power")
			}),
			becomes,
		},
	})
}
