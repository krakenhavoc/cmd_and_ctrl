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
			Then: func(ctx *Context, index int) error {
				if index == 0 {
					return SacrificeChoice{
						Player:     controller,
						Candidates: lands,
						Question:   name + " — sacrifice a land",
					}.Apply(ctx)
				}
				return SacrificePermanent{Target: source}.Apply(ctx)
			},
		}.Apply(ctx)
	}
}
