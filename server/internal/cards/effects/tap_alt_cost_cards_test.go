package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tap_alt_cost_cards_test.go — ADR 0135 §1 (#2030): alternative costs that
// tap permanents you control, against the real catalog. The tap reuses the
// activated abilities' TapOthersCost (owner decision 1): untapped, yours,
// not targeted, no summoning-sickness check (CR 302.6), paid with the spell
// on the stack.

const (
	orimsCureOracle        = "61eacaab-625b-499c-8602-f37ad5dda138"
	sivvisValorOracle      = "e02164ba-34f8-4a5f-a05b-dd3ef3f8ceae"
	angelicFavorOracle     = "37162c6e-ac31-4386-9d1e-76aa66070b35"
	ramosianRallyOracle    = "2731c237-eda1-4b74-a294-502d574f4435"
	lashknifeOracle        = "98d790c0-985f-44a4-b247-0951beea7637"
	prismaticStrandsOracle = "8e414ad6-ba19-44a1-a291-2e420734a6dd"
	battleScreechOracle    = "e73131cd-b454-405b-9539-9d777e232b9e"
	groupProjectOracle     = "e1ce210e-2d1d-4ab4-bfb7-8e36884797fc"
	zahidOracle            = "89b07037-64df-4e66-acd8-87d97df61e3a"
	tapAltCostKey          = "tap"
)

func TestTapAltCostCardsAreFull(t *testing.T) {
	for _, c := range []struct {
		oracle, key string
		n           int
		mana        string
	}{
		{orimsCureOracle, tapAltCostKey, 1, ""},
		{sivvisValorOracle, tapAltCostKey, 1, ""},
		{angelicFavorOracle, tapAltCostKey, 1, ""},
		{ramosianRallyOracle, tapAltCostKey, 1, ""},
		{lashknifeOracle, tapAltCostKey, 1, ""},
		{prismaticStrandsOracle, "flashback", 1, ""},
		{battleScreechOracle, "flashback", 3, ""},
		{groupProjectOracle, "flashback", 3, ""},
		{zahidOracle, tapAltCostKey, 1, "{3}{U}"},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 || offers[0].Key != c.key || offers[0].ManaCost != c.mana ||
			offers[0].TapOthers == nil || offers[0].CardPaymentCount() != c.n {
			t.Errorf("%s offers %+v, want one %q tapping %d for %q", spec.Name, offers, c.key, c.n, c.mana)
		}
	}
}

// tapTable is a main phase on seat 0's turn with an empty hand.
func tapTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}

func castTapping(g *game.Game, p *game.Player, card uuid.UUID, key string, targets []game.TargetRef, tap ...uuid.UUID) error {
	params := game.CastSpellParams{AlternativeCost: key, AltCostIDs: tap, Strict: true, Targets: targets}
	if key == "flashback" {
		params.FromZone = "graveyard"
	}
	return g.CastSpell(p.ID, card, params)
}

// Orim's Cure: offered only with a Plains; refuses a tapped creature, an
// opponent's creature and two creatures; accepts one that arrived this
// turn (CR 302.6) and a hexproof one (a cost does not target); taps it
// with the spell on the stack; and the shield prevents the next 4.
func TestOrimsCureIsCastByTappingAnUntappedCreature(t *testing.T) {
	g, me, opp := tapTable(t)
	cure := handCardOf(g, me, "Orim's Cure", "Instant", "{1}{W}", orimsCureOracle)
	fresh := pr7Creature(g, me.ID, "Fresh", 2, "G")
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, fresh)
		c.SummonedThisTurn = true
		c.Keywords = []string{"hexproof"}
	})
	already := pr7Creature(g, me.ID, "Tapped", 2, "G")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, already).Tapped = true })
	theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
	other := pr7Creature(g, me.ID, "Other", 2, "G")
	self := []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}

	if o := offerWithKey(grantedCardOffers(g, me.ID, cure, game.ZoneHand), tapAltCostKey); o != nil {
		t.Error("Orim's Cure's tap offer is listed with no Plains")
	}
	if err := castTapping(g, me, cure, tapAltCostKey, self, fresh); err == nil {
		t.Fatal("Orim's Cure was cast for its tap cost with no Plains")
	}
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if o := offerWithKey(grantedCardOffers(g, me.ID, cure, game.ZoneHand), tapAltCostKey); o == nil {
		t.Fatal("Orim's Cure's tap offer is not listed with a Plains and untapped creatures")
	}
	for name, ids := range map[string][]uuid.UUID{
		"an already tapped creature": {already},
		"an opponent's creature":     {theirs},
		"two creatures":              {fresh, other},
		"nothing":                    nil,
	} {
		if err := castTapping(g, me, cure, tapAltCostKey, self, ids...); err == nil {
			t.Fatalf("%s paid Orim's Cure", name)
		}
	}
	if tappedForTest(t, g, fresh) || tappedForTest(t, g, other) || tappedForTest(t, g, theirs) || !me.Hand.Contains(cure) {
		t.Fatal("a refused cast paid something")
	}
	if err := castTapping(g, me, cure, tapAltCostKey, self, fresh); err != nil {
		t.Fatalf("Orim's Cure for a creature that arrived this turn: %v", err)
	}
	if g.Stack.Size() != 1 || !tappedForTest(t, g, fresh) {
		t.Fatal("the creature was not tapped with Orim's Cure on the stack")
	}
	if tappedForTest(t, g, other) {
		t.Error("another creature was tapped too")
	}
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, theirs, me.ID, 6)
	if me.Life != life-2 {
		t.Errorf("life %d after 6 damage, want %d (4 prevented)", me.Life, life-2)
	}
}

// One creature is tapped once (CR 118.3): it can't pay the tap cost and
// convoke on the same cast, and the auto-tapper won't tap an artifact
// named to Zahid's cost for Zahid's {3}{U}.
func TestZahidTapsAnArtifactTheAutoTapperLeavesAlone(t *testing.T) {
	g, me, _ := tapTable(t)
	zahid := handCardOf(g, me, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", "{4}{U}{U}", zahidOracle)
	for i := 0; i < 4; i++ {
		apaPush(g, me.ID, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	creature := pr7Creature(g, me.ID, "Not an artifact", 2, "U")
	if err := g.CastSpell(me.ID, zahid, game.CastSpellParams{AlternativeCost: tapAltCostKey, AltCostIDs: []uuid.UUID{creature}, Strict: true, AutoTap: true}); err == nil {
		t.Fatal("a creature paid \"tap an untapped artifact\"")
	}
	if err := g.CastSpell(me.ID, zahid, game.CastSpellParams{AlternativeCost: tapAltCostKey, AltCostIDs: []uuid.UUID{rock}, Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Zahid for {3}{U} and an artifact: %v", err)
	}
	if !tappedForTest(t, g, rock) {
		t.Error("the artifact was not tapped")
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Island" && c.Tapped {
			tapped++
		}
	}
	if tapped != 4 {
		t.Errorf("%d Islands tapped, want 4 for {3}{U}", tapped)
	}
}

func TestATapCostCreatureCannotAlsoConvoke(t *testing.T) {
	const oracle = "test-0135-tap-and-convoke"
	registerForTest(t, Spec{OracleID: oracle, Name: "Test Convoke Tap", Completeness: CompletenessFull,
		TapCost: Convoke(),
		AlternativeCosts: []game.AlternativeCost{
			TapInsteadPaying("{1}", 1, "an untapped creature you control", Creature()),
		},
	})
	g, me, _ := tapTable(t)
	spell := handCardOf(g, me, "Test Convoke Tap", "Sorcery", "{5}", oracle)
	bear := pr7Creature(g, me.ID, "Bear", 2, "W")
	cast := func(convoke uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{
			AlternativeCost: tapAltCostKey, AltCostIDs: []uuid.UUID{bear}, TapIDs: []uuid.UUID{convoke}, Strict: true,
		})
	}
	if err := cast(bear); err == nil {
		t.Fatal("one creature paid a tap cost and convoke")
	}
	if tappedForTest(t, g, bear) {
		t.Error("the refused cast tapped the creature")
	}
	elf := pr7Creature(g, me.ID, "Elf", 1, "G")
	if err := cast(elf); err != nil {
		t.Fatalf("a second creature convoking the {1}: %v", err)
	}
	if !tappedForTest(t, g, bear) || !tappedForTest(t, g, elf) {
		t.Error("both creatures were not tapped")
	}
}

// Sivvi's Valor: damage to the target creature is dealt to you instead.
func TestSivvisValorRedirectsTheCreaturesDamageToYou(t *testing.T) {
	g, me, opp := tapTable(t)
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	valor := handCardOf(g, me, "Sivvi's Valor", "Instant", "{2}{W}", sivvisValorOracle)
	payer := pr7Creature(g, me.ID, "Payer", 1, "W")
	hero := pr7Creature(g, me.ID, "Hero", 2, "W")
	src := pr7Creature(g, opp.ID, "Source", 3, "R")
	if err := castTapping(g, me, valor, tapAltCostKey, cardRefs(hero), payer); err != nil {
		t.Fatalf("Sivvi's Valor for a creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	life := me.Life
	pr6Damage(t, g, src, hero, 3)
	if pr6Marked(g, hero) != 0 {
		t.Errorf("the target was dealt %d damage, want none", pr6Marked(g, hero))
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

// Angelic Favor: only during combat, whichever cost; the Angel arrives.
func TestAngelicFavorOnlyDuringCombat(t *testing.T) {
	g, me, _ := tapTable(t)
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	favor := handCardOf(g, me, "Angelic Favor", "Instant", "{3}{W}", angelicFavorOracle)
	payer := pr7Creature(g, me.ID, "Payer", 1, "W")
	if err := castTapping(g, me, favor, tapAltCostKey, nil, payer); err == nil {
		t.Fatal("Angelic Favor was cast in a main phase")
	}
	for i := 0; game.PhaseOf(g.Turn.Step) != game.PhaseCombat; i++ {
		if i > 8 {
			t.Fatal("never reached combat")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := castTapping(g, me, favor, tapAltCostKey, nil, payer); err != nil {
		t.Fatalf("Angelic Favor in combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	angels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && c.Controller == me.ID {
			angels++
		}
	}
	if angels != 1 {
		t.Errorf("%d Angels, want 1", angels)
	}
}

// Ramosian Rally pumps the team, the tapped creature too; Lashknife grants
// first strike to the creature that paid for it.
func TestRamosianRallyAndLashknifeForATappedCreature(t *testing.T) {
	g, me, _ := tapTable(t)
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	rally := handCardOf(g, me, "Ramosian Rally", "Instant", "{3}{W}", ramosianRallyOracle)
	a := pr7Creature(g, me.ID, "A", 1, "W")
	b := pr7Creature(g, me.ID, "B", 1, "W")
	if err := castTapping(g, me, rally, tapAltCostKey, nil, a); err != nil {
		t.Fatalf("Ramosian Rally: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c := findBattlefieldCardForTest(g, id); c == nil || c.Effective().Power != 2 {
			t.Errorf("%v did not get +1/+1", id)
		}
	}
	knife := handCardOf(g, me, "Lashknife", "Enchantment — Aura", "{1}{W}", lashknifeOracle)
	if err := castTapping(g, me, knife, tapAltCostKey, cardRefs(b), b); err != nil {
		t.Fatalf("Lashknife on the creature that paid for it: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, b); c == nil || !game.HasKeyword(c, "first strike") {
		t.Error("the enchanted creature does not have first strike")
	}
}

// Prismatic Strands: flashback taps a WHITE creature from the graveyard,
// the colour is chosen as it resolves, and the card is exiled.
func TestPrismaticStrandsFlashesBackByTappingAWhiteCreature(t *testing.T) {
	g, me, opp := tapTable(t)
	strands := graveyardCardOf(g, me, "Prismatic Strands", "Instant", "{2}{W}", prismaticStrandsOracle)
	green := pr7Creature(g, me.ID, "Green", 2, "G")
	white := pr7Creature(g, me.ID, "White", 2, "W")
	red := pr7Creature(g, opp.ID, "Red", 3, "R")
	blue := pr7Creature(g, opp.ID, "Blue", 3, "U")
	if err := castTapping(g, me, strands, "flashback", nil, green); err == nil {
		t.Fatal("a green creature paid \"tap an untapped white creature\"")
	}
	if err := castTapping(g, me, strands, "flashback", nil, white); err != nil {
		t.Fatalf("Prismatic Strands' flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(strands) {
		t.Error("the flashed-back Strands was not exiled")
	}
	life := me.Life
	pr7Hit(t, g, red, me.ID, 3)
	pr7Hit(t, g, blue, me.ID, 3)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d (red prevented, blue dealt)", me.Life, life-3)
	}
}

// Battle Screech: the two Birds the hand cast makes can tap for the
// flashback (the ruling), with a third white creature.
func TestBattleScreechFlashesBackWithItsOwnBirds(t *testing.T) {
	g, me, _ := tapTable(t)
	screech := handCardOf(g, me, "Battle Screech", "Sorcery", "{2}{W}{W}", battleScreechOracle)
	apaMana(me, "W", "W", "C", "C")
	if err := g.CastSpell(me.ID, screech, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Battle Screech from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	var birds []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.Controller == me.ID {
			birds = append(birds, c.InstanceID)
		}
	}
	if len(birds) != 2 || !me.Graveyard.Contains(screech) {
		t.Fatalf("birds %d, Screech in graveyard %v", len(birds), me.Graveyard.Contains(screech))
	}
	if err := castTapping(g, me, screech, "flashback", nil, birds...); err == nil {
		t.Fatal("two creatures paid \"tap three untapped white creatures\"")
	}
	knight := pr7Creature(g, me.ID, "Knight", 2, "W")
	if err := castTapping(g, me, screech, "flashback", nil, birds[0], birds[1], knight); err != nil {
		t.Fatalf("the flashback with the two new Birds: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.Controller == me.ID {
			n++
		}
	}
	if n != 4 || !g.Exile.Contains(screech) {
		t.Errorf("birds %d (want 4), Screech exiled %v", n, g.Exile.Contains(screech))
	}
}

// Group Project's flashback taps any three creatures.
func TestGroupProjectFlashesBackForThreeCreatures(t *testing.T) {
	g, me, _ := tapTable(t)
	project := graveyardCardOf(g, me, "Group Project", "Sorcery", "{1}{W}", groupProjectOracle)
	a := pr7Creature(g, me.ID, "A", 1, "G")
	b := pr7Creature(g, me.ID, "B", 1, "B")
	c := pr7Creature(g, me.ID, "C", 1, "R")
	if err := castTapping(g, me, project, "flashback", nil, a, b, c); err != nil {
		t.Fatalf("Group Project's flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Spirit" && card.Controller == me.ID {
			if card.Power != 2 || card.Toughness != 2 {
				t.Errorf("Spirit %d/%d, want 2/2", card.Power, card.Toughness)
			}
			return
		}
	}
	t.Error("no Spirit token")
}

// The wire and the bot: Orim's Cure's offer ships tap_options (the
// untapped creatures the caster controls, min = max = 1, not pay_options),
// a bystander sees none of it, and every enumerated tap payment is one
// the engine accepts.
func TestOrimsCureOfferShipsTapOptionsAndTheEnumeratorPaysIt(t *testing.T) {
	g, me, opp := tapTable(t)
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	cure := handCardOf(g, me, "Orim's Cure", "Instant", "{1}{W}", orimsCureOracle)
	mine := pr7Creature(g, me.ID, "Mine", 2, "W")
	tapped := pr7Creature(g, me.ID, "Tapped", 2, "W")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, tapped).Tapped = true })
	pr7Creature(g, opp.ID, "Theirs", 2, "W")

	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID != cure.String() {
				continue
			}
			for j := range s.Hand.Cards[i].AlternativeCosts {
				if s.Hand.Cards[i].AlternativeCosts[j].Key == tapAltCostKey {
					offer = &s.Hand.Cards[i].AlternativeCosts[j]
				}
			}
		}
	}
	if offer == nil || offer.TapOptions == nil {
		t.Fatalf("Orim's Cure's tap offer is not on the wire: %+v", offer)
	}
	if offer.PayOptions != nil || offer.SacrificeOptions != nil {
		t.Errorf("a tap offer shipped pay_options or sacrifice_options: %+v", offer)
	}
	to := offer.TapOptions
	if to.Min != 1 || to.Max != 1 || len(to.Cards) != 1 || to.Cards[0] != mine.String() {
		t.Errorf("tap_options = %+v, want 1/1 over the one untapped creature the caster controls", to)
	}
	if offer.PayLabel != "an untapped creature you control" || !strings.HasPrefix(offer.Label, "Tap an untapped creature") {
		t.Errorf("label %q, pay_label %q", offer.Label, offer.PayLabel)
	}

	type move struct {
		AlternativeCost string   `json:"alternative_cost"`
		AltCostIDs      []string `json:"alt_cost_ids"`
	}
	found := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != cure {
			continue
		}
		var p move
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != tapAltCostKey {
			continue
		}
		found++
		if len(p.AltCostIDs) != 1 || p.AltCostIDs[0] != mine.String() {
			t.Errorf("enumerated tap payment %v, want the untapped creature", p.AltCostIDs)
		}
		if !strings.Contains(m.Label, "tapping Mine") {
			t.Errorf("move label %q does not name the tapped creature", m.Label)
		}
	}
	if found == 0 {
		t.Fatal("the enumerator offers no Orim's Cure cast for its tap cost")
	}
}

// effects.Register holds a tap alternative cost to the shape the cast path
// reads.
func TestRegisterRefusesMalformedTapAlternativeCosts(t *testing.T) {
	another := TapInstead(1, "another untapped creature you control", nil, Creature())
	another.TapOthers.ExcludeSource = true
	xTap := TapInstead(1, "X untapped creatures you control", nil, Creature())
	xTap.TapOthers.Count = 0
	xTap.TapOthers.Filter.CountFromX = true
	two := TapInstead(1, "an untapped creature you control", nil, Creature())
	two.Sacrifice = sacrificeSpec("a creature", Creature()).WithCount(1, 1)
	zero := TapInstead(0, "no creatures", nil, Creature())
	for _, c := range []struct {
		name string
		ac   game.AlternativeCost
		want string
	}{
		{"another", another, "ExcludeSource"},
		{"X", xTap, "taps X permanents"},
		{"two components", two, "card-shaped payments"},
		{"zero", zero, "positive count"},
	} {
		msg := registerPanics(Spec{OracleID: "test-0135-tap-" + c.name, Name: "Test " + c.name, AlternativeCosts: []game.AlternativeCost{c.ac}})
		if !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}
