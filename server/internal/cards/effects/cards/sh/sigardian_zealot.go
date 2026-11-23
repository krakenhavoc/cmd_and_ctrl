package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigardian Zealot — Creature — Human Cleric {4}{G}, 3/3 (Midnight Hunt
// Commander, S82 MID batch 13):
//
//	"At the beginning of combat on your turn, choose any number of creatures
//	 with different powers. Each of them gets +X/+X and gains vigilance until
//	 end of turn, where X is this creature's power."
//
// X is this creature's power as the ability resolves, through the layers, or
// as it last existed if it has left the battlefield (CR 608.2h), and is read
// once, before the choice, so pumping the Zealot itself does not change what
// the others get. The choice is Sigarda's Vanguard's: every creature on the
// battlefield, any number, pairwise different powers (midB13DifferentPowers).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30aecffc-bbbf-44d0-8002-0ee5982f27b9",
		Name:         "Sigardian Zealot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat(
				"Sigardian Zealot — choose any number of creatures with different powers; each gets +X/+X and gains vigilance until end of turn, where X is this creature's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					creatures := effects.MidB13CreaturesOnBattlefield(g)
					if len(creatures) == 0 {
						return nil
					}
					effects.MidB13ChooseDifferentPowers(ctx, creatures, 0, len(creatures),
						"Sigardian Zealot — choose any number of creatures with different powers; each gets +X/+X and gains vigilance until end of turn",
						sigardianZealotPick, info.Power)
					return nil
				}),
		},
	})
}

// sigardianZealotPick gives each chosen creature +X/+X and vigilance until
// end of turn, X the Zealot's power read when the ability began to resolve
// (the first of the prompt's numbers). An on-disk key.
var sigardianZealotPick = effects.AskedPickRule("choose-cards/sigardian-zealot-pump", effects.MidB13DifferentPowers,
	func(ctx *effects.Context, r game.CardsChosen) error {
		if len(r.Nums) == 0 {
			return nil
		}
		x := r.Nums[0]
		for _, id := range r.Picked {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.ModifyPTMod(x, x), game.AddKeywordsMod("vigilance")},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Sigardian Zealot — +X/+X and vigilance until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
