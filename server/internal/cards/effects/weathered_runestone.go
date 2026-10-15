package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weathered Runestone — {2} Artifact:
//
//	"Nonland permanent cards in graveyards and libraries can't enter the
//	 battlefield.
//	 Players can't cast spells from graveyards or libraries."
//
// Grafdigger's Cage with a wider first sentence. Sentence one is an
// entry restriction (#1985, entry_restriction.go): an artifact,
// creature, enchantment, planeswalker or battle card leaving a graveyard
// or a library for the battlefield stays where it was (CR 101.2). A
// LAND card is named out, so a fetched or reanimated land still enters,
// and a token is not a card. Sentence two is the cast gate, the Cage's
// own constructor.
func init() {
	Register(Spec{
		OracleID:     "4386ae4e-560a-4251-942c-0f84ba7448fa",
		Name:         "Weathered Runestone",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			PlayersCantCastFrom("Players can't cast spells from graveyards or libraries.",
				game.ZoneGraveyard, game.ZoneLibrary),
		},
		EntryRestrictions: []game.EntryRestriction{
			CardsInZonesCantEnter("Nonland permanent cards in graveyards and libraries can't enter the battlefield.",
				func(c game.Card) bool { return c.IsPermanent() && !c.IsLand() },
				game.ZoneGraveyard, game.ZoneLibrary),
		},
	})
}
