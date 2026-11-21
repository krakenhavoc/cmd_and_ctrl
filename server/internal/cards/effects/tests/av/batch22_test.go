package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b22Lives snapshots every seat's life total.
func b22Lives(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// b22HandHasNamed reports whether a player holds a card by name.
func b22HandHasNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Whelming Wave
// was already on main from the S23 boardwipes (#382) and is pinned
// here so the table matches the issue's 34.
func TestBatch22CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B22WhelmingWaveAlreadyOID:   "Whelming Wave",
		effectstest.B22ConclaveMentorOracle:     "Conclave Mentor",
		effectstest.B22RograkhOracle:            "Rograkh, Son of Rohgahh",
		effectstest.B22ShivanGorgeOracle:        "Shivan Gorge",
		effectstest.B22DictateOfTheTwinGodsOID:  "Dictate of the Twin Gods",
		effectstest.B22SlaughterTheStrongOracle: "Slaughter the Strong",
		effectstest.B22TarnishedCitadelOracle:   "Tarnished Citadel",
		effectstest.B22SamwiseGamgeeOracle:      "Samwise Gamgee",
		effectstest.B22DisdainfulStrokeOracle:   "Disdainful Stroke",
		effectstest.B22ToxrillOracle:            "Toxrill, the Corrosive",
		effectstest.B22CabbageMerchantOracle:    "The Cabbage Merchant",
		effectstest.B22CauldronOfEssenceOracle:  "Cauldron of Essence",
		effectstest.B22EarthshakerDreadmawOID:   "Earthshaker Dreadmaw",
		effectstest.B22BerserkersOnslaughtOID:   "Berserkers' Onslaught",
		effectstest.B22FuryOracle:               "Fury",
		effectstest.B22WirewoodLodgeOracle:      "Wirewood Lodge",
		effectstest.B22WavebreakHippocampOracle: "Wavebreak Hippocamp",
		effectstest.B22SingularityRuptureOracle: "Singularity Rupture",
		effectstest.B22CodexShredderOracle:      "Codex Shredder",
		effectstest.B22WearDownOracle:           "Wear Down",
		effectstest.B22HullBreachOracle:         "Hull Breach",
		effectstest.B22LilianasCaressOracle:     "Liliana's Caress",
		effectstest.B22SabotenderOracle:         "Sabotender",
		effectstest.B22DawnOfHopeOracle:         "Dawn of Hope",
		effectstest.B22MycosynthWellspringOID:   "Mycosynth Wellspring",
		effectstest.B22EreborFlamesmithOracle:   "Erebor Flamesmith",
		effectstest.B22IorethOracle:             "Ioreth of the Healing House",
		effectstest.B22AethericAmplifierOracle:  "Aetheric Amplifier",
		effectstest.B22GodEternalOketraOracle:   "God-Eternal Oketra",
		effectstest.B22HarshMentorOracle:        "Harsh Mentor",
		// #1213 built the return-an-artifact cost component, so the
		// batch's fifth declared skip is a registered card now.
		"da46786e-28df-4638-ab3d-121011d2f150": "Master Transmuter",
	}
	if len(want) != 31 {
		t.Fatalf("the batch registers 30 cards plus Whelming Wave, the table lists %d", len(want))
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
	// The remaining declared skip must NOT be registered — it has a
	// cost the engine cannot express, and a spec would ship the card
	// stronger than printed.
	//
	// Two cards came OFF this list and are in `want` above: Harsh
	// Mentor with #1210's opponent-activation watch, and Master
	// Transmuter with #1213's return-an-artifact cost component. A
	// third, Odric, Master Tactician, came off with #4154's
	// declaration control and is pinned in
	// declaration_control_cards_test.go. Steelbane Hydra came off with
	// the remove-a-counter cost (#5011, tmc_b4_cards_test.go).
	for _, skipped := range []string{
		"119d719d-e965-45b4-9bc9-ac03211b10c2", // Survival of the Fittest — discard-a-card cost
	} {
		if _, ok := effects.Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestB22ShivanGorgeTapsForColorlessOrPingsTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gorge := effectstest.B12Push(g, me.ID, "Shivan Gorge", "Legendary Land", effectstest.B22ShivanGorgeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, gorge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// The two abilities share the tap.
	me.ManaPool.EmptyPool()
	effectstest.B06AddMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gorge, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a tapped Gorge cannot pay {T} again")
	}
	effectstest.B22Untap(g, gorge)
	before := b22Lives(g)
	effectstest.B16Activate(t, g, me.ID, gorge, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	if !effectstest.B20Tapped(t, g, gorge) {
		t.Error("the ping has a tap cost")
	}
}

func TestB22TarnishedCitadelColoredHalfDealsThreeColorlessHalfIsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	citadel := effectstest.SeedPermanentWithOracle(g, me.ID, "Tarnished Citadel", "Land", effectstest.B22TarnishedCitadelOracle)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, citadel, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if me.Life != life {
		t.Errorf("the {C} half hurt: %d → %d", life, me.Life)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	effectstest.B22Untap(g, citadel)
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, citadel, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, me.Life)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the coloured half must offer all five colours, got %+v", pick)
	}
}

func TestB22WirewoodLodgeUntapsAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lodge := effectstest.B12Push(g, me.ID, "Wirewood Lodge", "Land", effectstest.B22WirewoodLodgeOracle, 0, 0)
	elf := effectstest.B12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b08Tap(g, elf)
	b08Tap(g, bear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, lodge, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a Bear is not an Elf")
	}
	effectstest.B16Activate(t, g, me.ID, lodge, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(elf)})
	if effectstest.B20Tapped(t, g, elf) {
		t.Error("the Elf — anyone's — should be untapped")
	}
	if !effectstest.B20Tapped(t, g, lodge) {
		t.Error("the untap has a tap cost")
	}
}

// --- the counters family -------------------------------------------

func TestB22ConclaveMentorAddsACounterAndGainsItsPowerOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mentor := effectstest.B12Push(g, me.ID, "Conclave Mentor", "Creature — Centaur Cleric", effectstest.B22ConclaveMentorOracle, 2, 2)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)

	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", 1)
		_ = g.AddCounterForEffect(mentor, "+1/+1", 2)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
		_ = g.AddCounterForEffect(bear, "charge", 1)
	})
	if got := effectstest.CounterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("one +1/+1 counter on my creature should land as two, got %d", got)
	}
	if got := effectstest.CounterCount(g, mentor, "+1/+1"); got != 3 {
		t.Errorf("the Mentor grows itself: two should land as three, got %d", got)
	}
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature is not 'a creature you control', got %d", got)
	}
	if got := effectstest.CounterCount(g, bear, "charge"); got != 1 {
		t.Errorf("only +1/+1 counters are replaced, charge landed as %d", got)
	}

	// Dies: life equal to its power, counters included (2 + 3).
	life := me.Life
	effectstest.B18Kill(t, g, mentor)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (a 2/2 with three +1/+1 counters)", life, me.Life)
	}
}

func TestB22ToxrillSlimesShrinksAndMakesSlugs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toxrill := effectstest.B12Push(g, me.ID, "Toxrill, the Corrosive", "Legendary Creature — Slug Horror", effectstest.B22ToxrillOracle, 7, 7)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)

	// EACH end step — the controller's own first — slimes every
	// creature the controller does not control.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, theirs, "slime") != 1 || effectstest.CounterCount(g, wall, "slime") != 1 {
		t.Fatal("each creature you don't control gets a slime counter at the end step")
	}
	if effectstest.CounterCount(g, mine, "slime") != 0 || effectstest.CounterCount(g, toxrill, "slime") != 0 {
		t.Error("your own creatures are not slimed")
	}
	if effectstest.EffectivePower(t, g, wall) != -1 || effectstest.EffectiveToughness(t, g, wall) != 3 {
		t.Errorf("a slimed 0/4 should read -1/3, got %d/%d", effectstest.EffectivePower(t, g, wall), effectstest.EffectiveToughness(t, g, wall))
	}
	if effectstest.EffectiveToughness(t, g, mine) != 2 {
		t.Error("the -1/-1 is only for creatures you don't control")
	}

	// An opponent's end step counts too: the 2/2 reaches 0/0 and dies
	// with a slime counter on it — a Slug for Toxrill's controller.
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("two slime counters kill a 2/2")
	}
	if effectstest.CountBattlefieldNamed(g, me.ID, "Slug") != 1 {
		t.Errorf("a slimed creature dying makes a Slug, got %d", effectstest.CountBattlefieldNamed(g, me.ID, "Slug"))
	}
	if c, _ := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Slug")); c.Power != 1 || c.Toughness != 1 || !c.HasColor("B") {
		t.Error("a 1/1 black Slug")
	}

	// An unslimed opponent's creature dying is silent.
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	effectstest.B18Kill(t, g, fresh)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Slug") != 1 {
		t.Error("a creature without a slime counter makes no Slug")
	}

	// {U}{B}, Sacrifice a Slug: draw a card.
	slug := effectstest.FindBattlefieldByName(g, "Slug")
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "U", "B")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{slug}})
	if g.Battlefield.Contains(slug) {
		t.Error("the Slug is sacrificed as the cost")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// A Bear is not a Slug; Toxrill itself is.
	effectstest.B06AddMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{mine}}); err == nil {
		t.Error("a Bear cannot pay 'sacrifice a Slug'")
	}
	effectstest.B16Activate(t, g, me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{toxrill}})
	if g.Battlefield.Contains(toxrill) {
		t.Error("Toxrill is a Slug and may eat itself")
	}
}

func TestB22AethericAmplifierTapsForAnyColourAndDoublesCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amp := effectstest.B12Push(g, me.ID, "Aetheric Amplifier", "Artifact", effectstest.B22AethericAmplifierOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, amp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, the printed width, got %+v", pick)
	}
	// #730: an unanswered colour pick gates the cursor, and this test
	// walks the turn on below. The colour itself is beside the point.
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	effectstest.B22Untap(g, amp)
	theirs := effectstest.PushCounterCreature(g, opp.ID, "Their Hydra", "+1/+1", 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, "charge", 1) })
	// Sorcery speed: not during combat.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, amp, 0, game.ActivateAbilityParams{
		Modes: []int{0}, Targets: effectstest.B16TargetCard(theirs),
	}); err == nil {
		t.Fatal("activate only as a sorcery")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	// #764: a modal activated ability announces its modes WITH its
	// targets, in one indivisible step (CR 602.2b). The first bullet
	// targets; naming it without a target, or naming the untargeted
	// second bullet with one, is refused at announce.
	if err := g.ActivateCatalogAbility(me.ID, amp, 0, game.ActivateAbilityParams{Modes: []int{0}}); err == nil {
		t.Error("the first bullet targets — an announcement with no target is illegal")
	}
	if err := g.ActivateCatalogAbility(me.ID, amp, 0, game.ActivateAbilityParams{
		Modes: []int{1}, Targets: effectstest.B16TargetCard(theirs),
	}); err == nil {
		t.Error("the second bullet targets nothing — a target is illegal")
	}
	if err := g.ActivateCatalogAbility(me.ID, amp, 0, game.ActivateAbilityParams{
		Modes: []int{0, 1}, Targets: effectstest.B16TargetCard(theirs),
	}); err == nil {
		t.Error("choose ONE — two bullets is illegal")
	}
	effectstest.B16Activate(t, g, me.ID, amp, 0, game.ActivateAbilityParams{Modes: []int{0}, Targets: effectstest.B16TargetCard(theirs)})
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 6 {
		t.Errorf("+1/+1 counters 3 → %d, want 6", got)
	}
	if got := effectstest.CounterCount(g, theirs, "charge"); got != 2 {
		t.Errorf("each KIND doubles: charge 1 → %d, want 2", got)
	}
	if !effectstest.B20Tapped(t, g, amp) {
		t.Error("the activation has a tap cost")
	}

	// The SECOND bullet — the one the card shipped without until
	// #764: the activator's own player counters, every kind, no
	// target at all.
	effectstest.B22Untap(g, amp)
	g.WithWriteLock(func() {
		_ = g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 3)
		_ = g.AddPlayerCounterForEffect(me.ID, game.CounterPoison, 2)
	})
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, amp, 0, game.ActivateAbilityParams{Modes: []int{1}})
	if got := me.Counters[game.CounterEnergy]; got != 6 {
		t.Errorf("energy 3 → %d, want 6", got)
	}
	if got := me.Counters[game.CounterPoison]; got != 4 {
		t.Errorf("each KIND doubles: poison 2 → %d, want 4", got)
	}
	if got := effectstest.CounterCount(g, theirs, "+1/+1"); got != 6 {
		t.Error("the second bullet touches no permanent")
	}
	if spec, _ := effects.Lookup(effectstest.B22AethericAmplifierOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("both bullets ship in #764 — the caveat is gone")
	}
}

// --- the damage family ---------------------------------------------

func TestB22DictateOfTheTwinGodsDoublesEveryonesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dictate := effectstest.CastCatalogSpell(t, g, "Dictate of the Twin Gods", "Enchantment", effectstest.B22DictateOfTheTwinGodsOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dictate) {
		t.Fatal("the Dictate did not resolve")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, dictate), "flash") {
		t.Error("printed flash did not reach the effective abilities")
	}
	// The controller's Bolt at an opponent: 6.
	life := opp.Life
	b07BoltPlayer(t, g, opp.ID)
	if opp.Life != life-6 {
		t.Errorf("a Bolt under the Dictate: %d → %d, want -6", life, opp.Life)
	}
	// An OPPONENT's Bolt at the controller — every source, as printed.
	mine := me.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-6 {
		t.Errorf("an opponent's Bolt under the Dictate: %d → %d, want -6", mine, me.Life)
	}
}

func TestB22ShivanGorgeDamageIsDoubledByTheDictate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dictate of the Twin Gods", "Enchantment", effectstest.B22DictateOfTheTwinGodsOID, 0, 0)
	gorge := effectstest.B12Push(g, me.ID, "Shivan Gorge", "Legendary Land", effectstest.B22ShivanGorgeOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "C", "C")
	before := b22Lives(g)
	effectstest.B16Activate(t, g, me.ID, gorge, 0, game.ActivateAbilityParams{})
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (the Gorge is the source)", i, before[i], got)
		}
	}
}

func TestB22SabotenderPingsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sabotender := effectstest.B12Push(g, me.ID, "Sabotender", "Creature — Plant", effectstest.B22SabotenderOracle, 2, 1)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, sabotender), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}
	before := b22Lives(g)
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d: %d → %d, want -1", i, before[i], got)
		}
	}
	if me.Life != before[0] {
		t.Error("the controller is not an opponent")
	}
	// An opponent's land is not "a land YOU control".
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Forest", TypeLine: "Token Land — Forest"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.Seats[2].Life; got != before[2]-1 {
		t.Error("an opponent's land triggered my landfall")
	}
}

func TestB22EreborFlamesmithPingsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Erebor Flamesmith", "Creature — Dwarf Artificer", effectstest.B22EreborFlamesmithOracle, 2, 1)
	before := b22Lives(g)
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d after an instant: %d → %d, want -1", i, before[i], got)
		}
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d after a creature: %d → %d, want unchanged", i, before[i]-1, got)
		}
	}
}

func TestB22FuryDividesFourDamageAsChosenAndEvokesForARedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "Bear C", "Creature — Bear", 3, 3)

	fury := effectstest.CastCatalogSpell(t, g, "Fury", "Creature — Elemental Incarnation", effectstest.B22FuryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, fury), "double strike") {
		t.Error("printed double strike did not reach the effective abilities")
	}
	// #1563: the caster divides — 1 / 2 / 1, a split the old even
	// division (2 / 1 / 1, remainder to the first pick) never offered.
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2, c: 1}, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	if got := [3]int{effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b), effectstest.DamageMarkedOn(g, c)}; got != [3]int{1, 2, 1} {
		t.Errorf("the announced division 1/2/1 was not dealt: got %v", got)
	}

	// Evoke: pitch a red card, the ETB still fires, then Fury dies.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.ToMainForCost(t, g2)
	pitch := effectstest.HandCardFull(me2, "Lightning Bolt", "Instant", "{R}", "", []string{"R"})
	blue := effectstest.HandCardFull(me2, "Counterspell", "Instant", "{U}{U}", "", []string{"U"})
	victim := effectstest.B12Creature(g2, opp2.ID, "Victim", "Creature — Bear", 4, 4)
	fury2 := effectstest.HandCardFull(me2, "Fury", "Creature — Elemental Incarnation", "{3}{R}{R}", effectstest.B22FuryOracle, []string{"R"})
	if err := g2.CastSpell(me2.ID, fury2, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{blue},
	}); err == nil {
		t.Fatal("a blue card cannot pay 'exile a red card from your hand'")
	}
	if err := g2.CastSpell(me2.ID, fury2, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", AltCostIDs: []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("evoking Fury: %v", err)
	}
	if !g2.Exile.Contains(pitch) {
		t.Error("the evoke pitch was not exiled")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.B17PickCards(t, g2, me2.ID, victim)
	// #1529: the damage trigger and the evoke sacrifice are one CR
	// 603.3b batch; the damage resolves first.
	if !effectstest.AnswerTriggerOrderLastQueuedFirst(t, g2) {
		t.Error("the evoke sacrifice and the damage trigger were not offered for ordering")
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(victim) {
		t.Error("one target takes all 4")
	}
	if g2.Battlefield.Contains(fury2) || !me2.Graveyard.Contains(fury2) {
		t.Error("an evoked Fury is sacrificed after its trigger resolves")
	}
}

func TestB22BerserkersOnslaughtGivesAttackersDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Berserkers' Onslaught", "Enchantment", effectstest.B22BerserkersOnslaughtOID, 0, 0)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	home := effectstest.PushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, attacker), "double strike") {
		t.Fatal("the attacking creature should have double strike")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, home), "double strike") {
		t.Error("a creature that did not attack is not an attacking creature")
	}
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 2/2 with double strike deals 4: %d → %d", life, opp.Life)
	}

	// An opponent's attacker on their turn is not "you control".
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.DeclareAttack(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "double strike") {
		t.Error("an opponent's attacker is not yours")
	}
}

// --- the counterspell ----------------------------------------------

func TestB22DisdainfulStrokeCountersOnlyBigSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Disdainful Stroke", TypeLine: "Instant",
		OracleID: effectstest.B22DisdainfulStrokeOracle, Owner: me.ID, Controller: me.ID})
	// A spell with no printed cost is mana value 0.
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{Targets: effectstest.B16TargetCard(small)}); err == nil {
		t.Fatal("a mana value 0 spell is not 'mana value 4 or greater'")
	}
	effectstest.PassPriorityAroundTable(t, g)

	big := effectstest.CastWithCost(t, g, "Wurm", "Creature — Wurm", "{4}{G}{G}", "")
	stroke := effectstest.CastCatalogSpell(t, g, "Disdainful Stroke", "Instant", effectstest.B22DisdainfulStrokeOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !me.Graveyard.Contains(big) {
		t.Error("the six-drop should be countered")
	}
	if !me.Graveyard.Contains(stroke) {
		t.Error("the Stroke goes to the graveyard")
	}

	// X counts on the stack (CR 202.3e): a Hydra cast for X=3 is 5.
	hydra := effectstest.CastXSpell(t, g, "Hydra", "Creature — Hydra", "", "{X}{G}{G}", 3, nil)
	effectstest.CastCatalogSpell(t, g, "Disdainful Stroke", "Instant", effectstest.B22DisdainfulStrokeOracle, effectstest.B16TargetCard(hydra))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hydra) || !me.Graveyard.Contains(hydra) {
		t.Error("X is counted on the stack — the Hydra cast for X=3 has mana value 5")
	}
}

// --- the go-wide payoffs -------------------------------------------

func TestB22SamwiseGamgeeMakesFoodForNontokenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Samwise Gamgee", "Legendary Creature — Halfling Peasant", effectstest.B22SamwiseGamgeeOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatalf("a nontoken creature entering makes a Food, got %d", effectstest.CountBattlefieldNamed(g, me.ID, "Food"))
	}
	// A token, an opponent's creature and a noncreature are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Errorf("tokens, opponents' creatures and artifacts do not make Food, got %d", effectstest.CountBattlefieldNamed(g, me.ID, "Food"))
	}
	// #747: the three-Food ability ships at its printed count; its
	// engine test is in sacrifice_n_cards_test.go.
	spec, _ := effects.Lookup(effectstest.B22SamwiseGamgeeOracle)
	if len(spec.Activated) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the three-Food ability ships whole")
	}
}

func TestB22CabbageMerchantMakesFoodAndLosesItToCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "The Cabbage Merchant", "Legendary Creature — Human Citizen", effectstest.B22CabbageMerchantOracle, 2, 2)
	// My own noncreature spell is silent; an opponent's makes a Food.
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Fatal("your own spell is not an opponent's")
	}
	effectstest.B13OpponentCasts(t, g, opp, "Shock", "Instant", "", "{R}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatalf("an opponent's noncreature spell makes a Food, got %d", effectstest.CountBattlefieldNamed(g, me.ID, "Food"))
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.B20CastCreature(t, g, opp, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Error("an opponent's CREATURE spell is silent")
	}
	// A creature dealing combat damage to me: sacrifice a Food token.
	food := effectstest.FindBattlefieldByName(g, "Food")
	cabin := effectstest.B12Push(g, me.ID, "Gingerbread Cabin", "Land — Forest Food", "", 0, 0)
	attacker := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, attacker, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("combat damage to the Merchant's controller asks for a Food token")
	}
	if len(prompt.SacrificeOptions) != 1 || prompt.SacrificeOptions[0] != food {
		t.Errorf("only Food TOKENS are offered (not the Cabin), got %v", prompt.SacrificeOptions)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, food)
	if g.Battlefield.Contains(food) || !g.Battlefield.Contains(cabin) {
		t.Error("the Food token is sacrificed, the Cabin stays")
	}
	// Non-combat damage is silent.
	b07BoltPlayer(t, g, me.ID)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a Bolt is not combat damage")
	}
	spec, _ := effects.Lookup(effectstest.B22CabbageMerchantOracle)
	if len(spec.ManaAbilities) != 1 || spec.Completeness != effects.CompletenessFull {
		t.Error("the tap-two-Foods mana ability ships and the card is complete (#3172)")
	}
}

func TestB22CauldronOfEssenceDrainsOnDeathAndReanimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := effectstest.B12Push(g, me.ID, "Cauldron of Essence", "Artifact", effectstest.B22CauldronOfEssenceOracle, 0, 0)
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	before := b22Lives(g)
	effectstest.B18Kill(t, g, theirs)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("an opponent's creature dying is silent, seat %d moved", i)
		}
	}
	effectstest.B18Kill(t, g, mine)
	if me.Life != before[0]+1 {
		t.Errorf("you gain 1: %d → %d", before[0], me.Life)
	}
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d loses 1: %d → %d", i, before[i], got)
		}
	}

	// The reanimation: the sacrifice is the cost, its drain lands
	// above the ability, and the target comes back to the battlefield.
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.B06AddMana(me, "B", "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(mine),
	}); err == nil {
		t.Fatal("activate only as a sorcery")
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.B06AddMana(me, "B", "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(theirs),
	}); err == nil {
		t.Fatal("an opponent's graveyard is not yours")
	}
	mid := b22Lives(g)
	effectstest.B16Activate(t, g, me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.B16TargetCard(mine),
	})
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature is sacrificed as the cost")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature card comes back to the battlefield")
	}
	if me.Life != mid[0]+1 || g.Seats[1].Life != mid[1]-1 {
		t.Error("the sacrificed creature's death drains the table")
	}
	if !effectstest.B20Tapped(t, g, cauldron) {
		t.Error("the reanimation has a tap cost")
	}
}

func TestB22EarthshakerDreadmawDrawsPerOtherDinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 3, 3)
	effectstest.B12Creature(g, me.ID, "Ripjaw", "Creature — Dinosaur", 4, 5)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 3, 3)
	hand := me.Hand.Size()
	maw := effectstest.CastCatalogSpell(t, g, "Earthshaker Dreadmaw", "Creature — Dinosaur", effectstest.B22EarthshakerDreadmawOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, maw), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("drew %d, want 2 (two OTHER Dinosaurs you control)", got-hand)
	}
}

func TestB22GodEternalOketraMakesZombiesAndReturnsThirdFromTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oketra := effectstest.B12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", effectstest.B22GodEternalOketraOracle, 3, 6)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, oketra), "double strike") {
		t.Error("printed double strike did not reach the effective abilities")
	}
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Warrior") != 1 {
		t.Fatalf("a creature spell makes a Zombie Warrior, got %d", effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Warrior"))
	}
	zombie := effectstest.FindBattlefieldByName(g, "Zombie Warrior")
	if c, _ := effectstest.BattlefieldCard(g, zombie); c.Power != 4 || c.Toughness != 4 || !c.HasColor("B") {
		t.Error("a 4/4 black Zombie Warrior")
	}
	effectstest.AssertKeywords(t, g, zombie, "vigilance")
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountBattlefieldNamed(g, me.ID, "Zombie Warrior") != 1 {
		t.Error("a noncreature spell is silent")
	}

	// Dies: may go third from the top of the library.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(oketra) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(oketra) {
		t.Fatal("Oketra should have left the graveyard")
	}
	names := effectstest.LibraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Oketra" || names[0] == "God-Eternal Oketra" {
		t.Errorf("third from the top, got %v", names)
	}

	// Exiled from the battlefield: the same offer, declined.
	again := effectstest.B12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", effectstest.B22GodEternalOketraOracle, 3, 6)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(again) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(again) {
		t.Error("declining leaves Oketra in exile")
	}
	// Bounced: neither dies nor exiled, no offer.
	third := effectstest.B12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", effectstest.B22GodEternalOketraOracle, 3, 6)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(third) })
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a bounce is not 'dies or is put into exile'")
	}
}

// --- the enchantments ----------------------------------------------

func TestB22LilianasCaressDrainsTwoPerOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Liliana's Caress", "Enchantment", effectstest.B22LilianasCaressOracle, 0, 0)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, opp)
	effectstest.PushHandCard(g, me)
	life, mine := opp.Life, me.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("two discards: %d → %d, want -4", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Error("the controller's own discard is silent")
	}
}

func TestB22DawnOfHopeOffersToPayOnLifegainAndMakesLifelinkSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dawn := effectstest.B12Push(g, me.ID, "Dawn of Hope", "Enchantment", effectstest.B22DawnOfHopeOracle, 0, 0)
	hand := me.Hand.Size()
	effectstest.B14Gain(t, g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("gaining life asks the controller to pay {2}")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("paying draws one, drew %d", got-hand)
	}
	effectstest.B14Gain(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if got := me.Hand.Size(); got != hand+1 {
		t.Error("declining draws nothing")
	}
	effectstest.B14Gain(t, g, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's lifegain is not yours")
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "W", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, dawn, 0, game.ActivateAbilityParams{})
	soldier := effectstest.FindBattlefieldByName(g, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier")
	}
	if c, _ := effectstest.BattlefieldCard(g, soldier); c.Power != 1 || c.Toughness != 1 || !c.HasColor("W") {
		t.Error("a 1/1 white Soldier")
	}
	effectstest.AssertKeywords(t, g, soldier, "lifelink")
}

func TestB22WavebreakHippocampDrawsOnTheFirstSpellOfEachOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Wavebreak Hippocamp", "Enchantment Creature — Horse Fish", effectstest.B22WavebreakHippocampOracle, 2, 2)
	// On my own turn: silent.
	// Every cast helper puts the card in hand first, so a cast with
	// no draw leaves the hand size where it was.
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a spell on your own turn draws nothing, hand %d → %d", hand, got)
	}
	// On an opponent's turn: the first draws, the second does not.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand = me.Hand.Size()
	effectstest.B22CastInstantAs(t, g, me, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the first spell on an opponent's turn draws — hand %d → %d", hand, got)
	}
	effectstest.B22CastInstantAs(t, g, me, "Brainstorm")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the second spell that turn is silent — hand %d → %d", hand, got)
	}
	// The next opponent's turn resets the tally.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand = me.Hand.Size()
	effectstest.B22CastInstantAs(t, g, me, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("each opponent's turn has its own first spell — hand %d → %d", hand, got)
	}
}

// --- the sorceries -------------------------------------------------

func TestB22SlaughterTheStrongKeepsLowestPowerWithinFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := effectstest.B12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	myTwo := effectstest.B12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	myOne := effectstest.B12Creature(g, me.ID, "My One", "Creature — Elf", 1, 1)
	myZero := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	myOther := effectstest.B12Creature(g, me.ID, "My Other Two", "Creature — Bear", 2, 2)
	theirFour := effectstest.B12Creature(g, opp.ID, "Their Four", "Creature — Beast", 4, 4)
	theirOne := effectstest.B12Creature(g, opp.ID, "Their One", "Creature — Elf", 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Slaughter the Strong", "Sorcery", effectstest.B22SlaughterTheStrongOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Each player chooses; the active player's prompt comes first and
	// the set rule is total power 4 or less.
	mine := effectstest.ChooseCardsChoiceFor(g, me.ID)
	theirs := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if mine == nil || theirs == nil || mine.ChooseMin != 0 {
		t.Fatalf("every player with creatures is asked, for any number: %+v %+v", mine, theirs)
	}
	if err := g.ResolveChooseCards(mine.ID, me.ID, []uuid.UUID{myBig, myTwo}); err == nil {
		t.Fatal("a set with total power 7 was accepted")
	}
	// Me: 0 + 1 + 2 = 3 fits.
	effectstest.AnswerChooseCards(t, g, me.ID, myZero, myOne, myTwo)
	// Them: one 4-power creature over the 1/1: the player's choice, which
	// the old lowest-power auto-pick could not make.
	effectstest.AnswerChooseCards(t, g, opp.ID, theirFour)
	effectstest.PassPriorityAroundTable(t, g)
	for _, keep := range []uuid.UUID{myZero, myOne, myTwo} {
		if !g.Battlefield.Contains(keep) {
			t.Errorf("%s should be kept", keep)
		}
	}
	for _, gone := range []uuid.UUID{myBig, myOther} {
		if g.Battlefield.Contains(gone) {
			t.Errorf("%s should be sacrificed", gone)
		}
	}
	if !g.Battlefield.Contains(theirFour) || g.Battlefield.Contains(theirOne) {
		t.Error("each player keeps the set they chose")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("only creatures are sacrificed")
	}
	if !me.Graveyard.Contains(myBig) || !opp.Graveyard.Contains(theirOne) {
		t.Error("a sacrifice goes to its owner's graveyard")
	}
}

func TestB22SingularityRuptureWipesThenMillsHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	theirs := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	opp.Library.Cards, other.Library.Cards = nil, nil
	effectstest.SeedLibrary(opp, "a", "b", "c", "d", "e", "f", "g")
	effectstest.SeedLibrary(other, "a", "b", "c", "d")
	myLibrary := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Singularity Rupture", "Sorcery", effectstest.B22SingularityRuptureOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("destroy all creatures")
	}
	if got := opp.Library.Size(); got != 4 {
		t.Errorf("seven cards, half rounded down is three: library %d, want 4", got)
	}
	if got := other.Library.Size(); got != 2 {
		t.Errorf("four cards: library %d, want 2", got)
	}
	if got := me.Library.Size(); got != myLibrary {
		t.Errorf("an untargeted player is not milled: library %d → %d", myLibrary, got)
	}
	// Any number includes none.
	effectstest.CastCatalogSpell(t, g, "Singularity Rupture", "Sorcery", effectstest.B22SingularityRuptureOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != 4 {
		t.Error("with no targets nobody mills")
	}
}

func TestB22WearDownDestroysOneArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.SeedCreature(g, "Bear", opp.ID)
	if b09TryCast(t, g, "Wear Down", "Sorcery", effectstest.B22WearDownOracle, effectstest.B16TargetCard(bear)) == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	effectstest.CastCatalogSpell(t, g, "Wear Down", "Sorcery", effectstest.B22WearDownOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	spec, _ := effects.Lookup(effectstest.B22WearDownOracle)
	// ADR 0089: unpromised, one target; the gift's clause is the
	// two-target one (TestGiftWearDownPromisedDestroysTwo).
	if spec.Targets == nil || spec.Targets.Max != 1 || spec.Gift == nil || spec.Gift.Targets.Max != 2 {
		t.Error("one target without the gift, two with it")
	}
}

func TestB22HullBreachThreeModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	rock2 := effectstest.B12Permanent(g, opp.ID, "Rock 2", "Artifact")
	aura := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	aura2 := effectstest.B12Permanent(g, opp.ID, "Aura 2", "Enchantment")

	// Mode 1 refuses an enchantment; mode 2 refuses an artifact.
	if err := effectstest.B22TryModal(g, effectstest.B22HullBreachOracle, []int{0}, effectstest.B16TargetCard(aura)); err == nil {
		t.Error("mode 1 is 'target artifact'")
	}
	if err := effectstest.B22TryModal(g, effectstest.B22HullBreachOracle, []int{1}, effectstest.B16TargetCard(rock)); err == nil {
		t.Error("mode 2 is 'target enchantment'")
	}
	effectstest.CastModal(t, g, "Hull Breach", "Sorcery", effectstest.B22HullBreachOracle, []int{0}, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode 1 destroys the artifact")
	}
	effectstest.CastModal(t, g, "Hull Breach", "Sorcery", effectstest.B22HullBreachOracle, []int{1}, effectstest.B16TargetCard(aura))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode 2 destroys the enchantment")
	}
	// Mode 3: an artifact AND an enchantment.
	rock3 := effectstest.B12Permanent(g, opp.ID, "Rock 3", "Artifact")
	effectstest.CastModal(t, g, "Hull Breach", "Sorcery", effectstest.B22HullBreachOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock2}, {Kind: game.TargetCard, ID: aura2}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock2) || g.Battlefield.Contains(aura2) {
		t.Error("mode 3 destroys both")
	}
	// Two artifacts in mode 3 are refused: each slot has its own
	// predicate now.
	rock4 := effectstest.B12Permanent(g, opp.ID, "Rock 4", "Artifact")
	if err := effectstest.B22TryModal(g, effectstest.B22HullBreachOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock3}, {Kind: game.TargetCard, ID: rock4}}); err == nil {
		t.Error("two artifacts cannot fill 'target artifact and target enchantment'")
	}
	aura3 := effectstest.B12Permanent(g, opp.ID, "Aura 3", "Enchantment")
	aura4 := effectstest.B12Permanent(g, opp.ID, "Aura 4", "Enchantment")
	if err := effectstest.B22TryModal(g, effectstest.B22HullBreachOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura3}, {Kind: game.TargetCard, ID: aura4}}); err == nil {
		t.Error("two enchantments cannot fill it either")
	}
	if !g.Battlefield.Contains(rock3) || !g.Battlefield.Contains(rock4) {
		t.Error("a refused cast destroys nothing")
	}
	if spec, _ := effects.Lookup(effectstest.B22HullBreachOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the per-slot predicate landed: nothing is deferred")
	}
}

// --- the artifacts -------------------------------------------------

func TestB22CodexShredderMillsAndRecursForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := effectstest.B12Push(g, me.ID, "Codex Shredder", "Artifact", effectstest.B22CodexShredderOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	library := opp.Library.Size()
	effectstest.B16Activate(t, g, me.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if got := opp.Library.Size(); got != library-1 {
		t.Errorf("target player mills one: %d → %d", library, got)
	}
	if !effectstest.B20Tapped(t, g, shredder) {
		t.Fatal("the mill has a tap cost")
	}
	effectstest.B22Untap(g, shredder)
	relic := effectstest.B17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	theirs := effectstest.B17GraveyardCard(opp, "Their Card", "Instant", "{U}")
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shredder, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's graveyard is not yours")
	}
	effectstest.B16Activate(t, g, me.ID, shredder, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(relic)})
	if g.Battlefield.Contains(shredder) {
		t.Error("the Shredder is sacrificed as the cost")
	}
	if !me.Hand.Contains(relic) {
		t.Error("the card comes back to hand")
	}
}

func TestB22MycosynthWellspringFetchesABasicOnEntryAndOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Bayou", "Land — Swamp Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	well := effectstest.CastCatalogSpell(t, g, "Mycosynth Wellspring", "Artifact", effectstest.B22MycosynthWellspringOID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SearchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("entering offers a search")
	}
	if effectstest.SearchOptionNamed(g, prompt, "Bayou") != uuid.Nil {
		t.Error("a Bayou is not a basic land")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	if !b22HandHasNamed(me, "Forest") {
		t.Error("the Forest should be in hand")
	}
	// Dies: the same search again; "you may" declines.
	effectstest.B18Kill(t, g, well)
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("going to the graveyard from the battlefield offers the search again")
	}
	effectstest.AnswerSearchFailToFind(t, g, me.ID)
	if b22HandHasNamed(me, "Island") {
		t.Error("declining fetches nothing")
	}
	// Exiled from the battlefield: not "put into a graveyard".
	again := effectstest.B12Push(g, me.ID, "Mycosynth Wellspring", "Artifact", effectstest.B22MycosynthWellspringOID, 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(again) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("exile is not a graveyard")
	}
}

// --- the creatures with activations --------------------------------

func TestB22IorethUntapsAnotherPermanentOrTwoLegends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ioreth := effectstest.B12Push(g, me.ID, "Ioreth of the Healing House", "Legendary Creature — Human Cleric", effectstest.B22IorethOracle, 1, 4)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	legendA := effectstest.B12Push(g, me.ID, "Legend A", "Legendary Creature — Human", "", 2, 2)
	legendB := effectstest.B12Push(g, opp.ID, "Legend B", "Legendary Creature — Elf", "", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{land, legendA, legendB, bear} {
		b08Tap(g, id)
	}
	effectstest.AdvanceToMain(t, g)
	// "Another": Ioreth herself is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, ioreth, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(ioreth)}); err == nil {
		t.Fatal("Ioreth cannot untap herself")
	}
	effectstest.B16Activate(t, g, me.ID, ioreth, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(land)})
	if effectstest.B20Tapped(t, g, land) {
		t.Error("the land should be untapped")
	}
	effectstest.B22Untap(g, ioreth)
	// Two legendary creatures — anyone's; a Bear is not legendary.
	if err := g.ActivateCatalogAbility(me.ID, ioreth, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legendA}, {Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("a Bear is not a legendary creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, ioreth, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legendA}},
	}); err == nil {
		t.Fatal("the second ability needs two targets")
	}
	effectstest.B16Activate(t, g, me.ID, ioreth, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legendA}, {Kind: game.TargetCard, ID: legendB}},
	})
	if effectstest.B20Tapped(t, g, legendA) || effectstest.B20Tapped(t, g, legendB) {
		t.Error("both legends should be untapped")
	}
	if !effectstest.B20Tapped(t, g, ioreth) {
		t.Error("both abilities have a tap cost")
	}
}

func TestB22RograkhHasHisThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rograkh := effectstest.B12Push(g, me.ID, "Rograkh, Son of Rohgahh", "Legendary Creature — Kobold Warrior", effectstest.B22RograkhOracle, 0, 1)
	effectstest.AssertKeywords(t, g, rograkh, "first strike", "menace", "trample")
}
