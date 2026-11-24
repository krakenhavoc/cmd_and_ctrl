package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b02_cards_test.go — GAP2 batch 2 (#5948, S82): haunt (#6083),
// the replacements of losing the game and of planeswalking (#6096), and
// timed graveyard escape and unearth grants (#6086). Each test plays the
// real catalog entry.

const (
	gap2B02EuthanistOracle = "3cc82a02-2906-4411-aebf-474b0735bb53"
	gap2B02HunterOracle    = "a731f5bc-892c-4857-9590-4fa31c1caf00"
	gap2B02PontiffOracle   = "3620ef25-d9e2-4460-b25f-592ff7df681b"
	gap2B02ThroneOracle    = "a68f41a2-37ee-46e1-bf3c-164c78b003ca"
	gap2B02SusanOracle     = "0f21f3b1-aee9-4272-ad02-067aa5f6a7c8"
	gap2B02DialOracle      = "6316158c-2f66-4dc4-b78c-c6ff7c39bb77"
	gap2B02WelcomeOracle   = "a4b37d16-95b3-4143-a0b2-ad9f2aba91f8"
	gap2B02ArkOracle       = "bb1b4786-3804-486b-9fd9-85848d24d362"
)

func TestGap2B02CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap2B02EuthanistOracle:      "Orzhov Euthanist",
		gap2B02HunterOracle:         "Blind Hunter",
		gap2B02PontiffOracle:        "Orzhov Pontiff",
		gap2B02ThroneOracle:         "The Golden Throne",
		gap2B02SusanOracle:          "Susan Foreman",
		gap2B02DialOracle:           "Confession Dial",
		gap2B02WelcomeOracle:        "Welcome to . . .",
		gap2B02WelcomeOracle + "#1": "Jurassic Park",
		gap2B02ArkOracle:            "Ghost Ark",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// gap2B02Settle resolves the stack, answering each target pick with
// pick, and stops at any other prompt.
func gap2B02Settle(t *testing.T, g *game.Game, pick func(c *game.PendingChoice) (game.TargetRef, bool)) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if c := effectstest.FirstPendingTarget(g); c != nil {
			ref, ok := pick(c)
			if !ok {
				t.Fatalf("unexpected target pick: cards %v players %v", c.PickTargetCards, c.PickTargetPlayers)
			}
			if err := g.ResolvePickTarget(c.ID, c.Chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("the stack never settled")
}

func gap2B02Card(id uuid.UUID) func(*game.PendingChoice) (game.TargetRef, bool) {
	return func(c *game.PendingChoice) (game.TargetRef, bool) {
		if !effectstest.HasID(c.PickTargetCards, id) {
			return game.TargetRef{}, false
		}
		return game.TargetRef{Kind: game.TargetCard, ID: id}, true
	}
}

func gap2B02Player(id uuid.UUID) func(*game.PendingChoice) (game.TargetRef, bool) {
	return func(c *game.PendingChoice) (game.TargetRef, bool) {
		if !effectstest.HasID(c.PickTargetPlayers, id) {
			return game.TargetRef{}, false
		}
		return game.TargetRef{Kind: game.TargetPlayer, ID: id}, true
	}
}

// gap2B02Kill puts a creature into its owner's graveyard from the
// battlefield and puts what that triggered on its way.
func gap2B02Kill(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	effectstest.KillCreature(t, g, owner, id)
	g.RunStateChecksForTest()
}

func gap2B02Exiled(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

// gap2B02Haunt kills a haunt creature and has it haunt `victim`.
func gap2B02Haunt(t *testing.T, g *game.Game, owner, haunter, victim uuid.UUID) {
	t.Helper()
	gap2B02Kill(t, g, owner, haunter)
	gap2B02Settle(t, g, gap2B02Card(victim))
	c, ok := gap2B02Exiled(g, haunter)
	if !ok {
		t.Fatal("the haunt creature is not in exile")
	}
	if c.HauntedID != victim {
		t.Fatalf("it haunts %v, want %v", c.HauntedID, victim)
	}
}

// --- Blind Hunter --------------------------------------------------------

func TestGap2B02BlindHunterHauntsAndDrainsWhenTheHauntedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.DftABPush(g, me.ID, "Blind Hunter", "Creature — Bat", gap2B02HunterOracle, 2, 2)
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 2, 2)
	gap2B02Haunt(t, g, me.ID, hunter, bear)

	// Another creature dying is not the creature it haunts.
	myLife, oppLife := me.Life, opp.Life
	gap2B02Kill(t, g, opp.ID, other)
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("another creature's death drained: me %d→%d, opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}

	gap2B02Kill(t, g, opp.ID, bear)
	gap2B02Settle(t, g, gap2B02Player(opp.ID))
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("after the haunted creature died: me %d→%d, opp %d→%d; want +2 and -2", myLife, me.Life, oppLife, opp.Life)
	}
	if _, ok := gap2B02Exiled(g, hunter); !ok {
		t.Error("the haunting card left exile")
	}
}

func TestGap2B02BlindHunterDrainsAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	effectstest.DftABCast(t, g, me, game.Card{Name: "Blind Hunter", TypeLine: "Creature — Bat", OracleID: gap2B02HunterOracle,
		ManaCost: "{2}{W}{B}", Power: 2, Toughness: 2})
	gap2B02Settle(t, g, gap2B02Player(opp.ID))
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("enters: me %d→%d, opp %d→%d; want +2 and -2", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestGap2B02HauntEndsWhenTheHauntedCreatureBecomesANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.DftABPush(g, me.ID, "Blind Hunter", "Creature — Bat", gap2B02HunterOracle, 2, 2)
	bear := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	gap2B02Haunt(t, g, me.ID, hunter, bear)
	// Blinked: the creature that comes back is a new object (CR 400.7).
	effectstest.ExileCreature(t, g, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatalf("return: %v", err)
	}
	myLife, oppLife := me.Life, opp.Life
	gap2B02Kill(t, g, opp.ID, bear)
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if me.Life != myLife || opp.Life != oppLife {
		t.Errorf("a new object's death triggered the haunt: me %d→%d, opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
}

func TestGap2B02HauntWithNoOtherCreatureStaysInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hunter := effectstest.DftABPush(g, me.ID, "Blind Hunter", "Creature — Bat", gap2B02HunterOracle, 2, 2)
	gap2B02Kill(t, g, me.ID, hunter)
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if !me.Graveyard.Contains(hunter) {
		t.Error("with no creature to haunt, the card left its owner's graveyard")
	}
}

// --- Orzhov Euthanist ----------------------------------------------------

func TestGap2B02OrzhovEuthanistDestroysADamagedCreatureWhenTheHauntedOneDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	euth := effectstest.DftABPush(g, me.ID, "Orzhov Euthanist", "Creature — Human Assassin", gap2B02EuthanistOracle, 2, 2)
	haunted := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	hurt := effectstest.DftABPush(g, opp.ID, "Opp Wurm", "Creature — Wurm", "", 3, 3)
	whole := effectstest.DftABPush(g, opp.ID, "Opp Elk", "Creature — Elk", "", 3, 3)
	gap2B02Haunt(t, g, me.ID, euth, haunted)
	effectstest.Damage(t, g, whole, hurt, 1)
	gap2B02Kill(t, g, opp.ID, haunted)
	c := effectstest.FirstPendingTarget(g)
	if c == nil {
		t.Fatal("the haunted creature's death asked for no target")
	}
	if c.Chooser != me.ID {
		t.Errorf("the trigger is %v's, want the haunting card's owner", c.Chooser)
	}
	if effectstest.HasID(c.PickTargetCards, whole) || !effectstest.HasID(c.PickTargetCards, hurt) {
		t.Fatalf("offered %v, want the damaged creature and not the undamaged one", c.PickTargetCards)
	}
	gap2B02Settle(t, g, gap2B02Card(hurt))
	if !opp.Graveyard.Contains(hurt) || !g.Battlefield.Contains(whole) {
		t.Error("the damaged creature was not destroyed, or the other one was")
	}
}

// --- Orzhov Pontiff ------------------------------------------------------

func TestGap2B02OrzhovPontiffChoosesAModeWhenTheHauntedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pontiff := effectstest.DftABPush(g, me.ID, "Orzhov Pontiff", "Creature — Human Cleric", gap2B02PontiffOracle, 1, 1)
	mine := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DftABPush(g, opp.ID, "Opp Wurm", "Creature — Wurm", "", 3, 3)
	haunted := effectstest.DftABPush(g, opp.ID, "Opp Bear", "Creature — Bear", "", 2, 2)
	gap2B02Haunt(t, g, me.ID, pontiff, haunted)
	gap2B02Kill(t, g, opp.ID, haunted)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatal("the haunting card's owner was not asked for a mode")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, mine); p != 2 || th != 2 {
		t.Errorf("my creature is %d/%d, want 2/2", p, th)
	}
}

func TestGap2B02OrzhovPontiffPumpsYourCreaturesAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DftABPush(g, opp.ID, "Opp Wurm", "Creature — Wurm", "", 3, 3)
	effectstest.AdvanceToMain(t, g)
	pontiff := effectstest.DftABCast(t, g, me, game.Card{Name: "Orzhov Pontiff", TypeLine: "Creature — Human Cleric", OracleID: gap2B02PontiffOracle,
		ManaCost: "{1}{W}{B}", Power: 1, Toughness: 1})
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatal("no mode asked as it entered")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if p, th := effectstest.EffectivePTOf(t, g, mine); p != 3 || th != 3 {
		t.Errorf("my creature is %d/%d, want 3/3", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, pontiff); p != 2 || th != 2 {
		t.Errorf("the Pontiff is %d/%d, want 2/2", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 3 || th != 3 {
		t.Errorf("their creature is %d/%d, want 3/3", p, th)
	}
}

// --- The Golden Throne ---------------------------------------------------

func TestGap2B02GoldenThroneReplacesALossByLifeOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.DftABPush(g, me.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	g.WithWriteLock(func() { me.Life = -3 })
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("the player lost with The Golden Throne on the battlefield")
	}
	if me.Life != 1 {
		t.Errorf("life %d, want 1", me.Life)
	}
	if _, ok := gap2B02Exiled(g, throne); !ok {
		t.Error("The Golden Throne was not exiled")
	}
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Error("with the Throne gone, the player did not lose")
	}
}

func TestGap2B02GoldenThroneDoesNotStopPoisonForLong(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.DftABPush(g, me.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	g.WithWriteLock(func() {
		if me.Counters == nil {
			me.Counters = map[string]int{}
		}
		me.Counters[game.CounterPoison] = 10
		me.Poison = 10
	})
	from := len(g.Events)
	g.RunStateChecksForTest()
	// The first check's loss is replaced (the Throne goes to exile, which
	// the player's departure then takes out of the game, CR 800.4a); the
	// next check's is not.
	var replaced []game.Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventGameLossReplaced {
			replaced = append(replaced, ev)
		}
	}
	if len(replaced) != 1 || replaced[0].CardID != throne || replaced[0].Label != string(game.LossPoison) {
		t.Errorf("replaced losses %+v, want the one poison loss by the Throne", replaced)
	}
	if effectstest.BattlefieldCardFor(g, throne) != nil {
		t.Error("The Golden Throne is still on the battlefield")
	}
	if !me.Eliminated || me.LossCause != game.LossPoison {
		t.Errorf("eliminated %v by %q, want a loss to poison at the next check", me.Eliminated, me.LossCause)
	}
}

func TestGap2B02GoldenThroneReplacesALossByAnEffectButNotAConcession(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.DftABPush(g, me.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	var lost bool
	g.WithWriteLock(func() { lost, _ = g.LoseTheGameForEffect(me.ID, uuid.Nil) })
	if lost || me.Eliminated || me.Life != 1 {
		t.Fatalf("lost %v, eliminated %v, life %d; want the loss replaced and life 1", lost, me.Eliminated, me.Life)
	}
	if _, ok := gap2B02Exiled(g, throne); !ok {
		t.Error("The Golden Throne was not exiled")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[1]
	effectstest.DftABPush(g2, me2.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	if err := g2.Concede(me2.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !me2.Eliminated {
		t.Error("a concession was replaced")
	}
}

func TestGap2B02GoldenThroneOnlyKeepsItsControllerIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DftABPush(g, me.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	g.WithWriteLock(func() { opp.Life = 0 })
	g.RunStateChecksForTest()
	if !opp.Eliminated {
		t.Error("an opponent was kept in the game by my Throne")
	}
}

func TestGap2B02GoldenThroneSacrificesACreatureForThreeMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.DftABPush(g, me.ID, "The Golden Throne", "Legendary Artifact", gap2B02ThroneOracle, 0, 0)
	bear := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, throne, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Colors: []string{"W", "U", "B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 3 || !me.Graveyard.Contains(bear) {
		t.Errorf("pool %v, bear in graveyard %v; want three mana and the bear sacrificed", me.ManaPool, me.Graveyard.Contains(bear))
	}
}

// --- Susan Foreman -------------------------------------------------------

func gap2B02PlaneGame(t *testing.T) *game.Game {
	t.Helper()
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneFixture{
		Oracle: uuid.NewSHA1(uuid.NameSpaceOID, []byte("gap2-b02 plane")).String(), Name: "Test Plane", TypeLine: "Plane — Testworld"})
	return g
}

func gap2B02TopTwo(g *game.Game) (top, second uuid.UUID) {
	g.ReadSnapshot(func() {
		n := len(g.PlanarDeck.Cards)
		top, second = g.PlanarDeck.Cards[n-1].InstanceID, g.PlanarDeck.Cards[n-2].InstanceID
	})
	return top, second
}

func gap2B02FaceUp(g *game.Game) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		if n := len(g.PlanarFaceUp.Cards); n > 0 {
			id = g.PlanarFaceUp.Cards[n-1].InstanceID
		}
	})
	return id
}

func TestGap2B02SusanForemanOrdersTheTopTwoBeforeYouPlaneswalk(t *testing.T) {
	g := gap2B02PlaneGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABPush(g, me.ID, "Susan Foreman", "Legendary Creature — Time Lord", gap2B02SusanOracle, 1, 1)
	top, second := gap2B02TopTwo(g)
	var err error
	g.WithWriteLock(func() { err = g.PlaneswalkForEffect(me.ID) })
	if err != nil {
		t.Fatalf("PlaneswalkForEffect: %v", err)
	}
	if gap2B02FaceUp(g) == top || gap2B02FaceUp(g) == second {
		t.Fatal("planeswalked before the look-and-order was answered")
	}
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("Susan's controller was not asked")
	}
	g.ReadSnapshot(func() {
		n := len(g.PlanarDeck.Cards)
		if !g.PlanarDeck.Cards[n-1].IsKnownTo(me.ID) || !g.PlanarDeck.Cards[n-2].IsKnownTo(me.ID) {
			t.Error("the top two cards were not looked at")
		}
	})
	// The top card goes to the bottom: the second is the plane.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if got := gap2B02FaceUp(g); got != second {
		t.Errorf("planeswalked to %v, want the card that was second (%v)", got, second)
	}
	g.ReadSnapshot(func() {
		found := false
		for _, c := range g.PlanarDeck.Cards[:2] {
			if c.InstanceID == top {
				found = true
			}
		}
		if !found {
			t.Error("the card put on the bottom is not at the bottom of the planar deck")
		}
	})
}

func TestGap2B02SusanForemanCanKeepTheTopCard(t *testing.T) {
	g := gap2B02PlaneGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABPush(g, me.ID, "Susan Foreman", "Legendary Creature — Time Lord", gap2B02SusanOracle, 1, 1)
	top, _ := gap2B02TopTwo(g)
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if got := gap2B02FaceUp(g); got != top {
		t.Errorf("planeswalked to %v, want the top card (%v)", got, top)
	}
}

func TestGap2B02SusanForemanOnlyReplacesHerControllersPlaneswalk(t *testing.T) {
	g := gap2B02PlaneGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DftABPush(g, opp.ID, "Susan Foreman", "Legendary Creature — Time Lord", gap2B02SusanOracle, 1, 1)
	top, _ := gap2B02TopTwo(g)
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if effectstest.LatestOptionPickFor(g, opp.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Fatal("an opponent's Susan replaced the planeswalk")
	}
	if got := gap2B02FaceUp(g); got != top {
		t.Errorf("planeswalked to %v, want the top card (%v)", got, top)
	}
}

func TestGap2B02SusanForemanTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	susan := effectstest.DftABPush(g, me.ID, "Susan Foreman", "Legendary Creature — Time Lord", gap2B02SusanOracle, 1, 1)
	if err := g.ActivateManaAbility(me.ID, susan, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "G" {
		t.Errorf("pool %v, want {G}", me.ManaPool)
	}
}

// --- Confession Dial -----------------------------------------------------

func TestGap2B02ConfessionDialGivesALegendaryCreatureCardEscape(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dial := effectstest.DftABPush(g, me.ID, "Confession Dial", "Artifact", gap2B02DialOracle, 0, 0)
	legend, plain := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "Dead Legend", TypeLine: "Legendary Creature — Human", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: plain, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for range 3 {
		id := uuid.New()
		fodder = append(fodder, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dial, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("a nonlegendary creature card was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, dial, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(legend)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder, Strict: true}); err == nil {
		t.Fatal("the card the Dial did not target was escaped")
	}
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder[:2], Strict: true}); err == nil {
		t.Fatal("escaped by exiling only two other cards")
	}
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder, Strict: true}); err != nil {
		t.Fatalf("escape for its mana cost plus three exiled cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(legend) {
		t.Error("the escaped legend did not resolve")
	}
	for _, id := range fodder {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled to pay the escape", id)
		}
	}
}

func TestGap2B02ConfessionDialEscapeEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dial := effectstest.DftABPush(g, me.ID, "Confession Dial", "Artifact", gap2B02DialOracle, 0, 0)
	legend := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: legend, Name: "Dead Legend", TypeLine: "Legendary Creature — Human", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for range 3 {
		id := uuid.New()
		fodder = append(fodder, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, dial, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(legend)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder, Strict: true}); err == nil {
		t.Error("the granted escape outlived the turn")
	}
}

func TestGap2B02ConfessionDialSurveilsThreeAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Confession Dial", TypeLine: "Artifact", OracleID: gap2B02DialOracle, ManaCost: "{3}"})
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt as the Dial entered")
	}
	if n := len(c.ScryCards); n != 3 {
		t.Errorf("surveil looks at %d cards, want 3", n)
	}
}

// --- Ghost Ark -----------------------------------------------------------

func gap2B02ArkScene(t *testing.T) (g *game.Game, me *game.Player, ark, crewer, golem, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	ark = effectstest.DftABPush(g, me.ID, "Ghost Ark", "Artifact — Vehicle", gap2B02ArkOracle, 3, 3)
	crewer = effectstest.DftABPush(g, me.ID, "Crew Bear", "Creature — Bear", "", 2, 2)
	golem, bear = uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Dead Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{5}",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	return g, me, ark, crewer, golem, bear
}

func TestGap2B02GhostArkGivesArtifactCreatureCardsUnearthWhenCrewed(t *testing.T) {
	g, me, ark, crewer, golem, bear := gap2B02ArkScene(t)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("unearth before the Ark was crewed")
	}
	effectstest.CrewBy(t, g, me.ID, ark, crewer)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a nonartifact creature card gained unearth")
	}
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, golem)
	if c == nil {
		t.Fatal("the unearthed golem is not on the battlefield")
	}
	if !game.HasKeyword(c, "haste") {
		t.Error("the unearthed golem has no haste")
	}
}

func TestGap2B02GhostArkUnearthEndsWithTheTurn(t *testing.T) {
	g, me, ark, crewer, golem, _ := gap2B02ArkScene(t)
	effectstest.CrewBy(t, g, me.ID, ark, crewer)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the granted unearth outlived the turn")
	}
}

// --- Welcome to . . . // Jurassic Park -----------------------------------

func gap2B02WelcomeRow() cards.Card {
	return effectstest.TransformRow(gap2B02WelcomeOracle, "Welcome to . . .", "Enchantment — Saga", "{1}{G}{G}",
		"Jurassic Park", "Legendary Land", "", "", []string{"G"})
}

// gap2B02FirstOppPick picks `id` for the clause that offers it, and
// declines the clauses that offer nothing else.
func gap2B02FirstOppPick(t *testing.T, g *game.Game, me uuid.UUID, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		c := effectstest.FirstPendingTarget(g)
		if c == nil {
			return
		}
		if effectstest.HasID(c.PickTargetCards, id) {
			if err := g.ResolvePickTarget(c.ID, me, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.ResolvePickTargets(c.ID, me, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
	}
}

func TestGap2B02WelcomeToJurassicPark(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	rock := effectstest.DftABPush(g, opp.ID, "Opp Rock", "Artifact", "", 0, 0)
	saga := effectstest.ImportAndCast(t, g, gap2B02WelcomeRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	gap2B02FirstOppPick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, rock)
	if c == nil || !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Wall") || !game.HasKeyword(c, "defender") {
		t.Fatalf("chapter I: the rock is %+v, want a Wall artifact creature with defender", c)
	}
	if p, th := effectstest.EffectivePTOf(t, g, rock); p != 0 || th != 4 {
		t.Errorf("the rock is %d/%d, want 0/4", p, th)
	}

	// Chapter II: a 3/3 green trampling Dinosaur with haste this turn.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	dino := effectstest.FindBattlefieldByName(g, "Dinosaur")
	if dino == uuid.Nil {
		t.Fatal("chapter II made no Dinosaur")
	}
	d := effectstest.BattlefieldCardFor(g, dino)
	if !game.HasKeyword(d, "trample") || !game.HasKeyword(d, "haste") || d.CurrentPower() != 3 || d.CurrentToughness() != 3 {
		t.Errorf("the Dinosaur is %d/%d trample %v haste %v", d.CurrentPower(), d.CurrentToughness(),
			game.HasKeyword(d, "trample"), game.HasKeyword(d, "haste"))
	}

	// Chapter III: the Walls are destroyed while the Saga still holds them,
	// then it comes back as Jurassic Park.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	gap2B02Settle(t, g, func(*game.PendingChoice) (game.TargetRef, bool) { return game.TargetRef{}, false })
	if !opp.Graveyard.Contains(rock) {
		t.Error("chapter III did not destroy the Wall")
	}
	if effectstest.BattlefieldCardFor(g, saga) != nil {
		t.Fatal("the Saga object is still on the battlefield")
	}
	park := effectstest.FindBattlefieldByName(g, "Jurassic Park")
	if park == uuid.Nil {
		t.Fatal("no Jurassic Park")
	}
	effectstest.AssertFaceInvariant(t, g)

	// Jurassic Park: {G} for each Dinosaur, and escape for Dinosaur cards.
	if err := g.ActivateManaAbility(me.ID, park, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Jurassic Park: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool %v, want one {G} for the one Dinosaur", me.ManaPool)
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	dead, other := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Dinosaur", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var fodder []uuid.UUID
	for range 3 {
		id := uuid.New()
		fodder = append(fodder, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	}
	effectstest.FloatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder, Strict: true}); err == nil {
		t.Fatal("a non-Dinosaur card escaped")
	}
	if err := g.CastSpell(me.ID, dead, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: fodder, Strict: true}); err != nil {
		t.Fatalf("escape the Dinosaur card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the escaped Dinosaur did not resolve")
	}
}

func TestGap2B02WelcomeToWallEndsWhenTheSagaLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	rock := effectstest.DftABPush(g, opp.ID, "Opp Rock", "Artifact", "", 0, 0)
	creature := effectstest.DftABPush(g, opp.ID, "Opp Golem", "Artifact Creature — Golem", "", 3, 3)
	saga := effectstest.ImportAndCast(t, g, gap2B02WelcomeRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FirstPendingTarget(g); c != nil && effectstest.HasID(c.PickTargetCards, creature) {
		t.Error("an artifact creature was offered as a noncreature artifact")
	}
	gap2B02FirstOppPick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardFor(g, rock); c == nil || !c.IsCreature() {
		t.Fatal("the rock did not become a creature")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, saga); err != nil {
		t.Fatalf("destroy the Saga: %v", err)
	}
	g.RunStateChecksForTest()
	if c := effectstest.BattlefieldCardFor(g, rock); c == nil || c.IsCreature() || c.HasSubtype("Wall") {
		t.Error("the rock is still a Wall creature after the Saga left")
	}
}
