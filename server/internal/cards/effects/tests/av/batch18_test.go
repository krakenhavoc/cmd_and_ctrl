package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Two are rows
// in the Guildgate table and one is a filter-land one-liner, so a
// transposed row is invisible until someone plays that exact card.
func TestBatch18CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B18DreamstoneHedronOracle:   "Dreamstone Hedron",
		effectstest.B18SpringleafParadeOracle:   "Springleaf Parade",
		effectstest.B18StormfistCrusaderOracle:  "Stormfist Crusader",
		effectstest.B18ChartACourseOracle:       "Chart a Course",
		effectstest.B18WoundReflectionOracle:    "Wound Reflection",
		effectstest.B18SpectralSailorOracle:     "Spectral Sailor",
		effectstest.B18GrazilaxxOracle:          "Grazilaxx, Illithid Scholar",
		effectstest.B18KorSpiritdancerOracle:    "Kor Spiritdancer",
		effectstest.B18ManyPartingsOracle:       "Many Partings",
		effectstest.B18ObNixilisTheFallenOracle: "Ob Nixilis, the Fallen",
		effectstest.B18MikaeusTheLunarchOracle:  "Mikaeus, the Lunarch",
		effectstest.B18OgreSlumlordOracle:       "Ogre Slumlord",
		effectstest.B18HelpfulHunterOracle:      "Helpful Hunter",
		effectstest.B18OrzhovGuildgateOracle:    "Orzhov Guildgate",
		effectstest.B18JunkDiverOracle:          "Junk Diver",
		effectstest.B18VenserShaperSavantOracle: "Venser, Shaper Savant",
		effectstest.B18SplitUpOracle:            "Split Up",
		effectstest.B18HerosDownfallOracle:      "Hero's Downfall",
		effectstest.B18CycleOfRenewalOracle:     "Cycle of Renewal",
		effectstest.B18EurekaMomentOracle:       "Eureka Moment",
		effectstest.B18FalkenrathNobleOracle:    "Falkenrath Noble",
		effectstest.B18BitterbloomBearerOracle:  "Bitterbloom Bearer",
		effectstest.B18EmeriaAngelOracle:        "Emeria Angel",
		effectstest.B18KederektParasiteOracle:   "Kederekt Parasite",
		effectstest.B18WindsOfAbandonOracle:     "Winds of Abandon",
		effectstest.B18MagdaTheHoardmasterOrcl:  "Magda, the Hoardmaster",
		effectstest.B18SangromancerOracle:       "Sangromancer",
		effectstest.B18ThunderfootBalothOracle:  "Thunderfoot Baloth",
		effectstest.B18ArixmethesOracle:         "Arixmethes, Slumbering Isle",
		effectstest.B18RakdosGuildgateOracle:    "Rakdos Guildgate",
		effectstest.B18GarruksPackleaderOracle:  "Garruk's Packleader",
		effectstest.B18PerpetualTimepieceOracle: "Perpetual Timepiece",
		effectstest.B18WoodedBastionOracle:      "Wooded Bastion",
		effectstest.B18ThirstForKnowledgeOracle: "Thirst for Knowledge",
		effectstest.B18RighteousValkyrieOracle:  "Righteous Valkyrie",
		effectstest.B18KrenkosCommandOracle:     "Krenko's Command",
	}
	if len(want) != 36 {
		t.Fatalf("the batch registers 36 cards, the table lists %d", len(want))
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

func TestB18GuildgateRowsEnterTappedAndProduceTheirColours(t *testing.T) {
	for _, row := range []struct{ oracle, name, produced string }{
		{effectstest.B18OrzhovGuildgateOracle, "Orzhov Guildgate", "{W|B}"},
		{effectstest.B18RakdosGuildgateOracle, "Rakdos Guildgate", "{B|R}"},
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
	g := effectstest.NewCatalogGame(t)
	gate := effectstest.PlayLandFromHand(t, g, "Orzhov Guildgate", effectstest.B18OrzhovGuildgateOracle)
	effectstest.Top100AssertEnteredTapped(t, g, gate, "Orzhov Guildgate")
	if effectstest.TapEventsFor(g, gate) != 0 {
		t.Error("enters-tapped is a replacement, not a tap")
	}
}

func TestB18WoodedBastionFiltersHybridIntoTwoPicks(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B18WoodedBastionOracle)
	if !ok {
		t.Fatal("Wooded Bastion not registered")
	}
	if len(spec.ManaAbilities) != 2 || spec.ManaAbilities[0].Produced != "{C}" {
		t.Fatalf("want the painless {C} first, got %+v", spec.ManaAbilities)
	}
	filter := spec.ManaAbilities[1]
	if filter.Cost.Mana != "{G/W}" || filter.Produced != "{G|W}{G|W}" || filter.NarrowToCommanderIdentity {
		t.Errorf("filter ability %+v, want {G/W} in and two {G|W} picks out", filter)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bastion := effectstest.B12Push(g, me.ID, "Wooded Bastion", "Land", effectstest.B18WoodedBastionOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, bastion, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the filter needs {G/W} floating")
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateManaAbility(me.ID, bastion, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// #3939: one choice, asked once per pick.
	picks := 0
	for c := effectstest.RiderLatestManaPick(g, me.ID); c != nil && picks < 8; c = effectstest.RiderLatestManaPick(g, me.ID) {
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "W" || got[1] != "W" {
		t.Errorf("pool %v, want [W W] (the {G} was spent)", got)
	}
}

// --- the artifacts -------------------------------------------------

func TestB18DreamstoneHedronTapsForThreeAndCashesInForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hedron := effectstest.B12Push(g, me.ID, "Dreamstone Hedron", "Artifact", effectstest.B18DreamstoneHedronOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hedron, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 {
		t.Errorf("pool %v, want three colorless", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(hedron) })
	me.ManaPool.EmptyPool()
	effectstest.B06AddMana(me, "C", "C", "C")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, hedron, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("drew %d, want 3", got-hand)
	}
	if g.Battlefield.Contains(hedron) || !me.Graveyard.Contains(hedron) {
		t.Error("the Hedron is sacrificed as the cost")
	}
}

// The mill half. The exile-to-shuffle half landed with #1404 and is
// tested in exile_this_permanent_test.go.
func TestB18PerpetualTimepieceMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	piece := effectstest.B12Push(g, me.ID, "Perpetual Timepiece", "Artifact", effectstest.B18PerpetualTimepieceOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	grave, lib := me.Graveyard.Size(), me.Library.Size()
	effectstest.B16Activate(t, g, me.ID, piece, 0, game.ActivateAbilityParams{})
	if me.Graveyard.Size() != grave+2 || me.Library.Size() != lib-2 {
		t.Errorf("graveyard %d → %d, library %d → %d: want two milled", grave, me.Graveyard.Size(), lib, me.Library.Size())
	}
	if !effectstest.B16Tapped(t, g, piece) {
		t.Error("the Timepiece taps as its cost")
	}
}

// --- the token makers ----------------------------------------------

func TestB18KrenkosCommandMakesTwoGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Krenko's Command", "Sorcery", effectstest.B18KrenkosCommandOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Goblin"); n != 2 {
		t.Errorf("%d Goblins, want 2", n)
	}
}

func TestB18SpringleafParadeMakesXChangelingsThatTapForAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	parade := effectstest.B12PlayFromHand(t, g, "Springleaf Parade", "Enchantment", effectstest.B18SpringleafParadeOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(parade) {
		t.Fatal("the Parade did not resolve")
	}
	if n := effectstest.B16CountNamed(g, "Shapeshifter"); n != 3 {
		t.Fatalf("%d Shapeshifters, want X=3", n)
	}
	token := effectstest.FindBattlefieldByName(g, "Shapeshifter")
	if !effectstest.HasEffectiveKeyword(t, g, token, game.KeywordChangeling) {
		t.Error("the tokens have changeling")
	}
	c, _ := effectstest.BattlefieldCard(g, token)
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(parade) })
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(token) })
	if err := g.ActivateManaAbility(me.ID, token, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("without a Springleaf Parade the token has no mana ability")
	}
	// #1357 made the ETB a real trigger reading CastX(), and ADR 0093
	// made the grant a real layer-6 grant to EVERY creature token you
	// control, so nothing is declared.
	spec, _ := effects.Lookup(effectstest.B18SpringleafParadeOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Springleaf Parade has no simplification left — CompletenessFull")
	}
}

// ADR 0093: the grant reaches a creature token the Parade did not make.
func TestB18SpringleafParadeGrantsEveryCreatureTokenYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	parade := game.NewCard("Springleaf Parade", me.ID)
	parade.TypeLine = "Enchantment"
	parade.OracleID = effectstest.B18SpringleafParadeOracle
	effectstest.PushBattlefieldCardWithTimestamp(g, parade)
	goblin := effects.TokenCard("1/1 red Goblin")
	goblin.InstanceID, goblin.Owner, goblin.Controller = uuid.New(), me.ID, me.ID
	effectstest.PushBattlefieldCardWithTimestamp(g, goblin)
	theirs := effects.TokenCard("1/1 red Goblin")
	theirs.InstanceID, theirs.Owner, theirs.Controller = uuid.New(), opp.ID, opp.ID
	effectstest.PushBattlefieldCardWithTimestamp(g, theirs)
	if n := effectstest.GrantedManaRowsOf(g, goblin.InstanceID); n != 1 {
		t.Errorf("your Goblin token has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs.InstanceID); n != 0 {
		t.Errorf("an opponent's Goblin token has %d granted mana rows, want 0", n)
	}
}

func TestB18BitterbloomBearerLosesOneAndMakesAFaerieEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bitterbloom Bearer", "Creature — Faerie Rogue", effectstest.B18BitterbloomBearerOracle, 1, 1)
	life := me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, me.Life)
	}
	if n := effectstest.B16CountNamed(g, "Faerie"); n != 1 {
		t.Fatalf("%d Faeries, want 1", n)
	}
	faerie := effectstest.FindBattlefieldByName(g, "Faerie")
	if !effectstest.HasEffectiveKeyword(t, g, faerie, "flying") {
		t.Error("the Faerie has flying")
	}
	if c, _ := effectstest.BattlefieldCard(g, faerie); !c.HasColor("U") || !c.HasColor("B") {
		t.Error("the Faerie is blue and black")
	}
}

func TestB18EmeriaAngelMayMakeABirdOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Emeria Angel", "Creature — Angel", effectstest.B18EmeriaAngelOracle, 3, 3)
	effectstest.PlayLandFromHand(t, g, "Plains", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if bird == uuid.Nil {
		t.Fatal("no Bird")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") || effectstest.EffectivePower(t, g, bird) != 1 {
		t.Error("a 1/1 Bird with flying")
	}
}

func TestB18OgreSlumlordMakesDeathtouchRatsWhenNontokenCreaturesDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slumlord := effectstest.B12Push(g, me.ID, "Ogre Slumlord", "Creature — Ogre Rogue", effectstest.B18OgreSlumlordOracle, 3, 3)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	rat := effectstest.FindBattlefieldByName(g, "Rat")
	if rat == uuid.Nil {
		t.Fatal("no Rat")
	}
	if !effectstest.HasEffectiveKeyword(t, g, rat, "deathtouch") {
		t.Error("Rats you control have deathtouch")
	}
	if c, _ := effectstest.BattlefieldCard(g, rat); !c.HasColor("B") {
		t.Error("the Rat is black")
	}
	// A token dying does not trigger it; the Slumlord's own death
	// does not either.
	effectstest.B18Kill(t, g, rat)
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a token is not a nontoken creature")
	}
	effectstest.B18Kill(t, g, slumlord)
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("'another' excludes the Slumlord itself")
	}
}

// --- the draw spells -----------------------------------------------

func TestB18HelpfulHunterDrawsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Helpful Hunter", "Creature — Cat", effectstest.B18HelpfulHunterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: the Hunter was added and cast (net 0) and drew (+1)", hand, got)
	}
}

func TestB18SpectralSailorDrawsForFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sailor := effectstest.PushCatalogPermanent(g, me.ID, "Spectral Sailor", "Creature — Spirit Pirate", effectstest.B18SpectralSailorOracle, true)
	if !effectstest.HasEffectiveKeyword(t, g, sailor, "flying") || !effectstest.HasEffectiveKeyword(t, g, sailor, "flash") {
		t.Error("flash and flying are printed")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sailor, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{3}{U} must be paid")
	}
	effectstest.B06AddMana(me, "U", "C", "C", "C")
	hand := me.Hand.Size()
	// No tap in the cost: a summoning-sick Sailor can still draw.
	effectstest.B16Activate(t, g, me.ID, sailor, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
}

func TestB18EurekaMomentDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Eureka Moment", "Instant", effectstest.B18EurekaMomentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew two (+2)", hand, got)
	}
	if spec, _ := effects.Lookup(effectstest.B18EurekaMomentOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the put-a-land clause landed with #654; nothing is deferred any more")
	}
}

func TestB18ThirstForKnowledgeDrawsThreeThenOwesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Thirst for Knowledge", "Instant", effectstest.B18ThirstForKnowledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew three (+3)", hand, got)
	}
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Errorf("discard owed = %d, want 2", effectstest.DiscardOwed(g, me.ID))
	}
	if spec, _ := effects.Lookup(effectstest.B18ThirstForKnowledgeOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the artifact-discard alternative is offered: nothing is left to declare")
	}
}

// "Discard two cards unless you discard an artifact card": one artifact
// card is a complete answer, one other card is not, and two of anything
// is.
func TestB18ThirstForKnowledgeAcceptsOneArtifactOrTwoCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.HandCardFull(me, "Sol Ring", "Artifact", "{1}", "", nil)
	bear := effectstest.HandCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", nil)
	other := effectstest.HandCardFull(me, "Bolt", "Instant", "{R}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Thirst for Knowledge", "Instant", effectstest.B18ThirstForKnowledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 1 || c.ChooseMax != 2 {
		t.Fatalf("a floor of one and a ceiling of two: %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("one non-artifact card was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("one artifact card is a complete discard: %v", err)
	}
	if me.Hand.Contains(rock) || !me.Hand.Contains(bear) || !me.Hand.Contains(other) {
		t.Error("only the artifact card was discarded")
	}
}

func TestB18ThirstForKnowledgeAcceptsTwoOrdinaryCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.HandCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", nil)
	other := effectstest.HandCardFull(me, "Bolt", "Instant", "{R}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Thirst for Knowledge", "Instant", effectstest.B18ThirstForKnowledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, bear, other)
	if me.Hand.Contains(bear) || me.Hand.Contains(other) {
		t.Error("two cards of any kind are a complete discard")
	}
}

func TestB18ChartACourseDiscardsUnlessYouAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Chart a Course", "Sorcery", effectstest.B18ChartACourseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew two (+2)", hand, got)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("no attack yet: discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)

	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.CastCatalogSpell(t, g, "Chart a Course", "Sorcery", effectstest.B18ChartACourseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Errorf("attacked this turn: discard owed = %d, want 0", effectstest.DiscardOwed(g, me.ID))
	}
	// A new turn forgets the attack.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.CastCatalogSpell(t, g, "Chart a Course", "Sorcery", effectstest.B18ChartACourseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("next turn: discard owed = %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
}

func TestB18ManyPartingsFetchesABasicToHandAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Reliquary Tower", "Land"),
	)
	effectstest.CastCatalogSpell(t, g, "Many Partings", "Sorcery", effectstest.B18ManyPartingsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two basics for one slot: the searcher chooses")
	}
	if effectstest.SearchOptionNamed(g, c, "Reliquary Tower") != uuid.Nil {
		t.Error("a nonbasic is not offered")
	}
	if effectstest.B16CountNamed(g, "Food") != 0 {
		t.Error("the Food waits for the search")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")
	found := false
	for _, h := range me.Hand.Cards {
		if h.Name == "Island" {
			found = true
		}
	}
	if !found {
		t.Error("the Island goes to hand")
	}
	if effectstest.B16CountNamed(g, "Food") != 1 {
		t.Error("then a Food is created")
	}
}

func TestB18CycleOfRenewalSacrificesALandOnResolutionAndFetchesTwoTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Old Forest", "Basic Land — Forest")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
	)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Cycle of Renewal", TypeLine: "Instant — Lesson",
		OracleID: effectstest.B18CycleOfRenewalOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("nothing is sacrificed at cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the land is sacrificed as the spell resolves")
	}
	effectstest.TakeSearchPrompts(t, g)
	for _, name := range []string{"Forest", "Plains"} {
		l := effectstest.FindBattlefieldByName(g, name)
		if l == uuid.Nil {
			t.Fatalf("%s did not reach the battlefield", name)
		}
		if !effectstest.B16Tapped(t, g, l) {
			t.Errorf("%s enters tapped", name)
		}
	}
	if spec, _ := effects.Lookup(effectstest.B18CycleOfRenewalOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the sacrifice is part of the effect now")
	}
}

// --- removal -------------------------------------------------------

func TestB18HerosDownfallDestroysACreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Jace",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{"loyalty": 3},
	})
	rock := effectstest.B12Permanent(g, opp.ID, "Mana Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(g.Seats[0], "Hero's Downfall", "Instant", "", effectstest.B18HerosDownfallOracle, []string{"B"})
	if err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(rock)}); err == nil {
		t.Fatal("an artifact is not a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Hero's Downfall", "Instant", effectstest.B18HerosDownfallOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature is destroyed")
	}
	effectstest.CastCatalogSpell(t, g, "Hero's Downfall", "Instant", effectstest.B18HerosDownfallOracle, effectstest.B16TargetCard(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker is destroyed")
	}
}

func TestB18SplitUpDestroysTheTappedOrTheUntappedHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tappedMine := effectstest.B16Creature(g, me.ID, "Tapped Mine", "Creature — Bear", 2, 2, "G")
	untappedMine := effectstest.B16Creature(g, me.ID, "Untapped Mine", "Creature — Bear", 2, 2, "G")
	tappedTheirs := effectstest.B16Creature(g, opp.ID, "Tapped Theirs", "Creature — Bear", 2, 2, "G")
	untappedTheirs := effectstest.B16Creature(g, opp.ID, "Untapped Theirs", "Creature — Bear", 2, 2, "G")
	effectstest.B16Tap(g, tappedMine)
	effectstest.B16Tap(g, tappedTheirs)
	effectstest.B12PlayFromHand(t, g, "Split Up", "Sorcery", effectstest.B18SplitUpOracle, game.CastSpellParams{Modes: []int{0}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tappedMine) || g.Battlefield.Contains(tappedTheirs) {
		t.Error("mode 0 destroys every tapped creature")
	}
	if !g.Battlefield.Contains(untappedMine) || !g.Battlefield.Contains(untappedTheirs) {
		t.Error("mode 0 leaves the untapped ones")
	}
	effectstest.B12PlayFromHand(t, g, "Split Up", "Sorcery", effectstest.B18SplitUpOracle, game.CastSpellParams{Modes: []int{1}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(untappedMine) || g.Battlefield.Contains(untappedTheirs) {
		t.Error("mode 1 destroys every untapped creature")
	}
	// S30 (#470 / #446): the boardwipe indestructible gap this used
	// to require a declaration for is closed, so the card is Full.
	if spec, _ := effects.Lookup(effectstest.B18SplitUpOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Split Up implements every clause it prints")
	}
}

func TestB18WindsOfAbandonExilesAndTheControllerFetchesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.SeedSearchLibrary(opp,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
	)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Winds of Abandon", "Sorcery", "", effectstest.B18WindsOfAbandonOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(mine)}); err == nil {
		t.Fatal("your own creature is not a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("the target is exiled")
	}
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the creature's controller searches")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("the caster does not search")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Forest")
	forest := effectstest.FindBattlefieldByName(g, "Forest")
	if forest == uuid.Nil || !effectstest.B16Tapped(t, g, forest) {
		t.Error("the basic enters tapped under its controller")
	}
	if fc, _ := effectstest.BattlefieldCard(g, forest); fc.Controller != opp.ID {
		t.Error("the land is the searching player's")
	}
}

func TestB18WindsOfAbandonOverloadedExilesEachCreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	a := effectstest.B16Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := effectstest.B16Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	c := effectstest.B16Creature(g, other.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	effectstest.SeedSearchLibrary(opp,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Plains", "Basic Land — Plains"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	effectstest.SeedSearchLibrary(other, effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"))
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Winds of Abandon", "Sorcery", "", effectstest.B18WindsOfAbandonOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overloaded cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, x := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(x) {
			t.Errorf("%s survived the overload", x)
		}
		// Overload EXILES. A destroy would leave the battlefield too.
		if !g.Exile.Contains(x) {
			t.Errorf("%s left the battlefield but was not exiled", x)
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creatures are untouched")
	}
	// The opponent who lost two creatures runs one two-card search; the
	// one who lost one has exactly one basic and is asked too: a search
	// for a basic land may find nothing (CR 701.23b, #3053).
	pc := effectstest.SearchChoiceFor(g, opp.ID)
	if pc == nil || pc.Count != 2 {
		t.Fatalf("opp's search = %+v, want a two-card prompt", pc)
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Forest", "Plains")
	if effectstest.FindBattlefieldByName(g, "Forest") == uuid.Nil || effectstest.FindBattlefieldByName(g, "Plains") == uuid.Nil {
		t.Error("opp's two basics arrive")
	}
	if effectstest.FindBattlefieldByName(g, "Swamp") != uuid.Nil {
		t.Error("the other player's single basic waits for their answer")
	}
	effectstest.AnswerSearchNamed(t, g, other.ID, "Swamp")
	swamp := effectstest.FindBattlefieldByName(g, "Swamp")
	if swamp == uuid.Nil || !effectstest.B16Tapped(t, g, swamp) {
		t.Error("the other player's single basic arrives tapped once they take it")
	}
}

func TestB18VenserBouncesATargetPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.CastCatalogSpell(t, g, "Venser, Shaper Savant", "Legendary Creature — Human Wizard", effectstest.B18VenserShaperSavantOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the target permanent returns to its owner's hand")
	}
}

// --- the creatures with triggers -----------------------------------

func TestB18StormfistCrusaderEachPlayerDrawsAndLosesOneOnYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Stormfist Crusader", "Creature — Human Knight", effectstest.B18StormfistCrusaderOracle, 2, 2)
	crusader := effectstest.FindBattlefieldByName(g, "Stormfist Crusader")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if len(g.PendingTriggers) != 0 || effectstest.TriggerOnStack(g, crusader) != nil {
		t.Error("an opponent's upkeep is not yours")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	myLife, oppLife := me.Life, opp.Life
	myHand, oppHand := me.Hand.Size(), opp.Hand.Size()
	if len(g.PendingTriggers) == 0 && effectstest.TriggerOnStack(g, crusader) == nil {
		t.Fatal("your upkeep triggers it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 || opp.Life != oppLife-1 {
		t.Errorf("life me %d → %d, opp %d → %d: want -1 each", myLife, me.Life, oppLife, opp.Life)
	}
	if me.Hand.Size() != myHand+1 || opp.Hand.Size() != oppHand+1 {
		t.Errorf("hands me %d → %d, opp %d → %d: want +1 each", myHand, me.Hand.Size(), oppHand, opp.Hand.Size())
	}
}

func TestB18FalkenrathNobleDrainsATargetPlayerPerDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	noble := effectstest.B12Push(g, me.ID, "Falkenrath Noble", "Creature — Vampire Noble", effectstest.B18FalkenrathNobleOracle, 2, 2)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("opp %d → %d, me %d → %d: want -1 / +1", oppLife, opp.Life, myLife, me.Life)
	}
	// Its own death triggers it too.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(noble) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the Noble's own death drains: opp %d, want %d", opp.Life, oppLife-2)
	}
}

func TestB18SangromancerMayGainThreeOnOpponentDeathsAndDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sangromancer", "Creature — Vampire Shaman", effectstest.B18SangromancerOracle, 3, 3)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.B18Kill(t, g, mine)
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("your own creature dying is not an opponent's")
	}
	effectstest.B18Kill(t, g, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3 for the death", life, me.Life)
	}
	effectstest.HandCardFull(opp, "Junk", "Sorcery", "", "", nil)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("declined: life %d, want %d", me.Life, life+3)
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("accepted: life %d, want %d", me.Life, life+6)
	}
}

func TestB18ObNixilisTheFallenDrainsATargetAndGrowsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := effectstest.B12Push(g, me.ID, "Ob Nixilis, the Fallen", "Legendary Creature — Demon", effectstest.B18ObNixilisTheFallenOracle, 3, 3)
	life := opp.Life
	effectstest.PlayLandFromHand(t, g, "Swamp", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opp %d → %d, want -3", life, opp.Life)
	}
	if effectstest.B18Counters(t, g, ob, "+1/+1") != 3 {
		t.Errorf("Ob Nixilis has %d +1/+1 counters, want 3", effectstest.B18Counters(t, g, ob, "+1/+1"))
	}
	if c, _ := effectstest.BattlefieldCard(g, ob); c.CurrentPower() != 6 {
		t.Errorf("Ob Nixilis is a 3/3 with three counters, power %d, want 6", c.CurrentPower())
	}
	// Declining leaves both halves undone.
	life = opp.Life
	effectstest.PlayLandFromHand(t, g, "Swamp Two", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || effectstest.B18Counters(t, g, ob, "+1/+1") != 3 {
		t.Error("declined: no drain, no counters")
	}
}

func TestB18KederektParasitePingsADrawingOpponentWhileYouControlRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kederekt Parasite", "Creature — Horror", effectstest.B18KederektParasiteOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("no red permanent, no trigger")
	}
	red := effectstest.B16Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1, "R")
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 2) })
	if n := effectstest.B18TriggerPromptCount(g, me.ID); n != 2 {
		t.Fatalf("two cards drawn is two prompts, got %d", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opp %d → %d, want -2", life, opp.Life)
	}
	// Your own draw never triggers it.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("your own draw is not an opponent's")
	}
	// The intervening if is checked again on resolution.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(red) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("red permanent gone in response: opp %d, want %d", opp.Life, life-2)
	}
}

func TestB18GarruksPackleaderMayDrawWhenABigCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Garruk's Packleader", "Creature — Beast", effectstest.B18GarruksPackleaderOracle, 4, 4)
	effectstest.Top100CastCreature(t, g, me, "Small Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Fatal("power 2 is not 3 or greater")
	}
	hand := me.Hand.Size()
	effectstest.Top100CastCreature(t, g, me, "Big Beast", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: added and cast (net 0), drew (+1)", hand, got)
	}
}

func TestB18JunkDiverReturnsAnotherArtifactCardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	diver := effectstest.B12Push(g, me.ID, "Junk Diver", "Artifact Creature — Bird", effectstest.B18JunkDiverOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, diver, "flying") {
		t.Error("flying is printed")
	}
	relic := effectstest.Batch01GraveyardCard(me, "Sol Ring", "Artifact")
	bear := effectstest.Batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(diver) })
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, bear) || !effectstest.HasID(p.PickTargetCards, relic) {
		t.Error("artifact cards in your graveyard are offered, creature cards are not")
	}
	// #1337: the picker excludes the Diver ITSELF too — TargetsFrom /
	// AnotherTarget by instance rather than the old resolution-time
	// self-decline — so the gap that used to be a declared caveat is
	// closed.
	if effectstest.HasID(p.PickTargetCards, diver) {
		t.Error("Junk Diver itself is not offered — \"another\" excludes it by instance")
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the artifact card returns to hand")
	}
	if spec, _ := effects.Lookup(effectstest.B18JunkDiverOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("the self-pick gap (#1337) is closed")
	}
}

func TestB18RighteousValkyrieGainsToughnessForAngelsAndClerics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valkyrie := effectstest.B12Push(g, me.ID, "Righteous Valkyrie", "Creature — Angel Cleric", effectstest.B18RighteousValkyrieOracle, 2, 4)
	life := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Serra Angel", TypeLine: "Creature — Angel", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4 (the Angel's toughness)", life, me.Life)
	}
	effectstest.Top100CastCreature(t, g, me, "Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("a Beast is neither: life %d, want %d", me.Life, life+4)
	}
	// An opponent's Cleric is not yours.
	effectstest.AdvanceToMainOf(t, g, 1)
	cid := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: cid, Name: "Their Cleric", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, cid, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("an opponent's Cleric: life %d, want %d", me.Life, life+4)
	}
	// The +2/+2 was a declared gap until #1117 taught the layer engine
	// to invalidate on a life change. It is live now, and the line is
	// the TABLE's starting total plus 7 — 47 here, so 44 life is not
	// enough and the Valkyrie is still its printed 2/4.
	spec, _ := effects.Lookup(effectstest.B18RighteousValkyrieOracle)
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %q, want full now that the anthem ships", spec.Completeness)
	}
	if got := effectstest.EffectivePower(t, g, valkyrie); got != 2 {
		t.Errorf("power at %d life = %d, want 2 (below the +7 line)", me.Life, got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if got := effectstest.EffectivePower(t, g, valkyrie); got != 4 {
		t.Errorf("power at %d life = %d, want 4 (+2/+2 above the line)", me.Life, got)
	}
	if got := effectstest.EffectiveToughness(t, g, valkyrie); got != 6 {
		t.Errorf("toughness at %d life = %d, want 6", me.Life, got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	if got := effectstest.EffectivePower(t, g, valkyrie); got != 2 {
		t.Errorf("power right after dropping back to %d life = %d, want 2", me.Life, got)
	}
}

func TestB18KorSpiritdancerGrowsWithAurasAndDrawsOnAuraCasts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dancer := effectstest.B12Push(g, me.ID, "Kor Spiritdancer", "Creature — Kor Wizard", effectstest.B18KorSpiritdancerOracle, 0, 2)
	aura := effectstest.B12Permanent(g, me.ID, "Rancor", "Enchantment — Aura")
	equip := effectstest.B12Permanent(g, me.ID, "Test Blade", "Artifact — Equipment")
	if effectstest.EffectivePower(t, g, dancer) != 0 || effectstest.EffectiveToughness(t, g, dancer) != 2 {
		t.Fatal("unenchanted, a 0/2")
	}
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{aura, equip} {
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: dancer}); err != nil {
				t.Fatalf("AttachForEffect: %v", err)
			}
		}
	})
	if effectstest.EffectivePower(t, g, dancer) != 2 || effectstest.EffectiveToughness(t, g, dancer) != 4 {
		t.Errorf("one Aura and one Equipment = %d/%d, want 2/4", effectstest.EffectivePower(t, g, dancer), effectstest.EffectiveToughness(t, g, dancer))
	}
	hand := me.Hand.Size()
	effectstest.B18CastFromHand(t, g, me, "Pacifism", "Enchantment — Aura", "", game.CastSpellParams{})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: the Aura was added and cast (net 0), drew (+1)", hand, got)
	}
	effectstest.B18CastFromHand(t, g, me, "Plain Enchantment", "Enchantment", "", game.CastSpellParams{})
	if effectstest.B18TriggerPromptCount(g, me.ID) != 0 {
		t.Error("a non-Aura enchantment is not an Aura spell")
	}
}

func TestB18ThunderfootBalothIsALieutenant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baloth := effectstest.B12Push(g, me.ID, "Thunderfoot Baloth", "Creature — Beast", effectstest.B18ThunderfootBalothOracle, 5, 5)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	if effectstest.EffectivePower(t, g, baloth) != 5 || effectstest.EffectivePower(t, g, bear) != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("without your commander, nothing")
	}
	commander := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if effectstest.EffectivePower(t, g, baloth) != 7 || effectstest.EffectiveToughness(t, g, baloth) != 7 {
		t.Errorf("the Baloth = %d/%d, want 7/7", effectstest.EffectivePower(t, g, baloth), effectstest.EffectiveToughness(t, g, baloth))
	}
	if !effectstest.HasEffectiveKeyword(t, g, baloth, "trample") {
		t.Error("the Baloth keeps its printed trample")
	}
	if effectstest.EffectivePower(t, g, bear) != 4 || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("other creatures you control get +2/+2 and trample")
	}
	if effectstest.EffectivePower(t, g, commander) != 5 || !effectstest.HasEffectiveKeyword(t, g, commander, "trample") {
		t.Error("the commander is another creature you control")
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature is untouched")
	}
}

func TestB18ThunderfootBalothIgnoresAStolenCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	baloth := effectstest.B12Push(g, me.ID, "Thunderfoot Baloth", "Creature — Beast", effectstest.B18ThunderfootBalothOracle, 5, 5)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Commander", TypeLine: "Legendary Creature — Elf",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: me.ID, IsCommander: true,
	})
	if effectstest.EffectivePower(t, g, baloth) != 5 {
		t.Error("an opponent's commander you control is not YOUR commander")
	}
}

func TestB18WoundReflectionDoublesEachOpponentsLossAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Wound Reflection", "Enchantment", effectstest.B18WoundReflectionOracle, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3, "G")
	oppLife, otherLife, myLife := opp.Life, other.Life, me.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(other.ID))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5) })
	if opp.Life != oppLife-3 || other.Life != otherLife-3 {
		t.Fatalf("setup: opp %d other %d", opp.Life, other.Life)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-6 || other.Life != otherLife-6 {
		t.Error("a turn with no loss reflects nothing")
	}
}

func TestB18GrazilaxxMayBounceABlockedAttackerAndDrawsOncePerCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Grazilaxx, Illithid Scholar", "Legendary Creature — Horror", effectstest.B18GrazilaxxOracle, 3, 2)
	a := effectstest.B16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2, "G")
	b := effectstest.B16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2, "G")
	c := effectstest.B16Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2, "G")
	wall1 := effectstest.B16Creature(g, opp.ID, "Wall One", "Creature — Wall", 0, 4, "W")
	wall2 := effectstest.B16Creature(g, opp.ID, "Wall Two", "Creature — Wall", 0, 4, "W")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall1, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := effectstest.B18TriggerPromptCount(g, me.ID); n != 0 {
		t.Fatalf("#830: a click stages the pairing and announces nothing, got %d prompts", n)
	}
	if err := g.DeclareBlocker(wall2, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if n := effectstest.B18TriggerPromptCount(g, me.ID); n != 1 {
		t.Fatalf("a double block is one 'becomes blocked' (CR 506.4), got %d prompts", n)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the blocked creature returns to its owner's hand")
	}
	hand := me.Hand.Size()
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("two unblocked bears connect: opp %d, want %d", opp.Life, life-4)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d: one draw for one or more creatures connecting", hand, got)
	}
}

func TestB18MagdaMakesATappedTreasureOncePerTurnPerCrime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	magda := effectstest.B12Push(g, me.ID, "Magda, the Hoardmaster", "Legendary Creature — Dwarf Berserker", effectstest.B18MagdaTheHoardmasterOrcl, 2, 2)
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	// Targeting your own creature is not a crime.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Treasure") != 0 {
		t.Fatal("targeting your own creature is not a crime")
	}
	// Targeting an opponent is; the trigger resolves above the Bolt.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	if effectstest.TriggerOnStack(g, magda) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the crime triggers at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Treasure") != 1 {
		t.Fatalf("%d Treasures, want 1", effectstest.B16CountNamed(g, "Treasure"))
	}
	if !effectstest.B16Tapped(t, g, effectstest.FindBattlefieldByName(g, "Treasure")) {
		t.Error("the Treasure enters tapped")
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Treasure") != 1 {
		t.Error("only once each turn")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Treasure") != 2 {
		t.Error("a new turn, a new Treasure")
	}
	// #747: the Scorpion Dragon ability ships at its printed count;
	// its engine test is in sacrifice_n_cards_test.go.
	if spec, _ := effects.Lookup(effectstest.B18MagdaTheHoardmasterOrcl); spec.Completeness != effects.CompletenessFull || len(spec.Activated) != 1 {
		t.Error("the Scorpion Dragon ability ships whole")
	}
}

func TestB18MikaeusEntersWithXAndGrowsByTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mik := effectstest.B12PlayFromHand(t, g, "Mikaeus, the Lunarch", "Legendary Creature — Human Cleric", effectstest.B18MikaeusTheLunarchOracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mik) {
		t.Fatal("Mikaeus did not resolve")
	}
	if effectstest.B18Counters(t, g, mik, "+1/+1") != 2 {
		t.Fatalf("%d counters, want X=2", effectstest.B18Counters(t, g, mik, "+1/+1"))
	}
	if err := g.ActivateCatalogAbility(me.ID, mik, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a summoning-sick Mikaeus cannot tap")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mik {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	effectstest.B16Activate(t, g, me.ID, mik, 0, game.ActivateAbilityParams{})
	if effectstest.B18Counters(t, g, mik, "+1/+1") != 3 {
		t.Errorf("%d counters after the tap, want 3", effectstest.B18Counters(t, g, mik, "+1/+1"))
	}
	// #625: the team pump is the second ability; counter_cost_cards_test.go
	// drives it. The X-counter timing is still a declared caveat.
	// #1002 moved the X counters onto the CR 614 entry pipeline and
	// #691 made the X=0 body die as printed, so both caveats are gone
	// and both tap abilities still ship.
	if spec, _ := effects.Lookup(effectstest.B18MikaeusTheLunarchOracle); spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.Activated) != 2 {
		t.Errorf("both tap abilities ship and nothing is declared: %v %v %d", spec.Completeness, spec.Caveats, len(spec.Activated))
	}
}

func TestB18ArixmethesSleepsAsALandAndWakesAfterFiveSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arix := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: arix, Name: "Arixmethes, Slumbering Isle", TypeLine: "Legendary Creature — Kraken",
		OracleID: effectstest.B18ArixmethesOracle, Power: 12, Toughness: 12, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, arix, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(arix) {
		t.Fatal("Arixmethes did not resolve")
	}
	if !effectstest.B16Tapped(t, g, arix) || effectstest.TapEventsFor(g, arix) != 0 {
		t.Error("enters tapped, as a replacement")
	}
	if effectstest.B18Counters(t, g, arix, "slumber") != 5 {
		t.Fatalf("%d slumber counters, want 5", effectstest.B18Counters(t, g, arix, "slumber"))
	}
	c, _ := effectstest.BattlefieldCard(g, arix)
	if c.IsCreature() || !c.IsLand() || c.HasSubtype("Kraken") {
		t.Errorf("asleep it is a land and not a creature: types %v subtypes %v", effectstest.EffectiveTypes(t, g, arix), c.Effective().Subtypes)
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
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool %v, want [G U]", got)
	}
	// Each spell you cast may remove a counter; declining keeps it.
	effectstest.CastCatalogSpell(t, g, "Krenko's Command", "Sorcery", effectstest.B18KrenkosCommandOracle, nil)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B18Counters(t, g, arix, "slumber") != 5 {
		t.Error("declined: the counter stays")
	}
	for i := 0; i < 5; i++ {
		effectstest.CastCatalogSpell(t, g, "Krenko's Command", "Sorcery", effectstest.B18KrenkosCommandOracle, nil)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.B18Counters(t, g, arix, "slumber") != 0 {
		t.Fatalf("%d slumber counters after five removals, want 0", effectstest.B18Counters(t, g, arix, "slumber"))
	}
	c, _ = effectstest.BattlefieldCard(g, arix)
	if !c.IsCreature() || c.IsLand() || !c.HasSubtype("Kraken") {
		t.Errorf("awake it is a Kraken creature: types %v", effectstest.EffectiveTypes(t, g, arix))
	}
	if effectstest.EffectivePower(t, g, arix) != 12 || effectstest.EffectiveToughness(t, g, arix) != 12 {
		t.Error("a 12/12")
	}
	// A sixth spell with no counters left: nothing to remove, no prompt.
	effectstest.CastCatalogSpell(t, g, "Krenko's Command", "Sorcery", effectstest.B18KrenkosCommandOracle, nil)
	if effectstest.B18TriggerPromptCount(g, me.ID) == 1 {
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B18Counters(t, g, arix, "slumber") != 0 {
		t.Error("never negative")
	}
}
