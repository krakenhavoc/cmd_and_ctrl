package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shirei_a_test.go — the Shirei With The Demons Of Shizo deck request
// (#3173), slice A: the cards existing seams carry, plus the Village
// Rites review.

const (
	shReaperOracle      = "d3a50d73-3a1f-43aa-ab97-c6c0031bb734"
	shDreadfeastOracle  = "aa364a33-7c72-4096-b8ed-ca9d5515c3fe"
	shEaterOracle       = "3055381a-fb6b-47a3-8e77-a44768808e27"
	shKravOracle        = "316d49a0-620b-4942-a5e8-04cff15e6c01"
	shApostleOracle     = "595a15f0-77f3-4544-8acc-10630e12cc14"
	shExtractorOracle   = "1fb5c014-9366-4f9e-9728-303ff766002b"
	shDesecrationOracle = "ef75be00-1a88-47a1-a1ae-fe2c9881a798"
	shObNixilisOracle   = "73021bd2-3b60-4fe4-8858-6eb525be1422"
	shPlagueDroneOracle = "9552fc97-9e6d-4307-bf58-30bf956356bf"
	shArdynOracle       = "9ae13026-960a-4d31-b775-d47209a1e313"
	shArchfiendOracle   = "6beec974-891f-4354-8621-ac403c07509a"
	shShireiOracle      = "7b5c3d28-6d8b-488b-b088-567b24faadc7"
	shVillageRitesOracl = "365548fb-5acc-4a8a-b20b-26d28b7d029f"
)

func TestShireiSliceAIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		shReaperOracle:             "Reaper from the Abyss",
		shDreadfeastOracle:         "Dreadfeast Demon",
		shEaterOracle:              "Eater of Hope",
		shKravOracle:               "Krav, the Unredeemed",
		shApostleOracle:            "Shadowborn Apostle",
		shExtractorOracle:          "Extractor Demon",
		westvaleAbbeyOracle:        "Westvale Abbey",
		westvaleAbbeyOracle + "#1": "Ormendahl, Profane Prince",
		shDesecrationOracle:        "Desecration Demon",
		shObNixilisOracle:          "Ob Nixilis, Unshackled",
		shPlagueDroneOracle:        "Plague Drone",
		shArdynOracle:              "Ardyn, the Usurper",
		shArchfiendOracle:          "Archfiend of Spite",
		shShireiOracle:             "Shirei, Shizo's Caretaker",
		shVillageRitesOracl:        "Village Rites",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
	if spec, _ := Lookup(shArchfiendOracle); spec.Madness != "{3}{B}{B}" {
		t.Errorf("Archfiend of Spite madness = %q", spec.Madness)
	}
}

func shInZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == kind
}

// Reaper from the Abyss: no death, no trigger; a death this turn
// destroys a non-Demon at the end step, and a Demon is not a target.
func TestReaperFromTheAbyssDestroysANonDemonAfterADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Reaper from the Abyss", "Creature — Demon", shReaperOracle, 6, 6)
	demon := b12Creature(g, opp.ID, "Their Demon", "Creature — Demon", 5, 5)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	fodder := b12Creature(g, opp.ID, "Fodder", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepEnd)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step after a creature died")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: demon}); err == nil {
		t.Error("a Demon was accepted as the target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived the Reaper")
	}
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon was destroyed")
	}
}

func TestReaperFromTheAbyssDoesNothingWithoutADeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Reaper from the Abyss", "Creature — Demon", shReaperOracle, 6, 6)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the Reaper triggered with no death this turn")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear died")
	}
}

// Dreadfeast Demon: at your end step a non-Demon is sacrificed and a
// copy of the Demon is made; with nothing to sacrifice, no copy.
func TestDreadfeastDemonEatsANonDemonAndCopiesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dreadfeast Demon", "Creature — Demon", shDreadfeastOracle, 6, 6)
	other := b12Creature(g, me.ID, "My Demon", "Creature — Demon", 3, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a Demon was sacrificed")
	}
	if n := onBattlefieldNamed(g, "Dreadfeast Demon"); n != 2 {
		t.Errorf("Dreadfeast Demons on the battlefield = %d, want 2 (the original and its copy)", n)
	}
}

func TestDreadfeastDemonWithNothingToEatMakesNoCopy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dreadfeast Demon", "Creature — Demon", shDreadfeastOracle, 6, 6)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Dreadfeast Demon"); n != 1 {
		t.Errorf("Dreadfeast Demons = %d, want 1", n)
	}
}

// Eater of Hope: two OTHER creatures pay for the destroy; the Eater
// itself can't be one of them.
func TestEaterOfHopeSacrificesTwoOthersToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	eater := b12Push(g, me.ID, "Eater of Hope", "Creature — Demon", shEaterOracle, 6, 4)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)

	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, eater, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{eater, a},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err == nil {
		t.Fatal("the Eater paid with itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, eater, 1, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the victim and the two sacrificed creatures should all be gone")
	}
	if !g.Battlefield.Contains(eater) {
		t.Error("the Eater left")
	}
}

// Krav: X creatures sacrificed, the target player draws X and gains X,
// Krav gets X counters.
func TestKravSacrificesXForCardsLifeAndCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me, ally := g.Seats[0], g.Seats[1]
	krav := b12Push(g, me.ID, "Krav, the Unredeemed", "Legendary Creature — Demon", shKravOracle, 3, 3)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	hand, life := ally.Hand.Size(), ally.Life

	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, krav, 0, game.ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{a, b},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: ally.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := ally.Hand.Size() - hand; got != 2 {
		t.Errorf("target drew %d, want 2", got)
	}
	if got := ally.Life - life; got != 2 {
		t.Errorf("target gained %d, want 2", got)
	}
	if got := counterCount(g, krav, game.CounterPlusOne); got != 2 {
		t.Errorf("Krav's counters = %d, want 2", got)
	}
}

// Shadowborn Apostle: six Apostles go, a Demon comes out of the library.
func TestShadowbornApostleSixFetchADemon(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	var apostles []uuid.UUID
	for i := 0; i < 6; i++ {
		apostles = append(apostles, b12Push(g, me.ID, "Shadowborn Apostle", "Creature — Human Cleric", shApostleOracle, 1, 1))
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	demon := pushLibraryCardForTest(me, game.Card{Name: "Library Demon", TypeLine: "Creature — Demon", Power: 5, Toughness: 5})

	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, apostles[0], 0, game.ActivateAbilityParams{
		SacrificeIDs: append(append([]uuid.UUID(nil), apostles[:5]...), bear),
	}); err == nil {
		t.Fatal("a creature not named Shadowborn Apostle paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, apostles[0], 0, game.ActivateAbilityParams{
		SacrificeIDs: apostles,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, demon)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon is not on the battlefield")
	}
	for _, id := range apostles {
		if g.Battlefield.Contains(id) {
			t.Error("an Apostle stayed")
		}
	}
}

// Extractor Demon: another creature leaving makes a target player mill
// two, on a yes.
func TestExtractorDemonMillsWhenAnotherCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Extractor Demon", "Creature — Demon", shExtractorOracle, 5, 5)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := opp.Library.Size()

	castCatalogSpell(t, g, "Unsummon", "Instant", "837182db-1bf3-4a2c-bd01-1af9d9873561",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

// Westvale Abbey: the Cleric costs 1 life; five creatures flip it into
// an untapped Ormendahl.
func TestWestvaleAbbeyMakesClericsAndFlipsIntoOrmendahl(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: westvaleAbbeyOracle, Layout: game.LayoutTransform,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Westvale Abbey", TypeLine: "Land"},
			{Name: "Ormendahl, Profane Prince", TypeLine: "Legendary Creature — Demon", Colors: []string{"B"}, Power: 9, Toughness: 7},
		},
	}
	c.SetFace(0)
	abbey := pushBattlefieldCardWithTimestamp(g, c)
	life := me.Life

	fillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, abbey, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Cleric ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if n := onBattlefieldNamed(g, "Human Cleric"); n != 1 {
		t.Errorf("Human Clerics = %d, want 1", n)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(abbey) })
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		five = append(five, b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	fillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, abbey, 1, game.ActivateAbilityParams{SacrificeIDs: five}); err != nil {
		t.Fatalf("activate the transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	got := rvBattlefieldCard(g, abbey)
	if got == nil || got.Name != "Ormendahl, Profane Prince" {
		t.Fatalf("the Abbey did not transform: %+v", got)
	}
	if got.Tapped {
		t.Error("Ormendahl is tapped; the ability untaps it")
	}
	abilities := effectiveAbilities(t, g, abbey)
	for _, kw := range []string{"flying", "lifelink", "indestructible", "haste"} {
		if !containsString(abilities, kw) {
			t.Errorf("Ormendahl lacks %s: %v", kw, abilities)
		}
	}
}

// Desecration Demon: an opponent's sacrifice taps it and grows it.
func TestDesecrationDemonIsTappedByAnOpponentsSacrifice(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		demon := b12Push(g, me.ID, "Desecration Demon", "Creature — Demon", shDesecrationOracle, 6, 6)
		victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		appAnswerWith(t, g, appOnly(who, victim))
		d := rvBattlefieldCard(g, demon)
		if d == nil {
			t.Fatal("the Demon left")
		}
		if paid {
			if !d.Tapped || counterCount(g, demon, game.CounterPlusOne) != 1 || g.Battlefield.Contains(victim) {
				t.Errorf("paid: tapped=%v counters=%d victim on battlefield=%v",
					d.Tapped, counterCount(g, demon, game.CounterPlusOne), g.Battlefield.Contains(victim))
			}
			continue
		}
		if d.Tapped || counterCount(g, demon, game.CounterPlusOne) != 0 {
			t.Error("unpaid: the Demon was tapped or grew")
		}
	}
}

// Ob Nixilis, Unshackled: an opponent's own search costs them a creature
// and 10 life; another creature dying grows him.
func TestObNixilisUnshackledPunishesASearch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := b12Push(g, me.ID, "Ob Nixilis, Unshackled", "Legendary Creature — Demon", shObNixilisOracle, 4, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	pushLibraryCardForTest(opp, game.Card{Name: "Opponent's Card", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	life := opp.Life

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(opp.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, false)
	})
	answerSearchByID(t, g, opp.ID, opp.Library.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature was not sacrificed")
	}
	if opp.Life != life-10 {
		t.Errorf("opponent life %d → %d, want -10", life, opp.Life)
	}
	if got := counterCount(g, ob, game.CounterPlusOne); got != 1 {
		t.Errorf("Ob Nixilis counters = %d, want 1 for the sacrificed creature", got)
	}
}

// Plague Drone: an opponent's gain is a loss; its controller's is a gain.
func TestPlagueDroneTurnsAnOpponentsGainIntoLoss(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Plague Drone", "Creature — Demon", shPlagueDroneOracle, 3, 3)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 4)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	if opp.Life != theirs-4 {
		t.Errorf("opponent %d → %d, want -4", theirs, opp.Life)
	}
	if me.Life != mine+3 {
		t.Errorf("controller %d → %d, want +3", mine, me.Life)
	}
}

// Ardyn: Demons you control get the keywords; Starscourge exiles a
// creature card from a graveyard and copies it as a 5/5 black Demon.
func TestArdynGrantsDemonsAndReanimatesAsADemon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ardyn := b12Push(g, me.ID, "Ardyn, the Usurper", "Legendary Creature — Elder Human Noble", shArdynOracle, 4, 4)
	demon := b12Creature(g, me.ID, "My Demon", "Creature — Demon", 3, 3)
	theirDemon := b12Creature(g, opp.ID, "Their Demon", "Creature — Demon", 3, 3)
	for _, kw := range []string{"menace", "lifelink", "haste"} {
		if !containsString(effectiveAbilities(t, g, demon), kw) {
			t.Errorf("my Demon lacks %s", kw)
		}
		if containsString(effectiveAbilities(t, g, theirDemon), kw) {
			t.Errorf("an opponent's Demon has %s", kw)
		}
		if containsString(effectiveAbilities(t, g, ardyn), kw) {
			t.Errorf("Ardyn, not a Demon, has %s", kw)
		}
	}

	dead := pushGraveyardCardTyped(opp, "Dead Elf", "Creature — Elf Druid")
	advanceTo(t, g, game.StepBeginCombat)
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !shInZone(g, dead, game.ZoneExile) {
		t.Error("the card was not exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Dead Elf" && c.Controller == me.ID {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy")
	}
	p, tough := effectivePTOf(t, g, token.InstanceID)
	if p != 5 || tough != 5 || !token.HasSubtype("Demon") || token.HasSubtype("Elf") {
		t.Errorf("token = %d/%d %q, want a 5/5 Demon and no Elf", p, tough, token.TypeLine)
	}
	if !containsString(effectiveAbilities(t, g, token.InstanceID), "haste") {
		t.Error("the Demon token lacks Ardyn's haste")
	}
}

// Archfiend of Spite: an opponent's damage costs them life, or that
// many permanents, their choice; the controller's own damage does not
// trigger it.
func TestArchfiendOfSpitePunishesAnOpponentsDamage(t *testing.T) {
	for _, sacrifice := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fiend := b12Push(g, me.ID, "Archfiend of Spite", "Creature — Demon", shArchfiendOracle, 6, 6)
		shooter := b12Creature(g, opp.ID, "Shooter", "Creature — Human", 1, 1)
		x := b12Creature(g, opp.ID, "X", "Creature — Bear", 2, 2)
		y := b12Creature(g, opp.ID, "Y", "Creature — Bear", 2, 2)
		life := opp.Life

		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(shooter, fiend, 2) })
		passPriorityAroundTable(t, g)
		if sacrifice {
			answerOptionPick(t, g, opp.ID, 1)
			answerSacrifice(t, g, opp.ID, x)
			answerSacrifice(t, g, opp.ID, y)
		} else {
			answerOptionPick(t, g, opp.ID, 0)
		}
		passPriorityAroundTable(t, g)
		if sacrifice {
			if opp.Life != life || g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
				t.Errorf("sacrifice: life %d → %d, X on=%v, Y on=%v", life, opp.Life, g.Battlefield.Contains(x), g.Battlefield.Contains(y))
			}
			continue
		}
		if opp.Life != life-2 {
			t.Errorf("life %d → %d, want -2", life, opp.Life)
		}
	}
}

func TestArchfiendOfSpiteIgnoresItsControllersOwnDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fiend := b12Push(g, me.ID, "Archfiend of Spite", "Creature — Demon", shArchfiendOracle, 6, 6)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Human", 1, 1)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(mine, fiend, 2) })
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil || me.Life != life {
		t.Error("the controller's own damage triggered the Archfiend")
	}
}

// Shirei: a small creature of yours dying returns at the end step on a
// yes, while Shirei stays; a big one does not trigger.
func TestShireiReturnsASmallCreatureAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Shirei, Shizo's Caretaker", "Legendary Creature — Spirit", shShireiOracle, 2, 2)
	small := b12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(small)
		_ = g.DestroyPermanentForEffect(big)
	})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the 1/1 did not return")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the 2/2 returned")
	}
}

func TestShireiReturnsNothingOnceSheHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shirei := b12Push(g, me.ID, "Shirei, Shizo's Caretaker", "Legendary Creature — Spirit", shShireiOracle, 2, 2)
	small := b12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(small) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shirei) })
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked to return with Shirei gone")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the creature returned with Shirei gone")
	}
}
