package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guardian Archon — Creature — Archon {4}{W}{W}, 5/5 (Commander 2021):
//
//	"Flying
//	 As this creature enters, secretly choose an opponent.
//	 Reveal the player you chose: You and target permanent you control each
//	 gain protection from the chosen player until end of turn. Activate only
//	 once."
//
// The choice is made as the permanent enters (CR 614.12, the AsEnters hook,
// not a trigger) and stored on it, hidden from every seat but its chooser
// until the reveal, the way Stalking Leonin's is. Revealing is the cost
// (CR 602.2b): only the player who chose can pay it, and only while a
// choice exists. "Activate only once" is spent by the activation. On
// resolution the controller and the target permanent each gain protection
// from that one player until end of turn (CR 702.16); a permanent that is
// no longer a legal target gets nothing, the controller still does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7748659-0206-448f-8567-83565cb0b51a",
		Name:            "Guardian Archon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			ctx.Game.QueueSecretPlayerChoiceForEffect(card.Controller, card.InstanceID,
				"Guardian Archon — secretly choose an opponent (only you see the answer until you reveal it)",
				effects.Opponents.Seats(ctx, card.Controller))
			return nil
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Reveal the player you chose: You and target permanent you control each gain protection from the chosen player until end of turn. Activate only once.",
			Purpose: game.Purpose{
				Answers: game.AnswerProtect,
				Targets: game.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit}),
			},
			Cost:      effects.RevealSecretChoice(),
			Condition: effects.YouSecretlyChoseForThis,
			OnlyOnce:  true,
			Targets:   effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				chosen, _ := g.SecretChosenPlayerOf(item.SourceCardID)
				if chosen == uuid.Nil {
					return nil
				}
				quality := game.ProtectionFromPlayer(chosen)
				if err := (effects.GainPlayerKeyword{
					Player:  item.Controller,
					Keyword: quality,
					Label:   "Guardian Archon — you gain protection from the chosen player",
				}).Apply(ctx); err != nil {
					return err
				}
				if ts := ctx.LegalTargets(); len(ts) > 0 {
					return effects.GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{quality}, Label: "Guardian Archon — protection from the chosen player"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
