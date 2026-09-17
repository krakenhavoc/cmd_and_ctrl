package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch16_test.go — card-level coverage for the card-coverage
// roadmap's batch 16 (#309, `edhrec_rank` 1727–1829): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b16-prefixed.

const (
	b16DeathBaronOracle           = "99024aa8-5687-4d38-8a4b-feef42d6c1ff"
	b16FlameOfAnorOracle          = "ecd49d85-9c8c-4cc9-9a83-072ecd433677"
	b16SacredPeaksOracle          = "fb69bc57-f05a-41c2-9b7b-9a9761ef0cd3"
	b16VoltaicKeyOracle           = "09aeea91-b1dc-443f-a509-4758f052c0a7"
	b16RelicOfSauronOracle        = "b3a81bb1-cbd5-41a0-8dd4-faea06593f84"
	b16TombOfTheSpiritDragonOrcl  = "22f6391e-2634-440f-af1b-9581d1bff818"
	b16SunlitMarshOracle          = "a5e5a259-5fa7-4b01-93cb-a2b4aaf80927"
	b16DaxosOracle                = "9d3c7c96-056f-408e-a834-fa45a430d3d4"
	b16CircuitousRouteOracle      = "30afacd9-4680-4aac-8c22-584f9418822d"
	b16RiptideLaboratoryOracle    = "444d50dd-a44a-42db-bbf6-d0978e3bd6a3"
	b16UrabraskTheHiddenOracle    = "5b2ffb53-86b7-4665-a5c7-b85b035b6c81"
	b16TevalsJudgmentOracle       = "71fc2393-f55c-4b06-897c-fb7d4199b5f5"
	b16AureliaTheLawAboveOracle   = "2a800427-ff8c-4b3c-baee-85211b70656d"
	b16IzzetGuildgateOracle       = "bf75a3d1-f184-4b48-a913-21caee1db084"
	b16BartolomeDelPresidioOracle = "f47e4c56-0a0b-422d-bf3d-7a20ee289f15"
	b16VengefulBloodwitchOracle   = "4cdbc466-42fc-471f-beab-397caec18101"
	b16PsychicCorrosionOracle     = "328b42f1-d679-4f9c-80e3-38fe3b965d10"
	b16ThoughtScourOracle         = "83101ba8-a569-4827-8c53-9ca0dfcd59a7"
	b16GoblinMatronOracle         = "145737a7-c597-4dec-b752-207c2d0501e3"
	b16CollectorsVaultOracle      = "f460ff80-8175-43f7-9810-540cf3817085"
	b16DemonicCounselOracle       = "712e3479-722c-40a1-9b61-d5bdde93042b"
	b16LeadenMyrOracle            = "f62cabf0-df0d-4c4f-a93a-9340967d1775"
	b16ZacamaOracle               = "23270e4a-a222-4ffb-a946-d0e20d665187"
	b16DimirGuildgateOracle       = "52d14717-0cbc-4d7e-b546-54ea91580338"
	b16BraveTheSandsOracle        = "4e89bd75-f59d-4f08-be51-5660fbbba3c2"
	b16MurmuringMysticOracle      = "dcd4da46-5438-4454-8b1b-43ca51bda1f9"
	b16HorizonExplorerOracle      = "e8d20361-d9b7-4f9c-8ec5-3ac7c460dcb2"
	b16LyraDawnbringerOracle      = "592c91fc-6430-4c76-9460-65f047350f67"
	b16UnstoppablePlanOracle      = "3f6f4c98-ed7e-4fb1-8bfe-4210a39f77f2"
	b16DisallowOracle             = "88b51e15-6630-4e14-a6b8-db0aa12e34ef"
	b16WoodedRidgelineOracle      = "c2ca3e20-23ca-4d2a-88a1-5e98ff884abb"
	b16JuriOracle                 = "3364bad6-6d0a-4141-b410-86e3d9e1916e"
	b16TrueConvictionOracle       = "fc299c1c-50f3-492a-b6b8-a3664bb72ab7"
	b16BennieBracksOracle         = "17bd7ef7-8b4b-4a2d-a667-c751e10e2a47"

	b16SimicGuildgateOracle = "e8705df9-6439-4930-91b6-229f818559af"
)

// b16Creature seeds a non-catalog creature with a type line and
// colours, able to attack and tap.
func b16Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b16Activate activates ability idx on a permanent and settles it.
func b16Activate(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, card, idx, params); err != nil {
		t.Fatalf("ActivateCatalogAbility %d: %v", idx, err)
	}
	passPriorityAroundTable(t, g)
}

// b16Tapped reads a battlefield card's tapped state.
func b16Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b16CountNamed counts battlefield permanents with the given name.
func b16CountNamed(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// b16Tap taps a battlefield permanent directly.
func b16Tap(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
}

// b16TargetPlayer / b16TargetCard build target refs.
func b16TargetPlayer(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}

func b16TargetCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// b16PickPlayer answers a pick_target prompt with a player.
func b16PickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Five are rows
// in cycle tables (three Dominaria United duals, two Guildgates) and
// one is the first row of the Myr table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch16CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b16DeathBaronOracle:           "Death Baron",
		b16FlameOfAnorOracle:          "Flame of Anor",
		b16SacredPeaksOracle:          "Sacred Peaks",
		b16VoltaicKeyOracle:           "Voltaic Key",
		b16RelicOfSauronOracle:        "Relic of Sauron",
		b16TombOfTheSpiritDragonOrcl:  "Tomb of the Spirit Dragon",
		b16SunlitMarshOracle:          "Sunlit Marsh",
		b16DaxosOracle:                "Daxos, Blessed by the Sun",
		b16CircuitousRouteOracle:      "Circuitous Route",
		b16RiptideLaboratoryOracle:    "Riptide Laboratory",
		b16UrabraskTheHiddenOracle:    "Urabrask the Hidden",
		b16TevalsJudgmentOracle:       "Teval's Judgment",
		b16AureliaTheLawAboveOracle:   "Aurelia, the Law Above",
		b16IzzetGuildgateOracle:       "Izzet Guildgate",
		b16BartolomeDelPresidioOracle: "Bartolomé del Presidio",
		b16VengefulBloodwitchOracle:   "Vengeful Bloodwitch",
		b16PsychicCorrosionOracle:     "Psychic Corrosion",
		b16ThoughtScourOracle:         "Thought Scour",
		b16GoblinMatronOracle:         "Goblin Matron",
		b16CollectorsVaultOracle:      "Collector's Vault",
		b16DemonicCounselOracle:       "Demonic Counsel",
		b16LeadenMyrOracle:            "Leaden Myr",
		b16ZacamaOracle:               "Zacama, Primal Calamity",
		b16DimirGuildgateOracle:       "Dimir Guildgate",
		b16BraveTheSandsOracle:        "Brave the Sands",
		b16MurmuringMysticOracle:      "Murmuring Mystic",
		b16HorizonExplorerOracle:      "Horizon Explorer",
		b16LyraDawnbringerOracle:      "Lyra Dawnbringer",
		b16UnstoppablePlanOracle:      "Unstoppable Plan",
		b16DisallowOracle:             "Disallow",
		b16WoodedRidgelineOracle:      "Wooded Ridgeline",
		b16JuriOracle:                 "Juri, Master of the Revue",
		b16TrueConvictionOracle:       "True Conviction",
		b16BennieBracksOracle:         "Bennie Bracks, Zoologist",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- the lands -----------------------------------------------------

// The five cycle rows, each producing its own printed pair — the
// loop-leak canary the Temple cycle established.
func TestB16LandRowsProduceTheirPrintedColours(t *testing.T) {
	for _, row := range []struct{ oracle, name, produced string }{
		{b16SacredPeaksOracle, "Sacred Peaks", "{R|W}"},
		{b16SunlitMarshOracle, "Sunlit Marsh", "{W|B}"},
		{b16WoodedRidgelineOracle, "Wooded Ridgeline", "{R|G}"},
		{b16IzzetGuildgateOracle, "Izzet Guildgate", "{U|R}"},
		{b16DimirGuildgateOracle, "Dimir Guildgate", "{U|B}"},
	} {
		spec, ok := Lookup(row.oracle)
		if !ok {
			t.Errorf("%s not registered", row.name)
			continue
		}
		if len(spec.Replacements) != 1 {
			t.Errorf("%s: %d replacements, want 1 (enters tapped)", row.name, len(spec.Replacements))
		}
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != row.produced {
			t.Errorf("%s: mana abilities %+v, want one producing %s", row.name, spec.ManaAbilities, row.produced)
		}
	}
}

func TestB16SacredPeaksAndDimirGuildgateEnterTapped(t *testing.T) {
	g := newCatalogGame(t)
	peaks := playLandFromHand(t, g, "Sacred Peaks", b16SacredPeaksOracle)
	top100AssertEnteredTapped(t, g, peaks, "Sacred Peaks")
	if tapEventsFor(g, peaks) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	advanceToMainOf(t, g, 1)
	gate := playLandFromHand(t, g, "Dimir Guildgate", b16DimirGuildgateOracle)
	top100AssertEnteredTapped(t, g, gate, "Dimir Guildgate")
}

// --- the lords and team grants -------------------------------------

func TestB16DeathBaronBuffsSkeletonsAndOtherZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baron := b12Push(g, me.ID, "Death Baron", "Creature — Zombie Wizard", b16DeathBaronOracle, 2, 2)
	zombie := b16Creature(g, me.ID, "Walking Corpse", "Creature — Zombie", 2, 2, "B")
	skeleton := b16Creature(g, me.ID, "Drudge Skeletons", "Creature — Skeleton", 1, 1, "B")
	human := b16Creature(g, me.ID, "Bear Cub", "Creature — Human", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")

	if effectivePower(t, g, zombie) != 3 || !hasEffectiveKeyword(t, g, zombie, "deathtouch") {
		t.Error("another Zombie you control gets +1/+1 and deathtouch")
	}
	if effectivePower(t, g, skeleton) != 2 || !hasEffectiveKeyword(t, g, skeleton, "deathtouch") {
		t.Error("a Skeleton you control gets +1/+1 and deathtouch")
	}
	if effectivePower(t, g, baron) != 2 || hasEffectiveKeyword(t, g, baron, "deathtouch") {
		t.Error("the Baron is a Zombie, not an OTHER Zombie — no self-buff")
	}
	if effectivePower(t, g, human) != 2 || hasEffectiveKeyword(t, g, human, "deathtouch") {
		t.Error("a Human is untouched")
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's Zombie is untouched")
	}
}

func TestB16LyraDawnbringerBuffsOtherAngels(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lyra := b12Push(g, me.ID, "Lyra Dawnbringer", "Legendary Creature — Angel", b16LyraDawnbringerOracle, 5, 5)
	angel := b16Creature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4, "W")
	knight := b16Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2, "W")

	if effectivePower(t, g, angel) != 5 || effectiveToughness(t, g, angel) != 5 || !hasEffectiveKeyword(t, g, angel, "lifelink") {
		t.Error("another Angel gets +1/+1 and lifelink")
	}
	if effectivePower(t, g, lyra) != 5 || !hasEffectiveKeyword(t, g, lyra, "lifelink") || !hasEffectiveKeyword(t, g, lyra, "first strike") {
		t.Error("Lyra keeps her printed 5/5, flying, first strike and lifelink and does not buff herself")
	}
	if effectivePower(t, g, knight) != 2 || hasEffectiveKeyword(t, g, knight, "lifelink") {
		t.Error("a non-Angel is untouched")
	}
}

func TestB16TrueConvictionGrantsDoubleStrikeAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "True Conviction", "Enchantment", b16TrueConvictionOracle, nil)
	passPriorityAroundTable(t, g)

	if !hasEffectiveKeyword(t, g, mine, "double strike") || !hasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("your creatures have double strike and lifelink")
	}
	if hasEffectiveKeyword(t, g, theirs, "double strike") || hasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature gets nothing")
	}
	// Double strike plus lifelink: a 2/2 swinging unblocked deals 4
	// across the two substeps and gains 4.
	before, life := opp.Life, me.Life
	attackWith(t, g, opp.ID, mine)
	if opp.Life != before-4 {
		t.Errorf("opponent took %d, want 4 (double strike)", before-opp.Life)
	}
	if me.Life != life+4 {
		t.Errorf("controller gained %d, want 4 (lifelink)", me.Life-life)
	}
}

func TestB16BraveTheSandsGrantsVigilanceAndDeclaresTheBlockGap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Brave the Sands", "Enchantment", b16BraveTheSandsOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Fatal("your creatures have vigilance")
	}
	declareAttack(t, g, opp.ID, bear)
	if b16Tapped(t, g, bear) {
		t.Error("a vigilant attacker stays untapped")
	}
	spec, _ := Lookup(b16BraveTheSandsOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Error("the extra-block gap must be declared")
	}
}

func TestB16UrabraskGrantsHasteAndTapsOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Urabrask the Hidden", "Legendary Creature — Phyrexian Praetor", b16UrabraskTheHiddenOracle, 4, 4)
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("creatures you control have haste")
	}
	// Your own creature enters untapped; an opponent's creature spell
	// resolves tapped; an opponent's noncreature permanent is left
	// alone.
	own := top100CastCreature(t, g, me, "My Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, own) {
		t.Error("your own creature enters untapped")
	}
	advanceToMainOf(t, g, 1)
	theirs := top100CastCreature(t, g, opp, "Their Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	rock := handCardFull(opp, "Mana Rock", "Artifact", "", "", nil)
	if err := g.CastSpell(opp.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, rock) {
		t.Error("an opponent's artifact is not a creature")
	}
	// Declared: a token skips the entry pipeline and enters untapped.
	if spec, _ := Lookup(b16UrabraskTheHiddenOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the token gap must be declared")
	}
}

// --- the artifacts and utility lands -------------------------------

func TestB16RelicOfSauronMakesTwoPicksAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := b12Push(g, me.ID, "Relic of Sauron", "Artifact", b16RelicOfSauronOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if len(c.ColorOptions) != 3 {
				t.Errorf("each slot offers U, B and R, got %v", c.ColorOptions)
			}
			if err := g.ResolveManaChoice(c.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}

	// The loot: untap it, pay {3}, draw two, owe one discard.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	me.ManaPool.EmptyPool()
	b06AddMana(me, "C", "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("drew %d, want 2", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if !b16Tapped(t, g, relic) {
		t.Error("the Relic taps for the loot")
	}
}

func TestB16VoltaicKeyUntapsATargetArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	key := b12Push(g, me.ID, "Voltaic Key", "Artifact", b16VoltaicKeyOracle, 0, 0)
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b16Tap(g, rock)
	b16Tap(g, bear)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, key, 0, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if b16Tapped(t, g, rock) {
		t.Error("the target artifact untaps")
	}
	if !b16Tapped(t, g, key) {
		t.Error("the Key taps as its cost")
	}
	if !b16Tapped(t, g, bear) {
		t.Error("nothing else untaps")
	}
}

func TestB16TombOfTheSpiritDragonGainsLifePerColorlessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tomb := b12Push(g, me.ID, "Tomb of the Spirit Dragon", "Land", b16TombOfTheSpiritDragonOrcl, 0, 0)
	b16Creature(g, me.ID, "Eldrazi", "Creature — Eldrazi", 5, 5)
	b16Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	b16Creature(g, opp.ID, "Their Eldrazi", "Creature — Eldrazi", 5, 5)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	before := me.Life
	b16Activate(t, g, me.ID, tomb, 0, game.ActivateAbilityParams{})
	if me.Life != before+2 {
		t.Errorf("gained %d, want 2 (two colorless creatures you control)", me.Life-before)
	}
	if spec, _ := Lookup(b16TombOfTheSpiritDragonOrcl); len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Error("the Tomb taps for {C}")
	}
}

func TestB16RiptideLaboratoryBouncesAWizardYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lab := b12Push(g, me.ID, "Riptide Laboratory", "Land", b16RiptideLaboratoryOracle, 0, 0)
	wizard := b16Creature(g, me.ID, "Snapcaster Mage", "Creature — Human Wizard", 2, 1, "U")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Wizard", "Creature — Wizard", 2, 2, "U")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, lab, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a non-Wizard is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, lab, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's Wizard is not a legal target")
	}
	b06AddMana(me, "C", "U")
	b16Activate(t, g, me.ID, lab, 0, game.ActivateAbilityParams{Targets: b16TargetCard(wizard)})
	if !me.Hand.Contains(wizard) {
		t.Error("the Wizard returns to its owner's hand")
	}
}

func TestB16CollectorsVaultLootsAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vault := b12Push(g, me.ID, "Collector's Vault", "Artifact", b16CollectorsVaultOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if b16CountNamed(g, "Treasure") != 1 {
		t.Error("a Treasure is created")
	}
}

func TestB16LeadenMyrTapsForBlackOnceItCanTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	myr := pushCatalogPermanent(g, me.ID, "Leaden Myr", "Artifact Creature — Myr", b16LeadenMyrOracle, true)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Myr can't tap for mana")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == myr {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	if err := g.ActivateManaAbility(me.ID, myr, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

// --- the spells ----------------------------------------------------

func TestB16DisallowCountersASpellAndDeclaresTheAbilityGap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Disallow", "Instant", b16DisallowOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != 40 {
		t.Errorf("life = %d, the countered Bolt must not resolve", me.Life)
	}
	spec, _ := Lookup(b16DisallowOracle)
	if spec.Completeness != CompletenessCaveats {
		t.Error("the spells-only gap must be declared")
	}
}

func TestB16ThoughtScourMillsTheTargetThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand, grave, lib := me.Hand.Size(), opp.Graveyard.Size(), opp.Library.Size()
	castCatalogSpell(t, g, "Thought Scour", "Instant", b16ThoughtScourOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != grave+2 || opp.Library.Size() != lib-2 {
		t.Error("the target player mills two")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the caster draws one", hand, me.Hand.Size())
	}
}

func TestB16FlameOfAnorIsChooseOneAndResolvesEachMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4, "G")
	advanceToMain(t, g)

	cast := func(mode int, targets []game.TargetRef) error {
		id := handCardFull(me, "Flame of Anor", "Instant", "", b16FlameOfAnorOracle, []string{"U", "R"})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets})
	}
	if err := cast(2, b16TargetCard(bear)); err != nil {
		t.Fatalf("cast mode 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("5 damage kills a 4/4")
	}
	if err := cast(1, b16TargetCard(rock)); err != nil {
		t.Fatalf("cast mode 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	hand := opp.Hand.Size()
	if err := cast(0, b16TargetPlayer(opp.ID)); err != nil {
		t.Fatalf("cast mode 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the target player drew %d, want 2", opp.Hand.Size()-hand)
	}
	spec, _ := Lookup(b16FlameOfAnorOracle)
	if spec.Modes.Max != 1 || spec.Completeness != CompletenessCaveats {
		t.Error("the Wizard bonus gap must be declared, and the spell stays choose-one")
	}
}

func TestB16CircuitousRouteFetchesBasicsAndGatesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Simic Guildgate", TypeLine: "Land — Gate", OracleID: b16SimicGuildgateOracle},
		searchTestLand("Reliquary Tower", "Land"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Circuitous Route", "Sorcery", b16CircuitousRouteOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three matches for two slots: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Reliquary Tower") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic non-Gate land, and a creature, are not offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest", "Simic Guildgate")
	for _, name := range []string{"Forest", "Simic Guildgate"} {
		id := findBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("%s did not reach the battlefield", name)
		}
		if !b16Tapped(t, g, id) {
			t.Errorf("%s enters tapped", name)
		}
	}
}

func TestB16DemonicCounselTutorsADemonOrAnythingWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Griselbrand", TypeLine: "Legendary Creature — Demon"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	castCatalogSpell(t, g, "Demonic Counsel", "Sorcery", b16DemonicCounselOracle, nil)
	passPriorityAroundTable(t, g)
	if !b02bHandHasNamed(me, "Griselbrand") {
		t.Fatal("one Demon in the library: it goes straight to hand")
	}

	// Delirium: four card types in the graveyard opens the search to
	// any card.
	for _, c := range []game.Card{
		{Name: "Dead Bear", TypeLine: "Creature — Bear"},
		{Name: "Dead Rock", TypeLine: "Artifact"},
		{Name: "Dead Bolt", TypeLine: "Instant"},
		{Name: "Dead Land", TypeLine: "Land"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	castCatalogSpell(t, g, "Demonic Counsel", "Sorcery", b16DemonicCounselOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with delirium every card is a candidate, so the searcher chooses")
	}
	if searchOptionNamed(g, c, "Sol Ring") == uuid.Nil {
		t.Error("a non-Demon is offered under delirium")
	}
	answerSearchNamed(t, g, me.ID, "Sol Ring")
	if !b02bHandHasNamed(me, "Sol Ring") {
		t.Error("the chosen card reaches the hand")
	}
}

// --- the creatures -------------------------------------------------

func TestB16DaxosToughnessIsDevotionAndHeGainsOnEntersAndDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	daxos := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Daxos, Blessed by the Sun", TypeLine: "Legendary Enchantment Creature — Demigod",
		OracleID: b16DaxosOracle, ManaCost: "{W}{W}", Power: 2, Toughness: 0, Owner: me.ID, Controller: me.ID,
	})
	if got := effectiveToughness(t, g, daxos); got != 2 {
		t.Errorf("toughness %d, want 2 (his own {W}{W})", got)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment", ManaCost: "{1}{W}{W}",
		Owner: me.ID, Controller: me.ID,
	})
	if got := effectiveToughness(t, g, daxos); got != 4 {
		t.Errorf("toughness %d, want 4 (devotion 4)", got)
	}
	passPriorityAroundTable(t, g)

	before := me.Life
	bear := top100CastCreature(t, g, me, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature entering gains 1, got %d", me.Life-before)
	}
	b15Destroy(t, g, bear)
	if me.Life != before+2 {
		t.Errorf("a creature you control dying gains 1, got %d total", me.Life-before)
	}
}

func TestB16VengefulBloodwitchDrainsATargetOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	witch := b12Push(g, me.ID, "Vengeful Bloodwitch", "Creature — Vampire Warlock", b16VengefulBloodwitchOracle, 1, 1)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	life, oppLife, otherLife := me.Life, opp.Life, other.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	b16PickPlayer(t, g, me.ID, other.ID)
	passPriorityAroundTable(t, g)
	if other.Life != otherLife-1 || opp.Life != oppLife || me.Life != life+1 {
		t.Errorf("another creature dying: chosen opponent −1, you +1; got me %d opp %d other %d", me.Life-life, opp.Life-oppLife, other.Life-otherLife)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(witch) })
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+2 {
		t.Errorf("the Bloodwitch's own death drains too; got me +%d opp −%d", me.Life-life, oppLife-opp.Life)
	}
}

func TestB16GoblinMatronMaySearchForAGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Krenko", TypeLine: "Legendary Creature — Goblin"},
		game.Card{Name: "Goblin Guide", TypeLine: "Creature — Goblin Scout"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Goblin Matron", "Creature — Goblin", b16GoblinMatronOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ETB search is optional, so it always asks")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Goblin is not offered")
	}
	answerSearchNamed(t, g, me.ID, "Goblin Guide")
	if !b02bHandHasNamed(me, "Goblin Guide") {
		t.Error("the Goblin reaches the hand")
	}
}

func TestB16MurmuringMysticMakesABirdIllusionPerInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Murmuring Mystic", "Creature — Human Wizard", b16MurmuringMysticOracle, 1, 5)
	b15CastInstant(t, g, "Opt", "U")
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Bird Illusion") != 1 {
		t.Fatal("an instant makes one Bird Illusion")
	}
	bird := findBattlefieldByName(g, "Bird Illusion")
	if !hasEffectiveKeyword(t, g, bird, "flying") || effectivePower(t, g, bird) != 1 {
		t.Error("the token is a 1/1 with flying")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Bird Illusion") != 1 {
		t.Error("a creature spell makes nothing")
	}
}

func TestB16JuriGrowsOnSacrificeAndDealsItsPowerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	juri := b12Push(g, me.ID, "Juri, Master of the Revue", "Legendary Creature — Human Shaman", b16JuriOracle, 1, 1)
	fodder := b16Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1, "R")
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, juri, "+1/+1"); got != 2 {
		t.Fatalf("%d counters after two sacrifices, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if got := counterCount(g, juri, "+1/+1"); got != 2 {
		t.Fatalf("an opponent's creature being destroyed is not your sacrifice; %d counters", got)
	}

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(juri) })
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("Juri's death deals its last-known power (3) to the target, got %d", before-opp.Life)
	}
}

func TestB16BartolomeEatsAnotherCreatureOrArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bart := b12Push(g, me.ID, "Bartolomé del Presidio", "Legendary Creature — Vampire Knight", b16BartolomeDelPresidioOracle, 2, 1)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bart}}); err == nil {
		t.Fatal("Bartolomé can't sacrifice himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}}); err == nil {
		t.Fatal("an enchantment is neither a creature nor an artifact")
	}
	b16Activate(t, g, me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	b16Activate(t, g, me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(rock) {
		t.Error("the sacrificed permanents are gone")
	}
	if got := counterCount(g, bart, "+1/+1"); got != 2 {
		t.Errorf("%d counters after two sacrifices, want 2", got)
	}
}

func TestB16ZacamaUntapsLandsOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	b16Tap(g, land)
	b16Tap(g, rock)
	castCatalogSpell(t, g, "Zacama, Primal Calamity", "Legendary Creature — Elder Dinosaur", b16ZacamaOracle, nil)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, land) {
		t.Error("a cast Zacama untaps your lands")
	}
	if !b16Tapped(t, g, rock) {
		t.Error("only lands untap")
	}

	// Reanimated: the ETB fires, the intervening if fails, nothing
	// untaps.
	zacama := findBattlefieldByName(g, "Zacama, Primal Calamity")
	b15Destroy(t, g, zacama)
	b16Tap(g, land)
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(zacama, game.ZoneBattlefield, me.ID)
	})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, land) {
		t.Error("a reanimated Zacama was not cast — no untap")
	}
}

func TestB16ZacamaActivatedAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zacama := b12Push(g, me.ID, "Zacama, Primal Calamity", "Legendary Creature — Elder Dinosaur", b16ZacamaOracle, 9, 9)
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3, "G")
	rock := b12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "R")
	b16Activate(t, g, me.ID, zacama, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("{2}{R}: 3 damage kills a 3/3")
	}
	b06AddMana(me, "C", "C", "G")
	b16Activate(t, g, me.ID, zacama, 1, game.ActivateAbilityParams{Targets: b16TargetCard(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("{2}{G}: the artifact is destroyed")
	}
	before := me.Life
	b06AddMana(me, "C", "C", "W")
	b16Activate(t, g, me.ID, zacama, 2, game.ActivateAbilityParams{})
	if me.Life != before+3 {
		t.Errorf("{2}{W}: gained %d, want 3", me.Life-before)
	}
	if !hasEffectiveKeyword(t, g, zacama, "trample") || !hasEffectiveKeyword(t, g, zacama, "reach") || !hasEffectiveKeyword(t, g, zacama, "vigilance") {
		t.Error("reach, vigilance, trample")
	}
}

func TestB16BennieBracksDrawsAtAnyEndStepAfterATokenAndConvokes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	soldiers := pushTapCostSoldiers(g, me.ID, 2)
	id, err := castWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		b16BennieBracksOracle, game.CastSpellParams{TapIDs: soldiers})
	if err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	for _, s := range soldiers {
		if !tapCostTapped(g, s) {
			t.Error("convoke taps the named creatures")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Bennie did not resolve")
	}

	// No token this turn: the end step draws nothing.
	hand := me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no token created this turn — no draw")
	}

	// An opponent's turn: a token created by Bennie's controller
	// during it still draws at that turn's end step ("each end step").
	advanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 colorless Zombie"), 1) })
	hand = me.Hand.Size()
	batch01AdvanceToStepOf(t, g, 1, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d at an opponent's end step after making a token, want 1", me.Hand.Size()-hand)
	}
}

func TestB16AureliaFiresOncePerDeclarationAtThreeAndFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aurelia := b12Push(g, me.ID, "Aurelia, the Law Above", "Legendary Creature — Angel", b16AureliaTheLawAboveOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G"))
	}
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, bears[0], bears[1])
	if n := triggersOnStackFrom(g, aurelia); n != 0 {
		t.Fatalf("two attackers: %d triggers, want 0", n)
	}
	if err := g.DeclareAttacker(bears[2], opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, aurelia); n != 1 {
		t.Fatalf("three attackers: %d triggers, want 1 (draw)", n)
	}
	if err := g.DeclareAttacker(bears[3], opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, aurelia); n != 1 {
		t.Fatalf("four attackers: %d triggers, want still 1", n)
	}
	if err := g.DeclareAttacker(bears[4], opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, aurelia); n != 2 {
		t.Fatalf("five attackers: %d triggers, want 2 (draw + drain)", n)
	}
	lives := b15Lives(g)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1", me.Hand.Size()-hand)
	}
	for i, p := range g.Seats[1:] {
		if p.Life != lives[i]-3 {
			t.Errorf("opponent %d took %d, want 3", i, lives[i]-p.Life)
		}
	}
	if me.Life != life+3 {
		t.Errorf("gained %d, want 3", me.Life-life)
	}
}

func TestB16AureliaDrawsForAnOpponentsAttackToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Aurelia, the Law Above", "Legendary Creature — Angel", b16AureliaTheLawAboveOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G"))
	}
	advanceToMainOf(t, g, 1)
	hand := me.Hand.Size()
	attackWith(t, g, me.ID, bears...)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent attacking with three draws Aurelia's controller a card; drew %d", me.Hand.Size()-hand)
	}
}

func TestB16HorizonExplorerLandsEnterUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Horizon Explorer", "Creature — Insect Scout", b16HorizonExplorerOracle, 2, 4)
	gate := playLandFromHand(t, g, "Simic Guildgate", b16SimicGuildgateOracle)
	top100AssertEnteredUntapped(t, g, gate, "Simic Guildgate")
	if len(g.PendingChoices) != 0 {
		t.Fatal("no ordering prompt: the Explorer applies after the land's own effect")
	}

	// A basic fetched "onto the battlefield tapped" enters untapped;
	// a fetched Guildgate does not (declared).
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Dimir Guildgate", TypeLine: "Land — Gate", OracleID: b16DimirGuildgateOracle},
	)
	castCatalogSpell(t, g, "Circuitous Route", "Sorcery", b16CircuitousRouteOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Forest", "Dimir Guildgate")
	if b16Tapped(t, g, findBattlefieldByName(g, "Forest")) {
		t.Error("a fetched basic enters untapped under the Explorer")
	}
	if !b16Tapped(t, g, findBattlefieldByName(g, "Dimir Guildgate")) {
		t.Error("a fetched Guildgate enters tapped — the declared retreat")
	}
	if len(g.PendingChoices) != 0 || me.Library.Size() != 1 {
		t.Error("nothing is stranded and nothing prompts")
	}
}

func TestB16HorizonExplorerMakesALanderPerAttackThatFetches(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	explorer := b12Push(g, me.ID, "Horizon Explorer", "Creature — Insect Scout", b16HorizonExplorerOracle, 2, 4)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
	declareAttack(t, g, opp.ID, explorer, bear)
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Lander") != 1 {
		t.Fatalf("two attackers make one Lander, got %d", b16CountNamed(g, "Lander"))
	}
	lander := findBattlefieldByName(g, "Lander")
	// The Explorer leaves first, so the Lander's own "tapped" clause
	// is what is being read.
	b15Destroy(t, g, explorer)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, lander, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(lander) {
		t.Error("the Lander is sacrificed as a cost")
	}
	forest := findBattlefieldByName(g, "Forest")
	if forest == uuid.Nil || !b16Tapped(t, g, forest) {
		t.Error("the Lander fetches a basic land onto the battlefield tapped")
	}
}

func TestB16UnstoppablePlanUntapsNonlandsAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Unstoppable Plan", "Enchantment", b16UnstoppablePlanOracle, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b16Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	for _, id := range []uuid.UUID{bear, rock, land, theirs} {
		b16Tap(g, id)
	}
	batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, bear) || b16Tapped(t, g, rock) {
		t.Error("your nonland permanents untap")
	}
	if !b16Tapped(t, g, land) || !b16Tapped(t, g, theirs) {
		t.Error("lands and opponents' permanents stay tapped")
	}
}

func TestB16PsychicCorrosionMillsEachOpponentTwoPerDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Psychic Corrosion", "Enchantment", b16PsychicCorrosionOracle, 0, 0)
	advanceToMain(t, g)
	graves := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		graves = append(graves, p.Graveyard.Size())
	}
	mine := me.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != graves[i]+4 {
			t.Errorf("opponent %d milled %d, want 4 (two draws)", i, p.Graveyard.Size()-graves[i])
		}
	}
	if me.Graveyard.Size() != mine {
		t.Error("the controller mills nothing")
	}
}

func TestB16TevalsJudgmentTakesItsModesInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Teval's Judgment", "Enchantment", b16TevalsJudgmentOracle, 0, 0)
	advanceToMain(t, g)
	graveyard := func(p *game.Player, n int) []uuid.UUID {
		var ids []uuid.UUID
		for i := 0; i < n; i++ {
			ids = append(ids, pushGraveyardCardForTest(p, "Dead"))
		}
		return ids
	}

	// An opponent's graveyard leaving is not yours.
	theirs := graveyard(opp, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs[0]) })
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if me.Hand.Size() != hand {
		t.Fatal("an opponent's graveyard is not yours")
	}

	// Two cards at once: one trigger, first mode — draw.
	first := graveyard(me, 2)
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(first) })
	if n := len(g.PendingTriggers) + len(g.StackMeta); n != 1 {
		t.Fatalf("one or more cards leaving is one trigger, got %d", n)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("first batch draws: %d", me.Hand.Size()-hand)
	}

	// Second: a Treasure. Third: a Zombie Druid. Fourth: nothing.
	for i, want := range []struct {
		name string
		n    int
	}{{"Treasure", 1}, {"Zombie Druid", 1}, {"Zombie Druid", 1}} {
		card := graveyard(me, 1)
		g.WithWriteLock(func() { _ = g.BounceToHandForEffect(card[0]) })
		passPriorityAroundTable(t, g)
		if got := b16CountNamed(g, want.name); got != want.n {
			t.Errorf("batch %d: %d %s on the battlefield, want %d", i+2, got, want.name, want.n)
		}
	}
	if me.Hand.Size() != hand+1+3 {
		t.Errorf("hand grew by %d: one draw plus the three bounced cards, no more", me.Hand.Size()-hand)
	}

	// A new turn resets the tally: the first batch draws again.
	advanceToMainOf(t, g, 1)
	hand = me.Hand.Size()
	card := graveyard(me, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(card[0]) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Error("next turn, the first mode is available again")
	}

	// A flashback cast leaves the graveyard too: on my next turn,
	// flashing back Faithless Looting is the first batch — a draw,
	// on top of the Looting's own two.
	advanceToMainOf(t, g, 0)
	looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	hand = me.Hand.Size()
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	if triggerOnStack(g, findBattlefieldByName(g, "Teval's Judgment")) == nil {
		t.Fatal("a flashback cast leaves the graveyard: the trigger goes on the stack above the spell")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("drew %d, want 3 (Teval's one plus the Looting's two)", me.Hand.Size()-hand)
	}
}
