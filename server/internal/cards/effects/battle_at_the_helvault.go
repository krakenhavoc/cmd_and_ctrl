package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle at the Helvault — Enchantment — Saga {4}{W}{W} (Commander Masters,
// #5948):
//
//	"I, II — For each player, exile up to one target non-Saga, nonland
//	 permanent that player controls until this Saga leaves the battlefield.
//	 III — Create Avacyn, a legendary 8/8 white Angel creature token with
//	 flying, vigilance, and indestructible."
//
// Chapters I and II carry one "up to one" clause per seat, the first for
// you and the rest for the opponents in turn order, each bound to its own
// player so the CR 608.2b re-check asks "does THAT player still control
// it?". Every pick is exiled until the Saga leaves (CR 610.3), which it
// does after chapter III (CR 714.4), so everything returns then.
//
// No simplification.
func init() {
	const name = "Battle at the Helvault"
	const exileText = "for each player, exile up to one target non-Saga, nonland permanent that player controls until this Saga leaves the battlefield"
	targets := func() *game.TargetSpec {
		return gap110PerPlayerClauses("non-Saga, nonland permanent", Nonland(), Not(HasSubtype("Saga")))
	}
	Register(Spec{
		OracleID:     "076584c8-ec58-4711-8f59-4839a4fa88ad",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, exileText), targets(), gap110HelvaultExile),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, exileText), targets(), gap110HelvaultExile),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "create Avacyn, a legendary 8/8 white Angel creature token with flying, vigilance, and indestructible"),
				Do(CreateToken{Template: TokenCard("8/8 white Legendary Angel named Avacyn with flying, vigilance, and indestructible"), N: 1})),
		},
	})
}

// gap110HelvaultExile exiles every still-legal pick until this Saga leaves
// the battlefield, in announce order.
func gap110HelvaultExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileUntil{
			Target:     t.ID,
			ThisLeaves: true,
			Label:      "Battle at the Helvault — returns when the Saga leaves the battlefield",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
