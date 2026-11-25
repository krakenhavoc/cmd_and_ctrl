package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Start // Finish — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Start {2}{W}: "Create two 1/1 white Warrior creature tokens with vigilance."
//	Finish {2}{B}: "Aftermath. As an additional cost to cast this spell,
//	 sacrifice a creature. Destroy target creature."
//
// A split card with an aftermath half: two registered faces under the card's
// oracle ID and its "#1" suffix, Finish castable only from a graveyard and
// exiled as it resolves (CR 702.127a). Finish's creature sacrifice is paid on
// the cast (CR 118.8); the destroy then takes any target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ae2750e-3650-42df-a539-ea4cb72fa136",
		Name:         "Start",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior with vigilance"), N: 2}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:       "4ae2750e-3650-42df-a539-ea4cb72fa136#1",
		Name:           "Finish",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerSacOutlet},
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target creature"),
		OnResolve:      effects.DestroyTheTargetPermanent,
	})
}
