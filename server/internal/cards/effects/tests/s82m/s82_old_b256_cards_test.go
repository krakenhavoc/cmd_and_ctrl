package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b256_cards_test.go — Commander 2016, batch 256 (S82 older sets,
// #5537). Every test plays the real catalog entry through the engine.

const (
	oldB256VulturousOracle   = "11a28f60-2456-4ed5-b1cc-03c395f55e59"
	oldB256SolidarityOracle  = "3319b2df-7354-44b5-afa1-b9bc4a577254"
	oldB256NecroplasmOracle  = "aa1c40fb-c230-4e7f-99b1-4e2e8c39a168"
	oldB256BloodTyrantOracle = "ac74bd52-a450-486f-8161-816fd8d966d0"
	oldB256EliteOracle       = "ba6c3c72-c014-45c6-a0b4-59eb9a65303e"
	oldB256ShankerOracle     = "22052572-0b7d-4dd8-84a8-e2658f77b85e"
	oldB256WildernessOracle  = "478d1acf-93fa-4695-83be-693b839ebf80"
	oldB256ChimeraOracle     = "9745fc92-4955-48e3-a6c8-1300891cf078"
	oldB256NathOracle        = "87a44a87-a2a5-4193-a163-e4846e5b8387"
	oldB256MimicOracle       = "88929ea9-900f-4dbb-b16c-cf3bad4e410c"
	oldB256HorseOracle       = "8d0213d7-9e25-49d2-93fa-1d346cb2e869"
)

func TestOldB256EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB256VulturousOracle, oldB256SolidarityOracle, oldB256NecroplasmOracle, oldB256BloodTyrantOracle,
		oldB256EliteOracle, oldB256ShankerOracle, oldB256WildernessOracle, oldB256ChimeraOracle,
		oldB256NathOracle, oldB256MimicOracle, oldB256HorseOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func oldB256Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB256SpellheartChimeraCountsYourInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	chimera := effectstest.B12Push(g, me.ID, "Spellheart Chimera", "Creature — Chimera", oldB256ChimeraOracle, 0, 3)
	if p, _ := effectstest.PtOf(t, g, chimera); p != 0 {
		t.Fatalf("power %d with an empty graveyard, want 0", p)
	}
	effectstest.CmdrB17ToGraveyard(t, g, me, game.Card{Name: "Opt", TypeLine: "Instant"})
	effectstest.CmdrB17ToGraveyard(t, g, me, game.Card{Name: "Divination", TypeLine: "Sorcery"})
	effectstest.CmdrB17ToGraveyard(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.CmdrB17ToGraveyard(t, g, opp, game.Card{Name: "Their Bolt", TypeLine: "Instant"})
	if p, tt := effectstest.PtOf(t, g, chimera); p != 2 || tt != 3 {
		t.Errorf("the Chimera is %d/%d, want 2/3", p, tt)
	}
}

func TestOldB256WildernessElementalCountsOpposingNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	elemental := effectstest.B12Push(g, me.ID, "Wilderness Elemental", "Creature — Elemental", oldB256WildernessOracle, 0, 3)
	effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Temple", TypeLine: "Land"})
	effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Snow Mountain", TypeLine: "Snow Land — Mountain"})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Temple", TypeLine: "Land"})
	if p, tt := effectstest.PtOf(t, g, elemental); p != 2 || tt != 3 {
		t.Errorf("the Elemental is %d/%d, want 2/3 (two nonbasic lands of theirs)", p, tt)
	}
}

func TestOldB256AnkleShankerGivesEveryCreatureYouControlFirstStrikeAndDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	shanker := effectstest.B12Push(g, me.ID, "Ankle Shanker", "Creature — Goblin Berserker", oldB256ShankerOracle, 2, 2)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(shanker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{shanker, ally} {
		for _, kw := range []string{"first strike", "deathtouch"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", id, kw)
			}
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature gained deathtouch")
	}
}

func TestOldB256SolidarityOfHeroesDoublesCountersOnEachTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, game.CounterPlusOne, 3)
		_ = g.AddCounterForEffect(b, game.CounterPlusOne, 1)
	})
	spell := effectstest.HandCardOf(g, me, "Solidarity of Heroes", "Instant", "{1}{G}", oldB256SolidarityOracle)
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(a, b)}); err == nil {
		t.Fatal("two targets were cast for the price of one (strive)")
	}
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(a, b)}); err != nil {
		t.Fatalf("cast with two targets for {1}{G} + {1}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, a); got != 6 {
		t.Errorf("Bear A has %d counters, want 6", got)
	}
	if got := effectstest.PlusOnes(g, b); got != 2 {
		t.Errorf("Bear B has %d counters, want 2", got)
	}
}

func oldB256NecroplasmTable(t *testing.T, counters int) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	necro := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Necroplasm", OracleID: oldB256NecroplasmOracle,
		TypeLine: "Creature — Ooze", ManaCost: "{1}{B}{B}", Power: 1, Toughness: 1})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(necro, game.CounterPlusOne, counters) })
	victim := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{1}", Power: 2, Toughness: 2})
	spared := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Four Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{2}", Power: 2, Toughness: 2})
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return g, necro, victim, spared
}

func TestOldB256NecroplasmDestroysCreaturesWithManaValueEqualToItsCounters(t *testing.T) {
	g, necro, victim, spared := oldB256NecroplasmTable(t, 2)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the creature with mana value 2 survived two counters")
	}
	if _, ok := effectstest.BattlefieldCard(g, spared); !ok {
		t.Error("a creature with another mana value was destroyed")
	}
	if _, ok := effectstest.BattlefieldCard(g, necro); !ok {
		t.Error("Necroplasm (mana value 3, 2 counters) died")
	}
}

func TestOldB256NecroplasmDestroysItselfAtThreeCounters(t *testing.T) {
	g, necro, victim, _ := oldB256NecroplasmTable(t, 3)
	if _, ok := effectstest.BattlefieldCard(g, necro); ok {
		t.Error("Necroplasm survived its own mana value in counters")
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); !ok {
		t.Error("a mana value 2 creature died to three counters")
	}
}

func TestOldB256NecroplasmGetsACounterAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	necro := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Necroplasm", OracleID: oldB256NecroplasmOracle,
		TypeLine: "Creature — Ooze", ManaCost: "{1}{B}{B}", Power: 1, Toughness: 1})
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, necro); got != 1 {
		t.Errorf("%d counters after the upkeep, want 1", got)
	}
}

func TestOldB256VulturousZombieGrowsOnlyForCardsInOpposingGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	zombie := effectstest.B12Push(g, me.ID, "Vulturous Zombie", "Creature — Plant Zombie", oldB256VulturousOracle, 3, 3)
	mine := effectstest.PushCatalogHandCard(me, "My Card", "Sorcery", "")
	theirs := effectstest.PushCatalogHandCard(opp, "Their Card", "Sorcery", "")
	g.WithWriteLock(func() { g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: me.ID, Source: zombie, N: 1}) })
	effectstest.AnswerDiscard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, zombie); got != 0 {
		t.Fatalf("my own discard gave %d counters", got)
	}
	g.WithWriteLock(func() { g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp.ID, Source: zombie, N: 1}) })
	effectstest.AnswerDiscard(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, zombie); got != 1 {
		t.Errorf("an opponent's discard gave %d counters, want 1", got)
	}
}

func TestOldB256AcademyEliteEntersWithACounterPerInstantAndSorceryInAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.GyCard(me, "Opt", "Instant", "{U}", "")
	effectstest.GyCard(opp, "Their Bolt", "Instant", "{R}", "")
	effectstest.GyCard(opp, "Their Divination", "Sorcery", "{2}{U}", "")
	effectstest.GyCard(opp, "Their Bear", "Creature — Bear", "{1}", "")
	effectstest.CastCatalogSpell(t, g, "Academy Elite", "Creature — Human Wizard", oldB256EliteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var elite uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Academy Elite" {
			elite = c.InstanceID
		}
	}
	if elite == uuid.Nil {
		t.Fatal("Academy Elite is not on the battlefield")
	}
	if got := effectstest.PlusOnes(g, elite); got != 3 {
		t.Errorf("entered with %d counters, want 3", got)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "U", 3)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, elite, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, elite); got != 2 {
		t.Errorf("%d counters left, want 2", got)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a card drawn and a card discarded", hand, me.Hand.Size())
	}
}

func TestOldB256BloodTyrantDrainsEveryoneAtUpkeepAndGrowsByTheLifeLost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	tyrant := effectstest.B12Push(g, me.ID, "Blood Tyrant", "Creature — Vampire", oldB256BloodTyrantOracle, 5, 5)
	life, theirs := me.Life, opp.Life
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 || opp.Life != theirs-1 {
		t.Errorf("life %d -> %d and %d -> %d, want one lost each", life, me.Life, theirs, opp.Life)
	}
	if got := effectstest.PlusOnes(g, tyrant); got != len(g.Seats) {
		t.Errorf("%d counters, want %d (one per player)", got, len(g.Seats))
	}
}

func TestOldB256BloodTyrantGetsFiveCountersWhenAPlayerLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	tyrant := effectstest.B12Push(g, me.ID, "Blood Tyrant", "Creature — Vampire", oldB256BloodTyrantOracle, 5, 5)
	g.WithWriteLock(func() { opp.Life = 0 })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, tyrant); got != 5 {
		t.Errorf("%d counters after a player lost, want 5", got)
	}
}

func TestOldB256NathMakesElvesWhenAnOpponentDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	nath := effectstest.B12Push(g, me.ID, "Nath of the Gilt-Leaf", "Legendary Creature — Elf Warrior", oldB256NathOracle, 4, 4)
	card := effectstest.PushCatalogHandCard(opp, "Their Card", "Sorcery", "")
	g.WithWriteLock(func() { g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: opp.ID, Source: nath, N: 1}) })
	effectstest.AnswerDiscard(t, g, opp.ID, card)
	effectstest.PassPriorityUntilChoice(t, g)
	if p := effectstest.LatestTriggerPromptFor(g, me.ID); p != nil {
		if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB256Count(g, me.ID, "Elf Warrior"); got != 1 {
		t.Errorf("%d Elf Warrior tokens, want 1", got)
	}
}

func TestOldB256AkroanHorseGoesToAnOpponentAndMakesYouSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.CastCatalogSpell(t, g, "Akroan Horse", "Artifact Creature — Horse", oldB256HorseOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) != nil {
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB256Count(g, opp.ID, "Akroan Horse"); got != 1 {
		t.Fatalf("the chosen opponent controls %d Horses, want 1", got)
	}
	// On the new controller's upkeep, each of their opponents gets a Soldier.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB256Count(g, me.ID, "Soldier") != 1 {
		t.Errorf("the old controller has %d Soldiers, want 1", oldB256Count(g, me.ID, "Soldier"))
	}
	if oldB256Count(g, opp.ID, "Soldier") != 0 {
		t.Error("the Horse's controller got a Soldier")
	}
}

func TestOldB256ProgenitorMimicCopiesACreatureAndMakesTokenCopiesEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	ogre := effectstest.SeedCopyableCreature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Progenitor Mimic", "Creature — Shapeshifter", oldB256MimicOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, ogre)
	if got := oldB256Count(g, me.ID, "Their Ogre"); got != 1 {
		t.Fatalf("I control %d copies of the Ogre, want 1", got)
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB256Count(g, me.ID, "Their Ogre"); got != 2 {
		t.Errorf("I control %d Ogres after the upkeep, want 2 (the copy and its token)", got)
	}
}

func TestOldB256NathUpkeepMakesATargetOpponentDiscardAtRandomAndYouMakeAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Nath of the Gilt-Leaf", "Legendary Creature — Elf Warrior", oldB256NathOracle, 4, 4)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PushCatalogHandCard(opp, "Their Only Card", "Sorcery", "")
	for len(opp.Hand.Cards) > 1 {
		opp.Hand.Cards = opp.Hand.Cards[1:]
	}
	for i := 0; i < 10; i++ {
		effectstest.PassPriorityUntilChoice(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
			continue
		}
		if p := effectstest.LatestTriggerPromptFor(g, me.ID); p != nil {
			if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		break
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != 0 {
		t.Errorf("the opponent still has %d cards in hand", opp.Hand.Size())
	}
	if got := oldB256Count(g, me.ID, "Elf Warrior"); got != 1 {
		t.Errorf("%d Elf Warrior tokens after the discard, want 1", got)
	}
}
