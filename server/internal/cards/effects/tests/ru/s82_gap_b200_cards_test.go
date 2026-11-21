package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b200_cards_test.go — Double Masters 2022 artifacts and creatures,
// GAP batch 200 (#5948). Each test plays the real catalog entry through the
// engine.

const (
	gap200DragonArchOracle = "fdf21a39-7392-49b1-b9c4-ee8122631cab"
	gap200LashOracle       = "9d3f1b93-ec3b-47eb-bd93-89890fd33ccb"
	gap200NimOracle        = "66d41377-626d-4ae6-ba86-17bf0c8b3362"
	gap200StoneOracle      = "b2bd40a1-1857-4f4c-aed6-d4e224ac0ef7"
	gap200EmrakulOracle    = "900ca697-ad38-4b2b-bc74-2ff7eb6ea951"
	gap200ManglerOracle    = "71358110-4e28-4794-befa-ea6e27e720ef"
	gap200TrollOracle      = "61b1d7e5-6155-4204-b110-35a890551ec8"
	gap200SkullbriarOracle = "088ac883-6f44-4046-bf19-81c7fd952e55"
	gap200BrothersOracle   = "7fb63d9a-8d90-4b43-8390-924de2d7e32c"
)

func TestGap200EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gap200DragonArchOracle: "Dragon Arch",
		gap200LashOracle:       "Livewire Lash",
		gap200NimOracle:        "Nim Deathmantle",
		gap200StoneOracle:      "Thrumming Stone",
		gap200EmrakulOracle:    "Emrakul, the Aeons Torn",
		gap200ManglerOracle:    "Dreg Mangler",
		gap200TrollOracle:      "Lotleth Troll",
		gap200SkullbriarOracle: "Skullbriar, the Walking Grave",
		gap200BrothersOracle:   "Shattergang Brothers",
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

func gap200Card(g *game.Game, id uuid.UUID) game.Card {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	return game.Card{}
}

// --- Dragon Arch ---------------------------------------------------------------

func TestGap200DragonArchPutsOnlyAMulticoloredCreatureFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arch := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Arch", "Artifact", gap200DragonArchOracle, false)
	gold := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: gold, Name: "Gold Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{B}{G}",
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	mono := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mono, Name: "Green Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	goldSpell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: goldSpell, Name: "Gold Charm", TypeLine: "Instant", ManaCost: "{B}{G}",
		Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, arch, 0, game.ActivateAbilityParams{})
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt opens")
	}
	if !effectstest.HasID(c.ChooseCards, gold) {
		t.Error("the multicolored creature card is offered")
	}
	if effectstest.HasID(c.ChooseCards, mono) || effectstest.HasID(c.ChooseCards, goldSpell) {
		t.Error("a monocolored creature or a multicolored instant is offered")
	}
	if c.ChooseMin != 0 {
		t.Error("\"you MAY put\": declining is legal")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{gold}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(gold) || me.Hand.Contains(gold) {
		t.Error("the creature arrives on the battlefield")
	}
	if !effectstest.PermanentTapped(g, arch) {
		t.Error("{T} is part of the cost")
	}
}

func TestGap200DragonArchNeedsTwoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arch := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Arch", "Artifact", gap200DragonArchOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
}

// --- Livewire Lash -------------------------------------------------------------

func gap200Equipped(t *testing.T, g *game.Game, me *game.Player, oracle, name string) (equip, creature uuid.UUID) {
	t.Helper()
	equip = effectstest.SeedEquipment(g, me.ID, name, oracle)
	creature = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.EquipTo(t, g, me.ID, equip, creature)
	return equip, creature
}

func TestGap200LivewireLashPumpsAndPingsWhenTargetedBySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, bear := gap200Equipped(t, g, me, gap200LashOracle, "Livewire Lash")
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Fatalf("equipped Bear is %d/%d, want 4/2", p, tough)
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", "5748ebf1-24e3-499d-ab7c-c2cebd462a24", []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 damage", life, opp.Life)
	}
}

func TestGap200LivewireLashIgnoresAnAbilityTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, bear := gap200Equipped(t, g, me, gap200LashOracle, "Livewire Lash")
	other := effectstest.SeedEquipment(g, me.ID, "Spare Lash", gap200LashOracle)
	life := opp.Life
	effectstest.B06AddMana(me, "C", "C")
	effectstest.EquipTo(t, g, me.ID, other, bear) // an equip ability targets the creature
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OpenPickTarget(g) {
		t.Fatal("an ability targeting the creature asked for a target")
	}
	if opp.Life != life {
		t.Error("an ability targeting the creature dealt damage")
	}
}

// --- Nim Deathmantle -----------------------------------------------------------

func TestGap200NimDeathmantleGrantsItsStatsAndFormAndAColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, bear := gap200Equipped(t, g, me, gap200NimOracle, "Nim Deathmantle")
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveAbility(t, g, bear, "intimidate") {
		t.Error("equipped creature has no intimidate")
	}
	if cols := effectstest.EffectiveColors(t, g, bear); len(cols) != 1 || cols[0] != "B" {
		t.Errorf("colors %v, want black", cols)
	}
	card := gap200Card(g, bear)
	if !card.HasSubtype("Zombie") || card.HasSubtype("Bear") {
		t.Errorf("subtypes %q: want a Zombie and no longer a Bear", card.TypeLine)
	}
}

func TestGap200NimDeathmantleReturnsADyingCreatureAndAttachesForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	equip, bear := gap200Equipped(t, g, me, gap200NimOracle, "Nim Deathmantle")
	other := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) && effectstest.B12ZoneOf(g, other) != game.ZoneBattlefield {
		// the returned card is a new object but keeps the card's ID
		t.Fatalf("the Elk is in %v, want the battlefield", effectstest.B12ZoneOf(g, other))
	}
	if got := gap200Card(g, equip).AttachedTo.ID; got != other {
		t.Errorf("the Equipment is attached to %v, want the returned Elk", got)
	}
	_ = bear
}

func TestGap200NimDeathmantleDoesNothingWhenYouDeclineToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap200Equipped(t, g, me, gap200NimOracle, "Nim Deathmantle")
	other := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.KillCreature(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, other) != game.ZoneGraveyard {
		t.Errorf("the Elk is in %v, want the graveyard", effectstest.B12ZoneOf(g, other))
	}
}

func TestGap200NimDeathmantleIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gap200Equipped(t, g, me, gap200NimOracle, "Nim Deathmantle")
	theirs := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's creature offered the payment")
	}
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneGraveyard {
		t.Error("an opponent's creature left the graveyard")
	}
}

func TestGap200NimDeathmantleIgnoresATokenAndANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gap200Equipped(t, g, me, gap200NimOracle, "Nim Deathmantle")
	tok := effectstest.PushToken(g, me.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1})
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.KillCreature(t, g, me.ID, tok)
	effectstest.KillCreature(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a token or an artifact offered the payment")
	}
}

// --- Thrumming Stone -----------------------------------------------------------

func TestGap200ThrummingStoneRippleCastsSameNamedSpellsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thrumming Stone", "Legendary Artifact", gap200StoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.HandSpell(me, "Ripplet", "Sorcery", "{R}")
	effectstest.B06AddMana(me, "R")
	var top [4]uuid.UUID
	for i, name := range []string{"Ripplet", "Other", "Ripplet", "Another"} {
		top[3-i] = effectstest.SeedLibraryTop(me, name, "Sorcery", "{R}{R}{R}")
	}
	// top[3] is the first pushed: the deepest of the four. Push order made
	// "Another" the top card, then "Ripplet", "Other", "Ripplet".
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	ripplets := []uuid.UUID{top[1], top[3]}
	_ = effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, ripplets...)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, ripplets[0], game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the first Ripplet: %v", err)
	}
	// Casting from among them is each its own "may": the other is still on offer.
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, ripplets[1])
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if z := effectstest.B12ZoneOf(g, top[0]); z != game.ZoneLibrary {
		t.Errorf("the uncast card is in %v, want the library", z)
	}
	if z := effectstest.B12ZoneOf(g, ripplets[1]); z != game.ZoneLibrary {
		t.Errorf("the declined Ripplet is in %v, want the library", z)
	}
}

func TestGap200ThrummingStoneRippleCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thrumming Stone", "Legendary Artifact", gap200StoneOracle, false)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.HandSpell(me, "Ripplet", "Sorcery", "{R}")
	effectstest.B06AddMana(me, "R")
	same := effectstest.SeedLibraryTop(me, "Ripplet", "Sorcery", "{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a cast was offered although the reveal was declined")
	}
	if effectstest.B12ZoneOf(g, same) != game.ZoneLibrary {
		t.Error("the card left the library")
	}
}

// --- Emrakul, the Aeons Torn ---------------------------------------------------

func TestGap200EmrakulCastTakesAnExtraTurnAndCannotBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(gap200EmrakulOracle)
	if !spec.CantBeCountered {
		t.Error("Emrakul can be countered")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Emrakul, the Aeons Torn", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: gap200EmrakulOracle, ManaCost: "{15}", Power: 15, Toughness: 15, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// The cast trigger is on the stack above Emrakul, not waiting for it to resolve.
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(g.ExtraTurnsQueuedForEffect()); n != 1 {
		t.Fatalf("%d extra turns queued after the cast trigger, want 1", n)
	}
}

func TestGap200EmrakulFliesAndAnnihilates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	em := effectstest.PushCatalogPermanent(g, me.ID, "Emrakul, the Aeons Torn", "Legendary Creature — Eldrazi", gap200EmrakulOracle, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, em, "flying") {
		t.Error("Emrakul does not fly")
	}
	effectstest.PushAnnFodder(g, opp.ID, 8)
	effectstest.DeclareAttack(t, g, opp.ID, em)
	effectstest.AnnSettle(t, g, nil)
	if got := effectstest.AnnCountOn(g, opp.ID); got != 2 {
		t.Errorf("defender has %d permanents left, want 2 (annihilator 6)", got)
	}
}

func TestGap200EmrakulShufflesTheGraveyardIntoTheLibraryFromAnywhere(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	em := effectstest.PushCatalogPermanent(g, me.ID, "Emrakul, the Aeons Torn", "Legendary Creature — Eldrazi", gap200EmrakulOracle, false)
	effectstest.PushGraveyardCardTyped(me, "Old Bones", "Creature — Skeleton")
	lib := me.Library.Size()
	effectstest.KillCreature(t, g, me.ID, em)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard holds %d cards, want 0", me.Graveyard.Size())
	}
	if me.Library.Size() != lib+2 {
		t.Errorf("library %d -> %d, want +2 (Emrakul and the other card)", lib, me.Library.Size())
	}
}

func TestGap200EmrakulProtectionFromColoredSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	em := effectstest.PushCatalogPermanent(g, me.ID, "Emrakul, the Aeons Torn", "Legendary Creature — Eldrazi", gap200EmrakulOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(opp, "R", "R", "C", "C")
	// A red spell cannot target it (strictly).
	bolt := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: bolt, Name: "Red Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070",
		Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: em}}})
	if err == nil {
		t.Fatal("a red spell targeted Emrakul")
	}
	// A colorless spell can.
	rock := uuid.New()
	g.Seats[0].Hand.PushTop(game.Card{InstanceID: rock, Name: "Colorless Bolt", TypeLine: "Instant", ManaCost: "{1}", OracleID: "4457ed35-7c10-48c8-9776-456485fdf070",
		Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "C")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: em}}}); err != nil {
		t.Fatalf("a colorless spell could not target Emrakul: %v", err)
	}
}

// --- Dreg Mangler --------------------------------------------------------------

func TestGap200DregManglerHasHasteAndScavenges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	live := effectstest.PushCatalogPermanent(g, me.ID, "Dreg Mangler", "Creature — Plant Zombie", gap200ManglerOracle, true)
	if !effectstest.HasKeywordOnBattlefield(t, g, live, "haste") {
		t.Error("no haste")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	m, _ := effectstest.ActivateFromGraveyard(t, g, "Dreg Mangler", "Creature — Plant Zombie", gap200ManglerOracle, 3, 3, "{C}{C}{C}{B}{G}",
		game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, bear); got != 3 {
		t.Errorf("Bear has %d +1/+1 counters, want 3 (Dreg Mangler's power)", got)
	}
	if effectstest.B12ZoneOf(g, m) != game.ZoneExile {
		t.Error("scavenge exiles the card as a cost")
	}
}

// --- Lotleth Troll -------------------------------------------------------------

func TestGap200LotlethTrollGrowsByDiscardingACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.PushCatalogPermanent(g, me.ID, "Lotleth Troll", "Creature — Zombie Troll", gap200TrollOracle, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, troll, "trample") {
		t.Error("no trample")
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Sorcery Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	fodder := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, troll); got != 1 {
		t.Errorf("Troll has %d counters, want 1", got)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the creature card was not discarded")
	}
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Error("a noncreature card paid the discard")
	}
}

func TestGap200LotlethTrollRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	troll := effectstest.PushCatalogPermanent(g, me.ID, "Lotleth Troll", "Creature — Zombie Troll", gap200TrollOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, troll, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B15Destroy(t, g, troll)
	if !g.Battlefield.Contains(troll) {
		t.Error("the Troll died despite its regeneration shield")
	}
}

// --- Skullbriar, the Walking Grave ----------------------------------------------

func TestGap200SkullbriarGrowsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := effectstest.PushCatalogPermanent(g, me.ID, "Skullbriar, the Walking Grave", "Legendary Creature — Zombie Elemental", gap200SkullbriarOracle, true)
	if !effectstest.HasKeywordOnBattlefield(t, g, s, "haste") {
		t.Error("no haste")
	}
	effectstest.DeclareAttack(t, g, opp.ID, s)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr7Counters(g, s); got != 1 {
		t.Errorf("Skullbriar has %d counters, want 1", got)
	}
}

func TestGap200SkullbriarKeepsItsCountersExceptInAHandOrLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.PushCatalogPermanent(g, me.ID, "Skullbriar, the Walking Grave", "Legendary Creature — Zombie Elemental", gap200SkullbriarOracle, false)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(s, game.CounterPlusOne, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	counters := func() int {
		n := 0
		for _, z := range []*game.Zone{me.Graveyard, g.Exile, g.Battlefield, me.Hand, me.Library} {
			for _, c := range z.Cards {
				if c.InstanceID == s {
					n = c.Counters[game.CounterPlusOne]
				}
			}
		}
		return n
	}
	move := func(from, to game.ZoneRef) {
		t.Helper()
		if err := g.MoveCardByID(from, to, s); err != nil {
			t.Fatalf("move: %v", err)
		}
	}
	move(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID})
	if counters() != 3 {
		t.Fatalf("%d counters in the graveyard, want 3", counters())
	}
	move(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile})
	if counters() != 3 {
		t.Fatalf("%d counters in exile, want 3", counters())
	}
	move(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield})
	if counters() != 3 {
		t.Fatalf("%d counters back on the battlefield, want 3", counters())
	}
	move(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID})
	if counters() != 0 {
		t.Fatalf("%d counters in the hand, want 0", counters())
	}
}

// --- Shattergang Brothers ------------------------------------------------------

func TestGap200ShattergangBrothersEdictsEachKind(t *testing.T) {
	for _, tc := range []struct {
		idx      int
		mana     []string
		typeLine string
		sac      string
	}{
		{0, []string{"C", "C", "B"}, "Creature — Bear", "Creature — Elk"},
		{1, []string{"C", "C", "R"}, "Artifact", "Artifact"},
		{2, []string{"C", "C", "G"}, "Enchantment", "Enchantment"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
		bros := effectstest.PushCatalogPermanent(g, me.ID, "Shattergang Brothers", "Legendary Creature — Goblin Artificer", gap200BrothersOracle, false)
		fodder := effectstest.B12Permanent(g, me.ID, "Fodder", tc.sac)
		keepA := effectstest.B12Permanent(g, a.ID, "A1", tc.sac)
		keepB := effectstest.B12Permanent(g, b.ID, "B1", tc.sac)
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, tc.mana...)
		effectstest.B16Activate(t, g, me.ID, bros, tc.idx, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnnSettle(t, g, nil)
		if g.Battlefield.Contains(keepA) || g.Battlefield.Contains(keepB) {
			t.Errorf("ability %d: an opponent kept the only permanent of the kind", tc.idx)
		}
		if !g.Battlefield.Contains(bros) {
			t.Errorf("ability %d: the Brothers were sacrificed", tc.idx)
		}
	}
}

func TestGap200EmrakulTakesNoDamageFromAColoredSpellButIsNotImmuneToAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	em := effectstest.PushCatalogPermanent(g, opp.ID, "Emrakul, the Aeons Torn", "Legendary Creature — Eldrazi", gap200EmrakulOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Pyroclasm", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: "e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived Pyroclasm")
	}
	if dmg := gap200Card(g, em).DamageMarked; dmg != 0 {
		t.Errorf("Emrakul has %d damage from a red spell, want none", dmg)
	}
}
