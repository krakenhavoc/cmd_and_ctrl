package game

import (
	"testing"

	"github.com/google/uuid"
)

// s82_gap2_b05_seams_test.go — the engine seams S82 GAP2 batch 5 built
// (#6088, #6094): a spell's colour in layer 5 (spell_colors.go), an
// ability bundle a permanent spell gains (spell_grants.go), a copy with a
// specified new target (spell_copy_new_target.go, CR 707.10e), the "other
// creatures" form of the as-though-haste static, and the "until your next
// turn" spend rider. The printed cards are pinned in
// cards/effects/tests/ru/s82_gap2_b05_test.go.

func stackCardColors(g *Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Stack.Cards {
			if c.InstanceID == id {
				out = c.EffectiveColors()
			}
		}
	})
	return out
}

// CR 613.1e, 400.7a: a creature spell made green is green on the stack and
// as the creature it becomes, until end of turn.
func TestASpellColourChangeAppliesOnTheStackAndOnThePermanent(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	item := castSpellForControlTest(t, g, me, "Creature — Elf")
	g.WithWriteLock(func() {
		if !g.SetSpellColorsForEffect(uuid.New(), item.ID, []string{"G"}, g.UntilEndOfTurnDuration(), "test — it becomes green") {
			t.Fatal("SetSpellColorsForEffect refused a spell on the stack")
		}
	})
	if got := stackCardColors(g, item.ID); len(got) != 1 || got[0] != "G" {
		t.Fatalf("the spell is %v, want green", got)
	}
	resolveTop(t, g)
	if c := layeredBattlefieldCard(t, g, item.ID); len(c.EffectiveColors()) != 1 || c.EffectiveColors()[0] != "G" {
		t.Fatalf("the creature is %v, want green", c.EffectiveColors())
	}
	advancePastThisTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := layeredBattlefieldCard(t, g, item.ID); len(c.EffectiveColors()) != 0 {
		t.Errorf("the creature is still %v after the turn ended", c.EffectiveColors())
	}
}

// Nothing is registered for an object that is not a spell on the stack.
func TestASpellColourChangeRefusesWhatIsNotASpell(t *testing.T) {
	g := newActiveGame(t)
	before := len(g.ScopedEffects)
	g.WithWriteLock(func() {
		if g.SetSpellColorsForEffect(uuid.New(), uuid.New(), []string{"W"}, IndefiniteDuration(), "test") {
			t.Error("a colour change was registered for no spell")
		}
		if g.GrantAbilitiesToPermanentSpellForEffect(uuid.New(), uuid.New(), []string{"x"}, IndefiniteDuration(), "test") {
			t.Error("a grant was registered for no spell")
		}
	})
	if len(g.ScopedEffects) != before {
		t.Error("a record was written")
	}
}

// A bundle is refused on an instant spell: the stack step would not apply
// it, and an instant's abilities function on the stack (CR 113.6).
func TestABundleGrantRefusesAnInstantSpell(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	item := castSpellForControlTest(t, g, me, "Instant")
	g.WithWriteLock(func() {
		if g.GrantAbilitiesToPermanentSpellForEffect(uuid.New(), item.ID, []string{"x"}, IndefiniteDuration(), "test") {
			t.Error("a bundle was granted to an instant spell")
		}
	})
}

// The stack pin admits a colour and a bundle, and still nothing else.
func TestTheStackPinAdmitsColourAndBundleOnly(t *testing.T) {
	pin := []AffectedObject{PinStackObject(uuid.New(), 1)}
	for _, m := range []Mod{SetColorsMod("W"), GrantAbilitiesMod("x")} {
		if p := stackPinProblem(pin, []Mod{m}); p != "" {
			t.Errorf("%s refused: %s", m.Kind, p)
		}
	}
	if stackPinProblem(pin, []Mod{{Kind: ModModifyPT, Power: 1}}) == "" {
		t.Error("a P/T change was admitted on a stack pin")
	}
}

// CR 707.10e: the copy's every target is the specified player or object,
// and a copy whose specified target is illegal is not created.
func TestACopyWithASpecifiedTargetIsMadeOnlyWhenItIsLegal(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	retargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := pushRetargetCreature(g, other, "Bear")
	bolt := castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetPlayer, ID: opp.ID}})
	if sole, ok := g.SpellSoleTargetForEffect(bolt.ID); !ok || sole.ID != opp.ID {
		t.Fatalf("sole target %v %v, want the opponent", sole, ok)
	}
	stack := len(g.Stack.Cards)
	g.WithWriteLock(func() {
		if !g.CopySpellWithNewTargetForEffect(bolt.ID, me.ID, TargetRef{Kind: TargetCard, ID: bear}) {
			t.Error("a copy at a creature the spell can target was refused")
		}
	})
	if len(g.Stack.Cards) != stack+1 {
		t.Fatalf("stack %d -> %d, want one copy", stack, len(g.Stack.Cards))
	}
	top := g.Stack.Cards[len(g.Stack.Cards)-1].InstanceID
	if ts := g.StackMeta[top].Targets; len(ts) != 1 || ts[0].Kind != TargetCard || ts[0].ID != bear {
		t.Errorf("the copy targets %v, want the Bear", ts)
	}

	murder := castTargeted(t, g, me, retargetOneOracle, []TargetRef{{Kind: TargetCard, ID: bear}})
	stack = len(g.Stack.Cards)
	g.WithWriteLock(func() {
		if g.CopySpellWithNewTargetForEffect(murder.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: other.ID}) {
			t.Error("a creature-only spell was copied at a player")
		}
	})
	if len(g.Stack.Cards) != stack {
		t.Error("an illegal copy reached the stack")
	}
}

// "Other creatures you control": the static reaches the controller's other
// sick creatures and not its own source.
func TestActivateAsThoughHasteOthersSkipsItsSource(t *testing.T) {
	prev, prevOthers := CatalogActivateAsThoughHaste, CatalogActivateAsThoughHasteOthersOnly
	t.Cleanup(func() { CatalogActivateAsThoughHaste, CatalogActivateAsThoughHasteOthersOnly = prev, prevOthers })
	const key = "test-others-haste"
	CatalogActivateAsThoughHaste = func(k string) bool { return k == key }
	CatalogActivateAsThoughHasteOthersOnly = func(k string) bool { return k == key }

	g := newActiveGame(t)
	me := g.Seats[0]
	src := NewCard("Adept", me.ID)
	src.TypeLine, src.OracleID, src.SummonedThisTurn = "Creature — Wizard", key, true
	other := NewCard("Bear", me.ID)
	other.TypeLine, other.SummonedThisTurn = "Creature — Bear", true
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(src)
		g.Battlefield.PushTop(other)
		s, _ := g.battlefieldCardLocked(src.InstanceID)
		o, _ := g.battlefieldCardLocked(other.InstanceID)
		if !g.activateAsThoughHasteLocked(o, me.ID) {
			t.Error("the static does not reach another creature")
		}
		if g.activateAsThoughHasteLocked(s, me.ID) {
			t.Error("the \"other creatures\" static reaches its own source")
		}
	})
}

// CR 611.2b: "it gains hexproof until your next turn" from a spend rider
// lasts through the other players' turns and ends as its caster's next
// turn begins.
func TestSpellGainsRiderUntilYourNextTurn(t *testing.T) {
	rider := ManaSpendRider{
		Kind: ManaRiderSpellGains, When: []string{ManaRestrictCast, "type:Creature"},
		Keywords: []string{"hexproof"}, UntilYourNextTurn: true,
	}
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, StepPrecombatMain)
	g.WithWriteLock(func() { me.ManaPool.AddMana(riderToken("C", uuid.New(), rider)) })
	c := NewCard("Test Dragon", me.ID)
	c.TypeLine, c.ManaCost, c.Controller, c.Power, c.Toughness = "Creature — Dragon", "{1}", me.ID, 2, 2
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	resolveTop(t, g)
	if got := layeredBattlefieldCard(t, g, c.InstanceID); !HasKeyword(&got, "hexproof") {
		t.Fatal("the creature entered without hexproof")
	}
	advancePastThisTurn(t, g)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := layeredBattlefieldCard(t, g, c.InstanceID); !HasKeyword(&got, "hexproof") {
		t.Fatal("hexproof ended with its caster's turn")
	}
	for i := 0; g.Seats[g.Turn.ActiveSeat].ID != me.ID; i++ {
		if i > 8 {
			t.Fatal("never reached the caster's next turn")
		}
		advancePastThisTurn(t, g)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := layeredBattlefieldCard(t, g, c.InstanceID); HasKeyword(&got, "hexproof") {
		t.Error("hexproof lasted into its caster's next turn")
	}
}

// #3855, #6094: a permanent that reads the payment of every ability its
// controller activates makes that controller's activations read, and not
// another player's.
func TestYourActivationsReaderMakesAnActivationRead(t *testing.T) {
	prev := CatalogReadsManaSpent
	t.Cleanup(func() { CatalogReadsManaSpent = prev })
	const key = "test-reads-your-activations"
	CatalogReadsManaSpent = func(k string) ManaSpentReads {
		if k == key {
			return ReadsYourActivations
		}
		return 0
	}
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	watcher := NewCard("Adept", me.ID)
	watcher.TypeLine, watcher.OracleID = "Creature — Wizard", key
	rock := NewCard("Rock", opp.ID)
	rock.TypeLine = "Artifact"
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(watcher)
		if read, _ := g.activationPaymentReadLocked(me.ID, rock); !read {
			t.Error("the watcher's controller's activation does not read")
		}
		if read, _ := g.activationPaymentReadLocked(opp.ID, rock); read {
			t.Error("another player's activation reads")
		}
	})
}
