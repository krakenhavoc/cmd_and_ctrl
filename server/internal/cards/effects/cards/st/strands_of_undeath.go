package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strands of Undeath — Enchantment — Aura {3}{B} (S82 GAP, #5948):
//
//	"Enchant creature
//	 When this Aura enters, target player discards two cards.
//	 {B}: Regenerate enchanted creature."
//
// Dark Privilege's regenerate-the-host and Shrieking Grotesque's targeted
// discard. The discard target is the trigger's, the Aura's own target is the
// enchanted creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "897c2c66-fedc-4201-b187-db79414583fb",
		Name:         "Strands of Undeath",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Discards: 2},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Strands of Undeath — target player discards two cards", strandsOfUndeathDiscard),
				effects.TargetPlayer("target player")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Regenerate enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.RegenerateAttachedHost,
		}},
	})
}

func strandsOfUndeathDiscard(g *game.Game, item *game.StackItem) error {
	player, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
	if !ok {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 2})
	return nil
}
