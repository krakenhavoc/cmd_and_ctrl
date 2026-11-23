package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Theros — Battle — Siege, defense 4, for {2}{W} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, search your library for an Aura, God, or
//	 Demigod card, reveal it, put it into your hand, then shuffle."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the tutor and the defeated
// trigger that hands over Ephara, Ever-Sheltering. The search names a
// quality, so the searcher may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     InvasionOfTherosOracleID,
		Name:         "Invasion of Theros",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Theros — search for an Aura, God, or Demigod card",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.IsAura() || c.HasSubtypeWord("God") || c.HasSubtypeWord("Demigod")
						},
						Dest:    game.ZoneHand,
						Limit:   1,
						Reveal:  true,
						Shuffle: true,
						Source:  item.SourceCardID,
						Reason:  "Invasion of Theros — choose an Aura, God, or Demigod card to put into your hand",
					}.Apply(NewContext(g, item))
				}),
			DefeatedTrigger("Invasion of Theros — defeated: exile it, then cast Ephara, Ever-Sheltering", SiegeDefeated()),
		},
	})
}

// invasionOfTherosOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const InvasionOfTherosOracleID = "ba5f7f37-352a-4b9c-b648-0ba5c8d842fc"
