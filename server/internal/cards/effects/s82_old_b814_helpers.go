package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b814_helpers.go — shared pieces of the S82 OLD batch 814 cards
// (Alliances, card draw and library, #5634).

// oldB814DrawAtNextUpkeepBody is the delayed half of "<that player> draws N
// cards at the beginning of the next turn's upkeep": Params.Player draws
// Params.Amount cards (the trigger's controller with no player named). An
// on-disk key: never renamed, never reused.
var oldB814DrawAtNextUpkeepBody = game.DelayedBody("old-b814/draw-n-at-next-upkeep",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		player := p.Player
		if player == uuid.Nil {
			player = item.Controller
		}
		return DrawCards{Player: player, N: p.Amount}.Apply(NewContext(g, item))
	})

// OldB814DrawAtTheNextUpkeep is "<player> draws <n> card(s) at the beginning
// of the next turn's upkeep": a CR 603.7a delayed trigger for the very next
// upkeep at the table, controlled by whoever scheduled it. uuid.Nil is the
// controller. DrawACardAtTheNextUpkeep is the one-card form for the
// controller.
func OldB814DrawAtTheNextUpkeep(label string, player uuid.UUID, n int) ScheduleDelayedTrigger {
	return ScheduleDelayedTrigger{
		At:     game.StepUpkeep,
		Label:  label,
		Body:   oldB814DrawAtNextUpkeepBody,
		Params: game.EffectParams{Player: player, Amount: n},
	}
}

// OldB814DuringCombatBeforeBlockers is "Cast this spell only during combat
// before blockers are declared" (CR 506.8): the combat phase's beginning of
// combat and declare attackers steps.
func OldB814DuringCombatBeforeBlockers(g *game.Game, _ uuid.UUID, _ game.Card) bool {
	return game.PhaseOf(g.Turn.Step) == game.PhaseCombat &&
		(g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers)
}

// oldB814ReturnUnderOwnerBody is the delayed half of "return it to the
// battlefield under its owner's control": the card <Object> names goes from a
// graveyard to the battlefield under its OWNER's control, if it is still that
// object in a graveyard (CR 400.7). An on-disk key.
var oldB814ReturnUnderOwnerBody = game.DelayedBody("old-b814/return-the-object-under-its-owners-control",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		id := p.Object.ID
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.ObjectEpoch != p.Object.Epoch {
			return nil
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: c.Owner}.Apply(NewContext(g, item))
	})

// OldB814ReturnItAtTheNextEndStepAndSkipYourNextDrawStep is Ivory Gargoyle's
// dies body: "return it to the battlefield under its owner's control at the
// beginning of the next end step and you skip your next draw step". The
// delayed return is scheduled for the card that died while it is still in
// the graveyard as the trigger resolves; the skip is the controller's next
// draw step (CR 614.10).
func OldB814ReturnItAtTheNextEndStepAndSkipYourNextDrawStep(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := NewContext(g, item)
		id := item.Trigger.Event.CardID
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			if c, ok := g.LookupCardForEffect(id); ok {
				if err := (ScheduleDelayedTrigger{
					Label:  name + " — return it to the battlefield under its owner's control",
					Cards:  []uuid.UUID{id},
					Body:   oldB814ReturnUnderOwnerBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
				}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		g.SkipNextStepForEffect(ctx.Controller(), game.StepDraw, item.SourceCardID, name+" — you skip your next draw step")
		return nil
	}
}

// OldB814BasicLandCard matches a basic land card (Lodestone Bauble).
func OldB814BasicLandCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return IsBasicLand(c) }
}

// oldB814DiscardOneOfThem is the pick of Soldevi Sage's "then discard one of
// them": the chosen card is discarded through the one discard path. An
// on-disk key.
var oldB814DiscardOneOfThem = game.RegisterChooseCardsThen("choose-cards/old-b814-discard-one-of-them", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source, game.KeyedTail{})
})

// OldB814DrawThreeThenDiscardOneOfThem is Soldevi Sage's body: the controller
// draws three cards, then discards one of the cards drawn this way (CR 608.2c:
// not a card that was already in the hand). The drawn cards are the hand's
// new ones; a draw cut short by an empty library leaves fewer to choose from.
func OldB814DrawThreeThenDiscardOneOfThem(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	before := map[uuid.UUID]bool{}
	for _, c := range p.Hand.Cards {
		before[c.InstanceID] = true
	}
	if err := g.DrawNForEffect(player, 3); err != nil {
		return err
	}
	var drawn []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !before[c.InstanceID] {
			drawn = append(drawn, c.InstanceID)
		}
	}
	if len(drawn) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Soldevi Sage — discard one of the cards you drew",
		Cards:    drawn,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  oldB814DiscardOneOfThem,
	})
	return nil
}

// oldB814MayDraw is "that player may draw a card": the asked player (not the
// ability's controller) draws on a yes. An on-disk key.
var OldB814MayDraw = MayYes("confirm/old-b814-that-player-may-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	return DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
})

// OldB814ExileThisCreature is "Exile this creature" as an activated ability's
// effect, while the source is still the permanent the ability came from
// (CR 400.7).
func OldB814ExileThisCreature(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return ExileTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}
