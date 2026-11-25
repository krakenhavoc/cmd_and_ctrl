package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b072_helpers.go — shared bodies for Kaldheim batch 072.

// AttachedToTheSource narrows a target clause to permanents attached to the
// permanent whose ability it is: "target Equipment attached to it" (Shackles of
// Treachery's granted trigger). Wrap the whole clause, like RelativeToSource.
func AttachedToTheSource(spec *game.TargetSpec) *game.TargetSpec {
	spec.AttachedToTheSource = true
	return spec
}

// --- Reckless Crew ---------------------------------------------------------------

// OldB072ReckEquipmentAndVehicles is "the number of Vehicles you control plus
// the number of Equipment you control" as the spell resolves (CR 608.2h): an
// object that is both counts once for each.
func OldB072ReckEquipmentAndVehicles(ctx *Context) int {
	return CountYouControl(ctx, HasSubtype(ctx.TextSubtype("Vehicle"))) + CountYouControl(ctx, HasSubtype(ctx.TextSubtype("Equipment")))
}

// OldB072OfferAttach is "For each of those tokens, you may attach an Equipment
// you control to it": one question for each token that is still on the
// battlefield, in the order they were made, each offering the Equipment the
// controller controls and none. Each answer attaches its pick to its own token
// (CR 701.3); an Equipment may move from one token to the next, as each
// instruction is its own. With no Equipment to offer nothing is asked.
func OldB072OfferAttach(g *game.Game, controller, source uuid.UUID, tokens []uuid.UUID) {
	var equipment []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Equipment") {
			equipment = append(equipment, c.InstanceID)
		}
	}
	if len(equipment) == 0 {
		return
	}
	for _, token := range tokens {
		if !OnBattlefield(g, token) {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  controller,
			Source:   source,
			Question: "Reckless Crew — you may attach an Equipment you control to this Dwarf Berserker token",
			Cards:    equipment,
			Min:      0,
			Max:      1,
			LiveZone: game.ZoneBattlefield,
			ThenKey:  OldB072AttachPicked,
			Carry:    []uuid.UUID{token},
		})
	}
}

// OldB072AttachPicked attaches the picked Equipment to the token the question
// was about. An on-disk key: never renamed, never reused.
var OldB072AttachPicked = game.RegisterChooseCardsThen("choose-cards/old-b072-reckless-crew-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) != 1 || len(r.Picked) != 1 || !OnBattlefield(g, r.Carry[0]) || !OnBattlefield(g, r.Picked[0]) {
		return nil
	}
	return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
})

// OldB072TokensMade is the token creation's continuation: the tokens that
// entered (told as the tail's IDs) are each offered an Equipment. An on-disk
// key.
var OldB072TokensMade = CardTail("tokens-then/old-b072-reckless-crew-attach", func(ctx *Context, o game.TailOutcome) error {
	OldB072OfferAttach(ctx.Game, ctx.Controller(), ctx.Source(), o.IDs)
	return nil
})

// --- In Search of Greatness ------------------------------------------------------

// OldB072GreatnessAnswer is the offer's registered answer for In Search of
// Greatness: "If you don't, scry 1", told which of the offered cards were cast.
// An on-disk key: never renamed, never reused.
var OldB072GreatnessAnswer = game.RegisterChooseCardsThen("choose-cards/old-b072-in-search-of-greatness-answer", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		return nil
	}
	return Scry{Player: r.Chooser, N: 1}.Apply(KeyedContext(g, r.Chooser, r.Source))
})

// OldB072Greatness is the upkeep trigger's body: a permanent spell from the
// hand with mana value exactly 1 plus the greatest mana value among the other
// permanents its controller controls, cast free during the resolution
// (CR 608.2g); declining, or having none to cast, scries 1.
func OldB072Greatness(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	greatest := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.ManaValue() > greatest {
			greatest = c.ManaValue()
		}
	}
	want := greatest + 1
	spell := FreeCastSpell{Filter: game.PermissionFilter{PermanentOnly: true}}.WithManaValueAtLeast(want).WithManaValueAtMost(want)
	perm := spell.On(g.FreeCastFromHandPermission(item.Controller, item.SourceCardID, "In Search of Greatness"))
	candidates := FreeCastCandidates(g, &perm, AllHandCardIDs(g, item.Controller))
	if len(candidates) == 0 {
		return Scry{Player: item.Controller, N: 1}.Apply(ctx)
	}
	offered := g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      item.Controller,
		Source:       item.SourceCardID,
		Reason:       "In Search of Greatness",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordHandCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Keep it in your hand (scry 1)",
		Then:         OldB072GreatnessAnswer,
		Carry:        candidates,
	})
	if !offered {
		g.RunResolutionCastThenForEffect(OldB072GreatnessAnswer, item.Controller, item.SourceCardID, candidates)
	}
	return nil
}
