package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tiered_test.go — Tiered (CR 702.183a, #3777): exactly one bullet,
// its own cost added to the printed one. Proof cards: Cloud's Limit
// Break, Thunder Magic, Fire Magic, Tifa's Limit Break, Restoration
// Magic.

const (
	cloudsLimitBreakOracle  = "93df61b5-6abe-4603-9a6e-cba90a834173"
	thunderMagicOracle      = "623c1873-cd83-4a5f-b391-fd0d2bb89061"
	fireMagicOracle         = "c87db803-3490-41bc-af7a-78f99bcd77c8"
	tifasLimitBreakOracle   = "d0d90dd8-0642-489f-9afc-11c307f258f2"
	restorationMagicOracle  = "ddbe0d17-5cac-417c-afc0-e4aebae83205"
	cloudsLimitBreakCost    = "{1}{W}"
	thunderMagicPrintedCost = "{R}"
)

// tieredInHand puts a tiered spell with its printed cost in the active
// player's hand at a main phase and returns it.
func tieredInHand(t *testing.T, g *game.Game, name, oracle, cost string) (*game.Player, game.Card) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Instant", ManaCost: cost,
		OracleID: oracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(c)
	return me, c
}

func TestTieredIsExactlyOneMode(t *testing.T) {
	for _, oracle := range []string{cloudsLimitBreakOracle, thunderMagicOracle, fireMagicOracle,
		tifasLimitBreakOracle, restorationMagicOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Modes == nil || spec.Modes.Min != 1 || spec.Modes.Max != 1 || spec.Modes.Prompt != TieredPrompt {
			t.Errorf("%s: modes = %+v, want tiered (exactly one)", spec.Name, spec.Modes)
		}
		for i, o := range spec.Modes.Options {
			if o.Cost == "" {
				t.Errorf("%s bullet %d has no cost; every tiered bullet prints one", spec.Name, i)
			}
		}
	}
}

// The chosen bullet's cost joins the printed cost, and only the chosen
// one's: Thunder Magic is {R}, {3}{R} or {5}{R}{R}.
func TestTieredPricesTheChosenBullet(t *testing.T) {
	g := newCatalogGame(t)
	me, spell := tieredInHand(t, g, "Thunder Magic", thunderMagicOracle, thunderMagicPrintedCost)
	for mode, want := range []int{1, 4, 7} {
		price, err := g.PriceCast(me.ID, spell, game.CastSpellParams{Modes: []int{mode}})
		if err != nil {
			t.Fatalf("PriceCast mode %d: %v", mode, err)
		}
		if got := price.Total.ManaValue(); got != want {
			t.Errorf("mode %d: total %s (%d), want mana value %d", mode, price.Total.String(), got, want)
		}
	}
}

// Two bullets are refused, and so is none: tiered is "choose one".
func TestTieredRefusesTwoBulletsAndNone(t *testing.T) {
	for _, modes := range [][]int{{0, 1}, nil} {
		g := newCatalogGame(t)
		me, spell := tieredInHand(t, g, "Fire Magic", fireMagicOracle, "{R}")
		floatMana(t, g, me, "{R}{R}{R}")
		err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: modes, Strict: true})
		if err == nil {
			t.Errorf("modes %v: cast accepted, want refused", modes)
		}
	}
}

// Strict payment: Omnislash costs {1}{W} + {3}{W}. Five mana is not
// enough; six is, and it destroys every tapped creature and no other.
func TestCloudsLimitBreakOmnislashChargesItsTier(t *testing.T) {
	g := newCatalogGame(t)
	me, spell := tieredInHand(t, g, "Cloud's Limit Break", cloudsLimitBreakOracle, cloudsLimitBreakCost)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tappedA := b12Creature(g, opp.ID, "Tapped A", "Creature — Bear", 2, 2)
	tappedB := b12Creature(g, me.ID, "Tapped B", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	tapForTest(t, g, tappedA, tappedB)

	floatMana(t, g, me, "{W}{W}{C}{C}{C}")
	err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{2}, Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("Omnislash with five mana: err = %v, want insufficient mana", err)
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{2}, Strict: true}); err != nil {
		t.Fatalf("Omnislash with six mana: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(tappedA) || g.Battlefield.Contains(tappedB) {
		t.Error("Omnislash should destroy every tapped creature, yours too")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("Omnislash destroyed an untapped creature")
	}
}

// Cross-Slash costs nothing extra: {1}{W} casts it, and it refuses an
// untapped target.
func TestCloudsLimitBreakCrossSlash(t *testing.T) {
	g := newCatalogGame(t)
	me, spell := tieredInHand(t, g, "Cloud's Limit Break", cloudsLimitBreakOracle, cloudsLimitBreakCost)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := b12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	tapForTest(t, g, tapped)

	floatMana(t, g, me, "{C}{W}")
	err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{modeRef(game.TargetCard, untapped, 0, 0)}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Cross-Slash on an untapped creature: err = %v, want illegal target", err)
	}
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{modeRef(game.TargetCard, tapped, 0, 0)}}); err != nil {
		t.Fatalf("Cross-Slash for {1}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("Cross-Slash should destroy its tapped target")
	}
}

// Blade Beam: any number of tapped creatures, no two with one
// controller.
func TestCloudsLimitBreakBladeBeam(t *testing.T) {
	g := newCatalogGame(t)
	me, spell := tieredInHand(t, g, "Cloud's Limit Break", cloudsLimitBreakOracle, cloudsLimitBreakCost)
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	a := b12Creature(g, opp1.ID, "A", "Creature — Bear", 2, 2)
	a2 := b12Creature(g, opp1.ID, "A2", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp2.ID, "B", "Creature — Bear", 2, 2)
	tapForTest(t, g, a, a2, b)

	floatMana(t, g, me, "{C}{C}{W}")
	err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{modeRef(game.TargetCard, a, 0, 0), modeRef(game.TargetCard, a2, 0, 0)}})
	if err == nil {
		t.Fatal("Blade Beam on two creatures with one controller was accepted")
	}
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{modeRef(game.TargetCard, a, 0, 0), modeRef(game.TargetCard, b, 0, 0)}}); err != nil {
		t.Fatalf("Blade Beam for {2}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("Blade Beam should destroy both targets")
	}
	if !g.Battlefield.Contains(a2) {
		t.Error("Blade Beam destroyed a creature it did not target")
	}
}

func TestThunderMagicThundaraDealsFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	four := b12Creature(g, opp.ID, "Four", "Creature — Beast", 4, 4)
	castModal(t, g, "Thunder Magic", "Instant", thunderMagicOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, four, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) {
		t.Error("Thundara should deal 4 to the 4/4")
	}
}

func TestFireMagicFiraDealsTwoToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	two := b12Creature(g, g.Seats[1].ID, "Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, g.Seats[0].ID, "Three", "Creature — Bear", 3, 3)
	castModal(t, g, "Fire Magic", "Instant", fireMagicOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("Fira should kill the 2/2")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("Fira should not kill the 3/3")
	}
}

func TestTifasLimitBreakFinalHeavenTriples(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Tifa's Limit Break", "Instant", tifasLimitBreakOracle,
		[]int{2}, []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)})
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 6 || tt != 9 {
		t.Errorf("Final Heaven on a 2/3: %d/%d, want 6/9", p, tt)
	}
}

func TestTifasLimitBreakMeteorStrikesDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Tifa's Limit Break", "Instant", tifasLimitBreakOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)})
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tt != 6 {
		t.Errorf("Meteor Strikes on a 2/3: %d/%d, want 4/6", p, tt)
	}
}

func TestRestorationMagicCuragaProtectsYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	life := me.Life
	castModal(t, g, "Restoration Magic", "Instant", restorationMagicOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	abs := effectiveAbilities(t, g, mine)
	if !containsString(abs, "hexproof") || !containsString(abs, "indestructible") {
		t.Errorf("Curaga: your creature has %v, want hexproof and indestructible", abs)
	}
	if containsString(effectiveAbilities(t, g, theirs), "indestructible") {
		t.Error("Curaga protected an opponent's creature")
	}
	if me.Life != life+6 {
		t.Errorf("Curaga: life %d, want %d", me.Life, life+6)
	}
}

func TestRestorationMagicCuraTargetsAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	life := me.Life
	castModal(t, g, "Restoration Magic", "Instant", restorationMagicOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, mine, 0, 0)})
	passPriorityAroundTable(t, g)
	if !containsString(effectiveAbilities(t, g, mine), "hexproof") {
		t.Error("Cura: the target should gain hexproof")
	}
	if me.Life != life+3 {
		t.Errorf("Cura: life %d, want %d", me.Life, life+3)
	}
}
