package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
// Chapter I is a chain of ChoosePermanents legs, one per player in turn
// order starting with the Saga's controller (CR 101.4 and the card's own
// "starting with you"). Every leg offers the same shared pool, the
// permanents with mana value 3 or greater that the controller's opponents
// control, minus what earlier legs already took, so each answer narrows
// the next; each player picks up to one; the picks are exiled together
// once the last player has answered. It does not target.
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
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Legend of Yangchen — each player chooses up to one permanent with mana value 3 or greater from among permanents your opponents control; exile those permanents",
				yangchenChapterOne),
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
// yangchenRecord files the opponent who draws.
const yangchenRecord = "the-legend-of-yangchen/opponent"

// yangchenDraws is the opponent's draw and then yours: yours waits for a
// prompt theirs asks (a dredge offer; CR 608.2c, #4155).
var yangchenDraws = RegisterThenSteps("then-steps/the-legend-of-yangchen-draws",
	Instruction(func(ctx *Context) error {
		ids := ctx.StepRecord(yangchenRecord).IDs
		if len(ids) == 0 {
			return nil
		}
		return DrawCards{Player: ids[0], N: 3}.Apply(ctx)
	}),
	Instruction(func(ctx *Context) error {
		return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
	}),
)

func yangchenChapterTwoMayDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		opponent := t.ID
		return MayChoice{
			Question: "The Legend of Yangchen — have target opponent draw three cards?",
			Draws:    3, Drawer: opponent,
			Then:  yangchenAnswer,
			Carry: []uuid.UUID{opponent},
		}.Apply(ctx)
	}
	// CR 608.2b: the target left before this resolved.
	return nil
}

// yangchenChapterOne starts the chain at the Saga's controller.
func yangchenChapterOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return yangchenChoose(ctx, seatsStartingWith(g, ctx.Controller()), 0, nil)
}

// yangchenChoose asks players[i] for their pick from what is left of the
// pool, then the next player; after the last it exiles the lot. Only IDs
// are carried between legs.
func yangchenChoose(ctx *Context, players []uuid.UUID, i int, chosen []uuid.UUID) error {
	if i >= len(players) {
		for _, id := range chosen {
			if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	controller := ctx.Controller()
	taken := append([]uuid.UUID(nil), chosen...)
	return ChoosePermanents{
		Player:   players[i],
		Of:       []uuid.UUID{players[i]},
		Question: "The Legend of Yangchen — choose up to one permanent with mana value 3 or greater from among permanents your opponents control",
		Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			var pool []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == controller || slices.Contains(taken, c.InstanceID) {
					continue
				}
				if mv, ok := g.ManaValueForEffect(c); ok && mv >= 3 {
					pool = append(pool, c.InstanceID)
				}
			}
			return pool, 0, 1
		},
		Keyed: yangchenChosen,
		Carry: packIDLists(players, taken),
		Nums:  []int{i},
	}.Apply(ctx)
}

// yangchenChosen is one player's pick, then the next player's, keyed
// (#4290). The carry is the players in order, then the permanents chosen
// so far; the number is which player this was. Registered in init because
// the answer asks the next player, which names the key again.
var yangchenChosen PermanentsThen

func init() {
	yangchenChosen = RegisterPermanentsPick("prompt-run/the-legend-of-yangchen-exile", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			players, taken := unpackIDPair(r.Carry)
			i := 0
			if len(r.Nums) > 0 {
				i = r.Nums[0]
			}
			return yangchenChoose(ctx, players, i+1, append(taken, picked.Cards()...))
		},
	})
}

// seatsStartingWith lists the seats still in the game in turn order,
// beginning with `first`.
func seatsStartingWith(g *game.Game, first uuid.UUID) []uuid.UUID {
	n := len(g.Seats)
	start := 0
	for i, p := range g.Seats {
		if p != nil && p.ID == first {
			start = i
		}
	}
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		if p := g.Seats[(start+i)%n]; p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}

// yangchenAnswer is the "yes": the target opponent (the carried ID) draws
// three. Keyed (#4291).
var yangchenAnswer = MayYes("confirm/the-legend-of-yangchen-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	ctx.RecordStep(yangchenRecord, []uuid.UUID{carryAt(r, 0)})
	return yangchenDraws.Run(ctx)
})
