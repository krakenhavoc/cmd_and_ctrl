package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// b38Lands seeds n vanilla lands under owner.
func b38Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Permanent(g, owner, "Wastes", "Basic Land")
	}
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Turbulent Moor
// and Creosote Heath are rows in shared land cycles, so a transposed
// colour pair is invisible until somebody plays that exact land.
func TestBatch38CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B38TurbulentMoorOracle:       "Turbulent Moor",
		effectstest.B38CreosoteHeathOracle:       "Creosote Heath",
		effectstest.B38HedronCrawlerOracle:       "Hedron Crawler",
		effectstest.B38AlloyMyrOracle:            "Alloy Myr",
		effectstest.B38KingOfThePrideOracle:      "King of the Pride",
		effectstest.B38GaleriderSliverOracle:     "Galerider Sliver",
		effectstest.B38ArmageddonOracle:          "Armageddon",
		effectstest.B38InnocentBloodOracle:       "Innocent Blood",
		effectstest.B38AnnoyedAltisaurOracle:     "Annoyed Altisaur",
		effectstest.B38VirusBeetleOracle:         "Virus Beetle",
		effectstest.B38ObeliskOfUrdOracle:        "Obelisk of Urd",
		effectstest.B38FlameRiftOracle:           "Flame Rift",
		effectstest.B38PestilenceOracle:          "Pestilence",
		effectstest.B38FirespitterWhelpOracle:    "Firespitter Whelp",
		effectstest.B38SpiderManifestationOracle: "Spider Manifestation",
		effectstest.B38DesolationTwinOracle:      "Desolation Twin",
		effectstest.B38ViridianRevelOracle:       "Viridian Revel",
		effectstest.B38HydraOmnivoreOracle:       "Hydra Omnivore",
		effectstest.B38HareApparentOracle:        "Hare Apparent",
		effectstest.B38ColossificationOracle:     "Colossification",
		effectstest.B38HerosHeirloomOracle:       "Hero's Heirloom",
		effectstest.B38DoomedNecromancerOracle:   "Doomed Necromancer",
		effectstest.B38PhyrexianDelverOracle:     "Phyrexian Delver",
		effectstest.B38QuicksilverAmuletOracle:   "Quicksilver Amulet",
		effectstest.B38ArtfulDodgeOracle:         "Artful Dodge",
		effectstest.B38ExhumeOracle:              "Exhume",
		effectstest.B38SharedSummonsOracle:       "Shared Summons",
		effectstest.B38RelmsSketchingOracle:      "Relm's Sketching",
		effectstest.B38HeartwoodStorytellerOracl: "Heartwood Storyteller",
		effectstest.B38GrimServantOracle:         "Grim Servant",
		effectstest.B38MerrowReejereyOracle:      "Merrow Reejerey",
		effectstest.B38CanoptekScarabOracle:      "Canoptek Scarab Swarm",
		effectstest.B38TitaniasCommandOracle:     "Titania's Command",
		effectstest.B38SludgeMonsterOracle:       "Sludge Monster",
		effectstest.B38OpenTheVaultsOracle:       "Open the Vaults",
		effectstest.B38SpellStutterOracle:        "Spell Stutter",
		effectstest.B38WilsonOracle:              "Wilson, Refined Grizzly",
		effectstest.B38VinesOfVastwoodOracle:     "Vines of Vastwood",
		effectstest.B38DruidsDeliveranceOracle:   "Druid's Deliverance",
		effectstest.B38ManaSculptOracle:          "Mana Sculpt",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s: registered as %q", oracle, spec.Name)
		}
	}
}

// --- lands ---------------------------------------------------------

func TestB38TurbulentMoorEntersUntappedOnlyAgainstEightOpposingLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b38Lands(g, opp.ID, 4)
	b38Lands(g, other.ID, 3)
	// Your own lands never count toward the condition.
	b38Lands(g, me.ID, 6)
	tapped := effectstest.B12PlayFromHand(t, g, "Turbulent Moor", "Land — Plains Swamp", effectstest.B38TurbulentMoorOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands is not eight: it enters tapped")
	}
	b38Lands(g, other.ID, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	moor := effectstest.B12PlayFromHand(t, g, "Turbulent Moor", "Land — Plains Swamp", effectstest.B38TurbulentMoorOracle, game.CastSpellParams{})
	if effectstest.B16Tapped(t, g, moor) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	effectstest.B28TapForMana(t, g, me.ID, moor, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

func TestB38CreosoteHeathEntersTappedPingsAnOpponentAndTapsForGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := effectstest.B12PlayFromHand(t, g, "Creosote Heath", "Land — Desert", effectstest.B38CreosoteHeathOracle, game.CastSpellParams{})
	if !effectstest.B16Tapped(t, g, land) {
		t.Fatal("the Desert enters tapped")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || effectstest.HasID(p.PickTargetPlayers, me.ID) || !effectstest.HasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B28TapForMana(t, g, me.ID, land, "W")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

// --- mana creatures ------------------------------------------------

func TestB38HedronCrawlerAndAlloyMyrTapForTheirPrintedWidth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crawler := effectstest.PushCatalogPermanent(g, me.ID, "Hedron Crawler", "Artifact Creature — Construct", effectstest.B38HedronCrawlerOracle, false)
	effectstest.B28TapForMana(t, g, me.ID, crawler, "C")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("the Crawler taps for {C}: pool %v", got)
	}
	me.ManaPool = nil
	myr := effectstest.PushCatalogPermanent(g, me.ID, "Alloy Myr", "Artifact Creature — Myr", effectstest.B38AlloyMyrOracle, false)
	// Any of the five: the printed width, not the commander's identity.
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(myr) })
		me.ManaPool = nil
		effectstest.B28TapForMana(t, g, me.ID, myr, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("the Myr taps for {%s}: pool %v", colour, got)
		}
	}
}

// --- lords ---------------------------------------------------------

func TestB38KingOfThePridePumpsOtherCatsYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	king := effectstest.PushCatalogPermanent(g, me.ID, "King of the Pride", "Creature — Cat", effectstest.B38KingOfThePrideOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Leonin", "Creature — Cat", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	dog := effectstest.B12Creature(g, me.ID, "Dog", "Creature — Dog", 2, 2)
	if p, tough := effectstest.B38PT(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("your other Cat is 4/3: %d/%d", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Cat is untouched: %d/%d", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, dog); p != 2 || tough != 2 {
		t.Errorf("your Dog is untouched: %d/%d", p, tough)
	}
	// "Other" — the King is a Cat and does not pump itself.
	if p, tough := effectstest.B38PT(t, g, king); p != 1 || tough != 1 {
		t.Errorf("the King does not pump itself: %d/%d", p, tough)
	}
}

func TestB38GaleriderSliverGivesYourSliversFlyingAndNotTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galerider := effectstest.PushCatalogPermanent(g, me.ID, "Galerider Sliver", "Creature — Sliver", effectstest.B38GaleriderSliverOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Muscle Sliver", "Creature — Sliver", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "flying") {
		t.Error("your other Sliver flies")
	}
	// No "other": a lone Galerider Sliver flies.
	if !effectstest.HasEffectiveKeyword(t, g, galerider, "flying") {
		t.Error("the Galerider flies itself")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("modern Slivers say \"you control\" — an opponent's Sliver does not fly")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a Bear is not a Sliver")
	}
}

// --- sweepers ------------------------------------------------------

func TestB38ArmageddonDestroysEveryLandIncludingYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Armageddon", "Sorcery", effectstest.B38ArmageddonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your own land is destroyed too — there is no \"you control\" on the card")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's land is destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature is not a land")
	}
}

func TestB38InnocentBloodMakesEveryPlayerSacrificeIncludingYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Sol Ring", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Innocent Blood", "Sorcery", effectstest.B38InnocentBloodOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The caster chooses first, from their own creatures only.
	p := effectstest.SacrificeChoiceFor(g, me.ID)
	if p == nil {
		t.Fatalf("no sacrifice prompt for %s", me.Name)
	}
	opts := effectstest.SacrificeOptions(p)
	if effectstest.HasID(opts, theirs) {
		t.Error("you never pick an opponent's creature")
	}
	if effectstest.HasID(opts, rock) {
		t.Error("an artifact is not a creature")
	}
	if !effectstest.HasID(opts, mine) || !effectstest.HasID(opts, other) {
		t.Error("your own creatures are the choices you have")
	}
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the caster's own creature goes too, with the opponent's only one")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("only the chosen creature goes")
	}
}

func TestB38FlameRiftHitsEveryPlayerIncludingTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mineBefore := me.Life
	oppBefore := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Flame Rift", "Sorcery", effectstest.B38FlameRiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mineBefore-4 {
		t.Errorf("the caster takes 4 too: %d → %d", mineBefore, me.Life)
	}
	for i, before := range oppBefore {
		if got := g.Seats[i+1].Life; got != before-4 {
			t.Errorf("seat %d takes 4: %d → %d", i+1, before, got)
		}
	}
}

func TestB38PestilenceSweepsTheBoardAndTheTableThenEatsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pest := effectstest.PushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", effectstest.B38PestilenceOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My X/1", "Creature — Bear", 2, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their X/1", "Creature — Bear", 2, 1)
	myLife, theirLife := me.Life, opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, pest, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("one damage kills both X/1s — the caster's own included")
	}
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Errorf("every player takes one: %d→%d, %d→%d", myLife, me.Life, theirLife, opp.Life)
	}
	// With no creature left anywhere, the end step eats the enchantment.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pest) {
		t.Error("no creatures on the battlefield at the end step: Pestilence is sacrificed")
	}
}

func TestB38PestilenceSurvivesAnEndStepWithACreatureAnywhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pest := effectstest.PushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", effectstest.B38PestilenceOracle, false)
	// An OPPONENT's creature keeps it alive — the clause is the whole
	// battlefield, not "creatures you control".
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pest) {
		t.Error("an opponent's creature is still a creature on the battlefield")
	}
	_ = me
}

// --- cast triggers -------------------------------------------------

func TestB38AnnoyedAltisaurCascadesOnCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
	altisaur := effectstest.B38CastCreature(t, g, "Annoyed Altisaur", "Creature — Dinosaur", effectstest.B38AnnoyedAltisaurOracle, 6, 5)
	if !effectstest.HasEffectiveKeyword(t, g, altisaur, "reach") || !effectstest.HasEffectiveKeyword(t, g, altisaur, "trample") {
		t.Error("printed reach and trample reached the effective abilities")
	}
	// Cascade is a cast trigger: something was exiled off the top.
	if g.Exile.Size() == 0 && me.Library.Size() == 20 {
		t.Error("cascade exiled nothing off the top of the library")
	}
}

func TestB38DesolationTwinMakesItsTokenFromTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := len(g.Battlefield.Cards)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Desolation Twin", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.B38DesolationTwinOracle, Power: 10, Toughness: 10,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Desolation Twin: %v", err)
	}
	// The trigger is on the stack ABOVE the Twin: it resolves first,
	// so the token exists before the Twin does.
	if effectstest.TriggersOnStackFrom(g, id) != 1 {
		t.Fatal("the cast trigger did not fire from the stack (FromStack)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Eldrazi" && c.Controller == me.ID {
			tokens++
		}
	}
	if tokens != 1 {
		t.Errorf("one 10/10 Eldrazi token: got %d", tokens)
	}
	if len(g.Battlefield.Cards) != before+2 {
		t.Errorf("the Twin and its token: %d → %d", before, len(g.Battlefield.Cards))
	}
}

func TestB38FirespitterWhelpTriggersOnNoncreatureAndOnDragonsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Firespitter Whelp", "Creature — Dragon", effectstest.B38FirespitterWhelpOracle, false)

	before := effectstest.LifeOfOpponents(g)
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("a noncreature spell pings each opponent: seat %d %d → %d", i+1, b, got)
		}
	}

	before = effectstest.LifeOfOpponents(g)
	effectstest.B38CastCreature(t, g, "Shivan Dragon", "Creature — Dragon", "", 5, 5)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("a Dragon creature spell pings too: seat %d %d → %d", i+1, b, got)
		}
	}

	before = effectstest.LifeOfOpponents(g)
	effectstest.B38CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a non-Dragon creature spell does nothing: seat %d %d → %d", i+1, b, got)
		}
	}
}

func TestB38SpiderManifestationTapsForRedOrGreenAndUntapsOnABigSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spider := effectstest.PushCatalogPermanent(g, me.ID, "Spider Manifestation", "Creature — Spider Avatar", effectstest.B38SpiderManifestationOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, spider, "reach") {
		t.Error("printed reach reached the effective abilities")
	}
	effectstest.B28TapForMana(t, g, me.ID, spider, "G")
	if !effectstest.B16Tapped(t, g, spider) {
		t.Fatal("the mana ability taps it")
	}
	// A cheap spell does nothing.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, spider) {
		t.Error("a two-mana spell is not mana value 4 or greater")
	}
	// A four-drop untaps it.
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Big Thing", TypeLine: "Sorcery", ManaCost: "{2}{R}{G}",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Big Thing: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, spider) {
		t.Error("a mana value 4 spell untaps the Spider")
	}
}

func TestB38HeartwoodStorytellerAsksTheCastersOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Heartwood Storyteller", "Creature — Treefolk", effectstest.B38HeartwoodStorytellerOracl, false)
	// The Storyteller's OWN controller casts: they get nothing, the
	// rest of the table is asked.
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("the caster is not their own opponent — no prompt for them")
	}
	ask := effectstest.LatestConfirmFor(g, opp.ID)
	if ask == nil {
		t.Fatal("each of the caster's opponents is asked")
	}
	before := opp.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if opp.Hand.Size() != before+1 {
		t.Errorf("accepting draws a card: %d → %d", before, opp.Hand.Size())
	}
}

// --- enters triggers -----------------------------------------------

func TestB38VirusBeetleMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B38CastCreature(t, g, "Virus Beetle", "Artifact Creature — Insect", effectstest.B38VirusBeetleOracle, 1, 1)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if effectstest.LatestChooseCardsFor(g, p.ID) == nil {
			t.Errorf("no discard prompt for %s", p.Name)
		}
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("the Beetle's controller discards nothing")
	}
}

func TestB38HareApparentCountsOtherCopiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rabbits := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rabbit" && c.Controller == me.ID {
				n++
			}
		}
		return n
	}
	// The first copy sees no OTHER Hare Apparent and makes nothing.
	effectstest.B38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", effectstest.B38HareApparentOracle, 2, 2)
	if rabbits() != 0 {
		t.Errorf("the first copy makes no Rabbits: got %d", rabbits())
	}
	// A second sees one other and makes one.
	effectstest.B38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", effectstest.B38HareApparentOracle, 2, 2)
	if rabbits() != 1 {
		t.Errorf("the second copy makes one Rabbit: got %d", rabbits())
	}
	// A third sees two others and makes two more — the tokens
	// themselves are named Rabbit and never count.
	effectstest.B38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", effectstest.B38HareApparentOracle, 2, 2)
	if rabbits() != 3 {
		t.Errorf("the third copy makes two more Rabbits: got %d", rabbits())
	}
}

func TestB38PhyrexianDelverChargesTheCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B38Graveyard(g, me.ID, "Big Bear", "Creature — Bear", "{4}{G}", 5, 5)
	before := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Phyrexian Delver", TypeLine: "Creature — Phyrexian Zombie",
		OracleID: effectstest.B38PhyrexianDelverOracle, Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Phyrexian Delver: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card is reanimated")
	}
	if me.Life != before-5 {
		t.Errorf("you lose the CARD's mana value, 5: %d → %d", before, me.Life)
	}
}

func TestB38GrimServantTutorsWithinDevotionAndCostsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Devotion 3: two {B} pips already out, plus the Servant's own.
	effectstest.B27Push(g, me.ID, "Swamp Dweller", "Creature — Zombie", "", "{B}{B}", 2, 2, "B")
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{B}"},
		// A second legal pick, so the search is a real choice rather
		// than a forced one the engine settles without asking.
		game.Card{Name: "One Drop", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Five Drop", TypeLine: "Creature — Bear", ManaCost: "{4}{B}"},
	)
	before := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Grim Servant", TypeLine: "Creature — Zombie Warlock",
		OracleID: effectstest.B38GrimServantOracle, ManaCost: "{3}{B}", Colors: []string{"B"},
		Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grim Servant: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the tutor asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Three Drop") == uuid.Nil {
		t.Error("mana value 3 is within a devotion of 3")
	}
	if effectstest.SearchOptionNamed(g, c, "Five Drop") != uuid.Nil {
		t.Error("mana value 5 is beyond it")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Three Drop")
	if !effectstest.B02bHandHasNamed(me, "Three Drop") {
		t.Error("the pick goes to hand")
	}
	if me.Life != before-3 {
		t.Errorf("the 3 life is paid whatever the search found: %d → %d", before, me.Life)
	}
}

func TestB38CanoptekScarabSwarmPaysOnlyForArtifactsAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B38Graveyard(g, opp.ID, "Sol Ring", "Artifact", "{1}", 0, 0)
	effectstest.B38Graveyard(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.B38Graveyard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.B38Graveyard(g, opp.ID, "Shock", "Instant", "{R}", 0, 0)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Canoptek Scarab Swarm", TypeLine: "Artifact Creature — Insect",
		OracleID: effectstest.B38CanoptekScarabOracle, Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Canoptek Scarab Swarm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("the whole graveyard is exiled: %d left", opp.Graveyard.Size())
	}
	insects := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Insect" && c.Controller == me.ID {
			insects++
		}
	}
	if insects != 2 {
		t.Errorf("one Insect per artifact or land exiled — the Bear and the Shock pay nothing: got %d", insects)
	}
}

// --- attack and combat ---------------------------------------------

func TestB38HydraOmnivoreSpreadsWhatLandedToEachOtherOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, hit, other := g.Seats[0], g.Seats[1], g.Seats[2]
	hydra := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hydra Omnivore", TypeLine: "Creature — Hydra",
		OracleID: effectstest.B38HydraOmnivoreOracle, Power: 8, Toughness: 8,
		Owner: me.ID, Controller: me.ID,
	})
	hitBefore, otherBefore, mineBefore := hit.Life, other.Life, me.Life
	effectstest.DeclareAttack(t, g, hit.ID, hydra)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if hit.Life != hitBefore-8 {
		t.Errorf("the defending player takes the combat damage: %d → %d", hitBefore, hit.Life)
	}
	if other.Life != otherBefore-8 {
		t.Errorf("each OTHER opponent takes that much: %d → %d", otherBefore, other.Life)
	}
	if me.Life != mineBefore {
		t.Errorf("the Hydra's controller is not their own opponent: %d → %d", mineBefore, me.Life)
	}
}

func TestB38SludgeMonsterBlanksASlimedNonHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	horror := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other Horror", TypeLine: "Creature — Horror",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.B38CastCreature(t, g, "Sludge Monster", "Creature — Horror", effectstest.B38SludgeMonsterOracle, 5, 5)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, victim, "slime") != 1 {
		t.Fatalf("a slime counter landed: %d", effectstest.CounterOn(g, victim, "slime"))
	}
	if p, tough := effectstest.B38PT(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("a slimed non-Horror has base power and toughness 2/2: %d/%d", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("a slimed non-Horror loses all abilities")
	}
	// A Horror with a slime counter keeps everything.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(horror, "slime", 1) })
	if p, tough := effectstest.B38PT(t, g, horror); p != 4 || tough != 4 {
		t.Errorf("a Horror is exempt: %d/%d", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, horror, "flying") {
		t.Error("a Horror keeps its abilities")
	}
}

// --- attachments ---------------------------------------------------

func TestB38ColossificationTapsItsHostAndGrantsTwenty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Colossification", "Enchantment — Aura", effectstest.B38ColossificationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, bear); p != 22 || tough != 22 {
		t.Errorf("enchanted creature gets +20/+20: %d/%d", p, tough)
	}
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the Aura's enters trigger taps its own host")
	}
}

func TestB38HerosHeirloomGivesKeywordsOnlyToALegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	heirloom := effectstest.PushCatalogPermanent(g, me.ID, "Hero's Heirloom", "Artifact — Equipment", effectstest.B38HerosHeirloomOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	legend := effectstest.B12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.EquipTo(t, g, me.ID, heirloom, bear)
	if p, tough := effectstest.B38PT(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the +2/+1 is unconditional: %d/%d", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") || effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a nonlegendary host gets no keywords")
	}
	effectstest.EquipTo(t, g, me.ID, heirloom, legend)
	if !effectstest.HasEffectiveKeyword(t, g, legend, "trample") || !effectstest.HasEffectiveKeyword(t, g, legend, "haste") {
		t.Error("moving it to a legend switches both keywords on")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the old host keeps nothing")
	}
}

// --- activations ---------------------------------------------------

func TestB38DoomedNecromancerEatsItselfToReanimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	necro := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Necromancer", "Creature — Human Cleric Mercenary", effectstest.B38DoomedNecromancerOracle, false)
	mine := effectstest.B38Graveyard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("\"from YOUR graveyard\": an opponent's pile is not a legal target")
	}
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, necro, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)})
	if g.Battlefield.Contains(necro) {
		t.Error("sacrificing itself is part of the cost, paid on activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature card is reanimated")
	}
}

func TestB38QuicksilverAmuletPutsACreatureFromHandWithoutCastingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	amulet := effectstest.PushCatalogPermanent(g, me.ID, "Quicksilver Amulet", "Artifact", effectstest.B38QuicksilverAmuletOracle, false)
	fatty := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: fatty, Name: "Big Fatty", TypeLine: "Creature — Wurm", ManaCost: "{8}",
		Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID,
	})
	shock := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, amulet, 0, game.ActivateAbilityParams{})
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt opens")
	}
	if !effectstest.HasID(c.ChooseCards, fatty) {
		t.Error("the creature card is offered")
	}
	if effectstest.HasID(c.ChooseCards, shock) {
		t.Error("an instant cannot be put onto the battlefield")
	}
	if c.ChooseMin != 0 {
		t.Error("\"you MAY put\" — declining is a legal answer")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fatty}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(fatty) || me.Hand.Contains(fatty) {
		t.Error("the creature arrives on the battlefield")
	}
}

func TestB38MerrowReejereyPumpsOtherMerfolkAndTapsOrUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reejerey := effectstest.PushCatalogPermanent(g, me.ID, "Merrow Reejerey", "Creature — Merfolk Soldier", effectstest.B38MerrowReejereyOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Lord of Atlantis", "Creature — Merfolk", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	if p, tough := effectstest.B38PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your other Merfolk is 3/3: %d/%d", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Merfolk is untouched: %d/%d", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, reejerey); p != 1 || tough != 1 {
		t.Errorf("\"other\": the Reejerey does not pump itself: %d/%d", p, tough)
	}
	// Casting a Merfolk spell offers the tap-or-untap.
	victim := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	effectstest.B38CastCreature(t, g, "Merfolk Looter", "Creature — Merfolk Rogue", "", 1, 1)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Error("picking \"tap it\" taps the target permanent")
	}
}

// --- spells --------------------------------------------------------

func TestB38ObeliskOfUrdPumpsTheChosenTypeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Obelisk of Urd", "Artifact", effectstest.B38ObeliskOfUrdOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ReanimatedTypeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("\"as this enters, choose a creature type\" asks off the stack")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Goblin"); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice: %v", err)
	}
	if p, tough := effectstest.B38PT(t, g, goblin); p != 3 || tough != 3 {
		t.Errorf("a creature of the chosen type gets +2/+2: %d/%d", p, tough)
	}
	if p, tough := effectstest.B38PT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("anything else is untouched: %d/%d", p, tough)
	}
}

func TestB38ExhumeLetsEachPlayerReanimateTheirOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B38Graveyard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	// A noncreature in a graveyard is never offered.
	shock := effectstest.B38Graveyard(g, me.ID, "Shock", "Instant", "{R}", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Exhume", "Sorcery", effectstest.B38ExhumeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	mineAsk := effectstest.LatestChooseCardsFor(g, me.ID)
	if mineAsk == nil {
		t.Fatal("the caster is prompted for their own graveyard")
	}
	if effectstest.HasID(mineAsk.ChooseCards, theirs) {
		t.Error("nobody picks out of somebody else's pile")
	}
	if effectstest.HasID(mineAsk.ChooseCards, shock) {
		t.Error("an instant is not a creature card")
	}
	// CR 101.4: APNAP order, one player at a time.
	if effectstest.LatestChooseCardsFor(g, opp.ID) != nil {
		t.Fatal("the opponent is asked only after the active player has chosen")
	}
	if err := g.ResolveChooseCards(mineAsk.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("nothing enters until every player has chosen (#1867)")
	}
	oppAsk := effectstest.LatestChooseCardsFor(g, opp.ID)
	if oppAsk == nil {
		t.Fatal("each player is prompted, not just the caster")
	}
	if err := g.ResolveChooseCards(oppAsk.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards (opponent): %v", err)
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("both chosen creatures arrive, together")
	}
	var back game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs {
				back = c
			}
		}
	})
	if back.Controller != opp.ID {
		t.Errorf("each creature returns under its owner's control: %s", back.Controller)
	}
}

func TestB38SharedSummonsRefusesTwoOfTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", ManaCost: "{2}{G}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	effectstest.CastCatalogSpell(t, g, "Shared Summons", "Instant", effectstest.B38SharedSummonsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("an artifact is not a creature card")
	}
	// Two Bears share a name and are not a legal set.
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Error("\"with different names\" refuses two copies of the same name")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Bear", "Wolf")
	if !effectstest.B02bHandHasNamed(me, "Bear") || !effectstest.B02bHandHasNamed(me, "Wolf") {
		t.Error("two differently named creature cards reach hand")
	}
}

func TestB38RelmsSketchingCopiesALandNotJustACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	effectstest.CastCatalogSpell(t, g, "Relm's Sketching", "Sorcery", effectstest.B38RelmsSketchingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	effectstest.PassPriorityAroundTable(t, g)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Gaea's Cradle" && c.Controller == me.ID {
			copies++
		}
	}
	if copies != 1 {
		t.Errorf("a token copy of an opponent's land, under your control: got %d", copies)
	}
}

func TestB38ArtfulDodgeMakesACreatureUnblockableAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Artful Dodge", "Sorcery", effectstest.B38ArtfulDodgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if reason := effectstest.BlockRefusal(t, g, blocker, bear); reason == "" {
		t.Error("a creature that can't be blocked cannot be blocked by anything")
	}
	// The flashback copy is a second use, out of the graveyard.
	effectstest.AdvanceToMainOf(t, g, 0)
	id := effectstest.SeedGraveyardCard(t, g, "Artful Dodge", "Sorcery", effectstest.B38ArtfulDodgeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("a flashed-back sorcery is exiled rather than returning to the graveyard")
	}
}

func TestB38TitaniasCommandRunsTwoModesInPrintedOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B38Graveyard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.B38Graveyard(g, opp.ID, "Shock", "Instant", "{R}", 0, 0)
	before := me.Life
	// Bullets 0 and 2: exile a graveyard for life, and make two Bears.
	effectstest.CastModal(t, g, "Titania's Command", "Sorcery", effectstest.B38TitaniasCommandOracle, []int{0, 2},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("the target player's graveyard is exiled: %d left", opp.Graveyard.Size())
	}
	if me.Life != before+2 {
		t.Errorf("1 life per card exiled: %d → %d", before, me.Life)
	}
	bears := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == me.ID {
			bears++
		}
	}
	if bears != 2 {
		t.Errorf("two 2/2 green Bears: got %d", bears)
	}
}

// Its Auras choose their hosts as they enter (#3158); that half is
// aura_host_cards_test.go's.
func TestB38OpenTheVaultsRebuildsEveryonesArtifactsAndEnchantments(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B38Graveyard(g, me.ID, "Sol Ring", "Artifact", "{1}", 0, 0)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Signet", "Artifact", "{2}", 0, 0)
	ench := effectstest.B38Graveyard(g, me.ID, "Glorious Anthem", "Enchantment", "{1}{W}{W}", 0, 0)
	creature := effectstest.B38Graveyard(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Open the Vaults", "Sorcery", effectstest.B38OpenTheVaultsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(ench) {
		t.Error("your artifacts and enchantments come back")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("\"all graveyards\": an opponent's artifact comes back too")
	}
	var back game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == theirs {
				back = c
			}
		}
	})
	if back.Controller != opp.ID {
		t.Errorf("\"under their owners' control\": %s", back.Controller)
	}
	if g.Battlefield.Contains(creature) {
		t.Error("a creature card is neither an artifact nor an enchantment")
	}
}

func TestB38SpellStutterTaxesTwoPlusOnePerFaerie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Faerie A", "Creature — Faerie", 1, 1)
	effectstest.B12Creature(g, me.ID, "Faerie B", "Creature — Faerie", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Faerie", "Creature — Faerie", 1, 1)
	// The opponent casts something; the Stutter answers it.
	effectstest.AdvanceToMainOf(t, g, 1)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	stutter := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: stutter, Name: "Spell Stutter", TypeLine: "Instant",
		OracleID: effectstest.B38SpellStutterOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, stutter, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Spell Stutter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.AnswerPayUnless(t, g, opp.ID, false)
	// {2} plus {1} for each of the CASTER's two Faeries. The
	// opponent's own Faerie is not counted.
	if ask.PayCost != "{4}" {
		t.Errorf("the tax is {2} plus one per Faerie you control: %q", ask.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("declining the payment counters the spell")
	}
}

func TestB38ManaSculptCountersTheTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMainOf(t, g, 1)
	victim := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: victim, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	sculpt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: sculpt, Name: "Mana Sculpt", TypeLine: "Instant",
		OracleID: effectstest.B38ManaSculptOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, sculpt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Mana Sculpt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the target spell is countered")
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("a countered spell goes to its owner's graveyard")
	}
	// No Wizard, and a cast the engine did not charge: no refund is
	// scheduled. The refund itself is pinned in mana_spent_total_test.go
	// (#1735).
	if len(me.ManaPool) != 0 {
		t.Errorf("no refund without a Wizard: pool %v", effectstest.PoolColors(me))
	}
	if n := len(g.DelayedTriggers); n != 0 {
		t.Errorf("no Wizard, yet %d refund(s) scheduled", n)
	}
}

func TestB38DruidsDeliveranceSavesItsControllerAndNobodyElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	// The opponent is the attacker; we cast the Fog on our own turn's
	// combat is not possible, so give the turn to them.
	effectstest.AdvanceToMainOf(t, g, 1)
	attacker := effectstest.B12Creature(g, opp.ID, "Beater", "Creature — Ogre", 4, 4)
	second := effectstest.B12Creature(g, opp.ID, "Other Beater", "Creature — Ogre", 3, 3)
	deliverance := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: deliverance, Name: "Druid's Deliverance", TypeLine: "Instant",
		OracleID: effectstest.B38DruidsDeliveranceOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, deliverance, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Druid's Deliverance: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	mineBefore, otherBefore := me.Life, other.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatalf("DeclareAttacker at me: %v", err)
	}
	if err := g.DeclareAttacker(second, other.ID); err != nil {
		t.Fatalf("DeclareAttacker at the third seat: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mineBefore {
		t.Errorf("combat damage to the Deliverance's controller is prevented: %d → %d", mineBefore, me.Life)
	}
	if other.Life != otherBefore-3 {
		t.Errorf("a third player's combat damage is NOT prevented — this is not a Fog: %d → %d", otherBefore, other.Life)
	}
}

func TestB38ViridianRevelDrawsOnlyForAnOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Viridian Revel", "Enchantment", effectstest.B38ViridianRevelOracle, false)
	mine := effectstest.B12Permanent(g, me.ID, "My Signet", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Signet", "Artifact")

	// Your own artifact dying is somebody else's graveyard from the
	// Revel's point of view only if the OWNER is an opponent.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || effectstest.TriggerOnStack(g, mine) != nil {
		t.Error("an artifact that went to YOUR graveyard draws nothing")
	}

	before := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("an artifact in an opponent's graveyard draws: %d → %d", before, me.Hand.Size())
	}
}

func TestB38WilsonCannotBeCounteredAndWardsAgainstOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, ok := effects.Lookup(effectstest.B38WilsonOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("\"this spell can't be countered\" is the card's own rider on the Spec")
	}
	wilson := effectstest.PushCatalogPermanent(g, me.ID, "Wilson, Refined Grizzly", "Legendary Creature — Bear Warrior", effectstest.B38WilsonOracle, false)
	for _, kw := range []string{"reach", "vigilance", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, wilson, kw) {
			t.Errorf("printed %s reached the effective abilities", kw)
		}
	}
	// An opponent targeting him owes {2}.
	effectstest.AdvanceToMainOf(t, g, 1)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}",
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wilson}},
	}); err != nil {
		t.Fatalf("opponent targets Wilson: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("ward {2} asks the targeting opponent for the payment")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(bolt) {
		t.Error("declining the ward counters the spell")
	}
}
