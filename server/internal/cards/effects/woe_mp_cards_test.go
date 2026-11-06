package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_mp_cards_test.go — Wilds of Eldraine, the M–P batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeMPCoralsmithOracle = "05f3638f-3c66-47bb-8a82-ad173e152573"
	woeMPDaredevilOracle  = minecartDaredevilOracleID
	woeMPMintOracle       = "c8c88c33-b8df-4b19-b12c-79c9709af8f1"
	woeMPMotesOracle      = "09eec5a2-7835-4b24-9dd1-594612ee9152"
	woeMPValorOracle      = "bfac8072-2bbc-43ca-b141-9cd0799840f7"
	woeMPDreadknightOracl = mosswoodDreadknightOracleID
	woeMPNevaOracle       = "ba5895e1-f12b-43c7-b47c-376588d9d2d2"
	woeMPAttendantsOracle = obyrasAttendantsOracleID
	woeMPPlungeOracle     = "a272f15d-caa2-40e9-894f-19638bd5f375"
	woeMPHareOracle       = pollenShieldHareOracleID
	woeMPMerchantOracle   = "2f796869-7350-4886-b210-cb5b9cb3e78a"
	woeMPPestsOracle      = "c2a4066a-f911-4622-beb2-683e7e10822f"
	woeMPBearType         = "Creature — Bear"
)

func TestWoeMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeMPCoralsmithOracle:        "Merfolk Coralsmith",
		woeMPDaredevilOracle:         "Minecart Daredevil",
		woeMPDaredevilOracle + "#1":  "Ride the Rails",
		woeMPMintOracle:              "Mintstrosity",
		woeMPMotesOracle:             "Misleading Motes",
		woeMPValorOracle:             "Moment of Valor",
		woeMPDreadknightOracl:        "Mosswood Dreadknight",
		woeMPDreadknightOracl + "#1": "Dread Whispers",
		woeMPNevaOracle:              "Neva, Stalked by Nightmares",
		woeMPAttendantsOracle:        "Obyra's Attendants",
		woeMPAttendantsOracle + "#1": "Desperate Parry",
		woeMPPlungeOracle:            "Plunge into Winter",
		woeMPHareOracle:              "Pollen-Shield Hare",
		woeMPHareOracle + "#1":       "Hare Raising",
		woeMPMerchantOracle:          "Provisions Merchant",
		woeMPPestsOracle:             "Pests of Honor",
	} {
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

// woeMPReady clears summoning sickness so a pushed creature can attack.
func woeMPReady(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].SummonedThisTurn = false
					g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
				}
			}
		}
	})
}

// --- Merfolk Coralsmith ---------------------------------------------------------

func TestWoeMPCoralsmithPumpsForOneAndScriesTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	smith := b12Push(g, me.ID, "Merfolk Coralsmith", "Creature — Merfolk", woeMPCoralsmithOracle, 2, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump activated without {1}")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, smith, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{1}: +1/-1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, smith); p != 3 || tough != 2 {
		t.Fatalf("Coralsmith is %d/%d after one pump, want 3/2", p, tough)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("a scry was asked before it died")
	}
	woeB1Destroy(t, g, smith)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %v after it died, want one looking at two cards", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Minecart Daredevil // Ride the Rails ---------------------------------------

func TestWoeMPRideTheRailsPumpsTwoOneAndGoesOnAnAdventure(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "Bear", woeMPBearType, 2, 2)
	card := woeB1Adventure(me, woeMPDaredevilOracle,
		game.Face{Name: "Minecart Daredevil", TypeLine: "Creature — Dwarf Knight", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 4, Toughness: 2},
		game.Face{Name: "Ride the Rails", TypeLine: "Instant — Adventure", ManaCost: "{1}{R}", Colors: []string{"R"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(bear))
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the target is %d/%d, want 4/3", p, tough)
	}
	endTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
	}
}

// --- Mintstrosity -----------------------------------------------------------------

func TestWoeMPMintstrosityMakesAFoodWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	mint := b12Push(g, me.ID, "Mintstrosity", "Creature — Horror", woeMPMintOracle, 3, 1)
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food before it died", got)
	}
	woeB1Destroy(t, g, mint)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after it died, want 1", got)
	}
}

// --- Misleading Motes -------------------------------------------------------------

func TestWoeMPMisleadingMotesLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bottom bool
	}{{"top", false}, {"bottom", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := woeB1Seats(g)
			seedLibrary(opp, "X", "Y")
			bear := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
			castCatalogSpell(t, g, "Misleading Motes", "Instant", woeMPMotesOracle, cardRefs(bear))
			passPriorityAroundTable(t, g)
			if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
				t.Fatal("the caster was asked; the owner chooses")
			}
			c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
			if c == nil {
				t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
			}
			var top, bottom []uuid.UUID
			if tc.bottom {
				bottom = []uuid.UUID{bear}
			} else {
				top = []uuid.UUID{bear}
			}
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if g.Battlefield.Contains(bear) {
				t.Fatal("the creature is still on the battlefield")
			}
			cards := opp.Library.Cards
			// The bottom of a library is Cards[0], the top the last entry.
			if tc.bottom && cards[0].InstanceID != bear {
				t.Errorf("the bottom card is %q, want the creature", cards[0].Name)
			}
			if !tc.bottom && cards[len(cards)-1].InstanceID != bear {
				t.Errorf("the top card is %q, want the creature", cards[len(cards)-1].Name)
			}
		})
	}
}

// --- Moment of Valor --------------------------------------------------------------

func TestWoeMPMomentOfValorUntapsPumpsAndProtects(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "Bear", woeMPBearType, 2, 2)
	tapForTest(t, g, bear)
	castModal(t, g, "Moment of Valor", "Instant", woeMPValorOracle, []int{0}, []game.TargetRef{modeTargetRef(bear, 0)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the creature was not untapped")
	}
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the creature is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Fatal("the creature did not gain indestructible")
	}
	woeB1Destroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the indestructible creature was destroyed")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "indestructible") || effectivePower(t, g, bear) != 2 {
		t.Error("the first mode outlasted the turn")
	}
}

func TestWoeMPMomentOfValorDestroysOnlyAPowerFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	big := b12Creature(g, opp.ID, "Big", woeMPBearType, 4, 4)
	small := b12Creature(g, opp.ID, "Small", woeMPBearType, 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	card := game.Card{InstanceID: uuid.New(), Name: "Moment of Valor", TypeLine: "Instant", ManaCost: "{2}{W}",
		OracleID: woeMPValorOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{modeTargetRef(small, 0)}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("destroy mode at a power-3 creature: %v, want ErrIllegalTarget", err)
	}
	castModal(t, g, "Moment of Valor", "Instant", woeMPValorOracle, []int{1}, []game.TargetRef{modeTargetRef(big, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !opp.Graveyard.Contains(big) {
		t.Error("the power-4 creature was not destroyed")
	}
}

// --- Mosswood Dreadknight // Dread Whispers ---------------------------------------

func woeMPDreadknight(owner uuid.UUID) game.Card {
	return adventureCard(owner, adventureCase{oracle: woeMPDreadknightOracl, faces: [2]game.Face{
		{Name: "Mosswood Dreadknight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 3, Toughness: 2},
		{Name: "Dread Whispers", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{B}", Colors: []string{"B"}},
	}})
}

func TestWoeMPDreadWhispersDrawsAndLosesOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "Drawn")
	card := woeMPDreadknight(me.ID)
	me.Hand.PushTop(card)
	hand, life := me.Hand.Size(), me.Life
	woeB1CastAdventure(t, g, me, card, nil)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want the cast card replaced by one draw", hand, got)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
}

func TestWoeMPDreadknightHasTrampleAndCanBeCastAsAnAdventureFromTheGraveyardAfterItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "Drawn")
	knight := woeMPDreadknight(me.ID)
	g.Battlefield.PushTop(knight)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: knight.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	if !hasEffectiveKeyword(t, g, knight.InstanceID, "trample") {
		t.Error("Mosswood Dreadknight has no trample")
	}
	killCreature(t, g, me.ID, knight.InstanceID)
	drcSettle(t, g, me.ID)
	advanceTo(t, g, game.StepPrecombatMain)
	if !me.Graveyard.Contains(knight.InstanceID) {
		t.Fatal("setup: the Dreadknight should be in the graveyard")
	}
	if err := g.CastSpell(me.ID, knight.InstanceID, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("the creature half was castable from the graveyard")
	}
	if err := g.CastSpell(me.ID, knight.InstanceID, game.CastSpellParams{FromZone: "graveyard", Face: 1}); err != nil {
		t.Fatalf("casting Dread Whispers from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, knight.InstanceID) {
		t.Error("the Adventure did not go to exile after resolving from the graveyard")
	}
}

func TestWoeMPDreadknightPermissionLastsThroughYourNextTurnOnly(t *testing.T) {
	for _, tc := range []struct {
		name     string
		rounds   int
		castable bool
	}{{"your next turn", 1, true}, {"the turn after", 2, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := woeB1Seats(g)
			seedLibrary(me, "A", "B", "C", "D")
			knight := woeMPDreadknight(me.ID)
			g.Battlefield.PushTop(knight)
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: knight.InstanceID, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
			})
			killCreature(t, g, me.ID, knight.InstanceID)
			drcSettle(t, g, me.ID)
			for i := 0; i < tc.rounds*len(g.Seats); i++ {
				endTurn(t, g)
			}
			if g.Seats[g.Turn.ActiveSeat].ID != me.ID {
				t.Fatalf("setup: seat %d is active, want mine", g.Turn.ActiveSeat)
			}
			advanceTo(t, g, game.StepPrecombatMain)
			err := g.CastSpell(me.ID, knight.InstanceID, game.CastSpellParams{FromZone: "graveyard", Face: 1})
			if tc.castable && err != nil {
				t.Errorf("Dread Whispers was not castable from the graveyard on your next turn: %v", err)
			}
			if !tc.castable && err == nil {
				t.Error("Dread Whispers was castable from the graveyard after your next turn")
			}
		})
	}
}

// --- Neva, Stalked by Nightmares --------------------------------------------------

func TestWoeMPNevaReturnsACreatureOrEnchantmentCardOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	aura := socB11Graveyard(me, "Dead Aura", "Enchantment — Aura", "{W}")
	land := socB11Graveyard(me, "Dead Land", "Land", "")
	advanceTo(t, g, game.StepPrecombatMain)
	castFromHandAndResolve(t, g, me, game.Card{InstanceID: uuid.New(), Name: "Neva, Stalked by Nightmares",
		TypeLine: "Legendary Creature — Human Noble", OracleID: woeMPNevaOracle, Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID})
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, land) {
			t.Error("a land card was offered")
		}
		pickCard(t, g, me.ID, aura)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the enchantment card did not return to hand")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land card left the graveyard")
	}
}

func TestWoeMPNevaGrowsAndScriesWhenYourEnchantmentGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	neva := b12Push(g, me.ID, "Neva, Stalked by Nightmares", "Legendary Creature — Human Noble", woeMPNevaOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, neva, "menace") {
		t.Error("Neva has no menace")
	}
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	woeB1Destroy(t, g, theirs)
	if got := countersOn(g, neva, game.CounterPlusOne); got != 0 || scryChoiceFor(g, me.ID) != nil {
		t.Fatalf("an opponent's enchantment counted: %d counters", got)
	}
	mine := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	woeB1Destroy(t, g, mine)
	if got := countersOn(g, neva, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after my enchantment died, want 1", got)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	answerScryKeepAll(t, g, me.ID)
}

// --- Obyra's Attendants // Desperate Parry ----------------------------------------

func TestWoeMPDesperateParryShrinksPowerByFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 5, 5)
	card := woeB1Adventure(me, woeMPAttendantsOracle,
		game.Face{Name: "Obyra's Attendants", TypeLine: "Creature — Faerie Wizard", ManaCost: "{4}{U}", Colors: []string{"U"}, Power: 3, Toughness: 4},
		game.Face{Name: "Desperate Parry", TypeLine: "Instant — Adventure", ManaCost: "{1}{U}", Colors: []string{"U"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(foe))
	if p, tough := effectivePTOf(t, g, foe); p != 1 || tough != 5 {
		t.Errorf("the 5/5 is %d/%d, want 1/5", p, tough)
	}
}

func TestWoeMPObyrasAttendantsFly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	id := pushCatalogPermanent(g, me.ID, "Obyra's Attendants", "Creature — Faerie Wizard", woeMPAttendantsOracle, false)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Obyra's Attendants has no flying")
	}
}

// --- Plunge into Winter -----------------------------------------------------------

func TestWoeMPPlungeIntoWinterTapsThenScriesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	foe := b12Creature(g, opp.ID, "Foe", woeMPBearType, 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Plunge into Winter", "Instant", woeMPPlungeOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, foe); !c.Tapped {
		t.Error("the target was not tapped")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("the draw came before the scry was answered")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one draw", hand, got)
	}
}

func TestWoeMPPlungeIntoWinterWithNoTargetStillScriesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Plunge into Winter", "Instant", woeMPPlungeOracle, nil)
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one draw", hand, got)
	}
}

// --- Pollen-Shield Hare // Hare Raising -------------------------------------------

func TestWoeMPPollenShieldHarePumpsOnlyYourCreatureTokens(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Pollen-Shield Hare", "Creature — Rabbit", woeMPHareOracle, false)
	bear := b12Creature(g, me.ID, "Bear", woeMPBearType, 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1)
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1)
	})
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a nontoken creature got the anthem: %d/%d", p, tough)
	}
	for _, c := range g.Battlefield.Cards {
		if !IsToken(c) || !c.IsCreature() {
			continue
		}
		want := 1
		if c.Controller == me.ID {
			want = 2
		}
		if p, tough := effectivePTOf(t, g, c.InstanceID); p != want || tough != want {
			t.Errorf("token under %v is %d/%d, want %d/%d", c.Controller, p, tough, want, want)
		}
	}
}

func TestWoeMPHareRaisingPumpsByCreaturesYouControlAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "Bear", woeMPBearType, 2, 2)
	b12Creature(g, me.ID, "Friend", woeMPBearType, 1, 1)
	b12Creature(g, me.ID, "Other Friend", woeMPBearType, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	card := woeB1Adventure(me, woeMPHareOracle,
		game.Face{Name: "Pollen-Shield Hare", TypeLine: "Creature — Rabbit", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Hare Raising", TypeLine: "Sorcery — Adventure", ManaCost: "{G}", Colors: []string{"G"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("Hare Raising targeted a creature you don't control")
	}
	woeB1CastAdventure(t, g, me, card, cardRefs(bear))
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5 (+3/+3 for three creatures)", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the target did not gain vigilance")
	}
	// X was counted as it resolved.
	b12Creature(g, me.ID, "Late", woeMPBearType, 1, 1)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("a creature that came later changed X: power %d", p)
	}
}

// --- Provisions Merchant ----------------------------------------------------------

func TestWoeMPProvisionsMerchantMakesAFoodAndEatsItToPumpTheAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	advanceTo(t, g, game.StepPrecombatMain)
	merchant := uuid.New()
	castFromHandAndResolve(t, g, me, game.Card{InstanceID: merchant, Name: "Provisions Merchant",
		TypeLine: "Creature — Beast Peasant", OracleID: woeMPMerchantOracle, Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID})
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Food after it entered, want 1", got)
	}
	friend := b12Creature(g, me.ID, "Friend", woeMPBearType, 2, 2)
	home := b12Creature(g, me.ID, "Home", woeMPBearType, 2, 2)
	woeMPReady(g, merchant, friend, home)
	declareAttack(t, g, opp.ID, merchant, friend)
	passPriorityAroundTable(t, g)
	food := findBattlefieldByName(g, "Food")
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, food) {
		t.Fatal("the Food is chosen as the trigger resolves, with a floor of zero for the \"may\"")
	}
	answerOwnPermanents(t, g, me.ID, food)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the chosen Food was not sacrificed")
	}
	for name, id := range map[string]uuid.UUID{"the Merchant": merchant, "the other attacker": friend} {
		base := 2
		if id == merchant {
			base = 3
		}
		if p, tough := effectivePTOf(t, g, id); p != base+1 || tough != base+1 {
			t.Errorf("%s is %d/%d, want +1/+1", name, p, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("%s did not gain trample", name)
		}
	}
	if p := effectivePower(t, g, home); p != 2 || hasEffectiveKeyword(t, g, home, "trample") {
		t.Error("a creature that did not attack was pumped")
	}
}

func TestWoeMPProvisionsMerchantDeclinedPumpsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	merchant := b12Push(g, me.ID, "Provisions Merchant", "Creature — Beast Peasant", woeMPMerchantOracle, 3, 3)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FoodToken(), 1) })
	woeMPReady(g, merchant)
	declareAttack(t, g, opp.ID, merchant)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after declining, want 1", got)
	}
	if p := effectivePower(t, g, merchant); p != 3 || hasEffectiveKeyword(t, g, merchant, "trample") {
		t.Error("declining still pumped the attack")
	}
}

// --- Pests of Honor ---------------------------------------------------------------

func TestWoeMPPestsOfHonorGetsACounterAtCombatOnlyWhileCelebrating(t *testing.T) {
	// No celebration: nothing triggers as combat begins.
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	pests := pushCatalogPermanent(g, me.ID, "Pests of Honor", "Creature — Mouse", woeMPPestsOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if triggerOnStackFrom(g, "Pests of Honor") || countersOn(g, pests, game.CounterPlusOne) != 0 {
		t.Fatal("the trigger fired with no celebration")
	}

	// Celebration: one +1/+1 counter as combat begins.
	g2 := newCatalogGame(t)
	me2, _ := woeB1Seats(g2)
	pests2 := pushCatalogPermanent(g2, me2.ID, "Pests of Honor", "Creature — Mouse", woeMPPestsOracle, false)
	celebrationTokens(t, g2, me2.ID, 2)
	advanceTo(t, g2, game.StepBeginCombat)
	passPriorityAroundTable(t, g2)
	if got := countersOn(g2, pests2, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after a celebrating combat, want 1", got)
	}
}

func TestWoeMPPestsOfHonorDoesNotTriggerOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pests := pushCatalogPermanent(g, opp.ID, "Pests of Honor", "Creature — Mouse", woeMPPestsOracle, false)
	celebrationTokens(t, g, opp.ID, 2)
	celebrationTokens(t, g, me.ID, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, pests, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters at an opponent's combat, want none", got)
	}
}
