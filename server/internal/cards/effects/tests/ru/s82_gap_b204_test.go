package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b204_test.go — cards missing from sets since AFR, batch 204
// (#5948). Each test plays the real catalog entry and asserts what its text
// says.

const (
	gapB204Dreamshaper = "d06ef178-f02a-45a4-8193-140dee52c99d"
	gapB204Gargadon    = "406b4b1d-756b-4062-897d-3693832086a0"
	gapB204Champion    = "e1a22f41-0543-4e85-8fb8-93c018b461aa"
	gapB204Emissary    = "d78d7fd6-778b-4030-a776-29ddce83e33f"
	gapB204Pyromancer  = "02238355-9c84-4ae0-b850-269f460144a8"
	gapB204Arjun       = "ea904852-6322-4a24-9145-190acfce54ec"
	gapB204Bloodwater  = "94ca5db6-86e1-42b4-b776-977efa2e12b6"
	gapB204Duplicate   = "724767ca-bd84-4137-9133-4ad2e96b8428"
	gapB204Guildmage   = "08bb1f33-b232-4382-a268-4804d4e72d29"
	gapB204Mindwrack   = "c24b3135-5f23-461c-9e57-de747ca8978a"
)

func TestGapB204EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB204Dreamshaper: "Dreamshaper Shaman",
		gapB204Gargadon:    "Greater Gargadon",
		gapB204Champion:    "Labyrinth Champion",
		gapB204Emissary:    "Purphoros's Emissary",
		gapB204Pyromancer:  "Seasoned Pyromancer",
		gapB204Arjun:       "Arjun, the Shifting Flame",
		gapB204Bloodwater:  "Bloodwater Entity",
		gapB204Duplicate:   "Dack's Duplicate",
		gapB204Guildmage:   "League Guildmage",
		gapB204Mindwrack:   "Mindwrack Liege",
	} {
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

// gapB204Colour sets a battlefield card's colours (a pushed test card has no mana cost).
func gapB204Colour(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
		g.BumpLayerVersionForTest()
	})
}

// --- Labyrinth Champion ------------------------------------------------------

func TestGapB204LabyrinthChampionShootsWhenYouTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Labyrinth Champion", "Creature — Human Warrior", gapB204Champion, 2, 6)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: champ})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent's life %d -> %d, want -2", life, opp.Life)
	}
	// An opponent's spell at it is not heroic.
	life = opp.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: champ})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Error("an opponent's spell triggered heroic")
	}
}

// --- Purphoros's Emissary ----------------------------------------------------

func TestGapB204PurphorossEmissaryBestowsPlusThreeAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.BestowCard(t, g, "Purphoros's Emissary", "Enchantment Creature — Ox", gapB204Emissary, 3, 3)
	if alt := game.AlternativeCostByKey(gapB204Emissary, game.BestowKey); alt == nil || !alt.Bestow || alt.ManaCost != "{6}{R}" {
		t.Fatalf("bestow offer %+v", alt)
	}
	effectstest.CastBestowed(t, g, id, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("enchanted bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("the enchanted creature should have menace")
	}
}

func TestGapB204PurphorossEmissaryAsACreatureHasMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Purphoros's Emissary", "Enchantment Creature — Ox", gapB204Emissary, 3, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "menace") {
		t.Error("Purphoros's Emissary should have menace")
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("Purphoros's Emissary is %d/%d, want 3/3", p, tough)
	}
}

// --- Bloodwater Entity -------------------------------------------------------

func TestGapB204BloodwaterEntityPutsAnInstantOrSorceryOnTopOfYourLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bolt := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	id, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Bloodwater Entity", TypeLine: "Creature — Elemental", OracleID: gapB204Bloodwater, Power: 2, Toughness: 2}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickIfAsked(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	answerGapB204May(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() == 0 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != bolt {
		t.Error("the instant is not on top of the library")
	}
	if me.Graveyard.Contains(bolt) {
		t.Error("the instant is still in the graveyard")
	}
	for _, kw := range []string{"flying", "prowess"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("Bloodwater Entity does not have %s", kw)
		}
	}
}

func answerGapB204May(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	if effectstest.LatestConfirmFor(g, who) != nil {
		effectstest.AnswerMayChoice(t, g, who, yes)
		return
	}
	if prompt := effectstest.AwaitTriggerPrompt(t, g, who); prompt != nil {
		if err := g.ResolveTriggerPrompt(prompt.ID, who, yes); err != nil {
			t.Fatalf("ResolveTriggerPrompt: %v", err)
		}
	}
}

// --- Mindwrack Liege ---------------------------------------------------------

func TestGapB204MindwrackLiegeGivesBlueAndRedCreaturesPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Mindwrack Liege", "Creature — Horror", gapB204Mindwrack, 4, 4)
	mk := func(owner uuid.UUID, colors ...string) uuid.UUID {
		id := effectstest.B12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
		gapB204Colour(g, id, colors...)
		return id
	}
	blue, red, both, green, theirs := mk(me.ID, "U"), mk(me.ID, "R"), mk(me.ID, "U", "R"), mk(me.ID, "G"), mk(opp.ID, "U")
	for _, want := range []struct {
		id  uuid.UUID
		who string
		p   int
		t   int
	}{
		{liege, "the Liege", 4, 4}, {blue, "blue", 3, 3}, {red, "red", 3, 3}, {both, "blue and red", 4, 4},
		{green, "green", 2, 2}, {theirs, "an opponent's blue creature", 2, 2},
	} {
		if p, tough := effectstest.PtOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", want.who, p, tough, want.p, want.t)
		}
	}
}

func TestGapB204MindwrackLiegePutsABlueOrRedCreatureCardFromHandOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	liege := effectstest.B12Push(g, me.ID, "Mindwrack Liege", "Creature — Horror", gapB204Mindwrack, 4, 4)
	effectstest.AdvanceToMain(t, g)
	handCard := func(name, typeLine string, colors ...string) uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		return id
	}
	green := handCard("Green Bear", "Creature — Bear", "G")
	blueSpell := handCard("Blue Instant", "Instant", "U")
	red := handCard("Red Goblin", "Creature — Goblin", "R")

	effectstest.FloatForTest(g, me, "UUR")
	if err := g.ActivateCatalogAbility(me.ID, liege, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three mana")
	}
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, liege, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick from the hand")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[red] || offered[green] || offered[blueSpell] || len(offered) != 1 {
		t.Fatalf("offered %v, want only the red creature card", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{red}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effects.OnBattlefield(g, red) {
		t.Error("the red creature card did not enter the battlefield")
	}
}

// --- League Guildmage --------------------------------------------------------

func TestGapB204LeagueGuildmageDrawsForFourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gm := effectstest.B12Push(g, me.ID, "League Guildmage", "Creature — Human Wizard", gapB204Guildmage, 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.AddPool(t, g, me.ID, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three mana")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestGapB204LeagueGuildmageCopiesASpellWhoseManaValueIsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := effectstest.B12Push(g, me.ID, "League Guildmage", "Creature — Human Wizard", gapB204Guildmage, 2, 2)
	bolt := effectstest.CastCopyTarget(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	life := opp.Life

	// X = 2 does not match a mana value of 1 (a free Bolt in this fixture has
	// mana value 1 as printed).
	boltCard, _ := g.LookupCardForEffect(bolt)
	x := boltCard.ManaValue()
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, XValue: x + 1, Targets: effectstest.CardRefs(bolt)}); err == nil {
		t.Fatal("copied a spell whose mana value is not X")
	}
	effectstest.AddPool(t, g, me.ID, "{R}")
	for i := 0; i < x; i++ {
		effectstest.AddPool(t, g, me.ID, "{C}")
	}
	if err := g.ActivateCatalogAbility(me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, XValue: x, Targets: effectstest.CardRefs(bolt)}); err != nil {
		t.Fatalf("copy the Bolt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.CopyPromptFor(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("the opponent lost %d, want 6 (the spell and its copy)", got)
	}
}

// --- Seasoned Pyromancer -----------------------------------------------------

func TestGapB204SeasonedPyromancerDiscardsDrawsAndMakesAnElementalPerNonlandDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	land := effectstest.PushCatalogHandCard(me, "A Forest", "Basic Land — Forest", "")
	spell := effectstest.PushCatalogHandCard(me, "A Spell", "Sorcery", "")
	keep := effectstest.PushCatalogHandCard(me, "Keep", "Instant", "")
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Seasoned Pyromancer", TypeLine: "Creature — Human Shaman", OracleID: gapB204Pyromancer, Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, land, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(land) || !me.Graveyard.Contains(spell) || !me.Hand.Contains(keep) {
		t.Error("the chosen cards were not the ones discarded")
	}
	if me.Hand.Size() != 3 {
		t.Errorf("hand is %d cards, want 3 (kept one, drew two)", me.Hand.Size())
	}
	if n := effectstest.TokensNamedControlledBy(g, "Elemental", me.ID); n != 1 {
		t.Errorf("%d Elemental tokens, want 1 (one nonland card discarded)", n)
	}
}

func TestGapB204SeasonedPyromancerWithAnEmptyHandStillDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = me.Hand.Cards[:0]
	effectstest.Msc17Cast(t, g, me, game.Card{Name: "Seasoned Pyromancer", TypeLine: "Creature — Human Shaman", OracleID: gapB204Pyromancer, Power: 2, Toughness: 2}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 2 {
		t.Errorf("hand is %d cards, want 2 drawn", me.Hand.Size())
	}
	if n := effectstest.TokensNamedControlledBy(g, "Elemental", me.ID); n != 0 {
		t.Errorf("%d Elementals with nothing discarded, want 0", n)
	}
}

func TestGapB204SeasonedPyromancerMakesTwoElementalsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Seasoned Pyromancer", TypeLine: "Creature — Human Shaman", OracleID: gapB204Pyromancer,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with four mana")
	}
	effectstest.AddPool(t, g, me.ID, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Elemental", me.ID); n != 2 {
		t.Errorf("%d Elementals, want 2", n)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the card is still in the graveyard after being exiled as a cost")
	}
}

// --- Arjun, the Shifting Flame -----------------------------------------------

func TestGapB204ArjunBottomsYourHandAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arjun := effectstest.B12Push(g, me.ID, "Arjun, the Shifting Flame", "Legendary Creature — Sphinx Wizard", gapB204Arjun, 5, 5)
	if !effectstest.HasKeywordOnBattlefield(t, g, arjun, "flying") {
		t.Error("Arjun does not have flying")
	}
	effectstest.AdvanceToMain(t, g)
	me.Hand.Cards = me.Hand.Cards[:0]
	var old []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		old = append(old, effectstest.PushCatalogHandCard(me, n, "Instant", ""))
	}
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	order := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("the order prompt for the bottom of the library was not asked")
	}
	if len(order.ScryCards) != 3 {
		t.Fatalf("the pile is %d cards, want the three in hand", len(order.ScryCards))
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range old {
		if me.Hand.Contains(id) {
			t.Error("one of the old cards is still in hand")
		}
	}
	if me.Hand.Size() != 3 {
		t.Errorf("hand is %d cards, want 3 (three put on the bottom, three drawn)", me.Hand.Size())
	}
	bottom := me.Library.Cards[:3]
	got := 0
	for _, c := range bottom {
		for _, id := range old {
			if c.InstanceID == id {
				got++
			}
		}
	}
	if got != 3 {
		t.Errorf("%d of the old cards are at the bottom of the library, want 3", got)
	}
}

// --- Dack's Duplicate --------------------------------------------------------

func TestGapB204DacksDuplicateCopiesACreatureWithHasteAndDethrone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Dack's Duplicate", "Creature — Shapeshifter", gapB204Duplicate, nil)
	effectstest.ResolveWithCopyChoice(t, g, bear)
	g.RunStateChecksForTest()
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Grizzly Bears" || c.CurrentPower() != 2 {
		t.Fatalf("copy = %q %d/%d, want a Grizzly Bears 2/2", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("the copy does not have haste")
	}
	// Dethrone: attacking the life leader (tied) puts a +1/+1 counter on it.
	effectstest.LciB1Attack(t, g, id, opp.ID)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("the copy has %d +1/+1 counters after attacking the life leader, want 1", got)
	}
	// The original has neither.
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the original gained haste")
	}
}

// --- Dreamshaper Shaman ------------------------------------------------------

func TestGapB204DreamshaperShamanSacrificesAndRevealsUntilANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dreamshaper Shaman", "Enchantment Creature — Minotaur Shaman", gapB204Dreamshaper, 5, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	// Library from the top: a land, a sorcery, a creature, a land.
	me.Library.Cards = me.Library.Cards[:0]
	bottomLand := effectstest.TopLibraryCard(me, "Bottom Land", "Basic Land — Forest")
	hit := effectstest.TopLibraryCard(me, "Hit Creature", "Creature — Bear")
	sorcery := effectstest.TopLibraryCard(me, "Sorcery", "Sorcery")
	land := effectstest.TopLibraryCard(me, "Top Land", "Basic Land — Forest")
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerMayCounter(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, fodder) {
		t.Error("the Fodder was not sacrificed")
	}
	if !effects.OnBattlefield(g, hit) {
		t.Error("the first nonland permanent card was not put onto the battlefield")
	}
	if effects.OnBattlefield(g, sorcery) || effects.OnBattlefield(g, land) {
		t.Error("a land or an instant or sorcery card entered the battlefield")
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library has %d cards, want the land, the sorcery and the untouched bottom land", me.Library.Size())
	}
	ids := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards {
		ids[c.InstanceID] = true
	}
	if !ids[bottomLand] || !ids[sorcery] || !ids[land] {
		t.Error("the revealed misses and the unrevealed card should all be in the library")
	}
}

func TestGapB204DreamshaperShamanDoesNothingIfYouDeclineToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Dreamshaper Shaman", "Enchantment Creature — Minotaur Shaman", gapB204Dreamshaper, 5, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, fodder) {
		t.Error("a permanent was sacrificed though the mana was not paid")
	}
}

// --- Greater Gargadon --------------------------------------------------------

func TestGapB204GreaterGargadonSuspendsForTenAndCountsDownBySacrifice(t *testing.T) {
	spec, _ := effects.Lookup(gapB204Gargadon)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Counters != 10 || spec.SpecialActions[0].Cost != "{R}" {
		t.Fatalf("special actions %+v, want Suspend 10—{R}", spec.SpecialActions)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Greater Gargadon", TypeLine: "Creature — Beast", OracleID: gapB204Gargadon,
			ManaCost: "{9}{R}", Power: 9, Toughness: 7, Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterTime: 3}})
	})
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	land := effectstest.B12Permanent(g, me.ID, "A Forest", "Basic Land — Forest")
	rock := effectstest.B12Permanent(g, me.ID, "A Rock", "Artifact")
	enchant := effectstest.B12Permanent(g, me.ID, "An Enchantment", "Enchantment")

	// An enchantment is not a legal sacrifice.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{enchant}}); err == nil {
		t.Error("sacrificed an enchantment")
	}
	for i, sac := range []uuid.UUID{fodder, land} {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sac}}); err != nil {
			t.Fatalf("sacrifice %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := gapB204TimeCounters(g, id); got != 2-i {
			t.Errorf("after %d sacrifices the card has %d time counters, want %d", i+1, got, 2-i)
		}
	}
	// The last counter casts it for free (it has haste).
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("last sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCastPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, id) {
		t.Error("the Gargadon was not cast when its last time counter was removed")
	}
}

func gapB204TimeCounters(g *game.Game, id uuid.UUID) int {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c.Counters[game.CounterTime]
		}
	}
	return -1
}

func TestGapB204GreaterGargadonCantCountDownUnlessSuspended(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// Exiled with no time counters: not suspended.
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: id, Name: "Greater Gargadon", TypeLine: "Creature — Beast", OracleID: gapB204Gargadon,
			Power: 9, Toughness: 7, Owner: me.ID, Controller: me.ID})
	})
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("activated a Gargadon that is exiled but not suspended")
	}
	// In the hand it has no such ability either.
	hand := effectstest.HandCardFull(me, "Greater Gargadon", "Creature — Beast", "{9}{R}", gapB204Gargadon, nil)
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("activated a Gargadon in the hand")
	}
	if !effects.OnBattlefield(g, fodder) {
		t.Error("a refused activation still sacrificed the Fodder")
	}
}
