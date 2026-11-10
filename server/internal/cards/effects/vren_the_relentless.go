package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vren, the Relentless — Legendary Creature — Rat Rogue {2}{U}{B}, 3/4
// (Bloomburrow, #4467):
//
//	"Ward {2}
//	 If a creature an opponent controls would die, exile it instead.
//	 At the beginning of each end step, create X 1/1 black Rat creature
//	 tokens with "This token gets +1/+1 for each other Rat you control,"
//	 where X is the number of creatures that were exiled under your
//	 opponents' control this turn."
//
// The replacement is Garruk, Veiled Butcher's (rfOpponentCreatureWouldDie):
// the creature is exiled, so it never died and nothing that watches dying
// fires. It applies even if Vren leaves at the same time (the ruling).
// X is read as the trigger resolves from the engine's per-turn tally of
// creatures exiled from the battlefield under each player's control
// (PlayerTurnTally.CreaturesExiled): every exile counts, whatever caused it,
// tokens included. The Rat is the printed token with its own static
// ability (VrenRatToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1216900e-93e3-41a4-b354-9c81938639c2",
		Name:         "Vren, the Relentless",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: rfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Vren, the Relentless: exile instead of dying",
		}},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Vren, the Relentless — ward {2}"),
			On(game.EventBeginEndStep, AnyPlayer,
				"Vren, the Relentless — create X 1/1 black Rat tokens (X is the number of creatures exiled under your opponents' control this turn)",
				Do(Instruction(func(ctx *Context) error {
					x := 0
					for _, opp := range ctx.Opponents() {
						x += ctx.Game.TurnTallyFor(opp).CreaturesExiled
					}
					return CreateToken{Template: VrenRatToken(), N: x}.Apply(ctx)
				}))),
		},
	})
}
