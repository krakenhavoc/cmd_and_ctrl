package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chani, Blade of Sietch Tabr — Legendary Creature — Human Warrior {3}{R}{G},
// 3/3:
//
//	"Double strike
//	 You may play an additional land on each of your turns.
//	 Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature. If that land is a Desert, that creature gains vigilance,
//	 hexproof, and haste until end of turn."
//
// Aesi's extra land drop (Spec.AdditionalLandPlays) and a landfall trigger
// with a target chosen as it goes on the stack (CR 603.3d). The "that land"
// is the permanent that entered, read at resolution as it last existed
// (CR 608.2h, Context.TriggeringPermanent), so a Desert that has already left
// still grants the keywords; its subtype is read through the layers. A target
// gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "029f91b8-f5da-4ebb-835d-b6e42cffceaa",
		Name:                "Chani, Blade of Sietch Tabr",
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"double strike"},
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Landfall("Chani, Blade of Sietch Tabr — put a +1/+1 counter on target creature; if that land is a Desert, it gains vigilance, hexproof, and haste until end of turn (landfall)",
					chaniLandfall),
				effects.TargetCreature("target creature")),
		},
	})
}

func chaniLandfall(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	land, known := ctx.TriggeringPermanent()
	if !known || !effects.HasFold(land.Characteristic.Subtypes, "Desert") {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   id,
		Keywords: []string{"vigilance", "hexproof", "haste"},
		Label:    "Chani, Blade of Sietch Tabr — vigilance, hexproof, and haste",
	}.Apply(ctx)
}
