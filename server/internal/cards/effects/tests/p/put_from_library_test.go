package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLibraryToBattlefieldCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness effects.Completeness
	}{
		// #783 took Chaos Warp to `full`: the shuffle and the reveal
		// are the tuck's continuation now, so a commander's owner who
		// declines the command zone has it shuffled IN rather than
		// landing on top of an already-shuffled library.
		effectstest.PlChaosWarpOracle:            {"Chaos Warp", effects.CompletenessFull},
		effectstest.PlCoilingOracleOracle:        {"Coiling Oracle", effects.CompletenessFull},
		effectstest.PlGenesisWaveOracle:          {"Genesis Wave", effects.CompletenessFull},
		effectstest.PlRisenReefOracle:            {"Risen Reef", effects.CompletenessFull},
		effectstest.PlExplorersScopeOracle:       {"Explorer's Scope", effects.CompletenessFull},
		effectstest.PlThrasiosOracle:             {"Thrasios, Triton Hero", effects.CompletenessFull},
		effectstest.PlAnimistsOracle:             {"Animist's Awakening", effects.CompletenessFull},
		effectstest.PlRegaliaOracle:              {"The Regalia", effects.CompletenessFull},
		effectstest.PlLurkingOracle:              {"Lurking Predators", effects.CompletenessFull},
		effectstest.PlAtlaPalaniOracle:           {"Atla Palani, Nest Tender", effects.CompletenessFull},
		effectstest.PlUreniOracle:                {"Ureni of the Unwritten", effects.CompletenessFull},
		effectstest.PlMajesticOracle:             {"Majestic Genesis", effects.CompletenessFull},
		effectstest.PlGilgameshOracle:            {"Gilgamesh, Master-at-Arms", effects.CompletenessFull},
		effectstest.PlSkyhunterOracle:            {"Armored Skyhunter", effects.CompletenessFull},
		effects.EsikaGodOfTheTreeOracleID:        {"Esika, God of the Tree", effects.CompletenessFull},
		effects.EsikaGodOfTheTreeOracleID + "#1": {"The Prismatic Bridge", effects.CompletenessFull},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", want.name, oracle)
			continue
		}
		if spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: registered as %q with completeness %v, want %q / %v",
				oracle, spec.Name, spec.Completeness, want.name, want.completeness)
		}
	}
}

// --- Coiling Oracle ----------------------------------------------------

func TestCoilingOraclePutsARevealedLandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	before := len(g.Events)

	effectstest.CastCatalogSpell(t, g, "Coiling Oracle", "Creature — Snake Elf Druid", effectstest.PlCoilingOracleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(land) {
		t.Fatal("the revealed land did not enter")
	}
	if effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("nothing said tapped")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d: putting a land is not playing one (CR 305.4)", n)
	}
	if effectstest.PlSearchFired(g, before) {
		t.Error("a reveal is not a search")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("there is no choice on Coiling Oracle")
	}
}

func TestCoilingOraclePutsANonlandIntoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.PlTop(me, "Lightning Bolt", "Instant", "{R}")
	effectstest.CastCatalogSpell(t, g, "Coiling Oracle", "Creature — Snake Elf Druid", effectstest.PlCoilingOracleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the revealed nonland card did not go to hand")
	}
}

// --- Chaos Warp --------------------------------------------------------

// The owner shuffles the permanent in, then reveals and puts a
// permanent card onto the battlefield under THEIR control.
func TestChaosWarpReplacesThePermanentForItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	// A library of lands only, so whatever comes up is a permanent.
	opp.Library.Cards = nil
	for i := 0; i < 4; i++ {
		effectstest.PlTop(opp, "Forest", "Basic Land — Forest", "")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	onBattlefield := g.Battlefield.Size()
	before := len(g.Events)

	effectstest.CastCatalogSpell(t, g, "Chaos Warp", "Instant", effectstest.PlChaosWarpOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)

	// The shuffle decides whether a Forest or the Bear itself comes up,
	// so pin the card by the reveal rather than by the seed: exactly one
	// card was revealed, and that card is the one that entered.
	var revealed []uuid.UUID
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventRevealCards && ev.Actor == opp.ID {
			revealed = append(revealed, ev.CardID)
		}
	}
	if len(revealed) != 1 {
		t.Fatalf("revealed %d cards, want the top one", len(revealed))
	}
	entered, ok := g.LookupCardForEffect(revealed[0])
	if !ok || !g.Battlefield.Contains(revealed[0]) {
		t.Fatal("the revealed card is not the one that entered")
	}
	if entered.Controller != opp.ID {
		t.Errorf("%s entered under %s, want its owner", entered.Name, entered.Controller)
	}
	if g.Battlefield.Size() != onBattlefield {
		t.Errorf("battlefield %d → %d: one permanent left and one arrived", onBattlefield, g.Battlefield.Size())
	}
	if opp.Library.Size() != 4 {
		t.Errorf("owner's library = %d, want 4 (four lands + the bear − the card put)", opp.Library.Size())
	}
	if entered.Name != "Bear" && g.Battlefield.Contains(bear) {
		t.Error("a Forest came up, yet the Bear is still on the battlefield")
	}
}

// #773 review: a token is not a card (CR 108.2) and a token that has
// left the battlefield can't come back (CR 111.8). Chaos Warp on a token
// whose owner's library is empty tucks it, reveals it as the top card —
// and must not put it straight back, or the removal removed nothing.
func TestChaosWarpOnATokenDoesNotPutItBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	onBattlefield := g.Battlefield.Size()
	before := len(g.Events)

	effectstest.CastCatalogSpell(t, g, "Chaos Warp", "Instant", effectstest.PlChaosWarpOracle, effectstest.B16TargetCard(goblin))
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(goblin) {
		t.Fatal("the warped token came back onto the battlefield")
	}
	if g.Battlefield.Size() != onBattlefield-1 {
		t.Errorf("battlefield %d → %d, want one fewer", onBattlefield, g.Battlefield.Size())
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventEffectError {
			t.Errorf("a revealed token must read as \"not a permanent card\", not an engine error: %s", ev.ErrorMsg)
		}
	}
}

// A token already sitting in a library (from an earlier tuck) is never
// offered by "any number of permanent cards from among them", and it
// is not moved with the rest either (CR 111.8).
func TestGenesisWaveNeverOffersAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	token := effectstest.PlTop(me, "Goblin", "Token Creature — Goblin", "")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")

	effectstest.B12PlayFromHand(t, g, "Genesis Wave", "Sorcery", effectstest.PlGenesisWaveOracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Genesis Wave asked nothing")
	}
	if effectstest.HasID(pick.ChooseCards, token) {
		t.Error("a token was offered as a permanent card")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	if g.Battlefield.Contains(token) || me.Graveyard.Contains(token) {
		t.Error("the token changed zones")
	}
}

// The reveal-until family does not stop on a token: it is not "a land
// card", and "put that card onto the battlefield" could not bring it
// back (CR 111.8).
func TestTheRegaliaDoesNotStopOnAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	regalia := effectstest.B12Push(g, me.ID, "The Regalia", "Legendary Artifact Creature — Vehicle", effectstest.PlRegaliaOracle, 4, 4)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	token := effectstest.PlTop(me, "Dryad Arbor Copy", "Token Land Creature — Forest Dryad", "")

	effectstest.DeclareAttack(t, g, opp.ID, regalia)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Fatal("the reveal stopped on a token and put it onto the battlefield")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the reveal should have carried on to the land card under the token")
	}
}

// --- Genesis Wave ------------------------------------------------------

func TestGenesisWavePutsChosenPermanentsAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	notRevealed := effectstest.PlTop(me, "Deep Bolt", "Instant", "{R}")
	big := effectstest.PlTop(me, "Big", "Creature — Giant", "{5}")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	forest := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	before := len(g.Events)

	effectstest.B12PlayFromHand(t, g, "Genesis Wave", "Sorcery", effectstest.PlGenesisWaveOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Genesis Wave asked nothing")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2 (any number of the two eligible)", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, big) {
		t.Error("a mana value 5 card was offered for X = 3")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest, bear)

	if !g.Battlefield.Contains(forest) || !g.Battlefield.Contains(bear) {
		t.Fatal("the chosen permanents did not enter")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("a revealed card that was not put onto the battlefield must go to the graveyard")
	}
	if !me.Library.Contains(notRevealed) {
		t.Error("the fourth card was never revealed and must stay in the library")
	}
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventMill {
			t.Error("Genesis Wave's graveyard move is not a mill")
		}
	}
}

// --- Risen Reef --------------------------------------------------------

func TestRisenReefOffersTheLandTappedAndSendsADeclineToHand(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
		effectstest.CastCatalogSpell(t, g, "Risen Reef", "Creature — Elemental", effectstest.PlRisenReefOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)

		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("Risen Reef did not offer the land")
		}
		if accept {
			effectstest.AnswerChooseCards(t, g, me.ID, land)
			if !g.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, g, land) {
				t.Error("accepted: the land enters tapped")
			}
		} else {
			effectstest.AnswerChooseCards(t, g, me.ID)
			if !me.Hand.Contains(land) {
				t.Error("declined: \"if you don't put the card onto the battlefield, put it into your hand\"")
			}
		}
	}
}

func TestRisenReefTriggersForAnotherElementalAndHandsANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Risen Reef", "Creature — Elemental", effectstest.PlRisenReefOracle, 1, 1)
	bolt := effectstest.PlTop(me, "Lightning Bolt", "Instant", "{R}")
	effectstest.CastCatalogSpell(t, g, "Other Elemental", "Creature — Elemental", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a nonland card has nothing to offer")
	}
	if !me.Hand.Contains(bolt) {
		t.Error("another Elemental entering did not look, or the nonland did not go to hand")
	}
}

// --- Explorer's Scope --------------------------------------------------

func TestExplorersScopePutsALandTappedWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	scope := effectstest.SeedEquipment(g, me.ID, "Explorer's Scope", effectstest.PlExplorersScopeOracle)
	effectstest.B06AddMana(me, "C")
	effectstest.EquipTo(t, g, me.ID, scope, bear)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, g, land) {
		t.Error("the land enters tapped")
	}
}

// --- Thrasios ----------------------------------------------------------

func TestThrasiosScriesThenPutsALandOrDraws(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		land     bool
	}{{"a land", "Basic Land — Island", true}, {"a nonland", "Instant", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			thrasios := effectstest.B12Push(g, me.ID, "Thrasios, Triton Hero", "Legendary Creature — Merfolk Wizard", effectstest.PlThrasiosOracle, 1, 3)
			top := effectstest.PlTop(me, "Top", tc.typeLine, "")
			effectstest.B06AddMana(me, "C", "C", "C", "C")
			if err := g.ActivateCatalogAbility(me.ID, thrasios, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			var scry *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceScry {
					scry = c
				}
			}
			if scry == nil {
				t.Fatal("Thrasios did not scry first")
			}
			hand := me.Hand.Size()
			if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
				t.Fatal(err)
			}
			if tc.land {
				if !g.Battlefield.Contains(top) || !effectstest.TappedOnBattlefield(t, g, top) {
					t.Error("the revealed land enters tapped")
				}
				if me.Hand.Size() != hand {
					t.Error("a land was put, so nothing is drawn")
				}
			} else if !me.Hand.Contains(top) || me.Hand.Size() != hand+1 {
				t.Error("otherwise, draw a card")
			}
		})
	}
}

// --- Animist's Awakening -----------------------------------------------

func TestAnimistsAwakeningPutsEveryLandAndUntapsWithSpellMastery(t *testing.T) {
	for _, mastery := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if mastery {
			effectstest.Batch01GraveyardCard(me, "Bolt", "Instant")
			effectstest.Batch01GraveyardCard(me, "Rite", "Sorcery")
		}
		island := effectstest.PlTop(me, "Island", "Basic Land — Island", "")
		bolt := effectstest.PlTop(me, "Bolt", "Instant", "{R}")
		forest := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")

		effectstest.B12PlayFromHand(t, g, "Animist's Awakening", "Sorcery", effectstest.PlAnimistsOracle, game.CastSpellParams{XValue: 3})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
			t.Fatal("\"all land cards\" is not a choice")
		}
		for _, id := range []uuid.UUID{island, forest} {
			if !g.Battlefield.Contains(id) {
				t.Fatal("a revealed land did not enter")
			}
			if effectstest.TappedOnBattlefield(t, g, id) == mastery {
				t.Errorf("mastery %v: tapped %v", mastery, !mastery)
			}
		}
		if !effectstest.PlBottomIDs(me, 1)[bolt] {
			t.Error("the nonland card goes to the bottom")
		}
	}
}

// --- The Regalia -------------------------------------------------------

func TestTheRegaliaRevealsUntilALandOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// Crewed for the test: the type line makes it a creature so it can
	// be declared as an attacker without a crew activation.
	regalia := effectstest.B12Push(g, me.ID, "The Regalia", "Legendary Artifact Creature — Vehicle", effectstest.PlRegaliaOracle, 4, 4)
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	a := effectstest.PlTop(me, "A", "Instant", "")
	b := effectstest.PlTop(me, "B", "Sorcery", "")

	effectstest.DeclareAttack(t, g, opp.ID, regalia)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || !effectstest.TappedOnBattlefield(t, g, land) {
		t.Fatal("the land revealed enters tapped")
	}
	bottom := effectstest.PlBottomIDs(me, 2)
	if !bottom[a] || !bottom[b] {
		t.Error("the cards revealed above the land go to the bottom")
	}
	for _, id := range []uuid.UUID{a, b} {
		c, _ := g.LookupCardForEffect(id)
		if c.IsKnownTo(opp.ID) {
			t.Error("a card on the bottom in a random order is still known")
		}
	}
}

// --- Lurking Predators -------------------------------------------------

func TestLurkingPredatorsPutsACreatureOrOffersTheBottom(t *testing.T) {
	t.Run("a creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Lurking Predators", "Enchantment", effectstest.PlLurkingOracle, 0, 0)
		bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
		effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bear) {
			t.Error("the revealed creature did not enter")
		}
	})
	t.Run("a noncreature, put on the bottom", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Lurking Predators", "Enchantment", effectstest.PlLurkingOracle, 0, 0)
		bolt := effectstest.PlTop(me, "Rite", "Sorcery", "")
		effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		for i := 0; i < 8 && effectstest.ConfirmChoiceFor(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		c := effectstest.ConfirmChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("the bottom was not offered")
		}
		if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		if me.Library.Cards[0].InstanceID != bolt {
			t.Error("accepted: the card goes to the bottom")
		}
	})
}

// --- Atla Palani -------------------------------------------------------

func TestAtlaPalaniMakesEggsAndHatchesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	atla := effectstest.B12Push(g, me.ID, "Atla Palani, Nest Tender", "Legendary Creature — Human Shaman", effectstest.PlAtlaPalaniOracle, 2, 3)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, atla, 0, game.ActivateAbilityParams{})
	var egg uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Egg" && c.Controller == me.ID {
			egg = c.InstanceID
			if !game.HasKeyword(&c, "defender") {
				t.Error("the Egg has defender")
			}
		}
	}
	if egg == uuid.Nil {
		t.Fatal("no Egg token was created")
	}
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	skipped := effectstest.PlTop(me, "Rite", "Sorcery", "")

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(egg); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Egg died and nothing hatched")
	}
	if !effectstest.PlBottomIDs(me, 1)[skipped] {
		t.Error("the card revealed above the creature goes to the bottom")
	}
}

// --- Ureni -------------------------------------------------------------

func TestUreniOffersOnlyDragonsAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	other := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	dragon := effectstest.PlTop(me, "Shivan Dragon", "Creature — Dragon", "{4}{R}{R}")
	effectstest.CastCatalogSpell(t, g, "Ureni of the Unwritten", "Legendary Creature — Spirit Dragon", effectstest.PlUreniOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Ureni offered nothing")
	}
	if effectstest.HasID(pick.ChooseCards, other) || !effectstest.HasID(pick.ChooseCards, dragon) {
		t.Errorf("offered %v: only the Dragon creature card qualifies", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	c, _ := g.LookupCardForEffect(dragon)
	if c.IsKnownTo(g.Seats[1].ID) {
		t.Error("a LOOK made an opponent a knower")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dragon)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon did not enter")
	}
	if !effectstest.PlBottomIDs(me, 7)[other] {
		t.Error("the rest go to the bottom")
	}
}

// --- Majestic Genesis --------------------------------------------------

func TestMajesticGenesisRevealsByCommanderManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Command.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Commander", TypeLine: "Legendary Creature — Elf",
		ManaCost: "{1}{G}", IsCommander: true, Owner: me.ID, Controller: me.ID,
	})
	below := effectstest.PlTop(me, "Below", "Creature — Bear", "{1}{G}")
	bolt := effectstest.PlTop(me, "Rite", "Sorcery", "")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")

	effectstest.B12PlayFromHand(t, g, "Majestic Genesis", "Sorcery", effectstest.PlMajesticOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("nothing offered")
	}
	if effectstest.HasID(pick.ChooseCards, below) {
		t.Error("X is the commander's mana value, 2: the third card was never revealed")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the chosen permanent did not enter")
	}
	if !effectstest.PlBottomIDs(me, 1)[bolt] {
		t.Error("the rest go to the bottom")
	}
	if !me.Library.Contains(below) {
		t.Error("an unrevealed card moved")
	}
}

// --- Gilgamesh ---------------------------------------------------------

func TestGilgameshPutsAnyNumberOfEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	sword := effectstest.PlTop(me, "Sword", "Artifact — Equipment", "{3}")
	axe := effectstest.PlTop(me, "Axe", "Artifact — Equipment", "{2}")
	effectstest.CastCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", effectstest.PlGilgameshOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 2 || effectstest.HasID(pick.ChooseCards, bear) {
		t.Fatalf("want a 0..2 pick over the two Equipment, got %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sword, axe)
	if !g.Battlefield.Contains(sword) || !g.Battlefield.Contains(axe) {
		t.Fatal("both Equipment enter")
	}
	if !effectstest.PlBottomIDs(me, 4)[bear] {
		t.Error("the rest go to the bottom")
	}
}

// --- Armored Skyhunter -------------------------------------------------

func TestArmoredSkyhunterPutsAnEquipmentAndOffersTheAttach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sky := effectstest.B12Push(g, me.ID, "Armored Skyhunter", "Creature — Cat Knight", effectstest.PlSkyhunterOracle, 3, 3)
	aura := effectstest.PlTop(me, "Aura", "Enchantment — Aura", "{1}{W}")
	sword := effectstest.PlTop(me, "Sword", "Artifact — Equipment", "{3}")

	effectstest.DeclareAttack(t, g, opp.ID, sky)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no Equipment offered")
	}
	// #3158: an Aura card is offered too (aura_host_cards_test.go).
	if !effectstest.HasID(pick.ChooseCards, aura) || !effectstest.HasID(pick.ChooseCards, sword) {
		t.Error("an Aura or Equipment card among the six is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	if !g.Battlefield.Contains(sword) {
		t.Fatal("the Equipment did not enter")
	}
	attach := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if attach == nil || !effectstest.HasID(attach.ChooseCards, sky) {
		t.Fatal("the attach was not offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sky)
	c, _ := g.LookupCardForEffect(sword)
	if !c.IsAttachedTo(sky) {
		t.Error("the Equipment is not attached to the chosen creature")
	}
	if !me.Library.Contains(aura) {
		t.Error("the Aura goes to the bottom of the library")
	}
}

// --- The Prismatic Bridge ----------------------------------------------

func TestThePrismaticBridgeRevealsUntilACreatureOrPlaneswalkerOnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Prismatic Bridge", TypeLine: "Legendary Enchantment",
		OracleID: effects.EsikaGodOfTheTreeOracleID, ActiveFace: 1, Owner: me.ID, Controller: me.ID,
	})
	walker := effectstest.PlTop(me, "Walker", "Legendary Planeswalker — Test", "{3}")
	// With loyalty, or the state-based check (CR 704.5i) bins it on arrival.
	me.Library.Cards[len(me.Library.Cards)-1].StartingLoyalty = 3
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the Bridge triggered on an opponent's upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walker) {
		t.Fatal("the planeswalker card did not enter")
	}
	if !effectstest.PlBottomIDs(me, 1)[land] {
		t.Error("the land revealed above it goes to the bottom")
	}
}

// --- the primitive's error paths ---------------------------------------

// Then runs when a card of the batch cannot enter after all: whatever
// did enter is reported, and "the rest" still leave the top of the
// library.
//
// The Bear's replacement moves the Bear out of the library while its
// entry is evaluated. Until #1322 the batch's move then failed with
// ErrCardNotFound; since the batch can wait on a question, a card that
// left its zone before the landing is an ordinary outcome — milled or
// drawn while a shockland's question was open — so it is simply not
// part of the entry, the way the single-card finisher has always
// treated it. The Forest's entry is unaffected.
func TestPutFromLibraryThenRunsWhenABatchCardLeavesMidEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	floor := effectstest.PlTop(me, "Floor", "Instant", "{U}")
	rite := effectstest.PlTop(me, "Rite", "Sorcery", "{G}")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{1}{G}")
	forest := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	size := me.Library.Size()

	batchLeavesRan, batchLeavesRes = false, effects.PutFromLibraryResult{}
	var err error
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == bear && ev.NewZone == game.ZoneBattlefield
			},
			Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error {
				_, err := game.MoveCard(me.Library, me.Graveyard, bear)
				return err
			},
			Label: "Test: the Bear leaves the library mid-entry",
		})
		err = effects.PutFromLibraryOntoBattlefield{
			Player:   me.ID,
			Cards:    []uuid.UUID{forest, bear, rite},
			All:      true,
			ThenTail: batchLeavesThen,
		}.Apply(effects.NewContext(g, nil))
	})
	ran, res := batchLeavesRan, batchLeavesRes

	if err != nil {
		t.Errorf("err = %v; a card that left its zone is not part of the entry, not a failure", err)
	}
	if !ran {
		t.Fatal("Then did not run after a card of the batch could not enter")
	}
	if len(res.Entered) != 1 || res.Entered[0] != forest || !g.Battlefield.Contains(forest) {
		t.Errorf("entered %v, want only the Forest", res.Entered)
	}
	if len(res.Rest) != 1 || res.Rest[0] != rite {
		t.Errorf("rest %v, want [Rite]", res.Rest)
	}
	if !effectstest.PlBottomIDs(me, 1)[rite] {
		t.Error("the rest did not go to the bottom after the batch errored")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear, no longer in the library, was pulled back out of the graveyard")
	}
	if me.Library.Size() != size-2 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != floor {
		t.Error("the card under the revealed ones should now be the top card")
	}
}

// One card that cannot be binned does not keep the rest out of the
// graveyard; the first error is still reported.
func TestPutRestIntoGraveyardCarriesOnPastAnError(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PlTop(me, "A", "Instant", "{U}")
	b := effectstest.PlTop(me, "B", "Sorcery", "{G}")
	var err error
	g.WithWriteLock(func() {
		err = effects.PutRestIntoGraveyard(g, effects.PutFromLibraryResult{
			Player: me.ID,
			Rest:   []uuid.UUID{a, uuid.New(), b},
		})
	})
	if !errors.Is(err, game.ErrCardNotFound) {
		t.Errorf("err = %v, want ErrCardNotFound for the missing card", err)
	}
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("a card after the failing one stayed in the library")
	}
}

// "Put that card on the bottom" is about the card on top of the
// library. If it has left the library before the answer arrives, a yes
// does not pull it back out of the hand it went to.
func TestLurkingPredatorsLeavesACardThatLeftTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lurking Predators", "Enchantment", effectstest.PlLurkingOracle, 0, 0)
	rite := effectstest.PlTop(me, "Rite", "Sorcery", "")
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	for i := 0; i < 8 && effectstest.ConfirmChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	c := effectstest.ConfirmChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the bottom was not offered")
	}
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(me.Library, me.Hand, rite); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(rite) || me.Library.Contains(rite) {
		t.Error("a card that left the library was pulled back to its bottom")
	}
}

// batchLeavesRan and batchLeavesRes are what the probe rider below saw: a
// registered tail holds no closure, so the test reads them off package
// variables (the tests of this package do not run in parallel).
var (
	batchLeavesRan bool
	batchLeavesRes effects.PutFromLibraryResult
)

// batchLeavesThen records what it was told, then puts the rest on the
// bottom of the library.
var batchLeavesThen = effects.PutFromLibraryTail("test/put-from-library-batch-leaves-then", func(ctx *effects.Context, r effects.PutFromLibraryResult) error {
	batchLeavesRan, batchLeavesRes = true, r
	return effects.PutRestOnBottomInRandomOrder(ctx.Game, r)
})
