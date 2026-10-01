package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grievous Wound — Enchantment — Aura {3}{B}{B}:
//
//	"Enchant player
//	 Enchanted player can't gain life.
//	 Whenever enchanted player is dealt damage, they lose half their
//	 life, rounded up."
//
// An Aura on a PLAYER (Curse of Opulence's EnchantPlayer attachment).
// "Enchanted player can't gain life" is ADR 0107 §5's battlefield
// static in its enchanted-player form (CR 119.7, #1880), read off the
// Aura's attachment, so it ends the moment the Aura leaves.
//
// "Whenever enchanted player is dealt damage" triggers once per damage
// EVENT (CR 603.2c). The engine emits one damage event per source, so
// simultaneous damage from several sources — a combat damage step —
// would be several events; OncePerBatch collapses a batch into the one
// trigger the printed card has. The halving reads the life total as
// the trigger resolves (playerLosesHalfTheirLife, rounded up), and
// "they" is the player that was dealt the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d97b9d1-1138-40da-b593-8065ea6ed7b3",
		Name:         "Grievous Wound",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		CantGainLife: EnchantedPlayerCantGainLife(),
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Target
			}, "Grievous Wound — enchanted player loses half their life", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				return playerLosesHalfTheirLife(g, item, item.Trigger.Event.Target)
			})),
		},
	})
}
