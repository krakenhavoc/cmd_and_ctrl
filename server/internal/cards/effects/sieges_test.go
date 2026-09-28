package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sieges_test.go — #1572, the catalog half of CR 614.12's anchor-word
// choice ("As this enchantment enters, choose Khans or Dragons"). Each
// Siege is tested in BOTH modes, and each mode test also asserts the
// OTHER mode's line is absent: an anchor-word gate that leaked would
// hand the card both abilities, which a one-sided test cannot see.

const (
	frostcliffSiegeOracle   = "e7af31ff-fa1b-4225-b3bc-fde80ac72d0a"
	palaceSiegeOracle       = "4d819d35-5cd3-48f1-a77a-b7c3ff82c62e"
	citadelSiegeOracle      = "40f66e21-5b60-4e42-927d-65397e7ad544"
	barrensteppeSiegeOracle = "228d3307-69a9-45ce-a610-77754ed50877"
	outpostSiegeOracle      = "ebb24fc7-dc71-4712-8c2a-b5920f78e55d"
	frontierSiegeOracle     = "4cfaa5cf-cc3d-49a7-9544-38a8bb7e9ec1"
)

// answerAnchorWord answers the open as-enters prompt with `word`
// through the wire resolver — the prompt is half of what is tested.
func answerAnchorWord(t *testing.T, g *game.Game, chooser uuid.UUID, word string) {
	t.Helper()
	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no as-enters option prompt is open: %+v", g.PendingChoices)
	}
	for i, opt := range c.PickOptions {
		if opt.Label == word {
			if err := g.ResolveOptionPick(c.ID, chooser, i); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("%q is not offered by %v", word, optionWords(c))
}

func optionWords(c *game.PendingChoice) []string {
	out := make([]string, 0, len(c.PickOptions))
	for _, opt := range c.PickOptions {
		out = append(out, opt.Label)
	}
	return out
}

// castSiege casts the named Siege from the active seat's hand, lets it
// resolve, and answers its prompt with `word`.
func castSiege(t *testing.T, g *game.Game, name, oracle, word string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, name, "Enchantment", oracle, nil)
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, active.ID, word)
	if got := g.ChosenOptionOf(id); got != word {
		t.Fatalf("%s chose %q, want %q", name, got, word)
	}
	return id
}

// siegeTriggerKeys is the stack labels of the triggers the permanent
// has RIGHT NOW — the list ADR 0071's gate filters.
func siegeTriggerKeys(g *game.Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			for _, tr := range game.TriggersForCard(c) {
				out = append(out, tr.Key)
			}
		}
	})
	return out
}

func keysMention(keys []string, sub string) bool {
	for _, k := range keys {
		if strings.Contains(k, sub) {
			return true
		}
	}
	return false
}

// --- the shape, catalog-wide ------------------------------------------

// TestEveryAnchorWordGateIsOffered is the guard ChosenIs promises: a
// gate whose word the card's own prompt never offers would ship that
// line switched off forever, and an offered word that gates nothing is
// a choice that does nothing. Every registered spec with an anchor-word
// gate is run through its AsEnters hook and the two lists compared.
func TestEveryAnchorWordGateIsOffered(t *testing.T) {
	seen := 0
	for _, spec := range All() {
		gated := map[string]bool{}
		for _, d := range specDesignations(spec) {
			if d.Kind == game.DesignationChosenOption {
				gated[d.Option] = true
			}
		}
		if len(gated) == 0 {
			continue
		}
		seen++
		if spec.AsEnters == nil {
			t.Errorf("%s gates on an anchor word but asks no as-enters question", spec.Name)
			continue
		}
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushPermanentForTest(g, me.ID, spec.Name, spec.OracleID, "Enchantment")
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					if err := spec.AsEnters(&g.Battlefield.Cards[i], NewContext(g, nil)); err != nil {
						t.Errorf("%s AsEnters: %v", spec.Name, err)
					}
				}
			}
		})
		c := pendingOfKind(g, game.PendingChoiceOptionPick)
		if c == nil {
			t.Errorf("%s asked nothing as it entered", spec.Name)
			continue
		}
		offered := map[string]bool{}
		for _, w := range optionWords(c) {
			offered[w] = true
			if !gated[w] {
				t.Errorf("%s offers %q, which switches no ability on", spec.Name, w)
			}
		}
		for w := range gated {
			if !offered[w] {
				t.Errorf("%s gates a line on %q, which its prompt never offers (%v)", spec.Name, w, optionWords(c))
			}
		}
	}
	if seen < 6 {
		t.Errorf("only %d anchor-word cards registered; the Sieges should all be here", seen)
	}
}

// --- Frostcliff Siege -------------------------------------------------

// TestFrostcliffSiegeAsksJeskaiOrTemurAndHasNeitherLineUntilAnswered —
// the prompt, and the window before the answer.
func TestFrostcliffSiegeAsksJeskaiOrTemurAndHasNeitherLineUntilAnswered(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, active.ID, "Bear", 2, 2)

	id := castCatalogSpell(t, g, "Frostcliff Siege", "Enchantment", frostcliffSiegeOracle, nil)
	passPriorityAroundTable(t, g)

	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatalf("no as-enters prompt: %+v", g.PendingChoices)
	}
	if c.Chooser != active.ID || c.Source != id {
		t.Errorf("the controller chooses for the Siege: chooser %v source %v", c.Chooser, c.Source)
	}
	if got := strings.Join(optionWords(c), ","); got != "Jeskai,Temur" {
		t.Errorf("options %q, want the printed words in printed order", got)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("before the answer the Bear is %d power — Temur is not chosen yet", p)
	}
	if len(siegeTriggerKeys(g, id)) != 0 {
		t.Errorf("before the answer the Siege has triggers %v — Jeskai is not chosen yet", siegeTriggerKeys(g, id))
	}
}

// TestFrostcliffSiegeTemurPumpsYourCreaturesOnly — +1/+0, trample and
// haste on creatures you control, landing the moment Temur is chosen
// (the layer bump), and no Jeskai draw.
func TestFrostcliffSiegeTemurPumpsYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	id := castSiege(t, g, "Frostcliff Siege", frostcliffSiegeOracle, "Temur")

	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"trample", "haste"} {
		if !effectiveAbilitiesContain(t, g, mine, kw) {
			t.Errorf("your Bear lacks %s", kw)
		}
		if effectiveAbilitiesContain(t, g, theirs, kw) {
			t.Errorf("an opponent's creature got %s", kw)
		}
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Bear is %d power, want 2", p)
	}
	if keys := siegeTriggerKeys(g, id); len(keys) != 0 {
		t.Errorf("a Temur Siege has the Jeskai trigger too: %v", keys)
	}
}

// TestFrostcliffSiegeJeskaiDrawsOncePerPlayerConnectedWith — two
// creatures connecting with one player draw ONE card (CR 603.2c), and
// there is no Temur anthem.
func TestFrostcliffSiegeJeskaiDrawsOncePerPlayerConnectedWith(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	castSiege(t, g, "Frostcliff Siege", frostcliffSiegeOracle, "Jeskai")
	if p := effectivePower(t, g, a); p != 2 {
		t.Errorf("a Jeskai Siege pumped the Bear to %d power", p)
	}
	if effectiveAbilitiesContain(t, g, a, "haste") {
		t.Error("a Jeskai Siege granted haste")
	}

	handBefore := me.Hand.Size()
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("two creatures connecting with one player drew %d cards, want 1", got)
	}
}

// TestACopyOfASiegeDoesNotInheritItsChoice — CR 707.2: the chosen word
// is not a copiable value. A token copy of a Temur Siege enters with no
// answer of its own, so the anthem is not doubled by it.
func TestACopyOfASiegeDoesNotInheritItsChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := castSiege(t, g, "Frostcliff Siege", frostcliffSiegeOracle, "Temur")

	g.WithWriteLock(func() {
		if err := (CreateTokenCopy{Controller: me.ID, Copy: id, N: 1}).Apply(NewContext(g, nil)); err != nil {
			t.Fatalf("CreateTokenCopy: %v", err)
		}
	})
	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Frostcliff Siege" && c.InstanceID != id {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no token copy was made")
	}
	if got := g.ChosenOptionOf(copyID); got != "" {
		t.Errorf("the copy inherited the word %q", got)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("the Bear is %d power with one Temur Siege and an unanswered copy, want 3", p)
	}
	// The copy asks its OWN question as it enters (CR 614.12 applies to
	// the copy; the Fate Reforged rulings: a copy of a Siege makes a new
	// choice), and answering it is a separate choice for a separate
	// permanent — the original keeps its word whatever the copy says.
	c := pendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil || c.Source != copyID {
		t.Fatalf("the token copy asked no question of its own: %+v", g.PendingChoices)
	}
	answerAnchorWord(t, g, me.ID, "Jeskai")
	if got := g.ChosenOptionOf(copyID); got != "Jeskai" {
		t.Errorf("the copy chose %q, want Jeskai", got)
	}
	if got := g.ChosenOptionOf(id); got != "Temur" {
		t.Errorf("the copy's answer changed the original's word to %q", got)
	}
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("a Jeskai copy changed the Bear to %d power, want 3", p)
	}
}

// TestASiegeThatLeavesAndReturnsChoosesAgain — CR 400.7 end to end:
// bounced and recast, the Siege asks again and the old word is gone.
func TestASiegeThatLeavesAndReturnsChoosesAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := castSiege(t, g, "Frostcliff Siege", frostcliffSiegeOracle, "Temur")

	if err := g.SacrificePermanent(me.ID, id); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("with the Siege gone the Bear is %d power, want 2", p)
	}
	found := false
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id {
			found = true
			if c.ChosenOption != "" {
				t.Errorf("the Siege in the graveyard still remembers %q", c.ChosenOption)
			}
		}
	}
	if !found {
		t.Fatal("the Siege is not in the graveyard")
	}

	// Recast: a new entry asks again (CR 614.12 fires on each entry).
	again := castCatalogSpell(t, g, "Frostcliff Siege", "Enchantment", frostcliffSiegeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := g.ChosenOptionOf(again); got != "" {
		t.Errorf("a recast Siege arrived already knowing %q", got)
	}
	answerAnchorWord(t, g, me.ID, "Jeskai")
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a recast Jeskai Siege left the Bear at %d power, want 2", p)
	}
}

// --- Palace Siege -----------------------------------------------------

// TestPalaceSiegeKhansReturnsACreatureCardAtUpkeep — the Khans line
// targets a creature card in your graveyard; the Dragons drain does
// not happen.
func TestPalaceSiegeKhansReturnsACreatureCardAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	castSiege(t, g, "Palace Siege", palaceSiegeOracle, "Khans")
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	oppLife := opp.Life

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	advanceToUpkeepOf(t, g, seat)
	answerPickTarget(t, g, dead)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(dead) {
		t.Error("the targeted creature card did not come back to hand")
	}
	if opp.Life != oppLife {
		t.Errorf("a Khans Siege drained: opponent at %d, was %d", opp.Life, oppLife)
	}
}

// TestPalaceSiegeDragonsDrainsEachOpponentAtUpkeep — each opponent
// loses 2, you gain 2 once; no graveyard return is offered.
func TestPalaceSiegeDragonsDrainsEachOpponentAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := castSiege(t, g, "Palace Siege", palaceSiegeOracle, "Dragons")
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	lifeBefore := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lifeBefore[p.ID] = p.Life
	}
	advanceToUpkeepOf(t, g, seat)
	if c := pendingOfKind(g, game.PendingChoicePickTarget); c != nil {
		t.Fatalf("a Dragons Siege asked for a graveyard target: %+v", c)
	}
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		want := lifeBefore[p.ID] - 2
		if p.ID == me.ID {
			want = lifeBefore[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("%s at %d life, want %d", p.Name, p.Life, want)
		}
	}
	if me.Hand.Contains(dead) {
		t.Error("a Dragons Siege returned a creature card")
	}
	if keysMention(siegeTriggerKeys(g, id), "return") {
		t.Error("a Dragons Siege has the Khans trigger")
	}
}

// --- Citadel Siege ----------------------------------------------------

// TestCitadelSiegeKhansPutsTwoCountersAtYourCombat — beginning of
// combat on your turn, two +1/+1 counters on target creature you
// control.
func TestCitadelSiegeKhansPutsTwoCountersAtYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := castSiege(t, g, "Citadel Siege", citadelSiegeOracle, "Khans")

	advanceTo(t, g, game.StepBeginCombat)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, bear); n != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", n)
	}
	if keysMention(siegeTriggerKeys(g, id), "tap target") {
		t.Error("a Khans Siege has the Dragons trigger")
	}
}

// TestCitadelSiegeDragonsTapsTheActiveOpponentsCreature — on an
// opponent's turn, a creature THAT player controls is the only legal
// target; a third player's creature is not. Nothing on your own turn.
func TestCitadelSiegeDragonsTapsTheActiveOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	nextSeat := (seat + 1) % len(g.Seats)
	next := g.Seats[nextSeat]
	third := g.Seats[(seat+2)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	attacker := pushVanillaCreature(g, next.ID, "Their Bear", 2, 2)
	bystander := pushVanillaCreature(g, third.ID, "Bystander", 2, 2)
	castSiege(t, g, "Citadel Siege", citadelSiegeOracle, "Dragons")

	advanceTo(t, g, game.StepBeginCombat)
	if c := pendingOfKind(g, game.PendingChoicePickTarget); c != nil {
		t.Fatalf("a Dragons Siege triggered on its controller's own turn: %+v", c)
	}

	advanceToStepOf(t, g, nextSeat, game.StepBeginCombat)
	c := pickTargetPrompt(t, g)
	legal := map[uuid.UUID]bool{}
	for _, id := range c.PickTargetCards {
		legal[id] = true
	}
	if !legal[attacker] || legal[bystander] || legal[mine] {
		t.Errorf("the legal set is not exactly the active opponent's creatures: %v", c.PickTargetCards)
	}
	answerPickTarget(t, g, attacker)
	passPriorityAroundTable(t, g)
	if !battlefieldTapped(g, attacker) {
		t.Error("the active opponent's creature was not tapped")
	}
	if battlefieldTapped(g, bystander) {
		t.Error("a third player's creature was tapped")
	}
}

// plusOneCounters is the +1/+1 counter count on a battlefield card.
func plusOneCounters(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters["+1/+1"]
		}
	}
	return -1
}

// battlefieldTapped reports whether the battlefield card is tapped.
func battlefieldTapped(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}

// --- Barrensteppe Siege -----------------------------------------------

// TestBarrensteppeSiegeAbzanPutsACounterOnEachOfYourCreatures — end
// step, your creatures only; no edict.
func TestBarrensteppeSiegeAbzanPutsACounterOnEachOfYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	id := castSiege(t, g, "Barrensteppe Siege", barrensteppeSiegeOracle, "Abzan")

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, a); n != 1 {
		t.Errorf("Bear A has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(g, b); n != 1 {
		t.Errorf("Bear B has %d +1/+1 counters, want 1", n)
	}
	if n := plusOneCounters(g, theirs); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
	if keysMention(siegeTriggerKeys(g, id), "sacrifices") {
		t.Error("an Abzan Siege has the Mardu trigger")
	}
}

// TestBarrensteppeSiegeMarduEdictsOnlyAfterACreatureOfYoursDied — the
// intervening if: no death, no trigger; a death under your control,
// and each opponent sacrifices a creature of their choice.
func TestBarrensteppeSiegeMarduEdictsOnlyAfterACreatureOfYoursDied(t *testing.T) {
	t.Run("no death, no trigger", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
		castSiege(t, g, "Barrensteppe Siege", barrensteppeSiegeOracle, "Mardu")

		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if c := sacrificeChoiceFor(g, opp.ID); c != nil {
			t.Fatalf("an edict with no creature of yours dead: %+v", c)
		}
		if !g.Battlefield.Contains(theirs) {
			t.Error("the opponent's creature is gone")
		}
	})
	t.Run("a death, an edict", func(t *testing.T) {
		g := newCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me := g.Seats[seat]
		fodder := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
		keeper := pushVanillaCreature(g, me.ID, "Keeper", 2, 2)
		theirs := map[uuid.UUID]uuid.UUID{}
		for i := 1; i < len(g.Seats); i++ {
			p := g.Seats[(seat+i)%len(g.Seats)]
			theirs[p.ID] = pushVanillaCreature(g, p.ID, "Their Bear", 2, 2)
		}
		id := castSiege(t, g, "Barrensteppe Siege", barrensteppeSiegeOracle, "Mardu")
		if keysMention(siegeTriggerKeys(g, id), "+1/+1 counter") {
			t.Error("a Mardu Siege has the Abzan trigger")
		}
		if err := g.SacrificePermanent(me.ID, fodder); err != nil {
			t.Fatalf("SacrificePermanent: %v", err)
		}
		passPriorityAroundTable(t, g)

		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		if c := sacrificeChoiceFor(g, me.ID); c != nil {
			t.Error("the controller was asked to sacrifice — it is each OPPONENT")
		}
		for pid, cid := range theirs {
			answerSacrifice(t, g, pid, cid)
		}
		passPriorityAroundTable(t, g)
		for _, cid := range theirs {
			if g.Battlefield.Contains(cid) {
				t.Error("an opponent's creature survived the edict")
			}
		}
		if !g.Battlefield.Contains(keeper) {
			t.Error("your own creature was sacrificed")
		}
		if n := plusOneCounters(g, keeper); n != 0 {
			t.Errorf("a Mardu Siege put %d counters on your creature", n)
		}
	})
}

// --- Outpost Siege ----------------------------------------------------

// TestOutpostSiegeKhansImpulsesTheTopCardAtUpkeep — exiled with a PLAY
// grant for the controller; no damage trigger.
func TestOutpostSiegeKhansImpulsesTheTopCardAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := castSiege(t, g, "Outpost Siege", outpostSiegeOracle, "Khans")

	advanceToUpkeepOf(t, g, (seat+1)%len(g.Seats))
	top := seedLibrary(me, "Top Card")[0]
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	perm := exiledPermission(g, top)
	if perm.Player != me.ID {
		t.Errorf("the permission belongs to %v, want the controller", perm.Player)
	}
	if perm.CastOnly {
		t.Error("the printed text says PLAY — a land must not be stranded")
	}
	if keysMention(siegeTriggerKeys(g, id), "damage") {
		t.Error("a Khans Siege has the Dragons trigger")
	}
}

// TestOutpostSiegeDragonsPingsWhenYourCreatureLeaves — a creature you
// control leaving (here, dying) is 1 damage from the Siege to any
// target; an opponent's creature leaving is nothing; no impulse draw.
func TestOutpostSiegeDragonsPingsWhenYourCreatureLeaves(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	mine := pushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Their Fodder", 1, 1)
	id := castSiege(t, g, "Outpost Siege", outpostSiegeOracle, "Dragons")
	if keysMention(siegeTriggerKeys(g, id), "exile the top") {
		t.Error("a Dragons Siege has the Khans trigger")
	}

	if err := g.SacrificePermanent(opp.ID, theirs); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	if c := pendingOfKind(g, game.PendingChoicePickTarget); c != nil {
		t.Fatalf("an opponent's creature leaving triggered the Siege: %+v", c)
	}

	life := opp.Life
	if err := g.SacrificePermanent(me.ID, mine); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-1)
	}
}

// --- Frontier Siege ---------------------------------------------------

// TestFrontierSiegeKhansAddsGreenInEachOfYourMainPhases — {G}{G} in
// the first main phase and again in the second; no fight trigger.
func TestFrontierSiegeKhansAddsGreenInEachOfYourMainPhases(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := castSiege(t, g, "Frontier Siege", frontierSiegeOracle, "Khans")
	if keysMention(siegeTriggerKeys(g, id), "fights") {
		t.Error("a Khans Siege has the Dragons trigger")
	}

	advanceToPrecombatMainOf(t, g, (seat+1)%len(g.Seats))
	if n := countColor(poolColors(g.Seats[(seat+1)%len(g.Seats)]), "G"); n != 0 {
		t.Errorf("an opponent's main phase gave %d green", n)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if n := countColor(poolColors(me), "G"); n != 2 {
		t.Errorf("first main phase: %d green in the pool, want 2 (%v)", n, poolColors(me))
	}

	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if n := countColor(poolColors(me), "G"); n != 2 {
		t.Errorf("second main phase: %d green in the pool, want 2 (%v)", n, poolColors(me))
	}
}

func countColor(colors []string, want string) int {
	n := 0
	for _, c := range colors {
		if c == want {
			n++
		}
	}
	return n
}

// TestFrontierSiegeDragonsLetsAnEnteringFlierFight — a creature you
// control with flying enters, you say yes, it fights target creature
// you don't control; a creature without flying triggers nothing.
func TestFrontierSiegeDragonsLetsAnEnteringFlierFight(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castSiege(t, g, "Frontier Siege", frontierSiegeOracle, "Dragons")
	manaBefore := len(poolColors(me))

	ground := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ground Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	emitETBForTest(g, me.ID, ground)
	if c := pendingOfKind(g, game.PendingChoiceTriggerPrompt); c != nil {
		t.Fatalf("a creature without flying triggered the Siege: %+v", c)
	}

	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Roc", TypeLine: "Creature — Bird",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying"},
	})
	emitETBForTest(g, me.ID, flier)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("the 4/4 flier fought the 2/2 and it survived")
	}
	if c, ok := battlefieldCard(g, flier); !ok || c.DamageMarked != 2 {
		t.Errorf("the flier should have taken 2 damage back: %+v", c)
	}
	if got := len(poolColors(me)); got != manaBefore {
		t.Errorf("a Dragons Siege added mana: %d tokens, was %d", got, manaBefore)
	}
}

// emitETBForTest announces a pushed permanent's entry the way the
// entry pipeline would, so an "enters" trigger sees it.
func emitETBForTest(g *game.Game, controller, id uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: controller, CardID: id})
	})
}
