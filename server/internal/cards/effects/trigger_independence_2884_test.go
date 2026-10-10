package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_2884_test.go — #2884: "Ask only when the order
// matters" skips the CR 603.3b prompt for a batch whose items are
// pairwise independent, read off each row's declared footprint
// (footprint.go, game/trigger_independence.go). ADR 0018's #2884
// amendment.

const (
	essenceWardenOracle       = "6ca2a89e-7032-4864-b4e9-66f3178f90ab"
	firebrandArcherOracle     = "2e9289d6-dbc6-456d-88cf-d1f534e731d6"
	independenceTokenProbe    = "test-2884-token-probe"
	independenceCountingProbe = "test-2884-counting-probe"
)

func init() {
	// "Whenever you cast a noncreature spell, create a 1/1 colorless
	// Construct token": a declared token step, with a token that has no
	// text.
	Register(Spec{
		OracleID: independenceTokenProbe,
		Name:     "Independence Token Probe",
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Independence Token Probe — create a 1/1 Construct",
				Do(CreateToken{Template: game.Card{
					Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1,
				}, N: 1})),
		},
	})
	// "Whenever you cast a noncreature spell, put a +1/+1 counter on this
	// creature for each creature you control": a count of the board, in
	// a hand-written effect the registry cannot read.
	Register(Spec{
		OracleID: independenceCountingProbe,
		Name:     "Independence Counting Probe",
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Independence Counting Probe — a counter for each creature",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.IsCreature() {
							n++
						}
					}
					return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}

// pushColoredVivi seats Vivi Ornitier as the blue-red card she is, so
// Ugin's "permanent that's one or more colors" can name her.
func pushColoredVivi(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
		OracleID: viviOrnitierOracle, Colors: []string{"U", "R"}, Power: 0, Toughness: 3,
		Owner: owner, Controller: owner,
	})
}

// castColorlessSorcery casts a colorless noncreature spell from the
// active seat's hand: Vivi's and Ugin's triggers both fire on it.
func castColorlessSorcery(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "A Colorless Sorcery", "Sorcery", "", nil)
}

// settleUntilOrderPrompt passes priority until a trigger_order prompt
// opens or the stack is empty, and returns the prompt (or nil).
func settleUntilOrderPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16 && triggerOrderPrompt(g) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	return triggerOrderPrompt(g)
}

// stackSeqOf is the Seq of the triggered item from source on the stack.
func stackSeqOf(t *testing.T, g *game.Game, source uuid.UUID) int {
	t.Helper()
	item := triggerOnStack(g, source)
	if item == nil {
		t.Fatalf("no trigger from %s on the stack", source)
	}
	return int(item.Seq)
}

func TestTheRegistryDeclaresFootprints(t *testing.T) {
	cases := []struct {
		name   string
		oracle string
		row    int
		want   []game.FootprintKind
	}{
		{"Vivi Ornitier", viviOrnitierOracle, 0, []game.FootprintKind{game.FootprintCounterOnSource, game.FootprintDamageEachOpponent}},
		{"Ugin's cast trigger (a Build fill-in)", uginEyeOfTheStormsOracle, 0, []game.FootprintKind{game.FootprintExileTargets}},
		{"Ugin's battlefield trigger", uginEyeOfTheStormsOracle, 1, []game.FootprintKind{game.FootprintExileTargets}},
		{"Guttersnipe", b03GuttersnipeOracle, 0, []game.FootprintKind{game.FootprintDamageEachOpponent}},
		{"Firebrand Archer", firebrandArcherOracle, 0, []game.FootprintKind{game.FootprintDamageEachOpponent}},
		{"Soul Warden", b04SoulWardenOracle, 0, []game.FootprintKind{game.FootprintGainLife}},
		{"Impact Tremors is a hand-written closure", impactTremorsOracle, 0, nil},
		{"a counting closure", independenceCountingProbe, 0, nil},
		{"a token maker", independenceTokenProbe, 0, []game.FootprintKind{game.FootprintCreateToken}},
	}
	for _, c := range cases {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) <= c.row {
			t.Fatalf("%s: %d rows", c.name, len(rows))
		}
		var got []game.FootprintKind
		for _, s := range rows[c.row].Footprint {
			got = append(got, s.Kind)
		}
		if len(got) != len(c.want) {
			t.Errorf("%s: footprint %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: footprint %v, want %v", c.name, got, c.want)
			}
		}
	}
}

func TestFootprintOfRowRefusesWhatItCannotRead(t *testing.T) {
	gain := Do(GainLife{Amount: 1})
	for _, c := range []struct {
		name string
		row  game.TriggeredAbility
	}{
		{"a you may", game.TriggeredAbility{Effect: gain, OptionalPrompt: &game.TriggerOptionalPrompt{}}},
		{"a mode clause", game.TriggeredAbility{Effect: gain, Modes: &game.ModeSpec{}}},
		{"a closure", game.TriggeredAbility{Effect: func(*game.Game, *game.StackItem) error { return nil }}},
		{"an empty Do", game.TriggeredAbility{Effect: Do()}},
		{"a step with no declaration", game.TriggeredAbility{Effect: Do(GainLife{Amount: 1}, DiscardCards{N: 1})}},
		{"a scry with a Then", game.TriggeredAbility{Effect: Do(Scry{N: 1, Then: func(*game.Game) error { return nil }})}},
		{"a constant player", game.TriggeredAbility{Effect: Do(GainLife{Player: uuid.New(), Amount: 1})}},
		{"an exile with no target clause", game.TriggeredAbility{Effect: Do(ExileChosenTarget{})}},
	} {
		if fp := footprintOfRow(c.row); fp != nil {
			t.Errorf("%s: footprint %+v, want none", c.name, fp)
		}
	}
}

// The owner's report: Vivi and Ugin trigger off one colorless spell,
// and Ugin exiles a permanent that is not Vivi. Neither touches what the
// other reads, so the batch goes on the stack in the order it was
// collected and nobody is asked.
func TestViviAndUginOnAnotherPermanentDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vivi := pushColoredVivi(g, me.ID)
	ugin := pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
	victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
	toMain(t, g)
	oppLife := opp.Life

	castColorlessSorcery(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	if ch := triggerOrderPrompt(g); ch != nil {
		t.Fatalf("Vivi and Ugin on another permanent asked for an order: %+v", ch)
	}
	if stackSeqOf(t, g, vivi) > stackSeqOf(t, g, ugin) {
		t.Error("an independent batch goes on in the order it was collected: Vivi, then Ugin")
	}
	settleWithoutPrompts(t, g)
	if !inExile(g, victim) {
		t.Error("Ugin's target is exiled")
	}
	if got := counterCount(g, vivi, game.CounterPlusOne); got != 1 {
		t.Errorf("Vivi has %d +1/+1 counters, want 1", got)
	}
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d → %d, want -1", oppLife, opp.Life)
	}
}

// Ugin targeting Vivi: Ugin writes the object Vivi puts a counter on
// and deals damage from, so the order decides whether the counter lands.
func TestUginTargetingViviAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vivi := pushColoredVivi(g, me.ID)
	pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
	toMain(t, g)

	castColorlessSorcery(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, vivi)
	if ch := triggerOrderPromptFor(g, me.ID); ch == nil || len(ch.TriggerOrderIDs) != 2 {
		t.Fatalf("Ugin targeting Vivi must ask for an order, got %+v", ch)
	}
}

// Two different "deals damage to each opponent" triggers only lower
// opponents' life totals, and losses commute: no prompt.
func TestTwoDamageToEachOpponentTriggersDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
	pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
	oppLife := opp.Life

	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("two damage-to-each-opponent triggers asked for an order: %+v", ch)
	}
	if opp.Life != oppLife-3 {
		t.Errorf("opponent life %d → %d, want -3", oppLife, opp.Life)
	}
}

// The same pair asks when something on the board reacts to life
// changing — here an opponent's Ajani's Pridemate, whose trigger watches
// every life change. The check does not guess whether it would fire.
func TestDamageTriggersAskWhenTheBoardReactsToLifeChanges(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
	pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
	pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)

	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	if ch := settleUntilOrderPrompt(t, g); ch == nil || ch.Chooser != me.ID {
		t.Fatalf("a life-change watcher on the board must keep the prompt, got %+v", ch)
	}
}

// Damage from a lifelink source gains its controller life as it is
// dealt: not on the list, so the batch asks.
func TestDamageFromALifelinkSourceAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
	archer := pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == archer {
			g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "lifelink")
		}
	}

	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("damage from a lifelink source must keep the prompt")
	}
}

// Two different lifegain triggers commute — gains add up in any order.
// With a "whenever you gain life" payoff on the board, each gain
// triggers it between the two, and the batch asks.
func TestLifegainTriggersAskOnlyWithAPayoffOnTheBoard(t *testing.T) {
	for _, withPayoff := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		if withPayoff {
			pushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
		}
		pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
		pushCatalogPermanent(g, me.ID, "Essence Warden", "Creature — Elf Shaman", essenceWardenOracle, false)
		settleUntilOrderPrompt(t, g)
		for triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			settleUntilOrderPrompt(t, g)
		}

		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1) })
		ch := settleUntilOrderPrompt(t, g)
		if withPayoff && ch == nil {
			t.Error("a lifegain payoff on the board must keep the prompt")
		}
		if !withPayoff && ch != nil {
			t.Errorf("two lifegain triggers asked for an order: %+v", ch)
		}
	}
}

// A token maker beside a trigger that counts creatures: the token
// changes the count, and the count is a closure the check cannot read.
// It asks. The same token maker beside a damage trigger does not: the
// token has no text and nothing watches it enter.
func TestATokenMakerAsksBesideACountButNotBesideDamage(t *testing.T) {
	t.Run("beside a count", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
		pushCatalogPermanent(g, me.ID, "Independence Counting Probe", "Creature — Construct", independenceCountingProbe, false)
		castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a token maker beside a creature count must ask")
		}
	})
	t.Run("beside damage", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pushCatalogPermanent(g, me.ID, "Independence Token Probe", "Enchantment", independenceTokenProbe, false)
		pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
		castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
		if ch := settleUntilOrderPrompt(t, g); ch != nil {
			t.Fatalf("a token maker beside damage asked for an order: %+v", ch)
		}
	})
}

// A lethal batch: the opponent is at 1, so Vivi's damage takes them out
// of the game between the resolutions, and with them the permanent Ugin
// targets (CR 800.4a). Where Ugin's exile lands in the batch shows.
func TestDamageThatCanEndAnOpponentAsksBesideAnExileOfTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushColoredVivi(g, me.ID)
	pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
	victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
	toMain(t, g)
	opp.Life = 1

	castColorlessSorcery(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	if ch := triggerOrderPromptFor(g, me.ID); ch == nil {
		t.Fatal("damage that can end the target's controller must keep the prompt")
	}
}

// "Always ask" is unchanged: the independent pair still asks.
func TestAlwaysAskStillAsksForAnIndependentPair(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := g.SetTriggerOrderPreference(me.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	pushCatalogPermanent(g, me.ID, "Guttersnipe", "Creature — Goblin Shaman", b03GuttersnipeOracle, false)
	pushCatalogPermanent(g, me.ID, "Firebrand Archer", "Creature — Human Archer", firebrandArcherOracle, false)
	castCatalogSpell(t, g, "Sandbox Instant", "Instant", "", nil)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("always ask must ask for an independent pair")
	}
}

// APNAP, through a restore: a creature enters and both seats' Wardens
// trigger. The active seat's pair is independent and is not asked; the
// other seat asks for everything ("always"), so the whole drain waits
// on that prompt (CR 603.3b is APNAP). The table is captured and
// restored with the prompt open and four triggers waiting, and the
// restored game still knows the active seat's pair needs no order: the
// answer places it first (lowest on the stack), in the collected order,
// with the other seat's pair above it.
func TestAPNAPIndependentBatchSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	ap, nap := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.SetTriggerOrderPreference(nap.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	apSoul := pushCatalogPermanent(g, ap.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	apEssence := pushCatalogPermanent(g, ap.ID, "Essence Warden", "Creature — Elf Shaman", essenceWardenOracle, false)
	napSoul := pushCatalogPermanent(g, nap.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	napEssence := pushCatalogPermanent(g, nap.ID, "Essence Warden", "Creature — Elf Shaman", essenceWardenOracle, false)
	for i := 0; i < 8 && !(stackFullyEmpty(g) && len(g.PendingChoices) == 0); i++ {
		for triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	apLife, napLife := ap.Life, nap.Life

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(ap.ID, RedGoblinToken(), 1) })
	for i := 0; i < 4 && triggerOrderPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	if ch := triggerOrderPromptFor(g, ap.ID); ch != nil {
		t.Fatalf("the active seat's independent pair was asked for an order: %+v", ch)
	}
	if triggerOrderPromptFor(g, nap.ID) == nil {
		t.Fatal("the other seat always asks")
	}
	if len(g.PendingTriggers) != 4 {
		t.Fatalf("%d triggers waiting, want all four held behind the prompt", len(g.PendingTriggers))
	}

	r := restoreThroughJSON(t, g)
	rAP, rNAP := r.Seats[g.Turn.ActiveSeat], r.Seats[(g.Turn.ActiveSeat+1)%len(r.Seats)]
	if ch := triggerOrderPromptFor(r, rAP.ID); ch != nil {
		t.Fatalf("after a restore the active seat's pair was asked for an order: %+v", ch)
	}
	answerTriggerOrderInOfferedOrder(t, r)
	if triggerOrderPrompt(r) != nil {
		t.Fatal("answering the other seat's prompt must drain the batch")
	}
	soul, essence := stackSeqOf(t, r, apSoul), stackSeqOf(t, r, apEssence)
	if soul > essence {
		t.Error("the active seat's pair goes on in the collected order")
	}
	for _, id := range []uuid.UUID{napSoul, napEssence} {
		if s := stackSeqOf(t, r, id); s < soul || s < essence {
			t.Error("APNAP: the other seat's triggers go on above the active seat's")
		}
	}
	settleWithoutPrompts(t, r)
	if rAP.Life != apLife+2 || rNAP.Life != napLife+2 {
		t.Errorf("life %d/%d → %d/%d, want +2 each", apLife, napLife, rAP.Life, rNAP.Life)
	}
}
