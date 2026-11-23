package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vile Rebirth — Instant {B} (Magic 2013, #5568):
//
//	"Exile target creature card from a graveyard. Create a 2/2 black Zombie
//	 creature token."
//
// Cemetery Reaper's activated ability as a spell: a creature card in any
// graveyard. With its one target gone the spell does nothing, token
// and all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6845a62f-e0f1-448c-beeb-abac2dabd975",
		Name:         "Vile Rebirth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInAnyGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.BlackZombieToken(), N: 1}.Apply(ctx)
		},
	})
}
