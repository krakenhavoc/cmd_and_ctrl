package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reclamation Sage — 2/1 Elf Shaman for {2}{G}:
//
//	"When Reclamation Sage enters the battlefield, you may destroy
//	target artifact or enchantment."
//
// S19 sub-PR 3:
//   - OptionalPrompt drives the "you may" gate. Controller picks
//     yes → effect runs; no → trigger drops.
//   - Sandbox auto-pick: first opponent-controlled artifact or
//     enchantment, scanning g.BattlefieldCardsForEffect in the
//     order the cards arrived. Real target-picker UI is S20.
//   - If no legal target exists at fire time, the prompt still
//     queues (controller may want to confirm visibility) but the
//     yes-branch silently no-ops.
func init() {
	Register(Spec{
		OracleID: "032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		Name:     "Reclamation Sage",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				target := pickFirstOpponentNonland(g, source.Controller, true /* artifact */, true /* enchantment */, false /* land */)
				if target == uuid.Nil {
					return nil
				}
				ctx := NewContext(g, nil)
				_ = DestroyTarget{Target: target}.Apply(ctx)
				return nil
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Reclamation Sage — destroy target artifact or enchantment?",
			},
			// Warn the chooser when there's no opponent artifact /
			// enchantment to hit — the sandbox auto-picker only targets
			// opponents (S20 ships the real picker), so "Yes" would
			// silently no-op without this hint.
			HasLegalTarget: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return pickFirstOpponentNonland(g, source.Controller, true, true, false) != uuid.Nil
			},
		}},
	})
}

// pickFirstOpponentNonland returns the first battlefield card NOT
// controlled by `controller` whose type matches one of the enabled
// flags. Used by Reclamation Sage / Acidic Slime auto-targeting in
// S19 sub-PR 3 — the sandbox simplification stays consistent across
// the ETB-destroy cards. Returns uuid.Nil when no candidate exists.
//
// Caller must hold g.mu.
func pickFirstOpponentNonland(
	g *game.Game,
	controller uuid.UUID,
	wantArtifact, wantEnchantment, wantLand bool,
) uuid.UUID {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller {
			continue
		}
		if wantArtifact && c.IsArtifact() {
			return c.InstanceID
		}
		if wantEnchantment && c.IsEnchantment() {
			return c.InstanceID
		}
		if wantLand && c.IsLand() {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
