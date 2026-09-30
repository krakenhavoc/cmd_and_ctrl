package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vivivoltron_cards_test.go — the five cards added for the ViviVoltron
// deck request (#1640): Slip Out the Back, Diviner's Portent, Ancient
// Silver Dragon, Return the Favor and March of Swirling Mist.

const (
	slipOutTheBackOracle      = "82506e51-3fe1-446f-83c9-69e67583cefc"
	divinersPortentOracle     = "119585c7-ddfa-47ed-b2f8-488ebc156222"
	ancientSilverDragonOracle = "939556c8-e363-4453-a624-ca52b151467e"
	returnTheFavorOracle      = "30377bd5-99ab-4888-9ea7-e5a069b527e8"
	marchOfSwirlingMistOracle = "debc69ea-372a-4720-838a-16856cd50b07"
)

// vvPhasedOutCard is the phased-out copy of `id`, read off the holding
// slice (the card is not on the battlefield while phased out).
func vvPhasedOutCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.PhasedOutCardsForEffect() {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	return out, found
}

// vvRolls is every d20 result the log holds, in order.
func vvRolls(g *game.Game) []int {
	var out []int
	for _, ev := range randomEvents(g, game.EventRollDie) {
		out = append(out, ev.Amount)
	}
	return out
}

// --- Slip Out the Back ------------------------------------------------

// The counter goes on, then the creature phases out carrying it, and
// comes back with it at its controller's next untap step.
func TestSlipOutTheBackCountersThenPhasesOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVoltronCreature(g, me, "My Bear")

	castCatalogSpell(t, g, "Slip Out the Back", "Instant", slipOutTheBackOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	c, ok := vvPhasedOutCard(g, bear)
	if !ok {
		t.Fatal("the target creature should have phased out")
	}
	if c.Counters["+1/+1"] != 1 {
		t.Errorf("+1/+1 counters on the phased-out creature = %d, want 1", c.Counters["+1/+1"])
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a phased-out creature is not on the battlefield (CR 702.26b)")
	}

	passTurnsTo(t, g, 0)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature phases back in at its controller's untap step")
	}
	if p := currentPower(t, g, bear); p != 3 {
		t.Errorf("power after phasing in = %d, want 3 — the counter rode along", p)
	}
}

// --- Diviner's Portent ------------------------------------------------

// With fourteen cards in hand the total is at least 15 whatever the
// die says: scry X, then draw X.
func TestDivinersPortentScriesThenDrawsOnFifteenOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for me.Hand.Size() < 14 {
		me.Hand.PushTop(game.NewCard("filler", me.ID))
	}

	b12PlayFromHand(t, g, "Diviner's Portent", "Instant", divinersPortentOracle,
		game.CastSpellParams{XValue: 2})
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if rolls := vvRolls(g); len(rolls) != 1 {
		t.Fatalf("rolled %v, want one d20", rolls)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a total of 15 or more scries first")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry %d, want X = 2", len(c.ScryCards))
	}
	if me.Hand.Size() != handBefore {
		t.Error("the draw waits for the scry to be answered")
	}
	answerScryKeepAll(t, g, me.ID)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand after scry then draw = %d, want %d", got, handBefore+2)
	}
}

// With an empty hand a roll of 14 or less draws X and scries nothing.
// The roll is random, so fresh games are tried until one lands in
// that row (70% each); a roll of 15+ must take the other row.
func TestDivinersPortentDrawsOnFourteenOrLess(t *testing.T) {
	for attempt := 0; attempt < 40; attempt++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		me.Hand.Cards = nil

		b12PlayFromHand(t, g, "Diviner's Portent", "Instant", divinersPortentOracle,
			game.CastSpellParams{XValue: 3})
		passPriorityAroundTable(t, g)

		rolls := vvRolls(g)
		if len(rolls) != 1 {
			t.Fatalf("rolled %v, want one d20", rolls)
		}
		if rolls[0] >= 15 {
			if scryChoiceFor(g, me.ID) == nil {
				t.Fatalf("a roll of %d with an empty hand should scry", rolls[0])
			}
			continue
		}
		if scryChoiceFor(g, me.ID) != nil {
			t.Fatalf("a total of %d is the draw row, not the scry row", rolls[0])
		}
		if got := me.Hand.Size(); got != 3 {
			t.Fatalf("hand after drawing X = %d, want 3", got)
		}
		return
	}
	t.Fatal("never rolled 14 or less in 40 games")
}

// --- Ancient Silver Dragon -------------------------------------------

func TestAncientSilverDragonDrawsTheRollAndLiftsHandSize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 25; i++ {
		me.Library.PushTop(game.NewCard("filler", me.ID))
	}
	dragon := pushCatalogPermanent(g, me.ID, "Ancient Silver Dragon", "Creature — Elder Dragon",
		ancientSilverDragonOracle, false)
	if !hasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("Ancient Silver Dragon has flying")
	}
	handBefore := me.Hand.Size()

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: dragon,
			Target: g.Seats[1].ID, Amount: 8, Combat: true})
	})
	if len(vvRolls(g)) != 0 {
		t.Fatal("rolled before the trigger resolved")
	}
	passPriorityAroundTable(t, g)

	rolls := vvRolls(g)
	if len(rolls) != 1 {
		t.Fatalf("rolled %v, want one d20", rolls)
	}
	if got := me.Hand.Size(); got != handBefore+rolls[0] {
		t.Errorf("hand = %d, want %d (drew the roll of %d)", got, handBefore+rolls[0], rolls[0])
	}
	if me.MaxHandSize != game.NoMaxHandSize {
		t.Errorf("max hand size = %d, want no maximum", me.MaxHandSize)
	}

	// "For the rest of the game": the grant outlives the Dragon.
	g.WithWriteLock(func() {
		for i, c := range g.Battlefield.Cards {
			if c.InstanceID == dragon {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	if me.MaxHandSize != game.NoMaxHandSize {
		t.Error("the hand-size grant lapsed when the Dragon left")
	}
}

// --- Return the Favor -------------------------------------------------

// The copy bullet's clause: instants, sorceries and abilities; never a
// creature spell.
func TestReturnTheFavorCopyClause(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bears := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)

	spec, ok := Lookup(returnTheFavorOracle)
	if !ok || spec.Modes == nil || len(spec.Modes.Options) != 2 {
		t.Fatal("Return the Favor prints two Spree bullets")
	}
	for i, o := range spec.Modes.Options {
		if o.Cost != "{1}" {
			t.Errorf("bullet %d costs %q, want {1}", i, o.Cost)
		}
	}
	offered := taLegalForSpec(g, me.ID, spec.Modes.Options[0].Targets)
	if !offered[bolt] {
		t.Error("an instant spell is a legal copy target")
	}
	if !offered[ability] {
		t.Error("an opponent's triggered ability is a legal copy target")
	}
	if offered[bears] {
		t.Error("a creature spell is not an instant or sorcery")
	}
}

// Copying a spell: the copy is the caster's, and may be re-aimed.
func TestReturnTheFavorCopiesASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	castModal(t, g, "Return the Favor", "Instant", returnTheFavorOracle,
		[]int{0}, []game.TargetRef{modeRef(game.TargetCard, bolt, 0, 0)})

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victimA); got != 37 {
		t.Errorf("the original Bolt's target life = %d, want 37", got)
	}
	if got := lifeOf(g, victimB); got != 37 {
		t.Errorf("the copy's new target life = %d, want 37", got)
	}
}

// Copying an ability: the copy goes on the stack under the caster's
// control (CR 707.10b).
func TestReturnTheFavorCopiesAnAbility(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	ability := taAnnounce(t, g, opp.ID, src, "Their Rock — do a thing", nil, nil)

	castModal(t, g, "Return the Favor", "Instant", returnTheFavorOracle,
		[]int{0}, []game.TargetRef{modeRef(game.TargetCard, ability, 0, 0)})
	acPassUntilCopy(t, g)

	cp := acCopyOnStack(g)
	if cp == nil {
		t.Fatal("no copy of the ability on the stack")
	}
	if cp.Kind != game.StackItemTriggered {
		t.Errorf("a copy of a triggered ability is a triggered ability, got %v", cp.Kind)
	}
	if cp.Controller != me.ID {
		t.Error("the copy is controlled by Return the Favor's controller")
	}
}

// The change bullet is Bolt Bend's: mandatory, one slot.
func TestReturnTheFavorChangesTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	castModal(t, g, "Return the Favor", "Instant", returnTheFavorOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, bolt, 0, 0)})
	passPriorityAroundTable(t, g)

	prompt := latestRetarget(g, me)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("the change is mandatory: min %d, want 1", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if lifeOf(g, victimA) != 40 || lifeOf(g, victimB) != 37 {
		t.Errorf("the Bolt should have moved: A=%d B=%d", lifeOf(g, victimA), lifeOf(g, victimB))
	}
	if acCopiesOnStack(g) != 0 || lifeOf(g, me) != 40 {
		t.Error("the change bullet alone copies nothing")
	}
}

// Both bullets on one Bolt: the copy goes to a third seat, and the
// original is moved to a second one. Nobody is hit twice and the
// original target is spared.
func TestReturnTheFavorBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, victimA, victimB, victimC := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	castModal(t, g, "Return the Favor", "Instant", returnTheFavorOracle,
		[]int{0, 1}, []game.TargetRef{
			modeRef(game.TargetCard, bolt, 0, 0),
			modeRef(game.TargetCard, bolt, 1, 0),
		})

	answeredCopy, answeredChange := false, false
	for i := 0; i < 16 && !(answeredCopy && answeredChange); i++ {
		if p := latestPickTarget(g, me); p != nil && !answeredCopy {
			if err := g.ResolvePickTarget(p.ID, me,
				game.TargetRef{Kind: game.TargetPlayer, ID: victimC}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answeredCopy = true
			continue
		}
		if p := latestRetarget(g, me); p != nil && !answeredChange {
			if err := g.ResolveRetarget(p.ID, me,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
				t.Fatalf("ResolveRetarget: %v", err)
			}
			answeredChange = true
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !answeredCopy || !answeredChange {
		t.Fatalf("both bullets should ask: copy %v, change %v", answeredCopy, answeredChange)
	}
	passPriorityAroundTable(t, g)

	if got := lifeOf(g, victimA); got != 40 {
		t.Errorf("the original target life = %d, want 40 — the Bolt was moved", got)
	}
	if got := lifeOf(g, victimB); got != 37 {
		t.Errorf("the moved Bolt's target life = %d, want 37", got)
	}
	if got := lifeOf(g, victimC); got != 37 {
		t.Errorf("the copy's target life = %d, want 37", got)
	}
}

// --- March of Swirling Mist -------------------------------------------

func TestMarchOfSwirlingMistPhasesOutXCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVoltronCreature(g, me, "Mine")
	b := pushVoltronCreature(g, opp, "Theirs")
	c := pushVoltronCreature(g, opp, "Left Alone")

	b12PlayFromHand(t, g, "March of Swirling Mist", "Instant", marchOfSwirlingMistOracle,
		game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a},
			{Kind: game.TargetCard, ID: b},
		}})
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if _, ok := vvPhasedOutCard(g, id); !ok {
			t.Errorf("%v should have phased out", id)
		}
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature stays")
	}
}

// "Up to X targets" (#1738): fewer than X is fine; more is not.
func TestMarchOfSwirlingMistTakesUpToXTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushVoltronCreature(g, me, "Mine")
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "March of Swirling Mist", TypeLine: "Instant",
		OracleID: marchOfSwirlingMistOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}})
	if err != nil {
		t.Fatalf("one target with X = 2 is \"up to X\": %v", err)
	}
	spec, _ := Lookup(marchOfSwirlingMistOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("March of Swirling Mist keeps only its discount caveat: %q %v", spec.Completeness, spec.Caveats)
	}
}
