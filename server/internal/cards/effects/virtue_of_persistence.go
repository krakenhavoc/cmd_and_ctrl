package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Virtue of Persistence // Locthwain Scorn — adventure card (CR 715),
// oracle fae36da0…:
//
//	Virtue of Persistence — Enchantment {5}{B}{B}
//	  "At the beginning of your upkeep, put target creature card from a
//	   graveyard onto the battlefield under your control."
//	Locthwain Scorn — Sorcery — Adventure {1}{B}
//	  "Target creature gets -3/-3 until end of turn. You gain 2 life."
//
// Two keys, one card, as Virtue of Knowledge does: the enchantment
// keeps the bare oracle ID and the Adventure takes "<oracle_id>#1". What
// happens after Locthwain Scorn resolves — exile, and the permission to
// cast the enchantment later — is the engine's CR 715.3d branch.
//
// The upkeep trigger targets (CR 603.3d): "a graveyard" is every pile at
// the table, the pick is made as the ability goes on the stack, and it
// is re-checked at resolution (CR 608.2b). With no creature card in any
// graveyard the trigger is removed rather than put on the stack. The
// creature comes under the Virtue controller's control, not its owner's.
//
// Locthwain Scorn gains the life regardless of whether the -3/-3
// resolves on a still-legal target: if the target is gone the whole
// spell fizzles (a single target, CR 608.2b), life included.
//
// No simplification.
const virtueOfPersistenceOracleID = "fae36da0-bf9c-484f-b2fe-dd8ab2effc5c"

func init() {
	// Face 0 — the enchantment.
	Register(Spec{
		OracleID:     virtueOfPersistenceOracleID,
		Name:         "Virtue of Persistence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Virtue of Persistence — put target creature card from a graveyard onto the battlefield under your control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						targets := ctx.LegalTargets()
						if len(targets) == 0 {
							return nil
						}
						return ReturnFromGraveyard{
							Target:     targets[0].ID,
							Dest:       game.ZoneBattlefield,
							Controller: item.Controller,
						}.Apply(ctx)
					}),
				targetCreatureInAnyGraveyard()),
		},
	})

	// Face 1 — the Adventure, a sorcery while it is on the stack
	// (CR 715.3a).
	Register(Spec{
		OracleID:     virtueOfPersistenceOracleID + "#1",
		Name:         "Locthwain Scorn",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -3, Label: "Locthwain Scorn — -3/-3"}).Apply(ctx); err != nil {
					return err
				}
			}
			return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
		},
	})
}
