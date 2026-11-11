package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b24_test.go — Commander deck cards, batch 24 (#5368).

const (
	cmdrB24PriestOracle    = "524e5098-9c6d-43ad-80cb-52fd6d14ef53"
	cmdrB24MentorOracle    = "6eb5193b-5ef0-4261-acdc-958e161c8118"
	cmdrB24MonarchOracle   = "626a979f-bed7-4cb1-9ab7-c3f2bbe37781"
	cmdrB24DuneyardOracle  = "48edc348-93f6-4dce-9cc4-7244d76b6f4a"
	cmdrB24EndgameOracle   = "fedb1947-9fa5-4655-8d02-b370fa63dd16"
	cmdrB24AugurOracle     = "06eb43af-637a-4f73-8071-a7036d4c8c3d"
	cmdrB24CryptbreakerOrc = "16f75a65-d067-4f79-81b9-833c373cf4d2"
	cmdrB24ZombieMasterOrc = "5446c92f-ff22-4e9b-a2f6-e64c8560c1e0"
	cmdrB24DruidOracle     = "5bb32efd-9e58-4021-bbda-4ffccfa1d601"
	cmdrB24AngelOracle     = "3c49e98d-d33b-47fa-8b07-b38b367de2c7"
)

func TestCmdrB24CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdrB24PriestOracle:    "Priest of the Crossing",
		cmdrB24MentorOracle:    "Wizened Mentor",
		cmdrB24MonarchOracle:   "Lost Monarch of Ifnir",
		cmdrB24DuneyardOracle:  "Accursed Duneyard",
		cmdrB24EndgameOracle:   "Commence the Endgame",
		cmdrB24AugurOracle:     "Corpse Augur",
		cmdrB24CryptbreakerOrc: "Cryptbreaker",
		cmdrB24ZombieMasterOrc: "Zombie Master",
		cmdrB24DruidOracle:     "Druid of Purification",
		cmdrB24AngelOracle:     "Angel of Sanctions",
	} {
		cmdrB21AssertFull(t, oracle, name)
	}
}

func TestCmdrB24PriestOfTheCrossingCountsCreaturesThatDiedUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	priest := b12Push(g, me.ID, "Priest of the Crossing", "Creature — Zombie Bird Cleric", cmdrB24PriestOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := b12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	killCreature(t, g, me.ID, bear2)
	killCreature(t, g, opp.ID, theirs)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{priest, bear, token} {
		if got := plusOneCounters(g, id); got != 1 {
			t.Errorf("creature %s has %d counters, want 1 (one creature died under my control)", id, got)
		}
	}
}

func TestCmdrB24PriestDoesNothingWhenNothingDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	priest := b12Push(g, me.ID, "Priest of the Crossing", "Creature — Zombie Bird Cleric", cmdrB24PriestOracle, 3, 3)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, priest); got != 0 {
		t.Errorf("%d counters with no deaths", got)
	}
}

func TestCmdrB24WizenedMentorMakesAZombieOnceEachTurnForAnOpponentsActivation(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Wizened Mentor", "Creature — Zombie Cleric", cmdrB24MentorOracle, 2, 2)
	myKrenko := actPushKrenko(g, me.ID)
	theirKrenko := actPushKrenko(g, them.ID)
	theirElves := b21Push(g, them.ID, "Llanowar Elves", "Creature — Elf Druid", actLlanowarElvesOracle, 1, 1, "G")
	zombies := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie" && c.Controller == me.ID && c.IsToken() {
				n++
			}
		}
		return n
	}
	// Own activation: nothing.
	if err := g.ActivateCatalogAbility(me.ID, myKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if zombies() != 0 {
		t.Error("my own activation made a Zombie")
	}
	// A mana ability: nothing.
	if err := g.ActivateManaAbility(them.ID, theirElves, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if zombies() != 0 {
		t.Error("an opponent's mana ability made a Zombie")
	}
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if zombies() != 1 {
		t.Fatalf("%d Zombies after an opponent's activation, want 1", zombies())
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirKrenko {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(them.ID, theirKrenko, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if zombies() != 1 {
		t.Errorf("%d Zombies after a second activation the same turn, want still 1", zombies())
	}
}

func TestCmdrB24LostMonarchGrantsAfflictAndReturnsACreatureAfterZombieDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monarch := b12Push(g, me.ID, "Lost Monarch of Ifnir", "Creature — Zombie Noble", cmdrB24MonarchOracle, 4, 4)
	zombie := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 9)
	dead := cmdrB22GraveyardCreature(me, "Fallen", "{2}")
	life := opp.Life
	brAttack(t, g, monarch, zombie, human)
	if err := g.DeclareBlocker(blocker, zombie); err != nil {
		t.Fatalf("block: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 lost to the blocked Zombie's afflict", life, opp.Life)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil {
		t.Fatal("no creature pick in the second main phase")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the chosen creature card is not in hand")
	}
}

func TestCmdrB24LostMonarchDoesNothingWithoutZombieCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Lost Monarch of Ifnir", "Creature — Zombie Noble", cmdrB24MonarchOracle, 4, 4)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	cmdrB22GraveyardCreature(me, "Fallen", "{2}")
	library := me.Library.Size()
	brAttack(t, g, human)
	_ = opp
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Errorf("the library shrank %d -> %d though no Zombie dealt damage", library, me.Library.Size())
	}
}

func TestCmdrB24AccursedDuneyardRegeneratesATargetZombie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	yard := pushCatalogPermanent(g, me.ID, "Accursed Duneyard", "Land", cmdrB24DuneyardOracle, false)
	zombie := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, yard, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a Bear was a legal target")
	}
	b16Activate(t, g, me.ID, yard, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: zombie}}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(zombie) })
	if _, still := battlefieldCard(g, zombie); !still {
		t.Error("the regenerated Zombie was destroyed")
	}
}

func TestCmdrB24CommenceTheEndgameAmassesByHandSize(t *testing.T) {
	spec, _ := Lookup(cmdrB24EndgameOracle)
	if !spec.CantBeCountered {
		t.Error("Commence the Endgame can't be countered")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	for i := 0; i < 3; i++ {
		pushHandCardWithManaCost(me, "Filler", "Instant", "{1}")
	}
	castFromHand(t, g, game.Card{Name: "Commence the Endgame", TypeLine: "Instant", ManaCost: "{4}{U}{U}", OracleID: cmdrB24EndgameOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 5 {
		t.Fatalf("hand %d, want 5 (3 + 2 drawn)", me.Hand.Size())
	}
	var army game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Army") {
			army = c
		}
	}
	if army.InstanceID == uuid.Nil {
		t.Fatal("no Army token")
	}
	if got := plusOneCounters(g, army.InstanceID); got != 5 {
		t.Errorf("Army has %d counters, want 5 (cards in hand)", got)
	}
	if !army.HasSubtype("Zombie") {
		t.Error("the Army should be a Zombie")
	}
}

func TestCmdrB24CorpseAugurDrawsAndLosesByCreatureCardsInATargetGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	augur := b12Push(g, me.ID, "Corpse Augur", "Creature — Zombie Wizard", cmdrB24AugurOracle, 4, 2)
	for i := 0; i < 3; i++ {
		cmdrB22GraveyardCreature(opp, "Their dead", "{1}")
	}
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	killCreature(t, g, me.ID, augur)
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

func TestCmdrB24CryptbreakerMakesZombiesAndTapsThreeForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cb := b12Push(g, me.ID, "Cryptbreaker", "Creature — Zombie", cmdrB24CryptbreakerOrc, 1, 1)
	advanceToMain(t, g)
	discard := pushHandCardWithManaCost(me, "Junk", "Sorcery", "{1}")
	apaMana(me, "C", "B")
	b16Activate(t, g, me.ID, cb, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard cost was not paid")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.IsToken() && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no 2/2 Zombie token")
	}
	// Two Zombies (Cryptbreaker, the token) are not three; add a third and
	// tap all three, Cryptbreaker included.
	z3 := b12Creature(g, me.ID, "Zombie 3", "Creature — Zombie", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cb {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	hand, life := me.Hand.Size(), me.Life
	var zs []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Zombie") {
			zs = append(zs, c.InstanceID)
		}
	}
	if len(zs) != 3 || !hasID(zs, z3) {
		t.Fatalf("zombies = %v", zs)
	}
	if err := g.ActivateCatalogAbility(me.ID, cb, 1, game.ActivateAbilityParams{Strict: true, TapIDs: zs[:2]}); err == nil {
		t.Fatal("two Zombies paid a three-Zombie cost")
	}
	b16Activate(t, g, me.ID, cb, 1, game.ActivateAbilityParams{Strict: true, TapIDs: zs})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d; want +1 card and -1 life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestCmdrB24ZombieMasterGivesOtherZombiesSwampwalkAndRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zm := b12Push(g, me.ID, "Zombie Master", "Creature — Zombie", cmdrB24ZombieMasterOrc, 2, 3)
	mine := b12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !hasEffectiveKeyword(t, g, id, "swampwalk") {
			t.Errorf("Zombie %s lacks swampwalk", id)
		}
	}
	if hasEffectiveKeyword(t, g, zm, "swampwalk") || hasEffectiveKeyword(t, g, bear, "swampwalk") {
		t.Error("Zombie Master itself or a non-Zombie has swampwalk")
	}
	advanceToMain(t, g)
	apaMana(me, "B")
	b16Activate(t, g, me.ID, mine, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if _, still := battlefieldCard(g, mine); !still {
		t.Error("the Zombie with the granted regeneration was destroyed")
	}
	if err := g.ActivateCatalogAbility(me.ID, zm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Zombie Master itself has the regeneration ability")
	}
}

func TestCmdrB24DruidOfPurificationEveryPlayerMayChooseAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	push := func(owner uuid.UUID, name, tl string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: tl,
			Owner: owner, Controller: owner})
	}
	rock := push(opp1.ID, "Their Rock", "Artifact")
	aura := push(opp2.ID, "Their Aura", "Enchantment")
	spare := push(opp2.ID, "Spare Rock", "Artifact")
	mine := push(me.ID, "My Rock", "Artifact")
	land := push(opp1.ID, "Their Land", "Land")
	castAndResolveCreature(t, g, "Druid of Purification", "Creature — Human Druid", cmdrB24DruidOracle)
	passPriorityAroundTable(t, g)
	ask := func(who uuid.UUID) *game.PendingChoice {
		t.Helper()
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Chooser == who && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
				return pc
			}
		}
		t.Fatalf("no permanent prompt for %s", who)
		return nil
	}
	answer := func(c *game.PendingChoice, picks ...uuid.UUID) {
		t.Helper()
		var err error
		if c.Kind == game.PendingChoiceOwnPermanents {
			err = g.ResolveOwnPermanents(c.ID, c.Chooser, picks)
		} else {
			err = g.ResolveTheirPermanents(c.ID, c.Chooser, picks)
		}
		if err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	first := ask(me.ID)
	if hasID(first.ChooseCards, mine) || hasID(first.ChooseCards, land) || !hasID(first.ChooseCards, rock) || !hasID(first.ChooseCards, aura) {
		t.Fatalf("first pool = %v", first.ChooseCards)
	}
	answer(first, rock)
	second := ask(opp1.ID)
	if hasID(second.ChooseCards, rock) {
		t.Errorf("the second player may re-choose the first pick: %v", second.ChooseCards)
	}
	answer(second, aura)
	for i := 0; i < 8; i++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
				c = pc
			}
		}
		if c == nil {
			break
		}
		answer(c)
	}
	passPriorityAroundTable(t, g)
	for _, gone := range []uuid.UUID{rock, aura} {
		if _, still := battlefieldCard(g, gone); still {
			t.Errorf("chosen permanent %s survived", gone)
		}
	}
	for _, kept := range []uuid.UUID{spare, mine, land} {
		if _, still := battlefieldCard(g, kept); !still {
			t.Errorf("unchosen permanent %s was destroyed", kept)
		}
	}
}

func TestCmdrB24AngelOfSanctionsExilesUntilItLeavesAndEmbalms(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	angel := castAndResolveCreature(t, g, "Angel of Sanctions", "Creature — Angel", cmdrB24AngelOracle)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, land) || hasID(p.PickTargetCards, mine) {
			t.Errorf("target pool = %v, want only the opponent's nonland permanents", p.PickTargetCards)
		}
	}
	pickCard(t, g, me.ID, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Fatal("the target was not exiled")
	}
	killCreature(t, g, me.ID, angel)
	passPriorityAroundTable(t, g)
	if len(battlefieldIDsNamed(g, "Their Bear")) != 1 {
		t.Error("the exiled creature did not return when the Angel left")
	}
	// Embalm {5}{W} from the graveyard.
	apaMana(me, "C", "C", "C", "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, angel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("embalm: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel of Sanctions" && c.IsToken() {
			found = true
			if !c.HasSubtype("Zombie") || !c.HasSubtype("Angel") {
				t.Errorf("embalmed token = %q, want a Zombie Angel", c.TypeLine)
			}
		}
	}
	if !found {
		t.Error("no embalmed token")
	}
}
