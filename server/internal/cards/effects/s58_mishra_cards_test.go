package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_mishra_cards_test.go — S58 PR 4, the artifact cards Mishra's
// deck asked for (#2033).

const (
	s58MishraOracle            = "d3438037-3efd-4ce0-88ec-6d48ab521992"
	s58ArcumDagssonOracle      = "0fdfe986-0216-468b-aac8-bbcf588fd894"
	s58ArtificerClassOracle    = "6dccf583-1045-4058-86cc-ebbcc8de080e"
	s58CopyArtifactOracle      = "80bc56a9-40e0-48da-ae86-190e39c8a4a3"
	s58CybermenSquadronOracle  = "8d035688-089f-4d5c-bcad-9140a6c1681b"
	s58DiplomaticImmunityOrcl  = "f1a3153c-0200-4ff2-b7d5-48d23920bb3c"
	s58ForsakenMonumentOracle  = "7777fab1-df3f-467f-b9e2-46dd2bd2166e"
	s58KnightPaladinOracle     = "e2ab0f84-8402-472d-be66-9661b85e08fc"
	s58NautiloidShipOracle     = "613a8774-165e-4cf6-ad43-124f9ffc9980"
	s58PortalToPhyrexiaOracle  = "301d1d8e-d3fc-4010-8b29-4724fa0b31cd"
	s58MycosynthGardensOracle  = "03f5c566-825c-4c46-9c01-a2f9b1e70a13"
	s58ArtificerClassTypeLine  = "Enchantment — Class"
	s58MycosynthGardensTypes   = "Land — Sphere"
	s58MishraTypeLine          = "Legendary Creature — Human Artificer"
	s58NautiloidShipTypeLine   = "Artifact — Vehicle"
	s58PortalToPhyrexiaTypeLin = "Artifact"
)

// s58Rock seeds a plain (non-catalog) artifact with a mana cost.
func s58Rock(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// s58Named returns the battlefield cards named `name` that `controller`
// controls.
func s58Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

// s58Cast seeds a spell card into the active player's hand and casts it.
func s58Cast(t *testing.T, g *game.Game, c game.Card, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = active.ID, active.ID
	active.Hand.PushTop(c)
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, c.InstanceID, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// --- Artificer Class (CR 716) -----------------------------------------

// Level 1: the first artifact spell you cast each turn costs {1} less,
// and nothing else does.
func TestArtificerClassLevelOneDiscountsOnlyTheFirstArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	class := pushClass(g, me.ID, "Artificer Class", s58ArtificerClassTypeLine, s58ArtificerClassOracle)
	if got := levelOf(g, class); got != 1 {
		t.Fatalf("level = %d, want 1", got)
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, class))); got != 0 {
		t.Errorf("at level 1 the level-2 and level-3 triggers must not exist: got %d", got)
	}

	if got := priceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("first artifact spell: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a nonartifact spell is not discounted: %d, want 2", got)
	}
	if got := priceInHand(t, g, opp, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("an opponent's artifact spell is not discounted: %d, want 3", got)
	}

	// A nonartifact spell does not use the discount up.
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("after an instant, the first artifact spell is still discounted: %d, want 2", got)
	}

	castCatalogSpell(t, g, "First Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := g.CastTallyFor(me.ID).Artifact; got != 1 {
		t.Fatalf("artifact spells cast this turn = %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("the second artifact spell this turn: %d, want 3", got)
	}
}

// An artifact spell cast before the Class arrived was the turn's first,
// so nothing later that turn is discounted.
func TestArtificerClassCountsAnArtifactCastBeforeItArrived(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "First Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	pushClass(g, me.ID, "Artificer Class", s58ArtificerClassTypeLine, s58ArtificerClassOracle)
	if got := priceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("an artifact spell was already cast this turn: %d, want 3", got)
	}
}

// Level 2: "When this Class becomes level 2, reveal cards from the top of
// your library until you reveal an artifact card. Put that card into
// your hand and the rest on the bottom of your library in a random
// order." The level-3 line does not exist yet.
func TestArtificerClassLevelTwoRevealsUntilAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	class := pushClass(g, me.ID, "Artificer Class", s58ArtificerClassTypeLine, s58ArtificerClassOracle)
	rock := pushLibraryCardForTest(me, game.Card{Name: "Buried Rock", TypeLine: "Artifact"})
	advanceTo(t, g, game.StepPrecombatMain)
	hand, library := me.Hand.Size(), me.Library.Size()

	if err := levelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := levelOf(g, class); got != 2 {
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
	if got := len(game.TriggersForCard(layeredCard(t, g, class))); got != 1 {
		t.Errorf("at level 2 only the level-2 trigger exists: got %d", got)
	}

	// The level-3 end-step trigger does not exist at level 2.
	advanceToEndStepOf(t, g, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("the level-3 end-step trigger asked for a target at level 2")
	}
}

// A library with no artifact card reveals itself entirely and every
// card goes back; nothing reaches the hand.
func TestArtificerClassLevelTwoWithNoArtifactPutsEverythingBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	class := pushClass(g, me.ID, "Artificer Class", s58ArtificerClassTypeLine, s58ArtificerClassOracle)
	advanceTo(t, g, game.StepPrecombatMain)
	hand, library := me.Hand.Size(), me.Library.Size()
	if err := levelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	class := pushClass(g, me.ID, "Artificer Class", s58ArtificerClassTypeLine, s58ArtificerClassOracle)
	rock := s58Rock(g, me.ID, "Test Rock", "Artifact", "{2}")
	advanceTo(t, g, game.StepPrecombatMain)

	// CR 716.2a: level 3 cannot be gained from level 1.
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{5}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, class, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("level 3 from level 1 must be refused (CR 716.2a)")
	}
	if err := levelUpTo(t, g, me.ID, class, 0, "{1}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := levelUpTo(t, g, me.ID, class, 1, "{5}{U}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if got := levelOf(g, class); got != 3 {
		t.Fatalf("level = %d, want 3", got)
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, class))); got != 2 {
		t.Errorf("at level 3 both triggers exist: got %d", got)
	}
	// Level 1's discount is still there (CR 716.2a: "or greater").
	if got := priceInHand(t, g, me, "Rock", "Artifact", "{3}"); got != 2 {
		t.Errorf("level 1's discount at level 3: %d, want 2", got)
	}

	advanceToEndStepOf(t, g, 0)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	rocks := s58Named(g, me.ID, "Test Rock")
	if len(rocks) != 2 {
		t.Fatalf("Test Rocks = %d, want the original and a token copy", len(rocks))
	}
	tokens := 0
	for _, c := range rocks {
		if IsToken(c) {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("token copies = %d, want 1", tokens)
	}
}

// --- Mishra, Eminent One ------------------------------------------------

func TestMishraMakesAHastyWarformAndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Mishra, Eminent One", s58MishraTypeLine, s58MishraOracle, false)
	rock := s58Rock(g, me.ID, "Test Rock", "Artifact — Equipment", "{2}")
	bot := s58Rock(g, me.ID, "Test Bot", "Artifact Creature — Construct", "{2}")

	advanceTo(t, g, game.StepBeginCombat)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bot}); err == nil {
		t.Error("an artifact creature is not a noncreature artifact")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	warforms := s58Named(g, me.ID, "Mishra's Warform")
	if len(warforms) != 1 {
		t.Fatalf("Mishra's Warforms = %d, want 1", len(warforms))
	}
	w := warforms[0].InstanceID
	if !IsToken(warforms[0]) {
		t.Error("the Warform is a token")
	}
	if p, tough := effectivePower(t, g, w), effectiveToughness(t, g, w); p != 4 || tough != 4 {
		t.Errorf("Warform is %d/%d, want 4/4", p, tough)
	}
	card := layeredCard(t, g, w)
	if !card.IsArtifact() || !card.IsCreature() || !card.HasSubtype("Construct") || !card.HasSubtype("Equipment") {
		t.Errorf("Warform type line %q: want an Equipment Construct artifact creature", card.TypeLine)
	}
	if !containsString(effectiveAbilities(t, g, w), "haste") {
		t.Error("the Warform has haste this turn")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the original artifact stays")
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := len(s58Named(g, me.ID, "Mishra's Warform")); got != 0 {
		t.Errorf("Warforms after the end step = %d, want 0 (sacrificed)", got)
	}
}

// --- Arcum Dagsson -------------------------------------------------------

func TestArcumDagssonSacrificesThenItsControllerSearches(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arcum := pushCatalogPermanent(g, me.ID, "Arcum Dagsson", "Legendary Creature — Human Artificer", s58ArcumDagssonOracle, false)
	theirBot := s58Rock(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "{3}")
	bear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	pushLibraryCardForTest(opp, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	pushLibraryCardForTest(opp, game.Card{Name: "Their Golem", TypeLine: "Artifact Creature — Golem"})

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
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirBot) {
		t.Fatal("the artifact creature was not sacrificed")
	}
	search := searchChoiceFor(g, opp.ID)
	if search == nil {
		t.Fatal("the creature's controller is not offered the search")
	}
	if searchOptionNamed(g, search, "Their Golem") != uuid.Nil {
		t.Error("an artifact creature card is offered; the search is for a noncreature artifact")
	}
	answerSearchNamed(t, g, opp.ID, "Their Rock")
	passPriorityAroundTable(t, g)
	if got := len(s58Named(g, opp.ID, "Their Rock")); got != 1 {
		t.Errorf("Their Rock on the battlefield under its owner = %d, want 1", got)
	}
}

// --- Copy Artifact ------------------------------------------------------

func TestCopyArtifactEntersAsAnArtifactEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirRock := s58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")
	bear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := castCatalogSpell(t, g, "Copy Artifact", "Enchantment", s58CopyArtifactOracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if !hasID(pc.CopyOptions, theirRock) || hasID(pc.CopyOptions, bear) {
		t.Errorf("CopyOptions = %v, want the artifact and not the creature", pc.CopyOptions)
	}
	resolveWithCopyChoice(t, g, theirRock)

	c := layeredCard(t, g, id)
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
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bot := s58Rock(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "{3}")
	id := castCatalogSpell(t, g, "Copy Artifact", "Enchantment", s58CopyArtifactOracle, nil)
	resolveWithCopyChoice(t, g, bot)
	c := layeredCard(t, g, id)
	if !c.IsArtifact() || !c.IsCreature() || !c.IsEnchantment() || !c.HasSubtype("Construct") {
		t.Errorf("type line %q: want Artifact Creature Enchantment — Construct", c.TypeLine)
	}
}

// --- Cybermen Squadron ---------------------------------------------------

func TestCybermenSquadronGivesMyriadToNonlegendaryArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b21Push(g, me.ID, "Cybermen Squadron", "Artifact Creature — Cyberman", s58CybermenSquadronOracle, 5, 5)
	bot := pushTribalCreature(g, me.ID, "Test Bot", "Artifact Creature — Construct", 3, 3, "vigilance")

	declareAttack(t, g, opp1.ID, bot)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	attacking := map[uuid.UUID]int{}
	for _, c := range s58Named(g, me.ID, "Test Bot") {
		if c.InstanceID == bot {
			continue
		}
		if !c.Tapped || !IsToken(c) {
			t.Errorf("a myriad copy is a tapped token")
		}
		attacking[c.AttackingTarget]++
	}
	if attacking[opp2.ID] != 1 || attacking[opp3.ID] != 1 || attacking[opp1.ID] != 0 {
		t.Errorf("copies attack %v, want one at each non-defending opponent", attacking)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := len(s58Named(g, me.ID, "Test Bot")); got != 1 {
		t.Errorf("Test Bots after end of combat = %d, want the original", got)
	}
}

func TestCybermenSquadronSkipsLegendaryAndNonartifactAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b21Push(g, me.ID, "Cybermen Squadron", "Artifact Creature — Cyberman", s58CybermenSquadronOracle, 5, 5)
	legend := pushTribalCreature(g, me.ID, "Legend Bot", "Legendary Artifact Creature — Construct", 3, 3)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, legend, bear)
	if p := latestTriggerPrompt(g, me.ID); p != nil {
		t.Error("a legendary or nonartifact attacker has no myriad")
	}
}

// --- Diplomatic Immunity -------------------------------------------------

func TestDiplomaticImmunityGivesShroudAndHasIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Diplomatic Immunity", "Enchantment — Aura", s58DiplomaticImmunityOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !containsString(effectiveAbilities(t, g, bear), "shroud") {
		t.Error("the enchanted creature has shroud")
	}
	if !containsString(effectiveAbilities(t, g, aura), "shroud") {
		t.Error("the Aura itself has shroud")
	}
}

// --- Forsaken Monument ---------------------------------------------------

func TestForsakenMonumentPumpsColorlessCreaturesAndDoublesColorless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Forsaken Monument", "Legendary Artifact", s58ForsakenMonumentOracle, false)
	bot := b21Push(g, me.ID, "Test Bot", "Artifact Creature — Construct", "", 2, 2)
	bear := b21Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2, "G")
	theirBot := b21Push(g, opp.ID, "Their Bot", "Artifact Creature — Construct", "", 2, 2)
	if got := effectivePower(t, g, bot); got != 4 {
		t.Errorf("colorless creature you control: power %d, want 4", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("green creature: power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirBot); got != 2 {
		t.Errorf("an opponent's colorless creature: power %d, want 2", got)
	}

	gardens := pushCatalogPermanent(g, me.ID, "The Mycosynth Gardens", s58MycosynthGardensTypes, s58MycosynthGardensOracle, false)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	tapForMana(t, g, me.ID, gardens)
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"C", "C"}) {
		t.Errorf("a permanent tapped for {C}: pool %v, want {C}{C}", got)
	}
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); !reflect.DeepEqual(got, []string{"C", "C", "G"}) {
		t.Errorf("then a Forest: pool %v, want {C}{C}{G}", got)
	}
}

func TestForsakenMonumentGainsLifeOnColorlessSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Forsaken Monument", "Legendary Artifact", s58ForsakenMonumentOracle, false)
	life := me.Life
	s58Cast(t, g, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}}, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a green spell: life %d -> %d, want unchanged", life, me.Life)
	}
	s58Cast(t, g, game.Card{Name: "Test Rock", TypeLine: "Artifact", ManaCost: "{2}"}, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("a colorless spell: life %d -> %d, want +2", life, me.Life)
	}
}

// --- Knight Paladin ------------------------------------------------------

func TestKnightPaladinDealsFourToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Life
	}
	s58Cast(t, g, game.Card{Name: "Knight Paladin", TypeLine: "Artifact — Vehicle", OracleID: s58KnightPaladinOracle,
		ManaCost: "{5}", Power: 6, Toughness: 6}, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := before[i] - 4
		if p.ID == me.ID {
			want = before[i]
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d", i, p.Life, want)
		}
	}
	if !game.HasKeyword(&game.Card{OracleID: s58KnightPaladinOracle}, "trample") {
		t.Error("Knight Paladin prints trample")
	}
}

// --- Nautiloid Ship -------------------------------------------------------

func TestNautiloidShipExilesAGraveyardThenReturnsACreatureFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushGraveyardCardForTest(opp, "Their Dragon")
	spell := b17GraveyardCard(opp, "Their Bolt", "Instant", "{R}")
	mine := pushGraveyardCardForTest(me, "My Bear")

	ship := s58Cast(t, g, game.Card{Name: "Nautiloid Ship", TypeLine: s58NautiloidShipTypeLine, OracleID: s58NautiloidShipOracle,
		ManaCost: "{4}", Power: 5, Toughness: 5}, nil)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
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
	crewer := pushCrewerForTest(g, me.ID, "Crewer", 3)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crewer}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	attackWith(t, g, opp.ID, ship)
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choice of a creature card exiled with the Ship")
	}
	if !reflect.DeepEqual(pick.ChooseCards, []uuid.UUID{dragon}) {
		t.Errorf("offered %v, want only the creature card %v (not the instant %v)", pick.ChooseCards, dragon, spell)
	}
	answerChooseCards(t, g, me.ID, dragon)
	passPriorityAroundTable(t, g)
	if got := s58Named(g, me.ID, "Their Dragon"); len(got) != 1 {
		t.Errorf("Their Dragon under my control = %d, want 1", len(got))
	}
}

// --- Portal to Phyrexia ---------------------------------------------------

func TestPortalToPhyrexiaMakesEachOpponentSacrificeThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	myBear := pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 4; i++ {
		pushTribalCreature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	}
	pushTribalCreature(g, opp2.ID, "Lone Bear", "Creature — Bear", 2, 2)

	s58Cast(t, g, game.Card{Name: "Portal to Phyrexia", TypeLine: s58PortalToPhyrexiaTypeLin, OracleID: s58PortalToPhyrexiaOracle, ManaCost: "{9}"}, nil)
	for i := 0; i < 32; i++ {
		var c *game.PendingChoice
		for _, p := range g.Seats {
			if c = sacrificeChoiceFor(g, p.ID); c != nil {
				break
			}
		}
		if c == nil {
			if stackFullyEmpty(g) {
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
		if err := g.ResolveSacrificeChoice(c.ID, c.Chooser, c.SacrificeOptions[0]); err != nil {
			t.Fatalf("ResolveSacrificeChoice: %v", err)
		}
	}
	if got := countBattlefieldNamed(g, opp1.ID, "Their Bear"); got != 1 {
		t.Errorf("opponent with four creatures keeps %d, want 1", got)
	}
	if got := countBattlefieldNamed(g, opp2.ID, "Lone Bear"); got != 0 {
		t.Errorf("opponent with one creature keeps %d, want 0", got)
	}
	if !g.Battlefield.Contains(myBear) {
		t.Error("your own creature is not sacrificed")
	}
}

func TestPortalToPhyrexiaReanimatesAPhyrexianAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Portal to Phyrexia", s58PortalToPhyrexiaTypeLin, s58PortalToPhyrexiaOracle, false)
	dragon := pushGraveyardCardForTest(opp, "Their Dragon")

	advanceToUpkeepOf(t, g, 1)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's upkeep is not your upkeep")
	}
	advanceToUpkeepOf(t, g, 0)
	pickCard(t, g, me.ID, dragon)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, dragon)
	if !ok {
		t.Fatal("the creature card did not reach the battlefield")
	}
	if c.Controller != me.ID || c.Owner != opp.ID {
		t.Errorf("controller %v owner %v, want mine and theirs", c.Controller, c.Owner)
	}
	card := layeredCard(t, g, dragon)
	if !card.HasSubtype("Phyrexian") || !card.HasSubtype("Test") {
		t.Errorf("subtypes of %q: want Phyrexian in addition to its other types", card.TypeLine)
	}
}

// --- The Mycosynth Gardens ------------------------------------------------

func TestTheMycosynthGardensBecomesACopyOfAnArtifactWithManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gardens := pushCatalogPermanent(g, me.ID, "The Mycosynth Gardens", s58MycosynthGardensTypes, s58MycosynthGardensOracle, false)
	rock := s58Rock(g, me.ID, "Test Rock", "Artifact", "{2}")
	token := s58Rock(g, me.ID, "Token Rock", "Token Artifact", "{2}")
	theirs := s58Rock(g, opp.ID, "Their Rock", "Artifact", "{2}")

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
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, gardens)
	if c.Name != "Test Rock" {
		t.Errorf("name = %q, want the copied artifact's", c.Name)
	}
	if !c.IsArtifact() || c.IsLand() {
		t.Errorf("type line %q: want an artifact and no longer a land", c.TypeLine)
	}
}
