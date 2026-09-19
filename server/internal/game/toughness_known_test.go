package game

import (
	"testing"

	"github.com/google/uuid"
)

// toughness_known_test.go — #690 and #691: the toughness state-based
// action's stand-in skip (Card.ToughnessIsKnown) covers exactly the
// objects the engine has no toughness for.
//
//   - #690: a `*` creature whose characteristic-defining ability the
//     layer system COMPUTES is not a stand-in any more — Consuming
//     Aberration and Lord of Extinction at computed toughness 0 die,
//     and the damage checks below the skip finally see them at all.
//   - #691: an object with a printing behind it has a real printed 0 —
//     a Hangarback Walker cast for X=0 dies at once, as in paper.
//
// #683's own cases live in zero_toughness_test.go and still hold.

const cdaOracleForTest = "cda-pt-for-test"

// cdaSizeForTest is how big every coded-CDA fixture in this file is;
// the test sets it before each recompute. It stands in for "cards in
// all graveyards" without needing a graveyard.
var cdaSizeForTest int

// withCodedCDA installs a layer-7a characteristic-defining ability on
// cdaOracleForTest that sets power and toughness to cdaSizeForTest —
// Lord of Extinction's shape, with the count made explicit.
func withCodedCDA(t *testing.T) {
	t.Helper()
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != cdaOracleForTest {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7A_CDA,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power, c.Toughness = cdaSizeForTest, cdaSizeForTest
			},
		}}
	})
}

// pushCodedStar seats a `*`/`*` creature whose CDA the engine codes:
// the importer's stand-in 0 in Toughness, VariableToughness set, and
// an oracle ID the catalog answers for.
func pushCodedStar(g *Game, controller *Player) uuid.UUID {
	return pushTypedTestCard(g, Card{
		Name:              "Coded Star",
		TypeLine:          "Creature — Elemental",
		OracleID:          cdaOracleForTest,
		VariableToughness: true,
		Owner:             controller.ID,
		Controller:        controller.ID,
	})
}

// A coded `*`/`*` creature at computed toughness 0 is a real 0/0 and
// CR 704.5f puts it into its owner's graveyard. Lord of Extinction
// with every graveyard empty; Consuming Aberration against opponents
// who have discarded nothing (#690).
func TestCodedCDACreatureAtZeroToughnessDies(t *testing.T) {
	withCodedCDA(t)
	cdaSizeForTest = 0

	g := newActiveGame(t)
	me := g.Seats[0]
	star := pushCodedStar(g, me)

	runSBAsForTest(g)
	if g.Battlefield.Contains(star) {
		t.Error("a coded `*` creature computed at toughness 0 stayed on the battlefield — CR 704.5f")
	}
	if !me.Graveyard.Contains(star) {
		t.Error("it did not reach its owner's graveyard")
	}
}

// The same card at a computed toughness of 1 or more lives: the skip
// was never about the card, it was about the number.
func TestCodedCDACreatureAboveZeroLives(t *testing.T) {
	withCodedCDA(t)
	cdaSizeForTest = 3

	g := newActiveGame(t)
	me := g.Seats[0]
	star := pushCodedStar(g, me)

	runSBAsForTest(g)
	if !g.Battlefield.Contains(star) {
		t.Fatal("a coded `*` creature computed at toughness 3 died")
	}
	if got := layeredBattlefieldCard(t, g, star).CurrentToughness(); got != 3 {
		t.Errorf("computed toughness %d, want 3", got)
	}
}

// The skip covered the WHOLE creature, damage checks included, so a
// coded `*` creature could not be killed by damage either. It can now
// (CR 704.5g).
func TestCodedCDACreatureDiesToLethalDamage(t *testing.T) {
	withCodedCDA(t)
	cdaSizeForTest = 2

	g := newActiveGame(t)
	me := g.Seats[0]
	star := pushCodedStar(g, me)

	g.WithWriteLock(func() {
		if c := findBattlefieldCard(g, star); c != nil {
			c.DamageMarked = 2
		}
	})
	runSBAsForTest(g)
	if g.Battlefield.Contains(star) {
		t.Error("a coded `*` creature survived lethal damage — the stand-in skip was skipping its damage check too")
	}
}

// An UNCODED `*` creature is untouched: nothing computes its size, so
// its 0 is still the importer's stand-in and the skip is still what
// keeps Mortivore on the battlefield (#683).
func TestUncodedStarCreatureIsStillSkipped(t *testing.T) {
	withCodedCDA(t)
	cdaSizeForTest = 0

	g := newActiveGame(t)
	me := g.Seats[0]
	// Same shape as pushCodedStar but with an oracle ID the catalog
	// answers nothing for.
	star := pushTypedTestCard(g, Card{
		Name:              "Uncoded Star",
		TypeLine:          "Creature — Lhurgoyf",
		OracleID:          "no-such-card",
		VariableToughness: true,
		Owner:             me.ID,
		Controller:        me.ID,
	})

	runSBAsForTest(g)
	if !g.Battlefield.Contains(star) {
		t.Error("an uncoded `*` creature was killed — the stand-in skip regressed (#683)")
	}
}

// A `*` creature that a coded CDA sizes carries a real toughness even
// with a printing behind it, and the two agree: the CDA wins, because
// ToughnessIsKnown asks it first.
func TestToughnessIsKnownOrdersItsAnswers(t *testing.T) {
	printing := uuid.New().String()
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"a printed 2/2", Card{Toughness: 2}, true},
		{"a stand-in 0 with a counter on it", Card{VariableToughness: true, Counters: map[string]int{CounterPlusOne: 1}}, true},
		{"an uncoded `*`", Card{VariableToughness: true}, false},
		{"an uncoded `*` that lost its last counter", Card{VariableToughness: true, LostLastCounter: true}, false},
		{"a token that lost its last counter", Card{LostLastCounter: true}, true},
		{"a 0/0 token template", Card{}, false},
		{"a 0/0 with a printing", Card{ScryfallID: printing}, true},
		{"a `*` with a printing", Card{ScryfallID: printing, VariableToughness: true}, false},
	} {
		if got := tc.card.ToughnessIsKnown(); got != tc.want {
			t.Errorf("%s: ToughnessIsKnown = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// #691: a printed 0/0 that carries a printing is a printed 0/0. A
// Hangarback Walker cast for X=0 enters with no counters and dies at
// the next state-based check (CR 704.5f, CR 601.2b — X=0 is a legal
// announcement, it just does not survive one).
func TestPrintedZeroZeroWithAPrintingDies(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	walker := NewCard("Hangarback Walker", me.ID)
	walker.TypeLine = "Artifact Creature — Construct"
	walker.ScryfallID = uuid.New().String()
	walker.Controller = me.ID
	g.Battlefield.PushTop(walker)

	runSBAsForTest(g)
	if g.Battlefield.Contains(walker.InstanceID) {
		t.Error("a printed 0/0 with a printing behind it stayed on the battlefield — CR 704.5f (#691)")
	}
	if !me.Graveyard.Contains(walker.InstanceID) {
		t.Error("it did not reach its owner's graveyard")
	}
}

// The same body with no printing behind it is a fixture or a token
// template, and the engine has no toughness for it: still skipped.
func TestPrintedZeroZeroWithNoPrintingIsStillSkipped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	token := pushZeroZero(g, me, nil)

	runSBAsForTest(g)
	if !g.Battlefield.Contains(token) {
		t.Error("a 0/0 with no printing behind it was killed — the stand-in skip regressed")
	}
}

// A 0/0 with a printing that something else is holding up lives, and
// this is the case the old skip could not even reach: the skip fired
// before the toughness was computed, so an anthem on a printed 0/0
// never got a hearing.
func TestPrintedZeroZeroHeldUpByAnAnthemLives(t *testing.T) {
	const anthemOracle = "anthem-for-zero-zero-test"
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != anthemOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID != source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power++
				c.Toughness++
			},
		}}
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	pushTypedTestCard(g, Card{
		Name: "Anthem", TypeLine: "Enchantment", OracleID: anthemOracle,
		Owner: me.ID, Controller: me.ID,
	})
	walker := pushTypedTestCard(g, Card{
		Name: "Printed Zero", TypeLine: "Artifact Creature — Construct",
		ScryfallID: uuid.New().String(),
		Owner:      me.ID, Controller: me.ID,
	})

	runSBAsForTest(g)
	if !g.Battlefield.Contains(walker) {
		t.Fatal("a printed 0/0 under a +1/+1 anthem died at toughness 1")
	}
	if got := layeredBattlefieldCard(t, g, walker).CurrentToughness(); got != 1 {
		t.Errorf("computed toughness %d, want 1", got)
	}
}

// --- copies (#967's copiable values) ---------------------------------

// A Clone of a coded `*`/`*` creature is that creature: the copy
// carries the copied printing's `*` stand-in AND its oracle ID, so
// the layer pass gathers the SAME characteristic-defining ability for
// the copy and stamps PTDefined on it. The copy therefore answers
// branch 2, not branch 3, and dies with the original at a computed 0.
//
// It is the branch that has to work per OBJECT rather than per card:
// `PTDefined` is recomputed for every object the effect applies to,
// and the copy's AppliesTo names itself.
func TestCopyOfACodedStarCarriesPTDefined(t *testing.T) {
	withCodedCDA(t)

	for _, tc := range []struct {
		name string
		size int
		dies bool
	}{
		{"computed 0", 0, true},
		{"computed 3", 3, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cdaSizeForTest = tc.size
			g := newActiveGame(t)
			me := g.Seats[0]
			star := pushCodedStar(g, me)
			clone := pushTypedTestCard(g, Card{
				Name: "Clone", TypeLine: "Creature — Shapeshifter",
				Owner: me.ID, Controller: me.ID,
			})

			g.mu.Lock()
			src, _ := g.battlefieldCardLocked(star)
			dst, _ := g.battlefieldCardLocked(clone)
			dst.applyCopy(CopiableValuesOf(*src), *src)
			g.layerVersion.Add(1)
			g.mu.Unlock()

			// The copy took the stand-in 0 and the bit that says so,
			// which is what makes branch 2 load-bearing here.
			if c := findBattlefieldCard(g, clone); c == nil || !c.VariableToughness {
				t.Fatal("the copy did not carry VariableToughness off the copied printing")
			}

			runSBAsForTest(g)
			if got := g.Battlefield.Contains(clone); got == tc.dies {
				t.Errorf("copy on the battlefield = %v at computed toughness %d, want %v", got, tc.size, !tc.dies)
			}
			if g.Battlefield.Contains(star) == tc.dies {
				t.Error("the original and its copy disagreed")
			}
		})
	}
}

// CR 608.3f: a resolving copy of a permanent spell becomes a token,
// and `tokenCopyOfSpell` builds it out of `CopiableValuesOf`. That
// carries the copied card's ScryfallID along with its body, and the
// two travel together — which is the answer for this rule rather than
// a hazard: a token copy of a Hangarback Walker IS the printed 0/0 it
// copied and dies with no counters (branch 5), while a token copy of
// an uncoded `*` carries the stand-in bit with the stand-in 0 and
// stays skipped (branch 3).
//
// The case to guard against is a token that takes the printing
// WITHOUT the body it belongs to; these two rows are what say it
// cannot happen.
func TestTokenCopyOfAPermanentSpellReadsTheCopiedBody(t *testing.T) {
	printing := uuid.New().String()
	for _, tc := range []struct {
		name string
		src  Card
		want bool
	}{
		{
			name: "a printed 0/0",
			src: Card{
				Name: "Hangarback Walker", TypeLine: "Artifact Creature — Construct",
				ScryfallID: printing,
			},
			want: true,
		},
		{
			name: "an uncoded `*`",
			src: Card{
				Name: "Mortivore", TypeLine: "Creature — Lhurgoyf",
				ScryfallID: printing, VariableToughness: true,
			},
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok := tokenCopyOfSpell(tc.src)
			if !tok.IsToken() {
				t.Fatal("tokenCopyOfSpell did not make a token")
			}
			if tok.ScryfallID != tc.src.ScryfallID {
				t.Errorf("token ScryfallID = %q, want the copied printing %q", tok.ScryfallID, tc.src.ScryfallID)
			}
			if tok.VariableToughness != tc.src.VariableToughness {
				t.Errorf("token VariableToughness = %v, want the copied %v — the printing must not travel without the bit that reads it",
					tok.VariableToughness, tc.src.VariableToughness)
			}
			if got := tok.ToughnessIsKnown(); got != tc.want {
				t.Errorf("ToughnessIsKnown = %v, want %v", got, tc.want)
			}
		})
	}
}

// And the same through the real resolution path: a token copy of a
// printed 0/0 permanent spell is swept by CR 704.5f the moment it
// lands, rather than sitting on the battlefield as an unkillable 0/0.
func TestTokenCopyOfAPrintedZeroZeroDiesOnArrival(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	tok := tokenCopyOfSpell(Card{
		Name: "Hangarback Walker", TypeLine: "Artifact Creature — Construct",
		ScryfallID: uuid.New().String(),
	})
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	g.Battlefield.PushTop(tok)

	runSBAsForTest(g)
	if g.Battlefield.Contains(tok.InstanceID) {
		t.Error("a CR 608.3f token copy of a printed 0/0 stayed on the battlefield — CR 704.5f")
	}
}

// A layer 7b "set" defines the body too (CR 613.4b), so a 0/0 token
// that a Hallowed Haunting sizes is out of the skip while the
// enchantment is there — and a 7b set to 0/0 kills it.
func TestLayerSevenBSetDefinesTheBodyToo(t *testing.T) {
	const setOracle = "seven-b-set-for-test"
	size := 0
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != setOracle {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7B_Set,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool {
				return target.InstanceID != source.InstanceID
			},
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Power, c.Toughness = size, size
			},
		}}
	})

	g := newActiveGame(t)
	me := g.Seats[0]
	pushTypedTestCard(g, Card{
		Name: "Humility", TypeLine: "Enchantment", OracleID: setOracle,
		Owner: me.ID, Controller: me.ID,
	})
	// No printing, no counters: skipped on every branch but the
	// layer's.
	token := pushTypedTestCard(g, Card{
		Name: "Spirit", TypeLine: "Creature — Spirit Cleric",
		Owner: me.ID, Controller: me.ID,
	})

	runSBAsForTest(g)
	if g.Battlefield.Contains(token) {
		t.Error("a 0/0 token a layer-7b effect sets to 0/0 stayed on the battlefield")
	}
}
