package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	cardsx "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/x"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSquallDeckCardsAreRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		effectstest.SqArdennOracle, effectstest.SqCelestialOracle, effectstest.SqJirinaOracle, effectstest.SqMichikoOracle,
		effectstest.SqRinoaOracle, effectstest.SqDaggerOracle, effectstest.SqSigardasOracle, effectstest.SqSwordLSOracle, effectstest.SqTouchOracle,
		effectstest.SqFatedClashOracl, // #5207: its flash is a timing permission now
		cardsx.ZenosYaeGalvusOracleID, cardsx.ZenosYaeGalvusOracleID + "#1",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// --- Jirina, Dauntless General ---------------------------------------

func TestJirinaExilesTargetPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	a := effectstest.PushGraveyardCardTyped(opp, "Grave A", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(opp, "Grave B", "Sorcery")
	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Instant")
	effectstest.CastCatalogSpell(t, g, "Jirina, Dauntless General", "Legendary Creature — Human Soldier", effectstest.SqJirinaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InZone(opp.Graveyard, a) || effectstest.InZone(opp.Graveyard, b) {
		t.Error("the target player's graveyard was not exiled")
	}
	if !effectstest.InZone(me.Graveyard, mine) {
		t.Error("the other player's graveyard was touched")
	}
}

func TestJirinaSacrificeProtectsHumansOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	jirina := effectstest.PushCatalogPermanent(g, me.ID, "Jirina, Dauntless General", "Legendary Creature — Human Soldier", effectstest.SqJirinaOracle, false)
	human := effectstest.DcSeed(g, me.ID, "Villager", "Creature — Human Peasant", "", 1, 1)
	bear := effectstest.DcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirHuman := effectstest.DcSeed(g, opp.ID, "Their Human", "Creature — Human", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, jirina, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(jirina) {
		t.Error("Jirina was not sacrificed")
	}
	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, human, kw) {
			t.Errorf("your Human lacks %s", kw)
		}
		if effectstest.EffectiveAbilitiesContain(t, g, bear, kw) || effectstest.EffectiveAbilitiesContain(t, g, theirHuman, kw) {
			t.Errorf("a non-Human or an opponent's Human gained %s", kw)
		}
	}
}

// --- Michiko Konda, Truth Seeker -------------------------------------

func TestMichikoMakesTheDamagingPlayerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Michiko Konda, Truth Seeker", "Legendary Creature — Human Advisor", effectstest.SqMichikoOracle, false)
	shooter := effectstest.DcSeed(g, opp.ID, "Shooter", "Creature — Goblin", "", 1, 1)
	keep := effectstest.DcSeed(g, opp.ID, "Rock", "Artifact", "", 0, 0)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(shooter, me.ID, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, opp.ID)
	if p == nil {
		t.Fatal("the damaging player was not asked to sacrifice")
	}
	if !effectstest.HasID(p.ChooseCards, shooter) || !effectstest.HasID(p.ChooseCards, keep) {
		t.Errorf("candidates %v, want every permanent the opponent controls", p.ChooseCards)
	}
	effectstest.SqAnswerCards(t, g, opp.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(keep) || !g.Battlefield.Contains(shooter) {
		t.Error("the chosen permanent was not the one sacrificed")
	}
}

func TestMichikoIgnoresYourOwnSourcesAndDamageToCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Michiko Konda, Truth Seeker", "Legendary Creature — Human Advisor", effectstest.SqMichikoOracle, false)
	mine := effectstest.DcSeed(g, me.ID, "Mine", "Creature — Goblin", "", 1, 1)
	theirs := effectstest.DcSeed(g, opp.ID, "Theirs", "Creature — Goblin", "", 1, 1)
	victim := effectstest.DcSeed(g, me.ID, "Victim", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(mine, me.ID, 1)
		_ = g.DealDamageToCreatureForEffect(theirs, victim, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, opp.ID) != nil || effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("Michiko triggered on your own source or on damage to a creature")
	}
}

// --- Search for Dagger -----------------------------------------------

func TestSearchForDaggerTakesALegendaryCreatureWhenYourCommanderEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Search for Dagger", "Enchantment", effectstest.SqDaggerOracle, false)
	legend := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human"})
	plain := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear"})
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
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no look prompt when the commander entered")
	}
	if !effectstest.HasID(p.ChooseCards, legend) || effectstest.HasID(p.ChooseCards, plain) {
		t.Errorf("candidates %v, want only the legendary creature", p.ChooseCards)
	}
	effectstest.SqAnswerCards(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InZone(me.Hand, legend) {
		t.Error("the legendary creature did not go to hand")
	}
	if top, err := me.Library.Top(); err == nil && top.InstanceID == plain {
		t.Error("the rest stayed on top")
	}
}

func TestSearchForDaggerIgnoresANonCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Search for Dagger", "Enchantment", effectstest.SqDaggerOracle, false)
	effectstest.CastCatalogSpell(t, g, "Just a Legend", "Legendary Creature — Human", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("Search for Dagger triggered on a creature that is not your commander")
	}
}

// --- Sword of Light and Shadow ---------------------------------------

func TestSwordOfLightAndShadowGainsLifeAndReturnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	sword := effectstest.SeedPermanentWithOracle(g, me.ID, "Sword of Light and Shadow", "Artifact — Equipment", effectstest.SqSwordLSOracle)
	dead := effectstest.PushGraveyardCreature(g, me.ID, "Dead Friend", "{1}{W}")
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear})
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "protection from white") || !effectstest.EffectiveAbilitiesContain(t, g, bear, "protection from black") {
		t.Error("the equipped creature lacks protection from white and from black")
	}
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("advance: %v", err)
		}
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt after combat damage")
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KrileConfirm(t, g, me, true)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
	if !effectstest.InZone(me.Hand, dead) {
		t.Error("the creature card was not returned to hand")
	}
	if opp.Life != 40-4 {
		t.Errorf("opponent life %d, want 36 (+2/+2 on a 2/2)", opp.Life)
	}
}

// --- Celestial Armor -------------------------------------------------

func TestCelestialArmorAttachesAndProtectsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	armor := effectstest.CastCatalogSpell(t, g, "Celestial Armor", "Artifact — Equipment", effectstest.SqCelestialOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	a := effectstest.SqCard(t, g, armor)
	if a.AttachedTo.ID != bear {
		t.Fatalf("the Armor is attached to %v, want the bear", a.AttachedTo)
	}
	for _, kw := range []string{"hexproof", "indestructible", "flying"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, bear, kw) {
			t.Errorf("the creature lacks %s", kw)
		}
	}
	if c := effectstest.SqCard(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("bear is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Sigarda's Aid ---------------------------------------------------

func TestSigardasAidFlashesEquipmentAndAurasOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sigarda's Aid", "Enchantment", effectstest.SqSigardasOracle, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	equip := effectstest.PushCatalogHandCard(me, "Bonesplitter", "Artifact — Equipment", "")
	aura := effectstest.PushCatalogHandCard(me, "Some Aura", "Enchantment — Aura", "")
	rock := effectstest.PushCatalogHandCard(me, "Mind Stone", "Artifact", "")
	if err := g.CastSpell(me.ID, equip, game.CastSpellParams{}); err != nil {
		t.Errorf("an Equipment was refused at instant speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{}); err != nil {
		t.Errorf("an Aura was refused at instant speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err == nil {
		t.Error("a plain artifact was cast at instant speed")
	}
}

func TestSigardasAidAttachesAnEnteringEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sigarda's Aid", "Enchantment", effectstest.SqSigardasOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	equip := effectstest.CastCatalogSpell(t, g, "Bonesplitter", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.LatestTriggerPromptFor(g, me.ID); ask != nil {
		if err := g.ResolveTriggerPrompt(ask.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if c := effectstest.SqCard(t, g, equip); c.AttachedTo.ID != bear {
		t.Errorf("the Equipment is attached to %v, want the bear", c.AttachedTo)
	}
}

// --- Touch the Spirit Realm ------------------------------------------

func TestTouchTheSpiritRealmExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	theirs := effectstest.DcSeed(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	touch := effectstest.CastCatalogSpell(t, g, "Touch the Spirit Realm", "Enchantment", effectstest.SqTouchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !effectstest.InZone(g.Exile, theirs) {
		t.Fatal("the creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(touch) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.DcSeed(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	touch := effectstest.PushCatalogHandCard(me, "Touch the Spirit Realm", "Enchantment", effectstest.SqTouchOracle)
	effectstest.DcMana(t, g, me.ID, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, touch, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !effectstest.InZone(me.Graveyard, touch) {
		t.Error("channel did not discard the card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the artifact was not exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	mine := effectstest.DcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DcSeed(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	other := effectstest.DcSeed(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	otherTheirs := effectstest.DcSeed(g, opp.ID, "Other Theirs", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fated Clash", "Sorcery", effectstest.SqFatedClashOracl, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a target did not survive")
	}
	if g.Battlefield.Contains(other) || g.Battlefield.Contains(otherTheirs) {
		t.Error("a creature that was not a target survived")
	}
}

func TestFatedClashFlashOnlyWithAnAttackerAndABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 0, 4)
	clash := effectstest.PushCatalogHandCard(me, "Fated Clash", "Sorcery", effectstest.SqFatedClashOracl)
	fund := func() {
		g.WithWriteLock(func() {
			me.ManaPool = nil
			for _, c := range []string{"W", "W", "C", "C", "C"} {
				me.ManaPool.AddMana(game.ManaToken{Color: c})
			}
		})
	}
	flash := game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}, {Kind: game.TargetCard, ID: blocker}}}
	if cardsex.FatedClashCreaturesInCombat(g, me.ID) {
		t.Fatal("the condition holds outside combat")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if cardsex.FatedClashCreaturesInCombat(g, me.ID) {
		t.Error("the condition holds with an attacker and no blocker")
	}
	fund()
	if err := g.CastSpell(me.ID, clash, flash); err == nil {
		t.Fatal("flash-cast with an attacker and no blocker")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if !cardsex.FatedClashCreaturesInCombat(g, me.ID) {
		t.Error("the condition does not hold with an attacker and a blocker")
	}
	fund()
	if err := g.CastSpell(me.ID, clash, flash); err != nil {
		t.Fatalf("flash-cast with an attacker and a blocker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) || !g.Battlefield.Contains(blocker) {
		t.Error("a target did not survive the flashed Clash")
	}
}

// --- Rinoa, Angel Wing -----------------------------------------------

func TestRinoaPumpsFliersAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rinoa, Angel Wing", "Legendary Creature — Human Rebel Warlock", effectstest.SqRinoaOracle, false)
	flier := effectstest.PushVanillaCreature(g, me.ID, "Bird", 1, 1)
	ground := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SqCard(t, g, flier); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("flier is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, flier, "vigilance") {
		t.Error("the flier lacks vigilance")
	}
	if c := effectstest.SqCard(t, g, ground); c.CurrentPower() != 2 || effectstest.EffectiveAbilitiesContain(t, g, ground, "vigilance") {
		t.Error("a creature without flying was pumped")
	}
}

func TestRinoaReturnsADeadAttackerOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rinoa, Angel Wing", "Legendary Creature — Human Rebel Warlock", effectstest.SqRinoaOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(a) })
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil || !effectstest.HasID(p.ChooseCards, a) {
		t.Fatalf("no prompt offering the dead attacker: %+v", p)
	}
	effectstest.SqAnswerCards(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if p2 := effectstest.LatestChooseCards(g, me.ID); p2 != nil && p2 != p && effectstest.HasID(p2.ChooseCards, b) {
		t.Error("Rinoa offered a second return in the same turn")
	}
	if g.Battlefield.Contains(b) {
		t.Error("the second attacker came back")
	}
}

// --- Zenos yae Galvus // Shinryu -------------------------------------

func zenosCard(owner uuid.UUID) game.Card {
	c := effectstest.PraetorCard(owner, cardsx.ZenosYaeGalvusOracleID, "Zenos yae Galvus", "Legendary Creature — Human Noble Warrior", 4, 4,
		"Shinryu, Transcendent Rival", "Legendary Creature — Dragon", false)
	c.Faces[0].ManaCost = "{3}{B}{B}"
	c.Faces[1].Power, c.Faces[1].Toughness = 8, 8
	c.SetFace(0)
	return c
}

func TestZenosShrinksTheOthersAndTransformsWhenTheChosenLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	chosen := effectstest.DcSeed(g, opp.ID, "Rival", "Creature — Human", "", 3, 3)
	other := effectstest.DcSeed(g, opp.ID, "Other", "Creature — Bear", "", 3, 3)
	mine := effectstest.DcSeed(g, me.ID, "Mine", "Creature — Bear", "", 3, 3)
	z := zenosCard(me.ID)
	me.Hand.PushTop(z)
	if err := g.CastSpell(me.ID, z.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Zenos: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no choice of a creature")
	}
	if effectstest.HasID(p.ChooseCards, mine) || !effectstest.HasID(p.ChooseCards, chosen) {
		t.Errorf("candidates %v, want only opponents' creatures", p.ChooseCards)
	}
	effectstest.SqAnswerCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SqCard(t, g, chosen); c.CurrentPower() != 3 {
		t.Errorf("chosen creature is %d power, want 3", c.CurrentPower())
	}
	for _, id := range []uuid.UUID{other, mine} {
		if c := effectstest.SqCard(t, g, id); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
			t.Errorf("%s is %d/%d, want 1/1", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := effectstest.SqCard(t, g, z.InstanceID); c.CurrentPower() != 4 {
		t.Errorf("Zenos is %d power, want 4", c.CurrentPower())
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(chosen) })
	effectstest.PassPriorityAroundTable(t, g)
	// Shinryu asks for an opponent as it transforms.
	effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	zc := effectstest.SqCard(t, g, z.InstanceID)
	if zc.Name != "Shinryu, Transcendent Rival" {
		t.Fatalf("Zenos did not transform: %q", zc.Name)
	}
	if zc.ChosenPlayer != opp.ID {
		t.Errorf("Shinryu chose %v, want the opponent", zc.ChosenPlayer)
	}

	// The chosen player loses the game: you win.
	g.WithWriteLock(func() { _, _ = g.LoseTheGameForEffect(opp.ID, uuid.Nil) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.State == game.StateActive {
		t.Error("the game is still going after the chosen player lost")
	}
}

func TestZenosDoesNotTransformWhenAnotherCreatureLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	chosen := effectstest.DcSeed(g, opp.ID, "Rival", "Creature — Human", "", 5, 5)
	other := effectstest.DcSeed(g, opp.ID, "Other", "Creature — Bear", "", 5, 5)
	z := zenosCard(me.ID)
	me.Hand.PushTop(z)
	if err := g.CastSpell(me.ID, z.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Zenos: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SqAnswerCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SqCard(t, g, z.InstanceID); c.Name != "Zenos yae Galvus" {
		t.Errorf("Zenos transformed when an unchosen creature left: %q", c.Name)
	}
}

// --- Ardenn, Intrepid Archaeologist ----------------------------------

func TestArdennMovesChosenEquipmentToTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ardenn, Intrepid Archaeologist", "Legendary Creature — Kor Scout", effectstest.SqArdennOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	e1 := effectstest.DcSeed(g, me.ID, "Sword One", "Artifact — Equipment", "", 0, 0)
	e2 := effectstest.DcSeed(g, me.ID, "Sword Two", "Artifact — Equipment", "", 0, 0)
	rock := effectstest.DcSeed(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChooseCards(g, me.ID)
	if p == nil {
		t.Fatal("no choice of attachments")
	}
	if !effectstest.HasID(p.ChooseCards, e1) || !effectstest.HasID(p.ChooseCards, e2) || effectstest.HasID(p.ChooseCards, rock) {
		t.Errorf("candidates %v, want the two Equipment only", p.ChooseCards)
	}
	if p.ChooseMin != 0 {
		t.Errorf("floor %d, want 0 (you may)", p.ChooseMin)
	}
	effectstest.SqAnswerCards(t, g, me.ID, e1)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SqCard(t, g, e1); c.AttachedTo.ID != bear {
		t.Error("the chosen Equipment was not attached")
	}
	if c := effectstest.SqCard(t, g, e2); c.AttachedTo.ID == bear {
		t.Error("an unchosen Equipment was attached")
	}
}

func TestArdennOffersNoEquipmentForANoncreatureOrAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SqActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ardenn, Intrepid Archaeologist", "Legendary Creature — Kor Scout", effectstest.SqArdennOracle, false)
	effectstest.DcSeed(g, me.ID, "Sword One", "Artifact — Equipment", "", 0, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCards(g, me.ID) != nil {
		t.Error("an Equipment was offered for a player")
	}
}
