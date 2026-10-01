package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polluted Cistern // Dim Oubliette — Enchantment — Room (ADR 0103):
//
//	Polluted Cistern {1}{B}: "Whenever one or more cards are put into
//	 your graveyard from your library, each opponent loses 1 life for
//	 each card type among those cards."
//	Dim Oubliette {4}{B}: "When you unlock this door, mill three cards,
//	 then return a creature card from your graveyard to the battlefield."
//
// Polluted Cistern fires once per batch of library-to-graveyard moves
// (a three-card mill is one trigger) and reads the types of every card
// of that batch when it resolves. Dim Oubliette mills in the mill's
// continuation and then asks for a creature card from the whole
// graveyard, the milled cards included.
func init() {
	const cistern = "Polluted Cistern — each opponent loses 1 life for each card type among those cards"
	Register(Room(RoomSpec{
		OracleID:     "cfcf34ba-b412-447f-b200-d841371c0a1c",
		Name:         "Polluted Cistern // Dim Oubliette",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventMill, game.EventZoneMove},
				roomsBCardPutIntoYourGraveyardFromLibrary, cistern, roomsBEachOpponentLosesLifePerCardTypeMilled)),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Dim Oubliette — mill three cards, then return a creature card from your graveyard to the battlefield",
				roomsBMillThenReturnAChosenCreatureCard(3)),
		}},
	}))
}
