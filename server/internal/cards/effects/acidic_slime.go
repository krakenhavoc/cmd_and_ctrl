package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acidic Slime — 2/2 Ooze for {3}{G}{G}:
//
//	"Deathtouch. When Acidic Slime enters the battlefield, destroy
//	target artifact, enchantment, or land."
//
// S19 sub-PR 3:
//   - Mandatory ETB trigger (no "you may").
//   - Sandbox auto-pick: first opponent-controlled artifact /
//     enchantment / land, scanning g.BattlefieldCardsForEffect in
//     the order the cards arrived. The S20 smart-cast UI replaces
//     this with a real target picker.
//   - The pick happens in Build (when the trigger is put on the
//     stack, CR 603.3d) and is stamped onto item.Targets, so the
//     stack overlay shows it and the CR 608.2b re-check fizzles the
//     trigger if the target leaves before resolution.
//   - If no legal target exists at Build time, Build returns nil
//     and the trigger is never put on the stack (CR 603.3d).
func init() {
	Register(Spec{
		OracleID:        "21f45043-5419-4019-8b6c-e5294bd5f549",
		Name:            "Acidic Slime",
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				target := pickFirstOpponentNonland(g, source.Controller, true, true, true)
				if target == uuid.Nil {
					return nil
				}
				return destroyTargetTrigger(source, "Acidic Slime — destroy target permanent", target)
			},
		}},
	})
}

// destroyTargetTrigger builds a triggered-ability stack item that
// destroys the single card target on resolution. Shared by the
// ETB-destroy cards (Acidic Slime, Reclamation Sage): the target is
// recorded on the item so the engine's CR 608.2b re-check and the
// stack overlay both see it; the Effect reads it back off the item
// rather than closing over the ID, so an undo-restored item still
// points at the right card.
func destroyTargetTrigger(source *game.Card, label string, target uuid.UUID) *game.StackItem {
	item := game.NewTriggeredItem(source, label, func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
	})
	item.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	return item
}
