package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestShireiSliceAIsRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.ShReaperOracle:        "Reaper from the Abyss",
		effectstest.ShDreadfeastOracle:    "Dreadfeast Demon",
		effectstest.ShEaterOracle:         "Eater of Hope",
		effectstest.ShKravOracle:          "Krav, the Unredeemed",
		effectstest.ShApostleOracle:       "Shadowborn Apostle",
		effectstest.ShExtractorOracle:     "Extractor Demon",
		cardsw.WestvaleAbbeyOracle:        "Westvale Abbey",
		cardsw.WestvaleAbbeyOracle + "#1": "Ormendahl, Profane Prince",
		effectstest.ShDesecrationOracle:   "Desecration Demon",
		effectstest.ShObNixilisOracle:     "Ob Nixilis, Unshackled",
		effectstest.ShPlagueDroneOracle:   "Plague Drone",
		effectstest.ShArdynOracle:         "Ardyn, the Usurper",
		effectstest.ShArchfiendOracle:     "Archfiend of Spite",
		effectstest.ShShireiOracle:        "Shirei, Shizo's Caretaker",
		effectstest.ShVillageRitesOracl:   "Village Rites",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", name, spec.Completeness, spec.Caveats)
		}
	}
	if spec, _ := effects.Lookup(effectstest.ShArchfiendOracle); spec.Madness != "{3}{B}{B}" {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Reaper from the Abyss", "Creature — Demon", effectstest.ShReaperOracle, 6, 6)
	demon := effectstest.B12Creature(g, opp.ID, "Their Demon", "Creature — Demon", 5, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	fodder := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the end step after a creature died")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: demon}); err == nil {
		t.Error("a Demon was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived the Reaper")
	}
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon was destroyed")
	}
}

func TestReaperFromTheAbyssDoesNothingWithoutADeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Reaper from the Abyss", "Creature — Demon", effectstest.ShReaperOracle, 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the Reaper triggered with no death this turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear died")
	}
}

// Dreadfeast Demon: at your end step a non-Demon is sacrificed and a
// copy of the Demon is made; with nothing to sacrifice, no copy.
func TestDreadfeastDemonEatsANonDemonAndCopiesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dreadfeast Demon", "Creature — Demon", effectstest.ShDreadfeastOracle, 6, 6)
	other := effectstest.B12Creature(g, me.ID, "My Demon", "Creature — Demon", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a Demon was sacrificed")
	}
	if n := effectstest.OnBattlefieldNamed(g, "Dreadfeast Demon"); n != 2 {
		t.Errorf("Dreadfeast Demons on the battlefield = %d, want 2 (the original and its copy)", n)
	}
}

func TestDreadfeastDemonWithNothingToEatMakesNoCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dreadfeast Demon", "Creature — Demon", effectstest.ShDreadfeastOracle, 6, 6)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Dreadfeast Demon"); n != 1 {
		t.Errorf("Dreadfeast Demons = %d, want 1", n)
	}
}

// Eater of Hope: two OTHER creatures pay for the destroy; the Eater
// itself can't be one of them.
func TestEaterOfHopeSacrificesTwoOthersToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	eater := effectstest.B12Push(g, me.ID, "Eater of Hope", "Creature — Demon", effectstest.ShEaterOracle, 6, 4)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)

	effectstest.FillPoolColored(me, "B", 3)
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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, ally := g.Seats[0], g.Seats[1]
	krav := effectstest.B12Push(g, me.ID, "Krav, the Unredeemed", "Legendary Creature — Demon", effectstest.ShKravOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	hand, life := ally.Hand.Size(), ally.Life

	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, krav, 0, game.ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{a, b},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: ally.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := ally.Hand.Size() - hand; got != 2 {
		t.Errorf("target drew %d, want 2", got)
	}
	if got := ally.Life - life; got != 2 {
		t.Errorf("target gained %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, krav, game.CounterPlusOne); got != 2 {
		t.Errorf("Krav's counters = %d, want 2", got)
	}
}

// Shadowborn Apostle: six Apostles go, a Demon comes out of the library.
func TestShadowbornApostleSixFetchADemon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	var apostles []uuid.UUID
	for i := 0; i < 6; i++ {
		apostles = append(apostles, effectstest.B12Push(g, me.ID, "Shadowborn Apostle", "Creature — Human Cleric", effectstest.ShApostleOracle, 1, 1))
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	demon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Library Demon", TypeLine: "Creature — Demon", Power: 5, Toughness: 5})

	effectstest.FillPoolColored(me, "B", 2)
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
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, demon)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Extractor Demon", "Creature — Demon", effectstest.ShExtractorOracle, 5, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := opp.Library.Size()

	effectstest.CastCatalogSpell(t, g, "Unsummon", "Instant", "837182db-1bf3-4a2c-bd01-1af9d9873561",
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

// Westvale Abbey: the Cleric costs 1 life; five creatures flip it into
// an untapped Ormendahl.
func TestWestvaleAbbeyMakesClericsAndFlipsIntoOrmendahl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: cardsw.WestvaleAbbeyOracle, Layout: game.LayoutTransform,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Westvale Abbey", TypeLine: "Land"},
			{Name: "Ormendahl, Profane Prince", TypeLine: "Legendary Creature — Demon", Colors: []string{"B"}, Power: 9, Toughness: 7},
		},
	}
	c.SetFace(0)
	abbey := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	life := me.Life

	effectstest.FillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, abbey, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Cleric ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if n := effectstest.OnBattlefieldNamed(g, "Human Cleric"); n != 1 {
		t.Errorf("Human Clerics = %d, want 1", n)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(abbey) })
	var five []uuid.UUID
	for i := 0; i < 5; i++ {
		five = append(five, effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	effectstest.FillPoolColored(me, "C", 5)
	if err := g.ActivateCatalogAbility(me.ID, abbey, 1, game.ActivateAbilityParams{SacrificeIDs: five}); err != nil {
		t.Fatalf("activate the transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.RvBattlefieldCard(g, abbey)
	if got == nil || got.Name != "Ormendahl, Profane Prince" {
		t.Fatalf("the Abbey did not transform: %+v", got)
	}
	if got.Tapped {
		t.Error("Ormendahl is tapped; the ability untaps it")
	}
	abilities := effectstest.EffectiveAbilities(t, g, abbey)
	for _, kw := range []string{"flying", "lifelink", "indestructible", "haste"} {
		if !effectstest.ContainsString(abilities, kw) {
			t.Errorf("Ormendahl lacks %s: %v", kw, abilities)
		}
	}
}

// Desecration Demon: an opponent's sacrifice taps it and grows it.
func TestDesecrationDemonIsTappedByAnOpponentsSacrifice(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		demon := effectstest.B12Push(g, me.ID, "Desecration Demon", "Creature — Demon", effectstest.ShDesecrationOracle, 6, 6)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, victim))
		d := effectstest.RvBattlefieldCard(g, demon)
		if d == nil {
			t.Fatal("the Demon left")
		}
		if paid {
			if !d.Tapped || effectstest.CounterCount(g, demon, game.CounterPlusOne) != 1 || g.Battlefield.Contains(victim) {
				t.Errorf("paid: tapped=%v counters=%d victim on battlefield=%v",
					d.Tapped, effectstest.CounterCount(g, demon, game.CounterPlusOne), g.Battlefield.Contains(victim))
			}
			continue
		}
		if d.Tapped || effectstest.CounterCount(g, demon, game.CounterPlusOne) != 0 {
			t.Error("unpaid: the Demon was tapped or grew")
		}
	}
}

// Ob Nixilis, Unshackled: an opponent's own search costs them a creature
// and 10 life; another creature dying grows him.
func TestObNixilisUnshackledPunishesASearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := effectstest.B12Push(g, me.ID, "Ob Nixilis, Unshackled", "Legendary Creature — Demon", effectstest.ShObNixilisOracle, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Opponent's Card", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	life := opp.Life

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(opp.ID, func(game.Card) bool { return true }, game.ZoneHand, 1, false, false)
	})
	effectstest.AnswerSearchByID(t, g, opp.ID, opp.Library.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature was not sacrificed")
	}
	if opp.Life != life-10 {
		t.Errorf("opponent life %d → %d, want -10", life, opp.Life)
	}
	if got := effectstest.CounterCount(g, ob, game.CounterPlusOne); got != 1 {
		t.Errorf("Ob Nixilis counters = %d, want 1 for the sacrificed creature", got)
	}
}

// Plague Drone: an opponent's gain is a loss; its controller's is a gain.
func TestPlagueDroneTurnsAnOpponentsGainIntoLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Plague Drone", "Creature — Demon", effectstest.ShPlagueDroneOracle, 3, 3)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ardyn := effectstest.B12Push(g, me.ID, "Ardyn, the Usurper", "Legendary Creature — Elder Human Noble", effectstest.ShArdynOracle, 4, 4)
	demon := effectstest.B12Creature(g, me.ID, "My Demon", "Creature — Demon", 3, 3)
	theirDemon := effectstest.B12Creature(g, opp.ID, "Their Demon", "Creature — Demon", 3, 3)
	for _, kw := range []string{"menace", "lifelink", "haste"} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, demon), kw) {
			t.Errorf("my Demon lacks %s", kw)
		}
		if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirDemon), kw) {
			t.Errorf("an opponent's Demon has %s", kw)
		}
		if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, ardyn), kw) {
			t.Errorf("Ardyn, not a Demon, has %s", kw)
		}
	}

	dead := effectstest.PushGraveyardCardTyped(opp, "Dead Elf", "Creature — Elf Druid")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
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
	p, tough := effectstest.EffectivePTOf(t, g, token.InstanceID)
	if p != 5 || tough != 5 || !token.HasSubtype("Demon") || token.HasSubtype("Elf") {
		t.Errorf("token = %d/%d %q, want a 5/5 Demon and no Elf", p, tough, token.TypeLine)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, token.InstanceID), "haste") {
		t.Error("the Demon token lacks Ardyn's haste")
	}
}

// Archfiend of Spite: an opponent's damage costs them life, or that
// many permanents, their choice; the controller's own damage does not
// trigger it.
func TestArchfiendOfSpitePunishesAnOpponentsDamage(t *testing.T) {
	for _, sacrifice := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fiend := effectstest.B12Push(g, me.ID, "Archfiend of Spite", "Creature — Demon", effectstest.ShArchfiendOracle, 6, 6)
		shooter := effectstest.B12Creature(g, opp.ID, "Shooter", "Creature — Human", 1, 1)
		x := effectstest.B12Creature(g, opp.ID, "X", "Creature — Bear", 2, 2)
		y := effectstest.B12Creature(g, opp.ID, "Y", "Creature — Bear", 2, 2)
		life := opp.Life

		g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(shooter, fiend, 2) })
		effectstest.PassPriorityAroundTable(t, g)
		if sacrifice {
			effectstest.AnswerOptionPick(t, g, opp.ID, 1)
			effectstest.AnswerSacrifice(t, g, opp.ID, x)
			effectstest.AnswerSacrifice(t, g, opp.ID, y)
		} else {
			effectstest.AnswerOptionPick(t, g, opp.ID, 0)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fiend := effectstest.B12Push(g, me.ID, "Archfiend of Spite", "Creature — Demon", effectstest.ShArchfiendOracle, 6, 6)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Human", 1, 1)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(mine, fiend, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil || me.Life != life {
		t.Error("the controller's own damage triggered the Archfiend")
	}
}

// Shirei: a small creature of yours dying returns at the end step on a
// yes, while Shirei stays; a big one does not trigger.
func TestShireiReturnsASmallCreatureAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Shirei, Shizo's Caretaker", "Legendary Creature — Spirit", effectstest.ShShireiOracle, 2, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(small)
		_ = g.DestroyPermanentForEffect(big)
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the 1/1 did not return")
	}
	if g.Battlefield.Contains(big) {
		t.Error("the 2/2 returned")
	}
}

func TestShireiReturnsNothingOnceSheHasLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shirei := effectstest.B12Push(g, me.ID, "Shirei, Shizo's Caretaker", "Legendary Creature — Spirit", effectstest.ShShireiOracle, 2, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(small) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shirei) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked to return with Shirei gone")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the creature returned with Shirei gone")
	}
}
