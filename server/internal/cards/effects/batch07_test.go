package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch07_test.go — card-level coverage for the card-coverage
// roadmap's batch 07 (#300, `edhrec_rank` 798–901): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b07-prefixed.

const (
	b07TwinflameTyrantOracle       = "321b5cc6-8df6-4292-97a0-a6a3a22f3b55"
	b07SteelOverseerOracle         = "986ae327-f433-4c58-93dc-afc544b9bfcb"
	b07MesaEnchantressOracle       = "8f4b8a19-72f4-48ed-ac05-62a7c7525797"
	b07SaiMasterThopteristOracle   = "52241b9c-7a69-4176-9234-8bdab09d8e64"
	b07GhostQuarterOracle          = "2ec4288e-34c6-4831-a2c0-ba1ca1d9d1dc"
	b07YoungPyromancerOracle       = "5fac139a-07d3-4e6c-98e3-d98b199f7a6f"
	b07SoulsAttendantOracle        = "045a9d3d-c20d-427f-a77b-0bffc6f55526"
	b07CurseOfTheSwineOracle       = "5669ea7c-c4fc-494c-896b-4bce9b494817"
	b07UrzasTowerOracle            = "32fbb638-ab14-4e8b-a07a-d4c44e3496f2"
	b07KodamaOfTheWestTreeOracle   = "d69b1e68-8d8e-460b-9eb4-6a68be886197"
	b07MurderOracle                = "938b4e2c-88d9-4637-bc00-e228920c9a78"
	b07NadiersNightbladeOracle     = "391978f6-0bbc-41e8-9246-f7d0e21c7900"
	b07BloodthirstyConquerorOracle = "fd505c02-c59e-476d-8e88-35da862ddc23"
	b07AyaraOracle                 = "388168b3-ec68-4af2-b88c-6a5ec88c15f6"
	b07UrzasMineOracle             = "33e85a8a-86df-4cdc-a9cc-8cbabe92c3c0"
	b07FieryEmancipationOracle     = "52159875-354c-47f9-bb1c-cd65395fcc68"
	b07UrzasPowerPlantOracle       = "e11966cd-2ee3-4df4-b099-abf42dcdf0db"
	b07FieryIsletOracle            = "026f4a4b-eedd-44e1-9d37-ca4fb8d6db98"
	b07EleshNornOracle             = "958d71ff-c9f7-46f0-96ca-79e7f4d65a16"
	b07TemptWithDiscoveryOracle    = "4baa6145-216e-476b-b178-aaaa1e633701"
	b07CityOnFireOracle            = "41eec4e8-92d3-4f98-9346-3a3e3cc602ce"
	b07MysticGateOracle            = "e9f5feb2-2c1a-46ce-885a-4f378d7d10af"
	b07SplendidReclamationOracle   = "13fe5e46-77a6-45d8-ac0b-c3d740eccf86"
	b07MyrRetrieverOracle          = "d07d3be3-f69d-4484-8467-cffd43871788"
	b07TemurAscendancyOracle       = "e68dc47c-692f-4420-9799-eee104017273"
	b07SerumVisionsOracle          = "56956afd-db53-4542-816b-490c8b0bbcf7"
	b07EidolonOfBlossomsOracle     = "77ccbea1-70af-4194-adad-39a904221c75"
	b07FieldOfRuinOracle           = "f825c98f-a327-440b-8c0d-ebe02e23bfb7"
)

// b07LibraryLand seeds a land card into a player's library.
func b07LibraryLand(p *game.Player, name, typeLine string) uuid.UUID {
	return pushLibraryCardForTest(p, game.Card{Name: name, TypeLine: typeLine})
}

// b07LandsOnBattlefield counts the lands `controller` controls.
func b07LandsOnBattlefield(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsLand() {
			n++
		}
	}
	return n
}

// b07OpenSearchPrompts counts the search prompts open at the table,
// for the "never two over one library" invariant.
func b07OpenSearchPrompts(g *game.Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary {
			n++
		}
	}
	return n
}

// b07BoltPlayer casts Lightning Bolt at a player and settles it.
func b07BoltPlayer(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
	passPriorityAroundTable(t, g)
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The Tron
// lands are a loop over a table, so a transposed row is invisible
// until someone plays that exact card.
func TestBatch07CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b07TwinflameTyrantOracle:       "Twinflame Tyrant",
		b07SteelOverseerOracle:         "Steel Overseer",
		b07MesaEnchantressOracle:       "Mesa Enchantress",
		b07SaiMasterThopteristOracle:   "Sai, Master Thopterist",
		b07GhostQuarterOracle:          "Ghost Quarter",
		b07YoungPyromancerOracle:       "Young Pyromancer",
		b07SoulsAttendantOracle:        "Soul's Attendant",
		b07CurseOfTheSwineOracle:       "Curse of the Swine",
		b07UrzasTowerOracle:            "Urza's Tower",
		b07KodamaOfTheWestTreeOracle:   "Kodama of the West Tree",
		b07MurderOracle:                "Murder",
		b07NadiersNightbladeOracle:     "Nadier's Nightblade",
		b07BloodthirstyConquerorOracle: "Bloodthirsty Conqueror",
		b07AyaraOracle:                 "Ayara, First of Locthwain",
		b07UrzasMineOracle:             "Urza's Mine",
		b07FieryEmancipationOracle:     "Fiery Emancipation",
		b07UrzasPowerPlantOracle:       "Urza's Power Plant",
		b07FieryIsletOracle:            "Fiery Islet",
		b07EleshNornOracle:             "Elesh Norn, Grand Cenobite",
		b07TemptWithDiscoveryOracle:    "Tempt with Discovery",
		b07CityOnFireOracle:            "City on Fire",
		b07MysticGateOracle:            "Mystic Gate",
		b07SplendidReclamationOracle:   "Splendid Reclamation",
		b07MyrRetrieverOracle:          "Myr Retriever",
		b07TemurAscendancyOracle:       "Temur Ascendancy",
		b07SerumVisionsOracle:          "Serum Visions",
		b07EidolonOfBlossomsOracle:     "Eidolon of Blossoms",
		b07FieldOfRuinOracle:           "Field of Ruin",
	}
	if len(want) != 28 {
		t.Fatalf("the batch is 28 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness == CompletenessUnreviewed {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- damage multipliers --------------------------------------------

func TestB07TwinflameTyrantDoublesDamageToOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Twinflame Tyrant", "Creature — Dragon", b07TwinflameTyrantOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != oppBefore-6 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -6", oppBefore, opp.Life)
	}
	// Damage to yourself is not "to an opponent".
	b07BoltPlayer(t, g, me.ID)
	if me.Life != meBefore-3 {
		t.Errorf("a Bolt at yourself: %d → %d, want -3 (not doubled)", meBefore, me.Life)
	}
}

func TestB07FieryEmancipationTriplesEverything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Fiery Emancipation", "Enchantment", b07FieryEmancipationOracle, false)
	oppBefore, meBefore := opp.Life, me.Life

	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != oppBefore-9 {
		t.Errorf("a Bolt at an opponent: %d → %d, want -9", oppBefore, opp.Life)
	}
	// Your own sources hurting you are tripled too, as printed.
	b07BoltPlayer(t, g, me.ID)
	if me.Life != meBefore-9 {
		t.Errorf("a Bolt at yourself: %d → %d, want -9", meBefore, me.Life)
	}
}

func TestB07CityOnFireConvokesAndTriples(t *testing.T) {
	spec, ok := Lookup(b07CityOnFireOracle)
	if !ok || spec.TapCost == nil {
		t.Fatal("City on Fire must declare convoke as its tap cost")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "City on Fire", "Enchantment", b07CityOnFireOracle, false)
	before := opp.Life
	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != before-9 {
		t.Errorf("a Bolt under City on Fire: %d → %d, want -9", before, opp.Life)
	}
}

// --- cast triggers -------------------------------------------------

func TestB07MesaEnchantressMayDrawOnAnEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mesa Enchantress", "Creature — Human Druid", b07MesaEnchantressOracle, false)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}

	// A creature spell is not an enchantment spell: no prompt.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if len(g.PendingChoices) != 0 {
		t.Error("a creature spell prompted the Enchantress")
	}
	passPriorityAroundTable(t, g)
}

func TestB07SaiMakesAThopterPerArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", b07SaiMasterThopteristOracle, false)

	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Thopter") != 1 {
		t.Fatal("an artifact spell should make one Thopter")
	}
	thopter := findBattlefieldByName(g, "Thopter")
	if card, _ := battlefieldCard(g, thopter); !card.IsArtifact() || !card.IsCreature() {
		t.Error("the Thopter is an artifact creature")
	}
	if !eotHasAbility(effectiveAbilities(t, g, thopter), "flying") {
		t.Error("the Thopter flies")
	}

	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Thopter") != 1 {
		t.Error("a creature spell is not an artifact spell")
	}
}

func TestB07YoungPyromancerMakesAnElementalPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Young Pyromancer", "Creature — Human Shaman", b07YoungPyromancerOracle, false)

	b07BoltPlayer(t, g, opp.ID)
	castCatalogSpell(t, g, "Ponder", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Elemental"); n != 2 {
		t.Fatalf("an instant and a sorcery made %d Elementals, want 2", n)
	}
	elemental := findBattlefieldByName(g, "Elemental")
	if card, _ := battlefieldCard(g, elemental); !card.HasColor("R") {
		t.Error("the Elemental is red")
	}

	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Elemental") != 2 {
		t.Error("an artifact spell triggered the Pyromancer")
	}
}

// --- ETB triggers --------------------------------------------------

func TestB07SoulsAttendantMayGainOnAnyOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	castCatalogSpell(t, g, "Soul's Attendant", "Creature — Human Cleric", b07SoulsAttendantOracle, nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("the Attendant's own entry is not 'another creature'")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d, want +1 for an opponent's creature", before, me.Life)
	}

	// Declining is a real answer.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d → %d after declining, want still +1", before, me.Life)
	}
}

func TestB07EidolonOfBlossomsDrawsForItselfAndOtherEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Eidolon of Blossoms", "Enchantment Creature — Spirit", b07EidolonOfBlossomsOracle, nil)
	passPriorityAroundTable(t, g)
	// castCatalogSpell adds the spell to the hand and casting removes
	// it, so the difference is exactly the cards drawn.
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("drew %d, want 1 — the Eidolon's own entry draws (constellation counts itself)", got)
	}

	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d after another enchantment, want 2", got)
	}

	// A creature that is not an enchantment, and an opponent's
	// enchantment, are both silent.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	seedPermanentFor(g, opp.ID, "Their Glory", "Enchantment")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d, want still 2", got)
	}
}

func TestB07TemurAscendancyGrantsHasteAndMayDrawForPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Temur Ascendancy", "Enchantment", b07TemurAscendancyOracle, false)
	bear := seedCreature(g, "Bear", me.ID)
	if !eotHasAbility(effectiveAbilities(t, g, bear), "haste") {
		t.Error("creatures you control have haste")
	}
	before := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 colorless Beast"), 1) })
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatal("a 3/3 entering must not prompt")
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Wurm", TypeLine: "Token Creature — Wurm", Power: 5, Toughness: 5}, 1)
	})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1 for the 5/5", got)
	}
}

func TestB07AyaraDrainsOnBlackCreaturesAndEatsThemForCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := lifeOfOpponents(g)
	meBefore := me.Life

	ayara := castCatalogSpell(t, g, "Ayara, First of Locthwain", "Legendary Creature — Elf Noble", b07AyaraOracle, nil)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d after Ayara's own entry: %d → %d, want -1", i+1, b, got)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("Ayara's controller %d → %d, want +1", meBefore, me.Life)
	}

	// A black creature drains again; a green one does not.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, BlackZombieToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TokenCard("3/3 colorless Beast"), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d after a Zombie and a Beast: %d → %d, want -2", i+1, b, got)
		}
	}

	// The sacrifice cost takes another black creature, not a green
	// one and not Ayara herself. She was cast this turn, so shake
	// off the summoning sickness for the tap cost.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ayara {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	zombie := findBattlefieldByName(g, "Zombie")
	beast := findBattlefieldByName(g, "Beast")
	for _, bad := range []uuid.UUID{beast, ayara} {
		if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
			SacrificeIDs: []uuid.UUID{bad},
		}); err == nil {
			t.Errorf("%s was accepted for 'sacrifice another black creature'", bad)
		}
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{zombie},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if card, _ := battlefieldCard(g, ayara); !card.Tapped {
		t.Error("the ability has a tap cost")
	}
}

// --- life-loss and leaves-the-battlefield triggers -----------------

func TestB07BloodthirstyConquerorGainsWhatOpponentsLose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	conq := pushCatalogPermanent(g, me.ID, "Bloodthirsty Conqueror", "Creature — Vampire Knight", b07BloodthirstyConquerorOracle, false)
	if abs := effectiveAbilities(t, g, conq); !eotHasAbility(abs, "flying") || !eotHasAbility(abs, "deathtouch") {
		t.Error("printed flying and deathtouch did not reach the effective abilities")
	}
	before := me.Life

	b07BoltPlayer(t, g, opp.ID)
	if me.Life != before+3 {
		t.Fatalf("after a Bolt to an opponent: %d → %d, want +3", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -2) })
	passPriorityAroundTable(t, g)
	if me.Life != before+5 {
		t.Errorf("after a 2-life drain: %d → %d, want +5", before, me.Life)
	}
	// Your own loss is not an opponent's.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	passPriorityAroundTable(t, g)
	if me.Life != before+4 {
		t.Errorf("after my own loss: %d → %d, want +4", before, me.Life)
	}
}

func TestB07NadiersNightbladeDrainsWhenATokenLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Nadier's Nightblade", "Creature — Elf Warrior", b07NadiersNightbladeOracle, false)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	before := lifeOfOpponents(g)
	meBefore := me.Life

	// Sacrificed and bounced both count: "leaves the battlefield".
	treasure := findBattlefieldByName(g, "Treasure")
	goblin := uuid.Nil
	theirGoblin := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Goblin" {
			if c.Controller == me.ID {
				goblin = c.InstanceID
			} else {
				theirGoblin = c.InstanceID
			}
		}
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(treasure)
		_ = g.BounceToHandForEffect(goblin)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (two tokens left)", i+1, b, got)
		}
	}
	if me.Life != meBefore+2 {
		t.Errorf("controller %d → %d, want +2", meBefore, me.Life)
	}

	// An opponent's token and a nontoken creature are silent.
	bear := seedCreature(g, "Bear", me.ID)
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(theirGoblin)
		_ = g.SacrificePermanentForEffect(bear)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d → %d, want still -2", i+1, b, got)
		}
	}
}

func TestB07MyrRetrieverReturnsAnotherArtifactCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	retriever := pushCatalogPermanent(g, me.ID, "Myr Retriever", "Artifact Creature — Myr", b07MyrRetrieverOracle, false)
	relic := batch01GraveyardCard(me, "Sol Ring", "Artifact")
	bear := pushGraveyardCardForTest(me, "Dead Bear")

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(retriever) })
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	for _, id := range pick.PickTargetCards {
		if id == bear {
			t.Error("a creature card was offered for 'target artifact card'")
		}
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card did not come back to hand")
	}
}

// --- statics -------------------------------------------------------

func TestB07EleshNornPumpsYoursAndShrinksTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Mine", me.ID)
	small := seedCreature(g, "Their Small", opp.ID)
	big := b04Creature(g, opp.ID, "Their Big", 5, 5)

	norn := castCatalogSpell(t, g, "Elesh Norn, Grand Cenobite", "Legendary Creature — Phyrexian Praetor", b07EleshNornOracle, nil)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my 2/2 = %d/%d, want 4/4", p, tough)
	}
	if p := effectivePower(t, g, norn); p != 0 {
		t.Errorf("Norn pumps OTHER creatures, her own power = %d, want the printed 0 of the fixture", p)
	}
	if g.Battlefield.Contains(small) {
		t.Error("an opponent's 2/2 dies to -2/-2")
	}
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 3 || tough != 3 {
		t.Errorf("their 5/5 = %d/%d, want 3/3", p, tough)
	}
	if !eotHasAbility(effectiveAbilities(t, g, norn), "vigilance") {
		t.Error("printed vigilance did not reach the effective abilities")
	}
}

func TestB07KodamaGivesModifiedCreaturesTrampleAndRampsOnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Kodama of the West Tree", "Legendary Creature — Spirit", b07KodamaOfTheWestTreeOracle, false)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	countered := pushVanillaCreature(g, me.ID, "Countered", 2, 2)
	equipped := pushVanillaCreature(g, me.ID, "Equipped", 2, 2)
	theirs := seedCreature(g, "Theirs", victim.ID)
	blade := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blade", TypeLine: "Artifact — Equipment",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(countered, "+1/+1", 1)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
		if err := g.AttachForEffect(blade, game.TargetRef{Kind: game.TargetCard, ID: equipped}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	for _, tc := range []struct {
		id   uuid.UUID
		want bool
		why  string
	}{
		{countered, true, "a creature with a counter is modified"},
		{equipped, true, "an equipped creature is modified"},
		{plain, false, "an unmodified creature gets nothing"},
		{theirs, false, "an opponent's modified creature is not 'you control'"},
	} {
		if got := eotHasAbility(effectiveAbilities(t, g, tc.id), "trample"); got != tc.want {
			t.Errorf("%s: trample = %v, want %v", tc.why, got, tc.want)
		}
	}

	forest := b07LibraryLand(me, "Forest", "Basic Land — Forest")
	attackWith(t, g, victim.ID, countered, plain)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, forest)
	if !ok {
		t.Fatal("one modified creature connecting should fetch one basic")
	}
	if !card.Tapped {
		t.Error("the basic enters tapped")
	}
}

// --- activated abilities -------------------------------------------

func TestB07SteelOverseerGrowsEveryArtifactCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	overseer := pushCatalogPermanent(g, me.ID, "Steel Overseer", "Artifact Creature — Construct", b07SteelOverseerOracle, false)
	myr := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr", TypeLine: "Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := seedCreature(g, "Bear", me.ID)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Myr", TypeLine: "Artifact Creature — Myr",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{overseer, 1}, {myr, 1}, {bear, 0}, {theirs, 0}} {
		if card, _ := battlefieldCard(g, tc.id); card.Counters["+1/+1"] != tc.want {
			t.Errorf("%s: +1/+1 counters = %d, want %d", card.Name, card.Counters["+1/+1"], tc.want)
		}
	}
}

func TestB07SteelOverseerRespectsSummoningSickness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	overseer := pushCatalogPermanent(g, me.ID, "Steel Overseer", "Artifact Creature — Construct", b07SteelOverseerOracle, true)
	if err := g.ActivateCatalogAbility(me.ID, overseer, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Overseer must not tap")
	}
}

func TestB07FieryIsletPaysLifeForColourAndCashesInForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	islet := seedPermanentWithOracle(g, me.ID, "Fiery Islet", "Land", b07FieryIsletOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, islet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("the Islet asks U or R, got %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}

	fresh := seedPermanentWithOracle(g, me.ID, "Fiery Islet", "Land", b07FieryIsletOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, fresh, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fresh) {
		t.Error("the Islet is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestB07GhostQuarterDestroysALandAndOffersItsControllerABasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quarter := pushCatalogPermanent(g, me.ID, "Ghost Quarter", "Land", b07GhostQuarterOracle, false)
	coffers := seedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	swamp := b07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	b07LibraryLand(opp, "Swamp", "Basic Land — Swamp")

	if err := g.ActivateCatalogAbility(me.ID, quarter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(quarter) {
		t.Error("the Quarter is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the targeted land should be destroyed")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the activator does not search — the land's controller does")
	}
	answerSearchByID(t, g, opp.ID, swamp)
	card, ok := battlefieldCard(g, swamp)
	if !ok {
		t.Fatal("the victim's basic did not reach the battlefield")
	}
	if card.Tapped {
		t.Error("the basic enters untapped, as printed")
	}
}

func TestB07GhostQuarterOnYourOwnLandIsAFetch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	quarter := pushCatalogPermanent(g, me.ID, "Ghost Quarter", "Land", b07GhostQuarterOracle, false)
	mine := seedLandOnBattlefield(g, me.ID, "Wastes", "Basic Land")
	forest := b07LibraryLand(me, "Forest", "Basic Land — Forest")
	b07LibraryLand(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, quarter, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) {
		t.Error("targeting your own land should let you fetch")
	}
}

func TestB07FieldOfRuinDestroysANonbasicAndEveryoneSearches(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	field := pushCatalogPermanent(g, me.ID, "Field of Ruin", "Land", b07FieldOfRuinOracle, false)
	coffers := seedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	basic := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	// Seat 0 and seat 1 hold two basics each (a prompt), seat 2 holds
	// one (served without a prompt), seat 3 holds none.
	wanted := map[uuid.UUID]uuid.UUID{}
	for _, p := range []*game.Player{me, opp} {
		wanted[p.ID] = b07LibraryLand(p, "Plains", "Basic Land — Plains")
		b07LibraryLand(p, "Plains", "Basic Land — Plains")
	}
	only := b07LibraryLand(g.Seats[2], "Island", "Basic Land — Island")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}},
	}); err == nil {
		t.Fatal("a basic land was accepted for 'target nonbasic land'")
	}
	if err := g.ActivateCatalogAbility(me.ID, field, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: coffers}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(coffers) {
		t.Error("the targeted nonbasic should be destroyed")
	}
	if !g.Battlefield.Contains(only) {
		t.Error("a player with exactly one basic is served without a prompt")
	}
	if b07OpenSearchPrompts(g) != 2 {
		t.Fatalf("want a prompt for each of the two players with a choice, got %d", b07OpenSearchPrompts(g))
	}
	for _, p := range []*game.Player{me, opp} {
		answerSearchByID(t, g, p.ID, wanted[p.ID])
		if card, ok := battlefieldCard(g, wanted[p.ID]); !ok || card.Tapped {
			t.Errorf("%s's basic should be on the battlefield untapped", p.Name)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB07MurderDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := seedCreature(g, "Bear", opp.ID)
	castCatalogSpell(t, g, "Murder", "Instant", b07MurderOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature should be in its owner's graveyard")
	}
}

func TestB07CurseOfTheSwineExilesXAndHandsOutBoars(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	spared := seedCreature(g, "Spared", opp.ID)

	castXSpell(t, g, "Curse of the Swine", "Sorcery", b07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s should be in exile", id)
		}
	}
	if !g.Battlefield.Contains(spared) {
		t.Error("an untargeted creature survives")
	}
	if countBattlefieldNamed(g, me.ID, "Boar") != 1 || countBattlefieldNamed(g, opp.ID, "Boar") != 1 {
		t.Error("each exiled creature's controller gets one Boar")
	}
}

func TestB07CurseOfTheSwineDemandsExactlyXTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := seedCreature(g, "Theirs", opp.ID)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Curse of the Swine", TypeLine: "Sorcery",
		OracleID: b07CurseOfTheSwineOracle, ManaCost: "{X}{U}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Error("one target was accepted for X=2")
	}
}

func TestB07SerumVisionsDrawsThenScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Serum Visions", "Sorcery", b07SerumVisionsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	scry := scryChoiceFor(g, me.ID)
	if scry == nil || len(scry.ScryCards) != 2 {
		t.Fatalf("want a scry prompt over two cards, got %+v", scry)
	}
}

func TestB07SplendidReclamationReturnsEveryLandCardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := batch01GraveyardCard(me, "Forest", "Basic Land — Forest")
	coffers := batch01GraveyardCard(me, "Cabal Coffers", "Land")
	bear := pushGraveyardCardForTest(me, "Dead Bear")
	theirs := batch01GraveyardCard(opp, "Their Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Splendid Reclamation", "Sorcery", b07SplendidReclamationOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{forest, coffers} {
		card, ok := battlefieldCard(g, id)
		if !ok {
			t.Errorf("%s did not return", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("%s returned untapped", card.Name)
		}
		if card.Controller != me.ID {
			t.Errorf("%s returned under %s, want the caster", card.Name, card.Controller)
		}
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a creature card stays in the graveyard")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard is not 'your graveyard'")
	}
}

func TestB07TemptWithDiscoveryOffersEachOpponentInTurnAndPaysTheCasterBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	// Every library holds two lands, so every search is a real
	// prompt; the caster's has four, one for each search it can earn.
	for _, p := range []*game.Player{me, opp1, opp2, opp3} {
		for i := 0; i < 2; i++ {
			b07LibraryLand(p, "Forest", "Basic Land — Forest")
		}
	}
	b07LibraryLand(me, "Cabal Coffers", "Land")
	b07LibraryLand(me, "Gaea's Cradle", "Legendary Land")
	b07LibraryLand(me, "Bear", "Creature — Bear")

	castCatalogSpell(t, g, "Tempt with Discovery", "Sorcery", b07TemptWithDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)

	// Step 1: the caster's own search, and nothing else is open yet.
	step := func(who *game.Player, take bool, why string) {
		t.Helper()
		if b07OpenSearchPrompts(g) != 1 {
			t.Fatalf("%s: %d search prompts open, want exactly 1 (the offer is sequential)", why, b07OpenSearchPrompts(g))
		}
		c := searchChoiceFor(g, who.ID)
		if c == nil {
			t.Fatalf("%s: no search prompt for %s", why, who.Name)
		}
		for _, id := range c.SearchCards {
			card, _ := g.LookupCardForEffect(id)
			if !card.IsLand() {
				t.Errorf("%s: a non-land %q was offered for 'a land card'", why, card.Name)
			}
		}
		if take {
			answerSearchByID(t, g, who.ID, c.SearchCards[0])
		} else {
			answerSearchFailToFind(t, g, who.ID)
		}
	}
	step(me, true, "the caster's opening search")
	step(opp1, true, "opponent 1 is offered a land")
	step(me, true, "opponent 1 accepted, so the caster searches again")
	step(opp2, false, "opponent 2 is offered a land")
	step(opp3, true, "opponent 2 declined, so opponent 3 is next with no caster search between")
	step(me, true, "opponent 3 accepted, so the caster searches again")
	if b07OpenSearchPrompts(g) != 0 {
		t.Fatalf("after the last opponent nothing should be open, got %d", b07OpenSearchPrompts(g))
	}

	for _, tc := range []struct {
		p    *game.Player
		want int
	}{{me, 3}, {opp1, 1}, {opp2, 0}, {opp3, 1}} {
		if got := b07LandsOnBattlefield(g, tc.p.ID); got != tc.want {
			t.Errorf("%s controls %d lands, want %d", tc.p.Name, got, tc.want)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Tapped {
			t.Errorf("%s entered tapped; the printed text says untapped", c.Name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB07TronLandsScaleWithTheFullSet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tower := seedPermanentWithOracle(g, me.ID, "Urza's Tower", "Land — Urza's Tower", b07UrzasTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("a lone Tower: pool %v, want [C]", got)
	}

	me.ManaPool.EmptyPool()
	mine := seedPermanentWithOracle(g, me.ID, "Urza's Mine", "Land — Urza's Mine", b07UrzasMineOracle)
	seedPermanentWithOracle(g, me.ID, "Urza's Power Plant", "Land — Urza's Power-Plant", b07UrzasPowerPlantOracle)
	tower2 := seedPermanentWithOracle(g, me.ID, "Urza's Tower", "Land — Urza's Tower", b07UrzasTowerOracle)
	if err := g.ActivateManaAbility(me.ID, tower2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 3 {
		t.Errorf("Tower with Tron: pool %v, want [C C C]", got)
	}
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("Mine with Tron: pool %v, want [C C]", got)
	}
}

func TestB07MysticGateFiltersOneHybridIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gate := seedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", b07MysticGateOracle)

	if err := g.ActivateManaAbility(me.ID, gate, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}

	// The wrong colour floating: refused, Gate untouched.
	fresh := seedPermanentWithOracle(g, me.ID, "Mystic Gate", "Land", b07MysticGateOracle)
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("{W/U} was paid with {G}")
	}
	if card, _ := battlefieldCard(g, fresh); card.Tapped {
		t.Fatal("a refused activation must not tap the Gate")
	}

	// {U} pays the hybrid; the pool then asks twice from {W, U}.
	me.ManaPool.EmptyPool()
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateManaAbility(me.ID, fresh, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if len(c.ColorOptions) != 2 {
				t.Errorf("each slot must offer exactly W and U, got %v", c.ColorOptions)
			}
			if err := g.ResolveManaChoice(c.ID, me.ID, "W"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W]", got)
	}
}
