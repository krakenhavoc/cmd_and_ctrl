package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b17_helpers.go — the shared "except it's a 1/1 Spirit" clause of
// Donal, Herald of Wings and Haunting Imitation (CR 707.9b).

// vowB17SpiritCopyException is "except it's a 1/1 Spirit in addition to its
// other types": the copiable power and toughness become 1 and 1, and Spirit
// joins the subtypes. withFlying adds "and it has flying" (Haunting
// Imitation).
func vowB17SpiritCopyException(withFlying bool) func(t *game.Card) {
	return func(t *game.Card) {
		t.Power, t.Toughness = 1, 1
		t.VariableToughness = false
		t.PrintedPTKnown = true
		v := game.PrintedValues{TypeLine: t.TypeLine}
		v.AddSubtype("Spirit")
		t.TypeLine = v.TypeLine
		if withFlying && !hasKeywordWord(t.Keywords, "flying") {
			t.Keywords = append(t.Keywords, "flying")
		}
	}
}

// vowB17PutAttackingThen is putAttackingYourChoice (put_attacking.go) from
// the graveyard with a keyed tail that runs once the card has entered, told
// the IDs that did: Olivia, Crimson Bride's "It gains …" follows the entry.
// The attack target is asked as putAttackingYourChoice asks it, through
// this helper's own keyed answer.
func vowB17PutAttackingThen(ctx *Context, card uuid.UUID, name string, then game.TailRef) error {
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	tail := game.KeyedTail{Ref: then, Player: me, Source: ctx.Source()}
	if len(refs) == 1 {
		return vowB17PutAttacking(ctx.Game, me, card, refs[0].ID, tail)
	}
	options := make([]game.ChoiceOption, len(refs))
	carry := []uuid.UUID{card}
	for i, r := range refs {
		options[i] = game.ChoiceOption{Label: "Attack " + attackTargetName(ctx.Game, r.ID)}
		if r.Kind != game.AttackTargetPlayer {
			options[i].Cards = []uuid.UUID{r.ID}
		}
		carry = append(carry, r.ID)
	}
	return PickOption{
		Question: name + " — what does it attack?",
		Options:  options,
		ThenKey:  vowB17AttackAnswer,
		Carry:    carry,
	}.Apply(ctx)
}

func vowB17PutAttacking(g *game.Game, me, card, target uuid.UUID, tail game.KeyedTail) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target}, tail)
}

// vowB17AttackAnswer is the keyed answer to "what does it attack?" for
// Olivia, Crimson Bride. The grant tail is Olivia's alone.
var vowB17AttackAnswer = OptionPickThen("option-pick/olivia-crimson-bride-attack", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) < 2 {
		return nil
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return vowB17PutAttacking(ctx.Game, ctx.Controller(), card, targets[index],
		game.KeyedTail{Ref: oliviaCrimsonBrideGrantTail, Player: ctx.Controller(), Source: ctx.Source()})
})
