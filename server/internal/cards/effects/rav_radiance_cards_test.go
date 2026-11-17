package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_radiance_cards_test.go — radiance (#5736): "target X and each
// other X that shares a color with it", and the ten Ravnica: City of
// Guilds cards built on it (#5600). The first tests pin the shared set
// (radianceSet) through Cleansing Beam, the plainest card; the rest play
// each card's own text.

const (
	ravRadBatheInLightOracle    = "8e862a8e-7954-43ad-9bc0-09ce7b9859f0"
	ravRadBrightflameOracle     = "6473eac1-6786-47dd-8124-9f1c4b790ddf"
	ravRadCleansingBeamOracle   = "5c208258-f987-4d41-8eee-62e188d47ffb"
	ravRadInciteHysteriaOracle  = "0a992ad5-136f-482d-8390-fdfe745fe764"
	ravRadLeaveNoTraceOracle    = "a9367efe-82f2-43e4-9692-205567be6a0c"
	ravRadRallyOracle           = "5d4d064e-c4f4-4eca-955c-4539e66d230b"
	ravRadSurgeOfZealOracle     = "24dca209-9781-4bde-ad61-18c20e8ccbd8"
	ravRadWojekApothecaryOracle = "23ea480c-99e3-4cbf-8d24-9ea92230f57b"
	ravRadWojekEmbermageOracle  = "ac8630ac-defb-4774-ac71-d35fe02a7324"
	ravRadWojekSirenOracle      = "164b21ce-834b-4ec2-9094-33e51a46f9ef"
	ravRadCeruleanWispsOracle   = "a6605c50-558e-417c-8c75-6c45b06d6e13"
	ravRadFinalShowdownOracle   = "7e7ec3d6-a84f-4cc3-93f4-4d181d41e126"
	ravRadStartingLife          = 40
)

func TestRavRadianceEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ravRadBatheInLightOracle:    "Bathe in Light",
		ravRadBrightflameOracle:     "Brightflame",
		ravRadCleansingBeamOracle:   "Cleansing Beam",
		ravRadInciteHysteriaOracle:  "Incite Hysteria",
		ravRadLeaveNoTraceOracle:    "Leave No Trace",
		ravRadRallyOracle:           "Rally the Righteous",
		ravRadSurgeOfZealOracle:     "Surge of Zeal",
		ravRadWojekApothecaryOracle: "Wojek Apothecary",
		ravRadWojekEmbermageOracle:  "Wojek Embermage",
		ravRadWojekSirenOracle:      "Wojek Siren",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// ravRadCreature seeds a 3/3 non-catalog creature of the given colours
// (none for a colourless one) with any keywords.
func ravRadCreature(g *game.Game, owner uuid.UUID, name string, colors []string, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Colors: colors, Keywords: keywords,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner,
	})
}

// ravRadDamage reports the damage marked on each creature, in order.
func ravRadDamage(g *game.Game, ids ...uuid.UUID) []int {
	out := make([]int, len(ids))
	for i, id := range ids {
		out[i] = damageMarkedOn(g, id)
	}
	return out
}

func ravRadSame(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func ravRadCastBeam(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	// With its mana cost, so the spell is red (protection reads it).
	paCast(t, g, "Cleansing Beam", "Instant", "{4}{R}", ravRadCleansingBeamOracle, game.CastSpellParams{Targets: cardRefs(target)})
}

// --- the shared set, through Cleansing Beam -------------------------------

// A multicoloured target reaches every creature that has ANY of its
// colours, whoever controls it, and nothing else.
func TestRavRadianceMulticolouredTargetReachesEachCreatureSharingAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Boros Target", []string{"R", "W"})
	red := ravRadCreature(g, me.ID, "My Red", []string{"R"})
	white := ravRadCreature(g, opp.ID, "Their White", []string{"W"})
	greenWhite := ravRadCreature(g, g.Seats[2].ID, "Selesnya", []string{"G", "W"})
	blue := ravRadCreature(g, opp.ID, "Their Blue", []string{"U"})
	green := ravRadCreature(g, me.ID, "My Green", []string{"G"})
	colourless := ravRadCreature(g, opp.ID, "Golem", nil)

	ravRadCastBeam(t, g, target)
	passPriorityAroundTable(t, g)

	got := ravRadDamage(g, target, red, white, greenWhite, blue, green, colourless)
	if want := []int{2, 2, 2, 2, 0, 0, 0}; !ravRadSame(got, want) {
		t.Errorf("damage (target, red, white, green-white, blue, green, colourless) = %v, want %v", got, want)
	}
}

// The colours compared are the target's as the spell RESOLVES: a red
// target turned blue in response brings the blue creatures, not the red.
func TestRavRadianceReadsTheTargetsColoursAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	red := ravRadCreature(g, opp.ID, "Other Red", []string{"R"})
	blue := ravRadCreature(g, me.ID, "My Blue", []string{"U"})

	ravRadCastBeam(t, g, target)
	// In response: "Target creature becomes blue until end of turn."
	castCatalogSpell(t, g, "Cerulean Wisps", "Instant", ravRadCeruleanWispsOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	got := ravRadDamage(g, target, red, blue)
	if want := []int{2, 0, 2}; !ravRadSame(got, want) {
		t.Errorf("damage (recoloured target, red, blue) = %v, want %v", got, want)
	}
}

// A creature that arrives in response is in the set; the set is read as
// the spell resolves, not as it was cast.
func TestRavRadianceIncludesACreatureThatArrivedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	ravRadCastBeam(t, g, target)
	late := ravRadCreature(g, opp.ID, "Late Red", []string{"R"})
	passPriorityAroundTable(t, g)

	if got := ravRadDamage(g, target, late); !ravRadSame(got, []int{2, 2}) {
		t.Errorf("damage (target, arrived in response) = %v, want [2 2]", got)
	}
}

// The target leaving before resolution leaves no legal target, so the
// spell does nothing (CR 608.2b): no other creature is touched.
func TestRavRadianceDoesNothingWhenTheTargetHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	red := ravRadCreature(g, opp.ID, "Other Red", []string{"R"})
	mine := ravRadCreature(g, me.ID, "My Red", []string{"R"})

	ravRadCastBeam(t, g, target)
	hobDestroy(t, g, target)
	passPriorityAroundTable(t, g)

	if got := ravRadDamage(g, red, mine); !ravRadSame(got, []int{0, 0}) {
		t.Errorf("damage to the others after the target left = %v, want [0 0]", got)
	}
}

// A colourless target shares a colour with nothing, not even another
// colourless creature: only the target is affected.
func TestRavRadianceColourlessTargetAffectsOnlyItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Golem", nil)
	other := ravRadCreature(g, opp.ID, "Other Golem", nil)
	red := ravRadCreature(g, me.ID, "My Red", []string{"R"})

	ravRadCastBeam(t, g, target)
	passPriorityAroundTable(t, g)

	if got := ravRadDamage(g, target, other, red); !ravRadSame(got, []int{2, 0, 0}) {
		t.Errorf("damage (colourless target, colourless other, red) = %v, want [2 0 0]", got)
	}
}

// The others are not targeted (CR 115.10a): hexproof and shroud don't
// keep them out. Protection from red still prevents the damage itself
// (CR 702.16e).
func TestRavRadianceReachesHexproofAndShroudOthers(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	hexproof := ravRadCreature(g, opp.ID, "Hexproof Red", []string{"R"}, "hexproof")
	shroud := ravRadCreature(g, opp.ID, "Shroud Red", []string{"R"}, "shroud")
	proRed := ravRadCreature(g, opp.ID, "Pro-Red Red", []string{"R"}, "protection from red")

	ravRadCastBeam(t, g, target)
	passPriorityAroundTable(t, g)

	if got := ravRadDamage(g, target, hexproof, shroud, proRed); !ravRadSame(got, []int{2, 2, 2, 0}) {
		t.Errorf("damage (target, hexproof, shroud, protection from red) = %v, want [2 2 2 0]", got)
	}
}

// --- Brightflame ---------------------------------------------------------------

// "You gain life equal to the damage dealt this way" is what was actually
// dealt: a one-point shield takes 1 off one creature, protection from
// white takes all of another's, and a 1/1 dealt 3 still counts 3.
func TestRavRadianceBrightflameGainsTheDamageActuallyDealt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their White 1/1", TypeLine: "Creature — Test", Colors: []string{"W"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	shielded := ravRadCreature(g, opp.ID, "Shielded White", []string{"W"})
	proWhite := ravRadCreature(g, opp.ID, "Pro-White White", []string{"W"}, "protection from white")
	blue := ravRadCreature(g, opp.ID, "Their Blue", []string{"U"})
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, shielded, 1, false, "shield") })

	paCast(t, g, "Brightflame", "Sorcery", "{X}{R}{R}{W}{W}", ravRadBrightflameOracle,
		game.CastSpellParams{XValue: 3, Targets: cardRefs(target)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) {
		t.Error("the 1/1 target survived 3 damage")
	}
	if got := ravRadDamage(g, shielded, proWhite, blue); !ravRadSame(got, []int{2, 0, 0}) {
		t.Errorf("damage (shielded, protection from white, blue) = %v, want [2 0 0]", got)
	}
	if want := ravRadStartingLife + 3 + 2; me.Life != want {
		t.Errorf("Brightflame's caster is at %d life, want %d (3 + 2 dealt)", me.Life, want)
	}
}

// A damage replacement counts too: under Angrath's Marauders ("it deals
// double that damage instead") X = 1 deals 2 to each, and the gain is
// what was dealt.
func TestRavRadianceBrightflameGainsDoubledDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", b14AngrathsMaraudersOracle, 4, 4)
	target := ravRadCreature(g, opp.ID, "Their White", []string{"W"})
	other := ravRadCreature(g, opp.ID, "Their Boros", []string{"R", "W"})

	paCast(t, g, "Brightflame", "Sorcery", "{X}{R}{R}{W}{W}", ravRadBrightflameOracle,
		game.CastSpellParams{XValue: 1, Targets: cardRefs(target)})
	passPriorityAroundTable(t, g)

	if got := ravRadDamage(g, target, other); !ravRadSame(got, []int{2, 2}) {
		t.Errorf("damage (target, other) under a doubler = %v, want [2 2]", got)
	}
	if want := ravRadStartingLife + 4; me.Life != want {
		t.Errorf("Brightflame's caster is at %d life, want %d (2 + 2 dealt)", me.Life, want)
	}
}

func TestRavRadianceBrightflameWithItsTargetGoneGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their White", []string{"W"})
	other := ravRadCreature(g, opp.ID, "Other White", []string{"W"})
	paCast(t, g, "Brightflame", "Sorcery", "{X}{R}{R}{W}{W}", ravRadBrightflameOracle,
		game.CastSpellParams{XValue: 2, Targets: cardRefs(target)})
	hobDestroy(t, g, target)
	passPriorityAroundTable(t, g)

	if d := damageMarkedOn(g, other); d != 0 {
		t.Errorf("the other white creature took %d", d)
	}
	if me.Life != ravRadStartingLife {
		t.Errorf("a Brightflame that did nothing gained %d life", me.Life-ravRadStartingLife)
	}
}

// --- Bathe in Light --------------------------------------------------------

// The colour is chosen as the spell resolves and has nothing to do with
// the colours the creatures share.
func TestRavRadianceBatheInLightProtectsTheSetFromTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, me.ID, "My Selesnya", []string{"G", "W"})
	white := ravRadCreature(g, opp.ID, "Their White", []string{"W"})
	green := ravRadCreature(g, me.ID, "My Green", []string{"G"})
	blue := ravRadCreature(g, me.ID, "My Blue", []string{"U"})

	castCatalogSpell(t, g, "Bathe in Light", "Instant", ravRadBatheInLightOracle, cardRefs(target))
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	passPriorityAroundTable(t, g)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{{"the target", target, true}, {"a white creature", white, true}, {"a green creature", green, true}, {"a blue creature", blue, false}} {
		if got := p1g2HasKeyword(g, tc.id, "protection from red"); got != tc.want {
			t.Errorf("%s has protection from red: %v, want %v", tc.name, got, tc.want)
		}
	}
}

// --- Incite Hysteria -------------------------------------------------------

func ravRadCantBlock(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			out = c.Effective().Restrictions&game.CantBlock != 0
		}
	})
	return out
}

func TestRavRadianceInciteHysteriaGrantsCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	redBlack := ravRadCreature(g, opp.ID, "Rakdos", []string{"B", "R"})
	green := ravRadCreature(g, opp.ID, "Their Green", []string{"G"})

	castCatalogSpell(t, g, "Incite Hysteria", "Sorcery", ravRadInciteHysteriaOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	if !ravRadCantBlock(g, target) || !ravRadCantBlock(g, redBlack) {
		t.Errorf("can't block: target %v, red-black %v, want both", ravRadCantBlock(g, target), ravRadCantBlock(g, redBlack))
	}
	if ravRadCantBlock(g, green) {
		t.Error("a green creature can't block")
	}
}

// The creatures GAIN an ability, so losing all abilities afterwards takes
// it away (CR 613.7), as it would a printed "can't block".
func TestRavRadianceInciteHysteriaIsAnAbilityThatCanBeLost(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})

	castCatalogSpell(t, g, "Incite Hysteria", "Sorcery", ravRadInciteHysteriaOracle, cardRefs(target))
	passPriorityAroundTable(t, g)
	if !ravRadCantBlock(g, target) {
		t.Fatal("Incite Hysteria did not stop its target blocking")
	}
	// Final Showdown's first bullet: "All creatures lose all abilities
	// until end of turn."
	castCatalogSpellWithModes(t, g, "Final Showdown", "Instant", ravRadFinalShowdownOracle, []int{0})
	passPriorityAroundTable(t, g)
	if ravRadCantBlock(g, target) {
		t.Error("a creature that lost all its abilities still can't block")
	}
}

// --- Leave No Trace --------------------------------------------------------

func TestRavRadianceLeaveNoTraceDestroysEnchantmentsSharingAColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	enchantment := func(owner uuid.UUID, name string, colors ...string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment", Colors: colors, Owner: owner, Controller: owner,
		})
	}
	target := enchantment(opp.ID, "Their White Enchantment", "W")
	azorius := enchantment(me.ID, "My Azorius Enchantment", "W", "U")
	green := enchantment(opp.ID, "Their Green Enchantment", "G")
	whiteCreature := ravRadCreature(g, opp.ID, "Their White Creature", []string{"W"})

	castCatalogSpell(t, g, "Leave No Trace", "Instant", ravRadLeaveNoTraceOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(target) || g.Battlefield.Contains(azorius) {
		t.Errorf("still on the battlefield: target %v, white-blue enchantment %v",
			g.Battlefield.Contains(target), g.Battlefield.Contains(azorius))
	}
	if !g.Battlefield.Contains(green) || !g.Battlefield.Contains(whiteCreature) {
		t.Error("a green enchantment or a white creature was destroyed")
	}
}

// --- Rally the Righteous, Surge of Zeal, Wojek Siren -------------------------

func TestRavRadianceRallyTheRighteousUntapsAndPumpsThoseCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := ravRadCreature(g, me.ID, "My Red", []string{"R"})
	other := ravRadCreature(g, me.ID, "My Boros", []string{"R", "W"})
	green := ravRadCreature(g, me.ID, "My Green", []string{"G"})
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].Tapped = true
	}

	castCatalogSpell(t, g, "Rally the Righteous", "Instant", ravRadRallyOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	if isTapped(g, target) || isTapped(g, other) {
		t.Errorf("still tapped: target %v, other %v", isTapped(g, target), isTapped(g, other))
	}
	if !isTapped(g, green) {
		t.Error("a green creature was untapped")
	}
	for _, id := range []uuid.UUID{target, other} {
		if p, tough := ptOf(t, g, id); p != 5 || tough != 3 {
			t.Errorf("a rallied 3/3 is %d/%d, want 5/3", p, tough)
		}
	}
	if p, _ := ptOf(t, g, green); p != 3 {
		t.Errorf("a green creature's power is %d, want 3", p)
	}
}

func TestRavRadianceSurgeOfZealGrantsHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, me.ID, "My Red", []string{"R"})
	theirs := ravRadCreature(g, opp.ID, "Their Red", []string{"R"})
	blue := ravRadCreature(g, me.ID, "My Blue", []string{"U"})

	castCatalogSpell(t, g, "Surge of Zeal", "Instant", ravRadSurgeOfZealOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	if !p1g2HasKeyword(g, target, "haste") || !p1g2HasKeyword(g, theirs, "haste") {
		t.Error("a red creature has no haste")
	}
	if p1g2HasKeyword(g, blue, "haste") {
		t.Error("a blue creature has haste")
	}
}

func TestRavRadianceWojekSirenPumpsEachCreatureSharingAColour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := ravRadCreature(g, me.ID, "My White", []string{"W"})
	theirs := ravRadCreature(g, opp.ID, "Their Orzhov", []string{"W", "B"})
	black := ravRadCreature(g, opp.ID, "Their Black", []string{"B"})

	castCatalogSpell(t, g, "Wojek Siren", "Instant", ravRadWojekSirenOracle, cardRefs(target))
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{target, theirs} {
		if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("a white 3/3 is %d/%d, want 4/4", p, tough)
		}
	}
	if p, tough := ptOf(t, g, black); p != 3 || tough != 3 {
		t.Errorf("a black-only 3/3 is %d/%d, want 3/3", p, tough)
	}
}

// --- Wojek Embermage, Wojek Apothecary ---------------------------------------

func ravRadActivate(t *testing.T, g *game.Game, who, source, target uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(who, source, 0, game.ActivateAbilityParams{Targets: cardRefs(target)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The Embermage deals the damage, itself included when it shares the
// target's colour.
func TestRavRadianceWojekEmbermageDamagesTheSetItselfIncluded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wojek Embermage", TypeLine: "Creature — Human Wizard", OracleID: ravRadWojekEmbermageOracle,
		Colors: []string{"R"}, Power: 1, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	target := ravRadCreature(g, opp.ID, "Their Gruul", []string{"R", "G"})
	green := ravRadCreature(g, opp.ID, "Their Green", []string{"G"})
	blue := ravRadCreature(g, opp.ID, "Their Blue", []string{"U"})

	ravRadActivate(t, g, me.ID, mage, target)

	if got := ravRadDamage(g, target, green, mage, blue); !ravRadSame(got, []int{1, 1, 1, 0}) {
		t.Errorf("damage (target, green, the Embermage, blue) = %v, want [1 1 1 0]", got)
	}
	if !isTapped(g, mage) {
		t.Error("the Embermage is untapped after paying {T}")
	}
}

func TestRavRadianceWojekApothecaryShieldsEachCreatureOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	apothecary := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wojek Apothecary", TypeLine: "Creature — Human Cleric", OracleID: ravRadWojekApothecaryOracle,
		Colors: []string{"W"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	target := ravRadCreature(g, me.ID, "My White", []string{"W"})
	theirs := ravRadCreature(g, opp.ID, "Their White", []string{"W"}, "hexproof")
	blue := ravRadCreature(g, me.ID, "My Blue", []string{"U"})

	ravRadActivate(t, g, me.ID, apothecary, target)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{target, theirs, apothecary, blue} {
			if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 1); err != nil {
				t.Fatal(err)
			}
		}
	})
	passPriorityAroundTable(t, g)

	// The Apothecary is white too, so it shielded itself.
	if got := ravRadDamage(g, target, theirs, apothecary, blue); !ravRadSame(got, []int{0, 0, 0, 1}) {
		t.Errorf("damage after 1 each (target, hexproof white, the Apothecary, blue) = %v, want [0 0 0 1]", got)
	}
	if !isTapped(g, apothecary) {
		t.Error("the Apothecary is untapped after paying {T}")
	}
}
