package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umezawa's Jitte banks two charge counters whenever its equipped
// creature deals combat damage, including damage to a blocker: once per
// combat damage step, so a trampler that hits its blocker and the
// player banks two, not four (WheneverEquippedCreatureDealsCombatDamage,
// #5760). Its
// counter cost and modal choice are made when activated, and only the
// -1/-1 bullet targets. The pump reads the equipped creature when it
// resolves, using the Equipment's last-known attachment if it left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1da10d5c-36a8-473f-a5d4-782ad61d8057",
		Name:         "Umezawa's Jitte",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverEquippedCreatureDealsCombatDamage("Umezawa's Jitte — put two charge counters on it", func(g *game.Game, item *game.StackItem) error {
				if !effects.B09SourceStillOnBattlefield(g, item) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: "charge", N: 2}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Remove a charge counter from Umezawa's Jitte: Choose one — equipped creature gets +2/+2 until end of turn; or target creature gets -1/-1 until end of turn; or you gain 2 life.",
				Cost:    effects.RemoveCountersFromThis("charge", 1),
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Modes: effects.ChooseOne(
					effects.ModeDoing("Equipped creature gets +2/+2 until end of turn.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.PumpEquippedCreatureUntilEOT(ctx, 2, 2, "Umezawa's Jitte — +2/+2")
					}),
					effects.ModeDoing("Target creature gets -1/-1 until end of turn.", effects.TargetCreature("target creature"), func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						target, ok := effects.ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: target.ID, Power: -1, Toughness: -1, Label: "Umezawa's Jitte — -1/-1"}.Apply(ctx)
					}),
					effects.ModeDoing("You gain 2 life.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.GainLife{Amount: 2}.Apply(ctx)
					}),
				),
			},
			effects.EquipAbility("{2}"),
		},
	})
}
