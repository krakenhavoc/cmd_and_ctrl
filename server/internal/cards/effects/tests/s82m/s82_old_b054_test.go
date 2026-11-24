package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b054_test.go — Commander 2021, batch 54 (#5504): graveyard
// recursion. Each test plays the real catalog entry and asserts what its
// text says.

const (
	oldB054HealingTechnique = "1b99c32d-1d02-410e-84bf-9db2ca05f576"
	oldB054InfernalOffering = "348201c4-79de-4278-9896-d8b1567cc5cc"
	oldB054LivingLore       = "7c48c341-1d56-4210-9d86-181ebb206098"
	oldB054RevivalExp       = "913ec957-5f4e-46b4-a819-9448837f72a3"
	oldB054DebtorsKnell     = "a9ff368b-5208-4ada-bb99-eccbf8eaea28"
	oldB054PulseOfMurasa    = "d5b446a4-a761-42f0-acf0-ea0a1a9538d6"
	oldB054Ruxa             = "d760f2f0-6653-47c4-bd7b-aab2f398ddf2"
	oldB054SproutbackTrudge = "d9826660-0f34-4d82-8915-cf2796ab8d9a"
	oldB054VolcanicVision   = "fe3b4c74-2bf8-464a-8f55-9de401a8c4b9"
	oldB054CommuneWithEvil  = "0248bf01-14e0-4a39-ba3a-5177fe272fe9"
	oldB054StartingLife     = 40
)

func oldB054Seats(g *game.Game) (*game.Player, *game.Player) {
	i := g.Turn.ActiveSeat
	return g.Seats[i], g.Seats[(i+1)%len(g.Seats)]
}

func TestOldB054EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB054HealingTechnique: "Healing Technique",
		oldB054InfernalOffering: "Infernal Offering",
		oldB054LivingLore:       "Living Lore",
		oldB054RevivalExp:       "Revival Experiment",
		oldB054DebtorsKnell:     "Debtors' Knell",
		oldB054PulseOfMurasa:    "Pulse of Murasa",
		oldB054Ruxa:             "Ruxa, Patient Professor",
		oldB054SproutbackTrudge: "Sproutback Trudge",
		oldB054VolcanicVision:   "Volcanic Vision",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is %v, want Full", name, spec.Completeness)
		}
	}
}

// --- Pulse of Murasa --------------------------------------------------------

func TestOldB054PulseOfMurasaReturnsAnyGraveyardsCreatureToItsOwnersHandAndGainsSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	bear := effectstest.B17GraveyardCard(opp, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Pulse of Murasa", "Instant", oldB054PulseOfMurasa, effectstest.CardTargets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) {
		t.Error("the creature card did not return to its owner's hand")
	}
	if me.Hand.Contains(bear) {
		t.Error("the creature card went to the caster's hand")
	}
	if me.Life != oldB054StartingLife+6 {
		t.Errorf("life %d, want %d", me.Life, oldB054StartingLife+6)
	}
}

func TestOldB054PulseOfMurasaTakesALandCardAndRefusesAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	land := effectstest.B17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	spell := effectstest.B17GraveyardCard(me, "Shock", "Instant", "{R}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Pulse of Murasa", "Instant", oldB054PulseOfMurasa, effectstest.CardTargets(spell)); err == nil {
		t.Fatal("targeted an instant card")
	}
	effectstest.CastCatalogSpell(t, g, "Pulse of Murasa", "Instant", oldB054PulseOfMurasa, effectstest.CardTargets(land))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card did not return to hand")
	}
}

// --- Healing Technique ------------------------------------------------------

func oldB054Demonstrate(t *testing.T, g *game.Game, me, opp *game.Player, mine, theirs uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 12; i++ {
		switch {
		case effectstest.CopyPromptFor(g, me.ID) != nil:
			c := effectstest.CopyPromptFor(g, me.ID)
			if err := g.ResolvePickTargets(c.ID, me.ID, effectstest.CardTargets(mine)); err != nil {
				t.Fatalf("my copy keeps its target: %v", err)
			}
		case effectstest.CopyPromptFor(g, opp.ID) != nil:
			c := effectstest.CopyPromptFor(g, opp.ID)
			if err := g.ResolvePickTargets(c.ID, opp.ID, effectstest.CardTargets(theirs)); err != nil {
				t.Fatalf("the opponent's copy picks their card: %v", err)
			}
		case effectstest.LatestOptionPickFor(g, me.ID) != nil:
			effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
		default:
			i = 12
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB054HealingTechniqueReturnsACardGainsItsManaValueAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	card := effectstest.B17GraveyardCard(me, "Ogre", "Creature — Ogre", "{2}{R}")
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastFromHandForTest(t, g, me, "Healing Technique", "Sorcery", "{3}{G}", oldB054HealingTechnique,
		game.CastSpellParams{Targets: effectstest.CardTargets(card)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand")
	}
	if me.Life != oldB054StartingLife+3 {
		t.Errorf("life %d, want %d (the card's mana value)", me.Life, oldB054StartingLife+3)
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Healing Technique was not exiled")
	}
}

func TestOldB054HealingTechniqueRefusesAnOpponentsGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	theirs := effectstest.B17GraveyardCard(opp, "Ogre", "Creature — Ogre", "{2}{R}")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Healing Technique", TypeLine: "Sorcery", ManaCost: "{3}{G}", OracleID: oldB054HealingTechnique, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTargets(theirs), Strict: true}); err == nil {
		t.Fatal("targeted a card in an opponent's graveyard")
	}
}

func TestOldB054HealingTechniqueDemonstrateGivesTheOpponentTheirOwnCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	mine := effectstest.B17GraveyardCard(me, "Ogre", "Creature — Ogre", "{2}{R}")
	mine2 := effectstest.B17GraveyardCard(me, "Imp", "Creature — Imp", "{2}{B}")
	theirs := effectstest.B17GraveyardCard(opp, "Giant", "Creature — Giant", "{4}{R}")
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastFromHandForTest(t, g, me, "Healing Technique", "Sorcery", "{3}{G}", oldB054HealingTechnique,
		game.CastSpellParams{Targets: effectstest.CardTargets(mine)})
	oldB054Demonstrate(t, g, me, opp, mine2, theirs)
	if !me.Hand.Contains(mine) || !me.Hand.Contains(mine2) || !opp.Hand.Contains(theirs) {
		t.Fatalf("hands: mine back=%v/%v, theirs back=%v", me.Hand.Contains(mine), me.Hand.Contains(mine2), opp.Hand.Contains(theirs))
	}
	// The original and my copy each gain 3; the opponent's copy gains 5.
	if me.Life != oldB054StartingLife+3+3 {
		t.Errorf("my life %d, want %d", me.Life, oldB054StartingLife+6)
	}
	if opp.Life != oldB054StartingLife+5 {
		t.Errorf("the opponent's life %d, want %d", opp.Life, oldB054StartingLife+5)
	}
	if !g.Exile.Contains(spell) {
		t.Error("the original spell was not exiled")
	}
}

// --- Volcanic Vision --------------------------------------------------------

func TestOldB054VolcanicVisionReturnsTheSpellAndDamagesOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	bolt := effectstest.B17GraveyardCard(me, "Fireball", "Sorcery", "{2}{R}")
	small := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.CastFromHandForTest(t, g, me, "Volcanic Vision", "Sorcery", "{5}{R}{R}", oldB054VolcanicVision,
		game.CastSpellParams{Targets: effectstest.CardTargets(bolt)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the sorcery card did not return to hand")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived three damage")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("the 4/4 or my own creature died")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Volcanic Vision was not exiled")
	}
}

func TestOldB054VolcanicVisionRefusesACreatureCardTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	bear := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Volcanic Vision", TypeLine: "Sorcery", ManaCost: "{5}{R}{R}", OracleID: oldB054VolcanicVision, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.CardTargets(bear), Strict: true}); err == nil {
		t.Fatal("targeted a creature card")
	}
}

// --- Debtors' Knell ---------------------------------------------------------

func TestOldB054DebtorsKnellReanimatesFromAnyGraveyardUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Debtors' Knell", "Enchantment", oldB054DebtorsKnell, false)
	theirs := effectstest.B17GraveyardCard(opp, "Giant", "Creature — Giant", "{4}{R}")
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	c, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok {
		t.Fatal("the creature card did not enter")
	}
	if c.Controller != me.ID {
		t.Error("it entered under its owner's control, not the Knell's controller's")
	}
}

// --- Ruxa, Patient Professor ------------------------------------------------

func TestOldB054RuxaBuffsOnlyCreaturesWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ruxa, Patient Professor", "Legendary Creature — Bear Druid", oldB054Ruxa, false)
	vanilla := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, NoAbilitiesText: game.PrintedNo, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	if p, tough := effectstest.PtOf(t, g, vanilla); p != 3 || tough != 3 {
		t.Errorf("vanilla Bear is %d/%d, want 3/3", p, tough)
	}
	bp, bt := effectstest.PtOf(t, g, flyer)
	if bp != 1 || bt != 1 {
		t.Errorf("flying Bird is %d/%d, want its printed 1/1", bp, bt)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Wolf is %d/%d, want 2/2", p, tough)
	}
	if !effectstest.ApaLive(g, vanilla).Effective().AssignsAsThoughUnblocked {
		t.Error("the vanilla Bear may not assign its combat damage as though it weren't blocked")
	}
	if effectstest.ApaLive(g, flyer).Effective().AssignsAsThoughUnblocked {
		t.Error("the flying Bird may assign its combat damage as though it weren't blocked")
	}
}

func TestOldB054RuxaReturnsACreatureCardWithNoAbilitiesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	vanilla := uuid.New()
	flyer := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: vanilla, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, NoAbilitiesText: game.PrintedYes, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: flyer, Name: "Wind Drake", TypeLine: "Creature — Drake", Power: 2, Toughness: 2, NoAbilitiesText: game.PrintedNo, Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Ruxa, Patient Professor", "Legendary Creature — Bear Druid", "{2}{G}{G}", oldB054Ruxa, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, vanilla)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(vanilla) {
		t.Error("the creature card with no abilities did not return")
	}
	if me.Hand.Contains(flyer) {
		t.Error("a creature card with abilities was returned")
	}
}

// --- Revival Experiment -----------------------------------------------------

func TestOldB054RevivalExperimentSetRuleIsOneCardPerPermanentType(t *testing.T) {
	artCreature := game.Card{TypeLine: "Artifact Creature — Construct"}
	creature := game.Card{TypeLine: "Creature — Bear"}
	land := game.Card{TypeLine: "Land"}
	if !effects.OldB054OnePerPermanentType([]game.Card{artCreature, creature}) {
		t.Error("an artifact creature and a creature are two types, refused")
	}
	if !effects.OldB054OnePerPermanentType([]game.Card{artCreature, land}) {
		t.Error("an artifact creature and a land refused")
	}
	if effects.OldB054OnePerPermanentType([]game.Card{creature, {TypeLine: "Creature — Wolf"}}) {
		t.Error("two creature-only cards accepted")
	}
	if effects.OldB054OnePerPermanentType([]game.Card{{TypeLine: "Sorcery"}}) {
		t.Error("a sorcery accepted as a permanent type")
	}
	if !effects.OldB054OnePerPermanentType([]game.Card{{TypeLine: "Battle — Siege"}, {TypeLine: "Planeswalker — Jace"}, {TypeLine: "Enchantment"}}) {
		t.Error("a battle, a planeswalker and an enchantment refused")
	}
}

func TestOldB054RevivalExperimentReturnsOnePerTypeLosesThreeEachAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	bear := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	wolf := effectstest.B17GraveyardCard(me, "Wolf", "Creature — Wolf", "{2}{G}")
	land := effectstest.B17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	ench := effectstest.B17GraveyardCard(me, "Pacifism", "Enchantment — Aura", "{1}{W}")
	spell := effectstest.B17GraveyardCard(me, "Shock", "Instant", "{R}")
	effectstest.AdvanceToMain(t, g)
	card := effectstest.CastFromHandForTest(t, g, me, "Revival Experiment", "Sorcery", "{4}{B}{G}", oldB054RevivalExp, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the permanent cards")
	}
	if effectstest.HasID(pick.ChooseCards, spell) || len(pick.ChooseCards) != 4 {
		t.Errorf("offered %d cards, want the four permanent cards", len(pick.ChooseCards))
	}
	// Two creature-only cards are one too many for the creature type.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear, wolf}); err == nil {
		t.Fatal("two creature cards accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear, land, ench}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, land, ench} {
		if !g.Battlefield.Contains(id) {
			t.Error("a chosen card did not enter")
		}
	}
	if !me.Graveyard.Contains(wolf) {
		t.Error("the unchosen creature left the graveyard")
	}
	if me.Life != oldB054StartingLife-9 {
		t.Errorf("life %d, want %d (3 for each of three cards)", me.Life, oldB054StartingLife-9)
	}
	if !g.Exile.Contains(card) {
		t.Error("Revival Experiment was not exiled")
	}
}

func TestOldB054RevivalExperimentMayReturnNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	bear := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Revival Experiment", "Sorcery", "{4}{B}{G}", oldB054RevivalExp, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || me.Life != oldB054StartingLife {
		t.Errorf("returned or lost life with nothing chosen: life %d", me.Life)
	}
}

// --- Infernal Offering ------------------------------------------------------

func TestOldB054InfernalOfferingSacrificesDrawsAndReturnsInOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	myCard := effectstest.B17GraveyardCard(me, "Ogre", "Creature — Ogre", "{2}{R}")
	theirCard := effectstest.B17GraveyardCard(opp, "Giant", "Creature — Giant", "{4}{R}")
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.SeedLibrary(opp, "E", "F", "G", "H")
	myHand, oppHand := len(me.Hand.Cards), len(opp.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Infernal Offering", "Sorcery", oldB054InfernalOffering, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	for _, who := range []struct {
		p  *game.Player
		id uuid.UUID
	}{{me, mine}, {opp, theirs}} {
		if effectstest.SacrificeChoiceFor(g, who.p.ID) != nil {
			effectstest.AnswerSacrifice(t, g, who.p.ID, who.id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a creature was not sacrificed")
	}
	if got := len(me.Hand.Cards) - myHand; got != 2 {
		t.Errorf("I drew %d, want 2", got)
	}
	if got := len(opp.Hand.Cards) - oppHand; got != 2 {
		t.Errorf("the opponent drew %d, want 2", got)
	}
	// The second "choose an opponent", then my pick first.
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("I was not asked for a creature card")
	}
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("the opponent was asked before my card entered")
	}
	// Both sacrificed creatures are now in graveyards: mine offers the
	// Bear too.
	effectstest.AnswerChooseCards(t, g, me.ID, myCard)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(myCard) {
		t.Error("my creature card did not enter")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, theirCard)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirCard) {
		t.Error("the opponent's creature card did not enter")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == theirCard && c.Controller != opp.ID {
			t.Error("the opponent's card entered under the wrong control")
		}
	}
}

func TestOldB054InfernalOfferingOnlyThePlayerWhoSacrificedDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.SeedLibrary(opp, "E", "F", "G", "H")
	myHand, oppHand := len(me.Hand.Cards), len(opp.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Infernal Offering", "Sorcery", oldB054InfernalOffering, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSacrifice(t, g, me.ID, mine)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - myHand; got != 2 {
		t.Errorf("I drew %d, want 2", got)
	}
	if got := len(opp.Hand.Cards) - oppHand; got != 0 {
		t.Errorf("the opponent with no creature drew %d, want 0", got)
	}
}

// --- Living Lore ------------------------------------------------------------

func oldB054CastLore(t *testing.T, g *game.Game, me *game.Player, exile uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	lore := effectstest.CastFromHandForTest(t, g, me, "Living Lore", "Creature — Avatar", "{3}{U}", oldB054LivingLore, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the instant or sorcery card to exile")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("the exile is [%d,%d], want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, exile)
	effectstest.PassPriorityAroundTable(t, g)
	return lore
}

func TestOldB054LivingLoreExilesAnInstantOrSorcerySizedByItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	creature := effectstest.B17GraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.B17GraveyardCard(me, "Fireball", "Sorcery", "{3}{R}")
	other := effectstest.B17GraveyardCard(me, "Shock", "Instant", "{R}")
	effectstest.AdvanceToMain(t, g)
	lore := effectstest.CastFromHandForTest(t, g, me, "Living Lore", "Creature — Avatar", "{3}{U}", oldB054LivingLore, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if effectstest.HasID(pick.ChooseCards, creature) || len(pick.ChooseCards) != 2 {
		t.Errorf("offered %d cards, want the instant and the sorcery only", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || !me.Graveyard.Contains(other) {
		t.Error("the wrong card was exiled")
	}
	if p, tough := effectstest.PtOf(t, g, lore); p != 4 || tough != 4 {
		t.Errorf("Living Lore is %d/%d, want 4/4 (mana value 4)", p, tough)
	}
}

func TestOldB054LivingLoreDiesWithNoInstantOrSorceryToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	effectstest.AdvanceToMain(t, g)
	lore := effectstest.CastFromHandForTest(t, g, me, "Living Lore", "Creature — Avatar", "{3}{U}", oldB054LivingLore, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lore) {
		t.Error("a 0/0 Living Lore stayed on the battlefield")
	}
}

func TestOldB054LivingLoreSacrificesAfterCombatDamageToCastTheExiledCardFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Commune with Evil", TypeLine: "Sorcery", ManaCost: "{2}{B}", OracleID: oldB054CommuneWithEvil, Owner: me.ID, Controller: me.ID})
	})
	effectstest.SeedLibrary(me, "One", "Two", "Three", "Four", "Five")
	lore := oldB054CastLore(t, g, me, spell)
	effectstest.DealCombatDamageToPlayer(g, lore, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lore) {
		t.Fatal("Living Lore was not sacrificed")
	}
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast the exiled card")
	}
	if effectstest.KeywordCastCard(prompt) != spell {
		t.Error("the offer names a different card")
	}
	effectstest.AnswerKeywordCast(t, g, prompt, true)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChooseCardsFor(g, me.ID); pick != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards[0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != oldB054StartingLife+3 {
		t.Errorf("life %d, want %d: the exiled Commune with Evil was not cast", me.Life, oldB054StartingLife+3)
	}
}

func TestOldB054LivingLoreDecliningTheSacrificeKeepsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB054Seats(g)
	spell := effectstest.B17GraveyardCard(me, "Fireball", "Sorcery", "{3}{R}")
	lore := oldB054CastLore(t, g, me, spell)
	effectstest.DealCombatDamageToPlayer(g, lore, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lore) || !g.Exile.Contains(spell) {
		t.Error("declining changed the board")
	}
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil {
		t.Error("offered a cast without a sacrifice")
	}
}

// --- Sproutback Trudge ------------------------------------------------------

func oldB054TrudgeInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sproutback Trudge", TypeLine: "Creature — Fungus Beast", ManaCost: "{7}{G}{G}",
		Power: 9, Toughness: 7, OracleID: oldB054SproutbackTrudge, Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB054SproutbackTrudgeOffersTheCastAtEndStepOnlyAfterGainingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	trudge := oldB054TrudgeInGraveyard(me)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.KeywordCastPromptFor(g, me.ID) != nil {
		t.Fatal("offered the cast without gaining life this turn")
	}
	if !me.Graveyard.Contains(trudge) {
		t.Fatal("it left the graveyard")
	}
}

func TestOldB054SproutbackTrudgeCostsLessByLifeGainedAndCanBeCastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	trudge := oldB054TrudgeInGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.GainLifeNow(t, g, me.ID, me, 7)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast it from the graveyard after gaining life")
	}
	// {7}{G}{G} less the 7 life gained is {G}{G}: one {G} is not enough.
	granted := game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, trudge, granted); err == nil {
		t.Fatal("cast it for {G} alone")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, trudge, granted); err != nil {
		t.Fatalf("cast for {G}{G}: %v", err)
	}
	_ = prompt
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(trudge) {
		t.Error("Sproutback Trudge did not resolve onto the battlefield for {G}{G}")
	}
}

func TestOldB054SproutbackTrudgeDecliningLeavesItInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB054Seats(g)
	trudge := oldB054TrudgeInGraveyard(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.GainLifeNow(t, g, me.ID, me, 2)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer")
	}
	effectstest.AnswerKeywordCast(t, g, prompt, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(trudge) {
		t.Error("declining moved it")
	}
}
