package testsav

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b12_cards_test.go — Avatar: The Last Airbender batch 12 (#4459).

const (
	avB12RiotOracle     = "268d0056-0d3e-438a-ab17-625a6f5ec68a"
	avB12IrohOracle     = "8dd2e73a-d1b4-4d26-857f-9eb8e384c8ff"
	avB12JeongOracle    = "2158c73d-421b-4c94-af06-dd89cb8d3126"
	avB12JetOracle      = "5cca85e0-43e8-4b97-a43b-1b8f23f541d1"
	avB12MaiOracle      = "b5029f21-e0d6-4f22-9afd-79452de26490"
	avB12LizardOracle   = "a9bae56c-9666-42e8-94ec-03b0a964b142"
	avB12PriceOracle    = "5e6faa98-cf55-430d-87ce-8085f68aa20a"
	avB12RanShawOracle  = "a728685f-8670-4db2-ae02-3cf74eb3c402"
	avB12RhinoOracle    = "3ac94b28-4b0a-4121-99b9-9cd7dff7948c"
	avB12SolsticeOracle = "310c1280-dee8-426f-8440-6132a634a1e0"
)

func TestAvatarB12CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB12RiotOracle:     "How to Start a Riot",
		avB12IrohOracle:     "Iroh's Demonstration",
		avB12JeongOracle:    "Jeong Jeong, the Deserter",
		avB12JetOracle:      "Jet's Brainwashing",
		avB12MaiOracle:      "Mai, Jaded Edge",
		avB12LizardOracle:   "Mongoose Lizard",
		avB12PriceOracle:    "Price of Freedom",
		avB12RanShawOracle:  "Ran and Shaw",
		avB12RhinoOracle:    "Rough Rhino Cavalry",
		avB12SolsticeOracle: "Solstice Revelations",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB12Cast puts `c` into the active seat's hand, walks to a main phase
// and casts it with `params`.
func avB12Cast(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

func avB12Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func avB12Lesson(g *game.Game, p *game.Player) uuid.UUID {
	return effectstest.PushToGraveyard(g, p, game.Card{Name: "Old Lesson", TypeLine: "Sorcery — Lesson", ManaCost: "{1}"})
}

// --- How to Start a Riot --------------------------------------------------------

func TestAvatarB12RiotGrantsMenaceAndPumpsThePlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 3, 3)

	effectstest.CastCatalogSpell(t, g, "How to Start a Riot", "Instant — Lesson", avB12RiotOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetPlayer, ID: me.ID, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("the targeted creature has no menace")
	}
	if effectstest.HasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("an untargeted creature gained menace")
	}
	if got := effectstest.CurrentPower(t, g, mine); got != 4 {
		t.Errorf("my 2/2 has power %d, want 4", got)
	}
	if got := effectstest.CurrentPower(t, g, other); got != 3 {
		t.Errorf("my 1/1 has power %d, want 3", got)
	}
	if got := effectstest.CurrentPower(t, g, theirs); got != 3 {
		t.Errorf("the other player's creature has power %d, want 3", got)
	}
}

// --- Iroh's Demonstration -------------------------------------------------------

func TestAvatarB12IrohsDemonstrationPingsOnlyOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	sturdy := effectstest.PushVanillaCreature(g, opp.ID, "Sturdy", 1, 2)

	effectstest.B02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(mine) {
		t.Error("my own 1/1 died")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 1/1 survived")
	}
	if !g.Battlefield.Contains(sturdy) {
		t.Error("the opponent's 1/2 died to 1 damage")
	}
}

func TestAvatarB12IrohsDemonstrationDealsFourToATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	four := effectstest.PushVanillaCreature(g, opp.ID, "Four", 4, 4)

	if _, err := avB12Cast(t, g, game.Card{Name: "Iroh's Demonstration", TypeLine: "Sorcery — Lesson", OracleID: avB12IrohOracle},
		game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("the creature mode accepted a player target")
	}
	_ = me
	effectstest.B02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{1}, avB12Refs(four))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- Jeong Jeong, the Deserter --------------------------------------------------

func TestAvatarB12JeongJeongFirebendsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jj := effectstest.B27Push(g, me.ID, "Jeong Jeong, the Deserter", "Legendary Creature — Human Rebel Ally", avB12JeongOracle, "{2}{R}", 2, 3, "R")
	effectstest.DeclareAttack(t, g, opp.ID, jj)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	effectstest.MkAdvanceTo(t, g, game.StepEndCombat)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestAvatarB12JeongJeongCopiesTheNextLessonOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jj := effectstest.B27Push(g, me.ID, "Jeong Jeong, the Deserter", "Legendary Creature — Human Rebel Ally", avB12JeongOracle, "{2}{R}", 2, 3, "R")
	two := effectstest.PushVanillaCreature(g, opp.ID, "Two", 2, 2)

	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, jj, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, jj, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, jj, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })

	// A non-Lesson spell does not use the trigger up.
	oppLife := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Fatalf("opponent at %d after one Bolt, want %d (no copy)", opp.Life, oppLife-3)
	}

	// The Lesson is copied: two pings kill the 2/2.
	effectstest.B02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	for i := 0; i < 4 && !effectstest.StackFullyEmpty(g); i++ {
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(two) {
		t.Error("the 2/2 survived: the Lesson was not copied")
	}

	// The trigger is spent: the next Lesson is not copied.
	three := effectstest.PushVanillaCreature(g, opp.ID, "Another Two", 2, 2)
	effectstest.B02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("a second Lesson was copied too")
	}
}

// --- Jet's Brainwashing ---------------------------------------------------------

func TestAvatarB12JetsBrainwashingUnkickedStopsABlockAndMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Jet's Brainwashing", "Sorcery", avB12JetOracle, avB12Refs(bear))
	effectstest.PassPriorityAroundTable(t, g)

	c, _ := g.LookupCardForEffect(bear)
	if c.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the targeted creature can still block")
	}
	if c.Controller != opp.ID {
		t.Error("an unkicked Brainwashing took control")
	}
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestAvatarB12JetsBrainwashingKickedThreatens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, bear).Tapped = true })

	if _, err := avB12Cast(t, g, game.Card{Name: "Jet's Brainwashing", TypeLine: "Sorcery", OracleID: avB12JetOracle},
		game.CastSpellParams{Targets: avB12Refs(bear), OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindBattlefieldCardByID(g, bear)
	if c == nil {
		t.Fatal("the creature is gone")
	}
	if c.Controller != me.ID {
		t.Error("the kicked spell did not take control")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the creature has no haste")
	}
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Mai, Jaded Edge ------------------------------------------------------------

func TestAvatarB12MaiGetsADoubleStrikeCounterOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mai := effectstest.B27Push(g, me.ID, "Mai, Jaded Edge", "Legendary Creature — Human Noble", avB12MaiOracle, "{1}{R}", 1, 3, "R")

	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, mai, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mai, game.CounterDoubleStrike); got != 1 {
		t.Fatalf("double strike counters = %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mai, "double strike") {
		t.Error("Mai has no double strike")
	}
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, mai, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

func TestAvatarB12MaiHasProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mai := effectstest.B27Push(g, me.ID, "Mai, Jaded Edge", "Legendary Creature — Human Noble", avB12MaiOracle, "{1}{R}", 1, 3, "R")
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CurrentPower(t, g, mai); got != 2 {
		t.Errorf("Mai's power is %d after a noncreature spell, want 2", got)
	}
}

// --- Mongoose Lizard ------------------------------------------------------------

func TestAvatarB12MongooseLizardPingsAnyTargetOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life

	if _, err := avB12Cast(t, g, game.Card{Name: "Mongoose Lizard", TypeLine: "Creature — Mongoose Lizard",
		OracleID: avB12LizardOracle, ManaCost: "{4}{R}{R}", Power: 5, Toughness: 6, Colors: []string{"R"}}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

func TestAvatarB12MongooseLizardMountaincyclesForAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := effectstest.StapleLibraryCard(me, "Mountain", "Basic Land — Mountain")
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")

	lizard, _ := effectstest.CycleFromHand(t, g, "Mongoose Lizard", "Creature — Mongoose Lizard", avB12LizardOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		for _, id := range c.DiscardOptions {
			if card, ok := g.LookupCardForEffect(id); ok && card.Name == "Forest" {
				t.Error("a Forest was offered to mountaincycling")
			}
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, mountain)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(me, mountain) {
		t.Error("the Mountain is not in hand")
	}
	if effectstest.InHand(me, lizard) {
		t.Error("the Lizard was not discarded")
	}
}

// --- Price of Freedom -----------------------------------------------------------

func TestAvatarB12PriceOfFreedomDestroysOffersATappedBasicAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	forest := effectstest.StapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	effectstest.SeedLibrary(me, "Next Card")

	effectstest.CastCatalogSpell(t, g, "Price of Freedom", "Sorcery — Lesson", avB12PriceOracle, avB12Refs(relic))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the caster was offered the search; it belongs to the artifact's controller")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	got, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !got.Tapped || got.Controller != opp.ID {
		t.Errorf("the basic must be on the battlefield tapped under its owner, got %+v (found %v)", got, ok)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestAvatarB12PriceOfFreedomTargetsOnlyAnOpponentsArtifactOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	myLand := effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirBear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	for _, id := range []uuid.UUID{myLand, theirBear} {
		if _, err := avB12Cast(t, g, game.Card{Name: "Price of Freedom", TypeLine: "Sorcery — Lesson", OracleID: avB12PriceOracle},
			game.CastSpellParams{Strict: true, Targets: avB12Refs(id)}); err == nil {
			t.Errorf("%s was a legal target", id)
		}
	}
}

// --- Ran and Shaw ---------------------------------------------------------------

func avB12RanAndShaw() game.Card {
	return game.Card{Name: "Ran and Shaw", TypeLine: "Legendary Creature — Dragon", OracleID: avB12RanShawOracle,
		ManaCost: "{3}{R}{R}", Power: 4, Toughness: 4, Colors: []string{"R"}, Keywords: []string{"flying"}}
}

func avB12RanAndShawTokens(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ran and Shaw" && c.IsToken() && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func TestAvatarB12RanAndShawCopyThemselvesWithThreeDragonOrLessonCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB12Lesson(g, me)
	avB12Lesson(g, me)
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Old Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}", Power: 4, Toughness: 4})

	ran, err := avB12Cast(t, g, avB12RanAndShaw(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := avB12RanAndShawTokens(g, me.ID)
	if len(tokens) != 1 {
		t.Fatalf("%d token copies, want 1", len(tokens))
	}
	if strings.Contains(tokens[0].TypeLine, "Legendary") {
		t.Errorf("the token is legendary: %q", tokens[0].TypeLine)
	}
	if !g.Battlefield.Contains(ran) {
		t.Error("the original left the battlefield (legend rule?)")
	}
}

func TestAvatarB12RanAndShawNeedThreeCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB12Lesson(g, me)
	avB12Lesson(g, me)
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Old Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})

	if _, err := avB12Cast(t, g, avB12RanAndShaw(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB12RanAndShawTokens(g, me.ID)); got != 0 {
		t.Errorf("%d token copies with two Lessons and a Bear, want 0", got)
	}
}

func TestAvatarB12RanAndShawRecheckTheGraveyardOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB12Lesson(g, me)
	avB12Lesson(g, me)
	third := avB12Lesson(g, me)

	if _, err := avB12Cast(t, g, avB12RanAndShaw(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// Resolve the creature spell; the trigger goes on the stack.
	for i := 0; i < 8 && !avB12StackHasLabelPrefix(g, "Ran and Shaw — create"); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !avB12StackHasLabelPrefix(g, "Ran and Shaw — create") {
		t.Fatal("the enters trigger never reached the stack")
	}
	g.WithWriteLock(func() {
		var keep []game.Card
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID != third {
				keep = append(keep, c)
			}
		}
		me.Graveyard.Cards = keep
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(avB12RanAndShawTokens(g, me.ID)); got != 0 {
		t.Errorf("%d token copies after the graveyard fell to two, want 0", got)
	}
}

func avB12StackHasLabelPrefix(g *game.Game, prefix string) bool {
	for _, it := range g.StackMeta {
		if it != nil && strings.HasPrefix(it.Label, prefix) {
			return true
		}
	}
	return false
}

func TestAvatarB12RanAndShawPumpDragonsAndFirebendTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ran := effectstest.B27Push(g, me.ID, "Ran and Shaw", "Legendary Creature — Dragon", avB12RanShawOracle, "{3}{R}{R}", 4, 4, "R")
	drake := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Drake", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, ran, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CurrentPower(t, g, drake); got != 4 {
		t.Errorf("the other Dragon has power %d, want 4", got)
	}
	if got := effectstest.CurrentPower(t, g, ran); got != 6 {
		t.Errorf("Ran and Shaw have power %d, want 6", got)
	}
	if got := effectstest.CurrentPower(t, g, bear); got != 2 {
		t.Errorf("a non-Dragon has power %d, want 2", got)
	}

	effectstest.DeclareAttack(t, g, opp.ID, ran)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
}

// --- Rough Rhino Cavalry --------------------------------------------------------

func TestAvatarB12RoughRhinoCavalryExhaustGrowsAndTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhino := effectstest.B27Push(g, me.ID, "Rough Rhino Cavalry", "Creature — Human Mercenary", avB12RhinoOracle, "{4}{R}", 5, 5, "R")

	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the exhaust ability was activated for seven mana")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, rhino, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rhino, "trample") {
		t.Error("no trample")
	}
	effectstest.FillPool(me, 8)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

func TestAvatarB12RoughRhinoCavalryFirebendsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhino := effectstest.B27Push(g, me.ID, "Rough Rhino Cavalry", "Creature — Human Mercenary", avB12RhinoOracle, "{4}{R}", 5, 5, "R")
	effectstest.DeclareAttack(t, g, opp.ID, rhino)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, me, map[string]int{})
}

// --- Solstice Revelations -------------------------------------------------------

func avB12SolsticeTable(t *testing.T, mountains int, hitCost string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < mountains; i++ {
		effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	land := effectstest.DiscoverLibraryCard(me.ID, "Top Land", "Basic Land — Forest", "")
	hit := effectstest.DiscoverLibraryCard(me.ID, "The Hit", "Artifact", hitCost)
	under := effectstest.DiscoverLibraryCard(me.ID, "Under", "Artifact", "{1}")
	me.Library.Cards = []game.Card{under, hit, land} // the last card is the top
	return g, me, land.InstanceID, hit.InstanceID
}

func TestAvatarB12SolsticeRevelationsCastsALesserCardFree(t *testing.T) {
	g, me, land, hit := avB12SolsticeTable(t, 3, "{2}")
	effectstest.CastCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeOrHand, hit)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hit) {
		t.Error("the cast card is not on the battlefield")
	}
	if !effectstest.InExile(g, land) {
		t.Error("the land exiled on the way is not in exile")
	}
}

func TestAvatarB12SolsticeRevelationsPutsAnEqualCardIntoHand(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 3, "{3}")
	effectstest.CastCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a card with mana value equal to the Mountains was offered")
	}
	if !effectstest.InHand(me, hit) {
		t.Error("the card is not in hand")
	}
}

func TestAvatarB12SolsticeRevelationsDeclinePutsItIntoHand(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 3, "{2}")
	effectstest.CastCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(me, hit) {
		t.Error("the declined card is not in hand")
	}
}

func TestAvatarB12SolsticeRevelationsFlashesBack(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 0, "{0}")
	id := effectstest.PushToGraveyard(g, me, game.Card{Name: "Solstice Revelations", TypeLine: "Instant — Lesson", OracleID: avB12SolsticeOracle, ManaCost: "{2}{R}"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// No Mountains: even a mana value 0 card is not less than zero.
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a card was offered with no Mountains")
	}
	if !effectstest.InHand(me, hit) {
		t.Error("the card is not in hand")
	}
	if !effectstest.InExile(g, id) {
		t.Error("the flashed-back spell was not exiled")
	}
}
