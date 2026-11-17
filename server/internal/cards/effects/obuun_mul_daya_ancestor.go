package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obuun, Mul Daya Ancestor — Legendary Creature — Elf Spirit {1}{R}{G}{W},
// 3/3 (Zendikar Rising Commander, #5509):
//
//	"At the beginning of combat on your turn, up to one target land you
//	 control becomes an X/X Elemental creature with trample and haste until
//	 end of turn, where X is Obuun's power. It's still a land.
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature."
//
// The combat trigger is Destiny Spinner's animation on a beginning-of-combat
// trigger: one ScopedEffectFor record pinned to the land (CR 611.2c, CR
// 400.7) that adds the Creature type and the Elemental subtype (layer 4),
// trample and haste (layer 6) and sets the base power and toughness to X
// (layer 7b). Nothing is removed, so the land keeps its types and abilities
// (the ruling), and a later effect that sets power and toughness overwrites
// it by timestamp while counters and pumps apply on top (the rulings). "Up to
// one" is the clause's zero minimum. X is Obuun's power as the ability
// resolves (CR 608.2h): read from the battlefield while Obuun is there, and
// from his last known information once he has left. It is locked in then
// and does not follow his power later in the turn.
//
// Landfall (CR 207.2c) is Bristly Bill's targeted counter: any creature at
// the table, Obuun included.
//
// No simplification.
func init() {
	combat := Targeting(
		AtBeginningOfYourCombat("Obuun, Mul Daya Ancestor — up to one target land you control becomes an X/X Elemental creature with trample and haste until end of turn, where X is Obuun's power",
			obuunAnimateLand),
		TargetPermanent("up to one target land you control", Land(), YouControl()).WithCount(0, 1))
	landfall := Targeting(
		Landfall("Obuun, Mul Daya Ancestor — put a +1/+1 counter on target creature (landfall)", obuunLandfallCounter),
		TargetCreature("target creature"))
	Register(Spec{
		OracleID:     "28fcf855-6372-4d6b-b3fd-76f85247f309",
		Name:         "Obuun, Mul Daya Ancestor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{combat, landfall},
	})
}

// obuunAnimateLand makes the chosen land an X/X Elemental creature with
// trample and haste until end of turn, X Obuun's power now or as he last
// existed on the battlefield. No pick, or a land that left in response, does
// nothing.
func obuunAnimateLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	land := firstLegalCardTarget(ctx)
	if land == uuid.Nil {
		return nil
	}
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	x := info.Power
	return ScopedEffectFor{
		Target: land,
		Mods: []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
			game.SetBasePowerMod(x),
			game.SetBaseToughnessMod(x),
			game.AddKeywordsMod("trample", "haste"),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Obuun, Mul Daya Ancestor — an X/X Elemental creature with trample and haste until end of turn",
	}.Apply(ctx)
}

// obuunLandfallCounter puts a +1/+1 counter on the target creature, if it is
// still legal.
func obuunLandfallCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	return AddCounter{Target: id, Kind: "+1/+1", N: 1}.Apply(ctx)
}
