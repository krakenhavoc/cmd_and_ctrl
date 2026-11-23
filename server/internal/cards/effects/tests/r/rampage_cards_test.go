package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rampage_cards_test.go — the catalog half of #5769: the Alliances and
// Legends rampage creatures. Rampage N (CR 702.23) is the engine's
// keyword trigger (game/rampage.go, whose own tests cover the rule); each
// test here plays the real catalog entry.

const (
	rampageWarMakersOracle = "1bda7f60-f986-455b-9d56-bbd494e27c17"
	rampageGorillaOracle   = "6373c66d-edfc-44cb-af2a-2273e0ba53e0"
	rampageAerathiOracle   = "5cb495a2-c683-4066-b6ee-d0b7d8843cb9"
	rampageCrawOracle      = "152b621e-4c18-438c-9999-77e2685fda79"
	rampageFrostOracle     = "56ed6af3-8f85-4907-9364-eca71f21bb11"
	rampageHundingOracle   = "dd4f3a78-3167-42d0-8334-5ac36750a253"
	rampageMarhaultOracle  = "c202f124-4283-48f6-aef1-470ad9ded22b"
	rampageWolverineOracle = "d67d9d82-2786-48fc-8107-11d12c5813d0"
	rampageChromiumOracle  = "7c91ae5d-0320-46a7-98d2-df0918202478"
)

// rampageCards is every card of #5769: its name, printed size, N and the
// other keywords it prints.
var rampageCards = []struct {
	oracle, name, typeLine string
	power, toughness, n    int
	keywords               []string
}{
	{rampageWarMakersOracle, "Balduvian War-Makers", "Creature — Human Barbarian", 3, 3, 1, []string{"haste"}},
	{rampageGorillaOracle, "Gorilla Berserkers", "Creature — Ape Berserker", 2, 3, 2, []string{"trample"}},
	{rampageAerathiOracle, "Aerathi Berserker", "Creature — Human Berserker", 2, 4, 3, nil},
	{rampageCrawOracle, "Craw Giant", "Creature — Giant", 6, 4, 2, []string{"trample"}},
	{rampageFrostOracle, "Frost Giant", "Creature — Giant", 4, 4, 2, nil},
	{rampageHundingOracle, "Hunding Gjornersen", "Legendary Creature — Human Warrior", 5, 4, 1, nil},
	{rampageMarhaultOracle, "Marhault Elsdragon", "Legendary Creature — Elf Warrior", 4, 6, 1, nil},
	{rampageWolverineOracle, "Wolverine Pack", "Creature — Wolverine", 2, 4, 2, nil},
	{rampageChromiumOracle, "Chromium", "Legendary Creature — Elder Dragon", 7, 7, 2, []string{"flying"}},
}

func TestRampageCardsAreRegisteredFullWithTheirKeywords(t *testing.T) {
	for _, tc := range rampageCards {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.name)
			continue
		}
		if spec.Name != tc.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", tc.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		g := effectstest.NewCatalogGame(t)
		id := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, effectstest.ApaCreature(tc.name, tc.oracle, tc.power, tc.toughness))
		if got := game.RampageAmounts(effectstest.ApaLive(g, id)); len(got) != 1 || got[0] != tc.n {
			t.Errorf("%s has rampage amounts %v, want [%d]", tc.name, got, tc.n)
		}
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// The deck importer reads the number off the printed line, so an
// imported copy has the same one instance of rampage as the catalog's.
// Gorilla Berserkers' "Trample; rampage 2" is read by the line scan too
// (a semicolon separates keywords as a comma does, #5791) and declared
// by the catalog, and the merge keeps one instance, not two.
func TestRampageCardsImportWithOneInstance(t *testing.T) {
	for _, tc := range rampageCards {
		text := effectstest.SpellText(t, tc.oracle)
		if text == "" {
			t.Fatalf("%s has no oracle fixture", tc.name)
		}
		c := deck.ToGameCard(cards.Card{
			ID: uuid.New(), OracleID: uuid.MustParse(tc.oracle), Name: tc.name, Layout: "normal",
			TypeLine: tc.typeLine, Keywords: []string{"Rampage"}, OracleText: text,
		}, false)
		g := effectstest.NewCatalogGame(t)
		id := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, c)
		if got := game.RampageAmounts(effectstest.ApaLive(g, id)); len(got) != 1 || got[0] != tc.n {
			t.Errorf("imported %s has rampage amounts %v, want [%d]", tc.name, got, tc.n)
		}
	}
}

// Each card, blocked by one creature and by three: one is +0/+0, three
// are +2N/+2N (CR 702.23a).
func TestRampageCardsPumpForEachBlockerBeyondTheFirst(t *testing.T) {
	for _, tc := range rampageCards {
		for _, blockers := range []int{1, 3} {
			if tc.oracle == rampageGorillaOracle && blockers < 3 {
				continue // it can't be blocked except by three or more
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, effectstest.ApaCreature(tc.name, tc.oracle, tc.power, tc.toughness))
			effectstest.RampageBlock(t, g, id, effectstest.RampageBlockers(g, blockers)...)
			bonus := tc.n * (blockers - 1)
			c := effectstest.ApaLive(g, id)
			if c.CurrentPower() != tc.power+bonus || c.CurrentToughness() != tc.toughness+bonus {
				t.Errorf("%s blocked by %d is %d/%d, want %d/%d", tc.name, blockers,
					c.CurrentPower(), c.CurrentToughness(), tc.power+bonus, tc.toughness+bonus)
			}
		}
	}
}

// Gorilla Berserkers can't be blocked except by three or more creatures:
// two blockers are refused, and three give it +4/+4.
func TestGorillaBerserkersNeedsThreeBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, effectstest.ApaCreature("Gorilla Berserkers", rampageGorillaOracle, 2, 3))
	bs := effectstest.RampageBlockers(g, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: bs[0], Attacker: id}, {Blocker: bs[1], Attacker: id}})
	if !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("two blockers on Gorilla Berserkers = %v, want an illegal block", err)
	}
}

// Balduvian War-Makers has haste: it attacks the turn it arrives.
func TestBalduvianWarMakersAttacksTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.ApaPush(g, g.Seats[0].ID, g.Seats[0].ID, effectstest.ApaCreature("Balduvian War-Makers", rampageWarMakersOracle, 3, 3))
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = true })
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
		t.Fatalf("a hasty Balduvian War-Makers could not attack the turn it arrived: %v", err)
	}
}

// Chromium is sacrificed at its controller's upkeep unless they pay
// {W}{U}{B}, the Legends Elder Dragons' shape.
func TestChromiumIsSacrificedUnlessItsUpkeepIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		dragon := effectstest.LegB1Push(g, me.ID, "Chromium", effectstest.LegB1ElderDragonType, rampageChromiumOracle, "", 7, 7)
		effectstest.WalkToUpkeepOfSeat0(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.DcMana(t, g, me.ID, "{W}{U}{B}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if got := g.Battlefield.Contains(dragon); got != pay {
			t.Errorf("paid %v: Chromium on the battlefield = %v", pay, got)
		}
	}
}
