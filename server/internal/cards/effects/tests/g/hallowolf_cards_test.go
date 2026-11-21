package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hallowolf_cards_test.go — the eight Hallowolf (#4071) cards that sit on
// existing seams: Full Moon's Rise, Cemetery Prowler, Nightpack Ambusher,
// Castle Embereth, Howlpack Resurgence, Moonlight Hunt, Master's Rebuke,
// and Hollowhenge Overlord.

const (
	hwFullMoonsRise      = "3aef097b-711a-4463-a59c-831e4dd5e69a"
	hwCemeteryProwler    = "fa53cc50-f06f-44b0-850f-cfa6e0caa4c5"
	hwNightpackAmbusher  = "4d67357c-5939-45c8-a656-aaf76728a5a6"
	hwCastleEmbereth     = "91fbb25b-8521-483f-88b0-77778d25f7fd"
	hwHowlpackResurgence = "614c3f71-e231-4c12-9e6b-cb8ea1d100bf"
	hwMoonlightHunt      = "f38a6cb4-79dd-4c12-9896-8c6655e275f8"
	hwMastersRebuke      = "e68ff836-6e51-4c3a-91e5-f0d649297245"
	hwHollowhengeOverlrd = "0e2418f8-2f32-4f79-8ce7-44eae8e764cb"
)

func TestHWHallowolfCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]effects.Completeness{
		hwFullMoonsRise:      effects.CompletenessFull,
		hwCemeteryProwler:    effects.CompletenessFull,
		hwNightpackAmbusher:  effects.CompletenessFull,
		hwCastleEmbereth:     effects.CompletenessFull,
		hwHowlpackResurgence: effects.CompletenessFull,
		hwMoonlightHunt:      effects.CompletenessFull,
		hwMastersRebuke:      effects.CompletenessFull,
		hwHollowhengeOverlrd: effects.CompletenessFull,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != want {
			t.Errorf("%s: completeness %v, want %v", spec.Name, spec.Completeness, want)
		}
	}
}

func TestHWFullMoonsRisePumpsAndTramplesWerewolvesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rise := effectstest.RfcbPush(g, me.ID, "Full Moon's Rise", "Enchantment", hwFullMoonsRise, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Human Werewolf", 2, 2)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Werewolf", 2, 2)

	c := effectstest.LayeredCard(t, g, mine)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 2 || !effectstest.HwHasKeyword(c, "trample") {
		t.Errorf("my Werewolf: %d/%d trample=%v, want 3/2 trample", c.CurrentPower(), c.CurrentToughness(), effectstest.HwHasKeyword(c, "trample"))
	}
	if w := effectstest.LayeredCard(t, g, wolf); w.CurrentPower() != 2 || effectstest.HwHasKeyword(w, "trample") {
		t.Error("a Wolf that is not a Werewolf is untouched")
	}
	if o := effectstest.LayeredCard(t, g, theirs); o.CurrentPower() != 2 || effectstest.HwHasKeyword(o, "trample") {
		t.Error("an opponent's Werewolf is untouched")
	}

	// Sacrifice: a regeneration shield on each of my Werewolves only.
	if err := g.ActivateCatalogAbility(me.ID, rise, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, rise) != nil {
		t.Error("the enchantment is sacrificed")
	}
	if got := effectstest.RegenShieldsOn(t, g, mine); got != 1 {
		t.Errorf("my Werewolf has %d shields, want 1", got)
	}
	if got := effectstest.RegenShieldsOn(t, g, wolf); got != 0 {
		t.Errorf("a Wolf has %d shields, want 0", got)
	}
	if got := effectstest.RegenShieldsOn(t, g, theirs); got != 0 {
		t.Errorf("their Werewolf has %d shields, want 0", got)
	}
}

func TestHWHowlpackResurgenceBoostsWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Howlpack Resurgence", "Enchantment", hwHowlpackResurgence, nil)
	effectstest.PassPriorityAroundTable(t, g)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := effectstest.B12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{wolf, were} {
		c := effectstest.LayeredCard(t, g, id)
		if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !effectstest.HwHasKeyword(c, "trample") {
			t.Errorf("%s: %d/%d trample=%v, want 3/3 trample", c.Name, c.CurrentPower(), c.CurrentToughness(), effectstest.HwHasKeyword(c, "trample"))
		}
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2 || effectstest.HwHasKeyword(c, "trample") {
		t.Error("a Bear is untouched")
	}
}

func TestHWNightpackAmbusherAnthemSkipsItselfAndMakesAWolfOnAQuietTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	amb := effectstest.RfcbPush(g, me.ID, "Nightpack Ambusher", "Creature — Wolf", hwNightpackAmbusher, 4, 4)
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := effectstest.B12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if c := effectstest.LayeredCard(t, g, amb); c.CurrentPower() != 4 {
		t.Errorf("the Ambusher pumps only others: power %d", c.CurrentPower())
	}
	for _, id := range []uuid.UUID{wolf, were} {
		if c := effectstest.LayeredCard(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("%s: %d/%d, want 3/3", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2 {
		t.Error("a Bear is untouched")
	}

	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HwCount(g, me.ID, "Wolf"); got != 2 {
		t.Errorf("Wolves after a quiet turn = %d, want 2 (mine plus the token)", got)
	}
}

func TestHWNightpackAmbusherMakesNothingAfterYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.RfcbPush(g, me.ID, "Nightpack Ambusher", "Creature — Wolf", hwNightpackAmbusher, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HwCount(g, me.ID, "Wolf"); got != 0 {
		t.Errorf("a Wolf token appeared after a spell was cast: %d", got)
	}
}

func TestHWHollowhengeOverlordMakesAWolfPerWolfOrWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.RfcbPush(g, me.ID, "Hollowhenge Overlord", "Creature — Wolf", hwHollowhengeOverlrd, 4, 4)
	effectstest.B12Creature(g, me.ID, "Pup", "Creature — Wolf", 1, 1)
	effectstest.B12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)

	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	if g.Turn.ActiveSeat != 0 {
		effectstest.AdvanceToUpkeepOf(t, g, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Wolf") {
			tokens++
		}
	}
	if tokens != 3 {
		t.Errorf("Wolf tokens = %d, want 3 (Overlord, Pup, Were)", tokens)
	}
}

func TestHWMastersRebukeBitesOneSided(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Master's Rebuke", "Instant", hwMastersRebuke,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage kills the Bear")
	}
	if effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Error("the Beast takes nothing back")
	}
}

func TestHWMoonlightHuntSumsTheWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.B12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 3, 3)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 9, 9)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 6)
	effectstest.CastCatalogSpell(t, g, "Moonlight Hunt", "Instant", hwMoonlightHunt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 5 {
		t.Errorf("damage on the target = %d, want 5 (2 + 3, the Bear is no Wolf)", got)
	}
	// A creature of my own is not a legal target.
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wolf", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Moonlight Hunt", "Instant", hwMoonlightHunt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != game.ErrIllegalTarget {
		t.Errorf("targeting my own creature: err %v, want ErrIllegalTarget", err)
	}
}

func TestHWCastleEmberethPumpsCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castle := effectstest.PushCatalogPermanent(g, me.ID, "Castle Embereth", "Land", hwCastleEmbereth, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.SamiGiveColorless(me, 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("my creature %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("their creature power %d, want 2", c.CurrentPower())
	}
}

func TestHWCemeteryProwlerDiscountsSpellsSharingTypesWithItsExiledCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	golem := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Old Golem", TypeLine: "Artifact Creature — Golem", Owner: opp.ID, Controller: opp.ID})

	// Before any exile nothing is discounted.
	prowler := effectstest.CastCatalogSpell(t, g, "Cemetery Prowler", "Creature — Wolf", hwCemeteryProwler, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger queued no pick")
	}
	if !effectstest.HasID(pick.ChooseCards, bolt) || !effectstest.HasID(pick.ChooseCards, golem) {
		t.Errorf("candidates %v, want both graveyards' cards", pick.ChooseCards)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Fatalf("before the exile the price is %d, want 4", got)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the picked card is exiled")
	}
	if got := effectstest.PriceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 2 {
		t.Errorf("an instant after exiling an instant: price %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("a creature shares no type yet: price %d, want 4", got)
	}

	_ = prowler
}

// An attack exiles again, and a type is counted once per spell however
// many exiled cards have it.
func TestHWCemeteryProwlerAttackExilesAndTypesCountOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Old Golem", TypeLine: "Artifact Creature — Golem", Owner: opp.ID, Controller: opp.ID})
	prowler := effectstest.PushDiesCreatureForTest(g, me.ID, "Cemetery Prowler", hwCemeteryProwler, "Creature — Wolf", 3, 4)
	effectstest.HwDeclare(t, g, opp.ID, prowler)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger queued no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{golem}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 3 {
		t.Errorf("creature spell: price %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Thopter", "Artifact Creature — Thopter", "{3}{U}"); got != 2 {
		t.Errorf("artifact creature spell: price %d, want 2 ({1} per shared type)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{2}{R}"); got != 3 {
		t.Errorf("instant shares no type: price %d, want 3", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{G}"); got != 1 {
		t.Errorf("a one-pip spell is never below its coloured mana: price %d, want 1", got)
	}
}
