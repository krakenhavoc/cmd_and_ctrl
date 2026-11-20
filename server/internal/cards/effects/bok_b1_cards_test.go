package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bok_b1_cards_test.go — Betrayers of Kamigawa, batch 1 (#5602). Each
// test plays the real catalog entry and asserts what its text says.

const (
	bokB1MistbladeOracle      = "16856e22-6f9e-4f8e-b4ee-622a7c2ee453"
	bokB1ThroatSlitterOracle  = "56f782b7-d46c-4aa1-ae16-66d6401c72ae"
	bokB1FalseHopeOracle      = "49984248-4800-4f46-933a-fedd970ae910"
	bokB1HigureOracle         = "01197456-62f1-4bc0-9f2f-127f97f9b05d"
	bokB1SkullsnatcherOracle  = "5326381e-e5e7-4c3e-9565-4837b3ffaec0"
	bokB1TendoOracle          = "9e006a4b-8dde-4416-8cb4-8401562d0fd5"
	bokB1DisruptingOracle     = "2912d93d-ed51-41ef-b64f-3cb6581c09e5"
	bokB1LeechesOracle        = "923b725a-e189-4b61-b01a-c5056f1328ba"
	bokB1OkibaOracle          = "efe3b58f-6f83-45d5-beb1-7dde1cae9dd7"
	bokB1PatronMoonOracle     = "e514546a-ca03-48ba-90e1-7a0e145137f2"
	bokB1LifegiftOracle       = "c99d4be7-244c-4fb5-877a-8aee961a2666"
	bokB1WalkerOracle         = "451a3f52-9810-4137-93b1-9594e7727c29"
	bokB1ShukoOracle          = "8abe0577-8fdb-4e4a-a871-01b21732c961"
	bokB1GodsEyeOracle        = "a66008c9-1ede-4dcf-8d35-6c0ed2390996"
	bokB1PatronOrochiOracle   = "e245f2f6-2b5b-41bc-88b3-5b41ca813c5e"
	bokB1BlazingOracle        = "ccba143d-ed9b-4044-bdad-f010a0e38e47"
	bokB1TakenOracle          = "b3c5b7b6-8002-41f2-87fc-1ed7f9e103ee"
	bokB1TraprootOracle       = "ce7d555a-9197-4241-94ee-a164a65099a5"
	bokB1TerashisGraspOracle  = "d4738552-3a5e-43c5-a975-8b77618bacaf"
	bokB1YukoraOracle         = "10026343-a8cd-4a99-b4d9-ff7611f52052"
	bokB1LoamDwellerOracle    = "a61f7653-4c55-472f-b37c-617b94d0fecd"
	bokB1RibbonsOracle        = "22b76ab7-9658-41a3-b8bb-fc02de0631a4"
	bokB1ToilsOracle          = "a50be01d-821f-4fdc-b12a-4a1a76432a21"
	bokB1PatronNezumiOracle   = "038a8b6b-4529-4e06-bdff-9ee7ba291776"
	bokB1CrackOracle          = "eb107601-f4ff-4504-9e3f-3de63b0d9e6b"
	bokB1TeardropOracle       = "6f47a67e-1fcd-4f3f-81d0-1a6868606390"
	bokB1BlizzardHerderOracle = "16fe1f04-9bbf-43a3-ab14-395fdf4a3a43"
	bokB1WebOfWarOracle       = "9629e6b7-75e1-4447-bf20-4d4a43d6f60a"
	bokB1PatronKitsuneOracle  = "e75356a8-6268-4355-a0bd-114a5243d527"
	bokB1HerosDemiseOracle    = "69ec0b59-a0aa-4878-a3b3-130cea18fee0"
	bokB1SickeningOracle      = "2da8e9d6-e8a8-46a5-a478-3e81098d2a08"
	bokB1StreamOracle         = "9013214f-c066-4a09-871f-e9f8396f38a4"
	bokB1KyokiOracle          = "f5240f9a-597b-43e7-9876-74559df1354e"
	bokB1IndebtedOracle       = "fee392d9-a961-4f41-81ba-ab1b87c931fa"
	bokB1StirOracle           = "296dcc57-5651-4cb1-b89c-fcdd3f5311bd"
	bokB1BileUrchinOracle     = "7d406aa7-636a-45c1-903f-11c2ce2ef3e3"
	bokB1NourishingOracle     = "2aa0e0e5-cb6d-4518-8eff-d29f935486e0"
	bokB1UncheckedOracle      = "ea4052bc-cd37-467c-9fba-3f1ed537b781"
	bokB1IreOracle            = "4870ec65-8067-45fb-b7ec-a6eca87c7e77"
	bokB1FirstVolleyOracle    = "b6a89457-1581-46d4-b75b-c5b8704d9f6f"
	bokB1SilverstormOracle    = "3df0a392-bcc1-4456-bd6e-d044bcc83a16"
	bokB1PatronAkkiOracle     = "38dd86a2-652c-42b7-9b21-2fad059cc706"
	bokB1PusKamiOracle        = "9b2471d4-2821-4220-becc-94bebfdc456f"
)

func TestBokB1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		bokB1MistbladeOracle:      "Mistblade Shinobi",
		bokB1ThroatSlitterOracle:  "Throat Slitter",
		bokB1FalseHopeOracle:      "Kami of False Hope",
		bokB1HigureOracle:         "Higure, the Still Wind",
		bokB1SkullsnatcherOracle:  "Skullsnatcher",
		bokB1TendoOracle:          "Tendo Ice Bridge",
		bokB1DisruptingOracle:     "Disrupting Shoal",
		bokB1LeechesOracle:        "Blessing of Leeches",
		bokB1OkibaOracle:          "Okiba-Gang Shinobi",
		bokB1PatronMoonOracle:     "Patron of the Moon",
		bokB1LifegiftOracle:       "Lifegift",
		bokB1WalkerOracle:         "Walker of Secret Ways",
		bokB1ShukoOracle:          "Shuko",
		bokB1GodsEyeOracle:        "Gods' Eye, Gate to the Reikai",
		bokB1PatronOrochiOracle:   "Patron of the Orochi",
		bokB1BlazingOracle:        "Blazing Shoal",
		bokB1TakenOracle:          "That Which Was Taken",
		bokB1TraprootOracle:       "Traproot Kami",
		bokB1TerashisGraspOracle:  "Terashi's Grasp",
		bokB1YukoraOracle:         "Yukora, the Prisoner",
		bokB1LoamDwellerOracle:    "Loam Dweller",
		bokB1RibbonsOracle:        "Ribbons of the Reikai",
		bokB1ToilsOracle:          "Toils of Night and Day",
		bokB1PatronNezumiOracle:   "Patron of the Nezumi",
		bokB1CrackOracle:          "Crack the Earth",
		bokB1TeardropOracle:       "Teardrop Kami",
		bokB1BlizzardHerderOracle: "Akki Blizzard-Herder",
		bokB1WebOfWarOracle:       "In the Web of War",
		bokB1PatronKitsuneOracle:  "Patron of the Kitsune",
		bokB1HerosDemiseOracle:    "Hero's Demise",
		bokB1SickeningOracle:      "Sickening Shoal",
		bokB1StreamOracle:         "Stream of Consciousness",
		bokB1KyokiOracle:          "Kyoki, Sanity's Eclipse",
		bokB1IndebtedOracle:       "Indebted Samurai",
		bokB1StirOracle:           "Stir the Grave",
		bokB1BileUrchinOracle:     "Bile Urchin",
		bokB1NourishingOracle:     "Nourishing Shoal",
		bokB1UncheckedOracle:      "Unchecked Growth",
		bokB1IreOracle:            "Ire of Kaminari",
		bokB1FirstVolleyOracle:    "First Volley",
		bokB1SilverstormOracle:    "Silverstorm Samurai",
		bokB1PatronAkkiOracle:     "Patron of the Akki",
		bokB1PusKamiOracle:        "Pus Kami",
	}
	if len(want) != 43 {
		t.Fatalf("the batch registers 43 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBokB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Traproot Kami", "Creature — Spirit", bokB1TraprootOracle, []string{"defender", "reach"}},
		{"Patron of the Moon", "Legendary Creature — Spirit", bokB1PatronMoonOracle, []string{"flying"}},
		{"Silverstorm Samurai", "Creature — Fox Samurai", bokB1SilverstormOracle, []string{"flash"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
	if spec, _ := Lookup(bokB1LeechesOracle); len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Blessing of Leeches prints %v, want flash", spec.PrintedKeywords)
	}
}

// --- the ninjas -----------------------------------------------------------------------

func TestBokB1NinjasHaveTheirNinjutsuCost(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Mistblade Shinobi", bokB1MistbladeOracle, "{U}"},
		{"Throat Slitter", bokB1ThroatSlitterOracle, "{2}{B}"},
		{"Higure, the Still Wind", bokB1HigureOracle, "{2}{U}{U}"},
		{"Skullsnatcher", bokB1SkullsnatcherOracle, "{B}"},
		{"Okiba-Gang Shinobi", bokB1OkibaOracle, "{3}{B}"},
		{"Walker of Secret Ways", bokB1WalkerOracle, "{1}{U}"},
	} {
		abs := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abs) == 0 || abs[0].Cost.Mana != tc.mana || abs[0].Cost.ReturnToHand.Empty() {
			t.Errorf("%s: want a first ability of ninjutsu %s with a return cost", tc.name, tc.mana)
		}
	}
}

func TestBokB1MistbladeShinobiBouncesACreatureThatPlayerControls(t *testing.T) {
	g, _, _, me, opp := neoNinjutsu(t, "Mistblade Shinobi", bokB1MistbladeOracle, "Creature — Human Ninja", 1, 1, "{U}")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, g.Seats[2].ID, "Someone Else's Bear", "Creature — Bear", 2, 2)
	neoToDamage(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the damage trigger")
	}
	if !hasID(p.PickTargetCards, victim) || hasID(p.PickTargetCards, mine) || hasID(p.PickTargetCards, other) {
		t.Fatalf("the trigger offers %v, want only the damaged player's creature %v", p.PickTargetCards, victim)
	}
	pickIfAsked(t, g, me.ID, victim)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestBokB1ThroatSlitterDestroysANonblackCreature(t *testing.T) {
	g, _, _, me, opp := neoNinjutsu(t, "Throat Slitter", bokB1ThroatSlitterOracle, "Creature — Rat Ninja", 2, 2, "{2}{B}")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	black := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Zombie",
		TypeLine: "Creature — Zombie", ManaCost: "{1}{B}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	neoToDamage(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the damage trigger")
	}
	if !hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, black) {
		t.Fatalf("the trigger offers %v, want only the nonblack creature %v", p.PickTargetCards, bear)
	}
	pickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the nonblack creature was not destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature was destroyed")
	}
}

func TestBokB1OkibaGangShinobiMakesThatPlayerDiscardTwo(t *testing.T) {
	g, _, _, _, opp := neoNinjutsu(t, "Okiba-Gang Shinobi", bokB1OkibaOracle, "Creature — Rat Ninja", 3, 2, "{3}{B}")
	d2dSetHand(opp, 3)
	neoToDamage(t, g)
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no discard prompt for the damaged player")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("discard bounds [%d,%d], want [2,2]", c.ChooseMin, c.ChooseMax)
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the damaged player holds %d cards, want 1", opp.Hand.Size())
	}
}

func TestBokB1SkullsnatcherExilesUpToTwoCardsFromThatPlayersGraveyard(t *testing.T) {
	g, _, _, me, opp := neoNinjutsu(t, "Skullsnatcher", bokB1SkullsnatcherOracle, "Creature — Rat Ninja", 2, 1, "{B}")
	a := wthGraveyardCard(opp, "Their First", "Instant")
	b := wthGraveyardCard(opp, "Their Second", "Creature — Bear")
	c := wthGraveyardCard(opp, "Their Third", "Sorcery")
	mine := wthGraveyardCard(me, "My Card", "Instant")
	neoToDamage(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the damage trigger")
	}
	if !hasID(p.PickTargetCards, a) || !hasID(p.PickTargetCards, c) || hasID(p.PickTargetCards, mine) {
		t.Fatalf("the trigger offers %v, want only the damaged player's graveyard", p.PickTargetCards)
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 2 {
		t.Errorf("target bounds [%d,%d], want up to two", p.PickTargetMin, p.PickTargetMax)
	}
	pickIfAsked(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the two picked cards were not exiled")
	}
	if !opp.Graveyard.Contains(c) || !me.Graveyard.Contains(mine) {
		t.Error("an unpicked card left its graveyard")
	}
}

func TestBokB1HigureFindsANinjaWhenItConnects(t *testing.T) {
	g, _, _, me, _ := neoNinjutsu(t, "Higure, the Still Wind", bokB1HigureOracle, "Legendary Creature — Human Ninja", 3, 4, "{2}{U}{U}")
	ninja := pushLibraryCardForTest(me, game.Card{Name: "Some Ninja", TypeLine: "Creature — Rat Ninja"})
	rat := pushLibraryCardForTest(me, game.Card{Name: "Some Rat", TypeLine: "Creature — Rat Rogue"})
	neoToDamage(t, g)
	passPriorityAroundTable(t, g)
	open := chkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, open)
	open = chkB1OpenSearch(g)
	if !hasID(open.SearchCards, ninja) || hasID(open.SearchCards, rat) {
		t.Errorf("the search offers %v, want only the Ninja card %v", open.SearchCards, ninja)
	}
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(ninja) {
		t.Error("the Ninja card did not reach the hand")
	}
}

func TestBokB1HigureMakesANinjaUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	higure := b12Push(g, me.ID, "Higure, the Still Wind", "Legendary Creature — Human Ninja", bokB1HigureOracle, 3, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, higure, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "C", "C"); err == nil {
		t.Fatal("the ability targeted a creature that is not a Ninja")
	}
	ravB2MustActivate(t, g, me.ID, higure, 1, game.ActivateAbilityParams{Targets: cardRefs(higure)}, "C", "C")
	assertRestrictions(t, g, higure, game.CantBeBlocked)
	assertRestrictions(t, g, bear, 0)
}

func TestBokB1WalkerOfSecretWaysLooksAtThatPlayersHand(t *testing.T) {
	g, _, _, me, opp := neoNinjutsu(t, "Walker of Secret Ways", bokB1WalkerOracle, "Creature — Human Ninja", 1, 2, "{1}{U}")
	d2dSetHand(opp, 2)
	neoToDamage(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range opp.Hand.Cards {
		if !c.KnownBy[me.ID] {
			t.Errorf("%s in the damaged player's hand was not looked at", c.Name)
		}
	}
}

func TestBokB1WalkerOfSecretWaysReturnsANinjaOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := b12Push(g, me.ID, "Walker of Secret Ways", "Creature — Human Ninja", bokB1WalkerOracle, 1, 2)
	ninja := b12Creature(g, me.ID, "My Ninja", "Creature — Rat Ninja", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "U", "U"); err == nil {
		t.Fatal("the ability targeted a creature that is not a Ninja")
	}
	ravB2MustActivate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(ninja)}, "U", "U")
	if g.Battlefield.Contains(ninja) || !me.Hand.Contains(ninja) {
		t.Error("the Ninja did not return to its owner's hand")
	}
	advanceToNextSeatsTurn(t, g)
	if err := ravB2Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(walker)}, "U", "U"); err == nil {
		t.Error("the ability was activated on another player's turn")
	}
}

// --- the Shoals -------------------------------------------------------------------------

// bokB1Pitch is a Shoal cast for its alternative cost: `card` exiled from
// hand at X = x.
func bokB1Pitch(x int, card uuid.UUID, targets []game.TargetRef) game.CastSpellParams {
	return game.CastSpellParams{AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{card}, XValue: x, Targets: targets, Strict: true}
}

func TestBokB1BlazingShoalPitchesARedCardForItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	red := pushHandCardWithManaCost(me, "Red Giant", "Creature — Giant", "{3}{R}")
	blue := pushHandCardWithManaCost(me, "Blue Drake", "Creature — Drake", "{3}{U}")
	if _, err := castWithTapParams(t, g, "Blazing Shoal", "Instant — Arcane", "{X}{R}{R}", bokB1BlazingOracle, bokB1Pitch(4, blue, cardTarget(bear))); err == nil {
		t.Fatal("Blazing Shoal pitched a blue card")
	}
	if _, err := castWithTapParams(t, g, "Blazing Shoal", "Instant — Arcane", "{X}{R}{R}", bokB1BlazingOracle, bokB1Pitch(4, red, cardTarget(bear))); err != nil {
		t.Fatalf("pitching a mana value 4 red card at X = 4: %v", err)
	}
	if !g.Exile.Contains(red) {
		t.Fatal("the red card was not exiled as the cost")
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 6 || tough != 2 {
		t.Errorf("the target is %d/%d, want 6/2", p, tough)
	}
	dcCleanup(g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the bonus outlived the cleanup step (%d/%d)", p, tough)
	}
}

func TestBokB1SickeningShoalShrinksByTheBlackCardsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	wurm := b12Creature(g, opp.ID, "Their Wurm", "Creature — Wurm", 5, 5)
	black := pushHandCardWithManaCost(me, "Black Knight", "Creature — Human Knight", "{1}{B}{B}")
	if _, err := castWithTapParams(t, g, "Sickening Shoal", "Instant — Arcane", "{X}{B}{B}", bokB1SickeningOracle, bokB1Pitch(3, black, cardTarget(wurm))); err != nil {
		t.Fatalf("pitching a mana value 3 black card at X = 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, wurm); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tough)
	}
}

func TestBokB1NourishingShoalGainsX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	green := pushHandCardWithManaCost(me, "Green Wurm", "Creature — Wurm", "{4}{G}{G}")
	life := me.Life
	if _, err := castWithTapParams(t, g, "Nourishing Shoal", "Instant — Arcane", "{X}{G}{G}", bokB1NourishingOracle, bokB1Pitch(6, green, nil)); err != nil {
		t.Fatalf("pitching a mana value 6 green card at X = 6: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d → %d, want +6", life, me.Life)
	}
	life = me.Life
	castXSpell(t, g, "Nourishing Shoal", "Instant — Arcane", bokB1NourishingOracle, "{X}{G}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("hard-cast at X = 2: life %d → %d, want +2", life, me.Life)
	}
}

func TestBokB1DisruptingShoalCountersOnlyASpellOfManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushHandCardWithManaCost(me, "Small Drake", "Creature — Drake", "{1}{U}")
	big := pushHandCardWithManaCost(me, "Big Drake", "Creature — Drake", "{3}{U}{U}")

	wurm := castWithCost(t, g, "Wurm", "Creature — Wurm", "{3}{G}{G}", "")
	if _, err := castWithTapParams(t, g, "Disrupting Shoal", "Instant — Arcane", "{X}{U}{U}", bokB1DisruptingOracle, bokB1Pitch(2, small, b16TargetCard(wurm))); err != nil {
		t.Fatalf("pitching a mana value 2 blue card: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wurm) {
		t.Fatal("a mana value 5 spell was countered at X = 2")
	}

	elk := castWithCost(t, g, "Elk", "Creature — Elk", "{3}{G}{G}", "")
	if _, err := castWithTapParams(t, g, "Disrupting Shoal", "Instant — Arcane", "{X}{U}{U}", bokB1DisruptingOracle, bokB1Pitch(5, big, b16TargetCard(elk))); err != nil {
		t.Fatalf("pitching a mana value 5 blue card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(elk) || !me.Graveyard.Contains(elk) {
		t.Error("the mana value 5 spell was not countered at X = 5")
	}
}

// --- lands and artifacts ------------------------------------------------------------------

func TestBokB1TendoIceBridgeSpendsItsOneChargeCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bridge := b12PlayFromHand(t, g, "Tendo Ice Bridge", "Land", bokB1TendoOracle, game.CastSpellParams{})
	c := b12Card(t, g, bridge)
	if c.Counters[game.CounterCharge] != 1 || c.Tapped {
		t.Fatalf("entered with %d charge counters, tapped %v; want one and untapped", c.Counters[game.CounterCharge], c.Tapped)
	}
	if err := g.ActivateManaAbility(me.ID, bridge, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("remove a counter for {U}: %v", err)
	}
	if me.ManaPool.CountColor("U") != 1 || b12Card(t, g, bridge).Counters[game.CounterCharge] != 0 {
		t.Fatal("the counter was not spent for one blue mana")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bridge).Tapped = false })
	if err := g.ActivateManaAbility(me.ID, bridge, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("the coloured ability was offered with no charge counter")
	}
	if err := g.ActivateManaAbility(me.ID, bridge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
	if me.ManaPool.CountColor("C") != 1 {
		t.Error("the land did not tap for {C}")
	}
}

func TestBokB1GodsEyeTapsForColorlessAndLeavesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	eye := pushCatalogPermanent(g, me.ID, "Gods' Eye, Gate to the Reikai", "Legendary Land", bokB1GodsEyeOracle, false)
	if err := g.ActivateManaAbility(me.ID, eye, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.CountColor("C") != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d {C} of %d, want 1 of 1", me.ManaPool.CountColor("C"), me.ManaPool.Total())
	}
	destroyIt(t, g, eye)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
}

func TestBokB1ShukoPumpsForAZeroCostEquip(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shuko := b12Push(g, me.ID, "Shuko", "Artifact — Equipment", bokB1ShukoOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, shuko, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the equipped creature is %d/%d, want 3/2", p, tough)
	}
}

func TestBokB1ThatWhichWasTakenMakesAPermanentIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	taken := b12Push(g, me.ID, "That Which Was Taken", "Legendary Artifact", bokB1TakenOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, taken, 0, game.ActivateAbilityParams{Targets: cardRefs(taken)}, "C", "C", "C", "C"); err == nil {
		t.Fatal("That Which Was Taken targeted itself")
	}
	ravB2MustActivate(t, g, me.ID, taken, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}, "C", "C", "C", "C")
	if b12Card(t, g, theirs).Counters["divinity"] != 1 {
		t.Fatal("no divinity counter on the target")
	}
	destroyIt(t, g, theirs)
	destroyIt(t, g, other)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the permanent with a divinity counter was destroyed")
	}
	if g.Battlefield.Contains(other) {
		t.Error("a permanent with no divinity counter survived")
	}
}

// --- creatures ----------------------------------------------------------------------------

func TestBokB1TraprootKamiCountsEveryForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	kami := b12Push(g, me.ID, "Traproot Kami", "Creature — Spirit", bokB1TraprootOracle, 0, 0)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if p, tough := effectivePTOf(t, g, kami); p != 0 || tough != 3 {
		t.Errorf("Traproot Kami is %d/%d, want 0/3 with three Forests on the battlefield", p, tough)
	}
}

func TestBokB1KamiOfFalseHopePreventsCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	kami := b12Push(g, me.ID, "Kami of False Hope", "Creature — Spirit", bokB1FalseHopeOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(kami) {
		t.Fatal("the Kami was not sacrificed")
	}
	life := opp.Life
	attackWith(t, g, opp.ID, bear)
	if opp.Life != life {
		t.Errorf("combat damage was dealt (%d → %d)", life, opp.Life)
	}
}

func TestBokB1BileUrchinMakesAPlayerLoseLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	urchin := b12Push(g, me.ID, "Bile Urchin", "Creature — Spirit", bokB1BileUrchinOracle, 1, 1)
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, urchin, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 || g.Battlefield.Contains(urchin) {
		t.Errorf("life %d → %d, urchin on battlefield %v; want -1 and sacrificed", life, opp.Life, g.Battlefield.Contains(urchin))
	}
}

func TestBokB1TeardropKamiUntapsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kami := b12Push(g, me.ID, "Teardrop Kami", "Creature — Spirit", bokB1TeardropOracle, 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	if err := ravB2Activate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1) // untap it
	if isTapped(g, bear) {
		t.Error("the creature is still tapped")
	}
}

func TestBokB1PusKamiDestroysANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	// Its real mana cost, so its own soulshift has no target in the graveyard.
	kami := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pus Kami",
		TypeLine: "Creature — Spirit", OracleID: bokB1PusKamiOracle, ManaCost: "{5}{B}{B}",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	black := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Zombie",
		TypeLine: "Creature — Zombie", ManaCost: "{1}{B}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := ravB2Activate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{Targets: cardRefs(black)}, "B"); err == nil {
		t.Fatal("Pus Kami targeted a black creature")
	}
	ravB2MustActivate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "B")
	drcSettle(t, g, me.ID)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(kami) {
		t.Error("the target survived or Pus Kami was not sacrificed")
	}
}

func TestBokB1PusKamiSoulshiftsASpiritOfManaValueSixOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kami := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pus Kami", TypeLine: "Creature — Spirit", OracleID: bokB1PusKamiOracle,
		ManaCost: "{5}{B}{B}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	fits, tooBig := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: fits, Name: "Fits", TypeLine: "Creature — Spirit",
		ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: tooBig, Name: "Too Big", TypeLine: "Creature — Spirit",
		ManaCost: "{7}", Owner: me.ID, Controller: me.ID})
	killCreature(t, g, me.ID, kami)
	p := passUntilPickTarget(t, g, me.ID)
	if p == nil {
		t.Fatal("no soulshift target prompt")
	}
	if !hasID(p.PickTargetCards, fits) || hasID(p.PickTargetCards, tooBig) || hasID(p.PickTargetCards, kami) {
		t.Errorf("soulshift offers %v, want only the mana value 6 Spirit %v", p.PickTargetCards, fits)
	}
	pickCard(t, g, me.ID, fits)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(fits) {
		t.Error("soulshift did not return the Spirit card")
	}
}

func TestBokB1BushidoBodies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough, n        int
	}{
		{"Silverstorm Samurai", "Creature — Fox Samurai", bokB1SilverstormOracle, 3, 3, 1},
		{"Indebted Samurai", "Creature — Human Samurai", bokB1IndebtedOracle, 2, 3, 1},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
		})
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, id); p != tc.power+tc.n || tough != tc.tough+tc.n {
			t.Errorf("%s is %d/%d after becoming blocked, want %d/%d", tc.name, p, tough, tc.power+tc.n, tc.tough+tc.n)
		}
	}
}

func TestBokB1IndebtedSamuraiGrowsWhenASamuraiDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	samurai := b12Push(g, me.ID, "Indebted Samurai", "Creature — Human Samurai", bokB1IndebtedOracle, 2, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "My Other Samurai", "Creature — Human Samurai", 2, 2)
	killCreature(t, g, me.ID, bear)
	drcSettle(t, g, me.ID)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a non-Samurai's death triggered")
	}
	killCreature(t, g, me.ID, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if b12Card(t, g, samurai).Counters["+1/+1"] != 1 {
		t.Error("no +1/+1 counter after a Samurai died")
	}
}

func TestBokB1YukoraSacrificesYourNonOgres(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	yukora := b12Push(g, me.ID, "Yukora, the Prisoner", "Legendary Creature — Demon Spirit", bokB1YukoraOracle, 5, 5)
	ogre := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre Warrior", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	exileCreature(t, g, yukora)
	drcSettle(t, g, me.ID)
	if g.Battlefield.Contains(bear) {
		t.Error("your non-Ogre creature survived")
	}
	if !g.Battlefield.Contains(ogre) || !g.Battlefield.Contains(theirs) {
		t.Error("an Ogre or an opponent's creature was sacrificed")
	}
}

func TestBokB1AkkiBlizzardHerderMakesEachPlayerSacrificeALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	herder := b12Push(g, me.ID, "Akki Blizzard-Herder", "Creature — Goblin Shaman", bokB1BlizzardHerderOracle, 1, 1)
	lands := make([]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		lands[i] = b12Permanent(g, p.ID, "Mountain", "Basic Land — Mountain")
	}
	killCreature(t, g, me.ID, herder)
	drcSettle(t, g, me.ID)
	for i, id := range lands {
		if g.Battlefield.Contains(id) {
			t.Errorf("seat %d kept its land", i)
		}
	}
}

// --- "Whenever you cast a Spirit or Arcane spell" ---------------------------------------------

func TestBokB1LoamDwellerPutsALandOnASpiritOrArcaneSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Loam Dweller", "Creature — Spirit", bokB1LoamDwellerOracle, 2, 2)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a Bear spell triggered Loam Dweller")
	}
	castCatalogSpell(t, g, "An Arcane Spell", "Instant — Arcane", "", nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !isTapped(g, land) {
		t.Error("the land is not on the battlefield tapped")
	}
}

func TestBokB1KyokiMakesAnOpponentExileACard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Kyoki, Sanity's Eclipse", "Legendary Creature — Demon Spirit", bokB1KyokiOracle, 6, 4)
	d2dSetHand(opp, 2)
	card := opp.Hand.Cards[0].InstanceID
	castCatalogSpell(t, g, "A Spirit", "Creature — Spirit", "", nil)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, card)
	if !g.Exile.Contains(card) || opp.Hand.Size() != 1 {
		t.Error("the opponent did not exile the chosen card")
	}
}

// --- spells -----------------------------------------------------------------------------

func TestBokB1TerashisGraspDestroysAndGainsItsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	rock := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{3}"})
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Terashi's Grasp", "Sorcery — Arcane", bokB1TerashisGraspOracle, cardTarget(bear)); err == nil {
		t.Fatal("Terashi's Grasp targeted a creature")
	}
	life := me.Life
	castCatalogSpell(t, g, "Terashi's Grasp", "Sorcery — Arcane", bokB1TerashisGraspOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestBokB1HerosDemiseDestroysOnlyALegend(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	legend := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 3, 3)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Hero's Demise", "Instant", bokB1HerosDemiseOracle, cardTarget(bear)); err == nil {
		t.Fatal("Hero's Demise targeted a nonlegendary creature")
	}
	castCatalogSpell(t, g, "Hero's Demise", "Instant", bokB1HerosDemiseOracle, cardTarget(legend))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(legend) || !g.Battlefield.Contains(bear) {
		t.Error("Hero's Demise did not destroy the legend alone")
	}
}

func TestBokB1RibbonsOfTheReikaiDrawsPerSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	b12Creature(g, me.ID, "My Other Spirit", "Creature — Spirit", 1, 1)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Ribbons of the Reikai", "Sorcery — Arcane", bokB1RibbonsOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two draws for two Spirits", hand, me.Hand.Size())
	}
}

func TestBokB1CrackTheEarthIsAnEdict(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keep := b12Permanent(g, me.ID, "My Keeper", "Artifact")
	lose := b12Permanent(g, me.ID, "My Loser", "Artifact")
	theirs := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p.ID != me.ID {
			theirs = append(theirs, b12Permanent(g, p.ID, "Their Thing", "Enchantment"))
		}
	}
	castCatalogSpell(t, g, "Crack the Earth", "Sorcery — Arcane", bokB1CrackOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, lose)
	drcSettle(t, g, me.ID)
	if !g.Battlefield.Contains(keep) || g.Battlefield.Contains(lose) {
		t.Error("the caster did not sacrifice the permanent they chose")
	}
	for _, id := range theirs {
		if g.Battlefield.Contains(id) {
			t.Error("an opponent kept their only permanent")
		}
	}
}

func TestBokB1ToilsOfNightAndDayTapsOrUntapsTwoPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Permanent(g, me.ID, "My Land", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	castCatalogSpell(t, g, "Toils of Night and Day", "Instant — Arcane", bokB1ToilsOracle, cardRefs(theirs, mine))
	passPriorityAroundTable(t, g)
	first := latestOptionPickFor(g, me.ID)
	if first == nil || len(first.PickOptions) != 3 || first.PickOptions[0].Cards[0] != theirs {
		t.Fatalf("the first target is asked about first, with three answers; got %+v", first)
	}
	answerOptionPick(t, g, me.ID, 0) // tap it
	second := latestOptionPickFor(g, me.ID)
	if second == nil || second.PickOptions[0].Cards[0] != mine {
		t.Fatalf("the second target is asked about once the first is answered; got %+v", second)
	}
	answerOptionPick(t, g, me.ID, 1) // untap it
	if !isTapped(g, theirs) || isTapped(g, mine) {
		t.Errorf("tapped: theirs %v, mine %v; want tapped and untapped", isTapped(g, theirs), isTapped(g, mine))
	}
}

func TestBokB1StreamOfConsciousnessShufflesUpToFourCards(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	ids := []uuid.UUID{
		wthGraveyardCard(opp, "One", "Instant"), wthGraveyardCard(opp, "Two", "Sorcery"),
		wthGraveyardCard(opp, "Three", "Creature — Bear"), wthGraveyardCard(opp, "Four", "Artifact"),
	}
	left := wthGraveyardCard(opp, "Five", "Enchantment")
	library := opp.Library.Size()
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}}
	for _, id := range ids {
		targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 1})
	}
	castCatalogSpell(t, g, "Stream of Consciousness", "Instant — Arcane", bokB1StreamOracle, targets)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if opp.Graveyard.Contains(id) || !opp.Library.Contains(id) {
			t.Error("a target card was not shuffled into its owner's library")
		}
	}
	if !opp.Graveyard.Contains(left) || opp.Library.Size() != library+4 {
		t.Errorf("library %d → %d, want +4 with the fifth card left", library, opp.Library.Size())
	}
}

func TestBokB1StirTheGraveReturnsACreatureOfManaValueXOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small, big := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}{G}",
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if _, err := castWithTapParams(t, g, "Stir the Grave", "Sorcery", "{X}{B}", bokB1StirOracle,
		game.CastSpellParams{XValue: 2, Targets: cardTarget(big)}); err == nil {
		t.Fatal("Stir the Grave targeted a mana value 6 card at X = 2")
	}
	castXSpell(t, g, "Stir the Grave", "Sorcery", bokB1StirOracle, "{X}{B}", 2, cardTarget(small))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestBokB1UncheckedGrowthGivesASpiritTrample(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		trample  bool
	}{{"Creature — Spirit", true}, {"Creature — Bear", false}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := b12Creature(g, me.ID, "Target", tc.typeLine, 2, 2)
		castCatalogSpell(t, g, "Unchecked Growth", "Instant — Arcane", bokB1UncheckedOracle, cardTarget(id))
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, id); p != 6 || tough != 6 {
			t.Errorf("%s: %d/%d, want 6/6", tc.typeLine, p, tough)
		}
		if got := hasKeywordOnBattlefield(t, g, id, "trample"); got != tc.trample {
			t.Errorf("%s: trample %v, want %v", tc.typeLine, got, tc.trample)
		}
	}
}

func TestBokB1IreOfKaminariCountsArcaneCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	wthGraveyardCard(me, "Arcane One", "Instant — Arcane")
	wthGraveyardCard(me, "Arcane Two", "Sorcery — Arcane")
	wthGraveyardCard(me, "Plain Instant", "Instant")
	wthGraveyardCard(opp, "Their Arcane", "Instant — Arcane")
	life := opp.Life
	castCatalogSpell(t, g, "Ire of Kaminari", "Instant — Arcane", bokB1IreOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage for two Arcane cards", life, opp.Life)
	}
}

func TestBokB1FirstVolleyHitsTheCreatureAndItsController(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	life := opp.Life
	castCatalogSpell(t, g, "First Volley", "Instant — Arcane", bokB1FirstVolleyOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if c := b12Card(t, g, bear); c.DamageMarked != 1 {
		t.Errorf("the creature has %d damage marked, want 1", c.DamageMarked)
	}
	if opp.Life != life-1 {
		t.Errorf("its controller's life %d → %d, want -1", life, opp.Life)
	}
}

// --- enchantments -------------------------------------------------------------------------

func TestBokB1BlessingOfLeechesRegeneratesAndCostsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Blessing of Leeches", "Enchantment — Aura", bokB1LeechesOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	ravB2MustActivate(t, g, me.ID, aura, 0, game.ActivateAbilityParams{})
	destroyIt(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the enchanted creature was not regenerated")
	}
	advanceToUpkeepOf(t, g, 0)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d at your upkeep, want -1", life, me.Life)
	}
}

func TestBokB1LifegiftGainsOnALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Lifegift", "Enchantment", bokB1LifegiftOracle, 0, 0)
	life := me.Life
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
}

func TestBokB1InTheWebOfWarHastensYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "In the Web of War", "Enchantment", bokB1WebOfWarOracle, 0, 0)
	theirs := b12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Bear", "Creature — Bear", 2, 2)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the creature lacks haste")
	}
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 4/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature was pumped to %d power", p)
	}
	dcCleanup(g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 || hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the bonus outlived the cleanup step")
	}
}

// --- the Patrons --------------------------------------------------------------------------

func TestBokB1PatronsDeclareTheirOffering(t *testing.T) {
	for _, tc := range []struct{ name, oracle, label string }{
		{"Patron of the Moon", bokB1PatronMoonOracle, "Moonfolk offering"},
		{"Patron of the Orochi", bokB1PatronOrochiOracle, "Snake offering"},
		{"Patron of the Nezumi", bokB1PatronNezumiOracle, "Rat offering"},
		{"Patron of the Kitsune", bokB1PatronKitsuneOracle, "Fox offering"},
		{"Patron of the Akki", bokB1PatronAkkiOracle, "Goblin offering"},
	} {
		costs := game.OptionalCostsFor(tc.oracle)
		if len(costs) != 1 || costs[0].Key != OfferingKey || costs[0].Label != tc.label || !costs[0].AsThoughFlash ||
			!costs[0].ReducedBySacrificedManaCost || game.SacrificeCostCount(costs[0].Sacrifice) != 1 {
			t.Errorf("%s: want one %s", tc.name, tc.label)
		}
	}
}

func TestBokB1PatronOfTheKitsuneIsOfferedAtInstantSpeed(t *testing.T) {
	g, me, _ := emergeTable(t)
	patron := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: patron, Name: "Patron of the Kitsune", TypeLine: "Legendary Creature — Spirit",
			ManaCost: "{4}{W}{W}", OracleID: bokB1PatronKitsuneOracle, Power: 5, Toughness: 6,
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	})
	fox := apaPush(g, me.ID, me.ID, game.Card{Name: "My Fox", TypeLine: "Creature — Fox Samurai", ManaCost: "{1}{W}", Power: 2, Toughness: 2})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if got := offeringPrice(t, g, me, patron, offeringParams(false, fox)); got != "{3}{W}" {
		t.Errorf("offering a {1}{W} Fox: price %s, want {3}{W}", got)
	}
	ssToBeginCombat(t, g)
	if err := g.CastSpell(me.ID, patron, offeringParams(false, bear)); err == nil {
		t.Fatal("a Bear paid the Fox offering")
	}
	if err := g.CastSpell(me.ID, patron, offeringParams(false, fox)); err != nil {
		t.Fatalf("offering the Fox in combat: %v", err)
	}
	if !me.Graveyard.Contains(fox) {
		t.Fatal("the Fox was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, patron) == nil {
		t.Error("Patron of the Kitsune did not resolve")
	}
}

func TestBokB1PatronOfTheKitsuneGainsWheneverACreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Patron of the Kitsune", "Legendary Creature — Spirit", bokB1PatronKitsuneOracle, 5, 6)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	declareAttack(t, g, opp.ID, bear)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1 for one attacking creature", life, me.Life)
	}
}

func TestBokB1PatronOfTheAkkiPumpsYourTeamWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	patron := b12Push(g, me.ID, "Patron of the Akki", "Legendary Creature — Spirit", bokB1PatronAkkiOracle, 5, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, patron)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 4/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, patron); p != 7 {
		t.Errorf("the Patron's power is %d, want 7", p)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature was pumped to %d power", p)
	}
}

func TestBokB1PatronOfTheMoonPutsUpToTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	patron := b12Push(g, me.ID, "Patron of the Moon", "Legendary Creature — Spirit", bokB1PatronMoonOracle, 5, 4)
	a := handCardForTest(me, "Island", "Basic Land — Island", "")
	b := handCardForTest(me, "Plains", "Basic Land — Plains", "")
	c := handCardForTest(me, "Swamp", "Basic Land — Swamp", "")
	if err := ravB2Activate(t, g, me.ID, patron, 0, game.ActivateAbilityParams{}, "C"); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 2 || pick.ChooseMin != 0 {
		t.Fatalf("want a prompt for up to two land cards, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, a, b)
	for _, id := range []uuid.UUID{a, b} {
		if !g.Battlefield.Contains(id) || !isTapped(g, id) {
			t.Error("a picked land is not on the battlefield tapped")
		}
	}
	if !me.Hand.Contains(c) {
		t.Error("the third land left the hand")
	}
}

func TestBokB1PatronOfTheOrochiUntapsForestsAndGreenCreaturesOnce(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	patron := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Patron of the Orochi",
		TypeLine: "Legendary Creature — Spirit", ManaCost: "{6}{G}{G}", OracleID: bokB1PatronOrochiOracle,
		Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	island := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	elf := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Elf",
		TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	bear := b12Creature(g, me.ID, "Colourless Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{forest, island, elf, bear} {
			_ = g.TapTargetForEffect(id)
		}
	})
	ravB2MustActivate(t, g, me.ID, patron, 0, game.ActivateAbilityParams{})
	for _, id := range []uuid.UUID{forest, elf, patron} {
		if isTapped(g, id) {
			t.Errorf("%s is still tapped", b12Card(t, g, id).Name)
		}
	}
	for _, id := range []uuid.UUID{island, bear} {
		if !isTapped(g, id) {
			t.Errorf("%s was untapped", b12Card(t, g, id).Name)
		}
	}
	if err := ravB2Activate(t, g, me.ID, patron, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the ability was activated twice in one turn")
	}
}

func TestBokB1PatronOfTheNezumiDrainsAPermanentsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Patron of the Nezumi", "Legendary Creature — Spirit", bokB1PatronNezumiOracle, 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	oppLife, myLife := opp.Life, me.Life
	destroyIt(t, g, mine)
	drcSettle(t, g, me.ID)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("your own permanent dying drained someone (you %d → %d, them %d → %d)", myLife, me.Life, oppLife, opp.Life)
	}
	destroyIt(t, g, theirs)
	drcSettle(t, g, me.ID)
	if opp.Life != oppLife-1 {
		t.Errorf("the opponent's life %d → %d, want -1", oppLife, opp.Life)
	}
}
