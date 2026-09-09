package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// artifact_tokens_test.go — S21 sub-PR 4: Food / Clue / Blood carry
// their activated ability on the token template, because the
// catalog's hooks key on oracle ID and a token hasn't got one.

const (
	thrabenInspectorOracle = "caa02547-66e3-4e27-a2d3-5e94f3e7a069"
	ichorWellspringOracle  = "5b5ef43b-13fd-4461-8d2d-18be65e9a790"
)

func pushToken(g *game.Game, controller uuid.UUID, tmpl game.Card) uuid.UUID {
	tmpl.InstanceID = uuid.New()
	tmpl.Owner, tmpl.Controller = controller, controller
	g.Battlefield.PushTop(tmpl)
	return tmpl.InstanceID
}

func TestArtifactTokensCarryTheirAbilities(t *testing.T) {
	cases := []struct {
		name         string
		tmpl         game.Card
		tap, sacSelf bool
		mana         string
	}{
		{"Food", FoodToken(), true, true, "{2}"},
		{"Clue", ClueToken(), false, true, "{2}"},
		{"Blood", BloodToken(), true, true, "{1}"},
	}
	for _, tc := range cases {
		abs := game.ActivatedAbilitiesForCard(tc.tmpl)
		if len(abs) != 1 {
			t.Fatalf("%s: %d abilities, want 1", tc.name, len(abs))
		}
		got := abs[0].Cost
		if got.Tap != tc.tap || got.SacrificeSelf != tc.sacSelf || got.Mana != tc.mana {
			t.Errorf("%s cost = %+v, want tap=%v sac=%v mana=%s",
				tc.name, got, tc.tap, tc.sacSelf, tc.mana)
		}
	}
	// Powerstone is a mana ability, not an activated one.
	if ma := game.ManaAbilitiesForCard(PowerstoneToken()); len(ma) != 1 || ma[0].Produced != "{C}" {
		t.Errorf("Powerstone mana ability = %+v", ma)
	}
	if len(game.ActivatedAbilitiesForCard(PowerstoneToken())) != 0 {
		t.Errorf("Powerstone should have no stack-using ability")
	}
}

// Cracking a Clue: pay {2}, sacrifice it, draw a card. No tap in the
// cost, so it works the turn it arrives.
func TestCrackingAClueDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clue := pushToken(g, me.ID, ClueToken())
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	handBefore := me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Clue: %v", err)
	}
	if g.Battlefield.Contains(clue) {
		t.Errorf("the Clue should be sacrificed at announce")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the {2} should be spent: pool %+v", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}
}

// A Food needs to tap as well, so it can't be cracked the turn a
// creature-shaped source would be sick — but a Food is an artifact,
// never summoning-sick, so it works immediately.
func TestCrackingAFoodGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	food := pushToken(g, me.ID, FoodToken())
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := me.Life

	if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Food: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("life %d -> %d, want +3", before, me.Life)
	}
	if g.Battlefield.Contains(food) {
		t.Errorf("the Food should be gone")
	}
}

// --- producers ---------------------------------------------------

func TestThrabenInspectorInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Thraben Inspector", TypeLine: "Creature — Human Soldier",
		OracleID: thrabenInspectorOracle, Power: 1, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: id})
	})
	passPriorityAroundTable(t, g)

	if findBattlefieldByName(g, "Clue") == uuid.Nil {
		t.Fatalf("investigating should create a Clue token")
	}
	// And the Clue it made is a real one — crackable.
	clue := findBattlefieldByName(g, "Clue")
	if len(game.ActivatedAbilitiesForCard(cardByID(g, clue))) != 1 {
		t.Errorf("the created Clue carries no ability")
	}
}

// Ichor Wellspring draws on the way in AND on the way out — the
// first non-creature dies-trigger in the catalog.
func TestIchorWellspringDrawsOnBothHalves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Ichor Wellspring", TypeLine: "Artifact",
		OracleID: ichorWellspringOracle, Owner: me.ID, Controller: me.ID,
	})
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: id})
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Fatalf("ETB half: hand %d -> %d, want +1", before, me.Hand.Size())
	}

	// Sacrificing it (not just destroying) still counts as dying.
	if err := g.SacrificePermanent(me.ID, id); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+2 {
		t.Errorf("dies half: hand %d, want %d", me.Hand.Size(), before+2)
	}
}

func cardByID(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}
