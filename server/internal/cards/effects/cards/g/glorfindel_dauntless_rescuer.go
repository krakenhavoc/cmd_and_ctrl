package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorfindel, Dauntless Rescuer — Legendary Creature — Elf Noble {2}{G}, 3/2
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever you scry, choose one and Glorfindel gets +1/+1 until end of
//	 turn.
//	 • Glorfindel must be blocked this turn if able.
//	 • Glorfindel can't be blocked by more than one creature each combat
//	   this turn."
//
// An EventScry trigger the controller caused, modal with both bullets
// untargeted. The mode's requirement is a data record pinned to Glorfindel
// and swept at cleanup (BlockRequirementUntilEOT for "must be blocked",
// CantBeBlockedByMoreThanMod for the one-blocker cap, CR 514.2); the +1/+1
// is the same ability's other half and happens whichever mode is chosen.
// Scrying twice gives two triggers and two choices.
//
// No simplification.
func init() {
	const label = "Glorfindel, Dauntless Rescuer — choose one; Glorfindel gets +1/+1 until end of turn"
	// The +1/+1 is the ability's other half and happens whichever bullet is
	// chosen (exactly one always is), so each bullet's body does it first.
	pump := func(item *game.StackItem, ctx *effects.Context) bool {
		if !effects.OnBattlefield(ctx.Game, item.SourceCardID) {
			return false
		}
		_ = effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: "Glorfindel, Dauntless Rescuer — +1/+1 until end of turn"}.Apply(ctx)
		return true
	}
	trig := effects.On(game.EventScry, effects.ByYou, label, func(_ *game.Game, _ *game.StackItem) error { return nil })
	trig.Modes = effects.ChooseOne(
		effects.ModeDoing("Glorfindel must be blocked this turn if able.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			if !pump(item, ctx) {
				return nil
			}
			return effects.BlockRequirementUntilEOT{Target: item.SourceCardID, Kind: game.BlockRequirementMustBeBlocked,
				Label: "Glorfindel, Dauntless Rescuer — must be blocked this turn if able"}.Apply(ctx)
		}),
		effects.ModeDoing("Glorfindel can't be blocked by more than one creature each combat this turn.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			if !pump(item, ctx) {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, item.SourceCardID, nil, "Glorfindel, Dauntless Rescuer — can't be blocked by more than one creature this turn",
				game.CantBeBlockedByMoreThanMod(1))
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "93842030-2017-4233-a7ac-7112361c019f",
		Name:         "Glorfindel, Dauntless Rescuer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
	})
}
