package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b29_cards_test.go — Avatar: The Last Airbender Eternal, batch
// 29 (#5075): every card cast, triggered or activated through the engine.

const (
	avB29SeekerOracle     = "b58a70e2-3609-441a-993a-dccb9a2cf242"
	avB29NightmaresOracle = "0403bfb0-2174-4360-994d-68d8ca96fc55"
	avB29YueOracle        = "08c59e5e-d01c-49b7-a812-8ea75e9e945d"
	avB29SprayOracle      = "74b8976a-15cb-4604-a65d-ca9735cd6000"
	avB29WanShiTongOracle = "6927bd2e-cb05-487a-af86-89ede6e8dabd"
	avB29WhirlwindOracle  = "a4477852-66eb-4a54-b4c1-760cbf624f82"
	avB29CensorOracle     = "f13a8621-4d88-4c1f-8b6e-80f8ef383cd2"
	avB29OzaiOracle       = "5945306d-e792-4eae-999b-dfb44b10f062"
	avB29OccupationOracle = "2a99c0d6-3200-44ff-8d4c-f7a38cecb05f"
	avB29GiantFlyOracle   = "439ae30b-348d-480c-8d3d-fe105167bdc4"
)

func TestAvatarB29EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB29SeekerOracle:     "Dutiful Knowledge Seeker",
		avB29NightmaresOracle: "Nightmares and Daydreams",
		avB29YueOracle:        "Princess Yue",
		avB29SprayOracle:      "Unagi's Spray",
		avB29WanShiTongOracle: "Wan Shi Tong, All-Knowing",
		avB29WhirlwindOracle:  "Whirlwind Technique",
		avB29CensorOracle:     "Dai Li Censor",
		avB29OzaiOracle:       "Fire Lord Ozai",
		avB29OccupationOracle: "Fire Nation Occupation",
		avB29GiantFlyOracle:   "Giant Fly",
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

// avB29Live is a battlefield card with its layered characteristics read.
func avB29Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				_ = c.Effective()
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return out
}

func avB29Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func avB29Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && !stackFullyEmpty(g) && len(g.PendingChoices) == 0; i++ {
		passPriorityAroundTable(t, g)
	}
}

// avB29Hand puts a card into p's hand and returns its ID.
func avB29Hand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// --- Dutiful Knowledge Seeker ------------------------------------------------

func TestAvatarB29SeekerGrowsWhenACardIsPutIntoALibrary(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	seeker := b12Push(g, me.ID, "Dutiful Knowledge Seeker", "Creature — Fox Spirit", avB29SeekerOracle, 2, 2)
	dead := pushGraveyardCardForTest(opp, "Old Bear")

	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err == nil {
		t.Fatal("{3} was paid from an empty pool")
	}
	fillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	avB29Settle(t, g)
	if opp.Library.Cards[0].InstanceID != dead {
		t.Fatal("the card is not on the bottom of its owner's library")
	}
	if n := countersOn(g, seeker, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after one card was put into a library, want 1", n)
	}

	// Two permanents tucked by one instruction are one batch: one counter.
	a := b12Permanent(g, opp.ID, "Rock A", "Artifact")
	b := b12Permanent(g, opp.ID, "Rock B", "Artifact")
	g.BeginEventBatchForTest()
	g.WithWriteLock(func() {
		_ = g.TuckToLibraryForEffect(a, true)
		_ = g.TuckToLibraryForEffect(b, false)
	})
	g.BeginEventBatchForTest()
	avB29Settle(t, g)
	if n := countersOn(g, seeker, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a two-card batch, want 2", n)
	}

	// A mill puts nothing into a library.
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	g.BeginEventBatchForTest()
	avB29Settle(t, g)
	if n := countersOn(g, seeker, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a mill, want still 2", n)
	}
}

// --- Wan Shi Tong, All-Knowing -------------------------------------------------

func TestAvatarB29WanShiTongTucksAndMakesTwoSpirits(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	owl := avB29Hand(g, me, game.Card{Name: "Wan Shi Tong, All-Knowing", TypeLine: "Legendary Creature — Bird Spirit",
		OracleID: avB29WanShiTongOracle, ManaCost: "{3}{U}{U}", Power: 4, Toughness: 4})
	if err := g.CastSpell(me.ID, owl, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("a land was a legal target")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil || ask.LibraryTopDepth != 2 {
		t.Fatalf("the OWNER is asked second from the top or the bottom: %+v", ask)
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	g.BeginEventBatchForTest()
	avB29Settle(t, g)
	if opp.Library.Cards[0].InstanceID != bear {
		t.Fatal("the creature is not on the bottom of its owner's library")
	}
	if n := avB29Named(g, me.ID, "Spirit"); n != 2 {
		t.Fatalf("%d Spirits, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0) {
			t.Errorf("Spirit %d/%d %v, want a colorless 1/1", c.Power, c.Toughness, c.Colors)
		}
	}
}

// --- Nightmares and Daydreams ----------------------------------------------------

func TestAvatarB29NightmaresMillsTheSpellsManaValueUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	saga := castCatalogSpell(t, g, "Nightmares and Daydreams", "Enchantment — Saga", avB29NightmaresOracle, nil)
	avB29Settle(t, g)
	if !onBattlefield(g, saga) {
		t.Fatal("the Saga did not resolve")
	}

	// A creature spell does not fire it.
	library := opp.Library.Size()
	bear := avB29Hand(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell fired the delayed trigger")
	}
	avB29Settle(t, g)

	// A three-mana instant mills the target three.
	bolt := avB29Hand(g, me, game.Card{Name: "Big Instant", TypeLine: "Instant", ManaCost: "{1}{U}{U}"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the instant did not fire the delayed trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	avB29Settle(t, g)
	if got := library - opp.Library.Size(); got != 3 {
		t.Fatalf("the target milled %d cards, want 3", got)
	}

	// It lasts through the turn and ends as your next turn begins: a
	// sorcery now still fires it.
	sorc := avB29Hand(g, me, game.Card{Name: "Small Sorcery", TypeLine: "Sorcery", ManaCost: "{U}"})
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the sorcery: %v", err)
	}
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("a second instant or sorcery this turn did not fire it again")
	}
}

func TestAvatarB29NightmaresChapterFourDrawsThreeOffATwentyCardGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name  string
		grave int
		want  int
	}{{"nineteen", 19, 1}, {"twenty", 20, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
			for i := 0; i < tc.grave; i++ {
				pushGraveyardCardForTest(opp, "Old Card")
			}
			saga := pushCatalogPermanent(g, me.ID, "Nightmares and Daydreams", "Enchantment — Saga", avB29NightmaresOracle, false)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == saga {
						if g.Battlefield.Cards[i].Counters == nil {
							g.Battlefield.Cards[i].Counters = map[string]int{}
						}
						g.Battlefield.Cards[i].Counters[game.CounterLore] = 3
					}
				}
			})
			advanceToPrecombatMainOf(t, g, seat)
			hand := me.Hand.Size()
			// The precombat main is reached after the draw step, whose lore
			// counter fired chapter IV.
			avB29Settle(t, g)
			if got := me.Hand.Size() - hand; got != tc.want {
				t.Errorf("chapter IV drew %d, want %d", got, tc.want)
			}
		})
	}
}

// --- Princess Yue -----------------------------------------------------------------

func TestAvatarB29PrincessYueReturnsAsALandNamedMoon(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yue := b12Push(g, me.ID, "Princess Yue", "Legendary Creature — Human Noble Ally", avB29YueOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(yue) })
	avB29Settle(t, g)
	if !onBattlefield(g, yue) {
		t.Fatal("Princess Yue did not return")
	}
	moon := avB29Live(t, g, yue)
	if moon.Effective().Name != "Moon" || !moon.IsLand() || moon.IsCreature() || !moon.Tapped {
		t.Fatalf("returned as %q land %v creature %v tapped %v, want a tapped land named Moon",
			moon.Effective().Name, moon.IsLand(), moon.IsCreature(), moon.Tapped)
	}
	eff := moon.Effective()
	if !slices.Contains(eff.Supertypes, "Legendary") {
		t.Errorf("supertypes %v, want her still legendary", eff.Supertypes)
	}
	if slices.Contains(eff.Subtypes, "Human") || slices.Contains(eff.Subtypes, "Ally") {
		t.Errorf("subtypes %v: the creature types go with Creature", eff.Subtypes)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(yue) })
	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(yue)
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	idx := -1
	for i, a := range abs {
		if a.Produced == "{C}" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("Moon's mana abilities %+v, want the granted {T}: Add {C}", abs)
	}
	if err := g.ActivateManaAbility(me.ID, yue, idx, game.ManaAbilityParams{Ref: origins.Ref(idx)}); err != nil {
		t.Fatalf("tap Moon for {C}: %v", err)
	}
	mkWantPool(t, me, map[string]int{"C": 1})

	// Moon is a land: it dies to nothing, and if it is destroyed it was not
	// a nonland creature, so it stays in the graveyard.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(yue) })
	avB29Settle(t, g)
	if onBattlefield(g, yue) || !me.Graveyard.Contains(yue) {
		t.Error("a destroyed Moon came back")
	}
}

func TestAvatarB29PrincessYueScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	yue := b12Push(g, me.ID, "Princess Yue", "Legendary Creature — Human Noble Ally", avB29YueOracle, 3, 2)
	if err := g.ActivateCatalogAbility(me.ID, yue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %+v, want two cards", c)
	}
}

// --- Unagi's Spray --------------------------------------------------------------------

func TestAvatarB29UnagisSprayDrawsOnlyWithASeaCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		types string
		draw  int
	}{{"no sea creature", "Creature — Bear", 0}, {"a Serpent", "Creature — Serpent", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			b12Creature(g, me.ID, "My Creature", tc.types, 1, 1)
			ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
			hand := me.Hand.Size()
			castCatalogSpell(t, g, "Unagi's Spray", "Instant", avB29SprayOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ogre}})
			passPriorityAroundTable(t, g)
			if p := avB29Live(t, g, ogre).CurrentPower(); p != 1 {
				t.Errorf("power %d, want 1", p)
			}
			if got := me.Hand.Size() - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
		})
	}
}

// --- Whirlwind Technique ------------------------------------------------------------------

func TestAvatarB29WhirlwindTechniqueDrawsDiscardsAndAirbends(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Whirlwind Technique", "Instant — Lesson", avB29WhirlwindOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
		{Kind: game.TargetCard, ID: a, Slot: 1},
		{Kind: game.TargetCard, ID: b, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	avB29Settle(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (draw two, discard one)", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !g.Exile.Contains(id) {
			t.Fatalf("creature %s was not airbended", id)
		}
	}
	// Its owner may cast it for {2}.
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(a); ok {
			perm = g.CastPermissionForLocked(opp.ID, c, game.ZoneExile)
		}
	})
	if perm == nil || perm.Cost != AirbendCost {
		t.Errorf("the owner's permission %+v, want a {2} cast", perm)
	}
}

// --- Dai Li Censor -------------------------------------------------------------------------

func TestAvatarB29DaiLiCensorSacrificesForAPumpOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	censor := b12Push(g, me.ID, "Dai Li Censor", "Creature — Human Soldier Advisor", avB29CensorOracle, 2, 1)
	a := b12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	b := b12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1)
	fillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, censor, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{censor}}); err == nil {
		t.Fatal("the Censor sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, censor, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{a}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, a) {
		t.Error("the fodder was not sacrificed")
	}
	live := avB29Live(t, g, censor)
	if live.CurrentPower() != 4 || live.CurrentToughness() != 3 {
		t.Errorf("Censor %d/%d, want 4/3", live.CurrentPower(), live.CurrentToughness())
	}
	if err := g.ActivateCatalogAbility(me.ID, censor, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{b}}); err == nil {
		t.Error("activated twice in one turn")
	}
}

// --- Giant Fly -------------------------------------------------------------------------------

func TestAvatarB29GiantFlyGrowsWhenYouSacrificeAnother(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	fly := b12Push(g, me.ID, "Giant Fly", "Creature — Insect", avB29GiantFlyOracle, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	avB29Settle(t, g)
	if p := avB29Live(t, g, fly).CurrentPower(); p != 3 {
		t.Errorf("power %d after a sacrifice, want 3", p)
	}
	if tough := avB29Live(t, g, fly).CurrentToughness(); tough != 2 {
		t.Errorf("toughness %d, want 2 (+1/+0)", tough)
	}
}

// --- Fire Nation Occupation ---------------------------------------------------------------------

func TestAvatarB29FireNationOccupationMakesSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	castCatalogSpell(t, g, "Fire Nation Occupation", "Enchantment", avB29OccupationOracle, nil)
	avB29Settle(t, g)
	if n := avB29Named(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers after it entered, want 1", n)
	}
	// A spell on your own turn makes none.
	inst := avB29Hand(g, me, game.Card{Name: "Quick", TypeLine: "Instant", ManaCost: "{B}"})
	if err := g.CastSpell(me.ID, inst, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast on your turn: %v", err)
	}
	avB29Settle(t, g)
	if n := avB29Named(g, me.ID, "Soldier"); n != 1 {
		t.Fatalf("%d Soldiers after a spell on your own turn, want 1", n)
	}
	// One during an opponent's turn makes one.
	advanceToPrecombatMainOf(t, g, (seat+1)%4)
	inst2 := avB29Hand(g, me, game.Card{Name: "Quick Two", TypeLine: "Instant", ManaCost: "{B}"})
	if err := g.CastSpell(me.ID, inst2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast on an opponent's turn: %v", err)
	}
	avB29Settle(t, g)
	if n := avB29Named(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("%d Soldiers after a spell on an opponent's turn, want 2", n)
	}
}

// --- Fire Lord Ozai -------------------------------------------------------------------------------

func TestAvatarB29OzaiSacrificesForRedManaThatLastsTheCombat(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ozai := b12Push(g, me.ID, "Fire Lord Ozai", "Legendary Creature — Human Noble", avB29OzaiOracle, 4, 4)
	ogre := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	declareAttack(t, g, opp.ID, ozai)
	avB29Settle(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("no sacrifice prompt")
	}
	answerOwnPermanents(t, g, me.ID, ogre)
	avB29Settle(t, g)
	if onBattlefield(g, ogre) {
		t.Fatal("the Ogre was not sacrificed")
	}
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepCombatDamage)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

func TestAvatarB29OzaiPlaysOneExiledCardForFree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ozai := b12Push(g, me.ID, "Fire Lord Ozai", "Legendary Creature — Human Noble", avB29OzaiOracle, 4, 4)
	var tops []uuid.UUID
	for i := 1; i < 4; i++ {
		opp := g.Seats[(seat+i)%4]
		id := uuid.New()
		opp.Library.PushTop(game.Card{InstanceID: id, Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{5}{R}",
			Owner: opp.ID, Controller: opp.ID})
		tops = append(tops, id)
	}
	fillPoolColored(me, "B", 5)
	if err := g.ActivateCatalogAbility(me.ID, ozai, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{6} was paid with five mana")
	}
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, ozai, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	avB29Settle(t, g)
	for _, id := range tops {
		if !g.Exile.Contains(id) {
			t.Fatal("the top card of each opponent's library was not exiled")
		}
	}
	if err := g.CastSpell(me.ID, tops[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the first exiled card for free: %v", err)
	}
	avB29Settle(t, g)
	if err := g.CastSpell(me.ID, tops[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("cast a second of those cards")
	}
}
