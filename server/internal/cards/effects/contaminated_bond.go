package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contaminated Bond — Enchantment — Aura {1}{B} (Mirrodin, #5606):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks or blocks, its controller loses
//	 3 life."
//
// The Aura's trigger is Super-Soldier Serum's attacks-or-blocks
// condition on the enchanted creature, with the block counted once
// however many attackers the creature blocks (CR 509.3a, the numbered
// block events). "Its controller" is read as the trigger resolves: the
// creature's controller then, or as it last existed on the battlefield
// if it has left (CR 608.2h), and the player who declared it otherwise.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56aeaee8-e014-485e-b7ad-e26ade663bcf",
		Name:         "Contaminated Bond",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				switch ev.Kind {
				case game.EventAttack:
					return source.IsAttachedTo(ev.CardID)
				case game.EventBlock:
					return source.IsAttachedTo(ev.CardID) && ev.Amount <= 1
				}
				return false
			}, "Contaminated Bond — the enchanted creature's controller loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					loser := ev.Actor
					if c, ok := g.LookupCardForEffect(ev.CardID); ok && onBattlefield(g, ev.CardID) {
						loser = c.Controller
					} else if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
						loser = info.Controller
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, loser, -3)
				}),
		},
	})
}
