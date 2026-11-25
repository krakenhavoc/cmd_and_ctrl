package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Void — Sorcery {3}{B}{R} (Eternal Masters, S82 OLD, #5542):
//
//	"Choose a number. Destroy all artifacts and creatures with mana value
//	 equal to that number. Then target player reveals their hand and
//	 discards all nonland cards with mana value equal to the number."
//
// The player is chosen on cast; the number is asked as the spell resolves
// (Fatal Lore's choose-a-number prompt, any number from 0 up). All artifacts
// and creatures with that mana value are destroyed (a token or a land creature
// is a mana value 0 object, an artifact land counts), then, once the
// destruction has settled, the target player reveals their hand and discards
// every nonland card with that mana value, all at once and not by their
// choice. A target player who is gone by resolution makes only the discard
// not happen (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdb28fd3-70fb-43c3-b367-9cf0d2c46e06",
		Name:         "Void",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseNumber{
				Chooser:  ctx.Controller(),
				Min:      0,
				NoMax:    true,
				Question: "Void — choose a number",
				Unit:     game.PayAmountCards,
				Then:     voidNumberChosen,
				Carry:    []uuid.UUID{effects.FirstLegalTargetPlayer(ctx)},
			}.Apply(ctx)
		},
	})
}

// voidNumberChosen destroys the artifacts and creatures of that mana value and,
// once they are gone, hands on to the reveal and discard. An on-disk key: never
// renamed, never reused.
var voidNumberChosen = effects.NumberThen("number/s82-old-b281-void-number",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		victim := uuid.Nil
		if len(r.Carry) > 0 {
			victim = r.Carry[0]
		}
		tail := ctx.TailN(voidDiscard, r.Number)
		tail.Carry = []uuid.UUID{victim}
		return effects.DestroyAllMatching{
			Match:    effects.And(effects.Or(effects.Artifact(), effects.Creature()), effects.ManaValueIs(r.Number)),
			ThenTail: tail,
		}.Apply(ctx)
	})

// voidDiscard is the second sentence: the target player reveals their hand and
// discards every nonland card with the chosen mana value. An on-disk key.
var voidDiscard = effects.CardTail("then/s82-old-b281-void-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || o.Tail.Carry[0] == uuid.Nil {
		return nil
	}
	victim := o.Tail.Carry[0]
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !c.IsLand() && c.ManaValue() == o.Tail.N {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(victim, ids, ctx.Source(), game.KeyedTail{})
})
