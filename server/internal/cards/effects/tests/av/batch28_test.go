package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b28Push seeds a catalog permanent with a mana cost and colours.
func b28Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Beseech the
// Queen was already on main from the tutor work and is pinned here
// so the table matches the issue's 37 minus the nine declared skips.
func TestBatch28CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B28AbradedBluffsOracle:        "Abraded Bluffs",
		effectstest.B28DoppelgangOracle:           "Doppelgang",
		effectstest.B28ReviveTheShireOracle:       "Revive the Shire",
		effectstest.B28UncivilUnrestOracle:        "Uncivil Unrest",
		effectstest.B28NykthosParagonOracle:       "Nykthos Paragon",
		effectstest.B28BenevolentHydraOracle:      "Benevolent Hydra",
		effectstest.B28PhyrexianWalkerOracle:      "Phyrexian Walker",
		effectstest.B28HinterlandSanctifierOracle: "Hinterland Sanctifier",
		effectstest.B28UlvenwaldTrackerOracle:     "Ulvenwald Tracker",
		effectstest.B28CauldronFamiliarOracle:     "Cauldron Familiar",
		effectstest.B28DragonstormGlobeOracle:     "Dragonstorm Globe",
		effectstest.B28RepercussionOracle:         "Repercussion",
		effectstest.B28FatedFirepowerOracle:       "Fated Firepower",
		effectstest.B28RegalCaracalOracle:         "Regal Caracal",
		effectstest.B28TiamatOracle:               "Tiamat",
		effectstest.B28ArcanisOracle:              "Arcanis the Omnipotent",
		effectstest.B28MizziumMortarsOracle:       "Mizzium Mortars",
		effectstest.B28BeseechTheQueenAlreadyOID:  "Beseech the Queen",
		effectstest.B28KykarOracle:                "Kykar, Wind's Fury",
		effectstest.B28RavenousSquirrelOracle:     "Ravenous Squirrel",
		effectstest.B28KarametrasAcolyteOracle:    "Karametra's Acolyte",
		effectstest.B28ThunderbreakRegentOracle:   "Thunderbreak Regent",
		effectstest.B28ChivalricAllianceOracle:    "Chivalric Alliance",
		effectstest.B28MarchOfTheMultitudesOracle: "March of the Multitudes",
		effectstest.B28BrimazOracle:               "Brimaz, King of Oreskos",
		effectstest.B28StromkirkCaptainOracle:     "Stromkirk Captain",
		effectstest.B28LifebloodHydraOracle:       "Lifeblood Hydra",
		effectstest.B28SpawnbedProtectorOracle:    "Spawnbed Protector",
	}
	// #758 moves this former declared skip into the completed set.
	want[effectstest.B28HoldoutSettlementFullOracle] = "Holdout Settlement"
	// #4380 uses the existing counter cost and target-player search seams.
	want[effectstest.B28FertilidOracle] = "Fertilid"
	if len(want) != 30 {
		t.Fatalf("the completed batch table should have 30 cards, it lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
	// The remaining declared skips must NOT be registered — each needs a
	// seam the engine does not have, and a spec would ship the card
	// stronger than printed or as something other than itself.
	// Bribery was the ninth until #1230 gave the search primitive a
	// LibraryOwner parameter; it is registered now (bribery.go) and
	// dropped from this list. Hall of the Bandit Lord was the seventh
	// until #1547 gave mana spend riders (hall_of_the_bandit_lord.go).
	// Rampaging Yao Guai was the last of them and came off with the GAP
	// batch 63 (#5948): an XFromCast enters trigger's "any number of
	// targets with total mana value X or less" (rampaging_yao_guai.go).
}

// --- lands and vanilla ---------------------------------------------

func TestB28AbradedBluffsEntersTappedPingsAnOpponentAndTapsForEitherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := effectstest.B12PlayFromHand(t, g, "Abraded Bluffs", "Land — Desert", effectstest.B28AbradedBluffsOracle, game.CastSpellParams{})
	if !effectstest.B12Card(t, g, land).Tapped {
		t.Fatal("the Desert enters tapped")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p == nil || effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "W")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

func TestB28PhyrexianWalkerIsAVanillaZeroDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Phyrexian Walker", "Artifact Creature — Phyrexian Construct", effectstest.B28PhyrexianWalkerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Walker resolves to the battlefield")
	}
	if c := effectstest.B12Card(t, g, id); c.Controller != me.ID {
		t.Error("under its caster's control")
	}
	if spec, _ := effects.Lookup(effectstest.B28PhyrexianWalkerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("a vanilla creature is complete")
	}
}

// --- the spells ----------------------------------------------------

func TestB28ReviveTheShireReturnsAPermanentCardAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	bolt := effectstest.B17GraveyardCard(me, "Spent Bolt", "Instant", "{R}")
	if legal := effectstest.LegalCards(g, me.ID, effectstest.B28ReviveTheShireOracle); !legal[bear] || legal[bolt] {
		t.Fatal("a permanent card in your graveyard is legal; an instant is not")
	}
	effectstest.CastCatalogSpell(t, g, "Revive the Shire", "Sorcery", effectstest.B28ReviveTheShireOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the Bear is back in hand")
	}
	if got := effectstest.B28TokensNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("one Food: %d", got)
	}
}

func TestB28DoppelgangMakesXCopiesOfEachOfXTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	extra := effectstest.B12Permanent(g, opp.ID, "Their Idol", "Artifact")
	// Three targets on an X of two is refused: the count IS X.
	active := g.Seats[g.Turn.ActiveSeat]
	bad := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: bad, Name: "Doppelgang", TypeLine: "Sorcery", OracleID: effectstest.B28DoppelgangOracle,
		ManaCost: "{X}{X}{X}{G}{U}", Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, bad, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: extra},
	}}); err == nil {
		t.Fatal("X target permanents with X=2 refuses three")
	}
	effectstest.CastXSpell(t, g, "Doppelgang", "Sorcery", effectstest.B28DoppelgangOracle, "{X}{X}{X}{G}{U}", 2, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: rock},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "My Bear"); got != 2 {
		t.Errorf("two token Bears: %d", got)
	}
	if got := effectstest.B28TokensNamed(g, me.ID, "Their Rock"); got != 2 {
		t.Errorf("two token Rocks, under the caster's control: %d", got)
	}
	if got := effectstest.B28TokensNamed(g, opp.ID, "Their Rock"); got != 0 {
		t.Errorf("none for the opponent: %d", got)
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(rock) {
		t.Error("the originals stay")
	}
}

func TestB28MizziumMortarsTargetedAndOverloaded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	huge := effectstest.B12Creature(g, opp.ID, "Their Titan", "Creature — Giant", 6, 6)
	if legal := effectstest.LegalCards(g, me.ID, effectstest.B28MizziumMortarsOracle); legal[mine] || !legal[big] {
		t.Fatal("only creatures you don't control are legal targets")
	}
	effectstest.CastCatalogSpell(t, g, "Mizzium Mortars", "Sorcery", effectstest.B28MizziumMortarsOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("4 damage kills the 4/4")
	}
	if c := effectstest.B12Card(t, g, huge); c.DamageMarked != 0 {
		t.Error("the other creature is untouched by the targeted cast")
	}
	small := effectstest.B12Creature(g, opp.ID, "Their Rat", "Creature — Rat", 1, 1)
	effectstest.CastWithAltCost(t, g, "Mizzium Mortars", "Sorcery", effectstest.B28MizziumMortarsOracle, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("overloaded: each creature you don't control takes 4")
	}
	if c := effectstest.B12Card(t, g, huge); c.DamageMarked != 4 {
		t.Errorf("the 6/6 takes 4 and lives: %d marked", c.DamageMarked)
	}
	if c := effectstest.B12Card(t, g, mine); c.DamageMarked != 0 {
		t.Error("your own creatures are never in the sweep")
	}
}

func TestB28MarchOfTheMultitudesMakesXLifelinkSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "March of the Multitudes", "Instant", effectstest.B28MarchOfTheMultitudesOracle, "{X}{G}{W}{W}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Soldier"); got != 3 {
		t.Fatalf("three Soldiers: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && effects.IsToken(c) {
			effectstest.AssertKeywords(t, g, c.InstanceID, "lifelink")
		}
	}
	if spec, _ := effects.Lookup(effectstest.B28MarchOfTheMultitudesOracle); spec.TapCost == nil {
		t.Error("convoke is declared")
	}
}

// --- enchantments --------------------------------------------------

func TestB28UncivilUnrestRiotCounterAndDoubledDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Uncivil Unrest", "Enchantment", effectstest.B28UncivilUnrestOracle, 0, 0)
	bear := effectstest.CastCatalogSpell(t, g, "Riot Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	// ADR 0109 §10: riot is asked as the creature enters; take the
	// counter the second ability wants.
	effectstest.AnswerRiotFor(t, g, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 1 {
		t.Errorf("riot: a nontoken creature enters with a +1/+1 counter: %d", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && effects.IsToken(c) && c.Counters["+1/+1"] != 0 {
			t.Error("a token has no riot")
		}
	}
	// A countered attacker deals double; one without a counter does not.
	striker := effectstest.PushVanillaCreature(g, me.ID, "Striker", 2, 2)
	effectstest.B28AddCounter(t, g, striker, 1)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, striker, plain)
	if opp.Life != before-6-2 {
		t.Errorf("the 3/3 with a counter deals 6, the 2/2 deals 2: %d → %d", before, opp.Life)
	}
}

func TestB28FatedFirepowerAddsFireCountersToDamageAtOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 6)
	fp := effectstest.CastXSpell(t, g, "Fated Firepower", "Enchantment", effectstest.B28FatedFirepowerOracle, "{X}{R}{R}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, fp, "fire"); got != 2 {
		t.Fatalf("enters with X fire counters: %d", got)
	}
	before := opp.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-5 {
		t.Errorf("a Bolt at an opponent deals 3+2: %d → %d", before, opp.Life)
	}
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if c := effectstest.B12Card(t, g, theirs); c.DamageMarked != 5 {
		t.Errorf("a Bolt at their permanent deals 3+2: %d marked", c.DamageMarked)
	}
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	if c := effectstest.B12Card(t, g, mine); c.DamageMarked != 3 {
		t.Errorf("your own permanent takes the printed 3: %d marked", c.DamageMarked)
	}
	// An opponent's Bolt at you is not a source you control.
	mine2 := me.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine2-3 {
		t.Errorf("their Bolt deals the printed 3: %d → %d", mine2, me.Life)
	}
	effectstest.AssertKeywords(t, g, fp, "flash")
}

func TestB28RepercussionPunishesTheDamagedCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Repercussion", "Enchantment", effectstest.B28RepercussionOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Creature — Golem", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Golem", "Creature — Golem", 4, 4)
	before := effectstest.B28Lives(g)
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: theirs})
	if opp.Life != before[1]-3 {
		t.Errorf("their creature took 3, so they take 3: %d → %d", before[1], opp.Life)
	}
	if me.Life != before[0] {
		t.Error("nothing for the caster")
	}
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: mine})
	if me.Life != before[0]-3 {
		t.Errorf("your own creature dealt damage hits you too: %d → %d", before[0], me.Life)
	}
	// Damage to a player is not a creature being dealt damage.
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before[1]-3-3 {
		t.Errorf("a Bolt to the face is just the Bolt: %d", opp.Life)
	}
}

func TestB28ChivalricAllianceDrawsOnATwoCreatureAttackOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Chivalric Alliance", "Enchantment", effectstest.B28ChivalricAllianceOracle, 0, 0)
	a := effectstest.PushVanillaCreature(g, me.ID, "Knight A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Knight B", 2, 2)
	c := effectstest.PushVanillaCreature(g, me.ID, "Knight C", 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("three attackers is one draw: %d → %d", hand, me.Hand.Size())
	}
	// A lone attacker next turn draws nothing.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.AdvanceToMain(t, g)
	hand = me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("one attacker is no draw: %d → %d", hand, me.Hand.Size())
	}
	// #1381: the discard-cost Knight-making ability is registered now
	// (chivalric_alliance_test.go proves it works) and the card is
	// complete.
	if spec, _ := effects.Lookup(effectstest.B28ChivalricAllianceOracle); len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the discard-cost ability should be registered and the card complete")
	}
}

// --- creatures -----------------------------------------------------

func TestB28HinterlandSanctifierGainsOnEachOtherCreatureYouControlEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Hinterland Sanctifier", "Creature — Rabbit Cleric", effectstest.B28HinterlandSanctifierOracle, 1, 2)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("a creature entering is 1 life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("two tokens are 2 more: %d", me.Life)
	}
	effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("an opponent's creature is not yours: %d", me.Life)
	}
}

func TestB28CauldronFamiliarDrainsEachOpponentForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	before := effectstest.B28Lives(g)
	effectstest.CastCatalogSpell(t, g, "Cauldron Familiar", "Creature — Cat", effectstest.B28CauldronFamiliarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	want := []int{before[0] + 1, before[1] - 1, before[2] - 1, before[3] - 1}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want[i])
		}
	}
	// #1381: the sacrifice-a-Food graveyard ability is registered now
	// (cauldron_familiar_test.go proves it works) and the card is
	// complete.
	if spec, _ := effects.Lookup(effectstest.B28CauldronFamiliarOracle); len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the graveyard ability should be registered and the card complete")
	}
}

func TestB28NykthosParagonOffersCountersOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	paragon := effectstest.B12Push(g, me.ID, "Nykthos Paragon", "Enchantment Creature — Human Soldier", effectstest.B28NykthosParagonOracle, 4, 6)
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B14Gain(t, g, me.ID, 3)
	if effectstest.OwedMays(g, me.ID) == 0 {
		t.Fatal("a life gain asks the question")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 3 {
		t.Errorf("that many counters on each creature you control: Bear has %d", got)
	}
	if got := effectstest.CounterCount(g, paragon, "+1/+1"); got != 3 {
		t.Errorf("the Paragon included: %d", got)
	}
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("not on an opponent's: %d", got)
	}
	effectstest.B14Gain(t, g, me.ID, 2)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("only once each turn")
	}
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 3 {
		t.Errorf("no second placement this turn: %d", got)
	}
	// Next turn: declining leaves the once-per-turn unused.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.B14Gain(t, g, me.ID, 1)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Gain(t, g, me.ID, 4)
	if effectstest.OwedMays(g, me.ID) == 0 {
		t.Fatal("a declined offer does not use up the turn")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 7 {
		t.Errorf("3 then 4: %d", got)
	}
}

func TestB28BenevolentHydraEntersWithXAndGivesOthersAnExtraCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	hydra := effectstest.CastXSpell(t, g, "Benevolent Hydra", "Creature — Hydra", effectstest.B28BenevolentHydraOracle, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("enters with X counters, its own bonus not applying to itself: %d", got)
	}
	effectstest.B28AddCounter(t, g, bear, 1)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("one counter on another creature you control is two: %d", got)
	}
	effectstest.B28AddCounter(t, g, hydra, 1)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 4 {
		t.Errorf("the Hydra itself gets the printed one: %d", got)
	}
	effectstest.B28AddCounter(t, g, theirs, 1)
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature gets the printed one: %d", got)
	}
	// #625: the counter-moving tap ability ships; counter_cost_cards_test.go
	// drives it. The X-counter timing is still a declared caveat.
	if spec, _ := effects.Lookup(effectstest.B28BenevolentHydraOracle); len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the tap ability ships and the card is complete")
	}
}

func TestB28UlvenwaldTrackerMakesYourCreatureFightAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tracker := effectstest.B12Push(g, me.ID, "Ulvenwald Tracker", "Creature — Human Shaman", effectstest.B28UlvenwaldTrackerOracle, 1, 1)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Golem", 4, 4)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}})
	if g.Battlefield.Contains(theirs) {
		t.Error("the 4/4 kills the 2/2")
	}
	if c := effectstest.B12Card(t, g, mine); c.DamageMarked != 2 {
		t.Errorf("the 2/2 hits back for 2: %d marked", c.DamageMarked)
	}
	if !effectstest.B12Card(t, g, tracker).Tapped {
		t.Error("the Tracker tapped to activate")
	}
	// Their creature first is the wrong way round: refused at announce.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tracker) })
	other := effectstest.PushVanillaCreature(g, opp.ID, "Their Ox", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: other}, {Kind: game.TargetCard, ID: mine},
	}}); err == nil {
		t.Error("slot 0 must be yours — a pair that doesn't fit is refused")
	}
	// A creature cannot fight itself: "another".
	if err := g.ActivateCatalogAbility(me.ID, tracker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: mine},
	}}); err == nil {
		t.Error("the second target is another creature")
	}
	if effectstest.B12Card(t, g, tracker).Tapped {
		t.Error("a refused activation costs nothing")
	}
	if spec, _ := effects.Lookup(effectstest.B28UlvenwaldTrackerOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the per-slot clauses landed")
	}
}

func TestB28DragonstormGlobeGrowsEnteringDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	globe := effectstest.B12Push(g, me.ID, "Dragonstorm Globe", "Artifact", effectstest.B28DragonstormGlobeOracle, 0, 0)
	dragon := effectstest.CastCatalogSpell(t, g, "My Dragon", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, dragon, "+1/+1"); got != 1 {
		t.Errorf("a Dragon enters with an additional counter: %d", got)
	}
	bear := effectstest.CastCatalogSpell(t, g, "My Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 0 {
		t.Errorf("a Bear does not: %d", got)
	}
	theirs := effectstest.B13PlayAs(t, g, 1, "Their Dragon", "Creature — Dragon", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 0 {
		t.Errorf("an opponent's Dragon does not: %d", got)
	}
	effectstest.B28TapForMana(t, g, me.ID, globe, "U")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("any colour: pool %v", got)
	}
}

func TestB28RegalCaracalLordsCatsAndMakesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cat := effectstest.B12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	theirCat := effectstest.B12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	caracal := effectstest.B20CastCreature(t, g, me, "Regal Caracal", "Creature — Cat", effectstest.B28RegalCaracalOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Cat"); got != 2 {
		t.Fatalf("two Cat tokens: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat" && effects.IsToken(c) {
			if p := effectstest.EffectivePower(t, g, c.InstanceID); p != 2 {
				t.Errorf("a 1/1 token Cat under the Caracal is 2/2: power %d", p)
			}
			effectstest.AssertKeywords(t, g, c.InstanceID, "lifelink")
		}
	}
	if p := effectstest.EffectivePower(t, g, cat); p != 3 {
		t.Errorf("another Cat you control gets +1/+1: %d", p)
	}
	effectstest.AssertKeywords(t, g, cat, "lifelink")
	if p := effectstest.EffectivePower(t, g, caracal); p != 3 {
		t.Errorf("the Caracal itself is unchanged: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirCat); p != 2 {
		t.Errorf("an opponent's Cat is unchanged: %d", p)
	}
}

func TestB28StromkirkCaptainLordsOtherVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.B12Push(g, me.ID, "Stromkirk Captain", "Creature — Vampire Soldier", effectstest.B28StromkirkCaptainOracle, 2, 2)
	vamp := effectstest.B12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 2, 2)
	if p := effectstest.EffectivePower(t, g, vamp); p != 3 {
		t.Errorf("another Vampire you control is +1/+1: %d", p)
	}
	effectstest.AssertKeywords(t, g, vamp, "first strike")
	effectstest.AssertKeywords(t, g, captain, "first strike")
	if p := effectstest.EffectivePower(t, g, captain); p != 2 {
		t.Errorf("the Captain is 'other' — unchanged: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, human); p != 2 {
		t.Errorf("a Human is unchanged: %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Vampire is unchanged: %d", p)
	}
}

func TestB28TiamatTutorsDifferentlyNamedDragonsOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Dragon A", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Dragon B", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Dragon A", TypeLine: "Creature — Dragon"},
		game.Card{Name: "Tiamat", TypeLine: "Legendary Creature — Dragon God"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Tiamat", "Legendary Creature — Dragon God", effectstest.B28TiamatOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("cast Tiamat searches")
	}
	if effectstest.HasID(c.SearchCards, lib[3]) || effectstest.HasID(c.SearchCards, lib[4]) {
		t.Error("a second Tiamat and a Bear are not offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lib[0], lib[2]}); err == nil {
		t.Fatal("two Dragons with the same name are refused")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, lib[0], lib[1])
	if !me.Hand.Contains(lib[0]) || !me.Hand.Contains(lib[1]) {
		t.Error("the differently named Dragons are in hand")
	}
	// A Tiamat that was not cast does not search.
	dead := effectstest.B17GraveyardCard(me, "Tiamat", "Legendary Creature — Dragon God", "{2}{W}{U}{B}{R}{G}")
	g.Battlefield.Cards = nil
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(dead, game.ZoneBattlefield, me.ID)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a reanimated Tiamat has no 'if you cast it'")
	}
}

func TestB28ArcanisDrawsThreeAndBouncesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arcanis := effectstest.B12Push(g, me.ID, "Arcanis the Omnipotent", "Legendary Creature — Wizard", effectstest.B28ArcanisOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, arcanis, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+3 {
		t.Errorf("{T}: draw three: %d → %d", hand, me.Hand.Size())
	}
	if !effectstest.B12Card(t, g, arcanis).Tapped {
		t.Error("tapped to draw")
	}
	effectstest.B16Activate(t, g, me.ID, arcanis, 1, game.ActivateAbilityParams{})
	if !me.Hand.Contains(arcanis) || g.Battlefield.Contains(arcanis) {
		t.Error("{2}{U}{U}: back to its owner's hand")
	}
}

func TestB28KykarMakesSpiritsOnNoncreatureSpellsAndEatsThemForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kykar := effectstest.B12Push(g, me.ID, "Kykar, Wind's Fury", "Legendary Creature — Bird Wizard", effectstest.B28KykarOracle, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Some Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Fatalf("a noncreature spell makes a Spirit: %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("a creature spell does not: %d", got)
	}
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && effects.IsToken(c) {
			spirit = c.InstanceID
			effectstest.AssertKeywords(t, g, c.InstanceID, "flying")
		}
	}
	if err := g.ActivateManaAbility(me.ID, kykar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{spirit}}); err != nil {
		t.Fatalf("Sacrifice a Spirit: Add {R}: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("{R} in the pool: %v", got)
	}
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit was sacrificed")
	}
	bear := effectstest.FindBattlefieldByName(g, "Some Bear")
	if err := g.ActivateManaAbility(me.ID, kykar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear is not a Spirit")
	}
}

func TestB28RavenousSquirrelGrowsOnSacrificeAndCashesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	squirrel := effectstest.B12Push(g, me.ID, "Ravenous Squirrel", "Creature — Squirrel", effectstest.B28RavenousSquirrelOracle, 1, 1)
	food := effectstest.B12Permanent(g, me.ID, "Food", "Artifact — Food")
	land := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, squirrel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{food}})
	if g.Battlefield.Contains(food) {
		t.Fatal("the Food is sacrificed as the cost")
	}
	if got := effectstest.CounterCount(g, squirrel, "+1/+1"); got != 1 {
		t.Errorf("sacrificing an artifact grows the Squirrel: %d", got)
	}
	if me.Life != life+1 || me.Hand.Size() != hand+1 {
		t.Errorf("1 life and a card: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	if err := g.ActivateCatalogAbility(me.ID, squirrel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("a land is neither an artifact nor a creature")
	}
	// Sacrificing a creature to anything else grows it too.
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, squirrel, "+1/+1"); got != 2 {
		t.Errorf("a creature sacrificed elsewhere counts: %d", got)
	}
}

func TestB28KarametrasAcolyteTapsForDevotionToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	acolyte := b28Push(g, me.ID, "Karametra's Acolyte", "Creature — Human Druid", effectstest.B28KarametrasAcolyteOracle, "{3}{G}", 1, 4, "G")
	b28Push(g, me.ID, "Llanowar Visionary", "Creature — Elf Druid", "", "{2}{G}{G}", 2, 2, "G")
	b28Push(g, me.ID, "Sol Ring", "Artifact", "", "{1}", 0, 0)
	b28Push(g, g.Seats[1].ID, "Their Elf", "Creature — Elf", "", "{G}", 1, 1, "G")
	effectstest.B28TapForMana(t, g, me.ID, acolyte, "")
	if got := effectstest.PoolColors(me); len(got) != 3 {
		t.Errorf("devotion to green is 3 (one from the Acolyte, two from the Visionary): pool %v", got)
	}
	for _, c := range effectstest.PoolColors(me) {
		if c != "G" {
			t.Errorf("all green: %v", effectstest.PoolColors(me))
		}
	}
}

func TestB28ThunderbreakRegentPunishesOpponentsTargetingYourDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	regent := effectstest.B12Push(g, me.ID, "Thunderbreak Regent", "Creature — Dragon", effectstest.B28ThunderbreakRegentOracle, 4, 4)
	dragon := effectstest.B12Creature(g, me.ID, "My Dragon", "Creature — Dragon", 5, 5)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: dragon})
	if effectstest.TriggerOnStack(g, regent) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("targeting your Dragon triggers the Regent")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the targeting opponent takes 3: %d → %d", before, opp.Life)
	}
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: regent})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("the Regent itself is a Dragon: %d", opp.Life)
	}
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("a Bear is not a Dragon: %d", opp.Life)
	}
	// Your own spell at your own Dragon is not an opponent's.
	mine := me.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: dragon})
	if me.Life != mine || opp.Life != before-6 {
		t.Error("only an opponent's spell or ability triggers it")
	}
}

func TestB28BrimazMakesAnAttackingCatAndABlockingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brimaz := effectstest.B12Push(g, me.ID, "Brimaz, King of Oreskos", "Legendary Creature — Cat Soldier", effectstest.B28BrimazOracle, 3, 4)
	effectstest.AssertKeywords(t, g, brimaz, "vigilance")
	before := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, brimaz)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if got := effectstest.B28TokensNamed(g, me.ID, "Cat Soldier"); got != 1 {
		t.Fatalf("one Cat Soldier: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat Soldier" && effects.IsToken(c) {
			if c.AttackingTarget != opp.ID {
				t.Error("the token is attacking the player Brimaz attacks")
			}
			effectstest.AssertKeywords(t, g, c.InstanceID, "vigilance")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != before-4 {
		t.Errorf("Brimaz and the token connect for 3+1: %d → %d", before, opp.Life)
	}
	// Their turn: Brimaz blocks and makes another.
	effectstest.AangAdvanceToMain(t, g, 1)
	raider := effectstest.PushVanillaCreature(g, opp.ID, "Raider", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(brimaz, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// #830: the declaration announces at its lock-in, not at the
	// click, so the block trigger needs the priority wrap first.
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Cat Soldier"); got != 2 {
		t.Errorf("a second Cat Soldier for the block: %d", got)
	}
	// The second Cat Soldier is put onto the battlefield BLOCKING the
	// raider (CR 506.3d), not an untapped bystander.
	var blocker *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Cat Soldier" && effects.IsToken(*c) && c.BlockingTarget != uuid.Nil {
			blocker = c
		}
	}
	if blocker == nil || blocker.BlockingTarget != raider {
		t.Fatalf("the Cat Soldier Brimaz makes when it blocks should be blocking the raider: %+v", blocker)
	}
	if n := effectstest.TriggersOnStackFrom(g, brimaz); n != 0 || len(g.PendingTriggers) != 0 {
		t.Error("a token put onto the battlefield blocking was never declared a blocker")
	}
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	// The double-blocked raider assigns its damage before any of the
	// step's is dealt (#3665, CR 510.2).
	effectstest.B26AnswerDamageAssignments(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("Brimaz (3) and the blocking Cat Soldier (1) kill the 2/2 raider")
	}
	if me.Life != life {
		t.Errorf("the raider was blocked: the defender took %d", life-me.Life)
	}
	if spec, _ := effects.Lookup(effectstest.B28BrimazOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Brimaz works as printed")
	}
}

func TestB28LifebloodHydraPaysOutItsPowerOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hydra := effectstest.CastXSpell(t, g, "Lifeblood Hydra", "Creature — Hydra", effectstest.B28LifebloodHydraOracle, "{X}{G}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 3 {
		t.Fatalf("enters with X counters: %d", got)
	}
	effectstest.AssertKeywords(t, g, hydra, "trample")
	life, hand := me.Life, me.Hand.Size()
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: hydra})
	if g.Battlefield.Contains(hydra) {
		t.Fatal("3 damage kills the 3/3")
	}
	if me.Life != life+3 || me.Hand.Size() != hand+3 {
		t.Errorf("gain 3 and draw 3: life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
}

func TestB28SpawnbedProtectorReturnsAnEldraziAndMakesScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Spawnbed Protector", "Creature — Eldrazi", effectstest.B28SpawnbedProtectorOracle, 6, 8)
	eldrazi := effectstest.B17GraveyardCard(me, "Dead Eldrazi", "Creature — Eldrazi", "{5}")
	bear := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("with an Eldrazi creature card in the graveyard the end step asks for a target")
	}
	if effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, eldrazi) {
		t.Error("only Eldrazi creature cards are offered")
	}
	effectstest.PickCard(t, g, me.ID, eldrazi)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(eldrazi) {
		t.Error("the Eldrazi is back in hand")
	}
	if got := effectstest.B28TokensNamed(g, me.ID, "Eldrazi Scion"); got != 2 {
		t.Errorf("two Scions: %d", got)
	}
	// Without an Eldrazi card the untargeted declaration still makes Scions.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("nothing to target, nothing to ask")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B28TokensNamed(g, me.ID, "Eldrazi Scion"); got != 4 {
		t.Errorf("two more Scions with no target: %d", got)
	}
	var scion uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi Scion" && effects.IsToken(c) {
			scion = c.InstanceID
		}
	}
	if err := g.ActivateManaAbility(me.ID, scion, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Sacrifice this token: Add {C}: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("{C} in the pool: %v", got)
	}
}
