package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 666: Invasion — enters and dies triggers, graveyard recursion,
// artifacts, counters, counterspells and bounce.

const (
	oldB666Tectonic    = "0e68ee36-9649-490a-8483-fc8bdc23f161"
	oldB666AlloyGolem  = "410b7faf-365c-4b73-a64d-bddbe49cfded"
	oldB666Moat        = "dafd319e-1d4a-4e08-9167-a82ad8670f6f"
	oldB666CryptAngel  = "665f3e52-c536-481b-8b29-fe68f08d1fdf"
	oldB666Agenda      = "72177c93-ed1f-47bd-99a6-ad229a292d46"
	oldB666RootingKavu = "e7438616-cc42-4a6d-a7e3-5856cb0782fa"
	oldB666Filter      = "46484242-28ed-43d2-b33a-1e5c1771acca"
	oldB666Stakes      = "efd73b8a-1caa-4fc8-855a-cad33ea54097"
	oldB666KavuMonarch = "98cc734d-16d5-4c3c-98ca-a45efad27127"
	oldB666Distortion  = "b0a16783-31c2-4b2a-835d-774dc385f99a"
	oldB666Unmaking    = "bd7ad427-11d9-40d1-b292-71f985711665"
	oldB666Sunscape    = "e3be8e19-cec1-43ae-9e22-d7cb374570bb"
)

func TestOldB666Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB666Tectonic: "Tectonic Instability", oldB666AlloyGolem: "Alloy Golem", oldB666Moat: "Teferi's Moat",
		oldB666CryptAngel: "Crypt Angel", oldB666Agenda: "Yawgmoth's Agenda", oldB666RootingKavu: "Rooting Kavu",
		oldB666Filter: "Urza's Filter", oldB666Stakes: "Juntu Stakes", oldB666KavuMonarch: "Kavu Monarch",
		oldB666Distortion: "Temporal Distortion", oldB666Unmaking: "Barrin's Unmaking", oldB666Sunscape: "Sunscape Master",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB666Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// Tectonic Instability: any land entering taps every land its controller
// controls, the new one included; another player's lands stay as they were.
func TestOldB666TectonicInstabilityTapsTheEnteringLandsControllersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tectonic Instability", "Enchantment", oldB666Tectonic, false)
	mine := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.FreeSpellPermanent(g, opp.ID, "Island", "Basic Land — Island")
	played := effectstest.B12PlayFromHand(t, g, "Wastes", "Basic Land", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB666Tapped(t, g, mine) || !oldB666Tapped(t, g, played) {
		t.Error("the active player's lands, the new one included, should be tapped")
	}
	if oldB666Tapped(t, g, theirs) {
		t.Error("another player's land is not tapped by it")
	}
	// A land entering under the opponent's control taps the opponent's lands.
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, Actor: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB666Tapped(t, g, theirs) {
		t.Error("the opponent's land should be tapped when their land enters")
	}
}

// Alloy Golem: cast from hand it asks for a colour and is then exactly that
// colour, and still an artifact creature.
func TestOldB666AlloyGolemIsTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Alloy Golem", "Artifact Creature — Golem", "{6}", oldB666AlloyGolem, nil)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	effectstest.PassPriorityAroundTable(t, g)
	var colors []string
	var typeLine string
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, id)
		colors, typeLine = c.EffectiveColors(), c.TypeLine
	})
	if len(colors) != 1 || colors[0] != "G" {
		t.Errorf("colours %v, want only green", colors)
	}
	if !strings.Contains(typeLine, "Artifact") {
		t.Errorf("type line %q, want it still an artifact", typeLine)
	}
}

// Teferi's Moat: creatures of the chosen colour without flying can't attack
// you; flyers and creatures of other colours can.
func TestOldB666TeferisMoat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Teferi's Moat", "Enchantment", oldB666Moat, "R")
	redGround := effectstest.B27Push(g, opp.ID, "Red Bear", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	greenGround := effectstest.B27Push(g, opp.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redFlyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bird", TypeLine: "Creature — Bird",
		ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(redGround, me.ID); err == nil || !strings.Contains(err.Error(), "Teferi's Moat") {
		t.Errorf("a red creature without flying: err = %v, want the Moat's refusal", err)
	}
	if err := g.DeclareAttacker(greenGround, me.ID); err != nil {
		t.Errorf("a green creature may attack: %v", err)
	}
	if err := g.DeclareAttacker(redFlyer, me.ID); err != nil {
		t.Errorf("a red flyer may attack: %v", err)
	}
}

// Crypt Angel: returns a blue or red creature card, and only that.
func TestOldB666CryptAngelReturnsABlueOrRedCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	blue := effectstest.GraveyardCardOf(g, me, "Blue Merfolk", "Creature — Merfolk", "{1}{U}", "")
	green := effectstest.GraveyardCardOf(g, me, "Green Bear", "Creature — Bear", "{1}{G}", "")
	redSpell := effectstest.GraveyardCardOf(g, me, "Red Bolt", "Instant", "{R}", "")
	angel := effectstest.CastHoldCreature(t, g, "Crypt Angel", "Creature — Angel", oldB666CryptAngel, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil &&
		(effectstest.HasID(p.PickTargetCards, green) || effectstest.HasID(p.PickTargetCards, redSpell)) {
		t.Error("a green creature card or a red noncreature card is offered")
	}
	effectstest.PickCard(t, g, me.ID, blue)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(blue) {
		t.Error("the blue creature card was not returned")
	}
	effectstest.AssertKeywords(t, g, angel, "flying", "protection from white")
}

// Yawgmoth's Agenda: one spell a turn, but lands and spells come out of the
// graveyard, and cards headed for your graveyard are exiled.
func TestOldB666YawgmothsAgenda(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Yawgmoth's Agenda", "Enchantment", oldB666Agenda, false)
	land := effectstest.GraveyardCardOf(g, me, "Forest", "Basic Land — Forest", "", "")
	bear := effectstest.GraveyardCardOf(g, me, "Grave Bear", "Creature — Bear", "{1}", "")
	second := effectstest.GraveyardCardOf(g, me, "Second Bear", "Creature — Bear", "{1}", "")
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("playing a land from the graveyard: %v", err)
	}
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("casting a spell from the graveyard: %v", err)
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); !errors.Is(err, game.ErrCantCast) {
		t.Errorf("a second spell in one turn: err = %v, want ErrCantCast", err)
	}
	// Cards that would go to your graveyard are exiled; an opponent's are not.
	mine := effectstest.B12Creature(g, me.ID, "Dier", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Dier", "Creature — Bear", 1, 1)
	effectstest.AvB18Destroy(t, g, mine)
	effectstest.AvB18Destroy(t, g, theirs)
	if me.Graveyard.Contains(mine) || !g.Exile.Contains(mine) {
		t.Error("your dying creature should be exiled instead")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature goes to their graveyard")
	}
}

// Rooting Kavu: dying, you may exile it and shuffle your creature cards from
// the graveyard into the library.
func TestOldB666RootingKavuShufflesCreatureCardsAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kavu := effectstest.B12Push(g, me.ID, "Rooting Kavu", "Creature — Kavu", oldB666RootingKavu, 4, 3)
	bear := effectstest.GraveyardCardOf(g, me, "Grave Bear", "Creature — Bear", "{1}{G}", "")
	bolt := effectstest.GraveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	effectstest.AvB18Destroy(t, g, kavu)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(kavu) {
		t.Error("the Kavu should have been exiled")
	}
	if me.Graveyard.Contains(bear) || !me.Library.Contains(bear) {
		t.Error("the creature card should be shuffled into the library")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("a noncreature card stays in the graveyard")
	}
}

func TestOldB666RootingKavuMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kavu := effectstest.B12Push(g, me.ID, "Rooting Kavu", "Creature — Kavu", oldB666RootingKavu, 4, 3)
	bear := effectstest.GraveyardCardOf(g, me, "Grave Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AvB18Destroy(t, g, kavu)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(kavu) || !me.Graveyard.Contains(bear) {
		t.Error("declining leaves the graveyard as it was")
	}
}

// Urza's Filter: multicolored spells cost {2} less, for every player; mono and
// colourless spells don't.
func TestOldB666UrzasFilter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Urza's Filter", "Artifact", oldB666Filter, false)
	effectstest.AdvanceToMain(t, g)
	gold := oldB666HandSpell(g, me, "Gold Bolt", "{2}{R}{W}", "R", "W")
	if err := oldB666CastWith(g, me, gold, "{R}{W}"); err != nil {
		t.Fatalf("a {2}{R}{W} spell with only {R}{W} under the Filter: %v", err)
	}
	mono := oldB666HandSpell(g, me, "Mono Bolt", "{2}{R}", "R")
	if err := oldB666CastWith(g, me, mono, "{R}"); err == nil {
		t.Error("a monocolored spell is not discounted")
	}
}

func oldB666HandSpell(g *game.Game, p *game.Player, name, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost,
			Colors: colors, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}

func oldB666CastWith(g *game.Game, p *game.Player, id uuid.UUID, mana string) error {
	g.WithWriteLock(func() { _ = g.AddManaForEffect(p.ID, uuid.Nil, mana) })
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

// Juntu Stakes: tapped creatures with power 1 or less stay tapped through their
// controller's untap step; bigger ones untap.
func TestOldB666JuntuStakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Juntu Stakes", "Artifact", oldB666Stakes, false)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 2, 2)
	effectstest.TapCard(g, small)
	effectstest.TapCard(g, big)
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if !oldB666Tapped(t, g, small) {
		t.Error("a creature with power 1 or less should not untap")
	}
	if oldB666Tapped(t, g, big) {
		t.Error("a creature with power 2 untaps normally")
	}
}

// Kavu Monarch: every Kavu has trample, and another Kavu entering, whoever's,
// grows the Monarch; a non-Kavu doesn't.
func TestOldB666KavuMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monarch := effectstest.B12Push(g, me.ID, "Kavu Monarch", "Creature — Kavu", oldB666KavuMonarch, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Kavu", "Creature — Kavu", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, monarch, "trample")
	effectstest.AssertKeywords(t, g, theirs, "trample")
	if kw := effectstest.EffectiveAbilities(t, g, bear); effectstest.ContainsString(kw, "trample") {
		t.Error("a non-Kavu has no trample")
	}
	for _, id := range []uuid.UUID{theirs, bear} {
		id := id
		g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: opp.ID}) })
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, monarch, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1 (one other Kavu entered)", n)
	}
}

// Temporal Distortion: a tapped creature or land gets an hourglass counter,
// doesn't untap, and loses the counter at its controller's next upkeep.
func TestOldB666TemporalDistortion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Temporal Distortion", "Enchantment", oldB666Distortion, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, effectstestHourglass); n != 1 {
		t.Fatalf("hourglass counters %d after attacking, want 1", n)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, effectstestHourglass); n != 1 {
		t.Errorf("another player's upkeep must not remove it: %d", n)
	}
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if !oldB666Tapped(t, g, bear) {
		t.Error("the creature with an hourglass counter should not untap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, effectstestHourglass); n != 0 {
		t.Errorf("the upkeep trigger should remove it: %d", n)
	}
	if !oldB666Tapped(t, g, bear) {
		t.Error("removing the counter doesn't untap")
	}
}

const effectstestHourglass = "hourglass"

func TestOldB666TemporalDistortionLandsToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Temporal Distortion", "Enchantment", oldB666Distortion, false)
	effectstest.ToMainForCost(t, g)
	swamp := effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	if err := g.ActivateManaAbility(me.ID, swamp, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, swamp, effectstestHourglass); n != 1 {
		t.Errorf("a land tapped for mana gets an hourglass counter: %d", n)
	}
}

// Barrin's Unmaking: bounces a permanent that shares a colour with the most
// common colour (or one tied for it); not one that doesn't, nor a colourless one.
func TestOldB666BarrinsUnmaking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Red is the most common: two red permanents; one blue; one colourless.
	effectstest.B27Push(g, opp.ID, "Red A", "Creature — Bear", "", "{R}", 1, 1, "R")
	redB := effectstest.B27Push(g, opp.ID, "Red B", "Creature — Bear", "", "{R}", 1, 1, "R")
	blue := effectstest.B27Push(g, opp.ID, "Blue", "Creature — Bear", "", "{U}", 1, 1, "U")
	rock := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	cast := func(target uuid.UUID) {
		t.Helper()
		effectstest.CastCatalogSpell(t, g, "Barrin's Unmaking", "Instant", oldB666Unmaking, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast(blue)
	if !g.Battlefield.Contains(blue) {
		t.Error("blue isn't the most common colour; the permanent stays")
	}
	cast(rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("a colourless permanent shares no colour")
	}
	cast(redB)
	if g.Battlefield.Contains(redB) || !opp.Hand.Contains(redB) {
		t.Error("a red permanent shares the most common colour and returns to its owner's hand")
	}
	_ = me
}

// Sunscape Master: {G}{G},{T} pumps your team; {U}{U},{T} bounces a creature.
func TestOldB666SunscapeMaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := effectstest.B12Push(g, me.ID, "Sunscape Master", "Creature — Human Wizard", oldB666Sunscape, 2, 2)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 1, 1)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, ally); p != 3 {
		t.Errorf("ally power %d, want 3", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, foe); p != 1 {
		t.Errorf("an opponent's creature isn't pumped: %d", p)
	}
	effectstest.Untap(g, master)
	effectstest.FloatMana(t, g, me, "{U}{U}")
	if err := effectstest.ActivateOn(t, g, me.ID, master, 1, foe); err != nil {
		t.Fatalf("the bounce: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !opp.Hand.Contains(foe) {
		t.Error("the targeted creature should be back in its owner's hand")
	}
}
