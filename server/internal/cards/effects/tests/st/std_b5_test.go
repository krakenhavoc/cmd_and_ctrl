package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const nutrientBlockOracle = "04a0382f-7e4e-4450-ae70-18999379d875"

func TestNutrientBlockGainsLifeThenDrawsWhenItDies(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, nutrientBlockOracle, "Nutrient Block")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	block := effectstest.PushCatalogPermanent(g, me.ID, "Nutrient Block", "Artifact", nutrientBlockOracle, false)
	startLife, hand := effectstest.LifeOf(g, me.ID), effectstest.HandSizeOf(g, me)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, block, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != startLife+3 {
		t.Errorf("life %d → %d, want +3", startLife, got)
	}
	if effectstest.BattlefieldHas(g, block) {
		t.Errorf("Nutrient Block should have been sacrificed")
	}
	if got := effectstest.HandSizeOf(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn from the graveyard trigger", hand, got)
	}
}

var _ = uuid.Nil

const commandBridgeOracle = "87c8e1ed-258a-4a89-bcc6-211405e49692"

// Command Bridge asks for any untapped permanent (a creature will do),
// and sacrifices itself when you decline.
func TestCommandBridgeStaysWhenYouTapAnyPermanent(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, commandBridgeOracle, "Command Bridge")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Permanent(g, me.ID, "Bear", "Creature — Bear")
	land := effectstest.PlayLandFromHand(t, g, "Command Bridge", commandBridgeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
		t.Fatal("the Bridge did not enter tapped")
	}
	choice := effectstest.PayUnlessFor(g, me.ID)
	if choice == nil || choice.PayAction() == nil || choice.PayAction().Kind != game.PayActionTap {
		t.Fatalf("no tap payment asked: %+v", choice)
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, choice.PayAction()) })
	if len(opts) != 1 || opts[0] != bear {
		t.Fatalf("the payment offers %v, want only the untapped Bear (the tapped Bridge is not an option)", opts)
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("tap the Bear: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the Bridge was sacrificed although the payment was made")
	}
}

func TestCommandBridgeIsSacrificedWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.PlayLandFromHand(t, g, "Command Bridge", commandBridgeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.PayUnlessFor(g, me.ID)
	if choice == nil {
		t.Fatal("no payment asked")
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the Bridge was not sacrificed")
	}
}

// castStrict puts a card in the active seat's hand and casts it with the
// payment enforced.
func castStrict(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = active.ID, active.ID
	active.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	params.Strict = true
	return c.InstanceID, g.CastSpell(active.ID, c.InstanceID, params)
}

// --- Thrumming Hivepool ------------------------------------------------------

const hivepoolOracle = "04ea21b3-084b-4746-9b49-99b6ef4f0885"

func TestThrummingHivepoolGivesYourSliversDoubleStrikeAndHaste(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, hivepoolOracle, "Thrumming Hivepool")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thrumming Hivepool", "Artifact", hivepoolOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Sliver", "Creature — Sliver", "", false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Sliver", "Creature — Sliver", "", false)
	for _, kw := range []string{"double strike", "haste"} {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, mine), kw) {
			t.Errorf("a Sliver you control lacks %q", kw)
		}
		if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), kw) {
			t.Errorf("a Bear has %q", kw)
		}
		if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, theirs), kw) {
			t.Errorf("an opponent's Sliver has %q", kw)
		}
	}
}

func TestThrummingHivepoolMakesTwoSliversEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thrumming Hivepool", "Artifact", hivepoolOracle, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Sliver")
	if len(tokens) != 2 {
		t.Fatalf("%d Sliver tokens, want 2", len(tokens))
	}
}

func TestThrummingHivepoolCostsOneLessPerSliver(t *testing.T) {
	for _, c := range []struct {
		pool int
		ok   bool
	}{{3, false}, {4, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Sliver", "Creature — Sliver", "", false)
		effectstest.PushCatalogPermanent(g, me.ID, "Sliver", "Creature — Sliver", "", false)
		effectstest.AdvanceToMain(t, g)
		effectstest.FillPool(me, c.pool)
		card := game.Card{Name: "Thrumming Hivepool", TypeLine: "Artifact", OracleID: hivepoolOracle, ManaCost: "{6}"}
		_, err := castStrict(t, g, card, game.CastSpellParams{})
		if (err == nil) != c.ok {
			t.Errorf("pool %d with two Slivers (cost {4}): %v, want ok %v", c.pool, err, c.ok)
		}
	}
}

// --- Mightform Harmonizer ----------------------------------------------------

const mightformOracle = "8e025cda-71c5-4d06-8436-decd408667af"

func TestMightformHarmonizerDoublesPowerOnLandfall(t *testing.T) {
	spec := effectstest.RequireFullNoCaveats(t, mightformOracle, "Mightform Harmonizer")
	if len(spec.AlternativeCosts) != 1 {
		t.Fatalf("want the warp cost, got %d alternative costs", len(spec.AlternativeCosts))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	harm := effectstest.PushCatalogPermanent(g, me.ID, "Mightform Harmonizer", "Creature — Insect Druid", mightformOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 6 {
		t.Errorf("the Bear's power is %d, want 6 (3 doubled)", got)
	}
	if got := effectstest.EffectivePower(t, g, harm); got != 1 {
		t.Errorf("the Harmonizer's power is %d, want it untouched at the stand-in 1", got)
	}
}

// --- Rust Harvester ----------------------------------------------------------

const rustHarvesterOracle = "9ad956c8-71d3-436f-961e-e2dfc1d2ca70"

func TestRustHarvesterExilesAnArtifactAddsACounterThenDealsItsPower(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, rustHarvesterOracle, "Rust Harvester")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harv := effectstest.PushCatalogPermanent(g, me.ID, "Rust Harvester", "Artifact Creature — Robot", rustHarvesterOracle, false)
	rock := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.FillPool(me, 2)
	life := effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, harv, 0, game.ActivateAbilityParams{
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		ExileIDs: []uuid.UUID{bear},
		Strict:   true,
	}); err == nil {
		t.Fatal("exiling a non-artifact card should be refused")
	}
	if err := g.ActivateCatalogAbility(me.ID, harv, 0, game.ActivateAbilityParams{
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		ExileIDs: []uuid.UUID{rock},
		Strict:   true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(rock) {
		t.Error("the artifact card was not exiled")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d → %d, want 2 damage (1/1 with its new counter)", life, got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, harv); got != 2 {
		t.Errorf("power %d, want 2", got)
	}
}

// --- Dyadrine, Synthesis Amalgam ---------------------------------------------

const dyadrineOracle = "618a2272-62e0-4d0e-834d-7b81719eba12"

func TestDyadrineEntersWithACounterPerManaSpent(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, dyadrineOracle, "Dyadrine, Synthesis Amalgam")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPool(me, 2)
	id, err := castStrict(t, g, game.Card{Name: "Dyadrine, Synthesis Amalgam", TypeLine: "Legendary Artifact Creature — Construct",
		OracleID: dyadrineOracle, ManaCost: "{X}{G}{W}", Colors: []string{"G", "W"}}, game.CastSpellParams{XValue: 2})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCountOf(t, g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("Dyadrine entered with %d counters, want 4 (X=2 plus {G}{W})", got)
	}
}

func dyadrineAttack(t *testing.T, g *game.Game) (me *game.Player, a, b, dya uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	dya = effectstest.PushCatalogPermanent(g, me.ID, "Dyadrine, Synthesis Amalgam", "Legendary Artifact Creature — Construct", dyadrineOracle, false)
	counters := map[string]int{game.CounterPlusOne: 1}
	a = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: me.ID, Controller: me.ID})
	b = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: counters, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return me, a, b, dya
}

func TestDyadrineRemovesTwoCountersToDrawAndMakeARobot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hand := len(g.Seats[0].Hand.Cards)
	me, a, b, _ := dyadrineAttack(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCountOf(t, g, a, game.CounterPlusOne) != 0 || effectstest.CounterCountOf(t, g, b, game.CounterPlusOne) != 0 {
		t.Error("a counter was not removed from each of the two creatures")
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Robot"); got != 1 {
		t.Errorf("%d Robot tokens, want 1", got)
	}
}

func TestDyadrineDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hand := len(g.Seats[0].Hand.Cards)
	me, a, b, _ := dyadrineAttack(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCountOf(t, g, a, game.CounterPlusOne) != 1 || effectstest.CounterCountOf(t, g, b, game.CounterPlusOne) != 1 {
		t.Error("counters were removed although the player declined")
	}
	if len(me.Hand.Cards) != hand || effectstest.B30TokensNamed(g, me.ID, "Robot") != 0 {
		t.Error("declining still drew a card or made a Robot")
	}
}

// --- Survey Mechan -----------------------------------------------------------

const surveyMechanOracle = "53745903-de8f-42cf-a949-842e5831de30"

func TestSurveyMechanTenManaLessPerLandNameDealsDamageDrawsAndGains(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, surveyMechanOracle, "Survey Mechan")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mech := effectstest.PushCatalogPermanent(g, me.ID, "Survey Mechan", "Artifact Creature — Robot", surveyMechanOracle, false)
	for _, n := range []string{"Forest", "Island", "Swamp"} {
		effectstest.PushCatalogPermanent(g, me.ID, n, "Basic Land — "+n, "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: me.ID}}
	// Three differently named lands (the second Forest is not counted): {7}.
	effectstest.FillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err == nil {
		t.Fatal("six mana should not pay {10} less three")
	}
	effectstest.FillPool(me, 1)
	life, oppLife, hand := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Targets: target, Strict: true}); err != nil {
		t.Fatalf("seven mana should pay {10} less three: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-3 {
		t.Errorf("opponent life %d → %d, want 3 damage", oppLife, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+3 {
		t.Errorf("my life %d → %d, want +3", life, got)
	}
	if got := len(me.Hand.Cards); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards", hand, got)
	}
	if effectstest.BattlefieldHas(g, mech) {
		t.Error("the Mechan should have been sacrificed")
	}
}

// --- Xu-Ifit, Osteoharmonist -------------------------------------------------

const xuIfitOracle = "39e32fdf-93bf-4390-8275-0d7d10ad7360"

func TestXuIfitReturnsACreatureAsAnAbilitylessSkeletonAtSorcerySpeed(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, xuIfitOracle, "Xu-Ifit, Osteoharmonist")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	xu := effectstest.PushCatalogPermanent(g, me.ID, "Xu-Ifit, Osteoharmonist", "Legendary Creature — Human Wizard", xuIfitOracle, false)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID})
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}
	// Not at sorcery speed: in the beginning of combat.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, xu, 0, game.ActivateAbilityParams{Targets: tgt, Strict: true}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, xu, 0, game.ActivateAbilityParams{Targets: tgt, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, dead) {
		t.Fatal("the creature card did not return")
	}
	c := effectstest.BattlefieldCardCopy(t, g, dead)
	if !c.HasSubtype("Skeleton") || !c.HasSubtype("Bear") {
		t.Errorf("subtypes of the returned card: %q, want Skeleton in addition to Bear", c.TypeLine)
	}
	if ab := effectstest.EffectiveAbilities(t, g, dead); len(ab) != 0 {
		t.Errorf("the returned creature still has abilities %v", ab)
	}
}

// An "enters" trigger of the returned card never happens (the rulings).
func TestXuIfitStripsTheReturnedCardsEntersTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	xu := effectstest.PushCatalogPermanent(g, me.ID, "Xu-Ifit, Osteoharmonist", "Legendary Creature — Human Wizard", xuIfitOracle, false)
	// A catalog creature whose enters trigger would endure.
	spec := effects.Spec{Name: "Fortress Kin-Guard", OracleID: "fa3ed0f6-e44e-433d-80d3-824de5e6135a"}
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: spec.Name, TypeLine: "Creature — Dwarf Soldier", Power: 1, Toughness: 2,
		OracleID: spec.OracleID, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, xu, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, dead) {
		t.Fatal("the creature card did not return")
	}
	if got := effectstest.CounterCountOf(t, g, dead, game.CounterPlusOne); got != 0 {
		t.Errorf("Fortress Kin-Guard endured anyway (%d counters)", got)
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("Fortress Kin-Guard's enters trigger made %d Spirits", got)
	}
}

// --- Mm'menon, the Right Hand ------------------------------------------------

const mmenonOracle = "de58fa4b-eeeb-4dc8-aab9-6521b15678cc"

func TestMmenonGivesArtifactsRestrictedManaAndCastsArtifactsFromTheTop(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, mmenonOracle, "Mm'menon, the Right Hand")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mm'menon, the Right Hand", "Legendary Creature — Jellyfish Advisor", mmenonOracle, false)
	rock := effectstest.GaPut(g, me.ID, "Rock", "Artifact", "")
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	if n := effectstest.GrantedManaRowsOf(g, rock); n != 1 {
		t.Fatalf("the artifact has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 0 {
		t.Errorf("a non-artifact has %d granted rows, want 0", n)
	}
	effectstest.AdvanceToMain(t, g)
	// A noncreature, nonartifact top card is not castable.
	opt := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: opt, Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, opt, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Fatal("cast a non-artifact spell from the top of the library")
	}
	me.Library.Cards = me.Library.Cards[:len(me.Library.Cards)-1]
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	// The mana can't cast a spell from hand...
	if err := effectstest.TapGrantedMana(t, g, me.ID, rock, "U"); err != nil {
		t.Fatalf("tap the artifact: %v", err)
	}
	handRock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: handRock, Name: "Hand Rock", TypeLine: "Artifact", ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, handRock, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the restricted mana paid for a spell cast from hand")
	}
	// ...but pays for the artifact on top of the library.
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the top artifact with the restricted mana: %v", err)
	}
}

// --- Starfield Vocalist ------------------------------------------------------

const vocalistOracle = "4add809e-9caa-45ae-bb0d-4760a360b8e2"

func TestStarfieldVocalistDoublesEntersTriggersOfAnyPermanent(t *testing.T) {
	spec := effectstest.RequireFullNoCaveats(t, vocalistOracle, "Starfield Vocalist")
	if len(spec.TriggerDoublers) != 1 || len(spec.AlternativeCosts) != 1 {
		t.Fatalf("want one doubler and the warp cost, got %d and %d", len(spec.TriggerDoublers), len(spec.AlternativeCosts))
	}
	for _, vocalists := range []int{0, 1, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		for range vocalists {
			effectstest.PushCatalogPermanent(g, me.ID, "Starfield Vocalist", "Creature — Human Bard", vocalistOracle, false)
		}
		effectstest.CastFromHand(t, g, game.Card{
			Name: "Dusyut Earthcarver", TypeLine: "Artifact Creature — Dwarf Soldier", Power: 3, Toughness: 3,
			OracleID: effectstest.TdmDusyutEarthcarverOracle, ManaCost: "{3}{G}",
		}, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		// Each instance of the enters trigger is "it endures 3": a counter or a Spirit.
		for i := 0; i < 6 && effectstest.LatestOptionPickFor(g, me.ID) != nil; i++ {
			effectstest.AnswerOptionPick(t, g, me.ID, 0)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got, want := effectstest.B30TokensNamed(g, me.ID, "Spirit"), 1+vocalists; got != want {
			t.Errorf("%d Vocalists: %d Spirits from the enters trigger, want %d", vocalists, got, want)
		}
	}
}

// --- Sunset Saboteur ---------------------------------------------------------

const saboteurOracle = "c3731735-2b78-4a15-bff4-4d8559401f68"

func TestSunsetSaboteurHasWardDiscardACard(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, saboteurOracle, "Sunset Saboteur")
	g, _, opp, warded := effectstest.WardTable(t, "Sunset Saboteur", "Creature — Human Rogue", saboteurOracle)
	hand := effectstest.FillHand(opp, "Filler A", "Filler B")
	effectstest.BoltWardedCreature(t, g, opp, warded)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hand[0]) {
		t.Error("the ward's discard did not happen")
	}
}

func TestSunsetSaboteurPutsACounterOnAnOpponentsCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sab := effectstest.PushCatalogPermanent(g, me.ID, "Sunset Saboteur", "Creature — Human Rogue", saboteurOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sab, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoicePickTarget {
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
				t.Fatal("the trigger targeted a creature I control")
			}
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCountOf(t, g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("the opponent's creature has %d counters, want 1", got)
	}
}

// --- Pinnacle Emissary -------------------------------------------------------

const emissaryOracle = "522e8db5-857e-492f-afed-d60cd2ce3dcd"

func TestPinnacleEmissaryMakesABlockOnlyFlyersDroneForEachArtifactSpell(t *testing.T) {
	spec := effectstest.RequireFullNoCaveats(t, emissaryOracle, "Pinnacle Emissary")
	if len(spec.AlternativeCosts) != 1 {
		t.Fatalf("want the warp cost, got %d", len(spec.AlternativeCosts))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Pinnacle Emissary", "Artifact Creature — Robot", emissaryOracle, false)
	effectstest.CastFromHand(t, g, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHand(t, g, game.Card{Name: "Bolt Not", TypeLine: "Sorcery", ManaCost: "{1}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Drone")
	if len(tokens) != 1 {
		t.Fatalf("%d Drones, want 1 (one artifact spell, one sorcery)", len(tokens))
	}
	d := tokens[0]
	if !strings.Contains(d.TypeLine, "Artifact Creature — Drone") || d.Power != 1 || d.Toughness != 1 || len(d.Colors) != 0 {
		t.Errorf("Drone is %s %d/%d %v", d.TypeLine, d.Power, d.Toughness, d.Colors)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, d.InstanceID), "flying") {
		t.Error("the Drone does not fly")
	}
}

// --- Coordinated Maneuver ----------------------------------------------------

const maneuverOracle = "6ff39311-565e-4414-b32b-eb24c70b3090"

func TestCoordinatedManeuverDamagesByCreatureCountOrDestroysAnEnchantment(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, maneuverOracle, "Coordinated Maneuver")
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for range 3 {
		effectstest.PushVanillaCreature(g, me.ID, "Bear", 1, 1)
	}
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 5, 3)
	ench := effectstest.PushCatalogPermanent(g, opp.ID, "Pacifism-ish", "Enchantment", "", false)
	card := game.Card{Name: "Coordinated Maneuver", TypeLine: "Instant", OracleID: maneuverOracle, ManaCost: "{1}{W}"}
	// Mode 1 refuses a creature, mode 0 refuses an enchantment.
	effectstest.CastFromHand(t, g, card, game.CastSpellParams{Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, ench) {
		t.Error("mode two did not destroy the enchantment")
	}
	effectstest.CastFromHand(t, g, card, game.CastSpellParams{Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}})
	effectstest.PassPriorityAroundTable(t, g)
	// Three Bears are mine: 3 damage on a 5/3.
	if effectstest.BattlefieldHas(g, big) {
		t.Error("mode one dealt less than the three creatures I control")
	}
}

// --- Descendant of Storms ----------------------------------------------------

const descendantOracle = "ca773c52-ca64-463c-99cf-e71b59d2cff7"

func descendantAttack(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.PushCatalogPermanent(g, me.ID, "Descendant of Storms", "Creature — Human Soldier", descendantOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(d, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, d
}

func TestDescendantOfStormsPaysToEndure(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, descendantOracle, "Descendant of Storms")
	g, me, d := descendantAttack(t)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the attack trigger asked nothing")
	}
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "C", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // the counter
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestDescendantOfStormsCanMakeASpiritInstead(t *testing.T) {
	g, me, d := descendantAttack(t)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "C", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // the Spirit
	if got := effectstest.B30TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirit tokens, want 1", got)
	}
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after choosing the Spirit", got)
	}
}

func TestDescendantOfStormsDecliningDoesNothing(t *testing.T) {
	g, me, d := descendantAttack(t)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil || effectstest.CountersOn(g, d, game.CounterPlusOne) != 0 || effectstest.B30TokensNamed(g, me.ID, "Spirit") != 0 {
		t.Error("declining still endured")
	}
}

// --- Consult the Star Charts -------------------------------------------------

const consultOracle = "e921839f-9d91-41a9-bc89-016af3c757aa"

func consultTable(t *testing.T, kicked bool) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for range 3 {
		effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	}
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		// ids[0] is the top card; the top of a zone is the end of its slice.
		me.Library.Cards = nil
		for i := 0; i < 6; i++ {
			ids = append(ids, uuid.New())
		}
		for i := len(ids) - 1; i >= 0; i-- {
			me.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	params := game.CastSpellParams{}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	effectstest.CastFromHand(t, g, game.Card{Name: "Consult the Star Charts", TypeLine: "Instant", OracleID: consultOracle, ManaCost: "{1}{U}"}, params)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ids
}

func TestConsultTheStarChartsTakesOneOfTheTopXAndBottomsTheRest(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, consultOracle, "Consult the Star Charts")
	g, me, ids := consultTable(t, false)
	hand := len(me.Hand.Cards)
	// Three lands: only the top three are looked at.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 3 {
		t.Fatalf("prompt over %v, want the top three cards", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) || len(me.Hand.Cards) != hand+1 {
		t.Errorf("the chosen card did not reach the hand")
	}
	n := len(me.Library.Cards)
	if me.Library.Cards[n-1].InstanceID != ids[3] {
		t.Errorf("the card below the look is not on top")
	}
	tail := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:2] {
		tail[c.InstanceID] = true
	}
	if !tail[ids[0]] || !tail[ids[2]] {
		t.Errorf("the two cards not taken are not on the bottom")
	}
}

func TestConsultTheStarChartsKickedTakesTwo(t *testing.T) {
	g, me, ids := consultTable(t, true)
	hand := len(me.Hand.Cards)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0], ids[2])
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) || len(me.Hand.Cards) != hand+2 {
		t.Errorf("kicked, the two chosen cards did not reach the hand")
	}
}

// --- Pull Through the Weft ---------------------------------------------------

const weftOracle = "37bd9e54-c93e-4fec-9b21-6940da75ab7a"

func TestPullThroughTheWeftReturnsNonlandsToHandAndLandsTapped(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, weftOracle, "Pull Through the Weft")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mk := func(name, tl string) uuid.UUID {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		return id
	}
	bear, rock, spell := mk("Bear", "Creature — Bear"), mk("Rock", "Artifact"), mk("Opt", "Instant")
	f1, f2, f3 := mk("Forest", "Basic Land — Forest"), mk("Forest", "Basic Land — Forest"), mk("Forest", "Basic Land — Forest")
	ref := func(id uuid.UUID, slot int) game.TargetRef {
		return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
	}
	// A land can't be a "nonland permanent card", an instant isn't a permanent card.
	if err := effectstest.CastCatalogSpellErr(t, g, "Pull Through the Weft", "Sorcery", weftOracle,
		[]game.TargetRef{ref(spell, 0)}); err == nil {
		t.Fatal("targeted an instant card as a nonland permanent card")
	}
	hand := len(me.Hand.Cards)
	effectstest.CastFromHand(t, g, game.Card{Name: "Pull Through the Weft", TypeLine: "Sorcery", OracleID: weftOracle, ManaCost: "{3}{G}{G}"},
		game.CastSpellParams{Targets: []game.TargetRef{ref(bear, 0), ref(rock, 0), ref(f1, 1), ref(f2, 1)}, TargetsStamped: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(rock) || len(me.Hand.Cards) != hand+2 {
		t.Error("the two nonland permanent cards did not reach the hand")
	}
	for _, f := range []uuid.UUID{f1, f2} {
		c, ok := effectstest.BattlefieldCard(g, f)
		if !ok || !c.Tapped {
			t.Errorf("land %s did not enter tapped", f)
		}
	}
	if !me.Graveyard.Contains(f3) || !me.Graveyard.Contains(spell) {
		t.Error("an untargeted card left the graveyard")
	}
}

// --- Devastating Onslaught ---------------------------------------------------

const onslaughtOracle = "2e17cd4c-ebf0-40d0-b52d-b4d84bbcf7b3"

func TestDevastatingOnslaughtMakesXHastyCopiesSacrificedAtTheEnd(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, onslaughtOracle, "Devastating Onslaught")
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	big := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	land := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	card := game.Card{Name: "Devastating Onslaught", TypeLine: "Sorcery", OracleID: onslaughtOracle, ManaCost: "{X}{X}{R}", Colors: []string{"R"}}
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 4)
	if _, err := castStrict(t, g, card, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("targeted a land")
	}
	id, err := castStrict(t, g, card, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}})
	if err != nil {
		t.Fatalf("cast X=2 for {4}{R}: %v", err)
	}
	_ = id
	effectstest.PassPriorityAroundTable(t, g)
	tokens, _ := effectstest.CopiesOnBattlefield(g, "Big Bear")
	if len(tokens) != 2 {
		t.Fatalf("%d copies, want X=2", len(tokens))
	}
	for _, tok := range tokens {
		if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, tok.InstanceID), "haste") {
			t.Error("a copy lacks haste")
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if left, _ := effectstest.CopiesOnBattlefield(g, "Big Bear"); len(left) != 0 {
		t.Errorf("%d copies survived the end step", len(left))
	}
	if !effectstest.BattlefieldHas(g, big) {
		t.Error("the original was sacrificed")
	}
}

// --- Pinnacle Starcage -------------------------------------------------------

const starcageOracle = "10e0b563-4dfc-4530-a010-0ba40bdf02ac"

func starcageTable(t *testing.T) (g *game.Game, me *game.Player, cage, smallCreature, smallArtifact, bigCreature, bigArtifact uuid.UUID, token uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mk := func(owner uuid.UUID, name, tl, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: tl, ManaCost: cost,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	}
	smallCreature = mk(opp.ID, "Small Bear", "Creature — Bear", "{1}{G}")
	smallArtifact = mk(me.ID, "Small Rock", "Artifact", "{1}")
	bigCreature = mk(opp.ID, "Big Bear", "Creature — Bear", "{2}{G}")
	bigArtifact = mk(opp.ID, "Big Rock", "Artifact", "{3}")
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(opp.ID, effects.TokenCard("2/2 colorless Robot artifact"), 1, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("token: %v", err)
		}
		token = ids[0]
	})
	cage = effectstest.CastFromHand(t, g, game.Card{Name: "Pinnacle Starcage", TypeLine: "Artifact", OracleID: starcageOracle, ManaCost: "{1}{W}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return
}

func b5InExile(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.Exile.Contains(id) })
	return found
}

func TestPinnacleStarcageExilesSmallArtifactsAndCreaturesUntilItLeaves(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, starcageOracle, "Pinnacle Starcage")
	g, _, cage, smallC, smallA, bigC, bigA, token := starcageTable(t)
	if !b5InExile(g, smallC) || !b5InExile(g, smallA) {
		t.Fatal("a creature or artifact with mana value 2 or less stayed on the battlefield")
	}
	if !effectstest.BattlefieldHas(g, bigC) || !effectstest.BattlefieldHas(g, bigA) || !effectstest.BattlefieldHas(g, cage) {
		t.Fatal("a mana value 3 permanent, or the Starcage itself, was exiled")
	}
	if effectstest.BattlefieldHas(g, token) || b5InExile(g, token) {
		t.Error("the Robot token (mana value 0) should have been exiled and ceased to exist")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(cage); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// A card returning from exile is a new object with a new ID (CR 400.7).
	if !effectstest.BattlefieldHasName(g, "Small Bear") || !effectstest.BattlefieldHasName(g, "Small Rock") {
		t.Error("the exiled cards did not come back when the Starcage left the battlefield")
	}
	if effectstest.BattlefieldHas(g, token) {
		t.Error("the token came back")
	}
}

func TestPinnacleStarcageBuriesTheExiledCardsForRobotsAndSacrificesItself(t *testing.T) {
	g, me, cage, smallC, smallA, _, _, _ := starcageTable(t)
	effectstest.FillPoolColored(me, "W", 2)
	effectstest.FillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, cage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if b5InExile(g, smallC) || b5InExile(g, smallA) {
		t.Error("an exiled card was left in exile")
	}
	if !effectstest.InGraveyardOf(g, g.Seats[1].ID, smallC) || !effectstest.InGraveyardOf(g, me.ID, smallA) {
		t.Error("the cards did not go to their owners' graveyards")
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Robot"); got != 2 {
		t.Errorf("%d Robots, want one for each of the two cards put into a graveyard", got)
	}
	if effectstest.BattlefieldHas(g, cage) {
		t.Error("the Starcage was not sacrificed")
	}
	if effectstest.BattlefieldHasName(g, "Small Bear") || effectstest.BattlefieldHasName(g, "Small Rock") {
		t.Error("the sacrifice returned the buried cards")
	}
}

// --- Sothera, the Supervoid --------------------------------------------------

const sotheraOracle = "760ede92-a826-4144-9de8-84ba9b8743d1"

func TestSotheraExilesEachOpponentsChosenCreatureThenReanimatesAtTheEnd(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, sotheraOracle, "Sothera, the Supervoid")
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	sothera := effectstest.PushCatalogPermanent(g, me.ID, "Sothera, the Supervoid", "Legendary Enchantment", sotheraOracle, false)
	effectstest.PushVanillaCreature(g, me.ID, "My Spare", 1, 1)
	doomed := effectstest.PushVanillaCreature(g, me.ID, "My Doomed", 1, 1)
	a := effectstest.PushVanillaCreature(g, o1.ID, "Opp One A", 2, 2)
	b := effectstest.PushVanillaCreature(g, o1.ID, "Opp One B", 3, 3)
	only := effectstest.PushVanillaCreature(g, o2.ID, "Opp Two Only", 4, 4)
	_ = o3 // controls no creature: nothing to exile

	effectstest.WoeB1Destroy(t, g, doomed)
	// Opponent one has a real choice; the others do not.
	pick := effectstest.ChooseCardsChoiceFor(g, o1.ID)
	if pick == nil {
		t.Fatal("the first opponent was not asked which creature to exile")
	}
	if err := g.ResolveChooseCards(pick.ID, o1.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !b5InExile(g, b) || b5InExile(g, a) || effectstest.BattlefieldHas(g, b) {
		t.Error("opponent one's chosen creature was not exiled (and only it)")
	}
	if !b5InExile(g, only) {
		t.Error("opponent two's only creature was not exiled")
	}

	// o3 controls no creatures at my end step: Sothera is sacrificed and a
	// creature card exiled with it comes back under my control with two
	// counters; the choice is mine among the two exiled.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked which exiled creature to put onto the battlefield")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{only}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, sothera) {
		t.Error("Sothera was not sacrificed")
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Opp Two Only")
	if len(ids) != 1 {
		t.Fatalf("the chosen creature is on the battlefield %d times", len(ids))
	}
	if c, _ := effectstest.BattlefieldCard(g, ids[0]); c.Controller != me.ID {
		t.Fatalf("the chosen creature is not under my control: %+v", c)
	}
	if got := effectstest.CountersOn(g, ids[0], game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if !b5InExile(g, b) {
		t.Error("the other exiled creature should have stayed in exile")
	}
}

// With every player holding a creature, the end-step trigger never fires.
func TestSotheraEndStepNeedsAPlayerWithNoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sothera := effectstest.PushCatalogPermanent(g, me.ID, "Sothera, the Supervoid", "Legendary Enchantment", sotheraOracle, false)
	for _, s := range g.Seats {
		effectstest.PushVanillaCreature(g, s.ID, "Bear of "+s.Name, 1, 1)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, sothera) {
		t.Error("Sothera was sacrificed although every player controls a creature")
	}
}
