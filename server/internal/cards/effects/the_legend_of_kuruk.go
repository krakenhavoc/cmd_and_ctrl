package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Legend of Kuruk — Enchantment — Saga, {2}{U}{U}:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I, II — Scry 2, then draw a card.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// The FRONT face of a transforming Saga (ADR 0034, ADR 0079). Chapter
// III is the second verb ADR 0079 built — Fable of the Mirror-Breaker's
// shape exactly: exile then return is TWO zone changes, so the
// permanent that comes back is a new object (CR 400.7) with no lore
// counters, a fresh CR 613.7 timestamp and summoning sickness. See
// avatar_kuruk.go for the back face and fable_of_the_mirror_breaker.go
// for the worked example this is copied from.
//
// Chapters I and II print the identical clause — "Scry 2, then draw a
// card" is Preordain's whole body — so both chapters share one
// function rather than two copies of the same five lines.
//
// No simplification on this face.
func init() {
	Register(Spec{
		OracleID:     theLegendOfKurukOracleID,
		Name:         "The Legend of Kuruk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Legend of Kuruk — scry 2, then draw a card", kurukScryTwoThenDraw),
			ChapterTrigger(2, "The Legend of Kuruk — scry 2, then draw a card", kurukScryTwoThenDraw),
			ChapterTrigger(3, "The Legend of Kuruk — exile it, then return it transformed",
				ChapterExileAndReturnTransformed),
		},
	})
}

// theLegendOfKurukOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const theLegendOfKurukOracleID = "201b8936-18f4-4881-a1d4-c308cf3832bf"

// kurukScryTwoThenDraw is chapters I and II's shared body — Preordain
// written as a chapter Effect instead of an OnResolve. "Then" is load
// bearing for the same reason it is on Preordain: the scry only queues
// a prompt, so the draw belongs in Scry.Then rather than the next
// line, or it would resolve before the player finished deciding what
// to leave on top.
func kurukScryTwoThenDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := item.Controller
	return Scry{
		Player: controller,
		N:      2,
		Then: func(g *game.Game) error {
			return g.DrawNForEffect(controller, 1)
		},
	}.Apply(ctx)
}
