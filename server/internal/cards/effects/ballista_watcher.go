package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Ballista Watcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}: This creature deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Ballista Wielder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature deals 1 damage to any target. A creature dealt damage this way can't block this turn.",
			Cost:    ManaCost("{2}{R}"),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  ballistaWielderPing,
		}},
	})
}

func ballistaWielderPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller, source := item.Controller, item.SourceCardID
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	t := targets[0]
	if t.Kind != game.TargetCard {
		return DealDamage{Source: source, Target: t.ID, Amount: 1}.Apply(ctx)
	}
	id := t.ID
	return g.DealDamageToCreatureThenForEffect(source, id, 1, func(g *game.Game, dealt int) error {
		if dealt <= 0 {
			return nil
		}
		return ballistaWielderCantBlock(g, controller, source, id)
	})
}

func ballistaWielderCantBlock(g *game.Game, controller, source, creature uuid.UUID) error {
	c, ok := g.LookupCardForEffect(creature)
	if !ok || !c.IsCreature() {
		return nil
	}
	return RestrictUntilEOT{
		Target:       creature,
		Restrictions: game.CantBlock,
		Label:        "Ballista Wielder — can't block this turn",
	}.Apply(NewContext(g, &game.StackItem{Controller: controller, SourceCardID: source}))
}
