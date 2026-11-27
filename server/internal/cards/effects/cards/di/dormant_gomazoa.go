package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dormant Gomazoa — Creature — Jellyfish {1}{U}{U}, 5/5 (Rise of the Eldrazi,
// S82 OLD batch 484, #5581):
//
//	"Flying
//	 This creature enters tapped.
//	 This creature doesn't untap during your untap step.
//	 Whenever you become the target of a spell, you may untap this creature."
//
// The untap is Battered Golem's optional untap of itself. The trigger watches
// a player becoming a target (the Gomazoa's controller) of a spell on the
// stack, whoever controls that spell, once for each such spell; an ability
// that targets them does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "bc5a7015-c13b-4d92-8afe-50b7956c14dd",
		Name:                  "Dormant Gomazoa",
		Completeness:          effects.CompletenessFull,
		PrintedKeywords:       []string{"flying"},
		Replacements:          []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CardID == uuid.Nil && ev.Target == source.Controller &&
					ev.TargeterKind == game.StackItemSpell
			}, "Dormant Gomazoa — untap this creature", effects.B31UntapSelf), "Dormant Gomazoa — untap it?"),
		},
	})
}
