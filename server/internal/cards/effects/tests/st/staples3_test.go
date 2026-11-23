package testsst

// staples3_test.go — third staples batch: the cards that turn the S21
// sacrifice machinery into an engine (Pitiless Plunderer, Bastion of
// Remembrance, Vampiric Rites), plus Ravenous Chupacabra and Aura
// Shards on the targeted-trigger pipeline.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Pitiless Plunderer -----------------------------------------

func TestPitilessPlundererMakesTreasureOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", effectstest.PitilessPlundererOracle)
	victim := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)

	if err := g.SacrificePermanentForEffect(victim); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// "ANOTHER creature you control" — the Plunderer's own death must not
// trigger it, which is what separates it from Zulaport Cutthroat.
func TestPitilessPlundererIgnoresItsOwnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plunderer := effectstest.SeedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", effectstest.PitilessPlundererOracle)

	if err := g.SacrificePermanentForEffect(plunderer); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasure tokens, want 0 — its own death is not 'another'", got)
	}
}

// An opponent's creature dying is not "a creature YOU control".
func TestPitilessPlundererIgnoresOpponentDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Pitiless Plunderer",
		"Creature — Human Pirate", effectstest.PitilessPlundererOracle)
	theirs := effectstest.SeedCreature(g, "Grizzly Bears", opp.ID)

	if err := g.SacrificePermanentForEffect(theirs); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 0 {
		t.Errorf("%d Treasure tokens, want 0", got)
	}
}

// --- Bastion of Remembrance -------------------------------------

func TestBastionOfRemembranceMakesASoldierOnETB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Bastion of Remembrance", "Enchantment",
		effectstest.BastionOfRemembranceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, caster.ID, "Human Soldier"); got != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", got)
	}
}

// The drain hits every opponent for 1 and gains exactly 1 — not one
// per opponent — and it survives creature removal because the source
// is an enchantment.
func TestBastionOfRemembranceDrainsOnCreatureDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedPermanentWithOracle(g, me.ID, "Bastion of Remembrance",
		"Enchantment", effectstest.BastionOfRemembranceOracle)
	victim := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)

	myLife := me.Life
	oppLife := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppLife[p.ID] = p.Life
	}

	if err := g.SacrificePermanentForEffect(victim); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	prey := effectstest.SeedCreature(g, "Blightsteel Colossus", opp.ID)

	effectstest.CastCatalogSpell(t, g, "Ravenous Chupacabra", "Creature — Beast Horror",
		effectstest.RavenousChupacabraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PickCard(t, g, caster.ID, prey)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(prey) {
		t.Error("Chupacabra did not destroy the chosen creature")
	}
}

// "an opponent controls" — your own creatures are never in the legal
// set, so a board with only your creatures produces no prompt at all
// (CR 603.3d) rather than forcing you to shoot yourself.
func TestRavenousChupacabraNeverTargetsYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	mine := effectstest.SeedCreature(g, "Llanowar Elves", caster.ID)

	effectstest.CastCatalogSpell(t, g, "Ravenous Chupacabra", "Creature — Beast Horror",
		effectstest.RavenousChupacabraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p := effectstest.LatestPickTarget(g, caster.ID); p != nil {
		t.Error("Chupacabra prompted with only own creatures on board")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("Chupacabra destroyed the controller's own creature")
	}
}

// --- Aura Shards ------------------------------------------------

// Fires on ANY creature you control entering, not just its own ETB.
func TestAuraShardsTriggersOnCreatureETB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", effectstest.AuraShardsOracle)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	// Any creature entering under the Shards controller sets it off.
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, caster.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", effectstest.AuraShardsOracle)
	effectstest.SeedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	if err := g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1); err != nil {
		t.Fatalf("create opponent token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if p := effectstest.LatestPickTarget(g, caster.ID); p != nil {
		t.Error("Aura Shards triggered on an opponent's creature")
	}
	if c := effectstest.LatestTriggerPrompt(g, caster.ID); c != nil {
		t.Error("Aura Shards prompted on an opponent's creature")
	}
}

// The control for the test above: the same token path under the
// Shards' controller DOES trigger it. Without this, a broken ETB emit
// would make the negative case above meaningless.
func TestAuraShardsTriggersOnYourToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedPermanentWithOracle(g, caster.ID, "Aura Shards", "Enchantment", effectstest.AuraShardsOracle)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Sol Ring", "Artifact")

	if err := g.CreateTokenForEffect(caster.ID, effects.RedGoblinToken(), 1); err != nil {
		t.Fatalf("create own token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestPickTarget(g, caster.ID) == nil {
		t.Fatal("a token entering under your control must trigger Aura Shards")
	}
	effectstest.PickCard(t, g, caster.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, caster.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("Aura Shards did not destroy the chosen artifact")
	}
}

// --- Vampiric Rites ---------------------------------------------

// The cost pairs mana with sacrifice-ANOTHER, unlike Mind Stone's
// sacrifice-self.
func TestVampiricRitesCostShape(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.VampiricRitesOracle})
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
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[0]
	handBefore, lifeBefore := caster.Hand.Size(), caster.Life

	effectstest.CastCatalogSpell(t, g, "Ambition's Cost", "Sorcery", effectstest.AmbitionsCostOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// handBefore predates the spell being seeded, so its arrival and
	// departure cancel and the delta is the three cards drawn.
	if got := caster.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand delta %d, want 3", got)
	}
	if got := lifeBefore - caster.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}
