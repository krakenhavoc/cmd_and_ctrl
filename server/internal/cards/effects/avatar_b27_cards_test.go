package effects

import (
	"testing"

	"github.com/google/uuid"

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	pushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	mountain := pushLandFor(g, opp.ID, "Mountain", "Basic Land — Mountain")
	bender := b12Push(g, me.ID, "Swampbenders", "Creature — Human Druid Ally", avB27SwampbendersOracle, 0, 0)

	f := avB1Card(t, g, forest)
	if !f.HasSubtype("Swamp") || !f.HasSubtype("Forest") {
		t.Errorf("your Forest's subtypes %v, want Forest and Swamp", f.Effective().Subtypes)
	}
	if avB1Card(t, g, mountain).HasSubtype("Swamp") {
		t.Error("an opponent's Mountain became a Swamp: the grant is lands YOU control")
	}
	if c := avB1Card(t, g, bender); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Swampbenders is %d/%d, want 3/3 (two of your lands and their Swamp)", c.CurrentPower(), c.CurrentToughness())
	}
	tapped := false
	for i := 0; i < 3 && !tapped; i++ {
		g.WithWriteLock(func() { me.ManaPool = nil; _ = g.UntapTargetForEffect(forest) })
		tapped = g.ActivateManaAbility(me.ID, forest, i, game.ManaAbilityParams{}) == nil && mkColors(me)["B"] == 1
	}
	if !tapped {
		t.Fatal("the Forest has no {T}: Add {B}")
	}
	mkWantPool(t, me, map[string]int{"B": 1})
}

// --- Sokka's Charge ----------------------------------------------------------------

func TestAvatarB27SokkasChargeArmsYourAlliesOnYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sokka's Charge", "Enchantment", avB27ChargeOracle, 0, 0)
	ally := b12Creature(g, me.ID, "Ally Scout", "Creature — Human Ally", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 2, 2)
	advanceToMain(t, g)
	for _, kw := range []string{"double strike", "lifelink"} {
		if !hasEffectiveKeyword(t, g, ally, kw) {
			t.Errorf("your Ally lacks %s on your turn", kw)
		}
		if hasEffectiveKeyword(t, g, bear, kw) || hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("a non-Ally or an opponent's Ally has %s", kw)
		}
	}
	endTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("the turn did not pass")
	}
	if hasEffectiveKeyword(t, g, ally, "double strike") || hasEffectiveKeyword(t, g, ally, "lifelink") {
		t.Error("your Ally kept the grant on an opponent's turn")
	}
}

// --- Mai and Zuko ----------------------------------------------------------------------

func TestAvatarB27MaiAndZukoFlashesAlliesAndArtifactsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Mai and Zuko", "Legendary Creature — Human Noble Ally", avB27MaiZukoOracle, 3, 5)
	mkAdvanceTo(t, g, game.StepBeginCombat)
	rock := avB24Hand(g, me, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	ally := avB24Hand(g, me, game.Card{Name: "Ally Scout", TypeLine: "Creature — Human Ally", ManaCost: "{1}", Power: 1, Toughness: 1})
	bear := avB24Hand(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 2, Toughness: 2})
	for _, id := range []uuid.UUID{rock, ally} {
		floatMana(t, g, me, "{C}")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("an artifact or Ally spell was refused at instant speed: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a non-Ally creature spell was cast as though it had flash")
	}
}

func TestAvatarB27MaiAndZukoFirebendsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b12Push(g, me.ID, "Mai and Zuko", "Legendary Creature — Human Noble Ally", avB27MaiZukoOracle, 3, 5)
	declareAttack(t, g, opp.ID, duo)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
}

// --- Aang and Katara ----------------------------------------------------------------------

func TestAvatarB27AangAndKataraCountTappedArtifactsAndCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, rock, golem, bear, land, theirs)
	avB1Cast(t, g, me, game.Card{Name: "Aang and Katara", TypeLine: "Legendary Creature — Human Avatar Ally",
		OracleID: avB27AangKataraOracle, Power: 5, Toughness: 5}, game.CastSpellParams{})
	avB27Resolve(t, g)
	passPriorityAroundTable(t, g)
	if got := avB27Count(g, me.ID, "Ally"); got != 3 {
		t.Errorf("%d Ally tokens on entering, want 3 (rock, golem once, bear)", got)
	}
}

func TestAvatarB27AangAndKataraCountThemselvesWhenTheyAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b12Push(g, me.ID, "Aang and Katara", "Legendary Creature — Human Avatar Ally", avB27AangKataraOracle, 5, 5)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	tapForTest(t, g, rock)
	declareAttack(t, g, opp.ID, duo)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB14Forest(g, me.ID)
	advanceToMain(t, g)
	id := avB24Hand(g, me, game.Card{Name: "Toph, Greatest Earthbender", TypeLine: "Legendary Creature — Human Warrior Ally",
		ManaCost: "{2}{R}{G}", OracleID: avB27TophOracle, Power: 3, Toughness: 3})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatMana(t, g, me, "{C}{C}{R}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Toph: %v", err)
	}
	avB27Resolve(t, g)
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 4)
	if !hasEffectiveKeyword(t, g, land, "double strike") {
		t.Error("the land creature lacks double strike")
	}
	if hasEffectiveKeyword(t, g, id, "double strike") {
		t.Error("Toph herself (not a land) has double strike")
	}
}

// --- Sokka and Suki --------------------------------------------------------------------------------

func TestAvatarB27SokkaAndSukiAttachEquipmentToAnEnteringAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sokka and Suki", "Legendary Creature — Human Warrior Ally", avB27SokkaSukiOracle, 3, 3)
	sword := b12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	ally := avB1Cast(t, g, me, game.Card{Name: "Ally Scout", TypeLine: "Creature — Human Ally", Power: 1, Toughness: 1}, game.CastSpellParams{})
	avB27Resolve(t, g)
	pickCard(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, sword); !ok || c.AttachedTo.ID != ally {
		t.Fatalf("the Equipment is attached to %v, want the Ally %v", c.AttachedTo.ID, ally)
	}
	bear := avB1Cast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, game.CastSpellParams{})
	avB27Resolve(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a non-Ally entering triggered the attach")
	}
	if c, _ := battlefieldCard(g, sword); c.AttachedTo.ID == bear {
		t.Error("the Equipment moved to a non-Ally")
	}
}

func TestAvatarB27SokkaAndSukiMakeAnAllyForEachEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sokka and Suki", "Legendary Creature — Human Warrior Ally", avB27SokkaSukiOracle, 3, 3)
	avB1Cast(t, g, me, game.Card{Name: "Sword", TypeLine: "Artifact — Equipment"}, game.CastSpellParams{})
	avB27Resolve(t, g)
	passPriorityAroundTable(t, g)
	if got := avB27Count(g, me.ID, "Ally"); got != 1 {
		t.Fatalf("%d Ally tokens, want 1", got)
	}
	// The token is an Ally entering: its own attach trigger asks, and
	// declining ("up to one") leaves the Equipment where it is.
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the attach: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if avB27Count(g, me.ID, "Ally") != 1 {
		t.Error("declining the attach made more tokens")
	}
}

// --- Momo's Heist ------------------------------------------------------------------------------------

func TestAvatarB27MomosHeistStealsAnArtifactUntilTheEndStepSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tapForTest(t, g, rock)
	if err := castCatalogSpellErr(t, g, "Momo's Heist", "Sorcery", avB27HeistOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("Momo's Heist targeted a creature")
	}
	castCatalogSpell(t, g, "Momo's Heist", "Sorcery", avB27HeistOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, rock)
	if !ok || c.Controller != me.ID || c.Tapped {
		t.Fatalf("rock controller %v tapped %v, want yours and untapped", c.Controller, c.Tapped)
	}
	if !hasEffectiveKeyword(t, g, rock, "haste") {
		t.Error("the stolen artifact lacks haste")
	}
	mkAdvanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not sacrificed at the end step")
	}
}

// --- Uncle's Musings ------------------------------------------------------------------------------------

func TestAvatarB27UnclesMusingsReturnsOnePermanentPerColourAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	a := pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	b := pushGraveyardCardTyped(me, "Old Rock", "Artifact")
	c := pushGraveyardCardTyped(me, "Old Charm", "Enchantment")
	bolt := pushGraveyardCardTyped(me, "Old Bolt", "Instant")
	id := avB24Hand(g, me, game.Card{Name: "Uncle's Musings", TypeLine: "Sorcery", ManaCost: "{2}{G}{G}", OracleID: avB27MusingsOracle})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatMana(t, g, me, "{C}{U}{G}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
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
	answerChooseCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || !me.Graveyard.Contains(c) {
		t.Error("the two chosen cards did not return, or the third moved")
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("Uncle's Musings was not exiled")
	}
}

// --- Air Nomad Student ------------------------------------------------------------------------------------

func TestAvatarB27AirNomadStudentGrowsOnlyWhenItDidntAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	student := b12Push(g, me.ID, "Air Nomad Student", "Creature — Human Monk", avB27StudentOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, student, "flying") {
		t.Error("printed flying is missing")
	}
	mkAdvanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, student, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after a turn without attacking, want 1", n)
	}

	g2 := newCatalogGame(t)
	me, opp = g2.Seats[0], g2.Seats[1]
	attacker := b12Push(g2, me.ID, "Air Nomad Student", "Creature — Human Monk", avB27StudentOracle, 2, 2)
	declareAttack(t, g2, opp.ID, attacker)
	mkAdvanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if n := countersOn(g2, attacker, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after attacking, want 0", n)
	}
}

// --- The Duke, Rebel Sentry ------------------------------------------------------------------------------------

func TestAvatarB27TheDukeMovesHisCounterAndGrantsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duke := avB1Cast(t, g, me, game.Card{Name: "The Duke, Rebel Sentry", TypeLine: "Legendary Creature — Human Rebel Ally",
		OracleID: avB27DukeOracle, Power: 0, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := countersOn(g, duke, game.CounterPlusOne); n != 1 {
		t.Fatalf("The Duke entered with %d counters, want 1", n)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == duke {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	for _, bad := range []uuid.UUID{theirs, duke} {
		if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bad), CounterKind: game.CounterPlusOne, Strict: true}); err == nil {
			t.Fatalf("activated targeting %v: not another creature you control", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), CounterKind: game.CounterPlusOne, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, duke, game.CounterPlusOne) != 0 || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Errorf("Duke %d counters, bear %d; want 0 and 1", countersOn(g, duke, game.CounterPlusOne), countersOn(g, bear, game.CounterPlusOne))
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the bear lacks hexproof")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(duke) })
	if err := g.ActivateCatalogAbility(me.ID, duke, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear), CounterKind: game.CounterPlusOne, Strict: true}); err == nil {
		t.Error("activated with no counter to remove")
	}
}
