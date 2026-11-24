package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Euru, Acorn Scrounger — Legendary Creature — Squirrel Soldier
// {2}{B}{G}, 3/3 (Mystery Booster Commander, #5948):
//
//	"When Euru enters, you may forage. When you do, conjure a card named
//	 Chitterspitter onto the battlefield. (To forage, exile three cards
//	 from your graveyard or sacrifice a Food.)
//	 Whenever one or more Squirrels you control deal combat damage to a
//	 player, you may sacrifice a token. If you do, put an acorn counter on
//	 each permanent you control named Chitterspitter."
//
// The forage is Curious Forager's (MayForage, then a "when you do"
// reflexive trigger). The conjured Chitterspitter is a real card, built
// from its printed record and entered through the ordinary battlefield
// entry (ADR 0099 amendment 2026-11-24). The acorn is Chitterspitter's
// own "may sacrifice a token. If you do" prompt, counted on every
// Chitterspitter you control as the sacrifice settles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b01657ad-120f-4d00-a367-fd112d403700",
		Name:         "Euru, Acorn Scrounger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Euru, Acorn Scrounger — you may forage; when you do, conjure a Chitterspitter onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayForage(effects.NewContext(g, item), "Euru, Acorn Scrounger — forage to conjure a card named Chitterspitter onto the battlefield?", euruForaged)
				}),
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.HasSubtype("Squirrel"),
				"Euru, Acorn Scrounger — you may sacrifice a token; if you do, an acorn counter on each Chitterspitter you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Euru, Acorn Scrounger — sacrifice a token to put an acorn counter on each permanent you control named Chitterspitter?",
						effects.IsTokenPredicate(), false, euruAcorns)
				}),
		},
	})
}

// euruConjure is the conjure, checked at boot.
var euruConjure = effects.CheckConjure("Euru, Acorn Scrounger", effects.Conjure{Names: []string{"Chitterspitter"}, Zone: game.ZoneBattlefield})

// euruForaged is the paid half, keyed: the "when you do" reflexive
// trigger. An on-disk key.
var euruForaged = effects.PayUnlessBranches("pay-unless/euru-acorn-scrounger-conjure",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.WhenYouDo("Euru, Acorn Scrounger — conjure a card named Chitterspitter onto the battlefield", euruConjureBody).Apply(ctx)
	}, nil)

// euruConjureBody is the reflexive trigger's resolution. An on-disk key.
var euruConjureBody = game.ReflexiveBody("euru-acorn-scrounger/conjure-chitterspitter",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		return euruConjure.Apply(effects.NewContext(g, item))
	}), nil)

// euruAcorns is the acorn counters once a token has really gone. An
// on-disk key.
var euruAcorns = effects.RegisterSacrificedOne("prompt-run/euru-acorn-scrounger-acorns", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.CounterOnEachPermanentYouControlNamed{Name: "Chitterspitter", Kind: "acorn"}.Apply(ctx)
})
