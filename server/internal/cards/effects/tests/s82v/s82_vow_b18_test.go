package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b18_test.go — Innistrad: Crimson Vow, S82 batch 18.

const (
	vowScionOfOpulence    = "5818b31e-2e86-4230-937e-478e14534f91"
	vowEruth              = "a0380b63-58ef-4545-beec-6ad307bbc21b"
	vowDemonicBargain     = "ae18580a-1753-4dc8-b437-8a779b1a6e4d"
	vowStensiaUprising    = "a881bf6e-6247-4adb-9f9c-e4b928692573"
	vowImposingGrandeur   = "c9ef02d6-717d-4f2a-aa10-0af717d6360f"
	vowSpectralArcanist   = "aba60536-ffbd-480c-8e8f-9639bdc53d4b"
	vowCurseOfHospitality = "0fb70b94-0847-4155-8953-cb53fef82ef1"
	vowChandra            = "4686f776-2055-470c-bd99-7c4bb32902c0"
)

func TestS82VowB18CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		vowScionOfOpulence: "Scion of Opulence", vowEruth: "Eruth, Tormented Prophet",
		vowDemonicBargain: "Demonic Bargain", vowStensiaUprising: "Stensia Uprising",
		vowImposingGrandeur: "Imposing Grandeur", vowSpectralArcanist: "Spectral Arcanist",
		vowCurseOfHospitality: "Curse of Hospitality", vowChandra: "Chandra, Dressed to Kill",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// vowSettle passes priority until the stack is empty, picking `pick` for any
// target prompt the chooser is asked.
func vowSettle(t *testing.T, g *game.Game, chooser uuid.UUID, pick *game.TargetRef) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if pick == nil {
				t.Fatalf("unexpected pick_target prompt for %s", chooser)
			}
			if err := g.ResolvePickTarget(p.ID, p.Chooser, *pick); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("did not settle")
}

// vowExileHas counts the cards named `name` in exile.
func vowExileHas(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Exile.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func vowCount(g *game.Game, name string, controller uuid.UUID) int {
	return effectstest.CardsNamed(g, name, controller)
}

// --- Scion of Opulence ---------------------------------------------

func scionTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scion := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Scion of Opulence", OracleID: vowScionOfOpulence,
		TypeLine: "Creature — Vampire Noble", Power: 3, Toughness: 1, Colors: []string{"R"}})
	effectstest.AdvanceToMain(t, g)
	return g, me, scion
}

func TestS82ScionOfOpulenceMakesATreasureWhenItOrANontokenVampireDies(t *testing.T) {
	g, me, scion := scionTable(t)
	vamp := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Fang", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2})
	tokenVamp := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Fang Token", TypeLine: "Token Creature — Vampire", Power: 1, Toughness: 1})
	bear := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	opp := g.Seats[1]
	theirs := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Fang", TypeLine: "Creature — Vampire", Power: 2, Toughness: 2})

	effectstest.B11Die(t, g, me.ID, bear)
	effectstest.B11Die(t, g, me.ID, tokenVamp)
	effectstest.B11Die(t, g, opp.ID, theirs)
	if n := vowCount(g, "Treasure", me.ID); n != 0 {
		t.Fatalf("%d Treasures for a Bear, a Vampire token and an opposing Vampire, want 0", n)
	}
	effectstest.B11Die(t, g, me.ID, vamp)
	if n := vowCount(g, "Treasure", me.ID); n != 1 {
		t.Fatalf("%d Treasures after a nontoken Vampire died, want 1", n)
	}
	effectstest.B11Die(t, g, me.ID, scion)
	if n := vowCount(g, "Treasure", me.ID); n != 2 {
		t.Fatalf("%d Treasures after the Scion died, want 2", n)
	}
}

func TestS82ScionOfOpulenceSacrificesTwoArtifactsToExileAndPlayTheTop(t *testing.T) {
	g, me, scion := scionTable(t)
	a := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock A", TypeLine: "Artifact"})
	b := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock B", TypeLine: "Artifact"})
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})

	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, scion, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated with one artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, scion, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, scion}}); err == nil {
		t.Fatal("sacrificed a non-artifact")
	}
	effectstest.B16Activate(t, g, me.ID, scion, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}})
	effectstest.DrcSettle(t, g, me.ID)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if p := effectstest.ChandraTorchGrant(me, top); p == nil {
		t.Fatal("no permission to play the exiled card")
	}
	if vowCount(g, "Rock A", me.ID)+vowCount(g, "Rock B", me.ID) != 0 {
		t.Fatal("the artifacts were not sacrificed")
	}
}

// --- Eruth, Tormented Prophet --------------------------------------

func TestS82EruthExilesTwoInsteadOfDrawingAndYouMayPlayThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Eruth, Tormented Prophet", OracleID: vowEruth,
		TypeLine: "Legendary Creature — Human Wizard", Power: 2, Toughness: 4})
	var ids []uuid.UUID
	for _, n := range []string{"E1", "E2", "E3", "E4", "E5"} {
		id := uuid.New()
		ids = append(ids, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	// ids[4] (E5) is the top of the library.
	hand, lib := me.Hand.Size(), me.Library.Size()

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d → %d: Eruth must replace the draw", hand, me.Hand.Size())
	}
	if me.Library.Size() != lib-2 {
		t.Fatalf("library %d → %d, want two exiled", lib, me.Library.Size())
	}
	for _, id := range []uuid.UUID{ids[4], ids[3]} {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled", id)
		}
		if effectstest.ChandraTorchGrant(me, id) == nil {
			t.Errorf("no permission to play %v this turn", id)
		}
	}
	if effectstest.ChandraTorchGrant(me, ids[2]) != nil {
		t.Error("the third card got a permission")
	}

	// Draw two: each draw is replaced on its own, four more exiled.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Fatalf("hand %d after draw two, want %d", me.Hand.Size(), hand)
	}
	if got := vowExileHas(g, "E1") + vowExileHas(g, "E2") + vowExileHas(g, "E3") + vowExileHas(g, "E4"); got != 4 {
		t.Fatalf("%d of the next four exiled, want 4", got)
	}

	// An opponent's draw is theirs: Eruth says "you".
	oppHand := opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if opp.Hand.Size() != oppHand+1 {
		t.Fatalf("opponent hand %d → %d, want a normal draw", oppHand, opp.Hand.Size())
	}
}

func TestS82EruthWithOneCardLeftExilesItAndDoesNotLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Eruth, Tormented Prophet", OracleID: vowEruth,
		TypeLine: "Legendary Creature — Human Wizard", Power: 2, Toughness: 4})
	for me.Library.Size() > 1 {
		me.Library.PopTop()
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.DrcSettle(t, g, me.ID)
	if me.Library.Size() != 0 {
		t.Fatalf("library %d, want the last card exiled", me.Library.Size())
	}
	if me.Eliminated {
		t.Fatal("the draw was replaced, so no card was drawn from an empty library")
	}
}

// --- Demonic Bargain -----------------------------------------------

func TestS82DemonicBargainExilesThirteenThenSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	want := uuid.New()
	for i := 0; i < 10; i++ {
		me.Library.PushBottom(game.Card{InstanceID: uuid.New(), Name: "Pad", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	me.Library.PushBottom(game.Card{InstanceID: want, Name: "Wanted", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	if lib < 14 {
		t.Fatalf("library %d, the test needs more than thirteen", lib)
	}
	hand := me.Hand.Size()
	card := game.Card{InstanceID: uuid.New(), Name: "Demonic Bargain", TypeLine: "Sorcery", OracleID: vowDemonicBargain,
		ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Exile.Cards); got != 13 {
		t.Fatalf("%d cards exiled, want thirteen", got)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Wanted")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(want) {
		t.Fatal("the searched card is not in hand")
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d, want +1 (the spell left, the card arrived)", hand, me.Hand.Size())
	}
	if me.Library.Size() != lib-14 {
		t.Fatalf("library %d, want %d", me.Library.Size(), lib-14)
	}
}

func TestS82DemonicBargainOnAShortLibraryExilesWhatThereIsAndSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for me.Library.Size() > 5 {
		me.Library.PopTop()
	}
	card := game.Card{InstanceID: uuid.New(), Name: "Demonic Bargain", TypeLine: "Sorcery", OracleID: vowDemonicBargain,
		ManaCost: "{2}{B}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(g.Exile.Cards); got != 5 {
		t.Fatalf("%d cards exiled, want the five that were there", got)
	}
	if me.Library.Size() != 0 {
		t.Fatalf("library %d, want empty", me.Library.Size())
	}
}

// --- Stensia Uprising ----------------------------------------------

func stensiaTable(t *testing.T, others int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	st := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Stensia Uprising", OracleID: vowStensiaUprising,
		TypeLine: "Enchantment", Colors: []string{"R"}})
	for i := 0; i < others; i++ {
		effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.AdvanceToEndStep(t, g)
	return g, me, st
}

func TestS82StensiaUprisingMakesAHumanAndWithThirteenPermanentsMaySacrificeForSeven(t *testing.T) {
	g, me, st := stensiaTable(t, 11) // Stensia + 11 lands + the token = 13
	opp := g.Seats[1]
	before := opp.Life
	vowSettle(t, g, me.ID, nil)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human"); n != 1 {
		t.Fatalf("%d Human tokens, want 1", n)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	vowSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(st) {
		t.Fatal("Stensia Uprising was not sacrificed")
	}
	if opp.Life != before-7 {
		t.Fatalf("opponent life %d → %d, want 7 damage", before, opp.Life)
	}
}

func TestS82StensiaUprisingDecliningTheSacrificeKeepsItAndDealsNothing(t *testing.T) {
	g, me, st := stensiaTable(t, 11)
	opp := g.Seats[1]
	before := opp.Life
	vowSettle(t, g, me.ID, nil)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	vowSettle(t, g, me.ID, nil)
	if !g.Battlefield.Contains(st) || opp.Life != before {
		t.Fatalf("kept=%v life %d → %d", g.Battlefield.Contains(st), before, opp.Life)
	}
}

func TestS82StensiaUprisingAsksNothingWhenThePermanentCountIsNotThirteen(t *testing.T) {
	for _, others := range []int{10, 12} { // 12 and 14 with the token
		g, me, st := stensiaTable(t, others)
		vowSettle(t, g, me.ID, nil)
		if n := effectstest.CountBattlefieldNamed(g, me.ID, "Human"); n != 1 {
			t.Fatalf("others=%d: %d Human tokens, want 1", others, n)
		}
		if effectstest.LatestConfirmFor(g, me.ID) != nil || !g.Battlefield.Contains(st) {
			t.Fatalf("others=%d: asked about the sacrifice at the wrong count", others)
		}
	}
}

// --- Imposing Grandeur ---------------------------------------------

func vowCommander(g *game.Game, p *game.Player, name, mana string, zone string) uuid.UUID {
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human", ManaCost: mana,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID, IsCommander: true}
	switch zone {
	case "battlefield":
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	case "command":
		p.Command.PushTop(c)
	case "hand":
		p.Hand.PushTop(c)
	}
	return id
}

func TestS82ImposingGrandeurEachYesSayerDiscardsThenDrawsTheirCommandersValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, p1, p2, p3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.AdvanceToMain(t, g)
	vowCommander(g, me, "Cmdr Zero", "{2}{R}", "command") // 3 in the command zone
	vowCommander(g, p1, "Cmdr One", "{4}{G}", "battlefield")
	vowCommander(g, p1, "Cmdr One b", "{1}", "command") // greatest is 5
	vowCommander(g, p2, "Cmdr Two", "{6}", "hand")      // in hand: does not count
	card := game.Card{InstanceID: uuid.New(), Name: "Imposing Grandeur", TypeLine: "Sorcery", OracleID: vowImposingGrandeur,
		ManaCost: "{4}{R}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	handMe, handP2, handP3 := me.Hand.Size()-1, p2.Hand.Size(), p3.Hand.Size()
	libMe, libP1, libP2 := me.Library.Size(), p1.Library.Size(), p2.Library.Size()
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, p1.ID, true)
	effectstest.AnswerMayChoice(t, g, p2.ID, true)
	effectstest.AnswerMayChoice(t, g, p3.ID, false)
	effectstest.DrcSettle(t, g, me.ID)

	if me.Hand.Size() != 3 || me.Library.Size() != libMe-3 {
		t.Errorf("me: hand %d (was %d), library %d → %d; want 3 cards drawn", me.Hand.Size(), handMe, libMe, me.Library.Size())
	}
	if p1.Hand.Size() != 5 || p1.Library.Size() != libP1-5 {
		t.Errorf("p1: hand %d, library %d → %d; want 5 cards drawn", p1.Hand.Size(), libP1, p1.Library.Size())
	}
	if p2.Hand.Size() != 0 || p2.Library.Size() != libP2 {
		t.Errorf("p2: hand %d (was %d), library %d → %d; a commander in hand is worth zero", p2.Hand.Size(), handP2, libP2, p2.Library.Size())
	}
	if p3.Hand.Size() != handP3 {
		t.Errorf("p3 declined but their hand went %d → %d", handP3, p3.Hand.Size())
	}
}

// --- Spectral Arcanist ---------------------------------------------

func arcanistTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[0]
}

func arcanistCard(me *game.Player) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Spectral Arcanist", TypeLine: "Creature — Spirit Wizard",
		OracleID: vowSpectralArcanist, ManaCost: "{3}{U}", Power: 3, Toughness: 2, Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID}
}

func TestS82SpectralArcanistCastsASpellFromAnyGraveyardUpToItsSpiritCountThenExilesIt(t *testing.T) {
	g, me := arcanistTable(t)
	opp := g.Seats[1]
	cheap := uuid.New()  // mana value 1: castable with one Spirit (the Arcanist)
	pricey := uuid.New() // mana value 2: too big
	theirs := uuid.New() // an opponent's graveyard counts
	me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Pricey", TypeLine: "Sorcery", ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: opp.ID, Controller: opp.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Mine Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	oppLife := opp.Life

	effectstest.CastFromHandAndResolve(t, g, me, arcanistCard(me))
	prompt := effectstest.KeywordCastPromptFor(g, me.ID)
	if prompt == nil || prompt.Kind != game.PendingChoiceResolutionCast {
		t.Fatalf("no resolution_cast offer: %+v", prompt)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.DiscardOptions {
		offered[id] = true
	}
	if !offered[cheap] || !offered[theirs] || offered[pricey] {
		t.Fatalf("offered %v: want both one-mana instants (either graveyard) and not the two-mana sorcery", offered)
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, theirs, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast from their graveyard: %v", err)
	}
	// "An instant or sorcery spell": one cast spends the permission.
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true,
		AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("a second spell was cast on one trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	_ = oppLife
	if !g.Exile.Contains(theirs) {
		t.Fatal("the spell was not exiled instead of going to a graveyard")
	}
	if opp.Graveyard.Contains(theirs) {
		t.Fatal("the spell went to the graveyard")
	}
}

func TestS82SpectralArcanistWithNoFittingSpellOffersNothing(t *testing.T) {
	g, me := arcanistTable(t)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Pricey", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastFromHandAndResolve(t, g, me, arcanistCard(me))
	if p := effectstest.KeywordCastPromptFor(g, me.ID); p != nil {
		t.Fatalf("an offer with only a four-mana sorcery and one Spirit: %+v", p)
	}
}

func TestS82SpectralArcanistCountsEverySpirit(t *testing.T) {
	g, me := arcanistTable(t)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Ghost", TypeLine: "Creature — Spirit", Power: 1, Toughness: 1})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Ghost 2", TypeLine: "Creature — Spirit", Power: 1, Toughness: 1})
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Three Drop", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.CastFromHandAndResolve(t, g, me, arcanistCard(me)) // three Spirits with the Arcanist
	p := effectstest.KeywordCastPromptFor(g, me.ID)
	if p == nil || len(p.DiscardOptions) != 1 || p.DiscardOptions[0] != big {
		t.Fatalf("offer %+v, want the three-mana sorcery", p)
	}
}

// --- Curse of Hospitality ------------------------------------------

func TestS82CurseOfHospitalityGivesAttackersTrampleAndFeedsTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	curse := effectstest.CastCatalogSpell(t, g, "Curse of Hospitality", effectstest.CurseAuraTypeLine, vowCurseOfHospitality,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, curse); host.Kind != game.TargetPlayer || host.ID != victim.ID {
		t.Fatalf("Curse AttachedTo = %+v, want player %s", host, victim.ID)
	}
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	top := uuid.New()
	victim.Library.PushTop(game.Card{InstanceID: top, Name: "Victims Top", TypeLine: "Sorcery", ManaCost: "{1}{G}", Owner: victim.ID, Controller: victim.ID})
	bystanderTop := uuid.New()
	bystander.Library.PushTop(game.Card{InstanceID: bystanderTop, Name: "Bystander Top", TypeLine: "Sorcery", Owner: bystander.ID, Controller: bystander.ID})
	before := victim.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, victim.ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, bystander.ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	hasTrample := func(id uuid.UUID) bool {
		has := false
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == id {
					has = game.HasKeyword(&c, "trample")
				}
			}
		})
		return has
	}
	if !hasTrample(a) {
		t.Error("a creature attacking the enchanted player lacks trample")
	}
	if hasTrample(b) {
		t.Error("a creature attacking someone else has trample")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	vowSettle(t, g, me.ID, nil)
	if victim.Life != before-2 {
		t.Fatalf("victim life %d → %d", before, victim.Life)
	}
	if !g.Exile.Contains(top) {
		t.Fatal("the enchanted player did not exile the top card of their library")
	}
	if g.Exile.Contains(bystanderTop) {
		t.Error("a creature that hit someone else made the bystander exile a card")
	}
	grant := effectstest.ChandraTorchGrant(me, top)
	if grant == nil {
		t.Fatal("the creature's controller has no permission to play the card")
	}
	if !grant.AnyColor {
		t.Errorf("grant %+v: want mana of any colour to cast it", *grant)
	}
	if effectstest.ChandraTorchGrant(victim, top) != nil {
		t.Error("the enchanted player got the permission")
	}
}

// --- Chandra, Dressed to Kill ----------------------------------------

func chandraDressedTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chandra := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Dressed to Kill", "Legendary Planeswalker — Chandra", vowChandra, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 3) })
	effectstest.AdvanceToMain(t, g)
	return g, me, chandra
}

func TestS82ChandraDressedFirstPlusAddsRAndPingsAPlayer(t *testing.T) {
	g, me, chandra := chandraDressedTable(t)
	opp := g.Seats[1]
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != before-1 {
		t.Errorf("opponent life %d → %d, want 1 damage", before, opp.Life)
	}
	if effectstest.LoyaltyOf(g, chandra) != 4 {
		t.Errorf("loyalty %d, want 4", effectstest.LoyaltyOf(g, chandra))
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d, want {R}", me.ManaPool.Total())
	}
}

func TestS82ChandraDressedFirstPlusWithNoTargetStillAddsMana(t *testing.T) {
	g, me, chandra := chandraDressedTable(t)
	effectstest.B16Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	if me.ManaPool.Total() != 1 || effectstest.LoyaltyOf(g, chandra) != 4 {
		t.Errorf("pool %d, loyalty %d", me.ManaPool.Total(), effectstest.LoyaltyOf(g, chandra))
	}
}

func TestS82ChandraDressedSecondPlusGrantsACastOnlyForARedCard(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		want   bool
	}{{"red", []string{"R"}, true}, {"not red", []string{"G"}, false}, {"red and green", []string{"R", "G"}, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, chandra := chandraDressedTable(t)
			top := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Creature — Bear", ManaCost: "{1}", Colors: tc.colors,
				Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
			effectstest.B16Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
			if !g.Exile.Contains(top) {
				t.Fatal("the top card was not exiled")
			}
			grant := effectstest.ChandraTorchGrant(me, top)
			if (grant != nil) != tc.want {
				t.Fatalf("grant=%v, want %v", grant != nil, tc.want)
			}
			if grant != nil && !grant.CastOnly {
				t.Error("the permission must be to cast, not to play")
			}
		})
	}
}

func TestS82ChandraDressedUltimateExilesFiveGrantsRedCastsAndGivesTheEmblem(t *testing.T) {
	g, me, chandra := chandraDressedTable(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, 4) }) // 7
	red1, red2, green := uuid.New(), uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery"},
		{InstanceID: green, Name: "Green", TypeLine: "Sorcery", ManaCost: "{1}{G}", Colors: []string{"G"}},
		{InstanceID: red2, Name: "Red Two", TypeLine: "Instant", ManaCost: "{2}{R}", Colors: []string{"R"}},
		{InstanceID: red1, Name: "Red One", TypeLine: "Instant", ManaCost: "{1}{R}", Colors: []string{"R"}},
		{InstanceID: uuid.New(), Name: "Sixth", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	// Top five: Sixth, Red One, Red Two, Green, Filler... Sixth is on top, so
	// the five exiled are Sixth, Red One, Red Two, Green, Filler.
	effectstest.B16Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(chandra) {
		t.Fatal("Chandra at 0 loyalty stays on the battlefield")
	}
	for _, id := range []uuid.UUID{red1, red2} {
		if effectstest.ChandraTorchGrant(me, id) == nil {
			t.Errorf("no permission over red card %v", id)
		}
	}
	if effectstest.ChandraTorchGrant(me, green) != nil {
		t.Error("a green card may be cast")
	}
	if n := len(me.Emblems.Cards); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}

	// Cast Red One (mana value 2): two mana spent, so the emblem deals 2.
	effectstest.FloatForTest(g, me, "RR")
	before := opp.Life
	if err := g.CastSpell(me.ID, red1, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		// A target-less instant is fine too; retry untargeted.
		if err2 := g.CastSpell(me.ID, red1, game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err2 != nil {
			t.Fatalf("cast red card: %v / %v", err, err2)
		}
	}
	vowSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Fatalf("opponent life %d → %d, want 2 from the emblem (2 mana spent)", before, opp.Life)
	}
}
