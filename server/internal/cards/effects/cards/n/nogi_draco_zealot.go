package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nogi, Draco-Zealot — Legendary Creature — Kobold Shaman {1}{R}{R}, 3/3
// (Tarkir: Dragonstorm Commander):
//
//	"Dragon spells you cast cost {1} less to cast.
//	 Whenever Nogi attacks, if you control three or more Dragons, until
//	 end of turn, Nogi becomes a Dragon with base power and toughness 5/5
//	 and gains flying."
//
// The discount is Dragonlord's Servant's. The attack trigger's
// intervening "if" (CR 603.4) is checked as Nogi attacks and again as the
// ability resolves (InterveningIf), counting the Dragons you control then
// (Nogi is not one until the effect applies). The effect is one
// until-end-of-turn record on Nogi at one timestamp: layer 4 replaces its
// creature types with Dragon ("becomes a Dragon", so it is no longer a
// Kobold Shaman, and stays legendary), layer 6 gives flying, and layer 7b
// sets its base power and toughness to 5/5, under counters and anthems.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d04aa8c-7bff-42c6-bed3-0b18d2cb1529",
		Name:         "Nogi, Draco-Zealot",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Dragon spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.SpellOfCreatureType("Dragon")),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control three or more Dragons",
				Test: func(ctx *effects.Context) bool {
					return effects.B43CreaturesOfSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtypes("Dragon")[0]) >= 3
				},
			}, effects.WheneverThisAttacks("Nogi, Draco-Zealot — until end of turn, Nogi becomes a Dragon with base power and toughness 5/5 and gains flying",
				nogiBecomesADragon)),
		},
	})
}

// nogiBecomesADragon is the trigger's body: Nogi, if still on the
// battlefield, is a 5/5 flying Dragon until end of turn.
func nogiBecomesADragon(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: item.SourceCardID,
		Mods: []game.Mod{
			game.SetCreatureTypesMod(ctx.TextSubtypes("Dragon")...),
			game.AddKeywordsMod("flying"),
			game.SetBasePowerMod(5),
			game.SetBaseToughnessMod(5),
		},
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Nogi, Draco-Zealot — a Dragon with base power and toughness 5/5 and flying",
	}.Apply(ctx)
}
