package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acidic Slime — 2/2 Ooze for {3}{G}{G}:
//
//   "Deathtouch. When Acidic Slime enters the battlefield, destroy
//   target artifact, enchantment, or land."
//
// S19 sub-PR 3:
//   - Mandatory ETB trigger (no "you may"). Build always runs.
//   - Sandbox auto-pick: first opponent-controlled artifact /
//     enchantment / land, scanning g.BattlefieldCardsForEffect in
//     the order the cards arrived. The S20 smart-cast UI replaces
//     this with a real target picker.
//   - If no legal target exists, the trigger no-ops silently.
func init() {
	Register(Spec{
		OracleID:        "ff34ad03-9d5d-46a3-b00d-c6f6e92e3d4c",
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
				ctx := NewContext(g, nil)
				_ = DestroyTarget{Target: target}.Apply(ctx)
				return nil
			},
		}},
	})
}
