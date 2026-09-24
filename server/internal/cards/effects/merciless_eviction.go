package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Eviction — Sorcery {4}{W}{B}:
//
//	"Choose one —
//	 • Exile all artifacts.
//	 • Exile all creatures.
//	 • Exile all enchantments.
//	 • Exile all planeswalkers."
//
// EXILE rather than destroy is the reason to play it over Wrath of
// God: no dies-triggers, no regeneration, no recursion, and
// indestructible doesn't help. Against the aristocrats payoffs in the
// catalog that difference decides games — a Blood Artist board wiped
// by Damnation drains the table, wiped by Eviction it doesn't.
//
// Four untargeted modes, so ChooseOne carries no target clauses: the
// engine validates the pick at announce and OnResolve reads it back
// with ctx.HasMode in printed order (CR 608.2c). Only one can be
// chosen here, but the loop is written for the general case because
// Farewell is these same sweeps under ChooseN.
func init() {
	Register(Spec{
		OracleID:     "3c8d4999-e18b-48d8-8ed9-f2feaa38300d",
		Name:         "Merciless Eviction",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Exile all artifacts."),
			Mode("Exile all creatures."),
			Mode("Exile all enchantments."),
			Mode("Exile all planeswalkers."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			sweeps := []CardPredicate{Artifact(), Creature(), Enchantment(), Planeswalker()}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (ExileAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
