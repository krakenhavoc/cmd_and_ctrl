package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_payments_cards_test.go — the fourteen cards of ADR 0150
// PR 2 (#1903): any-player offers paid with life, damage, a sacrifice,
// a discard, or nothing at all.

const (
	AppAetherRift        = "2577ef8e-d85e-47af-ac34-629f31992751"
	AppDashHopes         = "6a4ace6b-ef3a-4d11-89d7-958f98dcc445"
	AppTemporalExtortion = "13da5255-2fc5-43a4-a44c-13a5d9ff77c3"
	AppBookBurning       = "644a8946-4890-4c2b-978f-02eedc95b5ad"
	AppBreakingPoint     = "a648b9ed-810a-41db-af40-3b5b4658db90"
	AppBrowbeat          = "11cd0c42-4148-4d82-9172-b1a1967fe319"
	AppBarbarianBully    = "5a88f616-986d-4e97-9457-26c783489123"
	AppArgothianWurm     = "f8aa1893-60e5-439e-9316-872a9f6b6c94"
	AppShivanWumpus      = "4b78222f-6973-42a3-987e-eb8cc8fd6c76"
	AppBrainGorgers      = "3b9369ae-8827-45d1-af85-c777cde8e1bc"
	AppProwlingPangolin  = "92042cad-09f5-464f-85ad-baad27e6d1eb"
	AppPhantasmagorian   = "248e0dfa-0bcb-402c-badf-ca59f02792f7"
	AppDistantMemories   = "12c78cc4-7fbe-4639-bef7-bf959a0d6d68"
)

// appOpenLink is the open link of an any-player offer, or nil.
func AppOpenLink(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.AnyPlayerOffer != nil {
			return c
		}
	}
	return nil
}

// appAnswerWith answers every link of the open any-player offer in
// turn: `pay` says whether the link's chooser pays and with which cards.
// It returns the seats asked, in order, and settles the stack after.
func AppAnswerWith(t *testing.T, g *game.Game, pay func(*game.PendingChoice) (bool, []uuid.UUID)) []uuid.UUID {
	t.Helper()
	var asked []uuid.UUID
	for i := 0; i < 8; i++ {
		link := AppOpenLink(g)
		if link == nil {
			PassPriorityAroundTable(t, g)
			return asked
		}
		asked = append(asked, link.Chooser)
		yes, cards := pay(link)
		if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, yes, cards); err != nil {
			t.Fatalf("answer link %d: %v", i, err)
		}
	}
	t.Fatal("the offer never ended")
	return nil
}

// appOnly pays as `who` (with `cards`, for a sacrifice or a discard)
// and declines as everyone else. uuid.Nil is nobody.
func AppOnly(who uuid.UUID, cards ...uuid.UUID) func(*game.PendingChoice) (bool, []uuid.UUID) {
	return func(c *game.PendingChoice) (bool, []uuid.UUID) {
		if who != uuid.Nil && c.Chooser == who {
			return true, cards
		}
		return false, nil
	}
}

// appToUpkeepOf walks the turn to `seat`'s upkeep.
func AppToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep; i++ {
		if i > 40 {
			t.Fatalf("never reached seat %d's upkeep (at seat %d %v)", seat, g.Turn.ActiveSeat, g.Turn.Step)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func AppInZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// appCastAsSeat puts a card in `p`'s hand and casts it at instant speed
// (the caller is already in a main phase).
func AppCastAsSeat(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}
