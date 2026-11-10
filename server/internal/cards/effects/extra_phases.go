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

	// AtThatCombat is "at the beginning of THAT combat, <body>" (Moraug,
	// Fury of Akoum; World at War): a delayed trigger bound to the first
	// combat phase this added (ADR 0188 Decision 5), so it fires at that
	// combat's beginning and at no other. Only Body, Label and Params are
	// read; At and the binding are set here. When no combat was added
	// (an anchor that did not hold) nothing is scheduled, which is the
	// card's text: there is no "that combat".
	AtThatCombat *ScheduleDelayedTrigger

	// DuringThatCombat is a rule in force "during that combat phase"
	// (Bumi, Unleashed: "Only land creatures can attack during that
	// combat phase"): a rules effect bound to the first combat phase
	// this added (#5277, ADR 0188 Decision 13), registered as that
	// combat begins and lasting until it ends. It applies in no other
	// combat, not even a newer added combat that runs first (CR 500.8).
	// When no combat was added nothing is bound, as for AtThatCombat.
	DuringThatCombat *CombatRule
}

// CombatRule is a "can't" rule for the whole of one combat phase,
// over a live scope (CR 611.2c: it changes no characteristic, so it
// reaches every object the scope names as the rule is asked).
type CombatRule struct {
	Scope        game.AffectedScope
	Restrictions game.Restriction
	Label        string
}

// OnlyLandCreaturesCanAttack is "Only land creatures can attack during
// that combat phase": "can't attack" on every creature that isn't a
// land, judged as attackers are declared (CR 508.1c).
func OnlyLandCreaturesCanAttack(label string) *CombatRule {
	return &CombatRule{Scope: game.ScopeCreaturesThatArentLands, Restrictions: game.CantAttack, Label: label}
}

func (a AddPhases) Apply(ctx *Context) error {
	ids := ctx.Game.AddPhasesForEffect(ctx.Source(), a.Anchor, a.Kinds...)
	combat := 0
	for i, k := range a.Kinds {
		if k == game.PhaseKindCombat && i < len(ids) {
			combat = ids[i]
			break
		}
	}
	if combat == 0 {
		return nil
	}
	if r := a.DuringThatCombat; r != nil {
		ctx.Game.DuringCombatPhaseRuleEffectForEffect(combat, ctx.Source(), r.Scope, ctx.Controller(),
			[]game.Mod{game.AddRestrictionsMod(r.Restrictions)}, r.Label)
	}
	if a.AtThatCombat == nil {
		return nil
	}
	dt := *a.AtThatCombat
	dt.At = game.StepBeginCombat
	dt.OnPhase = combat
	return dt.Apply(ctx)
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

// ExtraCombatAndMainAfterNthMain is "after the Nth main phase this turn,
// there is an additional combat phase followed by an additional main
// phase" (World at War). CR 505.1b counts every main phase of the turn,
// added ones included; a copy that resolves after the Nth has ended adds
// nothing (the 2010-06-15 ruling).
func ExtraCombatAndMainAfterNthMain(n int) AddPhases {
	return AddPhases{
		Anchor: game.PhaseAnchor{Kind: game.AnchorNthMainPhase, N: n},
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

// YouGetUpkeepStepsAfterThisPhase is "you get N additional upkeep steps
// after this phase" (Obeka, Splitter of Seconds; #5405). Each is a
// beginning phase holding only its upkeep step (CR 500.10), directly
// after the phase in progress, and "you get" adds nothing on another
// player's turn (CR 500.10a): the resolving ability's controller is the
// "you".
type YouGetUpkeepStepsAfterThisPhase struct {
	N int
}

func (a YouGetUpkeepStepsAfterThisPhase) Apply(ctx *Context) error {
	ctx.Game.AddUpkeepStepsAfterThisPhaseForEffect(ctx.Source(), ctx.Controller(), a.N)
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

// untapAllCreaturesYouControlAtThatCombat is the delayed-trigger body
// form of UntapAllCreaturesYouControl (Moraug, Fury of Akoum: "at the
// beginning of that combat, untap all creatures you control"). The
// controller is the trigger's, i.e. the player whose landfall added the
// combat.
func untapAllCreaturesYouControlAtThatCombat(g *game.Game, item *game.StackItem) error {
	return UntapAllCreaturesYouControl{}.Apply(NewContext(g, item))
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
