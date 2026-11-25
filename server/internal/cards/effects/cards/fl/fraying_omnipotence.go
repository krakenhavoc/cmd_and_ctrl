package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fraying Omnipotence — Sorcery {3}{B}{B} (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Each player loses half their life, then discards half the cards in their
//	 hand, then sacrifices half the creatures they control of their choice.
//	 Round up each time."
//
// Pox Plague's three instructions, each finishing before the next begins (CR
// 608.2c) and each count worked out when its instruction starts, rounded up:
// the life (S82OldB198EachPlayerLosesHalfTheirLifeRoundedUp), the hand
// (DiscardPrompt.HalfHandUp, each player's own choice in APNAP order), then the
// creatures they control now, each owner choosing its own (the Rush of Dread
// sacrifice seam). The sacrifice is a registered continuation so a table
// waiting on a discard is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b00068d-ad07-417c-bbb3-5f980615fb06",
		Name:         "Fraying Omnipotence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.S82OldB198EachPlayerLosesHalfTheirLifeRoundedUp),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
					game.DiscardPrompt{
						Source:     ctx.Source(),
						HalfHandUp: true,
						Question:   "Fraying Omnipotence — discard half the cards in your hand, rounded up",
					},
					frayingOmnipotenceSacrifice, []uuid.UUID{ctx.Controller()})
			}),
		),
	})
}

// frayingOmnipotenceSacrifice is "sacrifices half the creatures they control of
// their choice, round up", run once the last discard has landed.
var frayingOmnipotenceSacrifice = effects.RunThen("prompt-run/s82-old-b198-fraying-omnipotence-sacrifice", func(ctx *effects.Context, _ game.RunDone) error {
	var asks []uuid.UUID
	for _, seat := range effects.TablePlayers(ctx) {
		if p := ctx.PlayerByID(seat); p == nil || p.Eliminated {
			continue
		}
		owed := (len(effects.PermanentsControlledByMatching(ctx.Game, seat, effects.Creature())) + 1) / 2
		for i := 0; i < owed; i++ {
			asks = append(asks, seat)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, effects.SacrificeSpec("a creature", effects.Creature()),
		"Fraying Omnipotence — sacrifice half the creatures you control, rounded up")
	return nil
})
