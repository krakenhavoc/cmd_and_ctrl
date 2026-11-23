package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsfl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/fl"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b04_test.go — set BNS, batch 4: Special Guests (S82, #5929).

const (
	bnsDolmenGate      = "d53d000f-c92f-4d70-ae60-8be3791da8a2"
	bnsStarCompass     = "b5e9599f-9561-4e3f-9904-9a4dab98cabb"
	bnsPaintersServant = "baf93873-35e6-4bf2-bdcc-78a5206422fb"
	bnsCavalierOfDawn  = "7c520355-1ab8-4d6d-9a29-0f63d6b60024"
	bnsFaerieMacabre   = "c7d8beb6-472c-4758-9876-e8cbf98f4f72"
	bnsFiendArtisan    = "43b8456a-3333-4936-a09c-324327619c36"
	bnsFleshDuplicate  = "1869eb1e-7786-4fea-8a92-2bd839de47df"
	bnsBushwhacker     = "878b6808-48af-4047-ac35-bfc58bf92e7d"
	bnsGrimHaruspex    = "72aef1d8-11df-4bc2-af83-907b05df73a8"
	bnsKinsbaile       = "e21ed0d0-26b1-4ffe-971a-babf37f4492c"
	bnsMagusLibrary    = "69266cb7-50eb-43d2-96d6-fc3c85be49f7"
)

func TestS82BnsB04CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsDolmenGate: "Dolmen Gate", bnsStarCompass: "Star Compass", bnsPaintersServant: "Painter's Servant",
		bnsCavalierOfDawn: "Cavalier of Dawn", bnsFaerieMacabre: "Faerie Macabre", bnsFiendArtisan: "Fiend Artisan",
		bnsFleshDuplicate: "Flesh Duplicate", bnsBushwhacker: "Goblin Bushwhacker", bnsGrimHaruspex: "Grim Haruspex",
		bnsKinsbaile: "Kinsbaile Cavalier", bnsMagusLibrary: "Magus of the Library",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec)
		}
	}
}

// bnsMarked is the combat damage marked on a battlefield card.
func bnsMarked(g *game.Game, id uuid.UUID) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

// --- Dolmen Gate -----------------------------------------------------

func TestS82BnsB04DolmenGatePreventsCombatDamageToYourAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dolmen Gate", "Artifact", bnsDolmenGate, false)
	mk := func(owner uuid.UUID, attacking bool) uuid.UUID {
		c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 5, Owner: owner, Controller: owner}
		if attacking {
			c.AttackingTarget = opp.ID
		}
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	blocker := mk(opp.ID, false)
	attacker, idle, theirAttacker := mk(me.ID, true), mk(me.ID, false), mk(opp.ID, true)
	for _, id := range []uuid.UUID{attacker, idle, theirAttacker} {
		if err := g.MarkCombatDamage(blocker, id, 2); err != nil {
			t.Fatalf("MarkCombatDamage: %v", err)
		}
	}
	if got := bnsMarked(g, attacker); got != 0 {
		t.Errorf("my attacking creature took %d combat damage, want 0", got)
	}
	if got := bnsMarked(g, idle); got != 2 {
		t.Errorf("my non-attacking creature took %d, want 2", got)
	}
	if got := bnsMarked(g, theirAttacker); got != 2 {
		t.Errorf("the opponent's attacker took %d, want 2", got)
	}
}

// --- Star Compass ----------------------------------------------------

func TestS82BnsB04StarCompassEntersTappedAndNamesTheColoursOfYourBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Star Compass", TypeLine: "Artifact", OracleID: bnsStarCompass, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
		t.Fatalf("Star Compass on the battlefield=%v tapped=%v, want tapped", ok, c.Tapped)
	}
	if got := effectstest.Producible(g, id); len(got) != 0 {
		t.Errorf("with no basic land it could produce %v, want nothing", got)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Steam Vents", TypeLine: "Land — Island Mountain", Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: opp.ID, Controller: opp.ID})
	got := effectstest.Producible(g, id)
	slices.Sort(got)
	if len(got) != 1 || got[0] != "G" {
		t.Errorf("could produce %v, want only {G} (a nonbasic land and an opponent's basic do not count)", got)
	}
}

// --- Painter's Servant -----------------------------------------------

func TestS82BnsB04PaintersServantAddsTheChosenColourInEveryZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	inHand, inGrave, inLib := uuid.New(), uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: inHand, Name: "Hand Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: inGrave, Name: "Grave Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	opp.Library.PushTop(game.Card{InstanceID: inLib, Name: "Lib Rock", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	servant := effectstest.B38CastCreature(t, g, "Painter's Servant", "Artifact Creature — Scarecrow", bnsPaintersServant, 1, 3)
	effectstest.AnswerColor(t, g, me.ID, "U")
	effectstest.PassPriorityAroundTable(t, g)

	colours := func(id uuid.UUID) []string {
		var out []string
		g.ReadSnapshot(func() {
			if c, ok := g.LookupCardForEffect(id); ok {
				out = append(out, c.EffectiveColors()...)
			}
		})
		slices.Sort(out)
		return out
	}
	for name, tc := range map[string]struct {
		id   uuid.UUID
		want []string
	}{
		"the Servant itself":            {servant, []string{"U"}},
		"an opponent's creature":        {bear, []string{"U"}},
		"a green card in hand":          {inHand, []string{"G", "U"}},
		"a green card in a graveyard":   {inGrave, []string{"G", "U"}},
		"a colorless card in a library": {inLib, []string{"U"}},
	} {
		if got := colours(tc.id); !slices.Equal(got, tc.want) {
			t.Errorf("%s: colours %v, want %v", name, got, tc.want)
		}
	}
	// It leaves the battlefield and the colours go with it.
	effectstest.KillCreature(t, g, me.ID, servant)
	effectstest.DrcSettle(t, g, me.ID)
	if got := colours(inHand); !slices.Equal(got, []string{"G"}) {
		t.Errorf("after the Servant left, a green card in hand is %v, want {G}", got)
	}
}

// --- Faerie Macabre --------------------------------------------------

func TestS82BnsB04FaerieMacabreDiscardsItselfToExileUpToTwoCardsFromGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	faerie := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: faerie, Name: "Faerie Macabre", TypeLine: "Creature — Faerie Rogue", OracleID: bnsFaerieMacabre, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	mine, theirs, extra := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "A", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "B", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: extra, Name: "C", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	ref := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	if err := g.ActivateCatalogAbility(me.ID, faerie, 0, game.ActivateAbilityParams{Targets: ref(mine, theirs, extra)}); err == nil {
		t.Fatal("three targets were accepted for \"up to two\"")
	}
	if !me.Hand.Contains(faerie) {
		t.Fatal("the refused activation still discarded the Faerie")
	}
	if err := g.ActivateCatalogAbility(me.ID, faerie, 0, game.ActivateAbilityParams{Targets: ref(mine, theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(faerie) {
		t.Error("the Faerie was not discarded as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mine) || opp.Graveyard.Contains(theirs) {
		t.Error("a targeted card is still in a graveyard")
	}
	if !opp.Graveyard.Contains(extra) {
		t.Error("the untargeted card was exiled")
	}
}

// --- Goblin Bushwhacker ----------------------------------------------

func TestS82BnsB04GoblinBushwhackerPumpsOnlyWhenKicked(t *testing.T) {
	cast := func(kicked bool) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Goblin Bushwhacker", TypeLine: "Creature — Goblin Warrior", OracleID: bnsBushwhacker, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		var optional []int
		if kicked {
			optional = []int{0}
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		return g, id, mine, theirs
	}
	g, self, mine, theirs := cast(true)
	for name, id := range map[string]uuid.UUID{"the Bear": mine, "the Bushwhacker": self} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("kicked: %s has no haste", name)
		}
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("kicked: the Bear's power is %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, self); got != 2 {
		t.Errorf("kicked: the Bushwhacker's power is %d, want 2", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") || effectstest.EffectivePower(t, g, theirs) != 3 {
		t.Error("kicked: an opponent's creature was pumped")
	}
	g, self, mine, _ = cast(false)
	if effectstest.HasEffectiveKeyword(t, g, self, "haste") || effectstest.EffectivePower(t, g, mine) != 2 {
		t.Error("unkicked: the Bushwhacker pumped the team")
	}
}

// --- Grim Haruspex ---------------------------------------------------

func TestS82BnsB04GrimHaruspexDrawsForAnotherNontokenCreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	haruspex := effectstest.PushCatalogPermanent(g, me.ID, "Grim Haruspex", "Creature — Human Wizard", bnsGrimHaruspex, false)
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	die := func(id uuid.UUID, owner uuid.UUID) {
		effectstest.KillCreature(t, g, owner, id)
		effectstest.DrcSettle(t, g, me.ID)
	}
	before := me.Hand.Size()
	die(effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2), me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d after my nontoken creature died, want %d", got, before+1)
	}
	die(effectstest.B12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1), me.ID)
	die(effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3), opp.ID)
	die(effectstest.B12Permanent(g, me.ID, "Rock", "Artifact"), me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after a token, an opponent's creature and an artifact died, want %d", got, before+1)
	}
	die(haruspex, me.ID)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d after the Haruspex itself died, want %d", got, before+1)
	}
}

func TestS82BnsB04GrimHaruspexMorphIsTheFaceDownCastForThreeAndTurnsUpForB(t *testing.T) {
	spec, ok := effects.Lookup(bnsGrimHaruspex)
	if !ok || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "morph" {
		t.Fatalf("Grim Haruspex alternative costs = %+v, want one morph", spec.AlternativeCosts)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Grim Haruspex", TypeLine: "Creature — Human Wizard", OracleID: bnsGrimHaruspex, ManaCost: "{2}{B}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "morph", Strict: true}); err != nil {
		t.Fatalf("morph cast for {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the face-down Haruspex is not on the battlefield")
	}
	if e := c.Effective(); e.Power != 2 || e.Toughness != 2 {
		t.Errorf("face-down it is %d/%d, want 2/2", e.Power, e.Toughness)
	}
}

// --- Kinsbaile Cavalier ----------------------------------------------

func TestS82BnsB04KinsbaileCavalierGivesYourKnightsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Kinsbaile Cavalier", "Creature — Kithkin Knight", bnsKinsbaile, false)
	knight := effectstest.B12Creature(g, me.ID, "Human Knight", "Creature — Human Knight", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 2, 2)
	for name, want := range map[string]struct {
		id  uuid.UUID
		has bool
	}{"itself": {cav, true}, "my other Knight": {knight, true}, "my Bear": {bear, false}, "their Knight": {theirs, false}} {
		if got := effectstest.HasEffectiveKeyword(t, g, want.id, "double strike"); got != want.has {
			t.Errorf("%s: double strike=%v, want %v", name, got, want.has)
		}
	}
	effectstest.KillCreature(t, g, me.ID, cav)
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasEffectiveKeyword(t, g, knight, "double strike") {
		t.Error("my Knight kept double strike after the Cavalier left")
	}
}

// --- Magus of the Library --------------------------------------------

func TestS82BnsB04MagusOfTheLibraryDrawsOnlyWithExactlySevenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Library", "Creature — Human Wizard", bnsMagusLibrary, false)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	setHand := func(n int) {
		me.Hand.Cards = nil
		for i := 0; i < n; i++ {
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		}
	}
	activate := func() error {
		return g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{})
	}
	for _, n := range []int{6, 8} {
		setHand(n)
		if err := activate(); err == nil {
			t.Fatalf("the draw was activatable with %d cards in hand", n)
		}
		if c, _ := effectstest.BattlefieldCard(g, magus); c.Tapped {
			t.Fatalf("a refused activation (%d cards) tapped the Magus", n)
		}
	}
	setHand(7)
	if err := activate(); err != nil {
		t.Fatalf("activate with seven cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 8 {
		t.Errorf("hand %d after drawing from seven, want 8", got)
	}
}

func TestS82BnsB04MagusOfTheLibraryTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Library", "Creature — Human Wizard", bnsMagusLibrary, false)
	me.ManaPool = nil
	if err := g.ActivateManaAbility(me.ID, magus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Errorf("pool %+v, want one {C}", me.ManaPool)
	}
}

// --- Cavalier of Dawn ------------------------------------------------

func TestS82BnsB04CavalierOfDawnDestroysAPermanentAndGivesItsControllerAGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	golems := func() int {
		return effectstest.CountOnBattlefield(g, "Golem Token", opp.ID) + effectstest.CountOnBattlefield(g, "Golem", opp.ID)
	}
	cav := effectstest.B38CastCreature(t, g, "Cavalier of Dawn", "Creature — Elemental Knight", bnsCavalierOfDawn, 4, 6)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
		t.Fatal("a land was accepted as the target of \"nonland permanent\"")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the targeted creature was not destroyed")
	}
	if got := golems(); got != 1 {
		t.Errorf("the opponent has %d Golems, want 1", got)
	}
	if c, ok := effectstest.BattlefieldCard(g, cav); !ok || !effectstest.HasEffectiveKeyword(t, g, cav, "vigilance") {
		t.Errorf("Cavalier on the battlefield=%v with vigilance=%v", ok, effectstest.HasEffectiveKeyword(t, g, cav, "vigilance"))
	} else if e := c.Effective(); e.Power != 4 {
		t.Errorf("Cavalier power %d", e.Power)
	}
	var golem game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && (c.Name == "Golem Token" || c.Name == "Golem") {
			golem = c
		}
	}
	if e := golem.Effective(); e.Power != 3 || e.Toughness != 3 || len(e.Colors) != 0 || !golem.IsArtifact() || !golem.IsCreature() {
		t.Errorf("the token is %+v, want a 3/3 colorless artifact creature", golem)
	}
}

func TestS82BnsB04CavalierOfDawnGivesAGolemEvenToAnIndestructibleTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	god := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hardy", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.B38CastCreature(t, g, "Cavalier of Dawn", "Creature — Elemental Knight", bnsCavalierOfDawn, 4, 6)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, god)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, god); !ok {
		t.Error("an indestructible creature was destroyed")
	}
	if got := effectstest.CountOnBattlefield(g, "Golem Token", opp.ID) + effectstest.CountOnBattlefield(g, "Golem", opp.ID); got != 1 {
		t.Errorf("the opponent has %d Golems, want 1", got)
	}
}

func TestS82BnsB04CavalierOfDawnReturnsAnArtifactOrEnchantmentWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cav := effectstest.PushCatalogPermanent(g, me.ID, "Cavalier of Dawn", "Creature — Elemental Knight", bnsCavalierOfDawn, false)
	art, ench, creature := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: art, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: ench, Name: "Aura", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, me.ID, cav)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTarget(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: creature}); err == nil {
		t.Fatal("a creature card was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment card did not return to hand")
	}
	if me.Hand.Contains(art) || !me.Graveyard.Contains(art) {
		t.Error("the untargeted artifact moved")
	}
}

// --- Fiend Artisan ---------------------------------------------------

func TestS82BnsB04FiendArtisanGrowsWithCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := effectstest.PushCatalogPermanent(g, me.ID, "Fiend Artisan", "Creature — Nightmare", bnsFiendArtisan, false)
	pt := func() (int, int) {
		var p, tough int
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == art {
					e := c.Effective()
					p, tough = e.Power, e.Toughness
				}
			}
		})
		return p, tough
	}
	if p, tough := pt(); p != 1 || tough != 1 {
		t.Fatalf("empty graveyard: %d/%d, want 1/1", p, tough)
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if p, tough := pt(); p != 3 || tough != 3 {
		t.Errorf("two creature cards of mine in graveyards: %d/%d, want 3/3", p, tough)
	}
}

func TestS82BnsB04FiendArtisanSacrificesAnotherCreatureToFetchOneOfMatchingValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.PushCatalogPermanent(g, me.ID, "Fiend Artisan", "Creature — Nightmare", bnsFiendArtisan, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	small, big := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: big, Name: "Big Ogre", TypeLine: "Creature — Ogre", ManaCost: "{3}{B}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: small, Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	act := func(sac uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, art, 0, game.ActivateAbilityParams{XValue: 2, SacrificeIDs: []uuid.UUID{sac}})
	}
	if err := act(art); err == nil {
		t.Fatal("Fiend Artisan sacrificed itself to its own \"another creature\" cost")
	}
	if err := act(fodder); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, small); !ok {
		t.Error("the mana value 2 creature card was not put onto the battlefield")
	}
	if _, ok := effectstest.BattlefieldCard(g, big); ok {
		t.Error("the mana value 4 card came with X = 2")
	}
}

func TestS82BnsB04FiendArtisanOnlyAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := effectstest.PushCatalogPermanent(g, me.ID, "Fiend Artisan", "Creature — Nightmare", bnsFiendArtisan, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, art, 0, game.ActivateAbilityParams{XValue: 0, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("activated outside a main phase")
	}
	_ = opp
	if _, ok := effectstest.BattlefieldCard(g, fodder); !ok {
		t.Error("the refused activation sacrificed the creature")
	}
}

// --- Flesh Duplicate -------------------------------------------------

func fleshDuplicateCast(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Flesh Duplicate", TypeLine: "Creature — Shapeshifter Rebel", OracleID: bnsFleshDuplicate, ScryfallID: "00000000-0000-0000-0000-0000000000b4", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.ResolveWithCopyChoice(t, g, pick)
	return id
}

func TestS82BnsB04FleshDuplicateCopiesAndGainsVanishingThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bears := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	dup := fleshDuplicateCast(t, g, me, bears)
	c := effectstest.CopyBattlefieldCard(t, g, dup)
	if c.Name != "Grizzly Bears" {
		t.Fatalf("entered as %q, want a copy of Grizzly Bears", c.Name)
	}
	if got := c.Counters[game.CounterTime]; got != 3 {
		t.Fatalf("%d time counters, want 3", got)
	}
	if !slices.Contains(c.GrantedAbilities, game.GrantKey(cardsfl.FleshDuplicateVanishingGrant)) {
		t.Errorf("granted abilities %v lack the vanishing grant", c.GrantedAbilities)
	}
	if n := bnsTimeCounters(g, bears); n != 0 {
		t.Errorf("the copied Bears got %d time counters", n)
	}
	for want := 2; want >= 0; want-- {
		effectstest.Cmdr2B17NextUpkeepOfSeat0(t, g)
		if want > 0 {
			if got := bnsTimeCounters(g, dup); got != want {
				t.Fatalf("after an upkeep: %d time counters, want %d", got, want)
			}
		}
	}
	if _, here := effectstest.BattlefieldCard(g, dup); here {
		t.Error("the copy survived the removal of its last time counter")
	}
}

func bnsTimeCounters(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterTime]
			}
		}
	})
	return n
}

// A copied creature that already has vanishing keeps its own number: the
// copy enters with the counters the original card's vanishing gives it, not
// three more.
func TestS82BnsB04FleshDuplicateCopyingAVanishingCreatureKeepsItsOwnVanishing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hermit := effectstest.PushCatalogPermanent(g, me.ID, "Deep Forest Hermit", "Creature — Elf Druid", effectstest.Cmdr2B17HermitOracle, false)
	dup := fleshDuplicateCast(t, g, me, hermit)
	c := effectstest.CopyBattlefieldCard(t, g, dup)
	if c.Name != "Deep Forest Hermit" {
		t.Fatalf("entered as %q, want a copy of Deep Forest Hermit", c.Name)
	}
	if slices.Contains(c.GrantedAbilities, game.GrantKey(cardsfl.FleshDuplicateVanishingGrant)) {
		t.Error("a creature that already has vanishing was granted a second one")
	}
	if got := c.Counters[game.CounterTime]; got != 3 {
		t.Errorf("%d time counters, want the Hermit's own 3", got)
	}
}

// Declining the copy leaves the printed 0/0, which dies.
func TestS82BnsB04FleshDuplicateThatDeclinesTheCopyDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	dup := fleshDuplicateCast(t, g, me, uuid.Nil)
	if _, here := effectstest.BattlefieldCard(g, dup); here {
		t.Error("a declined Flesh Duplicate stayed on the battlefield as a 0/0")
	}
}
