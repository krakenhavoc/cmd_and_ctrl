package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pox Plague — Sorcery {B}{B}{B}{B}{B} (Secrets of Strixhaven, #4456):
//
//	"Each player loses half their life, then discards half the cards in
//	 their hand, then sacrifices half the permanents they control of their
//	 choice. Round down each time."
//
// Three instructions done to every player, one after the other (CR 608.2c),
// each finishing before the next begins, and each count is worked out when
// its instruction starts:
//
//   - Each player (the caster included) loses half their life, rounded down.
//   - Each player discards half the cards in their hand, rounded down
//     (DiscardPrompt.HalfHand), chosen by that player, asked in APNAP order
//     (EachPlayerDiscardsThenKeyedForEffect); a hand of one card discards none.
//   - Once every discard has landed, each player sacrifices half the
//     permanents they control, rounded down, of their choice (CR 701.21a):
//     APNAP choices, each seeing the earlier ones, and every chosen permanent
//     leaves together (CR 101.4, 608.2e). The count is read after the discards.
//
// The sacrifice is a registered continuation (poxPlagueSacrifice), so a table
// waiting on any seat's choice is a restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c505d304-9428-4532-87a2-167c69f16969",
		Name:         "Pox Plague",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(socB32EachPlayerLosesHalfTheirLife),
			Instruction(func(ctx *Context) error {
				return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
					game.DiscardPrompt{
						Source:   ctx.Source(),
						HalfHand: true,
						Question: "Pox Plague — discard half the cards in your hand",
					},
					poxPlagueSacrifice, []uuid.UUID{ctx.Controller()})
			}),
		),
	})
}

// poxPlagueSacrifice is "sacrifices half the permanents they control of their
// choice", run once the last discard has landed: each seat owes half its
// permanents, rounded down, counted now.
var poxPlagueSacrifice = RunThen("prompt-run/pox-plague-sacrifice", func(ctx *Context, _ game.RunDone) error {
	var asks []uuid.UUID
	for _, seat := range tablePlayers(ctx) {
		if p := ctx.PlayerByID(seat); p == nil || p.Eliminated {
			continue
		}
		owed := len(permanentsControlledByMatching(ctx.Game, seat, nil)) / 2
		for i := 0; i < owed; i++ {
			asks = append(asks, seat)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, nil, "Pox Plague — sacrifice half the permanents you control")
	return nil
})
