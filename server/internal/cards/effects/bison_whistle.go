package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bison Whistle — Artifact {1}{G} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"{1}, {T}: Look at the top card of your library. If it's a Bison
//	 card, you may put it onto the battlefield. If it's a creature card,
//	 you may reveal it and put it into your hand. Otherwise, you may put
//	 it into your graveyard."
//
// Three sentences read in order about the one looked-at card:
//
//   - A Bison permanent card is offered to the battlefield first (Risen
//     Reef's PutFromLibraryOntoBattlefield, optional, so the entry goes
//     through the CR 614 pipeline).
//   - If it is still on top and is a creature card, you may reveal it and
//     put it into your hand (Herald's Horn's reveal-and-take). A creature
//     you decline both ways stays on top: "Otherwise" is the else of the
//     creature sentence.
//   - A card that is not a creature card may be put into your graveyard
//     (a plain move, not a mill).
//
// "Bison" is read through the ability's text (CR 612.2). Declining every
// offer leaves the card where it was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47e77918-13ec-47ce-9256-5a3b5bf8d315",
		Name:         "Bison Whistle",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Look at the top card of your library. If it's a Bison card, you may put it onto the battlefield. If it's a creature card, you may reveal it and put it into your hand. Otherwise, you may put it into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
				if len(looked) == 0 {
					return nil
				}
				c, ok := g.LookupCardForEffect(looked[0])
				if !ok {
					return nil
				}
				bison := ctx.TextSubtype("Bison")
				if !c.HasSubtype(bison) {
					return bisonWhistleCreatureOrGraveyard(ctx, looked[0])
				}
				return PutFromLibraryOntoBattlefield{
					Player:   item.Controller,
					Cards:    looked,
					Match:    HasSubtype(bison),
					Max:      1,
					Optional: true,
					Label:    "Bison Whistle — you may put the Bison card onto the battlefield",
					ThenTail: bisonWhistleAfterBison,
				}.Apply(ctx)
			},
		}},
	})
}

// bisonWhistleAfterBison is the rest of the sentence once the Bison
// offer is answered: a card still on top goes on to the creature
// sentence.
var bisonWhistleAfterBison = PutFromLibraryTail("avatar-b30/bison-whistle-after-bison", func(ctx *Context, res PutFromLibraryResult) error {
	if len(res.Rest) == 0 {
		return nil
	}
	return bisonWhistleCreatureOrGraveyard(ctx, res.Rest[0])
})

// bisonWhistleCreatureOrGraveyard is the second and third sentences
// about the looked-at card.
func bisonWhistleCreatureOrGraveyard(ctx *Context, top uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(top)
	if !ok {
		return nil
	}
	if c.IsCreature() {
		return MayChoice{
			Player:   ctx.Controller(),
			Question: "Bison Whistle — reveal " + c.Name + " and put it into your hand?",
			YesLabel: "Put it into your hand",
			NoLabel:  "Leave it on top",
			Then:     bisonWhistleTakeCreature,
			Carry:    []uuid.UUID{top},
		}.Apply(ctx)
	}
	return MayChoice{
		Player:   ctx.Controller(),
		Question: "Bison Whistle — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     bisonWhistleToGraveyard,
		Carry:    []uuid.UUID{top},
	}.Apply(ctx)
}

var bisonWhistleTakeCreature = MayYes("avatar-b30/bison-whistle-take-creature", func(ctx *Context, r game.ConfirmAnswer) error {
	if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
		return nil
	}
	ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Bison Whistle — reveal the top card")
	return ctx.Game.BounceToHandForEffect(carryAt(r, 0))
})

var bisonWhistleToGraveyard = MayYes("avatar-b30/bison-whistle-to-graveyard", putCarriedLibraryCardIntoGraveyard)

// putCarriedLibraryCardIntoGraveyard is "you may put it into your
// graveyard" about one looked-at or revealed library card, the answer's
// first carry: a plain move (not a mill), made only while the card is
// still in the library. Shared by Bison Whistle and Kellan, Daring
// Traveler.
func putCarriedLibraryCardIntoGraveyard(ctx *Context, r game.ConfirmAnswer) error {
	if !bisonWhistleStillInLibrary(ctx.Game, carryAt(r, 0)) {
		return nil
	}
	return ctx.Game.PutIntoGraveyardForEffect(carryAt(r, 0))
}

// bisonWhistleStillInLibrary is the CR 608.2b check on the looked-at
// card: one that has left the library in the meantime is a new object.
func bisonWhistleStillInLibrary(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneLibrary
}
