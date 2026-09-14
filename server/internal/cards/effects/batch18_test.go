package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch18_test.go — card-level coverage for the card-coverage
// roadmap's batch 18 (#311, `edhrec_rank` 1930–2031): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, land play, attack, block or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b18-prefixed.

const (
	b18DreamstoneHedronOracle   = "0e2575be-c596-4c8c-bf07-7941ca065721"
	b18SpringleafParadeOracle   = "b1305916-53cc-4021-897e-bbefc65dce78"
	b18StormfistCrusaderOracle  = "0d3bebc6-662a-4cd4-ad1e-aeed0c5e04f6"
	b18ChartACourseOracle       = "05878e49-93ad-4144-9c50-a0bb86126c2e"
	b18WoundReflectionOracle    = "b09206a4-8b73-4125-8b79-53f6fd511b16"
	b18SpectralSailorOracle     = "a8fdbcdf-479d-4582-9ad5-9fbd4c740c29"
	b18GrazilaxxOracle          = "d22ff377-d282-4a28-9dce-96f25913dc96"
	b18KorSpiritdancerOracle    = "2bfc7467-1316-43d3-9ca6-ff36cc2de607"
	b18ManyPartingsOracle       = "92ea750e-62b7-4422-a527-ddf435759d08"
	b18ObNixilisTheFallenOracle = "154104b7-4125-4276-8bc7-9a6edfe48cb9"
	b18MikaeusTheLunarchOracle  = "82f3faa8-39fa-450b-843f-d60a4c36d8f7"
	b18OgreSlumlordOracle       = "0a5e3748-2e58-4e53-9653-8af4e21cf223"
	b18HelpfulHunterOracle      = "c0864adb-e9aa-40b6-91ff-a0646193e887"
	b18OrzhovGuildgateOracle    = "57b37df5-fee4-4720-931f-f0cb0a8b338c"
	b18JunkDiverOracle          = "08c595da-9305-42a9-b72f-5ccc546edc01"
	b18VenserShaperSavantOracle = "0f41cefc-d6ff-4db7-ba35-502b7e081de1"
	b18SplitUpOracle            = "2e82520a-9da3-49ae-b5c8-37e7ac8853fe"
	b18HerosDownfallOracle      = "03df6a57-37c9-46d3-83b3-4a6240100714"
	b18CycleOfRenewalOracle     = "77eda626-5cef-496d-b291-cfa1ee6fc1fb"
	b18EurekaMomentOracle       = "0e2c11b2-d95f-4402-9a4a-afd3f7ffb8be"
	b18FalkenrathNobleOracle    = "3739b179-bc81-4737-8376-66a57e16b942"
	b18BitterbloomBearerOracle  = "5effb651-f9e0-4c14-8192-bd0e132b8d5d"
	b18EmeriaAngelOracle        = "ea6616e4-db8d-4905-80f8-cb0162906850"
	b18KederektParasiteOracle   = "fc7b46af-6c07-455a-99c7-f1bccaa2a5f4"
	b18WindsOfAbandonOracle     = "499a4715-6776-461b-bee4-22ed172d455e"
	b18MagdaTheHoardmasterOrcl  = "7fd2beb8-f823-4723-beec-e59b62127490"
	b18SangromancerOracle       = "920445ab-0ac2-4de7-bc1c-f5e58eb4424c"
	b18ThunderfootBalothOracle  = "f55334d9-1ec2-4667-bf28-5a0384d6f053"
	b18ArixmethesOracle         = "caeb39ee-f0bb-4305-9e8b-b30ba0a74c78"
	b18RakdosGuildgateOracle    = "361f534b-39d1-4421-b5a8-d3813c62f86d"
	b18GarruksPackleaderOracle  = "13279222-422d-4447-9451-2463b0c714c6"
	b18PerpetualTimepieceOracle = "17b4778b-82b1-4845-ad08-00f3ff66877b"
	b18WoodedBastionOracle      = "61b85077-64aa-4bcc-890d-2d88da9543c0"
	b18ThirstForKnowledgeOracle = "939e6f71-185e-41f2-9d54-72cce06f1dce"
	b18RighteousValkyrieOracle  = "891d2690-7144-4f87-b6ef-96f2469780a9"
	b18KrenkosCommandOracle     = "9cfe86ae-eebe-44aa-a956-4b3e9e621105"
)

// b18Counters reads a battlefield card's counters of one kind.
func b18Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

// b18Kill destroys a battlefield permanent through the single-target
// verb, under the lock, and settles what it triggered up to any
// prompt.
func b18Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	passPriorityAroundTable(t, g)
}

// b18CastFromHand puts a card in a player's hand and casts it with
// the given params from a main phase of the active seat.
func b18CastFromHand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// b18TriggerPromptCount counts the open yes/no trigger prompts
// addressed to a player.
func b18TriggerPromptCount(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			n++
		}
	}
	return n
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Two are rows
// in the Guildgate table and one is a filter-land one-liner, so a
// transposed row is invisible until someone plays that exact card.
func TestBatch18CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b18DreamstoneHedronOracle:   "Dreamstone Hedron",
		b18SpringleafParadeOracle:   "Springleaf Parade",
		b18StormfistCrusaderOracle:  "Stormfist Crusader",
		b18ChartACourseOracle:       "Chart a Course",
		b18WoundReflectionOracle:    "Wound Reflection",
		b18SpectralSailorOracle:     "Spectral Sailor",
		b18GrazilaxxOracle:          "Grazilaxx, Illithid Scholar",
		b18KorSpiritdancerOracle:    "Kor Spiritdancer",
		b18ManyPartingsOracle:       "Many Partings",
		b18ObNixilisTheFallenOracle: "Ob Nixilis, the Fallen",
		b18MikaeusTheLunarchOracle:  "Mikaeus, the Lunarch",
		b18OgreSlumlordOracle:       "Ogre Slumlord",
		b18HelpfulHunterOracle:      "Helpful Hunter",
		b18OrzhovGuildgateOracle:    "Orzhov Guildgate",
		b18JunkDiverOracle:          "Junk Diver",
		b18VenserShaperSavantOracle: "Venser, Shaper Savant",
		b18SplitUpOracle:            "Split Up",
		b18HerosDownfallOracle:      "Hero's Downfall",
		b18CycleOfRenewalOracle:     "Cycle of Renewal",
		b18EurekaMomentOracle:       "Eureka Moment",
		b18FalkenrathNobleOracle:    "Falkenrath Noble",
		b18BitterbloomBearerOracle:  "Bitterbloom Bearer",
		b18EmeriaAngelOracle:        "Emeria Angel",
		b18KederektParasiteOracle:   "Kederekt Parasite",
		b18WindsOfAbandonOracle:     "Winds of Abandon",
		b18MagdaTheHoardmasterOrcl:  "Magda, the Hoardmaster",
		b18SangromancerOracle:       "Sangromancer",
		b18ThunderfootBalothOracle:  "Thunderfoot Baloth",
		b18ArixmethesOracle:         "Arixmethes, Slumbering Isle",
		b18RakdosGuildgateOracle:    "Rakdos Guildgate",
		b18GarruksPackleaderOracle:  "Garruk's Packleader",
		b18PerpetualTimepieceOracle: "Perpetual Timepiece",
		b18WoodedBastionOracle:      "Wooded Bastion",
		b18ThirstForKnowledgeOracle: "Thirst for Knowledge",
		b18RighteousValkyrieOracle:  "Righteous Valkyrie",
		b18KrenkosCommandOracle:     "Krenko's Command",
	}
	if len(want) != 36 {
		t.Fatalf("the batch registers 36 cards, the table lists %d", len(want))
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

func TestB18GuildgateRowsEnterTappedAndProduceTheirColours(t *testing.T) {
	for _, row := range []struct{ oracle, name, produced string }{
		{b18OrzhovGuildgateOracle, "Orzhov Guildgate", "{W|B}"},
		{b18RakdosGuildgateOracle, "Rakdos Guildgate", "{B|R}"},
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
	g := newCatalogGame(t)
	gate := playLandFromHand(t, g, "Orzhov Guildgate", b18OrzhovGuildgateOracle)
	top100AssertEnteredTapped(t, g, gate, "Orzhov Guildgate")
	if tapEventsFor(g, gate) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
}

func TestB18WoodedBastionFiltersHybridIntoTwoPicks(t *testing.T) {
	spec, ok := Lookup(b18WoodedBastionOracle)
	if !ok {
		t.Fatal("Wooded Bastion not registered")
	}
	if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Fatalf("want the painless {C} first, got %+v", spec.ManaAbilities)
	}
	filter := spec.ManaAbilities[1]
	if filter.Cost.Mana != "{G/W}" || filter.Produced != "{G|W}{G|W}" || !filter.IgnoreCommanderIdentity {
		t.Errorf("filter ability %+v, want {G/W} in and two {G|W} picks out", filter)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	bastion := b12Push(g, me.ID, "Wooded Bastion", "Land", b18WoodedBastionOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, bastion, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the filter needs {G/W} floating")
	}
	b06AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, bastion, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			picks++
			if err := g.ResolveManaChoice(c.ID, me.ID, "W"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if picks != 2 {
		t.Fatalf("want two colour picks, got %d", picks)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W] (the {G} was spent)", got)
	}
}

// --- the artifacts -------------------------------------------------

func TestB18DreamstoneHedronTapsForThreeAndCashesInForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hedron := b12Push(g, me.ID, "Dreamstone Hedron", "Artifact", b18DreamstoneHedronOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hedron, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 3 {
		t.Errorf("pool %v, want three colorless", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hedron) })
	me.ManaPool.EmptyPool()
	b06AddMana(me, "C", "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, hedron, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("drew %d, want 3", got-hand)
	}
	if g.Battlefield.Contains(hedron) || !me.Graveyard.Contains(hedron) {
		t.Error("the Hedron is sacrificed as the cost")
	}
}

func TestB18PerpetualTimepieceMillsTwoAndDeclaresTheShuffleGap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	piece := b12Push(g, me.ID, "Perpetual Timepiece", "Artifact", b18PerpetualTimepieceOracle, 0, 0)
	advanceToMain(t, g)
	grave, lib := me.Graveyard.Size(), me.Library.Size()
	b16Activate(t, g, me.ID, piece, 0, game.ActivateAbilityParams{})
	if me.Graveyard.Size() != grave+2 || me.Library.Size() != lib-2 {
		t.Errorf("graveyard %d → %d, library %d → %d: want two milled", grave, me.Graveyard.Size(), lib, me.Library.Size())
	}
	if !b16Tapped(t, g, piece) {
		t.Error("the Timepiece taps as its cost")
	}
	spec, _ := Lookup(b18PerpetualTimepieceOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Activated) != 1 {
		t.Error("the exile-to-shuffle gap must be declared, and only the mill ships")
	}
}

// --- the token makers ----------------------------------------------

func TestB18KrenkosCommandMakesTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Krenko's Command", "Sorcery", b18KrenkosCommandOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Goblin"); n != 2 {
		t.Errorf("%d Goblins, want 2", n)
	}
}

func TestB18SpringleafParadeMakesXChangelingsThatTapForAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	parade := b12PlayFromHand(t, g, "Springleaf Parade", "Enchantment", b18SpringleafParadeOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(parade) {
		t.Fatal("the Parade did not resolve")
	}
	if n := b16CountNamed(g, "Shapeshifter"); n != 3 {
		t.Fatalf("%d Shapeshifters, want X=3", n)
	}
	token := findBattlefieldByName(g, "Shapeshifter")
	if !hasEffectiveKeyword(t, g, token, game.KeywordChangeling) {
		t.Error("the tokens have changeling")
	}
	c, _ := battlefieldCard(g, token)
	if !c.HasSubtype("Goblin") || !c.HasSubtype("Elf") {
		t.Error("a changeling is every creature type")
	}
	// The mana ability: real while the Parade is here, gone when it
	// leaves. Summoning sickness applies to a creature's tap.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == token {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	if err := g.ActivateManaAbility(me.ID, token, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	picked := false
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == me.ID {
			if len(ch.ColorOptions) != 5 {
				t.Errorf("any colour is five options, got %v", ch.ColorOptions)
			}
			if err := g.ResolveManaChoice(ch.ID, me.ID, "B"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			picked = true
		}
	}
	if !picked {
		t.Fatal("the token's tap should ask for a colour")
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(parade) })
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(token) })
	if err := g.ActivateManaAbility(me.ID, token, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("without a Springleaf Parade the token has no mana ability")
	}
	spec, _ := Lookup(b18SpringleafParadeOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 2 {
		t.Error("the resolution-time tokens and the own-tokens-only grant must be declared")
	}
}

func TestB18BitterbloomBearerLosesOneAndMakesAFaerieEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bitterbloom Bearer", "Creature — Faerie Rogue", b18BitterbloomBearerOracle, 1, 1)
	life := me.Life
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if n := b16CountNamed(g, "Faerie"); n != 1 {
		t.Fatalf("%d Faeries, want 1", n)
	}
	faerie := findBattlefieldByName(g, "Faerie")
	if !hasEffectiveKeyword(t, g, faerie, "flying") {
		t.Error("the Faerie has flying")
	}
	if c, _ := battlefieldCard(g, faerie); !c.HasColor("U") || !c.HasColor("B") {
		t.Error("the Faerie is blue and black")
	}
}

func TestB18EmeriaAngelMayMakeABirdOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Emeria Angel", "Creature — Angel", b18EmeriaAngelOracle, 3, 3)
	playLandFromHand(t, g, "Plains", "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	bird := findBattlefieldByName(g, "Bird")
	if bird == uuid.Nil {
		t.Fatal("no Bird")
	}
	if !hasEffectiveKeyword(t, g, bird, "flying") || effectivePower(t, g, bird) != 1 {
		t.Error("a 1/1 Bird with flying")
	}
}

func TestB18OgreSlumlordMakesDeathtouchRatsWhenNontokenCreaturesDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slumlord := b12Push(g, me.ID, "Ogre Slumlord", "Creature — Ogre Rogue", b18OgreSlumlordOracle, 3, 3)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	b18Kill(t, g, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	rat := findBattlefieldByName(g, "Rat")
	if rat == uuid.Nil {
		t.Fatal("no Rat")
	}
	if !hasEffectiveKeyword(t, g, rat, "deathtouch") {
		t.Error("Rats you control have deathtouch")
	}
	if c, _ := battlefieldCard(g, rat); !c.HasColor("B") {
		t.Error("the Rat is black")
	}
	// A token dying does not trigger it; the Slumlord's own death
	// does not either.
	b18Kill(t, g, rat)
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a token is not a nontoken creature")
	}
	b18Kill(t, g, slumlord)
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("'another' excludes the Slumlord itself")
	}
}

// --- the draw spells -----------------------------------------------

func TestB18HelpfulHunterDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Helpful Hunter", "Creature — Cat", b18HelpfulHunterOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: the Hunter was added and cast (net 0) and drew (+1)", hand, got)
	}
}

func TestB18SpectralSailorDrawsForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sailor := pushCatalogPermanent(g, me.ID, "Spectral Sailor", "Creature — Spirit Pirate", b18SpectralSailorOracle, true)
	if !hasEffectiveKeyword(t, g, sailor, "flying") || !hasEffectiveKeyword(t, g, sailor, "flash") {
		t.Error("flash and flying are printed")
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sailor, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{U} must be paid")
	}
	b06AddMana(me, "U", "C", "C", "C")
	hand := me.Hand.Size()
	// No tap in the cost: a summoning-sick Sailor can still draw.
	b16Activate(t, g, me.ID, sailor, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB18EurekaMomentDrawsTwoAndDeclaresTheLandGap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Eureka Moment", "Instant", b18EurekaMomentOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew two (+2)", hand, got)
	}
	if spec, _ := Lookup(b18EurekaMomentOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the put-a-land gap must be declared")
	}
}

func TestB18ThirstForKnowledgeDrawsThreeThenOwesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Thirst for Knowledge", "Instant", b18ThirstForKnowledgeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew three (+3)", hand, got)
	}
	if g.DiscardPending[me.ID] != 2 {
		t.Errorf("discard owed = %d, want 2", g.DiscardPending[me.ID])
	}
	if spec, _ := Lookup(b18ThirstForKnowledgeOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the artifact-discard gap must be declared")
	}
}

func TestB18ChartACourseDiscardsUnlessYouAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Chart a Course", "Sorcery", b18ChartACourseOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew two (+2)", hand, got)
	}
	if g.DiscardPending[me.ID] != 1 {
		t.Fatalf("no attack yet: discard owed = %d, want 1", g.DiscardPending[me.ID])
	}
	delete(g.DiscardPending, me.ID)

	attackWith(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepPostcombatMain)
	castCatalogSpell(t, g, "Chart a Course", "Sorcery", b18ChartACourseOracle, nil)
	passPriorityAroundTable(t, g)
	if g.DiscardPending[me.ID] != 0 {
		t.Errorf("attacked this turn: discard owed = %d, want 0", g.DiscardPending[me.ID])
	}
	// A new turn forgets the attack.
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	castCatalogSpell(t, g, "Chart a Course", "Sorcery", b18ChartACourseOracle, nil)
	passPriorityAroundTable(t, g)
	if g.DiscardPending[me.ID] != 1 {
		t.Errorf("next turn: discard owed = %d, want 1", g.DiscardPending[me.ID])
	}
}

func TestB18ManyPartingsFetchesABasicToHandAndMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Reliquary Tower", "Land"),
	)
	castCatalogSpell(t, g, "Many Partings", "Sorcery", b18ManyPartingsOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics for one slot: the searcher chooses")
	}
	if searchOptionNamed(g, c, "Reliquary Tower") != uuid.Nil {
		t.Error("a nonbasic is not offered")
	}
	if b16CountNamed(g, "Food") != 0 {
		t.Error("the Food waits for the search")
	}
	answerSearchNamed(t, g, me.ID, "Island")
	found := false
	for _, h := range me.Hand.Cards {
		if h.Name == "Island" {
			found = true
		}
	}
	if !found {
		t.Error("the Island goes to hand")
	}
	if b16CountNamed(g, "Food") != 1 {
		t.Error("then a Food is created")
	}
}

func TestB18CycleOfRenewalSacrificesALandAndFetchesTwoTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12Permanent(g, me.ID, "Old Forest", "Basic Land — Forest")
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
	)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cycle of Renewal", TypeLine: "Instant — Lesson",
		OracleID: b18CycleOfRenewalOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Fatal("the land sacrifice is an additional cost")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed on cast")
	}
	passPriorityAroundTable(t, g)
	for _, name := range []string{"Forest", "Plains"} {
		l := findBattlefieldByName(g, name)
		if l == uuid.Nil {
			t.Fatalf("%s did not reach the battlefield", name)
		}
		if !b16Tapped(t, g, l) {
			t.Errorf("%s enters tapped", name)
		}
	}
	if spec, _ := Lookup(b18CycleOfRenewalOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the cost-not-effect posture must be declared")
	}
}

// --- removal -------------------------------------------------------

func TestB18HerosDownfallDestroysACreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Jace",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"loyalty": 3},
	})
	rock := b12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	advanceToMain(t, g)
	id := handCardFull(g.Seats[0], "Hero's Downfall", "Instant", "", b18HerosDownfallOracle, []string{"B"})
	if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Targets: b16TargetCard(rock)}); err == nil {
		t.Fatal("an artifact is not a legal target")
	}
	castCatalogSpell(t, g, "Hero's Downfall", "Instant", b18HerosDownfallOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is destroyed")
	}
	castCatalogSpell(t, g, "Hero's Downfall", "Instant", b18HerosDownfallOracle, b16TargetCard(walker))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed")
	}
}

func TestB18SplitUpDestroysTheTappedOrTheUntappedHalf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tappedMine := b16Creature(g, me.ID, "Tapped Mine", "Creature — Bear", 2, 2, "G")
	untappedMine := b16Creature(g, me.ID, "Untapped Mine", "Creature — Bear", 2, 2, "G")
	tappedTheirs := b16Creature(g, opp.ID, "Tapped Theirs", "Creature — Bear", 2, 2, "G")
	untappedTheirs := b16Creature(g, opp.ID, "Untapped Theirs", "Creature — Bear", 2, 2, "G")
	b16Tap(g, tappedMine)
	b16Tap(g, tappedTheirs)
	b12PlayFromHand(t, g, "Split Up", "Sorcery", b18SplitUpOracle, game.CastSpellParams{Modes: []int{0}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tappedMine) || g.Battlefield.Contains(tappedTheirs) {
		t.Error("mode 0 destroys every tapped creature")
	}
	if !g.Battlefield.Contains(untappedMine) || !g.Battlefield.Contains(untappedTheirs) {
		t.Error("mode 0 leaves the untapped ones")
	}
	b12PlayFromHand(t, g, "Split Up", "Sorcery", b18SplitUpOracle, game.CastSpellParams{Modes: []int{1}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(untappedMine) || g.Battlefield.Contains(untappedTheirs) {
		t.Error("mode 1 destroys every untapped creature")
	}
	// S30 (#470 / #446): the boardwipe indestructible gap this used
	// to require a declaration for is closed, so the card is Full.
	if spec, _ := Lookup(b18SplitUpOracle); spec.Completeness != CompletenessFull {
		t.Error("Split Up implements every clause it prints")
	}
}

func TestB18WindsOfAbandonExilesAndTheControllerFetchesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(opp,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
	)
	advanceToMain(t, g)
	id := handCardFull(me, "Winds of Abandon", "Sorcery", "", b18WindsOfAbandonOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(mine)}); err == nil {
		t.Fatal("your own creature is not a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(theirs)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("the target is exiled")
	}
	c := searchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the creature's controller searches")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("the caster does not search")
	}
	answerSearchNamed(t, g, opp.ID, "Forest")
	forest := findBattlefieldByName(g, "Forest")
	if forest == uuid.Nil || !b16Tapped(t, g, forest) {
		t.Error("the basic enters tapped under its controller")
	}
	if fc, _ := battlefieldCard(g, forest); fc.Controller != opp.ID {
		t.Error("the land is the searching player's")
	}
}

func TestB18WindsOfAbandonOverloadedExilesEachCreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	a := b16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	c := b16Creature(g, other.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	seedSearchLibrary(opp,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	seedSearchLibrary(other, searchTestLand("Swamp", "Basic Land — Swamp"))
	advanceToMain(t, g)
	id := handCardFull(me, "Winds of Abandon", "Sorcery", "", b18WindsOfAbandonOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overloaded cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(x) {
			t.Errorf("%s survived the overload", x)
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creatures are untouched")
	}
	// The opponent who lost two creatures runs one two-card search;
	// the one who lost one has exactly one basic and needs no prompt.
	pc := searchChoiceFor(g, opp.ID)
	if pc == nil || pc.Count != 2 {
		t.Fatalf("opp's search = %+v, want a two-card prompt", pc)
	}
	answerSearchNamed(t, g, opp.ID, "Forest", "Plains")
	if findBattlefieldByName(g, "Forest") == uuid.Nil || findBattlefieldByName(g, "Plains") == uuid.Nil {
		t.Error("opp's two basics arrive")
	}
	swamp := findBattlefieldByName(g, "Swamp")
	if swamp == uuid.Nil || !b16Tapped(t, g, swamp) {
		t.Error("the other player's single basic arrives tapped without a prompt")
	}
}

func TestB18VenserBouncesATargetPermanentAndDeclaresTheSpellGap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castCatalogSpell(t, g, "Venser, Shaper Savant", "Legendary Creature — Human Wizard", b18VenserShaperSavantOracle, nil)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the target permanent returns to its owner's hand")
	}
	spec, _ := Lookup(b18VenserShaperSavantOracle)
	if spec.Completeness != CompletenessCaveats {
		t.Error("the spell-half gap must be declared")
	}
	if len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil || len(spec.Triggered[0].Targets.Zones) != 1 || spec.Triggered[0].Targets.Zones[0] != game.ZoneBattlefield {
		t.Error("the target clause offers the battlefield only")
	}
}

// --- the creatures with triggers -----------------------------------

func TestB18StormfistCrusaderEachPlayerDrawsAndLosesOneOnYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Stormfist Crusader", "Creature — Human Knight", b18StormfistCrusaderOracle, 2, 2)
	crusader := findBattlefieldByName(g, "Stormfist Crusader")
	advanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers) != 0 || triggerOnStack(g, crusader) != nil {
		t.Error("an opponent's upkeep is not yours")
	}
	advanceToUpkeepOf(t, g, 0)
	myLife, oppLife := me.Life, opp.Life
	myHand, oppHand := me.Hand.Size(), opp.Hand.Size()
	if len(g.PendingTriggers) == 0 && triggerOnStack(g, crusader) == nil {
		t.Fatal("your upkeep triggers it")
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life me %d → %d, opp %d → %d: want -1 each", myLife, me.Life, oppLife, opp.Life)
	}
	if me.Hand.Size() != myHand+1 || opp.Hand.Size() != oppHand+1 {
		t.Errorf("hands me %d → %d, opp %d → %d: want +1 each", myHand, me.Hand.Size(), oppHand, opp.Hand.Size())
	}
}

func TestB18FalkenrathNobleDrainsATargetPlayerPerDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	noble := b12Push(g, me.ID, "Falkenrath Noble", "Creature — Vampire Noble", b18FalkenrathNobleOracle, 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("opp %d → %d, me %d → %d: want -1 / +1", oppLife, opp.Life, myLife, me.Life)
	}
	// Its own death triggers it too.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(noble) })
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the Noble's own death drains: opp %d, want %d", opp.Life, oppLife-2)
	}
}

func TestB18SangromancerMayGainThreeOnOpponentDeathsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sangromancer", "Creature — Vampire Shaman", b18SangromancerOracle, 3, 3)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	life := me.Life
	b18Kill(t, g, mine)
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("your own creature dying is not an opponent's")
	}
	b18Kill(t, g, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 for the death", life, me.Life)
	}
	handCardFull(opp, "Junk", "Sorcery", "", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("declined: life %d, want %d", me.Life, life+3)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("accepted: life %d, want %d", me.Life, life+6)
	}
}

func TestB18ObNixilisTheFallenDrainsATargetAndGrowsOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := b12Push(g, me.ID, "Ob Nixilis, the Fallen", "Legendary Creature — Demon", b18ObNixilisTheFallenOracle, 3, 3)
	life := opp.Life
	playLandFromHand(t, g, "Swamp", "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opp %d → %d, want -3", life, opp.Life)
	}
	if b18Counters(t, g, ob, "+1/+1") != 3 {
		t.Errorf("Ob Nixilis has %d +1/+1 counters, want 3", b18Counters(t, g, ob, "+1/+1"))
	}
	if c, _ := battlefieldCard(g, ob); c.CurrentPower() != 6 {
		t.Errorf("Ob Nixilis is a 3/3 with three counters, power %d, want 6", c.CurrentPower())
	}
	// Declining leaves both halves undone.
	life = opp.Life
	playLandFromHand(t, g, "Swamp Two", "")
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if opp.Life != life || b18Counters(t, g, ob, "+1/+1") != 3 {
		t.Error("declined: no drain, no counters")
	}
}

func TestB18KederektParasitePingsADrawingOpponentWhileYouControlRed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kederekt Parasite", "Creature — Horror", b18KederektParasiteOracle, 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("no red permanent, no trigger")
	}
	red := b16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	if n := b18TriggerPromptCount(g, me.ID); n != 2 {
		t.Fatalf("two cards drawn is two prompts, got %d", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opp %d → %d, want -2", life, opp.Life)
	}
	// Your own draw never triggers it.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("your own draw is not an opponent's")
	}
	// The intervening if is checked again on resolution.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(red) })
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("red permanent gone in response: opp %d, want %d", opp.Life, life-2)
	}
}

func TestB18GarruksPackleaderMayDrawWhenABigCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Garruk's Packleader", "Creature — Beast", b18GarruksPackleaderOracle, 4, 4)
	top100CastCreature(t, g, me, "Small Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("power 2 is not 3 or greater")
	}
	hand := me.Hand.Size()
	top100CastCreature(t, g, me, "Big Beast", 3, 3)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew (+1)", hand, got)
	}
}

func TestB18JunkDiverReturnsAnotherArtifactCardWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diver := b12Push(g, me.ID, "Junk Diver", "Artifact Creature — Bird", b18JunkDiverOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, diver, "flying") {
		t.Error("flying is printed")
	}
	relic := batch01GraveyardCard(me, "Sol Ring", "Artifact")
	bear := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(diver) })
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, bear) || !hasID(p.PickTargetCards, relic) {
		t.Error("artifact cards in your graveyard are offered, creature cards are not")
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card returns to hand")
	}
	if spec, _ := Lookup(b18JunkDiverOracle); spec.Completeness != CompletenessCaveats {
		t.Error("the self-pick gap must be declared")
	}
}

func TestB18RighteousValkyrieGainsToughnessForAngelsAndClerics(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", b18RighteousValkyrieOracle, 2, 4)
	life := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Serra Angel", TypeLine: "Creature — Angel", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4 (the Angel's toughness)", life, me.Life)
	}
	top100CastCreature(t, g, me, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("a Beast is neither: life %d, want %d", me.Life, life+4)
	}
	// An opponent's Cleric is not yours.
	advanceToMainOf(t, g, 1)
	cid := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: cid, Name: "Their Cleric", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, cid, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("an opponent's Cleric: life %d, want %d", me.Life, life+4)
	}
	spec, _ := Lookup(b18RighteousValkyrieOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Static) != 0 {
		t.Error("the life-threshold anthem gap must be declared, and no static ships")
	}
}

func TestB18KorSpiritdancerGrowsWithAurasAndDrawsOnAuraCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dancer := b12Push(g, me.ID, "Kor Spiritdancer", "Creature — Kor Wizard", b18KorSpiritdancerOracle, 0, 2)
	aura := b12Permanent(g, me.ID, "Rancor", "Enchantment — Aura")
	equip := b12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	if effectivePower(t, g, dancer) != 0 || effectiveToughness(t, g, dancer) != 2 {
		t.Fatal("unenchanted, a 0/2")
	}
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{aura, equip} {
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: dancer}); err != nil {
				t.Fatalf("AttachForEffect: %v", err)
			}
		}
	})
	if effectivePower(t, g, dancer) != 2 || effectiveToughness(t, g, dancer) != 4 {
		t.Errorf("one Aura and one Equipment = %d/%d, want 2/4", effectivePower(t, g, dancer), effectiveToughness(t, g, dancer))
	}
	hand := me.Hand.Size()
	b18CastFromHand(t, g, me, "Pacifism", "Enchantment — Aura", "", game.CastSpellParams{})
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: the Aura was added and cast (net 0), drew (+1)", hand, got)
	}
	b18CastFromHand(t, g, me, "Plain Enchantment", "Enchantment", "", game.CastSpellParams{})
	if b18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a non-Aura enchantment is not an Aura spell")
	}
}

func TestB18ThunderfootBalothIsALieutenant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baloth := b12Push(g, me.ID, "Thunderfoot Baloth", "Creature — Beast", b18ThunderfootBalothOracle, 5, 5)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if effectivePower(t, g, baloth) != 5 || effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("without your commander, nothing")
	}
	commander := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if effectivePower(t, g, baloth) != 7 || effectiveToughness(t, g, baloth) != 7 {
		t.Errorf("the Baloth = %d/%d, want 7/7", effectivePower(t, g, baloth), effectiveToughness(t, g, baloth))
	}
	if !hasEffectiveKeyword(t, g, baloth, "trample") {
		t.Error("the Baloth keeps its printed trample")
	}
	if effectivePower(t, g, bear) != 4 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("other creatures you control get +2/+2 and trample")
	}
	if effectivePower(t, g, commander) != 5 || !hasEffectiveKeyword(t, g, commander, "trample") {
		t.Error("the commander is another creature you control")
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature is untouched")
	}
}

func TestB18ThunderfootBalothIgnoresAStolenCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baloth := b12Push(g, me.ID, "Thunderfoot Baloth", "Creature — Beast", b18ThunderfootBalothOracle, 5, 5)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: me.ID, IsCommander: true,
	})
	if effectivePower(t, g, baloth) != 5 {
		t.Error("an opponent's commander you control is not YOUR commander")
	}
}

func TestB18WoundReflectionDoublesEachOpponentsLossAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Wound Reflection", "Enchantment", b18WoundReflectionOracle, 0, 0)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3, "G")
	oppLife, otherLife, myLife := opp.Life, other.Life, me.Life
	attackWith(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepPostcombatMain)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(other.ID))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5) })
	if opp.Life != oppLife-3 || other.Life != otherLife-3 {
		t.Fatalf("setup: opp %d other %d", opp.Life, other.Life)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-6 {
		t.Errorf("opp lost 3 in combat, loses 3 more: %d, want %d", opp.Life, oppLife-6)
	}
	if other.Life != otherLife-6 {
		t.Errorf("other lost 3 to the Bolt, loses 3 more: %d, want %d", other.Life, otherLife-6)
	}
	if me.Life != myLife-5 {
		t.Errorf("the controller's own loss is not reflected: %d, want %d", me.Life, myLife-5)
	}
	// Next turn's end step: nothing was lost, nothing is reflected.
	advanceToEndStepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-6 || other.Life != otherLife-6 {
		t.Error("a turn with no loss reflects nothing")
	}
}

func TestB18GrazilaxxMayBounceABlockedAttackerAndDrawsOncePerCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grazilaxx, Illithid Scholar", "Legendary Creature — Horror", b18GrazilaxxOracle, 3, 2)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	c := b16Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	wall1 := b16Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 4, "W")
	wall2 := b16Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 4, "W")
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall1, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := b18TriggerPromptCount(g, me.ID); n != 1 {
		t.Fatalf("one blocked attacker is one prompt, got %d", n)
	}
	if err := g.DeclareBlocker(wall2, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := b18TriggerPromptCount(g, me.ID); n != 1 {
		t.Fatalf("a second blocker on the same attacker is not a second 'becomes blocked', got %d prompts", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the blocked creature returns to its owner's hand")
	}
	hand := me.Hand.Size()
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("two unblocked bears connect: opp %d, want %d", opp.Life, life-4)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: one draw for one or more creatures connecting", hand, got)
	}
}

func TestB18MagdaMakesATappedTreasureOncePerTurnPerCrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magda := b12Push(g, me.ID, "Magda, the Hoardmaster", "Legendary Creature — Dwarf Berserker", b18MagdaTheHoardmasterOrcl, 2, 2)
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	advanceToMain(t, g)
	// Targeting your own creature is not a crime.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 0 {
		t.Fatal("targeting your own creature is not a crime")
	}
	// Targeting an opponent is; the trigger resolves above the Bolt.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	if triggerOnStack(g, magda) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the crime triggers at announce")
	}
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 1 {
		t.Fatalf("%d Treasures, want 1", b16CountNamed(g, "Treasure"))
	}
	if !b16Tapped(t, g, findBattlefieldByName(g, "Treasure")) {
		t.Error("the Treasure enters tapped")
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 1 {
		t.Error("only once each turn")
	}
	advanceToMainOf(t, g, 1)
	advanceToMainOf(t, g, 0)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	if b16CountNamed(g, "Treasure") != 2 {
		t.Error("a new turn, a new Treasure")
	}
	if spec, _ := Lookup(b18MagdaTheHoardmasterOrcl); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 0 {
		t.Error("the Scorpion Dragon gap must be declared, and no activated ability ships")
	}
}

func TestB18MikaeusEntersWithXAndGrowsByTapping(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mik := b12PlayFromHand(t, g, "Mikaeus, the Lunarch", "Legendary Creature — Human Cleric", b18MikaeusTheLunarchOracle, game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mik) {
		t.Fatal("Mikaeus did not resolve")
	}
	if b18Counters(t, g, mik, "+1/+1") != 2 {
		t.Fatalf("%d counters, want X=2", b18Counters(t, g, mik, "+1/+1"))
	}
	if err := g.ActivateCatalogAbility(me.ID, mik, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Mikaeus cannot tap")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mik {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	b16Activate(t, g, me.ID, mik, 0, game.ActivateAbilityParams{})
	if b18Counters(t, g, mik, "+1/+1") != 3 {
		t.Errorf("%d counters after the tap, want 3", b18Counters(t, g, mik, "+1/+1"))
	}
	if spec, _ := Lookup(b18MikaeusTheLunarchOracle); spec.Completeness != CompletenessCaveats || len(spec.Activated) != 1 {
		t.Error("the counter-removal gap must be declared, and only the self-grow ships")
	}
}

func TestB18ArixmethesSleepsAsALandAndWakesAfterFiveSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arix := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: arix, Name: "Arixmethes, Slumbering Isle", TypeLine: "Legendary Creature — Kraken",
		OracleID: b18ArixmethesOracle, Power: 12, Toughness: 12, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, arix, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(arix) {
		t.Fatal("Arixmethes did not resolve")
	}
	if !b16Tapped(t, g, arix) || tapEventsFor(g, arix) != 0 {
		t.Error("enters tapped, as a replacement")
	}
	if b18Counters(t, g, arix, "slumber") != 5 {
		t.Fatalf("%d slumber counters, want 5", b18Counters(t, g, arix, "slumber"))
	}
	c, _ := battlefieldCard(g, arix)
	if c.IsCreature() || !c.IsLand() || c.HasSubtype("Kraken") {
		t.Errorf("asleep it is a land and not a creature: types %v subtypes %v", effectiveTypes(t, g, arix), c.Effective().Subtypes)
	}
	if !c.IsLegendary() {
		t.Error("it keeps its supertype")
	}
	// The mana ability works untapped, with no summoning sickness —
	// it is not a creature.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(arix) })
	if err := g.ActivateManaAbility(me.ID, arix, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool %v, want [G U]", got)
	}
	// Each spell you cast may remove a counter; declining keeps it.
	castCatalogSpell(t, g, "Krenko's Command", "Sorcery", b18KrenkosCommandOracle, nil)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if b18Counters(t, g, arix, "slumber") != 5 {
		t.Error("declined: the counter stays")
	}
	for i := 0; i < 5; i++ {
		castCatalogSpell(t, g, "Krenko's Command", "Sorcery", b18KrenkosCommandOracle, nil)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
	}
	if b18Counters(t, g, arix, "slumber") != 0 {
		t.Fatalf("%d slumber counters after five removals, want 0", b18Counters(t, g, arix, "slumber"))
	}
	c, _ = battlefieldCard(g, arix)
	if !c.IsCreature() || c.IsLand() || !c.HasSubtype("Kraken") {
		t.Errorf("awake it is a Kraken creature: types %v", effectiveTypes(t, g, arix))
	}
	if effectivePower(t, g, arix) != 12 || effectiveToughness(t, g, arix) != 12 {
		t.Error("a 12/12")
	}
	// A sixth spell with no counters left: nothing to remove, no prompt.
	castCatalogSpell(t, g, "Krenko's Command", "Sorcery", b18KrenkosCommandOracle, nil)
	if b18TriggerPromptCount(g, me.ID) == 1 {
		answerLatestTriggerPrompt(t, g, me.ID, true)
	}
	passPriorityAroundTable(t, g)
	if b18Counters(t, g, arix, "slumber") != 0 {
		t.Error("never negative")
	}
}
