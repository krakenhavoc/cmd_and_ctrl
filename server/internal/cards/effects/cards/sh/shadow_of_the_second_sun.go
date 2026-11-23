package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow of the Second Sun — Enchantment — Aura {4}{U}{U} (Modern Horizons
// 3):
//
//	"Enchant player
//	 At the beginning of each of enchanted player's postcombat main phases,
//	 there is an additional beginning phase after this phase. (The end step
//	 happens after the added untap, upkeep, and draw steps.)"
//
// Curse of Clinging Webs' enchant-player shape with Sphinx of the Second Sun's
// added beginning phase (AddPhases; untap, upkeep and draw for the ACTIVE
// player, which is the enchanted player on their own turn, CR 500.8: the most
// recently added phase comes first). The step trigger watches the enchanted
// player's postcombat main phases, any added one included, and the Aura's
// controller controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ee9d018-02c7-43e7-ba2e-95e36d8ac0e8",
		Name:         "Shadow of the Second Sun",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventStepBegan && ev.Step == game.StepPostcombatMain &&
					source.AttachedTo.Kind == game.TargetPlayer && ev.Actor == source.AttachedTo.ID
			}, "Shadow of the Second Sun — an additional beginning phase after this phase", effects.Do(effects.AddPhases{
				Anchor: game.PhaseAnchor{Kind: game.AnchorThisPhase},
				Kinds:  []game.PhaseKind{game.PhaseKindBeginning},
			})),
		},
	})
}
