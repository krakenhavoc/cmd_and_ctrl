package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sun-Blessed Guardian // Furnace-Blessed Conqueror — {1}{W} Creature —
// Human Cleric 2/2 // Creature — Phyrexian Cleric 3/3 (March of the
// Machine):
//
//	Front: "{5}{R/P}: Transform this creature. Activate only as a sorcery."
//	Back:  "Whenever this creature attacks, create a tapped and attacking
//	        token that's a copy of it. Put a +1/+1 counter on that token
//	        for each +1/+1 counter on this creature. Sacrifice that token
//	        at the beginning of the next end step."
//
// The copy is made through CreateTokensAttackingYourChoice (never declared
// as an attacker, so it fires no "attacks" triggers, CR 508.4), taking the
// copiable values of the Conqueror as it is (CR 707.2). The number of
// counters is read when the trigger resolves, from the Conqueror on the
// battlefield or, if it has gone, from its last known information
// (CR 608.2h), and rides the tail as data. The tail puts the counters on
// the token, then schedules its sacrifice at the next end step.
//
// No simplification.
func init() {
	const oracle = "0b34581e-66df-4e2f-91fb-dddba632d203"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Sun-Blessed Guardian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{5}{R/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{5}{R/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Furnace-Blessed Conqueror",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Furnace-Blessed Conqueror — create a tapped and attacking copy of it with the same number of +1/+1 counters; sacrifice it at the next end step",
				furnaceBlessedConquerorCopy),
		},
	})
}

func furnaceBlessedConquerorCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	source := ctx.Source()
	tmpl, ok := TokenCopyTemplate(g, source)
	if !ok {
		return nil
	}
	counters := 0
	if info, ok := ctx.SourcePermanent(); ok && !info.Left {
		counters = info.Counters[game.CounterPlusOne]
	} else {
		counters = g.LastKnownCountersForEffect(source)[game.CounterPlusOne]
	}
	return CreateTokensAttackingYourChoice{
		Template: tmpl,
		Copy:     true,
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Furnace-Blessed Conqueror",
		Tail:     game.KeyedTail{Ref: furnaceBlessedConquerorTail, N: counters},
	}.Apply(ctx)
}

// furnaceBlessedConquerorTail puts the counters on the copy and sacrifices
// it at the next end step. An on-disk key: never renamed, never reused.
var furnaceBlessedConquerorTail = AttackingTokensThen("attacking-tokens/furnace-blessed-conqueror-counters-and-sacrifice",
	func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		if tail.N > 0 {
			if err := (AddCounter{Target: created[0], Kind: game.CounterPlusOne, N: tail.N}).Apply(ctx); err != nil {
				return err
			}
		}
		return ScheduleDelayedTrigger{
			Label: "Furnace-Blessed Conqueror — sacrifice the token",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	})
