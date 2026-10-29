package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galion, Elvenking's Butler — Legendary Creature — Elf Advisor {2}{G}{G}
// (The Hobbit), 4/4:
//
//	"Whenever Galion attacks, choose up to one other target creature you
//	 control. Its base power and toughness become equal to Galion's power
//	 and toughness until end of turn."
//
// #4454, S76. Flayer of Loyalties' base-power-and-toughness record
// (ScopedEffectFor with SetBasePTMods), with the numbers read as the trigger
// resolves from Galion's current power and toughness, counters and pumps
// included (his last-known values if he has left, CR 608.2h). The chosen
// creature's own counters and pumps still apply on top of the new base.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9efb3aa4-0023-4e0e-8bd4-f4dff1b3be66",
		Name:         "Galion, Elvenking's Butler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Galion — another creature's base power and toughness become equal to Galion's until end of turn", galionCopyStats),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
		},
	})
}

func galionCopyStats(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	src, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     game.SetBasePTMods(src.Power, src.Toughness),
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Galion — base power and toughness equal to Galion's",
	}.Apply(ctx)
}
