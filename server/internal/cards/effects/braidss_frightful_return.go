package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Braids's Frightful Return — Enchantment — Saga {2}{B} (Dominaria
// United):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I — You may sacrifice a creature. If you do, each opponent discards
//	 a card.
//	 II — Return target creature card from your graveyard to your hand.
//	 III — Target opponent may sacrifice a nonland, nontoken permanent of
//	 their choice. If they don't, they lose 2 life and you draw a card."
//
// Read ahead is the engine's (CR 702.155). Chapter I asks as it resolves
// over what you control then (maySacrificeOne), and each opponent's
// discard is their own choice. Chapter III asks the target opponent, as it
// resolves, for one nonland, nontoken permanent of theirs; naming none, or
// having none to name, is "they don't": they lose 2 life and you draw.
// The sacrifice goes through the engine's one sacrifice path (a permanent
// that can't be sacrificed is not offered, CR 701.21).
//
// No simplification.
func init() {
	const name = "Braids's Frightful Return"
	discard := ChapterTrigger(1, SagaChapterLabel(name, 1, "you may sacrifice a creature; if you do, each opponent discards a card"),
		func(g *game.Game, item *game.StackItem) error {
			return maySacrificeOne(g, item, name+" — sacrifice a creature? If you do, each opponent discards a card.",
				Creature(), false, braidsSacrificedACreature)
		})
	discard.Purpose = game.Purpose{Discards: 1}
	recur := ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "return target creature card from your graveyard to your hand"),
		TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
		returnFirstLegalGraveyardTargetToHand)
	edict := ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "target opponent may sacrifice a nonland, nontoken permanent; if they don't, they lose 2 life and you draw a card"),
		TargetPlayer("target opponent", Opponent()), braidsOpponentMaySacrifice)
	edict.Purpose = game.Purpose{Draws: 1}
	Register(Spec{
		OracleID:        "4567a42a-aa1c-4ca2-b5e2-b09acac3f276",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered:       []game.TriggeredAbility{discard, recur, edict},
	})
}

// braidsSacrificedACreature is chapter I's "if you do".
var braidsSacrificedACreature = RegisterSacrificedOne("braids-frightful-return/each-opponent-discards",
	func(ctx *Context, _ uuid.UUID) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) })

// braidsOpponentMaySacrifice is chapter III: the target opponent is asked,
// with a floor of zero for the "may".
func braidsOpponentMaySacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return ChoosePermanents{
		Player:   t.ID,
		Question: "Braids's Frightful Return — you may sacrifice a nonland, nontoken permanent; if you don't, you lose 2 life and its controller draws a card",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range permanentsControlledByMatching(g, of, And(Nonland(), Not(IsTokenPredicate()))) {
				if !g.CantBeSacrificedForEffect(id) {
					out = append(out, id)
				}
			}
			return out, 0, 1
		},
		Keyed: braidsEdictPick,
		Carry: []uuid.UUID{t.ID},
	}.Apply(ctx)
}

// braidsEdictPick is the opponent's answer: a sacrifice settles first, and
// no pick (or no permanent to offer) is "if they don't".
var braidsEdictPick = RegisterPermanentsPick("braids-frightful-return/edict", PermanentsDef{
	Sacrifice: true,
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		if picked.Count() > 0 || len(r.Carry) == 0 {
			return nil
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Carry[0], -2); err != nil {
			return err
		}
		return DrawCards{N: 1}.Apply(ctx)
	},
})
