package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horn of Valhalla // Ysgard's Call — an adventure card (Commander Legends:
// Battle for Baldur's Gate, S82 GAP #5948):
//
//	Horn of Valhalla {1}{W}, Artifact — Equipment:
//	"Equipped creature gets +1/+1 for each creature you control.
//	 Equip {3}"
//
//	Ysgard's Call {X}{W}{W}, Sorcery — Adventure:
//	"Create X 1/1 white Soldier creature tokens."
//
// The Equipment's pump is counted live (PumpAttachedPer, Banshee's Blade's), over
// the creatures its controller controls including the equipped one, so it moves
// as creatures come and go. The adventure half is the "<oracle_id>#1" spec with
// an announced X (XMatters); the engine exiles the resolved adventure with
// permission to cast the Equipment (CR 715.3d).
func init() {
	const oracleID = "d02179a0-a6b7-49ef-a73e-cf11cabb896f"
	creaturesYouControl := func(g *game.Game, source *game.Card) int {
		return effects.CountControlled(g, source.Controller, func(c game.Card) bool { return c.IsCreature() })
	}
	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Horn of Valhalla",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttachedPer(1, 1, creaturesYouControl)},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Ysgard's Call",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: ctx.X()}.Apply(ctx)
		},
	})
}
