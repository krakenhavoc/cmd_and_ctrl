package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_296b_test.go — card slice 296-b ("Cost reducers"), relates to
// #296. Each card gets a main-behaviour test plus, where the card has
// one, a refused/illegal case.

const (
	sapphireMedallionOracle   = "f1ef5698-14bd-42bf-aae5-2870699e4186"
	etheriumSculptorOracle    = "96b87445-6362-4a17-91b2-cbf203fd03fd"
	goblinAnarchomancerOracle = "e9319f13-1f2b-429c-9d61-e58a3cbec86a"
	emeraldMedallionOracle    = "d683024d-8fe7-4509-851b-7655fadc1de7"
	emryLurkerOfTheLochOracle = "da3e7d3d-2ca0-40c3-9602-fca37c92f507"
	cloudKeyOracle            = "2a838818-d590-4374-9a63-d9e6381a0f0d"
	goreclawOracle            = "befb211f-37ca-4083-98d4-9ff1f28be3f2"
	bontusMonumentOracle      = "940ba435-7abc-40f8-a5af-c1c653b3284e"
	theEarthCrystalOracle     = "5e7ef7fe-968b-4ada-9fe4-6fda0541aafc"
	danithaCapashenOracle     = "4b6377da-83e7-4519-9582-16a9c16b8faa"
	theFireCrystalOracle      = "57276db8-9d8d-4587-ae88-dfdc44343f17"
	stormcatchMentorOracle    = "b2252471-01ca-4d58-aa99-4ab0aa5eae12"
	archmageOfRunesOracle     = "28045b32-4c1a-40e5-a15d-524d0f8fe6ec"
	jhoirasFamiliarOracle     = "df4b5f7b-9d83-49b1-bd5e-77d0652eb34c"
	nightscapeFamiliarOracle  = "57296ea3-3c0d-49b7-bc08-0d0d8414e9ad"
)

// priceInHandWithPower is priceInHand plus a printed power, for the
// cards whose discount depends on the spell's power (Goreclaw).
func priceInHandWithPower(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, manaCost string, power int) int {
	t.Helper()
	base, err := game.ParseCost(manaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", manaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{
		Card: game.Card{
			InstanceID: uuid.New(),
			Name:       name,
			TypeLine:   typeLine,
			ManaCost:   manaCost,
			Power:      power,
			Owner:      seat.ID,
			Controller: seat.ID,
		},
		Controller: seat.ID,
		FromZone:   game.ZoneHand,
	})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

// --- Sapphire Medallion / Emerald Medallion -----------------------

func TestSapphireMedallionDiscountsBlueSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sapphire Medallion", "Artifact", sapphireMedallionOracle, false)

	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{1}{U}"); got != 1 {
		t.Errorf("blue spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("red spell should be unaffected: %d, want 1", got)
	}
}

func TestEmeraldMedallionDiscountsGreenSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Emerald Medallion", "Artifact", emeraldMedallionOracle, false)

	if got := priceInHand(t, g, me, "Rampant Growth", "Sorcery", "{1}{G}"); got != 1 {
		t.Errorf("green spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{R}"); got != 1 {
		t.Errorf("red spell should be unaffected: %d, want 1", got)
	}
}

// --- Etherium Sculptor ----------------------------------------------

func TestEtheriumSculptorDiscountsArtifactSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Etherium Sculptor", "Artifact Creature — Vedalken Artificer",
		etheriumSculptorOracle, false)

	if got := priceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 0 {
		t.Errorf("artifact spell: %d, want 0", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell should be unaffected: %d, want 2", got)
	}
}

// --- Goblin Anarchomancer -------------------------------------------

func TestGoblinAnarchomancerDiscountsRedOrGreenSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Goblin Anarchomancer", "Creature — Goblin Shaman",
		goblinAnarchomancerOracle, false)

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("red spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Giant Growth", "Instant", "{1}{G}"); got != 1 {
		t.Errorf("green spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Rampant Growth", "Sorcery", "{1}{G}{G}"); got != 2 {
		t.Errorf("Gruul (red-and-green counts once): %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("blue spell should be unaffected: %d, want 2", got)
	}
}

// --- Emry, Lurker of the Loch ---------------------------------------

func TestEmryAffinityMillAndGraveyardCastPermission(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		pushPermanent(g, me.ID, game.Card{Name: "Signet", TypeLine: "Artifact"})
	}

	// Affinity: {2}{U} with three artifacts discounts the two
	// generic pips, leaving {U} (mana value 1).
	if got := selfPricedMV(t, g, me, emryLurkerOfTheLochOracle,
		"Legendary Creature — Merfolk Wizard", "{2}{U}"); got != 1 {
		t.Errorf("Emry with three artifacts: %d, want 1", got)
	}

	// ETB mill four.
	before := me.Graveyard.Size()
	emry := castCatalogSpell(t, g, "Emry, Lurker of the Loch",
		"Legendary Creature — Merfolk Wizard", emryLurkerOfTheLochOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != before+4 {
		t.Errorf("graveyard after ETB: %d -> %d, want +4", before, got)
	}
	// The tap ability is being tested for its own sake, not for
	// timing — clear the summoning sickness a freshly cast Emry would
	// otherwise still have this turn (CR 302.6).
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == emry {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}

	// Tap ability: grant permission to cast an artifact card from the
	// graveyard this turn, at its own printed cost.
	sig := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: sig, Name: "Old Signet", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, emry, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: sig}},
	}); err != nil {
		t.Fatalf("Emry's tap ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, sig, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("casting the granted permission: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sig) {
		t.Error("the artifact Emry granted permission over never reached the battlefield")
	}
}

func TestEmryTapAbilityRefusesANonArtifactTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emry := pushCatalogPermanent(g, me.ID, "Emry, Lurker of the Loch",
		"Legendary Creature — Merfolk Wizard", emryLurkerOfTheLochOracle, false)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, emry, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a non-artifact card in the graveyard should not be a legal target")
	}
}

// --- Cloud Key --------------------------------------------------------

func TestCloudKeyDiscountsOnlyTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Cloud Key", "Artifact", cloudKeyOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, "instant")
	if got := g.ChosenOptionOf(id); got != "instant" {
		t.Fatalf("Cloud Key chose %q, want \"instant\"", got)
	}

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant spell (chosen type): %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell (not the chosen type) should be unaffected: %d, want 2", got)
	}
}

// --- Goreclaw, Terror of Qal Sisma ------------------------------------

func TestGoreclawDiscountsBigCreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Goreclaw, Terror of Qal Sisma",
		"Legendary Creature — Bear", goreclawOracle, false)

	if got := priceInHandWithPower(t, g, me, "Big Beater", "Creature — Beast", "{4}{G}{G}", 4); got != 4 {
		t.Errorf("power-4 creature spell: %d, want 4", got)
	}
	if got := priceInHandWithPower(t, g, me, "Small Beater", "Creature — Beast", "{4}{G}{G}", 3); got != 6 {
		t.Errorf("power-3 creature spell should be unaffected: %d, want 6", got)
	}
	if got := priceInHand(t, g, me, "Giant Growth", "Instant", "{G}"); got != 1 {
		t.Errorf("non-creature spell should be unaffected: %d, want 1", got)
	}
}

func TestGoreclawAttackPumpsOnlyBigCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]
	goreclaw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goreclaw, Terror of Qal Sisma",
		TypeLine: "Legendary Creature — Bear", OracleID: goreclawOracle,
		Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	small := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Small Beater", TypeLine: "Creature — Beast",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: goreclaw, Target: victim.ID},
	}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if p := effectivePower(t, g, goreclaw); p != 5 {
		t.Errorf("Goreclaw's own power after its attack trigger: %d, want 5", p)
	}
	if !eotHasAbility(effectiveAbilities(t, g, goreclaw), "trample") {
		t.Error("Goreclaw should gain trample from its own trigger")
	}
	if p := effectivePower(t, g, small); p != 2 {
		t.Errorf("a power-2 creature should not be pumped: %d, want 2", p)
	}
	if eotHasAbility(effectiveAbilities(t, g, small), "trample") {
		t.Error("a power-2 creature should not gain trample")
	}
}

// --- Bontu's Monument ---------------------------------------------------

func TestBontusMonumentDiscountsBlackCreaturesAndDrainsOnAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Bontu's Monument", "Legendary Artifact", bontusMonumentOracle, false)

	if got := priceInHand(t, g, me, "Black Bear", "Creature — Bear", "{2}{B}"); got != 2 {
		t.Errorf("black creature spell: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Green Bear", "Creature — Bear", "{2}{G}"); got != 3 {
		t.Errorf("non-black creature spell should be unaffected: %d, want 3", got)
	}

	lifeBefore := lifeOfOpponents(g)
	meBefore := me.Life
	castCatalogSpell(t, g, "Any Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range lifeBefore {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller life: %d -> %d, want +1", meBefore, me.Life)
	}

	// A noncreature spell doesn't drain.
	lifeMid := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range lifeMid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d drained by a noncreature spell", i+1)
		}
	}
}

// --- The Earth Crystal ---------------------------------------------------

func TestEarthCrystalDoublesCountersAndDistributesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crystal := pushCatalogPermanent(g, me.ID, "The Earth Crystal", "Legendary Artifact",
		theEarthCrystalOracle, false)

	if got := priceInHand(t, g, me, "Rampant Growth", "Sorcery", "{1}{G}"); got != 1 {
		t.Errorf("green spell: %d, want 1", got)
	}

	bear := seedCreature(g, "Bear", me.ID)
	if err := g.AddCounter(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("one +1/+1 counter doubled: %d, want 2", got)
	}

	bear2 := seedCreature(g, "Bear Two", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, crystal, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: bear2}},
		Distribution: map[uuid.UUID]int{bear: 1, bear2: 1},
	}); err != nil {
		t.Fatalf("distribute counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("bear after distribution (also doubled): %d, want 4 (2 + 1×2)", got)
	}
	if got := countersOn(g, bear2, game.CounterPlusOne); got != 2 {
		t.Errorf("bear two after distribution (doubled): %d, want 2", got)
	}
}

// --- Danitha Capashen, Paragon -------------------------------------------

func TestDanithaCapashenKeywordsAndAuraEquipmentDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Danitha Capashen, Paragon",
		"Legendary Creature — Human Knight", danithaCapashenOracle, false)
	for _, kw := range []string{"first strike", "vigilance", "lifelink"} {
		if !eotHasAbility(effectiveAbilities(t, g, id), kw) {
			t.Errorf("Danitha should have %q", kw)
		}
	}

	if got := priceInHand(t, g, me, "Aura Spell", "Enchantment — Aura", "{1}{G}"); got != 1 {
		t.Errorf("Aura spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Bonesplitter", "Artifact — Equipment", "{1}"); got != 0 {
		t.Errorf("Equipment spell: %d, want 0", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("plain creature spell should be unaffected: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Plain Enchantment", "Enchantment", "{1}{W}"); got != 2 {
		t.Errorf("plain enchantment (not an Aura) should be unaffected: %d, want 2", got)
	}
}

// --- The Fire Crystal ---------------------------------------------------

func TestFireCrystalDiscountHasteAndTokenCopy(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	victim := g.Seats[(seat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "The Fire Crystal", "Legendary Artifact", theFireCrystalOracle, false)

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("red spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("non-red spell should be unaffected: %d, want 2", got)
	}

	// Haste: a freshly entered creature can attack this turn.
	fresh := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fresh Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToStepOf(t, g, seat, game.StepDeclareAttackers)
	makeSick(t, g, fresh)
	if err := g.DeclareAttacker(fresh, victim.ID); err != nil {
		t.Errorf("a creature under The Fire Crystal's haste grant should be able to attack: %v", err)
	}
}

// --- Stormcatch Mentor ---------------------------------------------------

func TestStormcatchMentorKeywordsAndDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Stormcatch Mentor", "Creature — Otter Wizard",
		stormcatchMentorOracle, false)
	for _, kw := range []string{"haste", game.KeywordProwess} {
		if !eotHasAbility(effectiveAbilities(t, g, id), kw) {
			t.Errorf("Stormcatch Mentor should have %q", kw)
		}
	}

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell should be unaffected: %d, want 2", got)
	}
}

// --- Archmage of Runes ---------------------------------------------------

func TestArchmageOfRunesDiscountsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Archmage of Runes", "Creature — Giant Wizard",
		archmageOfRunesOracle, false)

	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant spell: %d, want 1", got)
	}

	before := me.Hand.Size()
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after casting an instant: %d -> %d, want +1 (draw trigger)", before, got)
	}

	mid := me.Hand.Size()
	castCatalogSpell(t, g, "Any Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != mid {
		t.Errorf("hand after casting a creature spell: %d -> %d, want unchanged", mid, got)
	}
}

// --- Jhoira's Familiar ---------------------------------------------------

func TestJhoirasFamiliarFlyingAndHistoricDiscount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Jhoira's Familiar", "Artifact Creature — Bird",
		jhoirasFamiliarOracle, false)
	if !eotHasAbility(effectiveAbilities(t, g, id), "flying") {
		t.Error("Jhoira's Familiar should have flying")
	}

	if got := priceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 0 {
		t.Errorf("artifact (historic) spell: %d, want 0", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Legendary Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("legendary (historic) spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Plain Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a plain, non-historic creature spell should be unaffected: %d, want 2", got)
	}
}

// --- Nightscape Familiar ---------------------------------------------------

func TestNightscapeFamiliarDiscountAndRegenerate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Nightscape Familiar", "Creature — Zombie",
		nightscapeFamiliarOracle, false)

	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{1}{U}"); got != 1 {
		t.Errorf("blue spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("red spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Giant Growth", "Instant", "{G}"); got != 1 {
		t.Errorf("green spell should be unaffected: %d, want 1", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Regenerate activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	if !g.Battlefield.Contains(id) {
		t.Error("a regenerated Nightscape Familiar should not be destroyed")
	}
}
