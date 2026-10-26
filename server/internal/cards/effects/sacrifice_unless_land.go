package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrifice_unless_land.go — "sacrifice this <permanent> unless you
// sacrifice a land", the upkeep tax of The Gitrog Monster and Territorial
// Dispute, as one effect body so the two cards cannot drift apart.

// sacrificeThisUnlessYouSacrificeALand is the trigger's whole effect: a
// two-way choice when the controller has a land to offer, and an
// unconditional self-sacrifice when there is not. It is not a CR 118.12
// "unless you pay" (that primitive is for a MANA payment), so it is a
// PickOption between the two sacrifices, and either branch is a real
// PendingChoice, so the upkeep step cannot advance past it unanswered.
// `name` is the permanent's printed name, for the prompts.
func sacrificeThisUnlessYouSacrificeALand(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		controller := item.Controller
		source := item.SourceCardID
		lands := permanentsControlledByMatching(g, controller, Land())
		if len(lands) == 0 {
			return SacrificePermanent{Target: source}.Apply(ctx)
		}
		return PickOption{
			Question: name + " — sacrifice a land, or sacrifice " + name,
			Options: []game.ChoiceOption{
				{Label: "Sacrifice a land"},
				{Label: "Sacrifice " + name},
			},
			ThenKey: sacrificeALandOrThisAnswer,
		}.Apply(ctx)
	}
}

// sacrificeALandOrThisAnswer is the upkeep tax's choice answered, keyed
// (#4253): it reads only the controller and the permanent, both on the
// prompt.
var sacrificeALandOrThisAnswer = OptionPickThen("option-pick/sacrifice-this-unless-you-sacrifice-a-land", func(ctx *Context, r game.OptionPicked) error {
	source := ctx.Source()
	if r.Index == 0 {
		name := "this permanent"
		if c, ok := ctx.Game.LookupCardForEffect(source); ok && c.Name != "" {
			name = c.Name
		}
		// The land is asked as a choose_cards prompt that still holds a
		// closure (SacrificeChoice), until that primitive is keyed.
		return SacrificeChoice{
			Player:     ctx.Controller(),
			Candidates: permanentsControlledByMatching(ctx.Game, ctx.Controller(), Land()),
			Question:   name + " — sacrifice a land",
		}.Apply(ctx)
	}
	return SacrificePermanent{Target: source}.Apply(ctx)
})
