package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	eternalAcrobatToastFlicker = "eternal-acrobat-toast/damage-flicker"
	eternalAcrobatToastUntap   = "eternal-acrobat-toast/tap-untap"
)

// Eternal Acrobat Toast — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Eternal", "Acrobat", "Toast"; three art stickers.
//	{TK}{TK} — Whenever this creature deals combat damage to a player, exile target creature you control, then return it to the battlefield under its owner's control.
//	{TK}{TK}{TK} — {T}: Untap another target permanent.
//	{TK}{TK}{TK} — 4/4
//	{TK}{TK}{TK}{TK}{TK} — 7/8
//
// The first sticker gives a combat-damage trigger that flickers a creature you control: it is exiled and returned to the battlefield under its owner's control, a new object (CR 400.7), so a sticker on it is kept (CR 123.5) but the object is new. The second gives a tap ability that untaps another target permanent.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60a69ddc-3289-4786-944d-27a82c5f0dc8",
		Name:         "Eternal Acrobat Toast",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: eternalAcrobatToastFlicker,
				Triggered: []game.TriggeredAbility{
					effects.Targeting(
						effects.WheneverThisDealsCombatDamageToAPlayer("Eternal Acrobat Toast sticker — exile target creature you control, then return it to the battlefield under its owner's control",
							func(g *game.Game, item *game.StackItem) error {
								ctx := effects.NewContext(g, item)
								id, ok := effects.StickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								return effects.Flicker{Target: id}.Apply(ctx)
							}),
						effects.TargetCreature("target creature you control", effects.YouControl())),
				},
				Text: "Whenever this creature deals combat damage to a player, exile target creature you control, then return it to the battlefield under its owner's control.",
			},
			{
				Key: eternalAcrobatToastUntap,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: Untap another target permanent.",
					Cost:    effects.TapCost(),
					Targets: effects.Another(effects.TargetPermanent("another target permanent")),
					Effect:  effects.UntapEachLegalTarget,
				}},
				Text: "{T}: Untap another target permanent.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: eternalAcrobatToastFlicker},
			{Grant: eternalAcrobatToastUntap},
		}},
	})
}
