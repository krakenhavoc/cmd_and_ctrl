package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunforger — Artifact — Equipment {3}:
//
//	"Equipped creature gets +4/+0.
//	 {R}{W}, Unattach this Equipment: Search your library for a red or white
//	 instant card with mana value 4 or less and cast that card without paying
//	 its mana cost. Then shuffle.
//	 Equip {3}"
//
// The pump and Equip {3} are the usual ones. "Unattach this Equipment" is a
// cost (game.AbilityCost.UnattachThis, CR 701.3d), paid at announce with the
// mana, so an Equipment attached to nothing can't activate the ability
// (CR 118.3) and the creature loses the +4/+0 before the ability resolves; the
// Equipment stays on the battlefield. The search takes a red or white instant
// card with mana value 4 or less (judged on the card in the library, X counted
// as 0) and may find nothing (CR 701.23b); the found card is left on top of
// the library and offered as a free cast during the resolution (CR 608.2g, the
// Emergent Ultimatum posture: the resolution_cast prompt, timing ignored, the
// spell's targets chosen as it is cast). The shuffle is the sentence after the
// cast, so it runs once the prompt is answered; a card that is declined or
// can't be cast is shuffled back in with the rest.
//
// No simplification beyond that posture.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1421070-a3cc-4af3-bf91-0f97372f4161",
		Name:         "Sunforger",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 0)},
		Activated: []effects.ActivatedAbility{
			{
				Label:     "{R}{W}, Unattach this Equipment: Search your library for a red or white instant card with mana value 4 or less and cast that card without paying its mana cost. Then shuffle.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.Plus(effects.ManaCost("{R}{W}"), effects.UnattachThisEquipment()),
				Condition: effects.SourceAttachedToACreature(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
						Player: ctx.Controller(),
						Source: ctx.Source(),
						Pred: func(c game.Card) bool {
							return c.HasCardType("instant") && c.ManaValue() <= 4 &&
								(c.HasColorWord("R") || c.HasColorWord("W"))
						},
						Dest:     game.ZoneLibrary,
						Limit:    1,
						ToTop:    true,
						Reason:   "Sunforger — search for a red or white instant card with mana value 4 or less",
						ThenTail: ctx.ItemTail(sunforgerFound),
					})
				},
			},
			effects.EquipAbility("{3}"),
		},
	})
}

// sunforgerFound offers the free cast of the card the search found, then
// shuffles. An on-disk key.
var sunforgerFound = effects.CardTail("sunforger/cast-found", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	controller := ctx.Controller()
	if len(o.IDs) == 0 {
		return g.ShuffleLibraryForEffect(controller)
	}
	perm := effects.FreeCastSpell{}.On(game.CastPermission{
		Player:     controller,
		Zone:       game.ZoneLibrary,
		Duration:   g.UntilEndOfTurnDuration(),
		Cost:       "{0}",
		AltCostKey: game.AltCostKeyGrantedCost,
		Timing:     game.TimingFlash,
		CastOnly:   true,
		Source:     ctx.Source(),
		SourceName: "Sunforger",
		Label:      "Sunforger — cast it without paying its mana cost",
	})
	if !g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       ctx.Source(),
		Reason:       "Sunforger",
		Cards:        o.IDs,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in your library",
		Then:         sunforgerShuffle,
		Carry:        []uuid.UUID{controller},
	}) {
		return g.ShuffleLibraryForEffect(controller)
	}
	return nil
})

// sunforgerShuffle is "Then shuffle", run once the cast offer is answered.
// An on-disk key.
var sunforgerShuffle = game.RegisterChooseCardsThen("choose-cards/sunforger-shuffle", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return g.ShuffleLibraryForEffect(r.Carry[0])
})
