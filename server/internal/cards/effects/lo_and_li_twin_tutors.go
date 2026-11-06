package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lo and Li, Twin Tutors — Legendary Creature — Human Advisor {4}{B},
// 2/2 (Avatar: The Last Airbender, #4459):
//
//	"When Lo and Li enter, search your library for a Lesson or Noble
//	 card, reveal it, put it into your hand, then shuffle.
//	 Noble creatures you control and Lesson spells you control have
//	 lifelink."
//
// The search names a quality, so the searcher may find nothing (CR
// 701.23b), and the library is shuffled either way. The static is two
// layer-6 grants: a lord over your Noble creatures, and Pestilent
// Spirit's grant over SPELLS (SpellsYouControlHave) for your Lesson
// spells while they are on the stack. A spell deals its damage from the
// stack, so a Lesson that deals damage gains you that much life; a
// Lesson that makes another source deal damage does not (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad643992-eb9f-4a4a-9b74-a5aee2337f30",
		Name:         "Lo and Li, Twin Tutors",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lo and Li, Twin Tutors — search for a Lesson or Noble card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Lesson") || c.HasSubtypeWord("Noble") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Lo and Li, Twin Tutors — a Lesson or Noble card, revealed, to hand",
					Source:    item.SourceCardID,
				}.Apply(NewContext(g, item))
			}),
		},
		Static: []game.StaticAbility{
			Labelled(TribalKeywordGrant(TribeFilter{Tribes: []string{"Noble"}, YoursOnly: true}, "lifelink"),
				"Noble creatures you control have lifelink."),
			Labelled(SpellsYouControlHave(Subtype("Lesson"), "lifelink"),
				"Lesson spells you control have lifelink."),
		},
	})
}
