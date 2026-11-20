package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doctor Strange, Surgeon — Legendary Creature — Human Doctor Hero {4}{W},
// 2/5 (Marvel Super Heroes Commander, #4455):
//
//	"Lifelink
//	 If you would gain life, you gain twice that much life instead.
//	 At the beginning of each combat, if you have at least 10 life more
//	 than your starting life total, creatures you control get +2/+2 and
//	 gain vigilance until end of turn."
//
// The doubling is Alhammarret's Archive's life replacement. The combat
// trigger has an intervening "if" (CR 603.4): the life test is checked as
// the combat begins (the row's AppliesTo) and again as the trigger
// resolves (If), against the table's own starting life total, which is 40
// in Commander. The set of creatures is locked when it resolves.
//
// No simplification.
func init() {
	trigger := effects.AtBeginningOfEachCombat("Doctor Strange, Surgeon — creatures you control get +2/+2 and gain vigilance until end of turn",
		effects.Do(effects.If{
			Cond: msc7TenAboveStarting,
			Then: []effects.Applier{
				effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 2, Label: "Doctor Strange, Surgeon — +2/+2"},
				effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"vigilance"}, Label: "Doctor Strange, Surgeon — vigilance"},
			},
		}))
	trigger.AppliesTo = effects.AllOf(trigger.AppliesTo, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.ControllerLifeAtLeast(g, source, effects.StartingLifeOf(g)+10)
	})
	effects.Register(effects.Spec{
		OracleID:        "43c31a03-2bbd-49a6-8b78-c5f3cab04a07",
		Name:            "Doctor Strange, Surgeon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Replacements: []game.ReplacementEffect{
			effects.YouGainTwiceThatMuchLife("Doctor Strange, Surgeon: gain twice that much life"),
		},
		Triggered: []game.TriggeredAbility{trigger},
	})
}

// msc7TenAboveStarting is "you have at least 10 life more than your
// starting life total", read as the trigger resolves.
var msc7TenAboveStarting = effects.CondBoard{
	Name: "at least 10 life more than your starting life total",
	Test: func(ctx *effects.Context) bool {
		p := ctx.PlayerByID(ctx.Controller())
		return p != nil && p.Life >= effects.StartingLifeOf(ctx.Game)+10
	},
}
