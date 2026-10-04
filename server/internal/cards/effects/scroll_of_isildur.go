package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scroll of Isildur — Enchantment — Saga {2}{U}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I — Gain control of up to one target artifact for as long as you
//	     control this Saga. The Ring tempts you.
//	 II — Tap up to two target creatures. Put a stun counter on each
//	     of them.
//	 III — Draw a card for each tapped creature target opponent
//	     controls."
//
// Chapter I's control lasts while you control the Saga (CR 611.2b), so
// it ends when the Saga is sacrificed after chapter III. If you no
// longer control the Saga as chapter I resolves, control doesn't change
// at all (2023-06-16 ruling); the Ring still tempts. Chapter III counts
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa9a45a5-0250-465b-962e-f2d45dea26f9",
		Name:         "Scroll of Isildur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "Scroll of Isildur — gain control of an artifact, then the Ring tempts you",
				TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1), scrollOfIsildurTake),
			ChapterTriggerTargeting(2, "Scroll of Isildur — tap up to two creatures and stun them",
				TargetCreature("up to two target creatures").WithCount(0, 2), scrollOfIsildurStun),
			ChapterTriggerTargeting(3, "Scroll of Isildur — draw a card for each tapped creature target opponent controls",
				TargetPlayer("target opponent", Opponent()), scrollOfIsildurDraw),
		},
	})
}

// scrollOfIsildurTake is chapter I.
func scrollOfIsildurTake(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetIDs(ctx) {
		d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
		if !ok {
			break
		}
		if err := (GainControl{Target: id, Duration: d, Label: "Scroll of Isildur — gain control"}).Apply(ctx); err != nil {
			return err
		}
	}
	return TheRingTemptsYou{}.Apply(ctx)
}

// scrollOfIsildurStun is chapter II.
func scrollOfIsildurStun(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetIDs(ctx) {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// scrollOfIsildurDraw is chapter III.
func scrollOfIsildurDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := legalTargetIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ids[0] && c.IsCreature() && c.Tapped {
			n++
		}
	}
	return DrawCards{N: n}.Apply(ctx)
}
