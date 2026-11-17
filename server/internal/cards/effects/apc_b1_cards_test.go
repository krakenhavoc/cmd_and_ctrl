package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apc_b1_cards_test.go — Apocalypse, batch 1 (#5614). Each test plays the
// real catalog entry and asserts what its text says.

const (
	apcB1AngelOracle     = "2be85439-4606-4e1e-8be7-56d9e31f77c4"
	apcB1SnakeOracle     = "161b6d06-e6a7-488e-9f9c-d266c1402a3a"
	apcB1GraspOracle     = "a2335149-d2db-48c1-9699-0df8ce12d4dd"
	apcB1HunterOracle    = "45baa5d1-79aa-4583-b938-0d5feaa5aa0c"
	apcB1DeadOracle      = "3a48541a-240b-4406-945c-1d0d032eca25"
	apcB1ConsumeOracle   = "53fd9fb1-c28f-4d73-bf98-94cdab418f21"
	apcB1ChargeOracle    = "6226db2b-d1a8-4c41-b802-67d3f65d2ca3"
	apcB1ChargerOracle   = "a19e0a7f-d356-4d2c-8414-70a40ce18674"
	apcB1BoltOracle      = "d5e4bf5e-1a66-4315-8c82-640c27977b88"
	apcB1FireIceOracle   = "ae92942b-919c-4ea9-b693-85fcef765d5a"
	apcB1NightDayOracle  = "cb32ab3e-ef76-4afb-959f-f8cc2d18416e"
	apcB1BlastOracle     = "2a9d9433-7b0f-45a3-88aa-9fd7e410c224"
	apcB1TreefolkOracle  = "04683bd7-b100-4fc8-8165-316c5508c255"
	apcB1DeedOracle      = "62e44e0d-eda0-4275-8367-49dab9a087c3"
	apcB1ShockOracleTest = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
)

func TestApcB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		apcB1AngelOracle:           "Lightning Angel",
		apcB1SnakeOracle:           "Mystic Snake",
		apcB1GraspOracle:           "Death Grasp",
		apcB1HunterOracle:          "Razorfin Hunter",
		apcB1DeadOracle:            "Llanowar Dead",
		apcB1ConsumeOracle:         "Consume Strength",
		apcB1ChargeOracle:          "Fervent Charge",
		apcB1ChargerOracle:         "Flowstone Charger",
		apcB1BoltOracle:            "Prophetic Bolt",
		apcB1FireIceOracle:         "Fire",
		apcB1FireIceOracle + "#1":  "Ice",
		apcB1NightDayOracle:        "Night",
		apcB1NightDayOracle + "#1": "Day",
		apcB1BlastOracle:           "Suffocating Blast",
		apcB1TreefolkOracle:        "Ebony Treefolk",
		apcB1DeedOracle:            "Pernicious Deed",
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

// apcB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func apcB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// apcB1CastHalf casts one half of a split card from the active player's
// hand with the given parameters, without resolving it.
func apcB1CastHalf(t *testing.T, g *game.Game, oracle string, faces [2]game.Face, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: faces[:],
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", faces[params.Face].Name, err)
	}
	return card.InstanceID
}

// apcB1OpponentShock puts an opponent's Shock aimed at `at` on the stack.
func apcB1OpponentShock(t *testing.T, g *game.Game, opp *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	shock := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", OracleID: apcB1ShockOracleTest,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, shock, game.CastSpellParams{Targets: []game.TargetRef{playerTarget(at)}}); err != nil {
		t.Fatalf("cast Shock: %v", err)
	}
	return shock
}

// --- Lightning Angel ---------------------------------------------------------------------

func TestApcB1LightningAngelHasFlyingVigilanceAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	angel := apcB1Push(g, g.Seats[0].ID, "Lightning Angel", "Creature — Angel", apcB1AngelOracle, "{1}{U}{R}{W}", 3, 4)
	for _, kw := range []string{"flying", "vigilance", "haste"} {
		if !hasKeywordOnBattlefield(t, g, angel, kw) {
			t.Errorf("Lightning Angel does not have %s", kw)
		}
	}
}

// --- Mystic Snake ------------------------------------------------------------------------

func TestApcB1MysticSnakeCountersTargetSpellAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	snake := castFromSeat(t, g, them, "Mystic Snake", "Creature — Snake", apcB1SnakeOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, victim)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, snake) {
		t.Error("Mystic Snake is not on the battlefield")
	}
	if battlefieldHas(g, victim) || !me.Graveyard.Contains(victim) {
		t.Error("the targeted spell was not countered into its owner's graveyard")
	}
}

// --- Death Grasp -------------------------------------------------------------------------

func TestApcB1DeathGraspDealsXAndGainsX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine, theirs := me.Life, opp.Life
	d1658Cast(t, g, "Death Grasp", "Sorcery", apcB1GraspOracle, game.CastSpellParams{
		XValue: 3, Targets: []game.TargetRef{playerTarget(opp.ID)},
	})
	passPriorityAroundTable(t, g)
	if got := theirs - opp.Life; got != 3 {
		t.Errorf("the target lost %d life, want 3", got)
	}
	if got := me.Life - mine; got != 3 {
		t.Errorf("the caster gained %d life, want 3", got)
	}
}

// --- Razorfin Hunter ---------------------------------------------------------------------

func TestApcB1RazorfinHunterPingsAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp, _ := blbB1Seats(g)
	hunter := apcB1Push(g, me.ID, "Razorfin Hunter", "Creature — Merfolk Goblin", apcB1HunterOracle, "{U}{R}", 1, 1)
	before := opp.Life
	if err := invB1Activate(g, me.ID, hunter, 0, playerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("the target lost %d life, want 1", got)
	}
	if !findCardAnywhere(t, g, hunter).Tapped {
		t.Error("Razorfin Hunter did not tap")
	}
}

// --- Llanowar Dead -----------------------------------------------------------------------

func TestApcB1LlanowarDeadTapsForBlack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := apcB1Push(g, me.ID, "Llanowar Dead", "Creature — Zombie Elf", apcB1DeadOracle, "{B}{G}", 2, 2)
	if err := g.ActivateManaAbility(me.ID, dead, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("B"); got != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d black of %d, want exactly {B}", got, me.ManaPool.Total())
	}
	if !findCardAnywhere(t, g, dead).Tapped {
		t.Error("Llanowar Dead did not tap")
	}
}

// --- Consume Strength --------------------------------------------------------------------

func TestApcB1ConsumeStrengthPumpsOneAndShrinksAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	if err := castCatalogSpellErr(t, g, "Consume Strength", "Instant", apcB1ConsumeOracle, cardRefs(mine, mine)); err == nil {
		t.Fatal("the same creature was accepted as both targets")
	}
	castCatalogSpell(t, g, "Consume Strength", "Instant", apcB1ConsumeOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("the first target is %d/%d, want 4/4", p, tough)
	}
	if battlefieldHas(g, theirs) {
		t.Error("the second target survived -2/-2")
	}
}

// --- Fervent Charge ----------------------------------------------------------------------

func TestApcB1FerventChargePumpsEachCreatureYouControlThatAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	apcB1Push(g, me.ID, "Fervent Charge", "Enchantment", apcB1ChargeOracle, "{1}{R}{W}{B}", 0, 0)
	a := seedBear(g, me.ID)
	b := seedBear(g, me.ID)
	home := seedBear(g, me.ID)
	declareAttack(t, g, them.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("an attacker is %d/%d, want 4/4", p, tough)
		}
	}
	if p, tough := ptOf(t, g, home); p != 2 || tough != 2 {
		t.Errorf("a creature that did not attack is %d/%d, want 2/2", p, tough)
	}
}

// --- Flowstone Charger -------------------------------------------------------------------

func TestApcB1FlowstoneChargerGetsPlusThreeMinusThreeWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	charger := apcB1Push(g, me.ID, "Flowstone Charger", "Creature — Beast", apcB1ChargerOracle, "{2}{R}{W}", 2, 5)
	declareAttack(t, g, them.ID, charger)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, charger); p != 5 || tough != 2 {
		t.Errorf("Flowstone Charger is %d/%d while attacking, want 5/2", p, tough)
	}
}

// --- Prophetic Bolt ----------------------------------------------------------------------

func TestApcB1PropheticBoltDealsFourThenTakesOneOfFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seeded := seedLibrary(me, "Look1", "Look2", "Look3", "Look4", "Filler")
	before := opp.Life
	castCatalogSpell(t, g, "Prophetic Bolt", "Instant", apcB1BoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 4 {
		t.Errorf("the target lost %d life, want 4", got)
	}
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no prompt to put one of the four into your hand")
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{seeded[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(seeded[2]) {
		t.Fatal("the chosen card did not reach the hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
		t.Fatalf("want a bottom placement over the three declined cards, got %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{seeded[3], seeded[0], seeded[1]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryBottomIDs(me, 3); !sameIDs(got, []uuid.UUID{seeded[3], seeded[0], seeded[1]}) {
		t.Errorf("the bottom three are %v, want them in the order chosen", got)
	}
}

// --- Fire // Ice -------------------------------------------------------------------------

var apcB1FireIceFaces = [2]game.Face{
	{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}", Colors: []string{"R"}},
	{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}", Colors: []string{"U"}},
}

func TestApcB1FireDividesTwoDamage(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	apcB1CastHalf(t, g, apcB1FireIceOracle, apcB1FireIceFaces, game.CastSpellParams{
		Face: 0, Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	passPriorityAroundTable(t, g)
	if da, db := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; da != 1 || db != 1 {
		t.Errorf("damage split %d/%d, want 1/1", da, db)
	}
}

func TestApcB1IceTapsTargetPermanentAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top")
	land := apcB1Push(g, opp.ID, "Their Land", "Land", "", "", 0, 0)
	apcB1CastHalf(t, g, apcB1FireIceOracle, apcB1FireIceFaces, game.CastSpellParams{Face: 1, Targets: cardTarget(land)})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !findCardAnywhere(t, g, land).Tapped {
		t.Error("the target permanent was not tapped")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Night // Day ------------------------------------------------------------------------

var apcB1NightDayFaces = [2]game.Face{
	{Name: "Night", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}},
	{Name: "Day", TypeLine: "Instant", ManaCost: "{2}{W}", Colors: []string{"W"}},
}

func TestApcB1NightShrinksTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	bear := seedBear(g, g.Seats[1].ID)
	apcB1CastHalf(t, g, apcB1NightDayOracle, apcB1NightDayFaces, game.CastSpellParams{Face: 0, Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}
}

func TestApcB1DayPumpsTheTargetPlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := seedBear(g, me.ID)
	b := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	apcB1CastHalf(t, g, apcB1NightDayOracle, apcB1NightDayFaces, game.CastSpellParams{
		Face: 1, Targets: []game.TargetRef{playerTarget(me.ID)},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("the target player's creature is %d/%d, want 3/3", p, tough)
		}
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("another player's creature is %d/%d, want 2/2", p, tough)
	}
}

// --- Suffocating Blast -------------------------------------------------------------------

func TestApcB1SuffocatingBlastCountersAndDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shock := apcB1OpponentShock(t, g, opp, me.ID)
	ogre := apcB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "", 3, 3)
	before := me.Life
	castCatalogSpell(t, g, "Suffocating Blast", "Instant", apcB1BlastOracle, cardRefs(shock, ogre))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(shock) || me.Life != before {
		t.Errorf("the Shock was not countered: in graveyard %v, life %d → %d", opp.Graveyard.Contains(shock), before, me.Life)
	}
	if battlefieldHas(g, ogre) {
		t.Error("the 3/3 survived 3 damage")
	}
}

func TestApcB1SuffocatingBlastStillDealsDamageWhenTheSpellIsGone(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	shock := apcB1OpponentShock(t, g, opp, me.ID)
	ogre := apcB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "", 3, 3)
	castCatalogSpell(t, g, "Suffocating Blast", "Instant", apcB1BlastOracle, cardRefs(shock, ogre))
	// The Shock leaves the stack before the Blast resolves.
	if err := g.CounterSpell(shock, nil); err != nil {
		t.Fatalf("counter the Shock: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, ogre) {
		t.Error("the creature half did not happen once the spell was gone")
	}
}

// --- Ebony Treefolk ----------------------------------------------------------------------

func TestApcB1EbonyTreefolkPumpsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	tree := apcB1Push(g, me.ID, "Ebony Treefolk", "Creature — Treefolk", apcB1TreefolkOracle, "{1}{B}{G}", 3, 3)
	dcMana(t, g, me.ID, "{B}{G}")
	if err := invB1Activate(g, me.ID, tree, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, tree); p != 4 || tough != 4 {
		t.Errorf("Ebony Treefolk is %d/%d, want 4/4", p, tough)
	}
}

// --- Pernicious Deed ---------------------------------------------------------------------

func TestApcB1PerniciousDeedDestroysArtifactsCreaturesAndEnchantmentsUpToX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	deed := apcB1Push(g, me.ID, "Pernicious Deed", "Enchantment", apcB1DeedOracle, "{1}{B}{G}", 0, 0)
	bear := apcB1Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	mine := apcB1Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	relic := apcB1Push(g, opp.ID, "Relic", "Artifact", "", "{1}", 0, 0)
	aura := apcB1Push(g, opp.ID, "Ward", "Enchantment", "", "{1}{W}", 0, 0)
	big := apcB1Push(g, opp.ID, "Big", "Creature — Giant", "", "{4}{R}", 5, 5)
	pricey := apcB1Push(g, opp.ID, "Pricey", "Artifact", "", "{3}", 0, 0)
	land := apcB1Push(g, opp.ID, "Land", "Land", "", "", 0, 0)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, deed, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(deed) {
		t.Fatal("Pernicious Deed was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"an opponent's 2-drop creature": bear, "your own 2-drop creature": mine, "a 1-drop artifact": relic, "a 2-drop enchantment": aura} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived X=2", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"a 5-drop creature": big, "a 3-drop artifact": pricey, "a land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed at X=2", name)
		}
	}
}
