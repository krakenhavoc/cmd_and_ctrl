package testsav

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Five are rows
// in cycle tables (three Dominaria United duals, two Guildgates) and
// one is the first row of the Myr table, so a transposed row is
// invisible until someone plays that exact card.
func TestBatch16CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B16DeathBaronOracle:           "Death Baron",
		effectstest.B16FlameOfAnorOracle:          "Flame of Anor",
		effectstest.B16SacredPeaksOracle:          "Sacred Peaks",
		effectstest.B16VoltaicKeyOracle:           "Voltaic Key",
		effectstest.B16RelicOfSauronOracle:        "Relic of Sauron",
		effectstest.B16TombOfTheSpiritDragonOrcl:  "Tomb of the Spirit Dragon",
		effectstest.B16SunlitMarshOracle:          "Sunlit Marsh",
		effectstest.B16DaxosOracle:                "Daxos, Blessed by the Sun",
		effectstest.B16CircuitousRouteOracle:      "Circuitous Route",
		effectstest.B16RiptideLaboratoryOracle:    "Riptide Laboratory",
		effectstest.B16UrabraskTheHiddenOracle:    "Urabrask the Hidden",
		effectstest.B16TevalsJudgmentOracle:       "Teval's Judgment",
		effectstest.B16AureliaTheLawAboveOracle:   "Aurelia, the Law Above",
		effectstest.B16IzzetGuildgateOracle:       "Izzet Guildgate",
		effectstest.B16BartolomeDelPresidioOracle: "Bartolomé del Presidio",
		effectstest.B16VengefulBloodwitchOracle:   "Vengeful Bloodwitch",
		effectstest.B16PsychicCorrosionOracle:     "Psychic Corrosion",
		effectstest.B16ThoughtScourOracle:         "Thought Scour",
		effectstest.B16GoblinMatronOracle:         "Goblin Matron",
		effectstest.B16CollectorsVaultOracle:      "Collector's Vault",
		effectstest.B16DemonicCounselOracle:       "Demonic Counsel",
		effectstest.B16LeadenMyrOracle:            "Leaden Myr",
		effectstest.B16ZacamaOracle:               "Zacama, Primal Calamity",
		effectstest.B16DimirGuildgateOracle:       "Dimir Guildgate",
		effectstest.B16BraveTheSandsOracle:        "Brave the Sands",
		effectstest.B16MurmuringMysticOracle:      "Murmuring Mystic",
		effectstest.B16HorizonExplorerOracle:      "Horizon Explorer",
		effectstest.B16LyraDawnbringerOracle:      "Lyra Dawnbringer",
		effectstest.B16UnstoppablePlanOracle:      "Unstoppable Plan",
		effectstest.B16DisallowOracle:             "Disallow",
		effectstest.B16WoodedRidgelineOracle:      "Wooded Ridgeline",
		effectstest.B16JuriOracle:                 "Juri, Master of the Revue",
		effectstest.B16TrueConvictionOracle:       "True Conviction",
		effectstest.B16BennieBracksOracle:         "Bennie Bracks, Zoologist",
	}
	if len(want) != 34 {
		t.Fatalf("the batch registers 34 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
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
		{effectstest.B16SacredPeaksOracle, "Sacred Peaks", "{R|W}"},
		{effectstest.B16SunlitMarshOracle, "Sunlit Marsh", "{W|B}"},
		{effectstest.B16WoodedRidgelineOracle, "Wooded Ridgeline", "{R|G}"},
		{effectstest.B16IzzetGuildgateOracle, "Izzet Guildgate", "{U|R}"},
		{effectstest.B16DimirGuildgateOracle, "Dimir Guildgate", "{U|B}"},
	} {
		spec, ok := effects.Lookup(row.oracle)
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
	g := effectstest.NewCatalogGame(t)
	peaks := effectstest.PlayLandFromHand(t, g, "Sacred Peaks", effectstest.B16SacredPeaksOracle)
	effectstest.Top100AssertEnteredTapped(t, g, peaks, "Sacred Peaks")
	if effectstest.TapEventsFor(g, peaks) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	gate := effectstest.PlayLandFromHand(t, g, "Dimir Guildgate", effectstest.B16DimirGuildgateOracle)
	effectstest.Top100AssertEnteredTapped(t, g, gate, "Dimir Guildgate")
}

// --- the lords and team grants -------------------------------------

func TestB16DeathBaronBuffsSkeletonsAndOtherZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baron := effectstest.B12Push(g, me.ID, "Death Baron", "Creature — Zombie Wizard", effectstest.B16DeathBaronOracle, 2, 2)
	zombie := effectstest.B16Creature(g, me.ID, "Walking Corpse", "Creature — Zombie", 2, 2, "B")
	skeleton := effectstest.B16Creature(g, me.ID, "Drudge Skeletons", "Creature — Skeleton", 1, 1, "B")
	human := effectstest.B16Creature(g, me.ID, "Bear Cub", "Creature — Human", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2, "B")

	if effectstest.EffectivePower(t, g, zombie) != 3 || !effectstest.HasEffectiveKeyword(t, g, zombie, "deathtouch") {
		t.Error("another Zombie you control gets +1/+1 and deathtouch")
	}
	if effectstest.EffectivePower(t, g, skeleton) != 2 || !effectstest.HasEffectiveKeyword(t, g, skeleton, "deathtouch") {
		t.Error("a Skeleton you control gets +1/+1 and deathtouch")
	}
	if effectstest.EffectivePower(t, g, baron) != 2 || effectstest.HasEffectiveKeyword(t, g, baron, "deathtouch") {
		t.Error("the Baron is a Zombie, not an OTHER Zombie — no self-buff")
	}
	if effectstest.EffectivePower(t, g, human) != 2 || effectstest.HasEffectiveKeyword(t, g, human, "deathtouch") {
		t.Error("a Human is untouched")
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's Zombie is untouched")
	}
}

func TestB16LyraDawnbringerBuffsOtherAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lyra := effectstest.B12Push(g, me.ID, "Lyra Dawnbringer", "Legendary Creature — Angel", effectstest.B16LyraDawnbringerOracle, 5, 5)
	angel := effectstest.B16Creature(g, me.ID, "Serra Angel", "Creature — Angel", 4, 4, "W")
	knight := effectstest.B16Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2, "W")

	if effectstest.EffectivePower(t, g, angel) != 5 || effectstest.EffectiveToughness(t, g, angel) != 5 || !effectstest.HasEffectiveKeyword(t, g, angel, "lifelink") {
		t.Error("another Angel gets +1/+1 and lifelink")
	}
	if effectstest.EffectivePower(t, g, lyra) != 5 || !effectstest.HasEffectiveKeyword(t, g, lyra, "lifelink") || !effectstest.HasEffectiveKeyword(t, g, lyra, "first strike") {
		t.Error("Lyra keeps her printed 5/5, flying, first strike and lifelink and does not buff herself")
	}
	if effectstest.EffectivePower(t, g, knight) != 2 || effectstest.HasEffectiveKeyword(t, g, knight, "lifelink") {
		t.Error("a non-Angel is untouched")
	}
}

func TestB16TrueConvictionGrantsDoubleStrikeAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "True Conviction", "Enchantment", effectstest.B16TrueConvictionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.HasEffectiveKeyword(t, g, mine, "double strike") || !effectstest.HasEffectiveKeyword(t, g, mine, "lifelink") {
		t.Error("your creatures have double strike and lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "double strike") || effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature gets nothing")
	}
	// Double strike plus lifelink: a 2/2 swinging unblocked deals 4
	// across the two substeps and gains 4.
	before, life := opp.Life, me.Life
	effectstest.AttackWith(t, g, opp.ID, mine)
	if opp.Life != before-4 {
		t.Errorf("opponent took %d, want 4 (double strike)", before-opp.Life)
	}
	if me.Life != life+4 {
		t.Errorf("controller gained %d, want 4 (lifelink)", me.Life-life)
	}
}

func TestB16BraveTheSandsGrantsVigilanceAndAnExtraBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Brave the Sands", "Enchantment", effectstest.B16BraveTheSandsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Fatal("your creatures have vigilance")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("a vigilant attacker stays untapped")
	}
	// #1706: the second line is live — each of your creatures can
	// block two attackers, an opponent's still one.
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if c := effectstest.E2Card(t, g, bear); game.BlockCapacity(&c) != 2 {
		t.Errorf("your creature's block capacity = %d, want 2", game.BlockCapacity(&c))
	}
	if c := effectstest.E2Card(t, g, theirs); game.BlockCapacity(&c) != 1 {
		t.Errorf("an opponent's creature's block capacity = %d, want 1", game.BlockCapacity(&c))
	}
	spec, _ := effects.Lookup(effectstest.B16BraveTheSandsOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Brave the Sands is complete since #1706")
	}
}

func TestB16UrabraskGrantsHasteAndTapsOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Urabrask the Hidden", "Legendary Creature — Phyrexian Praetor", effectstest.B16UrabraskTheHiddenOracle, 4, 4)
	mine := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("creatures you control have haste")
	}
	// Your own creature enters untapped; an opponent's creature spell
	// resolves tapped; an opponent's noncreature permanent is left
	// alone.
	own := effectstest.Top100CastCreature(t, g, me, "My Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, own) {
		t.Error("your own creature enters untapped")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	theirs := effectstest.Top100CastCreature(t, g, opp, "Their Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("an opponent's creature enters tapped")
	}
	rock := effectstest.HandCardFull(opp, "Mana Rock", "Artifact", "", "", nil)
	if err := g.CastSpell(opp.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, rock) {
		t.Error("an opponent's artifact is not a creature")
	}
	// #762: a created token runs the same entry pipeline, so an
	// opponent's creature token arrives tapped too.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1) })
	if !effectstest.B16Tapped(t, g, effectstest.FindBattlefieldByName(g, "Goblin")) {
		t.Error("an opponent's creature TOKEN enters tapped")
	}
	if spec, _ := effects.Lookup(effectstest.B16UrabraskTheHiddenOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the token gap is closed — the caveat must be gone")
	}
}

// --- the artifacts and utility lands -------------------------------

func TestB16RelicOfSauronMakesTwoPicksAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.B12Push(g, me.ID, "Relic of Sauron", "Artifact", effectstest.B16RelicOfSauronOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, relic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// #3939: one choice, asked once per pick.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("pool %v, want [R R]", got)
	}

	// The loot: untap it, pay {3}, draw two, owe one discard.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	me.ManaPool.EmptyPool()
	effectstest.B06AddMana(me, "C", "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, relic, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("drew %d, want 2", got-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	if !effectstest.B16Tapped(t, g, relic) {
		t.Error("the Relic taps for the loot")
	}
}

func TestB16VoltaicKeyUntapsATargetArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	key := effectstest.B12Push(g, me.ID, "Voltaic Key", "Artifact", effectstest.B16VoltaicKeyOracle, 0, 0)
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Tap(g, rock)
	effectstest.B16Tap(g, bear)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an artifact")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, key, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	if effectstest.B16Tapped(t, g, rock) {
		t.Error("the target artifact untaps")
	}
	if !effectstest.B16Tapped(t, g, key) {
		t.Error("the Key taps as its cost")
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("nothing else untaps")
	}
}

func TestB16TombOfTheSpiritDragonGainsLifePerColorlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tomb := effectstest.B12Push(g, me.ID, "Tomb of the Spirit Dragon", "Land", effectstest.B16TombOfTheSpiritDragonOrcl, 0, 0)
	effectstest.B16Creature(g, me.ID, "Eldrazi", "Creature — Eldrazi", 5, 5)
	effectstest.B16Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
	effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B16Creature(g, opp.ID, "Their Eldrazi", "Creature — Eldrazi", 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	before := me.Life
	effectstest.B16Activate(t, g, me.ID, tomb, 0, game.ActivateAbilityParams{})
	if me.Life != before+2 {
		t.Errorf("gained %d, want 2 (two colorless creatures you control)", me.Life-before)
	}
	if spec, _ := effects.Lookup(effectstest.B16TombOfTheSpiritDragonOrcl); len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Error("the Tomb taps for {C}")
	}
}

func TestB16RiptideLaboratoryBouncesAWizardYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lab := effectstest.B12Push(g, me.ID, "Riptide Laboratory", "Land", effectstest.B16RiptideLaboratoryOracle, 0, 0)
	wizard := effectstest.B16Creature(g, me.ID, "Snapcaster Mage", "Creature — Human Wizard", 2, 1, "U")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Wizard", "Creature — Wizard", 2, 2, "U")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, lab, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a non-Wizard is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, lab, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's Wizard is not a legal target")
	}
	effectstest.B06AddMana(me, "C", "U")
	effectstest.B16Activate(t, g, me.ID, lab, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(wizard)})
	if !me.Hand.Contains(wizard) {
		t.Error("the Wizard returns to its owner's hand")
	}
}

func TestB16CollectorsVaultLootsAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vault := effectstest.B12Push(g, me.ID, "Collector's Vault", "Artifact", effectstest.B16CollectorsVaultOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	// "Draw a card, then discard a card. Create a Treasure token." (CR 608.2c,
	// #4155): the Treasure waits for the discard's answer.
	if effectstest.B16CountNamed(g, "Treasure") != 0 {
		t.Error("the Treasure was created with the discard still open")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if effectstest.B16CountNamed(g, "Treasure") != 1 {
		t.Error("a Treasure is created")
	}
}

func TestB16LeadenMyrTapsForBlackOnceItCanTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	myr := effectstest.PushCatalogPermanent(g, me.ID, "Leaden Myr", "Artifact Creature — Myr", effectstest.B16LeadenMyrOracle, true)
	effectstest.AdvanceToMain(t, g)
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
}

// --- the spells ----------------------------------------------------

// TestB16DisallowCountersASpell — the spell third of the printed
// three-way clause. The ability thirds are in
// target_ability_cards_test.go, which is where the #1211 seam is
// proved; this test asserted the gap was DECLARED until that landed,
// and now asserts the card is whole.
func TestB16DisallowCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	effectstest.CastCatalogSpell(t, g, "Disallow", "Instant", effectstest.B16DisallowOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not countered")
	}
	if me.Life != 40 {
		t.Errorf("life = %d, the countered Bolt must not resolve", me.Life)
	}
	spec, _ := effects.Lookup(effectstest.B16DisallowOracle)
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("#1211 closed the ability gap: Disallow is %q", spec.Completeness)
	}
}

func TestB16ThoughtScourMillsTheTargetThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand, grave, lib := me.Hand.Size(), opp.Graveyard.Size(), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Thought Scour", "Instant", effectstest.B16ThoughtScourOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != grave+2 || opp.Library.Size() != lib-2 {
		t.Error("the target player mills two")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the caster draws one", hand, me.Hand.Size())
	}
}

func TestB16FlameOfAnorIsChooseOneAndResolvesEachMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 4, 4, "G")
	effectstest.AdvanceToMain(t, g)

	cast := func(mode int, targets []game.TargetRef) error {
		id := effectstest.HandCardFull(me, "Flame of Anor", "Instant", "", effectstest.B16FlameOfAnorOracle, []string{"U", "R"})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets})
	}
	if err := cast(2, effectstest.B16TargetCard(bear)); err != nil {
		t.Fatalf("cast mode 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("5 damage kills a 4/4")
	}
	if err := cast(1, effectstest.B16TargetCard(rock)); err != nil {
		t.Fatalf("cast mode 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is destroyed")
	}
	hand := opp.Hand.Size()
	if err := cast(0, effectstest.B16TargetPlayer(opp.ID)); err != nil {
		t.Fatalf("cast mode 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 {
		t.Errorf("the target player drew %d, want 2", opp.Hand.Size()-hand)
	}
	// Choosing two, with and without a Wizard, is
	// conditional_mode_count_test.go's (#1590).
}

func TestB16CircuitousRouteFetchesBasicsAndGatesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Simic Guildgate", TypeLine: "Land — Gate", OracleID: effectstest.B16SimicGuildgateOracle},
		effectstest.SearchTestLand("Reliquary Tower", "Land"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Circuitous Route", "Sorcery", effectstest.B16CircuitousRouteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("three matches for two slots: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Reliquary Tower") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic non-Gate land, and a creature, are not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Simic Guildgate")
	for _, name := range []string{"Forest", "Simic Guildgate"} {
		id := effectstest.FindBattlefieldByName(g, name)
		if id == uuid.Nil {
			t.Fatalf("%s did not reach the battlefield", name)
		}
		if !effectstest.B16Tapped(t, g, id) {
			t.Errorf("%s enters tapped", name)
		}
	}
}

func TestB16DemonicCounselTutorsADemonOrAnythingWithDelirium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Griselbrand", TypeLine: "Legendary Creature — Demon"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	effectstest.CastCatalogSpell(t, g, "Demonic Counsel", "Sorcery", effectstest.B16DemonicCounselOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !effectstest.B02bHandHasNamed(me, "Griselbrand") {
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
	effectstest.CastCatalogSpell(t, g, "Demonic Counsel", "Sorcery", effectstest.B16DemonicCounselOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with delirium every card is a candidate, so the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Sol Ring") == uuid.Nil {
		t.Error("a non-Demon is offered under delirium")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Sol Ring")
	if !effectstest.B02bHandHasNamed(me, "Sol Ring") {
		t.Error("the chosen card reaches the hand")
	}
}

// --- the creatures -------------------------------------------------

func TestB16DaxosToughnessIsDevotionAndHeGainsOnEntersAndDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	daxos := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Daxos, Blessed by the Sun", TypeLine: "Legendary Enchantment Creature — Demigod",
		OracleID: effectstest.B16DaxosOracle, ManaCost: "{W}{W}", Power: 2, Toughness: 0, Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectiveToughness(t, g, daxos); got != 2 {
		t.Errorf("toughness %d, want 2 (his own {W}{W})", got)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment", ManaCost: "{1}{W}{W}",
		Owner: me.ID, Controller: me.ID,
	})
	if got := effectstest.EffectiveToughness(t, g, daxos); got != 4 {
		t.Errorf("toughness %d, want 4 (devotion 4)", got)
	}
	effectstest.PassPriorityAroundTable(t, g)

	before := me.Life
	bear := effectstest.Top100CastCreature(t, g, me, "Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("another creature entering gains 1, got %d", me.Life-before)
	}
	effectstest.B15Destroy(t, g, bear)
	if me.Life != before+2 {
		t.Errorf("a creature you control dying gains 1, got %d total", me.Life-before)
	}
}

func TestB16VengefulBloodwitchDrainsATargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	witch := effectstest.B12Push(g, me.ID, "Vengeful Bloodwitch", "Creature — Vampire Warlock", effectstest.B16VengefulBloodwitchOracle, 1, 1)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	life, oppLife, otherLife := me.Life, opp.Life, other.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.B16PickPlayer(t, g, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if other.Life != otherLife-1 || opp.Life != oppLife || me.Life != life+1 {
		t.Errorf("another creature dying: chosen opponent −1, you +1; got me %d opp %d other %d", me.Life-life, opp.Life-oppLife, other.Life-otherLife)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(witch) })
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != life+2 {
		t.Errorf("the Bloodwitch's own death drains too; got me +%d opp −%d", me.Life-life, oppLife-opp.Life)
	}
}

func TestB16GoblinMatronMaySearchForAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Krenko", TypeLine: "Legendary Creature — Goblin"},
		game.Card{Name: "Goblin Guide", TypeLine: "Creature — Goblin Scout"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Goblin Matron", "Creature — Goblin", effectstest.B16GoblinMatronOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ETB search is optional, so it always asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Goblin is not offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Goblin Guide")
	if !effectstest.B02bHandHasNamed(me, "Goblin Guide") {
		t.Error("the Goblin reaches the hand")
	}
}

func TestB16MurmuringMysticMakesABirdIllusionPerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Murmuring Mystic", "Creature — Human Wizard", effectstest.B16MurmuringMysticOracle, 1, 5)
	b15CastInstant(t, g, "Opt", "U")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Bird Illusion") != 1 {
		t.Fatal("an instant makes one Bird Illusion")
	}
	bird := effectstest.FindBattlefieldByName(g, "Bird Illusion")
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") || effectstest.EffectivePower(t, g, bird) != 1 {
		t.Error("the token is a 1/1 with flying")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Bird Illusion") != 1 {
		t.Error("a creature spell makes nothing")
	}
}

func TestB16JuriGrowsOnSacrificeAndDealsItsPowerWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	juri := effectstest.B12Push(g, me.ID, "Juri, Master of the Revue", "Legendary Creature — Human Shaman", effectstest.B16JuriOracle, 1, 1)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1, "R")
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, juri, "+1/+1"); got != 2 {
		t.Fatalf("%d counters after two sacrifices, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, juri, "+1/+1"); got != 2 {
		t.Fatalf("an opponent's creature being destroyed is not your sacrifice; %d counters", got)
	}

	before := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(juri) })
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("Juri's death deals its last-known power (3) to the target, got %d", before-opp.Life)
	}
}

func TestB16BartolomeEatsAnotherCreatureOrArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bart := effectstest.B12Push(g, me.ID, "Bartolomé del Presidio", "Legendary Creature — Vampire Knight", effectstest.B16BartolomeDelPresidioOracle, 2, 1)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, me.ID, "Shrine", "Enchantment")
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bart}}); err == nil {
		t.Fatal("Bartolomé can't sacrifice himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}}); err == nil {
		t.Fatal("an enchantment is neither a creature nor an artifact")
	}
	effectstest.B16Activate(t, g, me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	effectstest.B16Activate(t, g, me.ID, bart, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(rock) {
		t.Error("the sacrificed permanents are gone")
	}
	if got := effectstest.CounterCount(g, bart, "+1/+1"); got != 2 {
		t.Errorf("%d counters after two sacrifices, want 2", got)
	}
}

func TestB16ZacamaUntapsLandsOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	effectstest.B16Tap(g, land)
	effectstest.B16Tap(g, rock)
	effectstest.CastCatalogSpell(t, g, "Zacama, Primal Calamity", "Legendary Creature — Elder Dinosaur", effectstest.B16ZacamaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("a cast Zacama untaps your lands")
	}
	if !effectstest.B16Tapped(t, g, rock) {
		t.Error("only lands untap")
	}

	// Reanimated: the ETB fires, the intervening if fails, nothing
	// untaps.
	zacama := effectstest.FindBattlefieldByName(g, "Zacama, Primal Calamity")
	effectstest.B15Destroy(t, g, zacama)
	effectstest.B16Tap(g, land)
	g.WithWriteLock(func() {
		_ = g.ReturnFromGraveyardUnderControlForEffect(zacama, game.ZoneBattlefield, me.ID)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("a reanimated Zacama was not cast — no untap")
	}
}

func TestB16ZacamaActivatedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zacama := effectstest.B12Push(g, me.ID, "Zacama, Primal Calamity", "Legendary Creature — Elder Dinosaur", effectstest.B16ZacamaOracle, 9, 9)
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3, "G")
	rock := effectstest.B12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "R")
	effectstest.B16Activate(t, g, me.ID, zacama, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)})
	if g.Battlefield.Contains(bear) {
		t.Error("{2}{R}: 3 damage kills a 3/3")
	}
	effectstest.B06AddMana(me, "C", "C", "G")
	effectstest.B16Activate(t, g, me.ID, zacama, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(rock)})
	if g.Battlefield.Contains(rock) {
		t.Error("{2}{G}: the artifact is destroyed")
	}
	before := me.Life
	effectstest.B06AddMana(me, "C", "C", "W")
	effectstest.B16Activate(t, g, me.ID, zacama, 2, game.ActivateAbilityParams{})
	if me.Life != before+3 {
		t.Errorf("{2}{W}: gained %d, want 3", me.Life-before)
	}
	if !effectstest.HasEffectiveKeyword(t, g, zacama, "trample") || !effectstest.HasEffectiveKeyword(t, g, zacama, "reach") || !effectstest.HasEffectiveKeyword(t, g, zacama, "vigilance") {
		t.Error("reach, vigilance, trample")
	}
}

func TestB16BennieBracksDrawsAtAnyEndStepAfterATokenAndConvokes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soldiers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	id, err := effectstest.CastWithTapParams(t, g, "Bennie Bracks, Zoologist", "Legendary Creature — Elf Druid", "{3}{W}",
		effectstest.B16BennieBracksOracle, game.CastSpellParams{TapIDs: soldiers})
	if err != nil {
		t.Fatalf("convoke cast: %v", err)
	}
	for _, s := range soldiers {
		if !effectstest.TapCostTapped(g, s) {
			t.Error("convoke taps the named creatures")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Bennie did not resolve")
	}

	// No token this turn: the end step draws nothing.
	hand := me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("no token created this turn — no draw")
	}

	// An opponent's turn: a token created by Bennie's controller
	// during it still draws at that turn's end step ("each end step").
	effectstest.AdvanceToMainOf(t, g, 1)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1) })
	hand = me.Hand.Size()
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d at an opponent's end step after making a token, want 1", me.Hand.Size()-hand)
	}
}

func TestB16AureliaFiresOncePerDeclarationAtThreeAndFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aurelia := effectstest.B12Push(g, me.ID, "Aurelia, the Law Above", "Legendary Creature — Angel", effectstest.B16AureliaTheLawAboveOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G"))
	}
	hand := me.Hand.Size()
	// #859: the declaration is one turn-based action announced at its
	// lock-in, so the five clicks are ONE declaration however they are
	// spread. Each ability fires once for it — never once per
	// attacker past the threshold, which is what the dedupe is for.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range bears {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	// Two differing triggers from one controller, but a draw and a
	// drain commute (both rows declare footprints, ADR 0018's
	// 2026-10-14 amendment), so no CR 603.3b prompt holds them.
	if ch := effectstest.TriggerOrderPrompt(g); ch != nil {
		t.Fatalf("a draw and a drain asked for an order: %+v", ch)
	}
	if n := effectstest.TriggersOnStackFrom(g, aurelia); n != 2 {
		t.Fatalf("five attackers: %d triggers, want 2 (draw + drain, one each)", n)
	}
	lives := b15Lives(g)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Aurelia, the Law Above", "Legendary Creature — Angel", effectstest.B16AureliaTheLawAboveOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G"))
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, me.ID, bears...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent attacking with three draws Aurelia's controller a card; drew %d", me.Hand.Size()-hand)
	}
}

func TestB16HorizonExplorerLandsEnterUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Horizon Explorer", "Creature — Insect Scout", effectstest.B16HorizonExplorerOracle, 2, 4)
	gate := effectstest.PlayLandFromHand(t, g, "Simic Guildgate", effectstest.B16SimicGuildgateOracle)
	effectstest.Top100AssertEnteredUntapped(t, g, gate, "Simic Guildgate")
	if len(g.PendingChoices) != 0 {
		t.Fatal("no ordering prompt: the Explorer applies after the land's own effect")
	}

	// A basic fetched "onto the battlefield tapped" enters untapped
	// with no prompt: the search seeds the flag, so the Explorer is
	// the only applicable effect.
	//
	// A fetched Guildgate is the one entry that raises the CR 616.1
	// ordering prompt, because the seeded flag and the Gate's own
	// clause are applicable at once. It used to be a declared retreat
	// — the helper left such a land alone because the search entry
	// could not pause without stranding it — and #478 made that entry
	// resumable, so #732 took the retreat out.
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		game.Card{Name: "Dimir Guildgate", TypeLine: "Land — Gate", OracleID: effectstest.B16DimirGuildgateOracle},
	)
	effectstest.CastCatalogSpell(t, g, "Circuitous Route", "Sorcery", effectstest.B16CircuitousRouteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Dimir Guildgate")
	if effectstest.B16Tapped(t, g, effectstest.FindBattlefieldByName(g, "Forest")) {
		t.Error("a fetched basic enters untapped under the Explorer")
	}
	order := effectstest.B06ReplacementOrderFor(g, me.ID)
	if order == nil {
		t.Fatal("a fetched Guildgate under the Explorer raised no CR 616 ordering prompt")
	}
	// The Explorer's clause last is the untapping order.
	var gateEff, explorerEff game.ReplacementEffectID
	for _, id := range order.ReplacementEffectIDs {
		if _, src := g.ReplacementOptionMetaForEffect(id); src == effectstest.FindBattlefieldByName(g, "Horizon Explorer") {
			explorerEff = id
		} else {
			gateEff = id
		}
	}
	if len(order.ReplacementEffectIDs) != 2 || gateEff == 0 || explorerEff == 0 {
		t.Fatalf("the prompt should list the Gate's clause and the Explorer's, got %v", order.ReplacementEffectIDs)
	}
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{gateEff, explorerEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	dimir := effectstest.FindBattlefieldByName(g, "Dimir Guildgate")
	if dimir == uuid.Nil {
		t.Fatal("the fetched Guildgate was stranded by the ordering prompt")
	}
	if effectstest.B16Tapped(t, g, dimir) {
		t.Error("a fetched Guildgate enters untapped under the Explorer once the order is answered")
	}
	if len(g.PendingChoices) != 0 || me.Library.Size() != 1 {
		t.Error("nothing is stranded and nothing prompts")
	}
}

func TestB16HorizonExplorerMakesALanderPerAttackThatFetches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	explorer := effectstest.B12Push(g, me.ID, "Horizon Explorer", "Creature — Insect Scout", effectstest.B16HorizonExplorerOracle, 2, 4)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.DeclareAttack(t, g, opp.ID, explorer, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if effectstest.B16CountNamed(g, "Lander") != 1 {
		t.Fatalf("two attackers make one Lander, got %d", effectstest.B16CountNamed(g, "Lander"))
	}
	lander := effectstest.FindBattlefieldByName(g, "Lander")
	// The Explorer leaves first, so the Lander's own "tapped" clause
	// is what is being read.
	effectstest.B15Destroy(t, g, explorer)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, lander, 0, game.ActivateAbilityParams{})
	effectstest.TakeSearchPrompts(t, g)
	if g.Battlefield.Contains(lander) {
		t.Error("the Lander is sacrificed as a cost")
	}
	forest := effectstest.FindBattlefieldByName(g, "Forest")
	if forest == uuid.Nil || !effectstest.B16Tapped(t, g, forest) {
		t.Error("the Lander fetches a basic land onto the battlefield tapped")
	}
}

func TestB16UnstoppablePlanUntapsNonlandsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Unstoppable Plan", "Enchantment", effectstest.B16UnstoppablePlanOracle, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	rock := effectstest.B12Permanent(g, me.ID, "Mana Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B16Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	for _, id := range []uuid.UUID{bear, rock, land, theirs} {
		effectstest.B16Tap(g, id)
	}
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, bear) || effectstest.B16Tapped(t, g, rock) {
		t.Error("your nonland permanents untap")
	}
	if !effectstest.B16Tapped(t, g, land) || !effectstest.B16Tapped(t, g, theirs) {
		t.Error("lands and opponents' permanents stay tapped")
	}
}

func TestB16PsychicCorrosionMillsEachOpponentTwoPerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Psychic Corrosion", "Enchantment", effectstest.B16PsychicCorrosionOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	graves := make([]int, 0, 3)
	for _, p := range g.Seats[1:] {
		graves = append(graves, p.Graveyard.Size())
	}
	mine := me.Graveyard.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats[1:] {
		if p.Graveyard.Size() != graves[i]+4 {
			t.Errorf("opponent %d milled %d, want 4 (two draws)", i, p.Graveyard.Size()-graves[i])
		}
	}
	if me.Graveyard.Size() != mine {
		t.Error("the controller mills nothing")
	}
}

// ADR 0097 (#1749): the controller chooses each mode, and "that
// hasn't been chosen this turn" is the engine's memory of what this
// enchantment's ability chose.
func TestB16TevalsJudgmentChoosesAModeNotChosenThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teval := effectstest.B12Push(g, me.ID, "Teval's Judgment", "Enchantment", effectstest.B16TevalsJudgmentOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	graveyard := func(p *game.Player, n int) []uuid.UUID {
		var ids []uuid.UUID
		for i := 0; i < n; i++ {
			ids = append(ids, effectstest.PushGraveyardCardForTest(p, "Dead"))
		}
		return ids
	}

	// An opponent's graveyard leaving is not yours.
	theirs := graveyard(opp, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs[0]) })
	if effectstest.ModePickChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's graveyard is not yours")
	}

	// Two cards at once: ONE trigger, and its controller chooses —
	// here the Zombie Druid first, which the old printed-order shape
	// could never give.
	first := graveyard(me, 2)
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(first) })
	if n := len(effectstest.ModePickChoicesFor(g, me.ID)); n != 1 {
		t.Fatalf("one or more cards leaving is one trigger, got %d prompts", n)
	}
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Zombie Druid") != 1 {
		t.Fatal("the chosen mode made a Zombie Druid")
	}

	// Second batch: the Zombie is used.
	card := graveyard(me, 1)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(card[0]) })
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1}) || !slices.Equal(c.ModeUsedIndex, []int{2}) {
		t.Fatalf("the draw and the Treasure remain: %+v", c)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	// Third: the draw alone. Fourth: nothing at all (CR 700.2b).
	hand := me.Hand.Size()
	card = graveyard(me, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(card[0]) })
	c = effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0}) {
		t.Fatalf("the draw alone: %+v", c)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the third batch drew: %d", me.Hand.Size()-hand)
	}
	card = graveyard(me, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(card[0]) })
	if effectstest.ModePickChoiceFor(g, me.ID) != nil || effectstest.TriggerOnStack(g, teval) != nil {
		t.Fatal("with all three used, the fourth trigger is removed")
	}
	if effectstest.B16CountNamed(g, "Treasure") != 1 || effectstest.B16CountNamed(g, "Zombie Druid") != 1 {
		t.Error("one of each, no more")
	}

	// A new turn resets the memory.
	effectstest.AdvanceToMainOf(t, g, 1)
	card = graveyard(me, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(card[0]) })
	if c := effectstest.ModePickChoiceFor(g, me.ID); c == nil || len(c.ModeOptionIndex) != 3 {
		t.Fatalf("next turn, every mode is available again: %+v", c)
	}
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)

	// A flashback cast leaves the graveyard too: the trigger goes on
	// the stack above the spell once its mode is chosen.
	effectstest.AdvanceToMainOf(t, g, 0)
	looting := effectstest.SeedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	hand = me.Hand.Size()
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	c = effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("a flashback cast leaves the graveyard: the trigger asks for its mode")
	}
	effectstest.AnswerMode(t, g, c, me.ID, 0)
	if effectstest.TriggerOnStack(g, teval) == nil {
		t.Fatal("the trigger goes on the stack above the spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("drew %d, want 3 (Teval's one plus the Looting's two)", me.Hand.Size()-hand)
	}
}
