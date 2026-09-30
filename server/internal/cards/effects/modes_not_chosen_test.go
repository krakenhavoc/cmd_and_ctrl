package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// modes_not_chosen_test.go — ADR 0097 (#1749): "choose one that
// hasn't been chosen [this turn]". Every test drives a real card (or,
// for the activated ability no catalogued card has yet, a registered
// fixture) through the engine's own verbs: the harvest, the mode_pick
// prompt, the activation gate and the bot's enumerator.

const monumentToEnduranceOracle = "e69e8de4-b521-4888-8074-17f1efe2f345"

// modePickChoicesFor is every open mode_pick prompt for a chooser, in
// queue order.
func modePickChoicesFor(g *game.Game, chooser uuid.UUID) []*game.PendingChoice {
	var out []*game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			out = append(out, c)
		}
	}
	return out
}

// enterCreatureFor creates one creature token for `p` — "another
// creature you control enters" for a Gala Greeters.
func enterCreatureFor(t *testing.T, g *game.Game, p uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(p, TokenCard("1/1 white Soldier"), n); err != nil {
			t.Fatalf("CreateTokenForEffect: %v", err)
		}
	})
}

func answerMode(t *testing.T, g *game.Game, c *game.PendingChoice, chooser uuid.UUID, mode int) {
	t.Helper()
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
}

// A used mode is not offered; it rides the prompt as the used list,
// and the answer gate refuses it.
func TestGalaGreetersOffersOnlyTheModesNotChosenThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greeters := b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)

	enterCreatureFor(t, g, me.ID, 1)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1, 2}) || len(c.ModeUsedIndex) != 0 {
		t.Fatalf("first trigger offers all three, nothing used: %+v", c)
	}
	if c.ModeNotChosen != game.ModeMemoryThisTurn {
		t.Errorf("the prompt says which restriction it is: %v", c.ModeNotChosen)
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)

	enterCreatureFor(t, g, me.ID, 1)
	c = modePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 2}) {
		t.Fatalf("second trigger offers the counter and the life: %+v", c)
	}
	if !slices.Equal(c.ModeUsedIndex, []int{1}) || c.ModeUsedLabel[0] != "Create a tapped Treasure token." {
		t.Errorf("the Treasure rides the prompt as used: %v %v", c.ModeUsedIndex, c.ModeUsedLabel)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err == nil {
		t.Fatal("the answer gate refuses a used mode")
	}
	answerMode(t, g, c, me.ID, 2)
	passPriorityAroundTable(t, g)

	enterCreatureFor(t, g, me.ID, 1)
	c = modePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0}) {
		t.Fatalf("third trigger has one mode left and must take it: %+v", c)
	}
	// The bot's enumerator reads the prompt: it offers the counter and
	// nothing else.
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			Modes []int `json:"modes"`
		}
		if json.Unmarshal(m.Params, &p) == nil && len(p.Modes) > 0 && !slices.Equal(p.Modes, []int{0}) {
			t.Errorf("the enumerator offered a used mode: %v", p.Modes)
		}
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, greeters, "+1/+1"); n != 1 {
		t.Errorf("the counter bullet: %d, want 1", n)
	}

	// CR 700.2b / the Breeches ruling: with all three used the fourth
	// instance is removed — no prompt, nothing on the stack.
	enterCreatureFor(t, g, me.ID, 1)
	if c := modePickChoiceFor(g, me.ID); c != nil {
		t.Fatalf("an exhausted trigger asks nothing: %+v", c)
	}
	if triggerOnStack(g, greeters) != nil {
		t.Error("an exhausted trigger is not put on the stack")
	}

	// The memory is "this turn": the next turn offers all three again.
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	enterCreatureFor(t, g, me.ID, 1)
	if c := modePickChoiceFor(g, me.ID); c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1, 2}) {
		t.Fatalf("a new turn resets the memory: %+v", c)
	}
}

// The Gala Greeters ruling: several instances of the ability triggered
// at once must choose different modes, and only the first three choose
// at all.
func TestGalaGreetersSimultaneousTriggersChooseDifferentModes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)

	enterCreatureFor(t, g, me.ID, 4)
	picks := modePickChoicesFor(g, me.ID)
	if len(picks) != 4 {
		t.Fatalf("four creatures entering at once queue four prompts, got %d", len(picks))
	}
	answerMode(t, g, picks[0], me.ID, 2)
	picks = modePickChoicesFor(g, me.ID)
	if len(picks) != 3 {
		t.Fatalf("three prompts left, got %d", len(picks))
	}
	for _, c := range picks {
		if slices.Contains(c.ModeOptionIndex, 2) || !slices.Contains(c.ModeUsedIndex, 2) {
			t.Errorf("the life mode is taken off every other open prompt: %v / used %v", c.ModeOptionIndex, c.ModeUsedIndex)
		}
	}
	answerMode(t, g, picks[0], me.ID, 0)
	picks = modePickChoicesFor(g, me.ID)
	if len(picks) != 2 || !slices.Equal(picks[0].ModeOptionIndex, []int{1}) {
		t.Fatalf("two prompts, each offering only the Treasure: %d", len(picks))
	}
	answerMode(t, g, picks[0], me.ID, 1)
	// "That choice is made only for the first three": the fourth
	// instance, left with nothing to choose, is withdrawn.
	if picks := modePickChoicesFor(g, me.ID); len(picks) != 0 {
		t.Fatalf("the fourth instance is withdrawn, not left open: %d prompts", len(picks))
	}
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 1 {
		t.Errorf("one Treasure, from the one instance that chose it: %d", b16CountNamed(g, "Treasure"))
	}
}

// The Demonic Pact ruling: a mode chosen for an instance that never
// resolves "still counts as being chosen". And CR 700.2g: a COPY of
// the trigger copies the mode and records nothing of its own.
func TestNotChosenIsRecordedAtTheChoiceNotAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greeters := b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)

	enterCreatureFor(t, g, me.ID, 1)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	item := triggerOnStack(g, greeters)
	if item == nil {
		t.Fatal("the trigger is on the stack")
	}
	g.WithWriteLock(func() {
		if err := (CounterTarget{StackID: item.ID}).Apply(ctxFor(g, &game.StackItem{Controller: g.Seats[1].ID})); err != nil {
			t.Fatalf("counter the trigger: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 0 {
		t.Fatal("the countered trigger made nothing")
	}
	enterCreatureFor(t, g, me.ID, 1)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || slices.Contains(c.ModeOptionIndex, 1) {
		t.Fatalf("a countered trigger still used its mode: %+v", c)
	}

	// Choose the life, then copy the trigger on the stack.
	life := me.Life
	answerMode(t, g, c, me.ID, 2)
	item = triggerOnStack(g, greeters)
	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Fatalf("CopyAbilityForEffect: %v", err)
		}
	})
	if modePickChoiceFor(g, me.ID) != nil {
		t.Fatal("a copy of a modal ability copies its modes; it is never asked (CR 700.2g)")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("the copy resolved the copied mode too: life %d → %d, want +4", life, me.Life)
	}
	var used []int
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, greeters)
		used = g.ModesChosenForEffect(b15GalaGreetersTrigger().Modes, game.ModeAbilityOf(c, b15GalaGreetersLabel))
	})
	if !slices.Equal(used, []int{1, 2}) {
		t.Errorf("the memory holds the two CHOSEN modes and nothing the copy did: %v", used)
	}
}

// The Demonic Pact ruling: "it doesn't matter who has chosen any
// particular mode". A change of control keeps the memory.
func TestNotChosenSurvivesAChangeOfControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greeters := b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)

	enterCreatureFor(t, g, me.ID, 1)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := (GainControl{Target: greeters, Controller: opp.ID}).Apply(ctxFor(g, &game.StackItem{Controller: opp.ID})); err != nil {
			t.Fatalf("GainControl: %v", err)
		}
	})
	enterCreatureFor(t, g, opp.ID, 1)
	c := modePickChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the new controller's creature triggers the stolen Greeters")
	}
	if !slices.Equal(c.ModeOptionIndex, []int{0, 2}) || !slices.Equal(c.ModeUsedIndex, []int{1}) {
		t.Errorf("the memory belongs to the object, not the controller: offered %v used %v", c.ModeOptionIndex, c.ModeUsedIndex)
	}
}

// "Ever": Silent Hallcreeper remembers across turns, runs out, and a
// new object (CR 400.7) remembers nothing.
func TestSilentHallcreeperModesAreOnceEachForThatObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	seat := g.Turn.ActiveSeat
	creeper := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Silent Hallcreeper", OracleID: silentHallcreeperOracle,
		TypeLine: "Enchantment Creature — Horror", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	connect := func() *game.PendingChoice {
		t.Helper()
		attackWith(t, g, opp.ID, creeper)
		passPriorityAroundTable(t, g)
		return modePickChoiceFor(g, me.ID)
	}

	// Turn 1: no other creature, so the copy bullet has no legal
	// target and is not offered at all (CR 603.3d).
	c := connect()
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1}) {
		t.Fatalf("counters and draw on offer: %+v", c)
	}
	if c.ModeNotChosen != game.ModeMemoryEver {
		t.Errorf("the prompt says the restriction has no duration: %v", c.ModeNotChosen)
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)

	// Next turn the draw is still used: "that hasn't been chosen" has
	// no duration.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
	c = connect()
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0}) || !slices.Equal(c.ModeUsedIndex, []int{1}) {
		t.Fatalf("the draw is remembered across turns: %+v", c)
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, creeper, game.CounterPlusOne); n != 2 {
		t.Fatalf("two +1/+1 counters: %d", n)
	}

	// Third connection: the copy bullet is the only one left and has
	// no target, so the trigger is removed with no prompt.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
	if c := connect(); c != nil {
		t.Fatalf("an exhausted trigger asks nothing: %+v", c)
	}
	if triggerOnStack(g, creeper) != nil {
		t.Error("and puts nothing on the stack")
	}

	// A flicker makes a new object with no memory of the modes chosen
	// (the 2024-09-20 ruling).
	g.WithWriteLock(func() {
		if err := (Flicker{Target: creeper}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	creeper = findBattlefieldByName(g, "Silent Hallcreeper")
	var memory map[string][]int
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, creeper)
		memory = c.ModesChosen
	})
	if len(memory) != 0 {
		t.Errorf("the returned Hallcreeper is a new object with no memory: %v", memory)
	}
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
	if c := connect(); c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1}) {
		t.Fatalf("the new object chooses afresh: %+v", c)
	}
}

// The third bullet: the Hallcreeper becomes a copy of another creature
// you control, indefinitely (CopyIndefinite).
func TestSilentHallcreeperBecomesACopyOfAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	creeper := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Silent Hallcreeper", OracleID: silentHallcreeperOracle,
		TypeLine: "Enchantment Creature — Horror", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Grizzly Bears", 2, 2)
	attackWith(t, g, opp.ID, creeper)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1, 2}) {
		t.Fatalf("all three on offer with another creature to copy: %+v", c)
	}
	answerMode(t, g, c, me.ID, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the copy bullet asks for its target")
	}
	for _, o := range pick.PickTargetCards {
		if o == creeper {
			t.Error("\"another\": the Hallcreeper is not its own target")
		}
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	var name string
	var power int
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, creeper)
		name, power = c.Effective().Name, c.Effective().Power
	})
	if name != "Grizzly Bears" || power != 2 {
		t.Errorf("the Hallcreeper became a copy of the Bears: %q %d", name, power)
	}
	// Indefinitely: still a copy after the turn ends.
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	g.ReadSnapshot(func() {
		c, _ := battlefieldCard(g, creeper)
		name = c.Effective().Name
	})
	if name != "Grizzly Bears" {
		t.Errorf("the copy has no duration: %q after the turn", name)
	}
}

// Monument to Endurance: one trigger per card discarded, each choosing
// a mode the others have not.
func TestMonumentToEnduranceChoosesADifferentModeForEachDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Monument to Endurance", "Artifact", monumentToEnduranceOracle, 0, 0)
	advanceToMain(t, g)
	emptyHandToLibrary(g, me)
	handCardFull(me, "Pitch A", "Instant", "{U}", "", nil)
	handCardFull(me, "Pitch B", "Instant", "{U}", "", nil)
	hand := me.Hand.Size()
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}

	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 2); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	picks := modePickChoicesFor(g, me.ID)
	if len(picks) != 2 {
		t.Fatalf("two discards, two triggers: %d", len(picks))
	}
	if picks[0].ModeOptionLabel[2] != "Each opponent loses 3 life." {
		t.Errorf("the bullets ride the prompt verbatim: %v", picks[0].ModeOptionLabel)
	}
	answerMode(t, g, picks[0], me.ID, 2)
	answerMode(t, g, modePickChoicesFor(g, me.ID)[0], me.ID, 0)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Life != lives[i+1]-3 {
			t.Errorf("opponent %d lost %d, want 3", i, lives[i+1]-p.Life)
		}
	}
	if me.Life != lives[0] {
		t.Error("the controller loses nothing")
	}
	if me.Hand.Size() != hand-2+1 {
		t.Errorf("two discarded, one drawn: hand %d → %d", hand, me.Hand.Size())
	}

	// A third discard: only the Treasure is left.
	handCardFull(me, "Pitch C", "Instant", "{U}", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	pick := modePickChoiceFor(g, me.ID)
	if pick == nil || !slices.Equal(pick.ModeOptionIndex, []int{1}) {
		t.Fatalf("the third discard offers the Treasure alone: %+v", pick)
	}
}

// --- the activated half, and the registration refusals -------------

const notChosenActivatedOracle = "test-not-chosen-activated"

// Kargan Intimidator's shape ("{1}: Choose one that hasn't been chosen
// this turn —") on a fixture, since no catalogued card has it yet.
func registerNotChosenActivated(t *testing.T) {
	t.Helper()
	registerForTest(t, Spec{
		OracleID: notChosenActivatedOracle,
		Name:     "Not-Chosen Fixture",
		Activated: []ActivatedAbility{{
			Label: "Pay 1 life: Choose one that hasn't been chosen this turn",
			Cost:  PayLife(1),
			Modes: ChooseOneNotChosenThisTurn(
				ModeDoing("Draw a card.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				}),
				ModeDoing("Gain 3 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
				}),
			),
		}},
	})
}

func activationModesOffered(g *game.Game, seat, source uuid.UUID) [][]int {
	var out [][]int
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "activate_ability" || m.Source != source {
			continue
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if json.Unmarshal(m.Params, &p) == nil {
			out = append(out, p.Modes)
		}
	}
	return out
}

func TestActivatedNotChosenRefusesAUsedModeWithNothingPaid(t *testing.T) {
	registerNotChosenActivated(t)
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	src := pushCatalogPermanent(g, me.ID, "Not-Chosen Fixture", "Artifact", notChosenActivatedOracle, false)

	if got := activationModesOffered(g, me.ID, src); len(got) != 2 {
		t.Fatalf("both modes enumerable at first: %v", got)
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Modes: []int{1}}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if me.Life != life-1 {
		t.Fatalf("the cost was paid: %d", me.Life)
	}
	if got := activationModesOffered(g, me.ID, src); len(got) != 1 || !slices.Equal(got[0], []int{0}) {
		t.Errorf("the enumerator never offers the used mode: %v", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Modes: []int{1}}); err == nil {
		t.Fatal("the used mode is refused")
	}
	if me.Life != life-1 {
		t.Errorf("a refused activation paid nothing: life %d, want %d", me.Life, life-1)
	}
	passPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Modes: []int{0}}); err != nil {
		t.Fatalf("the other mode is still available: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := activationModesOffered(g, me.ID, src); len(got) != 0 {
		t.Errorf("with both used, the ability is not enumerable at all: %v", got)
	}
}

func TestRegisterRefusesNotChosenOnASpellAndWithRepeatable(t *testing.T) {
	mustPanic(t, "spell's Modes", func() {
		Register(Spec{
			OracleID: "test-not-chosen-spell",
			Name:     "Not-Chosen Spell",
			Modes:    ChooseOneNotChosenThisTurn(Mode("A."), Mode("B.")),
		})
	})
	mustPanic(t, "Repeatable", func() {
		ms := ChooseOneNotChosen(Mode("A."), Mode("B."))
		ms.Repeatable, ms.Max = true, 2
		Register(Spec{
			OracleID: "test-not-chosen-repeatable",
			Name:     "Not-Chosen Repeatable",
			Triggered: []game.TriggeredAbility{func() game.TriggeredAbility {
				tr := WhenThisEnters("Not-Chosen Repeatable — choose", func(*game.Game, *game.StackItem) error { return nil })
				tr.Modes = ms
				return tr
			}()},
		})
	})
}

// A snapshot carries both halves of the memory: the turn's on the
// tally, the object's on the card.
func TestNotChosenMemorySurvivesARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	greeters := b12Push(g, me.ID, "Gala Greeters", "Creature — Elf Druid", b15GalaGreetersOracle, 1, 1)
	enterCreatureFor(t, g, me.ID, 1)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == greeters {
				g.Battlefield.Cards[i].ModesChosen = map[string][]int{"some ability": {0, 2}}
			}
		}
	})

	snap := g.CaptureSnapshot()
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var decoded game.GameSnapshot
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := decoded.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	var turn, ever []int
	restored.ReadSnapshot(func() {
		c, _ := battlefieldCard(restored, greeters)
		turn = restored.ModesChosenForEffect(b15GalaGreetersTrigger().Modes, game.ModeAbilityOf(c, b15GalaGreetersLabel))
		ever = c.ModesChosen["some ability"]
	})
	if !slices.Equal(turn, []int{1}) {
		t.Errorf("the turn's memory survived the restore: %v", turn)
	}
	if !slices.Equal(ever, []int{0, 2}) {
		t.Errorf("the object's memory survived the restore: %v", ever)
	}
	enterCreatureFor(t, restored, me.ID, 1)
	if c := modePickChoiceFor(restored, me.ID); c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 2}) {
		t.Fatalf("the restored game still refuses the used mode: %+v", c)
	}
}
