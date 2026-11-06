package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q, want %q", oracle, spec.Name, name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
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
	return pushToGraveyard(g, p, game.Card{Name: "Old Lesson", TypeLine: "Sorcery — Lesson", ManaCost: "{1}"})
}

// --- How to Start a Riot --------------------------------------------------------

func TestAvatarB12RiotGrantsMenaceAndPumpsThePlayersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)

	castCatalogSpell(t, g, "How to Start a Riot", "Instant — Lesson", avB12RiotOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs, Slot: 0}, {Kind: game.TargetPlayer, ID: me.ID, Slot: 1}})
	passPriorityAroundTable(t, g)

	if !hasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("the targeted creature has no menace")
	}
	if hasEffectiveKeyword(t, g, mine, "menace") {
		t.Error("an untargeted creature gained menace")
	}
	if got := currentPower(t, g, mine); got != 4 {
		t.Errorf("my 2/2 has power %d, want 4", got)
	}
	if got := currentPower(t, g, other); got != 3 {
		t.Errorf("my 1/1 has power %d, want 3", got)
	}
	if got := currentPower(t, g, theirs); got != 3 {
		t.Errorf("the other player's creature has power %d, want 3", got)
	}
}

// --- Iroh's Demonstration -------------------------------------------------------

func TestAvatarB12IrohsDemonstrationPingsOnlyOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	sturdy := pushVanillaCreature(g, opp.ID, "Sturdy", 1, 2)

	b02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	four := pushVanillaCreature(g, opp.ID, "Four", 4, 4)

	if _, err := avB12Cast(t, g, game.Card{Name: "Iroh's Demonstration", TypeLine: "Sorcery — Lesson", OracleID: avB12IrohOracle},
		game.CastSpellParams{Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("the creature mode accepted a player target")
	}
	_ = me
	b02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{1}, avB12Refs(four))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) {
		t.Error("the 4/4 survived 4 damage")
	}
}

// --- Jeong Jeong, the Deserter --------------------------------------------------

func TestAvatarB12JeongJeongFirebendsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jj := b27Push(g, me.ID, "Jeong Jeong, the Deserter", "Legendary Creature — Human Rebel Ally", avB12JeongOracle, "{2}{R}", 2, 3, "R")
	declareAttack(t, g, opp.ID, jj)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvanceTo(t, g, game.StepEndCombat)
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

func TestAvatarB12JeongJeongCopiesTheNextLessonOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jj := b27Push(g, me.ID, "Jeong Jeong, the Deserter", "Legendary Creature — Human Rebel Ally", avB12JeongOracle, "{2}{R}", 2, 3, "R")
	two := pushVanillaCreature(g, opp.ID, "Two", 2, 2)

	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, jj, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, jj, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, jj, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })

	// A non-Lesson spell does not use the trigger up.
	oppLife := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Fatalf("opponent at %d after one Bolt, want %d (no copy)", opp.Life, oppLife-3)
	}

	// The Lesson is copied: two pings kill the 2/2.
	b02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	for i := 0; i < 4 && !stackFullyEmpty(g); i++ {
		answerAnyTriggerOrderPrompt(t, g, me.ID)
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(two) {
		t.Error("the 2/2 survived: the Lesson was not copied")
	}

	// The trigger is spent: the next Lesson is not copied.
	three := pushVanillaCreature(g, opp.ID, "Another Two", 2, 2)
	b02cCastModalSpell(t, g, "Iroh's Demonstration", "Sorcery — Lesson", avB12IrohOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(three) {
		t.Error("a second Lesson was copied too")
	}
}

// --- Jet's Brainwashing ---------------------------------------------------------

func TestAvatarB12JetsBrainwashingUnkickedStopsABlockAndMakesAClue(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	castCatalogSpell(t, g, "Jet's Brainwashing", "Sorcery", avB12JetOracle, avB12Refs(bear))
	passPriorityAroundTable(t, g)

	c, _ := g.LookupCardForEffect(bear)
	if c.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the targeted creature can still block")
	}
	if c.Controller != opp.ID {
		t.Error("an unkicked Brainwashing took control")
	}
	if got := countOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestAvatarB12JetsBrainwashingKickedThreatens(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, bear).Tapped = true })

	if _, err := avB12Cast(t, g, game.Card{Name: "Jet's Brainwashing", TypeLine: "Sorcery", OracleID: avB12JetOracle},
		game.CastSpellParams{Targets: avB12Refs(bear), OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := findBattlefieldCardByID(g, bear)
	if c == nil {
		t.Fatal("the creature is gone")
	}
	if c.Controller != me.ID {
		t.Error("the kicked spell did not take control")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the creature has no haste")
	}
	if got := countOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Mai, Jaded Edge ------------------------------------------------------------

func TestAvatarB12MaiGetsADoubleStrikeCounterOnce(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mai := b27Push(g, me.ID, "Mai, Jaded Edge", "Legendary Creature — Human Noble", avB12MaiOracle, "{1}{R}", 1, 3, "R")

	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, mai, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, mai, game.CounterDoubleStrike); got != 1 {
		t.Fatalf("double strike counters = %d, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, mai, "double strike") {
		t.Error("Mai has no double strike")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, mai, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

func TestAvatarB12MaiHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mai := b27Push(g, me.ID, "Mai, Jaded Edge", "Legendary Creature — Human Noble", avB12MaiOracle, "{1}{R}", 1, 3, "R")
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := currentPower(t, g, mai); got != 2 {
		t.Errorf("Mai's power is %d after a noncreature spell, want 2", got)
	}
}

// --- Mongoose Lizard ------------------------------------------------------------

func TestAvatarB12MongooseLizardPingsAnyTargetOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life

	if _, err := avB12Cast(t, g, game.Card{Name: "Mongoose Lizard", TypeLine: "Creature — Mongoose Lizard",
		OracleID: avB12LizardOracle, ManaCost: "{4}{R}{R}", Power: 5, Toughness: 6, Colors: []string{"R"}}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

func TestAvatarB12MongooseLizardMountaincyclesForAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := stapleLibraryCard(me, "Mountain", "Basic Land — Mountain")
	stapleLibraryCard(me, "Forest", "Basic Land — Forest")

	lizard, _ := cycleFromHand(t, g, "Mongoose Lizard", "Creature — Mongoose Lizard", avB12LizardOracle, "{C}{C}")
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		for _, id := range c.DiscardOptions {
			if card, ok := g.LookupCardForEffect(id); ok && card.Name == "Forest" {
				t.Error("a Forest was offered to mountaincycling")
			}
		}
	}
	answerSearchByID(t, g, me.ID, mountain)
	passPriorityAroundTable(t, g)
	if !inHand(me, mountain) {
		t.Error("the Mountain is not in hand")
	}
	if inHand(me, lizard) {
		t.Error("the Lizard was not discarded")
	}
}

// --- Price of Freedom -----------------------------------------------------------

func TestAvatarB12PriceOfFreedomDestroysOffersATappedBasicAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	forest := stapleLibraryCard(opp, "Forest", "Basic Land — Forest")
	seedLibrary(me, "Next Card")

	castCatalogSpell(t, g, "Price of Freedom", "Sorcery — Lesson", avB12PriceOracle, avB12Refs(relic))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the caster was offered the search; it belongs to the artifact's controller")
	}
	answerSearchByID(t, g, opp.ID, forest)
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, forest)
	if !ok || !got.Tapped || got.Controller != opp.ID {
		t.Errorf("the basic must be on the battlefield tapped under its owner, got %+v (found %v)", got, ok)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestAvatarB12PriceOfFreedomTargetsOnlyAnOpponentsArtifactOrLand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	myLand := seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirBear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB12Lesson(g, me)
	avB12Lesson(g, me)
	pushToGraveyard(g, me, game.Card{Name: "Old Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}", Power: 4, Toughness: 4})

	ran, err := avB12Cast(t, g, avB12RanAndShaw(), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB12Lesson(g, me)
	avB12Lesson(g, me)
	pushToGraveyard(g, me, game.Card{Name: "Old Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2})

	if _, err := avB12Cast(t, g, avB12RanAndShaw(), game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := len(avB12RanAndShawTokens(g, me.ID)); got != 0 {
		t.Errorf("%d token copies with two Lessons and a Bear, want 0", got)
	}
}

func TestAvatarB12RanAndShawRecheckTheGraveyardOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
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
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ran := b27Push(g, me.ID, "Ran and Shaw", "Legendary Creature — Dragon", avB12RanShawOracle, "{3}{R}{R}", 4, 4, "R")
	drake := pushPermanent(g, me.ID, game.Card{Name: "Drake", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)

	fillPool(me, 3)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, ran, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := currentPower(t, g, drake); got != 4 {
		t.Errorf("the other Dragon has power %d, want 4", got)
	}
	if got := currentPower(t, g, ran); got != 6 {
		t.Errorf("Ran and Shaw have power %d, want 6", got)
	}
	if got := currentPower(t, g, bear); got != 2 {
		t.Errorf("a non-Dragon has power %d, want 2", got)
	}

	declareAttack(t, g, opp.ID, ran)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
}

// --- Rough Rhino Cavalry --------------------------------------------------------

func TestAvatarB12RoughRhinoCavalryExhaustGrowsAndTramples(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rhino := b27Push(g, me.ID, "Rough Rhino Cavalry", "Creature — Human Mercenary", avB12RhinoOracle, "{4}{R}", 5, 5, "R")

	fillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the exhaust ability was activated for seven mana")
	}
	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, rhino, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, rhino, "trample") {
		t.Error("no trample")
	}
	fillPool(me, 8)
	if err := g.ActivateCatalogAbility(me.ID, rhino, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

func TestAvatarB12RoughRhinoCavalryFirebendsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhino := b27Push(g, me.ID, "Rough Rhino Cavalry", "Creature — Human Mercenary", avB12RhinoOracle, "{4}{R}", 5, 5, "R")
	declareAttack(t, g, opp.ID, rhino)
	passPriorityAroundTable(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
	mkAdvanceTo(t, g, game.StepPostcombatMain)
	mkWantPool(t, me, map[string]int{})
}

// --- Solstice Revelations -------------------------------------------------------

func avB12SolsticeTable(t *testing.T, mountains int, hitCost string) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < mountains; i++ {
		seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	land := discoverLibraryCard(me.ID, "Top Land", "Basic Land — Forest", "")
	hit := discoverLibraryCard(me.ID, "The Hit", "Artifact", hitCost)
	under := discoverLibraryCard(me.ID, "Under", "Artifact", "{1}")
	me.Library.Cards = []game.Card{under, hit, land} // the last card is the top
	return g, me, land.InstanceID, hit.InstanceID
}

func TestAvatarB12SolsticeRevelationsCastsALesserCardFree(t *testing.T) {
	g, me, land, hit := avB12SolsticeTable(t, 3, "{2}")
	castCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeOrHand, hit)
	if err := castThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hit) {
		t.Error("the cast card is not on the battlefield")
	}
	if !inExile(g, land) {
		t.Error("the land exiled on the way is not in exile")
	}
}

func TestAvatarB12SolsticeRevelationsPutsAnEqualCardIntoHand(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 3, "{3}")
	castCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a card with mana value equal to the Mountains was offered")
	}
	if !inHand(me, hit) {
		t.Error("the card is not in hand")
	}
}

func TestAvatarB12SolsticeRevelationsDeclinePutsItIntoHand(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 3, "{2}")
	castCatalogSpell(t, g, "Solstice Revelations", "Instant — Lesson", avB12SolsticeOracle, nil)
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !inHand(me, hit) {
		t.Error("the declined card is not in hand")
	}
}

func TestAvatarB12SolsticeRevelationsFlashesBack(t *testing.T) {
	g, me, _, hit := avB12SolsticeTable(t, 0, "{0}")
	id := pushToGraveyard(g, me, game.Card{Name: "Solstice Revelations", TypeLine: "Instant — Lesson", OracleID: avB12SolsticeOracle, ManaCost: "{2}{R}"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	// No Mountains: even a mana value 0 card is not less than zero.
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a card was offered with no Mountains")
	}
	if !inHand(me, hit) {
		t.Error("the card is not in hand")
	}
	if !inExile(g, id) {
		t.Error("the flashed-back spell was not exiled")
	}
}
