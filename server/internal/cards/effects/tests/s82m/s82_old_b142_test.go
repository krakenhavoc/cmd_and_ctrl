package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b142_test.go — Throne of Eldraine, S82 older-sets batch 142
// (#5516): card draw and library. Each test plays the real catalog entry.

const (
	oldB142AdversaryOracle = "3afe4852-d506-45d7-8a19-ddcbc3dd90d4"
	oldB142StalwartsOracle = "9171248e-045e-4ed4-aa06-eea3ae3d8df0"
	oldB142OgreOracle      = "980a0d4d-b070-46c1-be8d-31bdda0d1c33"
	oldB142GadwickOracle   = "c7530ecd-e22f-41e7-a76b-de32809901c8"
	oldB142GwynOracle      = "ca411184-3413-492c-ab87-08c9fa1d4806"
	oldB142HospitalityOrcl = "d804b0e5-785a-4e6e-9fe7-38b8add826fa"
	oldB142InnkeeperOracle = "e3d7cea4-0d4c-4320-966d-17ef69b5e433"
	oldB142DragonOracle    = "e40d47cf-6232-41ec-8865-d12a335dfcee"
	oldB142MerchantOracle  = "f88a096e-7497-4297-ab65-f18769b1cc47"
	oldB142BonesplitterOrc = "452e3f5f-ce17-4682-966b-5cc100210aee"
)

func TestOldB142EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB142AdversaryOracle: "Oakhame Adversary", oldB142StalwartsOracle: "Rowan's Stalwarts",
		oldB142OgreOracle: "Skullknocker Ogre", oldB142GadwickOracle: "Gadwick, the Wizened",
		oldB142GwynOracle: "Syr Gwyn, Hero of Ashvale", oldB142HospitalityOrcl: "Oko's Hospitality",
		oldB142InnkeeperOracle: "Edgewall Innkeeper", oldB142DragonOracle: "Loch Dragon",
		oldB142MerchantOracle: "Merchant of the Vale", oldB142MerchantOracle + "#1": "Haggle",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Oakhame Adversary ----------------------------------------------------

func oldB142CastAdversary(g *game.Game, me *game.Player, pool int) error {
	card := game.Card{Name: "Oakhame Adversary", TypeLine: "Creature — Elf Warrior", ManaCost: "{3}{G}", OracleID: oldB142AdversaryOracle, Power: 2, Toughness: 3}
	return effectstest.SocB19CastWithPool(g, me, card, "G", pool, game.CastSpellParams{})
}

func TestOldB142OakhameAdversaryCostsTwoLessOnlyIfAnOpponentControlsAGreenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	if err := oldB142CastAdversary(g, me, 2); err == nil {
		t.Fatal("cast for {1}{G} with no green permanent on the opponent's side")
	}
	// Your own green permanent does not count.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := oldB142CastAdversary(g, me, 2); err == nil {
		t.Fatal("your own green permanent discounted it")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	if err := oldB142CastAdversary(g, me, 2); err != nil {
		t.Fatalf("cast for {1}{G} with an opposing green permanent: %v", err)
	}
}

func TestOldB142OakhameAdversaryDrawsOnCombatDamageToAPlayerAndHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adv := effectstest.B12Push(g, me.ID, "Oakhame Adversary", "Creature — Elf Warrior", oldB142AdversaryOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, adv, "deathtouch") {
		t.Error("no deathtouch")
	}
	before := effectstest.HandCount(g, me)
	effectstest.DealCombatDamageToPlayer(g, adv, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

// --- Rowan's Stalwarts ----------------------------------------------------

func TestOldB142RowansStalwartsFetchesRowanFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	rowan := effectstest.PushGraveyardCardWithTypeLineID(me, "Rowan, Fearless Sparkmage", "Legendary Planeswalker — Rowan")
	effectstest.CastCatalogSpell(t, g, "Rowan's Stalwarts", "Creature — Human Knight", oldB142StalwartsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, rowan)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rowan) {
		t.Fatal("Rowan was not fetched from the graveyard")
	}
}

func TestOldB142RowansStalwartsDecliningFetchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	rowan := effectstest.PushGraveyardCardWithTypeLineID(me, "Rowan, Fearless Sparkmage", "Legendary Planeswalker — Rowan")
	effectstest.CastCatalogSpell(t, g, "Rowan's Stalwarts", "Creature — Human Knight", oldB142StalwartsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(rowan) {
		t.Fatal("Rowan was fetched after declining")
	}
}

// --- Skullknocker Ogre ----------------------------------------------------

func TestOldB142SkullknockerOgreMakesTheOpponentDiscardAtRandomThenDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Skullknocker Ogre", "Creature — Ogre", oldB142OgreOracle, 4, 3)
	effectstest.D2dSetHand(opp, 2)
	gy := len(opp.Graveyard.Cards)
	effectstest.DealCombatDamageToPlayer(g, ogre, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, opp); got != 2 {
		t.Errorf("opponent hand %d, want 2 (discarded one, drew one)", got)
	}
	if got := len(opp.Graveyard.Cards) - gy; got != 1 {
		t.Errorf("%d cards entered the graveyard, want 1", got)
	}
}

func TestOldB142SkullknockerOgreWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Skullknocker Ogre", "Creature — Ogre", oldB142OgreOracle, 4, 3)
	effectstest.D2dSetHand(opp, 0)
	effectstest.DealCombatDamageToPlayer(g, ogre, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, opp); got != 0 {
		t.Errorf("opponent hand %d, want 0: no discard, so no draw", got)
	}
}

func TestOldB142SkullknockerOgreIgnoresDamageToItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ogre := effectstest.B12Push(g, me.ID, "Skullknocker Ogre", "Creature — Ogre", oldB142OgreOracle, 4, 3)
	effectstest.D2dSetHand(me, 2)
	effectstest.DealCombatDamageToPlayer(g, ogre, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards); got != 0 {
		t.Errorf("%d cards discarded for damage to its own controller", got)
	}
}

// --- Gadwick, the Wizened -------------------------------------------------

func TestOldB142GadwickDrawsXCardsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.D2dSetHand(me, 0)
	effectstest.PaCast(t, g, "Gadwick, the Wizened", "Legendary Creature — Human Wizard", "{X}{U}{U}{U}", oldB142GadwickOracle,
		game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != 3 {
		t.Errorf("hand %d, want 3", got)
	}
}

func TestOldB142GadwickTapsAnOpponentsNonlandPermanentWhenYouCastABlueSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gadwick, the Wizened", "Legendary Creature — Human Wizard", oldB142GadwickOracle, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Blue Thing", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}}
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// The trigger asks for its target; the land is not a legal one.
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Error("a land was a legal target")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("the Bear was not tapped")
	}
}

// --- Syr Gwyn, Hero of Ashvale -------------------------------------------

func TestOldB142SyrGwynGrantsEquipKnightZeroToYourEquipmentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Syr Gwyn, Hero of Ashvale", "Legendary Creature — Human Knight", oldB142GwynOracle, 5, 5)
	knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	bear := effectstest.SeedBear(g, me.ID)
	blade := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", oldB142BonesplitterOrc)
	theirs := effectstest.SeedEquipment(g, opp.ID, "Bonesplitter", oldB142BonesplitterOrc)
	if idx, _ := effectstest.BeGrantedRow(t, g, theirs, "Equip Knight {0}"); idx >= 0 {
		t.Error("an opponent's Equipment got equip Knight {0}")
	}
	idx, ref := effectstest.BeGrantedRow(t, g, blade, "Equip Knight {0}")
	if idx < 0 {
		t.Fatal("your Equipment has no granted equip Knight {0}")
	}
	if err := g.ActivateCatalogAbility(me.ID, blade, idx, game.ActivateAbilityParams{Ref: ref,
		Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("equip Knight {0} targeted a non-Knight")
	}
	effectstest.BeEquip(t, g, me.ID, blade, knight, "Equip Knight {0}")
	if p := effectstest.EffectivePower(t, g, knight); p != 4 {
		t.Errorf("the equipped Knight has %d power, want 4", p)
	}
}

func TestOldB142SyrGwynIgnoresAnUnequippedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Syr Gwyn, Hero of Ashvale", "Legendary Creature — Human Knight", oldB142GwynOracle, 5, 5)
	knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	plain := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Plain", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	blade := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", oldB142BonesplitterOrc)
	effectstest.BeEquip(t, g, me.ID, blade, knight, "Equip Knight {0}")
	hand, life := effectstest.HandCount(g, me), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandCount(g, me) != hand || me.Life != life {
		t.Fatal("an unequipped attacker triggered Syr Gwyn")
	}
}

func TestOldB142SyrGwynDrawsAndLosesLifeWhenAnEquippedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Syr Gwyn, Hero of Ashvale", "Legendary Creature — Human Knight", oldB142GwynOracle, 5, 5)
	knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	blade := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", oldB142BonesplitterOrc)
	effectstest.BeEquip(t, g, me.ID, blade, knight, "Equip Knight {0}")
	hand, life := effectstest.HandCount(g, me), me.Life
	effectstest.DeclareAttack(t, g, opp.ID, knight)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Oko's Hospitality ----------------------------------------------------

func TestOldB142OkosHospitalitySetsYourCreaturesToThreeThreeAndFetchesOko(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 6, 6)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	oko := effectstest.PushGraveyardCardWithTypeLineID(me, "Oko, the Trickster", "Legendary Planeswalker — Oko")
	effectstest.CastCatalogSpell(t, g, "Oko's Hospitality", "Instant", oldB142HospitalityOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, oko)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, small} {
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("your creature is %d/%d, want 3/3", p, tough)
		}
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
	if !me.Hand.Contains(oko) {
		t.Error("Oko was not fetched")
	}
}

func TestOldB142OkosHospitalityDoesNotChangeACreatureThatEntersLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Oko's Hospitality", "Instant", oldB142HospitalityOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Giant", 6, 6)
	if p, _ := effectstest.EffectivePTOf(t, g, late); p != 6 {
		t.Errorf("a creature that entered afterwards is %d power, want 6", p)
	}
}

// --- Adventure creatures --------------------------------------------------

func oldB142CastAdventureCreature(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
}

func TestOldB142EdgewallInnkeeperDrawsForACreatureSpellWithAnAdventureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Edgewall Innkeeper", "Creature — Human Peasant", oldB142InnkeeperOracle, 1, 1)
	effectstest.B13PlayAs(t, g, 0, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.HandCount(g, me)
	oldB142CastAdventureCreature(t, g, me)
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestOldB142EdgewallInnkeeperIgnoresTheAdventureHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Edgewall Innkeeper", "Creature — Human Peasant", oldB142InnkeeperOracle, 1, 1)
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	before := effectstest.HandCount(g, me)
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)
	if got := effectstest.HandCount(g, me); got != before-1 {
		t.Errorf("hand %d, want %d (only the cast card left)", got, before-1)
	}
}

// --- Loch Dragon ----------------------------------------------------------

func TestOldB142LochDragonLootsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.D2dSetHand(me, 2)
	effectstest.CastCatalogSpell(t, g, "Loch Dragon", "Creature — Dragon", oldB142DragonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.DiscardChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no discard prompt on entering")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != 2 {
		t.Errorf("hand %d, want 2 (discard one, draw one)", got)
	}
}

func TestOldB142LochDragonDeclinesTheLootOnAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dragon := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Loch Dragon", TypeLine: "Creature — Dragon",
		OracleID: oldB142DragonOracle, Power: 3, Toughness: 2})
	effectstest.D2dSetHand(me, 2)
	effectstest.DeclareAttack(t, g, opp.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on attacking")
	}
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != 2 {
		t.Errorf("declining changed the hand to %d, want 2", got)
	}
}

// --- Merchant of the Vale // Haggle ---------------------------------------

func TestOldB142MerchantOfTheValeDiscardsAsACostToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	merchant := effectstest.B12Push(g, me.ID, "Merchant of the Vale", "Creature — Human Peasant", oldB142MerchantOracle, 2, 3)
	effectstest.D2dSetHand(me, 0)
	effectstest.ApaMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, merchant, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.D2dSetHand(me, 1)
	filler := me.Hand.Cards[0].InstanceID
	effectstest.ApaMana(me, "R", "C", "C")
	effectstest.ApaActivate(t, g, me, merchant, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{filler}})
	if me.Hand.Contains(filler) {
		t.Error("the discard was not paid as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != 1 {
		t.Errorf("hand %d, want 1 (discarded one, drew one)", got)
	}
}

func TestOldB142HaggleLootsOptionallyAndGoesOnAnAdventure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, oldB142MerchantOracle,
		effectstest.HobA6Face("Merchant of the Vale", "Creature — Human Peasant", "{2}{R}", 2, 3, "R"),
		effectstest.HobA6Face("Haggle", "Instant — Adventure", "{R}", 0, 0, "R"))
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler", Owner: me.ID, Controller: me.ID})
	}
	start := effectstest.HandCount(g, me)
	effectstest.B10AddMana(me, "R")
	if err := effectstest.HobA6Cast(g, me, id, 1, nil); err != nil {
		t.Fatalf("cast Haggle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no optional discard prompt")
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != start-1 {
		t.Errorf("hand %d, want %d (Haggle cast, one discarded, one drawn)", got, start-1)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}
