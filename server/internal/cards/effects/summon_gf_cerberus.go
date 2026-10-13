package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: G.F. Cerberus — Enchantment Creature — Saga Dog {2}{R}{R}, 3/3:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Surveil 1.
//	 II — Double — When you next cast an instant or sorcery spell this
//	 turn, copy it. You may choose new targets for the copy.
//	 III — Triple — When you next cast an instant or sorcery spell this
//	 turn, copy it twice. You may choose new targets for the copies."
//
// A Saga creature (Summon: Bahamut's shape). Chapters II and III are
// Doublecast's delayed trigger: they wait for the next matching cast
// this turn and copy it from the triggered ability, above the spell,
// so the table gets a response window and the copies resolve first.
// Chapter III asks for two copies.
//
// No simplification.
func init() {
	instantOrSorcery := game.CastFilter{Types: []string{"Instant", "Sorcery"}}
	Register(Spec{
		OracleID:     "b6b22bac-853a-45a8-a74d-9904ec2b34fd",
		Name:         "Summon: G.F. Cerberus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel("Summon: G.F. Cerberus", 1, "surveil 1"), Do(Surveil{N: 1})),
			ChapterTrigger(2, SagaChapterLabel("Summon: G.F. Cerberus", 2, "Double — copy the next instant or sorcery you cast this turn"), func(g *game.Game, item *game.StackItem) error {
				return WhenYouNextCast("Summon: G.F. Cerberus — Double: copy that spell",
					instantOrSorcery, copyTheSpellBody).Apply(NewContext(g, item))
			}),
			ChapterTrigger(3, SagaChapterLabel("Summon: G.F. Cerberus", 3, "Triple — copy the next instant or sorcery you cast this turn twice"), func(g *game.Game, item *game.StackItem) error {
				return WhenYouNextCast("Summon: G.F. Cerberus — Triple: copy that spell twice",
					instantOrSorcery, copyTheSpellTwiceBody).Apply(NewContext(g, item))
			}),
		},
	})
}

// copyTheSpellTwiceBody is copyTheSpellYouJustCast with two copies
// (Summon: G.F. Cerberus's chapter III).
var copyTheSpellTwiceBody = game.SimpleDelayedBody("copy/the-spell-you-just-cast-twice", copyTheSpellYouJustCastTwice)

func copyTheSpellYouJustCastTwice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	return CopySpell{
		StackID:          cast[0],
		Controller:       item.Controller,
		Count:            2,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(ctx)
}
