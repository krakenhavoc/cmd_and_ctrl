package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b194_cards_test.go — OLD batch 194 (Commander 2018, #5526): eleven
// cards. Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	oldB194SentinelOracle    = "26396b80-e5b7-44f5-916c-56a9215a7a49"
	oldB194VesselOracle      = "8031285b-2add-42da-aa97-03c3f2fff226"
	oldB194RenaissanceOracle = "e2375f13-fd45-4087-abd1-c884acc6c736"
	oldB194CloudformOracle   = "2b9384a0-db2d-49e2-b121-6b81a0d77722"
	oldB194InfiltratorOracle = "39b359c1-48b8-43a7-b972-54764eadd2c7"
	oldB194LightformOracle   = "f2df375c-3741-409f-91d1-21bf632b2e1d"
	oldB194BargainOracle     = "444dea65-9be7-4b64-afd1-eebe44b285be"
	oldB194HagraOracle       = "73cbb4f0-90d6-4228-9341-791e3e9c4b48"
	oldB194ScionOracle       = "99c8647f-e08d-4217-b637-3f8f2bffdaa3"
	oldB194FinestHourOracle  = "05678253-a40f-4c11-8ac4-335b3a02532e"
	oldB194EnchantressOracle = "495fc654-7986-4d76-9d5c-bc1484e98de2"
)

func TestOldB194EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB194SentinelOracle:    "Silent Sentinel",
		oldB194VesselOracle:      "Vessel of Endless Rest",
		oldB194RenaissanceOracle: "Creeping Renaissance",
		oldB194CloudformOracle:   "Cloudform",
		oldB194InfiltratorOracle: "Jeskai Infiltrator",
		oldB194LightformOracle:   "Lightform",
		oldB194BargainOracle:     "Moonlight Bargain",
		oldB194HagraOracle:       "Retreat to Hagra",
		oldB194ScionOracle:       "Elderwood Scion",
		oldB194FinestHourOracle:  "Finest Hour",
		oldB194EnchantressOracle: "Yavimaya Enchantress",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB194In(z *game.Zone, id uuid.UUID) bool {
	for _, c := range z.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB194SilentSentinelReturnsAnEnchantmentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinel := effectstest.PushDiesCreatureForTest(g, me.ID, "Silent Sentinel", oldB194SentinelOracle, "Creature — Archon", 4, 6)
	ench := effectstest.PushGraveyardPermanent(me, "Pacifism", "Enchantment — Aura", "{1}{W}")
	notEnch := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Enchantment", "Enchantment", "{2}")

	effectstest.DeclareAttack(t, g, opp.ID, sentinel)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if !effectstest.HasID(prompt.PickTargetCards, ench) {
		t.Error("my enchantment card is not a legal target")
	}
	if effectstest.HasID(prompt.PickTargetCards, notEnch) || effectstest.HasID(prompt.PickTargetCards, theirs) {
		t.Error("a creature card or an opponent's graveyard card is offered")
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB194In(me.Graveyard, ench) {
		t.Error("the enchantment is still in the graveyard")
	}
	if c := effectstest.FindCardAnywhere(t, g, ench); c.Name != "Pacifism" || !oldB194In(g.Battlefield, ench) {
		t.Error("the enchantment did not return to the battlefield")
	}
}

func TestOldB194SilentSentinelIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinel := effectstest.PushDiesCreatureForTest(g, me.ID, "Silent Sentinel", oldB194SentinelOracle, "Creature — Archon", 4, 6)
	ench := effectstest.PushGraveyardPermanent(me, "Pacifism", "Enchantment — Aura", "{1}{W}")
	effectstest.DeclareAttack(t, g, opp.ID, sentinel)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB194In(me.Graveyard, ench) {
		t.Error("the enchantment left the graveyard after the player declined")
	}
}

func TestOldB194VesselTucksAnyGraveyardCardAndMakesAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.CastCatalogSpell(t, g, "Vessel of Endless Rest", "Artifact", oldB194VesselOracle, nil)
	effectstest.AdResolveSpell(t, g, spell)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB194In(opp.Graveyard, victim) {
		t.Fatal("the card is still in the graveyard")
	}
	if bottom := opp.Library.Cards[0].InstanceID; bottom != victim {
		t.Errorf("the card is not on the bottom of its owner's library (bottom %v)", bottom)
	}
	before := len(me.ManaPool.Units())
	if err := g.ActivateManaAbility(me.ID, spell, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool.Units()) != before+1 {
		t.Errorf("tapping added %d mana, want 1", len(me.ManaPool.Units())-before)
	}
}

func TestOldB194CreepingRenaissanceReturnsEveryCardOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushGraveyardPermanent(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	artCreature := effectstest.PushGraveyardPermanent(me, "Myr", "Artifact Creature — Myr", "{2}")
	forest := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	bolt := effectstest.PushGraveyardPermanent(me, "Lightning Bolt", "Instant", "{R}")
	theirBear := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Creeping Renaissance", "Sorcery", oldB194RenaissanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	words := effectstest.PickOptionWords(t, g, me.ID)
	if len(words) != 6 {
		t.Fatalf("options %v, want the six permanent types", words)
	}
	idx := -1
	for i, w := range words {
		if w == "creature" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no creature option in %v", words)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB194In(me.Hand, bear) || !oldB194In(me.Hand, artCreature) {
		t.Error("a creature card did not return to the hand")
	}
	if oldB194In(me.Hand, forest) || oldB194In(me.Hand, bolt) || oldB194In(me.Hand, theirBear) {
		t.Error("a card of another type, or an opponent's card, returned")
	}
	if !oldB194In(me.Graveyard, forest) || !oldB194In(opp.Graveyard, theirBear) {
		t.Error("cards that should have stayed in graveyards left")
	}
}

func TestOldB194CreepingRenaissanceFlashbackExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushGraveyardPermanent(me, "Forest", "Basic Land — Forest", "")
	effectstest.FillPoolColored(me, "G", 7)
	spell, err := effectstest.CastFlashbackFromGraveyard(t, g, oldB194RenaissanceOracle)
	if err != nil {
		t.Fatalf("flashback cast refused: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	words := effectstest.PickOptionWords(t, g, me.ID)
	for i, w := range words {
		if w == "land" {
			effectstest.AnswerOptionPick(t, g, me.ID, i)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB194In(me.Hand, forest) {
		t.Error("the land did not return")
	}
	if oldB194In(me.Graveyard, spell) {
		t.Error("the flashback spell is still in the graveyard instead of exile")
	}
}

// oldB194AuraManifest casts the enchantment with a known top card and resolves
// its enters trigger, returning the enchantment and the manifested card.
func oldB194AuraManifest(t *testing.T, g *game.Game, name, oracle string, top game.Card) (aura, manifested uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.Library.Cards = nil
	manifested = effectstest.PushLibraryCard(me, top)
	aura = effectstest.CastCatalogSpell(t, g, name, "Enchantment", oracle, nil)
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	return aura, manifested
}

func TestOldB194CloudformBecomesAnAuraOnTheManifestedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	aura, m := oldB194AuraManifest(t, g, "Cloudform", oldB194CloudformOracle,
		game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	c := effectstest.GaLayered(t, g, aura)
	if !c.IsAura() || !c.IsAttachedTo(m) {
		t.Fatalf("Cloudform aura=%v attached=%v, want an Aura on the manifested card", c.IsAura(), c.IsAttachedTo(m))
	}
	mc := effectstest.GaLayered(t, g, m)
	if !mc.FaceDown || mc.CurrentPower() != 2 || mc.CurrentToughness() != 2 || !mc.IsCreature() {
		t.Errorf("manifested card facedown=%v %d/%d creature=%v, want a face-down 2/2 creature", mc.FaceDown, mc.CurrentPower(), mc.CurrentToughness(), mc.IsCreature())
	}
	for _, kw := range []string{"flying", "hexproof"} {
		if !effectstest.HasEffectiveKeyword(t, g, m, kw) {
			t.Errorf("enchanted creature lacks %s", kw)
		}
	}
	if !oldB194In(g.Battlefield, aura) {
		t.Error("Cloudform left the battlefield")
	}
}

func TestOldB194CloudformWithAnEmptyLibraryIsPutIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	aura := effectstest.CastCatalogSpell(t, g, "Cloudform", "Enchantment", oldB194CloudformOracle, nil)
	effectstest.AdResolveSpell(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !oldB194In(me.Graveyard, aura) {
		t.Error("an Aura attached to nothing stayed on the battlefield")
	}
}

func TestOldB194LightformGrantsFlyingAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	aura, m := oldB194AuraManifest(t, g, "Lightform", oldB194LightformOracle,
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	c := effectstest.GaLayered(t, g, aura)
	if !c.IsAura() || !c.IsAttachedTo(m) {
		t.Fatalf("Lightform aura=%v attached=%v", c.IsAura(), c.IsAttachedTo(m))
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, m, kw) {
			t.Errorf("enchanted creature lacks %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, m, "hexproof") {
		t.Error("Lightform grants hexproof")
	}
}

func TestOldB194JeskaiInfiltratorCantBeBlockedWhileAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	inf := effectstest.B12Push(g, me.ID, "Jeskai Infiltrator", "Creature — Human Monk", oldB194InfiltratorOracle, 2, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AttackIntoBlocks(t, g, opp.ID, inf)
	if r := effectstest.BlockRefusal(t, g, blocker, inf); r == "" {
		t.Error("the Infiltrator was blockable while its controller controls no other creature")
	}
}

func TestOldB194JeskaiInfiltratorIsBlockableWithAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	inf := effectstest.B12Push(g, me.ID, "Jeskai Infiltrator", "Creature — Human Monk", oldB194InfiltratorOracle, 2, 3)
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AttackIntoBlocks(t, g, opp.ID, inf)
	if r := effectstest.BlockRefusal(t, g, blocker, inf); r != "" {
		t.Errorf("the Infiltrator was unblockable with another creature around: %v", r)
	}
}

// oldB194Manifested returns the face-down manifested permanents on the
// battlefield, as their layered views.
func oldB194Manifested(t *testing.T, g *game.Game) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.FaceDown && c.FaceDownKind == game.FaceDownManifested {
			out = append(out, effectstest.GaLayered(t, g, c.InstanceID))
		}
	}
	return out
}

func TestOldB194JeskaiInfiltratorManifestsItselfAndTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := effectstest.B12Push(g, me.ID, "Jeskai Infiltrator", "Creature — Human Monk", oldB194InfiltratorOracle, 2, 3)
	me.Library.Cards = nil
	second := effectstest.PushLibraryCard(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.DealCombatDamage(g, me.ID, inf, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	got := oldB194Manifested(t, g)
	if len(got) != 2 {
		t.Fatalf("%d manifested permanents, want 2 (the Infiltrator and the top card)", len(got))
	}
	for _, c := range got {
		if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.IsCreature() || c.Controller != me.ID {
			t.Errorf("manifested %d/%d creature=%v mine=%v, want a 2/2 creature I control", c.CurrentPower(), c.CurrentToughness(), c.IsCreature(), c.Controller == me.ID)
		}
	}
	if g.Exile.Size() != 0 {
		t.Errorf("%d cards left in exile", g.Exile.Size())
	}
	if !oldB194In(me.Library, second) || me.Library.Size() != 1 {
		t.Error("the second card of the library was taken, or the top one stayed")
	}
}

func TestOldB194JeskaiInfiltratorManifestsAnInstantFromTheTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inf := effectstest.B12Push(g, me.ID, "Jeskai Infiltrator", "Creature — Human Monk", oldB194InfiltratorOracle, 2, 3)
	me.Library.Cards = nil
	effectstest.PushLibraryCard(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.DealCombatDamage(g, me.ID, inf, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB194Manifested(t, g); len(got) != 2 {
		t.Fatalf("%d manifested permanents, want 2 (an instant is manifested too)", len(got))
	}
	if me.Library.Size() != 0 {
		t.Error("the top card is still in the library")
	}
}

func TestOldB194MoonlightBargainKeepsTheCardsYouPayFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	sixth := effectstest.PushLibraryCard(me, game.Card{Name: "Sixth", TypeLine: "Sorcery"})
	var ids []uuid.UUID
	for _, n := range []string{"E", "D", "C", "B", "A"} {
		ids = append(ids, effectstest.PushLibraryCard(me, game.Card{Name: n, TypeLine: "Sorcery"}))
	}
	// A is the top card, then B, C, D and E.
	top := []uuid.UUID{ids[4], ids[3], ids[2], ids[1], ids[0]}
	life := me.Life
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Moonlight Bargain", "Instant", oldB194BargainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Pay for the 1st and 3rd card, decline the rest (option 1 is "Pay 2 life").
	for i := range top {
		effectstest.AnswerOptionPick(t, g, me.ID, map[bool]int{true: 1, false: 0}[i%2 == 0])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-6 {
		t.Errorf("life %d -> %d, want 6 paid for three cards", life, me.Life)
	}
	for i, id := range top {
		kept := i%2 == 0
		if kept && !oldB194In(me.Hand, id) {
			t.Errorf("card %d was paid for but is not in the hand", i)
		}
		if !kept && !oldB194In(me.Graveyard, id) {
			t.Errorf("card %d was not paid for but is not in the graveyard", i)
		}
	}
	if got := me.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand grew by %d, want 3", got)
	}
	if !oldB194In(me.Library, sixth) {
		t.Error("the sixth card left the library")
	}
}

func TestOldB194MoonlightBargainWithOneLifeCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	a := effectstest.PushLibraryCard(me, game.Card{Name: "A", TypeLine: "Sorcery"})
	me.Life = 1
	effectstest.CastCatalogSpell(t, g, "Moonlight Bargain", "Instant", oldB194BargainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if words := effectstest.PickOptionWords(t, g, me.ID); len(words) != 1 {
		t.Fatalf("options %v, want only the graveyard", words)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 1 || !oldB194In(me.Graveyard, a) {
		t.Errorf("life=%d in graveyard=%v, want 1 and yes", me.Life, oldB194In(me.Graveyard, a))
	}
}

func oldB194HagraLand(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.PlayLandFromHand(t, g, "Forest", "")
}

func TestOldB194RetreatToHagraDrainsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Retreat to Hagra", oldB194HagraOracle, "Enchantment")
	myLife, theirLife := me.Life, opp.Life
	oldB194HagraLand(t, g)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life me %d -> %d, opponent %d -> %d, want +1 and -1", myLife, me.Life, theirLife, opp.Life)
	}
}

func TestOldB194RetreatToHagraPumpsAndGivesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushPermanentForTest(g, me.ID, "Retreat to Hagra", oldB194HagraOracle, "Enchantment")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB194HagraLand(t, g)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 0)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("bear is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("no deathtouch")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("deathtouch did not end with the turn")
	}
}

func TestOldB194ElderwoodScionTaxesOpponentsAndDiscountsYou(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	scion := effectstest.B12Push(g, me.ID, "Elderwood Scion", "Creature — Elemental", oldB194ScionOracle, 4, 4)
	for _, kw := range []string{"trample", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, scion, kw) {
			t.Errorf("no %s", kw)
		}
	}
	aura := func(p *game.Player) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: "Asha's Favor", TypeLine: "Enchantment — Aura", ManaCost: "{1}{W}",
				OracleID: "97e7ada6-dee1-4bb3-a42d-38a61d946bd4", Colors: []string{"W"}, Owner: p.ID, Controller: p.ID})
		})
		return id
	}
	effectstest.AdvanceToMain(t, g)
	mine := aura(me)
	effectstest.FlashIfFund(g, me, "W")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(scion)}); err != nil {
		t.Fatalf("my {1}{W} Aura at the Scion with only {W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.ActiveSeat == 1 {
			break
		}
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	theirs := aura(opp)
	effectstest.FlashIfFund(g, opp, "W", "C", "C")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(scion)}); err == nil {
		t.Fatal("an opponent cast a {1}{W} Aura at the Scion with {W}{C}{C}")
	}
	effectstest.FlashIfFund(g, opp, "W", "C", "C", "C")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(scion)}); err != nil {
		t.Fatalf("with {W}{C}{C}{C}: %v", err)
	}
}

func TestOldB194FinestHourAddsACombatAndUntapsTheLoneAttackerOnlyInTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hour := effectstest.PushPermanentForTest(g, me.ID, "Finest Hour", oldB194FinestHourOracle, "Enchantment")
	if !effectstest.HasEffectiveKeyword(t, g, hour, game.KeywordExalted) {
		t.Error("Finest Hour has no exalted")
	}
	atk := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, atk) {
		t.Error("the lone attacker is still tapped")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Errorf("after the first combat %+v, want another combat", next)
	}
	// The additional combat is not the first: nothing triggers.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, atk) {
		t.Error("the attacker was untapped by the second combat's attack")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the additional combat %+v, want the main phase", next)
	}
}

func TestOldB194FinestHourDoesNothingForTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Finest Hour", oldB194FinestHourOracle, "Enchantment")
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, a) {
		t.Error("an attacker was untapped though two attacked")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the combat %+v, want the main phase", next)
	}
}

func TestOldB194YavimayaEnchantressCountsEveryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	e := effectstest.B12Push(g, me.ID, "Yavimaya Enchantress", "Creature — Human Druid", oldB194EnchantressOracle, 2, 2)
	pt := func() (int, int) {
		c := effectstest.GaLayered(t, g, e)
		return c.CurrentPower(), c.CurrentToughness()
	}
	if p, tg := pt(); p != 2 || tg != 2 {
		t.Fatalf("with no enchantments it is %d/%d, want 2/2", p, tg)
	}
	effectstest.B12Push(g, me.ID, "Mine", "Enchantment", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Theirs", "Enchantment — Aura", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Both", "Enchantment Creature — Spirit", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Not", "Artifact", "", 1, 1)
	if p, tg := pt(); p != 5 || tg != 5 {
		t.Errorf("with three enchantments it is %d/%d, want 5/5", p, tg)
	}
}
