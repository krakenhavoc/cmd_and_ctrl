package cardsmi

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror-Style Master — Creature — Human Warrior {4}{R}{R}, 3/3 (March of
// the Machine Commander):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 Whenever this creature attacks, for each attacking modified creature
//	 you control, create a tapped and attacking token that's a copy of that
//	 creature. Exile those tokens at end of combat. (Equipment, Auras you
//	 control, and counters are modifications.)"
//
// The creatures to copy are fixed as the ability resolves: every attacking
// creature you control that is modified (CR 700.9, b07IsModified), this one
// included. The copies are made one at a time, each through the attack
// choice (Calamity's CreateTokensAttackingYourChoice): the token enters
// attacking a player or planeswalker of your choice (CR 508.4), is not
// "declared" so fires no attack trigger, and the tokens made so far ride the
// tail until the last one is in, so a single delayed trigger exiles them
// all at the beginning of the end of combat step. A token is a copy of the
// creature's copiable values (CR 707.2), so it carries none of its counters
// or attachments and is not modified itself.
//
// No simplification.
const mirrorStyleMasterGrant = "mirror-style-master/attack-copy"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a871d70-a46d-4357-8c7d-876f7a004eb7",
		Name:         "Mirror-Style Master",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       mirrorStyleMasterGrant,
			Triggered: []game.TriggeredAbility{mirrorStyleMasterAttack()},
			Text:      "Whenever this creature attacks, for each attacking modified creature you control, create a tapped and attacking token that's a copy of that creature. Exile those tokens at end of combat.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Mirror-Style Master", 1, mirrorStyleMasterGrant),
			mirrorStyleMasterAttack(),
		},
	})
}

func mirrorStyleMasterAttack() game.TriggeredAbility {
	return effects.WheneverThisAttacks("Mirror-Style Master — for each attacking modified creature you control, create a tapped and attacking token copy; exile the tokens at end of combat",
		func(g *game.Game, item *game.StackItem) error {
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.Controller == item.Controller && c.AttackingTarget != uuid.Nil && effects.B07IsModified(g, c) {
					ids = append(ids, c.InstanceID)
				}
			}
			return mirrorStyleRound(effects.NewContext(g, item), ids, nil)
		})
}

// mirrorStyleRound makes the copy of the next creature in `ids` and hands
// the rest, with the tokens made so far, to mirrorStyleNext; with none left
// it schedules the exile of every token made.
func mirrorStyleRound(ctx *effects.Context, ids, made []uuid.UUID) error {
	for len(ids) > 0 {
		id := ids[0]
		ids = ids[1:]
		tmpl, ok := effects.TokenCopyTemplate(ctx.Game, id)
		if !ok {
			continue
		}
		tail := ctx.ItemTail(mirrorStyleNext)
		tail.Carry = slices.Concat(ids, []uuid.UUID{uuid.Nil}, made)
		return effects.CreateTokensAttackingYourChoice{
			Template: tmpl,
			Copy:     true,
			N:        1,
			Tapped:   true,
			Label:    "Mirror-Style Master",
			Tail:     tail,
		}.Apply(ctx)
	}
	if len(made) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Mirror-Style Master — exile the tokens",
		Cards: made,
		Body:  effects.ExileListedCardsBody,
	}.Apply(ctx)
}

// mirrorStyleNext is the tail after one copy lands: the Carry is the
// creatures still to copy, a Nil, then the tokens made before this one. An
// on-disk key: never renamed, never reused.
var mirrorStyleNext game.TailRef

func init() {
	// Registered here, not in the var: the body starts the next round, which
	// names this tail.
	mirrorStyleNext = effects.AttackingTokensThen("attacking-tokens/mirror-style-master-next",
		func(ctx *effects.Context, created []uuid.UUID, tail game.KeyedTail) error {
			item := tail.Asked.StackItem()
			if item == nil {
				return nil
			}
			var ids, made []uuid.UUID
			seenNil := false
			for _, id := range tail.Carry {
				switch {
				case id == uuid.Nil:
					seenNil = true
				case seenNil:
					made = append(made, id)
				default:
					ids = append(ids, id)
				}
			}
			return mirrorStyleRound(effects.NewContext(ctx.Game, item), ids, append(made, created...))
		})
}
