package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_hobbit_recruit_cards_test.go — the recruit cards of The Hobbit
// (#4475). Recruit itself (CR 701.70a) is pinned in recruit_test.go;
// these pin each card's trigger or clause reaching it.

const (
	LakeTownLookoutOracle        = "cf765efe-884c-48e2-9edb-9d45cf2756dd"
	LongLakeNuisanceOracle       = "a833fdf1-db0c-4846-8452-d3b2059c2355"
	PatientInstructorOracle      = "bddd7e99-ec74-4ca6-9137-155b85695a95"
	EsgarothGarrisonOracle       = "84597236-2cfc-4654-b126-a8ee252f4d14"
	GreatGildedBoatOracle        = "d4607026-80b2-4fa6-99d5-694cd3b43868"
	TheQueenOfDaleOracle         = "a2ec1dd0-86c7-423d-b562-ed95b79bf8f7"
	BardsCompanyOracle           = "5c5bfbb2-0e63-4e43-b441-c4878983288f"
	SoundTheTrumpetsOracle       = "dbcf2a79-b455-43ac-9cf0-9e63708800f4"
	CelebrateTheMountainKingOrac = "d51136fa-3c13-48a5-83fd-51fe00010a4b"
	TheMountainKingsReturnOracle = "32ad5b3e-92c0-45be-b2e4-6f1794552f36"
)

// hobEnter puts c into its owner's hand and moves it onto the
// battlefield, so its entry triggers fire as they would from a cast.
func HobEnter(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	p := PlayerByIDForTest(g, owner)
	p.Hand.PushTop(c)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, c.InstanceID); err != nil {
		t.Fatalf("enter %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// hobAnswerTargets answers every open pick_target prompt of chooser's
// with the first offered card in want, or with nothing ("up to one").
func HobAnswerTargets(t *testing.T, g *game.Game, chooser uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		var refs []game.TargetRef
		for _, offered := range ch.PickTargetCards {
			for _, w := range want {
				if offered == w && len(refs) == 0 {
					refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: w})
				}
			}
		}
		if err := g.ResolvePickTargets(ch.ID, chooser, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
}

// hobAnswerOrder answers chooser's open trigger-order prompt, if any, in
// the order it offered, without passing priority.
func HobAnswerOrder(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			if err := g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			return
		}
	}
}
