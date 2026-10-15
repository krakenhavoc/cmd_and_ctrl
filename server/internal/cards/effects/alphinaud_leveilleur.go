package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alphinaud Leveilleur — Legendary Creature — Elf Wizard {3}{U}, 2/4:
//
//	"Partner with Alisaie Leveilleur (When this creature enters, target
//	 player may put Alisaie Leveilleur into their hand from their
//	 library, then shuffle.)
//	 Vigilance
//	 Eukrasia — Whenever you cast your second spell each turn, draw a
//	 card."
//
// Partner with is the shared entry trigger; the second-spell trigger is
// Breeches' and Avatar Yangchen's condition (the cast path bumps the
// per-turn tally before it emits EventCast, so a total of exactly two
// is the second spell).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d16783a4-9980-451b-9af0-33866980bc8d",
		Name:            "Alphinaud Leveilleur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Alphinaud Leveilleur", "Alisaie Leveilleur"),
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Alphinaud Leveilleur — Eukrasia: draw a card", Do(DrawCards{N: 1})),
		},
	})
}
