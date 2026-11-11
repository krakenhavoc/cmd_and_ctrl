package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusionist's Gambit — Instant {2}{U}{U}:
//
//	"Cast this spell only during the declare blockers step on an
//	 opponent's turn.
//	 Remove all attacking creatures from combat and untap them. After
//	 this phase, there is an additional combat phase. Each of those
//	 creatures attacks that combat if able. They can't attack you or
//	 planeswalkers you control that combat."
//
// The timing is a CastCondition: the declare blockers step of a turn
// whose active player is not the caster.
//
// "Those creatures" are the creatures attacking as the spell resolves,
// whoever controls them. Each is removed from combat (CR 506.4) and
// untapped, and the combat is added after this phase, whoever's turn it
// is (CR 500.8).
//
// The last two sentences are one rule bound to the combat this spell
// adds and to no other (ADR 0188 Decision 14, #5685): an attack
// requirement (CR 508.1d) and "can't attack you or planeswalkers you
// control" (CR 508.1c) over those creatures, registered as that combat
// begins and ending with it. A combat another effect adds later runs
// first (CR 500.8) and is not bound. A creature that has left the
// battlefield or changed controller when that combat begins is free of
// both. The protected player is the caster, fixed as the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "333745d9-e930-439b-94d6-3aeea2877f69",
		Name:         "Illusionist's Gambit",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			active := activePlayerIDOf(g)
			return active != uuid.Nil && active != controller && g.Turn.Step == game.StepDeclareBlockers
		},
		CastConditionLabel: "Cast this spell only during the declare blockers step on an opponent's turn.",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			those := AttackingCreatures(ctx.Game)
			for _, id := range those {
				if err := ctx.Game.RemoveFromCombatForEffect(id); err != nil {
					return err
				}
			}
			if err := untapEach(ctx, those); err != nil {
				return err
			}
			extra := ExtraCombatAfterThisPhase()
			extra.ThoseCreaturesThatCombat = &CombatObjectsRule{
				Objects: those,
				Mods: []game.Mod{
					game.AddAttackRequirementMod(uuid.Nil),
					game.CantAttackPlayerMod(item.Controller),
				},
				Label: "Illusionist's Gambit — attacks this combat if able, and can't attack its caster or their planeswalkers",
			}
			return extra.Apply(ctx)
		},
	})
}
