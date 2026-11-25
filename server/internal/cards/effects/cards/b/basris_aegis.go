package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basri's Aegis — Sorcery {2}{W}{W} (Core Set 2021, #5511):
//
//	"Put a +1/+1 counter on each of up to two target creatures. You may search
//	 your library and/or graveyard for a card named Basri, Devoted Paladin,
//	 reveal it, and put it into your hand. If you search your library this way,
//	 shuffle."
//
// The counters go on each target still legal as it resolves (CR 608.2b). The
// search is its own sentence, the planeswalker-fetch cycle's keyed "you may"
// answer (zones named first, shuffle only after a library search); it runs even
// with no creature to target, since "up to two" allows none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9825dab8-ca1a-44ea-b1d7-77c66f00351f",
		Name:         "Basri's Aegis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, id := range effects.LegalTargetCardIDs(ctx) {
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(effects.S82OldB108MayFetchNamed("Basri's Aegis", "Basri, Devoted Paladin")),
		),
	})
}
