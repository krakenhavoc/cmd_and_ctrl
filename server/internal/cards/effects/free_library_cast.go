package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_library_cast.go — "look at the top N cards of your library. You
// may cast a spell from among them without paying its mana cost. Put
// the rest on the bottom of your library in a random order" (Aetherworks
// Marvel, #2110).
//
// It is free_hand_cast.go's pick over cards the player has only LOOKED
// at, and Sunbird's Invocation's finish for a reveal. The differences
// from Sunbird are the two the printed words make:
//
//   - "look at" (CR 701.20), not "reveal": only the looker is marked a
//     knower (LookAtTopOfLibraryForEffect), so the prompt, the pile and
//     the chosen card are on nobody else's wire. The bottoming clears
//     the knowers (PutOnBottomInRandomOrderForEffect), because the
//     looker knows the cards are somewhere in the bottom few, not where.
//   - no mana value cap.
//
// The chosen card keeps its place and, once the rest are bottomed, is the
// top card of the library, still known to its owner alone. The grant is
// game.FreeCastFromLibraryPermission: free, any timing (CR 608.2g),
// closing on the holder's next priority pass, and a card that is not cast
// goes to the bottom with the rest, which is what "put the rest" means
// when the answer was "none". The cast itself is the ordinary cast_spell
// taken right after the pick (ADR 0066, ADR 0091, ADR 0099 §4).

// lookAtTopAndOfferFreeCast looks at the top n cards of `controller`'s
// library and asks which nonland card of them, if any, to cast free. The
// rest go to the bottom in a random order. Nothing is queued when no
// card of them can be cast, but the rest still go to the bottom.
func lookAtTopAndOfferFreeCast(g *game.Game, controller, source uuid.UUID, name string, n int) error {
	looked := g.LookAtTopOfLibraryForEffect(controller, n)
	if len(looked) == 0 {
		return nil
	}
	var eligible []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			eligible = append(eligible, id)
		}
	}
	finish := freeCastFromLookedFinisher(controller, source, name, looked)
	if len(eligible) == 0 {
		return finish(g, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     controller,
		Source:      source,
		Question:    name + " — cast a spell from among the top " + strconv.Itoa(len(looked)) + " cards of your library without paying its mana cost?",
		Cards:       eligible,
		Min:         0,
		Max:         1,
		Zone:        game.ZoneLibrary,
		Destination: game.ChooseCastFreeFromLibrary,
		Then:        finish,
	})
	return nil
}

// freeCastFromLookedFinisher is what happens once the pick is known: the
// rest of the looked-at cards go to the bottom in a random order, and
// the chosen card, if any, keeps its place and gets its free cast there.
// Everything it captures is an ID or a scalar, so an undo across the
// prompt resolves against the restored game.
func freeCastFromLookedFinisher(controller, source uuid.UUID, name string, looked []uuid.UUID) func(g *game.Game, picked []uuid.UUID) error {
	return func(g *game.Game, picked []uuid.UUID) error {
		var chosen uuid.UUID
		if len(picked) > 0 {
			chosen = picked[0]
		}
		var rest []uuid.UUID
		for _, id := range cardsStillInALibrary(g, looked) {
			if id != chosen {
				rest = append(rest, id)
			}
		}
		if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil {
			return err
		}
		if chosen != uuid.Nil {
			g.FreeCastFromLibraryForEffect(controller, source, chosen, name)
		}
		return nil
	}
}
