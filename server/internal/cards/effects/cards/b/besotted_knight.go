package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Besotted Knight // Betroth the Beast — Creature — Human Knight
// {3}{W}, 3/3 // Sorcery — Adventure {W} (CR 715):
//
//	Besotted Knight — no rules text.
//	Betroth the Beast — "Create a Royal Role token attached to target
//	 creature you control. (Enchanted creature gets +1/+1 and has ward
//	 {1}.)"
//
// Both faces register: the vanilla creature so the card does not wear
// the "unimplemented" badge in hand (its Adventure half has text), and
// the Adventure under "<oracle_id>#1". The Adventure lifecycle (exile
// on resolution, cast the creature later) is the engine's.
//
// No simplifications.
const besottedKnightOracleID = "c7b6d2cd-9105-404b-88f0-a67037fb2120"

func init() {
	effects.Register(effects.Spec{
		OracleID:     besottedKnightOracleID,
		Name:         "Besotted Knight",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     besottedKnightOracleID + "#1",
		Name:         "Betroth the Beast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				return effects.CreateRoleToken{Role: effects.RoleRoyal, Host: id}.Apply(ctx)
			}
			return nil
		},
	})
}
