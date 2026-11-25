package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Verazol, the Split Current — Legendary Creature — Serpent {X}{G}{U}, 0/0
// (Zendikar Rising, S82 OLD #5508):
//
//	"Verazol enters with a +1/+1 counter on it for each mana spent to cast
//	 it.
//	 Whenever you cast a kicked spell, you may remove two +1/+1 counters
//	 from Verazol. If you do, copy that spell. You may choose new targets
//	 for the copy. (A copy of a permanent spell becomes a token.)"
//
// The entry counters are the mana spent on the whole cost, X included
// (CR 601.2h). The cast trigger asks only while Verazol has two counters to
// remove; removing them is the yes branch, and the copy is made of the
// spell that was cast, from last-known information if it has since left the
// stack (CR 608.2h), with the chance to choose new targets (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d85e5398-fa97-4bcd-8b93-008f3f435e22",
		Name:           "Verazol, the Split Current",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ManaSpent },
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB092YouCastAKickedSpell,
				"Verazol — you may remove two +1/+1 counters from it; if you do, copy that spell",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || src.Counters[game.CounterPlusOne] < 2 {
						return nil
					}
					return effects.MayChoice{
						Question: "Verazol — remove two +1/+1 counters to copy that spell?",
						YesLabel: "Remove two counters and copy it",
						NoLabel:  "Decline",
						Then:     verazolCopy,
						Carry:    []uuid.UUID{item.Trigger.Event.CardID},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var verazolCopy = effects.MayYes("confirm/verazol-the-split-current-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || !effects.OnBattlefield(ctx.Game, ctx.Source()) || src.Counters[game.CounterPlusOne] < 2 {
		return nil
	}
	if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: -2}).Apply(ctx); err != nil {
		return err
	}
	return effects.CopySpell{
		StackID:          r.Carry[0],
		Controller:       ctx.Controller(),
		Count:            1,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(ctx)
})
