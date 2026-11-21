package testsav

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b25Hands reads every seat's hand size.
func b25Hands(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Hand.Size())
	}
	return out
}

// b25CastCommander seeds a commander in the active seat's command
// zone and casts it from there.
func b25CastCommander(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Command.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: active.ID, Controller: active.ID, IsCommander: true,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast %s: %v", name, err)
	}
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch25CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B25BloodPetOracle:            "Blood Pet",
		effectstest.B25VolatileFaultOracle:       "Volatile Fault",
		effectstest.B25QuickStudyOracle:          "Quick Study",
		effectstest.B25NadirKrakenOracle:         "Nadir Kraken",
		effectstest.B25HeritageReclamationOracle: "Heritage Reclamation",
		effectstest.B25BredForTheHuntOracle:      "Bred for the Hunt",
		effectstest.B25TevalOracle:               "Teval, the Balanced Scale",
		effectstest.B25ZenithChroniclerOracle:    "Zenith Chronicler",
		effectstest.B25PashalikMonsOracle:        "Pashalik Mons",
		effectstest.B25RiverOfTearsOracle:        "River of Tears",
		effectstest.B25RecklessHandlingOracle:    "Reckless Handling",
		effectstest.B25EssenceScatterOracle:      "Essence Scatter",
		effectstest.B25BladeHistorianOracle:      "Blade Historian",
		effectstest.B25CarefulStudyOracle:        "Careful Study",
		effectstest.B25ScorchedGeyserOracle:      "Scorched Geyser",
		effectstest.B25CosmosElixirOracle:        "Cosmos Elixir",
		effectstest.B25RimewoodFallsOracle:       "Rimewood Falls",
		effectstest.B25VegaOracle:                "Vega, the Watcher",
		effectstest.B25VoraciousHydraOracle:      "Voracious Hydra",
		// #1218: no longer a skip — see the note below the skip map.
		cardsco.CurrencyConverterOracle: "Currency Converter",
	}
	if len(want) != 20 {
		t.Fatalf("the batch registers 20 cards, the table lists %d", len(want))
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
	// The seven remaining declared skips must stay out until their seam
	// lands (Scholar of New Horizons, skipped for a counter-removal
	// cost, shipped with the Tidus deck request, #3246): an untap-step
	// event, an
	// opponent's non-mana choice at resolution (three cards), a
	// reveal-from-hand entry choice, a per-player "as though it had
	// flash" plus a trigger replacement, a legendary-sorcery cast
	// restriction, a card put from hand at resolution. Dawnsire, Sunstar
	// Dreadnought, skipped for a tap-another-creature cost, shipped
	// with station (S58 PR 7). Shadrix Silverquill, skipped for "a
	// beginning-of-combat event plus a modal trigger", shipped once both
	// existed (S79, #4834; shadrix_silverquill.go).
	//
	// Murkfiend Liege came off this list too (#5948): its untap clause is
	// Seedborn Muse's UntapStep permission, which the engine has built since
	// (murkfiend_liege.go).
	//
	// Two cards came off this list since it was written, each closing
	// a different seam. Cloudstone Curio, with #1223: "a
	// resolution-time choice constrained by the triggering permanent"
	// was the trigger-data seam, and its clause is now built from the
	// event (TriggeredAbility.TargetsFrom). Currency Converter, with
	// #1218: it was declared skipped here for "an 'exiled with this'
	// record plus an exile-to-graveyard move" and shipped using
	// b27ExiledWith (which did not exist when this batch ran) plus
	// ordinary primitives — see currency_converter.go — with no new
	// engine seam needed.
	for oracle, name := range map[string]string{
		"5b3b5f6a-375e-4479-9384-a942eda83f9b": "Gandalf the White",
	} {
		if _, ok := effects.Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #387 but is registered — update the issue", name)
		}
	}
}

// --- the mana sources ----------------------------------------------

func TestB25BloodPetCracksForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pet := effectstest.PushCatalogPermanent(g, me.ID, "Blood Pet", "Creature — Thrull", effectstest.B25BloodPetOracle, true)
	// Summoning sick, and it does not matter: no tap in the cost.
	if err := g.ActivateManaAbility(me.ID, pet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want B", got)
	}
	if g.Battlefield.Contains(pet) || !me.Graveyard.Contains(pet) {
		t.Error("the Thrull is sacrificed by its own ability")
	}
	if spec, _ := effects.Lookup(effectstest.B25BloodPetOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("a sacrifice-for-mana creature is whole")
	}
}

func TestB25CycleRowsEnterTappedAndTapForTheirColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Rimewood Falls always enters tapped.
	falls := effectstest.PlayLandFromHand(t, g, "Rimewood Falls", effectstest.B25RimewoodFallsOracle)
	if !effectstest.B25Tapped(t, g, falls) {
		t.Error("Rimewood Falls enters tapped")
	}
	// Scorched Geyser enters tapped with fewer than two basics...
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	geyser := effectstest.PlayLandFromHand(t, g, "Scorched Geyser", effectstest.B25ScorchedGeyserOracle)
	if !effectstest.B25Tapped(t, g, geyser) {
		t.Error("Scorched Geyser enters tapped without two basics")
	}
	// ...and untapped with two.
	effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	geyser2 := effectstest.PlayLandFromHand(t, g, "Scorched Geyser", effectstest.B25ScorchedGeyserOracle)
	if effectstest.B25Tapped(t, g, geyser2) {
		t.Error("Scorched Geyser enters untapped with two basics")
	}
	if err := g.ActivateManaAbility(me.ID, geyser2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(g.PendingChoices) == 0 || len(g.PendingChoices[len(g.PendingChoices)-1].ColorOptions) != 2 {
		t.Fatal("a dual offers its two colours")
	}
	for _, o := range []string{effectstest.B25RimewoodFallsOracle, effectstest.B25ScorchedGeyserOracle} {
		if spec, _ := effects.Lookup(o); spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is whole", spec.Name)
		}
	}
}

func TestB25RiverOfTearsIsBlueUntilYouPlayALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	river := effectstest.B21Push(g, me.ID, "River of Tears", "Land", effectstest.B25RiverOfTearsOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("no land played: pool %v, want U", got)
	}
	me.ManaPool = nil
	effectstest.B22Untap(g, river)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("a land played this turn: pool %v, want B", got)
	}
	// On an opponent's turn nobody has played a land for you. (The
	// River untaps in YOUR untap step, so untap it by hand here.)
	me.ManaPool = nil
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.B22Untap(g, river)
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("an opponent's turn: pool %v, want U", got)
	}
}

// --- the spells ----------------------------------------------------

func TestB25QuickStudyAndCarefulStudyDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Quick Study", "Instant", effectstest.B25QuickStudyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("Quick Study: hand %d → %d, want +2", hand, me.Hand.Size())
	}
	hand = me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Careful Study", "Sorcery", effectstest.B25CarefulStudyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("Careful Study: drew two, then a discard prompt; hand %d → %d", hand, me.Hand.Size())
	}
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Fatalf("Careful Study then asks the caster to discard two: %d owed", effectstest.DiscardOwed(g, me.ID))
	}
}

func TestB25EssenceScatterCountersOnlyACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A noncreature spell is refused as a target.
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.HandCardFull(me, "Essence Scatter", "Instant", "{1}{U}", effectstest.B25EssenceScatterOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err == nil {
		t.Error("a noncreature spell is refused as a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A creature spell is cast on its controller's own turn; the
	// Scatter answers it from across the table.
	bear := effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Essence Scatter", "Instant", effectstest.B25EssenceScatterOracle),
		game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("Essence Scatter at a creature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell is countered")
	}
}

func TestB25HeritageReclamationThreeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := effectstest.B12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	dead := effectstest.B17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")

	effectstest.CastModal(t, g, "Heritage Reclamation", "Instant", effectstest.B25HeritageReclamationOracle, []int{0}, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode one destroys the artifact")
	}
	effectstest.CastModal(t, g, "Heritage Reclamation", "Instant", effectstest.B25HeritageReclamationOracle, []int{1}, effectstest.B16TargetCard(shrine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("mode two destroys the enchantment")
	}
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Heritage Reclamation", "Instant", effectstest.B25HeritageReclamationOracle, []int{2}, effectstest.B16TargetCard(dead))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, dead) {
		t.Error("mode three exiles the graveyard card")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	// "Up to one": the third mode with no target is a cantrip.
	hand = me.Hand.Size()
	effectstest.CastModal(t, g, "Heritage Reclamation", "Instant", effectstest.B25HeritageReclamationOracle, []int{2}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("no target: still draws; hand %d → %d", hand, me.Hand.Size())
	}
	// A mode's clause is enforced.
	id := effectstest.HandCardFull(me, "Heritage Reclamation", "Instant", "{1}{G}", effectstest.B25HeritageReclamationOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: effectstest.B16TargetCard(shrine)}); err == nil {
		t.Error("the artifact mode refuses an enchantment")
	}
}

func TestB25RecklessHandlingTutorsThenDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := effectstest.StapleLibraryCard(me, "Sol Ring", "Artifact")
	lives := b22Lives(g)
	// An empty hand: the only card to discard is the tutored artifact.
	me.Hand.Cards = nil
	effectstest.CastCatalogSpell(t, g, "Reckless Handling", "Sorcery", effectstest.B25RecklessHandlingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if me.Hand.Contains(ring) || !me.Graveyard.Contains(ring) {
		t.Error("the artifact was tutored to hand and then discarded at random")
	}
	for i := 1; i < 4; i++ {
		if g.Seats[i].Life != lives[i]-2 {
			t.Errorf("an artifact discarded: seat %d takes 2 (%d → %d)", i, lives[i], g.Seats[i].Life)
		}
	}
	if me.Life != lives[0] {
		t.Error("the caster is not an opponent")
	}
	// No artifact in the library, a nonartifact in hand: the discard
	// is not an artifact, so no damage.
	lives = b22Lives(g)
	me.Hand.Cards = nil
	bear := effectstest.B20HandCard(me, "Bear", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Reckless Handling", "Sorcery", effectstest.B25RecklessHandlingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Error("with nothing found, a card is still discarded at random")
	}
	if opp.Life != lives[1] {
		t.Error("a nonartifact discard deals no damage")
	}
}

// --- the permanents ------------------------------------------------

func TestB25CosmosElixirGainsBelowStartingLifeAndDrawsAbove(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B21Push(g, me.ID, "Cosmos Elixir", "Artifact", effectstest.B25CosmosElixirOracle, 0, 0)
	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand {
		t.Errorf("at starting life: gain 2, no draw; life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	hand, life = me.Hand.Size(), me.Life
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life {
		t.Errorf("above starting life: draw, no gain; hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	// An opponent's end step is not yours.
	hand, life = me.Hand.Size(), me.Life
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Error("an opponent's end step does nothing")
	}
}

func TestB25VegaDrawsForCastsFromOutsideTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vega := effectstest.B21Push(g, me.ID, "Vega, the Watcher", "Legendary Creature — Bird Spirit", effectstest.B25VegaOracle, 2, 2, "W", "U")
	if !effectstest.HasEffectiveKeyword(t, g, vega, "flying") {
		t.Error("flying")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a cast from hand draws nothing")
	}
	hand = me.Hand.Size()
	b25CastCommander(t, g, "My Commander", "Legendary Creature — Human")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a cast from the command zone draws one: hand %d → %d", hand, me.Hand.Size())
	}
	if spec, _ := effects.Lookup(effectstest.B25VegaOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Vega is whole")
	}
}

func TestB25ZenithChroniclerDrawsTheTableOnAFirstGoldSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Zenith Chronicler", "Artifact Creature — Phyrexian Construct", effectstest.B25ZenithChroniclerOracle, 3, 1)
	hands := b25Hands(g)
	effectstest.B13OpponentCasts(t, g, opp, "Their Gold Spell", "Instant", "", "{W}{U}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	got := b25Hands(g)
	if got[0] != hands[0]+1 || got[2] != hands[2]+1 || got[3] != hands[3]+1 {
		t.Errorf("each OTHER player draws: %v → %v", hands, got)
	}
	if got[1] != hands[1] {
		t.Error("the caster does not draw")
	}
	// Their second gold spell this turn is not their first.
	hands = b25Hands(g)
	effectstest.B13OpponentCasts(t, g, opp, "Their Second Gold Spell", "Instant", "", "{B}{R}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b25Hands(g); got[0] != hands[0] {
		t.Error("a second multicolored spell in the same turn does nothing")
	}
	// A monocolored spell is not multicolored, whoever casts it.
	hands = b25Hands(g)
	effectstest.CastCatalogSpell(t, g, "My Mono Spell", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b25Hands(g); got[1] != hands[1] {
		t.Error("a monocolored spell does nothing")
	}
	// The controller's own first gold spell draws the opponents.
	hands = b25Hands(g)
	id := effectstest.HandCardFull(me, "My Gold Spell", "Instant", "{G}{U}", "", []string{"G", "U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got = b25Hands(g)
	if got[1] != hands[1]+1 || got[2] != hands[2]+1 || got[0] != hands[0] {
		t.Errorf("the controller's gold spell draws the opponents, not the controller: %v → %v", hands, got)
	}
	// A new turn resets the count.
	effectstest.AdvanceToMainOf(t, g, 1)
	hands = b25Hands(g)
	effectstest.B13OpponentCasts(t, g, opp, "Their Gold Spell Again", "Instant", "", "{W}{U}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b25Hands(g); got[0] != hands[0]+1 {
		t.Error("the next turn's first gold spell counts again")
	}
}

func TestB25BredForTheHuntDrawsWhenACounteredCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Bred for the Hunt", "Enchantment", effectstest.B25BredForTheHuntOracle, 0, 0)
	grown := effectstest.PushVanillaCreature(g, me.ID, "Grown Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, "+1/+1", 1) })
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, grown, plain)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("the countered creature connecting asks you to draw")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1 — the plain creature's damage asks nothing", me.Hand.Size()-hand)
	}
	if effectstest.B19HasTriggerPromptFor(g, me.ID) {
		t.Error("only the creature with a +1/+1 counter triggers")
	}
}

func TestB25NadirKrakenPaysOneToGrowAndSpawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.B21Push(g, me.ID, "Nadir Kraken", "Creature — Kraken", effectstest.B25NadirKrakenOracle, 2, 3, "U")
	effectstest.B25Draw(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("your draw asks you to pay {1}")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if effectstest.CounterCount(g, kraken, "+1/+1") != 1 {
		t.Errorf("paid: a +1/+1 counter, got %d", effectstest.CounterCount(g, kraken, "+1/+1"))
	}
	if effectstest.B16CountNamed(g, "Tentacle") != 1 {
		t.Fatal("paid: a Tentacle")
	}
	tentacle := effectstest.FindBattlefieldByName(g, "Tentacle")
	if c, _ := effectstest.BattlefieldCard(g, tentacle); c.Power != 1 || c.Toughness != 1 || !c.HasColor("U") {
		t.Error("a 1/1 blue Tentacle")
	}
	effectstest.B25Draw(g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if effectstest.CounterCount(g, kraken, "+1/+1") != 1 || effectstest.B16CountNamed(g, "Tentacle") != 1 {
		t.Error("declined: nothing")
	}
	effectstest.B25Draw(g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's draw is not yours")
	}
}

func TestB25BladeHistorianGivesAttackersDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := effectstest.B21Push(g, me.ID, "Blade Historian", "Creature — Human Cleric", effectstest.B25BladeHistorianOracle, 2, 3, "R", "W")
	home := effectstest.PushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, historian)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, historian), "double strike") {
		t.Fatal("the Historian attacking gives itself double strike")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, home), "double strike") {
		t.Error("a creature that did not attack is not an attacking creature")
	}
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 2/3 with double strike deals 4: %d → %d", life, opp.Life)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "double strike") {
		t.Error("an opponent's attacker is not yours")
	}
	if spec, _ := effects.Lookup(effectstest.B25BladeHistorianOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Blade Historian is a layer 6 static now: nothing is left to declare")
	}
}

func TestB25VolatileFaultDestroysANonbasicAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fault := effectstest.PushCatalogPermanent(g, me.ID, "Volatile Fault", "Land — Cave", effectstest.B25VolatileFaultOracle, false)
	cradle := effectstest.SeedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	forest := effectstest.SeedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	swamp := effectstest.StapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, fault, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(forest)}); err == nil {
		t.Error("a basic land is refused")
	}
	if err := g.ActivateCatalogAbility(me.ID, fault, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(cradle)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fault) {
		t.Error("the Fault is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cradle) {
		t.Fatal("the nonbasic land is destroyed")
	}
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the victim gets a (declinable) search for a basic")
	}
	if effectstest.B16CountNamed(g, "Treasure") != 0 {
		t.Error("the Treasure waits for the victim's answer")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, swamp)
	if !g.Battlefield.Contains(swamp) {
		t.Error("the victim's basic enters")
	}
	if effectstest.B16CountNamed(g, "Treasure") != 1 {
		t.Error("you create a Treasure")
	}
	if tr := effectstest.FindBattlefieldByName(g, "Treasure"); tr == uuid.Nil || effectstest.CardByID(g, tr).Controller != me.ID {
		t.Error("the Treasure is yours")
	}
}

func TestB25PashalikMonsPingsForGoblinDeathsAndMakesGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mons := effectstest.B21Push(g, me.ID, "Pashalik Mons", "Legendary Creature — Goblin Warrior", effectstest.B25PashalikMonsOracle, 2, 2, "R")
	goblin := effectstest.B16Creature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2, "R")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	effectstest.AdvanceToMain(t, g)
	life := opp.Life

	effectstest.B25Destroy(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Goblin dying is nothing")
	}
	effectstest.B25Destroy(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's Goblin dying is nothing")
	}
	effectstest.B25Destroy(g, goblin)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("another Goblin you control dying: 1 damage to the chosen target (%d → %d)", life, opp.Life)
	}

	// The activation sacrifices a Goblin — Mons himself is one — and
	// his own death pings too.
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, mons, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{mons}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("Mons dying to his own ability pings (%d → %d)", life, opp.Life)
	}
	if n := effectstest.B16CountNamed(g, "Goblin"); n != 2 {
		t.Errorf("two 1/1 red Goblins: %d", n)
	}
	if spec, _ := effects.Lookup(effectstest.B25PashalikMonsOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Mons is whole")
	}
}

func TestB25TevalMillsReturnsALandAndMakesZombiesWhenCardsLeave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teval := effectstest.B21Push(g, me.ID, "Teval, the Balanced Scale", "Legendary Creature — Spirit Dragon", effectstest.B25TevalOracle, 4, 4, "B", "G", "U")
	if !effectstest.HasEffectiveKeyword(t, g, teval, "flying") {
		t.Error("flying")
	}
	forest := effectstest.B17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	library := me.Library.Size()
	grave := me.Graveyard.Size()
	effectstest.DeclareAttack(t, g, opp.ID, teval)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if p == nil || !effectstest.HasID(p.ChooseCards, forest) || p.ChooseMin != 0 || p.ChooseMax != 1 {
		t.Fatal("you may return a land card: a floor of zero, asked after the mill")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library-3 {
		t.Errorf("mill three: library %d → %d", library, me.Library.Size())
	}
	if !g.Battlefield.Contains(forest) || !effectstest.B25Tapped(t, g, forest) {
		t.Error("the chosen land returns tapped")
	}
	if me.Graveyard.Size() != grave+3-1 {
		t.Errorf("graveyard %d → %d: three milled, one returned", grave, me.Graveyard.Size())
	}
	if n := effectstest.B16CountNamed(g, "Zombie Druid"); n != 1 {
		t.Errorf("the land leaving the graveyard makes a Zombie Druid: %d", n)
	}
	// Two cards leaving at once is one trigger.
	var milled []uuid.UUID
	for _, c := range me.Graveyard.Cards {
		milled = append(milled, c.InstanceID)
	}
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(milled) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Zombie Druid"); n != 2 {
		t.Errorf("one or more cards leaving is one Zombie: %d", n)
	}
	// With no land card in the graveyard the attack still mills.
	effectstest.AdvanceToMainOf(t, g, 0)
	library = me.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, teval)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("nothing to return: no prompt")
	}
	if me.Library.Size() != library-3 {
		t.Error("the mill does not depend on a land being in the graveyard")
	}
	if spec, _ := effects.Lookup(effectstest.B25TevalOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the land is chosen after the mill: nothing is left to declare")
	}
}

// The land is chosen after the mill, so one the trigger itself milled
// is a legal pick, and "you may" lets the player take none.
func TestB25TevalCanReturnALandItJustMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teval := effectstest.B21Push(g, me.ID, "Teval, the Balanced Scale", "Legendary Creature — Spirit Dragon", effectstest.B25TevalOracle, 4, 4, "B", "G", "U")
	milled := effectstest.B10LibraryTop(me, "Milled Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.B10LibraryTop(me, "Filler A", "Instant", "{R}", 0, 0)
	effectstest.B10LibraryTop(me, "Filler B", "Instant", "{R}", 0, 0)
	effectstest.DeclareAttack(t, g, opp.ID, teval)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if p == nil || !effectstest.HasID(p.ChooseCards, milled) {
		t.Fatal("a land the trigger milled is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, milled)
	if !g.Battlefield.Contains(milled) || !effectstest.B25Tapped(t, g, milled) {
		t.Error("the milled land returns tapped")
	}
}

func TestB25VoraciousHydraFightsOrDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	// Choose the fight mode: a real "choose one" prompt, then the target.
	hydra := effectstest.CastXSpell(t, g, "Voracious Hydra", "Creature — Hydra", effectstest.B25VoraciousHydraOracle, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, hydra, "+1/+1") != 3 {
		t.Fatalf("enters with X counters: %d", effectstest.CounterCount(g, hydra, "+1/+1"))
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil || !slices.Equal(c.ModeOptionIndex, []int{0, 1}) {
		t.Fatalf("both modes are offered while an opponent has a creature: %+v", c)
	}
	effectstest.AnswerMode(t, g, c, me.ID, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, theirs) || effectstest.HasID(p.PickTargetCards, mine) || p.PickTargetMin != 1 {
		t.Error("exactly one creature you don't control")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3/4 Hydra fights and kills the 2/2")
	}
	if c, _ := effectstest.BattlefieldCard(g, hydra); c.DamageMarked != 2 {
		t.Errorf("the Bear hits back for 2: %d", c.DamageMarked)
	}
	if effectstest.CounterCount(g, hydra, "+1/+1") != 3 {
		t.Error("the fight mode does not double")
	}
	// Choose the doubling mode: no target is asked, the counters double.
	theirs2 := effectstest.PushVanillaCreature(g, opp.ID, "Their Other Bear", 2, 2)
	hydra2 := effectstest.CastXSpell(t, g, "Voracious Hydra", "Creature — Hydra", effectstest.B25VoraciousHydraOracle, "{X}{G}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMode(t, g, effectstest.ModePickChoiceFor(g, me.ID), me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the doubling mode targets nothing")
	}
	if effectstest.CounterCount(g, hydra2, "+1/+1") != 4 {
		t.Errorf("doubling mode: 2 counters doubled to 4, got %d", effectstest.CounterCount(g, hydra2, "+1/+1"))
	}
	if !g.Battlefield.Contains(theirs2) {
		t.Error("no fight")
	}
	// No creature to fight at all: the fight mode is not offered.
	effectstest.B25Destroy(g, theirs2)
	effectstest.PassPriorityAroundTable(t, g)
	hydra3 := effectstest.CastXSpell(t, g, "Voracious Hydra", "Creature — Hydra", effectstest.B25VoraciousHydraOracle, "{X}{G}{G}", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ModePickChoiceFor(g, me.ID); c != nil {
		if slices.Contains(c.ModeOptionIndex, 1) {
			t.Fatalf("nothing to fight: the fight mode is not offered: %v", c.ModeOptionIndex)
		}
		effectstest.AnswerMode(t, g, c, me.ID, 0)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.CounterCount(g, hydra3, "+1/+1") != 2 {
		t.Errorf("1 counter doubled to 2, got %d", effectstest.CounterCount(g, hydra3, "+1/+1"))
	}
	if !effectstest.HasEffectiveKeyword(t, g, hydra3, "trample") {
		t.Error("trample")
	}
	if spec, _ := effects.Lookup(effectstest.B25VoraciousHydraOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the mode is a real choice now: nothing is declared")
	}
}
