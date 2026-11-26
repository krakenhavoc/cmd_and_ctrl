package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b318_helpers.go — Magic Origins batch 318 (S82, #5549): the
// spell mastery condition the batch's cards share.

// OldB318SpellMastery is "Spell mastery — if there are two or more instant
// and/or sorcery cards in your graveyard" as a resolution-time condition,
// read for the resolving spell's controller while it is still on the stack,
// so the spell does not count itself (CR 608.2n).
func OldB318SpellMastery() Condition {
	return CondBoard{
		Name: "spell mastery: two or more instant and/or sorcery cards in your graveyard",
		Test: func(ctx *Context) bool { return SpellMastery(ctx.Game, ctx.Controller()) },
	}
}

// OldB318SpellTargetsYou is "spell that targets you": the candidate on the
// stack has the caster as one of its announced targets (Psychic Rebuttal).
func OldB318SpellTargetsYou(g *game.Game, caster uuid.UUID, c game.Card) bool {
	item := g.StackItemForEffect(c.InstanceID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind == game.TargetPlayer && t.ID == caster {
			return true
		}
	}
	return false
}
