package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b27_cards_test.go — Avatar: The Last Airbender Eternal, batch 27
// (#5075): every card cast, triggered or activated through the engine.

const (
	avB27SwampbendersOracle = "feb00c6d-ff43-49bc-84d5-57e2d4651258"
	avB27ChargeOracle       = "dcc29e59-341c-4175-aa54-f535e2c3e355"
	avB27MaiZukoOracle      = "b86ba0aa-89f3-4924-b3c4-f7ca44468e9b"
	avB27AangKataraOracle   = "481c3e14-b670-4fab-aa9f-6ce5b514096d"
	avB27TophOracle         = "3741206f-8ef5-4b8f-ae4c-df3ab0a17683"
	avB27SokkaSukiOracle    = "e84becc0-6e15-4a74-aec0-9746b86ab5cc"
	avB27HeistOracle        = "b8f05217-1f61-4d6d-be17-7fb15f1d1e32"
	avB27MusingsOracle      = "6de489bf-6922-4f97-8001-efa8cb5a27cb"
	avB27StudentOracle      = "a468f6ab-c0a8-4295-8b01-7147d3f889b4"
	avB27DukeOracle         = "1c55047d-badc-49bb-b791-e9759577e4eb"
)

func TestAvatarB27EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB27SwampbendersOracle: "Swampbenders",
		avB27ChargeOracle:       "Sokka's Charge",
		avB27MaiZukoOracle:      "Mai and Zuko",
		avB27AangKataraOracle:   "Aang and Katara",
		avB27TophOracle:         "Toph, Greatest Earthbender",
		avB27SokkaSukiOracle:    "Sokka and Suki",
		avB27HeistOracle:        "Momo's Heist",
		avB27MusingsOracle:      "Uncle's Musings",
		avB27StudentOracle:      "Air Nomad Student",
		avB27DukeOracle:         "The Duke, Rebel Sentry",
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

// avB27Count counts the permanents named `name` that controller controls.
func avB27Count(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// avB27Resolve passes priority until the stack is empty or a prompt
// stops it.
func avB27Resolve(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			return
		}
	}
}

// --- Swampbenders ----------------------------------------------------------------

func TestAvatarB27SwampbendersMakesYourLandsSwampsAndCountsEverySwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mountain := effectstest.PushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	bender := effectstest.B12Push(g, me.ID, "Swampbenders", "Creature — Human Druid Ally", avB27SwampbendersOracle, 0, 0)

	f := effectstest.AvB1Card(t, g, forest)
	if !f.HasSubtype("Swamp") || !f.HasSubtype("Forest") {
		t.Errorf("your Forest's subtypes %v, want Forest and Swamp", f.Effective().Subtypes)
	}
	if effectstest.AvB1Card(t, g, mountain).HasSubtype("Swamp") {
		t.Error("an opponent's Mountain became a Swamp: the grant is lands YOU control")
	}
	if c := effectstest.AvB1Card(t, g, bender); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Swampbenders is %d/%d, want 3/3 (two of your lands and their Swamp)", c.CurrentPower(), c.CurrentToughness())
	}
	tapped := false
	for i := 0; i < 3 && !tapped; i++ {
		g.WithWriteLock(func() { me.ManaPool = nil; _ = g.UntapTargetForEffect(forest) })
		tapped = g.ActivateManaAbility(me.ID, forest, i, game.ManaAbilityParams{}) == nil && effectstest.MkColors(me)["B"] == 1
	}
	if !tapped {
		t.Fatal("the Forest has no {T}: Add {B}")
	}
	effectstest.MkWantPool(t, me, map[string]int{"B": 1})
}

// --- Sokka's Charge ----------------------------------------------------------------

func TestAvatarB27SokkasChargeArmsYourAlliesOnYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sokka's Charge", "Enchantment", avB27ChargeOracle, 0, 0)
	ally := effectstest.B12Creature(g, me.ID, "Ally Scout", "Creature — Human Ally", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for _, kw := range []string{"double strike", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, ally, kw) {
			t.Errorf("your Ally lacks %s on your turn", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, bear, kw) || effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("a non-Ally or an opponent's Ally has %s", kw)
		}
	}
	effectstest.EndTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("the turn did not pass")
	}
	if effectstest.HasEffectiveKeyword(t, g, ally, "double strike") || effectstest.HasEffectiveKeyword(t, g, ally, "lifelink") {
		t.Error("your Ally kept the grant on an opponent's turn")
	}
}

// --- Mai and Zuko ----------------------------------------------------------------------

func TestAvatarB27MaiAndZukoFlashesAlliesAndArtifactsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mai and Zuko", "Legendary Creature — Human Noble Ally", avB27MaiZukoOracle, 3, 5)
	effectstest.MkAdvanceTo(t, g, game.StepBeginCombat)
	rock := effectstest.AvB24Hand(g, me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	ally := effectstest.AvB24Hand(g, me, game.Card{Name: "Ally Scout", TypeLine: "Creature — Human Ally", ManaCost: "{1}", Power: 1, Toughness: 1})
	bear := effectstest.AvB24Hand(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2})
	for _, id := range []uuid.UUID{rock, ally} {
		effectstest.FloatMana(t, g, me, "{C}")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("an artifact or Ally spell was refused at instant speed: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Ally creature spell was cast as though it had flash")
	}
}

func TestAvatarB27MaiAndZukoFirebendsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := effectstest.B12Push(g, me.ID, "Mai and Zuko", "Legendary Creature — Human Noble Ally", avB27MaiZukoOracle, 3, 5)
	effectstest.DeclareAttack(t, g, opp.ID, duo)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
}

// --- Aang and Katara ----------------------------------------------------------------------

func TestAvatarB27AangAndKataraCountTappedArtifactsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, rock, golem, bear, land, theirs)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Aang and Katara", TypeLine: "Legendary Creature — Human Avatar Ally",
		OracleID: avB27AangKataraOracle, Power: 5, Toughness: 5}, game.CastSpellParams{})
	avB27Resolve(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB27Count(g, me.ID, "Ally"); got != 3 {
		t.Errorf("%d Ally tokens on entering, want 3 (rock, golem once, bear)", got)
	}
}

func TestAvatarB27AangAndKataraCountThemselvesWhenTheyAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := effectstest.B12Push(g, me.ID, "Aang and Katara", "Legendary Creature — Human Avatar Ally", avB27AangKataraOracle, 5, 5)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.TapForTest(t, g, rock)
	effectstest.DeclareAttack(t, g, opp.ID, duo)
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB27Count(g, me.ID, "Ally"); got != 2 {
		t.Errorf("%d Ally tokens on attacking, want 2 (the tapped attacker and the rock)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ally" && (c.Power != 1 || c.Toughness != 1 || !c.HasColor("W")) {
			t.Errorf("token %+v, want a 1/1 white Ally", c)
		}
	}
}

// --- Toph, Greatest Earthbender ------------------------------------------------------------------

func TestAvatarB27TophEarthbendsForTheManaSpentAndGivesDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.AvB14Forest(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB24Hand(g, me, game.Card{Name: "Toph, Greatest Earthbender", TypeLine: "Legendary Creature — Human Warrior Ally",
		ManaCost: "{2}{R}{G}", OracleID: avB27TophOracle, Power: 3, Toughness: 3})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatMana(t, g, me, "{C}{C}{R}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Toph: %v", err)
	}
	avB27Resolve(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 4)
	if !effectstest.HasEffectiveKeyword(t, g, land, "double strike") {
		t.Error("the land creature lacks double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
		t.Error("Toph herself (not a land) has double strike")
	}
}

// --- Sokka and Suki --------------------------------------------------------------------------------

func TestAvatarB27SokkaAndSukiAttachEquipmentToAnEnteringAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sokka and Suki", "Legendary Creature — Human Warrior Ally", avB27SokkaSukiOracle, 3, 3)
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	ally := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Ally Scout", TypeLine: "Creature — Human Ally", Power: 1, Toughness: 1}, game.CastSpellParams{})
	avB27Resolve(t, g)
	effectstest.PickCard(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, sword); !ok || c.AttachedTo.ID != ally {
		t.Fatalf("the Equipment is attached to %v, want the Ally %v", c.AttachedTo.ID, ally)
	}
	bear := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	avB27Resolve(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a non-Ally entering triggered the attach")
	}
	if c, _ := effectstest.BattlefieldCard(g, sword); c.AttachedTo.ID == bear {
		t.Error("the Equipment moved to a non-Ally")
	}
}

func TestAvatarB27SokkaAndSukiMakeAnAllyForEachEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sokka and Suki", "Legendary Creature — Human Warrior Ally", avB27SokkaSukiOracle, 3, 3)
	effectstest.AvB1Cast(t, g, me, game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"}, game.CastSpellParams{})
	avB27Resolve(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := avB27Count(g, me.ID, "Ally"); got != 1 {
		t.Fatalf("%d Ally tokens, want 1", got)
	}
	// The token is an Ally entering: its own attach trigger asks, and
	// declining ("up to one") leaves the Equipment where it is.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the attach: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if avB27Count(g, me.ID, "Ally") != 1 {
		t.Error("declining the attach made more tokens")
	}
}

// --- Momo's Heist ------------------------------------------------------------------------------------

func TestAvatarB27MomosHeistStealsAnArtifactUntilTheEndStepSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, rock)
	if err := effectstest.CastCatalogSpellErr(t, g, "Momo's Heist", "Sorcery", avB27HeistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("Momo's Heist targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Momo's Heist", "Sorcery", avB27HeistOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, rock)
	if !ok || c.Controller != me.ID || c.Tapped {
		t.Fatalf("rock controller %v tapped %v, want yours and untapped", c.Controller, c.Tapped)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rock, "haste") {
		t.Error("the stolen artifact lacks haste")
	}
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not sacrificed at the end step")
	}
}

// --- Uncle's Musings ------------------------------------------------------------------------------------

func TestAvatarB27UnclesMusingsReturnsOnePermanentPerColourAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "Old Rock", "Artifact")
	c := effectstest.PushGraveyardCardTyped(me, "Old Charm", "Enchantment")
	bolt := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	id := effectstest.AvB24Hand(g, me, game.Card{Name: "Uncle's Musings", TypeLine: "Sorcery", ManaCost: "{2}{G}{G}", OracleID: avB27MusingsOracle})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatMana(t, g, me, "{C}{U}{G}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if pick.ChooseMax != 2 {
		t.Errorf("up to %d cards, want 2 (green and blue were spent)", pick.ChooseMax)
	}
	for _, offered := range pick.ChooseCards {
		if offered == bolt {
			t.Error("an instant card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b, c}); err == nil {
		t.Error("three cards were accepted for an X of 2")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || !me.Graveyard.Contains(c) {
		t.Error("the two chosen cards did not return, or the third moved")
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("Uncle's Musings was not exiled")
	}
}

// --- Air Nomad Student ------------------------------------------------------------------------------------

func TestAvatarB27AirNomadStudentGrowsOnlyWhenItDidntAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	student := effectstest.B12Push(g, me.ID, "Air Nomad Student", "Creature — Human Monk", avB27StudentOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, student, "flying") {
		t.Error("printed flying is missing")
	}
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, student, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after a turn without attacking, want 1", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	me, opp := g2.Seats[0], g2.Seats[1]
	attacker := effectstest.B12Push(g2, me.ID, "Air Nomad Student", "Creature — Human Monk", avB27StudentOracle, 2, 2)
	effectstest.DeclareAttack(t, g2, opp.ID, attacker)
	effectstest.MkAdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.CountersOn(g2, attacker, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after attacking, want 0", n)
	}
}

// --- The Duke, Rebel Sentry ------------------------------------------------------------------------------------

func TestAvatarB27TheDukeMovesHisCounterAndGrantsHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duke := effectstest.AvB1Cast(t, g, me, game.Card{Name: "The Duke, Rebel Sentry", TypeLine: "Legendary Creature — Human Rebel Ally",
		OracleID: avB27DukeOracle, Power: 0, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, duke, game.CounterPlusOne); n != 1 {
		t.Fatalf("The Duke entered with %d counters, want 1", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == duke {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	for _, bad := range []uuid.UUID{theirs, duke} {
		if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bad), CounterKind: game.CounterPlusOne, Strict: true}); err == nil {
			t.Fatalf("activated targeting %v: not another creature you control", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear), CounterKind: game.CounterPlusOne, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, duke, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("Duke %d counters, bear %d; want 0 and 1", effectstest.CountersOn(g, duke, game.CounterPlusOne), effectstest.CountersOn(g, bear, game.CounterPlusOne))
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the bear lacks hexproof")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(duke) })
	if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear), CounterKind: game.CounterPlusOne, Strict: true}); err == nil {
		t.Error("activated with no counter to remove")
	}
}
