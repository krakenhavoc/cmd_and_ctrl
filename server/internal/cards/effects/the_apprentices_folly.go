package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Apprentice's Folly — Enchantment — Saga {2}{U}{R} (Wilds of
// Eldraine, #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Choose target nontoken creature you control that doesn't have
//	 the same name as a token you control. Create a token that's a copy of
//	 it, except it isn't legendary, is a Reflection in addition to its
//	 other types, and has haste.
//	 III — Sacrifice all Reflections you control."
//
// I and II are one printed chapter, so each is its own registered row with
// the same body. The target is chosen as the chapter goes on the stack, and
// the "doesn't have the same name" test is a predicate over your tokens read
// at announce and again as it resolves (CR 608.2b), so chapter II cannot pick
// the creature chapter I already copied. The token is CreateTokenCopy of the
// target with three template edits (CR 707.9a): the Legendary supertype goes,
// Reflection joins the creature types (and every other type stays), and haste
// is a printed keyword of the copy. Its own enters triggers work (the
// rulings). III sacrifices every Reflection you control as one simultaneous
// event, copies included, whatever made them so.
//
// No simplification.
func init() {
	const name = "The Apprentice's Folly"
	target := func() *game.TargetSpec {
		return TargetCreature("target nontoken creature you control that doesn't have the same name as a token you control",
			YouControl(), Not(IsTokenPredicate()), noTokenYouControlSharesItsName())
	}
	Register(Spec{
		OracleID:     "f45542a0-0d49-4b56-9c3e-663ff8d0529a",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "create a token that's a copy of target nontoken creature you control, except it isn't legendary, is a Reflection, and has haste"),
				target(), apprenticesFollyCopy),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "create a token that's a copy of target nontoken creature you control, except it isn't legendary, is a Reflection, and has haste"),
				target(), apprenticesFollyCopy),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "sacrifice all Reflections you control"),
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.HasSubtype("Reflection") {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return g.SacrificeAllKeyedForEffect(item.SourceCardID, ids, game.KeyedTail{})
				}),
		},
	})
}

// apprenticesFollyCopy is chapters I and II: the Reflection token copy.
func apprenticesFollyCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       t.ID,
		N:          1,
		Except: func(tok *game.Card) {
			tok.TypeLine = addedSubtypeTypeLine(removeLegendaryFromTypeLine(tok.TypeLine), "Reflection")
			TokenCopyGainsHaste(tok)
		},
	}.Apply(ctx)
}

// noTokenYouControlSharesItsName is "that doesn't have the same name as a
// token you control": no token the caster controls shares a name with the
// candidate (CR 201.2a), read through the layers.
func noTokenYouControlSharesItsName() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		for _, tok := range g.BattlefieldCardsForEffect() {
			if tok.IsToken() && tok.Controller == caster && tok.InstanceID != c.InstanceID && sharesANameWith(extractionNames(c))(tok) {
				return false
			}
		}
		return true
	}
}
