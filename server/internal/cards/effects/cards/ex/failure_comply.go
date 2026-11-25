package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Failure // Comply — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Failure {1}{U}: "Return target spell to its owner's hand."
//	Comply {W}: "Aftermath. Choose a card name. Until your next turn, your
//	 opponents can't cast spells with the chosen name."
//
// Two registered faces. Failure bounces the targeted spell (a copy ceases to
// exist, CR 707.10a). Comply asks for the name as it resolves (the name is free
// text, CR 201.2) and bars each opponent from casting spells with it until your
// next turn begins (CR 611.2b); it is cast only from a graveyard and exiled as
// it resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed8184f8-f6a7-4f97-9f4e-bf72077af754",
		Name:         "Failure",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.ReturnSpellToHand{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "ed8184f8-f6a7-4f97-9f4e-bf72077af754#1",
		Name:         "Comply",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.OldB243ChooseACardNameCastBan("Comply"),
	})
}
