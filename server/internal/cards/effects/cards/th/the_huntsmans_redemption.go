package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Huntsman's Redemption — Enchantment — Saga {2}{G} (Wilds of
// Eldraine, #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 3/3 green Beast creature token.
//	 II — You may sacrifice a creature. If you do, search your library for
//	 a creature or basic land card, reveal it, put it into your hand, then
//	 shuffle.
//	 III — Up to two target creatures each get +2/+2 and gain trample
//	 until end of turn."
//
// II asks as it resolves (Bolg's maySacrificeOne), over any creature you
// control, itself included, with a floor of none; the search is the "if
// you do", run only for a creature that really left (RegisterSacrificedOne).
// It is a tutor for either card type, revealed, and the library shuffles
// whether or not anything was found. III picks up to two targets as the
// chapter goes on the stack and gives each still legal one +2/+2 and
// trample.
//
// No simplification.
func init() {
	const name = "The Huntsman's Redemption"
	effects.Register(effects.Spec{
		OracleID:     "ae63fd5c-46ff-41fb-a1f6-6259336525c1",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 3/3 green Beast creature token"),
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "you may sacrifice a creature; if you do, search your library for a creature or basic land card"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, name+" — sacrifice a creature to search your library for a creature or basic land card?",
						effects.Creature(), false, huntsmanSacrificed)
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "up to two target creatures each get +2/+2 and gain trample until end of turn"),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: name + " — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: name + " — trample"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}

// huntsmanSacrificed is the "if you do": the tutor, keyed (#4290).
var huntsmanSacrificed = effects.RegisterSacrificedOne("prompt-run/the-huntsmans-redemption-sacrificed", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsCreature() || effects.IsBasicLand(c) },
		Dest:      game.ZoneHand,
		Limit:     1,
		Optional:  true,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "The Huntsman's Redemption — a creature or basic land card",
	}.Apply(ctx)
})
