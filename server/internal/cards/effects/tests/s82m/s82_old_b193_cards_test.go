package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b193_cards_test.go — Commander 2018, batch 193 (#5526). Every test
// plays the real catalog entry through the engine.

const (
	c18TreasureHuntOracle      = "05079479-86a6-4041-a395-83d325b6ddb7"
	c18EnigmaSphinxOracle      = "6fb84ecd-ab60-4c2c-987d-8a0749db5777"
	c18DreamCacheOracle        = "81338d64-8fc3-41ac-afc8-e2a79df1159b"
	c18MythUnboundOracle       = "83e047ac-aaa9-4bd8-9ea0-49761d39139c"
	c18AdarkarValkyrieOracle   = "2932baae-bdf1-455a-9ac4-23ba1f83e7e1"
	c18BorderlandExplorerOracl = "65bc4f8a-3226-4648-862b-3f62736ff1ae"
	c18PrototypePortalOracle   = "7f3b6c25-b855-4ea5-a8ad-d954a6cdffb7"
	c18EmissaryOracle          = "b415684e-40f0-43c1-b8d5-1a90b75c5a47"
	c18SpawningGroundsOracle   = "1961dd92-db0b-4f02-b9c8-08f760f4051b"
	c18RighteousAuthorityOracl = "386faeef-f5cc-4ef9-b8cb-6ef06b33cbd0"
	c18EpicProportionsOracle   = "ad8edfdf-6915-4fc2-9226-3aae95951935"
)

func TestC18B193EveryCardIsFull(t *testing.T) {
	for _, o := range []string{c18TreasureHuntOracle, c18EnigmaSphinxOracle, c18DreamCacheOracle, c18MythUnboundOracle,
		c18AdarkarValkyrieOracle, c18BorderlandExplorerOracl, c18PrototypePortalOracle, c18EmissaryOracle,
		c18SpawningGroundsOracle, c18RighteousAuthorityOracl, c18EpicProportionsOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s not registered Full without caveats", o)
		}
	}
}

// c18Hand puts a catalog card with a cost in a player's hand.
func c18Hand(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// c18Library pushes a card on top of a library.
func c18Library(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func c18InZone(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// --- Treasure Hunt ---

func TestC18B193TreasureHuntTakesEveryRevealedCardUpToTheNonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	deep := c18Library(me, "Deep Card", "Creature — Bear")
	hit := c18Library(me, "The Hit", "Sorcery")
	land2 := c18Library(me, "Second Land", "Basic Land — Island")
	land1 := c18Library(me, "First Land", "Basic Land — Island")
	spell := c18Hand(me, "Treasure Hunt", "Sorcery", c18TreasureHuntOracle, "{1}{U}")
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPool(me, 1)
	size := me.Library.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"first land": land1, "second land": land2, "nonland": hit} {
		if !c18InZone(me.Hand, id) {
			t.Errorf("the %s was not put into the hand", name)
		}
	}
	if c18InZone(me.Hand, deep) || !c18InZone(me.Library, deep) {
		t.Error("the card below the nonland left the library")
	}
	if me.Library.Size() != size-3 {
		t.Errorf("library %d, want %d", me.Library.Size(), size-3)
	}
}

func TestC18B193TreasureHuntWithOnlyLandsTakesTheWholeLibraryAndDoesNotLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, c18Library(me, "Island", "Basic Land — Island"))
	}
	spell := c18Hand(me, "Treasure Hunt", "Sorcery", c18TreasureHuntOracle, "{1}{U}")
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for _, id := range lands {
		if !c18InZone(me.Hand, id) {
			t.Error("a revealed land stayed in the library")
		}
	}
	if me.Library.Size() != 0 || me.Eliminated {
		t.Errorf("library %d, eliminated %v: revealing is not drawing", me.Library.Size(), me.Eliminated)
	}
}

// --- Enigma Sphinx ---

func c18Sphinx(g *game.Game, owner, controller uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Enigma Sphinx",
		TypeLine: "Artifact Creature — Sphinx", OracleID: c18EnigmaSphinxOracle, Power: 5, Toughness: 4,
		Owner: owner, Controller: controller})
}

func TestC18B193EnigmaSphinxGoesThirdFromTheTopWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := c18Sphinx(g, me.ID, me.ID)
	for i := 0; i < 5; i++ {
		c18Library(me, "Filler", "Creature — Bear")
	}
	effectstest.KillCreature(t, g, me.ID, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	n := len(me.Library.Cards)
	if !c18InZone(me.Library, id) || me.Library.Cards[n-3].InstanceID != id {
		t.Errorf("the Sphinx is not third from the top (library %d, in library %v)", n, c18InZone(me.Library, id))
	}
	if c18InZone(me.Graveyard, id) {
		t.Error("the Sphinx is still in the graveyard")
	}
}

func TestC18B193EnigmaSphinxStolenGoesToItsOwnersGraveyardAndStaysThere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := c18Sphinx(g, opp.ID, me.ID)
	for i := 0; i < 5; i++ {
		c18Library(opp, "Filler", "Creature — Bear")
	}
	effectstest.KillCreature(t, g, opp.ID, id)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !c18InZone(opp.Graveyard, id) || c18InZone(opp.Library, id) {
		t.Error("a Sphinx put into another player's graveyard was tucked: that is not \"your graveyard\"")
	}
}

// --- Epic Proportions ---

func TestC18B193EpicProportionsPumpsGrantsTrampleAndCanBeCastWithFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.NmAdvanceTo(t, g, game.StepUpkeep)
	aura := c18Hand(me, "Epic Proportions", "Enchantment — Aura", c18EpicProportionsOracle, "{4}{G}{G}")
	effectstest.FillPoolColored(me, "G", 2)
	effectstest.FillPool(me, 4)
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("a flash Aura in the upkeep: %v", err)
	}
	effectstest.AdResolveSpell(t, g, aura)
	if p, to := effectstest.EffectivePTOf(t, g, bear); p != 7 || to != 7 {
		t.Errorf("enchanted creature is %d/%d, want 7/7", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("enchanted creature lacks trample")
	}
}

// --- Spawning Grounds ---

func TestC18B193SpawningGroundsLetsTheLandsControllerMakeABeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.GaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.PushAuraOnLand(t, g, me.ID, "Spawning Grounds", c18SpawningGroundsOracle, land)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Error("the Aura's controller activated an ability of a land they do not control")
	}
	g.Turn.PriorityHolder = 1
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(opp.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatalf("the land's controller activates it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var beast uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" {
			beast = c.InstanceID
			if c.Controller != opp.ID {
				t.Error("the Beast belongs to the land's controller")
			}
		}
	}
	if beast == uuid.Nil {
		t.Fatal("no Beast was created")
	}
	if p, to := effectstest.EffectivePTOf(t, g, beast); p != 5 || to != 5 {
		t.Errorf("the Beast is %d/%d, want 5/5", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, beast, "trample") {
		t.Error("the Beast lacks trample")
	}
	if err := g.ActivateCatalogAbility(opp.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
		t.Error("the tapped land activated again")
	}
}

// --- Adarkar Valkyrie ---

func c18Valkyrie(g *game.Game, me uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Adarkar Valkyrie",
		TypeLine: "Snow Creature — Angel", OracleID: c18AdarkarValkyrieOracle, Power: 4, Toughness: 5, Owner: me, Controller: me})
}

func TestC18B193AdarkarValkyrieReturnsTheTargetUnderYourControlWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valk := c18Valkyrie(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, valk, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: valk}}}); err == nil {
		t.Fatal("the Valkyrie targeted itself")
	}
	effectstest.Pr7Activate(t, g, me.ID, valk, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.KillCreature(t, g, opp.ID, bystander)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !c18InZone(opp.Graveyard, bystander) {
		t.Error("a creature that was not the target came back")
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, victim)
	if !ok {
		t.Fatal("the target did not return to the battlefield")
	}
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("returned under %v (owner %v), want controller %v and owner %v", c.Controller, c.Owner, me.ID, opp.ID)
	}
}

func TestC18B193AdarkarValkyrieOnlyWatchesThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valk := c18Valkyrie(g, me.ID)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	effectstest.Pr7Activate(t, g, me.ID, valk, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
	effectstest.EndTurn(t, g)
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !c18InZone(opp.Graveyard, victim) {
		t.Error("the creature returned after the turn ended")
	}
}

// --- Dream Cache ---

func c18DreamCacheCast(t *testing.T, g *game.Game, me *game.Player) (drawn []uuid.UUID) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil; me.Library.Cards = nil })
	for i := 0; i < 6; i++ {
		drawn = append(drawn, c18Library(me, "Card"+string(rune('A'+i)), "Creature — Bear"))
	}
	spell := c18Hand(me, "Dream Cache", "Sorcery", c18DreamCacheOracle, "{2}{U}")
	effectstest.FillPoolColored(me, "U", 1)
	effectstest.FillPool(me, 2)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return drawn
}

func c18PlaceCards(t *testing.T, g *game.Game, me *game.Player, label string, a, b uuid.UUID) {
	t.Helper()
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.BlbSTPickOption(t, g, me.ID, label)
	c := effectstest.PendingOfKind(g, game.PendingChoicePutInLibrary)
	if c == nil {
		t.Fatalf("no ordering prompt: %+v", g.PendingChoices)
	}
	var err error
	if label == "Put both cards on top of your library" {
		err = g.ResolvePutInLibrary(c.ID, me.ID, nil, []uuid.UUID{a, b})
	} else {
		err = g.ResolvePutInLibrary(c.ID, me.ID, []uuid.UUID{a, b}, nil)
	}
	if err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestC18B193DreamCachePutsTwoCardsBothOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drawn := c18DreamCacheCast(t, g, me)
	if me.Hand.Size() != 3 {
		t.Fatalf("hand %d after drawing three, want 3", me.Hand.Size())
	}
	// The three drawn are the top three: F, E, D.
	a, b, keep := drawn[5], drawn[4], drawn[3]
	c18PlaceCards(t, g, me, "Put both cards on top of your library", a, b)
	n := len(me.Library.Cards)
	if me.Hand.Size() != 1 || !c18InZone(me.Hand, keep) {
		t.Errorf("hand %d, want only the card that was kept", me.Hand.Size())
	}
	if top := me.Library.Cards[n-1].InstanceID; top != a {
		t.Errorf("top card %v, want the first ordered card", top)
	}
	if second := me.Library.Cards[n-2].InstanceID; second != b {
		t.Errorf("second card %v, want the second ordered card", second)
	}
}

func TestC18B193DreamCachePutsTwoCardsBothOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drawn := c18DreamCacheCast(t, g, me)
	a, b := drawn[5], drawn[3]
	c18PlaceCards(t, g, me, "Put both cards on the bottom of your library", a, b)
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1", me.Hand.Size())
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[a] || !bottom[b] {
		t.Error("the two cards are not the bottom two of the library")
	}
}

// --- Borderland Explorer ---

func TestC18B193BorderlandExplorerEachWhoDiscardedMaySearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
	myLand := c18Library(me, "My Forest", "Basic Land — Forest")
	oppLand := c18Library(opp, "Their Forest", "Basic Land — Forest")
	thirdLand := c18Library(third, "Third Forest", "Basic Land — Forest")
	myDiscard := c18Hand(me, "My Spare", "Creature — Bear", "", "")
	oppKeep := c18Hand(opp, "Their Spare", "Creature — Bear", "", "")
	thirdDiscard := c18Hand(third, "Third Spare", "Creature — Bear", "", "")
	explorer := c18Hand(me, "Borderland Explorer", "Creature — Elf Scout", c18BorderlandExplorerOracl, "{1}{G}")
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, explorer, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, myDiscard)
	effectstest.AnswerChooseCards(t, g, opp.ID)
	effectstest.AnswerChooseCards(t, g, third.ID, thirdDiscard)
	// The two who discarded may now search; the one who declined does not.
	if effectstest.SearchChoiceFor(g, opp.ID) != nil {
		t.Error("a player who declined to discard was offered a search")
	}
	for _, who := range []struct {
		p    *game.Player
		land uuid.UUID
	}{{me, myLand}, {third, thirdLand}} {
		effectstest.AnswerSearchByID(t, g, who.p.ID, who.land)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !c18InZone(me.Hand, myLand) || !c18InZone(third.Hand, thirdLand) {
		t.Error("a player who discarded did not get their basic land")
	}
	if c18InZone(opp.Hand, oppLand) || !c18InZone(opp.Hand, oppKeep) || me.Hand.Size() != 1 {
		t.Error("the player who declined searched, or lost their card")
	}
	if !c18InZone(me.Graveyard, myDiscard) || !c18InZone(third.Graveyard, thirdDiscard) {
		t.Error("a discarded card is not in its owner's graveyard")
	}
}

// --- Prototype Portal ---

func TestC18B193PrototypePortalImprintsAnArtifactAndCopiesItForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	golem := c18Hand(me, "Test Golem", "Artifact Creature — Golem", "", "{2}")
	g.WithWriteLock(func() {
		me.Hand.Cards[len(me.Hand.Cards)-1].Power, me.Hand.Cards[len(me.Hand.Cards)-1].Toughness = 3, 3
	})
	notArtifact := c18Hand(me, "Plain Bear", "Creature — Bear", "", "{1}{G}")
	portal := c18Hand(me, "Prototype Portal", "Artifact", c18PrototypePortalOracle, "{4}")
	effectstest.FillPool(me, 4)
	if err := g.CastSpell(me.ID, portal, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatalf("no imprint prompt: %+v", g.PendingChoices)
	}
	for _, id := range prompt.ChooseCards {
		if id == notArtifact {
			t.Error("a nonartifact card was offered for the imprint")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, golem)
	effectstest.PassPriorityAroundTable(t, g)
	if !c18InZone(g.Exile, golem) {
		t.Fatal("the artifact card was not exiled")
	}
	price, err := g.PriceActivation(me.ID, portal, 0, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if price.Ability.Cost.DemandsX() || price.Total.String() != "{2}" {
		t.Fatalf("priced at %s (X asked: %v), want {2}", price.Total.String(), price.Ability.Cost.DemandsX())
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("one mana paid a {2} cost")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Test Golem" && c.IsToken() {
			tokens++
			if c.Controller != me.ID {
				t.Error("the copy is not under the Portal controller's control")
			}
		}
	}
	if tokens != 1 || !c18InZone(g.Exile, golem) {
		t.Errorf("%d token copies (want 1), card still in exile: %v", tokens, c18InZone(g.Exile, golem))
	}
}

func TestC18B193PrototypePortalWithNothingImprintedCostsOnlyTapAndMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	portal := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Prototype Portal",
		TypeLine: "Artifact", OracleID: c18PrototypePortalOracle, Owner: me.ID, Controller: me.ID})
	before := len(g.Battlefield.Cards)
	if err := g.ActivateCatalogAbility(me.ID, portal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with nothing imprinted: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before {
		t.Error("a token appeared with nothing imprinted")
	}
}

// --- Righteous Authority ---

func TestC18B193RighteousAuthorityCountsTheEnchantedCreaturesControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Righteous Authority", c18RighteousAuthorityOracl)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	for i := 0; i < 3; i++ {
		c18Hand(opp, "Their Card", "Creature — Bear", "", "")
	}
	for i := 0; i < 5; i++ {
		c18Hand(me, "My Card", "Creature — Bear", "", "")
	}
	if p, to := effectstest.EffectivePTOf(t, g, host); p != 5 || to != 5 {
		t.Errorf("enchanted creature is %d/%d, want 5/5 (its controller holds three cards)", p, to)
	}
	for i := 0; i < 2; i++ {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, opp.Hand.Cards[0].InstanceID); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if p, _ := effectstest.EffectivePTOf(t, g, host); p != 3 {
		t.Errorf("after discarding down to one card the creature is %d wide, want 3", p)
	}
}

func TestC18B193RighteousAuthorityDrawsTheCreaturesControllerAnExtraCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Cmdr2B19Attach(g, me.ID, host, "Righteous Authority", c18RighteousAuthorityOracl)
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.NmAdvanceTo(t, g, game.StepPrecombatMain)
	// One for the turn's draw step, one from the Aura.
	if got := opp.Hand.Size(); got != oppHand+2 {
		t.Errorf("the creature's controller holds %d cards after their draw step, want %d", got, oppHand+2)
	}
	if got := me.Hand.Size(); got != myHand {
		t.Errorf("the Aura's controller drew: hand %d, want %d", got, myHand)
	}
}

// --- Myth Unbound ---

func TestC18B193MythUnboundReducesTheCommanderByItsCommandZoneCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Myth Unbound",
		TypeLine: "Enchantment", OracleID: c18MythUnboundOracle, Owner: me.ID, Controller: me.ID})
	cmdr := game.NewCard("Test Commander", me.ID)
	cmdr.TypeLine, cmdr.ManaCost, cmdr.IsCommander, cmdr.Power, cmdr.Toughness = "Legendary Creature — Human", "{3}{G}{G}", true, 3, 3
	other := game.NewCard("Test Bear", me.ID)
	other.TypeLine, other.ManaCost = "Creature — Bear", "{3}{G}{G}"
	me.Command.PushTop(cmdr)
	me.Hand.PushTop(other)
	me.CommanderCasts[cmdr.InstanceID] = 2
	// The commander has been cast from the command zone twice: {3}{G}{G} less {2}.
	price := func(c game.Card, from game.ZoneKind) int {
		base, err := game.ParseCost(c.ManaCost)
		if err != nil {
			t.Fatalf("ParseCost: %v", err)
		}
		out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: c, Controller: me.ID, FromZone: from})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
		return out.ManaValue()
	}
	if got := price(cmdr, game.ZoneCommand); got != 3 {
		t.Errorf("commander priced at %d, want 3", got)
	}
	if got := price(cmdr, game.ZoneHand); got != 3 {
		t.Errorf("commander cast from hand priced at %d, want 3", got)
	}
	if got := price(other, game.ZoneHand); got != 5 {
		t.Errorf("a noncommander priced at %d, want the full 5", got)
	}
	me.CommanderCasts[cmdr.InstanceID] = 0
	if got := price(cmdr, game.ZoneCommand); got != 5 {
		t.Errorf("a commander never cast from the command zone priced at %d, want 5", got)
	}
}

func TestC18B193MythUnboundDrawsWhenYourCommanderIsPutIntoTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Myth Unbound",
		TypeLine: "Enchantment", OracleID: c18MythUnboundOracle, Owner: me.ID, Controller: me.ID})
	cmdr := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Commander",
		TypeLine: "Legendary Creature — Human", Power: 3, Toughness: 3, IsCommander: true, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Commander",
		TypeLine: "Legendary Creature — Human", Power: 3, Toughness: 3, IsCommander: true, Owner: opp.ID, Controller: opp.ID})
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	effectstest.AnswerCommanderReturn(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("an opponent's commander made you draw: hand %d, want %d", me.Hand.Size(), hand)
	}
	effectstest.KillCreature(t, g, me.ID, cmdr)
	g.RunStateChecksForTest()
	effectstest.AnswerCommanderReturn(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !c18InZone(me.Command, cmdr) {
		t.Fatal("the commander did not go to the command zone")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: one card for the commander", me.Hand.Size(), hand+1)
	}
}

// --- Emissary of Grudges ---

func c18Emissary(t *testing.T, g *game.Game, chosenIdx int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Emissary of Grudges", "Creature — Efreet", c18EmissaryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, chosenIdx)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Emissary of Grudges is not on the battlefield")
	}
	return id
}

func TestC18B193EmissaryRetargetsTheChosenPlayersSpellAimedAtYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	emissary := c18Emissary(t, g, 0)
	if w, by := g.SecretChosenPlayerOf(emissary); w != o1.ID || by != me.ID {
		t.Fatalf("secret choice = %v by %v, want %v by %v", w, by, o1.ID, me.ID)
	}
	bolt := effectstest.BoltFrom(t, g, o1, me.ID)
	g.Turn.PriorityHolder = 0
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("reveal and activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerRetargetWith(t, g, me.ID, effectstest.RtPlayerRef(o2.ID))
	life, o2life := me.Life, o2.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: the Bolt still hit me", me.Life, life)
	}
	if o2.Life != o2life-3 {
		t.Errorf("the new target's life %d, want %d", o2.Life, o2life-3)
	}
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Error("the ability activated a second time")
	}
}

func TestC18B193EmissaryDoesNothingToAnotherPlayersSpellButIsStillSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	emissary := c18Emissary(t, g, 0)
	bolt := effectstest.BoltFrom(t, g, o2, me.ID)
	life := me.Life
	g.Turn.PriorityHolder = 0
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestRetarget(g, me.ID) != nil {
		t.Fatal("a spell the chosen player did not control was offered new targets")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d: the Bolt was not left alone", me.Life, life-3)
	}
	_ = o1
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was not spent")
	}
}

func TestC18B193EmissaryLeavesAChosenPlayersSpellThatDoesNotTargetYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	emissary := c18Emissary(t, g, 0)
	bolt := effectstest.BoltFrom(t, g, o1, o2.ID)
	life := o2.Life
	g.Turn.PriorityHolder = 0
	if err := g.ActivateCatalogAbility(me.ID, emissary, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestRetarget(g, me.ID) != nil {
		t.Fatal("a spell that does not target you or your permanents was offered new targets")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o2.Life != life-3 {
		t.Errorf("life %d, want %d", o2.Life, life-3)
	}
}
