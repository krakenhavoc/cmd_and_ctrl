package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Legend of Yangchen — Enchantment — Saga, {3}{W}{W}:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I — Starting with you, each player chooses up to one permanent
//	    with mana value 3 or greater from among permanents your
//	    opponents control. Exile those permanents.
//	II — You may have target opponent draw three cards. If you do,
//	     draw three cards.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// The FRONT face of a transforming Saga (ADR 0034, ADR 0079). Chapter
// III is the exile-and-return verb, as on every other transforming
// Saga in this batch; see avatar_yangchen.go for the back face.
//
// SANDBOX SIMPLIFICATION on chapter I, and it is why this face is not
// Full. The printed ability asks EVERY player at the table, in turn
// order starting with the controller, to choose a permanent to exile
// from a SHARED pool — the engine has no prompt that lets a player
// other than a permanent's controller choose it at all (the
// "non-owner choosing among another player's permanents" seam,
// docs/engine-seams.md — Tragic Arrogance, Gluntch, the Bestower), and
// this chapter adds a second axis those two cards don't have: several
// players choosing from the same pool in one chapter. That whole
// shape is out of reach, but the controller's OWN half is not: "a
// permanent with mana value 3 or greater an opponent controls" is an
// ordinary target clause the controller announces, same as Despark's,
// so chapter I ships as "exile target permanent with mana value 3 or
// greater an opponent controls" — one exile instead of up to N, chosen
// only by the controller. TestRegisteredSagasDeclareContiguousChapters
// is also why this can't simply be dropped: a Saga that skips a
// chapter number is a typo the engine actively guards against, so a
// smaller-but-real chapter I is the right shape, not an absent one.
//
// Chapter II is Combustible Gearhulk's yes/no shape with the asker and
// the answerer swapped: THIS card's controller decides (MayChoice
// defaults its Player to the resolving effect's controller, which is
// exactly "you may" here) rather than the chosen opponent, which is
// why Gearhulk passes Player explicitly and this chapter does not.
func init() {
	Register(Spec{
		OracleID:     theLegendOfYangchenOracleID,
		Name:         "The Legend of Yangchen",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Chapter I only exiles one permanent, chosen by you — it doesn't also ask every other player to choose one from the same pool.",
		},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "The Legend of Yangchen — exile target permanent with mana value 3 or greater an opponent controls",
				TargetPermanent("target permanent with mana value 3 or greater an opponent controls",
					And(ManaValueGE(3), OpponentControls())),
				ExileFirstTarget),
			ChapterTriggerTargeting(2, "The Legend of Yangchen — you may have target opponent draw three cards",
				TargetPlayer("target opponent", Opponent()),
				yangchenChapterTwoMayDraw),
			ChapterTrigger(3, "The Legend of Yangchen — exile it, then return it transformed",
				ChapterExileAndReturnTransformed),
		},
	})
}

// theLegendOfYangchenOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const theLegendOfYangchenOracleID = "521a63cf-5e83-4649-9800-c62b2fc474d6"

// yangchenChapterTwoMayDraw is "You may have target opponent draw
// three cards. If you do, draw three cards." — the "you may" belongs
// to the chapter's controller, asked as the chapter resolves; the
// target opponent was already chosen at announce (CR 603.3d).
func yangchenChapterTwoMayDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		opponent := t.ID
		return MayChoice{
			Question: "The Legend of Yangchen — have target opponent draw three cards?",
			OnYes: func(ctx *Context) error {
				if err := (DrawCards{Player: opponent, N: 3}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
			},
		}.Apply(ctx)
	}
	// CR 608.2b: the target left before this resolved.
	return nil
}
