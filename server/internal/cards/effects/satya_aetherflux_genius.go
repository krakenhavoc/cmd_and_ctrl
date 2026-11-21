package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Satya, Aetherflux Genius — Legendary Creature — Human Artificer
// {1}{U}{R}{W}, 3/5 (Modern Horizons 3 Commander, #5948):
//
//	"Menace, haste
//	 Whenever Satya attacks, create a tapped and attacking token that's a
//	 copy of up to one other target nontoken creature you control. You get
//	 {E}{E} (two energy counters). At the beginning of the next end step,
//	 sacrifice that token unless you pay an amount of {E} equal to its
//	 mana value."
//
// "Up to one" means the trigger goes on the stack with no target and still
// gives the energy. The copy enters tapped and attacking the player (or
// planeswalker's controller) Satya was attacking; it was never declared as
// an attacker, so attack triggers do not fire for it (CR 508.4). The
// energy comes first, so it can pay for the token that same turn. At the
// next end step (of any turn) a delayed trigger asks for energy equal to
// the token's mana value as it resolves and sacrifices the token unless it
// is paid (CR 118.12); a copy of a creature with mana value 0 costs
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce8af7dd-49df-4419-9776-4157efc60b0f",
		Name:            "Satya, Aetherflux Genius",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "haste"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: ThisAttacked,
			Key:       "Satya, Aetherflux Genius — create a tapped and attacking token copy; you get {E}{E}",
			Targets: Another(TargetCreature("up to one other target nontoken creature you control",
				YouControl(), Not(IsTokenPredicate())).WithCount(0, 1)),
			Effect: satyaAttackEffect,
		}},
	})
}

func satyaAttackEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var copyOf uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			copyOf = t.ID
			break
		}
	}
	var created []uuid.UUID
	if defender := b17DefendingPlayer(g, item.Trigger.Event); copyOf != uuid.Nil && defender != uuid.Nil {
		cursor := b25LastEventSeq(g)
		if err := (CreateTokenCopy{
			Controller: item.Controller,
			Copy:       copyOf,
			N:          1,
			Except: func(tok *game.Card) {
				tok.Tapped = true
				tok.AttackingTarget = defender
			},
		}).Apply(ctx); err != nil {
			return err
		}
		created = b27TokensCreatedByAfter(g, item.Controller, cursor)
	}
	if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
		return err
	}
	if len(created) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Satya, Aetherflux Genius — sacrifice the token unless you pay {E} equal to its mana value",
		Cards: created,
		Body:  gapB45SatyaSacrificeBody,
	}.Apply(ctx)
}
