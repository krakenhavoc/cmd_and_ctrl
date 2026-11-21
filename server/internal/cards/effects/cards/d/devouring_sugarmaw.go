package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devouring Sugarmaw // Have for Dinner — Creature — Horror {2}{B}{B}, 6/6
// // Instant — Adventure {1}{W} (Wilds of Eldraine, #4471):
//
//	Devouring Sugarmaw — "Menace, trample
//	 At the beginning of your upkeep, you may sacrifice an artifact,
//	 enchantment, or token. If you don't, tap this creature."
//	Have for Dinner — "Create a 1/1 white Human creature token and a Food
//	 token."
//
// The upkeep is Apocalypse Demon's choice (sacrifice, or tap), made as the
// trigger resolves: "you may" is the option to sacrifice and "if you
// don't" is declining it, or having nothing to sacrifice, which taps the
// Sugarmaw without asking. The sacrificed permanent is any artifact,
// enchantment or token you control, the Sugarmaw included (it is a
// creature, so only if it is somehow one of those).
//
// No simplification.
const devouringSugarmawOracleID = "c8bc28ea-5630-49d5-963a-5789fa83f3c7"

func init() {
	effects.Register(effects.Spec{
		OracleID:        devouringSugarmawOracleID,
		Name:            "Devouring Sugarmaw",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Devouring Sugarmaw — you may sacrifice an artifact, enchantment, or token; if you don't, tap this creature", devouringSugarmawUpkeep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     devouringSugarmawOracleID + "#1",
		Name:         "Have for Dinner",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1},
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}

var sugarmawFodder = effects.Or(effects.Artifact(), effects.Enchantment(), effects.IsTokenPredicate())

func devouringSugarmawUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.PermanentsControlledByMatching(g, item.Controller, sugarmawFodder)) == 0 {
		return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
	return effects.PickOption{
		Question: "Devouring Sugarmaw — sacrifice an artifact, enchantment, or token, or tap Devouring Sugarmaw",
		Options: []game.ChoiceOption{
			{Label: "Tap Devouring Sugarmaw"},
			{Label: "Sacrifice an artifact, enchantment, or token"},
		},
		ThenKey: devouringSugarmawAnswer,
	}.Apply(ctx)
}

// devouringSugarmawAnswer is the upkeep's choice answered, keyed (#4253).
var devouringSugarmawAnswer = effects.OptionPickThen("option-pick/devouring-sugarmaw-sacrifice-or-tap", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index == 1 {
		ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(),
			effects.SacrificeSpec("an artifact, enchantment, or token", sugarmawFodder),
			"Devouring Sugarmaw — sacrifice an artifact, enchantment, or token")
		return nil
	}
	return effects.TapTarget{Target: ctx.Source()}.Apply(ctx)
})
