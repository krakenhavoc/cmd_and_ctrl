package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vadrok, Apex of Thunder — Legendary Creature — Elemental Dinosaur Cat
// {U}{R}{W}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Mutate {1}{W/U}{R}{R} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Flying, first strike
//	 Whenever this creature mutates, you may cast target noncreature card with
//	 mana value 3 or less from your graveyard without paying its mana cost."
//
// Mutate is the engine's (ADR 0223). The target is chosen as the trigger goes
// on the stack and re-checked as it resolves (CR 608.2b); the cast is made
// during the resolution (CR 608.2g) through the shared free graveyard cast, a
// {0} cast-only permission offered with the resolution_cast prompt, whose
// decline is the "you may". Unlike Efreet Flamepainter the spell is not
// exiled: it goes to the graveyard as any spell does.
//
// No simplification.
func init() {
	spell := effects.FreeCastSpell{Filter: game.PermissionFilter{NoncreatureOnly: true}}.WithManaValueAtMost(3)
	effects.Register(effects.Spec{
		OracleID:         "775c41fa-2316-43c9-af34-ac9c9b0db7e2",
		Name:             "Vadrok, Apex of Thunder",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "first strike"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{1}{W/U}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Vadrok, Apex of Thunder — you may cast target noncreature card with mana value 3 or less from your graveyard without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.S82OldB111FirstGraveyardTarget(ctx)
					if !ok {
						return nil
					}
					effects.Cmdr2S42FreeGraveyardCast(ctx, "Vadrok, Apex of Thunder", spell, false, []uuid.UUID{id})
					return nil
				}),
				effects.TargetCardInGraveyard("target noncreature card with mana value 3 or less from your graveyard", effects.YouOwn(), effects.Noncreature(), effects.ManaValueLE(3))),
		},
	})
}
