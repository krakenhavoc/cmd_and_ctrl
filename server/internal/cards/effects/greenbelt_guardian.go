package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greenbelt Guardian — Creature — Elf Ranger {1}{G}, 2/2:
//
//	"{G}: Target creature gains trample until end of turn.
//	 Exhaust — {3}{G}: Put three +1/+1 counters on this creature.
//	 (Activate each exhaust ability only once.)"
//
// The printed proof that exhaust is PER ABILITY (#1181). The Elf has
// two activated abilities and only the second is an exhaust ability:
// spending the exhaust must leave the repeatable {G} trample pump
// exactly as it was, and firing the pump nine times must not touch the
// exhaust. That falls out of the record's key rather than out of a
// check — it is keyed by (object, ability label), so the two abilities
// address two different cells and neither can see the other's.
//
// The same shape is what lets one card print three exhaust abilities
// and activate each once (Loot, the Pathfinder, not in the catalog:
// its first exhaust ability is a MANA ability, and the mana path does
// not write the activation record — see ADR 0020's exhaust addendum).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d8aa053-289d-40d9-baa7-9bd1c5b8e957",
		Name:         "Greenbelt Guardian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{G}: Target creature gains trample until end of turn.",
				Cost:    ManaCost("{G}"),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   target,
						Keywords: []string{"trample"},
						Label:    "Greenbelt Guardian — trample until end of turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "Exhaust — {3}{G}: Put three +1/+1 counters on this creature.",
				Exhaust: true,
				Cost:    ManaCost("{3}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterPlusOne,
						N:      3,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
