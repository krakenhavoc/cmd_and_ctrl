package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delina, Wild Mage — Legendary Creature — Elf Shaman {3}{R}, 3/2
// (Adventures in the Forgotten Realms, S82):
//
//	"Whenever Delina attacks, choose target creature you control, then
//	 roll a d20.
//	 1—14 | Create a tapped and attacking token that's a copy of that
//	 creature, except it's not legendary and it has "At end of combat,
//	 exile this token."
//	 15—20 | Create one of those tokens. You may roll again."
//
// The target is chosen as the trigger goes on the stack and re-read as it
// resolves; with it gone the ability does nothing and rolls nothing. The copy
// is of the creature's copiable values, from last-known information if it has
// left (CR 707.2, 608.2h), and "not legendary" and the quoted ability are
// part of the copy (CR 707.9a, TokenCopyGainsAbility): the token has its own
// "at end of combat, exile this token", so a copy of the token has it too. It
// enters tapped and attacking a player or planeswalker of the controller's
// choice (CR 508.4) and fires no attack trigger. On 15—20 the token comes
// first and "you may roll again" is asked once the token has landed (CR
// 706.3c): a yes rolls another d20 (same die, same table), with the same
// creature, and so on for as long as the player says yes.
//
// No simplification.
const delinaExileGrant = "delina/exile-at-end-of-combat"

func init() {
	// Registered here, not in vars: the tails name each other in a ring (a
	// roll, a token, the "roll again" question, a roll).
	delinaRolled = effects.CardTail("dice/delina-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
			return nil
		}
		id := o.Tail.Carry[0]
		tmpl, ok := effects.TokenCopyTemplate(ctx.Game, id)
		if !ok {
			return nil
		}
		tmpl.TypeLine = effects.RemoveLegendaryFromTypeLine(tmpl.TypeLine)
		effects.TokenCopyGainsAbility(&tmpl, delinaExileGrant)
		tail := ctx.ItemTail(delinaTokenMade)
		tail.Carry = []uuid.UUID{id}
		if o.Rolls[0] >= 15 {
			tail.N = 1
		}
		return effects.CreateTokensAttackingYourChoice{
			Template: tmpl,
			Copy:     true,
			N:        1,
			Tapped:   true,
			Label:    "Delina, Wild Mage",
			Tail:     tail,
		}.Apply(ctx)
	})
	delinaTokenMade = effects.AttackingTokensThen("attacking-tokens/delina-token-made",
		func(ctx *effects.Context, _ []uuid.UUID, tail game.KeyedTail) error {
			item := tail.Asked.StackItem()
			if tail.N == 0 || item == nil {
				return nil
			}
			return effects.MayChoice{
				Question: "Delina, Wild Mage — roll a d20 again?",
				Then:     delinaRollAgain,
				Carry:    tail.Carry,
			}.Apply(effects.NewContext(ctx.Game, item))
		})
	delinaRollAgain = effects.MayYes("confirm/delina-roll-again", func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.RollDice{Sides: 20, N: 1, Then: delinaRolled, TailCarry: r.Carry}.Apply(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "8d57dadb-c689-4691-b331-aba15daf46ff",
		Name:         "Delina, Wild Mage",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  delinaExileGrant,
			Text: "At end of combat, exile this token.",
			Triggered: []game.TriggeredAbility{
				effects.AtEachStep(game.StepEndCombat, "Delina's token — exile this token at end of combat", func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
			},
		}},
		Triggered: []game.TriggeredAbility{delinaAttack()},
	})
}

func delinaAttack() game.TriggeredAbility {
	t := effects.WheneverThisAttacks("Delina, Wild Mage — choose target creature you control, then roll a d20",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.RollDice{Sides: 20, N: 1, Then: delinaRolled, TailCarry: []uuid.UUID{id}}.Apply(ctx)
		})
	t.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	return t
}

// The ring of tails, set in init. On-disk keys: never renamed, never reused.
var (
	delinaRolled    game.TailRef
	delinaTokenMade game.TailRef
	delinaRollAgain game.ConfirmThen
)
