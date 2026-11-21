package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b74_cards_test.go — S82 GAP batch 74 (#5948), the Ravnica
// Remastered red, blue-red and red-white creatures that were missing. Each
// test plays the catalog entry through the engine.

const (
	gapBurningProphetOracle = "1e462ce2-15d3-473c-8f48-9c9363fc65eb"
	gapIlhargOracle         = "8d571129-9030-47e0-9624-a49fb63e5a1b"
	gapPitDragonOracle      = "63dd611e-1c36-4e5d-ba45-87c7f629cd54"
	gapMaakaOracle          = "185db9c7-3936-473c-9220-ae255527aaf9"
	gapDodgerOracle         = "fde3c8df-cc57-4e64-b78b-301e103d8487"
	gapCracklingDrakeOracle = "c8b4eba6-88ce-42d7-a235-5a4765a4a981"
	gapIzzetGuildmageOracle = "34778e85-cd02-40c8-855c-4f4529aa458b"
	gapPetrahydroxOracle    = "545c516f-85a4-4a59-8cfa-03da27121a57"
	gapAureliaOracle        = "2e1e0385-1537-4cb6-9d24-e4593d9a8f4a"
)

func TestGapB74CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gapBurningProphetOracle: "Burning Prophet", gapIlhargOracle: "Ilharg, the Raze-Boar",
		gapPitDragonOracle: "Rakdos Pit Dragon", gapMaakaOracle: "Rubblebelt Maaka",
		gapDodgerOracle: "Tin Street Dodger", gapCracklingDrakeOracle: "Crackling Drake",
		gapIzzetGuildmageOracle: "Izzet Guildmage", gapPetrahydroxOracle: "Petrahydrox",
		gapAureliaOracle: "Aurelia, Exemplar of Justice",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

// gapPower is a permanent's power with its counters and until-end-of-turn
// effects counted.
func gapPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var p int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s is not on the battlefield", id)
		}
		p = c.CurrentPower()
	})
	return p
}

// r35GapCastCreature casts a creature card with the given base power and
// toughness for free from the active seat's hand in a main phase.
func r35GapCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Burning Prophet ---------------------------------------------------------

func TestGapB74BurningProphetPumpsAndScriesOnNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	prophet := effectstest.B12Push(g, me.ID, "Burning Prophet", "Creature — Human Wizard", gapBurningProphetOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil || gapPower(t, g, prophet) != 1 {
		t.Fatal("a creature spell triggered the Prophet")
	}
	effectstest.CastCatalogSpell(t, g, "A Charm", "Instant", "", nil)
	effectstest.AnswerScryIfAsked(t, g, me)
	if p := gapPower(t, g, prophet); p != 2 {
		t.Errorf("power %d after a noncreature spell, want 2", p)
	}
	effectstest.CastCatalogSpell(t, g, "A Second Charm", "Instant", "", nil)
	effectstest.AnswerScryIfAsked(t, g, me)
	if p := gapPower(t, g, prophet); p != 3 {
		t.Errorf("power %d after two noncreature spells, want 3", p)
	}
}

// --- Rubblebelt Maaka ---------------------------------------------------------

func TestGapB74RubblebeltMaakaBloodrushesFromTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	maaka := effectstest.HandCardForTest(me, "Rubblebelt Maaka", "Creature — Cat", gapMaakaOracle)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AddPool(t, g, me.ID, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, maaka, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(maaka)}); err == nil {
		t.Fatal("a card in the hand is not an attacking creature")
	}
	effectstest.B16Activate(t, g, me.ID, maaka, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)})
	effectstest.PassPriorityAroundTable(t, g)
	if p := gapPower(t, g, atk); p != 5 {
		t.Errorf("the attacker has power %d, want 5", p)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, atk); p != 5 || tough != 5 {
		t.Errorf("the attacker is %d/%d, want 5/5", p, tough)
	}
	if z, _ := effectstest.CardWhere(g, me, maaka); z != "graveyard" {
		t.Errorf("the Maaka is in %s, want the graveyard (discarded to pay)", z)
	}
}

// --- Tin Street Dodger --------------------------------------------------------

func TestGapB74TinStreetDodgerCantBeBlockedExceptByDefendersWhenPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dodger := effectstest.B12Push(g, me.ID, "Tin Street Dodger", "Creature — Goblin Rogue", gapDodgerOracle, 1, 1)
	bear := effectstest.BrCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.BrCreature(g, opp.ID, "Wall", "Creature — Wall", 0, 4, "defender")
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{R}")
	effectstest.P7Activate(t, g, me, dodger, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.BrAttack(t, g, dodger)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, dodger), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(wall, dodger); err != nil {
		t.Fatalf("a creature with defender blocks: %v", err)
	}
}

func TestGapB74TinStreetDodgerCanBeBlockedUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dodger := effectstest.B12Push(g, me.ID, "Tin Street Dodger", "Creature — Goblin Rogue", gapDodgerOracle, 1, 1)
	bear := effectstest.BrCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dodger, "haste") {
		t.Error("no haste")
	}
	effectstest.BrAttack(t, g, dodger)
	if err := g.DeclareBlocker(bear, dodger); err != nil {
		t.Fatalf("an unpaid Dodger can be blocked: %v", err)
	}
}

// --- Rakdos Pit Dragon --------------------------------------------------------

func TestGapB74RakdosPitDragonHasDoubleStrikeOnlyWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Rakdos Pit Dragon", "Creature — Dragon", gapPitDragonOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	g.BumpLayerVersionForTest()
	effectstest.CheckState(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "double strike") {
		t.Error("no double strike with an empty hand")
	}
	effectstest.HandCardForTest(me, "A Card", "Instant", "")
	g.BumpLayerVersionForTest()
	effectstest.CheckState(t, g)
	if effectstest.HasEffectiveKeyword(t, g, dragon, "double strike") {
		t.Error("double strike with a card in hand")
	}
	// A card in an opponent's hand does not matter.
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	effectstest.HandCardForTest(g.Seats[1], "Theirs", "Instant", "")
	g.BumpLayerVersionForTest()
	effectstest.CheckState(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "double strike") {
		t.Error("an opponent's hand stopped hellbent")
	}
}

func TestGapB74RakdosPitDragonFliesAndFirebreathsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := effectstest.B12Push(g, me.ID, "Rakdos Pit Dragon", "Creature — Dragon", gapPitDragonOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Fatal("flying before paying")
	}
	effectstest.AddPool(t, g, me.ID, "{R}{R}")
	effectstest.P7Activate(t, g, me, dragon, 0, game.ActivateAbilityParams{Strict: true})
	if !effectstest.HasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("no flying after {R}{R}")
	}
	effectstest.AddPool(t, g, me.ID, "{R}{R}")
	effectstest.P7Activate(t, g, me, dragon, 1, game.ActivateAbilityParams{Strict: true})
	effectstest.P7Activate(t, g, me, dragon, 1, game.ActivateAbilityParams{Strict: true})
	if p, _ := effectstest.EffectivePTOf(t, g, dragon); p != 5 {
		t.Errorf("power %d after two {R}, want 5", p)
	}
}

// --- Crackling Drake -------------------------------------------------------------

func TestGapB74CracklingDrakePowerCountsInstantsAndSorceriesYouOwnInExileAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drake := effectstest.B12Push(g, me.ID, "Crackling Drake", "Creature — Drake", gapCracklingDrakeOracle, 0, 4)
	effectstest.AdvanceToMain(t, g)
	add := func(zone *game.Zone, owner uuid.UUID, typeLine string) {
		zone.PushTop(game.Card{InstanceID: uuid.New(), Name: "C", TypeLine: typeLine, Owner: owner, Controller: owner})
	}
	add(me.Graveyard, me.ID, "Instant")
	add(me.Graveyard, me.ID, "Sorcery")
	add(me.Graveyard, me.ID, "Creature — Bear")
	add(opp.Graveyard, opp.ID, "Instant")
	add(g.Exile, me.ID, "Sorcery")
	add(g.Exile, me.ID, "Instant")
	add(g.Exile, me.ID, "Land")
	add(g.Exile, opp.ID, "Instant")
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if p, tough := effectstest.EffectivePTOf(t, g, drake); p != 4 || tough != 4 {
		t.Errorf("the Drake is %d/%d, want 4/4 (two in the graveyard, two in exile)", p, tough)
	}
}

func TestGapB74CracklingDrakeDrawsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	drake := r35GapCastCreature(t, g, "Crackling Drake", "Creature — Drake", gapCracklingDrakeOracle, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(drake) {
		t.Fatal("the Drake did not enter")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d (drew one)", me.Hand.Size(), before+1)
	}
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("no flying")
	}
}

// --- Izzet Guildmage ---------------------------------------------------------------

func TestGapB74IzzetGuildmageCopiesAnInstantOrSorceryOfManaValueTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.B12Push(g, me.ID, "Izzet Guildmage", "Creature — Human Wizard", gapIzzetGuildmageOracle, 2, 2)
	bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	life := opp.Life
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bolt)}); err == nil {
		t.Fatal("the sorcery ability targeted an instant")
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("copy the instant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	r35AnswerCopyTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("the opponent lost %d, want 6 (the spell and its copy)", got)
	}
}

func TestGapB74IzzetGuildmageCantCopyAnOpponentsSpellOrAnExpensiveOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.B12Push(g, me.ID, "Izzet Guildmage", "Creature — Human Wizard", gapIzzetGuildmageOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	pricey := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pricey, Name: "Pricey", TypeLine: "Instant", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	if err := g.CastSpell(me.ID, pricey, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(pricey)}); err == nil {
		t.Error("copied a spell with mana value 3")
	}
	_ = opp
}

// --- Petrahydrox -----------------------------------------------------------------------

func TestGapB74PetrahydroxReturnsToHandWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hydrox := effectstest.B12Push(g, me.ID, "Petrahydrox", "Creature — Weird", gapPetrahydroxOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardRefs(hydrox))
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, hydrox); z != "hand" {
		t.Errorf("the Petrahydrox is in %s, want the hand", z)
	}
	if z, _ := effectstest.CardWhere(g, me, bolt); z != "graveyard" {
		t.Errorf("the Bolt is in %s, want the graveyard (no legal target)", z)
	}
}

// --- Aurelia, Exemplar of Justice -------------------------------------------------------

func gapAureliaTable(t *testing.T, colors ...string) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aurelia := effectstest.B12Push(g, me.ID, "Aurelia, Exemplar of Justice", "Legendary Creature — Angel", gapAureliaOracle, 2, 5)
	target := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Target", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: colors})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	return g, aurelia, target
}

func TestGapB74AureliaGivesTrampleToARedCreatureAndVigilanceToAWhiteOne(t *testing.T) {
	for _, tc := range []struct {
		name           string
		colors         []string
		trample, vigil bool
		wantPower      int
	}{
		{"red", []string{"R"}, true, false, 4},
		{"white", []string{"W"}, false, true, 4},
		{"red and white", []string{"R", "W"}, true, true, 4},
		{"green", []string{"G"}, false, false, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, target := gapAureliaTable(t, tc.colors...)
			if got := effectstest.HasEffectiveKeyword(t, g, target, "trample"); got != tc.trample {
				t.Errorf("trample %v, want %v", got, tc.trample)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, target, "vigilance"); got != tc.vigil {
				t.Errorf("vigilance %v, want %v", got, tc.vigil)
			}
			if p, _ := effectstest.EffectivePTOf(t, g, target); p != tc.wantPower {
				t.Errorf("power %d, want %d", p, tc.wantPower)
			}
		})
	}
}

func TestGapB74AureliaCanTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aurelia := effectstest.B12Push(g, me.ID, "Aurelia, Exemplar of Justice", "Legendary Creature — Angel", gapAureliaOracle, 2, 5)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, aurelia)
		c.Colors = []string{"R", "W"}
	})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, aurelia)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, aurelia); p != 4 {
		t.Errorf("Aurelia has power %d, want 4", p)
	}
}

// --- Ilharg, the Raze-Boar ---------------------------------------------------------------

func gapIlhargAttacks(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	ilharg := effectstest.B12Push(g, me.ID, "Ilharg, the Raze-Boar", "Legendary Creature — Boar God", gapIlhargOracle, 6, 6)
	return g, me, opp, ilharg
}

func TestGapB74IlhargPutsACreatureCardAttackingAndItReturnsAtEndStep(t *testing.T) {
	g, me, opp, ilharg := gapIlhargAttacks(t)
	bear := effectstest.HandCardForTest(me, "Hill Giant", "Creature — Giant", "")
	effectstest.HandCardForTest(me, "A Charm", "Instant", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, ilharg)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no card offered")
	}
	if slices.Contains(choice.ChooseCards, ilharg) || len(choice.ChooseCards) != 1 || choice.ChooseCards[0] != bear {
		t.Fatalf("offered %v; want only the creature card in hand", choice.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	z, c := effectstest.CardWhere(g, me, bear)
	if z != "battlefield" || !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Fatalf("the put creature is in %s tapped=%v attacking=%v, want tapped and attacking", z, c.Tapped, c.AttackingTarget)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, bear); z != "hand" {
		t.Errorf("at the end step the creature is in %s, want the hand", z)
	}
}

func TestGapB74IlhargMayPutNothing(t *testing.T) {
	g, me, opp, ilharg := gapIlhargAttacks(t)
	bear := effectstest.HandCardForTest(me, "Hill Giant", "Creature — Giant", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, ilharg)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, bear); z != "hand" {
		t.Errorf("declined, yet the creature is in %s", z)
	}
}

func TestGapB74IlhargGoesThirdFromTheTopWhenItDiesOrIsExiled(t *testing.T) {
	for _, tc := range []struct {
		name string
		move func(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID)
	}{
		{"dies", func(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
			effectstest.KillCreature(t, g, me.ID, id)
		}},
		{"exiled", func(t *testing.T, g *game.Game, _ *game.Player, id uuid.UUID) { effectstest.ExileCreature(t, g, id) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _, ilharg := gapIlhargAttacks(t)
			effectstest.AdvanceToMain(t, g)
			tc.move(t, g, me, ilharg)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
			if got := slices.IndexFunc(me.Library.Cards, func(c game.Card) bool { return c.InstanceID == ilharg }); got != len(me.Library.Cards)-3 {
				t.Errorf("Ilharg is at index %d of %d cards, want third from the top (index %d)", got, len(me.Library.Cards), len(me.Library.Cards)-3)
			}
		})
	}
}

func TestGapB74IlhargMayStayInTheGraveyard(t *testing.T) {
	g, me, _, ilharg := gapIlhargAttacks(t)
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, ilharg)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, ilharg); z != "graveyard" {
		t.Errorf("Ilharg is in %s, want the graveyard", z)
	}
}
