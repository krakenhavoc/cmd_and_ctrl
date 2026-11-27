package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanoth — Creature — Beast {3}{R}{R}, 3/3 (Duel Decks: Izzet vs. Golgari):
//
//	"At the beginning of your upkeep, you may look at the top card of your
//	 library. You may cast it without paying its mana cost if it's an instant
//	 or sorcery spell."
//
// Two "may"s: the look is the trigger's own (Optional); the cast is offered
// inside the resolution (CR 608.2g), once, over the top card when some face of
// it is an instant or sorcery spell (FreeInstantOrSorcery, judged on the face
// being cast). A card that is not offered, or is declined, stays on top: it was
// only looked at, so nothing moves and only its controller knows what it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b4fa1ba-9597-4db1-a77c-975ed839f57f",
		Name:         "Galvanoth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep(galvanothLabel, galvanothLook), "Galvanoth — look at the top card of your library?"),
		},
	})
}

const galvanothLabel = "Galvanoth — look at the top card of your library; you may cast it without paying its mana cost if it's an instant or sorcery spell"

func galvanothLook(g *game.Game, item *game.StackItem) error {
	you, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfLibraryForEffect(you, 1)
	if len(looked) == 0 {
		return nil
	}
	perm := effects.FreeInstantOrSorcery().On(g.FreeCastFromLibraryPermission(you, source, "Galvanoth"))
	perm.LapseOnPass = game.LapseStaysInHand
	if len(effects.FreeCastCandidates(g, &perm, looked)) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       "Galvanoth",
		Cards:        looked,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it on top",
		Carry:        looked,
	})
	return nil
}
