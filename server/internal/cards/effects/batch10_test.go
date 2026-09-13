package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch10_test.go — card-level coverage for the card-coverage
// roadmap's batch 10 (#303, `edhrec_rank` 1111–1213): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play or attack. Helpers from
// the earlier batch test files are reused by name; new ones are
// b10-prefixed.

const (
	b10MaraudingBlightPriestOracle = "814b87fe-2a75-4ff2-8637-7e69e3fb285b"
	b10DrakusethOracle             = "060deaff-44d6-4f03-9568-bcb7add80255"
	b10KalonianHydraOracle         = "7bd36106-04fe-481f-b16e-e076dcbb183b"
	b10MarbleDiamondOracle         = "910488bf-66ab-415e-973b-1262b2ab7454"
	b10FirebrandArcherOracle       = "2e9289d6-dbc6-456d-88cf-d1f534e731d6"
	b10LumraOracle                 = "97a84e9d-bfc4-4ca2-b1e8-908dba56ccdb"
	b10SecretRendezvousOracle      = "34b90d3e-f48c-41ff-b3e4-5cab7a9cd597"
	b10WaterloggedGroveOracle      = "70fa2eba-565e-4fed-adc9-7f5d9fcbf1fa"
	b10RadiantSummitOracle         = "5dd0cc44-4647-4857-ad3b-22494099d08a"
	b10MagusOfTheWheelOracle       = "df13dafb-f82e-42ac-b6dc-efe82af5db58"
	b10RecruiterOfTheGuardOracle   = "d521a329-a53a-4962-810a-2abed80df260"
	b10SheoldredsEdictOracle       = "217062f5-96f1-454c-9507-17f34ef37070"
	b10TitaniaOracle               = "d0ade00d-a496-441d-9b7e-7dc033d3292c"
	b10SoulShatterOracle           = "615927c2-3fb0-4e64-a1a7-55fb56de1423"
	b10VindicateOracle             = "63c1ac21-e3d8-40c2-8c09-3f31c52992ef"
	b10KedissOracle                = "d9c87cc2-943e-49b6-becc-748857549617"
	b10TheShireOracle              = "9abf9a0e-8e7d-406b-a01d-d4870b30134e"
	b10MikokoroOracle              = "a4580a1d-141e-449b-9018-e0258130634b"
	b10DamningVerdictOracle        = "355ed7ef-bfa6-4538-99fb-a2d203eb7005"
	b10CityOfTraitorsOracle        = "f161111d-9747-47b3-bb10-3c8bded32e21"
	b10EssenceWardenOracle         = "6ca2a89e-7032-4864-b4e9-66f3178f90ab"
	b10SuturePriestOracle          = "c4d36522-3ace-4bfb-bf2d-1a366f458698"
	b10GoblinEngineerOracle        = "c1d6cce8-085f-42cb-8b0c-b6fbbf88b16a"
	b10SirenStormtamerOracle       = "59c7496b-19a2-478b-b28f-6f153c2458ae"
	b10CrackleWithPowerOracle      = "273f5483-b67e-4dd6-bba8-c0a047fa34d7"
	b10NotionThiefOracle           = "f8dab16e-1d50-443e-9431-8b6f1cf61c9c"
	b10EzurisPredationOracle       = "d0dd425b-fdba-41b4-b9e6-f5161610bd7e"
	b10DragonsHoardOracle          = "8cf77dc4-763b-41b7-a5da-0ef2734f08e6"
	b10ArchonOfSunsGraceOracle     = "48a99dce-0aa9-4aac-81df-cec5f94c639d"
	b10PatriarsSealOracle          = "ad95eac2-5b48-4102-9988-a6d7b1ec30e0"
	b10MarwynOracle                = "ee35de1c-aef1-4bd4-85fd-fe77bc927790"
	b10MountDoomOracle             = "995c8dac-fd27-468a-abd4-02372cf0c850"
	b10WastelandOracle             = "09a70ae8-3859-4a09-901d-dce063fa3b5f"
	b10FireLitThicketOracle        = "d99a1d9a-7721-4331-bf22-1c6ee0bd825a"
	b10PrimalVigorOracle           = "c665544f-557b-4631-a1dc-39571470ca2e"

	b10RampantGrowthOracle = "8539f295-5d58-4436-a73a-b9277c4c7795"
)

// b10Awake clears summoning sickness on a creature that was cast
// this turn, so a test can attack or tap with it in the same turn.
func b10Awake(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// b10Creature pushes a live creature with the given type line, P/T
// and (optionally) mana cost, with a real entry timestamp and no
// summoning sickness.
func b10Creature(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b10LibraryTop pushes a card onto the TOP of a player's library — the
// next card drawn or milled.
func b10LibraryTop(p *game.Player, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b10AddMana floats the named mana in a player's pool.
func b10AddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// b10ResolveAllManaPicks answers every open colour pick for a player
// with `color`, returning how many were answered.
func b10ResolveAllManaPicks(t *testing.T, g *game.Game, chooser uuid.UUID, color string) int {
	t.Helper()
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == chooser {
			if err := g.ResolveManaChoice(c.ID, chooser, color); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			n++
		}
	}
	return n
}

// b10OpponentCastsBolt puts a Lightning Bolt in an opponent's hand and
// casts it at the given target, returning the spell's ID.
func b10OpponentCastsBolt(t *testing.T, g *game.Game, opp *game.Player, target game.TargetRef) uuid.UUID {
	t.Helper()
	return batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "{R}", []game.TargetRef{target})
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Marble Diamond
// and Radiant Summit are rows in cycle tables, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch10CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b10MaraudingBlightPriestOracle: "Marauding Blight-Priest",
		b10DrakusethOracle:             "Drakuseth, Maw of Flames",
		b10KalonianHydraOracle:         "Kalonian Hydra",
		b10MarbleDiamondOracle:         "Marble Diamond",
		b10FirebrandArcherOracle:       "Firebrand Archer",
		b10LumraOracle:                 "Lumra, Bellow of the Woods",
		b10SecretRendezvousOracle:      "Secret Rendezvous",
		b10WaterloggedGroveOracle:      "Waterlogged Grove",
		b10RadiantSummitOracle:         "Radiant Summit",
		b10MagusOfTheWheelOracle:       "Magus of the Wheel",
		b10RecruiterOfTheGuardOracle:   "Recruiter of the Guard",
		b10SheoldredsEdictOracle:       "Sheoldred's Edict",
		b10TitaniaOracle:               "Titania, Protector of Argoth",
		b10SoulShatterOracle:           "Soul Shatter",
		b10VindicateOracle:             "Vindicate",
		b10KedissOracle:                "Kediss, Emberclaw Familiar",
		b10TheShireOracle:              "The Shire",
		b10MikokoroOracle:              "Mikokoro, Center of the Sea",
		b10DamningVerdictOracle:        "Damning Verdict",
		b10CityOfTraitorsOracle:        "City of Traitors",
		b10EssenceWardenOracle:         "Essence Warden",
		b10SuturePriestOracle:          "Suture Priest",
		b10GoblinEngineerOracle:        "Goblin Engineer",
		b10SirenStormtamerOracle:       "Siren Stormtamer",
		b10CrackleWithPowerOracle:      "Crackle with Power",
		b10NotionThiefOracle:           "Notion Thief",
		b10EzurisPredationOracle:       "Ezuri's Predation",
		b10DragonsHoardOracle:          "Dragon's Hoard",
		b10ArchonOfSunsGraceOracle:     "Archon of Sun's Grace",
		b10PatriarsSealOracle:          "Patriar's Seal",
		b10MarwynOracle:                "Marwyn, the Nurturer",
		b10MountDoomOracle:             "Mount Doom",
		b10WastelandOracle:             "Wasteland",
		b10FireLitThicketOracle:        "Fire-Lit Thicket",
		b10PrimalVigorOracle:           "Primal Vigor",
	}
	if len(want) != 35 {
		t.Fatalf("the batch is 35 cards, the table lists %d", len(want))
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
		// The battle-land table predates the completeness field and
		// declares nothing for any of its rows; Radiant Summit
		// inherits that, and stamping the whole cycle is not this
		// batch's call.
		if spec.Completeness == CompletenessUnreviewed && oracle != b10RadiantSummitOracle {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- lifegain and cast payoffs ------------------------------------

func TestB10MaraudingBlightPriestDrainsTheTableWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	priest := pushCatalogPermanent(g, me.ID, "Marauding Blight-Priest", "Creature — Vampire Cleric", b10MaraudingBlightPriestOracle, false)
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(priest, me.ID, 3) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 (one gain, one trigger)", i+1, b, got)
		}
	}

	// Losing life, or an opponent gaining, is silent.
	mid := lifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(priest, me.ID, -2)
		_ = g.ChangePlayerLifeForEffect(priest, g.Seats[1].ID, 2)
	})
	passPriorityAroundTable(t, g)
	for i, b := range mid {
		want := b
		if i == 0 {
			want = b + 2
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent %d: %d → %d, want %d", i+1, b, got, want)
		}
	}
}

func TestB10FirebrandArcherPingsOnNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", b10FirebrandArcherOracle, false)
	before := lifeOfOpponents(g)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		want := b - 1
		if i == 0 {
			want -= 3
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent %d: %d → %d, want %d", i+1, b, got, want)
		}
	}

	mid := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a creature spell pinged opponent %d: %d → %d", i+1, b, got)
		}
	}
}

func TestB10EssenceWardenGainsOnAnyCreatureButNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	castCatalogSpell(t, g, "Essence Warden", "Creature — Elf Shaman", b10EssenceWardenOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Fatalf("its own entry gained life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 2)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("two opposing creatures: life %d → %d, want +2", before, me.Life)
	}
}

func TestB10SuturePriestAsksOnBothSidesOfTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Suture Priest", "Creature — Phyrexian Cleric", b10SuturePriestOracle, false)
	myBefore, theirBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != myBefore+1 || opp.Life != theirBefore {
		t.Errorf("my creature: me %d → %d (want +1), opp %d → %d (want same)", myBefore, me.Life, theirBefore, opp.Life)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != myBefore+1 || opp.Life != theirBefore-1 {
		t.Errorf("their creature: me %d → %d (want +1), opp %d → %d (want -1)", myBefore, me.Life, theirBefore, opp.Life)
	}

	// Declining does nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if opp.Life != theirBefore-1 {
		t.Errorf("a declined prompt still drained: %d", opp.Life)
	}
}

// --- attack triggers ----------------------------------------------

func TestB10DrakusethDealsFourThenThreeToUpToTwoOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drak := b10Creature(g, me.ID, "Drakuseth, Maw of Flames", "Legendary Creature — Dragon", "{4}{R}{R}{R}", 7, 7)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drak {
				g.Battlefield.Cards[i].OracleID = b10DrakusethOracle
			}
		}
	})
	bear := seedCreature(g, "Bear", opp.ID)
	big := b10Creature(g, opp.ID, "Big", "Creature — Giant", "", 4, 4)
	before := opp.Life

	declareAttack(t, g, opp.ID, drak)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p.PickTargetMin != 1 || p.PickTargetMax != 3 {
		t.Fatalf("the clause is one to three targets, got %d..%d", p.PickTargetMin, p.PickTargetMax)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID},
		{Kind: game.TargetCard, ID: bear},
		{Kind: game.TargetCard, ID: big},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the first target takes 4: %d → %d", before, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage should kill the 2/2")
	}
	if got := damageMarkedOn(g, big); got != 3 {
		t.Errorf("the 4/4 has %d damage marked, want 3", got)
	}
}

func TestB10KalonianHydraEntersWithFourAndDoublesYourTeamOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := castCatalogSpell(t, g, "Kalonian Hydra", "Creature — Hydra", b10KalonianHydraOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hydra, "+1/+1"); got != 4 {
		t.Fatalf("entered with %d +1/+1 counters, want 4", got)
	}
	if !eotHasAbility(effectiveAbilities(t, g, hydra), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	if p := effectivePower(t, g, hydra); p != 0 {
		t.Errorf("printed power should be 0 before counters, got %d", p)
	}

	b10Awake(g, hydra)
	bear := pushCounterCreature(g, me.ID, "Bear", "+1/+1", 1)
	theirs := pushCounterCreature(g, opp.ID, "Theirs", "+1/+1", 2)
	declareAttack(t, g, opp.ID, hydra)
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{hydra, 8, "the Hydra's four double to eight"},
		{bear, 2, "one counter doubles to two"},
		{theirs, 2, "an opponent's creature is untouched"},
	} {
		if got := counterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.why, got, tc.want)
		}
	}
}

func TestB10KedissSpreadsCommanderDamageToTheOtherOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, struck := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Kediss, Emberclaw Familiar", "Legendary Creature — Elemental Lizard", b10KedissOracle, false)
	cmdr := b10Creature(g, me.ID, "General", "Legendary Creature — Human", "{2}{R}", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	grunt := b10Creature(g, me.ID, "Grunt", "Creature — Human", "", 2, 2)
	before := lifeOfOpponents(g)

	dealCombatDamageToPlayer(g, cmdr, struck.ID, 5)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		want := b - 5
		if i == 0 {
			want = b // the fake event moved no life; the struck player is skipped by the trigger
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent %d: %d → %d, want %d", i+1, b, got, want)
		}
	}

	mid := lifeOfOpponents(g)
	dealCombatDamageToPlayer(g, grunt, struck.ID, 2)
	passPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a non-commander triggered Kediss: opponent %d %d → %d", i+1, b, got)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB10SecretRendezvousDrawsThreeForYouAndTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine, theirs, others := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	castCatalogSpell(t, g, "Secret Rendezvous", "Sorcery", b10SecretRendezvousOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != mine+3 {
		t.Errorf("caster: hand %d → %d, want +3 (the spell was added and cast in between)", mine, me.Hand.Size())
	}
	if opp.Hand.Size() != theirs+3 {
		t.Errorf("target: hand %d → %d, want +3", theirs, opp.Hand.Size())
	}
	if other.Hand.Size() != others {
		t.Errorf("a non-target drew: %d → %d", others, other.Hand.Size())
	}
}

func TestB10VindicateDestroysAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Vindicate", "Sorcery", b10VindicateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the land should be in its owner's graveyard")
	}
}

func TestB10SheoldredsEdictModesAskForTheRightKindOfPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := seedCreature(g, "Bear", opp.ID)
	seedCreature(g, "Mine", me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1) })
	goblin := findBattlefieldByName(g, "Goblin")

	castModal(t, g, "Sheoldred's Edict", "Instant", b10SheoldredsEdictOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the opponent should be asked to sacrifice a nontoken creature")
	}
	if len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != bear {
		t.Errorf("nontoken mode offered %v, want only the bear", c.SacrificeOptions)
	}
	if sacrificeChoiceFor(g, me.ID) != nil || sacrificeChoiceFor(g, other.ID) != nil {
		t.Error("only opponents with a matching creature are asked")
	}
	answerSacrifice(t, g, opp.ID, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the bear was not sacrificed")
	}

	castModal(t, g, "Sheoldred's Edict", "Instant", b10SheoldredsEdictOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	c = sacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != goblin {
		t.Fatalf("token mode should offer only the Goblin, got %+v", c)
	}
	answerSacrifice(t, g, opp.ID, goblin)

	castModal(t, g, "Sheoldred's Edict", "Instant", b10SheoldredsEdictOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("nobody controls a planeswalker, so nobody is asked")
	}
}

func TestB10SoulShatterTakesEachOpponentsBiggestCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	small := b10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	dragon := b10Creature(g, opp.ID, "Dragon", "Creature — Dragon", "{3}{R}{R}", 4, 4)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}{U}{U}",
		Owner: other.ID, Controller: other.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	mine := b10Creature(g, me.ID, "Mine", "Creature — Giant", "{5}{G}{G}", 7, 7)

	castCatalogSpell(t, g, "Soul Shatter", "Instant", b10SoulShatterOracle, nil)
	passPriorityAroundTable(t, g)
	c := sacrificeChoiceFor(g, opp.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != dragon {
		t.Fatalf("the first opponent should be offered only the mana-value-5 Dragon, got %+v", c)
	}
	answerSacrifice(t, g, opp.ID, dragon)
	c = sacrificeChoiceFor(g, other.ID)
	if c == nil || len(c.SacrificeOptions) != 1 || c.SacrificeOptions[0] != walker {
		t.Fatalf("the second opponent should be offered their only planeswalker, got %+v", c)
	}
	answerSacrifice(t, g, other.ID, walker)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster is not an opponent")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("only the greatest mana value goes")
	}
}

func TestB10DamningVerdictSparesCreaturesWithAnyCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := seedCreature(g, "Plain", me.ID)
	grown := pushCounterCreature(g, opp.ID, "Grown", "+1/+1", 1)
	shrunk := pushCounterCreature(g, opp.ID, "Shrunk", "-1/-1", 1)
	theirPlain := seedCreature(g, "Their Plain", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Damning Verdict", "Sorcery", b10DamningVerdictOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{plain, theirPlain} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a creature with no counters survived: %s", id)
		}
	}
	for _, id := range []uuid.UUID{grown, shrunk, rock} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a countered creature or a noncreature was destroyed: %s", id)
		}
	}
}

func TestB10CrackleWithPowerDealsFiveXToExactlyXTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	b1, b2 := opp.Life, other.Life
	castXSpell(t, g, "Crackle with Power", "Sorcery", b10CrackleWithPowerOracle, "{X}{X}{X}{R}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != b1-10 || other.Life != b2-10 {
		t.Errorf("X=2: %d → %d and %d → %d, want -10 each", b1, opp.Life, b2, other.Life)
	}

	// The target count IS X: two targets on an X of 1 is refused.
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Crackle with Power", TypeLine: "Sorcery",
		OracleID: b10CrackleWithPowerOracle, ManaCost: "{X}{X}{X}{R}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID},
	}}); err == nil {
		t.Error("two targets were accepted on X=1")
	}
}

func TestB10EzurisPredationMakesABeastPerOpposingCreatureAndFightsEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := seedCreature(g, "Bear", opp.ID)
	big := b10Creature(g, opp.ID, "Big", "Creature — Giant", "", 5, 5)
	small := b10Creature(g, other.ID, "Small", "Creature — Rat", "", 1, 1)
	mine := seedCreature(g, "Mine", me.ID)

	castCatalogSpell(t, g, "Ezuri's Predation", "Sorcery", b10EzurisPredationOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, small} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a 4-power fight should have killed %s", id)
		}
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 4 {
		t.Errorf("the 5/5 should survive with 4 damage, marked %d", damageMarkedOn(g, big))
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's creature is not one of \"those creatures\"")
	}
	if got := countBattlefieldNamed(g, me.ID, "Phyrexian Beast"); got != 2 {
		t.Errorf("three Beasts minus the one the 5/5 killed = 2, got %d", got)
	}
}

func TestB10EzurisPredationDoesNothingWithNoOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Ezuri's Predation", "Sorcery", b10EzurisPredationOracle, nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Phyrexian Beast") != 0 {
		t.Error("no creatures, no Beasts")
	}
}

// --- ETB creatures -------------------------------------------------

func TestB10LumraMillsFourThenReturnsEveryLandTappedAndCountsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	old := pushGraveyardPermanent(me, "Old Swamp", "Basic Land — Swamp", "")
	// Top of the library, top-first: the next four are milled.
	spell := b10LibraryTop(me, "Spell", "Instant", "{U}", 0, 0)
	island := b10LibraryTop(me, "Island", "Basic Land — Island", "", 0, 0)
	bear := b10LibraryTop(me, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	plains := b10LibraryTop(me, "Plains", "Basic Land — Plains", "", 0, 0)
	fifth := b10LibraryTop(me, "Fifth", "Basic Land — Mountain", "", 0, 0)
	_ = fifth

	lumra := castCatalogSpell(t, g, "Lumra, Bellow of the Woods", "Legendary Creature — Elemental Bear", b10LumraOracle, nil)
	passPriorityAroundTable(t, g)
	// The fifth card was drawn by nothing: it was the top card, so it
	// is milled; the four milled are fifth, plains, bear, island.
	for _, id := range []uuid.UUID{fifth, plains, island, old} {
		card, ok := battlefieldCard(g, id)
		if !ok {
			t.Errorf("land %s should have returned to the battlefield", id)
			continue
		}
		if !card.Tapped {
			t.Errorf("land %s should have returned tapped", id)
		}
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the milled creature stays in the graveyard")
	}
	if me.Graveyard.Contains(spell) || !me.Library.Contains(spell) {
		t.Error("only four cards are milled")
	}
	// One Forest already there plus four returned.
	if p, tg := effectivePower(t, g, lumra), effectiveToughness(t, g, lumra); p != 5 || tg != 5 {
		t.Errorf("Lumra is %d/%d, want 5/5 for five lands", p, tg)
	}
	if !eotHasAbility(effectiveAbilities(t, g, lumra), "reach") || !eotHasAbility(effectiveAbilities(t, g, lumra), "vigilance") {
		t.Error("printed reach and vigilance did not reach the effective abilities")
	}
}

func TestB10RecruiterOfTheGuardFindsOnlyToughnessTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "Recruiter of the Guard", "Creature — Human Soldier", b10RecruiterOfTheGuardOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the optional search should always prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Fatalf("options %v, want only the 1/1 Elf", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Elf should be in hand")
	}
}

func TestB10TitaniaReturnsALandAndPaysOutWhenYourLandsDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")

	castCatalogSpell(t, g, "Titania, Protector of Argoth", "Legendary Creature — Elemental", b10TitaniaOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != forest {
		t.Fatalf("'target land card in your graveyard' offered %v, want only the Forest", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest should be back on the battlefield")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(forest) })
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("a land of mine dying should make one 5/3 Elemental, got %d", got)
	}
	elemental := battlefieldIDNamed(g, me.ID, "Elemental")
	if card, _ := battlefieldCard(g, elemental); card.Power != 5 || card.Toughness != 3 {
		t.Errorf("the token is %d/%d, want 5/3", card.Power, card.Toughness)
	}

	// An opponent's land, and a land of mine that is bounced, are silent.
	theirs := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mine := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(theirs)
		_ = g.BounceToHandForEffect(mine)
	})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("still one Elemental expected, got %d", got)
	}
}

func TestB10GoblinEngineerEntombsAnArtifactAndWeldsOneBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	engineer := castCatalogSpell(t, g, "Goblin Engineer", "Creature — Goblin Artificer", b10GoblinEngineerOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the optional search should prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Fatalf("options %v, want only the artifact", c.SearchCards)
	}
	answerSearchByID(t, g, me.ID, ids[1])
	if !me.Graveyard.Contains(ids[1]) {
		t.Fatal("the artifact should be in the graveyard, not in hand")
	}

	b10Awake(g, engineer)
	fodder := seedPermanentFor(g, me.ID, "Fodder", "Artifact")
	bear := seedCreature(g, "Bear", me.ID)
	pricey := pushGraveyardPermanent(me, "Pricey", "Artifact", "{4}")
	b10AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: pricey}},
	}); err == nil {
		t.Fatal("a mana-value-4 artifact was accepted as a target")
	}
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: ids[1]}},
	}); err == nil {
		t.Fatal("a creature was accepted for 'sacrifice an artifact'")
	}
	if err := g.ActivateCatalogAbility(me.ID, engineer, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: ids[1]}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the artifact is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the Ring should be back on the battlefield")
	}
}

func TestB10ArchonOfSunsGraceMakesFlyingPegasiWithLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archon := b10Creature(g, me.ID, "Archon of Sun's Grace", "Creature — Archon", "{2}{W}{W}", 3, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == archon {
				g.Battlefield.Cards[i].OracleID = b10ArchonOfSunsGraceOracle
			}
		}
	})
	if abilities := effectiveAbilities(t, g, archon); !eotHasAbility(abilities, "flying") || !eotHasAbility(abilities, "lifelink") {
		t.Error("printed flying and lifelink did not reach the effective abilities")
	}

	castCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Pegasus"); got != 1 {
		t.Fatalf("an enchantment entering should make one Pegasus, got %d", got)
	}
	pegasus := battlefieldIDNamed(g, me.ID, "Pegasus")
	if abilities := effectiveAbilities(t, g, pegasus); !eotHasAbility(abilities, "flying") || !eotHasAbility(abilities, "lifelink") {
		t.Errorf("the Pegasus should fly and have lifelink from the Archon, got %v", abilities)
	}

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Glory", TypeLine: "Token Enchantment"}, 1)
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Pegasus"); got != 1 {
		t.Errorf("an opponent's enchantment or a creature triggered constellation: %d Pegasi", got)
	}
}

func TestB10MarwynGrowsOnElvesAndTapsForHerPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	marwyn := pushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", b10MarwynOracle, false)
	if err := g.ActivateManaAbility(me.ID, marwyn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("power 1: pool %v, want [G]", got)
	}

	castCatalogSpell(t, g, "Elf", "Creature — Elf Warrior", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, marwyn, "+1/+1"); got != 1 {
		t.Fatalf("an Elf entering should put one counter on Marwyn, got %d", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, marwyn, "+1/+1"); got != 1 {
		t.Errorf("a non-Elf grew Marwyn: %d counters", got)
	}

	me.ManaPool.EmptyPool()
	b08Untap(g, marwyn)
	if err := g.ActivateManaAbility(me.ID, marwyn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("power 2: pool %v, want [G G]", got)
	}
}

func TestB10DragonsHoardCountsDragonsAndTapsForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hoard := pushCatalogPermanent(g, me.ID, "Dragon's Hoard", "Artifact", b10DragonsHoardOracle, false)
	castCatalogSpell(t, g, "Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hoard, "gold"); got != 1 {
		t.Fatalf("a Dragon entering should add one gold counter, got %d", got)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, hoard, "gold"); got != 1 {
		t.Errorf("a non-Dragon added a counter: %d", got)
	}
	if err := g.ActivateManaAbility(me.ID, hoard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("'any color' must offer all five, got %+v", pick)
	}
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: b10DragonsHoardOracle})); n != 0 {
		t.Errorf("the draw ability is declared missing; found %d activated abilities", n)
	}
}

// --- replacements --------------------------------------------------

func TestB10NotionThiefStealsOpponentsDrawsOutsideTheirDrawStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedReplacementPermanent(g, b10NotionThiefOracle, "Notion Thief", me.ID)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()

	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	if me.Hand.Size() != mine+2 || opp.Hand.Size() != theirs {
		t.Errorf("two stolen draws: me %d → %d (want +2), opp %d → %d (want same)", mine, me.Hand.Size(), theirs, opp.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if me.Hand.Size() != mine+3 {
		t.Errorf("my own draw is untouched: %d", me.Hand.Size())
	}

	// The opponent's own draw step is left alone.
	mine, theirs = me.Hand.Size(), opp.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("expected the draw step, at %s", g.Turn.Step)
	}
	if opp.Hand.Size() != theirs+1 || me.Hand.Size() != mine {
		t.Errorf("draw step: opp %d → %d (want +1), me %d → %d (want same)", theirs, opp.Hand.Size(), mine, me.Hand.Size())
	}
}

func TestB10PrimalVigorDoublesEveryonesPlusOneCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedReplacementPermanent(g, b10PrimalVigorOracle, "Primal Vigor", me.ID)
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	rock := seedPermanentFor(g, opp.ID, "Rock", "Artifact")

	for _, id := range []uuid.UUID{mine, theirs} {
		if err := g.AddCounter(id, "+1/+1", 1); err != nil {
			t.Fatal(err)
		}
		if got := counterCount(g, id, "+1/+1"); got != 2 {
			t.Errorf("%s: one +1/+1 counter became %d, want 2", id, got)
		}
	}
	if err := g.AddCounter(theirs, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	if got := counterCount(g, theirs, "-1/-1"); got != 1 {
		t.Errorf("a -1/-1 counter was doubled: %d", got)
	}
	if err := g.AddCounter(rock, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := counterCount(g, rock, "+1/+1"); got != 1 {
		t.Errorf("a noncreature's counter was doubled: %d", got)
	}
}

// --- activated abilities ------------------------------------------

func TestB10MagusOfTheWheelWheelsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	magus := pushCatalogPermanent(g, me.ID, "Magus of the Wheel", "Creature — Human Wizard", b10MagusOfTheWheelOracle, false)
	toMain(t, g)
	b10AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(magus) || !me.Graveyard.Contains(magus) {
		t.Error("the Magus is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %d: hand %d, want 7", i, p.Hand.Size())
		}
	}
}

func TestB10MikokoroDrawsForEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := seedPermanentWithOracle(g, me.ID, "Mikokoro, Center of the Sea", "Legendary Land", b10MikokoroOracle)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	b10AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != before[i]+1 {
			t.Errorf("seat %d: hand %d → %d, want +1", i, before[i], p.Hand.Size())
		}
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the land is tapped; its mana ability should be refused")
	}
}

func TestB10PatriarsSealUntapsOnlyALegendaryCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := pushCatalogPermanent(g, me.ID, "Patriar's Seal", "Artifact", b10PatriarsSealOracle, false)
	legend := b10Creature(g, me.ID, "Legend", "Legendary Creature — Human", "", 2, 2)
	plain := b10Creature(g, me.ID, "Plain", "Creature — Human", "", 2, 2)
	theirs := b10Creature(g, opp.ID, "Theirs", "Legendary Creature — Human", "", 2, 2)
	for _, id := range []uuid.UUID{legend, plain, theirs} {
		b08Tap(g, id)
	}
	b10AddMana(me, "C")
	for _, bad := range []uuid.UUID{plain, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}},
		}); err == nil {
			t.Errorf("%s was accepted for 'target legendary creature you control'", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, seal, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if card, _ := battlefieldCard(g, legend); card.Tapped {
		t.Error("the legend should be untapped")
	}
	if err := g.ActivateManaAbility(me.ID, seal, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Seal is tapped from the untap ability")
	}
}

func TestB10SirenStormtamerCountersASpellAimedAtYouOrYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	siren := pushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", b10SirenStormtamerOracle, false)
	before := me.Life

	// A spell at someone else is not a legal target.
	stray := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID})
	b10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stray}},
	}); err == nil {
		t.Fatal("a spell targeting another player was accepted")
	}
	if !g.Battlefield.Contains(siren) {
		t.Fatal("a refused activation must not sacrifice the Stormtamer")
	}
	passPriorityAroundTable(t, g)

	bolt := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(siren) {
		t.Error("the Stormtamer is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("the Bolt should have been countered: life %d → %d", before, me.Life)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt should be in its owner's graveyard")
	}
}

// The printed ruling: sacrificing the Stormtamer to counter a spell
// aimed only at the Stormtamer leaves that spell targeting a creature
// in a graveyard, so the ability fizzles. A spell at another creature
// of yours is countered as usual.
func TestB10SirenStormtamerCannotSaveItselfButCanSaveAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	siren := pushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", b10SirenStormtamerOracle, false)
	bear := seedCreature(g, "Bear", me.ID)

	bolt := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	b10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt at the bear should have been countered")
	}

	siren2 := pushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", b10SirenStormtamerOracle, false)
	bolt2 := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: siren2})
	b10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren2, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt2}},
	}); err != nil {
		t.Fatalf("a spell at the Stormtamer itself is a legal target at announce: %v", err)
	}
	fizzles := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventFizzle && ev.Source == siren2 {
			fizzles++
		}
	}
	passPriorityAroundTable(t, g)
	after := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventFizzle && ev.Source == siren2 {
			after++
		}
	}
	if after != fizzles+1 {
		t.Errorf("the ability should fizzle once its only protected creature is in the graveyard: %d fizzles", after-fizzles)
	}
}

func TestB10MountDoomThreeAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := seedPermanentWithOracle(g, me.ID, "Mount Doom", "Legendary Land", b10MountDoomOracle)
	lifeBefore := me.Life
	if err := g.ActivateManaAbility(me.ID, doom, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("the mana costs a life: %d → %d", lifeBefore, me.Life)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask B or R, got %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "B")

	b08Untap(g, doom)
	me.ManaPool.EmptyPool()
	b10AddMana(me, "C", "B", "R")
	oppBefore := lifeOfOpponents(g)
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doom, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	passPriorityAroundTable(t, g)
	for i, b := range oppBefore {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i+1, b, got)
		}
	}

	b08Untap(g, doom)
	me.ManaPool.EmptyPool()
	b10AddMana(me, "C", "C", "C", "C", "C", "B", "R")
	ring := seedPermanentFor(g, me.ID, "The One Ring", "Legendary Artifact")
	rock := seedPermanentFor(g, me.ID, "Rock", "Artifact")
	keepMine := seedCreature(g, "Keep Mine", me.ID)
	keepTheirs := seedCreature(g, "Keep Theirs", opp.ID)
	loseMine := seedCreature(g, "Lose Mine", me.ID)
	loseTheirs := seedCreature(g, "Lose Theirs", opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, doom, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: keepMine}},
	}); err == nil {
		t.Fatal("a non-legendary artifact was accepted for the sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, doom, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{ring},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: keepMine},
			{Kind: game.TargetCard, ID: keepTheirs},
		},
	}); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	if g.Battlefield.Contains(doom) || g.Battlefield.Contains(ring) {
		t.Error("Mount Doom and the legendary artifact are sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{keepMine, keepTheirs} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a chosen creature was destroyed: %s", id)
		}
	}
	for _, id := range []uuid.UUID{loseMine, loseTheirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("an unchosen creature survived: %s", id)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("only creatures are destroyed")
	}
}

func TestB10WastelandDestroysOnlyANonbasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	waste := seedPermanentWithOracle(g, me.ID, "Wasteland", "Land", b10WastelandOracle)
	basic := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := seedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	if err := g.ActivateCatalogAbility(me.ID, waste, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: basic}},
	}); err == nil {
		t.Fatal("a basic land was accepted for 'target nonbasic land'")
	}
	if !g.Battlefield.Contains(waste) {
		t.Fatal("a refused activation must not sacrifice the Wasteland")
	}
	if err := g.ActivateCatalogAbility(me.ID, waste, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: nonbasic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nonbasic) || !g.Battlefield.Contains(basic) {
		t.Error("the nonbasic land should be destroyed and the basic untouched")
	}
}

// --- lands ---------------------------------------------------------

func TestB10MarbleDiamondEntersTappedAndTapsForWhite(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := castCatalogSpell(t, g, "Marble Diamond", "Artifact", b10MarbleDiamondOracle, nil)
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, rock, "Marble Diamond")
	b08Untap(g, rock)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB10RadiantSummitEntersUntappedWithTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	tapped := playLandFromHand(t, g, "Radiant Summit", b10RadiantSummitOracle)
	top100AssertEnteredTapped(t, g, tapped, "Radiant Summit (one basic)")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	seedLandOnBattlefield(g2, me2.ID, "Mountain", "Basic Land — Mountain")
	seedLandOnBattlefield(g2, me2.ID, "Plains", "Basic Land — Plains")
	untapped := playLandFromHand(t, g2, "Radiant Summit", b10RadiantSummitOracle)
	top100AssertEnteredUntapped(t, g2, untapped, "Radiant Summit (two basics)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g2, me2.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask R or W, got %+v", pick)
	}
}

func TestB10WaterloggedGroveCostsALifeAndCashesInForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	grove := seedPermanentWithOracle(g, me.ID, "Waterlogged Grove", "Land", b10WaterloggedGroveOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, grove, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask G or U, got %+v", pick)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "G")

	b08Untap(g, grove)
	me.ManaPool.EmptyPool()
	b10AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, grove, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(grove) {
		t.Error("the Grove is sacrificed as a cost, at announce")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestB10TheShireEntersUntappedWithALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	tapped := playLandFromHand(t, g, "The Shire", b10TheShireOracle)
	top100AssertEnteredTapped(t, g, tapped, "The Shire (no legend)")

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	b10Creature(g2, me2.ID, "Legend", "Legendary Creature — Halfling", "", 1, 1)
	untapped := playLandFromHand(t, g2, "The Shire", b10TheShireOracle)
	top100AssertEnteredUntapped(t, g2, untapped, "The Shire (with a legend)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me2); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: b10TheShireOracle})); n != 0 {
		t.Errorf("the Food ability is declared missing; found %d activated abilities", n)
	}
}

func TestB10CityOfTraitorsGoesWhenYouPlayALandButNotWhenYouFetchOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	city := pushCatalogPermanent(g, me.ID, "City of Traitors", "Land", b10CityOfTraitorsOracle, false)
	if err := g.ActivateManaAbility(me.ID, city, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}

	// A land put onto the battlefield from the library is not a play.
	me.ManaPool.EmptyPool()
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castCatalogSpell(t, g, "Rampant Growth", "Sorcery", b10RampantGrowthOracle, nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Forest") == uuid.Nil {
		t.Fatal("Rampant Growth should have fetched the Forest")
	}
	if !g.Battlefield.Contains(city) {
		t.Fatal("a fetched land is not a land play; the City should still be there")
	}

	playLandFromHand(t, g, "Swamp", "")
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(city) || !me.Graveyard.Contains(city) {
		t.Error("playing a land should have sacrificed the City")
	}
}

// Pins the declared gap: a land RETURNED from a graveyard reads as a
// land play and costs the City. Flips when a land-play event lands.
func TestB10CityOfTraitorsAlsoGoesWhenALandReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	city := pushCatalogPermanent(g, me.ID, "City of Traitors", "Land", b10CityOfTraitorsOracle, false)
	forest := pushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(forest, game.ZoneBattlefield, uuid.Nil)
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(city) {
		t.Error("declared gap: a land returned from the graveyard counts as a play (weaker than printed); if this now passes, close the caveat")
	}
}

func TestB10FireLitThicketIsAFilterLand(t *testing.T) {
	spec, ok := Lookup(b10FireLitThicketOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if len(spec.ManaAbilities) != 2 {
		t.Fatalf("%d mana abilities, want 2", len(spec.ManaAbilities))
	}
	if spec.ManaAbilities[0].Produced != "{C}" {
		t.Errorf("ability 0 must be the {C} half, got %q", spec.ManaAbilities[0].Produced)
	}
	filter := spec.ManaAbilities[1]
	if filter.Cost.Mana != "{R/G}" || filter.Produced != "{R|G}{R|G}" || !filter.IgnoreCommanderIdentity {
		t.Errorf("filter half is %+v, want a {R/G} cost and two {R|G} picks", filter)
	}
}

func TestB10FireLitThicketFiltersOneHybridIntoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	thicket := seedPermanentWithOracle(g, me.ID, "Fire-Lit Thicket", "Land", b10FireLitThicketOracle)
	b10AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, thicket, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := b10ResolveAllManaPicks(t, g, me.ID, "R"); n != 2 {
		t.Fatalf("want two colour picks, got %d", n)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}
}
