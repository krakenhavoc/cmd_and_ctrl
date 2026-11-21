package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illuna, Apex of Wishes — Legendary Creature — Beast Elemental Dinosaur
// {2}{G}{U}{R}, 6/6 (Ikoria: Lair of Behemoths; reprinted in the S82
// sets' Commander decks, #5908):
//
//	"Mutate {3}{R/G}{U}{U} (If you cast this spell for its mutate cost,
//	 put it over or under target non-Human creature you own. They mutate
//	 into the creature on top plus all abilities from under it.)
//	 Flying, trample
//	 Whenever this creature mutates, exile cards from the top of your
//	 library until you exile a nonland permanent card. Put that card onto
//	 the battlefield or into your hand."
//
// Mutate is the engine's (ADR 0223): Everquill Phoenix's shape. The
// merged creature has every component's abilities (CR 702.140e), so
// flying, trample and the trigger work from underneath too, and the
// trigger fires on the merge whichever card ended on top (CR 702.140d).
//
// The run is the shared exile-until primitive (MillToZone with a
// registered stop clause, a restore point at every leg) stopping on the
// first nonland permanent card; the cards exiled on the way stay in
// exile. A library with no such card exiles all of itself and puts
// nothing anywhere. The controller then picks battlefield or hand for
// the card it stopped on (a yes/no, a restore point): onto the
// battlefield under your control, or into your hand. A card that
// chose the battlefield but could not enter (an Aura with nothing to
// enchant, CR 303.4g) stays in exile, as printed: it was not put into
// your hand instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "08daa30f-7ce0-4a72-b539-fbb0c585fd73",
		Name:             "Illuna, Apex of Wishes",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{R/G}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Illuna, Apex of Wishes — exile cards from the top of your library until you exile a nonland permanent card, then put it onto the battlefield or into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    game.MillUntil{Ref: illunaUntilNonlandPermanent},
						ThenTail: ctx.ItemTail(illunaExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// illunaUntilNonlandPermanent is the run's stop clause: the last card
// exiled is a nonland permanent card. An on-disk key.
var illunaUntilNonlandPermanent = game.RegisterMillUntil("until/illuna-nonland-permanent-card", func(landed []game.Card, _ int) bool {
	return len(landed) > 0 && illunaHit(landed[len(landed)-1])
})

// illunaHit is "a nonland permanent card".
func illunaHit(c game.Card) bool { return c.IsPermanent() && !c.IsLand() }

// illunaExiled is the clause after the run: ask battlefield or hand for
// the card it stopped on, if that card is still in exile. An on-disk key.
var illunaExiled = effects.CardTail("illuna-apex-of-wishes/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	hit := o.IDs[len(o.IDs)-1]
	c, ok := ctx.Game.LookupCardForEffect(hit)
	if z := ctx.Game.FindCardZoneForEffect(hit); !ok || z == nil || z.Kind != game.ZoneExile || !illunaHit(c) {
		return nil
	}
	_, err := ctx.Game.QueueConfirmKeyedForEffect(game.ConfirmPrompt{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Question:     "Illuna, Apex of Wishes — put " + c.Name + " onto the battlefield? Otherwise it goes into your hand.",
		AcceptLabel:  "Put it onto the battlefield",
		DeclineLabel: "Put it into my hand",
	}, illunaPut, []uuid.UUID{hit})
	return err
})

// illunaPut is the answer: accept puts the exiled card onto the
// battlefield under the chooser's control, decline puts it into its
// owner's (the chooser's) hand. A card that left exile meanwhile is left
// alone. An on-disk key: never renamed, never reused.
var illunaPut = game.RegisterConfirmThen("confirm/illuna-apex-of-wishes-put", func(g *game.Game, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	if r.Accepted {
		_, err := g.ReturnFromExileToBattlefieldForEffect(id, r.Chooser, false)
		return err
	}
	return g.BounceToHandForEffect(id)
})
