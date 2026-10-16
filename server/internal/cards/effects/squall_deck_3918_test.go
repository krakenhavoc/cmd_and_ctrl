package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// squall_deck_3918_test.go — the Squall, SeeD Mercenary deck request
// (#3918): Ardenn, Celestial Armor, Fated Clash, Jirina, Michiko Konda,
// Rinoa, Search for Dagger, Sigarda's Aid, Sword of Light and Shadow,
// Touch the Spirit Realm, Zenos yae Galvus // Shinryu.

const (
	sqArdennOracle    = "7c521612-21c6-4c4b-b10e-66e70928609b"
	sqCelestialOracle = "4f1b68d7-7d16-4bb1-af5f-f1708b23ca73"
	sqFatedClashOracl = "610d9f58-4182-49cb-855c-9b6030e7f714"
	sqJirinaOracle    = "fda76032-7e6f-416f-8874-2a0642837522"
	sqMichikoOracle   = "fbe190eb-fd4b-40e7-9a17-48c96b60f653"
	sqRinoaOracle     = "47b29cf9-8ef3-4b6a-a9a3-3ab822c5dea4"
	sqDaggerOracle    = "9bdd6ece-b78c-4814-ac40-b48f91b405b5"
	sqSigardasOracle  = "4f5f2792-0cc4-4caa-b798-2ee9597f2524"
	sqSwordLSOracle   = "01c8a554-45a6-46ca-b40a-af0fd4529e40"
	sqTouchOracle     = "52f0dcbd-ccf3-4525-8e55-2d38b4991834"
)

func TestSquallDeckCardsAreRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		sqArdennOracle, sqCelestialOracle, sqFatedClashOracl, sqJirinaOracle, sqMichikoOracle,
		sqRinoaOracle, sqDaggerOracle, sqSigardasOracle, sqSwordLSOracle, sqTouchOracle,
		zenosYaeGalvusOracleID, zenosYaeGalvusOracleID + "#1",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// sqActive is the active player and the next seat.
func sqActive(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func sqAnswerCards(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	p := latestChooseCards(g, chooser)
	if p == nil {
		t.Fatalf("no choose_cards prompt for %s; pending %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(p.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	return p
}

func sqCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return dcCard(t, g, id)
}

// --- Jirina, Dauntless General ---------------------------------------

func TestJirinaExilesTargetPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	a := pushGraveyardCardTyped(opp, "Grave A", "Creature — Bear")
	b := pushGraveyardCardTyped(opp, "Grave B", "Sorcery")
	mine := pushGraveyardCardTyped(me, "Mine", "Instant")
	castCatalogSpell(t, g, "Jirina, Dauntless General", "Legendary Creature — Human Soldier", sqJirinaOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if inZone(opp.Graveyard, a) || inZone(opp.Graveyard, b) {
		t.Error("the target player's graveyard was not exiled")
	}
	if !inZone(me.Graveyard, mine) {
		t.Error("the other player's graveyard was touched")
	}
}

func TestJirinaSacrificeProtectsHumansOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	jirina := pushCatalogPermanent(g, me.ID, "Jirina, Dauntless General", "Legendary Creature — Human Soldier", sqJirinaOracle, false)
	human := dcSeed(g, me.ID, "Villager", "Creature — Human Peasant", "", 1, 1)
	bear := dcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirHuman := dcSeed(g, opp.ID, "Their Human", "Creature — Human", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, jirina, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(jirina) {
		t.Error("Jirina was not sacrificed")
	}
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectiveAbilitiesContain(t, g, human, kw) {
			t.Errorf("your Human lacks %s", kw)
		}
		if effectiveAbilitiesContain(t, g, bear, kw) || effectiveAbilitiesContain(t, g, theirHuman, kw) {
			t.Errorf("a non-Human or an opponent's Human gained %s", kw)
		}
	}
}

// --- Michiko Konda, Truth Seeker -------------------------------------

func TestMichikoMakesTheDamagingPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Michiko Konda, Truth Seeker", "Legendary Creature — Human Advisor", sqMichikoOracle, false)
	shooter := dcSeed(g, opp.ID, "Shooter", "Creature — Goblin", "", 1, 1)
	keep := dcSeed(g, opp.ID, "Rock", "Artifact", "", 0, 0)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(shooter, me.ID, 1)
	})
	passPriorityAroundTable(t, g)
	p := latestChooseCards(g, opp.ID)
	if p == nil {
		t.Fatal("the damaging player was not asked to sacrifice")
	}
	if !hasID(p.ChooseCards, shooter) || !hasID(p.ChooseCards, keep) {
		t.Errorf("candidates %v, want every permanent the opponent controls", p.ChooseCards)
	}
	sqAnswerCards(t, g, opp.ID, keep)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keep) || !g.Battlefield.Contains(shooter) {
		t.Error("the chosen permanent was not the one sacrificed")
	}
}

func TestMichikoIgnoresYourOwnSourcesAndDamageToCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Michiko Konda, Truth Seeker", "Legendary Creature — Human Advisor", sqMichikoOracle, false)
	mine := dcSeed(g, me.ID, "Mine", "Creature — Goblin", "", 1, 1)
	theirs := dcSeed(g, opp.ID, "Theirs", "Creature — Goblin", "", 1, 1)
	victim := dcSeed(g, me.ID, "Victim", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(mine, me.ID, 1)
		_ = g.DealDamageToCreatureForEffect(theirs, victim, 1)
	})
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, opp.ID) != nil || latestChooseCards(g, me.ID) != nil {
		t.Error("Michiko triggered on your own source or on damage to a creature")
	}
}

// --- Search for Dagger -----------------------------------------------

func TestSearchForDaggerTakesALegendaryCreatureWhenYourCommanderEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Search for Dagger", "Enchantment", sqDaggerOracle, false)
	legend := pushLibraryCardForTest(me, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human"})
	plain := pushLibraryCardForTest(me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear"})
	// Two more on top so the legend is within the top six.
	for i := 0; i < 2; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	// pushLibraryCardForTest bottoms; move the two named cards to the top.
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{legend, plain} {
			c, _ := me.Library.Remove(id)
			me.Library.PushTop(c)
		}
	})
	cmdr := game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Knight",
		Owner: me.ID, Controller: me.ID, IsCommander: true, Power: 3, Toughness: 3}
	me.Hand.PushTop(cmdr)
	if err := g.CastSpell(me.ID, cmdr.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast commander: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no look prompt when the commander entered")
	}
	if !hasID(p.ChooseCards, legend) || hasID(p.ChooseCards, plain) {
		t.Errorf("candidates %v, want only the legendary creature", p.ChooseCards)
	}
	sqAnswerCards(t, g, me.ID, legend)
	passPriorityAroundTable(t, g)
	if !inZone(me.Hand, legend) {
		t.Error("the legendary creature did not go to hand")
	}
	if top, err := me.Library.Top(); err == nil && top.InstanceID == plain {
		t.Error("the rest stayed on top")
	}
}

func TestSearchForDaggerIgnoresANonCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Search for Dagger", "Enchantment", sqDaggerOracle, false)
	castCatalogSpell(t, g, "Just a Legend", "Legendary Creature — Human", "", nil)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("Search for Dagger triggered on a creature that is not your commander")
	}
}

// --- Sword of Light and Shadow ---------------------------------------

func TestSwordOfLightAndShadowGainsLifeAndReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	sword := seedPermanentWithOracle(g, me.ID, "Sword of Light and Shadow", "Artifact — Equipment", sqSwordLSOracle)
	dead := pushGraveyardCreature(g, me.ID, "Dead Friend", "{1}{W}")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear})
	})
	if !effectiveAbilitiesContain(t, g, bear, "protection from white") || !effectiveAbilitiesContain(t, g, bear, "protection from black") {
		t.Error("the equipped creature lacks protection from white and from black")
	}
	life := me.Life
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt after combat damage")
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	krileConfirm(t, g, me, true)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if !inZone(me.Hand, dead) {
		t.Error("the creature card was not returned to hand")
	}
	if opp.Life != 40-4 {
		t.Errorf("opponent life %d, want 36 (+2/+2 on a 2/2)", opp.Life)
	}
}

// --- Celestial Armor -------------------------------------------------

func TestCelestialArmorAttachesAndProtectsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	armor := castCatalogSpell(t, g, "Celestial Armor", "Artifact — Equipment", sqCelestialOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	a := sqCard(t, g, armor)
	if a.AttachedTo.ID != bear {
		t.Fatalf("the Armor is attached to %v, want the bear", a.AttachedTo)
	}
	for _, kw := range []string{"hexproof", "indestructible", "flying"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("the creature lacks %s", kw)
		}
	}
	if c := sqCard(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("bear is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Sigarda's Aid ---------------------------------------------------

func TestSigardasAidFlashesEquipmentAndAurasOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Sigarda's Aid", "Enchantment", sqSigardasOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	equip := pushCatalogHandCard(me, "Bonesplitter", "Artifact — Equipment", "")
	aura := pushCatalogHandCard(me, "Some Aura", "Enchantment — Aura", "")
	rock := pushCatalogHandCard(me, "Mind Stone", "Artifact", "")
	if err := g.CastSpell(me.ID, equip, game.CastSpellParams{}); err != nil {
		t.Errorf("an Equipment was refused at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{}); err != nil {
		t.Errorf("an Aura was refused at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err == nil {
		t.Error("a plain artifact was cast at instant speed")
	}
}

func TestSigardasAidAttachesAnEnteringEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Sigarda's Aid", "Enchantment", sqSigardasOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	equip := castCatalogSpell(t, g, "Bonesplitter", "Artifact — Equipment", "", nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if ask := latestTriggerPromptFor(g, me.ID); ask != nil {
		if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c := sqCard(t, g, equip); c.AttachedTo.ID != bear {
		t.Errorf("the Equipment is attached to %v, want the bear", c.AttachedTo)
	}
}

// --- Touch the Spirit Realm ------------------------------------------

func TestTouchTheSpiritRealmExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	theirs := dcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	touch := castCatalogSpell(t, g, "Touch the Spirit Realm", "Enchantment", sqTouchOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !inZone(g.Exile, theirs) {
		t.Fatal("the creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(touch) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("the exiled creature did not return when the enchantment left")
	}
}

func TestTouchTheSpiritRealmChannelReturnsAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	theirs := dcSeed(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	touch := pushCatalogHandCard(me, "Touch the Spirit Realm", "Enchantment", sqTouchOracle)
	dcMana(t, g, me.ID, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, touch, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !inZone(me.Graveyard, touch) {
		t.Error("channel did not discard the card")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the artifact was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Rock" && c.Controller == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("the artifact did not return at the end step under its owner's control")
	}
}

// --- Fated Clash -----------------------------------------------------

func TestFatedClashSavesBothTargetsAndDestroysTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	mine := dcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := dcSeed(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	other := dcSeed(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	otherTheirs := dcSeed(g, opp.ID, "Other Theirs", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Fated Clash", "Sorcery", sqFatedClashOracl, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a target did not survive")
	}
	if g.Battlefield.Contains(other) || g.Battlefield.Contains(otherTheirs) {
		t.Error("a creature that was not a target survived")
	}
}

func TestFatedClashFlashOnlyWithAnAttackerAndABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 0, 4)
	clash := pushCatalogHandCard(me, "Fated Clash", "Sorcery", sqFatedClashOracl)
	fund := func() {
		g.WithWriteLock(func() {
			me.ManaPool = nil
			for _, c := range []string{"W", "W", "C", "C", "C"} {
				me.ManaPool.AddMana(game.ManaToken{Color: c})
			}
		})
	}
	flash := game.CastSpellParams{Strict: true, AlternativeCost: "flash-if-attacking-and-blocking",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}, {Kind: game.TargetCard, ID: blocker}}}
	if fatedClashCreaturesInCombat(g, me.ID) {
		t.Fatal("the condition holds outside combat")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if fatedClashCreaturesInCombat(g, me.ID) {
		t.Error("the condition holds with an attacker and no blocker")
	}
	fund()
	if err := g.CastSpell(me.ID, clash, flash); err == nil {
		t.Fatal("flash-cast with an attacker and no blocker")
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	finishBlockDeclarations(t, g)
	if !fatedClashCreaturesInCombat(g, me.ID) {
		t.Error("the condition does not hold with an attacker and a blocker")
	}
	fund()
	if err := g.CastSpell(me.ID, clash, flash); err != nil {
		t.Fatalf("flash-cast with an attacker and a blocker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) || !g.Battlefield.Contains(blocker) {
		t.Error("a target did not survive the flashed Clash")
	}
}

// --- Rinoa, Angel Wing -----------------------------------------------

func TestRinoaPumpsFliersAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Rinoa, Angel Wing", "Legendary Creature — Human Rebel Warlock", sqRinoaOracle, false)
	flier := pushVanillaCreature(g, me.ID, "Bird", 1, 1)
	ground := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if c := sqCard(t, g, flier); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("flier is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectiveAbilitiesContain(t, g, flier, "vigilance") {
		t.Error("the flier lacks vigilance")
	}
	if c := sqCard(t, g, ground); c.CurrentPower() != 2 || effectiveAbilitiesContain(t, g, ground, "vigilance") {
		t.Error("a creature without flying was pumped")
	}
}

func TestRinoaReturnsADeadAttackerOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Rinoa, Angel Wing", "Legendary Creature — Human Rebel Warlock", sqRinoaOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	passPriorityAroundTable(t, g)
	p := latestChooseCards(g, me.ID)
	if p == nil || !hasID(p.ChooseCards, a) {
		t.Fatalf("no prompt offering the dead attacker: %+v", p)
	}
	sqAnswerCards(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Name == "Bear A" {
			back = &c
		}
	}
	if back == nil {
		t.Fatal("the attacker did not return")
	}
	if !back.Tapped || back.Counters[game.CounterFlying] != 1 {
		t.Errorf("returned tapped=%v flying counters=%d, want tapped with one", back.Tapped, back.Counters[game.CounterFlying])
	}

	// The second death this turn finds the use spent.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(b) })
	passPriorityAroundTable(t, g)
	if p2 := latestChooseCards(g, me.ID); p2 != nil && p2 != p && hasID(p2.ChooseCards, b) {
		t.Error("Rinoa offered a second return in the same turn")
	}
	if g.Battlefield.Contains(b) {
		t.Error("the second attacker came back")
	}
}

// --- Zenos yae Galvus // Shinryu -------------------------------------

func zenosCard(owner uuid.UUID) game.Card {
	c := praetorCard(owner, zenosYaeGalvusOracleID, "Zenos yae Galvus", "Legendary Creature — Human Noble Warrior", 4, 4,
		"Shinryu, Transcendent Rival", "Legendary Creature — Dragon", false)
	c.Faces[0].ManaCost = "{3}{B}{B}"
	c.Faces[1].Power, c.Faces[1].Toughness = 8, 8
	c.SetFace(0)
	return c
}

func TestZenosShrinksTheOthersAndTransformsWhenTheChosenLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	chosen := dcSeed(g, opp.ID, "Rival", "Creature — Human", "", 3, 3)
	other := dcSeed(g, opp.ID, "Other", "Creature — Bear", "", 3, 3)
	mine := dcSeed(g, me.ID, "Mine", "Creature — Bear", "", 3, 3)
	z := zenosCard(me.ID)
	me.Hand.PushTop(z)
	if err := g.CastSpell(me.ID, z.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Zenos: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no choice of a creature")
	}
	if hasID(p.ChooseCards, mine) || !hasID(p.ChooseCards, chosen) {
		t.Errorf("candidates %v, want only opponents' creatures", p.ChooseCards)
	}
	sqAnswerCards(t, g, me.ID, chosen)
	passPriorityAroundTable(t, g)
	if c := sqCard(t, g, chosen); c.CurrentPower() != 3 {
		t.Errorf("chosen creature is %d power, want 3", c.CurrentPower())
	}
	for _, id := range []uuid.UUID{other, mine} {
		if c := sqCard(t, g, id); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
			t.Errorf("%s is %d/%d, want 1/1", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := sqCard(t, g, z.InstanceID); c.CurrentPower() != 4 {
		t.Errorf("Zenos is %d power, want 4", c.CurrentPower())
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(chosen) })
	passPriorityAroundTable(t, g)
	// Shinryu asks for an opponent as it transforms.
	answerChosenPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	zc := sqCard(t, g, z.InstanceID)
	if zc.Name != "Shinryu, Transcendent Rival" {
		t.Fatalf("Zenos did not transform: %q", zc.Name)
	}
	if zc.ChosenPlayer != opp.ID {
		t.Errorf("Shinryu chose %v, want the opponent", zc.ChosenPlayer)
	}

	// The chosen player loses the game: you win.
	g.WithWriteLock(func() { _, _ = g.LoseTheGameForEffect(opp.ID, uuid.Nil) })
	passPriorityAroundTable(t, g)
	if g.State == game.StateActive {
		t.Error("the game is still going after the chosen player lost")
	}
}

func TestZenosDoesNotTransformWhenAnotherCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	chosen := dcSeed(g, opp.ID, "Rival", "Creature — Human", "", 5, 5)
	other := dcSeed(g, opp.ID, "Other", "Creature — Bear", "", 5, 5)
	z := zenosCard(me.ID)
	me.Hand.PushTop(z)
	if err := g.CastSpell(me.ID, z.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Zenos: %v", err)
	}
	passPriorityAroundTable(t, g)
	sqAnswerCards(t, g, me.ID, chosen)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if c := sqCard(t, g, z.InstanceID); c.Name != "Zenos yae Galvus" {
		t.Errorf("Zenos transformed when an unchosen creature left: %q", c.Name)
	}
}

// --- Ardenn, Intrepid Archaeologist ----------------------------------

func TestArdennMovesChosenEquipmentToTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ardenn, Intrepid Archaeologist", "Legendary Creature — Kor Scout", sqArdennOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	e1 := dcSeed(g, me.ID, "Sword One", "Artifact — Equipment", "", 0, 0)
	e2 := dcSeed(g, me.ID, "Sword Two", "Artifact — Equipment", "", 0, 0)
	rock := dcSeed(g, me.ID, "Rock", "Artifact", "", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	p := latestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no choice of attachments")
	}
	if !hasID(p.ChooseCards, e1) || !hasID(p.ChooseCards, e2) || hasID(p.ChooseCards, rock) {
		t.Errorf("candidates %v, want the two Equipment only", p.ChooseCards)
	}
	if p.ChooseMin != 0 {
		t.Errorf("floor %d, want 0 (you may)", p.ChooseMin)
	}
	sqAnswerCards(t, g, me.ID, e1)
	passPriorityAroundTable(t, g)
	if c := sqCard(t, g, e1); c.AttachedTo.ID != bear {
		t.Error("the chosen Equipment was not attached")
	}
	if c := sqCard(t, g, e2); c.AttachedTo.ID == bear {
		t.Error("an unchosen Equipment was attached")
	}
}

func TestArdennOffersNoEquipmentForANoncreatureOrAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := sqActive(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Ardenn, Intrepid Archaeologist", "Legendary Creature — Kor Scout", sqArdennOracle, false)
	dcSeed(g, me.ID, "Sword One", "Artifact — Equipment", "", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Error("an Equipment was offered for a player")
	}
}
