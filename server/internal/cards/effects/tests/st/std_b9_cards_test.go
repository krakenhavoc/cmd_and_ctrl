package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b9_cards_test.go — Standard batches 9 and 10, Tarkir: Dragonstorm
// leftovers that needed a seam (#4463). Each test plays the real catalog
// entry and asserts what its text says.

const (
	stdB9TradeRouteEnvoyOracle   = "95d51048-be32-4c5f-8b75-fca3a891c9d6"
	stdB9TravelingBotanistOracle = "14934ab0-41ab-44f8-bb28-2b0457eb6f5f"
	stdB9WardenOracle            = "13bf8103-3db3-4ee6-a6bb-441aabfd9e41"
	stdB9DragonclawStrikeOracle  = "69e0d9f1-f9dd-4614-ad9f-13e7153710ce"
	stdB9EffortlessMasterOracle  = "185e0892-194f-43ac-853f-1e142bf8476e"
	stdB9FangkeepersOracle       = "06343800-b555-4a13-ad4d-19e37d200138"
	stdB9FelotharOracle          = "fd666669-ad23-4e41-b78c-1cd7b82eb8b4"
	stdB9GurmagNightwatchOracle  = "fde50a0d-9bc6-45f4-873e-3de81a513fac"
	stdB9HostOracle              = "3431338b-928a-47f2-bc2f-f3c5d8eda8d0"
	stdB9KotisOracle             = "2eb2f6a2-e31a-4d3e-93ce-100394899014"
	stdB9LieInWaitOracle         = "4dc0831a-0eb1-4f7d-9591-6e670e180116"
	stdB9SiegebreakerOracle      = "1f82e4b9-e67f-4f34-ac4c-57b0b300ee42"
	stdB9RediscoverOracle        = "680a8e14-053f-4cde-a060-e1b36873cf14"
	stdB9RiverwheelOracle        = "917aef2b-0e6d-460e-8745-8da8423d1592"
	stdB9StalwartOracle          = "45392882-e4b9-4c74-a10f-985c2b4e2be9"
	stdB9BattlecrierOracle       = "fadf984c-c224-432a-87b9-b34f929b8409"
	stdB9ThunderOfUnityOracle    = "8419d1d5-bb0e-4a2d-bb6f-67957d035dde"
	stdB9ZurgoOracle             = "393d0bf8-5f2e-4eaa-91a1-9132429a7371"
	stdB9EmbermouthOracle        = "b5208c2c-b839-42f6-8717-ee4bbaf1385b"
	stdB9CorrodingOracle         = "4c55ca3b-1254-4d44-96ac-1e1cadaf49f5"
)

func TestStdB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB9TradeRouteEnvoyOracle:   "Trade Route Envoy",
		stdB9TravelingBotanistOracle: "Traveling Botanist",
		stdB9WardenOracle:            "Warden of the Grove",
		stdB9DragonclawStrikeOracle:  "Dragonclaw Strike",
		stdB9EffortlessMasterOracle:  "Effortless Master",
		stdB9FangkeepersOracle:       "Fangkeeper's Familiar",
		stdB9FelotharOracle:          "Felothar, Dawn of the Abzan",
		stdB9GurmagNightwatchOracle:  "Gurmag Nightwatch",
		stdB9HostOracle:              "Host of the Hereafter",
		stdB9KotisOracle:             "Kotis, the Fangkeeper",
		stdB9LieInWaitOracle:         "Lie in Wait",
		stdB9SiegebreakerOracle:      "Mardu Siegebreaker",
		stdB9RediscoverOracle:        "Rediscover the Way",
		stdB9RiverwheelOracle:        "Riverwheel Sweep",
		stdB9StalwartOracle:          "Stalwart Successor",
		stdB9BattlecrierOracle:       "Temur Battlecrier",
		stdB9ThunderOfUnityOracle:    "Thunder of Unity",
		stdB9ZurgoOracle:             "Zurgo, Thunder's Decree",
		stdB9EmbermouthOracle:        "Embermouth Sentinel",
		stdB9CorrodingOracle:         "Corroding Dragonstorm",
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

// stdB9Push puts a creature on the active player's battlefield.
func stdB9Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, typeLine, oracle, power, toughness)
}

// --- Trade Route Envoy ---------------------------------------------------------

func TestStdB9TradeRouteEnvoyDrawsWithACounterCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	hand := me.Hand.Size()
	envoy := effectstest.BlbGLCastCreature(t, g, "Trade Route Envoy", "Creature — Dog Soldier", stdB9TradeRouteEnvoyOracle, 4, 3, nil)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (a card was drawn)", got, hand+1)
	}
	if got := effectstest.CountersOn(g, envoy, game.CounterPlusOne); got != 0 {
		t.Errorf("the Envoy got %d +1/+1 counters although it drew", got)
	}
}

func TestStdB9TradeRouteEnvoyGrowsWithoutACounterCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	envoy := effectstest.BlbGLCastCreature(t, g, "Trade Route Envoy", "Creature — Dog Soldier", stdB9TradeRouteEnvoyOracle, 4, 3, nil)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (nothing drawn)", got, hand)
	}
	if got := effectstest.CountersOn(g, envoy, game.CounterPlusOne); got != 1 {
		t.Errorf("the Envoy has %d +1/+1 counters, want 1", got)
	}
}

func TestStdB9TradeRouteEnvoyGrowsWhenTheLibraryCannotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	envoy := effectstest.BlbGLCastCreature(t, g, "Trade Route Envoy", "Creature — Dog Soldier", stdB9TradeRouteEnvoyOracle, 4, 3, nil)
	if got := effectstest.CountersOn(g, envoy, game.CounterPlusOne); got != 1 {
		t.Errorf("no card could be drawn: the Envoy has %d +1/+1 counters, want 1", got)
	}
}

// --- Warden of the Grove -------------------------------------------------------

func TestStdB9WardenGrowsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	warden := stdB9Push(g, me.ID, "Warden of the Grove", "Creature — Hydra", stdB9WardenOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, warden, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after my end step, want 1", got)
	}
}

func TestStdB9WardenMakesAnEnteringCreatureEndureForItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	warden := stdB9Push(g, me.ID, "Warden of the Grove", "Creature — Hydra", stdB9WardenOracle, 2, 2)
	effectstest.StdB9AddCounter(t, g, warden, game.CounterPlusOne, 2)
	effectstest.StdB9AddCounter(t, g, warden, "quest", 1) // every kind counts: X is 3
	bear := effectstest.BlbGLCastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2, nil)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the Bear endured for %d counters, want 3", got)
	}

	// The Spirit option makes an X/X white Spirit instead.
	cub := effectstest.BlbGLCastCreature(t, g, "Cub", "Creature — Bear", "", 2, 2, nil)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 3 || spirits[0].Toughness != 3 {
		t.Fatalf("Spirits %+v, want one 3/3", spirits)
	}
	if got := effectstest.CountersOn(g, cub, game.CounterPlusOne); got != 0 {
		t.Errorf("choosing the Spirit still put %d counters on the Cub", got)
	}
}

func TestStdB9WardenIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	warden := stdB9Push(g, me.ID, "Warden of the Grove", "Creature — Hydra", stdB9WardenOracle, 2, 2)
	effectstest.StdB9AddCounter(t, g, warden, game.CounterPlusOne, 1)
	stdB9Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1}.Apply(
			effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: warden}))
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("a token or an opponent's creature entering made something endure")
	}
}

// --- Host of the Hereafter -----------------------------------------------------

func TestStdB9HostEntersWithTwoCountersAndMovesTheirCountersOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	host := effectstest.BlbGLCastCreature(t, g, "Host of the Hereafter", "Creature — Zombie Warlock", stdB9HostOracle, 2, 2, nil)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 2 {
		t.Fatalf("the Host entered with %d +1/+1 counters, want 2", got)
	}
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, bear, game.CounterPlusOne, 1)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	effectstest.StdB9Kill(t, g, bear)
	effectstest.AnswerPickTarget(t, g, host)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 3 {
		t.Errorf("the Host has %d +1/+1 counters, want 3", got)
	}
	if got := effectstest.CountersOn(g, host, "quest"); got != 1 {
		t.Errorf("the Host has %d quest counters, want 1 (every kind moves)", got)
	}
}

func TestStdB9HostGivesItsOwnCountersToAnotherCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	host := effectstest.BlbGLCastCreature(t, g, "Host of the Hereafter", "Creature — Zombie Warlock", stdB9HostOracle, 2, 2, nil)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9Kill(t, g, host)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
}

func TestStdB9HostIgnoresACounterlessCreatureAndAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbGLCastCreature(t, g, "Host of the Hereafter", "Creature — Zombie Warlock", stdB9HostOracle, 2, 2, nil)
	plain := stdB9Push(g, me.ID, "Plain", "Creature — Bear", "", 2, 2)
	effectstest.StdB9Kill(t, g, plain)
	if pickTargetOpen(g) {
		t.Error("a creature with no counters dying triggered the Host")
	}
	theirs := stdB9Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, theirs, game.CounterPlusOne, 1)
	effectstest.StdB9Kill(t, g, theirs)
	if pickTargetOpen(g) {
		t.Error("an opponent's creature dying triggered the Host")
	}
}

func pickTargetOpen(g *game.Game) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return true
		}
	}
	return false
}

// --- Stalwart Successor --------------------------------------------------------

func TestStdB9StalwartSuccessorGrowsACreatureOnTheFirstCountersOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Stalwart Successor", "Creature — Human Warrior", stdB9StalwartOracle, 3, 2)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("first counters on the Bear: %d +1/+1 counters, want 1", got)
	}
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("second counters the same turn: %d +1/+1 counters, want still 1", got)
	}
	other := stdB9Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, other, game.CounterPlusOne, 2)
	if got := effectstest.CountersOn(g, other, game.CounterPlusOne); got != 3 {
		t.Errorf("a different creature's first counters: %d +1/+1 counters, want 3 (two, plus one)", got)
	}
}

func TestStdB9StalwartSuccessorResetsEachTurnAndIgnoresOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Stalwart Successor", "Creature — Human Warrior", stdB9StalwartOracle, 3, 2)
	theirs := stdB9Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, theirs, "quest", 1)
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d +1/+1 counters", got)
	}
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the next turn's first counters: %d +1/+1 counters, want 2", got)
	}
}

func TestStdB9StalwartSuccessorCountsTheCountersACreatureEntersWith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Stalwart Successor", "Creature — Human Warrior", stdB9StalwartOracle, 3, 2)
	host := effectstest.BlbGLCastCreature(t, g, "Host of the Hereafter", "Creature — Zombie Warlock", stdB9HostOracle, 2, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterPlusOne); got != 3 {
		t.Errorf("the Host entered with %d +1/+1 counters, want 3 (two, and one from the Successor)", got)
	}
}

// --- Effortless Master ---------------------------------------------------------

func TestStdB9EffortlessMasterEntersWithCountersAfterTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	// A first spell this turn, then the Master is the second.
	if err := g.CastSpell(me.ID, effectstest.HandCardForTest(me, "Cantrip", "Instant", ""), game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	master := effectstest.BlbGLCastCreature(t, g, "Effortless Master", "Creature — Orc Monk", stdB9EffortlessMasterOracle, 4, 3, nil)
	if got := effectstest.CountersOn(g, master, game.CounterPlusOne); got != 2 {
		t.Errorf("the Master entered with %d +1/+1 counters after a second spell, want 2", got)
	}
	for _, kw := range []string{"vigilance", "menace"} {
		if !effectstest.HasEffectiveKeyword(t, g, master, kw) {
			t.Errorf("the Master lacks %s", kw)
		}
	}
}

func TestStdB9EffortlessMasterEntersBareAsTheFirstSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	master := effectstest.BlbGLCastCreature(t, g, "Effortless Master", "Creature — Orc Monk", stdB9EffortlessMasterOracle, 4, 3, nil)
	if got := effectstest.CountersOn(g, master, game.CounterPlusOne); got != 0 {
		t.Errorf("the Master entered with %d +1/+1 counters as the first spell, want 0", got)
	}
}

// --- Traveling Botanist --------------------------------------------------------

func TestStdB9TravelingBotanistTakesALandWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	botanist := stdB9Push(g, me.ID, "Traveling Botanist", "Creature — Dog Scout", stdB9TravelingBotanistOracle, 2, 3)
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.DeclareAttack(t, g, opp.ID, botanist)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if !me.Hand.Contains(land) || me.Library.Contains(land) {
		t.Error("the land should have gone to my hand")
	}
}

func TestStdB9TravelingBotanistLandDeclinedMayGoToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	botanist := stdB9Push(g, me.ID, "Traveling Botanist", "Creature — Dog Scout", stdB9TravelingBotanistOracle, 2, 3)
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.DeclareAttack(t, g, opp.ID, botanist)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false) // not into the hand
	effectstest.AnswerMayChoice(t, g, me.ID, true)  // into the graveyard
	if me.Hand.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the declined land should have gone to the graveyard")
	}
}

func TestStdB9TravelingBotanistNonLandOffersOnlyTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	botanist := stdB9Push(g, me.ID, "Traveling Botanist", "Creature — Dog Scout", stdB9TravelingBotanistOracle, 2, 3)
	spell := effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	effectstest.DeclareAttack(t, g, opp.ID, botanist)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if me.Hand.Contains(spell) || !me.Library.Contains(spell) {
		t.Error("declining should leave a non-land card on top")
	}
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked again after declining")
	}
}

func TestStdB9TravelingBotanistNonLandMayGoToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	botanist := stdB9Push(g, me.ID, "Traveling Botanist", "Creature — Dog Scout", stdB9TravelingBotanistOracle, 2, 3)
	spell := effectstest.SeedLibraryTop(me, "Bolt", "Instant", "{R}")
	effectstest.DeclareAttack(t, g, opp.ID, botanist)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if !me.Graveyard.Contains(spell) {
		t.Error("the card should be in the graveyard")
	}
}

func TestStdB9TravelingBotanistDoesNotTriggerWhenAVigilantAttackerStaysUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	botanist := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Traveling Botanist", TypeLine: "Creature — Dog Scout", OracleID: stdB9TravelingBotanistOracle,
		Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID, Keywords: []string{"vigilance"},
	})
	effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.DeclareAttack(t, g, opp.ID, botanist)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a vigilance attacker did not become tapped")
	}
}

// --- Dragonclaw Strike ---------------------------------------------------------

func TestStdB9DragonclawStrikeDoublesThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := stdB9Push(g, me.ID, "Mine", "Creature — Bear", "", 3, 2)
	theirs := stdB9Push(g, opp.ID, "Theirs", "Creature — Ox", "", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Dragonclaw Strike", "Sorcery", stdB9DragonclawStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	got, _ := effectstest.BattlefieldCard(g, mine)
	if got.CurrentPower() != 6 || got.CurrentToughness() != 4 {
		t.Errorf("my creature is %d/%d, want 6/4 after doubling", got.CurrentPower(), got.CurrentToughness())
	}
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("their 2/6 should have died to 6 damage from the doubled creature")
	}
	if got.DamageMarked != 2 {
		t.Errorf("my creature took %d damage, want 2", got.DamageMarked)
	}
}

func TestStdB9DragonclawStrikeWithoutASecondTargetOnlyDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := stdB9Push(g, me.ID, "Mine", "Creature — Bear", "", 3, 2)
	theirs := stdB9Push(g, opp.ID, "Theirs", "Creature — Ox", "", 2, 6)
	effectstest.CastCatalogSpell(t, g, "Dragonclaw Strike", "Sorcery", stdB9DragonclawStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
	})
	effectstest.PassPriorityAroundTable(t, g)
	got, _ := effectstest.BattlefieldCard(g, mine)
	if got.CurrentPower() != 6 || got.CurrentToughness() != 4 || got.DamageMarked != 0 {
		t.Errorf("my creature is %d/%d with %d damage, want 6/4 and none", got.CurrentPower(), got.CurrentToughness(), got.DamageMarked)
	}
	if ox, _ := effectstest.BattlefieldCard(g, theirs); ox.DamageMarked != 0 {
		t.Error("with no second target nothing should fight")
	}
}

func TestStdB9DragonclawStrikeRefusesMyCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mine := stdB9Push(g, me.ID, "Mine", "Creature — Bear", "", 3, 2)
	other := stdB9Push(g, me.ID, "Other", "Creature — Bear", "", 3, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Dragonclaw Strike", "Sorcery", stdB9DragonclawStrikeOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: other, Slot: 1},
	})
	if err == nil {
		t.Error("the fight target must be a creature an opponent controls")
	}
}

// --- Fangkeeper's Familiar -----------------------------------------------------

func stdB9ChooseMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, chooser)
	if c == nil {
		t.Fatalf("no mode_pick prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestStdB9FangkeepersFamiliarGainsLifeAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.BlbGLCastCreature(t, g, "Fangkeeper's Familiar", "Creature — Snake", stdB9FangkeepersOracle, 3, 3, nil)
	stdB9ChooseMode(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("life %d, want %d", got, life+3)
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 3 {
		t.Errorf("want a surveil 3 prompt over three cards: %+v", c)
	}
}

func TestStdB9FangkeepersFamiliarDestroysAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	aura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.BlbGLCastCreature(t, g, "Fangkeeper's Familiar", "Creature — Snake", stdB9FangkeepersOracle, 3, 3, nil)
	stdB9ChooseMode(t, g, me.ID, 1)
	effectstest.AnswerPickTarget(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); ok {
		t.Error("the enchantment should be destroyed")
	}
}

func TestStdB9FangkeepersFamiliarCountersACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bear, Name: "Their Bear", TypeLine: "Creature — Bear", Keywords: []string{"flash"}, Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("their cast: %v", err)
	}
	familiar := effectstest.HandCardForTest(me, "Fangkeeper's Familiar", "Creature — Snake", stdB9FangkeepersOracle)
	if err := g.CastSpell(me.ID, familiar, game.CastSpellParams{}); err != nil {
		t.Fatalf("my cast: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	stdB9ChooseMode(t, g, me.ID, 2)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Error("their creature spell should have been countered into their graveyard")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the countered creature entered the battlefield")
	}
}

// --- Felothar, Dawn of the Abzan -----------------------------------------------

func TestStdB9FelotharSacrificesToGrowEveryCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	theirs := stdB9Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	felothar := effectstest.BlbGLCastCreature(t, g, "Felothar, Dawn of the Abzan", "Legendary Creature — Human Warrior", stdB9FelotharOracle, 3, 3, nil)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 {
		t.Fatalf("want a may-sacrifice prompt: %+v", prompt)
	}
	if effectstest.HasID(prompt.ChooseCards, land) {
		t.Error("a land was offered for the sacrifice")
	}
	if !effectstest.HasID(prompt.ChooseCards, relic) || !effectstest.HasID(prompt.ChooseCards, felothar) {
		t.Error("the artifact and Felothar himself should be offered")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Bear": bear, "Felothar": felothar} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, got)
		}
	}
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestStdB9FelotharDecliningTheSacrificeGivesNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.BlbGLCastCreature(t, g, "Felothar, Dawn of the Abzan", "Legendary Creature — Human Warrior", stdB9FelotharOracle, 3, 3, nil)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters after declining", got)
	}
}

func TestStdB9FelotharTriggersWhenItAttacksAndMaySacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	felothar := stdB9Push(g, me.ID, "Felothar, Dawn of the Abzan", "Legendary Creature — Human Warrior", stdB9FelotharOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, felothar)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, felothar) // sacrifice Felothar himself
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, felothar); ok {
		t.Error("Felothar should have been sacrificed")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

// --- Gurmag Nightwatch ---------------------------------------------------------

func TestStdB9GurmagNightwatchKeepsOneOnTopAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.BlbGLCastCreature(t, g, "Gurmag Nightwatch", "Creature — Human Ranger", stdB9GurmagNightwatchOracle, 3, 3, nil)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if top != ids[1] {
		t.Errorf("the card kept on top is %v, want B", top)
	}
	if !me.Graveyard.Contains(ids[0]) || !me.Graveyard.Contains(ids[2]) {
		t.Error("the other two looked-at cards should be in the graveyard")
	}
	if me.Graveyard.Contains(ids[3]) || !me.Library.Contains(ids[3]) {
		t.Error("the fourth card was not looked at and must stay in the library")
	}
}

func TestStdB9GurmagNightwatchMayPutEverythingIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.BlbGLCastCreature(t, g, "Gurmag Nightwatch", "Creature — Human Ranger", stdB9GurmagNightwatchOracle, 3, 3, nil)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%v should be in the graveyard", id)
		}
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the fourth card should stay in the library")
	}
}

// --- Kotis, the Fangkeeper -----------------------------------------------------

func TestStdB9KotisExilesXCardsAndCastsTheCheapOnesFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	kotis := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kotis, the Fangkeeper", TypeLine: "Legendary Creature — Zombie Warrior", OracleID: stdB9KotisOracle,
		Power: 2, Toughness: 1, Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	deep := effectstest.SeedLibraryTop(opp, "Deep", "Sorcery", "{1}")
	pricey := effectstest.SeedLibraryTop(opp, "Pricey", "Sorcery", "{4}")
	cheap := effectstest.SeedLibraryTop(opp, "Cheap", "Instant", "{1}")
	effectstest.DeclareAttack(t, g, opp.ID, kotis)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	// X = 2: the top two cards (Cheap, Pricey) are exiled.
	if !g.Exile.Contains(pricey) && !opp.Library.Contains(pricey) {
		t.Fatal("Pricey vanished")
	}
	if !opp.Library.Contains(deep) {
		t.Error("only X = 2 cards are exiled; Deep stays in the library")
	}
	offer := effectstest.KeywordCastPromptFor(g, me.ID)
	if offer == nil {
		t.Fatalf("no free-cast offer for Kotis's controller: %+v", g.PendingChoices)
	}
	if got := effectstest.KeywordCastCard(offer); got != cheap {
		t.Fatalf("the offer is for %v, want the mana value 1 card", got)
	}
	effectstest.AnswerKeywordCast(t, g, offer, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(cheap) {
		t.Error("Cheap should have been cast and resolved into its owner's graveyard")
	}
	if !g.Exile.Contains(pricey) {
		t.Error("the mana value 4 card can't be cast with X = 2 and stays in exile")
	}
}

// --- Lie in Wait ---------------------------------------------------------------

func TestStdB9LieInWaitReturnsTheCardAndDamagesByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	dead := effectstest.PushGraveyardCardTyped(me, "Big Dead", "Creature — Ogre")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == dead {
				me.Graveyard.Cards[i].Power = 3
			}
		}
	})
	victim := stdB9Push(g, opp.ID, "Victim", "Creature — Ox", "", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Lie in Wait", "Sorcery", stdB9LieInWaitOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) || me.Graveyard.Contains(dead) {
		t.Error("the creature card should return to my hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("3 damage should have killed the 2/3")
	}
}

func TestStdB9LieInWaitNeedsBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	dead := effectstest.PushGraveyardCardTyped(me, "Big Dead", "Creature — Ogre")
	err := effectstest.CastCatalogSpellErr(t, g, "Lie in Wait", "Sorcery", stdB9LieInWaitOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead, Slot: 0},
	})
	if err == nil {
		t.Error("Lie in Wait can't be cast without a creature on the battlefield as its second target")
	}
}

func TestStdB9LieInWaitRefusesANoncreatureCardInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	spell := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	victim := stdB9Push(g, opp.ID, "Victim", "Creature — Ox", "", 2, 3)
	err := effectstest.CastCatalogSpellErr(t, g, "Lie in Wait", "Sorcery", stdB9LieInWaitOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: spell, Slot: 0},
		{Kind: game.TargetCard, ID: victim, Slot: 1},
	})
	if err == nil {
		t.Error("the first target must be a creature card")
	}
}

// --- Mardu Siegebreaker --------------------------------------------------------

func stdB9CastSiegebreaker(t *testing.T, g *game.Game, exile uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.BlbGLCastCreature(t, g, "Mardu Siegebreaker", "Creature — Human Warrior", stdB9SiegebreakerOracle, 4, 4, nil)
	if exile != uuid.Nil {
		effectstest.AnswerPickTarget(t, g, exile)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id
}

func TestStdB9SiegebreakerExilesUntilItLeavesAndReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	siege := stdB9CastSiegebreaker(t, g, bear)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Fatal("the Bear should be exiled")
	}
	for _, kw := range []string{"deathtouch", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, siege, kw) {
			t.Errorf("Siegebreaker lacks %s", kw)
		}
	}
	effectstest.StdB9Kill(t, g, siege)
	if got := len(effectstest.SocB8Named(g, me.ID, "Bear")); got != 1 {
		t.Errorf("%d Bears under my control, want the exiled one back", got)
	}
}

func TestStdB9SiegebreakerAttackMakesATappedCopyAttackingEachOpponentThenSacrificesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	// Counters and effects on the exiled creature are not copied.
	effectstest.StdB9AddCounter(t, g, bear, game.CounterPlusOne, 2)
	siege := stdB9CastSiegebreaker(t, g, bear)
	effectstest.DeclareAttack(t, g, opp.ID, siege)
	effectstest.PassPriorityAroundTable(t, g)
	copies := effectstest.SocB8Named(g, me.ID, "Bear")
	if len(copies) != 3 {
		t.Fatalf("%d Bear tokens, want one per opponent (3)", len(copies))
	}
	attacked := map[uuid.UUID]bool{}
	for _, c := range copies {
		if !c.IsToken() || !c.Tapped || c.Power != 2 || c.Toughness != 2 {
			t.Errorf("token %+v: want a tapped 2/2 token copy with no counters", c)
		}
		if got := effectstest.CountersOn(g, c.InstanceID, game.CounterPlusOne); got != 0 {
			t.Errorf("a token copy has %d counters", got)
		}
		var defender uuid.UUID
		g.ReadSnapshot(func() { defender = g.DefendingPlayerForAttackerForEffect(c.InstanceID) })
		attacked[defender] = true
	}
	for _, p := range []*game.Player{opp, other, third} {
		if !attacked[p.ID] {
			t.Errorf("no token attacks %s", p.Name)
		}
	}
	// "At the beginning of your next end step, sacrifice those tokens."
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Bear")); got != 0 {
		t.Errorf("%d token copies remain after the end step, want 0", got)
	}
}

func TestStdB9SiegebreakerWithNothingExiledMakesNoCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	siege := stdB9CastSiegebreaker(t, g, uuid.Nil)
	// "Up to one": decline the target.
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the up-to-one target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, siege)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Bear")); got != 1 {
		t.Errorf("%d Bears, want only the original", got)
	}
}

func TestStdB9SiegebreakerCopiesANoncreatureCardTappedAndNotAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	// An animated artifact creature, exiled, is just an artifact card in exile.
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	siege := stdB9CastSiegebreaker(t, g, golem)
	effectstest.DeclareAttack(t, g, opp.ID, siege)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Golem")); got != 3 {
		t.Errorf("%d Golem tokens, want 3", got)
	}
}

// --- Rediscover the Way --------------------------------------------------------

func TestStdB9RediscoverTheWayDigsOnChaptersIAndII(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	saga := effectstest.CastCatalogSpell(t, g, "Rediscover the Way", "Enchantment — Saga", stdB9RediscoverOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// I: look at the top three, one into hand, the rest on the bottom.
	hand := me.Hand.Size()
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(ids[1]) {
		t.Error("chapter I: B should be in my hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 2 {
		t.Fatalf("the two cards not taken should be ordered onto the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[2], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{ids[2], ids[0]}) {
		t.Errorf("the bottom two are %v, want [C A]", got)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[3] {
		t.Error("the fourth card is now on top")
	}
	if effectstest.LoreCountersOn(g, saga) != 1 {
		t.Fatalf("lore counters %d, want 1", effectstest.LoreCountersOn(g, saga))
	}
}

func TestStdB9RediscoverTheWayGivesDoubleStrikeToEachNoncreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := stdB9Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	ox := stdB9Push(g, me.ID, "Ox", "Creature — Ox", "", 2, 2)
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rediscover the Way", TypeLine: "Enchantment — Saga", OracleID: stdB9RediscoverOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	effectstest.AdvanceToMain(t, g)
	// Chapter III arrives with the next lore counter.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, saga); ok {
		t.Fatal("the Saga should be sacrificed after chapter III resolves")
	}
	// A creature spell triggers nothing.
	effectstest.HandCardForTest(me, "Grizzly", "Creature — Bear", "")
	effectstest.CastCreatureCard(t, g, "Grizzly", "Creature — Bear", "", 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered chapter III's ability")
	}
	// A noncreature spell triggers it, and every one does, for the turn.
	for i, who := range []uuid.UUID{bear, ox} {
		effectstest.CastCatalogSpell(t, g, "Cantrip", "Instant", "", nil)
		pick := effectstest.LatestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatalf("cast %d: no target prompt for the double strike trigger", i)
		}
		effectstest.AnswerPickTarget(t, g, who)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, who, "double strike") {
			t.Errorf("cast %d: the chosen creature lacks double strike", i)
		}
	}
}

// --- Riverwheel Sweep ----------------------------------------------------------

func TestStdB9RiverwheelSweepTapsStunsAndOffersOneExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	victim := stdB9Push(g, opp.ID, "Victim", "Creature — Ox", "", 2, 3)
	top := effectstest.SeedLibrary(me, "First", "Second", "Third")
	effectstest.CastCatalogSpell(t, g, "Riverwheel Sweep", "Sorcery", stdB9RiverwheelOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	v, _ := effectstest.BattlefieldCard(g, victim)
	if !v.Tapped || v.Counters[game.CounterStun] != 3 {
		t.Errorf("victim tapped %v with %d stun counters, want tapped with 3", v.Tapped, v.Counters[game.CounterStun])
	}
	if !g.Exile.Contains(top[0]) || !g.Exile.Contains(top[1]) || g.Exile.Contains(top[2]) {
		t.Fatal("the top two cards should be exiled")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Error("the chosen card should be playable")
	}
	if effectstest.B12ExiledPlayableBy(t, g, top[0], me.ID) {
		t.Error("the other exiled card must not be playable")
	}
	// "Until the end of your next turn": still there on my next turn,
	// gone once it has ended.
	seat := g.Turn.ActiveSeat
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	if !effectstest.B12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Error("the card should still be playable during my next turn")
	}
	effectstest.EndTurn(t, g)
	if effectstest.B12ExiledPlayableBy(t, g, top[1], me.ID) {
		t.Error("the permission should have ended with my next turn")
	}
}

func TestStdB9RiverwheelSweepFizzlesWithoutALegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	victim := stdB9Push(g, opp.ID, "Victim", "Creature — Ox", "", 2, 3)
	top := effectstest.SeedLibrary(me, "First", "Second")
	effectstest.CastCatalogSpell(t, g, "Riverwheel Sweep", "Sorcery", stdB9RiverwheelOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.StdB9Kill(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(top[0]) {
		t.Error("the spell should not resolve with its only target gone: nothing is exiled")
	}
}

// --- Temur Battlecrier ---------------------------------------------------------

func TestStdB9TemurBattlecrierDiscountsSpellsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Temur Battlecrier", "Creature — Orc Ranger", stdB9BattlecrierOracle, 4, 3)
	stdB9Push(g, me.ID, "Big", "Creature — Ox", "", 5, 5)
	stdB9Push(g, me.ID, "Small", "Creature — Ox", "", 3, 3)
	effectstest.AdvanceToMain(t, g)
	// Two creatures with power 4 or greater: {1}{1}{1} costs {1}.
	effectstest.FillPool(me, 1)
	effectstest.CastFromHandForTest(t, g, me, "Pricey", "Sorcery", "{3}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("the pool still holds %d mana", me.ManaPool.Total())
	}
}

func TestStdB9TemurBattlecrierNeedsPowerFourAndOnlyHelpsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Temur Battlecrier", "Creature — Orc Ranger", stdB9BattlecrierOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	// One big creature (itself): {3} costs {2}; one mana is not enough.
	effectstest.FillPool(me, 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pricey", TypeLine: "Sorcery", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a {3} spell with one big creature costs {2}: one mana can't pay it")
	}
	effectstest.FillPool(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("two mana should pay for the discounted {3} spell: %v", err)
	}
}

func TestStdB9TemurBattlecrierDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stdB9Push(g, opp.ID, "Temur Battlecrier", "Creature — Orc Ranger", stdB9BattlecrierOracle, 4, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(opp, 2)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Flash Spell", TypeLine: "Instant", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("it isn't their turn: no discount, so two mana can't pay {3}")
	}
	_ = me
}

// --- Thunder of Unity ----------------------------------------------------------

func TestStdB9ThunderOfUnityChapterIDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand, life := me.Hand.Size(), effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Thunder of Unity", "Enchantment — Saga", stdB9ThunderOfUnityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (two drawn)", got, hand+2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
}

func TestStdB9ThunderOfUnityChaptersIIAndIIIDrainForEachCreatureThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Thunder of Unity", TypeLine: "Enchantment — Saga", OracleID: stdB9ThunderOfUnityOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 1},
	})
	opps := []*game.Player{g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]}
	before := effectstest.LifeOf(g, opps[0].ID)
	// The lore counter arrives as the precombat main phase begins: chapter II.
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, saga) != 2 {
		t.Fatalf("lore counters %d, want 2", effectstest.LoreCountersOn(g, saga))
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCreatureCard(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCreatureCard(t, g, "Ox", "Creature — Ox", "", 2, 2)
	for _, o := range opps {
		if got := effectstest.LifeOf(g, o.ID); got != before-2 {
			t.Errorf("%s is at %d, want %d (two creatures entered)", o.Name, got, before-2)
		}
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("I am at %d, want %d", got, life+2)
	}
	// An opponent's creature entering doesn't drain.
	stdB9Push(g, opps[0].ID, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("an opponent's creature drained: I am at %d", got)
	}
}

// --- Zurgo, Thunder's Decree ---------------------------------------------------

func TestStdB9ZurgoWarriorsSurviveTheMobilizeSacrificeInYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	zurgo := stdB9Push(g, me.ID, "Zurgo, Thunder's Decree", "Legendary Creature — Orc Warrior", stdB9ZurgoOracle, 2, 4)
	effectstest.DeclareAttack(t, g, opp.ID, zurgo)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	warriors := effectstest.SocB8Named(g, me.ID, "Warrior")
	if len(warriors) != 2 {
		t.Fatalf("%d Warrior tokens, want 2 (mobilize 2)", len(warriors))
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Warrior")); got != 2 {
		t.Errorf("%d Warriors after my end step, want both to survive", got)
	}
	// Outside the end step the tokens can be sacrificed like any permanent.
	effectstest.EndTurn(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(warriors[0].InstanceID) })
	if got := len(effectstest.SocB8Named(g, me.ID, "Warrior")); got != 1 {
		t.Errorf("%d Warriors after a sacrifice on another turn, want 1", got)
	}
}

func TestStdB9ZurgoGoneBeforeTheEndStepLetsTheTokensBeSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	zurgo := stdB9Push(g, me.ID, "Zurgo, Thunder's Decree", "Legendary Creature — Orc Warrior", stdB9ZurgoOracle, 2, 4)
	effectstest.DeclareAttack(t, g, opp.ID, zurgo)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	effectstest.StdB9Kill(t, g, zurgo)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.SocB8Named(g, me.ID, "Warrior")); got != 0 {
		t.Errorf("%d Warriors remain, want 0 without Zurgo", got)
	}
}

// --- Embermouth Sentinel -------------------------------------------------------

func TestStdB9EmbermouthSentinelPutsABasicLandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Shock Land", "Land — Mountain Forest"),
		game.Card{Name: "Spell", TypeLine: "Sorcery"})
	effectstest.BlbGLCastCreature(t, g, "Embermouth Sentinel", "Artifact Creature — Chimera", stdB9EmbermouthOracle, 2, 1, nil)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || effectstest.SearchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Fatalf("the search should offer only the basic land: %+v", c)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if top, _ := me.Library.Top(); top.InstanceID != ids[0] {
		t.Error("the basic land should be on top of the library")
	}
	if _, ok := effectstest.BattlefieldCard(g, ids[0]); ok {
		t.Error("with no Dragon the land stays in the library")
	}
}

func TestStdB9EmbermouthSentinelPutsItOnTheBattlefieldTappedWithADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	stdB9Push(g, me.ID, "Dragon", "Creature — Dragon", "", 4, 4)
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.BlbGLCastCreature(t, g, "Embermouth Sentinel", "Artifact Creature — Chimera", stdB9EmbermouthOracle, 2, 1, nil)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	land, ok := effectstest.BattlefieldCard(g, ids[0])
	if !ok || !land.Tapped {
		t.Errorf("the land should be on the battlefield tapped: %v tapped=%v", ok, land.Tapped)
	}
}

func TestStdB9EmbermouthSentinelMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.BlbGLCastCreature(t, g, "Embermouth Sentinel", "Artifact Creature — Chimera", stdB9EmbermouthOracle, 2, 1, nil)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the search: %v", err)
	}
	if !me.Library.Contains(ids[0]) {
		t.Error("the land left the library on a declined search")
	}
}

// --- Corroding Dragonstorm -----------------------------------------------------

func TestStdB9CorrodingDragonstormDrainsAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.SeedLibrary(me, "A", "B", "C")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Corroding Dragonstorm", "Enchantment", stdB9CorrodingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i <= 3; i++ {
		if got := effectstest.LifeOf(g, g.Seats[(seat+i)%4].ID); got != 38 {
			t.Errorf("opponent %d is at %d, want 38", i, got)
		}
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want %d", got, life+2)
	}
	if c := effectstest.SurveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Errorf("want a surveil 2 prompt: %+v", c)
	}
}

func TestStdB9CorrodingDragonstormReturnsToHandWhenADragonEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	enchantment := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Corroding Dragonstorm", TypeLine: "Enchantment", OracleID: stdB9CorrodingOracle,
		Owner: me.ID, Controller: me.ID,
	})
	// An opponent's Dragon, or my non-Dragon, does nothing.
	stdB9Push(g, opp.ID, "Their Dragon", "Creature — Dragon", "", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCreatureCard(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if _, ok := effectstest.BattlefieldCard(g, enchantment); !ok {
		t.Fatal("neither an opponent's Dragon nor a Bear should bounce it")
	}
	effectstest.CastCreatureCard(t, g, "Dragon", "Creature — Dragon", "", 4, 4)
	if _, ok := effectstest.BattlefieldCard(g, enchantment); ok {
		t.Error("my Dragon entering should return the enchantment to my hand")
	}
	if !me.Hand.Contains(enchantment) {
		t.Error("the enchantment should be in my hand")
	}
}
