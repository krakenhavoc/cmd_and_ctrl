package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breena, the Demagogue — Legendary Creature — Bird Warlock {1}{W}{B},
// 1/3 (EDHREC rank 1878):
//
//	"Flying
//	 Whenever a player attacks one of your opponents, if that opponent
//	 has more life than another of your opponents, that attacking
//	 player draws a card and you put two +1/+1 counters on a creature
//	 you control."
//
// The politics commander: attack the opponent who is ahead and
// Breena pays you a card while her controller grows a creature.
// "Attacks one of your opponents" is read off EventAttack's target —
// through the defending player, so an attack at that opponent's
// planeswalker or battle counts (S27) — and the attacker may be
// anyone, Breena's controller included. The ability triggers ONCE
// PER OPPONENT ATTACKED per combat, not once per creature: the
// engine emits EventAttack per creature, so the guard is the
// engine's own once-per-batch key with its PLAYER dimension
// (OncePerBatchPerPlayer, CR 603.2c / #784, reading the attack
// through the defending player). One declaration is one batch, so
// the later creatures aimed at the same opponent are declined while
// a split attack on two opponents still triggers twice. The stack
// label carries the attacked player's name, which is what tells two
// simultaneous triggers apart and what the once-per-turn tally below
// keys on.
//
// The intervening if (CR 603.4) — the attacked opponent has more
// life than at least one OTHER opponent of Breena's controller — is
// checked as the trigger would fire and again on resolution.
//
// Sandbox simplification, declared (the Azorius Chancery posture):
// "a creature you control" is a resolution-time choice, and the
// pick_target prompt is the one picker the engine has, so the
// creature is picked as the trigger goes on the stack. Two
// consequences, both weaker than printed: opponents see the pick
// before the ability resolves, and a creature of yours with hexproof
// or shroud cannot be picked. Breena herself is always a legal
// choice, so the trigger is never dropped for want of a target.
//
// One more declared retreat, also weaker: with an extra combat in
// the same turn the ability would not fire again for the same
// opponent (Aurelia's note — the engine has no extra combats). The
// retreat that used to sit beside it is gone with #784: an open
// creature pick for one attacked opponent no longer swallows a
// second opponent's trigger, because the guard is keyed on the batch
// and the opponent rather than on what is in flight.
func init() {
	Register(Spec{
		OracleID:     "d11e627b-8a48-411d-a261-2c9a02a758ba",
		Name:         "Breena, the Demagogue",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The creature that gets the two +1/+1 counters is picked when the trigger goes on the stack rather than as it resolves, so opponents can respond to the choice, and a creature with hexproof or shroud can't be picked.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			OncePerBatch: true,
			BatchKey:     PerPlayer,
			Key:          b17BreenaKey,
			Watches:      []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				opp := b17DefendingPlayer(g, ev)
				if opp == uuid.Nil || !b17OpponentHasMoreLifeThanAnother(g, source.Controller, opp) {
					return false
				}
				// The once-per-combat half is the engine's batch guard,
				// keyed per attacked opponent; this is the extra-combat
				// retreat below, keyed on the same opponent's label.
				return !b11TriggeredThisTurn(g, source.InstanceID, b17BreenaLabel(g, opp))
			},
			Targets: TargetCreature("a creature you control", YouControl()),
			// A fill-in Build (ADR 0041 P9): the stack label names the
			// attacked opponent, a fact of the moment Breena triggered.
			// The effect re-derives that opponent from item.Trigger at
			// resolution rather than closing over it.
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				opp := b17DefendingPlayer(g, ev)
				return game.NewTriggeredItem(source, b17BreenaLabel(g, opp))
			},
			Effect: breenaEffect,
		}},
	})
}

// breenaEffect is Breena's resolution: read the attacker and the
// attacked opponent off item.Trigger.Event rather than a closure, so
// a restored item resolves against the current board's answer to the
// intervening if.
func breenaEffect(g *game.Game, item *game.StackItem) error {
	ev := item.Trigger.Event
	opp := b17DefendingPlayer(g, ev)
	if !b17OpponentHasMoreLifeThanAnother(g, item.Controller, opp) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: ev.Actor, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}
