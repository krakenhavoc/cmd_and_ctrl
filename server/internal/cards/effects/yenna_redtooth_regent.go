package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yenna, Redtooth Regent — Legendary Creature — Elf Noble {2}{G}{W}, 4/4
// (Wilds of Eldraine, #4471):
//
//	"{2}, {T}: Choose target enchantment you control that doesn't have the
//	 same name as another permanent you control. Create a token that's a
//	 copy of it, except it isn't legendary. If the token is an Aura, untap
//	 Yenna, then scry 2. Activate only as a sorcery."
//
// The target is an enchantment you control with no OTHER permanent of yours
// sharing its name, judged at announce and again at resolution (CR 608.2b).
// The token is CreateTokenCopiesTogether with "isn't legendary" as a
// template edit (CR 707.9a). A copy of an Aura enters without being cast and
// chooses what it enchants as it enters (CR 303.4f, the engine's), and with
// no legal host no token is created: the rider reads the tokens that really
// landed (the tail), so "if the token is an Aura" is then false and Yenna
// stays tapped. It untaps Yenna (still the permanent that activated) and
// then scries 2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0adede77-5181-4190-8066-4797cefe0189",
		Name:         "Yenna, Redtooth Regent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Choose target enchantment you control that doesn't have the same name as another permanent you control. " +
				"Create a token that's a copy of it, except it isn't legendary. If the token is an Aura, untap Yenna, then scry 2. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}"), TapCost()),
			SorcerySpeed: true,
			Targets: TargetPermanent("target enchantment you control that doesn't have the same name as another permanent you control",
				Enchantment(), YouControl(), noOtherPermanentYouControlSharesItsName()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return CreateTokenCopiesTogether{
					Controller: item.Controller,
					Copies:     []uuid.UUID{t.ID},
					Except:     func(tok *game.Card) { tok.TypeLine = removeLegendaryFromTypeLine(tok.TypeLine) },
					Then:       ctx.ItemTail(yennaAfterCopy),
				}.Apply(ctx)
			},
		}},
	})
}

// yennaAfterCopy is the rider: if a token that landed is an Aura, untap
// Yenna (while she is still the permanent that activated), then scry 2. An
// on-disk key.
var yennaAfterCopy = CardTail("yenna-redtooth-regent/after-copy", func(ctx *Context, o game.TailOutcome) error {
	isAura := false
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsAura() {
			isAura = true
		}
	}
	if !isAura {
		return nil
	}
	if sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		if err := (UntapTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
			return err
		}
	}
	return Scry{N: 2}.Apply(ctx)
})

// noOtherPermanentYouControlSharesItsName is "that doesn't have the same
// name as another permanent you control": the candidate itself doesn't
// count against it (CR 201.2a).
func noOtherPermanentYouControlSharesItsName() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		match := sharesANameWith(extractionNames(c))
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == caster && p.InstanceID != c.InstanceID && match(p) {
				return false
			}
		}
		return true
	}
}
