package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dwarvenConfluencerOracle = "d483f61b-6000-43ca-b5eb-8ab1c924ffd6"

func TestDwarvenConfluencerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(dwarvenConfluencerOracle)
	if !ok || spec.Name != "Dwarven Confluencer" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Dwarven Confluencer spec = %+v, registered %v", spec, ok)
	}
}

func dwarvenCountNamed(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

func TestDwarvenConfluencerDestroysTheLandAndItsControllerGetsTheToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Confluencer", "Creature — Dwarf", dwarvenConfluencerOracle, false)
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToMain(t, g)

	target := []game.TargetRef{{Kind: game.TargetCard, ID: land}}
	effectstest.B16Activate(t, g, me.ID, dwarf, 0, game.ActivateAbilityParams{Targets: target, Strict: true})
	if !effectstest.B16Tapped(t, g, dwarf) {
		t.Error("the Confluencer did not tap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Fatal("the land survived")
	}
	if n := dwarvenCountNamed(g, opp.ID, "Mana Confluence"); n != 1 {
		t.Errorf("the land's controller has %d Mana Confluence tokens, want 1", n)
	}
	if n := dwarvenCountNamed(g, me.ID, "Mana Confluence"); n != 0 {
		t.Errorf("the Confluencer's controller got %d tokens, want 0", n)
	}
}

func TestDwarvenConfluencerRefusesATokenLandAndTheTokenIsAPayLifeLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Confluencer", "Creature — Dwarf", dwarvenConfluencerOracle, false)
	effectstest.MakeTokens(t, g, opp.ID, effects.ManaConfluenceToken(), 1)
	token := effectstest.FindBattlefieldByName(g, "Mana Confluence")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: token}}}); err == nil {
		t.Fatal("targeted a token land")
	}
	if effectstest.B16Tapped(t, g, dwarf) {
		t.Error("a refused activation tapped the Confluencer")
	}
	tc, _ := effectstest.LookupBattlefield(g, token)
	if !tc.IsLand() || tc.IsCreature() {
		t.Errorf("token type line %q, want a land", tc.TypeLine)
	}
	life := opp.Life
	if err := g.ActivateManaAbility(opp.ID, token, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("tapping the token for mana: %v", err)
	}
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
}
