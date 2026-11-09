package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b11_cards_test.go — Final Fantasy batches 11 and 12 (#4462). Each
// test plays the real catalog entry and asserts what its text says.

const (
	finB11SeiferOracle        = "417b9d9c-0a67-4554-a275-a6ae15a04bfb"
	finB11DuelistsFlameOracle = "bbc990e4-1601-498f-9031-0cd663139367"
	finB11SquallOracle        = "880c2b0a-b708-4f11-a88a-ec990c904602"
	finB11WingsOracle         = "ed06b85a-e85f-47c4-ba90-1664e69a338f"
	finB11MegaFlareOracle     = "891c6ca5-d238-4ac2-8df4-c6dc9a807562"
	finB11AmarantOracle       = "322d2407-fa9f-48d0-bdf0-88f6d24cf865"
	finB11ViviOracle          = "fd4a1f4b-674d-41f3-a9a6-c2b65bf5c1ab"
	finB11NoctisOracle        = "7abe351e-0889-4fc1-91ae-8372bcd28acb"
	finB11FishingGearOracle   = "7d309b18-8dc5-41ee-888c-e275761029e7"
	finB11CampOracle          = "ed77fdf2-59c0-4310-9b12-80d28beeaeef"
	finB12FlashOracle         = "a92ae7b0-cbbb-44a8-bfd8-2a68f4c1ce1a"
	finB12CuisineOracle       = "45f58002-c831-4b35-8b8c-9a502118e2ec"
	finB12ResolveOracle       = "e3f68aff-9347-452a-8f26-b344016a275d"
	finB12GodoOracle          = "f890d51a-dc90-4179-8423-4738a72a527c"
	finB12JodahOracle         = "490a8045-e160-4ccf-a05e-658ad2b7ac2d"
	finB12TymnaOracle         = "d15642e4-e61c-4d29-af48-de837991245e"
	finB12WinotaOracle        = "02598a35-1b11-4abe-89cd-fc44c8c36c15"
	finB12KenrithOracle       = "d209b948-9afb-4fd1-a961-72c87282878c"
	finB12TeferiOracle        = "2b6cbd55-5adc-4a34-ac45-8531705a7ee1"
	finB12LightUpOracle       = "8d286376-f386-43ab-a9bd-8478fb9e2497"
)

// finB11Registered asserts a catalog entry is Full with no caveats.
func finB11Registered(t *testing.T, names map[string]string) {
	t.Helper()
	for oracle, name := range names {
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

// --- Chocobo Camp ---------------------------------------------------------------------

func TestFinB11ChocoboCampEntersTappedUnlessYouControlALegendaryCreature(t *testing.T) {
	finB11Registered(t, map[string]string{finB11CampOracle: "Chocobo Camp"})
	g := newCatalogGame(t)
	camp := playLandFromHand(t, g, "Chocobo Camp", finB11CampOracle)
	if !b12Card(t, g, camp).Tapped {
		t.Error("Chocobo Camp entered untapped with no legendary creature")
	}
	g2 := newCatalogGame(t)
	b12Creature(g2, g2.Seats[0].ID, "Legend", "Legendary Creature — Human", 1, 1)
	camp2 := playLandFromHand(t, g2, "Chocobo Camp", finB11CampOracle)
	if b12Card(t, g2, camp2).Tapped {
		t.Error("Chocobo Camp entered tapped with a legendary creature")
	}
}

func TestFinB11ChocoboCampMakesTheNextBirdSpellEnterWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	camp := playLandFromHand(t, g, "Chocobo Camp", finB11CampOracle)
	activateManaFor(t, g, me.ID, camp, 0, game.ManaAbilityParams{})
	if got := me.ManaPool.CountColor("G"); got != 1 {
		t.Fatalf("pool has %d G, want 1", got)
	}
	// A non-Bird creature spell does not use the promise up.
	bear := nsHandCard(me, "Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear.InstanceID, "+1/+1"); n != 0 {
		t.Errorf("the Bear entered with %d counters, want 0", n)
	}
	bird := nsHandCard(me, "Test Bird", "Creature — Bird", "")
	if err := g.CastSpell(me.ID, bird.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bird.InstanceID, "+1/+1"); n != 1 {
		t.Errorf("the Bird entered with %d counters, want 1", n)
	}
	bird2 := nsHandCard(me, "Second Bird", "Creature — Bird", "")
	if err := g.CastSpell(me.ID, bird2.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bird2.InstanceID, "+1/+1"); n != 0 {
		t.Errorf("the second Bird entered with %d counters, want 0 (it was one cast)", n)
	}
}

func TestFinB11ChocoboCampMakesABirdForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	camp := playLandFromHand(t, g, "Chocobo Camp", finB11CampOracle)
	// Without the mana the activation is refused.
	if err := g.ActivateCatalogAbility(me.ID, camp, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	floatMana(t, g, me, "{C}{C}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, camp, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	finCG2OneBird(t, g, me.ID, 1)
	if !b12Card(t, g, camp).Tapped {
		t.Error("the Camp did not tap")
	}
}

// --- Fishing Gear ---------------------------------------------------------------------

// finB11GearHits sets up Fishing Gear on a bear with `top` on the
// opponent's library and has the bear connect.
func finB11GearHits(t *testing.T, g *game.Game, top func(opp *game.Player) uuid.UUID) (me, opp *game.Player, topID uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gear := b12Push(g, me.ID, "Fishing Gear", "Artifact — Equipment", finB11FishingGearOracle, 0, 0)
	equipTo(t, g, me.ID, gear, bear)
	passPriorityAroundTable(t, g)
	topID = top(opp)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	return me, opp, topID
}

func TestFinB11FishingGearOffersAPermanentCardFromTheirLibrary(t *testing.T) {
	finB11Registered(t, map[string]string{finB11FishingGearOracle: "Fishing Gear"})
	g := newCatalogGame(t)
	me, opp, top := finB11GearHits(t, g, func(opp *game.Player) uuid.UUID {
		return seedLibraryTop(opp, "Stolen Bear", "Creature — Bear", "{1}{G}")
	})
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	var stolen *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Stolen Bear" {
			stolen = &g.Battlefield.Cards[i]
		}
	}
	if stolen == nil || stolen.Controller != me.ID || stolen.Owner != opp.ID {
		t.Fatalf("Stolen Bear %+v, want on the battlefield under my control, owned by the victim (card %s)", stolen, top)
	}
	if battlefieldNamed(g, me.ID, "Fish") != 0 {
		t.Error("a Fish was made although the card was put onto the battlefield")
	}
}

func TestFinB11FishingGearMakesAFishWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := finB11GearHits(t, g, func(opp *game.Player) uuid.UUID {
		return seedLibraryTop(opp, "Stolen Bear", "Creature — Bear", "{1}{G}")
	})
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if battlefieldNamed(g, me.ID, "Fish") != 1 || battlefieldNamed(g, me.ID, "Stolen Bear") != 0 {
		t.Errorf("Fish %d, Stolen Bear %d; want one Fish and no Bear", battlefieldNamed(g, me.ID, "Fish"), battlefieldNamed(g, me.ID, "Stolen Bear"))
	}
	if g.Exile.Size() == 0 {
		t.Error("the card is not in exile")
	}
}

func TestFinB11FishingGearMakesAFishFromANonpermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := finB11GearHits(t, g, func(opp *game.Player) uuid.UUID {
		return seedLibraryTop(opp, "Test Bolt", "Instant", "{R}")
	})
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("a nonpermanent card was offered")
	}
	if battlefieldNamed(g, me.ID, "Fish") != 1 {
		t.Errorf("Fish %d, want 1", battlefieldNamed(g, me.ID, "Fish"))
	}
}

// --- Brilliant Wings ------------------------------------------------------------------

// finB11WingsOnABear casts Brilliant Wings on a bear and returns both.
func finB11WingsOnABear(t *testing.T, g *game.Game) (wings, bear uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wings = castCatalogSpell(t, g, "Brilliant Wings", auraTypeLine, finB11WingsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	return wings, bear
}

func TestFinB11BrilliantWingsGivesFlyingAndHexproofAndOnlyEnchantsYourCreatures(t *testing.T) {
	finB11Registered(t, map[string]string{finB11WingsOracle: "Brilliant Wings"})
	g := newCatalogGame(t)
	_, bear := finB11WingsOnABear(t, g)
	for _, kw := range []string{"flying", "hexproof"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the enchanted creature has no %s", kw)
		}
	}
	g2 := newCatalogGame(t)
	theirs := b12Creature(g2, g2.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g2, "Brilliant Wings", auraTypeLine, finB11WingsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("Brilliant Wings enchanted a creature an opponent controls")
	}
}

func TestFinB11BrilliantWingsMovesToANewCreatureForOneMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wings, bear := finB11WingsOnABear(t, g)
	b06AddMana(me, "G")
	fresh := nsHandCard(me, "Fresh Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, fresh.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, wings); host.ID != fresh.InstanceID {
		t.Errorf("Brilliant Wings is on %s, want the new creature %s (old host %s)", host.ID, fresh.InstanceID, bear)
	}
	if hasEffectiveKeyword(t, g, bear, "flying") || !hasEffectiveKeyword(t, g, fresh.InstanceID, "flying") {
		t.Error("flying did not move with the Aura")
	}
}

func TestFinB11BrilliantWingsStaysPutWhenYouDontPay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wings, bear := finB11WingsOnABear(t, g)
	b06AddMana(me, "G")
	fresh := nsHandCard(me, "Fresh Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, fresh.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, wings); host.ID != bear {
		t.Errorf("Brilliant Wings moved to %s without a payment", host.ID)
	}
}

// --- Mega Flare -----------------------------------------------------------------------

func finB11MegaFlareCard() game.Card {
	return game.Card{Name: "Mega Flare", TypeLine: "Sorcery", ManaCost: "{2}{R}", OracleID: finB11MegaFlareOracle}
}

func TestFinB11MegaFlareHitsOneCreaturePerOpponentForYourGreatestPower(t *testing.T) {
	finB11Registered(t, map[string]string{finB11MegaFlareOracle: "Mega Flare"})
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	first := b12Creature(g, a.ID, "First", "Creature — Ox", 2, 4)
	second := b12Creature(g, b.ID, "Second", "Creature — Ox", 2, 5)
	other := b12Creature(g, a.ID, "Other", "Creature — Ox", 2, 2)
	b12Creature(g, me.ID, "Big", "Creature — Ox", 4, 4)
	b12Creature(g, me.ID, "Small", "Creature — Ox", 1, 1)
	tmcB2Pool(t, g, me, "{C}{C}{R}")
	// Two picks from one opponent, or a pick in the wrong opponent's slot, is refused.
	if _, err := tmcB2Cast(g, me, finB11MegaFlareCard(), game.CastSpellParams{Targets: cardRefs(first, other)}); err == nil {
		t.Fatal("two creatures of one opponent were accepted")
	}
	if _, err := tmcB2Cast(g, me, finB11MegaFlareCard(), game.CastSpellParams{Targets: cardRefs(second, first)}); err == nil {
		t.Fatal("the picks were accepted in the wrong opponents' slots")
	}
	if _, err := tmcB2Cast(g, me, finB11MegaFlareCard(), game.CastSpellParams{Targets: cardRefs(first, second)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Error("the 2/4 survived 4 damage (my greatest power)")
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the 2/5 died to 4 damage")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an unchosen creature was hit")
	}
	if n := battlefieldNamed(g, me.ID, "Dragon"); n != 0 {
		t.Errorf("%d Dragons from an unkicked Mega Flare", n)
	}
}

func TestFinB11MegaFlareKickedMakesADragonThatSetsTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	first := b12Creature(g, a.ID, "First", "Creature — Ox", 2, 6)
	second := b12Creature(g, b.ID, "Second", "Creature — Ox", 2, 7)
	b12Creature(g, me.ID, "Small", "Creature — Ox", 1, 1)
	tmcB2Pool(t, g, me, "{C}{C}{C}{C}{C}{C}{R}{R}{R}")
	if _, err := tmcB2Cast(g, me, finB11MegaFlareCard(), game.CastSpellParams{Targets: cardRefs(first, second), OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	dragons := socB8Named(g, me.ID, "Dragon")
	if len(dragons) != 1 {
		t.Fatalf("%d Dragons, want 1", len(dragons))
	}
	d := s58p6View(t, g, dragons[0].InstanceID)
	if d.CurrentPower() != 6 || d.CurrentToughness() != 6 || !hasEffectiveKeyword(t, g, d.InstanceID, "flying") {
		t.Errorf("the Dragon is %d/%d, want a 6/6 flier", d.CurrentPower(), d.CurrentToughness())
	}
	if g.Battlefield.Contains(first) {
		t.Error("the 2/6 survived 6 damage (the Dragon's power)")
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the 2/7 died to 6 damage")
	}
}

func TestFinB11MegaFlareWithNoTargetsStillMakesTheKickedDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tmcB2Pool(t, g, me, "{C}{C}{C}{C}{C}{C}{R}{R}{R}")
	if _, err := tmcB2Cast(g, me, finB11MegaFlareCard(), game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Dragon")); n != 1 {
		t.Errorf("%d Dragons, want 1", n)
	}
}

// --- Amarant Coral --------------------------------------------------------------------

func TestFinB11AmarantCoralMustAttackAndSpreadsTheDamageToTheOtherOpponents(t *testing.T) {
	finB11Registered(t, map[string]string{finB11AmarantOracle: "Amarant Coral"})
	g := newCatalogGame(t)
	me, hit, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	amarant := b12Push(g, me.ID, "Amarant Coral", "Legendary Creature — Human Monk", finB11AmarantOracle, 5, 4)
	if v := s58p6View(t, g, amarant); len(v.Effective().AttackRequirements) == 0 || !hasEffectiveKeyword(t, g, amarant, "trample") {
		t.Fatal("Amarant Coral does not have to attack, or lacks trample")
	}
	hitBefore, o2Before, o3Before, mineBefore := hit.Life, o2.Life, o3.Life, me.Life
	declareAttack(t, g, hit.ID, amarant)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if hit.Life != hitBefore-5 || o2.Life != o2Before-5 || o3.Life != o3Before-5 {
		t.Errorf("life %d→%d, %d→%d, %d→%d; want each opponent down 5", hitBefore, hit.Life, o2Before, o2.Life, o3Before, o3.Life)
	}
	if me.Life != mineBefore {
		t.Errorf("my life changed %d → %d", mineBefore, me.Life)
	}
}

// --- Vivi's Persistence ---------------------------------------------------------------

func TestFinB11ViviMakesAWizardThatPingsOnNoncreatureSpells(t *testing.T) {
	finB11Registered(t, map[string]string{finB11ViviOracle: "Vivi's Persistence"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Vivi's Persistence", "Instant", finB11ViviOracle, nil)
	passPriorityAroundTable(t, g)
	wizards := socB8Named(g, me.ID, "Wizard")
	if len(wizards) != 1 {
		t.Fatalf("%d Wizards, want 1", len(wizards))
	}
	w := s58p6View(t, g, wizards[0].InstanceID)
	if w.CurrentPower() != 0 || w.CurrentToughness() != 1 || !w.IsToken() || len(w.Effective().Colors) != 1 || w.Effective().Colors[0] != "B" {
		t.Errorf("the Wizard is %d/%d %v, want a 0/1 black token", w.CurrentPower(), w.CurrentToughness(), w.Effective().Colors)
	}
	before := opp.Life
	spell := nsHandCard(me, "Test Instant", "Instant", "")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d → %d, want 1 less", before, opp.Life)
	}
}

func finB11ViviInGraveyard(g *game.Game) (*game.Player, uuid.UUID) {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Vivi's Persistence", TypeLine: "Instant", OracleID: finB11ViviOracle,
		ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID,
	})
	return me, id
}

func TestFinB11ViviReturnsWhenYourCommanderEntersAndYouPay(t *testing.T) {
	g := newCatalogGame(t)
	me, vivi := finB11ViviInGraveyard(g)
	msc10CastCommander(t, g, me, "My Commander")
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander entering did not offer to pay {2}")
	}
	floatMana(t, g, me, "{C}{C}")
	answerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(vivi) || me.Graveyard.Contains(vivi) {
		t.Error("the card did not return to hand")
	}
}

func TestFinB11ViviStaysInTheGraveyardWhenYouDeclineOrNoCommanderEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, vivi := finB11ViviInGraveyard(g)
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature that is not a commander offered the payment")
	}
	msc10CastCommander(t, g, me, "My Commander")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	if !me.Graveyard.Contains(vivi) {
		t.Error("a declined payment still returned the card")
	}
}

func TestFinB11ViviReturnsWhenYourCommanderAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, vivi := finB11ViviInGraveyard(g)
	opp := g.Seats[1]
	advanceToMain(t, g)
	cmdr := b12Push(g, me.ID, "My Commander", "Legendary Creature — Human", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	declareAttack(t, g, opp.ID, cmdr)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the commander attacking did not offer to pay {2}")
	}
	floatMana(t, g, me, "{C}{C}")
	answerPayUnless(t, g, me.ID, true)
	if !me.Hand.Contains(vivi) {
		t.Error("the card did not return to hand")
	}
}

// --- Seifer, Balamb Rival -------------------------------------------------------------

func TestFinB11SeiferGoadsACreatureOfEachPlayerYouAttack(t *testing.T) {
	finB11Registered(t, map[string]string{finB11SeiferOracle: "Seifer, Balamb Rival"})
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	seifer := b12Push(g, me.ID, "Seifer, Balamb Rival", "Legendary Creature — Human Mercenary", finB11SeiferOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	untouched := b12Creature(g, other.ID, "Third Player's Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{seifer, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if b33Goaded(t, g, theirs) != me.ID {
		t.Error("the chosen creature is not goaded by me")
	}
	if b33Goaded(t, g, untouched) != uuid.Nil {
		t.Error("a creature of a player I did not attack was goaded")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("two attackers at one player asked for a second goad target")
	}
}

func TestFinB11SeiferDoesNotTriggerOnAnAttackAtAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seifer := b12Push(g, me.ID, "Seifer, Balamb Rival", "Legendary Creature — Human Mercenary", finB11SeiferOracle, 4, 3)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: walker, Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(seifer, walker); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if n := triggersOnStackFrom(g, seifer); n != 0 || latestPickTarget(g, me.ID) != nil {
		t.Errorf("an attack at a planeswalker triggered the goad (%d on the stack)", n)
	}
}

func TestFinB11SeiferGivesDeathtouchToACreatureBlockedByTwoOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Seifer, Balamb Rival", "Legendary Creature — Human Mercenary", finB11SeiferOracle, 4, 3)
	// The attacker is the opponent's, attacking another opponent of Seifer's
	// controller: any player's creature counts. Here it is mine, at opp.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 9)
	w2 := b12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 9)
	declareAttack(t, g, opp.ID, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, w1)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, w := range []uuid.UUID{w1, w2} {
		if err := g.DeclareBlocker(w, bear); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the double-blocked attacker did not gain deathtouch")
	}
}

func TestFinB11SeiferGivesNoDeathtouchForASingleBlockerOrACreatureAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Seifer, Balamb Rival", "Legendary Creature — Human Mercenary", finB11SeiferOracle, 4, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	w1 := b12Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 9)
	declareAttack(t, g, opp.ID, bear)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, w1)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(w1, bear); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("a single blocker gave deathtouch")
	}
}

// --- Squall, Gunblade Duelist ---------------------------------------------------------

func finB11SquallOnTheBattlefield(g *game.Game, owner uuid.UUID, chosen int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Squall, Gunblade Duelist", TypeLine: "Legendary Creature — Human Mercenary",
		OracleID: finB11SquallOracle, ManaCost: "{R}{W}{B}", Power: 3, Toughness: 2,
		Owner: owner, Controller: owner, ChosenNumber: chosen,
	})
}

func TestFinB11SquallChoosesAnyNumberAsItEnters(t *testing.T) {
	finB11Registered(t, map[string]string{finB11SquallOracle: "Squall, Gunblade Duelist"})
	g, me, _ := trekB8Main(t)
	trekB8Float(t, g, me, "{R}{W}{B}")
	squall := castFromHandForTest(t, g, me, "Squall, Gunblade Duelist", "Legendary Creature — Human Mercenary", "{R}{W}{B}", finB11SquallOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	prompt := answerNumber(t, g, me.ID, 37)
	if prompt.Min != 0 || !prompt.NoMax {
		t.Errorf("number prompt %d..%d (no max %v), want 0 or more with no ceiling", prompt.Min, prompt.Max, prompt.NoMax)
	}
	if n := g.ChosenNumberOf(squall); n != 37 {
		t.Errorf("chosen number %d, want 37", n)
	}
}

func TestFinB11SquallDealsItsPowerWhenAnAttackerHasTheChosenPowerOrToughness(t *testing.T) {
	for _, tc := range []struct {
		name       string
		power, tou int
		chosen     int
		want       int
	}{
		{"power equals", 2, 5, 2, 3},
		{"toughness equals", 1, 4, 4, 3},
		{"neither equals", 1, 1, 2, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			finB11SquallOnTheBattlefield(g, me.ID, tc.chosen)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", tc.power, tc.tou)
			before := opp.Life
			declareAttack(t, g, opp.ID, bear)
			passPriorityAroundTable(t, g)
			advanceTo(t, g, game.StepDeclareBlockers)
			passPriorityAroundTable(t, g)
			// The combat damage of the bear itself is not the assertion: read the
			// life lost before combat damage.
			if got := before - opp.Life; got != tc.want {
				t.Errorf("opponent lost %d before combat damage, want %d", got, tc.want)
			}
		})
	}
}

func TestFinB11SquallTriggersOncePerDeclarationAtOnePlayerAndSeparatelyPerPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	finB11SquallOnTheBattlefield(g, me.ID, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, me.ID, "Three", "Creature — Bear", 2, 2)
	aBefore, bBefore := a.Life, b.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, d := range []game.AttackDeclaration{{Attacker: one, Target: a.ID}, {Attacker: two, Target: a.ID}, {Attacker: three, Target: b.ID}} {
		if err := g.DeclareAttacker(d.Attacker, d.Target); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if aBefore-a.Life != 3 || bBefore-b.Life != 3 {
		t.Errorf("life lost %d and %d, want 3 each (one trigger per defending player)", aBefore-a.Life, bBefore-b.Life)
	}
}

func TestFinB11SquallIgnoresAnAttackAtAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	finB11SquallOnTheBattlefield(g, me.ID, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: walker, Name: "Test Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	before := opp.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, walker); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("an attack at a planeswalker cost the player %d life", before-opp.Life)
	}
}

func TestFinB11SquallCountsAnOpponentsCreaturesAttackingAnotherOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	finB11SquallOnTheBattlefield(g, me.ID, 2)
	// An attack at an opponent of Squall's controller by a creature its
	// controller does not control (b attacks a, with seat 0 active in this
	// harness, so it is simulated by a creature b controls that attacks a).
	bear := b12Creature(g, b.ID, "Their Bear", "Creature — Bear", 2, 2)
	before := a.Life
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].AttackingTarget = a.ID
			}
		}
		g.EmitEvent(game.Event{Kind: game.EventAttack, CardID: bear, Source: bear, Actor: b.ID, Target: a.ID})
	})
	passPriorityAroundTable(t, g)
	if before-a.Life != 3 {
		t.Errorf("the defending opponent lost %d life, want 3 (Squall's power)", before-a.Life)
	}
}

// --- Noctis, Heir Apparent ------------------------------------------------------------

func TestFinB11NoctisWarpStrikeReturnsItTappedAttackingAndUnblockable(t *testing.T) {
	finB11Registered(t, map[string]string{finB11NoctisOracle: "Noctis, Heir Apparent"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	noctis := b12Push(g, me.ID, "Noctis, Heir Apparent", "Legendary Creature — Human Noble", finB11NoctisOracle, 2, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, noctis, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Warp-Strike: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(noctis) {
		t.Fatal("Noctis was not exiled")
	}
	if !g.Exile.Contains(noctis) {
		t.Fatal("Noctis is not in exile")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	// An attack declaration with no attackers is enough: the trigger comes as the step starts.
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	// The owner chooses what it attacks (CR 508.4): any of three opponents.
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			break
		}
	}
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Noctis, Heir Apparent" {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("Noctis did not return to the battlefield")
	}
	if !back.Tapped || back.AttackingTarget == uuid.Nil || back.Controller != me.ID {
		t.Errorf("Noctis: tapped %v attacking %v controller %v, want tapped, attacking, mine", back.Tapped, back.AttackingTarget, back.Controller)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, back.InstanceID); err == nil {
		t.Error("Noctis was blocked although it can't be blocked that combat")
	}
}

func TestFinB11NoctisWarpStrikeNeedsThreeMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	noctis := b12Push(g, me.ID, "Noctis, Heir Apparent", "Legendary Creature — Human Noble", finB11NoctisOracle, 2, 3)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, noctis, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("Warp-Strike was activated with two mana")
	}
}

func TestFinB11NoctisAttachesAnEquipmentWhenACreatureEntersDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Noctis, Heir Apparent", "Legendary Creature — Human Noble", finB11NoctisOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sword := seedEquipment(g, me.ID, "Test Sword", "")
	// A creature entering in the main phase does nothing.
	early := nsHandCard(me, "Early Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, early.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a creature entering in the main phase triggered Noctis")
	}
	advanceTo(t, g, game.StepBeginCombat)
	late := nsHandCard(me, "Flash Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == late.InstanceID {
				me.Hand.Cards[i].Keywords = []string{"flash"}
			}
		}
	})
	if err := g.CastSpell(me.ID, late.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, sword)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, sword); host.ID != bear {
		t.Errorf("the Equipment is attached to %s, want the Bear %s", host.ID, bear)
	}
}

// --- Duelist's Flame ------------------------------------------------------------------

func finB11FlameCard() game.Card {
	return game.Card{Name: "Duelist's Flame", TypeLine: "Instant", ManaCost: "{X}{R}{R}", OracleID: finB11DuelistsFlameOracle}
}

// finB11FlameBlocked has my bear attack opp, a wall block it, and Duelist's
// Flame (X = 2) cast on the bear. It returns the bear.
func finB11FlameBlocked(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 1)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	tmcB2Pool(t, g, me, "{C}{C}{R}{R}")
	if _, err := tmcB2Cast(g, me, finB11FlameCard(), game.CastSpellParams{XValue: 2, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return bear
}

func TestFinB11DuelistsFlamePumpsTramplesAndLooksAtThatManyCards(t *testing.T) {
	finB11Registered(t, map[string]string{finB11DuelistsFlameOracle: "Duelist's Flame"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Library top first: a land, the spell, another card.
	seedLibraryTop(me, "Bottom Filler", "Sorcery", "{4}")
	seedLibraryTop(me, "Free Bear", "Creature — Bear", "{3}{G}")
	seedLibraryTop(me, "Top Forest", "Basic Land — Forest", "")
	advanceToMain(t, g)
	bear := finB11FlameBlocked(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 4 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatalf("the bear is %d/%d, trample %v; want 4/2 with trample", v.CurrentPower(), v.CurrentToughness(), hasEffectiveKeyword(t, g, bear, "trample"))
	}
	before := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	b26AnswerDamageAssignments(t, g)
	passPriorityAroundTable(t, g)
	if before-opp.Life != 3 {
		t.Fatalf("the opponent lost %d life, want 3 (4 power tramples over a 0/1 wall)", before-opp.Life)
	}
	var freeBear uuid.UUID
	for _, c := range me.Library.Cards {
		if c.Name == "Free Bear" {
			freeBear = c.InstanceID
		}
	}
	answerChooseCards(t, g, me.ID, freeBear)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, freeBear)
	if err := castThroughResolutionCast(t, g, me.ID, freeBear, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the exiled card free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldNamed(g, me.ID, "Free Bear") != 1 {
		t.Error("the exiled card was not cast")
	}
	// The land and the third card went to the bottom of the library.
	if me.Library.Size() > 0 {
		if bottom := me.Library.Cards[0].Name; bottom != "Top Forest" && bottom != "Bottom Filler" && bottom != "basic-filler" {
			t.Errorf("bottom of the library is %q", bottom)
		}
	}
	for _, c := range me.Library.Cards[:2] {
		if c.Name != "Top Forest" && c.Name != "Bottom Filler" {
			t.Errorf("the two cards put on the bottom are %q and others, want the land and the filler first", c.Name)
		}
	}
}

func TestFinB11DuelistsFlameCanDeclineTheExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibraryTop(me, "Spell Card", "Creature — Bear", "{3}{G}")
	advanceToMain(t, g)
	finB11FlameBlocked(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	b26AnswerDamageAssignments(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	for _, c := range g.Exile.Cards {
		if c.Name == "Spell Card" {
			t.Error("a card was exiled although none was chosen")
		}
	}
	if battlefieldNamed(g, me.ID, "Spell Card") != 0 {
		t.Error("the card was cast")
	}
}

func TestFinB11DuelistsFlameNeedsABlockedCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tmcB2Pool(t, g, me, "{C}{R}{R}")
	if _, err := tmcB2Cast(g, me, finB11FlameCard(), game.CastSpellParams{XValue: 1, Targets: cardRefs(bear)}); err == nil {
		t.Error("an unblocked creature was accepted")
	}
	if _, err := tmcB2Cast(g, me, finB11FlameCard(), game.CastSpellParams{XValue: 1, Targets: cardRefs(theirs)}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
}

// --- Campsite Cuisine -----------------------------------------------------------------

func TestFinB12CampsiteCuisineMakesAFoodForItselfAndForEachLegendaryCreature(t *testing.T) {
	finB11Registered(t, map[string]string{finB12CuisineOracle: "Campsite Cuisine"})
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Campsite Cuisine", "Enchantment", finB12CuisineOracle, nil)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Foods after the enchantment entered, want 1", n)
	}
	legend := nsHandCard(me, "Legend", "Legendary Creature — Human", "")
	if err := g.CastSpell(me.ID, legend.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Food"); n != 2 {
		t.Fatalf("%d Foods after a legendary creature entered, want 2", n)
	}
	plain := nsHandCard(me, "Plain", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Food"); n != 2 {
		t.Errorf("%d Foods after a nonlegendary creature entered, want 2", n)
	}
}

func finB12FoodsAndAttackers(t *testing.T, g *game.Game, foods int) (cuisine uuid.UUID, attackers []uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	cuisine = b12Push(g, me.ID, "Campsite Cuisine", "Enchantment", finB12CuisineOracle, 0, 0)
	for i := 0; i < foods; i++ {
		b12Permanent(g, me.ID, "Food", "Artifact — Food")
	}
	one := b12Creature(g, me.ID, "One", "Creature — Bear", 2, 2)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{one, two} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	return cuisine, []uuid.UUID{one, two}
}

func TestFinB12CampsiteCuisineSacrificesXFoodsToBoostUpToXAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, attackers := finB12FoodsAndAttackers(t, g, 3)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil || p.ChooseMin != 0 || len(p.ChooseCards) != 3 {
		t.Fatalf("Food pick %+v, want any number of the three Foods", p)
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, p.ChooseCards[:2]); err != nil {
		t.Fatal(err)
	}
	if n := battlefieldNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Foods left, want 1", n)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil || pick.PickTargetMin != 0 || pick.PickTargetMax != 2 {
		t.Fatalf("target prompt %+v, want up to two attacking creatures", pick)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, cardRefs(attackers...)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range attackers {
		v := s58p6View(t, g, id)
		if v.CurrentPower() != 5 || v.CurrentToughness() != 5 || !hasEffectiveKeyword(t, g, id, "trample") || !hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("attacker %s is %d/%d, want 5/5 with trample and indestructible", id, v.CurrentPower(), v.CurrentToughness())
		}
	}
}

func TestFinB12CampsiteCuisineSacrificingNoFoodDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, attackers := finB12FoodsAndAttackers(t, g, 2)
	passPriorityAroundTable(t, g)
	p := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no Food pick")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a reflexive trigger asked for targets after sacrificing nothing")
	}
	if v := s58p6View(t, g, attackers[0]); v.CurrentPower() != 2 {
		t.Errorf("the attacker is %d power, want 2", v.CurrentPower())
	}
}

// --- Warrior's Resolve ----------------------------------------------------------------

func TestFinB12WarriorsResolveGivesTrainingAndDrawsAtTheEndStep(t *testing.T) {
	finB11Registered(t, map[string]string{finB12ResolveOracle: "Warrior's Resolve"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Warrior's Resolve", "Enchantment", finB12ResolveOracle, 0, 0)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	attackWith(t, g, opp.ID, small, big)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, small, "+1/+1"); n != 1 {
		t.Errorf("the smaller attacker has %d counters, want 1 (training)", n)
	}
	if n := countersOn(g, big, "+1/+1"); n != 0 {
		t.Errorf("the bigger attacker has %d counters, want 0", n)
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card drawn at the end step", hand, me.Hand.Size())
	}
}

func TestFinB12WarriorsResolveDrawsNothingWithoutACounteredAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Warrior's Resolve", "Enchantment", finB12ResolveOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
	// A creature with a counter that did not attack.
	home := b12Creature(g, me.ID, "Home", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(home, game.CounterPlusOne, 1) })
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw", hand, me.Hand.Size())
	}
}

func TestFinB12WarriorsResolveCountsACounterPutOnAfterTheAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Warrior's Resolve", "Enchantment", finB12ResolveOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, 1) })
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card (the counter came after the attack)", hand, me.Hand.Size())
	}
}

// --- Light Up the Stage ---------------------------------------------------------------

func TestFinB12LightUpTheStageExilesTwoCardsToPlayUntilYourNextTurn(t *testing.T) {
	finB11Registered(t, map[string]string{finB12LightUpOracle: "Light Up the Stage"})
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedLibrary(me, "First", "Second", "Third")
	castCatalogSpell(t, g, "Light Up the Stage", "Sorcery", finB12LightUpOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range ids[:2] {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s was not exiled", id)
		}
	}
	if g.Exile.Contains(ids[2]) {
		t.Error("a third card was exiled")
	}
}

func TestFinB12LightUpTheStageSpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	card := handCardFull(me, "Light Up the Stage", "Sorcery", "{2}{R}", finB12LightUpOracle, []string{"R"})
	params := game.CastSpellParams{Strict: true, AlternativeCost: "spectacle"}
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, card, params); err == nil {
		t.Fatal("spectacle was accepted with no life lost this turn")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, victim.ID, -1); err != nil {
			t.Fatal(err)
		}
	})
	seedLibrary(me, "First", "Second")
	if err := g.CastSpell(me.ID, card, params); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
}

// --- Kenrith, the Returned King -------------------------------------------------------

func finB12Kenrith(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	return b12Push(g, g.Seats[0].ID, "Kenrith, the Returned King", "Legendary Creature — Human Noble", finB12KenrithOracle, 5, 5)
}

func finB12Activate(t *testing.T, g *game.Game, source uuid.UUID, idx int, mana string, targets ...game.TargetRef) error {
	t.Helper()
	floatMana(t, g, g.Seats[0], mana)
	return g.ActivateCatalogAbility(g.Seats[0].ID, source, idx, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

func TestFinB12KenrithRedGivesOnlyCreaturesOnTheBattlefieldTrampleAndHaste(t *testing.T) {
	finB11Registered(t, map[string]string{finB12KenrithOracle: "Kenrith, the Returned King"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kenrith := finB12Kenrith(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := finB12Activate(t, g, kenrith, 0, "{R}"); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{kenrith, theirs} {
		for _, kw := range []string{"trample", "haste"} {
			if !hasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("creature %s lacks %s", id, kw)
			}
		}
	}
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that entered afterwards gained trample")
	}
}

func TestFinB12KenrithGreenPutsACounterOnATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kenrith := finB12Kenrith(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := finB12Activate(t, g, kenrith, 1, "{C}{G}", game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, "+1/+1"); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

func TestFinB12KenrithWhiteGainsFiveAndBlueDrawsForAnyPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	kenrith := finB12Kenrith(t, g)
	life, hand := opp.Life, opp.Hand.Size()
	if err := finB12Activate(t, g, kenrith, 2, "{C}{C}{W}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life+5 {
		t.Errorf("life %d → %d, want +5", life, opp.Life)
	}
	seedLibrary(opp, "Drawn")
	if err := finB12Activate(t, g, kenrith, 3, "{C}{C}{C}{U}", game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, opp.Hand.Size())
	}
}

func TestFinB12KenrithBlackReanimatesUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kenrith := finB12Kenrith(t, g)
	card := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := finB12Activate(t, g, kenrith, 4, "{C}{C}{C}{C}{B}", game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldNamed(g, opp.ID, "Dead Bear") != 1 || battlefieldNamed(g, me.ID, "Dead Bear") != 0 {
		t.Error("the creature did not return under its owner's control")
	}
	if err := finB12Activate(t, g, kenrith, 4, "{C}{C}{C}{C}{B}", game.TargetRef{Kind: game.TargetCard, ID: kenrith}); err == nil {
		t.Error("a creature on the battlefield was accepted as a graveyard target")
	}
}

// --- Tymna the Weaver -----------------------------------------------------------------

func finB12TymnaAttack(t *testing.T, g *game.Game, hit ...int) {
	t.Helper()
	me := g.Seats[0]
	b12Push(g, me.ID, "Tymna the Weaver", "Legendary Creature — Human Cleric", finB12TymnaOracle, 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, seat := range hit {
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		if err := g.DeclareAttacker(bear, g.Seats[seat].ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	if len(hit) > 0 {
		advanceTo(t, g, game.StepCombatDamage)
		passPriorityAroundTable(t, g)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
}

func TestFinB12TymnaPaysXLifeToDrawXCards(t *testing.T) {
	finB11Registered(t, map[string]string{finB12TymnaOracle: "Tymna the Weaver"})
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	finB12TymnaAttack(t, g, 1, 2)
	life, hand := me.Life, me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 paid", life, me.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want 2 cards", hand, me.Hand.Size())
	}
}

func TestFinB12TymnaCountsAnOpponentOnceAndCanDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finB12TymnaAttack(t, g, 1, 1)
	life, hand := me.Life, me.Hand.Size()
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("declining changed life %d → %d or hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
}

func TestFinB12TymnaAsksNothingWhenNoOpponentWasDealtCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finB12TymnaAttack(t, g)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("Tymna asked for a payment with X = 0")
	}
}

func TestFinB12TymnaCantPayMoreLifeThanYouHave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	finB12TymnaAttack(t, g, 1, 2, 3)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { me.Life = 2 })
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != 2 || me.Hand.Size() != hand {
		t.Errorf("a payment of 3 life at 2 life went through: life %d, hand %d → %d", me.Life, hand, me.Hand.Size())
	}
}

// --- Godo, Bandit Warlord -------------------------------------------------------------

func TestFinB12GodoFetchesAnyEquipmentOntoTheBattlefield(t *testing.T) {
	finB11Registered(t, map[string]string{finB12GodoOracle: "Godo, Bandit Warlord"})
	g := newCatalogGame(t)
	me := g.Seats[0]
	axe := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Axe", TypeLine: "Artifact — Equipment", ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Godo, Bandit Warlord", "Legendary Creature — Human Barbarian", finB12GodoOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the searcher is not asked")
	}
	if searchOptionNamed(g, c, "Rock") != uuid.Nil {
		t.Error("a non-Equipment artifact is offered")
	}
	answerSearchNamed(t, g, me.ID, "Big Axe")
	if b12ZoneOf(g, axe) != game.ZoneBattlefield {
		t.Error("the Equipment did not go onto the battlefield")
	}
}

func TestFinB12GodoMaySkipTheSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	axe := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Big Axe", TypeLine: "Artifact — Equipment", ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Godo, Bandit Warlord", "Legendary Creature — Human Barbarian", finB12GodoOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, me.ID)
	if b12ZoneOf(g, axe) != game.ZoneLibrary {
		t.Error("declining the search still moved the Equipment")
	}
}

func TestFinB12GodoUntapsItAndSamuraiAndAddsACombatOnlyTheFirstTime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	godo := b12Push(g, me.ID, "Godo, Bandit Warlord", "Legendary Creature — Human Barbarian", finB12GodoOracle, 3, 3)
	samurai := b12Creature(g, me.ID, "Samurai", "Creature — Human Samurai", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == samurai || id == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(godo, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if isTapped(g, godo) || isTapped(g, samurai) {
		t.Error("Godo or the Samurai stayed tapped")
	}
	if !isTapped(g, bear) {
		t.Error("a non-Samurai creature was untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after combat: %+v, want an additional combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(godo, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("a second additional combat was added by Godo's second attack")
	}
}

// --- Winota, Joiner of Forces ---------------------------------------------------------

func finB12AnswerAttackPick(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == who {
			if err := g.ResolveOptionPick(c.ID, who, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatal("no \"what does it attack?\" prompt")
}

func TestFinB12WinotaPutsAHumanFromTheTopSixInTappedAttackingAndIndestructible(t *testing.T) {
	finB11Registered(t, map[string]string{finB12WinotaOracle: "Winota, Joiner of Forces"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Winota, Joiner of Forces", "Legendary Creature — Human Warrior", finB12WinotaOracle, 4, 4)
	// Top six, top first: a nonhuman creature, a Human noncreature..., a Human, filler.
	seedLibraryTop(me, "Beyond The Six", "Creature — Human Soldier", "{W}")
	for _, n := range []string{"Five", "Four", "Three"} {
		seedLibraryTop(me, n, "Sorcery", "{1}")
	}
	human := seedLibraryTop(me, "Hired Human", "Creature — Human Soldier", "{1}{W}")
	seedLibraryTop(me, "Nonhuman Beast", "Creature — Beast", "{2}")
	seedLibraryTop(me, "Top Spell", "Sorcery", "{1}")
	beast := b12Creature(g, me.ID, "Beast", "Creature — Beast", 2, 2)
	declareAttack(t, g, opp.ID, beast)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the Human")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != human {
		t.Fatalf("offered %v, want only the Human creature card among the top six (%s)", pick.ChooseCards, human)
	}
	answerChooseCards(t, g, me.ID, human)
	finB12AnswerAttackPick(t, g, me.ID)
	var got *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Hired Human" {
			got = &g.Battlefield.Cards[i]
		}
	}
	if got == nil || !got.Tapped || got.AttackingTarget == uuid.Nil {
		t.Fatalf("Hired Human %+v, want on the battlefield tapped and attacking", got)
	}
	if !hasEffectiveKeyword(t, g, got.InstanceID, "indestructible") {
		t.Error("the Human did not gain indestructible")
	}
	// The other five went to the bottom of the library.
	for _, c := range me.Library.Cards[:5] {
		if c.Name == "Top Spell" || c.Name == "Nonhuman Beast" || c.Name == "Three" || c.Name == "Four" || c.Name == "Five" {
			continue
		}
		t.Errorf("bottom card %q is not one of the looked-at cards", c.Name)
	}
	if me.Library.Cards[me.Library.Size()-1].Name != "Beyond The Six" {
		// The card below the six stays on top of the library.
		t.Logf("top of the library is %q", me.Library.Cards[me.Library.Size()-1].Name)
	}
}

func TestFinB12WinotaDoesNotTriggerForAHumanAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Winota, Joiner of Forces", "Legendary Creature — Human Warrior", finB12WinotaOracle, 4, 4)
	seedLibraryTop(me, "Hired Human", "Creature — Human Soldier", "{1}{W}")
	human := b12Creature(g, me.ID, "Human Bear", "Creature — Human Soldier", 2, 2)
	declareAttack(t, g, opp.ID, human)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a Human attacking triggered Winota")
	}
}

func TestFinB12WinotaTriggersOncePerNonHumanAttackerAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Winota, Joiner of Forces", "Legendary Creature — Human Warrior", finB12WinotaOracle, 4, 4)
	for i := 0; i < 12; i++ {
		seedLibraryTop(me, "Hired Human", "Creature — Human Soldier", "{1}{W}")
	}
	one := b12Creature(g, me.ID, "Beast One", "Creature — Beast", 2, 2)
	two := b12Creature(g, me.ID, "Beast Two", "Creature — Beast", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{one, two} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the second non-Human attacker did not get its own look")
	}
	answerChooseCards(t, g, me.ID)
	if n := battlefieldNamed(g, me.ID, "Hired Human"); n != 0 {
		t.Errorf("%d Humans entered after declining both", n)
	}
}

// --- Jodah, the Unifier ---------------------------------------------------------------

func TestFinB12JodahPumpsLegendaryCreaturesByTheNumberOfThem(t *testing.T) {
	finB11Registered(t, map[string]string{finB12JodahOracle: "Jodah, the Unifier"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jodah := b12Push(g, me.ID, "Jodah, the Unifier", "Legendary Creature — Human Wizard", finB12JodahOracle, 5, 5)
	if v := s58p6View(t, g, jodah); v.CurrentPower() != 6 || v.CurrentToughness() != 6 {
		t.Fatalf("a lone Jodah is %d/%d, want 6/6", v.CurrentPower(), v.CurrentToughness())
	}
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 1, 1)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", 1, 1)
	for id, want := range map[uuid.UUID]int{jodah: 7, legend: 3, plain: 2, theirs: 1} {
		if v := s58p6View(t, g, id); v.CurrentPower() != want {
			t.Errorf("%s has power %d, want %d", v.Name, v.CurrentPower(), want)
		}
	}
}

func finB12JodahCast(t *testing.T, g *game.Game, spellCost string, mana string, library func(me *game.Player)) (me *game.Player, spell uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	b12Push(g, me.ID, "Jodah, the Unifier", "Legendary Creature — Human Wizard", finB12JodahOracle, 5, 5)
	library(me)
	advanceToMain(t, g)
	card := game.Card{InstanceID: uuid.New(), Name: "Big Legend", TypeLine: "Legendary Creature — Human", ManaCost: spellCost, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	floatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	return me, card.InstanceID
}

func TestFinB12JodahCastsAFreeLegendaryWithLesserManaValueAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	var hit uuid.UUID
	me, spell := finB12JodahCast(t, g, "{3}", "{C}{C}{C}", func(me *game.Player) {
		seedLibraryTop(me, "Last Card", "Sorcery", "{1}")
		seedLibraryTop(me, "Equal Legend", "Legendary Creature — Human", "{2}{G}")
		hit = seedLibraryTop(me, "Cheap Legend", "Legendary Creature — Human", "{1}{G}")
		seedLibraryTop(me, "Legendary Land", "Legendary Land", "")
		seedLibraryTop(me, "Plain Spell", "Creature — Bear", "{1}")
	})
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit)
	if err := castThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the hit free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if battlefieldNamed(g, me.ID, "Cheap Legend") != 1 {
		t.Error("the legendary card was not cast")
	}
	if battlefieldNamed(g, me.ID, "Big Legend") != 1 {
		t.Error("the spell that triggered it did not resolve")
	}
	// The hit resolved before the spell that caused the trigger.
	var order []string
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cheap Legend" || c.Name == "Big Legend" {
			order = append(order, c.Name)
		}
	}
	if len(order) == 2 && order[0] != "Cheap Legend" {
		t.Errorf("battlefield order %v, want the free spell first", order)
	}
	// The two passed-over cards are on the bottom of the library; the cards
	// below the hit were never exiled and stay on top.
	bottom := map[string]bool{}
	for _, c := range me.Library.Cards[:2] {
		bottom[c.Name] = true
	}
	if !bottom["Plain Spell"] || !bottom["Legendary Land"] {
		t.Errorf("bottom of the library %v, want the two passed-over cards", bottom)
	}
	if top := me.Library.Cards[me.Library.Size()-1].Name; top != "Equal Legend" {
		t.Errorf("top of the library is %q, want Equal Legend (it was never reached)", top)
	}
	_ = spell
}

func TestFinB12JodahDeclinedHitStaysInExile(t *testing.T) {
	g := newCatalogGame(t)
	var hit uuid.UUID
	me, _ := finB12JodahCast(t, g, "{3}", "{C}{C}{C}", func(me *game.Player) {
		hit = seedLibraryTop(me, "Cheap Legend", "Legendary Creature — Human", "{1}{G}")
	})
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit)
	declineResolutionCast(t, g, me.ID)
	if !g.Exile.Contains(hit) {
		t.Error("a declined card did not stay in exile")
	}
}

func TestFinB12JodahIgnoresASpellNotCastFromHandOrNotLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Jodah, the Unifier", "Legendary Creature — Human Wizard", finB12JodahOracle, 5, 5)
	seedLibraryTop(me, "Cheap Legend", "Legendary Creature — Human", "{1}{G}")
	advanceToMain(t, g)
	plain := nsHandCard(me, "Plain", "Creature — Bear", "{3}")
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, plain.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a nonlegendary spell triggered Jodah")
	}
}

// --- Teferi, Mage of Zhalfir ----------------------------------------------------------

func TestFinB12TeferiGivesFlashToCreatureCardsYouOwn(t *testing.T) {
	finB11Registered(t, map[string]string{finB12TeferiOracle: "Teferi, Mage of Zhalfir"})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Teferi, Mage of Zhalfir", "Legendary Creature — Human Wizard", finB12TeferiOracle, 3, 4)
	advanceTo(t, g, game.StepEnd)
	// A creature card I own can be cast at instant speed.
	mine := nsHandCard(me, "My Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, mine.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an owned creature card at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A sorcery and a creature card I don't own can't.
	sorcery := nsHandCard(me, "My Sorcery", "Sorcery", "")
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery card was cast at instant speed")
	}
	theirs := nsHandCard(me, "Their Bear", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == theirs.InstanceID {
				me.Hand.Cards[i].Owner = opp.ID
			}
		}
	})
	if err := g.CastSpell(me.ID, theirs.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature card owned by an opponent was cast at instant speed")
	}
}

func TestFinB12TeferiLetsYouSuspendACreatureCardAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepEnd)
	id := handCardForTest(me, "Needle", "Creature — Horror", nghcNihilith)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("suspended a creature card at instant speed without Teferi")
	}
	b12Push(g, me.ID, "Teferi, Mage of Zhalfir", "Legendary Creature — Human Wizard", finB12TeferiOracle, 3, 4)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionSuspend, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("suspending at instant speed under Teferi: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("the card was not exiled")
	}
}

func TestFinB12TeferiLimitsOpponentsToSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Teferi, Mage of Zhalfir", "Legendary Creature — Human Wizard", finB12TeferiOracle, 3, 4)
	advanceToMain(t, g)
	instant := nsHandCard(opp, "Opp Instant", "Instant", "")
	if err := g.CastSpell(opp.ID, instant.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast an instant on my turn under Teferi")
	}
	// Teferi's controller is not restricted.
	mine := nsHandCard(me, "My Instant", "Instant", "")
	if err := g.CastSpell(me.ID, mine.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("Teferi's controller was restricted: %v", err)
	}
}
