package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// put_from_library_test.go — the catalog half of #745: the
// PutFromLibraryOntoBattlefield primitive and the cards that print
// "put [it / a card from among them] onto the battlefield" off the top
// of a library.
//
// The engine half (the move, the batch, the random-order bottom) is
// pinned in server/internal/game/put_from_library_test.go. These tests
// are about each card: what is offered, what enters and how, and where
// the rest go.

const (
	plChaosWarpOracle      = "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8"
	plCoilingOracleOracle  = "69fd4ddf-9ed8-4c56-bef3-9944daf05e4f"
	plGenesisWaveOracle    = "e2487868-f386-438e-a73f-b494f6d35fac"
	plRisenReefOracle      = "2ae71e86-4400-4a30-9077-4d57a43e7395"
	plExplorersScopeOracle = "a563ede9-b92f-4285-88f8-abcbdd017742"
	plThrasiosOracle       = "3d867016-2601-4a37-a73d-308898d3bd37"
	plAnimistsOracle       = "6f1bfe50-b61a-4fdd-963c-40d59117bdf4"
	plRegaliaOracle        = "27a7610e-acbe-4a2b-9f61-81b383eb20a5"
	plLurkingOracle        = "15fbb7b1-c62d-4f82-9f35-2c10299779f4"
	plAtlaPalaniOracle     = "b56cebe0-3752-4ce5-afbd-911543784015"
	plUreniOracle          = "99c2d3ef-e5b4-48cd-b3f5-de9b02c7c36a"
	plMajesticOracle       = "15039c85-31e2-4a2b-82f8-2f8270ff9a00"
	plGilgameshOracle      = "c9432e87-38f6-4889-8f07-ae7cd045e532"
	plSkyhunterOracle      = "b4dbbf56-d7df-4183-bb48-cfc6b4d0468f"
)

// plTop pushes a card onto the TOP of a player's library.
func plTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// plBottomIDs returns the bottom n library card IDs as a set.
func plBottomIDs(p *game.Player, n int) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for _, c := range p.Library.Cards[:n] {
		out[c.InstanceID] = true
	}
	return out
}

func plSearchFired(g *game.Game, from int) bool {
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary {
			return true
		}
	}
	return false
}

func TestLibraryToBattlefieldCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness Completeness
	}{
		plChaosWarpOracle:                {"Chaos Warp", CompletenessCaveats},
		plCoilingOracleOracle:            {"Coiling Oracle", CompletenessFull},
		plGenesisWaveOracle:              {"Genesis Wave", CompletenessFull},
		plRisenReefOracle:                {"Risen Reef", CompletenessFull},
		plExplorersScopeOracle:           {"Explorer's Scope", CompletenessFull},
		plThrasiosOracle:                 {"Thrasios, Triton Hero", CompletenessFull},
		plAnimistsOracle:                 {"Animist's Awakening", CompletenessFull},
		plRegaliaOracle:                  {"The Regalia", CompletenessFull},
		plLurkingOracle:                  {"Lurking Predators", CompletenessFull},
		plAtlaPalaniOracle:               {"Atla Palani, Nest Tender", CompletenessFull},
		plUreniOracle:                    {"Ureni of the Unwritten", CompletenessFull},
		plMajesticOracle:                 {"Majestic Genesis", CompletenessFull},
		plGilgameshOracle:                {"Gilgamesh, Master-at-Arms", CompletenessCaveats},
		plSkyhunterOracle:                {"Armored Skyhunter", CompletenessCaveats},
		esikaGodOfTheTreeOracleID:        {"Esika, God of the Tree", CompletenessCaveats},
		esikaGodOfTheTreeOracleID + "#1": {"The Prismatic Bridge", CompletenessFull},
	} {
		spec, ok := Lookup(oracle)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	before := len(g.Events)

	castCatalogSpell(t, g, "Coiling Oracle", "Creature — Snake Elf Druid", plCoilingOracleOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(land) {
		t.Fatal("the revealed land did not enter")
	}
	if tappedOnBattlefield(t, g, land) {
		t.Error("nothing said tapped")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("land drops used = %d: putting a land is not playing one (CR 305.4)", n)
	}
	if plSearchFired(g, before) {
		t.Error("a reveal is not a search")
	}
	if len(g.PendingChoices) != 0 {
		t.Error("there is no choice on Coiling Oracle")
	}
}

func TestCoilingOraclePutsANonlandIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	castCatalogSpell(t, g, "Coiling Oracle", "Creature — Snake Elf Druid", plCoilingOracleOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the revealed nonland card did not go to hand")
	}
}

// --- Chaos Warp --------------------------------------------------------

// The owner shuffles the permanent in, then reveals and puts a
// permanent card onto the battlefield under THEIR control.
func TestChaosWarpReplacesThePermanentForItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	// A library of lands only, so whatever comes up is a permanent.
	opp.Library.Cards = nil
	for i := 0; i < 4; i++ {
		plTop(opp, "Forest", "Basic Land — Forest", "")
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	onBattlefield := g.Battlefield.Size()
	before := len(g.Events)

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	opp.Library.Cards = nil
	goblin := b12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	onBattlefield := g.Battlefield.Size()
	before := len(g.Events)

	castCatalogSpell(t, g, "Chaos Warp", "Instant", plChaosWarpOracle, b16TargetCard(goblin))
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	token := plTop(me, "Goblin", "Token Creature — Goblin", "")
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")

	b12PlayFromHand(t, g, "Genesis Wave", "Sorcery", plGenesisWaveOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Genesis Wave asked nothing")
	}
	if hasID(pick.ChooseCards, token) {
		t.Error("a token was offered as a permanent card")
	}
	answerChooseCards(t, g, me.ID, bear)
	if g.Battlefield.Contains(token) || me.Graveyard.Contains(token) {
		t.Error("the token changed zones")
	}
}

// The reveal-until family does not stop on a token: it is not "a land
// card", and "put that card onto the battlefield" could not bring it
// back (CR 111.8).
func TestTheRegaliaDoesNotStopOnAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	regalia := b12Push(g, me.ID, "The Regalia", "Legendary Artifact Creature — Vehicle", plRegaliaOracle, 4, 4)
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	token := plTop(me, "Dryad Arbor Copy", "Token Land Creature — Forest Dryad", "")

	declareAttack(t, g, opp.ID, regalia)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Fatal("the reveal stopped on a token and put it onto the battlefield")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the reveal should have carried on to the land card under the token")
	}
}

// --- Genesis Wave ------------------------------------------------------

func TestGenesisWavePutsChosenPermanentsAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	notRevealed := plTop(me, "Deep Bolt", "Instant", "{R}")
	big := plTop(me, "Big", "Creature — Giant", "{5}")
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	forest := plTop(me, "Forest", "Basic Land — Forest", "")
	before := len(g.Events)

	b12PlayFromHand(t, g, "Genesis Wave", "Sorcery", plGenesisWaveOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Genesis Wave asked nothing")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2 (any number of the two eligible)", pick.ChooseMin, pick.ChooseMax)
	}
	if hasID(pick.ChooseCards, big) {
		t.Error("a mana value 5 card was offered for X = 3")
	}
	answerChooseCards(t, g, me.ID, forest, bear)

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
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := plTop(me, "Forest", "Basic Land — Forest", "")
		castCatalogSpell(t, g, "Risen Reef", "Creature — Elemental", plRisenReefOracle, nil)
		passPriorityAroundTable(t, g)

		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatal("Risen Reef did not offer the land")
		}
		if accept {
			answerChooseCards(t, g, me.ID, land)
			if !g.Battlefield.Contains(land) || !tappedOnBattlefield(t, g, land) {
				t.Error("accepted: the land enters tapped")
			}
		} else {
			answerChooseCards(t, g, me.ID)
			if !me.Hand.Contains(land) {
				t.Error("declined: \"if you don't put the card onto the battlefield, put it into your hand\"")
			}
		}
	}
}

func TestRisenReefTriggersForAnotherElementalAndHandsANonland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Risen Reef", "Creature — Elemental", plRisenReefOracle, 1, 1)
	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	castCatalogSpell(t, g, "Other Elemental", "Creature — Elemental", "", nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a nonland card has nothing to offer")
	}
	if !me.Hand.Contains(bolt) {
		t.Error("another Elemental entering did not look, or the nonland did not go to hand")
	}
}

// --- Explorer's Scope --------------------------------------------------

func TestExplorersScopePutsALandTappedWhenTheEquippedCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	scope := seedEquipment(g, me.ID, "Explorer's Scope", plExplorersScopeOracle)
	b06AddMana(me, "C")
	equipTo(t, g, me.ID, scope, bear)
	land := plTop(me, "Forest", "Basic Land — Forest", "")

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !tappedOnBattlefield(t, g, land) {
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			thrasios := b12Push(g, me.ID, "Thrasios, Triton Hero", "Legendary Creature — Merfolk Wizard", plThrasiosOracle, 1, 3)
			top := plTop(me, "Top", tc.typeLine, "")
			b06AddMana(me, "C", "C", "C", "C")
			if err := g.ActivateCatalogAbility(me.ID, thrasios, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
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
				if !g.Battlefield.Contains(top) || !tappedOnBattlefield(t, g, top) {
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		if mastery {
			batch01GraveyardCard(me, "Bolt", "Instant")
			batch01GraveyardCard(me, "Rite", "Sorcery")
		}
		island := plTop(me, "Island", "Basic Land — Island", "")
		bolt := plTop(me, "Bolt", "Instant", "{R}")
		forest := plTop(me, "Forest", "Basic Land — Forest", "")

		b12PlayFromHand(t, g, "Animist's Awakening", "Sorcery", plAnimistsOracle, game.CastSpellParams{XValue: 3})
		passPriorityAroundTable(t, g)
		if chooseCardsChoiceFor(g, me.ID) != nil {
			t.Fatal("\"all land cards\" is not a choice")
		}
		for _, id := range []uuid.UUID{island, forest} {
			if !g.Battlefield.Contains(id) {
				t.Fatal("a revealed land did not enter")
			}
			if tappedOnBattlefield(t, g, id) == mastery {
				t.Errorf("mastery %v: tapped %v", mastery, !mastery)
			}
		}
		if !plBottomIDs(me, 1)[bolt] {
			t.Error("the nonland card goes to the bottom")
		}
	}
}

// --- The Regalia -------------------------------------------------------

func TestTheRegaliaRevealsUntilALandOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	// Crewed for the test: the type line makes it a creature so it can
	// be declared as an attacker without a crew activation.
	regalia := b12Push(g, me.ID, "The Regalia", "Legendary Artifact Creature — Vehicle", plRegaliaOracle, 4, 4)
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	a := plTop(me, "A", "Instant", "")
	b := plTop(me, "B", "Sorcery", "")

	declareAttack(t, g, opp.ID, regalia)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || !tappedOnBattlefield(t, g, land) {
		t.Fatal("the land revealed enters tapped")
	}
	bottom := plBottomIDs(me, 2)
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Lurking Predators", "Enchantment", plLurkingOracle, 0, 0)
		bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
		b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(bear) {
			t.Error("the revealed creature did not enter")
		}
	})
	t.Run("a noncreature, put on the bottom", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b12Push(g, me.ID, "Lurking Predators", "Enchantment", plLurkingOracle, 0, 0)
		bolt := plTop(me, "Rite", "Sorcery", "")
		b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		for i := 0; i < 8 && confirmChoiceFor(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		c := confirmChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	atla := b12Push(g, me.ID, "Atla Palani, Nest Tender", "Legendary Creature — Human Shaman", plAtlaPalaniOracle, 2, 3)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, atla, 0, game.ActivateAbilityParams{})
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
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	skipped := plTop(me, "Rite", "Sorcery", "")

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(egg); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Egg died and nothing hatched")
	}
	if !plBottomIDs(me, 1)[skipped] {
		t.Error("the card revealed above the creature goes to the bottom")
	}
}

// --- Ureni -------------------------------------------------------------

func TestUreniOffersOnlyDragonsAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	other := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	dragon := plTop(me, "Shivan Dragon", "Creature — Dragon", "{4}{R}{R}")
	castCatalogSpell(t, g, "Ureni of the Unwritten", "Legendary Creature — Spirit Dragon", plUreniOracle, nil)
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Ureni offered nothing")
	}
	if hasID(pick.ChooseCards, other) || !hasID(pick.ChooseCards, dragon) {
		t.Errorf("offered %v: only the Dragon creature card qualifies", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	c, _ := g.LookupCardForEffect(dragon)
	if c.IsKnownTo(g.Seats[1].ID) {
		t.Error("a LOOK made an opponent a knower")
	}
	answerChooseCards(t, g, me.ID, dragon)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon did not enter")
	}
	if !plBottomIDs(me, 7)[other] {
		t.Error("the rest go to the bottom")
	}
}

// --- Majestic Genesis --------------------------------------------------

func TestMajesticGenesisRevealsByCommanderManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Command.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Commander", TypeLine: "Legendary Creature — Elf",
		ManaCost: "{1}{G}", IsCommander: true, Owner: me.ID, Controller: me.ID,
	})
	below := plTop(me, "Below", "Creature — Bear", "{1}{G}")
	bolt := plTop(me, "Rite", "Sorcery", "")
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")

	b12PlayFromHand(t, g, "Majestic Genesis", "Sorcery", plMajesticOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("nothing offered")
	}
	if hasID(pick.ChooseCards, below) {
		t.Error("X is the commander's mana value, 2: the third card was never revealed")
	}
	answerChooseCards(t, g, me.ID, bear)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the chosen permanent did not enter")
	}
	if !plBottomIDs(me, 1)[bolt] {
		t.Error("the rest go to the bottom")
	}
	if !me.Library.Contains(below) {
		t.Error("an unrevealed card moved")
	}
}

// --- Gilgamesh ---------------------------------------------------------

func TestGilgameshPutsAnyNumberOfEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	sword := plTop(me, "Sword", "Artifact — Equipment", "{3}")
	axe := plTop(me, "Axe", "Artifact — Equipment", "{2}")
	castCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", plGilgameshOracle, nil)
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 2 || hasID(pick.ChooseCards, bear) {
		t.Fatalf("want a 0..2 pick over the two Equipment, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, sword, axe)
	if !g.Battlefield.Contains(sword) || !g.Battlefield.Contains(axe) {
		t.Fatal("both Equipment enter")
	}
	if !plBottomIDs(me, 4)[bear] {
		t.Error("the rest go to the bottom")
	}
}

// --- Armored Skyhunter -------------------------------------------------

func TestArmoredSkyhunterPutsAnEquipmentAndOffersTheAttach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sky := b12Push(g, me.ID, "Armored Skyhunter", "Creature — Cat Knight", plSkyhunterOracle, 3, 3)
	aura := plTop(me, "Aura", "Enchantment — Aura", "{1}{W}")
	sword := plTop(me, "Sword", "Artifact — Equipment", "{3}")

	declareAttack(t, g, opp.ID, sky)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no Equipment offered")
	}
	if hasID(pick.ChooseCards, aura) {
		t.Error("Auras are a declared omission and must not be offered")
	}
	answerChooseCards(t, g, me.ID, sword)
	if !g.Battlefield.Contains(sword) {
		t.Fatal("the Equipment did not enter")
	}
	attach := chooseCardsChoiceFor(g, me.ID)
	if attach == nil || !hasID(attach.ChooseCards, sky) {
		t.Fatal("the attach was not offered")
	}
	answerChooseCards(t, g, me.ID, sky)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Prismatic Bridge", TypeLine: "Legendary Enchantment",
		OracleID: esikaGodOfTheTreeOracleID, ActiveFace: 1, Owner: me.ID, Controller: me.ID,
	})
	walker := plTop(me, "Walker", "Legendary Planeswalker — Test", "{3}")
	// With loyalty, or the state-based check (CR 704.5i) bins it on arrival.
	me.Library.Cards[len(me.Library.Cards)-1].StartingLoyalty = 3
	land := plTop(me, "Forest", "Basic Land — Forest", "")

	advanceToUpkeepOf(t, g, 1)
	if g.Battlefield.Contains(walker) {
		t.Fatal("the Bridge triggered on an opponent's upkeep")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walker) {
		t.Fatal("the planeswalker card did not enter")
	}
	if !plBottomIDs(me, 1)[land] {
		t.Error("the land revealed above it goes to the bottom")
	}
}
