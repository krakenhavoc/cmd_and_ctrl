package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b061_cards_test.go — Kaldheim, batch 061: each test plays the real
// catalog entry through the engine.

const (
	oldB061GravenLore    = "0d9df234-520c-40aa-a37d-070181cf695c"
	oldB061BlessingFrost = "240d650b-97ea-40ad-8c1b-fe1ea52d861f"
	oldB061Runeforge     = "2705a288-3b9d-4452-a320-28b86331aa09"
	oldB061Avenger       = "47c67df1-52f5-459f-b83e-6b971be9b294"
	oldB061Yew           = "5b1dc298-403d-484d-90dd-e1c9b73e0188"
	oldB061Pyre          = "691ba449-40e1-4976-8dab-9b1741ea37b8"
	oldB061Arcanist      = "6bdb251b-51fe-46cc-8b40-540d081c9419"
	oldB061Demon         = "d551fae8-6b40-4118-b9db-29d2f5dff405"
	oldB061Fury          = "e5f83bfc-ff74-4bb9-82c8-a07fedc6e916"
	oldB061Glory         = "fab29bc7-1f0a-4df4-b8f1-c008ce5583d2"
	oldB061Aldergard     = "ff3bc67b-f435-4bb7-9c21-8d24d8808c1a"
)

func TestOldB061EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB061GravenLore:    "Graven Lore",
		oldB061BlessingFrost: "Blessing of Frost",
		oldB061Runeforge:     "Runeforge Champion",
		oldB061Avenger:       "Skemfar Avenger",
		oldB061Yew:           "Rootless Yew",
		oldB061Pyre:          "Pyre of Heroes",
		oldB061Arcanist:      "Frostpyre Arcanist",
		oldB061Demon:         "Burning-Rune Demon",
		oldB061Fury:          "Fire Giant's Fury",
		oldB061Glory:         "Search for Glory",
		oldB061Aldergard:     "Spirit of the Aldergard",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

// oldB061Lands puts n basic lands onto the battlefield, snow or not.
func oldB061Lands(g *game.Game, owner uuid.UUID, n int, snow bool) {
	oldB061ColorLands(g, owner, n, snow, "Island")
}

// oldB061ColorLands is oldB061Lands for a named basic land type.
func oldB061ColorLands(g *game.Game, owner uuid.UUID, n int, snow bool, basic string) {
	for i := 0; i < n; i++ {
		if snow {
			effectstest.BlbB1Push(g, owner, "Snow-Covered "+basic, "Basic Snow Land — "+basic, "", 0, 0)
		} else {
			effectstest.BlbB1Push(g, owner, basic, "Basic Land — "+basic, "", 0, 0)
		}
	}
}

// oldB061Cast casts a hand card with the table's lands.
func oldB061Cast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestOldB061SpiritOfTheAldergardPumpsPerOtherSnowPermanentAndFetchesASnowLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Snow-Covered Forest", "Basic Snow Land — Forest"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	oldB061Lands(g, me.ID, 2, true)
	oldB061Lands(g, me.ID, 1, false)
	spirit := effectstest.CastRingCreature(t, g, "Spirit of the Aldergard", "Snow Creature — Bear Spirit", oldB061Aldergard, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("only snow land cards are legal finds")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Snow-Covered Forest")
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		found = found || h.Name == "Snow-Covered Forest"
	}
	if !found {
		t.Error("the snow land is not in hand")
	}
	// Two other snow permanents; the Spirit does not count itself.
	if p, tough := effectstest.PtOf(t, g, spirit); p != 2 || tough != 4 {
		t.Errorf("the Spirit is %d/%d, want 2/4", p, tough)
	}
}

func TestOldB061RootlessYewFetchesABigCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := game.Card{Name: "Big Toughness", TypeLine: "Creature — Wall", Power: 0, Toughness: 6}
	powerful := game.Card{Name: "Big Power", TypeLine: "Creature — Ogre", Power: 6, Toughness: 1}
	small := game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 5, Toughness: 5}
	effectstest.SeedSearchLibrary(me, big, powerful, small)
	yew := effectstest.B12Push(g, me.ID, "Rootless Yew", "Creature — Treefolk", oldB061Yew, 5, 4)
	effectstest.KillCreature(t, g, me.ID, yew)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Small") != uuid.Nil {
		t.Error("a 5/5 is not a legal find")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Big Power")
	effectstest.PassPriorityAroundTable(t, g)
	got := false
	for _, h := range me.Hand.Cards {
		got = got || h.Name == "Big Power"
	}
	if !got {
		t.Error("the 6-power creature is not in hand")
	}
}

func TestOldB061SkemfarAvengerDrawsForNontokenElvesAndBerserkersOnly(t *testing.T) {
	cases := []struct {
		name, typeLine string
		token          bool
		other          bool
		draws          bool
	}{
		{"Elf", "Creature — Elf Druid", false, false, true},
		{"Berserker", "Creature — Human Berserker", false, false, true},
		{"Bear", "Creature — Bear", false, false, false},
		{"Elf token", "Token Creature — Elf", true, false, false},
		{"opponent's Elf", "Creature — Elf", false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.B12Push(g, me.ID, "Skemfar Avenger", "Creature — Elf Berserker", oldB061Avenger, 3, 1)
			owner := me
			if tc.other {
				owner = opp
			}
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: tc.name, TypeLine: tc.typeLine, Power: 1, Toughness: 1,
				Owner: owner.ID, Controller: owner.ID})
			hand, life := me.Hand.Size(), me.Life
			effectstest.KillCreature(t, g, owner.ID, id)
			effectstest.PassPriorityAroundTable(t, g)
			drew := me.Hand.Size() == hand+1 && me.Life == life-1
			if drew != tc.draws {
				t.Errorf("drew=%v (hand %d->%d, life %d->%d), want %v", drew, hand, me.Hand.Size(), life, me.Life, tc.draws)
			}
		})
	}
}

func TestOldB061GravenLoreScriesForSnowManaThenDrawsThree(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMainForCost(t, g)
		oldB061Lands(g, me.ID, 5, snow)
		id := effectstest.HandCardFull(me, "Graven Lore", "Snow Instant", "{3}{U}{U}", oldB061GravenLore, []string{"U"})
		hand := me.Hand.Size()
		oldB061Cast(t, g, me, id)
		effectstest.PassPriorityAroundTable(t, g)
		sc := effectstest.ScryChoiceFor(g, me.ID)
		if snow != (sc != nil) {
			t.Errorf("snow=%v: scry prompt open = %v", snow, sc != nil)
		}
		if sc != nil {
			if len(sc.ScryCards) != 5 {
				t.Errorf("scry %d cards, want 5 (one per snow mana)", len(sc.ScryCards))
			}
			effectstest.AnswerScryKeepAll(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand-1+3 {
			t.Errorf("snow=%v: hand %d -> %d, want the spell gone and three drawn", snow, hand, got)
		}
	}
}

func TestOldB061SearchForGloryFindsTheThreeKindsAndGainsLifePerSnowMana(t *testing.T) {
	for _, snow := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.SeedSearchLibrary(me,
			effectstest.SearchTestLand("Plain Bear", "Creature — Bear"),
			effectstest.SearchTestLand("Snow Rock", "Snow Artifact"),
			effectstest.SearchTestLand("Old King", "Legendary Creature — Elf"),
			effectstest.SearchTestLand("A Saga", "Enchantment — Saga"),
			effectstest.SearchTestLand("Snow Instant", "Snow Instant"),
		)
		effectstest.ToMainForCost(t, g)
		oldB061ColorLands(g, me.ID, 3, snow, "Plains")
		life := me.Life
		id := effectstest.HandCardFull(me, "Search for Glory", "Snow Sorcery", "{2}{W}", oldB061Glory, []string{"W"})
		oldB061Cast(t, g, me, id)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no search offered")
		}
		for _, bad := range []string{"Plain Bear", "Snow Instant"} {
			if effectstest.SearchOptionNamed(g, c, bad) != uuid.Nil {
				t.Errorf("%s is not a snow permanent, legendary or a Saga", bad)
			}
		}
		for _, ok := range []string{"Snow Rock", "Old King", "A Saga"} {
			if effectstest.SearchOptionNamed(g, c, ok) == uuid.Nil {
				t.Errorf("%s should be a legal find", ok)
			}
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Old King")
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if snow {
			want += 3
		}
		if me.Life != want {
			t.Errorf("snow=%v: life %d, want %d", snow, me.Life, want)
		}
	}
}

func TestOldB061BlessingOfFrostDistributesCountersThenDrawsForPowerFourOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.ToMainForCost(t, g)
	oldB061ColorLands(g, me.ID, 4, true, "Forest")
	id := effectstest.HandCardFull(me, "Blessing of Frost", "Snow Sorcery", "{3}{G}", oldB061BlessingFrost, []string{"G"})
	hand := me.Hand.Size()
	oldB061Cast(t, g, me, id)
	effectstest.PassPriorityAroundTable(t, g)
	// X = 4: pick two of the three bears, 3 on the first, the rest on the second.
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.AnswerNumber(t, g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, a, "+1/+1"); n != 3 {
		t.Errorf("A has %d counters, want 3", n)
	}
	if n := effectstest.CounterCount(g, b, "+1/+1"); n != 1 {
		t.Errorf("B has %d counters, want 1", n)
	}
	if n := effectstest.CounterCount(g, c, "+1/+1"); n != 0 {
		t.Errorf("C has %d counters, want none", n)
	}
	// A is 5/5 now, Big is 4/4: two creatures with power 4 or greater.
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d -> %d, want the spell gone and two cards drawn", hand, got)
	}
	_ = big
}

func TestOldB061BlessingOfFrostWithoutSnowManaOnlyDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.ToMainForCost(t, g)
	oldB061ColorLands(g, me.ID, 4, false, "Forest")
	id := effectstest.HandCardFull(me, "Blessing of Frost", "Snow Sorcery", "{3}{G}", oldB061BlessingFrost, []string{"G"})
	hand := me.Hand.Size()
	oldB061Cast(t, g, me, id)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, a, "+1/+1"); n != 0 {
		t.Errorf("%d counters with no snow mana spent", n)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want the spell gone and one card drawn", hand, got)
	}
}

func TestOldB061FrostpyreArcanistSearchesForTheNameOfAGraveyardCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Lightning Bolt", "Instant"),
		effectstest.SearchTestLand("Divination", "Sorcery"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastRingCreature(t, g, "Frostpyre Arcanist", "Creature — Giant Wizard", oldB061Arcanist, 2, 5)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Divination") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("only an instant or sorcery sharing a graveyard card's name is a legal find")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Lightning Bolt")
	effectstest.PassPriorityAroundTable(t, g)
	got := false
	for _, h := range me.Hand.Cards {
		got = got || h.Name == "Lightning Bolt"
	}
	if !got {
		t.Error("Lightning Bolt is not in hand")
	}
}

func TestOldB061FrostpyreArcanistCostsOneLessWithAGiantOrWizard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		reduced        bool
	}{
		{"Giant", "Creature — Giant", true},
		{"Wizard", "Creature — Human Wizard", true},
		{"Bear", "Creature — Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Creature(g, me.ID, tc.name, tc.typeLine, 1, 1)
			effectstest.ToMainForCost(t, g)
			oldB061Lands(g, me.ID, 4, false)
			id := effectstest.HandCardFull(me, "Frostpyre Arcanist", "Creature — Giant Wizard", "{4}{U}", oldB061Arcanist, []string{"U"})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true})
			if (err == nil) != tc.reduced {
				t.Errorf("cast with four lands: err=%v, want castable=%v", err, tc.reduced)
			}
		})
	}
}

func TestOldB061RuneforgeChampionSearchesLibraryAndOrGraveyardForARune(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Rune of Speed", "Enchantment — Aura Rune"),
		effectstest.SearchTestLand("Grizzly Bears", "Creature — Bear"),
	)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rune of Might", TypeLine: "Enchantment — Aura Rune", Owner: me.ID, Controller: me.ID})
	effectstest.CastRingCreature(t, g, "Runeforge Champion", "Creature — Dwarf Warrior", oldB061Runeforge, 2, 3)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if err := g.BeginSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("BeginSearchLibrary: %v", err)
	}
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if effectstest.SearchOptionNamed(g, c, "Rune of Might") == uuid.Nil {
		t.Error("the graveyard Rune is not offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Rune of Speed") != uuid.Nil {
		t.Error("the library Rune is offered although only the graveyard was searched")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Rune of Might")
	effectstest.PassPriorityAroundTable(t, g)
	got := false
	for _, h := range me.Hand.Cards {
		got = got || h.Name == "Rune of Might"
	}
	if !got {
		t.Error("the Rune is not in hand")
	}
}

func TestOldB061RuneforgeChampionLetsYouPayOneForRuneSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Runeforge Champion", "Creature — Dwarf Warrior", oldB061Runeforge, 2, 3)
	effectstest.ToMainForCost(t, g)
	rune := effectstest.HandCardFull(me, "Test Rune", "Enchantment — Aura Rune", "{3}{R}", "", []string{"R"})
	plain := effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{3}{R}", "", []string{"R"})
	offers := effectstest.GrantedCardOffers(g, me.ID, rune, game.ZoneHand)
	if effectstest.OfferWithKey(offers, effects.GrantedAltCostOneRuneSpells) == nil {
		t.Error("a Rune spell is not offered the {1} cost")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, plain, game.ZoneHand), effects.GrantedAltCostOneRuneSpells) != nil {
		t.Error("a non-Rune spell is offered the {1} cost")
	}
}

func TestOldB061BurningRuneDemonOpponentChoosesWhichCardYouKeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Burning-Rune Demon", "Creature — Demon Berserker"),
	)
	_ = ids
	effectstest.CastRingCreature(t, g, "Burning-Rune Demon", "Creature — Demon Berserker", oldB061Demon, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Burning-Rune Demon") != uuid.Nil {
		t.Error("a card named Burning-Rune Demon is not a legal find")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island", "Swamp")
	effectstest.PassPriorityAroundTable(t, g)
	// The controller picks the opponent, who then chooses a card.
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	var island, swamp uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			if card, ok := g.LookupCardForEffect(id); ok {
				switch card.Name {
				case "Island":
					island = id
				case "Swamp":
					swamp = id
				}
			}
		}
	}
	if island == uuid.Nil || swamp == uuid.Nil {
		t.Fatal("the two found cards left the library before the opponent chose")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, island)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(island) {
		t.Error("the chosen card is not in the controller's hand")
	}
	if !me.Graveyard.Contains(swamp) {
		t.Error("the other card is not in the controller's graveyard")
	}
}

func TestOldB061PyreOfHeroesFetchesTheNextCreatureThatSharesAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	elf := game.Card{InstanceID: uuid.New(), Name: "Elf Two", TypeLine: "Creature — Elf Druid", ManaCost: "{1}{G}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
	g.Battlefield.PushTop(elf)
	pyre := effectstest.PushCatalogPermanent(g, me.ID, "Pyre of Heroes", "Artifact", oldB061Pyre, false)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Elf Three", TypeLine: "Creature — Elf Warrior", ManaCost: "{1}{G}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Bear Three", TypeLine: "Creature — Bear", ManaCost: "{1}{G}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Elf Four", TypeLine: "Creature — Elf", ManaCost: "{2}{G}{G}", Power: 3, Toughness: 3},
	)
	oldB061Lands(g, me.ID, 2, false)
	if err := g.ActivateCatalogAbility(me.ID, pyre, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true,
		SacrificeIDs: []uuid.UUID{elf.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear Three") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Elf Four") != uuid.Nil {
		t.Error("only a creature that shares a type and costs exactly one more is legal")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Elf Three")
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.SearchFetchedCard(g, "Elf Three"); !ok {
		t.Error("Elf Three is not on the battlefield")
	}
}

func TestOldB061FireGiantsFuryPumpsAGiantAndExilesCardsForItsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	oldB061ColorLands(g, me.ID, 2, false, "Mountain")
	id := effectstest.HandCardFull(me, "Fire Giant's Fury", "Sorcery", "{1}{R}", oldB061Fury, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("a non-Giant was accepted as the target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, giant); p != 5 || tough != 5 {
		t.Fatalf("the Giant is %d/%d, want 5/5", p, tough)
	}
	lib := me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, giant)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 5 {
		t.Errorf("%d cards exiled from the library, want 5 (the combat damage)", got)
	}
	exiled := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == me.ID {
			exiled++
		}
	}
	if exiled != 5 {
		t.Errorf("%d of my cards in exile, want 5", exiled)
	}
}
