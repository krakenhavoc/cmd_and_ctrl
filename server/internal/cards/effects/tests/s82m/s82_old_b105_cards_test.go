package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b105_cards_test.go — Core Set 2021, batch 105 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	m21B105SanctumCalmWaters = "1cd1e3a8-e09d-4b61-8244-92b317132d6a"
	m21B105LilianasScorn     = "2359db1c-6029-4748-9f09-d57bb0ae92e9"
	m21B105SanctumOfAll      = "63f60dfc-2bc4-411f-ae53-504156fe0102"
	m21B105HistorianZhalfir  = "87e9065d-814a-47fa-9e92-19fb6fd6d78c"
	m21B105AlpineHoundmaster = "88939cb3-5f61-4857-95cb-9446c6c1f1aa"
	m21B105TrackDown         = "8f28b65b-4b73-4080-9cdc-d236ad99f39c"
	m21B105Niambi            = "91c9dbd8-a8d8-4a4e-8408-3ad92e974d11"
	m21B105PortcullisVine    = "a217f386-11b5-48f0-b173-3b0700ba295f"
	m21B105FranticInventory  = "a7865331-4c53-46db-b2d7-7e5cb21b6b9d"
	m21B105GnarledSage       = "c36d7c2b-204f-432e-b8e2-ffb3ffb49231"
	m21B105GarruksWarsteed   = "d4911e00-eed1-4c8c-8779-8a3f9406ead7"
	m21B105CryptLurker       = "f0cbf361-d7dc-4556-a654-71c08785095a"

	m21B105Bear = "Creature — Bear"
)

func m21B105Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestM21B105EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m21B105SanctumCalmWaters: "Sanctum of Calm Waters",
		m21B105LilianasScorn:     "Liliana's Scorn",
		m21B105SanctumOfAll:      "Sanctum of All",
		m21B105HistorianZhalfir:  "Historian of Zhalfir",
		m21B105AlpineHoundmaster: "Alpine Houndmaster",
		m21B105TrackDown:         "Track Down",
		m21B105Niambi:            "Niambi, Esteemed Speaker",
		m21B105PortcullisVine:    "Portcullis Vine",
		m21B105FranticInventory:  "Frantic Inventory",
		m21B105GnarledSage:       "Gnarled Sage",
		m21B105GarruksWarsteed:   "Garruk's Warsteed",
		m21B105CryptLurker:       "Crypt Lurker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestM21B105PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Garruk's Warsteed", "Creature — Rhino", m21B105GarruksWarsteed, 3, 5, []string{"vigilance"}},
		{"Niambi, Esteemed Speaker", "Legendary Creature — Human Cleric", m21B105Niambi, 2, 1, []string{"flash"}},
		{"Portcullis Vine", "Creature — Plant Wall", m21B105PortcullisVine, 0, 3, []string{"defender"}},
		{"Gnarled Sage", "Creature — Treefolk Druid", m21B105GnarledSage, 4, 4, []string{"reach"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

// --- the fetch family -----------------------------------------------------------

func TestM21B105LilianasScornDestroysThenMayFetchLiliana(t *testing.T) {
	for _, zone := range []string{"library", "graveyard", "declined"} {
		t.Run(zone, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := m21B105Seats(g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B105Bear, 2, 2)
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			var lili uuid.UUID
			switch zone {
			case "graveyard":
				lili = effectstest.PushGraveyardCardWithTypeLineID(me, "Liliana, Death Mage", "Legendary Planeswalker — Liliana")
			default:
				lili = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Liliana, Death Mage", TypeLine: "Legendary Planeswalker — Liliana"})
			}
			effectstest.CastCatalogSpell(t, g, "Liliana's Scorn", "Sorcery", m21B105LilianasScorn, effectstest.B16TargetCard(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Fatal("the creature was not destroyed")
			}
			if zone == "declined" {
				effectstest.AnswerMayChoice(t, g, me.ID, false)
				effectstest.PassPriorityAroundTable(t, g)
				if me.Hand.Contains(lili) {
					t.Error("declining still took Liliana")
				}
				return
			}
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			if zone == "graveyard" {
				effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			}
			effectstest.AnswerSearchByID(t, g, me.ID, lili)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(lili) {
				t.Errorf("Liliana did not come from the %s", zone)
			}
		})
	}
}

func TestM21B105LilianasScornFizzlesWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B105Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B105Bear, 2, 2)
	lili := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Liliana, Death Mage", TypeLine: "Legendary Planeswalker — Liliana"})
	effectstest.CastCatalogSpell(t, g, "Liliana's Scorn", "Sorcery", m21B105LilianasScorn, effectstest.B16TargetCard(victim))
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || me.Hand.Contains(lili) {
		t.Error("a spell whose only target is gone searched anyway (CR 608.2b)")
	}
}

func TestM21B105GarruksWarsteedFetchesGarruk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B105Seats(g)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	garruk := effectstest.PushGraveyardCardWithTypeLineID(me, "Garruk, Savage Herald", "Legendary Planeswalker — Garruk")
	effectstest.CastCatalogSpell(t, g, "Garruk's Warsteed", "Creature — Rhino", m21B105GarruksWarsteed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, garruk)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(garruk) {
		t.Error("Garruk was not fetched from the graveyard")
	}
}

func TestM21B105AlpineHoundmasterFetchesOneOfEachName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B105Seats(g)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	dog1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Alpine Watchdog", TypeLine: "Creature — Dog"})
	dog2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Alpine Watchdog", TypeLine: "Creature — Dog"})
	cur := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Igneous Cur", TypeLine: "Creature — Elemental Dog"})
	effectstest.CastCatalogSpell(t, g, "Alpine Houndmaster", "Creature — Human Warrior", m21B105AlpineHoundmaster, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 3 {
		t.Fatalf("offered %d cards, want the three named ones", len(c.SearchCards))
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{dog1, dog2}); err == nil {
		t.Fatal("two cards of one name were accepted")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dog1, cur)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dog1) || !me.Hand.Contains(cur) || me.Hand.Contains(dog2) {
		t.Error("the Watchdog and the Cur should be in hand, the second Watchdog not")
	}
	_ = opp
}

func TestM21B105AlpineHoundmasterGrowsWithTheOtherAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B105Seats(g)
	hound := effectstest.B12Push(g, me.ID, "Alpine Houndmaster", "Creature — Human Warrior", m21B105AlpineHoundmaster, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "A", m21B105Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", m21B105Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hound, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, hound); p != 4 || tough != 2 {
		t.Errorf("the Houndmaster is %d/%d with two other attackers, want 4/2", p, tough)
	}
}

// --- Track Down -----------------------------------------------------------------

func TestM21B105TrackDownDrawsOnACreatureOrLandOnTop(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		draws          bool
	}{
		{"Bear", m21B105Bear, true},
		{"Forest", "Basic Land — Forest", true},
		{"Bolt", "Instant", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := m21B105Seats(g)
			effectstest.SeedLibrary(me, "Two", "Three", "Four", "Five")
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: c.name, TypeLine: c.typeLine, Owner: me.ID, Controller: me.ID})
			effectstest.CastCatalogSpell(t, g, "Track Down", "Sorcery", m21B105TrackDown, nil)
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if ch := effectstest.ScryChoiceFor(g, me.ID); ch == nil || len(ch.ScryCards) != 3 {
				t.Fatalf("scry 3 prompt: %+v", ch)
			}
			effectstest.AnswerScryKeepAll(t, g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
			want := hand
			if c.draws {
				want = hand + 1
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
		})
	}
}

// --- Historian, Frantic Inventory, Gnarled Sage ---------------------------------

func TestM21B105HistorianDrawsOnlyWithATeferi(t *testing.T) {
	for _, teferi := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := m21B105Seats(g)
		h := effectstest.B12Push(g, me.ID, "Historian of Zhalfir", "Creature — Human Wizard", m21B105HistorianZhalfir, 3, 3)
		if teferi {
			pw := effectstest.PushCatalogPermanent(g, me.ID, "Teferi, Timeless Voyager", "Legendary Planeswalker — Teferi", "", false)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(pw, game.CounterLoyalty, 4) })
		} else {
			pw := effectstest.PushCatalogPermanent(g, me.ID, "Not a Teferi", "Legendary Planeswalker — Jace", "", false)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(pw, game.CounterLoyalty, 4) })
		}
		hand := me.Hand.Size()
		effectstest.DeclareAttack(t, g, opp.ID, h)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if teferi {
			want = hand + 1
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("a Teferi %v: hand %d, want %d", teferi, got, want)
		}
	}
}

func TestM21B105FranticInventoryDrawsOnePlusTheCopiesInTheGraveyard(t *testing.T) {
	for _, copies := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B105Seats(g)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
		for i := 0; i < copies; i++ {
			effectstest.PushGraveyardCardWithTypeLineID(me, "Frantic Inventory", "Instant")
		}
		effectstest.PushGraveyardCardWithTypeLineID(me, "Other Instant", "Instant")
		effectstest.CastCatalogSpell(t, g, "Frantic Inventory", "Instant", m21B105FranticInventory, nil)
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != 1+copies {
			t.Errorf("%d copies in the graveyard: drew %d, want %d", copies, got, 1+copies)
		}
	}
}

func TestM21B105GnarledSageGrowsAfterTwoDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B105Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	sage := effectstest.B12Push(g, me.ID, "Gnarled Sage", "Creature — Treefolk Druid", m21B105GnarledSage, 4, 4)
	effectstest.AdvanceToMain(t, g)
	base := g.TurnTallyFor(me.ID).CardsDrawn
	for drawn := base; drawn < 2; drawn++ {
		if p, tough := effectstest.PtOf(t, g, sage); p != 4 || tough != 4 || effectstest.HasEffectiveKeyword(t, g, sage, "vigilance") {
			t.Fatalf("after %d draws the Sage is %d/%d, want a plain 4/4", drawn, p, tough)
		}
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatal(err)
		}
	}
	if p, tough := effectstest.PtOf(t, g, sage); p != 4 || tough != 6 {
		t.Errorf("after two draws the Sage is %d/%d, want 4/6", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sage, "vigilance") {
		t.Error("no vigilance after two draws")
	}
}

// --- Portcullis Vine, Niambi ----------------------------------------------------

func TestM21B105PortcullisVineSacrificesOnlyADefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := m21B105Seats(g)
	vine := effectstest.B12Push(g, me.ID, "Portcullis Vine", "Creature — Plant Wall", m21B105PortcullisVine, 0, 3)
	wall := effectstest.B12Push(g, me.ID, "Wall", "Creature — Wall", "", 0, 4)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, wall).Keywords = []string{"defender"} })
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B105Bear, 2, 2)
	effectstest.ApaMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, vine, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a creature without defender")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "G", "G")
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, vine, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wall}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestM21B105NiambiBouncesAndGainsLifeEqualToItsManaValue(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B105Seats(g)
		bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Costly Bear", TypeLine: m21B105Bear,
			ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Niambi, Esteemed Speaker", "Legendary Creature — Human Cleric", m21B105Niambi, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
				t.Error("the Bear did not return to hand")
			}
			if me.Life != life+3 {
				t.Errorf("life %d, want %d", me.Life, life+3)
			}
		} else if !g.Battlefield.Contains(bear) || me.Life != life {
			t.Error("declining still bounced or gained life")
		}
	}
}

func TestM21B105NiambiDiscardsALegendaryCardToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := m21B105Seats(g)
	niambi := effectstest.B12Push(g, me.ID, "Niambi, Esteemed Speaker", "Legendary Creature — Human Cleric", m21B105Niambi, 2, 1)
	effectstest.SeedLibrary(me, "A", "B", "C")
	plain := effectstest.EoHand(me, "Plain", "Creature — Bear")
	legend := effectstest.EoHand(me, "Legend", "Legendary Creature — Human")
	effectstest.ApaMana(me, "G", "W", "U")
	if err := g.ActivateCatalogAbility(me.ID, niambi, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{plain}}); err == nil {
		t.Fatal("a nonlegendary card paid the discard")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "G", "W", "U")
	hand := me.Hand.Size()
	effectstest.ApaActivate(t, g, me, niambi, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{legend}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(legend) {
		t.Error("the legendary card was not discarded")
	}
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Crypt Lurker ---------------------------------------------------------------

func TestM21B105CryptLurkerSacrificesOrDiscardsACreatureToDraw(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option int
	}{{"sacrifice", 0}, {"discard", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.EoTable(t)
			effectstest.SeedLibrary(me, "A", "B")
			land := effectstest.EoHand(me, "Forest", "Basic Land — Forest")
			zombie := effectstest.EoHand(me, "Zombie", "Creature — Zombie")
			bear := effectstest.B12Creature(g, me.ID, "Bear", m21B105Bear, 2, 2)
			effectstest.EnterFromHand(t, g, me.ID, "Crypt Lurker", "Creature — Horror", m21B105CryptLurker)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.option == 1 {
				if err := g.ResolvePayUnlessWithOption(effectstest.EoPrompt(g, me.ID).ID, me.ID, true, &tc.option, []uuid.UUID{land}); err == nil {
					t.Fatal("a land paid a creature-card discard")
				}
			}
			hand := me.Hand.Size()
			if tc.option == 0 {
				effectstest.EoPay(t, g, me.ID, 0, bear)
			} else {
				effectstest.EoPay(t, g, me.ID, 1, zombie)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := hand + 1
			if tc.option == 1 {
				want = hand - 1 + 1
			}
			if got := me.Hand.Size(); got != want {
				t.Errorf("hand %d, want %d", got, want)
			}
			if tc.option == 0 && g.Battlefield.Contains(bear) {
				t.Error("the Bear was not sacrificed")
			}
			if tc.option == 1 && me.Hand.Contains(zombie) {
				t.Error("the Zombie was not discarded")
			}
		})
	}
}

func TestM21B105CryptLurkerAsksNothingOfALurkerAlone(t *testing.T) {
	// The Lurker itself is a creature it may sacrifice, so it is always asked.
	g, me, _ := effectstest.EoTable(t)
	effectstest.SeedLibrary(me, "A", "B")
	lurker := effectstest.EnterFromHand(t, g, me.ID, "Crypt Lurker", "Creature — Horror", m21B105CryptLurker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EoOfferedOptions(t, g, me.ID); !got[0] || got[1] {
		t.Fatalf("offered %v, want the sacrifice only", got)
	}
	effectstest.EoPay(t, g, me.ID, 0, lurker)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lurker) {
		t.Error("the Lurker was not sacrificed")
	}
}

// --- Shrines --------------------------------------------------------------------

func m21B105AtFirstMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestM21B105SanctumOfCalmWatersDrawsXThenDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B105Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of Calm Waters", "Legendary Enchantment — Shrine", m21B105SanctumCalmWaters)
	effectstest.B12Permanent(g, me.ID, "Other Shrine", "Enchantment — Shrine")
	effectstest.B12Permanent(g, g.Seats[1].ID, "Their Shrine", "Enchantment — Shrine")
	m21B105AtFirstMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("drew %d cards, want 2 (two Shrines you control)", got-hand)
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 1 {
		t.Errorf("owes %d discards, want 1", n)
	}
}

func TestM21B105SanctumOfCalmWatersMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B105Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of Calm Waters", "Legendary Enchantment — Shrine", m21B105SanctumCalmWaters)
	m21B105AtFirstMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if me.Hand.Size() != hand || effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Error("declining still drew or discarded")
	}
}

func TestM21B105SanctumOfAllFetchesAShrineFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B105Seats(g)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	shrine := effectstest.PushGraveyardCardWithTypeLineID(me, "Sanctum of Test", "Legendary Enchantment — Shrine")
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Library Shrine", TypeLine: "Enchantment — Shrine"})
	effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of All", "Legendary Enchantment — Shrine", m21B105SanctumOfAll)
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil && c.SearchIsUndecided() {
		if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, shrine)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(shrine) || me.Graveyard.Contains(shrine) {
		t.Error("the Shrine did not reach the battlefield")
	}
}

func TestM21B105SanctumOfAllDoublesOtherShrinesTriggersAtSixShrines(t *testing.T) {
	for _, tc := range []struct {
		others int // Shrines besides Sanctum of All and Sanctum of Calm Waters
		want   int
	}{{3, 1}, {4, 2}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B105Seats(g)
		effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L")
		effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of All", "Legendary Enchantment — Shrine", m21B105SanctumOfAll)
		effectstest.SeedPermanentWithOracle(g, me.ID, "Sanctum of Calm Waters", "Legendary Enchantment — Shrine", m21B105SanctumCalmWaters)
		for i := 0; i < tc.others; i++ {
			effectstest.B12Permanent(g, me.ID, "Plain Shrine", "Enchantment — Shrine")
		}
		m21B105AtFirstMain(t, g)
		asked := 0
		for i := 0; i < 4; i++ {
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LatestConfirmFor(g, me.ID) == nil {
				break
			}
			asked++
			effectstest.AnswerMayChoice(t, g, me.ID, false)
		}
		if asked != tc.want {
			t.Errorf("%d Shrines: Sanctum of Calm Waters triggered %d times, want %d", tc.others+2, asked, tc.want)
		}
	}
}
