package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        devouringSugarmawOracleID,
		Name:            "Devouring Sugarmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "trample"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Devouring Sugarmaw — you may sacrifice an artifact, enchantment, or token; if you don't, tap this creature", devouringSugarmawUpkeep),
		},
	})
	Register(Spec{
		OracleID:     devouringSugarmawOracleID + "#1",
		Name:         "Have for Dinner",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("1/1 white Human"), N: 1},
			CreateToken{Template: FoodToken(), N: 1},
		),
	})
}

var sugarmawFodder = Or(Artifact(), Enchantment(), IsTokenPredicate())

func devouringSugarmawUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(permanentsControlledByMatching(g, item.Controller, sugarmawFodder)) == 0 {
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
	return PickOption{
		Question: "Devouring Sugarmaw — sacrifice an artifact, enchantment, or token, or tap Devouring Sugarmaw",
		Options: []game.ChoiceOption{
			{Label: "Tap Devouring Sugarmaw"},
			{Label: "Sacrifice an artifact, enchantment, or token"},
		},
		ThenKey: devouringSugarmawAnswer,
	}.Apply(ctx)
}

// devouringSugarmawAnswer is the upkeep's choice answered, keyed (#4253).
var devouringSugarmawAnswer = OptionPickThen("option-pick/devouring-sugarmaw-sacrifice-or-tap", func(ctx *Context, r game.OptionPicked) error {
	if r.Index == 1 {
		ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(),
			sacrificeSpec("an artifact, enchantment, or token", sugarmawFodder),
			"Devouring Sugarmaw — sacrifice an artifact, enchantment, or token")
		return nil
	}
	return TapTarget{Target: ctx.Source()}.Apply(ctx)
})
