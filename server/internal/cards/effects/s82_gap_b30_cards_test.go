package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gap30Hoodwink     = "a2fe10b8-857e-4881-85ef-24a0f96a3d79"
	gap30ManaShort    = "48207d1c-448a-4e1b-974a-642dfea75933"
	gap30Dispute      = "48390b8d-db41-48a5-9bac-316660ab7252"
	gap30Reset        = "4faa5a70-ac0c-4a29-8bec-218a35a29fdf"
	gap30Paradoxical  = "2cce7250-f8f9-4c0a-87b1-ba3c5b8ed6b3"
	gap30Gush         = "16d8ee99-8ec8-429f-9ba7-818a74b6f910"
	gap30BasiliskGate = "8733a4fc-4068-4af4-9598-dc3d895e8556"
)

func TestGapB30CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gap30Hoodwink:     "Hoodwink",
		gap30ManaShort:    "Mana Short",
		gap30Dispute:      "Mystical Dispute",
		gap30Reset:        "Reset",
		gap30Paradoxical:  "Paradoxical Outcome",
		gap30Gush:         "Gush",
		gap30BasiliskGate: "Basilisk Gate",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB30HoodwinkBouncesALandAndRefusesACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Hoodwink", "Instant", gap30Hoodwink, cardTarget(bear)); err == nil {
		t.Errorf("Hoodwink was cast at a creature")
	}
	castCatalogSpell(t, g, "Hoodwink", "Instant", gap30Hoodwink, cardTarget(land))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(land) {
		t.Errorf("the land did not return to its owner's hand")
	}
}

func TestGapB30ManaShortTapsTheLandsAndEmptiesThePool(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	apaMana(opp, "G", "G")
	castCatalogSpell(t, g, "Mana Short", "Instant", gap30ManaShort, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if c := apaLive(g, land); c == nil || !c.Tapped {
		t.Errorf("the land is not tapped")
	}
	if got := opp.ManaPool.Total(); got != 0 {
		t.Errorf("unspent mana %d, want 0", got)
	}
}

func TestGapB30MysticalDisputeCostsTwoLessAgainstABlueSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	blue := handCardFull(me, "Blue Thing", "Sorcery", "{1}{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the blue spell: %v", err)
	}
	apaMana(me, "U")
	dispute := handCardFull(me, "Mystical Dispute", "Instant", "{2}{U}", gap30Dispute, []string{"U"})
	if err := g.CastSpell(me.ID, dispute, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: blue}}}); err != nil {
		t.Errorf("Mystical Dispute for {U} against a blue spell: %v", err)
	}
}

func TestGapB30MysticalDisputeIsFullPriceAgainstANonblueSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	red := handCardFull(me, "Red Thing", "Sorcery", "{1}{R}", "", []string{"R"})
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the red spell: %v", err)
	}
	apaMana(me, "U")
	dispute := handCardFull(me, "Mystical Dispute", "Instant", "{2}{U}", gap30Dispute, []string{"U"})
	if err := g.CastSpell(me.ID, dispute, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: red}}}); err == nil {
		t.Errorf("Mystical Dispute for {U} against a red spell")
	}
}

func TestGapB30ResetOnlyOnAnOpponentsTurnAfterUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == land {
				c.Tapped = true
			}
		}
	})
	advanceToMain(t, g)
	apaMana(me, "U", "U")
	id := handCardFull(me, "Reset", "Instant", "{U}{U}", gap30Reset, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatalf("Reset was cast on its controller's own turn")
	}
	endTurn(t, g)
	advanceToMain(t, g)
	apaMana(me, "U", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Reset on an opponent's main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := apaLive(g, land); c == nil || c.Tapped {
		t.Errorf("the land was not untapped")
	}
}

func TestGapB30ParadoxicalOutcomeReturnsAndDrawsPerCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Permanent(g, me.ID, "Rock A", "Artifact")
	b := b12Permanent(g, me.ID, "Rock B", "Artifact")
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Paradoxical Outcome", "Instant", gap30Paradoxical, cardTargets(a, b))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("the artifacts did not return")
	}
	if !g.Battlefield.Contains(land) {
		t.Errorf("the land left the battlefield")
	}
	// Two returned cards in hand, two drawn, and the spell itself left the hand.
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand size %d, want %d", got, hand+4)
	}
}

func TestGapB30ParadoxicalOutcomeRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErrStrict(t, g, "Paradoxical Outcome", "Instant", gap30Paradoxical, cardTarget(land)); err == nil {
		t.Errorf("Paradoxical Outcome targeted a land")
	}
}

func TestGapB30GushReturnsTwoIslandsAsACostAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	i1 := freeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	i2 := freeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	i3 := freeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	gush := handCardFull(me, "Gush", "Instant", "{4}{U}", gap30Gush, []string{"U"})
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, gush, game.CastSpellParams{Strict: true, AlternativeCost: "return",
		AltCostIDs: []uuid.UUID{i1}}); err == nil {
		t.Fatalf("Gush was cast returning one Island")
	}
	if err := g.CastSpell(me.ID, gush, game.CastSpellParams{Strict: true, AlternativeCost: "return",
		AltCostIDs: []uuid.UUID{i1, i2}}); err != nil {
		t.Fatalf("Gush returning two Islands: %v", err)
	}
	if !me.Hand.Contains(i1) || !me.Hand.Contains(i2) || !g.Battlefield.Contains(i3) {
		t.Errorf("the two Islands did not return (and only them)")
	}
	passPriorityAroundTable(t, g)
	// Hand: -Gush, +2 Islands, +2 drawn.
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand size %d, want %d", got, hand+3)
	}
}

func TestGapB30BasiliskGatePumpsByTheGatesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	gate := b21Push(g, me.ID, "Basilisk Gate", "Land — Gate", gap30BasiliskGate, 0, 0)
	b21Push(g, me.ID, "Other Gate", "Land — Gate", "", 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := apaLive(g, bear); c == nil || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
}
