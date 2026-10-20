package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Berserker's Frenzy — Instant {2}{R}:
//
//	"Cast this spell only before combat or during combat before blockers
//	 are declared.
//	 Roll two d20 and ignore the lower roll.
//	 1—14 | Choose any number of creatures. They block this turn if able.
//	 15—20 | You choose which creatures block this turn and how those
//	 creatures block."
//
// The roll is two d20 from the keyed random stream with the lower one
// ignored (Game.RollDiceIgnoringLowestForEffect, CR 706.6, #4174): only
// the higher roll exists as far as the game is concerned, so it is the
// only die a "whenever you roll" ability sees and the only result the
// table reads.
//
// 1-14: the caster chooses any number of creatures, on any board, as the
// spell resolves (a choice and not a target: CR 608.2c, hexproof and
// protection don't stop it), through ChoosePermanents with one leg per
// player's board, in turn order. Each creature chosen gets a block
// requirement for the turn (CR 509.1c; a record pinned to that creature,
// and a creature that leaves and returns is a new object, CR 400.7).
//
// 15-20: one DeclarationControl for the turn, blockers only
// (game/declaration_control.go, #4154), as Master Warcraft's second half:
// in every declare blockers step this turn the caster declares every other
// defending player's blocks, within the normal rules for blocking.
//
// The cast window is "before combat or during combat before blockers are
// declared" (CR 506.8b, 506.8c, 506.8g): before the combat phase (the
// beginning phase and the precombat main phase) or in a combat phase
// before its declare blockers step begins, any combat phase of a turn
// with several ("during combat", CR 506.8c). A turn whose combat is
// skipped keeps the window to the end of the precombat main phase (CR
// 506.8e), and the postcombat main phase and ending phase are outside it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5cd51d39-aebc-4367-9157-da70b8d9b1a3",
		Name:         "Berserker's Frenzy",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return beforeBlockersAreDeclared(g)
		},
		CastConditionLabel: "Cast this spell only before combat or during combat before blockers are declared.",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			rolls, err := rollDiceIgnoringLowest(ctx, 20, 2, 1)
			if err != nil || len(rolls) == 0 {
				return err
			}
			if rolls[0] >= 15 {
				ctx.Game.BeginDeclarationControlForEffect(game.DeclarationControl{
					Declarer:   item.Controller,
					Source:     item.SourceCardID,
					SourceName: "Berserker's Frenzy",
					Blockers:   true,
				})
				return nil
			}
			return ChoosePermanents{
				Question: "Berserker's Frenzy — choose any number of creatures; they block this turn if able",
				Of:       apnapLivingSeats(ctx.Game),
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && c.IsCreature() {
							out = append(out, c.InstanceID)
						}
					}
					return out, 0, len(out)
				},
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					for _, id := range picked.Cards() {
						if !onBattlefield(ctx.Game, id) {
							continue
						}
						if err := (BlockRequirementUntilEOT{
							Target: id,
							Kind:   game.BlockRequirementBlocks,
							Label:  "Berserker's Frenzy — blocks this turn if able",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}.Apply(ctx)
		},
	})
}

// beforeBlockersAreDeclared is Berserker's Frenzy's cast window, "before
// combat or during combat before blockers are declared" (CR 506.8b: before
// the declare blockers step begins, whether or not any blockers are
// declared): the beginning phase, the precombat main phase, and the
// beginning of combat and declare attackers steps of a combat phase. The
// declare blockers step and everything after it, the postcombat main phase
// and the ending phase are outside it.
//
// An added main phase between two combat phases is a postcombat main
// phase, so it is outside the window: "before combat" names the first
// combat phase (CR 506.8d) and the added one is only reachable "during
// combat" (CR 506.8c).
func beforeBlockersAreDeclared(g *game.Game) bool {
	switch game.PhaseOf(g.Turn.Step) {
	case game.PhaseBeginning, game.PhasePrecombatMain:
		return true
	case game.PhaseCombat:
		return g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers
	}
	return false
}
