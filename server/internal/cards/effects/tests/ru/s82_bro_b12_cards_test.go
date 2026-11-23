package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	broB12ArbalestOracle    = "06bea0b2-1d0f-4d64-a13e-723d08042e83"
	broB12SpiderOracle      = "78d710ea-fbf5-4855-b874-b53478bf4a87"
	broB12WurmOracle        = "5b5b9390-bb3a-4f71-809a-cfe2a6566b59"
	broB12NetOracle         = "3f285781-ee03-4a58-a85e-a33f5efffc9d"
	broB12RefineryOracle    = "a23c4378-9242-46a0-9f62-ca91f0baa428"
	broB12HunterOracle      = "08dcda8e-704d-444e-a7ca-34586a7f42a7"
	broB12DevastationOracle = "03c1e376-1410-4321-90fe-a6997e8db85a"
	broB12VisionsOracle     = "634b46a6-ec8b-41f0-8eb6-1e9144035847"
	broB12BearerOracle      = "de6fd93e-16d8-4855-a655-1cc0755c5d15"
	broB12UrzaOracle        = "8ad7fdd2-bff5-4434-a3f9-1f5ce0ed573b"
	broB12MishraOracle      = "4eba0e29-985b-4763-9a70-0946ba044b43"
)

func TestBroB12CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB12ArbalestOracle: "Arbalest Engineers", broB12SpiderOracle: "Thran Spider",
		broB12WurmOracle: "Sarinth Greatwurm", broB12NetOracle: "Static Net",
		broB12RefineryOracle: "Slagstone Refinery", broB12HunterOracle: "Wreck Hunter",
		broB12DevastationOracle: "Terisiare's Devastation", broB12VisionsOracle: "Visions of Phyrexia",
		broB12BearerOracle: "Battery Bearer", broB12UrzaOracle: "Urza, Powerstone Prodigy",
		broB12MishraOracle: "Mishra, Excavation Prodigy",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// broB12Powerstones counts the Powerstones a player controls and how
// many of them are tapped.
func broB12Powerstones(g *game.Game, who uuid.UUID) (total, tapped int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == who && c.Name == "Powerstone" && c.IsToken() {
			total++
			if c.Tapped {
				tapped++
			}
		}
	}
	return total, tapped
}

func broB12WantStones(t *testing.T, g *game.Game, who uuid.UUID, want int) {
	t.Helper()
	total, tapped := broB12Powerstones(g, who)
	if total != want || tapped != want {
		t.Errorf("Powerstones: %d (%d tapped), want %d, all tapped", total, tapped, want)
	}
}

// broB12Settle resolves the stack, taking the engine's order for any
// trigger-order prompt on the way.
func broB12Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.TriggerOrderPrompt(g) == nil {
			return
		}
	}
}

// broB12Discard discards the given cards from a hand together, as one
// effect (a discard of your own choosing, so the cards are a batch).
func broB12Discard(t *testing.T, g *game.Game, who uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DiscardCardsKeyedForEffect(who, ids, uuid.Nil, game.KeyedTail{}); err != nil {
			t.Errorf("discard: %v", err)
		}
	})
}

// --- Static Net ----------------------------------------------------------

func TestBroB12StaticNetExilesUntilItLeavesAndMakesAPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Land — Forest")
	life := me.Life
	net := effectstest.CastCatalogSpell(t, g, "Static Net", "Enchantment", broB12NetOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, land) {
		t.Fatal("a land is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	broB12Settle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear should be exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	broB12WantStones(t, g, me.ID, 1)

	// It comes back when the Net leaves (CR 610.3).
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, net); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	broB12Settle(t, g)
	g.RunStateChecksForTest()
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" && c.Controller == opp.ID {
			back = true
		}
	}
	if !back {
		t.Error("the Bear should return when Static Net leaves")
	}
}

// --- Sarinth Greatwurm ---------------------------------------------------

func TestBroB12SarinthGreatwurmMakesAPowerstoneForEveryLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sarinth Greatwurm", "Creature — Wurm", broB12WurmOracle, false)
	// My land, then an opponent's.
	effectstest.B12PlayFromHand(t, g, "My Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	broB12WantStones(t, g, me.ID, 1)
	theirs := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, theirs); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	broB12WantStones(t, g, me.ID, 2)
	if total, _ := broB12Powerstones(g, opp.ID); total != 0 {
		t.Errorf("the opponent got %d Powerstones, want 0", total)
	}
}

// --- Thran Spider --------------------------------------------------------

func TestBroB12ThranSpiderGivesYouAndTargetOpponentATappedPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := effectstest.CastAndResolveCreature(t, g, "Thran Spider", "Artifact Creature — Spider", broB12SpiderOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, id := range p.PickTargetPlayers {
		if id == me.ID {
			t.Fatal("you are not a legal target")
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
	broB12WantStones(t, g, opp.ID, 1)
	if total, _ := broB12Powerstones(g, g.Seats[2].ID); total != 0 {
		t.Errorf("an untargeted opponent got %d Powerstones", total)
	}
	effectstest.AssertKeywords(t, g, spider, "reach")
}

func TestBroB12ThranSpiderDigsForAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.PushCatalogPermanent(g, me.ID, "Thran Spider", "Artifact Creature — Spider", broB12SpiderOracle, false)
	me.Library.Cards = nil
	deep := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: deep, Name: "Deep Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	var rock uuid.UUID
	for _, typeLine := range []string{"Creature — Bear", "Artifact", "Land", "Sorcery"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		if typeLine == "Artifact" {
			rock = id
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, spider, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d cards offered, want the one artifact among the top four", len(c.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact is not in hand")
	}
	if n := len(me.Library.Cards); n != 4 || me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the untouched fifth card should be on top over the three bottomed ones: %d cards", n)
	}
}

func TestBroB12ThranSpiderRefusesAnActivationWithoutSevenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.PushCatalogPermanent(g, me.ID, "Thran Spider", "Artifact Creature — Spider", broB12SpiderOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, spider, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("six mana pays for a {7} ability")
	}
}

// --- Arbalest Engineers --------------------------------------------------

func broB12CastArbalest(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.CastAndResolveCreature(t, g, "Arbalest Engineers", "Creature — Human Artificer", broB12ArbalestOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return id
}

func TestBroB12ArbalestEngineersDamageMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	broB12CastArbalest(t, g, 0)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	broB12Settle(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, opp.Life)
	}
}

func TestBroB12ArbalestEngineersCounterMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	broB12CastArbalest(t, g, 1)
	effectstest.PickCard(t, g, me.ID, bear)
	broB12Settle(t, g)
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
	effectstest.AssertKeywords(t, g, bear, "trample", "haste")
	// Until end of turn only.
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	c, _ := g.LookupCardForEffect(bear)
	if effectstest.HasKeywordForTest(c.Effective().Abilities, "trample") {
		t.Error("the keywords should wear off at end of turn")
	}
}

func TestBroB12ArbalestEngineersPowerstoneMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	broB12CastArbalest(t, g, 2)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

// --- Terisiare's Devastation ---------------------------------------------

func broB12CastDevastation(t *testing.T, g *game.Game, x int) {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Terisiare's Devastation", TypeLine: "Sorcery", ManaCost: "{X}{2}{B}{B}",
		OracleID: broB12DevastationOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	for i := 0; i < x; i++ {
		effectstest.B06AddMana(me, "C")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	broB12Settle(t, g)
	g.RunStateChecksForTest()
}

func TestBroB12TerisiaresDevastationLosesLifeMakesStonesAndShrinksByArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	small := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 5, 5)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	broB12CastDevastation(t, g, 2)
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
	broB12WantStones(t, g, me.ID, 2)
	// Three artifacts (the rock and two Powerstones): -3/-3.
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 should die at -3/-3")
	}
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{mine, 2}, {big, 3}} {
		card, ok := g.LookupCardForEffect(c.id)
		if !ok || !g.Battlefield.Contains(c.id) {
			t.Fatalf("%s should survive", card.Name)
		}
		if got := card.CurrentPower(); got != c.want {
			t.Errorf("%s has power %d, want %d (the caster's creatures shrink too)", card.Name, got, c.want)
		}
	}
}

func TestBroB12TerisiaresDevastationWithXZeroOnlyShrinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	broB12CastDevastation(t, g, 0)
	if me.Life != life {
		t.Errorf("life changed to %d", me.Life)
	}
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Errorf("%d Powerstones with X = 0", total)
	}
	if card, _ := g.LookupCardForEffect(bear); card.CurrentPower() != 1 {
		t.Errorf("the Bear has power %d, want 1 (one artifact)", card.CurrentPower())
	}
}

func TestBroB12TerisiaresDevastationNeedsItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Terisiare's Devastation", TypeLine: "Sorcery", ManaCost: "{X}{2}{B}{B}",
		OracleID: broB12DevastationOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true}); err == nil {
		t.Error("X = 2 is cast with only four mana")
	}
}

// --- Wreck Hunter --------------------------------------------------------

func TestBroB12WreckHunterCountsNonlandCardsPutThereFromTheBattlefieldThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toGraveyard := func(id uuid.UUID, owner uuid.UUID) {
		t.Helper()
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	// Counted: a creature and an artifact of theirs from the battlefield.
	toGraveyard(effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2), opp.ID)
	toGraveyard(effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact"), opp.ID)
	// Not counted: a land, a token, a card from hand, one of mine.
	toGraveyard(effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest"), opp.ID)
	toGraveyard(effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	}), opp.ID)
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Discarded", TypeLine: "Creature — Elf", Owner: opp.ID, Controller: opp.ID})
	toGraveyard(effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2), me.ID)

	effectstest.CastAndResolveCreature(t, g, "Wreck Hunter", "Creature — Human Artificer", broB12HunterOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 2)
	if total, _ := broB12Powerstones(g, opp.ID); total != 0 {
		t.Errorf("the target player got %d Powerstones", total)
	}
}

func TestBroB12WreckHunterCanTargetYouAndIgnoresLastTurnsCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// A card that has been in the graveyard since an earlier turn.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	fresh := effectstest.B12Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, fresh); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.CastAndResolveCreature(t, g, "Wreck Hunter", "Creature — Human Artificer", broB12HunterOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, me.ID)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

func TestBroB12WreckHunterWithNothingCountedMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastAndResolveCreature(t, g, "Wreck Hunter", "Creature — Human Artificer", broB12HunterOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	broB12Settle(t, g)
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Errorf("%d Powerstones for an empty count", total)
	}
}

// --- Slagstone Refinery --------------------------------------------------

func TestBroB12SlagstoneRefineryTriggersOnYourNontokenArtifactsGoingToGraveyardOrExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Slagstone Refinery", "Artifact", broB12RefineryOracle, false)
	move := func(id uuid.UUID, to game.ZoneRef) {
		t.Helper()
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, to, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
		broB12Settle(t, g)
	}
	grave := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}
	exile := game.ZoneRef{Kind: game.ZoneExile}

	// Not counted: a token, an opponent's artifact, a nonartifact, a bounce.
	move(effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure", Owner: me.ID, Controller: me.ID,
	}), grave)
	move(effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact"), game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID})
	move(effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2), grave)
	move(effectstest.B12Permanent(g, me.ID, "My Bounced Rock", "Artifact"), game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID})
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Fatalf("%d Powerstones before any counted artifact left", total)
	}
	// Counted: an artifact to the graveyard, an artifact creature to exile.
	move(effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact"), grave)
	broB12WantStones(t, g, me.ID, 1)
	move(effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2), exile)
	broB12WantStones(t, g, me.ID, 2)
}

func TestBroB12SlagstoneRefineryTriggersOnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	refinery := effectstest.PushCatalogPermanent(g, me.ID, "Slagstone Refinery", "Artifact", broB12RefineryOracle, false)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, refinery); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

func TestBroB12SlagstoneRefineryMakesOnePerArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Slagstone Refinery", "Artifact", broB12RefineryOracle, false)
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	for _, id := range []uuid.UUID{a, b} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 2)
}

// --- Visions of Phyrexia -------------------------------------------------

func broB12VisionsUpkeep(t *testing.T, g *game.Game, top game.Card) {
	t.Helper()
	me := g.Seats[0]
	// Seat 3's turn is under way, so seat 0's own turn 1 is behind us.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 3)
	effectstest.PushCatalogPermanent(g, me.ID, "Visions of Phyrexia", "Enchantment", broB12VisionsOracle, false)
	top.InstanceID, top.Owner, top.Controller = uuid.New(), me.ID, me.ID
	me.Library.PushTop(top)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	// The draw step has not happened: the exiled card is the one we put on top.
	broB12Settle(t, g)
	if !g.Exile.Contains(top.InstanceID) {
		t.Fatalf("the top card was not exiled: library top is %v", me.Library.Cards[len(me.Library.Cards)-1].Name)
	}
}

func TestBroB12VisionsOfPhyrexiaMakesAPowerstoneIfYouPlayNothingFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	broB12VisionsUpkeep(t, g, game.Card{Name: "Spare Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.AdvanceToEndStep(t, g)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

func TestBroB12VisionsOfPhyrexiaMakesNothingIfYouCastFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{Name: "Spare Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
	broB12VisionsUpkeep(t, g, card)
	var id uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Name == "Spare Bear" {
			id = c.InstanceID
		}
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	broB12Settle(t, g)
	effectstest.AdvanceToEndStep(t, g)
	broB12Settle(t, g)
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Errorf("%d Powerstones after casting from exile, want 0", total)
	}
}

func TestBroB12VisionsOfPhyrexiaMakesNothingIfYouPlayALandFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	broB12VisionsUpkeep(t, g, game.Card{Name: "Spare Forest", TypeLine: "Basic Land — Forest"})
	var id uuid.UUID
	for _, c := range g.Exile.Cards {
		if c.Name == "Spare Forest" {
			id = c.InstanceID
		}
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	broB12Settle(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the land was not played")
	}
	effectstest.AdvanceToEndStep(t, g)
	broB12Settle(t, g)
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Errorf("%d Powerstones after playing a land from exile, want 0", total)
	}
}

func TestBroB12VisionsOfPhyrexiaCastFromHandStillMakesAPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	broB12VisionsUpkeep(t, g, game.Card{Name: "Spare Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastAndResolveCreature(t, g, "Hand Bear", "Creature — Bear", "")
	effectstest.AdvanceToEndStep(t, g)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

// --- Battery Bearer ------------------------------------------------------

func TestBroB12BatteryBearerGivesCreaturesRestrictedColorlessMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Battery Bearer", "Creature — Human Artificer", broB12BearerOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, bear, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("a creature you control taps for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 1 {
		t.Errorf("pool holds %d {C}, want 1", got)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("an opponent's creature got the ability")
	}
}

func TestBroB12BatteryBearerManaCantCastANonartifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Battery Bearer", "Creature — Human Artificer", broB12BearerOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, bear, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast("Cheap Bear", "Creature — Bear"); err == nil {
		t.Error("the restricted mana paid for a nonartifact spell")
	}
	if err := cast("Cheap Rock", "Artifact"); err != nil {
		t.Errorf("the restricted mana should pay for an artifact spell: %v", err)
	}
}

func TestBroB12BatteryBearerDrawsOnABigArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Battery Bearer", "Creature — Human Artificer", broB12BearerOracle, false)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []struct {
		name, typeLine, cost string
		draws                bool
	}{
		{"Six Rock", "Artifact", "{6}", true},
		{"Five Rock", "Artifact", "{5}", false},
		{"Six Beast", "Creature — Beast", "{6}", false},
		{"Seven Golem", "Artifact Creature — Golem", "{7}", true},
	} {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, Owner: me.ID, Controller: me.ID})
		effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C")
		before := me.Hand.Size()
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", c.name, err)
		}
		broB12Settle(t, g)
		me.ManaPool.EmptyPool()
		want := before - 1
		if c.draws {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("%s: hand %d -> %d, want %d", c.name, before, got, want)
		}
	}
}

// --- Urza, Powerstone Prodigy and Mishra, Excavation Prodigy -------------

func TestBroB12UrzaLootsAndMakesAPowerstoneWhenYouDiscardAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	urza := effectstest.PushCatalogPermanent(g, me.ID, "Urza, Powerstone Prodigy", "Legendary Creature — Human Artificer", broB12UrzaOracle, false)
	rock := uuid.New()
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Spare Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, urza, 0, game.ActivateAbilityParams{})
	broB12Settle(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, rock)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
	effectstest.AssertKeywords(t, g, urza, "vigilance")
}

func TestBroB12UrzaIgnoresANonartifactDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	urza := effectstest.PushCatalogPermanent(g, me.ID, "Urza, Powerstone Prodigy", "Legendary Creature — Human Artificer", broB12UrzaOracle, false)
	bear := uuid.New()
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Spare Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, urza, 0, game.ActivateAbilityParams{})
	broB12Settle(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, bear)
	broB12Settle(t, g)
	if total, _ := broB12Powerstones(g, me.ID); total != 0 {
		t.Errorf("%d Powerstones for a nonartifact discard", total)
	}
}

func TestBroB12UrzaTriggersOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Urza, Powerstone Prodigy", "Legendary Creature — Human Artificer", broB12UrzaOracle, false)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
		broB12Discard(t, g, me.ID, id)
		broB12Settle(t, g)
	}
	broB12WantStones(t, g, me.ID, 1)
}

func TestBroB12UrzaMakesOneForTwoArtifactsDiscardedTogether(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Urza, Powerstone Prodigy", "Legendary Creature — Human Artificer", broB12UrzaOracle, false)
	effectstest.AdvanceToMain(t, g)
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	for _, id := range ids {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	}
	broB12Discard(t, g, me.ID, ids...)
	broB12Settle(t, g)
	broB12WantStones(t, g, me.ID, 1)
}

func TestBroB12MishraDiscardsToDrawAndAddsRedRedForAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mishra := effectstest.PushCatalogPermanent(g, me.ID, "Mishra, Excavation Prodigy", "Legendary Creature — Human Artificer", broB12MishraOracle, false)
	rock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Spare Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, mishra, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{rock}})
	broB12Settle(t, g)
	if me.Hand.Size() != hand { // one discarded, one drawn
		t.Errorf("hand %d -> %d, want unchanged", hand, me.Hand.Size())
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("pool holds %d {R}, want 2", got)
	}
	effectstest.AssertKeywords(t, g, mishra, "haste")
}

func TestBroB12MishraAddsNothingForANonartifactDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mishra := effectstest.PushCatalogPermanent(g, me.ID, "Mishra, Excavation Prodigy", "Legendary Creature — Human Artificer", broB12MishraOracle, false)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Spare Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, mishra, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{bear}})
	broB12Settle(t, g)
	if got := me.ManaPool.CountColor("R"); got != 0 {
		t.Errorf("pool holds %d {R}, want 0", got)
	}
}
