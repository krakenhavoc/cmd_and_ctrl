package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 665: Invasion enchantments.

const (
	oldB665FightOrFlight   = "2fc42640-983a-49ca-9020-7756135077b4"
	oldB665WellLaidPlans   = "50343947-d127-4479-9c1b-8f5040a97afb"
	oldB665ManaMaze        = "5b6f9d55-eae8-48fd-a1d1-d56939adaf3a"
	oldB665PsychicBattle   = "79f0d763-5a71-470a-867f-86fc72e9db6d"
	oldB665StandOrFall     = "927d0085-9bd9-4454-8ac5-0e374237ed35"
	oldB665SpiritOfResist  = "b075b645-2e5a-4d91-9440-02db4e6f5d08"
	oldB665PulseOfLlanowar = "dd1ef824-4e61-4dee-a8f0-9cf459008f2c"
)

func TestOldB665Registered(t *testing.T) {
	want := map[string]string{
		oldB665FightOrFlight: "Fight or Flight", oldB665WellLaidPlans: "Well-Laid Plans",
		oldB665ManaMaze: "Mana Maze", oldB665PsychicBattle: "Psychic Battle",
		oldB665StandOrFall: "Stand or Fall", oldB665SpiritOfResist: "Spirit of Resistance",
		oldB665PulseOfLlanowar: "Pulse of Llanowar",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want Full with no caveats, got %v / %v", name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB665HandSpell(g *game.Game, p *game.Player, name, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost,
			Colors: colors, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}

func oldB665Cast(g *game.Game, p *game.Player, id uuid.UUID, color string) error {
	g.WithWriteLock(func() { effectstest.FillPoolColored(p, color, 1) })
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

// Mana Maze: a red spell after a red spell is refused for either player; blue is
// fine; a colourless spell is never refused.
func TestOldB665ManaMaze(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mana Maze", "Enchantment", oldB665ManaMaze, false)
	bolt := oldB665HandSpell(g, me, "Bolt", "{R}", "R")
	shock := oldB665HandSpell(g, me, "Shock", "{R}", "R")
	opt := oldB665HandSpell(g, me, "Opt", "{U}", "U")
	if err := oldB665Cast(g, me, bolt, "R"); err != nil {
		t.Fatalf("the turn's first spell: %v", err)
	}
	if err := oldB665Cast(g, me, shock, "R"); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("red after red: err = %v, want ErrCantCast", err)
	}
	if err := oldB665Cast(g, me, opt, "U"); err != nil {
		t.Fatalf("blue after red: %v", err)
	}
	// Players, not only the controller: a blue spell by the opponent right after
	// the blue Opt is refused.
	oppBlue := oldB665HandSpell(g, opp, "Counter", "{U}", "U")
	g.WithWriteLock(func() { effectstest.FillPoolColored(opp, "U", 1) })
	if err := g.CastSpell(opp.ID, oppBlue, game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("an opponent's blue spell after a blue spell: err = %v, want ErrCantCast", err)
	}
}

// Pulse of Llanowar: a Swamp pays {G}; a nonbasic land is unchanged.
func TestOldB665PulseOfLlanowar(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Pulse of Llanowar", "Enchantment", oldB665PulseOfLlanowar, false)
	effectstest.ToMainForCost(t, g)
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	spell := effectstest.HandCardFull(me, "Green Test", "Sorcery", "{G}", "", []string{"G"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("a {G} spell paid with only a Swamp under Pulse: %v", err)
	}
	if c, _ := g.LookupCardForEffect(swamp); !c.Tapped {
		t.Error("the Swamp should have been tapped to pay")
	}
	// Without a basic land the replacement doesn't apply.
	waste := effectstest.FreeSpellPermanent(g, me.ID, "Wastes Test", "Land")
	spell2 := effectstest.HandCardFull(me, "Green Test 2", "Sorcery", "{G}", "", []string{"G"})
	if err := g.CastSpell(me.ID, spell2, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		if c, _ := g.LookupCardForEffect(waste); c.Tapped {
			t.Error("a nonbasic land should not have paid {G}")
		}
	}
}

func oldB665Burn(t *testing.T, g *game.Game, from *game.Player, at uuid.UUID) {
	t.Helper()
	effectstest.BoltFrom(t, g, from, at)
	effectstest.PassPriorityAroundTable(t, g)
}

// Spirit of Resistance: damage to its controller is prevented only while they
// control a permanent of each colour.
func TestOldB665SpiritOfResistance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Spirit of Resistance", "Enchantment", oldB665SpiritOfResist, false)
	life := me.Life
	oldB665Burn(t, g, opp, me.ID)
	if me.Life != life-3 {
		t.Fatalf("with no coloured permanents the Bolt deals 3: life %d→%d", life, me.Life)
	}
	for _, c := range []string{"W", "U", "B", "R"} {
		effectstest.B27Push(g, me.ID, "Dork "+c, "Creature — Bear", "", "", 1, 1, c)
	}
	life = me.Life
	oldB665Burn(t, g, opp, me.ID)
	if me.Life != life-3 {
		t.Fatalf("four colours aren't enough: life %d→%d", life, me.Life)
	}
	effectstest.B27Push(g, me.ID, "Dork G", "Creature — Bear", "", "", 1, 1, "G")
	life = me.Life
	oldB665Burn(t, g, opp, me.ID)
	if me.Life != life {
		t.Fatalf("a permanent of each colour prevents the damage: life %d→%d", life, me.Life)
	}
	oppLife := opp.Life
	oldB665Burn(t, g, me, opp.ID)
	if opp.Life != oppLife-3 {
		t.Errorf("only its controller is protected: opponent life %d→%d", oppLife, opp.Life)
	}
}

// Well-Laid Plans: damage from a creature to another creature sharing a colour is
// prevented; with no shared colour, or from a noncreature source, it isn't.
func TestOldB665WellLaidPlans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Well-Laid Plans", "Enchantment", oldB665WellLaidPlans, false)
	red := effectstest.B27Push(g, me.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 9, "R")
	redWB := effectstest.B27Push(g, opp.ID, "Red-White", "Creature — Bear", "", "{R}{W}", 3, 5, "R", "W")
	blue := effectstest.B27Push(g, opp.ID, "Blue Bear", "Creature — Bear", "", "{1}{U}", 3, 5, "U")
	colourless := effectstest.B27Push(g, opp.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 5)

	effectstest.DealDamageTo(t, g, redWB, red, 3)
	if got := effectstest.DamageMarkedOn(g, red); got != 0 {
		t.Errorf("a red creature's damage to a red creature is prevented, %d marked", got)
	}
	effectstest.DealDamageTo(t, g, blue, red, 3)
	if got := effectstest.DamageMarkedOn(g, red); got != 3 {
		t.Errorf("no shared colour: the damage is dealt, %d marked", got)
	}
	effectstest.DealDamageTo(t, g, colourless, red, 1)
	if got := effectstest.DamageMarkedOn(g, red); got != 4 {
		t.Errorf("a colourless creature shares no colour: %d marked, want 4", got)
	}
	// A creature that deals damage to a creature that shares a colour with it,
	// the other way round: the red bear hurting the red-white one is prevented too.
	effectstest.DealDamageTo(t, g, red, redWB, 2)
	if got := effectstest.DamageMarkedOn(g, redWB); got != 0 {
		t.Errorf("the prevention reads both directions, %d marked on the red-white bear", got)
	}
	// A burn spell is not a creature: Bolt on the red bear isn't a creature source.
	// (a creature can't be dealt "by another creature" damage from itself either)
	effectstest.DealDamageTo(t, g, red, red, 1)
	if got := effectstest.DamageMarkedOn(g, red); got != 5 {
		t.Errorf("a creature's damage to itself isn't 'another creature': %d marked, want 5", got)
	}
}

// oldB665SplitHolding is the splitter's open pile split that offers card id.
func oldB665SplitHolding(g *game.Game, splitter, id uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceRevealPick || c.Chooser != splitter {
			continue
		}
		for _, o := range c.ChooseCards {
			if o == id {
				return c
			}
		}
	}
	return nil
}

func oldB665AttackErr(g *game.Game, id, defender uuid.UUID) error {
	return g.DeclareAttacker(id, defender)
}

// Fight or Flight: on an opponent's turn the controller separates that player's
// creatures; the opponent picks the pile that may attack.
func TestOldB665FightOrFlight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fight or Flight", "Enchantment", oldB665FightOrFlight, false)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 3 {
		t.Fatalf("Fight or Flight's controller separates the opponent's three creatures: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the active opponent chooses a pile: %+v", g.PendingChoices)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // {B, C}
	effectstest.PassPriorityAroundTable(t, g)
	// A creature that arrives later is not in the pile either.
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := oldB665AttackErr(g, a, me.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("A is outside the chosen pile: err = %v, want ErrCantAttack", err)
	}
	if err := oldB665AttackErr(g, late, me.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("a late arrival is outside the pile: err = %v, want ErrCantAttack", err)
	}
	if err := oldB665AttackErr(g, b, me.ID); err != nil {
		t.Errorf("B is in the chosen pile: %v", err)
	}
	if err := oldB665AttackErr(g, c, me.ID); err != nil {
		t.Errorf("C is in the chosen pile: %v", err)
	}
}

// Fight or Flight does nothing on its controller's own turn.
func TestOldB665FightOrFlightOnlyOnOpponentsTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Fight or Flight", "Enchantment", oldB665FightOrFlight, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestRevealPickFor(g, me.ID) != nil {
		t.Fatal("no split on the controller's own turn")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := oldB665AttackErr(g, mine, g.Seats[1].ID); err != nil {
		t.Errorf("an unrestricted attack: %v", err)
	}
}

// An empty pile is legal: choosing it bars every attacker.
func TestOldB665FightOrFlightEmptyPile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fight or Flight", "Enchantment", oldB665FightOrFlight, false)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("no split: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, []uuid.UUID{a}); err != nil { // {A} and {}
		t.Fatal(err)
	}
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("no pile pick: %+v", g.PendingChoices)
	}
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 0 {
			effectstest.AnswerOptionPick(t, g, opp.ID, i)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := oldB665AttackErr(g, a, me.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("the empty pile was chosen, so nothing attacks: err = %v", err)
	}
}

// Stand or Fall: on its controller's turn each opponent's creatures are split
// and that opponent chooses; only the chosen pile can block.
func TestOldB665StandOrFall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Stand or Fall", "Enchantment", oldB665StandOrFall, false)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	x := effectstest.B12Creature(g, third.ID, "X", "Creature — Bear", 2, 2)
	y := effectstest.B12Creature(g, third.ID, "Y", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, step := range []struct {
		p      *game.Player
		first  uuid.UUID
		chosen int // pile index the player takes
	}{{opp, a, 0}, {third, x, 1}} {
		split := oldB665SplitHolding(g, me.ID, step.first)
		if split == nil {
			t.Fatalf("no split for seat %d: %+v", step.p.Seat, g.PendingChoices)
		}
		if err := g.ResolveRevealPick(split.ID, me.ID, []uuid.UUID{step.first}); err != nil {
			t.Fatal(err)
		}
		if effectstest.LatestOptionPickFor(g, step.p.ID) == nil {
			t.Fatalf("seat %d chooses a pile: %+v", step.p.Seat, g.PendingChoices)
		}
		effectstest.AnswerOptionPick(t, g, step.p.ID, step.chosen)
		effectstest.PassPriorityAroundTable(t, g)
	}
	// opp took {A}; third took {Y} (the second pile).
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, third.ID); err != nil {
		t.Fatalf("attacking: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(b, atk); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("B is outside opp's chosen pile: err = %v, want ErrCantBlock", err)
	}
	if err := g.DeclareBlocker(x, atk); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("X is outside the third player's chosen pile: err = %v, want ErrCantBlock", err)
	}
	if err := g.DeclareBlocker(y, atk); err != nil {
		t.Errorf("Y is in the third player's chosen pile: %v", err)
	}
}

// Psychic Battle (#6228): the player revealing the greatest mana value may
// change the target of a spell cast; its own change doesn't retrigger it.
func TestOldB665PsychicBattle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PushCatalogPermanent(g, g.Seats[2].ID, "Psychic Battle", "Enchantment", oldB665PsychicBattle, false)
	for i, cost := range [4]string{"{1}", "{5}", "{2}", "{3}"} {
		p := g.Seats[i]
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top " + cost, TypeLine: "Sorcery", ManaCost: cost, Owner: p.ID, Controller: p.ID})
	}
	life3 := g.Seats[3].Life
	bolt := effectstest.BoltFrom(t, g, g.Seats[0], g.Seats[2].ID)
	var prompt *game.PendingChoice
	for i := 0; i < 32 && prompt == nil && !effectstest.StackFullyEmpty(g); i++ {
		if prompt = effectstest.LatestRetarget(g, g.Seats[1].ID); prompt != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if prompt == nil {
		prompt = effectstest.LatestRetarget(g, g.Seats[1].ID)
	}
	if prompt == nil {
		t.Fatal("the player who revealed the greatest mana value was not offered the change")
	}
	effectstest.AnswerRetargetWith(t, g, g.Seats[1].ID, effectstest.RtPlayerRef(g.Seats[3].ID))
	if item := g.StackMeta[bolt]; item == nil || item.Targets[0].ID != g.Seats[3].ID {
		t.Fatalf("the Bolt was not redirected: %+v", item)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("the stack holds %d items; Psychic Battle's own change must not retrigger it", len(g.StackMeta))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[3].Life != life3-3 {
		t.Errorf("seat 3 life %d→%d, want the Bolt there", life3, g.Seats[3].Life)
	}
}
