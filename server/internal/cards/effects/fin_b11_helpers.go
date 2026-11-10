package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b11_helpers.go — shared pieces of Final Fantasy batch 11 (#4462).

// finB11PutAttackingThen is putAttackingYourChoice from `from` (exile or a
// library) with one addition: a keyed tail that is told the permanent that
// entered (uuid.Nil when nothing did). Noctis, Heir Apparent's Warp-Strike
// uses it for "it can't be blocked that combat" and Winota, Joiner of Forces
// for "it gains indestructible until end of turn", which belong to the new
// object.
//
// The attack target is asked of the controller when there is more than
// one (CR 508.4), exactly as putAttackingYourChoice asks it; the card must
// still be in `from` when it moves, checked again after the answer.
func finB11PutAttackingThen(ctx *Context, card uuid.UUID, from game.ZoneKind, name string, then game.TailRef, answer game.OptionPickThen) error {
	me := ctx.Controller()
	refs := ctx.Game.AttackTargetsForEffect(me)
	if len(refs) == 0 {
		return nil
	}
	if len(refs) == 1 {
		return finB11PutAttacking(ctx.Game, me, card, from, refs[0].ID, ctx.Tail(then))
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
		ThenKey:  answer,
		Carry:    carry,
	}.Apply(ctx)
}

// finB11PutAttacking puts the card, if it is still in `from`, onto the
// battlefield under `me`'s control tapped and attacking `target`.
func finB11PutAttacking(g *game.Game, me, card uuid.UUID, from game.ZoneKind, target uuid.UUID, then game.KeyedTail) error {
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != from {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: card, From: from}},
		game.ZoneEntryOptions{Controller: me, Tapped: true, Attacking: target}, then)
}

// finB11PickedAttack reads the answer to "what does it attack?": the card
// and the attack targets ride the prompt's carry, and the index names one.
func finB11PickedAttack(r game.OptionPicked) (card, target uuid.UUID, err error) {
	if len(r.Carry) < 2 {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: no attack target to pick", game.ErrInvalidParam)
	}
	card, targets := r.Carry[0], r.Carry[1:]
	index := r.Index
	if index < 0 || index >= len(targets) {
		index = 0
	}
	return card, targets[index], nil
}

// finB12OwnedCreatureCardsHaveFlash is Teferi, Mage of Zhalfir's "creature
// cards you own that aren't on the battlefield have flash": a timing grant
// for creature spells the caster owns, from any zone but the battlefield
// (a cast never begins there).
func finB12OwnedCreatureCardsHaveFlash(label string) game.CastTimingRule {
	rule := CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true}, label)
	rule.OwnedByCaster = true
	return rule
}
