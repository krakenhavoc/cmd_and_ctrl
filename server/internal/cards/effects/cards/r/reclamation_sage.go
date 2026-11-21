package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclamation Sage — 2/1 Elf Shaman for {2}{G}:
//
//	"When Reclamation Sage enters the battlefield, you may destroy
//	target artifact or enchantment."
//
// S19 sub-PR 3 shipped this on the trigger pipeline with a sandbox
// auto-picker ("first opponent artifact or enchantment"). S20
// sub-PR 2 replaces that with a real target: Targets declares the
// clause, so the engine drops the trigger when nothing qualifies
// (CR 603.3d), asks "you may" first, then queues a pick_target
// prompt whose legal set the controller clicks on the board. Any
// artifact or enchantment qualifies — including your own, as in
// paper. The chosen ref lands in item.Targets and is re-checked at
// resolution (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "032ec6e2-6cc3-4a97-9cc7-3233f5e11904",
		Name:         "Reclamation Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Key:     "Reclamation Sage — destroy target artifact or enchantment",
			Effect:  effects.DestroyChosenPermanent,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Reclamation Sage — destroy an artifact or enchantment?",
			},
		}},
	})
}
