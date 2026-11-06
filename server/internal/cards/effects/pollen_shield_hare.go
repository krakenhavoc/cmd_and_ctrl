package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pollen-Shield Hare // Hare Raising — Creature — Rabbit {1}{W}, 2/2 //
// Sorcery — Adventure {G} (Wilds of Eldraine, #4471):
//
//	Pollen-Shield Hare — "Creature tokens you control get +1/+1."
//	Hare Raising — "Target creature you control gains vigilance and gets
//	 +X/+X until end of turn, where X is the number of creatures you
//	 control. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The anthem is Intangible Virtue's +1/+1 half (layer 7c) over
// creature tokens the Hare's controller controls, re-read on every
// recompute. Hare Raising counts X once, as it resolves, the target
// included (CR 608.2h), and puts the pump and the vigilance on the target
// as one until-end-of-turn record, so a creature that enters later
// changes nothing (CR 611.2c).
//
// No simplification.
const pollenShieldHareOracleID = "3ef02f2b-b611-40c3-a1d5-8b34989643cf"

func init() {
	Register(Spec{
		OracleID:     pollenShieldHareOracleID,
		Name:         "Pollen-Shield Hare",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creature tokens you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b05CreatureTokenYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
	Register(Spec{
		OracleID:     pollenShieldHareOracleID + "#1",
		Name:         "Hare Raising",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "vigilance")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
			x := teamAvatarCreatures(ctx.Game, ctx.Item)
			return untilEndOfTurn(ctx, id, nil, "Hare Raising — +X/+X and vigilance",
				game.ModifyPTMod(x, x), game.AddKeywordsMod("vigilance"))
		})),
	})
}
