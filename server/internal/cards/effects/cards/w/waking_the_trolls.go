package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Waking the Trolls — Enchantment — Saga {4}{R}{G} (Kaldheim, #5505):
//
//	"I — Destroy target land.
//	 II — Put target land card from a graveyard onto the battlefield under
//	 your control.
//	 III — Choose target opponent. If they control fewer lands than you,
//	 create a number of 4/4 green Troll Warrior creature tokens with trample
//	 equal to the difference."
//
// Chapter II names a land card in any graveyard and puts it onto the
// battlefield under the Saga's controller's control, whoever owned it.
// Chapter III targets an opponent and, as it resolves, compares the lands each
// of you controls then; a target that has become illegal does nothing, and
// neither does a comparison that is not in your favour.
//
// No simplification.
func init() {
	const name = "Waking the Trolls"
	effects.Register(effects.Spec{
		OracleID:     "c8cdabc4-e147-491b-9c44-96e57f6ec3b4",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "destroy target land"),
				effects.TargetPermanent("target land", effects.Land()),
				effects.DestroyFirstLegalCardTarget),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "put target land card from a graveyard onto the battlefield under your control"),
				effects.TargetCardInGraveyard("target land card from a graveyard", effects.Land()),
				effects.ReanimateChosenUnderYourControl("Waking the Trolls")),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "choose target opponent; if they control fewer lands than you, create a number of 4/4 green Troll Warrior creature tokens with trample equal to the difference"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					opp, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					diff := len(effects.LandsControlledByPlayer(g, item.Controller)) - len(effects.LandsControlledByPlayer(g, opp))
					if diff <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("4/4 green Troll Warrior with trample"), N: diff}.Apply(ctx)
				}),
		},
	})
}
