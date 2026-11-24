package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Fourteenth Doctor — Legendary Creature — Time Lord Doctor {R/G}{W}{U},
// 3/4 (Secret Lair Drop):
//
//	"When you cast this spell, reveal the top fourteen cards of your library.
//	 Put all Doctor cards revealed this way into your graveyard and the rest
//	 on the bottom of your library in a random order.
//	 You may have The Fourteenth Doctor enter as a copy of a Doctor card in
//	 your graveyard that was put there from your library this turn. If you
//	 do, it gains haste until end of turn."
//
// The cast trigger goes on the stack above the spell (CR 603.3), so it has
// resolved by the time the spell does. It puts the Doctors into the
// graveyard (not a mill), which stamps each as put there from the library this
// turn (Card.PutInGraveyardFromLibraryThisTurn, the Kagha clause). The copy is
// Body Double's graveyard-copy entry replacement over those cards only; the
// haste is the copy's except clause asking for a haste record that ends at
// cleanup (game.ReplacementEvent.EntersWithHasteTillEOT), and it is given only
// when the copy was chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0afcf1eb-ac2c-4bb0-8823-9fe45968c2d9",
		Name:         "The Fourteenth Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("The Fourteenth Doctor — reveal the top fourteen cards; Doctors go to your graveyard, the rest to the bottom in a random order",
				effects.S82OldB021FourteenthDoctorReveal),
		},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("The Fourteenth Doctor", effects.S82OldB021DoctorCardsFromYourLibraryThisTurn, effects.S82OldB021GainsHasteUntilEndOfTurn),
		},
	})
}
