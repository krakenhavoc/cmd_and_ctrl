package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_phases.go — the catalog side of ADR 0059 Decisions 3, 4 and 8
// (#753, sub-PR 2b): additional combat, main and beginning phases, an
// additional step, and the per-turn attack history the cards that add
// them read.
//
// A card never touches the turn cursor. It says what the turn gets and
// where, and the engine splices it into the turn plan
// (game.AddPhasesForEffect, game.AddStepAfterCurrentForEffect). The
// most recently added phase happens first (CR 500.8), with no rule of
// its own.

// AddPhases is "after this [main] phase, there is an additional <kind>
// phase [followed by an additional <kind> phase]". Kinds are in the
// printed order. A no-op when the anchor does not hold — an
// AnchorThisMainPhase resolving outside a main phase adds nothing, as
// the Relentless Assault ruling says.
type AddPhases struct {
	Anchor game.PhaseAnchor
	Kinds  []game.PhaseKind
}

func (a AddPhases) Apply(ctx *Context) error {
	ctx.Game.AddPhasesForEffect(ctx.Source(), a.Anchor, a.Kinds...)
	return nil
}

// ExtraCombatAfterThisPhase is "after this phase, there is an
// additional combat phase" (Aurelia, the Warleader; Karlach; Hellkite
// Charger). No main phase comes with it: end of combat goes straight
// to the next beginning of combat (the Karlach and Aurelia rulings).
func ExtraCombatAfterThisPhase() AddPhases {
	return AddPhases{
		Anchor: game.PhaseAnchor{Kind: game.AnchorThisPhase},
		Kinds:  []game.PhaseKind{game.PhaseKindCombat},
	}
}

// ExtraCombatAndMainAfterThisMain is "after this main phase, there is
// an additional combat phase followed by an additional main phase"
// (Relentless Assault, Seize the Day, Aggravated Assault).
func ExtraCombatAndMainAfterThisMain() AddPhases {
	return AddPhases{
		Anchor: game.PhaseAnchor{Kind: game.AnchorThisMainPhase},
		Kinds:  []game.PhaseKind{game.PhaseKindCombat, game.PhaseKindMain},
	}
}

// AddStepAfterThisStep is "there is an additional <step> after this
// step" (CR 500.9; Y'shtola Rhul's end step).
type AddStepAfterThisStep struct {
	Step game.Step
}

func (a AddStepAfterThisStep) Apply(ctx *Context) error {
	ctx.Game.AddStepAfterCurrentForEffect(ctx.Source(), a.Step)
	return nil
}

// IsFirstCombatPhase is "if it's the first combat phase of the turn"
// (Karlach, Genji Glove, Raiyuu).
func IsFirstCombatPhase(g *game.Game) bool {
	return g.IsFirstCombatPhaseForEffect()
}

// CreaturesThatAttackedThisTurn is every creature on the battlefield
// that attacked this turn, whoever controls it. Read off the per-turn
// attack history, per OBJECT: a creature that left and came back is a
// new object that has not attacked (CR 400.7). A creature put onto the
// battlefield attacking was never declared as an attacker, so it has
// not attacked either.
func CreaturesThatAttackedThisTurn(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && g.AttackedThisTurn(c.InstanceID) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// AttackingCreatures is every creature attacking right now.
func AttackingCreatures(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// untapEach untaps each permanent named, one at a time. A permanent
// that is already untapped does nothing (CR 701.26b).
func untapEach(ctx *Context, ids []uuid.UUID) error {
	for _, id := range ids {
		if err := (UntapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// UntapCreaturesThatAttackedThisTurn is "untap all creatures that
// attacked this turn" (Relentless Assault, Full Throttle).
type UntapCreaturesThatAttackedThisTurn struct{}

func (UntapCreaturesThatAttackedThisTurn) Apply(ctx *Context) error {
	return untapEach(ctx, CreaturesThatAttackedThisTurn(ctx.Game))
}

// UntapAttackingCreatures is "untap all attacking creatures" (Karlach,
// Hellkite Charger).
type UntapAttackingCreatures struct{}

func (UntapAttackingCreatures) Apply(ctx *Context) error {
	return untapEach(ctx, AttackingCreatures(ctx.Game))
}

// UntapAllCreaturesYouControl is "untap all creatures you control"
// (Aurelia, the Warleader; Aggravated Assault; Éomer).
type UntapAllCreaturesYouControl struct{}

func (UntapAllCreaturesYouControl) Apply(ctx *Context) error {
	return b16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
}

// untapCreaturesThatAttackedThisTurn is the delayed-trigger body form
// of UntapCreaturesThatAttackedThisTurn (Full Throttle).
func untapCreaturesThatAttackedThisTurn(g *game.Game, item *game.StackItem) error {
	return UntapCreaturesThatAttackedThisTurn{}.Apply(NewContext(g, item))
}

// untapAllYouControlThenExtraCombat is "untap all creatures you
// control. After this phase, there is an additional combat phase"
// (Aurelia, the Warleader; Éomer, Marshal of Rohan).
var untapAllYouControlThenExtraCombat = Do(UntapAllCreaturesYouControl{}, ExtraCombatAfterThisPhase())

// untapAttackersThenExtraCombat is "untap all attacking creatures and
// after this phase, there is an additional combat phase" (Hellkite
// Charger's paid half).
func untapAttackersThenExtraCombat(ctx *Context) error {
	if err := (UntapAttackingCreatures{}).Apply(ctx); err != nil {
		return err
	}
	return ExtraCombatAfterThisPhase().Apply(ctx)
}
