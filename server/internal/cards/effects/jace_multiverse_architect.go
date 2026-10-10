package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace, Multiverse Architect — Legendary Planeswalker — Jace
// {1}{W}{U}{B}{R}, loyalty 4 (Reality Fracture commander set):
//
//	"At the beginning of combat on each opponent's turn, they may pay
//	 {2}. If they don't, creatures they control can't attack Jaces you
//	 control this turn.
//	 +1: Draw two cards, then put a card from your hand on the bottom of
//	     your library.
//	 −3: Exile another target planeswalker or creature you control.
//	     Reveal cards from the top of your library until you reveal a
//	     creature or planeswalker card. Put that card onto the battlefield
//	     and the rest on the bottom of your library in a random order.
//	 Jace, Multiverse Architect can be your commander."
//
// The combat tax (#2719, ADR 0063's 2026-10-10 amendment) is the
// pay-or-else every begin-of-step "pay or else" uses (UpkeepPayUnless),
// asked of the opponent whose turn it is, so the table does not reach
// declare attackers with the question open. On a decline the opponent
// gets a this-turn attack restriction scoped to "Jaces you control":
// the Jace's controller and their other planeswalkers stay open. The
// +1, the −3 and commander eligibility are as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3321c134-bf5b-4f63-8035-fca0bbdfa86b",
		Name:         "Jace, Multiverse Architect",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, false), ByAnOpponent),
				"Jace, Multiverse Architect — that player may pay {2}", jaceMultiverseArchitectCombatTax),
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Draw two cards, then put a card from your hand on the bottom of your library.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
						return err
					}
					return rfPutCardFromHandOnBottom(ctx, "Jace, Multiverse Architect — put a card from your hand on the bottom of your library")
				},
			},
			{
				Label:   "−3: Exile another target planeswalker or creature you control. Reveal cards from the top of your library until you reveal a creature or planeswalker card. Put that card onto the battlefield and the rest on the bottom of your library in a random order.",
				Cost:    LoyaltyCost(-3),
				Targets: Another(TargetPermanent("another target planeswalker or creature you control", Or(Creature(), Planeswalker()), YouControl())),
				Effect:  rfExileTargetThenRevealCreatureOrPlaneswalker,
			},
		},
	})
}

// jaceMultiverseArchitectCombatTax is "they may pay {2}. If they don't,
// creatures they control can't attack Jaces you control this turn."
// "They" is the opponent whose beginning of combat it is; "you" is the
// Jace's controller as the trigger resolves.
func jaceMultiverseArchitectCombatTax(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	opponent, controller := item.Trigger.Event.Actor, item.Controller
	return UpkeepPayUnless{
		Chooser:  opponent,
		Cost:     "{2}",
		Question: "Jace, Multiverse Architect — pay {2}, or your creatures can't attack its controller's Jaces this turn?",
		OnDecline: func(ctx *Context) error {
			ctx.Game.GrantCantAttackPlayerThisTurnForEffect(opponent, controller,
				game.CantAttackScope{PlayerExempt: true, PlaneswalkersOnly: true, Subtype: "Jace"},
				"Jace, Multiverse Architect", ctx.Source())
			return nil
		},
	}.Apply(NewContext(g, item))
}
