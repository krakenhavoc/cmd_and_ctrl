package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b095_cards_test.go — Zendikar Rising, batch 095 (#5508). Every test
// plays the real catalog entry through the engine.

const (
	oldB095RelicGolemOracle         = "14d9bb24-df7b-4f41-aeca-aad639b280db"
	oldB095SneakingGuideOracle      = "3eabd27c-bd65-479a-9743-e90e4273c016"
	oldB095StoneworkPackbeastOracle = "782c47c5-b108-415c-9b70-a033ac52956f"
	oldB095LinvalaOracle            = "afd8ff3e-1d57-4fc5-91dc-1a8400c98787"
	oldB095TazriOracle              = "ce703e1d-b7b2-4b8b-a12c-714fc152d0a5"
	oldB095KorBlademasterOracle     = "194260af-aba5-4d18-a69b-eedb97ed4a4b"
	oldB095ShatterskullOracle       = "2019004c-ebcf-4665-8736-f01ce7d8fcd2"
	oldB095MerfolkFalconerOracle    = "77d5d435-d794-46c0-b630-a460973cf75a"
	oldB095VeteranAdventurerOracle  = "ea3401a9-76db-4f38-b973-2b7754087b97"
	oldB095DeadlyAllianceOracle     = "0d2cdcc6-da65-4f66-ba38-111efad44919"
	oldB095JourneyOracle            = "31e8c1e1-520a-4e64-a221-203afc819fcf"
	oldB095ShadowsVerdictOracle     = "83e99691-bc25-4b3e-b828-670ef2027592"
	oldB095ScorchRiderOracle        = "8165acd4-ac62-47f1-97dd-01860aee29ee"
)

func TestOldB095EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB095RelicGolemOracle, oldB095SneakingGuideOracle, oldB095StoneworkPackbeastOracle, oldB095LinvalaOracle,
		oldB095TazriOracle, oldB095KorBlademasterOracle, oldB095ShatterskullOracle, oldB095MerfolkFalconerOracle,
		oldB095VeteranAdventurerOracle, oldB095DeadlyAllianceOracle, oldB095JourneyOracle, oldB095ShadowsVerdictOracle,
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

// oldB095FillGraveyard leaves p with n cards in their graveyard, the last one
// arriving by a real zone move, which drops the layer cache.
func oldB095FillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n-1; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	effectstest.B27Kill(g, effectstest.B12Creature(g, p.ID, "Doomed Bear", "Creature — Bear", 1, 1))
}

func oldB095PartySize(g *game.Game, who uuid.UUID) (n int) {
	g.WithWriteLock(func() { n = effects.FinB9PartySize(g, who) })
	return n
}

// oldB095Party seats a Cleric, a Rogue and a Warrior for `who`.
func oldB095Party(g *game.Game, who uuid.UUID) {
	effectstest.B12Creature(g, who, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, who, "Thief", "Creature — Human Rogue", 1, 1)
	effectstest.B12Creature(g, who, "Fighter", "Creature — Human Warrior", 1, 1)
}

// oldB095Cast casts a spell with a real mana cost, strictly, from `p`'s hand.
func oldB095Cast(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, targets []game.TargetRef) error {
	_, err := effectstest.CastFromHandErr(g, p, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true, Targets: targets})
	return err
}

// --- Relic Golem --------------------------------------------------------------

func TestOldB095RelicGolemCanAttackAndBlockOnlyAgainstEightInAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	golem := effectstest.B12Push(g, me.ID, "Relic Golem", "Artifact Creature — Golem", oldB095RelicGolemOracle, 6, 6)
	oldB095FillGraveyard(g, opp, 7)
	effectstest.AssertRestrictions(t, g, golem, game.CantAttack|game.CantBlock)
	oldB095FillGraveyard(g, opp, 1)
	effectstest.AssertRestrictions(t, g, golem, 0)
}

func TestOldB095RelicGolemIgnoresItsOwnersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	golem := effectstest.B12Push(g, me.ID, "Relic Golem", "Artifact Creature — Golem", oldB095RelicGolemOracle, 6, 6)
	oldB095FillGraveyard(g, me, 10)
	effectstest.AssertRestrictions(t, g, golem, game.CantAttack|game.CantBlock)
}

func TestOldB095RelicGolemMillsTheTargetPlayerTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	golem := effectstest.PushCatalogPermanent(g, me.ID, "Relic Golem", "Artifact Creature — Golem", oldB095RelicGolemOracle, false)
	library := opp.Library.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("the opponent milled %d, want 2", got)
	}
}

// --- Sneaking Guide -----------------------------------------------------------

func TestOldB095SneakingGuideMakesASmallCreatureUnblockableAndRefusesABigOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	guide := effectstest.PushCatalogPermanent(g, me.ID, "Sneaking Guide", "Creature — Goblin Rogue", oldB095SneakingGuideOracle, false)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, guide, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(big)}); err == nil {
		t.Fatal("targeted a creature with power 3")
	}
	effectstest.B16Activate(t, g, me.ID, guide, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(small)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, small, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, big, 0)
}

// --- Stonework Packbeast and Veteran Adventurer -------------------------------

func TestOldB095AlsoAClericRogueWarriorAndWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	beast := effectstest.B12Push(g, me.ID, "Stonework Packbeast", "Artifact Creature — Beast", oldB095StoneworkPackbeastOracle, 2, 1)
	adv := effectstest.B12Push(g, me.ID, "Veteran Adventurer", "Creature — Human", oldB095VeteranAdventurerOracle, 5, 5)
	for _, id := range []uuid.UUID{beast, adv} {
		have := effectstest.EffectiveSubtypes(t, g, id)
		for _, st := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
			if !effects.EotHasType(have, st) {
				t.Errorf("%v is missing %s", have, st)
			}
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, adv, "vigilance") {
		t.Error("the Adventurer has no vigilance")
	}
	if n := oldB095PartySize(g, me.ID); n != 2 {
		t.Errorf("a party of %d from two four-type creatures, want 2", n)
	}
}

func TestOldB095StoneworkPackbeastFillsTheLastSlotOfAFullParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	oldB095Party(g, me.ID)
	if n := oldB095PartySize(g, me.ID); n != 3 {
		t.Fatalf("party of %d before the Packbeast, want 3", n)
	}
	effectstest.B12Push(g, me.ID, "Stonework Packbeast", "Artifact Creature — Beast", oldB095StoneworkPackbeastOracle, 2, 1)
	if n := oldB095PartySize(g, me.ID); n != 4 {
		t.Errorf("party of %d with the Packbeast, want 4", n)
	}
}

func TestOldB095StoneworkPackbeastMakesAnyColorForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Stonework Packbeast", "Artifact Creature — Beast", oldB095StoneworkPackbeastOracle, 2, 1)
	effectstest.SamiGiveColorless(me, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v after paying {2} for one blue mana, want [U]", got)
	}
}

// --- party discounts ----------------------------------------------------------

func TestOldB095ShatterskullMinotaurCostsLessPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "C")
	if err := oldB095Cast(g, me, "Shatterskull Minotaur", "Creature — Minotaur Warrior", "{4}{R}{R}", oldB095ShatterskullOracle, nil); err == nil {
		t.Fatal("cast {4}{R}{R} for three mana with a party of two")
	}
	effectstest.B06AddMana(me, "C")
	if err := oldB095Cast(g, me, "Shatterskull Minotaur", "Creature — Minotaur Warrior", "{4}{R}{R}", oldB095ShatterskullOracle, nil); err != nil {
		t.Fatalf("cast for {2}{R}{R}: %v", err)
	}
}

func TestOldB095VeteranAdventurerCostsLessPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	oldB095Party(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G", "C")
	if err := oldB095Cast(g, me, "Veteran Adventurer", "Creature — Human", "{5}{G}", oldB095VeteranAdventurerOracle, nil); err == nil {
		t.Fatal("cast {5}{G} for two mana with a party of three")
	}
	effectstest.B06AddMana(me, "C")
	if err := oldB095Cast(g, me, "Veteran Adventurer", "Creature — Human", "{5}{G}", oldB095VeteranAdventurerOracle, nil); err != nil {
		t.Fatalf("cast for {2}{G}: %v", err)
	}
}

func TestOldB095DeadlyAllianceCostsLessAndDestroysACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "C", "C")
	if err := oldB095Cast(g, me, "Deadly Alliance", "Instant", "{4}{B}", oldB095DeadlyAllianceOracle, effectstest.CardRefs(victim)); err == nil {
		t.Fatal("cast {4}{B} for three mana with a party of one")
	}
	effectstest.B06AddMana(me, "C")
	if err := oldB095Cast(g, me, "Deadly Alliance", "Instant", "{4}{B}", oldB095DeadlyAllianceOracle, effectstest.CardRefs(land)); err == nil {
		t.Fatal("targeted a land")
	}
	if err := oldB095Cast(g, me, "Deadly Alliance", "Instant", "{4}{B}", oldB095DeadlyAllianceOracle, effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("cast for {3}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
}

func TestOldB095JourneyToOblivionExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C")
	if err := oldB095Cast(g, me, "Journey to Oblivion", "Enchantment", "{4}{W}", oldB095JourneyOracle, nil); err == nil {
		t.Fatal("cast {4}{W} for two mana with a party of two")
	}
	effectstest.B06AddMana(me, "C")
	id, err := effectstest.CastFromHandErr(g, me, "Journey to Oblivion", "Enchantment", "{4}{W}", oldB095JourneyOracle, game.CastSpellParams{Strict: true})
	if err != nil {
		t.Fatalf("cast for {2}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || effectstest.TmtB1Count(g, opp.ID, "Ogre") != 0 {
		t.Fatal("the target was not exiled")
	}
	effectstest.B27Kill(g, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, opp.ID, "Ogre") != 1 {
		t.Error("the exiled permanent did not return when the enchantment left")
	}
}

// --- Shadows' Verdict ---------------------------------------------------------

func TestOldB095ShadowsVerdictExilesSmallCreaturesPlaneswalkersAndTheirGraveyardCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	smallMine := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	bigTheirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	smallTheirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Jace", ManaCost: "{1}{U}", Owner: opp.ID, Controller: opp.ID})
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	deadSmall := uuid.New()
	deadBig := uuid.New()
	deadSpell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: deadSmall, Name: "Dead Small", TypeLine: "Creature — Bear", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: deadBig, Name: "Dead Big", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: deadSpell, Name: "Dead Spell", TypeLine: "Sorcery", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C")
	if err := oldB095Cast(g, me, "Shadows' Verdict", "Sorcery", "{3}{B}{B}", oldB095ShadowsVerdictOracle, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"your small creature": smallMine, "their cheap creature": smallTheirs, "their planeswalker": walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"their big creature": bigTheirs, "their artifact": rock} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was exiled", name)
		}
	}
	if me.Graveyard.Contains(deadSmall) {
		t.Error("the cheap creature card stayed in your graveyard")
	}
	if !opp.Graveyard.Contains(deadBig) || !opp.Graveyard.Contains(deadSpell) {
		t.Error("an expensive creature card or a sorcery was exiled from their graveyard")
	}
	if z := effectstest.B12ZoneOf(g, smallMine); z != game.ZoneExile {
		t.Errorf("your small creature is in %v, want exile", z)
	}
	if z := effectstest.B12ZoneOf(g, deadSmall); z != game.ZoneExile {
		t.Errorf("the cheap creature card is in %v, want exile", z)
	}
}

// --- Merfolk Falconer ---------------------------------------------------------

func TestOldB095MerfolkFalconerScriesOnlyForKickedSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	falconer := effectstest.B12Push(g, me.ID, "Merfolk Falconer", "Creature — Merfolk Wizard", oldB095MerfolkFalconerOracle, 4, 4)
	if !effectstest.HasEffectiveKeyword(t, g, falconer, "flying") {
		t.Error("no flying")
	}
	effectstest.CastHoldCreature(t, g, "Scorch Rider", "Creature — Human Warrior", oldB095ScorchRiderOracle, 4, 3)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("an unkicked spell scried")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Scorch Rider", TypeLine: "Creature — Human Warrior", OracleID: oldB095ScorchRiderOracle,
		Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("CastSpell kicked: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a kicked spell did not scry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scrying %d cards, want 2", len(c.ScryCards))
	}
}

// --- Kor Blademaster ----------------------------------------------------------

func TestOldB095KorBlademasterGrantsDoubleStrikeToEquippedWarriorsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	master := effectstest.B12Push(g, me.ID, "Kor Blademaster", "Creature — Kor Warrior", oldB095KorBlademasterOracle, 1, 1)
	warrior := effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := effectstest.FinABEquipment(g, me.ID, "Sword")
	if !effectstest.HasEffectiveKeyword(t, g, master, "double strike") {
		t.Error("the Blademaster has no double strike of its own")
	}
	if effectstest.HasEffectiveKeyword(t, g, warrior, "double strike") {
		t.Error("an unequipped Warrior has double strike")
	}
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("attach to the Bear: %v", err)
		}
	})
	if effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("an equipped non-Warrior has double strike")
	}
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: warrior}); err != nil {
			t.Fatalf("attach to the Warrior: %v", err)
		}
	})
	if !effectstest.HasEffectiveKeyword(t, g, warrior, "double strike") {
		t.Error("an equipped Warrior has no double strike")
	}
}

// --- Tazri, Beacon of Unity ---------------------------------------------------

func TestOldB095TazriCostsLessAndLooksForUpToTwoPartyCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	oldB095Party(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W")
	if err := oldB095Cast(g, me, "Tazri, Beacon of Unity", "Legendary Creature — Human Warrior", "{4}{W}", oldB095TazriOracle, nil); err == nil {
		t.Fatal("cast {4}{W} for one mana with a party of three")
	}
	effectstest.B06AddMana(me, "C")
	if err := oldB095Cast(g, me, "Tazri, Beacon of Unity", "Legendary Creature — Human Warrior", "{4}{W}", oldB095TazriOracle, nil); err != nil {
		t.Fatalf("cast for {1}{W}: %v", err)
	}
}

func TestOldB095TazriLookAtSixTakesUpToTwoPartyCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	tazri := effectstest.PushCatalogPermanent(g, me.ID, "Tazri, Beacon of Unity", "Legendary Creature — Human Warrior", oldB095TazriOracle, false)
	me.Library.Cards = nil
	var cleric, rogue, wizard, ally, bear uuid.UUID
	push := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	for i := 0; i < 4; i++ {
		push("Filler", "Sorcery")
	}
	bear = push("Bear", "Creature — Bear")
	ally = push("Ally", "Creature — Kor Ally")
	wizard = push("Wizard", "Creature — Human Wizard")
	rogue = push("Rogue", "Creature — Human Rogue")
	cleric = push("Cleric", "Creature — Human Cleric")
	effectstest.AdvanceToMain(t, g)
	for _, c := range []string{"U", "B", "R", "G", "C", "C", "C", "C"} {
		effectstest.B06AddMana(me, c)
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, tazri, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt for the look: %+v", g.PendingChoices)
	}
	if c.ChooseMax != 2 {
		t.Errorf("may take up to %d, want 2", c.ChooseMax)
	}
	if len(c.ChooseCards) != 4 {
		t.Errorf("%d candidates, want the Ally, Wizard, Rogue and Cleric (not the Bear or the filler)", len(c.ChooseCards))
	}
	for _, id := range c.ChooseCards {
		if id == bear {
			t.Error("the Bear is offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cleric, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || !me.Hand.Contains(cleric) || !me.Hand.Contains(ally) {
		t.Errorf("hand %d (had %d), want the Cleric and the Ally taken", me.Hand.Size(), hand)
	}
	if me.Hand.Contains(bear) || me.Hand.Contains(wizard) || me.Hand.Contains(rogue) {
		t.Error("a card that was not chosen reached the hand")
	}
}

// --- Linvala, Shield of Sea Gate ----------------------------------------------

func oldB095PickOption(t *testing.T, g *game.Game, who uuid.UUID, idx int) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, who)
	if c == nil {
		t.Fatalf("no option prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(c.ID, who, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

func TestOldB095LinvalaDetainsOnlyWithAFullParty(t *testing.T) {
	for _, full := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		linvala := effectstest.B12Push(g, me.ID, "Linvala, Shield of Sea Gate", "Legendary Creature — Angel Wizard", oldB095LinvalaOracle, 3, 3)
		_ = linvala
		effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
		effectstest.B12Creature(g, me.ID, "Thief", "Creature — Human Rogue", 1, 1)
		if full {
			effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 1, 1)
		}
		victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		p := effectstest.LatestPickTarget(g, me.ID)
		if !full {
			if p != nil {
				t.Fatal("Linvala asked for a target without a full party")
			}
			continue
		}
		if p == nil {
			t.Fatal("no target prompt with a full party")
		}
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
		want := game.CantAttack | game.CantBlock | game.CantActivate | game.CantActivateMana
		if got := effectstest.RestrictionsOf(t, g, victim); got&want != want {
			t.Errorf("the target's restrictions are %v, want it detained", got.Names())
		}
	}
}

func TestOldB095LinvalaSacrificeGivesTheChosenKeywordToYourCreatures(t *testing.T) {
	for idx, kw := range []string{"hexproof", "indestructible"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
		linvala := effectstest.PushCatalogPermanent(g, me.ID, "Linvala, Shield of Sea Gate", "Legendary Creature — Angel Wizard", oldB095LinvalaOracle, false)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
		effectstest.AdvanceToMain(t, g)
		effectstest.B16Activate(t, g, me.ID, linvala, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		oldB095PickOption(t, g, me.ID, idx)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(linvala) {
			t.Error("Linvala was not sacrificed")
		}
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("your Bear did not gain %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("their Ogre gained %s", kw)
		}
	}
}
