package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merfolk batch (#5368, cmdr2 batch 5).

const (
	cmdr2B05SvyelunOracle     = "872bebff-720d-459b-8f86-fb9129a8d1fb"
	cmdr2B05KopalaOracle      = "0237cccc-3161-4d91-8e69-e3d315dcb246"
	cmdr2B05BanneretOracle    = "0d6202c7-b19c-429a-85c4-332024381dd7"
	cmdr2B05SurgespannerOracl = "3f15349a-6ac6-4091-95a3-ae9f593e4650"
	cmdr2B05SingerOracle      = "7df43670-c948-464d-9bb9-db672e755f0c"
	cmdr2B05MihailOracle      = "6c901ff4-15c3-468f-9a1b-4c69a7bbc31a"
	cmdr2B05DeeprootOracle    = "764e8e6e-0ace-4fc8-8ca5-468a10ae3b73"
	cmdr2B05KumenaOracle      = "e7125493-febc-4e8e-abdc-5fc4f94a281c"
	cmdr2B05MistbinderOracle  = "be56dfea-1ece-4f06-b1cc-d7425b5018b4"
	cmdr2B05ShorestalkerOracl = "71a60247-17e5-44bd-9fed-f7d550b09ef3"
)

func cmdr2B05Card(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tt int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tt, Owner: owner, Controller: owner,
	})
}

func TestCmdr2B05Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B05SvyelunOracle:     "Svyelun of Sea and Sky",
		cmdr2B05MistbinderOracle:  "Merfolk Mistbinder",
		cmdr2B05ShorestalkerOracl: "Triton Shorestalker",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || spec.Name != name {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

func TestCmdr2B05ShorestalkerCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := cmdr2B05Card(g, me.ID, "Triton Shorestalker", "Creature — Merfolk Rogue", cmdr2B05ShorestalkerOracl, 1, 1)
	assertRestrictions(t, g, id, game.CantBeBlocked)
}

func TestCmdr2B05MistbinderPumpsOnlyYourOtherMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mb := cmdr2B05Card(g, me.ID, "Merfolk Mistbinder", "Creature — Merfolk Shaman", cmdr2B05MistbinderOracle, 2, 2)
	mine := b39Creature(g, me.ID, "Mine", "Creature — Merfolk", 1, 1)
	theirs := b39Creature(g, opp.ID, "Theirs", "Creature — Merfolk", 1, 1)
	bear := seedBear(g, me.ID)
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("your Merfolk power %d, want 2", got)
	}
	for name, id := range map[string]uuid.UUID{"self": mb, "opponent's": theirs} {
		want := 1
		if name == "self" {
			want = 2
		}
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("%s power %d, want %d", name, got, want)
		}
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("bear power %d, want 2", got)
	}
}

func TestCmdr2B05SvyelunIndestructibleNeedsTwoOtherMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sv := cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	if hasKeywordOf(g, sv, "indestructible") {
		t.Fatal("alone, Svyelun is not indestructible")
	}
	b39Creature(g, me.ID, "One", "Creature — Merfolk", 1, 1)
	if hasKeywordOf(g, sv, "indestructible") {
		t.Fatal("one other Merfolk is not enough")
	}
	b39Creature(g, me.ID, "Two", "Creature — Merfolk", 1, 1)
	if !hasKeywordOf(g, sv, "indestructible") {
		t.Fatal("two other Merfolk make Svyelun indestructible")
	}
}

func TestCmdr2B05SvyelunOpponentsMerfolkDoNotCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sv := cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	b39Creature(g, opp.ID, "A", "Creature — Merfolk", 1, 1)
	b39Creature(g, opp.ID, "B", "Creature — Merfolk", 1, 1)
	if hasKeywordOf(g, sv, "indestructible") {
		t.Fatal("only Merfolk you control count")
	}
}

func TestCmdr2B05SvyelunAttackDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sv := cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	advanceToMain(t, g)
	before := me.Hand.Size()
	declareAttack(t, g, opp.ID, sv)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestCmdr2B05SvyelunGivesOtherMerfolkWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	merfolk := b12Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 6; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if hasPayUnlessFor(g, opp.ID) {
		t.Error("a non-Merfolk has ward")
	}
	passPriorityAroundTable(t, g)
	castAtWardedCreature(t, g, opp, merfolk)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward prompt for a spell aimed at another Merfolk")
	}
}

// --- Kopala ---------------------------------------------------------

const cmdr2B05ProbeOracle = "test-cmdr2-b05-probe"

func init() {
	// {1}: target creature. A probe ability, usable by any player.
	Register(Spec{
		OracleID: cmdr2B05ProbeOracle,
		Name:     "B05 Probe",
		Activated: []ActivatedAbility{{
			Label:   "{1}: Target creature does nothing.",
			Cost:    ManaCost("{1}"),
			Targets: TargetCreature("target creature"),
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}

func TestCmdr2B05KopalaTaxesOpponentsSpellsAtYourMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B05Card(g, me.ID, "Kopala, Warden of Waves", "Legendary Creature — Merfolk Wizard", cmdr2B05KopalaOracle, 2, 2)
	merfolk := b39Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }

	if got := priceOpponentsBoltAgainst(t, g, opp, at(merfolk)); got != 3 {
		t.Errorf("opponent's spell at my Merfolk costs %d, want 3", got)
	}
	if got := priceOpponentsBoltAgainst(t, g, opp, at(bear)); got != 1 {
		t.Errorf("opponent's spell at my non-Merfolk costs %d, want 1", got)
	}
	if got := priceOpponentsBoltAgainst(t, g, opp, at(theirs)); got != 1 {
		t.Errorf("opponent's spell at their own Merfolk costs %d, want 1", got)
	}
	if got := priceOpponentsBoltAgainst(t, g, me, at(merfolk)); got != 1 {
		t.Errorf("my own spell at my Merfolk costs %d, want 1", got)
	}
}

func TestCmdr2B05KopalaTaxesOpponentsAbilitiesAtYourMerfolk(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B05Card(g, me.ID, "Kopala, Warden of Waves", "Legendary Creature — Merfolk Wizard", cmdr2B05KopalaOracle, 2, 2)
	merfolk := b39Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := seedBear(g, me.ID)
	probe := pushCatalogPermanent(g, opp.ID, "B05 Probe", "Artifact", cmdr2B05ProbeOracle, false)

	if err := strictProbe(t, g, opp.ID, probe, "{C}", merfolk); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("an opponent's ability at my Merfolk with {1}: err = %v, want insufficient mana", err)
	}
	drainPool(t, g, opp.ID)
	if err := strictProbe(t, g, opp.ID, probe, "{C}{C}{C}", merfolk); err != nil {
		t.Fatalf("with {3} the ability is payable: %v", err)
	}
	drainPool(t, g, opp.ID)
	probe2 := pushCatalogPermanent(g, opp.ID, "B05 Probe", "Artifact", cmdr2B05ProbeOracle, false)
	if err := strictProbe(t, g, opp.ID, probe2, "{C}", bear); err != nil {
		t.Fatalf("an ability at a non-Merfolk is untaxed: %v", err)
	}
}

// --- Stonybrook Banneret -------------------------------------------

func TestCmdr2B05BanneretDiscountsMerfolkAndWizardSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cmdr2B05Card(g, me.ID, "Stonybrook Banneret", "Creature — Merfolk Wizard", cmdr2B05BanneretOracle, 1, 1)
	for _, c := range []struct {
		name, typeLine, cost string
		seat                 *game.Player
		want                 int
	}{
		{"Merfolk", "Creature — Merfolk", "{1}{U}", me, 1},
		{"Wizard", "Creature — Human Wizard", "{2}{U}", me, 2},
		{"Both", "Creature — Merfolk Wizard", "{3}", me, 2},
		{"Bear", "Creature — Bear", "{1}{G}", me, 2},
		{"Opponent's Merfolk", "Creature — Merfolk", "{1}{U}", opp, 2},
	} {
		if got := priceInHand(t, g, c.seat, c.name, c.typeLine, c.cost); got != c.want {
			t.Errorf("%s costs %d, want %d", c.name, got, c.want)
		}
	}
}

func TestCmdr2B05BanneretHasIslandwalk(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	b := cmdr2B05Card(g, me.ID, "Stonybrook Banneret", "Creature — Merfolk Wizard", cmdr2B05BanneretOracle, 1, 1)
	bear := pushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	seedLand(g, them.ID, "Island", "Basic Land — Island", "")
	attackIntoBlocks(t, g, them.ID, b)
	if r := blockRefusal(t, g, bear, b); r != game.BlockReasonLandwalk {
		t.Fatalf("blocking an islandwalker against an Island: reason %q, want landwalk", r)
	}
}
