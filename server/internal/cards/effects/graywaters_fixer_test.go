package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graywaters_fixer_test.go — #5429: a granted graveyard ability whose mana
// cost is the card's own mana value (game.AbilityCost.ManaValueFromSource).

const graywatersFixerOracle = "93438939-3909-4e59-bfe4-760ec73eb57d"

// gfBoard is the Fixer on my battlefield at my precombat main.
func gfBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	fixer := pushCatalogPermanent(g, me.ID, "Graywater's Fixer", "Creature — Lizard Mercenary", graywatersFixerOracle, false)
	return g, me, fixer
}

// gfDead puts an uncatalogued card into p's graveyard by a real move (a
// discard's zone change), so the zone grant sees it arrive.
func gfDead(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
		t.Fatalf("move %s to the graveyard: %v", name, err)
	}
	return id
}

// gfEncoreCost is the mana of the one activated row the card has, or
// ok=false when it has none.
func gfEncoreCost(g *game.Game, id uuid.UUID) (cost string, ok bool) {
	g.ReadSnapshot(func() {
		c, found := g.LookupCardForEffect(id)
		if !found {
			return
		}
		rows := game.ActivatedAbilitiesForCard(c)
		if len(rows) == 1 {
			cost, ok = rows[0].Cost.Mana, true
		}
	})
	return cost, ok
}

func gfTokens(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller && c.IsToken() {
				n++
			}
		}
	})
	return n
}

// Each outlaw creature card in my graveyard has the row, priced at its own
// mana value; other cards, and other players' outlaws, have nothing.
func TestGraywatersFixerGrantsEncorePricedByEachCardsManaValue(t *testing.T) {
	g, me, _ := gfBoard(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, tc := range []struct {
		name, typeLine, cost, want string
	}{
		{"Cheap Rogue", "Creature — Human Rogue", "{B}", "{1}"},
		{"Big Pirate", "Creature — Orc Pirate", "{3}{R}{R}", "{5}"},
		{"Hybrid Assassin", "Creature — Vampire Assassin", "{1}{U/B}{B}", "{3}"},
		{"X Warlock", "Creature — Human Warlock", "{X}{B}{B}", "{2}"},
		{"Free Mercenary", "Artifact Creature — Construct Mercenary", "{0}", "{0}"},
	} {
		id := gfDead(t, g, me, tc.name, tc.typeLine, tc.cost)
		if got, ok := gfEncoreCost(g, id); !ok || got != tc.want {
			t.Errorf("%s (%s): encore cost %q (has a row: %v), want %q", tc.name, tc.cost, got, ok, tc.want)
		}
	}
	for _, tc := range []struct{ name, typeLine string }{
		{"Dead Bear", "Creature — Bear"},
		{"Rogue's Passage", "Land"},
		{"Pirate's Pillage", "Sorcery"},
		{"Kindred Rogue Spell", "Kindred Instant — Rogue"},
	} {
		id := gfDead(t, g, me, tc.name, tc.typeLine, "{2}")
		if _, ok := gfEncoreCost(g, id); ok {
			t.Errorf("%s (%s) was given encore; it is not an outlaw creature card", tc.name, tc.typeLine)
		}
	}
	theirs := gfDead(t, g, opp, "Their Rogue", "Creature — Human Rogue", "{B}")
	if _, ok := gfEncoreCost(g, theirs); ok {
		t.Error("an outlaw in another player's graveyard was given encore")
	}
}

// The granted encore is the keyword: its cost is the card's mana value and
// the exile, and it makes one hasty token copy per opponent.
func TestGraywatersFixerGrantedEncoreCostsTheManaValueAndMakesATokenPerOpponent(t *testing.T) {
	g, me, _ := gfBoard(t)
	rogue := gfDead(t, g, me, "Dead Rogue", "Creature — Human Rogue", "{1}{B}{B}")

	// Two mana cannot pay encore {3}.
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("encore {3} was activated with two mana")
	}
	if !me.Graveyard.Contains(rogue) {
		t.Fatal("a refused activation exiled the card")
	}
	// Any three mana pay it: the cost is generic, not the card's {1}{B}{B}.
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("encore for three generic mana: %v", err)
	}
	if !g.Exile.Contains(rogue) {
		t.Fatal("encore's cost exiles the card from the graveyard")
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Fatalf("%d mana left after paying encore {3} from three", left)
	}
	passPriorityAroundTable(t, g)
	if n := gfTokens(g, "Dead Rogue", me.ID); n != len(g.Seats)-1 {
		t.Fatalf("%d token copies, want one per opponent (%d)", n, len(g.Seats)-1)
	}
	var toks []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dead Rogue" {
				toks = append(toks, c.InstanceID)
			}
		}
	})
	for _, tok := range toks {
		if !hasEffectiveKeyword(t, g, tok, "haste") {
			t.Errorf("token %s has no haste", tok)
		}
	}
	// Each token attacks its own opponent, and all are sacrificed at the
	// beginning of the next end step.
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	for _, tok := range toks {
		var must uuid.UUID
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(tok)
			for _, r := range c.Effective().AttackRequirements {
				if r.MustAttack != uuid.Nil {
					must = r.MustAttack
				}
			}
		})
		if must == uuid.Nil || must == me.ID {
			t.Fatalf("token %s is not bound to an opponent", tok)
		}
		if err := g.DeclareAttacker(tok, must); err != nil {
			t.Fatalf("declare the token at its opponent: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with every token at its own opponent: %v", err)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := gfTokens(g, "Dead Rogue", me.ID); n != 0 {
		t.Errorf("%d tokens survived the end step", n)
	}
}

// "Activate only as a sorcery", and only its owner activates it.
func TestGraywatersFixerGrantedEncoreIsSorcerySpeedAndTheOwners(t *testing.T) {
	g, me, _ := gfBoard(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rogue := gfDead(t, g, me, "Dead Rogue", "Creature — Human Rogue", "{B}")
	floatForTest(g, opp, "C")
	if err := g.ActivateCatalogAbility(opp.ID, rogue, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an opponent activated encore on a card in my graveyard")
	}
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("encore was activated in the beginning of combat step")
	}
	if !me.Graveyard.Contains(rogue) {
		t.Fatal("a refused activation exiled the card")
	}
}

// The row leaves with the Fixer, but an encore already on the stack
// resolves (CR 113.7a).
func TestGraywatersFixerGrantEndsWhenItLeavesButAnActivatedEncoreResolves(t *testing.T) {
	g, me, fixer := gfBoard(t)
	first := gfDead(t, g, me, "First Rogue", "Creature — Human Rogue", "{B}")
	second := gfDead(t, g, me, "Second Rogue", "Creature — Human Rogue", "{B}")
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("encore: %v", err)
	}
	killCreature(t, g, me.ID, fixer)
	if _, ok := gfEncoreCost(g, second); ok {
		t.Fatal("a card still has encore after the Fixer left the battlefield")
	}
	// The Fixer is a Mercenary, but in the graveyard it grants nothing,
	// itself included.
	if _, ok := gfEncoreCost(g, fixer); ok {
		t.Fatal("the Fixer's own card has encore in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if n := gfTokens(g, "First Rogue", me.ID); n != len(g.Seats)-1 {
		t.Fatalf("%d token copies from the encore already activated, want %d", n, len(g.Seats)-1)
	}
}

// A card that prints encore keeps its own row beside the granted one, each
// at its own price.
func TestGraywatersFixerLeavesAPrintedEncoreBesideTheGrantedOne(t *testing.T) {
	g, me, _ := gfBoard(t)
	id := pushCatalogGraveyardCard(me, "Impulsive Pilferer", "Creature — Goblin Pirate", impulsivePilfererOracle, 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].ManaCost = "{R}"
			}
		}
	})
	g.RunStateChecksForTest()
	var costs []string
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		for _, r := range game.ActivatedAbilitiesForCard(c) {
			costs = append(costs, r.Cost.Mana)
		}
	})
	if len(costs) != 2 || costs[0] != "{3}{R}" || costs[1] != "{1}" {
		t.Fatalf("encore rows cost %v, want the printed {3}{R} and the granted {1}", costs)
	}
}
