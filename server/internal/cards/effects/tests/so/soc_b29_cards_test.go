package testsso

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b29_cards_test.go — Secrets of Strixhaven, batch 29 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB29FlourishOracle    = "1cb419cc-d8e4-4676-987f-089cebccc374"
	socB29MatterbendOracle  = "94688e25-28cb-45aa-95ea-0d6f547cae57"
	socB29PoetOracle        = "37731c49-1238-4e0e-8993-204e36c1f4ee"
	socB29MicaOracle        = "fc90f74d-11f3-4bd4-b8a1-1cf2ffcd5b73"
	socB29BiomancerOracle   = "03af9d24-2df6-4318-82f2-8abdd22206f1"
	socB29MindIntoOracle    = "9f9e4be0-814a-4cab-948b-114769035a4a"
	socB29MindRootsOracle   = "5d32dbf3-c157-4543-8813-da88f463de0e"
	socB29MoltenNoteOracle  = "8c179b8f-3a83-49d6-9e7f-9e17c293ccbb"
	socB29ReckoningOracle   = "8e9576c1-b0c2-4be5-9005-35b51164f78e"
	socB29BearType          = "Creature — Bear"
	socB29InstantTypeLine   = "Instant"
	socB29SorceryTypeLine   = "Sorcery"
	socB29ArtifactTypeLine  = "Artifact"
	socB29LandTypeLine      = "Basic Land — Forest"
	socB29EnchantmentType   = "Enchantment"
	socB29ReckoningManaCost = "{3}{W}{W}{B}{B}"
)

func TestSocB29EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB29FlourishOracle:   "Masterful Flourish",
		socB29MatterbendOracle: "Matterbending Mage",
		socB29PoetOracle:       "Melancholic Poet",
		socB29MicaOracle:       "Mica, Reader of Ruins",
		socB29BiomancerOracle:  "Mindful Biomancer",
		socB29MindIntoOracle:   "Mind into Matter",
		socB29MindRootsOracle:  "Mind Roots",
		socB29MoltenNoteOracle: "Molten Note",
		socB29ReckoningOracle:  "Moment of Reckoning",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB29HasKeyword reads a keyword through the layered view.
func socB29HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, keyword string) bool {
	t.Helper()
	v := effectstest.S58p6View(t, g, id)
	return game.HasKeyword(&v, keyword)
}

// --- Masterful Flourish ------------------------------------------------------

func TestSocB29FlourishGivesPlusOneAndIndestructibleUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB29BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Masterful Flourish", socB29InstantTypeLine, socB29FlourishOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bear); p != 3 || tg != 2 {
		t.Errorf("%d/%d after the spell, want 3/2", p, tg)
	}
	if !socB29HasKeyword(t, g, bear, "indestructible") {
		t.Error("the Bear is not indestructible")
	}
	seq := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == seq; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if p, _ := effectstest.SocB14PT(t, g, bear); p != 2 {
		t.Errorf("power %d next turn, want the +1/+0 gone", p)
	}
	if socB29HasKeyword(t, g, bear, "indestructible") {
		t.Error("indestructible lasted past the turn")
	}
}

func TestSocB29FlourishSurvivesLethalDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB29BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Masterful Flourish", socB29InstantTypeLine, socB29FlourishOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].DamageMarked = 5
			}
		}
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(bear) {
		t.Error("an indestructible creature died to damage")
	}
}

func TestSocB29FlourishRefusesACreatureYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Ogre", "Creature — Ogre", 3, 3)
	err := effectstest.CastCatalogSpellErr(t, g, "Masterful Flourish", socB29InstantTypeLine, socB29FlourishOracle, effectstest.CardRefs(theirs))
	if err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
}

// --- Matterbending Mage -------------------------------------------------------

func TestSocB29MatterbendingMageBouncesAnotherCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastAndResolveCreature(t, g, "Matterbending Mage", "Creature — Human Wizard", socB29MatterbendOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Errorf("the Ogre is on the battlefield %v and in its owner's hand %v, want hand", g.Battlefield.Contains(theirs), opp.Hand.Contains(theirs))
	}
}

func TestSocB29MatterbendingMageCanBounceYourOwnOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Bear", socB29BearType, 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Matterbending Mage", "Creature — Human Wizard", socB29MatterbendOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("the Bear did not return to my hand")
	}
}

func TestSocB29MatterbendingMageCannotTargetItselfAndTargetsNothingByChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	mage := effectstest.CastAndResolveCreature(t, g, "Matterbending Mage", "Creature — Human Wizard", socB29MatterbendOracle)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mage}); err == nil {
		t.Error("the Mage was accepted as its own target for \"other target creature\"")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("declining the \"up to one\": %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(mage) {
		t.Error("something left the battlefield with no target chosen")
	}
}

func TestSocB29MatterbendingMageEntersQuietlyWithNoOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mage := effectstest.CastAndResolveCreature(t, g, "Matterbending Mage", "Creature — Human Wizard", socB29MatterbendOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mage) {
		t.Error("the Mage is not on the battlefield")
	}
}

// socB29MageAttacks puts a Mage on the battlefield, casts the given spell and
// reports whether an opposing creature can block the Mage.
func socB29MageCanBeBlockedAfter(t *testing.T, cast func(t *testing.T, g *game.Game, me *game.Player)) bool {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Matterbending Mage", "Creature — Human Wizard", socB29MatterbendOracle, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", socB29BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	cast(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mage, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, mage) == nil
}

func TestSocB29MatterbendingMageCantBeBlockedAfterYouCastASpellWithX(t *testing.T) {
	blockable := socB29MageCanBeBlockedAfter(t, func(t *testing.T, g *game.Game, me *game.Player) {
		c := game.Card{Name: "X Spell", TypeLine: socB29SorceryTypeLine, ManaCost: "{X}{R}"}
		if err := effectstest.SocB19CastWithPool(g, me, c, "R", 2, game.CastSpellParams{XValue: 1}); err != nil {
			t.Fatalf("cast: %v", err)
		}
	})
	if blockable {
		t.Error("the Mage was blocked after its controller cast a spell with {X}")
	}
}

func TestSocB29MatterbendingMageStaysBlockableAfterASpellWithoutX(t *testing.T) {
	blockable := socB29MageCanBeBlockedAfter(t, func(t *testing.T, g *game.Game, me *game.Player) {
		c := game.Card{Name: "Plain Spell", TypeLine: socB29SorceryTypeLine, ManaCost: "{1}{R}"}
		if err := effectstest.SocB19CastWithPool(g, me, c, "R", 2, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
	})
	if !blockable {
		t.Error("a spell without {X} made the Mage unblockable")
	}
}

func TestSocB29MatterbendingMageIsBlockableWithNoSpellCast(t *testing.T) {
	if !socB29MageCanBeBlockedAfter(t, func(*testing.T, *game.Game, *game.Player) {}) {
		t.Error("the Mage cannot be blocked with no X spell cast")
	}
}

// --- Melancholic Poet -----------------------------------------------------------

func TestSocB29PoetDrainsWhenYourSpellTargetsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Melancholic Poet", "Creature — Elf Bard", socB29PoetOracle, 2, 2)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	life := me.Life
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Locust Spray", socB29InstantTypeLine, effectstest.SocB17LocustSprayOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("my life %d -> %d, want +1", life, me.Life)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != before[p.ID]-1 {
			t.Errorf("opponent life %d -> %d, want -1", before[p.ID], p.Life)
		}
	}
}

func TestSocB29PoetIgnoresSpellsThatDoNotTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Melancholic Poet", "Creature — Elf Bard", socB29PoetOracle, 2, 2)
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Plain Instant", socB29InstantTypeLine, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear Cub", socB29BearType, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("life changed (%d -> %d, %d -> %d) with no creature-targeting instant or sorcery", life, me.Life, oppLife, opp.Life)
	}
}

// --- Mica, Reader of Ruins --------------------------------------------------------

func socB29MicaBoard(g *game.Game, me *game.Player, artifacts int) (mica uuid.UUID, rocks []uuid.UUID) {
	mica = effectstest.B12Push(g, me.ID, "Mica, Reader of Ruins", "Legendary Creature — Human Artificer", socB29MicaOracle, 4, 4)
	for i := 0; i < artifacts; i++ {
		rocks = append(rocks, effectstest.PushTypedCard(g, me.ID, "Rock", socB29ArtifactTypeLine, "{1}"))
	}
	return mica, rocks
}

func TestSocB29MicaSacrificesAnArtifactToCopyAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	_, rocks := socB29MicaBoard(g, me, 1)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, me.ID, rocks); err != nil {
		t.Fatalf("sacrifice the artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rocks[0]) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d cards, want 4: the Divination and its copy", got)
	}
}

func TestSocB29MicaDecliningTheSacrificeCopiesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	_, rocks := socB29MicaBoard(g, me, 1)
	hand := me.Hand.Size()
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(prompt.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rocks[0]) {
		t.Error("the artifact was sacrificed anyway")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2: only the Divination", got)
	}
}

func TestSocB29MicaAsksNothingWithoutAnArtifactAndIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	socB29MicaBoard(g, me, 0)
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("asked for an artifact to sacrifice with none in play")
	}
	effectstest.PushTypedCard(g, me.ID, "Rock", socB29ArtifactTypeLine, "{1}")
	effectstest.CastCatalogSpell(t, g, "Bear Cub", socB29BearType, "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("a creature spell offered the copy")
	}
}

func TestSocB29MicaWardAsksThreeLife(t *testing.T) {
	g, _, opp, mica := effectstest.WardTable(t, "Mica, Reader of Ruins", "Legendary Creature — Human Artificer", socB29MicaOracle)
	blade := effectstest.CastAtWardedCreature(t, g, opp, mica)
	life := opp.Life
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if prompt.LifeCost != 3 {
		t.Errorf("prompt LifeCost = %d, want 3", prompt.LifeCost)
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mica) || !opp.Graveyard.Contains(blade) || opp.Life != life {
		t.Error("declining the ward must counter the spell and cost nothing")
	}
}

// --- Mindful Biomancer ---------------------------------------------------------------

func TestSocB29BiomancerGainsOneLifeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Mindful Biomancer", "Creature — Dryad Druid", socB29BiomancerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestSocB29BiomancerPumpsOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bio := effectstest.B12Push(g, me.ID, "Mindful Biomancer", "Creature — Dryad Druid", socB29BiomancerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "GGGGGG")
	if err := g.ActivateCatalogAbility(me.ID, bio, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.SocB14PT(t, g, bio); p != 4 || tg != 4 {
		t.Fatalf("%d/%d after one activation, want 4/4", p, tg)
	}
	if err := g.ActivateCatalogAbility(me.ID, bio, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a second activation in the same turn was accepted")
	}
	if p, tg := effectstest.SocB14PT(t, g, bio); p != 4 || tg != 4 {
		t.Errorf("%d/%d after a refused second activation, want 4/4", p, tg)
	}
}

func TestSocB29BiomancerPumpIsRefusedWithoutItsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bio := effectstest.B12Push(g, me.ID, "Mindful Biomancer", "Creature — Dryad Druid", socB29BiomancerOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRR")
	if err := g.ActivateCatalogAbility(me.ID, bio, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid {2}{G} with no green mana")
	}
}

// --- Mind into Matter ---------------------------------------------------------------------

func socB29CastMindIntoMatter(t *testing.T, g *game.Game, me *game.Player, x int) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mind into Matter", TypeLine: socB29SorceryTypeLine, ManaCost: "{X}{G}{U}",
		OracleID: socB29MindIntoOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "GU"+strings.Repeat("G", x))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x}); err != nil {
		t.Fatalf("cast with X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB29MindIntoMatterDrawsXThenPutsAPermanentTappedAtOrBelowX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	cheap := effectstest.HandCardForTest(me, "Cheap Bear", socB29BearType, "")
	dear := effectstest.HandCardForTest(me, "Dear Bear", socB29BearType, "")
	spell := effectstest.HandCardForTest(me, "Spell", socB29SorceryTypeLine, "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			switch me.Hand.Cards[i].InstanceID {
			case cheap:
				me.Hand.Cards[i].ManaCost = "{1}{G}"
			case dear:
				me.Hand.Cards[i].ManaCost = "{3}{G}"
			case spell:
				me.Hand.Cards[i].ManaCost = "{G}"
			}
		}
	})
	library := me.Library.Size()
	socB29CastMindIntoMatter(t, g, me, 2)
	if got := library - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want X=2", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put prompt")
	}
	if !effectstest.HasID(pick.ChooseCards, cheap) {
		t.Error("the mana value 2 creature card is not offered")
	}
	if effectstest.HasID(pick.ChooseCards, dear) {
		t.Error("a mana value 4 card is offered for X=2")
	}
	if effectstest.HasID(pick.ChooseCards, spell) {
		t.Error("a sorcery card is offered: it is not a permanent card")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) || !effectstest.B16Tapped(t, g, cheap) {
		t.Error("the chosen card is not on the battlefield tapped")
	}
	if me.Hand.Contains(cheap) {
		t.Error("the chosen card is still in hand")
	}
}

func TestSocB29MindIntoMatterCanPutACardJustDrawnAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := effectstest.SeedLibraryTop(me, "Forest", socB29LandTypeLine, "")
	socB29CastMindIntoMatter(t, g, me, 1)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, land) {
		t.Fatal("the land drawn this way is not offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Hand.Contains(land) {
		t.Error("declining should leave the land in hand")
	}
}

func TestSocB29MindIntoMatterWithXZeroPutsALandAndAsksNothingWithoutACandidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := effectstest.HandCardForTest(me, "Forest", socB29LandTypeLine, "")
	socB29CastMindIntoMatter(t, g, me, 0)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || !effectstest.B16Tapped(t, g, land) {
		t.Error("the land is not on the battlefield tapped")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	g2.WithWriteLock(func() { me2.Hand.Cards = nil })
	effectstest.HandCardForTest(me2, "Spell", socB29SorceryTypeLine, "")
	socB29CastMindIntoMatter(t, g2, me2, 0)
	if effectstest.ChooseCardsChoiceFor(g2, me2.ID) != nil {
		t.Error("asked to put a card with no permanent card in hand")
	}
}

// --- Mind Roots -------------------------------------------------------------------------------

func socB29CastMindRoots(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mind Roots", TypeLine: socB29SorceryTypeLine, ManaCost: "{1}{B}{G}",
		OracleID: socB29MindRootsOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "BGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func socB29SetHand(p *game.Player) { p.Hand.Cards = nil }

func TestSocB29MindRootsDiscardsTwoAndPutsADiscardedLandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { socB29SetHand(opp) })
	land := effectstest.HandCardForTest(opp, "Forest", socB29LandTypeLine, "")
	bear := effectstest.HandCardForTest(opp, "Bear", socB29BearType, "")
	keep := effectstest.HandCardForTest(opp, "Keeper", socB29SorceryTypeLine, "")
	socB29CastMindRoots(t, g, me, opp.ID)
	effectstest.AnswerDiscard(t, g, opp.ID, land, bear)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the caster was not offered the discarded land")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("pick %d..%d, want up to one", pick.ChooseMin, pick.ChooseMax)
	}
	if effectstest.HasID(pick.ChooseCards, bear) || effectstest.HasID(pick.ChooseCards, keep) {
		t.Error("a nonland card is offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok {
		t.Fatal("the land is not on the battlefield")
	}
	if c.Controller != me.ID || !c.Tapped {
		t.Errorf("the land is controlled by %v, tapped=%v: want mine, tapped", c.Controller == me.ID, c.Tapped)
	}
	if !opp.Graveyard.Contains(bear) || !opp.Hand.Contains(keep) {
		t.Error("the Bear should be in their graveyard and the third card still in hand")
	}
}

func TestSocB29MindRootsMayDeclineTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { socB29SetHand(opp) })
	land := effectstest.HandCardForTest(opp, "Forest", socB29LandTypeLine, "")
	bear := effectstest.HandCardForTest(opp, "Bear", socB29BearType, "")
	socB29CastMindRoots(t, g, me, opp.ID)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, land, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !opp.Graveyard.Contains(land) {
		t.Error("declining should leave the land in the graveyard")
	}
}

func TestSocB29MindRootsOffersOnlyOneOfTwoDiscardedLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { socB29SetHand(opp) })
	a := effectstest.HandCardForTest(opp, "Forest", socB29LandTypeLine, "")
	b := effectstest.HandCardForTest(opp, "Forest", socB29LandTypeLine, "")
	socB29CastMindRoots(t, g, me, opp.ID)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, a, b)
		effectstest.PassPriorityAroundTable(t, g)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) {
		t.Fatal("both discarded lands should be offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Fatal("two lands were accepted for \"up to one\"")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(b) || g.Battlefield.Contains(a) || !opp.Graveyard.Contains(a) {
		t.Error("exactly the chosen land should have entered")
	}
}

func TestSocB29MindRootsWithNoLandJustDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { socB29SetHand(opp) })
	a := effectstest.HandCardForTest(opp, "Bear", socB29BearType, "")
	b := effectstest.HandCardForTest(opp, "Ogre", socB29BearType, "")
	socB29CastMindRoots(t, g, me, opp.ID)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, a, b)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("offered a land when none was discarded")
	}
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("both cards should have been discarded")
	}
}

func TestSocB29MindRootsAgainstAHandOfOneLandStillOffersIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { socB29SetHand(opp) })
	land := effectstest.HandCardForTest(opp, "Forest", socB29LandTypeLine, "")
	socB29CastMindRoots(t, g, me, opp.ID)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Error("the only card in their hand, a land, was discarded and should have been put onto the battlefield")
	}
}

func TestSocB29MindRootsCanTargetYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { socB29SetHand(me) })
	land := effectstest.HandCardForTest(me, "Forest", socB29LandTypeLine, "")
	bear := effectstest.HandCardForTest(me, "Bear", socB29BearType, "")
	socB29CastMindRoots(t, g, me, me.ID)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		effectstest.AnswerDiscard(t, g, me.ID, land, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) || !me.Graveyard.Contains(bear) {
		t.Error("discarding to myself should put the land onto the battlefield and the Bear in the graveyard")
	}
}

// --- Molten Note --------------------------------------------------------------------------------

func socB29CastMoltenNote(t *testing.T, g *game.Game, me *game.Player, x int, target uuid.UUID) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Molten Note", TypeLine: socB29SorceryTypeLine, ManaCost: "{X}{R}{W}",
		OracleID: socB29MoltenNoteOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "RW"+strings.Repeat("R", x))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(target)}); err != nil {
		t.Fatalf("cast with X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB29MoltenNoteDealsDamageEqualToTheManaSpentAndUntapsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 9)
	mine := effectstest.B12Creature(g, me.ID, "Bear", socB29BearType, 2, 2)
	theirTapped := effectstest.B12Creature(g, opp.ID, "Tapped Ogre", "Creature — Ogre", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == mine || id == theirTapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	socB29CastMoltenNote(t, g, me, 3, ogre)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 5 {
		t.Errorf("damage %d, want 5: X=3 plus {R}{W}", got)
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("my creature stayed tapped")
	}
	if !effectstest.B16Tapped(t, g, theirTapped) {
		t.Error("an opponent's creature was untapped")
	}
}

func TestSocB29MoltenNoteWithXZeroDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 9)
	socB29CastMoltenNote(t, g, me, 0, ogre)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
}

func TestSocB29MoltenNoteFlashbackDealsEightAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 12)
	mine := effectstest.B12Creature(g, me.ID, "Bear", socB29BearType, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mine {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Molten Note", TypeLine: socB29SorceryTypeLine, ManaCost: "{X}{R}{W}",
		OracleID: socB29MoltenNoteOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "RWRRRRRR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback",
		Strict: true, Targets: effectstest.CardRefs(ogre)}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 8 {
		t.Errorf("damage %d, want 8: {6}{R}{W} was the mana spent", got)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back Molten Note returned to the graveyard instead of being exiled")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("the flashback did not untap my creature")
	}
}

func TestSocB29MoltenNoteRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.PushTypedCard(g, g.Seats[1].ID, "Rock", socB29ArtifactTypeLine, "{1}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Molten Note", socB29SorceryTypeLine, socB29MoltenNoteOracle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("targeted a noncreature artifact")
	}
}

// --- Moment of Reckoning ---------------------------------------------------------------------------

func socB29CastReckoning(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Moment of Reckoning", TypeLine: socB29SorceryTypeLine, ManaCost: socB29ReckoningManaCost,
		OracleID: socB29ReckoningOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestSocB29ReckoningDestroysAndReturnsInOneCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	b := effectstest.B12Permanent(g, opp.ID, "Rock", socB29ArtifactTypeLine)
	back := effectstest.GraveyardCardOf(g, me, "Old Stone", socB29ArtifactTypeLine, "{2}", "")
	if err := socB29CastReckoning(t, g, []int{0, 0, 1}, []game.TargetRef{
		effectstest.ModeTargetRef(a, 0), effectstest.ModeTargetRef(b, 1), effectstest.ModeTargetRef(back, 2),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a destroy mode left its target alive")
	}
	c, ok := effectstest.BattlefieldCard(g, back)
	if !ok || c.Controller != me.ID {
		t.Error("the card in my graveyard did not return under my control")
	}
}

func TestSocB29ReckoningChoosesUpToFourIncludingTheSameModeRepeatedly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	var victims []uuid.UUID
	var targets []game.TargetRef
	for i := 0; i < 4; i++ {
		id := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		victims = append(victims, id)
		targets = append(targets, effectstest.ModeTargetRef(id, i))
	}
	if err := socB29CastReckoning(t, g, []int{0, 0, 0, 0}, targets); err != nil {
		t.Fatalf("four destroys: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for _, id := range victims {
		if g.Battlefield.Contains(id) {
			t.Error("one of the four Ogres survived")
		}
	}
}

func TestSocB29ReckoningRefusesAFifthMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	var targets []game.TargetRef
	for i := 0; i < 5; i++ {
		targets = append(targets, effectstest.ModeTargetRef(effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3), i))
	}
	if err := socB29CastReckoning(t, g, []int{0, 0, 0, 0, 0}, targets); err == nil {
		t.Fatal("five modes were accepted for \"choose up to four\"")
	}
}

func TestSocB29ReckoningCanChooseASingleMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := socB29CastReckoning(t, g, []int{0}, []game.TargetRef{effectstest.ModeTargetRef(ogre, 0)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre was not destroyed")
	}
}

func TestSocB29ReckoningWillNotDestroyALandOrReturnALandOrCreatureFromElsewhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", socB29LandTypeLine)
	if err := socB29CastReckoning(t, g, []int{0}, []game.TargetRef{effectstest.ModeTargetRef(land, 0)}); err == nil {
		t.Fatal("a land was accepted as \"target nonland permanent\"")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	theirs := effectstest.GraveyardCardOf(g2, opp2, "Their Stone", socB29ArtifactTypeLine, "{2}", "")
	if err := socB29CastReckoning(t, g2, []int{1}, []game.TargetRef{effectstest.ModeTargetRef(theirs, 0)}); err == nil {
		t.Fatal("a card in an opponent's graveyard was accepted")
	}
	landCard := effectstest.GraveyardCardOf(g2, me2, "Forest", socB29LandTypeLine, "", "")
	if err := socB29CastReckoning(t, g2, []int{1}, []game.TargetRef{effectstest.ModeTargetRef(landCard, 0)}); err == nil {
		t.Fatal("a land card was accepted as \"target nonland permanent card\"")
	}
}

func TestSocB29ReckoningTheSameCardChosenTwiceReturnsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	back := effectstest.GraveyardCardOf(g, me, "Old Stone", socB29ArtifactTypeLine, "{2}", "")
	if err := socB29CastReckoning(t, g, []int{1, 1}, []game.TargetRef{effectstest.ModeTargetRef(back, 0), effectstest.ModeTargetRef(back, 1)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Old Stone" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Old Stones on the battlefield, want 1", n)
	}
}
