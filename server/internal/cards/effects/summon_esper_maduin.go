package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Esper Maduin — the back face of Esper Origins, Enchantment
// Creature — Saga Elemental, 4/4 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Reveal the top card of your library. If it's a permanent card, put
//	 it into your hand.
//	 II — Add {G}{G}.
//	 III — Other creatures you control get +2/+2 and gain trample until end
//	 of turn."
//
// A Saga creature (Summon: Esper Ramuh's chapter shape). Chapter I reveals
// the top card and takes it only if it is a permanent card; any other card
// stays on top, revealed. Chapter II adds the mana as it resolves; it uses
// the stack and is not a mana ability. Chapter III locks in the creatures
// you control other than this one as it resolves (CR 611.2c), a creature
// that enters later gets nothing. Registered under "<oracle_id>#1".
//
// No simplification.
func init() {
	const name = "Summon: Esper Maduin"
	others := func(ctx *Context) CardPredicate {
		self := ctx.Source()
		return And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })
	}
	Register(Spec{
		OracleID:     esperOriginsOracleID + "#1",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "reveal the top card of your library; if it's a permanent card, put it into your hand"),
				func(g *game.Game, item *game.StackItem) error {
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1,
						name+" — reveals the top card of their library")
					if len(revealed) == 0 {
						return nil
					}
					if c, ok := g.LookupCardForEffect(revealed[0]); ok && c.IsPermanent() {
						g.TakeFromLibraryToHandForEffect(item.Controller, revealed)
					}
					return nil
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "add {G}{G}"), Do(AddMana{Produced: "{G}{G}"})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "other creatures you control get +2/+2 and gain trample until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					match := others(ctx)
					if err := (BoostUntilEOT{Match: match, Power: 2, Toughness: 2, Label: name + " — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: name + " — trample until end of turn"}.Apply(ctx)
				}),
		},
	})
}
