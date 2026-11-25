package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wings of Hubris — Artifact — Equipment {2} (Theros Beyond Death, #5514):
//
//	"Equipped creature has flying.
//	 Sacrifice this Equipment: Equipped creature can't be blocked this turn.
//	 Sacrifice it at the beginning of the next end step.
//	 Equip {1}"
//
// Lightning Spear's sacrifice cost with Rogue's Passage's restriction. The
// "equipped creature" is read as the Equipment last was (CR 608.2h), since
// sacrificing it as the cost unattaches it: the attached object rides the
// source's last-known permanent. "It" in the last sentence is that creature
// (the Equipment is already gone): a delayed trigger sacrifices that same
// object at the beginning of the next end step, if it is still on the
// battlefield under the ability's controller (CR 603.7c). An Equipment that
// was attached to nothing makes neither.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b56efee-bdc8-4516-9c0d-b28ca1b0a918",
		Name:         "Wings of Hubris",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice this Equipment: Equipped creature can't be blocked this turn. Sacrifice it at the beginning of the next end step.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.SacrificeThis(),
				Effect:  wingsOfHubrisResolve,
			},
			effects.EquipAbility("{1}"),
		},
	})
}

// wingsOfHubrisResolve makes the creature the Equipment was attached to
// unblockable this turn and schedules its sacrifice.
func wingsOfHubrisResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.AttachedObject == nil {
		return nil
	}
	obj := *info.AttachedObject
	c, found := g.LookupCardForEffect(obj.ID)
	if !found || !effects.OnBattlefield(g, obj.ID) || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	if err := (effects.RestrictUntilEOT{
		Target:       obj.ID,
		Restrictions: game.CantBeBlocked,
		Label:        "Wings of Hubris — can't be blocked",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Wings of Hubris — sacrifice that creature",
		Body:   effects.Cmdr2S12SoulstokeSacrifice,
		Params: game.EffectParams{Object: obj},
	}.Apply(ctx)
}
