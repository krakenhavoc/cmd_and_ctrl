package testsru

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Artificer Class (CR 716) -----------------------------------------

// Level 1: the first artifact spell you cast each turn costs {1} less,
// and nothing else does.
func TestArtificerClassLevelOneDiscountsOnlyTheFirstArtifactSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	class := effectstest.PushClass(g, me.ID, "Artificer Class", effectstest.S58ArtificerClassTypeLine, effectstest.S58ArtificerClassOracle)
	if got := effectstest.LevelOf(g, class); got != 1 {
		t.Fatalf("level = %d, want 1", got)
	}
	if got := len(game.TriggersForCard(effectstest.LayeredCard(t, g, class))); got != 0 {
		t.Errorf("at level 1 the level-2 and level-3 triggers must not exist: got %d", got)
	}

	if got := effectstest.PriceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("first artifact spell: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a nonartifact spell is not discounted: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("an opponent's artifact spell is not discounted: %d, want 3", got)
	}

	// A nonartifact spell does not use the discount up.
	effectstest.CastCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("after an instant, the first artifact spell is still discounted: %d, want 2", got)
	}

	effectstest.CastCatalogSpell(t, g, "First Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.CastTallyFor(me.ID).Artifact; got != 1 {
		t.Fatalf("artifact spells cast this turn = %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("the second artifact spell this turn: %d, want 3", got)
	}
}

// An artifact spell cast before the Class arrived was the turn's first,
// so nothing later that turn is discounted.
func TestArtificerClassCountsAnArtifactCastBeforeItArrived(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "First Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PushClass(g, me.ID, "Artificer Class", effectstest.S58ArtificerClassTypeLine, effectstest.S58ArtificerClassOracle)
	if got := effectstest.PriceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("an artifact spell was already cast this turn: %d, want 3", got)
	}
}

// Level 2: "When this Class becomes level 2, reveal cards from the top of
// your library until you reveal an artifact card. Put that card into
// your hand and the rest on the bottom of your library in a random
// order." The level-3 line does not exist yet.
func TestArtificerClassLevelTwoRevealsUntilAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	class := effectstest.PushClass(g, me.ID, "Artificer Class", effectstest.S58ArtificerClassTypeLine, effectstest.S58ArtificerClassOracle)
	rock := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Buried Rock", TypeLine: "Artifact"})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand, library := me.Hand.Size(), me.Library.Size()

	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 2 {
		t.Fatalf("level = %d, want 2", got)
	}
	if !me.Hand.Contains(rock) {
		t.Fatal("the revealed artifact card did not reach the hand")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (only the artifact)", got)
	}
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library = %d, want %d (the rest went to the bottom)", got, library-1)
	}
	if got := len(game.TriggersForCard(effectstest.LayeredCard(t, g, class))); got != 1 {
		t.Errorf("at level 2 only the level-2 trigger exists: got %d", got)
	}

	// The level-3 end-step trigger does not exist at level 2.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the level-3 end-step trigger asked for a target at level 2")
	}
}

// A library with no artifact card reveals itself entirely and every
// card goes back; nothing reaches the hand.
func TestArtificerClassLevelTwoWithNoArtifactPutsEverythingBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	class := effectstest.PushClass(g, me.ID, "Artificer Class", effectstest.S58ArtificerClassTypeLine, effectstest.S58ArtificerClassOracle)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand, library := me.Hand.Size(), me.Library.Size()
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("hand %d -> %d, library %d -> %d; want both unchanged", hand, me.Hand.Size(), library, me.Library.Size())
	}
}

// Level 3: at the beginning of your end step, a token copy of target
// artifact you control. The level-up abilities climb in order and at
// sorcery speed.
func TestArtificerClassLevelThreeCopiesAnArtifactAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	class := effectstest.PushClass(g, me.ID, "Artificer Class", effectstest.S58ArtificerClassTypeLine, effectstest.S58ArtificerClassOracle)
	rock := effectstest.S58Rock(g, me.ID, "Test Rock", "Artifact", "{2}")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)

	// CR 716.2a: level 3 cannot be gained from level 1.
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{5}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, class, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("level 3 from level 1 must be refused (CR 716.2a)")
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{5}{U}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if got := effectstest.LevelOf(g, class); got != 3 {
		t.Fatalf("level = %d, want 3", got)
	}
	if got := len(game.TriggersForCard(effectstest.LayeredCard(t, g, class))); got != 2 {
		t.Errorf("at level 3 both triggers exist: got %d", got)
	}
	// Level 1's discount is still there (CR 716.2a: "or greater").
	if got := effectstest.PriceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("level 1's discount at level 3: %d, want 2", got)
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	rocks := effectstest.S58Named(g, me.ID, "Test Rock")
	if len(rocks) != 2 {
		t.Fatalf("Test Rocks = %d, want the original and a token copy", len(rocks))
	}
	tokens := 0
	for _, c := range rocks {
		if effects.IsToken(c) {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("token copies = %d, want 1", tokens)
	}
}

// --- Mishra, Eminent One ------------------------------------------------

func TestMishraMakesAHastyWarformAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Mishra, Eminent One", effectstest.S58MishraTypeLine, effectstest.S58MishraOracle, false)
	rock := effectstest.S58Rock(g, me.ID, "Test Rock", "Artifact — Equipment", "{2}")
	bot := effectstest.S58Rock(g, me.ID, "Test Bot", "Artifact Creature — Construct", "{2}")

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bot}); err == nil {
		t.Error("an artifact creature is not a noncreature artifact")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)

	warforms := effectstest.S58Named(g, me.ID, "Mishra's Warform")
	if len(warforms) != 1 {
		t.Fatalf("Mishra's Warforms = %d, want 1", len(warforms))
	}
	w := warforms[0].InstanceID
	if !effects.IsToken(warforms[0]) {
		t.Error("the Warform is a token")
	}
	if p, tough := effectstest.EffectivePower(t, g, w), effectstest.EffectiveToughness(t, g, w); p != 4 || tough != 4 {
		t.Errorf("Warform is %d/%d, want 4/4", p, tough)
	}
	card := effectstest.LayeredCard(t, g, w)
	if !card.IsArtifact() || !card.IsCreature() || !card.HasSubtype("Construct") || !card.HasSubtype("Equipment") {
		t.Errorf("Warform type line %q: want an Equipment Construct artifact creature", card.TypeLine)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, w), "haste") {
		t.Error("the Warform has haste this turn")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the original artifact stays")
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S58Named(g, me.ID, "Mishra's Warform")); got != 0 {
		t.Errorf("Warforms after the end step = %d, want 0 (sacrificed)", got)
	}
}

// --- Arcum Dagsson -------------------------------------------------------

func TestArcumDagssonSacrificesThenItsControllerSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arcum := effectstest.PushCatalogPermanent(g, me.ID, "Arcum Dagsson", "Legendary Creature — Human Artificer", effectstest.S58ArcumDagssonOracle, false)
	theirBot := effectstest.S58Rock(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "{3}")
	bear := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Their Golem", TypeLine: "Artifact Creature — Golem"})

	if err := g.ActivateCatalogAbility(me.ID, arcum, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a non-artifact creature target = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, arcum, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirBot}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBot) {
		t.Fatal("the artifact creature was not sacrificed")
	}
	search := effectstest.SearchChoiceFor(g, opp.ID)
	if search == nil {
		t.Fatal("the creature's controller is not offered the search")
	}
	if effectstest.SearchOptionNamed(g, search, "Their Golem") != uuid.Nil {
		t.Error("an artifact creature card is offered; the search is for a noncreature artifact")
	}
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Their Rock")
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S58Named(g, opp.ID, "Their Rock")); got != 1 {
		t.Errorf("Their Rock on the battlefield under its owner = %d, want 1", got)
	}
}

// --- Copy Artifact ------------------------------------------------------

func TestCopyArtifactEntersAsAnArtifactEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirRock := effectstest.S58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")
	bear := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := effectstest.CastCatalogSpell(t, g, "Copy Artifact", "Enchantment", effectstest.S58CopyArtifactOracle, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !effectstest.HasID(pc.CopyOptions, theirRock) || effectstest.HasID(pc.CopyOptions, bear) {
		t.Errorf("CopyOptions = %v, want the artifact and not the creature", pc.CopyOptions)
	}
	effectstest.ResolveWithCopyChoice(t, g, theirRock)

	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Their Rock" {
		t.Errorf("name = %q, want the copied artifact's", c.Name)
	}
	if !c.IsArtifact() || !c.IsEnchantment() {
		t.Errorf("type line %q: want an artifact AND an enchantment", c.TypeLine)
	}
	if c.Controller != me.ID {
		t.Error("the copy is mine")
	}
}

func TestCopyArtifactCopyingAnArtifactCreatureIsAllThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bot := effectstest.S58Rock(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "{3}")
	id := effectstest.CastCatalogSpell(t, g, "Copy Artifact", "Enchantment", effectstest.S58CopyArtifactOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, bot)
	c := effectstest.LayeredCard(t, g, id)
	if !c.IsArtifact() || !c.IsCreature() || !c.IsEnchantment() || !c.HasSubtype("Construct") {
		t.Errorf("type line %q: want Artifact Creature Enchantment — Construct", c.TypeLine)
	}
}

// --- Cybermen Squadron ---------------------------------------------------

func TestCybermenSquadronGivesMyriadToNonlegendaryArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B21Push(g, me.ID, "Cybermen Squadron", "Artifact Creature — Cyberman", effectstest.S58CybermenSquadronOracle, 5, 5)
	bot := effectstest.PushTribalCreature(g, me.ID, "Test Bot", "Artifact Creature — Construct", 3, 3, "vigilance")

	effectstest.DeclareAttack(t, g, opp1.ID, bot)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	attacking := map[uuid.UUID]int{}
	for _, c := range effectstest.S58Named(g, me.ID, "Test Bot") {
		if c.InstanceID == bot {
			continue
		}
		if !c.Tapped || !effects.IsToken(c) {
			t.Errorf("a myriad copy is a tapped token")
		}
		attacking[c.AttackingTarget]++
	}
	if attacking[opp2.ID] != 1 || attacking[opp3.ID] != 1 || attacking[opp1.ID] != 0 {
		t.Errorf("copies attack %v, want one at each non-defending opponent", attacking)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S58Named(g, me.ID, "Test Bot")); got != 1 {
		t.Errorf("Test Bots after end of combat = %d, want the original", got)
	}
}

func TestCybermenSquadronSkipsLegendaryAndNonartifactAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B21Push(g, me.ID, "Cybermen Squadron", "Artifact Creature — Cyberman", effectstest.S58CybermenSquadronOracle, 5, 5)
	legend := effectstest.PushTribalCreature(g, me.ID, "Legend Bot", "Legendary Artifact Creature — Construct", 3, 3)
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, legend, bear)
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
		t.Error("a legendary or nonartifact attacker has no myriad")
	}
}

// --- Diplomatic Immunity -------------------------------------------------

func TestDiplomaticImmunityGivesShroudAndHasIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Diplomatic Immunity", "Enchantment — Aura", effectstest.S58DiplomaticImmunityOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "shroud") {
		t.Error("the enchanted creature has shroud")
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, aura), "shroud") {
		t.Error("the Aura itself has shroud")
	}
}

// --- Forsaken Monument ---------------------------------------------------

func TestForsakenMonumentPumpsColorlessCreaturesAndDoublesColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Forsaken Monument", "Legendary Artifact", effectstest.S58ForsakenMonumentOracle, false)
	bot := effectstest.B21Push(g, me.ID, "Test Bot", "Artifact Creature — Construct", "", 2, 2)
	bear := effectstest.B21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	theirBot := effectstest.B21Push(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "", 2, 2)
	if got := effectstest.EffectivePower(t, g, bot); got != 4 {
		t.Errorf("colorless creature you control: power %d, want 4", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("green creature: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirBot); got != 2 {
		t.Errorf("an opponent's colorless creature: power %d, want 2", got)
	}

	gardens := effectstest.PushCatalogPermanent(g, me.ID, "The Mycosynth Gardens", effectstest.S58MycosynthGardensTypes, effectstest.S58MycosynthGardensOracle, false)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.TapForMana(t, g, me.ID, gardens)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"C", "C"}) {
		t.Errorf("a permanent tapped for {C}: pool %v, want {C}{C}", got)
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); !reflect.DeepEqual(got, []string{"C", "C", "G"}) {
		t.Errorf("then a Forest: pool %v, want {C}{C}{G}", got)
	}
}

func TestForsakenMonumentGainsLifeOnColorlessSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Forsaken Monument", "Legendary Artifact", effectstest.S58ForsakenMonumentOracle, false)
	life := me.Life
	effectstest.S58Cast(t, g, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a green spell: life %d -> %d, want unchanged", life, me.Life)
	}
	effectstest.S58Cast(t, g, game.Card{Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{2}"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("a colorless spell: life %d -> %d, want +2", life, me.Life)
	}
}

// --- Knight Paladin ------------------------------------------------------

func TestKnightPaladinDealsFourToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	effectstest.S58Cast(t, g, game.Card{Name: "Knight Paladin", TypeLine: "Artifact — Vehicle", OracleID: effectstest.S58KnightPaladinOracle,
		ManaCost: "{5}", Power: 6, Toughness: 6}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 4
		if p.ID == me.ID {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d", i, p.Life, want)
		}
	}
	if !game.HasKeyword(&game.Card{OracleID: effectstest.S58KnightPaladinOracle}, "trample") {
		t.Error("Knight Paladin prints trample")
	}
}

// --- Nautiloid Ship -------------------------------------------------------

func TestNautiloidShipExilesAGraveyardThenReturnsACreatureFromIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.PushGraveyardCardForTest(opp, "Their Dragon")
	spell := effectstest.B17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	mine := effectstest.PushGraveyardCardForTest(me, "My Bear")

	ship := effectstest.S58Cast(t, g, game.Card{Name: "Nautiloid Ship", TypeLine: effectstest.S58NautiloidShipTypeLine, OracleID: effectstest.S58NautiloidShipOracle,
		ManaCost: "{4}", Power: 5, Toughness: 5}, nil)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Fatalf("their graveyard still holds %d cards", opp.Graveyard.Size())
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("only the target player's graveyard is exiled")
	}

	// Crew it and attack this turn: it is not summoning sick for the test.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == ship {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	crewer := effectstest.PushCrewerForTest(g, me.ID, "Crewer", 3)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackWith(t, g, opp.ID, ship)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice of a creature card exiled with the Ship")
	}
	if !reflect.DeepEqual(pick.ChooseCards, []uuid.UUID{dragon}) {
		t.Errorf("offered %v, want only the creature card %v (not the instant %v)", pick.ChooseCards, dragon, spell)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.S58Named(g, me.ID, "Their Dragon"); len(got) != 1 {
		t.Errorf("Their Dragon under my control = %d, want 1", len(got))
	}
}

// --- Portal to Phyrexia ---------------------------------------------------

func TestPortalToPhyrexiaMakesEachOpponentSacrificeThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	myBear := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 4; i++ {
		effectstest.PushTribalCreature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	}
	effectstest.PushTribalCreature(g, opp2.ID, "Lone Bear", "Creature — Bear", 2, 2)

	effectstest.S58Cast(t, g, game.Card{Name: "Portal to Phyrexia", TypeLine: effectstest.S58PortalToPhyrexiaTypeLin, OracleID: effectstest.S58PortalToPhyrexiaOracle, ManaCost: "{9}"}, nil)
	for i := 0; i < 32; i++ {
		var c *game.PendingChoice
		for _, p := range g.Seats {
			if c = effectstest.SacrificeChoiceFor(g, p.ID); c != nil {
				break
			}
		}
		if c == nil {
			if effectstest.StackFullyEmpty(g) {
				break
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			continue
		}
		if c.Chooser == me.ID {
			t.Fatal("the Portal's controller is not asked")
		}
		// #3664: a seat owing three creatures names all three in one pick.
		opts := effectstest.SacrificeOptions(c)
		if c.ChooseMax != 3 || len(opts) < 3 {
			t.Fatalf("the pick takes %d of %d creatures, want 3", c.ChooseMax, len(opts))
		}
		effectstest.AnswerSacrifice(t, g, c.Chooser, opts[:3]...)
	}
	if got := effectstest.CountBattlefieldNamed(g, opp1.ID, "Their Bear"); got != 1 {
		t.Errorf("opponent with four creatures keeps %d, want 1", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, opp2.ID, "Lone Bear"); got != 0 {
		t.Errorf("opponent with one creature keeps %d, want 0", got)
	}
	if !g.Battlefield.Contains(myBear) {
		t.Error("your own creature is not sacrificed")
	}
}

func TestPortalToPhyrexiaReanimatesAPhyrexianAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Portal to Phyrexia", effectstest.S58PortalToPhyrexiaTypeLin, effectstest.S58PortalToPhyrexiaOracle, false)
	dragon := effectstest.PushGraveyardCardForTest(opp, "Their Dragon")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's upkeep is not your upkeep")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PickCard(t, g, me.ID, dragon)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, dragon)
	if !ok {
		t.Fatal("the creature card did not reach the battlefield")
	}
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %v owner %v, want mine and theirs", c.Controller, c.Owner)
	}
	card := effectstest.LayeredCard(t, g, dragon)
	if !card.HasSubtype("Phyrexian") || !card.HasSubtype("Test") {
		t.Errorf("subtypes of %q: want Phyrexian in addition to its other types", card.TypeLine)
	}
}

// --- The Mycosynth Gardens ------------------------------------------------

func TestTheMycosynthGardensBecomesACopyOfAnArtifactWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gardens := effectstest.PushCatalogPermanent(g, me.ID, "The Mycosynth Gardens", effectstest.S58MycosynthGardensTypes, effectstest.S58MycosynthGardensOracle, false)
	rock := effectstest.S58Rock(g, me.ID, "Test Rock", "Artifact", "{2}")
	token := effectstest.S58Rock(g, me.ID, "Token Rock", "Token Artifact", "{2}")
	theirs := effectstest.S58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")

	for _, bad := range []struct {
		target uuid.UUID
		x      int
		why    string
	}{
		{token, 2, "a token"},
		{theirs, 2, "an artifact you don't control"},
		{rock, 3, "mana value 2 with X=3"},
	} {
		if err := g.ActivateCatalogAbility(me.ID, gardens, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad.target}},
			XValue:  bad.x,
		}); !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("%s: %v, want ErrIllegalTarget", bad.why, err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, gardens, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		XValue:  2,
	}); err != nil {
		t.Fatalf("X=2 targeting a mana value 2 artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, gardens)
	if c.Name != "Test Rock" {
		t.Errorf("name = %q, want the copied artifact's", c.Name)
	}
	if !c.IsArtifact() || c.IsLand() {
		t.Errorf("type line %q: want an artifact and no longer a land", c.TypeLine)
	}
}
