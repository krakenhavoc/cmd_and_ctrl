package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Find // Finality — split card (Modern Horizons 3 Commander, #5948):
//
//	Find {B/G}{B/G} — Sorcery: "Return up to two target creature cards from
//	  your graveyard to your hand."
//	Finality {4}{B}{G} — Sorcery: "You may put two +1/+1 counters on a
//	  creature you control. Then all creatures get -4/-4 until end of turn."
//
// Each half is an ADR 0034 face (ADR 0103): Find under the bare oracle ID,
// Finality under "#1". Only one half can be cast (CR 709.3; the card has no
// fuse). Find is Blood Fountain's return of up to two creature cards.
// Finality is Choking Miasma's shape: printed order is resolution order (CR
// 608.2c), so the optional counters go on first and the creature they land
// on is already bigger when the -4/-4 arrives. The counters are not
// targeted: the creature is chosen as the spell resolves, and with none to
// choose the clause does nothing. The -4/-4 locks the set of creatures as it
// begins (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "212af747-2ac1-4d33-9456-8b3cc8988f5d",
		Name:         "Find",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "212af747-2ac1-4d33-9456-8b3cc8988f5d#1",
		Name:         "Finality",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 4}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChoosePermanents{
					Question:   "Finality — you may put two +1/+1 counters on a creature you control",
					Candidates: effects.R34UpToOneCreatureYouControl,
					Keyed:      finalityChosen,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -4, Toughness: -4, Label: "Finality — -4/-4 until end of turn"}.Apply(ctx)
			}),
		),
	})
}

// finalityChosen puts the two counters on the creature chosen, keyed (#4290).
var finalityChosen = effects.RegisterPermanentsPick("prompt-run/finality-counters", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
