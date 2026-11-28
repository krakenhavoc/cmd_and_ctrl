package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b608_helpers.go — Eighth Edition, batch 608 (#5607): the life-gain
// artifacts, Sanctimony and Blessed Reversal.

// OldB608CastASpellOfColor is the condition of Wooden Sphere's and Ivory Cup's
// triggers: any player cast a spell that is the color the source's text names
// (CR 612.2), read as the spell sits on the stack.
func OldB608CastASpellOfColor(letter string) func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || source == nil {
			return false
		}
		spell, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		want := source.TextColor(letter)
		for _, c := range spell.EffectiveColors() {
			if strings.EqualFold(c, want) {
				return true
			}
		}
		return false
	}
}

// OldB608MayPayOneToGainOne is "you may pay {1}. If you do, you gain 1 life."
func OldB608MayPayOneToGainOne(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MayPay{
			Chooser:  item.Controller,
			Cost:     "{1}",
			Question: label + " — pay {1} to gain 1 life?",
			Then:     oldB608GainOneLife,
		}.Apply(NewContext(g, item))
	}
}

// oldB608GainOneLife is the paid branch: gain 1 life. An on-disk key.
var oldB608GainOneLife = PayUnlessBranches("pay-unless/old-b608-gain-1-life",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)

// OldB608AnOpponentTappedAMountain is Sanctimony's condition: a land that is a
// Mountain (as the Enchantment's text now says the word, CR 612.2) was tapped
// for mana by an opponent of its controller.
func OldB608AnOpponentTappedAMountain(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Actor == source.Controller || !B20LandTappedForMana(ev, g) {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.Source)
	return ok && land.HasSubtype(source.TextSubtype("Mountain"))
}

// OldB608BlessedReversal is "You gain 3 life for each creature attacking you",
// counted as the spell resolves; a creature attacking one of your planeswalkers
// is not attacking you.
func OldB608BlessedReversal(_ *game.StackItem, ctx *Context) error {
	n := CreaturesAttackingPlayer(ctx.Game, ctx.Controller())
	if n <= 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: 3 * n}.Apply(ctx)
}

// OldB608PeachGardenOath is "You gain 2 life for each creature you control",
// counted as the spell resolves.
func OldB608PeachGardenOath(_ *game.StackItem, ctx *Context) error {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: 2 * n}.Apply(ctx)
}

// OldB608GainThreeOrPreventThree is the printed text of Healing Leaves and
// Healing Salve alike: "Choose one — • Target player gains 3 life. • Prevent the
// next 3 damage that would be dealt to any target this turn." `name` words the
// shield's log label.
func OldB608GainThreeOrPreventThree(name string) *game.ModeSpec {
	return ChooseOne(
		ModeDoing("Target player gains 3 life.", TargetPlayer("target player"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return GainLife{Player: t.ID, Amount: 3}.Apply(ctx)
			}),
		ModeWithPurpose(ModeDoing("Prevent the next 3 damage that would be dealt to any target this turn.", TargetAny(),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return PreventNextDamage{Target: t.ID, Amount: 3, Label: name + ": prevent the next 3 damage"}.Apply(ctx)
			}), game.Purpose{Answers: game.AnswerPrevent}),
	)
}
