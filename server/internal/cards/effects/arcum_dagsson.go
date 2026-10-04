package effects

import (
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
	Register(Spec{
		OracleID:     "0fdfe986-0216-468b-aac8-bbcf588fd894",
		Name:         "Arcum Dagsson",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target artifact creature's controller sacrifices it. That player may search their library for a noncreature artifact card, put it onto the battlefield, then shuffle.",
			Cost:    TapCost(),
			Targets: TargetCreature("target artifact creature", Artifact()),
			Effect:  arcumDagssonSacrificeThenSearch,
		}},
	})
}

// arcumDagssonSacrificeThenSearch is the ability's resolution.
func arcumDagssonSacrificeThenSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	player, ok := controllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	return SacrificePermanent{Target: target, Then: func(ctx *Context, _ bool) error {
		return SearchLibrary{
			Player:    player,
			Predicate: func(c game.Card) bool { return c.IsArtifact() && !c.IsCreature() },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Optional:  true,
			Reason:    "Arcum Dagsson — you may search for a noncreature artifact card",
		}.Apply(ctx)
	}}.Apply(ctx)
}
