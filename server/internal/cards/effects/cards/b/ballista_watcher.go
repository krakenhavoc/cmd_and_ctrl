package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ballista Watcher // Ballista Wielder — {2}{R}{R} Creature — Human
// Soldier Werewolf 4/3 // Creature — Werewolf 5/5 (#2586, ADR 0132):
//
//	Front: "{2}{R}, {T}: This creature deals 1 damage to any target.
//	        Daybound"
//	Back:  "{2}{R}: This creature deals 1 damage to any target. A
//	        creature dealt damage this way can't block this turn.
//	        Nightbound"
//
// The back face's rider is read from the damage's continuation: the
// restriction lands only on a creature that was actually dealt damage,
// so a prevented ping leaves it free to block. It is a CR 509.1b
// restriction that ends with the turn, and a creature that left and came
// back is a new object that blocks freely (CR 400.7).
//
// No simplification.
func init() {
	const oracle = "b6811c31-fcd3-4d00-89d5-cf974575a87c"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Ballista Watcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Ballista Wielder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to any target. A creature dealt damage this way can't block this turn.",
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  ballistaWielderPing,
		}},
	})
}

func ballistaWielderPing(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	controller, source := item.Controller, item.SourceCardID
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	if t.Kind != game.TargetCard {
		return effects.DealDamage{Source: source, Target: t.ID, Amount: 1}.Apply(ctx)
	}
	// The rider is a keyed tail (#4253), told the creature and the damage
	// dealt, so a prompt the damage pauses on is a restore point.
	return g.DealDamageToCreatureKeyedForEffect(source, t.ID, 1, game.KeyedTail{Ref: ballistaWielderDealt, Player: controller, Source: source})
}

// ballistaWielderDealt is "a creature dealt damage this way can't block
// this turn", told the creature and the damage dealt. An on-disk key.
var ballistaWielderDealt = game.RegisterTail("then/ballista-wielder-cant-block", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.IDs) == 0 {
		return nil
	}
	return ballistaWielderCantBlock(g, o.Tail.Player, o.Tail.Source, o.IDs[0])
})

func ballistaWielderCantBlock(g *game.Game, controller, source, creature uuid.UUID) error {
	c, ok := g.LookupCardForEffect(creature)
	if !ok || !c.IsCreature() {
		return nil
	}
	return effects.RestrictUntilEOT{
		Target:       creature,
		Restrictions: game.CantBlock,
		Label:        "Ballista Wielder — can't block this turn",
	}.Apply(effects.NewContext(g, &game.StackItem{Controller: controller, SourceCardID: source}))
}
