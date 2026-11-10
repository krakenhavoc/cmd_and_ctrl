package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emeria Shepherd — Creature — Angel {5}{W}{W}, 4/4:
//
//	"Flying
//	 Landfall — Whenever a land you control enters, you may return
//	 target nonland permanent card from your graveyard to your hand. If
//	 that land is a Plains, you may return that nonland permanent card
//	 to the battlefield instead."
//
// The target is chosen as the ability goes on the stack. As it resolves
// the player picks: leave it, return it to hand, and, when the entering
// land is a Plains (read as it is now, even if it has left, CR 608.2h),
// return it to the battlefield instead.
func init() {
	t := Landfall("Emeria Shepherd — you may return target nonland permanent card from your graveyard to your hand, or to the battlefield if that land is a Plains",
		emeriaShepherdOffer)
	t.Targets = TargetCardInGraveyard("target nonland permanent card from your graveyard", YouOwn(), Permanent(), Nonland())
	Register(Spec{
		OracleID:        "473da1b6-232b-4428-a66d-30262246b47a",
		Name:            "Emeria Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

func emeriaShepherdOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	name := "the card"
	if c, ok := g.LookupCardForEffect(ids[0]); ok {
		name = c.Name
	}
	options := []game.ChoiceOption{
		{Label: "Leave it in your graveyard"},
		{Label: "Return it to your hand"},
	}
	if rfCreatureCEnteredLandHasSubtype(g, item, ctx.TextSubtype("Plains")) {
		options = append(options, game.ChoiceOption{Label: "Return it to the battlefield instead"})
	}
	return PickOption{
		Question: "Emeria Shepherd — return " + name + "?",
		Options:  options,
		ThenKey:  emeriaShepherdAnswered,
		Carry:    ids[:1],
	}.Apply(ctx)
}

// emeriaShepherdAnswered runs the pick: index 1 is hand, 2 the
// battlefield. An on-disk key.
var emeriaShepherdAnswered = OptionPickThen("option-pick/emeria-shepherd-return", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 1 {
		return nil
	}
	dest := game.ZoneHand
	if r.Index == 2 {
		dest = game.ZoneBattlefield
	}
	return ReturnFromGraveyard{Target: r.Carry[0], Dest: dest}.Apply(ctx)
})
