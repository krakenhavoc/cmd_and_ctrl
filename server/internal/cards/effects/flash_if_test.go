package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_if_test.go — #5207: "You may cast this spell as though it had
// flash if <condition>" is a timing permission (Spec.FlashIf,
// game/flash_if.go), and "if you pay {2} more" / "if you behold a
// Dragon" an optional cost that opens the window (CR 601.3c). Neither is
// an alternative cost, so a flash cast also takes one (CR 118.9a). The
// enumerator and the view halves of the X ceiling live in internal/legal
// and internal/protocol; this file is the announce gate and the cards.

const silverScrutinyFlashOracle = "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e"

// flashIfHandCard puts a catalog card in `me`'s hand.
func flashIfHandCard(g *game.Game, me *game.Player, name, typeLine, mana, oracle string) uuid.UUID {
	card := eacHandCard(g, me, name, typeLine, mana)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].OracleID = oracle
			}
		}
	})
	return card
}

func ssFlashTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := p7Table(t)
	return g, me, flashIfHandCard(g, me, "Silver Scrutiny", "Sorcery", "{X}{U}{U}", silverScrutinyFlashOracle)
}

func ssFund(g *game.Game, me *game.Player, n int) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for i := 0; i < n; i++ {
			c := "C"
			if i < 2 {
				c = "U"
			}
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func ssToBeginCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func flashIfFund(g *game.Game, me *game.Player, colors ...string) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for _, c := range colors {
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

// flashIfOmniscience puts Omniscience ("You may cast spells from your
// hand without paying their mana costs") on `me`'s battlefield: an
// alternative cost from another source (CR 118.9).
func flashIfOmniscience(g *game.Game, me *game.Player) {
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Omniscience", TypeLine: "Enchantment",
		OracleID: omniscienceOracle, Owner: me.ID, Controller: me.ID,
	})
}

// At instant speed the card's own permission opens the window for the
// printed cost, and an X at the ceiling is accepted.
func TestSilverScrutinyFlashesForXUpToThree(t *testing.T) {
	g, me, card := ssFlashTable(t)
	ssToBeginCombat(t, g)
	ssFund(g, me, 5)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("flash cast with X=3: %v", err)
	}
	if got := g.StackMeta[card].XValue; got != 3 {
		t.Errorf("XValue = %d, want 3", got)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left: %d, want 0 (the printed cost was paid)", len(me.ManaPool))
	}
}

// X = 4 at instant speed is refused (CR 601.3b), and nothing moves.
func TestSilverScrutinyRefusesXFourAtInstantSpeed(t *testing.T) {
	g, me, card := ssFlashTable(t)
	ssToBeginCombat(t, g)
	ssFund(g, me, 6)
	err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 4})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("flash cast with X=4: err = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(card) {
		t.Errorf("a refused cast reached the stack")
	}
	var inHand bool
	g.ReadSnapshot(func() { inHand = me.Hand.Contains(card) })
	if !inHand {
		t.Errorf("the refused cast left the hand")
	}
}

// The sorcery-speed cast does not need the permission, so it is not
// bound by it: X = 10 is fine in a main phase.
func TestSilverScrutinySorcerySpeedCastTakesAnyX(t *testing.T) {
	g, me, card := ssFlashTable(t)
	advanceToMain(t, g)
	ssFund(g, me, 12)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 10}); err != nil {
		t.Fatalf("sorcery-speed cast with X=10: %v", err)
	}
	if got := g.StackMeta[card].XValue; got != 10 {
		t.Errorf("XValue = %d, want 10", got)
	}
}

// Silver Scrutiny is Full and draws X.
func TestSilverScrutinyIsFullAndDrawsX(t *testing.T) {
	spec, ok := Lookup(silverScrutinyFlashOracle)
	if !ok {
		t.Fatal("Silver Scrutiny is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || len(spec.AlternativeCosts) != 0 {
		t.Errorf("completeness = %v, caveats = %v, offers = %d; want Full, none, none", spec.Completeness, spec.Caveats, len(spec.AlternativeCosts))
	}
	if spec.FlashIf == nil || spec.FlashIf.XCeiling != 3 {
		t.Errorf("FlashIf = %+v, want an X ceiling of 3", spec.FlashIf)
	}

	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Silver Scrutiny", "Sorcery", silverScrutinyFlashOracle, "{X}{U}{U}", 4, nil)
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
}

// The heart of #5207: a flash cast also claims an alternative cost from
// another source. Bard's Company with a Human out, in the beginning of
// combat step, cast for Omniscience's "without paying its mana cost".
func TestFlashIfCastAlsoTakesAnotherAlternativeCost(t *testing.T) {
	g, me, _ := p7Table(t)
	company := flashIfHandCard(g, me, "Bard's Company", "Creature — Human Citizen", "{2}{W}{U}", bardsCompanyOracle)
	b12Creature(g, me.ID, "Human Soldier", "Creature — Human Soldier", 1, 1)
	flashIfOmniscience(g, me)
	ssToBeginCombat(t, g)
	flashIfFund(g, me)
	if err := g.CastSpell(me.ID, company, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); err != nil {
		t.Fatalf("flash cast through Omniscience: %v", err)
	}
	if !g.Stack.Contains(company) {
		t.Fatal("Bard's Company is not on the stack")
	}
}

// The condition false: no instant-speed cast, for any cost.
func TestFlashIfConditionFalseRefusesTheInstantSpeedCast(t *testing.T) {
	g, me, _ := p7Table(t)
	company := flashIfHandCard(g, me, "Bard's Company", "Creature — Human Citizen", "{2}{W}{U}", bardsCompanyOracle)
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	flashIfOmniscience(g, me)
	ssToBeginCombat(t, g)
	flashIfFund(g, me, "W", "U", "C", "C")
	if err := g.CastSpell(me.ID, company, game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("printed cost with no Human: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, company, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Omniscience's cost with no Human: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if g.Stack.Contains(company) {
		t.Error("a refused cast reached the stack")
	}
}

// "If you pay {2} more": the {2} is an optional additional cost, so a
// flash cast through Omniscience pays only the {2} (CR 601.2f), and one
// that does not announce it stays shut (CR 601.3c).
func TestFlashForMoreCombinesWithAFreeCast(t *testing.T) {
	g, me, _ := p7Table(t)
	antics := flashIfHandCard(g, me, "Asinine Antics", "Sorcery", "{2}{U}{U}", asininAnticsOracle)
	flashIfOmniscience(g, me)
	ssToBeginCombat(t, g)
	flashIfFund(g, me, "C", "C")
	if err := g.CastSpell(me.ID, antics, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("free cast without the {2}: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, antics, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("free cast paying {2} more: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left: %d, want 0 (the {2} was paid)", len(me.ManaPool))
	}
}

// "If it targets a commander", cast through Omniscience at instant speed:
// the commander is a legal target and a plain creature is not.
func TestFlashIfTargetsHoldsUnderAnotherAlternativeCost(t *testing.T) {
	g, me, ward, commander, bear := timelyWardTable(t)
	flashIfOmniscience(g, me)
	ssToBeginCombat(t, g)
	free := func(id uuid.UUID) game.CastSpellParams {
		return game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostFree,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
	}
	if err := g.CastSpell(me.ID, ward, free(bear)); err == nil {
		t.Fatal("free flash cast onto a creature that is not a commander")
	}
	if err := g.CastSpell(me.ID, ward, free(commander)); err != nil {
		t.Fatalf("free flash cast onto a commander: %v", err)
	}
}

// No card builds "as though it had flash if" as an alternative cost any
// more, and each one is Full.
func TestFlashIfCardsAreFull(t *testing.T) {
	for _, oracle := range []string{
		"63957ec0-89e6-4be2-8740-846ca56ee48c", // Scarring Memories
		"82aa1402-4b1c-4343-aba5-dd3271cc1ac5", // Molten Exhale
		asininAnticsOracle,
		"1b639537-a8fe-4615-b937-afa478fdec0f", // Oakshade Stalker
		silverScrutinyFlashOracle,
		timelyWardOracle,
		"610d9f58-4182-49cb-855c-9b6030e7f714", // Fated Clash
		bardsCompanyOracle,
		"9f5277e8-1cbe-478c-a8bc-ac028f13305e", // Harbinger of the Tides
		"fb1a6986-39f7-4592-9039-7e30f5ccd009", // Mystical Tether
		"98e36c6e-a868-4679-a5b5-6f02097f560b", // Serpent of the Pass
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v %q, want Full with no caveats", spec.Name, spec.Completeness, spec.Caveats)
		}
		for _, ac := range spec.AlternativeCosts {
			if ac.AsThoughFlash {
				t.Errorf("%s still offers %q as an alternative cost", spec.Name, ac.Key)
			}
		}
		opens := spec.FlashIf != nil
		for _, oc := range spec.OptionalCosts {
			opens = opens || oc.AsThoughFlash
		}
		if !opens {
			t.Errorf("%s has neither a FlashIf nor an optional cost that opens the window", spec.Name)
		}
	}
}

// Register refuses a FlashIf it cannot apply, at boot.
func TestFlashIfRegisterChecks(t *testing.T) {
	cases := map[string]Spec{
		"negative ceiling":   {FlashIf: &game.FlashIf{XCeiling: -1}},
		"nothing to narrow":  {FlashIf: FlashIfTargets("a commander", EnchantCreature(ACommander()))},
		"no condition":       {FlashIf: &game.FlashIf{}},
		"with the cleanup":   {FlashIf: FlashIfXAtMost(3), FlashWithCleanupSacrifice: true},
		"optional not flash": {OptionalCosts: []game.AdditionalCost{{Key: "k", ManaCost: "{2}", AsThoughFlash: true}}},
	}
	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("Register accepted the spec")
				}
			}()
			spec.OracleID = "test-flash-if-register-" + strings.ReplaceAll(name, " ", "-")
			spec.Name = "Test Flash If " + name
			Register(spec)
		})
	}
}
