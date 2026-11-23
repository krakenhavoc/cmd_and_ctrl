package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arcum Dagsson — Legendary Creature — Human Artificer {3}{U}, 2/2:
//
//	"{T}: Target artifact creature's controller sacrifices it. That
//	 player may search their library for a noncreature artifact card,
//	 put it onto the battlefield, then shuffle."
//
// The ability targets the artifact creature, not its controller (the
// 2006 ruling). "That player" is that creature's controller as the
// ability resolves, read before the sacrifice (CR 701.21a: only a
// permanent's controller can sacrifice it).
//
// The search is not "if they do". The ruling says a legal target that
// was not sacrificed (a Sigarda, Host of Herons on the board) still
// lets its controller search, so the search runs from the sacrifice's
// continuation whether or not the creature left. It runs from the
// continuation, rather than on the next line, so a sacrificed
// commander's CR 903.9 question is answered first. An illegal target
// on resolution (CR 608.2b) means no sacrifice and no search.
//
// The search is "may" (Optional), may find nothing, and puts the card
// onto the battlefield under the searcher's control, which is the
// library's owner.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fdfe986-0216-468b-aac8-bbcf588fd894",
		Name:         "Arcum Dagsson",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target artifact creature's controller sacrifices it. That player may search their library for a noncreature artifact card, put it onto the battlefield, then shuffle.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
			Effect:  arcumDagssonSacrificeThenSearch,
		}},
	})
}

// arcumDagssonSacrificeThenSearch is the ability's resolution.
func arcumDagssonSacrificeThenSearch(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	player, ok := effects.ControllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	return effects.SacrificePermanent{Target: target, ThenTail: effects.WithCarry(ctx.ItemTail(arcumDagssonSearch), player)}.Apply(ctx)
}

// arcumDagssonSearch is "that creature's controller may search their
// library for a noncreature artifact card", after the sacrifice, keyed
// (#4253 slice 10a): the controller is Carry[0].
var arcumDagssonSearch = effects.CardTail("then/arcum-dagsson-search", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.SearchLibrary{
		Player:    o.Tail.Carry[0],
		Predicate: func(c game.Card) bool { return c.IsArtifact() && !c.IsCreature() },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Optional:  true,
		Reason:    "Arcum Dagsson — you may search for a noncreature artifact card",
	}.Apply(ctx)
})
