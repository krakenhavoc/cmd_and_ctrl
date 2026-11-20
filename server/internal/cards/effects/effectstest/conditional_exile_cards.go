package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_exile_cards_test.go — #3192, ADR 0066's second
// 2026-10-13 amendment. One test per card that exiles with a
// conditional or source-bound permission. The engine model is pinned in
// game/cast_permission_condition_test.go; these pin that each card says
// the right thing with it.

const (
	BorosStrikeCaptainOracle    = "30b20932-0d9a-447f-b934-1daa8c44a678"
	GoblinResearcherOracle      = "c8762dfa-a027-4e1c-a2fa-bfdab1de102a"
	RobberOfTheRichOracle       = "4df0d381-fc05-4d30-9307-48963aa2ecfd"
	PickUpThePaceOracle         = "873e739c-edf4-46a7-9024-5caaf87f507e"
	NerivCracklingVanguardOracl = "c8f827d7-4740-43c8-b494-c9e0479e0bf7"
	PossibilityTechnicianOracle = "ebbf84f7-d142-4de5-bfc1-ea86f082d733"
	LightningSecuritySergeantOr = "e1626880-0fa3-441a-9978-b293fc87c418"
	VictorManchaRunawayOracle   = "606a83ae-c774-4abc-ba30-c113cbd1c05b"
	NivixAerieOracle            = "9c482f1d-08b4-4882-918c-448a556d3fbe"
)

// nextTurnOfMine rolls the table round to the controller's next turn.
func NextTurnOfMine(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceToMainOf(t, g, 1)
	AdvanceToMainOf(t, g, 0)
}

// topLibraryCard puts a recognisable card on top of the library and
// returns its ID.
func TopLibraryCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}
