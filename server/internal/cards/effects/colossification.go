package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Colossification — Enchantment — Aura for {5}{G}{G} (EDHREC rank
// 4020):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature gets +20/+20."
//
// The biggest single pump in the game, with the drawback written into
// the entry: the creature it lands on is tapped, so it does not
// attack the turn the Aura resolves and it cannot block before your
// next turn. The deck that plays it either has vigilance, an untapper,
// or an Aura-recursion plan.
//
// It is in the batch because the ETB is the whole tension of the card
// and it is an unusual shape: an Aura trigger that acts on ITS OWN
// HOST rather than drawing a card or hitting a target. The host is
// read off the attachment relation at resolution, which is the only
// correct time to read it — an Aura that has already fallen off (its
// creature died in response to the trigger, or a Totem Armor
// redirected it) taps nothing, because there is nothing it is
// attached to.
//
// Note it says "enchanted creature", not "target creature": nothing
// about the tap targets, so a creature with hexproof or protection
// from green that the Aura is somehow already on is still tapped.
//
// The +20/+20 is a layer 7c modification, not a base-P/T set, so it
// stacks with counters and with other anthems the way the printed card
// does — and it goes away the instant the Aura does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05c5ee65-651b-4bfb-b57f-305026507135",
		Name:         "Colossification",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(20, 20),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Colossification — tap enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					host, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || host.AttachedTo.ID == uuid.Nil {
						return nil
					}
					return TapTarget{Target: host.AttachedTo.ID}.Apply(NewContext(g, item))
				}),
		},
	})
}
