package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// forced_attack_helpers.go - the pieces of the "attacks this turn if able,
// destroy it at the beginning of the next end step if it didn't attack"
// family (#4062): Nettling Imp, Norritt, Arcum's Whistle and Siren's
// Call, all of which print the "before attackers are declared" window.

// DuringAnOpponentsTurn - "Activate only during an opponent's turn"
// (Nettling Imp): the active player is not the activator. In a
// multiplayer game that is any other player's turn.
func DuringAnOpponentsTurn() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return !IsYourTurn(g, controller)
	}
}

// DuringAnOpponentsTurnBeforeAttackersAreDeclared is "Activate only
// during an opponent's turn, before attackers are declared" (Nettling
// Imp).
func DuringAnOpponentsTurnBeforeAttackersAreDeclared() ActivationCondition {
	return AllConditions(DuringAnOpponentsTurn(), BeforeAttackersAreDeclared())
}

// NonWall is "non-Wall": not a creature of the Wall type.
func NonWall() CardPredicate { return Not(Subtype("Wall")) }

// ControlledByTheActivePlayer is "the active player controls": the
// candidate is under the control of whoever's turn it is.
func ControlledByTheActivePlayer() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
			return false
		}
		active := g.Seats[g.Turn.ActiveSeat]
		return active != nil && c.Controller == active.ID
	}
}

// ControlledSinceTheBeginningOfTheTurn is "controlled continuously since
// the beginning of the turn" (CR 302.6): the same reading summoning
// sickness takes - it came under its controller's control this turn if
// it is flagged as arriving this turn.
func ControlledSinceTheBeginningOfTheTurn() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.SummonedThisTurn }
}

// forcedAttackTarget is the clause of Nettling Imp, Norritt and Arcum's
// Whistle: "target non-Wall creature the active player has controlled
// continuously since the beginning of the turn".
func forcedAttackTarget() *game.TargetSpec {
	return TargetCreature(
		"target non-Wall creature the active player has controlled continuously since the beginning of the turn",
		NonWall(), ControlledByTheActivePlayer(), ControlledSinceTheBeginningOfTheTurn())
}

// destroyUnlessItAttackedBody is "destroy it if it didn't attack this
// turn": the permanent p.Object names, if it is still that object on the
// battlefield and was not declared as an attacker this turn.
var destroyUnlessItAttackedBody = game.DelayedBody("destroy/the-object-unless-it-attacked", destroyUnlessItAttacked)

func destroyUnlessItAttacked(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, id) || g.AttackedThisTurn(id) {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(NewContext(g, item))
}

// mustAttackOrBeDestroyed makes the creature attack this turn if able
// (CR 508.1d) and sets up the destruction at the beginning of the next end
// step if it didn't. The delayed trigger is the resolving ability's own
// (CR 603.7), put on the stack at that end step.
func mustAttackOrBeDestroyed(ctx *Context, creature uuid.UUID, label string) error {
	c, ok := ctx.Game.LookupCardForEffect(creature)
	if !ok {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   creature,
		Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    label + " - attacks this turn if able",
	}).Apply(ctx); err != nil {
		return err
	}
	if NonWall()(ctx.Game, ctx.Controller(), c) {
		return ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label + " - destroy " + c.Name + " if it didn't attack",
			Cards:  []uuid.UUID{creature},
			Body:   destroyUnlessItAttackedBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: creature, Epoch: c.ObjectEpoch}},
		}.Apply(ctx)
	}
	return nil
}

// forceTheChosenCreature is the Effect of "choose target non-Wall
// creature the active player has controlled continuously since the
// beginning of the turn. That creature attacks this turn if able. Destroy
// it at the beginning of the next end step if it didn't attack this
// turn" (Nettling Imp, Norritt).
func forceTheChosenCreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return mustAttackOrBeDestroyed(ctx, id, label)
	}
}
