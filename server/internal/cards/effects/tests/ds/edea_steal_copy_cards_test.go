package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// e2CastOwnedBy puts a card OWNED by `owner` into the active seat's
// hand and casts it — "a spell you don't own".
func e2CastOwnedBy(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// e2HandIDs lists a player's hand.
func e2HandIDs(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

// --- Don Andres, the Renegade ---------------------------------------

const donAndresOracle = "060ef981-db05-436b-b1c1-f55071375344"

func TestDonAndresBuffsCreaturesYouControlButDontOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Don Andres, the Renegade", "Legendary Creature — Vampire Pirate", donAndresOracle, 4, 3)
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	stolen := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	e2Steal(t, g, stolen, me.ID)

	if got := effectstest.EffectivePower(t, g, stolen); got != 4 {
		t.Errorf("stolen creature power %d, want 4 (+2/+2)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, stolen); got != 4 {
		t.Errorf("stolen creature toughness %d, want 4", got)
	}
	for _, kw := range []string{"menace", "deathtouch"} {
		if !effectstest.EffectiveAbilitiesContain(t, g, stolen, kw) {
			t.Errorf("stolen creature lacks %s", kw)
		}
	}
	if !effectstest.HasString(effectstest.EffectiveSubtypes(t, g, stolen), "Pirate") {
		t.Errorf("stolen creature is not a Pirate: %v", effectstest.EffectiveSubtypes(t, g, stolen))
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("a creature you own: power %d, want the printed 2", got)
	}

	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.EffectivePower(t, g, stolen); got != 2 {
		t.Errorf("back with its owner: power %d, want 2", got)
	}
}

func TestDonAndresMakesTreasuresOnlyForANoncreatureSpellYouDontOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Don Andres, the Renegade", "Legendary Creature — Vampire Pirate", donAndresOracle, 4, 3)

	e2CastOwnedBy(t, g, me.ID, "My Opt", "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	e2CastOwnedBy(t, g, opp.ID, "Their Bear", "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.E2Treasures(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures from your own spell and a creature spell, want 0", n)
	}

	e2CastOwnedBy(t, g, opp.ID, "Their Opt", "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := effectstest.E2Treasures(g, me.ID)
	if n != 2 || tapped != 2 {
		t.Errorf("Treasures %d (tapped %d), want two tapped", n, tapped)
	}
}

// --- Yahenni, Undying Partisan ---------------------------------------

const yahenniOracle = "fdba89eb-1cf5-46e6-9d09-1adb9bc40fcd"

func TestYahenniGrowsOnlyWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	yahenni := effectstest.B43Catalog(g, me.ID, "Yahenni, Undying Partisan", "Legendary Creature — Aetherborn Vampire", yahenniOracle, 2, 2)
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	stolen := effectstest.CtrlPushCreature(g, opp.ID, "Stolen Bear")
	e2Steal(t, g, stolen, me.ID)

	effectstest.E2Destroy(t, g, mine)
	effectstest.E2Destroy(t, g, stolen)
	if n := effectstest.CounterOn(g, yahenni, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters from your own creatures dying, want 0 (the stolen one was yours when it died)", n)
	}
	effectstest.E2Destroy(t, g, theirs)
	if n := effectstest.CounterOn(g, yahenni, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after an opponent's creature died, want 1", n)
	}
}

func TestYahenniSacrificesAnotherCreatureForIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	yahenni := effectstest.B43Catalog(g, me.ID, "Yahenni, Undying Partisan", "Legendary Creature — Aetherborn Vampire", yahenniOracle, 2, 2)
	fodder := effectstest.CtrlPushCreature(g, me.ID, "Fodder")

	if err := g.ActivateCatalogAbility(me.ID, yahenni, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{yahenni}}); err == nil {
		t.Fatal("Yahenni sacrificed itself — the cost is ANOTHER creature")
	}
	effectstest.B16Activate(t, g, me.ID, yahenni, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if z := effectstest.E2Zone(g, fodder); z != game.ZoneGraveyard {
		t.Errorf("the sacrificed creature is in %q", z)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, yahenni, "indestructible") {
		t.Error("Yahenni did not gain indestructible")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, yahenni, "haste") {
		t.Error("Yahenni lost its printed haste")
	}
}

// --- Gogo, Master of Mimicry -----------------------------------------

func TestGogoCopiesAnAbilityXTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gogo := effectstest.PushCatalogPermanent(g, me.ID, "Gogo, Master of Mimicry", "Legendary Creature — Wizard", cardsg.GogoMasterOfMimicryOracleID, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	activation := effectstest.AcItemLabelled(g, "an activation")
	if !effectstest.AcLegalAbilityTargets(g, me.ID, cardsg.GogoMasterOfMimicryOracleID, 0)[activation] {
		t.Fatal("an activated ability you control must be offered")
	}

	effectstest.FloatForTest(g, me, "CCCC")
	if err := g.ActivateCatalogAbility(me.ID, gogo, 0, game.ActivateAbilityParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: activation}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility X=2: %v", err)
	}
	for i := 0; i < 16 && effectstest.AcCopiesOnStack(g) < 2; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if n := effectstest.AcCopiesOnStack(g); n != 2 {
		t.Errorf("X=2 made %d copies, want 2", n)
	}
}

func TestGogoRefusesXZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gogo := effectstest.PushCatalogPermanent(g, me.ID, "Gogo, Master of Mimicry", "Legendary Creature — Wizard", cardsg.GogoMasterOfMimicryOracleID, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other", "Artifact", "", false)
	if err := g.ActivateAbility(me.ID, other, game.AbilityParams{Label: "an activation"}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	activation := effectstest.AcItemLabelled(g, "an activation")
	if err := g.ActivateCatalogAbility(me.ID, gogo, 0, game.ActivateAbilityParams{
		XValue:  0,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: activation}},
	}); err == nil {
		t.Fatal("X = 0 was accepted — \"X can't be 0\"")
	}
	// "This ability can't be copied" is #1574's Uncopyable bit; its
	// tests are in gogo_sephiroth_hooks_test.go.
}

func TestSephirothEntersAndMaySacrificeToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	effectstest.ImportAndCast(t, g, effectstest.SephirothRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no sacrifice pick when Sephiroth entered")
	}
	if p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Errorf("pick bounds %d..%d, want 0..1 (\"you MAY sacrifice another creature\")", p.ChooseMin, p.ChooseMax)
	}
	hand := len(me.Hand.Cards)
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	// The bear dying is "another creature dies": the drain targets.
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, bear); z != game.ZoneGraveyard {
		t.Errorf("the sacrificed bear is in %q", z)
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("drew %d after sacrificing, want 1", got)
	}
}

func TestSephirothTransformsOnTheFourthDrainAndMakesTheEmblem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	var victims []uuid.UUID
	for i := 0; i < 5; i++ {
		victims = append(victims, effectstest.CtrlPushCreature(g, opp.ID, "Victim"))
	}
	seph := effectstest.ImportToBattlefield(t, g, effectstest.SephirothRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		if err := g.ResolveOwnPermanents(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the entry sacrifice: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	myLife, theirLife := me.Life, opp.Life

	for i := 0; i < 4; i++ {
		if effects.TransformedPermanent(effectstest.E2Card(t, g, seph)) {
			t.Fatalf("transformed after %d drains, want only on the fourth", i)
		}
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victims[i]) })
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effects.TransformedPermanent(effectstest.E2Card(t, g, seph)) {
		t.Fatal("Sephiroth did not transform on the fourth resolution")
	}
	if me.Emblems == nil || me.Emblems.Size() != 1 {
		t.Fatal("Super Nova: want exactly one emblem")
	}
	if opp.Life != theirLife-4 || me.Life != myLife+4 {
		t.Errorf("life %d/%d, want %d/%d after four drains", me.Life, opp.Life, myLife+4, theirLife-4)
	}

	// Now the emblem drains, and the back face does not.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victims[4]) })
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-5 {
		t.Errorf("the emblem's drain: opponent at %d, want %d", opp.Life, theirLife-5)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, seph, "flying") {
		t.Error("the back face has no flying")
	}
}

func TestHostageTakerExilesUntilItLeavesAndReturnsToTheOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	taker := effectstest.CastAndResolveCreature(t, g, "Hostage Taker", "Creature — Human Pirate", effectstest.HostageTakerOracle)

	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(prompt.PickTargetCards, taker) {
		t.Error("\"ANOTHER target creature\": Hostage Taker is offered itself")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if z := effectstest.E2Zone(g, victim); z != game.ZoneExile {
		t.Fatalf("the creature is in %q, want exile", z)
	}
	perm := effectstest.E2Permission(g, victim)
	if perm == nil || perm.Player != me.ID || !perm.CastOnly || !perm.AnyType {
		t.Fatalf("permission %+v, want a cast-only any-type grant to the Taker's controller", perm)
	}

	effectstest.E2Destroy(t, g, taker)
	// A card returning from exile is a new object with a new instance
	// ID (CR 400.7), so it is found by name.
	back, ok := effectstest.E2BattlefieldNamed(g, "Their Bear")
	if !ok {
		t.Fatal("after the Taker left the creature is not back on the battlefield")
	}
	if back.Controller != opp.ID || back.Owner != opp.ID {
		t.Errorf("returned under %s, want its owner %s", back.Controller, opp.ID)
	}
}

func TestHostageTakerCastingTheCardKeepsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	taker := effectstest.CastAndResolveCreature(t, g, "Hostage Taker", "Creature — Human Pirate", effectstest.HostageTakerOracle)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, victim, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("the cast creature is controlled by %s, want %s", got, me.ID)
	}
	effectstest.E2Destroy(t, g, taker)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("the Taker leaving took back a card that was cast: controller %s", got)
	}
}

// TestHostageTakerThatLeftFirstExilesNothing pins CR 610.3b.
func TestHostageTakerThatLeftFirstExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Bear")
	taker := effectstest.CastAndResolveCreature(t, g, "Hostage Taker", "Creature — Human Pirate", effectstest.HostageTakerOracle)
	effectstest.PickCard(t, g, me.ID, victim)
	// Kill the Taker with its entry trigger still on the stack.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(taker) })
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, victim); z != game.ZoneBattlefield {
		t.Errorf("the creature is in %q; a Taker that has already left exiles nothing", z)
	}
}

// --- Gonti, Night Minister ---------------------------------------------

const gontiOracle = "e16f79c1-ffbb-4893-b62a-d4e9d15e2b16"

func TestGontiExilesTheTopCardForTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, me.ID, "Gonti, Night Minister", "Legendary Creature — Aetherborn Rogue", gontiOracle, 3, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Rogue", 2, 2)
	top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID

	effectstest.AttackWith(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)

	if z := effectstest.E2Zone(g, top); z != game.ZoneExile {
		t.Fatalf("the opponent's top card is in %q, want exile", z)
	}
	perm := effectstest.E2Permission(g, top)
	if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("permission %+v, want a play (not cast-only) any-type grant to you", perm)
	}
	effectstest.E2AssertFaceDownFor(t, g, top, me.ID)
}

// TestGontiHandsTheCardToTheOtherCreaturesController pins the
// symmetric half: an opponent's creature hitting another of your
// opponents gives THAT creature's controller the card, and damage to
// Gonti's own controller does nothing.
func TestGontiHandsTheCardToTheOtherCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, oppA, oppB := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B43Catalog(g, me.ID, "Gonti, Night Minister", "Legendary Creature — Aetherborn Rogue", gontiOracle, 3, 4)
	theirs := effectstest.B12Creature(g, oppA.ID, "Their Rogue", "Creature — Rogue", 2, 2)
	top := oppB.Library.Cards[len(oppB.Library.Cards)-1].InstanceID
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: theirs, Target: oppB.ID, Amount: 2, Combat: true, Actor: oppA.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.E2Permission(g, top)
	if perm == nil || perm.Player != oppA.ID {
		t.Errorf("permission %+v, want the grant to the creature's controller %s", perm, oppA.ID)
	}

	mine := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: theirs, Target: me.ID, Amount: 2, Combat: true, Actor: oppA.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, mine); z != game.ZoneLibrary {
		t.Errorf("damage to Gonti's controller exiled their own top card (zone %q)", z)
	}
}

func TestGontiGivesATreasureToWhoeverCastsASpellTheyDontOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B43Catalog(g, opp.ID, "Gonti, Night Minister", "Legendary Creature — Aetherborn Rogue", gontiOracle, 3, 4)
	e2CastOwnedBy(t, g, me.ID, "My Opt", "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.E2Treasures(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures for casting your own spell, want 0", n)
	}
	e2CastOwnedBy(t, g, opp.ID, "Their Bear", "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.E2Treasures(g, me.ID); n != 1 {
		t.Errorf("%d Treasures for the caster of a spell they don't own, want 1 (to the caster, not Gonti's controller)", n)
	}
	if n, _ := effectstest.E2Treasures(g, opp.ID); n != 0 {
		t.Errorf("Gonti's controller got %d Treasures, want 0", n)
	}
}

// --- Outrageous Robbery -------------------------------------------------

func TestOutrageousRobberyExilesXForYouToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	n := len(opp.Library.Cards)
	top := []uuid.UUID{opp.Library.Cards[n-1].InstanceID, opp.Library.Cards[n-2].InstanceID}
	effectstest.B12PlayFromHand(t, g, "Outrageous Robbery", "Instant", "5b194438-6946-45dd-8d77-c9de8c115d09",
		game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := n - len(opp.Library.Cards); got != 2 {
		t.Errorf("exiled %d cards, want X = 2", got)
	}
	for _, id := range top {
		perm := effectstest.E2Permission(g, id)
		if perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
			t.Errorf("card %s permission %+v, want an any-type play grant to the caster", id, perm)
		}
		effectstest.E2AssertFaceDownFor(t, g, id, me.ID)
	}
}

// TestOutrageousRobberyCardIsCastableWithAnyType casts a stolen {2}{C}
// card off three Swamps — "mana of any type", {C} included (#1573).
func TestOutrageousRobberyCardIsCastableWithAnyType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seer := game.NewCard("Their Seer", opp.ID)
	seer.TypeLine = "Creature — Eldrazi"
	seer.ManaCost = "{2}{C}"
	opp.Library.PushTop(seer)
	effectstest.B12PlayFromHand(t, g, "Outrageous Robbery", "Instant", "5b194438-6946-45dd-8d77-c9de8c115d09",
		game.CastSpellParams{XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.E2AssertFaceDownFor(t, g, seer.InstanceID, me.ID)
	effectstest.E2Swamps(g, me.ID, 3)
	if err := g.CastSpell(me.ID, seer.InstanceID, game.CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("cast the stolen {2}{C} card off three Swamps: %v", err)
	}
	if c := effectstest.E2Card(t, g, seer.InstanceID); c.FaceDown {
		t.Error("the cast card is still face down")
	}
}

// TestHostageTakerCardIsCastableWithAnyType: a stolen {2}{C} creature
// is castable off three Swamps (#1573).
func TestHostageTakerCardIsCastableWithAnyType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Their Seer")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].ManaCost = "{2}{C}"
			}
		}
	})
	effectstest.CastAndResolveCreature(t, g, "Hostage Taker", "Creature — Human Pirate", effectstest.HostageTakerOracle)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.E2Swamps(g, me.ID, 3)
	if err := g.CastSpell(me.ID, victim, game.CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("cast the stolen {2}{C} creature off three Swamps: %v", err)
	}
}

// --- Zara, Renegade Recruiter -------------------------------------------

func TestZaraPutsACreatureFromTheirHandAttackingAndReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zara := effectstest.B43Catalog(g, me.ID, "Zara, Renegade Recruiter", "Legendary Creature — Human Pirate", "cd2720c2-522c-4fdc-9cef-9c5ce250fe7b", 4, 3)
	beast := uuid.New()
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: beast, Name: "Their Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})

	effectstest.DeclareAttack(t, g, opp.ID, zara)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick from the defending player's hand")
	}
	if effectstest.HasID(pick.ChooseCards, spell) || !effectstest.HasID(pick.ChooseCards, beast) {
		t.Errorf("candidates %v, want the creature card only", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 {
		t.Errorf("floor %d, want 0 — \"you MAY put\"", pick.ChooseMin)
	}
	if !effectstest.E2Card(t, g, spell).KnownBy[me.ID] {
		t.Error("Zara's controller did not look at the whole hand")
	}
	if effectstest.E2Card(t, g, spell).KnownBy[g.Seats[2].ID] {
		t.Error("a third player saw the hand — it is a look, not a reveal")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.E2Card(t, g, beast)
	if z := effectstest.E2Zone(g, beast); z != game.ZoneBattlefield {
		t.Fatalf("the creature is in %q, want the battlefield", z)
	}
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %s owner %s, want yours and still theirs", c.Controller, c.Owner)
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped %v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
	}
	if n := effectstest.AttackEventsFor(g, beast); n != 0 {
		t.Errorf("the creature was DECLARED as an attacker %d times — CR 506.3c", n)
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, beast); z != game.ZoneHand {
		t.Fatalf("at the end step the creature is in %q, want its owner's hand", z)
	}
	if !effectstest.HasID(e2HandIDs(opp), beast) {
		t.Error("the creature went to a hand that is not its owner's")
	}
}

func TestZaraMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zara := effectstest.B43Catalog(g, me.ID, "Zara, Renegade Recruiter", "Legendary Creature — Human Pirate", "cd2720c2-522c-4fdc-9cef-9c5ce250fe7b", 4, 3)
	beast := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: beast, Name: "Their Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})

	effectstest.DeclareAttack(t, g, opp.ID, zara)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, beast); z != game.ZoneHand {
		t.Errorf("declined, but the creature is in %q", z)
	}
}

// --- Seize the Spotlight ------------------------------------------------

func TestSeizeTheSpotlightFameAndFortune(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	star := effectstest.CtrlPushCreature(g, a.ID, "A's Star")
	other := effectstest.CtrlPushCreature(g, a.ID, "A's Other")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == star {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})

	effectstest.CastCatalogSpell(t, g, "Seize the Spotlight", "Sorcery", "3d1b09a3-f151-4218-974b-02bb6159247a", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, a.ID, 0) // fame
	effectstest.AnswerOptionPick(t, g, b.ID, 1) // fortune
	effectstest.AnswerOptionPick(t, g, c.ID, 1) // fortune

	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTheirPermanents, me.ID)
	if pick == nil {
		t.Fatal("no pick of the fame player's creature")
	}
	if !effectstest.HasID(pick.ChooseCards, star) || !effectstest.HasID(pick.ChooseCards, other) {
		t.Errorf("candidates %v, want both of A's creatures", pick.ChooseCards)
	}
	hand := len(me.Hand.Cards)
	if err := g.ResolveTheirPermanents(pick.ID, me.ID, []uuid.UUID{star}); err != nil {
		t.Fatalf("ResolveTheirPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.ControllerOf(t, g, star); got != me.ID {
		t.Errorf("fame: controller %s, want %s", got, me.ID)
	}
	if got := effectstest.ControllerOf(t, g, other); got != a.ID {
		t.Error("the creature not chosen changed hands")
	}
	if effectstest.B16Tapped(t, g, star) || effectstest.SummoningSickOf(t, g, star) {
		t.Error("fame: the stolen creature must be untapped and hasty")
	}
	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("fortune ×2: drew %d, want 2", got)
	}
	if n, _ := effectstest.E2Treasures(g, me.ID); n != 2 {
		t.Errorf("fortune ×2: %d Treasures, want 2", n)
	}
}

func TestShatterskullSmashingSplitsXAmongTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 30)
	// #1563: the caster divides — 1/2, the reverse of the old even
	// division's 2/1.
	effectstest.B12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", effectstest.ShatterskullOracle,
		game.CastSpellParams{XValue: 3, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2}})
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("X=3 over two targets divided 1/2: got %d/%d", d1, d2)
	}
}

func TestShatterskullSmashingDoublesAtSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 30)
	effectstest.B12PlayFromHand(t, g, "Shatterskull Smashing", "Sorcery", effectstest.ShatterskullOracle,
		game.CastSpellParams{XValue: 6, Targets: effectstest.CardRefs(wall)})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, wall).DamageMarked; d != 12 {
		t.Errorf("X=6 at one target: %d damage, want twice X = 12", d)
	}
}

func TestShatterskullSmashingRefusesAThirdTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Wall", 0, 30)
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Shatterskull Smashing", TypeLine: "Sorcery",
		OracleID: effectstest.ShatterskullOracle, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3, Targets: effectstest.CardRefs(a, b, c)}); err == nil {
		t.Error("three targets accepted — \"up to two\"")
	}
}
