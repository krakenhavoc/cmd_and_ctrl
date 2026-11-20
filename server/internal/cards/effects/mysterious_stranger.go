package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mysterious Stranger — Creature — Human Rogue {2}{R}{R}, 3/2 (Fallout,
// #5948):
//
//	"Flash
//	 When this creature enters, for each graveyard with an instant or sorcery
//	 card in it, exile target instant or sorcery card from that graveyard. If
//	 two or more cards are exiled this way, choose one of them at random and
//	 copy it. You may cast the copy without paying its mana cost."
//
// One target per graveyard that holds an instant or sorcery card as the
// trigger goes on the stack (CR 603.3d), and exactly that many: the clause is
// built per trigger from the table (TargetsFrom) with a count of the
// graveyards that qualify and a rule that no two picks come from the same
// one (gapB63EachDifferentGraveyard). With no instant or sorcery in any
// graveyard there is no target and nothing happens. The cards that really
// reached exile are the set: with two or more, one is chosen at random
// through the keyed random stream and copied (CopyAndCast, CR 707.12), the
// copy castable for free while the ability is still resolving (CR 608.2g);
// with one, nothing is copied.
//
// No simplification.
func init() {
	clause := func(n int) *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card from each graveyard that has one", Or(Instant(), Sorcery())).
			WithCount(n, n).EachDifferent(gapB63EachDifferentGraveyard())
	}
	enters := WhenThisEnters("Mysterious Stranger — exile target instant or sorcery card from each graveyard that has one; if two or more, copy one at random",
		mysteriousStrangerResolve)
	enters.TargetsFrom = func(_ game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
		owners := map[uuid.UUID]bool{}
		for _, p := range g.Seats {
			if p == nil || p.Graveyard == nil {
				continue
			}
			for _, c := range p.Graveyard.Cards {
				if c.IsInstant() || c.IsSorcery() {
					owners[p.ID] = true
					break
				}
			}
		}
		if len(owners) == 0 {
			return nil
		}
		return clause(len(owners))
	}
	Register(Spec{
		OracleID:        "4cd2e1c2-5f14-4ac9-9c88-899af6ad9f80",
		Name:            "Mysterious Stranger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

func mysteriousStrangerResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(mysteriousStrangerExiled))
}

// mysteriousStrangerExiled is the rest of the ability once the exile has
// settled: with two or more cards exiled, one at random is copied and the
// copy may be cast for free. An on-disk key: never renamed, never reused.
var mysteriousStrangerExiled = CardTail("mysterious-stranger/copy-one-at-random", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) < 2 {
		return nil
	}
	pick := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), o.IDs, 1)
	if len(pick) == 0 {
		return nil
	}
	return CopyAndCast{Cards: pick, Free: true}.Apply(ctx)
})
