package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, batch 33 (Modern Horizons 2: token makers, #5501).

const (
	oldB033VerdantCommand = "5a5d426b-e5eb-44f0-b5fe-c258bec6d59e"
	oldB033Rokiric        = "64c15098-f893-44f2-bf7f-162db8ab80e9"
	oldB033Sanctuary      = "693d20e9-ad92-44ad-9190-d247c94660eb"
	oldB033FaeOffering    = "885d68de-db97-4c9e-bb44-8e561a482dde"
	oldB033Apprentice     = "8fb9ff2d-1e90-4a26-8397-bb981fda2f7b"
	oldB033Aerialist      = "dd39cccb-095c-4ef4-8423-f4d328849295"
	oldB033Magus          = "e48951ed-1c32-449a-898e-e5b1c0d0ee73"
	oldB033Hermit         = "f7f4ddc9-cea0-42f9-bbca-2ad87a24bfc8"
	oldB033Traprunner     = "fbdffef0-9119-4650-9f92-58e8a664db6f"
	oldB033Collector      = "fdb51cc1-ce0e-4920-b123-1cdd570fc38b"
	oldB033JaredOracle    = "f9485c76-122d-4faf-9194-b84ed9c41378"
)

func TestOldB033CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB033VerdantCommand: "Verdant Command",
		oldB033Rokiric:        "General Ferrous Rokiric",
		oldB033Sanctuary:      "Squirrel Sanctuary",
		oldB033FaeOffering:    "Fae Offering",
		oldB033Apprentice:     "Breya's Apprentice",
		oldB033Aerialist:      "Burdened Aerialist",
		oldB033Magus:          "Magus of the Bridge",
		oldB033Hermit:         "Underworld Hermit",
		oldB033Traprunner:     "Goblin Traprunner",
		oldB033Collector:      "Specimen Collector",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB033Count(g *game.Game, name string, controller uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

func oldB033Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	var out game.ZoneKind
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(id); z != nil {
			out = z.Kind
		}
	})
	return out
}

func oldB033Tapped(g *game.Game, name string, controller uuid.UUID) (tapped, total int) {
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				total++
				if c.Tapped {
					tapped++
				}
			}
		}
	})
	return
}

// castColored casts a free spell of the given colours as `p`.
func oldB033CastColored(t *testing.T, g *game.Game, p *game.Player, typeLine string, colors []string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Spell", TypeLine: typeLine, Colors: colors, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// --- Verdant Command ---------------------------------------------------------

func TestOldB033VerdantCommandTokensAndLifeGoToTheNamedPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Verdant Command", TypeLine: "Instant", OracleID: oldB033VerdantCommand, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Modes:   []int{0, 3},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, total := oldB033Tapped(g, "Squirrel", opp.ID); total != 2 || tapped != 2 {
		t.Errorf("opponent has %d Squirrels (%d tapped), want two tapped", total, tapped)
	}
	if oldB033Count(g, "Squirrel", me.ID) != 0 {
		t.Error("the caster got Squirrels")
	}
	if opp.Life != life {
		t.Errorf("the opponent's life changed to %d", opp.Life)
	}
	if me.Life != 43 && me.Life != 40+3 {
		t.Errorf("caster life %d, want 3 more than a starting 40", me.Life)
	}
}

func TestOldB033VerdantCommandExilesAGraveyardCardAndCountersAPlaneswalkerAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gy := effectstest.PushCatalogGraveyardCard(opp, "Old Bear", "Creature — Bear", "", 2, 2)
	jared := effectstest.PushCatalogPermanent(g, me.ID, "Jared Carthalion", "Legendary Planeswalker — Jared", oldB033JaredOracle, false)
	effectstest.PutCounters(t, g, jared, game.CounterLoyalty, 5)
	if err := g.ActivateCatalogAbility(me.ID, jared, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	ability := effectstest.AcItemLabelled(g, "Jared Carthalion — +1: Create a 3/3 Kavu creature token with trample that's all colors.")
	if ability == uuid.Nil {
		ability = effectstest.AcItemOnStack(g, "+1:").ID
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Verdant Command", TypeLine: "Instant", OracleID: oldB033VerdantCommand, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Modes:   []int{1, 2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ability}, {Kind: game.TargetCard, ID: gy}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Zone(g, gy) != game.ZoneExile {
		t.Errorf("the graveyard card is in %q, want exile", oldB033Zone(g, gy))
	}
	if oldB033Count(g, "Kavu", me.ID) != 0 {
		t.Error("the loyalty ability was not countered: a Kavu entered")
	}
}

func TestOldB033VerdantCommandRefusesANonLoyaltyAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	app := effectstest.PushCatalogPermanent(g, me.ID, "Breya's Apprentice", "Artifact Creature — Human Artificer", oldB033Apprentice, false)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Aid", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, app, 0, game.ActivateAbilityParams{
		Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: app}},
		SacrificeIDs: []uuid.UUID{app},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	ability := effectstest.AcItemOnStack(g, "{T}, Sacrifice")
	if ability == nil {
		t.Fatal("no ability on the stack")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Verdant Command", TypeLine: "Instant", OracleID: oldB033VerdantCommand, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Modes:   []int{1, 3},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ability.ID}, {Kind: game.TargetPlayer, ID: me.ID}},
	})
	if err == nil {
		t.Error("a creature's ability is not a loyalty ability of a planeswalker")
	}
}

// --- General Ferrous Rokiric --------------------------------------------------

func TestOldB033RokiricMakesAGolemOnlyForMulticoloredSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "General Ferrous Rokiric", "Legendary Creature — Human Soldier", oldB033Rokiric, false)
	effectstest.AdvanceToMain(t, g)
	oldB033CastColored(t, g, me, "Instant", []string{"R"})
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB033Count(g, "Golem", me.ID); n != 0 {
		t.Fatalf("a monocolored spell made %d Golems", n)
	}
	oldB033CastColored(t, g, me, "Instant", []string{"R", "W"})
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB033Count(g, "Golem", me.ID); n != 1 {
		t.Fatalf("a multicolored spell made %d Golems, want 1", n)
	}
	golem := effectstest.BattlefieldTokenNamed(t, g, "Golem")
	if golem.Power != 4 || golem.Toughness != 4 || len(golem.Colors) != 2 || !golem.IsArtifact() || !golem.IsCreature() {
		t.Errorf("Golem is %+v, want a 4/4 red and white artifact creature", golem)
	}
}

// --- Squirrel Sanctuary --------------------------------------------------------

func TestOldB033SquirrelSanctuaryEntersWithASquirrelAndBouncesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanct := effectstest.CastCatalogSpell(t, g, "Squirrel Sanctuary", "Enchantment", oldB033Sanctuary, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Count(g, "Squirrel", me.ID) != 1 {
		t.Fatalf("the Sanctuary made %d Squirrels, want 1", oldB033Count(g, "Squirrel", me.ID))
	}
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "G", 1)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if oldB033Zone(g, sanct) != game.ZoneHand {
		t.Errorf("the Sanctuary is in %q, want the hand", oldB033Zone(g, sanct))
	}
}

func TestOldB033SquirrelSanctuaryIgnoresTokensAndStaysWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sanct := effectstest.CastCatalogSpell(t, g, "Squirrel Sanctuary", "Enchantment", oldB033Sanctuary, nil)
	effectstest.PassPriorityAroundTable(t, g)
	tok := effectstest.BattlefieldTokenNamed(t, g, "Squirrel")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tok.InstanceID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a token dying offered the payment")
	}
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if oldB033Zone(g, sanct) != game.ZoneBattlefield {
		t.Errorf("the Sanctuary is in %q after declining, want the battlefield", oldB033Zone(g, sanct))
	}
}

// --- Fae Offering ---------------------------------------------------------------

func TestOldB033FaeOfferingNeedsBothSpellKinds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Fae Offering", "Enchantment", oldB033FaeOffering, false)
	effectstest.AdvanceToMain(t, g)
	oldB033CastColored(t, g, me, "Creature — Bear", []string{"G"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB033Count(g, "Clue", me.ID) + oldB033Count(g, "Food", me.ID) + oldB033Count(g, "Treasure", me.ID); n != 0 {
		t.Fatalf("a creature spell alone made %d tokens", n)
	}
}

func TestOldB033FaeOfferingMakesClueFoodAndTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Fae Offering", "Enchantment", oldB033FaeOffering, false)
	effectstest.AdvanceToMain(t, g)
	oldB033CastColored(t, g, me, "Creature — Bear", []string{"G"})
	effectstest.PassPriorityAroundTable(t, g)
	oldB033CastColored(t, g, me, "Instant", []string{"G"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, n := range []string{"Clue", "Food", "Treasure"} {
		if oldB033Count(g, n, me.ID) != 1 {
			t.Errorf("%s tokens: %d, want 1", n, oldB033Count(g, n, me.ID))
		}
	}
}

// --- Breya's Apprentice ----------------------------------------------------------

func TestOldB033ApprenticeEntersWithAThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Breya's Apprentice", "Artifact Creature — Human Artificer", oldB033Apprentice, nil)
	effectstest.PassPriorityAroundTable(t, g)
	th := effectstest.BattlefieldTokenNamed(t, g, "Thopter")
	if th.Power != 1 || th.Toughness != 1 || len(th.Colors) != 0 || !effectstest.P1g2HasKeyword(g, th.InstanceID, "flying") {
		t.Errorf("Thopter is %+v, want a colorless 1/1 flyer", th)
	}
	if oldB033Count(g, "Thopter", me.ID) != 1 {
		t.Error("expected one Thopter")
	}
}

func TestOldB033ApprenticePumpsAfterSacrificingAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	app := effectstest.PushCatalogPermanent(g, me.ID, "Breya's Apprentice", "Artifact Creature — Human Artificer", oldB033Apprentice, false)
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, app, 0, game.ActivateAbilityParams{
		Strict: true, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{rock},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4", got)
	}
	if oldB033Zone(g, rock) != game.ZoneGraveyard {
		t.Errorf("the artifact is in %q, want the graveyard", oldB033Zone(g, rock))
	}
}

func TestOldB033ApprenticeExileModePlaysTheTopCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	app := effectstest.PushCatalogPermanent(g, me.ID, "Breya's Apprentice", "Artifact Creature — Human Artificer", oldB033Apprentice, false)
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	me.Library.Cards = nil
	top := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, app, 0, game.ActivateAbilityParams{
		Strict: true, Modes: []int{0}, SacrificeIDs: []uuid.UUID{rock},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Zone(g, top) != game.ZoneExile {
		t.Fatalf("the top card is in %q, want exile", oldB033Zone(g, top))
	}
	if !effectstest.B12ExiledPlayableBy(t, g, top, me.ID) {
		t.Error("the exiled card is not playable")
	}
}

// --- Burdened Aerialist ------------------------------------------------------------

func TestOldB033AerialistGainsFlyingWhenYouSacrificeAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Burdened Aerialist", "Creature — Human Pirate", oldB033Aerialist, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Count(g, "Treasure", me.ID) != 1 {
		t.Fatal("the Aerialist did not make a Treasure")
	}
	var air uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Burdened Aerialist" {
				air = c.InstanceID
			}
		}
	})
	if effectstest.HasEffectiveKeyword(t, g, air, "flying") {
		t.Fatal("the Aerialist flies before any sacrifice")
	}
	treasure := effectstest.BattlefieldTokenNamed(t, g, "Treasure")
	if err := g.SacrificePermanent(me.ID, treasure.InstanceID); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, air, "flying") {
		t.Error("sacrificing a Treasure should give the Aerialist flying")
	}
}

func TestOldB033AerialistIgnoresANontokenSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	air := effectstest.PushCatalogPermanent(g, me.ID, "Burdened Aerialist", "Creature — Human Pirate", oldB033Aerialist, false)
	rock := effectstest.PushArtifactToBattlefieldForTest(g, me.ID, "Rock", "Artifact", "")
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatalf("SacrificePermanent: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, air, "flying") {
		t.Error("a nontoken sacrifice gave flying")
	}
}

// --- Magus of the Bridge --------------------------------------------------------------

func TestOldB033MagusMakesAZombieForYourNontokenCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Bridge", "Creature — Human Wizard", oldB033Magus, false)
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mine); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Count(g, "Zombie", me.ID) != 1 {
		t.Fatalf("Zombies: %d, want 1", oldB033Count(g, "Zombie", me.ID))
	}
	// A token of yours dying makes nothing.
	tok := effectstest.BattlefieldTokenNamed(t, g, "Zombie")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, tok.InstanceID); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Count(g, "Zombie", me.ID) != 0 {
		t.Error("a token dying made another Zombie")
	}
	// The Magus is untouched by your own deaths, and exiled by an opponent's.
	if oldB033Zone(g, magus) != game.ZoneBattlefield {
		t.Fatalf("the Magus is in %q", oldB033Zone(g, magus))
	}
	theirs := effectstest.PushDiesCreatureForTest(g, opp.ID, "Theirs", "", "Creature — Bear", 2, 2)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Zone(g, magus) != game.ZoneExile {
		t.Errorf("the Magus is in %q, want exile", oldB033Zone(g, magus))
	}
	if oldB033Count(g, "Zombie", me.ID) != 0 {
		t.Error("an opponent's creature made a Zombie")
	}
}

// --- Underworld Hermit -----------------------------------------------------------------

func TestOldB033HermitCountsDevotionToBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		g.Battlefield.Cards = append(g.Battlefield.Cards, game.Card{
			InstanceID: uuid.New(), Name: "Two Black", TypeLine: "Creature — Bear", ManaCost: "{1}{B}{B}",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.CastCatalogSpell(t, g, "Underworld Hermit", "Creature — Human Peasant", oldB033Hermit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The test copy of the Hermit has no mana cost, so only the other
	// permanent counts: two.
	if n := oldB033Count(g, "Squirrel", me.ID); n != 2 {
		t.Errorf("Squirrels: %d, want 2", n)
	}
}

// --- Goblin Traprunner ---------------------------------------------------------------------

func TestOldB033TraprunnerFlipsThreeCoinsAndMakesAttackingGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	run := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Traprunner", "Creature — Goblin", oldB033Traprunner, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, run)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if ask.CoinCount != 3 {
		t.Fatalf("the call is for %d coins, want 3", ask.CoinCount)
	}
	before := len(effectstest.RandomEvents(g, game.EventFlipCoin))
	if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	evs := effectstest.RandomEvents(g, game.EventFlipCoin)[before:]
	if len(evs) != 3 {
		t.Fatalf("%d coins flipped, want 3", len(evs))
	}
	wins := 0
	for _, e := range evs {
		if e.Won {
			wins++
		}
	}
	// Each win's Goblin asks what it attacks; take the defaults.
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	}
	if n := oldB033Count(g, "Goblin", me.ID); n != wins {
		t.Errorf("Goblins: %d, want %d (one per win)", n, wins)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Goblin" && effects.IsToken(c) && (!c.Tapped || c.AttackingTarget == uuid.Nil) {
				t.Errorf("a Goblin token is not tapped and attacking: %+v", c)
			}
		}
	})
}

// --- Specimen Collector ------------------------------------------------------------------------

func TestOldB033CollectorEntersWithSquirrelAndCrabThenCopiesATokenWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	col := effectstest.CastCatalogSpell(t, g, "Specimen Collector", "Creature — Vedalken Wizard", oldB033Collector, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB033Count(g, "Squirrel", me.ID) != 1 || oldB033Count(g, "Crab", me.ID) != 1 {
		t.Fatalf("Squirrels %d, Crabs %d, want one each", oldB033Count(g, "Squirrel", me.ID), oldB033Count(g, "Crab", me.ID))
	}
	crab := effectstest.BattlefieldTokenNamed(t, g, "Crab")
	if crab.Power != 0 || crab.Toughness != 3 {
		t.Errorf("Crab is %d/%d, want 0/3", crab.Power, crab.Toughness)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, col); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if effectstest.PassUntilPickTarget(t, g, me.ID) == nil {
		t.Fatal("the death trigger asked for no target")
	}
	effectstest.PickCard(t, g, me.ID, crab.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB033Count(g, "Crab", me.ID); n != 2 {
		t.Errorf("Crabs: %d, want 2 (the copy)", n)
	}
}
