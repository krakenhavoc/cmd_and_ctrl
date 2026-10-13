package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_test.go — slice kuja-a of the Kuja 9's ultimate deck request
// (#3213): one test group per card.

const (
	fandanielOracle    = "3909ff79-2103-407c-a812-a5a5a996fce2"
	goliathOracle      = "9207facd-fbf9-4faa-ac44-6c0bcac81551"
	kefkaOracle        = "fe5690d3-547b-4ce9-8e94-77fdc0e9c5c6"
	krarkThumblessOrcl = "18e421c0-2879-4ad8-b027-397f168dfbc2"
	nehebEternalOracle = "b0d3fd57-1420-4dd1-8410-9d659655287f"
	queenBrahneOracle  = "0527b7b5-a55d-4720-af1f-17fc1178f4a6"
	cerberusOracle     = "b6b22bac-853a-45a8-a74d-9904ec2b34fd"
	tersaOracle        = "9918eb7b-054e-4b1c-afc7-1163ab642d50"
	deadpoolOracle     = "4b96c22a-0d5b-44fd-b326-5f3ffcc3917b"
)

// kujaGame seats Kuja (imported with both faces) under seat 0.
func kujaGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	row := transformRow(kujaOracleID, "Kuja, Genome Sorcerer", "Legendary Creature — Human Mutant Wizard", "{2}{B}{R}",
		"Trance Kuja, Fate Defied", "Legendary Creature — Avatar Wizard", "4", "6", []string{"B", "R"})
	return g, importToBattlefield(t, g, row, g.Seats[0])
}

func wizardTokens(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == controller && c.Name == "Wizard" {
			out = append(out, c)
		}
	}
	return out
}

func addWizardTokens(t *testing.T, g *game.Game, controller uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		_, err = g.CreateTokensForEffect(controller, NoncreatureCastWizardToken(), n, game.TokenEntryOptions{})
	})
	if err != nil {
		t.Fatalf("create Wizard tokens: %v", err)
	}
}

func TestKujaEndStepMakesATappedWizardAndStaysFrontBelowFourWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	me := g.Seats[0]
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	toks := wizardTokens(g, me.ID)
	if len(toks) != 1 {
		t.Fatalf("Wizard tokens = %d, want 1", len(toks))
	}
	if !toks[0].Tapped {
		t.Error("the Wizard token must enter tapped")
	}
	if got := battlefieldCardFor(g, kuja).ActiveFace; got != 0 {
		t.Errorf("Kuja turned over with only two Wizards (ActiveFace = %d)", got)
	}
}

func TestKujaTransformsOnceHeControlsFourWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	me := g.Seats[0]
	// Kuja + two tokens already; the end-step token is the fourth Wizard.
	addWizardTokens(t, g, me.ID, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := len(wizardTokens(g, me.ID)); got != 3 {
		t.Fatalf("Wizard tokens = %d, want 3", got)
	}
	c := battlefieldCardFor(g, kuja)
	if c.ActiveFace != 1 || c.Name != "Trance Kuja, Fate Defied" {
		t.Errorf("Kuja = face %d %q, want Trance Kuja", c.ActiveFace, c.Name)
	}
}

func TestKujaDoesNotCountAnotherPlayersWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	addWizardTokens(t, g, g.Seats[1].ID, 3)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardFor(g, kuja).ActiveFace; got != 0 {
		t.Errorf("Kuja counted an opponent's Wizards (ActiveFace = %d)", got)
	}
}

// pingWithWizard casts a noncreature spell with one Wizard token under
// seat 0 and returns the damage the first opponent took.
func pingWithWizard(t *testing.T, g *game.Game) int {
	t.Helper()
	opp := g.Seats[1]
	before := opp.Life
	castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	passPriorityAroundTable(t, g)
	return before - opp.Life
}

func TestTranceKujaDoublesWizardDamage(t *testing.T) {
	g, kuja := kujaGame(t)
	addWizardTokens(t, g, g.Seats[0].ID, 1)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(kuja) })
	if got := pingWithWizard(t, g); got != 2 {
		t.Errorf("a Wizard's 1 damage under Trance Kuja = %d, want 2", got)
	}
}

func TestFrontFaceKujaDoesNotDoubleWizardDamage(t *testing.T) {
	g, _ := kujaGame(t)
	addWizardTokens(t, g, g.Seats[0].ID, 1)
	if got := pingWithWizard(t, g); got != 1 {
		t.Errorf("front-face Kuja: ping = %d, want 1", got)
	}
}

// Flare Star is about Wizards YOU control, and only Wizards.
func TestTranceKujaIgnoresOpponentWizardsAndNonWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(kuja) })
	oppWizard := b27Push(g, opp.ID, "Their Wizard", "Creature — Human Wizard", "", "{1}{U}", 1, 1, "U")
	bear := b27Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	deal := func(src uuid.UUID) int {
		before := me.Life
		if src == bear {
			before = opp.Life
		}
		target := me.ID
		if src == bear {
			target = opp.ID
		}
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, target, 1) })
		if src == bear {
			return before - opp.Life
		}
		return before - me.Life
	}
	if got := deal(oppWizard); got != 1 {
		t.Errorf("an opponent's Wizard dealt %d, want 1 (undoubled)", got)
	}
	if got := deal(bear); got != 1 {
		t.Errorf("a non-Wizard of yours dealt %d, want 1 (undoubled)", got)
	}
}
