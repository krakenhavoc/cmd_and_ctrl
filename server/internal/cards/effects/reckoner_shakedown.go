package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckoner Shakedown — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You may choose a nonland card
//	 from it. If you do, that player discards that card. If you don't,
//	 put two +1/+1 counters on a creature or Vehicle you control."
//
// The optional revealed-hand pick (#2115, ADR 0116's 2026-10-05
// amendment): the whole table sees the hand (CR 701.20a), and you may
// choose a nonland card or nothing. A chosen card is discarded. If you
// choose nothing — or there is no nonland card to choose — you choose a
// creature or Vehicle you control and it gets two +1/+1 counters. That
// choice is made after the pick (its 2022-02-18 ruling), it is not
// targeted, and with no creature or Vehicle nothing else happens (the
// same ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15159091-84ec-4b8c-8d84-8a8bb8baf27d",
		Name:         "Reckoner Shakedown",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player:   TargetedPlayer(ctx),
				Filter:   Nonland(),
				Label:    "nonland card",
				Optional: true,
				Then:     reckonerShakedownIfYouDont,
			}.Apply(ctx)
		},
	})
}

// reckonerShakedownIfYouDont is "If you don't, put two +1/+1 counters
// on a creature or Vehicle you control."
var reckonerShakedownIfYouDont = RevealedPickThen("revealed-pick/reckoner-shakedown-two-counters",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) > 0 {
			return nil
		}
		you := ctx.Controller()
		var options []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == you && (c.IsCreature() || c.HasSubtype("Vehicle")) {
				options = append(options, c.InstanceID)
			}
		}
		if len(options) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     pick.Source,
			Question:   "Reckoner Shakedown — put two +1/+1 counters on a creature or Vehicle you control",
			Cards:      options,
			Min:        1,
			Max:        1,
			// Re-checked on submit: a permanent can leave between the
			// question and the answer.
			Zone: game.ZoneBattlefield,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				for _, id := range picked {
					// You put them (CR 120.3d's placer).
					if err := g.AddCounterByForEffect(you, id, "+1/+1", 2); err != nil {
						return err
					}
				}
				return nil
			},
		})
		return nil
	})
