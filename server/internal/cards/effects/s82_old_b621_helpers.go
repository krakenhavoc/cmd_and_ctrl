package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b621_helpers.go — shared pieces of Judgment batch 621 (#5611).

// OldB621CardToBottomOfLibrary puts the first legal target card, wherever
// it is in a graveyard, on the bottom of its owner's library.
func OldB621CardToBottomOfLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return PutInLibraryInAnyOrder{
			Cards:     []uuid.UUID{t.ID},
			From:      game.ZoneGraveyard,
			Placement: game.LibraryPlaceBottom,
			Label:     "put the card on the bottom of its owner's library",
		}.Apply(ctx)
	}
	return nil
}

// OldB621CastFromAGraveyard passes for a spell on the stack that was cast from
// a graveyard (Laquatus's Disdain): flashback, Yawgmoth's Will and the rest.
func OldB621CastFromAGraveyard() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		return item != nil && item.CastFromZone == game.ZoneGraveyard
	}
}

// OldB621CounterTheTargetSpell is the first sentence of Laquatus's Disdain:
// "Counter target spell cast from a graveyard." A target that has left the
// stack does nothing (CR 608.2b).
var OldB621CounterTheTargetSpell = Instruction(func(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		return CounterTarget{StackID: t.ID}.Apply(ctx)
	}
	return nil
})

// OldB621CounterUnlessItExilesItsGraveyard is the first sentence of Grip of
// Amnesia: "Counter target spell unless its controller exiles all cards from
// their graveyard." The spell's controller is asked as the ability resolves
// and the stack waits for the answer. "All cards" is as many as the graveyard
// holds now, so the only way to pay is to exile every one of them; a spell
// whose controller has an empty graveyard has nothing to exile and is not
// countered (CR 118.12: the payment is made in full).
var OldB621CounterUnlessItExilesItsGraveyard = Instruction(func(ctx *Context) error {
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	item := ctx.Game.StackItemForEffect(targets[0].ID)
	if item == nil {
		return nil
	}
	chooser := item.Controller
	n := B31GraveyardSize(ctx.Game, chooser)
	if n == 0 {
		return nil
	}
	action := &game.PayAction{Kind: game.PayActionExileFromGraveyard, Count: n}
	return ctx.Game.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
		StackItem: targets[0].ID,
		Chooser:   chooser,
		Source:    ctx.Source(),
		Cost:      "Exile all cards from your graveyard",
		Action:    action,
		Question:  "Grip of Amnesia — exile all cards from your graveyard, or your spell is countered",
	})
})

// OldB621GraveConsequences is "Each player may exile any number of cards from
// their graveyard. Then each player loses 1 life for each card in their
// graveyard. Draw a card." Each player is asked in APNAP order (CR 101.4) about
// their own graveyard, every pick is exiled together once the last has
// answered, then the life is lost and the controller draws.
var OldB621GraveConsequences = Instruction(func(ctx *Context) error {
	g := ctx.Game
	players := ApnapOrder(g, TablePlayers(ctx))
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		cands := GraveyardCardIDs(ctx, id, func(game.Card) bool { return true })
		if len(cands) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   "Grave Consequences — you may exile any number of cards from your graveyard",
			Cards:      cands,
			Min:        0,
			Max:        0,
			LiveZone:   game.ZoneGraveyard,
		})
	}
	if len(legs) == 0 {
		return oldB621GraveConsequencesFinish(ctx)
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, oldB621GraveConsequencesExile, ctx.Source(), ctx.RunData(players))
	return err
})

var oldB621GraveConsequencesExile = RunThen("prompt-run/old-b621-grave-consequences-exile", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	var all []uuid.UUID
	for _, id := range r.Carry {
		all = append(all, picks.By(id)...)
	}
	if len(all) == 0 {
		return oldB621GraveConsequencesFinish(ctx)
	}
	return ctx.Game.ExileCardsKeyedForEffect(all, ctx.ItemTail(oldB621GraveConsequencesThen))
})

var oldB621GraveConsequencesThen = CardTail("old-b621/grave-consequences-then", func(ctx *Context, _ game.TailOutcome) error {
	return oldB621GraveConsequencesFinish(ctx)
})

func oldB621GraveConsequencesFinish(ctx *Context) error {
	for _, id := range ApnapOrder(ctx.Game, TablePlayers(ctx)) {
		if n := B31GraveyardSize(ctx.Game, id); n > 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -n); err != nil {
				return err
			}
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// OldB621OpponentOwns passes for a card its caster does not own: "from an
// opponent's graveyard".
func OldB621OpponentOwns() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner != caster }
}

// OldB621ExileTargetCreatureCardItDealsItsPower is Selfless Exorcist's ability:
// the target creature card is exiled, then it deals damage equal to its power
// (read where it was, in the graveyard, before it moved) to the source.
func OldB621ExileTargetCreatureCardItDealsItsPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	power := 0
	if c, found := g.LookupCardForEffect(id); found {
		power = c.CurrentPower()
	}
	if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return DealDamage{Source: id, Target: item.SourceCardID, Amount: power}.Apply(ctx)
}

// OldB621PulsemageAdvocate is the whole of the Pulsemage Advocate's ability:
// the three cards of the first clause go to their owners' hands, and the
// creature card of the second returns to the battlefield under your control.
// A target that is no longer legal is skipped (CR 608.2b).
func OldB621PulsemageAdvocate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		dest := game.ZoneHand
		var controller uuid.UUID
		if t.Slot == 1 {
			dest = game.ZoneBattlefield
			controller = ctx.Controller()
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: dest, Controller: controller}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB621NomadMythmaker is "Put target Aura card from a graveyard onto the
// battlefield under your control attached to a creature you control": the
// Aura chooses its host among the creatures you control as it enters (CR
// 303.4f), and with none it stays in the graveyard (CR 303.4g).
func OldB621NomadMythmaker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := LegalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:                 ids[:1],
		Controller:              ctx.Controller(),
		AuraHostsCreaturesOnly:  true,
		AuraHostsYouControlOnly: true,
	}.Apply(ctx)
}
