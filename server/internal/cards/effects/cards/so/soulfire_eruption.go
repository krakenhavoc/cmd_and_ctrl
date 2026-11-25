package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulfire Eruption — Sorcery {6}{R}{R}{R}:
//
//	"Choose any number of target creatures, planeswalkers, and/or players.
//	 For each of them, exile the top card of your library, then Soulfire
//	 Eruption deals damage equal to that card's mana value to that
//	 permanent or player. You may play the exiled cards until the end of
//	 your next turn."
//
// A loop over the targets still legal as the spell resolves (CR 608.2b),
// in the order they were chosen (CR 608.2c, #6166): for each one, the top
// card of the caster's library is exiled with the impulse permission
// (play it, a land included, until the end of the caster's next turn),
// and then the spell deals that card's mana value as damage to that
// target (a land, or nothing exiled, deals none). The next target's card
// is exiled only once the previous one's damage is dealt (a keyed damage
// door, so a CR 616 prompt on the damage is a restore point too), so each amount
// reads its own card. Each step is a registered tail carrying the targets
// still to go, so a pause on the way (a commander's CR 903.9 question) is
// a restore point.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     SoulfireEruptionOracleID,
		Name:         "Soulfire Eruption",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreaturePlaneswalkerOrPlayer("any number of target creatures, planeswalkers, and/or players").WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				ids = append(ids, t.ID)
			}
			return soulfireEruptionNext(ctx, ids)
		},
	})
}

// SoulfireEruptionOracleID is the card's oracle_id.
const SoulfireEruptionOracleID = "089dd3b7-c3ed-4734-8789-a3c1ed9a4e40"

// soulfireEruptionNext exiles the top card for the first target of
// `rest`, then deals its damage from the tail.
func soulfireEruptionNext(ctx *effects.Context, rest []uuid.UUID) error {
	if len(rest) == 0 {
		return nil
	}
	perm := game.CastPermission{Duration: ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())}
	tail := ctx.ItemTail(soulfireEruptionExiled)
	tail.Carry = append([]uuid.UUID(nil), rest...)
	return ctx.Game.ExileTopWithPermissionKeyedForEffect(ctx.Controller(), ctx.Controller(), 1, perm, tail)
}

// soulfireEruptionExiled deals the exiled card's mana value to the first
// carried target, then goes on with the rest (on-disk key
// "soulfire-eruption/exiled"); soulfireEruptionDealt is its sequel.
var soulfireEruptionExiled, soulfireEruptionDealt game.TailRef

// The two tails call each other through soulfireEruptionNext, so they are
// registered here rather than in their declarations (an initialization
// cycle otherwise).
func init() {
	soulfireEruptionExiled = effects.CardTail("soulfire-eruption/exiled", soulfireEruptionExiledBody)
	soulfireEruptionDealt = effects.CardTail("soulfire-eruption/dealt", soulfireEruptionDealtBody)
}

func soulfireEruptionExiledBody(ctx *effects.Context, o game.TailOutcome) error {
	rest := o.Tail.Carry
	if len(rest) == 0 {
		return nil
	}
	target, rest := rest[0], rest[1:]
	mv := 0
	if len(o.IDs) > 0 {
		if c, ok := ctx.Game.LookupCardForEffect(o.IDs[0]); ok {
			mv = c.ManaValue()
		}
	}
	next := ctx.ItemTail(soulfireEruptionDealt)
	next.Carry = append([]uuid.UUID(nil), rest...)
	switch {
	case mv > 0 && ctx.Game.PlayerByIDForEffect(target) != nil:
		return ctx.Game.DealDamageToPlayerKeyedForEffect(ctx.Source(), target, mv, next)
	case mv > 0 && effects.OnBattlefield(ctx.Game, target):
		return ctx.Game.DealDamageToCreatureKeyedForEffect(ctx.Source(), target, mv, next)
	}
	return soulfireEruptionNext(ctx, rest)
}

// soulfireEruptionDealtBody goes on with the targets still to go once a
// target's damage is dealt (on-disk key "soulfire-eruption/dealt").
func soulfireEruptionDealtBody(ctx *effects.Context, o game.TailOutcome) error {
	return soulfireEruptionNext(ctx, o.Tail.Carry)
}
