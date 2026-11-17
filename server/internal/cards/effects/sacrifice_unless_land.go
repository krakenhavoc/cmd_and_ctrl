package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sacrifice_unless_land.go — "sacrifice this <permanent> unless you
// sacrifice a land", the upkeep tax of The Gitrog Monster and Territorial
// Dispute, as one effect body so the two cards cannot drift apart.
// Contamination's "unless you sacrifice a creature" is the same body over
// creatures.

// sacrificeThisUnlessYouSacrificeALand is the trigger's whole effect for
// a land. `name` is the permanent's printed name, for the prompts.
func sacrificeThisUnlessYouSacrificeALand(name string) Effect {
	return sacrificeThisUnlessYouSacrificeA(name, "a land", Land(), sacrificeALandOrThisAnswer)
}

// sacrificeThisUnlessYouSacrificeA is the tax for any kind of permanent:
// a two-way choice when the controller has one (`what`, matching `match`)
// to offer, and an unconditional self-sacrifice when there is not. It is
// not a CR 118.12 "unless you pay" (that primitive is for a MANA payment),
// so it is a PickOption between the two sacrifices, and either branch is
// a real PendingChoice, so the upkeep step cannot advance past it
// unanswered. `answer` is the keyed continuation registered for the same
// `what` and `match` (sacrificeOneOrThisAnswer).
func sacrificeThisUnlessYouSacrificeA(name, what string, match CardPredicate, answer game.OptionPickThen) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		controller := item.Controller
		source := item.SourceCardID
		if len(permanentsControlledByMatching(g, controller, match)) == 0 {
			return SacrificePermanent{Target: source}.Apply(ctx)
		}
		return PickOption{
			Question: name + " — sacrifice " + what + ", or sacrifice " + name,
			Options: []game.ChoiceOption{
				{Label: "Sacrifice " + what},
				{Label: "Sacrifice " + name},
			},
			ThenKey: answer,
		}.Apply(ctx)
	}
}

// sacrificeALandOrThisAnswer is the land tax's choice answered, keyed
// (#4253): it reads only the controller and the permanent, both on the
// prompt.
var sacrificeALandOrThisAnswer = sacrificeOneOrThisAnswer("option-pick/sacrifice-this-unless-you-sacrifice-a-land", "a land", Land())

// sacrificeACreatureOrThisAnswer is Contamination's.
var sacrificeACreatureOrThisAnswer = sacrificeOneOrThisAnswer("option-pick/sacrifice-this-unless-you-sacrifice-a-creature", "a creature", Creature())

// sacrificeOneOrThisAnswer registers the continuation of one kind of
// tax under `key`, an on-disk identity: never renamed, never reused.
// Index 0 asks the controller for `what`; any other index sacrifices the
// permanent.
func sacrificeOneOrThisAnswer(key, what string, match CardPredicate) game.OptionPickThen {
	return OptionPickThen(key, func(ctx *Context, r game.OptionPicked) error {
		source := ctx.Source()
		if r.Index == 0 {
			name := "this permanent"
			if c, ok := ctx.Game.LookupCardForEffect(source); ok && c.Name != "" {
				name = c.Name
			}
			// The permanent is asked as a choose_cards prompt that still
			// holds a closure (SacrificeChoice), until that primitive is
			// keyed.
			return SacrificeChoice{
				Player:     ctx.Controller(),
				Candidates: permanentsControlledByMatching(ctx.Game, ctx.Controller(), match),
				Question:   name + " — sacrifice " + what,
			}.Apply(ctx)
		}
		return SacrificePermanent{Target: source}.Apply(ctx)
	})
}
