package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b504_test.go — Alara Reborn, older-sets batch 504 (#5586). Every test
// plays the real catalog entry through the engine.

const (
	oldB504Defiler  = "34894b92-7ebd-49be-9a67-7dc6428aaab9"
	oldB504Hellion  = "7ef3f240-05a5-4fc9-a977-a729010a53c2"
	oldB504Demon    = "91d0a557-7a80-46d5-b823-392a84a6be44"
	oldB504Shield   = "6d9aaa70-b2cc-424c-8095-a0b2fa272db3"
	oldB504Whip     = "a588f047-02b4-4c6b-bce2-6e4d564296ec"
	oldB504Unscythe = "d16976b8-4eda-4415-bc16-c20d958b6f59"
	oldB504Mage     = "a5d02339-9f6d-486e-87d8-5ec2274cf680"
	oldB504Garg     = "acc74d59-f587-46a2-a90b-865f721d494d"
	oldB504Beetle   = "c2dbdf9d-3ac1-47d8-9ad8-57e7888e5cdd"
	oldB504Anathem  = "137b5708-50a9-4f4c-84e6-9043be9f1989"
	oldB504Grixis   = "21f507fd-6a26-455c-879a-f838e0aa2a43"
)

func TestOldB504EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB504Defiler: "Defiler of Souls", oldB504Hellion: "Blitz Hellion", oldB504Demon: "Illusory Demon",
		oldB504Shield: "Shield of the Righteous", oldB504Whip: "Demonspine Whip", oldB504Unscythe: "Unscythe, Killer of Kings",
		oldB504Mage: "Meddling Mage", oldB504Garg: "Nulltread Gargantuan", oldB504Beetle: "Giant Ambush Beetle",
		oldB504Anathem: "Anathemancer", oldB504Grixis: "Grixis Sojourners",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB504Colored(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: colors, Owner: owner, Controller: owner,
	})
}

// --- Defiler of Souls ---------------------------------------------------------

func TestOldB504DefilerMakesTheUpkeepPlayerSacrificeAMonocoloredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Defiler of Souls", "Creature — Demon", oldB504Defiler, false)
	mono := oldB504Colored(g, opp.ID, "Mono", "R")
	gold := oldB504Colored(g, opp.ID, "Gold", "R", "G")
	colorless := oldB504Colored(g, opp.ID, "Colorless")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChooseCardsFor(g, opp.ID); c != nil {
		if len(c.ChooseCards) != 1 || c.ChooseCards[0] != mono {
			t.Errorf("offered %v, want only the monocolored creature", c.ChooseCards)
		}
		effectstest.AnswerChooseCards(t, g, opp.ID, mono)
	}
	if g.Battlefield.Contains(mono) {
		t.Error("the monocolored creature was not sacrificed")
	}
	if !g.Battlefield.Contains(gold) || !g.Battlefield.Contains(colorless) {
		t.Error("a multicolored or colorless creature was sacrificed")
	}
}

func TestOldB504DefilerAlsoHitsItsControllersOwnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Defiler of Souls", "Creature — Demon", oldB504Defiler, false)
	mine := oldB504Colored(g, me.ID, "Mine", "U")
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, mine)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the controller's own monocolored creature was not sacrificed at their upkeep")
	}
}

// --- Blitz Hellion ------------------------------------------------------------

func TestOldB504HellionShufflesItselfIntoItsOwnersLibraryAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hellion := effectstest.PushCatalogPermanent(g, me.ID, "Blitz Hellion", "Creature — Hellion", oldB504Hellion, false)
	lib := me.Library.Size()
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hellion) {
		t.Fatal("the Hellion is still on the battlefield")
	}
	if effectstest.ZoneOfCard(g, hellion) != game.ZoneLibrary {
		t.Errorf("the Hellion is in %q, want the library", effectstest.ZoneOfCard(g, hellion))
	}
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib+1)
	}
}

func TestOldB504HellionGoesBackToItsOwnerAtAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hellion := effectstest.PushCatalogPermanent(g, opp.ID, "Blitz Hellion", "Creature — Hellion", oldB504Hellion, false)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, hellion) != game.ZoneLibrary {
		t.Errorf("the Hellion is in %q after an end step, want the library", effectstest.ZoneOfCard(g, hellion))
	}
}

// --- Illusory Demon -----------------------------------------------------------

func TestOldB504DemonSacrificesItselfWhenYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.PushCatalogPermanent(g, me.ID, "Illusory Demon", "Creature — Demon Illusion", oldB504Demon, false)
	effectstest.CastCatalogSpell(t, g, "Plain Spell", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(demon) {
		t.Error("the Demon survived its controller casting a spell")
	}
}

func TestOldB504DemonIgnoresAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.PushCatalogPermanent(g, me.ID, "Illusory Demon", "Creature — Demon Illusion", oldB504Demon, false)
	effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: "Their Spell", TypeLine: "Instant"})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon was sacrificed to an opponent's spell")
	}
}

// --- Shield of the Righteous --------------------------------------------------

func TestOldB504ShieldGivesPlusTwoToughnessVigilanceAndKeepsABlockedAttackerTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 5)
	blocker := effectstest.HmlCreature(g, opp.ID, "Blocker", "", 1, 4)
	shield := effectstest.PushCatalogPermanent(g, opp.ID, "Shield of the Righteous", "Artifact — Equipment", oldB504Shield, false)
	effectstest.AttachForTest(t, g, shield, blocker)
	if p, tough := effectstest.EffectivePTOf(t, g, blocker); p != 1 || tough != 6 {
		t.Errorf("equipped creature is %d/%d, want 1/6", p, tough)
	}
	if !effectstest.P1g2HasKeyword(g, blocker, "vigilance") {
		t.Error("the equipped creature has no vigilance")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if skips := effectstest.HmlUntapSkips(t, g, attacker); len(skips) != 1 {
		t.Errorf("the blocked attacker carries %d untap marks, want 1", len(skips))
	}
	if skips := effectstest.HmlUntapSkips(t, g, blocker); len(skips) != 0 {
		t.Errorf("the blocker marked itself: %+v", skips)
	}
}

// --- Demonspine Whip ----------------------------------------------------------

func TestOldB504WhipGivesTheEquippedCreaturePlusXPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Demonspine Whip", "Artifact — Equipment", oldB504Whip, false)
	effectstest.AttachForTest(t, g, whip, bear)
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{XValue: 3}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("equipped creature is %d/%d, want 5/2", p, tough)
	}
}

func TestOldB504WhipDoesNothingWhenUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Demonspine Whip", "Artifact — Equipment", oldB504Whip, false)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("an unattached Whip changed a creature's power to %d", p)
	}
}

// --- Unscythe, Killer of Kings ------------------------------------------------

func oldB504UnscytheBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bearer := effectstest.B12Creature(g, me.ID, "Bearer", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	unscythe := effectstest.PushCatalogPermanent(g, me.ID, "Unscythe, Killer of Kings", "Legendary Artifact — Equipment", oldB504Unscythe, false)
	effectstest.AttachForTest(t, g, unscythe, bearer)
	return g, me, bearer, victim
}

func TestOldB504UnscytheExilesACreatureTheBearerDamagedAndMakesAZombie(t *testing.T) {
	g, me, bearer, victim := oldB504UnscytheBoard(t)
	if p, tough := effectstest.EffectivePTOf(t, g, bearer); p != 5 || tough != 5 {
		t.Errorf("equipped creature is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.P1g2HasKeyword(g, bearer, "first strike") {
		t.Error("the equipped creature has no first strike")
	}
	effectstest.P2Ping(t, g, bearer, victim, 2)
	g.RunStateChecksForTest()
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if !effectstest.InExile(g, victim) {
		t.Errorf("the dead creature is in %q, want exile", effectstest.ZoneOfCard(g, victim))
	}
	if n := effectstest.TokensNamedControlledBy(g, "Zombie", me.ID); n != 1 {
		t.Errorf("%d Zombie tokens, want 1", n)
	}
}

func TestOldB504UnscytheMayDeclineAndIgnoresCreaturesTheBearerDidNotDamage(t *testing.T) {
	g, me, bearer, victim := oldB504UnscytheBoard(t)
	effectstest.P2Ping(t, g, bearer, victim, 2)
	g.RunStateChecksForTest()
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if effectstest.InExile(g, victim) || effectstest.TokensNamedControlledBy(g, "Zombie", me.ID) != 0 {
		t.Error("declined, yet the card was exiled or a Zombie was made")
	}

	g2, me2, _, _ := oldB504UnscytheBoard(t)
	other := effectstest.B12Creature(g2, g2.Seats[1].ID, "Undamaged", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g2, g2.Seats[1].ID, other)
	effectstest.B06AnswerAllTriggerPrompts(t, g2, me2.ID, true)
	if effectstest.InExile(g2, other) || effectstest.TokensNamedControlledBy(g2, "Zombie", me2.ID) != 0 {
		t.Error("a creature the bearer never damaged was exiled")
	}
}

// --- Meddling Mage ------------------------------------------------------------

func TestOldB504MageBarsTheNamedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Meddling Mage", "Creature — Human Wizard", oldB504Mage, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatalf("no card-name prompt as it entered: %+v", g.PendingChoices)
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	const bolt = "4457ed35-7c10-48c8-9776-456485fdf070"
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", bolt, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("the named spell was cast")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Errorf("a differently named spell was refused: %v", err)
	}
}

// --- Nulltread Gargantuan -----------------------------------------------------

func oldB504PickOwn(t *testing.T, g *game.Game, me, id uuid.UUID) {
	t.Helper()
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me)
	if p == nil {
		t.Fatal("no own-permanents pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me, []uuid.UUID{id}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
}

func TestOldB504GargantuanPutsAChosenCreatureOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Nulltread Gargantuan", "Creature — Beast", oldB504Garg, 5, 6)
	effectstest.PassPriorityAroundTable(t, g)
	oldB504PickOwn(t, g, me.ID, bear)
	if effectstest.ZoneOfCard(g, bear) != game.ZoneLibrary {
		t.Errorf("the chosen creature is in %q, want the library", effectstest.ZoneOfCard(g, bear))
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != bear {
		t.Error("the chosen creature is not on top of the library")
	}
}

func TestOldB504GargantuanMayChooseItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	garg := effectstest.CastHoldCreature(t, g, "Nulltread Gargantuan", "Creature — Beast", oldB504Garg, 5, 6)
	effectstest.PassPriorityAroundTable(t, g)
	oldB504PickOwn(t, g, me.ID, garg)
	if effectstest.ZoneOfCard(g, garg) != game.ZoneLibrary {
		t.Errorf("the Gargantuan is in %q, want the library", effectstest.ZoneOfCard(g, garg))
	}
}

// --- Giant Ambush Beetle ------------------------------------------------------

func TestOldB504BeetleForcesTheTargetToBlockIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	effectstest.AdvanceToMain(t, g)
	forced := effectstest.ReqCreature(g, opp.ID, "Forced", "", 2, 2)
	other := effectstest.ReqCreature(g, opp.ID, "Other", "", 2, 2)
	beetle := effectstest.CastFromHandForTest(t, g, me, "Giant Ambush Beetle", "Creature — Insect", "{3}{R}{R}", oldB504Beetle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: forced}, true)
	effectstest.AttackAndTrigger(t, g, beetle)
	effectstest.OnToBlockers(t, g)
	if owed := effectstest.MustBlockNow(g); owed[forced] != beetle {
		t.Fatalf("must_block = %v, want the forced creature owing a block of the Beetle", owed)
	}
	if err := effectstest.ReqBlock(t, g, beetle, other); err != nil {
		t.Fatalf("a legal block was refused outright: %v", err)
	}
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, beetle, forced); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("the forced block was still refused: %v", err)
	}
}

func TestOldB504BeetleAskedNothingWhenTheMayIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	effectstest.AdvanceToMain(t, g)
	forced := effectstest.ReqCreature(g, opp.ID, "Forced", "", 2, 2)
	beetle := effectstest.CastFromHandForTest(t, g, me, "Giant Ambush Beetle", "Creature — Insect", "{3}{R}{R}", oldB504Beetle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: forced}, false)
	effectstest.AttackAndTrigger(t, g, beetle)
	effectstest.OnToBlockers(t, g)
	if owed := effectstest.MustBlockNow(g); owed != nil {
		t.Errorf("must_block = %v, want nothing owed after a declined may", owed)
	}
}

// --- Anathemancer -------------------------------------------------------------

func TestOldB504AnathemancerDealsDamageEqualToTheTargetsNonbasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Nonbasic A", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Nonbasic B", TypeLine: "Land — Desert", Owner: opp.ID, Controller: opp.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Snow Island", TypeLine: "Basic Snow Land — Island", Owner: opp.ID, Controller: opp.ID})
	life := opp.Life
	effectstest.CastFromHandForTest(t, g, me, "Anathemancer", "Creature — Zombie Wizard", "{1}{B}{R}", oldB504Anathem, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d (two nonbasic lands)", opp.Life, life-2)
	}
}

func TestOldB504AnathemancerHasUnearth(t *testing.T) {
	var found *game.ActivatedAbilityShape
	for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: oldB504Anathem}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			ab := ab
			found = &ab
		}
	}
	if found == nil {
		t.Fatal("no graveyard ability")
	}
	if found.Cost.Mana != "{5}{B}{R}" || !found.SorcerySpeed {
		t.Errorf("unearth cost %q sorcery %v, want {5}{B}{R} at sorcery speed", found.Cost.Mana, found.SorcerySpeed)
	}
}

// --- Grixis Sojourners --------------------------------------------------------

func TestOldB504GrixisCycledMayExileACardFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushGraveyardCardTyped(opp, "Old Card", "Sorcery")
	effectstest.CycleFromHand(t, g, "Grixis Sojourners", "Creature — Zombie Ogre", oldB504Grixis, "{C}{C}{B}")
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, true)
	if !effectstest.InExile(g, target) {
		t.Errorf("the card is in %q, want exile", effectstest.ZoneOfCard(g, target))
	}
}

func TestOldB504GrixisDiesMayExileACardOnceAndMayDecline(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		target := effectstest.PushGraveyardCardTyped(opp, "Old Card", "Sorcery")
		grixis := effectstest.PushCatalogPermanent(g, me.ID, "Grixis Sojourners", "Creature — Zombie Ogre", oldB504Grixis, false)
		effectstest.KillCreature(t, g, me.ID, grixis)
		effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, yes)
		if effectstest.InExile(g, target) != yes {
			t.Errorf("yes=%v: card in exile %v", yes, effectstest.InExile(g, target))
		}
	}
}
