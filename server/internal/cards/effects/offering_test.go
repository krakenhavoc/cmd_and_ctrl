package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// offering_test.go — #5657: offering (CR 702.48) and Blast-Furnace
// Hellkite against the real catalog. The offering is an optional
// additional cost that sacrifices one artifact, takes that artifact's
// mana cost off the total cost symbol by symbol (CR 702.48c) after any
// increase (CR 601.2f), and opens the instant-speed window only for the
// cast that announces it (CR 702.48a, 601.3c).

const blastFurnaceHellkiteOracle = "c13079d5-6749-43e5-a8ee-ef3af31ff679"

// offeringHellkite puts Blast-Furnace Hellkite in p's hand.
func offeringHellkite(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Blast-Furnace Hellkite", TypeLine: "Creature — Dragon",
			ManaCost: "{7}{R}{R}", OracleID: blastFurnaceHellkiteOracle, Power: 5, Toughness: 5,
			Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}

// offeringArtifact puts an artifact with mana cost `cost` on owner's
// battlefield.
func offeringArtifact(g *game.Game, owner uuid.UUID, name, cost string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Artifact", ManaCost: cost})
}

// offeringParams is a cast that announces the offering over `sac`.
func offeringParams(strict bool, sac ...uuid.UUID) game.CastSpellParams {
	return game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: sac, Strict: strict, AutoTap: strict}
}

// offeringPrice is what the engine's one pricer charges for `card` under
// `params`.
func offeringPrice(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, params game.CastSpellParams) string {
	t.Helper()
	var out string
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, params)
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total.String()
	})
	return out
}

func TestBlastFurnaceHellkiteIsFullAndDeclaresArtifactOffering(t *testing.T) {
	spec, ok := Lookup(blastFurnaceHellkiteOracle)
	if !ok {
		t.Fatal("Blast-Furnace Hellkite is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
	costs := game.OptionalCostsFor(blastFurnaceHellkiteOracle)
	if len(costs) != 1 {
		t.Fatalf("%d optional costs, want the offering", len(costs))
	}
	oc := costs[0]
	if oc.Key != OfferingKey || oc.Label != "Artifact offering" || !oc.Optional || !oc.AsThoughFlash ||
		!oc.ReducedBySacrificedManaCost || oc.Sacrifice == nil || game.SacrificeCostCount(oc.Sacrifice) != 1 || oc.ManaCost != "" {
		t.Errorf("optional cost %+v, want an artifact offering", oc)
	}
}

// CR 702.48c: generic reduces generic; a coloured or colourless symbol
// reduces mana of the same type, and any excess reduces generic. CR
// 202.1b: a token has no mana cost. CR 107.3g: an {X} on the battlefield
// is 0. CR 601.2f: never below {0}, which the pricer prints as the empty
// cost.
func TestOfferingIsPricedByTheSacrificedPermanentsManaCost(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	if got := offeringPrice(t, g, me, kite, game.CastSpellParams{}); got != "{7}{R}{R}" {
		t.Fatalf("without the offering: %s, want {7}{R}{R}", got)
	}
	for _, c := range []struct{ name, cost, want string }{
		{"generic off generic", "{4}", "{3}{R}{R}"},
		{"a red symbol off a red symbol", "{2}{R}", "{5}{R}"},
		{"both red symbols", "{R}{R}", "{7}"},
		{"a third red symbol off generic", "{R}{R}{R}", "{6}"},
		{"a blue symbol off generic", "{1}{U}", "{5}{R}{R}"},
		{"a colourless symbol off generic", "{C}", "{6}{R}{R}"},
		{"a hybrid symbol takes its red half", "{R/G}", "{7}{R}"},
		{"an X is 0", "{X}{2}", "{5}{R}{R}"},
		{"more generic than the spell has", "{10}", "{R}{R}"},
		{"more than the whole cost", "{8}{R}{R}{R}", ""},
		{"no mana cost", "", "{7}{R}{R}"},
	} {
		sac := offeringArtifact(g, me.ID, c.name, c.cost)
		if got := offeringPrice(t, g, me, kite, offeringParams(false, sac)); got != c.want {
			t.Errorf("%s (%q): price %s, want %s", c.name, c.cost, got, c.want)
		}
	}
	// A face-down permanent has no mana cost (CR 708.2).
	down := apaPush(g, me.ID, me.ID, game.Card{Name: "Hidden", TypeLine: "Artifact", ManaCost: "{6}", FaceDown: true})
	if got := offeringPrice(t, g, me, kite, offeringParams(false, down)); got != "{7}{R}{R}" {
		t.Errorf("a face-down artifact: price %s, want {7}{R}{R}", got)
	}
}

// CR 601.2f: the increase is added before the reduction is taken off.
func TestOfferingReductionComesAfterAnIncrease(t *testing.T) {
	g, me, opp := emergeTable(t)
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Sphere of Resistance", TypeLine: "Artifact", ManaCost: "{2}", OracleID: sphereOfResistanceOracle})
	kite := offeringHellkite(g, me)
	four := offeringArtifact(g, me.ID, "Four", "{4}")
	eight := offeringArtifact(g, me.ID, "Eight", "{8}")
	if got := offeringPrice(t, g, me, kite, offeringParams(false, four)); got != "{4}{R}{R}" {
		t.Errorf("over a four-drop under Sphere: %s, want {4}{R}{R}", got)
	}
	if got := offeringPrice(t, g, me, kite, offeringParams(false, eight)); got != "{R}{R}" {
		t.Errorf("over an eight-drop under Sphere: %s, want {R}{R}", got)
	}
}

// CR 702.48a, 601.3c: the offering opens the instant-speed window for the
// cast that announces it and no other. End to end under the strict gate:
// five Mountains pay {3}{R}{R} over a four-drop, which is sacrificed with
// the spell on the stack (CR 702.48b, 601.2h).
func TestBlastFurnaceHellkiteIsOfferedAtInstantSpeed(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	four := offeringArtifact(g, me.ID, "Four", "{4}")
	lands := emergeLands(g, me.ID, 4, "Mountain")
	ssToBeginCombat(t, g)
	if err := g.CastSpell(me.ID, kite, game.CastSpellParams{Strict: true, AutoTap: true}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("cast without the offering in combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, kite, offeringParams(true, four)); err == nil {
		t.Fatal("the Hellkite was offered over a four-drop with four Mountains for {3}{R}{R}")
	}
	if findBattlefieldCardForTest(g, four) == nil || !me.Hand.Contains(kite) || tappedForTest(t, g, lands[0]) {
		t.Fatal("a refused cast paid something")
	}
	lands = append(lands, emergeLands(g, me.ID, 1, "Mountain")...)
	if err := g.CastSpell(me.ID, kite, offeringParams(true, four)); err != nil {
		t.Fatalf("offering a four-drop with five Mountains: %v", err)
	}
	if findBattlefieldCardForTest(g, four) != nil || !me.Graveyard.Contains(four) {
		t.Fatal("the artifact was not sacrificed with the spell on the stack")
	}
	for _, id := range lands {
		if !tappedForTest(t, g, id) {
			t.Error("a Mountain was left untapped paying {3}{R}{R}")
		}
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, kite) == nil {
		t.Error("Blast-Furnace Hellkite did not resolve")
	}
}

// Without the offering the Hellkite is a nine-mana creature at sorcery
// speed, and nothing is sacrificed.
func TestBlastFurnaceHellkiteCastsForItsManaCostWithoutTheOffering(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	rock := offeringArtifact(g, me.ID, "Rock", "{4}")
	emergeLands(g, me.ID, 8, "Mountain")
	if err := g.CastSpell(me.ID, kite, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("the Hellkite was cast for eight mana")
	}
	emergeLands(g, me.ID, 1, "Mountain")
	if err := g.CastSpell(me.ID, kite, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast for {7}{R}{R} with nine Mountains: %v", err)
	}
	if findBattlefieldCardForTest(g, rock) == nil {
		t.Error("a cast without the offering sacrificed the artifact")
	}
}

// The offering sacrifices exactly one artifact its caster controls.
func TestOfferingRefusesTheWrongPayment(t *testing.T) {
	g, me, opp := emergeTable(t)
	kite := offeringHellkite(g, me)
	theirs := offeringArtifact(g, opp.ID, "Theirs", "{9}")
	bear := emergeFodder(g, me.ID, "Bear", "{9}", 2, 2)
	land := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	one := offeringArtifact(g, me.ID, "One", "{9}")
	two := offeringArtifact(g, me.ID, "Two", "{9}")
	for name, sac := range map[string][]uuid.UUID{
		"an opponent's artifact": {theirs},
		"a creature":             {bear},
		"a land":                 {land},
		"nothing":                nil,
		"two artifacts":          {one, two},
	} {
		if err := g.CastSpell(me.ID, kite, offeringParams(false, sac...)); err == nil {
			t.Errorf("%s paid the offering", name)
		}
	}
	if err := g.CastSpell(me.ID, kite, game.CastSpellParams{SacrificeIDs: []uuid.UUID{one}}); err == nil {
		t.Error("an artifact was sacrificed to a cast that announced no offering")
	}
	if !me.Hand.Contains(kite) || findBattlefieldCardForTest(g, one) == nil {
		t.Fatal("a refused cast paid something")
	}
}

// The payment record on the stack item (the announced cost and the
// permanent it sacrificed) survives a restore, and the spell resolves.
func TestOfferedSpellSurvivesARestore(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	rock := offeringArtifact(g, me.ID, "Rock", "{7}{R}{R}")
	ssToBeginCombat(t, g)
	if err := g.CastSpell(me.ID, kite, offeringParams(true, rock)); err != nil {
		t.Fatalf("offering a {7}{R}{R} artifact for {0}: %v", err)
	}
	restored := restoreRoundTrip(t, g, true)
	var paid game.PaidCost
	restored.ReadSnapshot(func() {
		if item := restored.StackMeta[kite]; item != nil {
			paid = item.Paid
		}
	})
	if len(paid.OptionalCosts) != 1 || paid.OptionalCosts[0] != 0 || len(paid.SacrificedObjects) != 1 || paid.SacrificedObjects[0].ID != rock {
		t.Fatalf("restored PaidCost %+v, want the offering over the artifact", paid)
	}
	passPriorityAroundTable(t, restored)
	if findBattlefieldCardForTest(restored, kite) == nil {
		t.Error("the restored Hellkite did not resolve")
	}
}

// The enumerator prices each payment on its own: in combat, with five
// Mountains, the Hellkite is offered only with the offering announced and
// only over the artifact that makes it affordable.
func TestOfferingEnumeratorOffersOnlyAffordablePayments(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	emergeLands(g, me.ID, 5, "Mountain")
	four := offeringArtifact(g, me.ID, "Four", "{4}")
	six := offeringArtifact(g, me.ID, "Six", "{4}{R}{R}")
	two := offeringArtifact(g, me.ID, "Two", "{2}")
	free := offeringArtifact(g, me.ID, "Free", "")
	offered := func() (map[uuid.UUID]bool, int) {
		with, without := map[uuid.UUID]bool{}, 0
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != legal.TypeCastSpell || m.Source != kite {
				continue
			}
			var p struct {
				OptionalCosts []int       `json:"optional_costs"`
				SacrificeIDs  []uuid.UUID `json:"sacrifice_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			if len(p.OptionalCosts) == 0 {
				without++
				continue
			}
			if len(p.SacrificeIDs) != 1 {
				t.Fatalf("an offered cast names %d permanents", len(p.SacrificeIDs))
			}
			with[p.SacrificeIDs[0]] = true
		}
		return with, without
	}
	check := func(when string) {
		with, without := offered()
		if without != 0 {
			t.Errorf("%s: the nine-mana cast is offered with five Mountains", when)
		}
		if !with[four] || !with[six] {
			t.Errorf("%s: offered %v, want the four- and six-mana artifacts", when, with)
		}
		if with[two] || with[free] {
			t.Errorf("%s: offered %v, which includes a payment the seat can't afford", when, with)
		}
	}
	check("main phase")
	ssToBeginCombat(t, g)
	check("combat")
}

// The view: outside a main phase the printed claim is open only with the
// offering announced (`opens_flash_required`), and the offer carries the
// artifacts that can pay it.
func TestOfferingOnTheWireOpensTheWindow(t *testing.T) {
	g, me, _ := emergeTable(t)
	kite := offeringHellkite(g, me)
	rock := offeringArtifact(g, me.ID, "Rock", "{4}")
	card := func() *protocol.CardView {
		v := protocol.ViewOfGameFor(g, me.ID.String())
		for _, s := range v.Seats {
			for i := range s.Hand.Cards {
				if s.Hand.Cards[i].InstanceID == kite.String() {
					return &s.Hand.Cards[i]
				}
			}
		}
		t.Fatal("the Hellkite is not in the view")
		return nil
	}
	if c := card(); c.OpensFlashRequired {
		t.Error("the offering is required in a main phase")
	}
	ssToBeginCombat(t, g)
	c := card()
	if !c.OpensFlashRequired {
		t.Error("in combat the cast is not marked open only with the offering")
	}
	if len(c.OptionalCosts) != 1 || c.OptionalCosts[0].Key != OfferingKey || !c.OptionalCosts[0].OpensFlash ||
		c.OptionalCosts[0].SacrificeOptions == nil || len(c.OptionalCosts[0].SacrificeOptions.Cards) != 1 ||
		c.OptionalCosts[0].SacrificeOptions.Cards[0] != rock.String() {
		t.Errorf("optional costs %+v, want the offering over the one artifact", c.OptionalCosts)
	}
}

// "Creatures attacking your opponents have double strike": every creature
// attacking a player who is the Hellkite's controller's opponent, whoever
// controls it, and not one attacking its controller.
func TestBlastFurnaceHellkiteGivesDoubleStrikeToCreaturesAttackingItsOpponents(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Turn.ActiveSeat
	attacker, kiteSide, other := g.Seats[a], g.Seats[(a+1)%4], g.Seats[(a+2)%4]
	dftABPush(g, kiteSide.ID, "Blast-Furnace Hellkite", "Creature — Dragon", blastFurnaceHellkiteOracle, 5, 5)
	atOther := dftABPush(g, attacker.ID, "Bear A", "Creature — Bear", "", 2, 2)
	atKite := dftABPush(g, attacker.ID, "Bear B", "Creature — Bear", "", 2, 2)
	idle := dftABPush(g, attacker.ID, "Bear C", "Creature — Bear", "", 2, 2)
	if hasEffectiveKeyword(t, g, atOther, "double strike") {
		t.Fatal("a creature that is not attacking has double strike")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atOther, other.ID); err != nil {
		t.Fatalf("attack %v: %v", other.ID, err)
	}
	if err := g.DeclareAttacker(atKite, kiteSide.ID); err != nil {
		t.Fatalf("attack %v: %v", kiteSide.ID, err)
	}
	lockInAttacks(t, g)
	if !hasEffectiveKeyword(t, g, atOther, "double strike") {
		t.Error("another player's creature attacking one of the Hellkite's controller's opponents lacks double strike")
	}
	if hasEffectiveKeyword(t, g, atKite, "double strike") {
		t.Error("a creature attacking the Hellkite's controller has double strike")
	}
	if hasEffectiveKeyword(t, g, idle, "double strike") {
		t.Error("a creature that did not attack has double strike")
	}
}

// And the Hellkite's controller's own attackers have it.
func TestBlastFurnaceHellkiteGivesItsControllersAttackersDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	a := g.Turn.ActiveSeat
	me, opp := g.Seats[a], g.Seats[(a+1)%4]
	dftABPush(g, me.ID, "Blast-Furnace Hellkite", "Creature — Dragon", blastFurnaceHellkiteOracle, 5, 5)
	bear := dftABPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the controller's attacker lacks double strike")
	}
}

// effects.Register holds the flag to an optional sacrifice of one
// permanent.
func TestRegisterRefusesOfferingWithoutOneOptionalSacrifice(t *testing.T) {
	two := Offering("artifact", Artifact())
	two.Sacrifice = two.Sacrifice.WithCount(2, 2)
	none := Kicker("{4}")
	none.ReducedBySacrificedManaCost = true
	mandatory := SacrificeCost("an artifact", Artifact())
	mandatory.ReducedBySacrificedManaCost = true
	branch := *SacrificeCost("an artifact", Artifact()).Keyed("sacrifice")
	branch.ReducedBySacrificedManaCost = true
	life := *(&game.AdditionalCost{PayLife: 3, Label: "Pay 3 life"}).Keyed("life")
	for name, spec := range map[string]Spec{
		"two":       {OptionalCosts: []game.AdditionalCost{two}},
		"none":      {OptionalCosts: []game.AdditionalCost{none}},
		"mandatory": {AdditionalCost: mandatory},
		"branch":    {AdditionalCost: &game.AdditionalCost{Either: []game.AdditionalCost{branch, life}}},
	} {
		spec.OracleID, spec.Name = "test-5657-offering-"+name, "Test "+name
		if msg := registerPanics(spec); !strings.Contains(msg, "CR 702.48a") {
			t.Errorf("%s: panic %q, want the offering guard", name, msg)
		}
	}
}
