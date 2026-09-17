package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch22_test.go — card-level coverage for the card-coverage
// roadmap's batch 22 (#384, `edhrec_rank` 2335–2435): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, land play or step change.
// Helpers from the earlier batch test files are reused by name; new
// ones are b22-prefixed.

const (
	b22WhelmingWaveAlreadyOID   = "e510eaaf-6497-480f-baa8-f4796b5f1086"
	b22ConclaveMentorOracle     = "a2fe5937-212c-4e71-8d6e-f408b38100aa"
	b22RograkhOracle            = "584cee10-f18c-4633-95cc-f2e7a11841ac"
	b22ShivanGorgeOracle        = "e90a1381-c9c1-4f57-928c-5d19dc065274"
	b22DictateOfTheTwinGodsOID  = "3d960d33-623a-4415-ae00-f8cffbc15f5a"
	b22SlaughterTheStrongOracle = "7a3569a0-a55f-40c9-9588-92db35e26567"
	b22TarnishedCitadelOracle   = "66ae2562-68e9-4c77-ba0a-57f8ff37f656"
	b22SamwiseGamgeeOracle      = "7ce37c26-91ea-493f-bfc3-a890d4538bc1"
	b22DisdainfulStrokeOracle   = "11e02134-7b1a-46a4-a89e-7539dd1efada"
	b22ToxrillOracle            = "c71b2325-bde6-4364-a93b-8477ffeb25d8"
	b22CabbageMerchantOracle    = "e31808bb-fb6e-487b-a973-db44961c84ee"
	b22CauldronOfEssenceOracle  = "a2f8cde8-bf7b-4234-89f0-a95f9dc937e3"
	b22EarthshakerDreadmawOID   = "454cbc6a-b7f0-445e-842c-5db267917a18"
	b22BerserkersOnslaughtOID   = "85d2948c-1a79-418d-80c4-bc1012a4d313"
	b22FuryOracle               = "fbf9f8c5-849f-45d5-8129-5fc683c21a04"
	b22WirewoodLodgeOracle      = "1275653f-de4e-4fe9-aad8-88555fa11680"
	b22WavebreakHippocampOracle = "3405c8a9-a8d6-4b45-9b64-94141076603b"
	b22SingularityRuptureOracle = "e1976b6c-7e43-4f4a-b082-f95495a1b260"
	b22CodexShredderOracle      = "ef7b11ab-24e7-4e7c-91a7-920bade6e60b"
	b22WearDownOracle           = "27905301-333e-4cdd-90cf-188159fcf8e9"
	b22HullBreachOracle         = "2da232d8-580f-4116-b977-2c59cd21b5a4"
	b22LilianasCaressOracle     = "a4aec0d6-13fa-4709-b1a9-2f483f032744"
	b22SabotenderOracle         = "65385311-1158-4e94-892a-683997706ca8"
	b22DawnOfHopeOracle         = "d7a38484-2acc-49a6-b32d-d54dddb14d31"
	b22MycosynthWellspringOID   = "9ec43ec6-b625-4be8-8f79-3679e6657dbc"
	b22EreborFlamesmithOracle   = "6eb93545-a1a6-4447-82c5-421d8e9f023d"
	b22IorethOracle             = "af6e4c3e-0276-4f72-9a70-25868fe8bba5"
	b22AethericAmplifierOracle  = "295cd8e1-0830-46c1-9957-556afd4bcee6"
	b22GodEternalOketraOracle   = "3b03358d-f87e-4939-afc9-5ee3f044146a"
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

// b22Untap untaps a battlefield card through the effect API.
func b22Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

// b22CastInstantAs puts an instant in a player's hand and casts it
// from wherever the cursor is — the non-active seat's instant-speed
// cast, for a "during each opponent's turn" trigger.
func b22CastInstantAs(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: "{U}",
		Owner: p.ID, Controller: p.ID,
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Whelming Wave
// was already on main from the S23 boardwipes (#382) and is pinned
// here so the table matches the issue's 34.
func TestBatch22CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b22WhelmingWaveAlreadyOID:   "Whelming Wave",
		b22ConclaveMentorOracle:     "Conclave Mentor",
		b22RograkhOracle:            "Rograkh, Son of Rohgahh",
		b22ShivanGorgeOracle:        "Shivan Gorge",
		b22DictateOfTheTwinGodsOID:  "Dictate of the Twin Gods",
		b22SlaughterTheStrongOracle: "Slaughter the Strong",
		b22TarnishedCitadelOracle:   "Tarnished Citadel",
		b22SamwiseGamgeeOracle:      "Samwise Gamgee",
		b22DisdainfulStrokeOracle:   "Disdainful Stroke",
		b22ToxrillOracle:            "Toxrill, the Corrosive",
		b22CabbageMerchantOracle:    "The Cabbage Merchant",
		b22CauldronOfEssenceOracle:  "Cauldron of Essence",
		b22EarthshakerDreadmawOID:   "Earthshaker Dreadmaw",
		b22BerserkersOnslaughtOID:   "Berserkers' Onslaught",
		b22FuryOracle:               "Fury",
		b22WirewoodLodgeOracle:      "Wirewood Lodge",
		b22WavebreakHippocampOracle: "Wavebreak Hippocamp",
		b22SingularityRuptureOracle: "Singularity Rupture",
		b22CodexShredderOracle:      "Codex Shredder",
		b22WearDownOracle:           "Wear Down",
		b22HullBreachOracle:         "Hull Breach",
		b22LilianasCaressOracle:     "Liliana's Caress",
		b22SabotenderOracle:         "Sabotender",
		b22DawnOfHopeOracle:         "Dawn of Hope",
		b22MycosynthWellspringOID:   "Mycosynth Wellspring",
		b22EreborFlamesmithOracle:   "Erebor Flamesmith",
		b22IorethOracle:             "Ioreth of the Healing House",
		b22AethericAmplifierOracle:  "Aetheric Amplifier",
		b22GodEternalOketraOracle:   "God-Eternal Oketra",
	}
	if len(want) != 29 {
		t.Fatalf("the batch registers 28 cards plus Whelming Wave, the table lists %d", len(want))
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
	// The five declared skips must NOT be registered — each has a
	// cost or a trigger the engine cannot express, and a spec would
	// ship the card stronger than printed.
	for _, skipped := range []string{
		"119d719d-e965-45b4-9bc9-ac03211b10c2", // Survival of the Fittest — discard-a-card cost
		"e38e3723-05f5-4a51-8364-1cda19f9cc49", // Steelbane Hydra — remove-a-counter cost
		"168336e2-d795-4b75-bf21-ce128b0dd7b0", // Harsh Mentor — no ability-activated event
		"ade898df-14a5-460b-94f6-1f3f74d3ff95", // Odric, Master Tactician — choose the blocks
		"da46786e-28df-4638-ab3d-121011d2f150", // Master Transmuter — return-an-artifact cost
	} {
		if _, ok := Lookup(skipped); ok {
			t.Errorf("%s is a declared skip and must not be registered", skipped)
		}
	}
}

// --- the lands -----------------------------------------------------

func TestB22ShivanGorgeTapsForColorlessOrPingsTheTable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gorge := b12Push(g, me.ID, "Shivan Gorge", "Legendary Land", b22ShivanGorgeOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, gorge, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	// The two abilities share the tap.
	me.ManaPool.EmptyPool()
	b06AddMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gorge, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a tapped Gorge cannot pay {T} again")
	}
	b22Untap(g, gorge)
	before := b22Lives(g)
	b16Activate(t, g, me.ID, gorge, 0, game.ActivateAbilityParams{})
	for i, p := range g.Seats {
		want := before[i]
		if i != 0 {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %d: %d → %d, want %d", i, before[i], p.Life, want)
		}
	}
	if !b20Tapped(t, g, gorge) {
		t.Error("the ping has a tap cost")
	}
}

func TestB22TarnishedCitadelColoredHalfDealsThreeColorlessHalfIsFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	citadel := seedPermanentWithOracle(g, me.ID, "Tarnished Citadel", "Land", b22TarnishedCitadelOracle)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, citadel, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} half: %v", err)
	}
	if me.Life != life {
		t.Errorf("the {C} half hurt: %d → %d", life, me.Life)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
	b22Untap(g, citadel)
	me.ManaPool.EmptyPool()
	if err := g.ActivateManaAbility(me.ID, citadel, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the coloured half: %v", err)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, me.Life)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the coloured half must offer all five colours, got %+v", pick)
	}
}

func TestB22WirewoodLodgeUntapsAnElf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lodge := b12Push(g, me.ID, "Wirewood Lodge", "Land", b22WirewoodLodgeOracle, 0, 0)
	elf := b12Creature(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b08Tap(g, elf)
	b08Tap(g, bear)
	advanceToMain(t, g)
	b06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, lodge, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a Bear is not an Elf")
	}
	b16Activate(t, g, me.ID, lodge, 0, game.ActivateAbilityParams{Targets: b16TargetCard(elf)})
	if b20Tapped(t, g, elf) {
		t.Error("the Elf — anyone's — should be untapped")
	}
	if !b20Tapped(t, g, lodge) {
		t.Error("the untap has a tap cost")
	}
}

// --- the counters family -------------------------------------------

func TestB22ConclaveMentorAddsACounterAndGainsItsPowerOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mentor := b12Push(g, me.ID, "Conclave Mentor", "Creature — Centaur Cleric", b22ConclaveMentorOracle, 2, 2)
	bear := seedCreature(g, "Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)

	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, "+1/+1", 1)
		_ = g.AddCounterForEffect(mentor, "+1/+1", 2)
		_ = g.AddCounterForEffect(theirs, "+1/+1", 1)
		_ = g.AddCounterForEffect(bear, "charge", 1)
	})
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("one +1/+1 counter on my creature should land as two, got %d", got)
	}
	if got := counterCount(g, mentor, "+1/+1"); got != 3 {
		t.Errorf("the Mentor grows itself: two should land as three, got %d", got)
	}
	if got := counterCount(g, theirs, "+1/+1"); got != 1 {
		t.Errorf("an opponent's creature is not 'a creature you control', got %d", got)
	}
	if got := counterCount(g, bear, "charge"); got != 1 {
		t.Errorf("only +1/+1 counters are replaced, charge landed as %d", got)
	}

	// Dies: life equal to its power, counters included (2 + 3).
	life := me.Life
	b18Kill(t, g, mentor)
	if me.Life != life+5 {
		t.Errorf("life %d → %d, want +5 (a 2/2 with three +1/+1 counters)", life, me.Life)
	}
}

func TestB22ToxrillSlimesShrinksAndMakesSlugs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toxrill := b12Push(g, me.ID, "Toxrill, the Corrosive", "Legendary Creature — Slug Horror", b22ToxrillOracle, 7, 7)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)

	// EACH end step — the controller's own first — slimes every
	// creature the controller does not control.
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if counterCount(g, theirs, "slime") != 1 || counterCount(g, wall, "slime") != 1 {
		t.Fatal("each creature you don't control gets a slime counter at the end step")
	}
	if counterCount(g, mine, "slime") != 0 || counterCount(g, toxrill, "slime") != 0 {
		t.Error("your own creatures are not slimed")
	}
	if effectivePower(t, g, wall) != -1 || effectiveToughness(t, g, wall) != 3 {
		t.Errorf("a slimed 0/4 should read -1/3, got %d/%d", effectivePower(t, g, wall), effectiveToughness(t, g, wall))
	}
	if effectiveToughness(t, g, mine) != 2 {
		t.Error("the -1/-1 is only for creatures you don't control")
	}

	// An opponent's end step counts too: the 2/2 reaches 0/0 and dies
	// with a slime counter on it — a Slug for Toxrill's controller.
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("two slime counters kill a 2/2")
	}
	if countBattlefieldNamed(g, me.ID, "Slug") != 1 {
		t.Errorf("a slimed creature dying makes a Slug, got %d", countBattlefieldNamed(g, me.ID, "Slug"))
	}
	if c, _ := battlefieldCard(g, findBattlefieldByName(g, "Slug")); c.Power != 1 || c.Toughness != 1 || !c.HasColor("B") {
		t.Error("a 1/1 black Slug")
	}

	// An unslimed opponent's creature dying is silent.
	fresh := b12Creature(g, opp.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	b18Kill(t, g, fresh)
	if countBattlefieldNamed(g, me.ID, "Slug") != 1 {
		t.Error("a creature without a slime counter makes no Slug")
	}

	// {U}{B}, Sacrifice a Slug: draw a card.
	slug := findBattlefieldByName(g, "Slug")
	advanceToMainOf(t, g, 0)
	b06AddMana(me, "U", "B")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{slug}})
	if g.Battlefield.Contains(slug) {
		t.Error("the Slug is sacrificed as the cost")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	// A Bear is not a Slug; Toxrill itself is.
	b06AddMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{mine}}); err == nil {
		t.Error("a Bear cannot pay 'sacrifice a Slug'")
	}
	b16Activate(t, g, me.ID, toxrill, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{toxrill}})
	if g.Battlefield.Contains(toxrill) {
		t.Error("Toxrill is a Slug and may eat itself")
	}
}

func TestB22AethericAmplifierTapsForAnyColourAndDoublesCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	amp := b12Push(g, me.ID, "Aetheric Amplifier", "Artifact", b22AethericAmplifierOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, amp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("any colour, the printed width, got %+v", pick)
	}
	// #730: an unanswered colour pick gates the cursor, and this test
	// walks the turn on below. The colour itself is beside the point.
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	b22Untap(g, amp)
	theirs := pushCounterCreature(g, opp.ID, "Their Hydra", "+1/+1", 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, "charge", 1) })
	// Sorcery speed: not during combat.
	advanceTo(t, g, game.StepDeclareAttackers)
	b06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, amp, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("activate only as a sorcery")
	}
	advanceToMainOf(t, g, 0)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, amp, 0, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)})
	if got := counterCount(g, theirs, "+1/+1"); got != 6 {
		t.Errorf("+1/+1 counters 3 → %d, want 6", got)
	}
	if got := counterCount(g, theirs, "charge"); got != 2 {
		t.Errorf("each KIND doubles: charge 1 → %d, want 2", got)
	}
	if !b20Tapped(t, g, amp) {
		t.Error("the activation has a tap cost")
	}
}

// --- the damage family ---------------------------------------------

func TestB22DictateOfTheTwinGodsDoublesEveryonesDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dictate := castCatalogSpell(t, g, "Dictate of the Twin Gods", "Enchantment", b22DictateOfTheTwinGodsOID, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dictate) {
		t.Fatal("the Dictate did not resolve")
	}
	if !hasAbility(effectiveAbilities(t, g, dictate), "flash") {
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
	b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	passPriorityAroundTable(t, g)
	if me.Life != mine-6 {
		t.Errorf("an opponent's Bolt under the Dictate: %d → %d, want -6", mine, me.Life)
	}
}

func TestB22ShivanGorgeDamageIsDoubledByTheDictate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dictate of the Twin Gods", "Enchantment", b22DictateOfTheTwinGodsOID, 0, 0)
	gorge := b12Push(g, me.ID, "Shivan Gorge", "Legendary Land", b22ShivanGorgeOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "R", "C", "C")
	before := b22Lives(g)
	b16Activate(t, g, me.ID, gorge, 0, game.ActivateAbilityParams{})
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-2 {
			t.Errorf("opponent %d: %d → %d, want -2 (the Gorge is the source)", i, before[i], got)
		}
	}
}

func TestB22SabotenderPingsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sabotender := b12Push(g, me.ID, "Sabotender", "Creature — Plant", b22SabotenderOracle, 2, 1)
	if !hasAbility(effectiveAbilities(t, g, sabotender), "reach") {
		t.Error("printed reach did not reach the effective abilities")
	}
	before := b22Lives(g)
	playLandFromHand(t, g, "Mountain", "")
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
	if got := g.Seats[2].Life; got != before[2]-1 {
		t.Error("an opponent's land triggered my landfall")
	}
}

func TestB22EreborFlamesmithPingsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Erebor Flamesmith", "Creature — Dwarf Artificer", b22EreborFlamesmithOracle, 2, 1)
	before := b22Lives(g)
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d after an instant: %d → %d, want -1", i, before[i], got)
		}
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for i := 1; i < len(g.Seats); i++ {
		if got := g.Seats[i].Life; got != before[i]-1 {
			t.Errorf("opponent %d after a creature: %d → %d, want unchanged", i, before[i]-1, got)
		}
	}
}

func TestB22FuryDividesFourDamageEvenlyAndEvokesForARedCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 3, 3)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 3, 3)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 3, 3)

	fury := castCatalogSpell(t, g, "Fury", "Creature — Elemental Incarnation", b22FuryOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, fury), "double strike") {
		t.Error("printed double strike did not reach the effective abilities")
	}
	// Three targets in the order picked: 2 / 1 / 1.
	b17PickCards(t, g, me.ID, a, b, c)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, a); got != 2 {
		t.Errorf("the first pick takes the remainder: %d damage, want 2", got)
	}
	if damageMarkedOn(g, b) != 1 || damageMarkedOn(g, c) != 1 {
		t.Errorf("the rest take 1 each: %d / %d", damageMarkedOn(g, b), damageMarkedOn(g, c))
	}

	// Evoke: pitch a red card, the ETB still fires, then Fury dies.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	toMainForCost(t, g2)
	pitch := handCardFull(me2, "Lightning Bolt", "Instant", "{R}", "", []string{"R"})
	blue := handCardFull(me2, "Counterspell", "Instant", "{U}{U}", "", []string{"U"})
	victim := b12Creature(g2, opp2.ID, "Victim", "Creature — Bear", 4, 4)
	fury2 := handCardFull(me2, "Fury", "Creature — Elemental Incarnation", "{3}{R}{R}", b22FuryOracle, []string{"R"})
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
	passPriorityAroundTable(t, g2)
	b17PickCards(t, g2, me2.ID, victim)
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(victim) {
		t.Error("one target takes all 4")
	}
	if g2.Battlefield.Contains(fury2) || !me2.Graveyard.Contains(fury2) {
		t.Error("an evoked Fury is sacrificed after its trigger resolves")
	}
}

func TestB22BerserkersOnslaughtGivesAttackersDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Berserkers' Onslaught", "Enchantment", b22BerserkersOnslaughtOID, 0, 0)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	home := pushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, attacker), "double strike") {
		t.Fatal("the attacking creature should have double strike")
	}
	if hasAbility(effectiveAbilities(t, g, home), "double strike") {
		t.Error("a creature that did not attack is not an attacking creature")
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("a 2/2 with double strike deals 4: %d → %d", life, opp.Life)
	}

	// An opponent's attacker on their turn is not "you control".
	advanceToNextSeatsTurn(t, g)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if hasAbility(effectiveAbilities(t, g, theirs), "double strike") {
		t.Error("an opponent's attacker is not yours")
	}
}

// --- the counterspell ----------------------------------------------

func TestB22DisdainfulStrokeCountersOnlyBigSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Disdainful Stroke", TypeLine: "Instant",
		OracleID: b22DisdainfulStrokeOracle, Owner: me.ID, Controller: me.ID})
	// A spell with no printed cost is mana value 0.
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{Targets: b16TargetCard(small)}); err == nil {
		t.Fatal("a mana value 0 spell is not 'mana value 4 or greater'")
	}
	passPriorityAroundTable(t, g)

	big := castWithCost(t, g, "Wurm", "Creature — Wurm", "{4}{G}{G}", "")
	stroke := castCatalogSpell(t, g, "Disdainful Stroke", "Instant", b22DisdainfulStrokeOracle, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !me.Graveyard.Contains(big) {
		t.Error("the six-drop should be countered")
	}
	if !me.Graveyard.Contains(stroke) {
		t.Error("the Stroke goes to the graveyard")
	}

	// X counts on the stack (CR 202.3e): a Hydra cast for X=3 is 5.
	hydra := castXSpell(t, g, "Hydra", "Creature — Hydra", "", "{X}{G}{G}", 3, nil)
	castCatalogSpell(t, g, "Disdainful Stroke", "Instant", b22DisdainfulStrokeOracle, b16TargetCard(hydra))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hydra) || !me.Graveyard.Contains(hydra) {
		t.Error("X is counted on the stack — the Hydra cast for X=3 has mana value 5")
	}
}

// --- the go-wide payoffs -------------------------------------------

func TestB22SamwiseGamgeeMakesFoodForNontokenCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Samwise Gamgee", "Legendary Creature — Halfling Peasant", b22SamwiseGamgeeOracle, 2, 2)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatalf("a nontoken creature entering makes a Food, got %d", countBattlefieldNamed(g, me.ID, "Food"))
	}
	// A token, an opponent's creature and a noncreature are silent.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Errorf("tokens, opponents' creatures and artifacts do not make Food, got %d", countBattlefieldNamed(g, me.ID, "Food"))
	}
	// #747: the three-Food ability ships at its printed count; its
	// engine test is in sacrifice_n_cards_test.go.
	spec, _ := Lookup(b22SamwiseGamgeeOracle)
	if len(spec.Activated) != 1 || spec.Completeness != CompletenessFull {
		t.Error("the three-Food ability ships whole")
	}
}

func TestB22CabbageMerchantMakesFoodAndLosesItToCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "The Cabbage Merchant", "Legendary Creature — Human Citizen", b22CabbageMerchantOracle, 2, 2)
	// My own noncreature spell is silent; an opponent's makes a Food.
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 0 {
		t.Fatal("your own spell is not an opponent's")
	}
	b13OpponentCasts(t, g, opp, "Shock", "Instant", "", "{R}", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Fatalf("an opponent's noncreature spell makes a Food, got %d", countBattlefieldNamed(g, me.ID, "Food"))
	}
	advanceToNextSeatsTurn(t, g)
	b20CastCreature(t, g, opp, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 {
		t.Error("an opponent's CREATURE spell is silent")
	}
	// A creature dealing combat damage to me: sacrifice a Food token.
	food := findBattlefieldByName(g, "Food")
	cabin := b12Push(g, me.ID, "Gingerbread Cabin", "Land — Forest Food", "", 0, 0)
	attacker := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	dealCombatDamageToPlayer(g, attacker, me.ID, 2)
	passPriorityAroundTable(t, g)
	prompt := sacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("combat damage to the Merchant's controller asks for a Food token")
	}
	if len(prompt.SacrificeOptions) != 1 || prompt.SacrificeOptions[0] != food {
		t.Errorf("only Food TOKENS are offered (not the Cabin), got %v", prompt.SacrificeOptions)
	}
	answerSacrifice(t, g, me.ID, food)
	if g.Battlefield.Contains(food) || !g.Battlefield.Contains(cabin) {
		t.Error("the Food token is sacrificed, the Cabin stays")
	}
	// Non-combat damage is silent.
	b07BoltPlayer(t, g, me.ID)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a Bolt is not combat damage")
	}
	spec, _ := Lookup(b22CabbageMerchantOracle)
	if len(spec.ManaAbilities) != 0 || spec.Completeness != CompletenessCaveats {
		t.Error("the tap-two-Foods mana ability is a declared gap")
	}
}

func TestB22CauldronOfEssenceDrainsOnDeathAndReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cauldron := b12Push(g, me.ID, "Cauldron of Essence", "Artifact", b22CauldronOfEssenceOracle, 0, 0)
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	before := b22Lives(g)
	b18Kill(t, g, theirs)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("an opponent's creature dying is silent, seat %d moved", i)
		}
	}
	b18Kill(t, g, mine)
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
	fodder := seedCreature(g, "Fodder", me.ID)
	advanceTo(t, g, game.StepDeclareAttackers)
	b06AddMana(me, "B", "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: b16TargetCard(mine),
	}); err == nil {
		t.Fatal("activate only as a sorcery")
	}
	advanceToMainOf(t, g, 0)
	b06AddMana(me, "B", "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: b16TargetCard(theirs),
	}); err == nil {
		t.Fatal("an opponent's graveyard is not yours")
	}
	mid := b22Lives(g)
	b16Activate(t, g, me.ID, cauldron, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: b16TargetCard(mine),
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
	if !b20Tapped(t, g, cauldron) {
		t.Error("the reanimation has a tap cost")
	}
}

func TestB22EarthshakerDreadmawDrawsPerOtherDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 3, 3)
	b12Creature(g, me.ID, "Ripjaw", "Creature — Dinosaur", 4, 5)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Dinosaur", "Creature — Dinosaur", 3, 3)
	hand := me.Hand.Size()
	maw := castCatalogSpell(t, g, "Earthshaker Dreadmaw", "Creature — Dinosaur", b22EarthshakerDreadmawOID, nil)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, maw), "trample") {
		t.Error("printed trample did not reach the effective abilities")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("drew %d, want 2 (two OTHER Dinosaurs you control)", got-hand)
	}
}

func TestB22GodEternalOketraMakesZombiesAndReturnsThirdFromTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oketra := b12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", b22GodEternalOketraOracle, 3, 6)
	if !hasAbility(effectiveAbilities(t, g, oketra), "double strike") {
		t.Error("printed double strike did not reach the effective abilities")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Zombie Warrior") != 1 {
		t.Fatalf("a creature spell makes a Zombie Warrior, got %d", countBattlefieldNamed(g, me.ID, "Zombie Warrior"))
	}
	zombie := findBattlefieldByName(g, "Zombie Warrior")
	if c, _ := battlefieldCard(g, zombie); c.Power != 4 || c.Toughness != 4 || !c.HasColor("B") {
		t.Error("a 4/4 black Zombie Warrior")
	}
	assertKeywords(t, g, zombie, "vigilance")
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Zombie Warrior") != 1 {
		t.Error("a noncreature spell is silent")
	}

	// Dies: may go third from the top of the library.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(oketra) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(oketra) {
		t.Fatal("Oketra should have left the graveyard")
	}
	names := libraryTopNames(me, 3)
	if len(names) != 3 || names[2] != "God-Eternal Oketra" || names[0] == "God-Eternal Oketra" {
		t.Errorf("third from the top, got %v", names)
	}

	// Exiled from the battlefield: the same offer, declined.
	again := b12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", b22GodEternalOketraOracle, 3, 6)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(again) })
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(again) {
		t.Error("declining leaves Oketra in exile")
	}
	// Bounced: neither dies nor exiled, no offer.
	third := b12Push(g, me.ID, "God-Eternal Oketra", "Legendary Creature — Zombie God", b22GodEternalOketraOracle, 3, 6)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(third) })
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a bounce is not 'dies or is put into exile'")
	}
}

// --- the enchantments ----------------------------------------------

func TestB22LilianasCaressDrainsTwoPerOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Liliana's Caress", "Enchantment", b22LilianasCaressOracle, 0, 0)
	pushHandCard(g, opp)
	pushHandCard(g, opp)
	pushHandCard(g, me)
	life, mine := opp.Life, me.Life
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("two discards: %d → %d, want -4", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Error("the controller's own discard is silent")
	}
}

func TestB22DawnOfHopeOffersToPayOnLifegainAndMakesLifelinkSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dawn := b12Push(g, me.ID, "Dawn of Hope", "Enchantment", b22DawnOfHopeOracle, 0, 0)
	hand := me.Hand.Size()
	b14Gain(t, g, me.ID, 3)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("gaining life asks the controller to pay {2}")
	}
	b06AddMana(me, "C", "C")
	answerPayUnless(t, g, me.ID, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("paying draws one, drew %d", got-hand)
	}
	b14Gain(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	if got := me.Hand.Size(); got != hand+1 {
		t.Error("declining draws nothing")
	}
	b14Gain(t, g, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's lifegain is not yours")
	}

	advanceToMain(t, g)
	b06AddMana(me, "W", "C", "C", "C")
	b16Activate(t, g, me.ID, dawn, 0, game.ActivateAbilityParams{})
	soldier := findBattlefieldByName(g, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier")
	}
	if c, _ := battlefieldCard(g, soldier); c.Power != 1 || c.Toughness != 1 || !c.HasColor("W") {
		t.Error("a 1/1 white Soldier")
	}
	assertKeywords(t, g, soldier, "lifelink")
}

func TestB22WavebreakHippocampDrawsOnTheFirstSpellOfEachOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Wavebreak Hippocamp", "Enchantment Creature — Horse Fish", b22WavebreakHippocampOracle, 2, 2)
	// On my own turn: silent.
	// Every cast helper puts the card in hand first, so a cast with
	// no draw leaves the hand size where it was.
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("a spell on your own turn draws nothing, hand %d → %d", hand, got)
	}
	// On an opponent's turn: the first draws, the second does not.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	hand = me.Hand.Size()
	b22CastInstantAs(t, g, me, "Opt")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the first spell on an opponent's turn draws — hand %d → %d", hand, got)
	}
	b22CastInstantAs(t, g, me, "Brainstorm")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the second spell that turn is silent — hand %d → %d", hand, got)
	}
	// The next opponent's turn resets the tally.
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	hand = me.Hand.Size()
	b22CastInstantAs(t, g, me, "Opt")
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("each opponent's turn has its own first spell — hand %d → %d", hand, got)
	}
}

// --- the sorceries -------------------------------------------------

func TestB22SlaughterTheStrongKeepsLowestPowerWithinFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myBig := b12Creature(g, me.ID, "My Big", "Creature — Wurm", 5, 5)
	myTwo := b12Creature(g, me.ID, "My Two", "Creature — Bear", 2, 2)
	myOne := b12Creature(g, me.ID, "My One", "Creature — Elf", 1, 1)
	myZero := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 4)
	myOther := b12Creature(g, me.ID, "My Other Two", "Creature — Bear", 2, 2)
	theirFour := b12Creature(g, opp.ID, "Their Four", "Creature — Beast", 4, 4)
	theirOne := b12Creature(g, opp.ID, "Their One", "Creature — Elf", 1, 1)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")

	castCatalogSpell(t, g, "Slaughter the Strong", "Sorcery", b22SlaughterTheStrongOracle, nil)
	passPriorityAroundTable(t, g)
	// Me: 0 + 1 + 2 = 3 fits; the next 2 would make 5, so it and the
	// 5/5 go.
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
	// Them: 1 fits; 1 + 4 = 5 does not.
	if !g.Battlefield.Contains(theirOne) || g.Battlefield.Contains(theirFour) {
		t.Error("each player keeps their own lowest-power set")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("only creatures are sacrificed")
	}
	if !me.Graveyard.Contains(myBig) || !opp.Graveyard.Contains(theirFour) {
		t.Error("a sacrifice goes to its owner's graveyard")
	}
}

func TestB22SingularityRuptureWipesThenMillsHalf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := seedCreature(g, "My Bear", me.ID)
	theirs := seedCreature(g, "Their Bear", opp.ID)
	opp.Library.Cards, other.Library.Cards = nil, nil
	seedLibrary(opp, "a", "b", "c", "d", "e", "f", "g")
	seedLibrary(other, "a", "b", "c", "d")
	myLibrary := me.Library.Size()
	castCatalogSpell(t, g, "Singularity Rupture", "Sorcery", b22SingularityRuptureOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: other.ID}})
	passPriorityAroundTable(t, g)
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
	castCatalogSpell(t, g, "Singularity Rupture", "Sorcery", b22SingularityRuptureOracle, nil)
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != 4 {
		t.Error("with no targets nobody mills")
	}
}

func TestB22WearDownDestroysOneArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := seedCreature(g, "Bear", opp.ID)
	if b09TryCast(t, g, "Wear Down", "Sorcery", b22WearDownOracle, b16TargetCard(bear)) == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	castCatalogSpell(t, g, "Wear Down", "Sorcery", b22WearDownOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	spec, _ := Lookup(b22WearDownOracle)
	if spec.Targets == nil || spec.Targets.Max != 1 || spec.Completeness != CompletenessCaveats {
		t.Error("the gift is a declared gap: one target, never two")
	}
}

func TestB22HullBreachThreeModes(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	rock2 := b12Permanent(g, opp.ID, "Rock 2", "Artifact")
	aura := b12Permanent(g, opp.ID, "Aura", "Enchantment")
	aura2 := b12Permanent(g, opp.ID, "Aura 2", "Enchantment")

	// Mode 1 refuses an enchantment; mode 2 refuses an artifact.
	if err := b22TryModal(g, b22HullBreachOracle, []int{0}, b16TargetCard(aura)); err == nil {
		t.Error("mode 1 is 'target artifact'")
	}
	if err := b22TryModal(g, b22HullBreachOracle, []int{1}, b16TargetCard(rock)); err == nil {
		t.Error("mode 2 is 'target enchantment'")
	}
	castModal(t, g, "Hull Breach", "Sorcery", b22HullBreachOracle, []int{0}, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("mode 1 destroys the artifact")
	}
	castModal(t, g, "Hull Breach", "Sorcery", b22HullBreachOracle, []int{1}, b16TargetCard(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("mode 2 destroys the enchantment")
	}
	// Mode 3: an artifact AND an enchantment.
	rock3 := b12Permanent(g, opp.ID, "Rock 3", "Artifact")
	castModal(t, g, "Hull Breach", "Sorcery", b22HullBreachOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: aura2}, {Kind: game.TargetCard, ID: rock2}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock2) || g.Battlefield.Contains(aura2) {
		t.Error("mode 3 destroys both")
	}
	// Two artifacts in mode 3: the declared gap — only the first dies.
	rock4 := b12Permanent(g, opp.ID, "Rock 4", "Artifact")
	castModal(t, g, "Hull Breach", "Sorcery", b22HullBreachOracle, []int{2},
		[]game.TargetRef{{Kind: game.TargetCard, ID: rock3}, {Kind: game.TargetCard, ID: rock4}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock3) {
		t.Error("the first artifact picked is destroyed")
	}
	if !g.Battlefield.Contains(rock4) {
		t.Error("a second artifact fills no slot — weaker than printed, never two artifacts")
	}
}

// b22TryModal attempts a modal cast and returns the engine's answer.
func b22TryModal(g *game.Game, oracle string, modes []int, targets []game.TargetRef) error {
	active := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			return err
		}
	}
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Hull Breach", TypeLine: "Sorcery",
		OracleID: oracle, Owner: active.ID, Controller: active.ID})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
	if err != nil {
		_, _ = active.Hand.Remove(id)
	}
	return err
}

// --- the artifacts -------------------------------------------------

func TestB22CodexShredderMillsAndRecursForItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shredder := b12Push(g, me.ID, "Codex Shredder", "Artifact", b22CodexShredderOracle, 0, 0)
	advanceToMain(t, g)
	library := opp.Library.Size()
	b16Activate(t, g, me.ID, shredder, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	if got := opp.Library.Size(); got != library-1 {
		t.Errorf("target player mills one: %d → %d", library, got)
	}
	if !b20Tapped(t, g, shredder) {
		t.Fatal("the mill has a tap cost")
	}
	b22Untap(g, shredder)
	relic := b17GraveyardCard(me, "Sol Ring", "Artifact", "{1}")
	theirs := b17GraveyardCard(opp, "Their Card", "Instant", "{U}")
	b06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, shredder, 1, game.ActivateAbilityParams{Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's graveyard is not yours")
	}
	b16Activate(t, g, me.ID, shredder, 1, game.ActivateAbilityParams{Targets: b16TargetCard(relic)})
	if g.Battlefield.Contains(shredder) {
		t.Error("the Shredder is sacrificed as the cost")
	}
	if !me.Hand.Contains(relic) {
		t.Error("the card comes back to hand")
	}
}

func TestB22MycosynthWellspringFetchesABasicOnEntryAndOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Bayou", "Land — Swamp Forest"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	well := castCatalogSpell(t, g, "Mycosynth Wellspring", "Artifact", b22MycosynthWellspringOID, nil)
	passPriorityAroundTable(t, g)
	prompt := searchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("entering offers a search")
	}
	if searchOptionNamed(g, prompt, "Bayou") != uuid.Nil {
		t.Error("a Bayou is not a basic land")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	if !b22HandHasNamed(me, "Forest") {
		t.Error("the Forest should be in hand")
	}
	// Dies: the same search again; "you may" declines.
	b18Kill(t, g, well)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("going to the graveyard from the battlefield offers the search again")
	}
	answerSearchFailToFind(t, g, me.ID)
	if b22HandHasNamed(me, "Island") {
		t.Error("declining fetches nothing")
	}
	// Exiled from the battlefield: not "put into a graveyard".
	again := b12Push(g, me.ID, "Mycosynth Wellspring", "Artifact", b22MycosynthWellspringOID, 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(again) })
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("exile is not a graveyard")
	}
}

// --- the creatures with activations --------------------------------

func TestB22IorethUntapsAnotherPermanentOrTwoLegends(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ioreth := b12Push(g, me.ID, "Ioreth of the Healing House", "Legendary Creature — Human Cleric", b22IorethOracle, 1, 4)
	land := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	legendA := b12Push(g, me.ID, "Legend A", "Legendary Creature — Human", "", 2, 2)
	legendB := b12Push(g, opp.ID, "Legend B", "Legendary Creature — Elf", "", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{land, legendA, legendB, bear} {
		b08Tap(g, id)
	}
	advanceToMain(t, g)
	// "Another": Ioreth herself is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, ioreth, 0, game.ActivateAbilityParams{Targets: b16TargetCard(ioreth)}); err == nil {
		t.Fatal("Ioreth cannot untap herself")
	}
	b16Activate(t, g, me.ID, ioreth, 0, game.ActivateAbilityParams{Targets: b16TargetCard(land)})
	if b20Tapped(t, g, land) {
		t.Error("the land should be untapped")
	}
	b22Untap(g, ioreth)
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
	b16Activate(t, g, me.ID, ioreth, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legendA}, {Kind: game.TargetCard, ID: legendB}},
	})
	if b20Tapped(t, g, legendA) || b20Tapped(t, g, legendB) {
		t.Error("both legends should be untapped")
	}
	if !b20Tapped(t, g, ioreth) {
		t.Error("both abilities have a tap cost")
	}
}

func TestB22RograkhHasHisThreeKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rograkh := b12Push(g, me.ID, "Rograkh, Son of Rohgahh", "Legendary Creature — Kobold Warrior", b22RograkhOracle, 0, 1)
	assertKeywords(t, g, rograkh, "first strike", "menace", "trample")
}
