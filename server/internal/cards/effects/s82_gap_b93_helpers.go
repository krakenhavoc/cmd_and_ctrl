package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b93_helpers.go — card-side helpers of set GAP, batch 93
// (Doctor Who enchantments, S82, #5948).

// gapDoctorsCompanionCards are the oracle IDs of the cards that print
// "Doctor's companion" (CR 702.124m). The keyword is a deck-construction rule
// the engine keeps nowhere on a card (internal/deck reads it off the oracle
// text), so a search that asks for "a card with doctor's companion" reads
// this set, built from the Scryfall dump.
var gapDoctorsCompanionCards = map[string]bool{
	"50d0d870-fe3d-44a2-ad6e-93307a1fb468": true, // Ace, Fearless Rebel
	"af566158-23b2-4f91-8d59-a58342a9f576": true, // Adric, Mathematical Genius
	"aa7d3e99-fc14-4f62-bf00-c0428c11e4f2": true, // Amy Pond
	"0e61a062-f13a-4958-91de-909650c662a8": true, // Barbara Wright
	"0ebe67ed-8a71-4bbc-8e58-cebbd7b2d83f": true, // Bill Potts
	"542eae01-058d-4959-92a8-49d6b1e33ecd": true, // Clara Oswald
	"b84ab039-901e-4635-b57e-3f2dac239f38": true, // Dan Lewis
	"0d6e76aa-2966-48b7-be4e-2d3e2cc94fc1": true, // Donna Noble
	"7986908d-8af2-465a-8597-26e92ca2f1c9": true, // Graham O'Brien
	"32c9ce8c-a19d-49e4-bc6d-c7182fa00b51": true, // Ian Chesterton
	"c9f074cb-6dcb-4c25-8332-9b15e4881d33": true, // Jamie McCrimmon
	"65ea3638-1766-4096-88a8-662a0d6faf0c": true, // Jo Grant
	"413bc18b-1579-41b5-a778-3a29def52bb0": true, // K-9, Mark I
	"4ae972ea-d5d8-4064-9d91-5b2040d7dcbf": true, // Leela, Sevateem Warrior
	"5b7ccdd2-40aa-4619-a147-0792dc02b2e5": true, // Martha Jones
	"81c515b7-9174-4b34-ad07-4144e3dbaaac": true, // Nardole, Resourceful Cyborg
	"07802bd6-29ca-40c5-beba-c2baf7a92555": true, // Nyssa of Traken
	"2fadf4aa-1b7e-4e72-ad20-809a59ddccf8": true, // Peri Brown
	"3346a585-d35a-4f8b-b9ab-7e453ebd15f4": true, // Romana II
	"2f56de72-80d7-48c6-9b57-84fbd252c699": true, // Rose Tyler
	"6c06e2d5-1473-4ce6-bcf6-10dfbbe5f801": true, // Ryan Sinclair
	"da268892-b190-4eee-9a2b-b0b9bb763033": true, // Sarah Jane Smith
	"0f21f3b1-aee9-4272-ad02-067aa5f6a7c8": true, // Susan Foreman
	"ac46f435-5b25-4b46-b89f-175bca81b9ad": true, // Tegan Jovanka
	"cdf5c31c-c324-471f-b6d7-06a06e91daa3": true, // Vislor Turlough
	"93637599-d9b4-4f7e-a5c0-5a92656e941e": true, // Yasmin Khan
}

// gapIsDoctorCard is "a Doctor card": a card with the Doctor subtype.
func GapIsDoctorCard(c game.Card) bool { return c.HasSubtypeWord("Doctor") }

// gapHasDoctorsCompanion is "a card with doctor's companion".
func GapHasDoctorsCompanion(c game.Card) bool { return gapDoctorsCompanionCards[c.OracleID] }

// gapDoctorSpell is a Doctor spell, the cost-reduction predicate of the
// Human token of The Eleventh Hour.
func gapDoctorSpell() CostPredicate {
	return func(q game.CostQuery) bool { return GapIsDoctorCard(q.Card) }
}

// gapDoctorHumanToken is the 1/1 white Human creature token of The Eleventh
// Hour: "Doctor spells you cast cost {1} less to cast."
func GapDoctorHumanToken() game.Card { return TokenFromCatalog(printedGapDoctorHumanToken) }

// printedGapDoctorHumanToken is that Human as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func printedGapDoctorHumanToken() TokenTemplate {
	return TokenTemplate{
		Slug: "human-doctor-spells-cost-less",
		Card: game.Card{
			Name:      "Human",
			TypeLine:  "Token Creature — Human",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Doctor spells you cast cost {1} less to cast.", YourSpell(), gapDoctorSpell()),
		},
		Text: "Doctor spells you cast cost {1} less to cast.",
	}
}
