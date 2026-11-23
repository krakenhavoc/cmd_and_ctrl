package testsg

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a7_cards_test.go — The Hobbit / The Hobbit Commander, half A, seventh
// batch (#4454): modal spells, landfall and attack triggers, and a few
// evasive bodies. Each test casts or plays the real catalog entry and
// asserts what the card's text says.

const (
	hobA7ReverentHowlOracle      = "86214b4c-a3ad-4372-b5fe-8f742a7a3d4c"
	hobA7DuskwatchHunterOracle   = "6283e5ca-9924-478b-89a5-36924accb9aa"
	hobA7OpportunistOracle       = "61773534-8da8-48fa-b51a-73c0a9a751c4"
	hobA7WarglingOracle          = "ce107656-f222-4d9a-9ba2-b33b4d47bdb5"
	hobA7StalwartOracle          = "bd4ab393-9538-4230-b2a1-bc77099481c9"
	hobA7RaftSteererOracle       = "d0f62148-1015-49b8-b61f-a4332325a6b0"
	hobA7MeditatorOracle         = "399cd159-7376-421c-b001-c26f1cbadced"
	hobA7ConfusticateOracle      = "7a7d9887-3cc4-4fb9-9109-1315903ace45"
	hobA7NimrodelWatcherOracle   = "1b8de1c5-1e23-4c84-bc5a-a14853ad8c7a"
	hobA7NoriOracle              = "03253076-9add-4efb-9227-b0a062c86534"
	hobA7CantankerousKeepersOrcl = "77c97dad-17cc-4a28-a48b-907c044d9034"
	hobA7BejeweledWargOracle     = "051ff7e0-dd00-4467-8796-a5d1c21934ed"
	hobA7ShortswordOracle        = "9cc52afb-9788-46f6-9f9b-347513f8a64f"
	hobA7LonelyMountainOracle    = "3678c06f-8a33-4a6d-bf20-5b92d5c05a95"
)

func TestHobA7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA7ReverentHowlOracle:      "Reverent Howl",
		hobA7DuskwatchHunterOracle:   "Duskwatch Hunter",
		hobA7OpportunistOracle:       "Gundabad Opportunist",
		hobA7WarglingOracle:          "Wargling",
		hobA7StalwartOracle:          "Iron Hills Stalwart",
		hobA7RaftSteererOracle:       "Elven Raft-Steerer",
		hobA7MeditatorOracle:         "Mirkwood Meditator",
		hobA7ConfusticateOracle:      "Confusticate and Bebother",
		hobA7NimrodelWatcherOracle:   "Nimrodel Watcher",
		hobA7NoriOracle:              "Nori, Teller of Tales",
		hobA7CantankerousKeepersOrcl: "Cantankerous Keepers",
		hobA7BejeweledWargOracle:     "Bejeweled Warg",
		hobA7ShortswordOracle:        "Dwarven Shortsword",
		hobA7LonelyMountainOracle:    "The Lonely Mountain",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// hobA7PlayerRef is a player target announced for mode occurrence 0.
func hobA7PlayerRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// hobA7CardRef is a card target announced for mode occurrence 0.
func hobA7CardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// hobA7Counters reads a battlefield card's +1/+1 counters after the layers
// settle.
func hobA7Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.S58p6View(t, g, id).Counters[game.CounterPlusOne]
}

// --- Reverent Howl ---------------------------------------------------------

// The first bullet: the targeted player draws two cards and loses 2 life.
func TestHobA7ReverentHowlMakesAnOpponentDrawTwoAndLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	hand, life := opp.Hand.Size(), opp.Life

	effectstest.B02cCastModalSpell(t, g, "Reverent Howl", "Instant", hobA7ReverentHowlOracle, []int{0},
		[]game.TargetRef{hobA7PlayerRef(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	if got := opp.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d", got, hand+2)
	}
	if opp.Life != life-2 {
		t.Errorf("life = %d, want %d", opp.Life, life-2)
	}
}

// You may name yourself: the card is a draw spell as often as a punisher.
func TestHobA7ReverentHowlCanTargetItsCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	hand, life := me.Hand.Size(), me.Life

	effectstest.B02cCastModalSpell(t, g, "Reverent Howl", "Instant", hobA7ReverentHowlOracle, []int{0},
		[]game.TargetRef{hobA7PlayerRef(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	// The spell left the hand when it was cast.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d", got, hand+2)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

// The second bullet: +2/+2 and lifelink on the target creature.
func TestHobA7ReverentHowlPumpsAndGivesLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)

	effectstest.B02cCastModalSpell(t, g, "Reverent Howl", "Instant", hobA7ReverentHowlOracle, []int{1},
		[]game.TargetRef{hobA7CardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.S58p6View(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("target is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "lifelink") {
		t.Error("the target should have lifelink")
	}
	o := effectstest.S58p6View(t, g, other)
	if o.CurrentPower() != 2 || game.HasKeyword(&o, "lifelink") {
		t.Error("an untargeted creature changed")
	}
}

// Each bullet takes only its own kind of target.
func TestHobA7ReverentHowlBulletsRefuseTheOtherKindOfTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	for _, tc := range []struct {
		mode int
		ref  game.TargetRef
	}{
		{0, hobA7CardRef(bear)},
		{1, hobA7PlayerRef(opp.ID)},
	} {
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Reverent Howl", TypeLine: "Instant",
			OracleID: hobA7ReverentHowlOracle, Owner: active.ID, Controller: active.ID})
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: []int{tc.mode}, Targets: []game.TargetRef{tc.ref}}); err == nil {
			t.Errorf("mode %d accepted a %s target", tc.mode, tc.ref.Kind)
		}
	}
}

// --- Duskwatch Hunter ------------------------------------------------------

func TestHobA7DuskwatchHunterPutsACounterOnTargetCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Duskwatch Hunter", "Creature — Wolf", hobA7DuskwatchHunterOracle, 3, 1)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)

	if got := hobA7Counters(t, g, bear); got != 1 {
		t.Errorf("their Bear has %d +1/+1 counters, want 1: the target is any creature", got)
	}
}

func TestHobA7DuskwatchHunterMayTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hunter := effectstest.CastHoldCreature(t, g, "Duskwatch Hunter", "Creature — Wolf", hobA7DuskwatchHunterOracle, 3, 1)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: hunter}, true)

	c := effectstest.S58p6View(t, g, hunter)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("Hunter is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobA7DuskwatchHunterCannotBeBlockedByTokensButByOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.B12Push(g, me.ID, "Duskwatch Hunter", "Creature — Wolf", hobA7DuskwatchHunterOracle, 3, 1)
	token := effectstest.B12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, hunter)

	offered := effectstest.BrOffered(t, g, opp.ID)
	if effectstest.BrOffers(offered, token, hunter) || !effectstest.BrOffers(offered, bear, hunter) {
		t.Errorf("the enumerator disagrees with the rule: %v", offered)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(token, hunter), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(bear, hunter); err != nil {
		t.Fatalf("a nontoken creature blocks the Hunter: %v", err)
	}
}

// --- Gundabad Opportunist --------------------------------------------------

func TestHobA7GundabadOpportunistExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Impulse Land", TypeLine: "Basic Land — Forest"})
	under := uuid.New()
	me.Library.Cards = append([]game.Card{{InstanceID: under, Name: "Underneath", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}}, me.Library.Cards...)

	effectstest.CastHoldCreature(t, g, "Gundabad Opportunist", "Creature — Goblin Rogue", hobA7OpportunistOracle, 4, 2)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)

	if me.Library.Contains(top) || !g.Exile.Contains(top) {
		t.Fatal("the top card should be exiled")
	}
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card carries no permission to play it")
	}
	if !me.Library.Contains(under) || g.Exile.Contains(under) {
		t.Error("only the top card is exiled")
	}
}

// --- Wargling --------------------------------------------------------------

func hobA7WarglingTable(t *testing.T, friendPower int) (g *game.Game, wargling, friend, theirs uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wargling = effectstest.B12Push(g, me.ID, "Wargling", "Creature — Wolf", hobA7WarglingOracle, 2, 2)
	friend = effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", friendPower, 4)
	theirs = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, wargling)
	effectstest.PassPriorityAroundTable(t, g)
	return g, wargling, friend, theirs
}

func TestHobA7WarglingFerociousPumpsItselfAndGivesTheTeamTrample(t *testing.T) {
	g, wargling, friend, theirs := hobA7WarglingTable(t, 4)

	if c := effectstest.S58p6View(t, g, wargling); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("Wargling is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, wargling, "trample") || !effectstest.HasEffectiveKeyword(t, g, friend, "trample") {
		t.Error("creatures I control should have trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

func TestHobA7WarglingDoesNothingWithoutAPowerFourCreature(t *testing.T) {
	g, wargling, friend, _ := hobA7WarglingTable(t, 3)

	if c := effectstest.S58p6View(t, g, wargling); c.CurrentPower() != 2 {
		t.Errorf("Wargling's power is %d, want 2: ferocious needs a power 4 creature", c.CurrentPower())
	}
	if effectstest.HasEffectiveKeyword(t, g, friend, "trample") {
		t.Error("trample was granted without ferocious")
	}
}

// --- Iron Hills Stalwart ---------------------------------------------------

// hobA7CastStalwart casts the Stalwart and answers its two target prompts: the
// Equipment first, then the creature, or none when `creature` is uuid.Nil.
func hobA7CastStalwart(t *testing.T, g *game.Game, me uuid.UUID, equipment, creature uuid.UUID) {
	t.Helper()
	effectstest.CastHoldCreature(t, g, "Iron Hills Stalwart", "Creature — Dwarf Warrior", hobA7StalwartOracle, 4, 5)
	first := effectstest.LatestPickTarget(g, me)
	if first == nil {
		t.Fatal("no prompt for the Equipment")
	}
	if err := g.ResolvePickTarget(first.ID, me, game.TargetRef{Kind: game.TargetCard, ID: equipment}); err != nil {
		t.Fatalf("pick the Equipment: %v", err)
	}
	second := effectstest.LatestPickTarget(g, me)
	if second == nil {
		t.Fatal("no prompt for the creature")
	}
	var picks []game.TargetRef
	if creature != uuid.Nil {
		picks = []game.TargetRef{{Kind: game.TargetCard, ID: creature, Slot: 1}}
	}
	if err := g.ResolvePickTargets(second.ID, me, picks); err != nil {
		t.Fatalf("pick the creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestHobA7IronHillsStalwartAttachesAnEquipmentToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hobA7CastStalwart(t, g, me.ID, sword, bear)

	s := effectstest.FindBattlefieldCardByID(g, sword)
	if s == nil || !s.IsAttachedTo(bear) {
		t.Fatalf("the Sword should be attached to the Bear: %+v", s)
	}
}

// The Equipment moves off the creature it was on, and the new host may be
// the Stalwart itself.
func TestHobA7IronHillsStalwartMovesAnEquipmentOntoItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AttachForEffect(sword, hobA7CardRef(bear)); err != nil {
		t.Fatalf("attach: %v", err)
	}
	effectstest.CastHoldCreature(t, g, "Iron Hills Stalwart", "Creature — Dwarf Warrior", hobA7StalwartOracle, 4, 5)
	first := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(first.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: sword}); err != nil {
		t.Fatalf("pick the Equipment: %v", err)
	}
	second := effectstest.LatestPickTarget(g, me.ID)
	var stalwart uuid.UUID
	for _, id := range second.PickTargetCards {
		if id != bear {
			stalwart = id
		}
	}
	if err := g.ResolvePickTargets(second.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: stalwart, Slot: 1}}); err != nil {
		t.Fatalf("pick the Stalwart: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	s := effectstest.FindBattlefieldCardByID(g, sword)
	if s == nil || !s.IsAttachedTo(stalwart) || s.IsAttachedTo(bear) {
		t.Fatalf("the Sword should have moved from the Bear to the Stalwart: %+v", s)
	}
}

// "Up to one": with no creature chosen, nothing is attached.
func TestHobA7IronHillsStalwartMayChooseNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hobA7CastStalwart(t, g, me.ID, sword, uuid.Nil)

	if s := effectstest.FindBattlefieldCardByID(g, sword); s == nil || s.AttachedTo.ID != uuid.Nil {
		t.Errorf("the Sword should stay unattached: %+v", s)
	}
}

// The first target is an Equipment you control, so with none the trigger has
// nothing to do and never asks.
func TestHobA7IronHillsStalwartNeedsAnEquipmentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushPermanent(g, g.Seats[1].ID, game.Card{Name: "Their Sword", TypeLine: "Artifact — Equipment"})
	effectstest.CastHoldCreature(t, g, "Iron Hills Stalwart", "Creature — Dwarf Warrior", hobA7StalwartOracle, 4, 5)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("asked for a target with no Equipment of mine on the battlefield")
	}
}

// --- Landfall --------------------------------------------------------------

// hobA7Landfall plays a Forest for the active seat and answers the landfall
// trigger's "choose one" with `mode`, leaving any target prompt of the chosen
// bullet open.
func hobA7Landfall(t *testing.T, g *game.Game, me uuid.UUID, mode int) {
	t.Helper()
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatal("landfall did not ask which mode")
	}
	if err := g.ResolveModePick(c.ID, me, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

// --- Elven Raft-Steerer ----------------------------------------------------

func TestHobA7ElvenRaftSteererTapsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Elven Raft-Steerer", "Creature — Elf Pilot", hobA7RaftSteererOracle, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hobA7Landfall(t, g, me.ID, 0)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the tap bullet")
	}
	if effectstest.HasID(p.PickTargetCards, mine) || !effectstest.HasID(p.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: only a creature an opponent controls", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || !c.Tapped {
		t.Error("their Bear should be tapped")
	}
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || c.Tapped {
		t.Error("my Bear must stay untapped")
	}
}

func TestHobA7ElvenRaftSteererUntapsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Elven Raft-Steerer", "Creature — Elf Pilot", hobA7RaftSteererOracle, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssistCardTapped(t, g, mine)
	effectstest.AssistCardTapped(t, g, theirs)
	hobA7Landfall(t, g, me.ID, 1)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the untap bullet")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("legal targets %v: only a creature I control", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || c.Tapped {
		t.Error("my Bear should be untapped")
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || !c.Tapped {
		t.Error("their Bear must stay tapped")
	}
}

func TestHobA7ConfusticateCountersASpellUnlessItsControllerPaysFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.HobA7BearSpell(t, g, me)
	if err := effectstest.RfCastModal(t, g, "Confusticate and Bebother", hobA7ConfusticateOracle, []int{0}, effectstest.RfCardTarget(bear)); err != nil {
		t.Fatalf("cast Confusticate and Bebother: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("declining the {4} should counter the spell")
	}
}

func TestHobA7ConfusticateLetsTheSpellResolveWhenFourIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.HobA7BearSpell(t, g, me)
	if err := effectstest.RfCastModal(t, g, "Confusticate and Bebother", hobA7ConfusticateOracle, []int{0}, effectstest.RfCardTarget(bear)); err != nil {
		t.Fatalf("cast Confusticate and Bebother: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("paying the {4} should let the spell resolve")
	}
}

// Three mana is not enough: the tax is {4}, not the {3} of its cousins.
func TestHobA7ConfusticateAsksForFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.HobA7BearSpell(t, g, me)
	if err := effectstest.RfCastModal(t, g, "Confusticate and Bebother", hobA7ConfusticateOracle, []int{0}, effectstest.RfCardTarget(bear)); err != nil {
		t.Fatalf("cast Confusticate and Bebother: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, me.ID)
	var cost string
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			cost = c.PayCost
		}
	}
	if cost != "{4}" {
		t.Errorf("the tax is %q, want {4}", cost)
	}
}

func TestHobA7ConfusticateRefusesAnythingButASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.RfCastModal(t, g, "Confusticate and Bebother", hobA7ConfusticateOracle, []int{0}, effectstest.RfCardTarget(bear)); err == nil {
		t.Error("a creature on the battlefield is not a legal target for the counter bullet")
	}
}

func TestHobA7ConfusticateLootsForTwoAndDiscardsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()
	if err := effectstest.RfCastModal(t, g, "Confusticate and Bebother", hobA7ConfusticateOracle, []int{1}, nil); err != nil {
		t.Fatalf("cast Confusticate and Bebother: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand = %d, want %d after drawing two", got, hand+2)
	}
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 1 {
		t.Fatalf("owed discards = %d, want 1", owed)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand = %d, want %d after discarding one", got, hand+1)
	}
}

// --- Mirkwood Meditator ----------------------------------------------------

func hobA7Meditator(t *testing.T, yes bool, counters int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := game.Card{
		InstanceID: uuid.New(), Name: "Mirkwood Meditator", TypeLine: "Creature — Elf Druid", OracleID: hobA7MeditatorOracle,
		Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	}
	if counters > 0 {
		card.Counters = map[string]int{game.CounterPlusOne: counters}
	}
	meditator := effectstest.PushBattlefieldCardWithTimestamp(g, card)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
	return g, meditator
}

func TestHobA7MirkwoodMeditatorBecomesFourTwoWhenYouSayYes(t *testing.T) {
	g, id := hobA7Meditator(t, true, 0)
	if c := effectstest.S58p6View(t, g, id); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("Meditator is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestHobA7MirkwoodMeditatorStaysTwoFourWhenYouSayNo(t *testing.T) {
	g, id := hobA7Meditator(t, false, 0)
	if c := effectstest.S58p6View(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 4 {
		t.Errorf("Meditator is %d/%d, want 2/4", c.CurrentPower(), c.CurrentToughness())
	}
}

// Only the BASE power and toughness change, so a counter still adds to it.
func TestHobA7MirkwoodMeditatorKeepsItsCounters(t *testing.T) {
	g, id := hobA7Meditator(t, true, 1)
	if c := effectstest.S58p6View(t, g, id); c.CurrentPower() != 5 || c.CurrentToughness() != 3 {
		t.Errorf("Meditator is %d/%d, want 5/3: a 4/2 base plus a +1/+1 counter", c.CurrentPower(), c.CurrentToughness())
	}
}

// The change ends with the turn.
func TestHobA7MirkwoodMeditatorReturnsToTwoFourNextTurn(t *testing.T) {
	g, id := hobA7Meditator(t, true, 0)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.S58p6View(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 4 {
		t.Errorf("Meditator is %d/%d next turn, want 2/4", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Nimrodel Watcher ------------------------------------------------------

// hobA7Scry has `who` scry 1 and keeps the card on top.
func hobA7Scry(t *testing.T, g *game.Game, who *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { g.ScryForEffect(who.ID, uuid.Nil, 1) })
	scry := effectstest.ScryChoiceFor(g, who.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, who.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func hobA7WatcherTable(t *testing.T) (g *game.Game, watcher uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.SeedLibrary(g.Seats[1], "One", "Two", "Three")
	watcher = effectstest.B12Push(g, me.ID, "Nimrodel Watcher", "Creature — Elf Scout", hobA7NimrodelWatcherOracle, 2, 1)
	return g, watcher
}

func TestHobA7NimrodelWatcherGetsPlusOneAndCannotBeBlockedWhenYouScry(t *testing.T) {
	g, watcher := hobA7WatcherTable(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	hobA7Scry(t, g, g.Seats[0])

	if c := effectstest.S58p6View(t, g, watcher); c.CurrentPower() != 3 || c.CurrentToughness() != 1 {
		t.Errorf("Watcher is %d/%d, want 3/1", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.BrAttack(t, g, watcher)
	if err := g.DeclareBlocker(bear, watcher); err == nil {
		t.Error("a creature blocked the Watcher after its controller scried")
	}
}

// "This ability triggers only once each turn": a second scry adds nothing.
func TestHobA7NimrodelWatcherTriggersOnlyOncePerTurn(t *testing.T) {
	g, watcher := hobA7WatcherTable(t)
	effectstest.AdvanceToMain(t, g)
	hobA7Scry(t, g, g.Seats[0])
	hobA7Scry(t, g, g.Seats[0])

	if c := effectstest.S58p6View(t, g, watcher); c.CurrentPower() != 3 {
		t.Errorf("Watcher's power is %d, want 3: only the first scry of the turn counts", c.CurrentPower())
	}
}

func TestHobA7NimrodelWatcherIgnoresAnOpponentsScry(t *testing.T) {
	g, watcher := hobA7WatcherTable(t)
	effectstest.AdvanceToMain(t, g)
	hobA7Scry(t, g, g.Seats[1])

	if c := effectstest.S58p6View(t, g, watcher); c.CurrentPower() != 2 {
		t.Errorf("Watcher's power is %d, want 2: only your own scry counts", c.CurrentPower())
	}
}

// --- Nori, Teller of Tales -------------------------------------------------

func TestHobA7NoriGivesAnotherAttackerFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nori := effectstest.B12Push(g, me.ID, "Nori, Teller of Tales", "Legendary Creature — Dwarf Bard", hobA7NoriOracle, 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, nori, friend)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Nori's trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, nori) || !effectstest.HasID(p.PickTargetCards, friend) || effectstest.HasID(p.PickTargetCards, idle) {
		t.Errorf("legal targets %v: any attacking creature, and only those", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, friend)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, friend, "first strike") {
		t.Error("the chosen attacker should have first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, nori, "first strike") || effectstest.HasEffectiveKeyword(t, g, idle, "first strike") {
		t.Error("only the chosen attacker gains first strike")
	}
}

func TestHobA7NoriMayGiveFirstStrikeToHerself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nori := effectstest.B12Push(g, me.ID, "Nori, Teller of Tales", "Legendary Creature — Dwarf Bard", hobA7NoriOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, nori)
	effectstest.PickCard(t, g, me.ID, nori)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, nori, "first strike") {
		t.Error("Nori should gain first strike")
	}
}

// --- Cantankerous Keepers --------------------------------------------------

func TestHobA7CantankerousKeepersCostsOneLessForEachElfYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Elf A", TypeLine: "Creature — Elf Druid", Power: 1, Toughness: 1})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Elf B", TypeLine: "Creature — Elf Warrior", Power: 1, Toughness: 1})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Elf", TypeLine: "Creature — Elf Scout", Power: 1, Toughness: 1})

	if got := effectstest.SelfPricedMV(t, g, me, hobA7CantankerousKeepersOrcl, "Creature — Elf Soldier", "{5}{G}"); got != 4 {
		t.Errorf("Keepers with two Elves: mana value %d, want 4", got)
	}
	if got := effectstest.SelfPricedMV(t, g, opp, hobA7CantankerousKeepersOrcl, "Creature — Elf Soldier", "{5}{G}"); got != 5 {
		t.Errorf("Keepers for the opponent with one Elf: mana value %d, want 5", got)
	}
}

func TestHobA7CantankerousKeepersMillsFourAndTakesEveryElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	deep := uuid.New()
	var ids []uuid.UUID
	for _, c := range []struct{ Name, TypeLine string }{
		{"Deep Elf", "Creature — Elf Druid"}, // the fifth card: stays in the library
		{"Elf One", "Creature — Elf Warrior"},
		{"Not An Elf", "Creature — Bear"},
		{"Elf Two", "Creature — Elf Scout"},
		{"A Land", "Basic Land — Forest"},
	} {
		id := uuid.New()
		if c.Name == "Deep Elf" {
			id = deep
		}
		ids = append(ids, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: c.Name, TypeLine: c.TypeLine, Owner: me.ID, Controller: me.ID})
	}
	// Library, top first: A Land, Elf Two, Not An Elf, Elf One | Deep Elf.
	effectstest.CastHoldCreature(t, g, "Cantankerous Keepers", "Creature — Elf Soldier", hobA7CantankerousKeepersOrcl, 4, 3)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) {
		t.Error("both milled Elves should be in hand")
	}
	if !me.Graveyard.Contains(ids[2]) || !me.Graveyard.Contains(ids[4]) {
		t.Error("the milled non-Elves should stay in the graveyard")
	}
	if me.Graveyard.Contains(ids[1]) || me.Graveyard.Contains(ids[3]) {
		t.Error("the Elves must leave the graveyard")
	}
	if !me.Library.Contains(deep) {
		t.Error("the fifth card, an Elf, was not milled and must stay in the library")
	}
}

// --- Bejeweled Warg --------------------------------------------------------

func hobA7WargTable(t *testing.T, mode int) (g *game.Game, me *game.Player, warg, wolf, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	warg = effectstest.B12Push(g, me.ID, "Bejeweled Warg", "Creature — Wolf", hobA7BejeweledWargOracle, 3, 2)
	wolf = effectstest.B12Creature(g, me.ID, "Pack Wolf", "Creature — Wolf", 2, 2)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, warg, g.Seats[1].ID, 3)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("combat damage did not ask which mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	return g, me, warg, wolf, bear
}

func TestHobA7BejeweledWargPutsACounterOnAnotherWolf(t *testing.T) {
	g, me, warg, wolf, bear := hobA7WargTable(t, 0)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the counter bullet")
	}
	if !effectstest.HasID(p.PickTargetCards, warg) || !effectstest.HasID(p.PickTargetCards, wolf) || effectstest.HasID(p.PickTargetCards, bear) {
		t.Errorf("legal targets %v: Wolves I control only", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, wolf)
	effectstest.PassPriorityAroundTable(t, g)

	if got := hobA7Counters(t, g, wolf); got != 1 {
		t.Errorf("the Wolf has %d +1/+1 counters, want 1", got)
	}
	if got := hobA7Counters(t, g, warg); got != 0 {
		t.Errorf("the Warg has %d +1/+1 counters, want 0", got)
	}
	if effectstest.FindBattlefieldByName(g, "Treasure") != uuid.Nil {
		t.Error("the counter bullet also made a Treasure")
	}
}

func TestHobA7BejeweledWargCreatesATreasure(t *testing.T) {
	g, me, warg, wolf, _ := hobA7WargTable(t, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.NamedOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("%d Treasures under my control, want 1", n)
	}
	if hobA7Counters(t, g, warg) != 0 || hobA7Counters(t, g, wolf) != 0 {
		t.Error("the Treasure bullet also put a counter on a Wolf")
	}
}

// Another creature's combat damage and the Warg's own noncombat damage do
// not trigger it.
func TestHobA7BejeweledWargIgnoresOtherCreaturesAndNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warg := effectstest.B12Push(g, me.ID, "Bejeweled Warg", "Creature — Wolf", hobA7BejeweledWargOracle, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Pack Wolf", "Creature — Wolf", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, other, opp.ID, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: warg, Target: opp.ID, Amount: 3})
	})
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("the Warg triggered off damage it did not deal in combat")
	}
}

// --- Dwarven Shortsword ----------------------------------------------------

// The Shortsword makes a 2/2 red Dwarf and attaches itself to it, so the
// Dwarf is 3/4.
func TestHobA7DwarvenShortswordMakesADwarfAndAttachesToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.CastCatalogSpell(t, g, "Dwarven Shortsword", "Artifact — Equipment", hobA7ShortswordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	dwarf := effectstest.FindBattlefieldByName(g, "Dwarf")
	if dwarf == uuid.Nil {
		t.Fatal("no Dwarf token")
	}
	c := effectstest.S58p6View(t, g, dwarf)
	if c.Controller != me.ID || !c.IsToken() || !c.HasSubtype("Dwarf") || !c.HasColorWord("R") {
		t.Errorf("the token is %+v, want a red Dwarf token under my control", c)
	}
	if c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("the Dwarf is %d/%d, want 3/4 (2/2 plus +1/+2)", c.CurrentPower(), c.CurrentToughness())
	}
	if s := effectstest.FindBattlefieldCardByID(g, sword); s == nil || !s.IsAttachedTo(dwarf) {
		t.Error("the Shortsword should be attached to the Dwarf")
	}
}

// Equip moves the Shortsword off the Dwarf and onto another creature.
func TestHobA7DwarvenShortswordEquipMovesItToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sword := effectstest.CastCatalogSpell(t, g, "Dwarven Shortsword", "Artifact — Equipment", hobA7ShortswordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.EquipTo(t, g, me.ID, sword, bear)

	if c := effectstest.S58p6View(t, g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 3/4", c.CurrentPower(), c.CurrentToughness())
	}
	dwarf := effectstest.FindBattlefieldByName(g, "Dwarf")
	if c := effectstest.S58p6View(t, g, dwarf); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the Dwarf is %d/%d after the Sword moved, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- The Lonely Mountain ---------------------------------------------------

func TestHobA7LonelyMountainEntersTappedWithoutAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.PlayLandFromHand(t, g, "The Lonely Mountain", hobA7LonelyMountainOracle)
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("the land should enter tapped with no Equipment")
	}
}

func TestHobA7LonelyMountainEntersUntappedWithAnEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushPermanent(g, g.Seats[g.Turn.ActiveSeat].ID, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	land := effectstest.PlayLandFromHand(t, g, "The Lonely Mountain", hobA7LonelyMountainOracle)
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || c.Tapped {
		t.Error("the land should enter untapped with an Equipment")
	}
}

// Someone else's Equipment does not count.
func TestHobA7LonelyMountainIgnoresAnOpponentsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushPermanent(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, game.Card{Name: "Their Sword", TypeLine: "Artifact — Equipment"})
	land := effectstest.PlayLandFromHand(t, g, "The Lonely Mountain", hobA7LonelyMountainOracle)
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("an opponent's Equipment must not untap the land")
	}
}

func TestHobA7LonelyMountainTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "The Lonely Mountain", TypeLine: "Land — Mountain", OracleID: hobA7LonelyMountainOracle})
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("pool = %d mana, want 1", got)
	}
}

// hobA7MountainDwarf activates the Mountain's token ability with `mana`
// floating and the payment enforced.
func hobA7MountainDwarf(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, mana string) error {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect(%s): %v", mana, err)
	}
	return g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true})
}

func hobA7MountainTable(t *testing.T, equipment int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < equipment; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Sword", TypeLine: "Artifact — Equipment"})
	}
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "The Lonely Mountain", TypeLine: "Land — Mountain", OracleID: hobA7LonelyMountainOracle})
	return g, me, land
}

func TestHobA7LonelyMountainMakesADwarfForFiveMana(t *testing.T) {
	g, me, land := hobA7MountainTable(t, 0)
	if err := hobA7MountainDwarf(t, g, me, land, "{R}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate with {4}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	dwarf := effectstest.FindBattlefieldByName(g, "Dwarf")
	if dwarf == uuid.Nil {
		t.Fatal("no Dwarf token")
	}
	if c := effectstest.S58p6View(t, g, dwarf); c.CurrentPower() != 2 || c.CurrentToughness() != 2 || c.Controller != me.ID {
		t.Errorf("the token is %d/%d under %s, want a 2/2 under me", c.CurrentPower(), c.CurrentToughness(), c.Controller)
	}
	if c := effectstest.FindBattlefieldCardByID(g, land); c == nil || !c.Tapped {
		t.Error("the ability's cost taps the land")
	}
}

func TestHobA7LonelyMountainRefusesFourMana(t *testing.T) {
	g, me, land := hobA7MountainTable(t, 0)
	err := hobA7MountainDwarf(t, g, me, land, "{R}{C}{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("activate with four mana: err = %v, want insufficient mana", err)
	}
}

// {1} less for each Equipment you control, down to {R}.
func TestHobA7LonelyMountainCostsOneLessForEachEquipment(t *testing.T) {
	g, me, land := hobA7MountainTable(t, 2)
	if err := hobA7MountainDwarf(t, g, me, land, "{R}{C}"); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("two Equipment make it {2}{R}; two mana: err = %v, want insufficient mana", err)
	}
	g, me, land = hobA7MountainTable(t, 2)
	if err := hobA7MountainDwarf(t, g, me, land, "{R}{C}{C}"); err != nil {
		t.Fatalf("activate with {2}{R} and two Equipment: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Dwarf") == uuid.Nil {
		t.Error("no Dwarf token")
	}
}

func TestHobA7LonelyMountainCostStopsAtRed(t *testing.T) {
	g, me, land := hobA7MountainTable(t, 6)
	if err := hobA7MountainDwarf(t, g, me, land, "{R}"); err != nil {
		t.Fatalf("six Equipment make it {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Dwarf") == uuid.Nil {
		t.Error("no Dwarf token")
	}
}

// Sorcery speed: not in the middle of combat.
func TestHobA7LonelyMountainIsSorcerySpeed(t *testing.T) {
	g, me, land := hobA7MountainTable(t, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := hobA7MountainDwarf(t, g, me, land, "{R}{C}{C}{C}{C}"); err == nil {
		t.Error("the ability activated outside a main phase")
	}
}
