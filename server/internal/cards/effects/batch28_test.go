package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch28_test.go — card-level coverage for the card-coverage
// roadmap's batch 28 (#390, `edhrec_rank` 2940–3040): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, block, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b28-prefixed.

const (
	b28AbradedBluffsOracle         = "ca7d093c-0533-493f-9ad3-8af30118fbfc"
	b28DoppelgangOracle            = "d04c6375-25dd-4882-b8b3-a3b0d3081f32"
	b28ReviveTheShireOracle        = "67afcfef-3f0c-4823-b690-01732f28a005"
	b28UncivilUnrestOracle         = "bd655e8b-f192-4635-9e23-357b6f89ef8f"
	b28NykthosParagonOracle        = "367ac7e2-5056-48f3-a296-87d6cb1f7b54"
	b28BenevolentHydraOracle       = "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0"
	b28PhyrexianWalkerOracle       = "7af75024-6c9b-4844-aeb6-81de25464822"
	b28HinterlandSanctifierOracle  = "d812fc6d-b96d-4986-b171-9f3feee603dc"
	b28UlvenwaldTrackerOracle      = "63a37c08-7e38-4fe2-9a71-cc4604c4f831"
	b28CauldronFamiliarOracle      = "a7dc2e62-1c50-4ed7-b71f-2d782a447a5e"
	b28DragonstormGlobeOracle      = "f6de5bd7-7704-4a0c-a27a-9e565e49f5e9"
	b28RepercussionOracle          = "f6069a4f-744e-43e8-9f8b-7c13b8b0187f"
	b28FatedFirepowerOracle        = "13daa21c-278d-45bd-9a6e-a77d6a558453"
	b28RegalCaracalOracle          = "aa571676-b390-4b8a-baea-4c4cd60c01f6"
	b28TiamatOracle                = "f00e4df1-13fb-4514-8abb-92954068689a"
	b28ArcanisOracle               = "8a7183cc-161c-444d-a889-a17519c8061b"
	b28MizziumMortarsOracle        = "48ddba1e-2ad7-463f-9307-d2379a800e51"
	b28BeseechTheQueenAlreadyOID   = "cf94cafc-527e-4b27-8a28-7807435aaccf"
	b28KykarOracle                 = "71a80491-eea2-4661-9446-26a87efbf4a8"
	b28RavenousSquirrelOracle      = "8fb48d65-406b-4231-83ea-9cf3bb8dff76"
	b28KarametrasAcolyteOracle     = "e8cf80a1-6908-4adb-88e7-2015672d4905"
	b28ThunderbreakRegentOracle    = "bb04f927-0348-4a14-9e74-381f18083f6a"
	b28ChivalricAllianceOracle     = "9b8d9236-0452-4509-aca4-a53a399fff85"
	b28MarchOfTheMultitudesOracle  = "0c26ab0d-80f6-4e5b-9d0e-af17c1519583"
	b28BrimazOracle                = "49471b65-be9d-4ded-b2e5-dfc81b306b54"
	b28StromkirkCaptainOracle      = "599c3fd6-3309-4b5d-adec-9c4062848ad5"
	b28LifebloodHydraOracle        = "b14d05c0-fe10-4079-a90e-0aea1a8fd375"
	b28SpawnbedProtectorOracle     = "256a7f54-0e8e-4d22-a0f7-ff2830e8884e"
	b28RampagingYaoGuaiSkipOracle  = "d37e8f75-c7cb-4270-bb2f-0125e972ed15"
	b28TectonicGiantSkipOracle     = "0e5e46e3-f9af-47cc-a740-bf808d5cb4b1"
	b28HallOfTheBanditLordSkipOID  = "32fe7ac4-86f5-44af-9f73-ee8f6a9ce2ba"
	b28HoldoutSettlementFullOracle = "e6b77545-de5c-4f4a-b7ea-83498fb33ba8"
	b28FertilidSkipOracle          = "21f1c6d7-8289-44b2-b88f-c09e202be200"
	b28MeathookMassacreIISkipOID   = "68957dca-df5c-4051-af22-407cb7e47200"
	b28TimeStretchSkipOracle       = "72e56963-a9dd-44dc-a4d3-992b4d89dd28"
	b28LoyalGuardianSkipOracle     = "de1cfa83-2049-48ab-b093-2f120a3ef4f2"
)

// b28Lives snapshots every seat's life total.
func b28Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b28TokensNamed counts the battlefield tokens `controller` controls
// with the given name.
func b28TokensNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name && IsToken(c) {
			n++
		}
	}
	return n
}

// b28Push seeds a catalog permanent with a mana cost and colours.
func b28Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b28AddCounter puts n +1/+1 counters on a permanent through the
// effect path, so replacements see it.
func b28AddCounter(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, "+1/+1", n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

// b28TapForMana activates a permanent's first mana ability and
// answers every colour pick it opened with `color`.
func b28TapForMana(t *testing.T, g *game.Game, controller, card uuid.UUID, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if color != "" {
		b10ResolveAllManaPicks(t, g, controller, color)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Beseech the
// Queen was already on main from the tutor work and is pinned here
// so the table matches the issue's 37 minus the nine declared skips.
func TestBatch28CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b28AbradedBluffsOracle:        "Abraded Bluffs",
		b28DoppelgangOracle:           "Doppelgang",
		b28ReviveTheShireOracle:       "Revive the Shire",
		b28UncivilUnrestOracle:        "Uncivil Unrest",
		b28NykthosParagonOracle:       "Nykthos Paragon",
		b28BenevolentHydraOracle:      "Benevolent Hydra",
		b28PhyrexianWalkerOracle:      "Phyrexian Walker",
		b28HinterlandSanctifierOracle: "Hinterland Sanctifier",
		b28UlvenwaldTrackerOracle:     "Ulvenwald Tracker",
		b28CauldronFamiliarOracle:     "Cauldron Familiar",
		b28DragonstormGlobeOracle:     "Dragonstorm Globe",
		b28RepercussionOracle:         "Repercussion",
		b28FatedFirepowerOracle:       "Fated Firepower",
		b28RegalCaracalOracle:         "Regal Caracal",
		b28TiamatOracle:               "Tiamat",
		b28ArcanisOracle:              "Arcanis the Omnipotent",
		b28MizziumMortarsOracle:       "Mizzium Mortars",
		b28BeseechTheQueenAlreadyOID:  "Beseech the Queen",
		b28KykarOracle:                "Kykar, Wind's Fury",
		b28RavenousSquirrelOracle:     "Ravenous Squirrel",
		b28KarametrasAcolyteOracle:    "Karametra's Acolyte",
		b28ThunderbreakRegentOracle:   "Thunderbreak Regent",
		b28ChivalricAllianceOracle:    "Chivalric Alliance",
		b28MarchOfTheMultitudesOracle: "March of the Multitudes",
		b28BrimazOracle:               "Brimaz, King of Oreskos",
		b28StromkirkCaptainOracle:     "Stromkirk Captain",
		b28LifebloodHydraOracle:       "Lifeblood Hydra",
		b28SpawnbedProtectorOracle:    "Spawnbed Protector",
	}
	// #758 moves this former declared skip into the completed set.
	want[b28HoldoutSettlementFullOracle] = "Holdout Settlement"
	if len(want) != 29 {
		t.Fatalf("the batch registers 28 cards plus one already on main, the table lists %d", len(want))
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
	}
	// The seven declared skips must NOT be registered — each needs a
	// seam the engine does not have, and a spec would ship the card
	// stronger than printed or as something other than itself.
	// Bribery was the ninth until #1230 gave the search primitive a
	// LibraryOwner parameter; it is registered now (bribery.go) and
	// dropped from this list.
	for _, skipped := range []string{
		b28RampagingYaoGuaiSkipOracle, // an enters trigger cannot read the spell's X
		b28TectonicGiantSkipOracle,    // a modal triggered ability
		b28HallOfTheBanditLordSkipOID, // a "if that mana is spent on" rider
		b28FertilidSkipOracle,         // counter-removal cost
		b28MeathookMassacreIISkipOID,  // an opponent's life-payment choice with the consequence on decline; finality counters
		b28TimeStretchSkipOracle,      // extra turns
		b28LoyalGuardianSkipOracle,    // a beginning-of-combat trigger event
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- lands and vanilla ---------------------------------------------

func TestB28AbradedBluffsEntersTappedPingsAnOpponentAndTapsForEitherColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := b12PlayFromHand(t, g, "Abraded Bluffs", "Land — Desert", b28AbradedBluffsOracle, game.CastSpellParams{})
	if !b12Card(t, g, land).Tapped {
		t.Fatal("the Desert enters tapped")
	}
	if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

func TestB28PhyrexianWalkerIsAVanillaZeroDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Phyrexian Walker", "Artifact Creature — Phyrexian Construct", b28PhyrexianWalkerOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Walker resolves to the battlefield")
	}
	if c := b12Card(t, g, id); c.Controller != me.ID {
		t.Error("under its caster's control")
	}
	if spec, _ := Lookup(b28PhyrexianWalkerOracle); spec.Completeness != CompletenessFull {
		t.Error("a vanilla creature is complete")
	}
}

// --- the spells ----------------------------------------------------

func TestB28ReviveTheShireReturnsAPermanentCardAndMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	bolt := b17GraveyardCard(me, "Spent Bolt", "Instant", "{R}")
	if legal := legalCards(g, me.ID, b28ReviveTheShireOracle); !legal[bear] || legal[bolt] {
		t.Fatal("a permanent card in your graveyard is legal; an instant is not")
	}
	castCatalogSpell(t, g, "Revive the Shire", "Sorcery", b28ReviveTheShireOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the Bear is back in hand")
	}
	if got := b28TokensNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("one Food: %d", got)
	}
}

func TestB28DoppelgangMakesXCopiesOfEachOfXTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	extra := b12Permanent(g, opp.ID, "Their Idol", "Artifact")
	// Three targets on an X of two is refused: the count IS X.
	active := g.Seats[g.Turn.ActiveSeat]
	bad := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: bad, Name: "Doppelgang", TypeLine: "Sorcery", OracleID: b28DoppelgangOracle,
		ManaCost: "{X}{X}{X}{G}{U}", Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, bad, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: extra},
	}}); err == nil {
		t.Fatal("X target permanents with X=2 refuses three")
	}
	castXSpell(t, g, "Doppelgang", "Sorcery", b28DoppelgangOracle, "{X}{X}{X}{G}{U}", 2, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock},
	})
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "My Bear"); got != 2 {
		t.Errorf("two token Bears: %d", got)
	}
	if got := b28TokensNamed(g, me.ID, "Their Rock"); got != 2 {
		t.Errorf("two token Rocks, under the caster's control: %d", got)
	}
	if got := b28TokensNamed(g, opp.ID, "Their Rock"); got != 0 {
		t.Errorf("none for the opponent: %d", got)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("the originals stay")
	}
}

func TestB28MizziumMortarsTargetedAndOverloaded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	huge := b12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 6, 6)
	if legal := legalCards(g, me.ID, b28MizziumMortarsOracle); legal[mine] || !legal[big] {
		t.Fatal("only creatures you don't control are legal targets")
	}
	castCatalogSpell(t, g, "Mizzium Mortars", "Sorcery", b28MizziumMortarsOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("4 damage kills the 4/4")
	}
	if c := b12Card(t, g, huge); c.DamageMarked != 0 {
		t.Error("the other creature is untouched by the targeted cast")
	}
	small := b12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	castWithAltCost(t, g, "Mizzium Mortars", "Sorcery", b28MizziumMortarsOracle, "overload")
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("overloaded: each creature you don't control takes 4")
	}
	if c := b12Card(t, g, huge); c.DamageMarked != 4 {
		t.Errorf("the 6/6 takes 4 and lives: %d marked", c.DamageMarked)
	}
	if c := b12Card(t, g, mine); c.DamageMarked != 0 {
		t.Error("your own creatures are never in the sweep")
	}
}

func TestB28MarchOfTheMultitudesMakesXLifelinkSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "March of the Multitudes", "Instant", b28MarchOfTheMultitudesOracle, "{X}{G}{W}{W}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Soldier"); got != 3 {
		t.Fatalf("three Soldiers: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && IsToken(c) {
			assertKeywords(t, g, c.InstanceID, "lifelink")
		}
	}
	if spec, _ := Lookup(b28MarchOfTheMultitudesOracle); spec.TapCost == nil {
		t.Error("convoke is declared")
	}
}

// --- enchantments --------------------------------------------------

func TestB28UncivilUnrestRiotCounterAndDoubledDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Uncivil Unrest", "Enchantment", b28UncivilUnrestOracle, 0, 0)
	bear := castCatalogSpell(t, g, "Riot Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("riot: a nontoken creature enters with a +1/+1 counter: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && IsToken(c) && c.Counters["+1/+1"] != 0 {
			t.Error("a token has no riot")
		}
	}
	// A countered attacker deals double; one without a counter does not.
	striker := pushVanillaCreature(g, me.ID, "Striker", 2, 2)
	b28AddCounter(t, g, striker, 1)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	before := opp.Life
	attackWith(t, g, opp.ID, striker, plain)
	if opp.Life != before-6-2 {
		t.Errorf("the 3/3 with a counter deals 6, the 2/2 deals 2: %d → %d", before, opp.Life)
	}
}

func TestB28FatedFirepowerAddsFireCountersToDamageAtOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 6)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	fp := castXSpell(t, g, "Fated Firepower", "Enchantment", b28FatedFirepowerOracle, "{X}{R}{R}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, fp, "fire"); got != 2 {
		t.Fatalf("enters with X fire counters: %d", got)
	}
	before := opp.Life
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-5 {
		t.Errorf("a Bolt at an opponent deals 3+2: %d → %d", before, opp.Life)
	}
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if c := b12Card(t, g, theirs); c.DamageMarked != 5 {
		t.Errorf("a Bolt at their permanent deals 3+2: %d marked", c.DamageMarked)
	}
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	if c := b12Card(t, g, mine); c.DamageMarked != 3 {
		t.Errorf("your own permanent takes the printed 3: %d marked", c.DamageMarked)
	}
	// An opponent's Bolt at you is not a source you control.
	mine2 := me.Life
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	passPriorityAroundTable(t, g)
	if me.Life != mine2-3 {
		t.Errorf("their Bolt deals the printed 3: %d → %d", mine2, me.Life)
	}
	assertKeywords(t, g, fp, "flash")
}

func TestB28RepercussionPunishesTheDamagedCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Repercussion", "Enchantment", b28RepercussionOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	mine := b12Creature(g, me.ID, "My Golem", "Creature — Golem", 4, 4)
	before := b28Lives(g)
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if opp.Life != before[1]-3 {
		t.Errorf("their creature took 3, so they take 3: %d → %d", before[1], opp.Life)
	}
	if me.Life != before[0] {
		t.Error("nothing for the caster")
	}
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	if me.Life != before[0]-3 {
		t.Errorf("your own creature dealt damage hits you too: %d → %d", before[0], me.Life)
	}
	// Damage to a player is not a creature being dealt damage.
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before[1]-3-3 {
		t.Errorf("a Bolt to the face is just the Bolt: %d", opp.Life)
	}
}

func TestB28ChivalricAllianceDrawsOnATwoCreatureAttackOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Chivalric Alliance", "Enchantment", b28ChivalricAllianceOracle, 0, 0)
	a := pushVanillaCreature(g, me.ID, "Knight A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Knight B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "Knight C", 2, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a, b, c)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("three attackers is one draw: %d → %d", hand, me.Hand.Size())
	}
	// A lone attacker next turn draws nothing.
	b12ToMyNextUpkeep(t, g)
	advanceToMain(t, g)
	hand = me.Hand.Size()
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("one attacker is no draw: %d → %d", hand, me.Hand.Size())
	}
	// #1381: the discard-cost Knight-making ability is registered now
	// (chivalric_alliance_test.go proves it works) and the card is
	// complete.
	if spec, _ := Lookup(b28ChivalricAllianceOracle); len(spec.Activated) != 1 || spec.Completeness != CompletenessFull {
		t.Error("the discard-cost ability should be registered and the card complete")
	}
}

// --- creatures -----------------------------------------------------

func TestB28HinterlandSanctifierGainsOnEachOtherCreatureYouControlEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Hinterland Sanctifier", "Creature — Rabbit Cleric", b28HinterlandSanctifierOracle, 1, 2)
	before := me.Life
	castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("a creature entering is 1 life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 2) })
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("two tokens are 2 more: %d", me.Life)
	}
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("an opponent's creature is not yours: %d", me.Life)
	}
}

func TestB28CauldronFamiliarDrainsEachOpponentForOne(t *testing.T) {
	g := newCatalogGame(t)
	before := b28Lives(g)
	castCatalogSpell(t, g, "Cauldron Familiar", "Creature — Cat", b28CauldronFamiliarOracle, nil)
	passPriorityAroundTable(t, g)
	want := []int{before[0] + 1, before[1] - 1, before[2] - 1, before[3] - 1}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want[i])
		}
	}
	// #1381: the sacrifice-a-Food graveyard ability is registered now
	// (cauldron_familiar_test.go proves it works) and the card is
	// complete.
	if spec, _ := Lookup(b28CauldronFamiliarOracle); len(spec.Activated) != 1 || spec.Completeness != CompletenessFull {
		t.Error("the graveyard ability should be registered and the card complete")
	}
}

func TestB28NykthosParagonOffersCountersOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paragon := b12Push(g, me.ID, "Nykthos Paragon", "Enchantment Creature — Human Soldier", b28NykthosParagonOracle, 4, 6)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	b14Gain(t, g, me.ID, 3)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("a life gain asks the question")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 3 {
		t.Errorf("that many counters on each creature you control: Bear has %d", got)
	}
	if got := counterCount(g, paragon, "+1/+1"); got != 3 {
		t.Errorf("the Paragon included: %d", got)
	}
	if got := counterCount(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("not on an opponent's: %d", got)
	}
	b14Gain(t, g, me.ID, 2)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("only once each turn")
	}
	if got := counterCount(g, bear, "+1/+1"); got != 3 {
		t.Errorf("no second placement this turn: %d", got)
	}
	// Next turn: declining leaves the once-per-turn unused.
	b12ToMyNextUpkeep(t, g)
	b14Gain(t, g, me.ID, 1)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	b14Gain(t, g, me.ID, 4)
	if latestTriggerPrompt(g, me.ID) == nil {
		t.Fatal("a declined offer does not use up the turn")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 7 {
		t.Errorf("3 then 4: %d", got)
	}
}

func TestB28BenevolentHydraEntersWithXAndGivesOthersAnExtraCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	hydra := castXSpell(t, g, "Benevolent Hydra", "Creature — Hydra", b28BenevolentHydraOracle, "{X}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("enters with X counters, its own bonus not applying to itself: %d", got)
	}
	b28AddCounter(t, g, bear, 1)
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("one counter on another creature you control is two: %d", got)
	}
	b28AddCounter(t, g, hydra, 1)
	if got := counterCount(g, hydra, "+1/+1"); got != 4 {
		t.Errorf("the Hydra itself gets the printed one: %d", got)
	}
	b28AddCounter(t, g, theirs, 1)
	if got := counterCount(g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature gets the printed one: %d", got)
	}
	// #625: the counter-moving tap ability ships; counter_cost_cards_test.go
	// drives it. The X-counter timing is still a declared caveat.
	if spec, _ := Lookup(b28BenevolentHydraOracle); len(spec.Activated) != 1 || spec.Completeness != CompletenessCaveats {
		t.Error("the tap ability ships, and the X-counter caveat is still declared")
	}
}

func TestB28UlvenwaldTrackerMakesYourCreatureFightAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tracker := b12Push(g, me.ID, "Ulvenwald Tracker", "Creature — Human Shaman", b28UlvenwaldTrackerOracle, 1, 1)
	mine := pushVanillaCreature(g, me.ID, "My Golem", 4, 4)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}})
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 kills the 2/2")
	}
	if c := b12Card(t, g, mine); c.DamageMarked != 2 {
		t.Errorf("the 2/2 hits back for 2: %d marked", c.DamageMarked)
	}
	if !b12Card(t, g, tracker).Tapped {
		t.Error("the Tracker tapped to activate")
	}
	// Their creature first is the wrong way round: nothing happens.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tracker) })
	other := pushVanillaCreature(g, opp.ID, "Their Ox", 3, 3)
	b16Activate(t, g, me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: other}, {Kind: game.TargetCard, ID: mine},
	}})
	if c := b12Card(t, g, other); c.DamageMarked != 0 {
		t.Error("slot 0 must be yours — a pair that doesn't fit does nothing")
	}
}

func TestB28DragonstormGlobeGrowsEnteringDragons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	globe := b12Push(g, me.ID, "Dragonstorm Globe", "Artifact", b28DragonstormGlobeOracle, 0, 0)
	dragon := castCatalogSpell(t, g, "My Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, dragon, "+1/+1"); got != 1 {
		t.Errorf("a Dragon enters with an additional counter: %d", got)
	}
	bear := castCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, "+1/+1"); got != 0 {
		t.Errorf("a Bear does not: %d", got)
	}
	theirs := b13PlayAs(t, g, 1, "Their Dragon", "Creature — Dragon", "")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's Dragon does not: %d", got)
	}
	b28TapForMana(t, g, me.ID, globe, "U")
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("any colour: pool %v", got)
	}
}

func TestB28RegalCaracalLordsCatsAndMakesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := b12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	theirCat := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	caracal := b20CastCreature(t, g, me, "Regal Caracal", "Creature — Cat", b28RegalCaracalOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Cat"); got != 2 {
		t.Fatalf("two Cat tokens: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && IsToken(c) {
			if p := effectivePower(t, g, c.InstanceID); p != 2 {
				t.Errorf("a 1/1 token Cat under the Caracal is 2/2: power %d", p)
			}
			assertKeywords(t, g, c.InstanceID, "lifelink")
		}
	}
	if p := effectivePower(t, g, cat); p != 3 {
		t.Errorf("another Cat you control gets +1/+1: %d", p)
	}
	assertKeywords(t, g, cat, "lifelink")
	if p := effectivePower(t, g, caracal); p != 3 {
		t.Errorf("the Caracal itself is unchanged: %d", p)
	}
	if p := effectivePower(t, g, theirCat); p != 2 {
		t.Errorf("an opponent's Cat is unchanged: %d", p)
	}
}

func TestB28StromkirkCaptainLordsOtherVampires(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := b12Push(g, me.ID, "Stromkirk Captain", "Creature — Vampire Soldier", b28StromkirkCaptainOracle, 2, 2)
	vamp := b12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 2, 2)
	human := b12Creature(g, me.ID, "My Human", "Creature — Human", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2)
	if p := effectivePower(t, g, vamp); p != 3 {
		t.Errorf("another Vampire you control is +1/+1: %d", p)
	}
	assertKeywords(t, g, vamp, "first strike")
	assertKeywords(t, g, captain, "first strike")
	if p := effectivePower(t, g, captain); p != 2 {
		t.Errorf("the Captain is 'other' — unchanged: %d", p)
	}
	if p := effectivePower(t, g, human); p != 2 {
		t.Errorf("a Human is unchanged: %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Vampire is unchanged: %d", p)
	}
}

func TestB28TiamatTutorsDifferentlyNamedDragonsOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := seedSearchLibrary(me,
		game.Card{Name: "Dragon A", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Dragon B", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Dragon A", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Tiamat", TypeLine: "Legendary Creature — Dragon God"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Tiamat", "Legendary Creature — Dragon God", b28TiamatOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("cast Tiamat searches")
	}
	if hasID(c.SearchCards, lib[3]) || hasID(c.SearchCards, lib[4]) {
		t.Error("a second Tiamat and a Bear are not offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lib[0], lib[2]}); err == nil {
		t.Fatal("two Dragons with the same name are refused")
	}
	answerSearchByID(t, g, me.ID, lib[0], lib[1])
	if !me.Hand.Contains(lib[0]) || !me.Hand.Contains(lib[1]) {
		t.Error("the differently named Dragons are in hand")
	}
	// A Tiamat that was not cast does not search.
	dead := b17GraveyardCard(me, "Tiamat", "Legendary Creature — Dragon God", "{2}{W}{U}{B}{R}{G}")
	g.Battlefield.Cards = nil
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(dead, game.ZoneBattlefield, me.ID)
	})
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("a reanimated Tiamat has no 'if you cast it'")
	}
}

func TestB28ArcanisDrawsThreeAndBouncesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arcanis := b12Push(g, me.ID, "Arcanis the Omnipotent", "Legendary Creature — Wizard", b28ArcanisOracle, 3, 4)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, arcanis, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+3 {
		t.Errorf("{T}: draw three: %d → %d", hand, me.Hand.Size())
	}
	if !b12Card(t, g, arcanis).Tapped {
		t.Error("tapped to draw")
	}
	b16Activate(t, g, me.ID, arcanis, 1, game.ActivateAbilityParams{})
	if !me.Hand.Contains(arcanis) || g.Battlefield.Contains(arcanis) {
		t.Error("{2}{U}{U}: back to its owner's hand")
	}
}

func TestB28KykarMakesSpiritsOnNoncreatureSpellsAndEatsThemForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kykar := b12Push(g, me.ID, "Kykar, Wind's Fury", "Legendary Creature — Bird Wizard", b28KykarOracle, 3, 3)
	castCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Fatalf("a noncreature spell makes a Spirit: %d", got)
	}
	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("a creature spell does not: %d", got)
	}
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && IsToken(c) {
			spirit = c.InstanceID
			assertKeywords(t, g, c.InstanceID, "flying")
		}
	}
	if err := g.ActivateManaAbility(me.ID, kykar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{spirit}}); err != nil {
		t.Fatalf("Sacrifice a Spirit: Add {R}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("{R} in the pool: %v", got)
	}
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit was sacrificed")
	}
	bear := findBattlefieldByName(g, "Some Bear")
	if err := g.ActivateManaAbility(me.ID, kykar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear is not a Spirit")
	}
}

func TestB28RavenousSquirrelGrowsOnSacrificeAndCashesIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	squirrel := b12Push(g, me.ID, "Ravenous Squirrel", "Creature — Squirrel", b28RavenousSquirrelOracle, 1, 1)
	food := b12Permanent(g, me.ID, "Food", "Artifact — Food")
	land := seedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, squirrel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{food}})
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed as the cost")
	}
	if got := counterCount(g, squirrel, "+1/+1"); got != 1 {
		t.Errorf("sacrificing an artifact grows the Squirrel: %d", got)
	}
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("1 life and a card: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	if err := g.ActivateCatalogAbility(me.ID, squirrel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land is neither an artifact nor a creature")
	}
	// Sacrificing a creature to anything else grows it too.
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, squirrel, "+1/+1"); got != 2 {
		t.Errorf("a creature sacrificed elsewhere counts: %d", got)
	}
}

func TestB28KarametrasAcolyteTapsForDevotionToGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	acolyte := b28Push(g, me.ID, "Karametra's Acolyte", "Creature — Human Druid", b28KarametrasAcolyteOracle, "{3}{G}", 1, 4, "G")
	b28Push(g, me.ID, "Llanowar Visionary", "Creature — Elf Druid", "", "{2}{G}{G}", 2, 2, "G")
	b28Push(g, me.ID, "Sol Ring", "Artifact", "", "{1}", 0, 0)
	b28Push(g, g.Seats[1].ID, "Their Elf", "Creature — Elf", "", "{G}", 1, 1, "G")
	b28TapForMana(t, g, me.ID, acolyte, "")
	if got := poolColors(me); len(got) != 3 {
		t.Errorf("devotion to green is 3 (one from the Acolyte, two from the Visionary): pool %v", got)
	}
	for _, c := range poolColors(me) {
		if c != "G" {
			t.Errorf("all green: %v", poolColors(me))
		}
	}
}

func TestB28ThunderbreakRegentPunishesOpponentsTargetingYourDragons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	regent := b12Push(g, me.ID, "Thunderbreak Regent", "Creature — Dragon", b28ThunderbreakRegentOracle, 4, 4)
	dragon := b12Creature(g, me.ID, "My Dragon", "Creature — Dragon", 5, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: dragon})
	if triggerOnStack(g, regent) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("targeting your Dragon triggers the Regent")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the targeting opponent takes 3: %d → %d", before, opp.Life)
	}
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: regent})
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("the Regent itself is a Dragon: %d", opp.Life)
	}
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("a Bear is not a Dragon: %d", opp.Life)
	}
	// Your own spell at your own Dragon is not an opponent's.
	mine := me.Life
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: dragon})
	if me.Life != mine || opp.Life != before-6 {
		t.Error("only an opponent's spell or ability triggers it")
	}
}

func TestB28BrimazMakesAnAttackingCatAndABlockingOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brimaz := b12Push(g, me.ID, "Brimaz, King of Oreskos", "Legendary Creature — Cat Soldier", b28BrimazOracle, 3, 4)
	assertKeywords(t, g, brimaz, "vigilance")
	before := opp.Life
	declareAttack(t, g, opp.ID, brimaz)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Cat Soldier"); got != 1 {
		t.Fatalf("one Cat Soldier: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat Soldier" && IsToken(c) {
			if c.AttackingTarget != opp.ID {
				t.Error("the token is attacking the player Brimaz attacks")
			}
			assertKeywords(t, g, c.InstanceID, "vigilance")
		}
	}
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != before-4 {
		t.Errorf("Brimaz and the token connect for 3+1: %d → %d", before, opp.Life)
	}
	// Their turn: Brimaz blocks and makes another.
	aangAdvanceToMain(t, g, 1)
	raider := pushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(brimaz, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// #830: the declaration announces at its lock-in, not at the
	// click, so the block trigger needs the priority wrap first.
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Cat Soldier"); got != 2 {
		t.Errorf("a second Cat Soldier for the block: %d", got)
	}
	if spec, _ := Lookup(b28BrimazOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the not-blocking token is a declared gap")
	}
}

func TestB28LifebloodHydraPaysOutItsPowerOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := castXSpell(t, g, "Lifeblood Hydra", "Creature — Hydra", b28LifebloodHydraOracle, "{X}{G}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("enters with X counters: %d", got)
	}
	assertKeywords(t, g, hydra, "trample")
	life, hand := me.Life, me.Hand.Size()
	b12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: hydra})
	if g.Battlefield.Contains(hydra) {
		t.Fatal("3 damage kills the 3/3")
	}
	if me.Life != life+3 || me.Hand.Size() != hand+3 {
		t.Errorf("gain 3 and draw 3: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
}

func TestB28SpawnbedProtectorReturnsAnEldraziAndMakesScions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Spawnbed Protector", "Creature — Eldrazi", b28SpawnbedProtectorOracle, 6, 8)
	eldrazi := b17GraveyardCard(me, "Dead Eldrazi", "Creature — Eldrazi", "{5}")
	bear := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("with an Eldrazi creature card in the graveyard the end step asks for a target")
	}
	if hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, eldrazi) {
		t.Error("only Eldrazi creature cards are offered")
	}
	pickCard(t, g, me.ID, eldrazi)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(eldrazi) {
		t.Error("the Eldrazi is back in hand")
	}
	if got := b28TokensNamed(g, me.ID, "Eldrazi Scion"); got != 2 {
		t.Errorf("two Scions: %d", got)
	}
	// Without an Eldrazi card the untargeted declaration still makes Scions.
	b12ToMyNextUpkeep(t, g)
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("nothing to target, nothing to ask")
	}
	passPriorityAroundTable(t, g)
	if got := b28TokensNamed(g, me.ID, "Eldrazi Scion"); got != 4 {
		t.Errorf("two more Scions with no target: %d", got)
	}
	var scion uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi Scion" && IsToken(c) {
			scion = c.InstanceID
		}
	}
	if err := g.ActivateManaAbility(me.ID, scion, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Sacrifice this token: Add {C}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("{C} in the pool: %v", got)
	}
}
