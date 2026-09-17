package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch25_test.go — card-level coverage for the card-coverage
// roadmap's batch 25 (#387, `edhrec_rank` 2638–2737): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, draw, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b25-prefixed.

const (
	b25BloodPetOracle            = "e05c6c80-a91a-45e0-b991-0014fd5a6472"
	b25VolatileFaultOracle       = "95c44f28-f7fa-4785-83b9-0d81be0db0c8"
	b25QuickStudyOracle          = "7bff8d4a-1c7d-48b8-b3e3-737dd6f01823"
	b25NadirKrakenOracle         = "5a73d439-bd78-42e1-90ae-eedd30536881"
	b25HeritageReclamationOracle = "09955b4b-6052-4c27-8b63-9548482b3c5c"
	b25BredForTheHuntOracle      = "fc7cda6f-7e5e-4a56-8d67-ffdea7edf269"
	b25TevalOracle               = "c8cbf0ec-ec98-4cb3-8068-60e92bbd740d"
	b25ZenithChroniclerOracle    = "9fed6494-3bdd-4bc6-901f-da5fb623b152"
	b25PashalikMonsOracle        = "bdb94ccd-1bb5-4bb7-9539-e1b1c97c19c5"
	b25RiverOfTearsOracle        = "8a83d284-75a0-4901-b7d9-c4b7586ee327"
	b25RecklessHandlingOracle    = "f1ea7dc5-01cb-4780-947c-08ea9324c52f"
	b25EssenceScatterOracle      = "46665089-aa3d-44c3-964d-6638dfbb5782"
	b25BladeHistorianOracle      = "314f0a96-5e91-42a3-9d75-3f641f22a9ee"
	b25CarefulStudyOracle        = "32e9aa23-fc0a-4b82-82f9-1659c304428c"
	b25ScorchedGeyserOracle      = "f808b510-907a-4c3c-aea1-efb825c8e13e"
	b25CosmosElixirOracle        = "ed7300f4-831a-4ba4-b5e6-ceba8d079eaa"
	b25RimewoodFallsOracle       = "983739cd-0b36-40d9-9a03-7b6aa7ffd0df"
	b25VegaOracle                = "95e66851-7aff-4559-91b4-6b8a95c6e1f8"
	b25VoraciousHydraOracle      = "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3"
)

// b25Draw draws n cards for a player through the effect API and
// leaves whatever it triggered pending.
func b25Draw(g *game.Game, p uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p, n) })
}

// b25Destroy destroys a permanent through the effect API and leaves
// whatever it triggered pending.
func b25Destroy(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

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
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("commander cast %s: %v", name, err)
	}
	return id
}

// b25Tapped reads a battlefield card's tapped state.
func b25Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name.
func TestBatch25CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b25BloodPetOracle:            "Blood Pet",
		b25VolatileFaultOracle:       "Volatile Fault",
		b25QuickStudyOracle:          "Quick Study",
		b25NadirKrakenOracle:         "Nadir Kraken",
		b25HeritageReclamationOracle: "Heritage Reclamation",
		b25BredForTheHuntOracle:      "Bred for the Hunt",
		b25TevalOracle:               "Teval, the Balanced Scale",
		b25ZenithChroniclerOracle:    "Zenith Chronicler",
		b25PashalikMonsOracle:        "Pashalik Mons",
		b25RiverOfTearsOracle:        "River of Tears",
		b25RecklessHandlingOracle:    "Reckless Handling",
		b25EssenceScatterOracle:      "Essence Scatter",
		b25BladeHistorianOracle:      "Blade Historian",
		b25CarefulStudyOracle:        "Careful Study",
		b25ScorchedGeyserOracle:      "Scorched Geyser",
		b25CosmosElixirOracle:        "Cosmos Elixir",
		b25RimewoodFallsOracle:       "Rimewood Falls",
		b25VegaOracle:                "Vega, the Watcher",
		b25VoraciousHydraOracle:      "Voracious Hydra",
	}
	if len(want) != 19 {
		t.Fatalf("the batch registers 19 cards, the table lists %d", len(want))
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
	// The twelve declared skips must stay out until their seam
	// lands: a counter-removal cost, an untap-step event, a
	// resolution-time choice constrained by the triggering permanent,
	// an "exiled with this" record plus an exile-to-graveyard move, an opponent's
	// non-mana choice at resolution (three cards), a reveal-from-hand
	// entry choice, a per-player "as though it had flash" plus a
	// trigger replacement, a legendary-sorcery cast restriction, a
	// card put from hand at resolution, a beginning-of-combat event
	// plus a modal trigger, and a tap-another-creature cost.
	for oracle, name := range map[string]string{
		"3c7ea603-c985-4107-806b-a467b5fcca36": "Scholar of New Horizons",
		"61d28182-498f-4bbc-bb7a-c5e1ef872dda": "Murkfiend Liege",
		"5cd2fd32-4da2-40eb-b003-c0b9a9ec91c1": "Cloudstone Curio",
		"981298e6-ddee-49c0-9377-f47f019b4138": "Currency Converter",
		"8e356df5-ca92-4be2-871e-8965c2510fbe": "Tempt with Vengeance",
		"42b9d383-3fe2-4fc8-ab86-f80a288d502b": "Murmuring Bosk",
		"5b3b5f6a-375e-4479-9384-a942eda83f9b": "Gandalf the White",
		"8de8027b-072a-474b-b76b-dc49c417cc55": "Primevals' Glorious Rebirth",
		"3b6ef144-bb98-4686-9718-204f1c3cf020": "Worldsoul's Rage",
		"2143f413-7baa-4fd6-a7ca-32f52bfa553f": "Shadrix Silverquill",
		"afc9436b-8cad-4916-929d-ff33a37b42d5": "Dawnsire, Sunstar Dreadnought",
		"d08ce18a-0fd9-47ac-ad2f-6934b60070f1": "Rakdos, Patron of Chaos",
	} {
		if _, ok := Lookup(oracle); ok {
			t.Errorf("%s is declared skipped on #387 but is registered — update the issue", name)
		}
	}
}

// --- the mana sources ----------------------------------------------

func TestB25BloodPetCracksForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pet := pushCatalogPermanent(g, me.ID, "Blood Pet", "Creature — Thrull", b25BloodPetOracle, true)
	// Summoning sick, and it does not matter: no tap in the cost.
	if err := g.ActivateManaAbility(me.ID, pet, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want B", got)
	}
	if g.Battlefield.Contains(pet) || !me.Graveyard.Contains(pet) {
		t.Error("the Thrull is sacrificed by its own ability")
	}
	if spec, _ := Lookup(b25BloodPetOracle); spec.Completeness != CompletenessFull {
		t.Error("a sacrifice-for-mana creature is whole")
	}
}

func TestB25CycleRowsEnterTappedAndTapForTheirColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Rimewood Falls always enters tapped.
	falls := playLandFromHand(t, g, "Rimewood Falls", b25RimewoodFallsOracle)
	if !b25Tapped(t, g, falls) {
		t.Error("Rimewood Falls enters tapped")
	}
	// Scorched Geyser enters tapped with fewer than two basics...
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	geyser := playLandFromHand(t, g, "Scorched Geyser", b25ScorchedGeyserOracle)
	if !b25Tapped(t, g, geyser) {
		t.Error("Scorched Geyser enters tapped without two basics")
	}
	// ...and untapped with two.
	seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	seedLandOnBattlefield(g, me.ID, "Mountain", "Basic Land — Mountain")
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	geyser2 := playLandFromHand(t, g, "Scorched Geyser", b25ScorchedGeyserOracle)
	if b25Tapped(t, g, geyser2) {
		t.Error("Scorched Geyser enters untapped with two basics")
	}
	if err := g.ActivateManaAbility(me.ID, geyser2, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(g.PendingChoices) == 0 || len(g.PendingChoices[len(g.PendingChoices)-1].ColorOptions) != 2 {
		t.Fatal("a dual offers its two colours")
	}
	for _, o := range []string{b25RimewoodFallsOracle, b25ScorchedGeyserOracle} {
		if spec, _ := Lookup(o); spec.Completeness != CompletenessFull {
			t.Errorf("%s is whole", spec.Name)
		}
	}
}

func TestB25RiverOfTearsIsBlueUntilYouPlayALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	river := b21Push(g, me.ID, "River of Tears", "Land", b25RiverOfTearsOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("no land played: pool %v, want U", got)
	}
	me.ManaPool = nil
	b22Untap(g, river)
	playLandFromHand(t, g, "Forest", "")
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("a land played this turn: pool %v, want B", got)
	}
	// On an opponent's turn nobody has played a land for you. (The
	// River untaps in YOUR untap step, so untap it by hand here.)
	me.ManaPool = nil
	advanceToMainOf(t, g, 1)
	b22Untap(g, river)
	if err := g.ActivateManaAbility(me.ID, river, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("an opponent's turn: pool %v, want U", got)
	}
}

// --- the spells ----------------------------------------------------

func TestB25QuickStudyAndCarefulStudyDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Quick Study", "Instant", b25QuickStudyOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("Quick Study: hand %d → %d, want +2", hand, me.Hand.Size())
	}
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Careful Study", "Sorcery", b25CarefulStudyOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("Careful Study: drew two, then a discard prompt; hand %d → %d", hand, me.Hand.Size())
	}
	if discardOwed(g, me.ID) != 2 {
		t.Fatalf("Careful Study then asks the caster to discard two: %d owed", discardOwed(g, me.ID))
	}
}

func TestB25EssenceScatterCountersOnlyACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// A noncreature spell is refused as a target.
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := handCardFull(me, "Essence Scatter", "Instant", "{1}{U}", b25EssenceScatterOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt)}); err == nil {
		t.Error("a noncreature spell is refused as a target")
	}
	passPriorityAroundTable(t, g)
	// A creature spell is cast on its controller's own turn; the
	// Scatter answers it from across the table.
	bear := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Essence Scatter", "Instant", b25EssenceScatterOracle),
		game.CastSpellParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("Essence Scatter at a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell is countered")
	}
}

func TestB25HeritageReclamationThreeModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	shrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	dead := b17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")

	castModal(t, g, "Heritage Reclamation", "Instant", b25HeritageReclamationOracle, []int{0}, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode one destroys the artifact")
	}
	castModal(t, g, "Heritage Reclamation", "Instant", b25HeritageReclamationOracle, []int{1}, b16TargetCard(shrine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("mode two destroys the enchantment")
	}
	hand := me.Hand.Size()
	castModal(t, g, "Heritage Reclamation", "Instant", b25HeritageReclamationOracle, []int{2}, b16TargetCard(dead))
	passPriorityAroundTable(t, g)
	if !inExile(g, dead) {
		t.Error("mode three exiles the graveyard card")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("and draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	// "Up to one": the third mode with no target is a cantrip.
	hand = me.Hand.Size()
	castModal(t, g, "Heritage Reclamation", "Instant", b25HeritageReclamationOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("no target: still draws; hand %d → %d", hand, me.Hand.Size())
	}
	// A mode's clause is enforced.
	id := handCardFull(me, "Heritage Reclamation", "Instant", "{1}{G}", b25HeritageReclamationOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: b16TargetCard(shrine)}); err == nil {
		t.Error("the artifact mode refuses an enchantment")
	}
}

func TestB25RecklessHandlingTutorsThenDiscardsAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ring := stapleLibraryCard(me, "Sol Ring", "Artifact")
	lives := b22Lives(g)
	// An empty hand: the only card to discard is the tutored artifact.
	me.Hand.Cards = nil
	castCatalogSpell(t, g, "Reckless Handling", "Sorcery", b25RecklessHandlingOracle, nil)
	passPriorityAroundTable(t, g)
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
	bear := b20HandCard(me, "Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Reckless Handling", "Sorcery", b25RecklessHandlingOracle, nil)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bear) {
		t.Error("with nothing found, a card is still discarded at random")
	}
	if opp.Life != lives[1] {
		t.Error("a nonartifact discard deals no damage")
	}
}

// --- the permanents ------------------------------------------------

func TestB25CosmosElixirGainsBelowStartingLifeAndDrawsAbove(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b21Push(g, me.ID, "Cosmos Elixir", "Artifact", b25CosmosElixirOracle, 0, 0)
	hand, life := me.Hand.Size(), me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size() != hand {
		t.Errorf("at starting life: gain 2, no draw; life %d → %d, hand %d → %d", life, me.Life, hand, me.Hand.Size())
	}
	advanceToNextSeatsTurn(t, g)
	advanceToMainOf(t, g, 0)
	hand, life = me.Hand.Size(), me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life {
		t.Errorf("above starting life: draw, no gain; hand %d → %d, life %d → %d", hand, me.Hand.Size(), life, me.Life)
	}
	// An opponent's end step is not yours.
	hand, life = me.Hand.Size(), me.Life
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Error("an opponent's end step does nothing")
	}
}

func TestB25VegaDrawsForCastsFromOutsideTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vega := b21Push(g, me.ID, "Vega, the Watcher", "Legendary Creature — Bird Spirit", b25VegaOracle, 2, 2, "W", "U")
	if !hasEffectiveKeyword(t, g, vega, "flying") {
		t.Error("flying")
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Error("a cast from hand draws nothing")
	}
	hand = me.Hand.Size()
	b25CastCommander(t, g, "My Commander", "Legendary Creature — Human")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a cast from the command zone draws one: hand %d → %d", hand, me.Hand.Size())
	}
	if spec, _ := Lookup(b25VegaOracle); spec.Completeness != CompletenessFull {
		t.Error("Vega is whole")
	}
}

func TestB25ZenithChroniclerDrawsTheTableOnAFirstGoldSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Zenith Chronicler", "Artifact Creature — Phyrexian Construct", b25ZenithChroniclerOracle, 3, 1)
	hands := b25Hands(g)
	b13OpponentCasts(t, g, opp, "Their Gold Spell", "Instant", "", "{W}{U}", nil)
	passPriorityAroundTable(t, g)
	got := b25Hands(g)
	if got[0] != hands[0]+1 || got[2] != hands[2]+1 || got[3] != hands[3]+1 {
		t.Errorf("each OTHER player draws: %v → %v", hands, got)
	}
	if got[1] != hands[1] {
		t.Error("the caster does not draw")
	}
	// Their second gold spell this turn is not their first.
	hands = b25Hands(g)
	b13OpponentCasts(t, g, opp, "Their Second Gold Spell", "Instant", "", "{B}{R}", nil)
	passPriorityAroundTable(t, g)
	if got := b25Hands(g); got[0] != hands[0] {
		t.Error("a second multicolored spell in the same turn does nothing")
	}
	// A monocolored spell is not multicolored, whoever casts it.
	hands = b25Hands(g)
	castCatalogSpell(t, g, "My Mono Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := b25Hands(g); got[1] != hands[1] {
		t.Error("a monocolored spell does nothing")
	}
	// The controller's own first gold spell draws the opponents.
	hands = b25Hands(g)
	id := handCardFull(me, "My Gold Spell", "Instant", "{G}{U}", "", []string{"G", "U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	got = b25Hands(g)
	if got[1] != hands[1]+1 || got[2] != hands[2]+1 || got[0] != hands[0] {
		t.Errorf("the controller's gold spell draws the opponents, not the controller: %v → %v", hands, got)
	}
	// A new turn resets the count.
	advanceToMainOf(t, g, 1)
	hands = b25Hands(g)
	b13OpponentCasts(t, g, opp, "Their Gold Spell Again", "Instant", "", "{W}{U}", nil)
	passPriorityAroundTable(t, g)
	if got := b25Hands(g); got[0] != hands[0]+1 {
		t.Error("the next turn's first gold spell counts again")
	}
}

func TestB25BredForTheHuntDrawsWhenACounteredCreatureConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Bred for the Hunt", "Enchantment", b25BredForTheHuntOracle, 0, 0)
	grown := pushVanillaCreature(g, me.ID, "Grown Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grown, "+1/+1", 1) })
	plain := pushVanillaCreature(g, me.ID, "Plain Bear", 2, 2)
	declareAttack(t, g, opp.ID, grown, plain)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !b19HasTriggerPromptFor(g, me.ID) {
		t.Fatal("the countered creature connecting asks you to draw")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("drew %d, want 1 — the plain creature's damage asks nothing", me.Hand.Size()-hand)
	}
	if b19HasTriggerPromptFor(g, me.ID) {
		t.Error("only the creature with a +1/+1 counter triggers")
	}
}

func TestB25NadirKrakenPaysOneToGrowAndSpawn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := b21Push(g, me.ID, "Nadir Kraken", "Creature — Kraken", b25NadirKrakenOracle, 2, 3, "U")
	b25Draw(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("your draw asks you to pay {1}")
	}
	b06AddMana(me, "C")
	answerPayUnless(t, g, me.ID, true)
	if counterCount(g, kraken, "+1/+1") != 1 {
		t.Errorf("paid: a +1/+1 counter, got %d", counterCount(g, kraken, "+1/+1"))
	}
	if b16CountNamed(g, "Tentacle") != 1 {
		t.Fatal("paid: a Tentacle")
	}
	tentacle := findBattlefieldByName(g, "Tentacle")
	if c, _ := battlefieldCard(g, tentacle); c.Power != 1 || c.Toughness != 1 || !c.HasColor("U") {
		t.Error("a 1/1 blue Tentacle")
	}
	b25Draw(g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	if counterCount(g, kraken, "+1/+1") != 1 || b16CountNamed(g, "Tentacle") != 1 {
		t.Error("declined: nothing")
	}
	b25Draw(g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's draw is not yours")
	}
}

func TestB25BladeHistorianGivesAttackersDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	historian := b21Push(g, me.ID, "Blade Historian", "Creature — Human Cleric", b25BladeHistorianOracle, 2, 3, "R", "W")
	home := pushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	declareAttack(t, g, opp.ID, historian)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, historian), "double strike") {
		t.Fatal("the Historian attacking gives itself double strike")
	}
	if hasAbility(effectiveAbilities(t, g, home), "double strike") {
		t.Error("a creature that did not attack is not an attacking creature")
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 2/3 with double strike deals 4: %d → %d", life, opp.Life)
	}
	advanceToNextSeatsTurn(t, g)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if hasAbility(effectiveAbilities(t, g, theirs), "double strike") {
		t.Error("an opponent's attacker is not yours")
	}
	if spec, _ := Lookup(b25BladeHistorianOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the attack-declaration gap is declared")
	}
}

func TestB25VolatileFaultDestroysANonbasicAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fault := pushCatalogPermanent(g, me.ID, "Volatile Fault", "Land — Cave", b25VolatileFaultOracle, false)
	cradle := seedLandOnBattlefield(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	forest := seedLandOnBattlefield(g, opp.ID, "Forest", "Basic Land — Forest")
	swamp := stapleLibraryCard(opp, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, fault, 0, game.ActivateAbilityParams{Targets: b16TargetCard(forest)}); err == nil {
		t.Error("a basic land is refused")
	}
	if err := g.ActivateCatalogAbility(me.ID, fault, 0, game.ActivateAbilityParams{Targets: b16TargetCard(cradle)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fault) {
		t.Error("the Fault is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cradle) {
		t.Fatal("the nonbasic land is destroyed")
	}
	if searchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the victim gets a (declinable) search for a basic")
	}
	if b16CountNamed(g, "Treasure") != 0 {
		t.Error("the Treasure waits for the victim's answer")
	}
	answerSearchByID(t, g, opp.ID, swamp)
	if !g.Battlefield.Contains(swamp) {
		t.Error("the victim's basic enters")
	}
	if b16CountNamed(g, "Treasure") != 1 {
		t.Error("you create a Treasure")
	}
	if tr := findBattlefieldByName(g, "Treasure"); tr == uuid.Nil || cardByID(g, tr).Controller != me.ID {
		t.Error("the Treasure is yours")
	}
}

func TestB25PashalikMonsPingsForGoblinDeathsAndMakesGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mons := b21Push(g, me.ID, "Pashalik Mons", "Legendary Creature — Goblin Warrior", b25PashalikMonsOracle, 2, 2, "R")
	goblin := b16Creature(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", 2, 2, "R")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", 1, 1, "R")
	advanceToMain(t, g)
	life := opp.Life

	b25Destroy(g, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Goblin dying is nothing")
	}
	b25Destroy(g, theirs)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's Goblin dying is nothing")
	}
	b25Destroy(g, goblin)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("another Goblin you control dying: 1 damage to the chosen target (%d → %d)", life, opp.Life)
	}

	// The activation sacrifices a Goblin — Mons himself is one — and
	// his own death pings too.
	b06AddMana(me, "R", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, mons, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{mons}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("Mons dying to his own ability pings (%d → %d)", life, opp.Life)
	}
	if n := b16CountNamed(g, "Goblin"); n != 2 {
		t.Errorf("two 1/1 red Goblins: %d", n)
	}
	if spec, _ := Lookup(b25PashalikMonsOracle); spec.Completeness != CompletenessFull {
		t.Error("Mons is whole")
	}
}

func TestB25TevalMillsReturnsALandAndMakesZombiesWhenCardsLeave(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	teval := b21Push(g, me.ID, "Teval, the Balanced Scale", "Legendary Creature — Spirit Dragon", b25TevalOracle, 4, 4, "B", "G", "U")
	if !hasEffectiveKeyword(t, g, teval, "flying") {
		t.Error("flying")
	}
	forest := b17GraveyardCard(me, "Forest", "Basic Land — Forest", "")
	library := me.Library.Size()
	grave := me.Graveyard.Size()
	declareAttack(t, g, opp.ID, teval)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, forest) || p.PickTargetMin != 0 {
		t.Error("up to one land card in your graveyard")
	}
	pickCard(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-3 {
		t.Errorf("mill three: library %d → %d", library, me.Library.Size())
	}
	if !g.Battlefield.Contains(forest) || !b25Tapped(t, g, forest) {
		t.Error("the chosen land returns tapped")
	}
	if me.Graveyard.Size() != grave+3-1 {
		t.Errorf("graveyard %d → %d: three milled, one returned", grave, me.Graveyard.Size())
	}
	if n := b16CountNamed(g, "Zombie Druid"); n != 1 {
		t.Errorf("the land leaving the graveyard makes a Zombie Druid: %d", n)
	}
	// Two cards leaving at once is one trigger.
	var milled []uuid.UUID
	for _, c := range me.Graveyard.Cards {
		milled = append(milled, c.InstanceID)
	}
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(milled) })
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Zombie Druid"); n != 2 {
		t.Errorf("one or more cards leaving is one Zombie: %d", n)
	}
	// With no land card in the graveyard the attack still mills.
	advanceToMainOf(t, g, 0)
	library = me.Library.Size()
	declareAttack(t, g, opp.ID, teval)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("nothing to return: no prompt")
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != library-3 {
		t.Error("the mill does not depend on a land being in the graveyard")
	}
	if spec, _ := Lookup(b25TevalOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the pick-before-mill gap is declared")
	}
}

func TestB25VoraciousHydraFightsOrDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	hydra := castXSpell(t, g, "Voracious Hydra", "Creature — Hydra", b25VoraciousHydraOracle, "{X}{G}{G}", 3, nil)
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra, "+1/+1") != 3 {
		t.Fatalf("enters with X counters: %d", counterCount(g, hydra, "+1/+1"))
	}
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, theirs) || hasID(p.PickTargetCards, mine) || p.PickTargetMin != 0 {
		t.Error("up to one creature you don't control")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3/4 Hydra fights and kills the 2/2")
	}
	if c, _ := battlefieldCard(g, hydra); c.DamageMarked != 2 {
		t.Errorf("the Bear hits back for 2: %d", c.DamageMarked)
	}
	if counterCount(g, hydra, "+1/+1") != 3 {
		t.Error("the fight mode does not double")
	}
	// Pick no target: the counters double instead.
	theirs2 := pushVanillaCreature(g, opp.ID, "Their Other Bear", 2, 2)
	hydra2 := castXSpell(t, g, "Voracious Hydra", "Creature — Hydra", b25VoraciousHydraOracle, "{X}{G}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	p = latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("picking no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, hydra2, "+1/+1") != 4 {
		t.Errorf("no target: 2 counters doubled to 4, got %d", counterCount(g, hydra2, "+1/+1"))
	}
	if !g.Battlefield.Contains(theirs2) {
		t.Error("no fight")
	}
	// No creature to fight at all: no prompt, the counters double.
	b25Destroy(g, theirs2)
	passPriorityAroundTable(t, g)
	hydra3 := castXSpell(t, g, "Voracious Hydra", "Creature — Hydra", b25VoraciousHydraOracle, "{X}{G}{G}", 1, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("nothing to fight: no prompt")
	}
	if counterCount(g, hydra3, "+1/+1") != 2 {
		t.Errorf("1 counter doubled to 2, got %d", counterCount(g, hydra3, "+1/+1"))
	}
	if !hasEffectiveKeyword(t, g, hydra3, "trample") {
		t.Error("trample")
	}
	if spec, _ := Lookup(b25VoraciousHydraOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the X-on-resolve and mode-by-target gaps are declared")
	}
}
