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
//     yes → the trigger goes on the stack; no → trigger drops.
//   - Sandbox auto-pick: first opponent-controlled artifact or
//     enchantment, scanning g.BattlefieldCardsForEffect in the
//     order the cards arrived. Real target-picker UI is S20.
//   - The pick is stamped onto item.Targets when the trigger is
//     put on the stack (Build runs on "yes"), so the stack overlay
//     shows it and the CR 608.2b re-check fizzles the trigger if
//     the target leaves before resolution.
//   - If no legal target exists at fire time, the prompt still
//     queues (HasLegalTarget flags the warning) and a "yes" builds
//     nothing — the trigger never reaches the stack.
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
				return destroyTargetTrigger(source, "Reclamation Sage — destroy target artifact or enchantment", target)
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
