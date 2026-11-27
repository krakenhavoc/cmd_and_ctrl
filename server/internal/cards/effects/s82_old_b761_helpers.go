package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b761_helpers.go — shared pieces of the Tempest batch 761 cards
// (#5628).

// S82OldB761ReturnExiledCreaturesUnderYourControl is "return each creature card
// exiled with this <permanent> to the battlefield under your control": every
// creature card still in exile that the source object exiled (its last-known
// record, since the sacrifice that pays the cost has already moved it), put
// onto the battlefield together under the ability's controller (Cold Storage).
func S82OldB761ReturnExiledCreaturesUnderYourControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	return ReturnFromExileTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
}

// S82OldB761ExileTopFiveFaceDownLinked is Duplicity's "exile the top five cards
// of your library face down": one at a time, nobody may look (CR 406.3), each
// linked to this enchantment object (CR 607.2a). A short library gives up what
// it has.
func S82OldB761ExileTopFiveFaceDownLinked(g *game.Game, item *game.StackItem) error {
	for i := 0; i < 5; i++ {
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
			return nil
		}
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		if _, err := g.ExileFaceDownLinkedForEffect(item.SourceCardID, item.Controller, top); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB761DuplicityUpkeep is the yes to "you may exile all cards from your hand
// face down. If you do, put all other cards you own exiled with this enchantment
// into your hand": the other cards are the ones it already holds, read before the
// hand goes (CR 607.2a), so the hand just exiled does not come straight back.
func S82OldB761DuplicityUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var prior []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.Owner == item.Controller {
			prior = append(prior, c.InstanceID)
		}
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	var hand []uuid.UUID
	for _, c := range p.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	for _, id := range hand {
		if _, err := g.ExileFaceDownLinkedForEffect(item.SourceCardID, item.Controller, id); err != nil {
			return err
		}
	}
	for _, id := range prior {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB761DiscardACard is "discard a card": the controller picks it.
func S82OldB761DiscardACard(g *game.Game, item *game.StackItem) error {
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 1})
	return nil
}

// S82OldB761ExiledWithToGraveyards is "put all cards exiled with this enchantment
// into their owner's graveyard": every card the source object (its last-known
// record, since it may have left) still holds in exile (Duplicity).
func S82OldB761ExiledWithToGraveyards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	for _, id := range ids {
		if err := g.PutIntoGraveyardForEffect(id); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB761BeforeCombat is false from the beginning of combat on: the
// activation window "only before combat" (Maddening Imp) is the beginning phase
// and the precombat main phase, and a turn whose combat has begun is out of it.
func S82OldB761BeforeCombat() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		switch game.PhaseOf(g.Turn.Step) {
		case game.PhaseBeginning, game.PhasePrecombatMain:
			return !g.DeclareAttackersHasBegunThisTurnForEffect() && g.TurnTally.PhasesBegun[game.PhaseKindCombat] == 0
		}
		return false
	}
}

// S82OldB761ScheduleOracleTurn files Oracle en-Vec's delayed trigger for the
// beginning of `player`'s next upkeep, naming the chosen creatures as the
// objects they are now.
func S82OldB761ScheduleOracleTurn(ctx *Context, player uuid.UUID, picked []uuid.UUID) error {
	var refs []game.ObjectRef
	for _, id := range picked {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && OnBattlefield(ctx.Game, id) {
			refs = append(refs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		}
	}
	return ScheduleDelayedTrigger{
		At:     game.StepUpkeep,
		TurnOf: player,
		Label:  "Oracle en-Vec — the chosen creatures attack this turn if able, and other creatures can't attack",
		Body:   oracleEnVecTurnBegins,
		Params: game.EffectParams{Player: player, Objects: refs},
	}.Apply(ctx)
}

// oracleEnVecTurnBegins is the beginning of the chosen player's turn: every
// chosen creature still the same object must attack and is destroyed at the end
// step if it did not, and no other creature may attack this turn.
var oracleEnVecTurnBegins = game.DelayedBody("oracle-en-vec/the-turn-begins", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	var still []game.ObjectRef
	for _, o := range p.Objects {
		c, ok := g.LookupCardForEffect(o.ID)
		if !ok || c.ObjectEpoch != o.Epoch || !OnBattlefield(g, o.ID) {
			continue
		}
		still = append(still, o)
		if err := (ScopedEffectFor{
			Target:   o.ID,
			Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Oracle en-Vec - attacks this turn if able",
		}).Apply(ctx); err != nil {
			return err
		}
		if err := (ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  "Oracle en-Vec - destroy " + c.Name + " if it didn't attack",
			Cards:  []uuid.UUID{o.ID},
			Body:   destroyUnlessItAttackedBody,
			Params: game.EffectParams{Object: o},
		}).Apply(ctx); err != nil {
			return err
		}
	}
	mods := []game.Mod{game.AddRestrictionsMod(game.CantAttack)}
	scope := game.ScopeCreatures
	if len(still) > 0 {
		mods = []game.Mod{game.OnlyTheseCanAttackMod(still)}
		scope = game.ScopeCreaturesExceptObjects
	}
	g.RegisterScopedRuleEffectForEffect(ctx.Source(), scope, ctx.Controller(), mods,
		DurationUntilEndOfTurn(ctx), "Oracle en-Vec — other creatures can't attack this turn")
	return nil
})
