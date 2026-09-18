package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch38_test.go — card-level coverage for the card-coverage roadmap's
// batch 38 (#401, `edhrec_rank` 3953–4052). One test per observable
// behaviour, driven through a real cast, activation or attack rather
// than by calling primitives directly.

const (
	b38TurbulentMoorOracle       = "2eb4da30-2600-4a7f-8e6c-6a090faa9a8d"
	b38CreosoteHeathOracle       = "c116b787-5f7e-47ef-a694-58709770dd32"
	b38HedronCrawlerOracle       = "0b9a4e06-b21d-4cbe-906f-9dbd08dbe5d3"
	b38AlloyMyrOracle            = "efb0394c-2a45-4dd8-bca3-08704056fa31"
	b38KingOfThePrideOracle      = "96d0e4dd-6cc2-4349-ac44-785b50f8dd90"
	b38GaleriderSliverOracle     = "1ad78038-44f9-4599-84db-fc1a87d9ed45"
	b38ArmageddonOracle          = "c9ed8b01-959a-47d6-891e-0abbdccf6e4f"
	b38InnocentBloodOracle       = "6791ec3c-c397-4087-8c8c-84d3797df415"
	b38AnnoyedAltisaurOracle     = "a8135179-51ef-454e-98ff-69137440339f"
	b38VirusBeetleOracle         = "e1e489f6-37e5-469e-bca3-0e9d3fc3fa95"
	b38ObeliskOfUrdOracle        = "e2ef3a24-9e78-47fc-9192-049aa0ddb7a0"
	b38FlameRiftOracle           = "28479aa8-d1cd-421a-8bbb-0594bb4dd410"
	b38PestilenceOracle          = "dafe63ef-f3d6-45e7-877a-573da92ba85e"
	b38FirespitterWhelpOracle    = "1cad92ce-55c8-4b78-8d8c-56645ef8e6ee"
	b38SpiderManifestationOracle = "8281f2b9-e81b-48da-812a-8713b2adf8ab"
	b38DesolationTwinOracle      = "c0cb1f37-1679-42c7-a794-81e088157eeb"
	b38ViridianRevelOracle       = "c24b0c7d-1a4b-4eca-8d83-9e81ed3b3b27"
	b38HydraOmnivoreOracle       = "8f504855-f3df-4284-a189-e799bcddf620"
	b38HareApparentOracle        = "3c1619bd-db5e-4df6-a196-0a9d62374f6d"
	b38ColossificationOracle     = "05c5ee65-651b-4bfb-b57f-305026507135"
	b38HerosHeirloomOracle       = "f707dfb5-dfa9-471b-ac29-5435c30069cc"
	b38DoomedNecromancerOracle   = "155422a0-a0cd-4399-8ed9-fa68ac2c80a6"
	b38PhyrexianDelverOracle     = "a13cbac0-4c76-4970-b61e-5f4e020ee95c"
	b38QuicksilverAmuletOracle   = "e1096a98-a631-43f5-97d8-325001d608b3"
	b38ArtfulDodgeOracle         = "c174dcbb-03a0-439c-b3d8-ed61bd46dc67"
	b38ExhumeOracle              = "fbe61f74-1b3c-4e12-8758-7029872c9ff1"
	b38SharedSummonsOracle       = "c2c9dbd0-2062-4ee1-b32e-8eeacd95589c"
	b38RelmsSketchingOracle      = "dd6601c6-810d-4883-89d2-5d31419fb1cc"
	b38HeartwoodStorytellerOracl = "82e6da87-f8a4-4897-bf0f-c0f2cd06b8b1"
	b38GrimServantOracle         = "1caa7fa2-a881-4a65-a87e-12a974520c83"
	b38MerrowReejereyOracle      = "6b7e6ae4-2ee1-44bf-ac93-79fc87494515"
	b38CanoptekScarabOracle      = "5b7f6cc3-8d4e-42e5-a908-7ebe3bccaf1e"
	b38TitaniasCommandOracle     = "7aae0a3d-8882-4485-a126-f06ea6593dcb"
	b38SludgeMonsterOracle       = "2802fc92-49a0-43c8-bc10-24882eeee3f1"
	b38OpenTheVaultsOracle       = "1e9c473e-bd65-4e1f-b2ab-cac58dc581c9"
	b38SpellStutterOracle        = "f9bf996e-ae14-40a3-a8d8-f5725ddaa270"
	b38WilsonOracle              = "d2766fd7-5cf9-4037-9f34-9ae3982c613a"
	b38VinesOfVastwoodOracle     = "8998d211-5ca6-41d2-bc64-f51f70bd41e5"
	b38DruidsDeliveranceOracle   = "fde7645a-5f02-4d5f-b38c-8390f325899e"
	b38ManaSculptOracle          = "35e2f82e-7ca3-4a92-9134-b7999eef5337"
)

// --- shared fixtures -----------------------------------------------

// b38PT reads a battlefield card's post-layer power and toughness.
func b38PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	p, tough, found := 0, 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				eff := c.Effective()
				p, tough, found = eff.Power, eff.Toughness, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return p, tough
}

// b38Lands seeds n vanilla lands under owner.
func b38Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, owner, "Wastes", "Basic Land")
	}
}

// b38CastCreature casts a creature with a real body from the active
// seat's hand and settles the spell and whatever it triggered.
func b38CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// b38Graveyard seeds a card into a player's graveyard with a real
// mana cost, so a mana-value rider has something to read.
func b38Graveyard(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string, power, toughness int) uuid.UUID {
	p := g.PlayerByID(owner)
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. Turbulent Moor
// and Creosote Heath are rows in shared land cycles, so a transposed
// colour pair is invisible until somebody plays that exact land.
func TestBatch38CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b38TurbulentMoorOracle:       "Turbulent Moor",
		b38CreosoteHeathOracle:       "Creosote Heath",
		b38HedronCrawlerOracle:       "Hedron Crawler",
		b38AlloyMyrOracle:            "Alloy Myr",
		b38KingOfThePrideOracle:      "King of the Pride",
		b38GaleriderSliverOracle:     "Galerider Sliver",
		b38ArmageddonOracle:          "Armageddon",
		b38InnocentBloodOracle:       "Innocent Blood",
		b38AnnoyedAltisaurOracle:     "Annoyed Altisaur",
		b38VirusBeetleOracle:         "Virus Beetle",
		b38ObeliskOfUrdOracle:        "Obelisk of Urd",
		b38FlameRiftOracle:           "Flame Rift",
		b38PestilenceOracle:          "Pestilence",
		b38FirespitterWhelpOracle:    "Firespitter Whelp",
		b38SpiderManifestationOracle: "Spider Manifestation",
		b38DesolationTwinOracle:      "Desolation Twin",
		b38ViridianRevelOracle:       "Viridian Revel",
		b38HydraOmnivoreOracle:       "Hydra Omnivore",
		b38HareApparentOracle:        "Hare Apparent",
		b38ColossificationOracle:     "Colossification",
		b38HerosHeirloomOracle:       "Hero's Heirloom",
		b38DoomedNecromancerOracle:   "Doomed Necromancer",
		b38PhyrexianDelverOracle:     "Phyrexian Delver",
		b38QuicksilverAmuletOracle:   "Quicksilver Amulet",
		b38ArtfulDodgeOracle:         "Artful Dodge",
		b38ExhumeOracle:              "Exhume",
		b38SharedSummonsOracle:       "Shared Summons",
		b38RelmsSketchingOracle:      "Relm's Sketching",
		b38HeartwoodStorytellerOracl: "Heartwood Storyteller",
		b38GrimServantOracle:         "Grim Servant",
		b38MerrowReejereyOracle:      "Merrow Reejerey",
		b38CanoptekScarabOracle:      "Canoptek Scarab Swarm",
		b38TitaniasCommandOracle:     "Titania's Command",
		b38SludgeMonsterOracle:       "Sludge Monster",
		b38OpenTheVaultsOracle:       "Open the Vaults",
		b38SpellStutterOracle:        "Spell Stutter",
		b38WilsonOracle:              "Wilson, Refined Grizzly",
		b38VinesOfVastwoodOracle:     "Vines of Vastwood",
		b38DruidsDeliveranceOracle:   "Druid's Deliverance",
		b38ManaSculptOracle:          "Mana Sculpt",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
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
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b38Lands(g, opp.ID, 4)
	b38Lands(g, other.ID, 3)
	// Your own lands never count toward the condition.
	b38Lands(g, me.ID, 6)
	tapped := b12PlayFromHand(t, g, "Turbulent Moor", "Land — Plains Swamp", b38TurbulentMoorOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, tapped) {
		t.Fatal("seven opposing lands is not eight: it enters tapped")
	}
	b38Lands(g, other.ID, 1)
	advanceToMainOf(t, g, 0)
	moor := b12PlayFromHand(t, g, "Turbulent Moor", "Land — Plains Swamp", b38TurbulentMoorOracle, game.CastSpellParams{})
	if b16Tapped(t, g, moor) {
		t.Fatal("eight opposing lands between two opponents: it enters untapped")
	}
	b28TapForMana(t, g, me.ID, moor, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B}: pool %v", got)
	}
}

func TestB38CreosoteHeathEntersTappedPingsAnOpponentAndTapsForGreenOrWhite(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[2]
	before := opp.Life
	land := b12PlayFromHand(t, g, "Creosote Heath", "Land — Desert", b38CreosoteHeathOracle, game.CastSpellParams{})
	if !b16Tapped(t, g, land) {
		t.Fatal("the Desert enters tapped")
	}
	p := latestPickTarget(g, me.ID)
	if p == nil || hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Fatalf("the trigger asks for an opponent, never you: %+v", p)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("target opponent takes 1: %d → %d", before, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "W")
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Errorf("tapped for {W}: pool %v", got)
	}
}

// --- mana creatures ------------------------------------------------

func TestB38HedronCrawlerAndAlloyMyrTapForTheirPrintedWidth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crawler := pushCatalogPermanent(g, me.ID, "Hedron Crawler", "Artifact Creature — Construct", b38HedronCrawlerOracle, false)
	b28TapForMana(t, g, me.ID, crawler, "C")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("the Crawler taps for {C}: pool %v", got)
	}
	me.ManaPool = nil
	myr := pushCatalogPermanent(g, me.ID, "Alloy Myr", "Artifact Creature — Myr", b38AlloyMyrOracle, false)
	// Any of the five: the printed width, not the commander's identity.
	for _, colour := range []string{"W", "U", "B", "R", "G"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(myr) })
		me.ManaPool = nil
		b28TapForMana(t, g, me.ID, myr, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("the Myr taps for {%s}: pool %v", colour, got)
		}
	}
}

// --- lords ---------------------------------------------------------

func TestB38KingOfThePridePumpsOtherCatsYouControlOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	king := pushCatalogPermanent(g, me.ID, "King of the Pride", "Creature — Cat", b38KingOfThePrideOracle, false)
	mine := b12Creature(g, me.ID, "Leonin", "Creature — Cat", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	dog := b12Creature(g, me.ID, "Dog", "Creature — Dog", 2, 2)
	if p, tough := b38PT(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("your other Cat is 4/3: %d/%d", p, tough)
	}
	if p, tough := b38PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Cat is untouched: %d/%d", p, tough)
	}
	if p, tough := b38PT(t, g, dog); p != 2 || tough != 2 {
		t.Errorf("your Dog is untouched: %d/%d", p, tough)
	}
	// "Other" — the King is a Cat and does not pump itself.
	if p, tough := b38PT(t, g, king); p != 1 || tough != 1 {
		t.Errorf("the King does not pump itself: %d/%d", p, tough)
	}
}

func TestB38GaleriderSliverGivesYourSliversFlyingAndNotTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	galerider := pushCatalogPermanent(g, me.ID, "Galerider Sliver", "Creature — Sliver", b38GaleriderSliverOracle, false)
	mine := b12Creature(g, me.ID, "Muscle Sliver", "Creature — Sliver", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "flying") {
		t.Error("your other Sliver flies")
	}
	// No "other": a lone Galerider Sliver flies.
	if !hasEffectiveKeyword(t, g, galerider, "flying") {
		t.Error("the Galerider flies itself")
	}
	if hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("modern Slivers say \"you control\" — an opponent's Sliver does not fly")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("a Bear is not a Sliver")
	}
}

// --- sweepers ------------------------------------------------------

func TestB38ArmageddonDestroysEveryLandIncludingYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Armageddon", "Sorcery", b38ArmageddonOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Sol Ring", "Artifact")
	castCatalogSpell(t, g, "Innocent Blood", "Sorcery", b38InnocentBloodOracle, nil)
	passPriorityAroundTable(t, g)
	// One prompt per player, each offering only their own creatures.
	for _, seat := range []*game.Player{me, opp} {
		if sacrificeChoiceFor(g, seat.ID) == nil {
			t.Fatalf("no sacrifice prompt for %s", seat.Name)
		}
	}
	p := sacrificeChoiceFor(g, me.ID)
	if hasID(p.SacrificeOptions, theirs) {
		t.Error("you never pick an opponent's creature")
	}
	if hasID(p.SacrificeOptions, rock) {
		t.Error("an artifact is not a creature")
	}
	if !hasID(p.SacrificeOptions, mine) {
		t.Error("your own creature is the only choice you have")
	}
	answerSacrifice(t, g, me.ID, mine)
	if g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature goes too")
	}
}

func TestB38FlameRiftHitsEveryPlayerIncludingTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mineBefore := me.Life
	oppBefore := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Flame Rift", "Sorcery", b38FlameRiftOracle, nil)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pest := pushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", b38PestilenceOracle, false)
	mine := b12Creature(g, me.ID, "My X/1", "Creature — Bear", 2, 1)
	theirs := b12Creature(g, opp.ID, "Their X/1", "Creature — Bear", 2, 1)
	myLife, theirLife := me.Life, opp.Life
	advanceToMain(t, g)
	b06AddMana(me, "B")
	b16Activate(t, g, me.ID, pest, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("one damage kills both X/1s — the caster's own included")
	}
	if me.Life != myLife-1 || opp.Life != theirLife-1 {
		t.Errorf("every player takes one: %d→%d, %d→%d", myLife, me.Life, theirLife, opp.Life)
	}
	// With no creature left anywhere, the end step eats the enchantment.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pest) {
		t.Error("no creatures on the battlefield at the end step: Pestilence is sacrificed")
	}
}

func TestB38PestilenceSurvivesAnEndStepWithACreatureAnywhere(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pest := pushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", b38PestilenceOracle, false)
	// An OPPONENT's creature keeps it alive — the clause is the whole
	// battlefield, not "creatures you control".
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pest) {
		t.Error("an opponent's creature is still a creature on the battlefield")
	}
	_ = me
}

// --- cast triggers -------------------------------------------------

func TestB38AnnoyedAltisaurCascadesOnCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
	altisaur := b38CastCreature(t, g, "Annoyed Altisaur", "Creature — Dinosaur", b38AnnoyedAltisaurOracle, 6, 5)
	if !hasEffectiveKeyword(t, g, altisaur, "reach") || !hasEffectiveKeyword(t, g, altisaur, "trample") {
		t.Error("printed reach and trample reached the effective abilities")
	}
	// Cascade is a cast trigger: something was exiled off the top.
	if g.Exile.Size() == 0 && me.Library.Size() == 20 {
		t.Error("cascade exiled nothing off the top of the library")
	}
}

func TestB38DesolationTwinMakesItsTokenFromTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := len(g.Battlefield.Cards)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Desolation Twin", TypeLine: "Creature — Eldrazi",
		OracleID: b38DesolationTwinOracle, Power: 10, Toughness: 10,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Desolation Twin: %v", err)
	}
	// The trigger is on the stack ABOVE the Twin: it resolves first,
	// so the token exists before the Twin does.
	if triggersOnStackFrom(g, id) != 1 {
		t.Fatal("the cast trigger did not fire from the stack (FromStack)")
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Firespitter Whelp", "Creature — Dragon", b38FirespitterWhelpOracle, false)

	before := lifeOfOpponents(g)
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("a noncreature spell pings each opponent: seat %d %d → %d", i+1, b, got)
		}
	}

	before = lifeOfOpponents(g)
	b38CastCreature(t, g, "Shivan Dragon", "Creature — Dragon", "", 5, 5)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("a Dragon creature spell pings too: seat %d %d → %d", i+1, b, got)
		}
	}

	before = lifeOfOpponents(g)
	b38CastCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("a non-Dragon creature spell does nothing: seat %d %d → %d", i+1, b, got)
		}
	}
}

func TestB38SpiderManifestationTapsForRedOrGreenAndUntapsOnABigSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spider := pushCatalogPermanent(g, me.ID, "Spider Manifestation", "Creature — Spider Avatar", b38SpiderManifestationOracle, false)
	if !hasEffectiveKeyword(t, g, spider, "reach") {
		t.Error("printed reach reached the effective abilities")
	}
	b28TapForMana(t, g, me.ID, spider, "G")
	if !b16Tapped(t, g, spider) {
		t.Fatal("the mana ability taps it")
	}
	// A cheap spell does nothing.
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, spider) {
		t.Error("a two-mana spell is not mana value 4 or greater")
	}
	// A four-drop untaps it.
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Big Thing", TypeLine: "Sorcery", ManaCost: "{2}{R}{G}",
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Big Thing: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, spider) {
		t.Error("a mana value 4 spell untaps the Spider")
	}
}

func TestB38HeartwoodStorytellerAsksTheCastersOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Heartwood Storyteller", "Creature — Treefolk", b38HeartwoodStorytellerOracl, false)
	// The Storyteller's OWN controller casts: they get nothing, the
	// rest of the table is asked.
	castCatalogSpell(t, g, "Shock", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("the caster is not their own opponent — no prompt for them")
	}
	ask := latestConfirmFor(g, opp.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b38CastCreature(t, g, "Virus Beetle", "Artifact Creature — Insect", b38VirusBeetleOracle, 1, 1)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if latestChooseCardsFor(g, p.ID) == nil {
			t.Errorf("no discard prompt for %s", p.Name)
		}
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the Beetle's controller discards nothing")
	}
}

func TestB38HareApparentCountsOtherCopiesOnly(t *testing.T) {
	g := newCatalogGame(t)
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
	b38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", b38HareApparentOracle, 2, 2)
	if rabbits() != 0 {
		t.Errorf("the first copy makes no Rabbits: got %d", rabbits())
	}
	// A second sees one other and makes one.
	b38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", b38HareApparentOracle, 2, 2)
	if rabbits() != 1 {
		t.Errorf("the second copy makes one Rabbit: got %d", rabbits())
	}
	// A third sees two others and makes two more — the tokens
	// themselves are named Rabbit and never count.
	b38CastCreature(t, g, "Hare Apparent", "Creature — Rabbit Noble", b38HareApparentOracle, 2, 2)
	if rabbits() != 3 {
		t.Errorf("the third copy makes two more Rabbits: got %d", rabbits())
	}
}

func TestB38PhyrexianDelverChargesTheCardsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b38Graveyard(g, me.ID, "Big Bear", "Creature — Bear", "{4}{G}", 5, 5)
	before := me.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Phyrexian Delver", TypeLine: "Creature — Phyrexian Zombie",
		OracleID: b38PhyrexianDelverOracle, Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Phyrexian Delver: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card is reanimated")
	}
	if me.Life != before-5 {
		t.Errorf("you lose the CARD's mana value, 5: %d → %d", before, me.Life)
	}
}

func TestB38GrimServantTutorsWithinDevotionAndCostsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Devotion 3: two {B} pips already out, plus the Servant's own.
	b27Push(g, me.ID, "Swamp Dweller", "Creature — Zombie", "", "{B}{B}", 2, 2, "B")
	seedSearchLibrary(me,
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
		OracleID: b38GrimServantOracle, ManaCost: "{3}{B}", Colors: []string{"B"},
		Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grim Servant: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the tutor asks")
	}
	if searchOptionNamed(g, c, "Three Drop") == uuid.Nil {
		t.Error("mana value 3 is within a devotion of 3")
	}
	if searchOptionNamed(g, c, "Five Drop") != uuid.Nil {
		t.Error("mana value 5 is beyond it")
	}
	answerSearchNamed(t, g, me.ID, "Three Drop")
	if !b02bHandHasNamed(me, "Three Drop") {
		t.Error("the pick goes to hand")
	}
	if me.Life != before-3 {
		t.Errorf("the 3 life is paid whatever the search found: %d → %d", before, me.Life)
	}
}

func TestB38CanoptekScarabSwarmPaysOnlyForArtifactsAndLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b38Graveyard(g, opp.ID, "Sol Ring", "Artifact", "{1}", 0, 0)
	b38Graveyard(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	b38Graveyard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	b38Graveyard(g, opp.ID, "Shock", "Instant", "{R}", 0, 0)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Canoptek Scarab Swarm", TypeLine: "Artifact Creature — Insect",
		OracleID: b38CanoptekScarabOracle, Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Canoptek Scarab Swarm: %v", err)
	}
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, hit, other := g.Seats[0], g.Seats[1], g.Seats[2]
	hydra := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hydra Omnivore", TypeLine: "Creature — Hydra",
		OracleID: b38HydraOmnivoreOracle, Power: 8, Toughness: 8,
		Owner: me.ID, Controller: me.ID,
	})
	hitBefore, otherBefore, mineBefore := hit.Life, other.Life, me.Life
	declareAttack(t, g, hit.ID, hydra)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	horror := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other Horror", TypeLine: "Creature — Horror",
		Power: 4, Toughness: 4, Keywords: []string{"flying"},
		Owner: opp.ID, Controller: opp.ID,
	})
	b38CastCreature(t, g, "Sludge Monster", "Creature — Horror", b38SludgeMonsterOracle, 5, 5)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if counterOn(g, victim, "slime") != 1 {
		t.Fatalf("a slime counter landed: %d", counterOn(g, victim, "slime"))
	}
	if p, tough := b38PT(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("a slimed non-Horror has base power and toughness 2/2: %d/%d", p, tough)
	}
	if hasEffectiveKeyword(t, g, victim, "flying") {
		t.Error("a slimed non-Horror loses all abilities")
	}
	// A Horror with a slime counter keeps everything.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(horror, "slime", 1) })
	if p, tough := b38PT(t, g, horror); p != 4 || tough != 4 {
		t.Errorf("a Horror is exempt: %d/%d", p, tough)
	}
	if !hasEffectiveKeyword(t, g, horror, "flying") {
		t.Error("a Horror keeps its abilities")
	}
}

// --- attachments ---------------------------------------------------

func TestB38ColossificationTapsItsHostAndGrantsTwenty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Colossification", "Enchantment — Aura", b38ColossificationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, tough := b38PT(t, g, bear); p != 22 || tough != 22 {
		t.Errorf("enchanted creature gets +20/+20: %d/%d", p, tough)
	}
	if !b16Tapped(t, g, bear) {
		t.Error("the Aura's enters trigger taps its own host")
	}
}

func TestB38HerosHeirloomGivesKeywordsOnlyToALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	heirloom := pushCatalogPermanent(g, me.ID, "Hero's Heirloom", "Artifact — Equipment", b38HerosHeirloomOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	legend := b12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	advanceToMain(t, g)
	equipTo(t, g, me.ID, heirloom, bear)
	if p, tough := b38PT(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("the +2/+1 is unconditional: %d/%d", p, tough)
	}
	if hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a nonlegendary host gets no keywords")
	}
	equipTo(t, g, me.ID, heirloom, legend)
	if !hasEffectiveKeyword(t, g, legend, "trample") || !hasEffectiveKeyword(t, g, legend, "haste") {
		t.Error("moving it to a legend switches both keywords on")
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the old host keeps nothing")
	}
}

// --- activations ---------------------------------------------------

func TestB38DoomedNecromancerEatsItselfToReanimate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	necro := pushCatalogPermanent(g, me.ID, "Doomed Necromancer", "Creature — Human Cleric Mercenary", b38DoomedNecromancerOracle, false)
	mine := b38Graveyard(g, me.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := b38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, necro, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err == nil {
		t.Fatal("\"from YOUR graveyard\": an opponent's pile is not a legal target")
	}
	b06AddMana(me, "B")
	b16Activate(t, g, me.ID, necro, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)})
	if g.Battlefield.Contains(necro) {
		t.Error("sacrificing itself is part of the cost, paid on activation")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature card is reanimated")
	}
}

func TestB38QuicksilverAmuletPutsACreatureFromHandWithoutCastingIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	amulet := pushCatalogPermanent(g, me.ID, "Quicksilver Amulet", "Artifact", b38QuicksilverAmuletOracle, false)
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
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, amulet, 0, game.ActivateAbilityParams{})
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt opens")
	}
	if !hasID(c.ChooseCards, fatty) {
		t.Error("the creature card is offered")
	}
	if hasID(c.ChooseCards, shock) {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reejerey := pushCatalogPermanent(g, me.ID, "Merrow Reejerey", "Creature — Merfolk Soldier", b38MerrowReejereyOracle, false)
	mine := b12Creature(g, me.ID, "Lord of Atlantis", "Creature — Merfolk", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	if p, tough := b38PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your other Merfolk is 3/3: %d/%d", p, tough)
	}
	if p, tough := b38PT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's Merfolk is untouched: %d/%d", p, tough)
	}
	if p, tough := b38PT(t, g, reejerey); p != 1 || tough != 1 {
		t.Errorf("\"other\": the Reejerey does not pump itself: %d/%d", p, tough)
	}
	// Casting a Merfolk spell offers the tap-or-untap.
	victim := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	b38CastCreature(t, g, "Merfolk Looter", "Creature — Merfolk Rogue", "", 1, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, victim) {
		t.Error("picking \"tap it\" taps the target permanent")
	}
}

// --- spells --------------------------------------------------------

func TestB38ObeliskOfUrdPumpsTheChosenTypeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	goblin := b12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Obelisk of Urd", "Artifact", b38ObeliskOfUrdOracle, nil)
	passPriorityAroundTable(t, g)
	c := reanimatedTypeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("\"as this enters, choose a creature type\" asks off the stack")
	}
	if err := g.ResolveCreatureTypeChoice(c.ID, me.ID, "Goblin"); err != nil {
		t.Fatalf("ResolveCreatureTypeChoice: %v", err)
	}
	if p, tough := b38PT(t, g, goblin); p != 3 || tough != 3 {
		t.Errorf("a creature of the chosen type gets +2/+2: %d/%d", p, tough)
	}
	if p, tough := b38PT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("anything else is untouched: %d/%d", p, tough)
	}
}

func TestB38ExhumeLetsEachPlayerReanimateTheirOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b38Graveyard(g, me.ID, "My Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirs := b38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	// A noncreature in a graveyard is never offered.
	shock := b38Graveyard(g, me.ID, "Shock", "Instant", "{R}", 0, 0)
	castCatalogSpell(t, g, "Exhume", "Sorcery", b38ExhumeOracle, nil)
	passPriorityAroundTable(t, g)
	mineAsk := latestChooseCardsFor(g, me.ID)
	if mineAsk == nil {
		t.Fatal("the caster is prompted for their own graveyard")
	}
	if hasID(mineAsk.ChooseCards, theirs) {
		t.Error("nobody picks out of somebody else's pile")
	}
	if hasID(mineAsk.ChooseCards, shock) {
		t.Error("an instant is not a creature card")
	}
	if latestChooseCardsFor(g, opp.ID) == nil {
		t.Fatal("each player is prompted, not just the caster")
	}
	if err := g.ResolveChooseCards(mineAsk.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the chosen creature arrives")
	}
	oppAsk := latestChooseCardsFor(g, opp.ID)
	if err := g.ResolveChooseCards(oppAsk.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards (opponent): %v", err)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf", ManaCost: "{2}{G}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
	)
	castCatalogSpell(t, g, "Shared Summons", "Instant", b38SharedSummonsOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the search asks")
	}
	if searchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("an artifact is not a creature card")
	}
	// Two Bears share a name and are not a legal set.
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Error("\"with different names\" refuses two copies of the same name")
	}
	answerSearchNamed(t, g, me.ID, "Bear", "Wolf")
	if !b02bHandHasNamed(me, "Bear") || !b02bHandHasNamed(me, "Wolf") {
		t.Error("two differently named creature cards reach hand")
	}
}

func TestB38RelmsSketchingCopiesALandNotJustACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, opp.ID, "Gaea's Cradle", "Legendary Land")
	castCatalogSpell(t, g, "Relm's Sketching", "Sorcery", b38RelmsSketchingOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	castCatalogSpell(t, g, "Artful Dodge", "Sorcery", b38ArtfulDodgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, bear)
	if reason := blockRefusal(t, g, blocker, bear); reason == "" {
		t.Error("a creature that can't be blocked cannot be blocked by anything")
	}
	// The flashback copy is a second use, out of the graveyard.
	advanceToMainOf(t, g, 0)
	id := seedGraveyardCard(t, g, "Artful Dodge", "Sorcery", b38ArtfulDodgeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("a flashed-back sorcery is exiled rather than returning to the graveyard")
	}
}

func TestB38TitaniasCommandRunsTwoModesInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b38Graveyard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	b38Graveyard(g, opp.ID, "Shock", "Instant", "{R}", 0, 0)
	before := me.Life
	// Bullets 0 and 2: exile a graveyard for life, and make two Bears.
	castModal(t, g, "Titania's Command", "Sorcery", b38TitaniasCommandOracle, []int{0, 2},
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
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

func TestB38OpenTheVaultsRebuildsEveryonesArtifactsButLeavesAuras(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b38Graveyard(g, me.ID, "Sol Ring", "Artifact", "{1}", 0, 0)
	theirs := b38Graveyard(g, opp.ID, "Their Signet", "Artifact", "{2}", 0, 0)
	ench := b38Graveyard(g, me.ID, "Glorious Anthem", "Enchantment", "{1}{W}{W}", 0, 0)
	aura := b38Graveyard(g, me.ID, "Rancor", "Enchantment — Aura", "{G}", 0, 0)
	creature := b38Graveyard(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	castCatalogSpell(t, g, "Open the Vaults", "Sorcery", b38OpenTheVaultsOracle, nil)
	passPriorityAroundTable(t, g)
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
	// Declared caveat: Auras are left where they are.
	if g.Battlefield.Contains(aura) {
		t.Error("the declared simplification leaves Aura cards in their graveyards")
	}
}

func TestB38SpellStutterTaxesTwoPlusOnePerFaerie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Faerie A", "Creature — Faerie", 1, 1)
	b12Creature(g, me.ID, "Faerie B", "Creature — Faerie", 1, 1)
	b12Creature(g, opp.ID, "Their Faerie", "Creature — Faerie", 1, 1)
	// The opponent casts something; the Stutter answers it.
	advanceToMainOf(t, g, 1)
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
		OracleID: b38SpellStutterOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, stutter, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Spell Stutter: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := answerPayUnless(t, g, opp.ID, false)
	// {2} plus {1} for each of the CASTER's two Faeries. The
	// opponent's own Faerie is not counted.
	if ask.PayCost != "{4}" {
		t.Errorf("the tax is {2} plus one per Faerie you control: %q", ask.PayCost)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("declining the payment counters the spell")
	}
}

func TestB38ManaSculptCountersTheTargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMainOf(t, g, 1)
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
		OracleID: b38ManaSculptOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, sculpt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("cast Mana Sculpt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("the target spell is countered")
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("a countered spell goes to its owner's graveyard")
	}
	// Declared caveat: no mana refund on the next main phase.
	if len(me.ManaPool) != 0 {
		t.Errorf("the refund is not implemented: pool %v", poolColors(me))
	}
}

func TestB38VinesOfVastwoodGrantsHexproofToYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Declared caveat: only a creature you control is a legal target.
	vines := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: vines, Name: "Vines of Vastwood", TypeLine: "Instant",
		OracleID: b38VinesOfVastwoodOracle, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, vines, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("the narrowed clause refuses an opponent's creature")
	}
	if err := g.CastSpell(me.ID, vines, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("cast Vines of Vastwood: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("\"can't be the target of spells or abilities your opponents control\" is hexproof")
	}
	// Declared caveat: no kicker, so no +4/+4.
	if p, tough := b38PT(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("kicker is not implemented, so the body is untouched: %d/%d", p, tough)
	}
}

func TestB38DruidsDeliveranceSavesItsControllerAndNobodyElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	// The opponent is the attacker; we cast the Fog on our own turn's
	// combat is not possible, so give the turn to them.
	advanceToMainOf(t, g, 1)
	attacker := b12Creature(g, opp.ID, "Beater", "Creature — Ogre", 4, 4)
	second := b12Creature(g, opp.ID, "Other Beater", "Creature — Ogre", 3, 3)
	deliverance := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: deliverance, Name: "Druid's Deliverance", TypeLine: "Instant",
		OracleID: b38DruidsDeliveranceOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, deliverance, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Druid's Deliverance: %v", err)
	}
	passPriorityAroundTable(t, g)
	mineBefore, otherBefore := me.Life, other.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatalf("DeclareAttacker at me: %v", err)
	}
	if err := g.DeclareAttacker(second, other.ID); err != nil {
		t.Fatalf("DeclareAttacker at the third seat: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Life != mineBefore {
		t.Errorf("combat damage to the Deliverance's controller is prevented: %d → %d", mineBefore, me.Life)
	}
	if other.Life != otherBefore-3 {
		t.Errorf("a third player's combat damage is NOT prevented — this is not a Fog: %d → %d", otherBefore, other.Life)
	}
}

func TestB38ViridianRevelDrawsOnlyForAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Viridian Revel", "Enchantment", b38ViridianRevelOracle, false)
	mine := b12Permanent(g, me.ID, "My Signet", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Signet", "Artifact")

	// Your own artifact dying is somebody else's graveyard from the
	// Revel's point of view only if the OWNER is an opponent.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil || triggerOnStack(g, mine) != nil {
		t.Error("an artifact that went to YOUR graveyard draws nothing")
	}

	before := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("an artifact in an opponent's graveyard draws: %d → %d", before, me.Hand.Size())
	}
}

func TestB38WilsonCannotBeCounteredAndWardsAgainstOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, ok := Lookup(b38WilsonOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("\"this spell can't be countered\" is the card's own rider on the Spec")
	}
	wilson := pushCatalogPermanent(g, me.ID, "Wilson, Refined Grizzly", "Legendary Creature — Bear Warrior", b38WilsonOracle, false)
	for _, kw := range []string{"reach", "vigilance", "trample"} {
		if !hasEffectiveKeyword(t, g, wilson, kw) {
			t.Errorf("printed %s reached the effective abilities", kw)
		}
	}
	// An opponent targeting him owes {2}.
	advanceToMainOf(t, g, 1)
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
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("ward {2} asks the targeting opponent for the payment")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(bolt) {
		t.Error("declining the ward counters the spell")
	}
}
