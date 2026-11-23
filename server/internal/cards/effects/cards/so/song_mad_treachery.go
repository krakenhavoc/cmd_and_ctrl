package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Song-Mad Treachery // Song-Mad Ruins — Modal DFC. Front: Sorcery
// {3}{R}{R}: "Gain control of target creature until end of turn. Untap
// that creature. It gains haste until end of turn." Back: Land
// (mdfc_lands.go).
//
// The Act of Treason shape: a layer 2 control effect that ends at cleanup,
// the untap, and the haste grant (the creature is summoning-sick under its
// new controller, CR 302.6, which the haste lifts). The shared
// threatenAndGrant runs the three in that order. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81b61770-2ed5-4a50-84d0-97790002fc5a",
		Name:         "Song-Mad Treachery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.ThreatenAndGrant(ctx, id, "Song-Mad Treachery", "haste")
		},
	})
}
