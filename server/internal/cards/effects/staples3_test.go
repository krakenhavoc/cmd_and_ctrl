package effects

// staples3_test.go — third staples batch: the cards that turn the S21
// sacrifice machinery into an engine (Pitiless Plunderer, Bastion of
// Remembrance, Vampiric Rites), plus Ravenous Chupacabra and Aura
// Shards on the targeted-trigger pipeline.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	pitilessPlundererOracle    = "a784481f-eccb-4112-bb38-04a659319660"
	bastionOfRemembranceOracle = "c7f33cea-2ec8-4081-9208-a5b1d86721b3"
	ravenousChupacabraOracle   = "7b459306-149b-4f43-abc1-2dd70c748c0e"
	auraShardsOracle           = "8d03d050-391c-4311-8c42-4ee632d40fdc"
	vampiricRitesOracle        = "660de988-b6fb-4f36-8006-42af3e7f908d"
	ambitionsCostOracle        = "84de4fec-2f38-4293-93d3-b3882c5aac14"
)

// seedPermanentWithOracle puts a catalog permanent on the battlefield
// so its triggered abilities are live without casting it.
func seedPermanentWithOracle(g *game.Game, owner uuid.UUID, name, typeLine, oracleID string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Power:      1,
		Toughness:  4,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// --- Pitiless Plunderer -----------------------------------------

func TestPitilessPlundererMakesTreasureOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", pitilessPlundererOracle)
	victim := seedCreature(g, "Doomed Traveler", me.ID)

	if err := g.SacrificePermanentForEffect(victim); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// "ANOTHER creature you control" — the Plunderer's own death must not
// trigger it, which is what separates it from Zulaport Cutthroat.
func TestPitilessPlundererIgnoresItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plunderer := seedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", pitilessPlundererOracle)

	if err := g.SacrificePermanentForEffect(plunderer); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasure tokens, want 0 — its own death is not 'another'", got)
	}
}

// An opponent's creature dying is not "a creature YOU control".
func TestPitilessPlundererIgnoresOpponentDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", pitilessPlundererOracle)
	theirs := seedCreature(g, "Grizzly Bears", opp.ID)

	if err := g.SacrificePermanentForEffect(theirs); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasure tokens, want 0", got)
	}
}

// --- Bastion of Remembrance -------------------------------------

func TestBastionOfRemembranceMakesASoldierOnETB(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]

	castCatalogSpell(t, g, "Bastion of Remembrance", "Enchantment",
		bastionOfRemembranceOracle, nil)
	passPriorityAroundTable(t, g)

	if got := countBattlefieldNamed(g, caster.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}
}

// The drain hits every opponent for 1 and gains exactly 1 — not one
// per opponent — and it survives creature removal because the source
// is an enchantment.
func TestBastionOfRemembranceDrainsOnCreatureDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Bastion of Remembrance",
		"Enchantment", bastionOfRemembranceOracle)
	victim := seedCreature(g, "Doomed Traveler", me.ID)

	myLife := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}

	if err := g.SacrificePermanentForEffect(victim); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := me.Life - myLife; got != 1 {
		t.Errorf("controller gained %d life, want exactly 1 regardless of table size", got)
	}
	for _, p := range g.Seats[1:] {
		if got := oppLife[p.ID] - p.Life; got != 1 {
			t.Errorf("opponent %s lost %d life, want 1", p.Name, got)
		}
	}
}

// --- Ravenous Chupacabra ----------------------------------------

// Mandatory targeted ETB trigger: no yes/no prompt, straight to the
// target pick.
func TestRavenousChupacabraDestroysOnETB(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	prey := seedCreature(g, "Blightsteel Colossus", opp.ID)

	castCatalogSpell(t, g, "Ravenous Chupacabra", "Creature — Beast Horror",
		ravenousChupacabraOracle, nil)
	passPriorityAroundTable(t, g)

	pickCard(t, g, caster.ID, prey)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(prey) {
		t.Error("Chupacabra did not destroy the chosen creature")
	}
}

// "an opponent controls" — your own creatures are never in the legal
// set, so a board with only your creatures produces no prompt at all
// (CR 603.3d) rather than forcing you to shoot yourself.
func TestRavenousChupacabraNeverTargetsYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	mine := seedCreature(g, "Llanowar Elves", caster.ID)

	castCatalogSpell(t, g, "Ravenous Chupacabra", "Creature — Beast Horror",
		ravenousChupacabraOracle, nil)
	passPriorityAroundTable(t, g)

	if p := latestPickTarget(g, caster.ID); p != nil {
		t.Error("Chupacabra prompted with only own creatures on board")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("Chupacabra destroyed the controller's own creature")
	}
}

// --- Aura Shards ------------------------------------------------

// Fires on ANY creature you control entering, not just its own ETB.
func TestAuraShardsTriggersOnCreatureETB(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", auraShardsOracle)
	rock := seedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	// Any creature entering under the Shards controller sets it off.
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	answerLatestTriggerPrompt(t, g, caster.ID, true)
	passPriorityAroundTable(t, g)
	pickCard(t, g, caster.ID, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Aura Shards did not destroy the chosen artifact")
	}
}

// An opponent's creature entering must not trigger it.
//
// The creature enters via CreateTokenForEffect rather than a direct
// battlefield push: a push emits no EventETB, which would make this
// test pass whether or not the predicate is right. The paired
// controller-side case below proves the token path really does emit
// ETB, so a silent no-op here is a genuine negative.
func TestAuraShardsIgnoresOpponentCreatures(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", auraShardsOracle)
	seedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	if err := g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1); err != nil {
		t.Fatalf("create opponent token: %v", err)
	}
	passPriorityAroundTable(t, g)

	if p := latestPickTarget(g, caster.ID); p != nil {
		t.Error("Aura Shards triggered on an opponent's creature")
	}
	if c := latestTriggerPrompt(g, caster.ID); c != nil {
		t.Error("Aura Shards prompted on an opponent's creature")
	}
}

// The control for the test above: the same token path under the
// Shards' controller DOES trigger it. Without this, a broken ETB emit
// would make the negative case above meaningless.
func TestAuraShardsTriggersOnYourToken(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", auraShardsOracle)
	rock := seedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	if err := g.CreateTokenForEffect(caster.ID, RedGoblinToken(), 1); err != nil {
		t.Fatalf("create own token: %v", err)
	}
	passPriorityAroundTable(t, g)

	if latestTriggerPrompt(g, caster.ID) == nil {
		t.Fatal("a token entering under your control must trigger Aura Shards")
	}
	answerLatestTriggerPrompt(t, g, caster.ID, true)
	passPriorityAroundTable(t, g)
	pickCard(t, g, caster.ID, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Aura Shards did not destroy the chosen artifact")
	}
}

// --- Vampiric Rites ---------------------------------------------

// The cost pairs mana with sacrifice-ANOTHER, unlike Mind Stone's
// sacrifice-self.
func TestVampiricRitesCostShape(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: vampiricRitesOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want 1", len(abilities))
	}
	cost := abilities[0].Cost
	if cost.SacrificeOther == nil {
		t.Error("cost should sacrifice another creature, not itself")
	}
	if cost.SacrificeSelf {
		t.Error("Vampiric Rites must not sacrifice itself — it's the outlet")
	}
	if cost.Mana != "{1}{B}" {
		t.Errorf("mana %q, want {1}{B}", cost.Mana)
	}
	if cost.Tap {
		t.Error("no tap component — an enchantment doesn't tap")
	}
}

// --- Ambition's Cost --------------------------------------------

func TestAmbitionsCostDrawsThreeLosesThree(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore, lifeBefore := caster.Hand.Size(), caster.Life

	castCatalogSpell(t, g, "Ambition's Cost", "Sorcery", ambitionsCostOracle, nil)
	passPriorityAroundTable(t, g)

	// handBefore predates the spell being seeded, so its arrival and
	// departure cancel and the delta is the three cards drawn.
	if got := caster.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand delta %d, want 3", got)
	}
	if got := lifeBefore - caster.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}
