package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boltThenRedirect casts a Bolt at `victimA`, then `redirect` aimed
// at the Bolt, and passes until the redirect has resolved and its
// prompt (if any) is open. Returns the Bolt's instance id.
func boltThenRedirect(t *testing.T, g *game.Game, victimA uuid.UUID, redirectName, redirectOracle string) uuid.UUID {
	t.Helper()
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	// castCatalogSpell's fixture carries no printed cost, and Imp's
	// Mischief charges its controller the redirected spell's MANA
	// VALUE — so the Bolt needs its {R} to be worth a life.
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == bolt {
				g.Stack.Cards[i].ManaCost = "{R}"
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, redirectName, "Instant", redirectOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	return bolt
}

// Deflecting Swat is CR 115.7c: "you may choose new targets". The
// whole of batch 01's #294 entry in one test.
func TestDeflectingSwatRedirectsTheBolt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := boltThenRedirect(t, g, victimA, "Deflecting Swat", effectstest.DeflectingSwatOracle)

	prompt := effectstest.LatestRetarget(g, me)
	if prompt == nil {
		t.Fatal("Deflecting Swat opened no retarget prompt")
	}
	if prompt.PickTargetMin != 0 {
		t.Errorf(`"you may choose new targets" offers min %d, want 0`, prompt.PickTargetMin)
	}
	if !effectstest.HasID(prompt.PickTargetPlayers, victimB) {
		t.Fatalf("the Bolt's alternatives should include every other seat: %v", prompt.PickTargetPlayers)
	}
	if effectstest.HasID(prompt.PickTargetPlayers, victimA) {
		t.Error("the seat the Bolt already points at is offered as an alternative")
	}
	if err := g.ResolveRetarget(prompt.ID, me,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[bolt].Targets; len(got) != 1 || got[0].ID != victimB {
		t.Fatalf("the Bolt's target did not move: %+v", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victimA); got != 40 {
		t.Errorf("the original target's life = %d, want 40 — the Bolt was redirected", got)
	}
	if got := effectstest.LifeOf(g, victimB); got != 37 {
		t.Errorf("the new target's life = %d, want 37", got)
	}
}

// The "you may" half: declining leaves the Bolt exactly where it was.
func TestDeflectingSwatMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA := g.Seats[0].ID, g.Seats[1].ID
	boltThenRedirect(t, g, victimA, "Deflecting Swat", effectstest.DeflectingSwatOracle)

	prompt := effectstest.LatestRetarget(g, me)
	if prompt == nil {
		t.Fatal("no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, me, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, victimA); got != 37 {
		t.Errorf("declined redirect: original target life = %d, want 37", got)
	}
}

// Deflecting Swat's free cast is the Commander Legends offer shape,
// gated on a commander you CONTROL.
func TestDeflectingSwatFreeCastNeedsACommanderYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	swat := effectstest.VisibleHandCard(me, "Deflecting Swat", "Instant", "{2}{R}", effectstest.DeflectingSwatOracle, []string{"R"})
	if keys := effectstest.OfferedKeys(t, g, me, swat); effectstest.HasKey(keys, "free") {
		t.Errorf("the free cast is offered with no commander on the battlefield: %v", keys)
	}
	cmdr := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Their General", TypeLine: "Legendary Creature — Test", Power: 3, Toughness: 3})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cmdr {
			g.Battlefield.Cards[i].IsCommander = true
		}
	}
	if keys := effectstest.OfferedKeys(t, g, me, swat); !effectstest.HasKey(keys, "free") {
		t.Errorf("the free cast is missing with a commander on the battlefield: %v", keys)
	}
}

// Bolt Bend is CR 115.7b: "change the target … with a single
// target". Mandatory, and the clause refuses a spell with two
// targets rather than offering a prompt it cannot ask.
func TestBoltBendChangesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	bolt := boltThenRedirect(t, g, victimA, "Bolt Bend", effectstest.BoltBendOracle)

	prompt := effectstest.LatestRetarget(g, me)
	if prompt == nil {
		t.Fatal("Bolt Bend opened no retarget prompt")
	}
	if prompt.PickTargetMin != 1 {
		t.Errorf("a mandatory change offers min %d, want 1", prompt.PickTargetMin)
	}
	if err := g.ResolveRetarget(prompt.ID, me, nil); err == nil {
		t.Error("a mandatory change was declined")
	}
	if err := g.ResolveRetarget(prompt.ID, me,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[bolt].Targets; got[0].ID != victimB {
		t.Fatalf("the Bolt's target did not move: %+v", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 40 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("damage landed on the wrong seat: A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
}

// "This spell costs {3} less to cast if you control a creature with
// power 4 or greater" — the discount half of the same card.
func TestBoltBendCostsThreeLessWithABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BoltBendOracle, "Instant", "{3}{R}"); got != 4 {
		t.Errorf("Bolt Bend with no creatures: %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Small", TypeLine: "Creature — Test", Power: 3, Toughness: 3})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BoltBendOracle, "Instant", "{3}{R}"); got != 4 {
		t.Errorf("Bolt Bend with a 3-power creature: %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Test", Power: 4, Toughness: 4})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.BoltBendOracle, "Instant", "{3}{R}"); got != 1 {
		t.Errorf("Bolt Bend with a 4-power creature: %d, want 1", got)
	}
}

// Misdirection: the pitch cost, then the same redirect. The one card
// in the family that ships with no caveat.
func TestMisdirectionPitchesABlueCardAndRedirects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	effectstest.ToMainForCost(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})

	mis := effectstest.VisibleHandCard(me, "Misdirection", "Instant", "{3}{U}{U}", effectstest.MisdirectionOracle, []string{"U"})
	pitch := effectstest.VisibleHandCard(me, "Brainstorm", "Instant", "{U}", "test-blue-pitch-misdirection", []string{"U"})
	if keys := effectstest.OfferedKeys(t, g, me, mis); !effectstest.HasKey(keys, "pitch") {
		t.Fatalf("Misdirection's pitch is not offered with a blue card in hand: %v", keys)
	}
	if err := g.CastSpell(me.ID, mis, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("CastSpell Misdirection: %v", err)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the pitched card is still in hand")
	}
	effectstest.PassPriorityAroundTable(t, g)

	prompt := effectstest.LatestRetarget(g, me.ID)
	if prompt == nil {
		t.Fatal("Misdirection opened no retarget prompt")
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 40 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("damage landed on the wrong seat: A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
}

// Imp's Mischief redirects and charges its controller the redirected
// spell's mana value in life — Lightning Bolt's is 1.
func TestImpsMischiefRedirectsAndCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	before := effectstest.LifeOf(g, me)
	boltThenRedirect(t, g, victimA, "Imp's Mischief", effectstest.ImpsMischiefOracle)

	prompt := effectstest.LatestRetarget(g, me)
	if prompt == nil {
		t.Fatal("Imp's Mischief opened no retarget prompt")
	}
	if got := effectstest.LifeOf(g, me); got != before-1 {
		t.Errorf("life after Imp's Mischief on a Bolt = %d, want %d (the Bolt's mana value)", got, before-1)
	}
	if err := g.ResolveRetarget(prompt.ID, me,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimB}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 40 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("damage landed on the wrong seat: A=%d B=%d", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
}

// "With a single target" is a clause predicate, not a
// resolution-time refusal: a two-target spell is never offered to
// Bolt Bend's picker in the first place.
func TestSingleTargetClauseExcludesAMultiTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0], g.Seats[1].ID, g.Seats[2].ID
	effectstest.ToMainForCost(t, g)
	single := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	// Bite Down-shaped: two targets, one clause each.
	double := effectstest.CastCatalogSpell(t, g, "Bite Down", "Instant", effectstest.B26BiteDownOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: effectstest.PushPermanent(g, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Test", Power: 5, Toughness: 5}), Slot: 0},
		{Kind: game.TargetCard, ID: effectstest.PushPermanent(g, victimB, game.Card{Name: "Theirs", TypeLine: "Creature — Test", Power: 2, Toughness: 2}), Slot: 1},
	})

	bend := effectstest.VisibleHandCard(me, "Bolt Bend", "Instant", "{3}{R}", effectstest.BoltBendOracle, []string{"R"})
	var legal *game.LegalTargets
	g.WithWriteLock(func() {
		legal = &game.LegalTargets{}
		*legal = g.LegalTargetsForEffect(
			game.SourceObject(me.ID, &game.Card{InstanceID: bend, OracleID: effectstest.BoltBendOracle}),
			game.TargetSpecFor(effectstest.BoltBendOracle))
	})
	if !effectstest.HasID(legal.Cards, single) {
		t.Errorf("a one-target Bolt is not offered to Bolt Bend: %v", legal.Cards)
	}
	if effectstest.HasID(legal.Cards, double) {
		t.Error("a two-target spell reached Bolt Bend's picker despite \"with a single target\"")
	}
}
