package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skewer Slinger — Creature — Dwarf Knight {1}{R}, 1/3 (Wilds of
// Eldraine, #4471):
//
//	"Reach
//	 Whenever this creature blocks or becomes blocked by a creature, this
//	 creature deals 1 damage to that creature."
//
// Lim-Dûl's Cohort's trigger: the lock-in emits one EventBlock per
// (blocker, attacker) pair (#830), so it fires once for each creature the
// Slinger blocks or is blocked by. The Slinger deals the damage as an
// object, so a Slinger that has left the battlefield still deals it; a
// creature that has left the battlefield takes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e8433a2-ad71-4d30-84fe-12b709197824",
		Name:            "Skewer Slinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return skewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Skewer Slinger — 1 damage to that creature", skewerSlingerPokes),
		},
	})
}

// skewerSlingerFoe is the other creature of a block pair that names the
// Slinger, or uuid.Nil when the pair does not name it.
func skewerSlingerFoe(ev game.Event, self uuid.UUID) uuid.UUID {
	switch self {
	case ev.CardID:
		return ev.Target
	case ev.Target:
		return ev.CardID
	}
	return uuid.Nil
}

func skewerSlingerPokes(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	foe := skewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
	if !ok || foe == uuid.Nil {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: foe, Amount: 1}.Apply(ctx)
}
