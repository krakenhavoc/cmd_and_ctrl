package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Marble Diamond
// and Radiant Summit are rows in cycle tables, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch10CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B10MaraudingBlightPriestOracle: "Marauding Blight-Priest",
		effectstest.B10DrakusethOracle:             "Drakuseth, Maw of Flames",
		effectstest.B10KalonianHydraOracle:         "Kalonian Hydra",
		effectstest.B10MarbleDiamondOracle:         "Marble Diamond",
		effectstest.B10FirebrandArcherOracle:       "Firebrand Archer",
		effectstest.B10LumraOracle:                 "Lumra, Bellow of the Woods",
		effectstest.B10SecretRendezvousOracle:      "Secret Rendezvous",
		effectstest.B10WaterloggedGroveOracle:      "Waterlogged Grove",
		effectstest.B10RadiantSummitOracle:         "Radiant Summit",
		effectstest.B10MagusOfTheWheelOracle:       "Magus of the Wheel",
		effectstest.B10RecruiterOfTheGuardOracle:   "Recruiter of the Guard",
		effectstest.B10SheoldredsEdictOracle:       "Sheoldred's Edict",
		effectstest.B10TitaniaOracle:               "Titania, Protector of Argoth",
		effectstest.B10SoulShatterOracle:           "Soul Shatter",
		effectstest.B10VindicateOracle:             "Vindicate",
		effectstest.B10KedissOracle:                "Kediss, Emberclaw Familiar",
		effectstest.B10TheShireOracle:              "The Shire",
		effectstest.B10MikokoroOracle:              "Mikokoro, Center of the Sea",
		effectstest.B10DamningVerdictOracle:        "Damning Verdict",
		effectstest.B10CityOfTraitorsOracle:        "City of Traitors",
		effectstest.B10EssenceWardenOracle:         "Essence Warden",
		effectstest.B10SuturePriestOracle:          "Suture Priest",
		effectstest.B10GoblinEngineerOracle:        "Goblin Engineer",
		effectstest.B10SirenStormtamerOracle:       "Siren Stormtamer",
		effectstest.B10CrackleWithPowerOracle:      "Crackle with Power",
		effectstest.B10NotionThiefOracle:           "Notion Thief",
		effectstest.B10EzurisPredationOracle:       "Ezuri's Predation",
		effectstest.B10DragonsHoardOracle:          "Dragon's Hoard",
		effectstest.B10ArchonOfSunsGraceOracle:     "Archon of Sun's Grace",
		effectstest.B10PatriarsSealOracle:          "Patriar's Seal",
		effectstest.B10MarwynOracle:                "Marwyn, the Nurturer",
		effectstest.B10MountDoomOracle:             "Mount Doom",
		effectstest.B10WastelandOracle:             "Wasteland",
		effectstest.B10FireLitThicketOracle:        "Fire-Lit Thicket",
		effectstest.B10PrimalVigorOracle:           "Primal Vigor",
	}
	if len(want) != 35 {
		t.Fatalf("the batch is 35 cards, the table lists %d", len(want))
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
		// The battle-land table predates the completeness field and
		// declares nothing for any of its rows; Radiant Summit
		// inherits that, and stamping the whole cycle is not this
		// batch's call.
		if spec.Completeness == effects.CompletenessUnreviewed && oracle != effectstest.B10RadiantSummitOracle {
			t.Errorf("%s ships without a completeness declaration", name)
		}
	}
}

// --- lifegain and cast payoffs ------------------------------------

func TestB10MaraudingBlightPriestDrainsTheTableWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Marauding Blight-Priest", "Creature — Vampire Cleric", effectstest.B10MaraudingBlightPriestOracle, false)
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(priest, me.ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1 (one gain, one trigger)", i+1, b, got)
		}
	}

	// Losing life, or an opponent gaining, is silent.
	mid := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(priest, me.ID, -2)
		_ = g.ChangePlayerLifeForEffect(priest, g.Seats[1].ID, 2)
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", effectstest.B10FirebrandArcherOracle, false)
	before := effectstest.LifeOfOpponents(g)

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		want := b - 1
		if i == 0 {
			want -= 3
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent %d: %d → %d, want %d", i+1, b, got, want)
		}
	}

	mid := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a creature spell pinged opponent %d: %d → %d", i+1, b, got)
		}
	}
}

func TestB10EssenceWardenGainsOnAnyCreatureButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Essence Warden", "Creature — Elf Shaman", effectstest.B10EssenceWardenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Fatalf("its own entry gained life: %d → %d", before, me.Life)
	}
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 2)
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("two opposing creatures: life %d → %d, want +2", before, me.Life)
	}
}

func TestB10SuturePriestAsksOnBothSidesOfTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Suture Priest", "Creature — Phyrexian Cleric", effectstest.B10SuturePriestOracle, false)
	myBefore, theirBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myBefore+1 || opp.Life != theirBefore {
		t.Errorf("my creature: me %d → %d (want +1), opp %d → %d (want same)", myBefore, me.Life, theirBefore, opp.Life)
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myBefore+1 || opp.Life != theirBefore-1 {
		t.Errorf("their creature: me %d → %d (want +1), opp %d → %d (want -1)", myBefore, me.Life, theirBefore, opp.Life)
	}

	// Declining does nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirBefore-1 {
		t.Errorf("a declined prompt still drained: %d", opp.Life)
	}
}

// --- attack triggers ----------------------------------------------

func TestB10DrakusethDealsFourThenThreeToUpToTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drak := effectstest.B10Creature(g, me.ID, "Drakuseth, Maw of Flames", "Legendary Creature — Dragon", "{4}{R}{R}{R}", 7, 7)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == drak {
				g.Battlefield.Cards[i].OracleID = effectstest.B10DrakusethOracle
			}
		}
	})
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	big := effectstest.B10Creature(g, opp.ID, "Big", "Creature — Giant", "", 4, 4)
	before := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, drak)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the first target takes 4: %d → %d", before, opp.Life)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("3 damage should kill the 2/2")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 3 {
		t.Errorf("the 4/4 has %d damage marked, want 3", got)
	}
}

func TestB10KalonianHydraEntersWithFourAndDoublesYourTeamOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.CastCatalogSpell(t, g, "Kalonian Hydra", "Creature — Hydra", effectstest.B10KalonianHydraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hydra, "+1/+1"); got != 4 {
		t.Fatalf("entered with %d +1/+1 counters, want 4", got)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, hydra), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	if p := effectstest.EffectivePower(t, g, hydra); p != 0 {
		t.Errorf("printed power should be 0 before counters, got %d", p)
	}

	effectstest.B10Awake(g, hydra)
	bear := effectstest.PushCounterCreature(g, me.ID, "Bear", "+1/+1", 1)
	theirs := effectstest.PushCounterCreature(g, opp.ID, "Theirs", "+1/+1", 2)
	effectstest.DeclareAttack(t, g, opp.ID, hydra)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{hydra, 8, "the Hydra's four double to eight"},
		{bear, 2, "one counter doubles to two"},
		{theirs, 2, "an opponent's creature is untouched"},
	} {
		if got := effectstest.CounterCount(g, tc.id, "+1/+1"); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.why, got, tc.want)
		}
	}
}

func TestB10KedissSpreadsCommanderDamageToTheOtherOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, struck := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Kediss, Emberclaw Familiar", "Legendary Creature — Elemental Lizard", effectstest.B10KedissOracle, false)
	cmdr := effectstest.B10Creature(g, me.ID, "General", "Legendary Creature — Human", "{2}{R}", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	grunt := effectstest.B10Creature(g, me.ID, "Grunt", "Creature — Human", "", 2, 2)
	before := effectstest.LifeOfOpponents(g)

	effectstest.DealCombatDamageToPlayer(g, cmdr, struck.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		want := b - 5
		if i == 0 {
			want = b // the fake event moved no life; the struck player is skipped by the trigger
		}
		if got := g.Seats[i+1].Life; got != want {
			t.Errorf("opponent %d: %d → %d, want %d", i+1, b, got, want)
		}
	}

	mid := effectstest.LifeOfOpponents(g)
	effectstest.DealCombatDamageToPlayer(g, grunt, struck.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range mid {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a non-commander triggered Kediss: opponent %d %d → %d", i+1, b, got)
		}
	}
}

// --- spells --------------------------------------------------------

func TestB10SecretRendezvousDrawsThreeForYouAndTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine, theirs, others := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Secret Rendezvous", "Sorcery", effectstest.B10SecretRendezvousOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Vindicate", "Sorcery", effectstest.B10VindicateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("the land should be in its owner's graveyard")
	}
}

func TestB10SheoldredsEdictModesAskForTheRightKindOfPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	effectstest.SeedCreature(g, "Mine", me.ID)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	goblin := effectstest.FindBattlefieldByName(g, "Goblin")

	// Each mode leaves the opponent exactly one matching permanent, so
	// the pick is forced and collected without a prompt (ADR 0175,
	// #3664): what the mode admits shows in what is sacrificed.
	effectstest.CastModal(t, g, "Sheoldred's Edict", "Instant", effectstest.B10SheoldredsEdictOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil || effectstest.SacrificeChoiceFor(g, other.ID) != nil {
		t.Error("only opponents with a matching creature are asked")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the bear, the only nontoken creature, was not sacrificed")
	}
	if !g.Battlefield.Contains(goblin) {
		t.Error("nontoken mode took the Goblin token")
	}

	effectstest.CastModal(t, g, "Sheoldred's Edict", "Instant", effectstest.B10SheoldredsEdictOracle, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) {
		t.Error("token mode did not take the Goblin")
	}
	if effectstest.FindBattlefieldByName(g, "Mine") == uuid.Nil {
		t.Error("the caster is not an opponent")
	}

	effectstest.CastModal(t, g, "Sheoldred's Edict", "Instant", effectstest.B10SheoldredsEdictOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("nobody controls a planeswalker, so nobody is asked")
	}
}

func TestB10SoulShatterTakesEachOpponentsBiggestCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	small := effectstest.B10Creature(g, opp.ID, "Small", "Creature — Bear", "{1}{G}", 2, 2)
	dragon := effectstest.B10Creature(g, opp.ID, "Dragon", "Creature — Dragon", "{3}{R}{R}", 4, 4)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}{U}{U}",
		Owner: other.ID, Controller: other.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	mine := effectstest.B10Creature(g, me.ID, "Mine", "Creature — Giant", "{5}{G}{G}", 7, 7)

	effectstest.CastCatalogSpell(t, g, "Soul Shatter", "Instant", effectstest.B10SoulShatterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Each opponent has exactly one permanent of greatest mana value,
	// so neither is asked (ADR 0175's forced pick, #3664).
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster is not an opponent")
	}
	if g.Battlefield.Contains(dragon) {
		t.Error("the first opponent's mana-value-5 Dragon was not sacrificed")
	}
	if g.Battlefield.Contains(walker) {
		t.Error("the second opponent's only planeswalker was not sacrificed")
	}
	if !g.Battlefield.Contains(small) || !g.Battlefield.Contains(mine) {
		t.Error("only the greatest mana value goes")
	}
}

func TestB10DamningVerdictSparesCreaturesWithAnyCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := effectstest.SeedCreature(g, "Plain", me.ID)
	grown := effectstest.PushCounterCreature(g, opp.ID, "Grown", "+1/+1", 1)
	shrunk := effectstest.PushCounterCreature(g, opp.ID, "Shrunk", "-1/-1", 1)
	theirPlain := effectstest.SeedCreature(g, "Their Plain", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Damning Verdict", "Sorcery", effectstest.B10DamningVerdictOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	opp, other := g.Seats[1], g.Seats[2]
	b1, b2 := opp.Life, other.Life
	effectstest.CastXSpell(t, g, "Crackle with Power", "Sorcery", effectstest.B10CrackleWithPowerOracle, "{X}{X}{X}{R}{R}", 2,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != b1-10 || other.Life != b2-10 {
		t.Errorf("X=2: %d → %d and %d → %d, want -10 each", b1, opp.Life, b2, other.Life)
	}

	// The target count IS X: two targets on an X of 1 is refused.
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Crackle with Power", TypeLine: "Sorcery",
		OracleID: effectstest.B10CrackleWithPowerOracle, ManaCost: "{X}{X}{X}{R}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID},
	}}); err == nil {
		t.Error("two targets were accepted on X=1")
	}
}

func TestB10EzurisPredationMakesABeastPerOpposingCreatureAndFightsEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	big := effectstest.B10Creature(g, opp.ID, "Big", "Creature — Giant", "", 5, 5)
	small := effectstest.B10Creature(g, other.ID, "Small", "Creature — Rat", "", 1, 1)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)

	effectstest.CastCatalogSpell(t, g, "Ezuri's Predation", "Sorcery", effectstest.B10EzurisPredationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, small} {
		if g.Battlefield.Contains(id) {
			t.Errorf("a 4-power fight should have killed %s", id)
		}
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 4 {
		t.Errorf("the 5/5 should survive with 4 damage, marked %d", effectstest.DamageMarkedOn(g, big))
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's creature is not one of \"those creatures\"")
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Phyrexian Beast"); got != 2 {
		t.Errorf("three Beasts minus the one the 5/5 killed = 2, got %d", got)
	}
}

func TestB10EzurisPredationDoesNothingWithNoOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Ezuri's Predation", "Sorcery", effectstest.B10EzurisPredationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Phyrexian Beast") != 0 {
		t.Error("no creatures, no Beasts")
	}
}

// --- ETB creatures -------------------------------------------------

func TestB10LumraMillsFourThenReturnsEveryLandTappedAndCountsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	old := effectstest.PushGraveyardPermanent(me, "Old Swamp", "Basic Land — Swamp", "")
	// Top of the library, top-first: the next four are milled.
	spell := effectstest.B10LibraryTop(me, "Spell", "Instant", "{U}", 0, 0)
	island := effectstest.B10LibraryTop(me, "Island", "Basic Land — Island", "", 0, 0)
	bear := effectstest.B10LibraryTop(me, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	plains := effectstest.B10LibraryTop(me, "Plains", "Basic Land — Plains", "", 0, 0)
	fifth := effectstest.B10LibraryTop(me, "Fifth", "Basic Land — Mountain", "", 0, 0)
	_ = fifth

	lumra := effectstest.CastCatalogSpell(t, g, "Lumra, Bellow of the Woods", "Legendary Creature — Elemental Bear", effectstest.B10LumraOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The fifth card was drawn by nothing: it was the top card, so it
	// is milled; the four milled are fifth, plains, bear, island.
	for _, id := range []uuid.UUID{fifth, plains, island, old} {
		card, ok := effectstest.BattlefieldCard(g, id)
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
	if p, tg := effectstest.EffectivePower(t, g, lumra), effectstest.EffectiveToughness(t, g, lumra); p != 5 || tg != 5 {
		t.Errorf("Lumra is %d/%d, want 5/5 for five lands", p, tg)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, lumra), "reach") || !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, lumra), "vigilance") {
		t.Error("printed reach and vigilance did not reach the effective abilities")
	}
}

func TestB10RecruiterOfTheGuardFindsOnlyToughnessTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
	)
	effectstest.CastCatalogSpell(t, g, "Recruiter of the Guard", "Creature — Human Soldier", effectstest.B10RecruiterOfTheGuardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the optional search should always prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Fatalf("options %v, want only the 1/1 Elf", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Error("the Elf should be in hand")
	}
}

func TestB10TitaniaReturnsALandAndPaysOutWhenYourLandsDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")

	effectstest.CastCatalogSpell(t, g, "Titania, Protector of Argoth", "Legendary Creature — Elemental", effectstest.B10TitaniaOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != forest {
		t.Fatalf("'target land card in your graveyard' offered %v, want only the Forest", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the Forest should be back on the battlefield")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(forest) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("a land of mine dying should make one 5/3 Elemental, got %d", got)
	}
	elemental := effectstest.BattlefieldIDNamed(g, me.ID, "Elemental")
	if card, _ := effectstest.BattlefieldCard(g, elemental); card.Power != 5 || card.Toughness != 3 {
		t.Errorf("the token is %d/%d, want 5/3", card.Power, card.Toughness)
	}

	// An opponent's land, and a land of mine that is bounced, are silent.
	theirs := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(theirs)
		_ = g.BounceToHandForEffect(mine)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Elemental"); got != 1 {
		t.Errorf("still one Elemental expected, got %d", got)
	}
}

func TestB10GoblinEngineerEntombsAnArtifactAndWeldsOneBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	engineer := effectstest.CastCatalogSpell(t, g, "Goblin Engineer", "Creature — Goblin Artificer", effectstest.B10GoblinEngineerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the optional search should prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Fatalf("options %v, want only the artifact", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	if !me.Graveyard.Contains(ids[1]) {
		t.Fatal("the artifact should be in the graveyard, not in hand")
	}

	effectstest.B10Awake(g, engineer)
	fodder := effectstest.SeedPermanentFor(g, me.ID, "Fodder", "Artifact")
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	pricey := effectstest.PushGraveyardPermanent(me, "Pricey", "Artifact", "{4}")
	effectstest.B10AddMana(me, "R")
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
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the Ring should be back on the battlefield")
	}
}

func TestB10ArchonOfSunsGraceMakesFlyingPegasiWithLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archon := effectstest.B10Creature(g, me.ID, "Archon of Sun's Grace", "Creature — Archon", "{2}{W}{W}", 3, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == archon {
				g.Battlefield.Cards[i].OracleID = effectstest.B10ArchonOfSunsGraceOracle
			}
		}
	})
	if abilities := effectstest.EffectiveAbilities(t, g, archon); !effects.EotHasAbility(abilities, "flying") || !effects.EotHasAbility(abilities, "lifelink") {
		t.Error("printed flying and lifelink did not reach the effective abilities")
	}

	effectstest.CastCatalogSpell(t, g, "Glory", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Pegasus"); got != 1 {
		t.Fatalf("an enchantment entering should make one Pegasus, got %d", got)
	}
	pegasus := effectstest.BattlefieldIDNamed(g, me.ID, "Pegasus")
	if abilities := effectstest.EffectiveAbilities(t, g, pegasus); !effects.EotHasAbility(abilities, "flying") || !effects.EotHasAbility(abilities, "lifelink") {
		t.Errorf("the Pegasus should fly and have lifelink from the Archon, got %v", abilities)
	}

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Glory", TypeLine: "Token Enchantment"}, 1)
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Pegasus"); got != 1 {
		t.Errorf("an opponent's enchantment or a creature triggered constellation: %d Pegasi", got)
	}
}

func TestB10MarwynGrowsOnElvesAndTapsForHerPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := effectstest.PushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)
	if err := g.ActivateManaAbility(me.ID, marwyn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("power 1: pool %v, want [G]", got)
	}

	effectstest.CastCatalogSpell(t, g, "Elf", "Creature — Elf Warrior", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, marwyn, "+1/+1"); got != 1 {
		t.Fatalf("an Elf entering should put one counter on Marwyn, got %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, marwyn, "+1/+1"); got != 1 {
		t.Errorf("a non-Elf grew Marwyn: %d counters", got)
	}

	me.ManaPool.EmptyPool()
	effectstest.B08Untap(g, marwyn)
	if err := g.ActivateManaAbility(me.ID, marwyn, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("power 2: pool %v, want [G G]", got)
	}
}

func TestB10DragonsHoardCountsDragonsAndTapsForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hoard := effectstest.PushCatalogPermanent(g, me.ID, "Dragon's Hoard", "Artifact", effectstest.B10DragonsHoardOracle, false)
	effectstest.CastCatalogSpell(t, g, "Dragon", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hoard, "gold"); got != 1 {
		t.Fatalf("a Dragon entering should add one gold counter, got %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, hoard, "gold"); got != 1 {
		t.Errorf("a non-Dragon added a counter: %d", got)
	}
	if err := g.ActivateManaAbility(me.ID, hoard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("'any color' must offer all five, got %+v", pick)
	}
	// #625: the draw ability ships; counter_cost_cards_test.go drives it.
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.B10DragonsHoardOracle})); n != 1 {
		t.Errorf("the draw ability should be wired; found %d activated abilities", n)
	}
}

// --- replacements --------------------------------------------------

func TestB10NotionThiefStealsOpponentsDrawsOutsideTheirDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedReplacementPermanent(g, effectstest.B10NotionThiefOracle, "Notion Thief", me.ID)
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
	effectstest.AdvanceToUpkeepOf(t, g, 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedReplacementPermanent(g, effectstest.B10PrimalVigorOracle, "Primal Vigor", me.ID)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	rock := effectstest.SeedPermanentFor(g, opp.ID, "Rock", "Artifact")

	for _, id := range []uuid.UUID{mine, theirs} {
		if err := g.AddCounter(id, "+1/+1", 1); err != nil {
			t.Fatal(err)
		}
		if got := effectstest.CounterCount(g, id, "+1/+1"); got != 2 {
			t.Errorf("%s: one +1/+1 counter became %d, want 2", id, got)
		}
	}
	if err := g.AddCounter(theirs, "-1/-1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.CounterCount(g, theirs, "-1/-1"); got != 1 {
		t.Errorf("a -1/-1 counter was doubled: %d", got)
	}
	if err := g.AddCounter(rock, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.CounterCount(g, rock, "+1/+1"); got != 1 {
		t.Errorf("a noncreature's counter was doubled: %d", got)
	}
}

// --- activated abilities ------------------------------------------

func TestB10MagusOfTheWheelWheelsTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Wheel", "Creature — Human Wizard", effectstest.B10MagusOfTheWheelOracle, false)
	effectstest.ToMain(t, g)
	effectstest.B10AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(magus) || !me.Graveyard.Contains(magus) {
		t.Error("the Magus is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != 7 {
			t.Errorf("seat %d: hand %d, want 7", i, p.Hand.Size())
		}
	}
}

func TestB10MikokoroDrawsForEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Mikokoro, Center of the Sea", "Legendary Land", effectstest.B10MikokoroOracle)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	effectstest.B10AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seal := effectstest.PushCatalogPermanent(g, me.ID, "Patriar's Seal", "Artifact", effectstest.B10PatriarsSealOracle, false)
	legend := effectstest.B10Creature(g, me.ID, "Legend", "Legendary Creature — Human", "", 2, 2)
	plain := effectstest.B10Creature(g, me.ID, "Plain", "Creature — Human", "", 2, 2)
	theirs := effectstest.B10Creature(g, opp.ID, "Theirs", "Legendary Creature — Human", "", 2, 2)
	for _, id := range []uuid.UUID{legend, plain, theirs} {
		b08Tap(g, id)
	}
	effectstest.B10AddMana(me, "C")
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
	effectstest.PassPriorityAroundTable(t, g)
	if card, _ := effectstest.BattlefieldCard(g, legend); card.Tapped {
		t.Error("the legend should be untapped")
	}
	if err := g.ActivateManaAbility(me.ID, seal, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Seal is tapped from the untap ability")
	}
}

func TestB10SirenStormtamerCountersASpellAimedAtYouOrYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	siren := effectstest.PushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", effectstest.B10SirenStormtamerOracle, false)
	before := me.Life

	// A spell at someone else is not a legal target.
	stray := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: other.ID})
	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stray}},
	}); err == nil {
		t.Fatal("a spell targeting another player was accepted")
	}
	if !g.Battlefield.Contains(siren) {
		t.Fatal("a refused activation must not sacrifice the Stormtamer")
	}
	effectstest.PassPriorityAroundTable(t, g)

	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(siren) {
		t.Error("the Stormtamer is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	siren := effectstest.PushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", effectstest.B10SirenStormtamerOracle, false)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	bolt := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.B10AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, siren, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt at the bear should have been countered")
	}

	siren2 := effectstest.PushCatalogPermanent(g, me.ID, "Siren Stormtamer", "Creature — Siren Pirate Wizard", effectstest.B10SirenStormtamerOracle, false)
	bolt2 := effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: siren2})
	effectstest.B10AddMana(me, "U")
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := effectstest.SeedPermanentWithOracle(g, me.ID, "Mount Doom", "Legendary Land", effectstest.B10MountDoomOracle)
	lifeBefore := me.Life
	if err := g.ActivateManaAbility(me.ID, doom, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != lifeBefore-1 {
		t.Errorf("the mana costs a life: %d → %d", lifeBefore, me.Life)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask B or R, got %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "B")

	effectstest.B08Untap(g, doom)
	me.ManaPool.EmptyPool()
	effectstest.B10AddMana(me, "C", "B", "R")
	oppBefore := effectstest.LifeOfOpponents(g)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, doom, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range oppBefore {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i+1, b, got)
		}
	}

	effectstest.B08Untap(g, doom)
	me.ManaPool.EmptyPool()
	effectstest.B10AddMana(me, "C", "C", "C", "C", "C", "B", "R")
	ring := effectstest.SeedPermanentFor(g, me.ID, "The One Ring", "Legendary Artifact")
	rock := effectstest.SeedPermanentFor(g, me.ID, "Rock", "Artifact")
	keepMine := effectstest.SeedCreature(g, "Keep Mine", me.ID)
	keepTheirs := effectstest.SeedCreature(g, "Keep Theirs", opp.ID)
	loseMine := effectstest.SeedCreature(g, "Lose Mine", me.ID)
	loseTheirs := effectstest.SeedCreature(g, "Lose Theirs", opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, doom, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock},
	}); err == nil {
		t.Fatal("a non-legendary artifact was accepted for the sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, doom, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{ring},
	}); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	if g.Battlefield.Contains(doom) || g.Battlefield.Contains(ring) {
		t.Error("Mount Doom and the legendary artifact are sacrificed as the cost")
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the survivors are not chosen while the ability is on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 2 || !effectstest.HasID(pick.ChooseCards, loseTheirs) {
		t.Fatalf("one pick of up to two over every creature, asked as the ability resolves: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, keepMine, keepTheirs)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	waste := effectstest.SeedPermanentWithOracle(g, me.ID, "Wasteland", "Land", effectstest.B10WastelandOracle)
	basic := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	nonbasic := effectstest.SeedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nonbasic) || !g.Battlefield.Contains(basic) {
		t.Error("the nonbasic land should be destroyed and the basic untouched")
	}
}

// --- lands ---------------------------------------------------------

func TestB10MarbleDiamondEntersTappedAndTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.CastCatalogSpell(t, g, "Marble Diamond", "Artifact", effectstest.B10MarbleDiamondOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Top100AssertEnteredTapped(t, g, rock, "Marble Diamond")
	effectstest.B08Untap(g, rock)
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("pool %v, want [W]", got)
	}
}

func TestB10RadiantSummitEntersUntappedWithTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	tapped := effectstest.PlayLandFromHand(t, g, "Radiant Summit", effectstest.B10RadiantSummitOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "Radiant Summit (one basic)")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.SeedLandOnBattlefield(g2, me2.ID, "Mountain", "Basic Land — Mountain")
	effectstest.SeedLandOnBattlefield(g2, me2.ID, "Plains", "Basic Land — Plains")
	untapped := effectstest.PlayLandFromHand(t, g2, "Radiant Summit", effectstest.B10RadiantSummitOracle)
	effectstest.Top100AssertEnteredUntapped(t, g2, untapped, "Radiant Summit (two basics)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g2, me2.ID)
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask R or W, got %+v", pick)
	}
}

func TestB10WaterloggedGroveCostsALifeAndCashesInForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grove := effectstest.SeedPermanentWithOracle(g, me.ID, "Waterlogged Grove", "Land", effectstest.B10WaterloggedGroveOracle)
	before := me.Life
	if err := g.ActivateManaAbility(me.ID, grove, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want -1", before, me.Life)
	}
	if pick := effectstest.RiderLatestManaPick(g, me.ID); pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("must ask G or U, got %+v", pick)
	}
	effectstest.B10ResolveAllManaPicks(t, g, me.ID, "G")

	effectstest.B08Untap(g, grove)
	me.ManaPool.EmptyPool()
	effectstest.B10AddMana(me, "C")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, grove, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(grove) {
		t.Error("the Grove is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}
}

func TestB10TheShireEntersUntappedWithALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	tapped := effectstest.PlayLandFromHand(t, g, "The Shire", effectstest.B10TheShireOracle)
	effectstest.Top100AssertEnteredTapped(t, g, tapped, "The Shire (no legend)")

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B10Creature(g2, me2.ID, "Legend", "Legendary Creature — Halfling", "", 1, 1)
	untapped := effectstest.PlayLandFromHand(t, g2, "The Shire", effectstest.B10TheShireOracle)
	effectstest.Top100AssertEnteredUntapped(t, g2, untapped, "The Shire (with a legend)")
	if err := g2.ActivateManaAbility(me2.ID, untapped, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me2); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	// #1381: the Food ability is registered now (the_shire_test.go
	// proves it works).
	if n := len(game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.B10TheShireOracle})); n != 1 {
		t.Errorf("the Food ability should be registered; found %d activated abilities, want 1", n)
	}
}

func TestB10CityOfTraitorsGoesWhenYouPlayALandButNotWhenYouFetchOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	city := effectstest.PushCatalogPermanent(g, me.ID, "City of Traitors", "Land", effectstest.B10CityOfTraitorsOracle, false)
	if err := g.ActivateManaAbility(me.ID, city, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want [C C]", got)
	}

	// A land put onto the battlefield from the library is not a play.
	me.ManaPool.EmptyPool()
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Rampant Growth", "Sorcery", effectstest.B10RampantGrowthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if effectstest.FindBattlefieldByName(g, "Forest") == uuid.Nil {
		t.Fatal("Rampant Growth should have fetched the Forest")
	}
	if !g.Battlefield.Contains(city) {
		t.Fatal("a fetched land is not a land play; the City should still be there")
	}

	effectstest.PlayLandFromHand(t, g, "Swamp", "")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(city) || !me.Graveyard.Contains(city) {
		t.Error("playing a land should have sacrificed the City")
	}
}

// #1326: a land RETURNED from a graveyard by an effect is not a play
// (CR 305.4, Event.Played), and the City survives it.
func TestB10CityOfTraitorsSurvivesALandReturnedFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	city := effectstest.PushCatalogPermanent(g, me.ID, "City of Traitors", "Land", effectstest.B10CityOfTraitorsOracle, false)
	forest := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(forest, game.ZoneBattlefield, uuid.Nil)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(city) {
		t.Error("a land returned from the graveyard by an effect is not a land play; the City should still be there")
	}
}

func TestB10FireLitThicketIsAFilterLand(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B10FireLitThicketOracle)
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
	if filter.Cost.Mana != "{R/G}" || filter.Produced != "{R|G}{R|G}" || filter.NarrowToCommanderIdentity {
		t.Errorf("filter half is %+v, want a {R/G} cost and two {R|G} picks", filter)
	}
}

func TestB10FireLitThicketFiltersOneHybridIntoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thicket := effectstest.SeedPermanentWithOracle(g, me.ID, "Fire-Lit Thicket", "Land", effectstest.B10FireLitThicketOracle)
	effectstest.B10AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, thicket, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := effectstest.B10ResolveAllManaPicks(t, g, me.ID, "R"); n != 2 {
		t.Fatalf("want two colour picks, got %d", n)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}
}

// #2611: a token Marwyn that dies while its counter trigger waits on
// the stack has ceased to exist, so the trigger's AddCounter finds no
// card anywhere. "Put a counter on this" with nothing to put it on
// does nothing (CR 608.2b); it must not surface as an effect error.
func TestB10MarwynTokenGoneBeforeTriggerResolvesDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	marwyn := effectstest.PushCatalogPermanent(g, me.ID, "Marwyn, the Nurturer", "Token Legendary Creature — Elf Druid", effectstest.B10MarwynOracle, false)

	effectstest.CastCatalogSpell(t, g, "Elf", "Creature — Elf Warrior", "", nil)
	for i := 0; i < 8 && effectstest.TriggerOnStack(g, marwyn) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.TriggerOnStack(g, marwyn) == nil {
		t.Fatal("Marwyn's counter trigger never reached the stack")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(marwyn); err != nil {
			t.Errorf("destroy Marwyn: %v", err)
		}
	})
	g.RunStateChecksForTest() // CR 704.5d: the dead token ceases to exist
	if _, ok := g.LookupCardForEffect(marwyn); ok {
		t.Fatal("test setup: the token Marwyn should have ceased to exist")
	}
	before := len(g.Events)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the trigger should do nothing", n)
	}
}
