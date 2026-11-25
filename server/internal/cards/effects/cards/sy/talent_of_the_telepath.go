package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talent of the Telepath — Sorcery {2}{U}{U} (Magic Origins, #5549):
//
//	"Target opponent reveals the top seven cards of their library. You may
//	 cast an instant or sorcery spell from among them without paying its
//	 mana cost. Then that player puts the rest into their graveyard.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, you may cast up to two instant and/or sorcery spells
//	 from among the revealed cards instead of one."
//
// The seven cards are revealed to the table, then the caster is offered the
// instants and sorceries among them through the resolution-cast prompt
// (CR 608.2g), cast from the opponent's library; with spell mastery, read as
// the spell resolves so it does not count itself, a cast is followed by a
// second offer over what is left. Once the casts are done every revealed
// card still in the library, the declined ones included, goes into its
// owner's graveyard. A target that is illegal on resolution does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3a26df4-a9d5-48f5-942f-76862bdcd146",
		Name:         "Talent of the Telepath",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var opp uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					opp = t.ID
				}
			}
			if opp == uuid.Nil {
				return nil
			}
			g := ctx.Game
			revealed := g.RevealTopOfLibraryForEffect(opp, ctx.Source(), 7,
				"Talent of the Telepath — revealed from the top of the library")
			if len(revealed) == 0 {
				return nil
			}
			then := talentOfTheTelepathOnly
			if effects.SpellMastery(g, ctx.Controller()) {
				then = talentOfTheTelepathFirstOfTwo
			}
			return talentOfTheTelepathOffer(g, ctx.Controller(), ctx.Source(), revealed, revealed, then)
		},
	})
}

const talentOfTheTelepathName = "Talent of the Telepath"

// talentOfTheTelepathOffer asks the caster to cast one instant or sorcery
// among `from` (the revealed cards still in a library) without paying its
// mana cost, `all` being every card revealed. With nothing to cast, or no
// resolution open to ask in, the rest go to the graveyard at once.
func talentOfTheTelepathOffer(g *game.Game, caster, source uuid.UUID, all, from []uuid.UUID, then game.ChooseCardsThen) error {
	perm := effects.FreeInstantOrSorcery().On(g.FreeCastFromLibraryPermission(caster, source, talentOfTheTelepathName))
	perm.LapseOnPass = game.LapseStaysInHand
	eligible := effects.FreeCastCandidates(g, &perm, effects.CardsStillInALibrary(g, from))
	if len(eligible) == 0 {
		return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, all))
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      caster,
		Source:       source,
		Reason:       talentOfTheTelepathName,
		Cards:        eligible,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put the rest into their graveyard",
		Then:         then,
		Carry:        all,
	}) {
		return nil
	}
	return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, all))
}

// talentOfTheTelepathOnly ends the card: the rest go into the graveyard.
// On-disk key, never renamed.
var talentOfTheTelepathOnly = game.RegisterChooseCardsThen("resolution-cast/talent-of-the-telepath-rest",
	func(g *game.Game, r game.CardsChosen) error {
		return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, r.Carry))
	})

// talentOfTheTelepathFirstOfTwo is the first offer under spell mastery:
// after a cast the caster is offered a second spell among the cards left.
// On-disk key, never renamed.
var talentOfTheTelepathFirstOfTwo = game.RegisterChooseCardsThen("resolution-cast/talent-of-the-telepath-first",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 {
			return effects.RestIntoGraveyard(g, effects.CardsStillInALibrary(g, r.Carry))
		}
		return talentOfTheTelepathOffer(g, r.Chooser, r.Source, r.Carry, r.Carry, talentOfTheTelepathOnly)
	})
