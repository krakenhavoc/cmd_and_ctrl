package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kujaGame seats Kuja (imported with both faces) under seat 0.
func kujaGame(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	row := effectstest.TransformRow(cardskr.KujaOracleID, "Kuja, Genome Sorcerer", "Legendary Creature — Human Mutant Wizard", "{2}{B}{R}",
		"Trance Kuja, Fate Defied", "Legendary Creature — Avatar Wizard", "4", "6", []string{"B", "R"})
	return g, effectstest.ImportToBattlefield(t, g, row, g.Seats[0])
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
		_, err = g.CreateTokensForEffect(controller, effects.NoncreatureCastWizardToken(), n, game.TokenEntryOptions{})
	})
	if err != nil {
		t.Fatalf("create Wizard tokens: %v", err)
	}
}

func TestKujaEndStepMakesATappedWizardAndStaysFrontBelowFourWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	toks := wizardTokens(g, me.ID)
	if len(toks) != 1 {
		t.Fatalf("Wizard tokens = %d, want 1", len(toks))
	}
	if !toks[0].Tapped {
		t.Error("the Wizard token must enter tapped")
	}
	if got := effectstest.BattlefieldCardFor(g, kuja).ActiveFace; got != 0 {
		t.Errorf("Kuja turned over with only two Wizards (ActiveFace = %d)", got)
	}
}

func TestKujaTransformsOnceHeControlsFourWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	me := g.Seats[0]
	// Kuja + two tokens already; the end-step token is the fourth Wizard.
	addWizardTokens(t, g, me.ID, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(wizardTokens(g, me.ID)); got != 3 {
		t.Fatalf("Wizard tokens = %d, want 3", got)
	}
	c := effectstest.BattlefieldCardFor(g, kuja)
	if c.ActiveFace != 1 || c.Name != "Trance Kuja, Fate Defied" {
		t.Errorf("Kuja = face %d %q, want Trance Kuja", c.ActiveFace, c.Name)
	}
}

func TestKujaDoesNotCountAnotherPlayersWizards(t *testing.T) {
	g, kuja := kujaGame(t)
	addWizardTokens(t, g, g.Seats[1].ID, 3)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardFor(g, kuja).ActiveFace; got != 0 {
		t.Errorf("Kuja counted an opponent's Wizards (ActiveFace = %d)", got)
	}
}

// pingWithWizard casts a noncreature spell with one Wizard token under
// seat 0 and returns the damage the first opponent took.
func pingWithWizard(t *testing.T, g *game.Game) int {
	t.Helper()
	opp := g.Seats[1]
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	oppWizard := effectstest.B27Push(g, opp.ID, "Their Wizard", "Creature — Human Wizard", "", "{1}{U}", 1, 1, "U")
	bear := effectstest.B27Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
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
