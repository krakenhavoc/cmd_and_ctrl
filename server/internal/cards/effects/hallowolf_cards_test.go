package effects

import (
	"testing"

	"github.com/google/uuid"

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

func hwHasKeyword(c game.Card, kw string) bool {
	for _, a := range c.Effective().Abilities {
		if a == kw {
			return true
		}
	}
	return false
}

// hwDeclare declares the creatures as one attack declaration (CR 508.1),
// so a trigger that reads the whole declaration sees all of it.
func hwDeclare(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	var decls []game.AttackDeclaration
	for _, id := range attackers {
		decls = append(decls, game.AttackDeclaration{Attacker: id, Target: defender})
	}
	if _, err := g.DeclareAttackers(decls); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
}

func hwCount(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

func TestHWHallowolfCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]Completeness{
		hwFullMoonsRise:      CompletenessFull,
		hwCemeteryProwler:    CompletenessFull,
		hwNightpackAmbusher:  CompletenessFull,
		hwCastleEmbereth:     CompletenessFull,
		hwHowlpackResurgence: CompletenessFull,
		hwMoonlightHunt:      CompletenessFull,
		hwMastersRebuke:      CompletenessFull,
		hwHollowhengeOverlrd: CompletenessFull,
	} {
		spec, ok := Lookup(oracle)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rise := rfcbPush(g, me.ID, "Full Moon's Rise", "Enchantment", hwFullMoonsRise, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Human Werewolf", 2, 2)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Werewolf", 2, 2)

	c := layeredCard(t, g, mine)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 2 || !hwHasKeyword(c, "trample") {
		t.Errorf("my Werewolf: %d/%d trample=%v, want 3/2 trample", c.CurrentPower(), c.CurrentToughness(), hwHasKeyword(c, "trample"))
	}
	if w := layeredCard(t, g, wolf); w.CurrentPower() != 2 || hwHasKeyword(w, "trample") {
		t.Error("a Wolf that is not a Werewolf is untouched")
	}
	if o := layeredCard(t, g, theirs); o.CurrentPower() != 2 || hwHasKeyword(o, "trample") {
		t.Error("an opponent's Werewolf is untouched")
	}

	// Sacrifice: a regeneration shield on each of my Werewolves only.
	if err := g.ActivateCatalogAbility(me.ID, rise, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, rise) != nil {
		t.Error("the enchantment is sacrificed")
	}
	if got := regenShieldsOn(t, g, mine); got != 1 {
		t.Errorf("my Werewolf has %d shields, want 1", got)
	}
	if got := regenShieldsOn(t, g, wolf); got != 0 {
		t.Errorf("a Wolf has %d shields, want 0", got)
	}
	if got := regenShieldsOn(t, g, theirs); got != 0 {
		t.Errorf("their Werewolf has %d shields, want 0", got)
	}
}

func TestHWHowlpackResurgenceBoostsWolvesAndWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Howlpack Resurgence", "Enchantment", hwHowlpackResurgence, nil)
	passPriorityAroundTable(t, g)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := b12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{wolf, were} {
		c := layeredCard(t, g, id)
		if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !hwHasKeyword(c, "trample") {
			t.Errorf("%s: %d/%d trample=%v, want 3/3 trample", c.Name, c.CurrentPower(), c.CurrentToughness(), hwHasKeyword(c, "trample"))
		}
	}
	if c := layeredCard(t, g, bear); c.CurrentPower() != 2 || hwHasKeyword(c, "trample") {
		t.Error("a Bear is untouched")
	}
}

func TestHWNightpackAmbusherAnthemSkipsItselfAndMakesAWolfOnAQuietTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	amb := rfcbPush(g, me.ID, "Nightpack Ambusher", "Creature — Wolf", hwNightpackAmbusher, 4, 4)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := b12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if c := layeredCard(t, g, amb); c.CurrentPower() != 4 {
		t.Errorf("the Ambusher pumps only others: power %d", c.CurrentPower())
	}
	for _, id := range []uuid.UUID{wolf, were} {
		if c := layeredCard(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("%s: %d/%d, want 3/3", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if c := layeredCard(t, g, bear); c.CurrentPower() != 2 {
		t.Error("a Bear is untouched")
	}

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := hwCount(g, me.ID, "Wolf"); got != 2 {
		t.Errorf("Wolves after a quiet turn = %d, want 2 (mine plus the token)", got)
	}
}

func TestHWNightpackAmbusherMakesNothingAfterYouCastASpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfcbPush(g, me.ID, "Nightpack Ambusher", "Creature — Wolf", hwNightpackAmbusher, 4, 4)
	castCatalogSpell(t, g, "Cantrip", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := hwCount(g, me.ID, "Wolf"); got != 0 {
		t.Errorf("a Wolf token appeared after a spell was cast: %d", got)
	}
}

func TestHWHollowhengeOverlordMakesAWolfPerWolfOrWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rfcbPush(g, me.ID, "Hollowhenge Overlord", "Creature — Wolf", hwHollowhengeOverlrd, 4, 4)
	b12Creature(g, me.ID, "Pup", "Creature — Wolf", 1, 1)
	b12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)

	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	if g.Turn.ActiveSeat != 0 {
		advanceToUpkeepOf(t, g, 0)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Beast", "Creature — Beast", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Master's Rebuke", "Instant", hwMastersRebuke,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("4 damage kills the Bear")
	}
	if damageMarkedOn(g, mine) != 0 {
		t.Error("the Beast takes nothing back")
	}
}

func TestHWMoonlightHuntSumsTheWolvesAndWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 3, 3)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 9, 9)
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 6)
	castCatalogSpell(t, g, "Moonlight Hunt", "Instant", hwMoonlightHunt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, victim); got != 5 {
		t.Errorf("damage on the target = %d, want 5 (2 + 3, the Bear is no Wolf)", got)
	}
	// A creature of my own is not a legal target.
	mine := b12Creature(g, me.ID, "Mine", "Creature — Wolf", 1, 1)
	if err := castCatalogSpellErr(t, g, "Moonlight Hunt", "Instant", hwMoonlightHunt,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != game.ErrIllegalTarget {
		t.Errorf("targeting my own creature: err %v, want ErrIllegalTarget", err)
	}
}

func TestHWCastleEmberethPumpsCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castle := pushCatalogPermanent(g, me.ID, "Castle Embereth", "Land", hwCastleEmbereth, false)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	samiGiveColorless(me, 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, castle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("my creature %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("their creature power %d, want 2", c.CurrentPower())
	}
}

func TestHWCemeteryProwlerDiscountsSpellsSharingTypesWithItsExiledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	golem := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Old Golem", TypeLine: "Artifact Creature — Golem", Owner: opp.ID, Controller: opp.ID})

	// Before any exile nothing is discounted.
	prowler := castCatalogSpell(t, g, "Cemetery Prowler", "Creature — Wolf", hwCemeteryProwler, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger queued no pick")
	}
	if !hasID(pick.ChooseCards, bolt) || !hasID(pick.ChooseCards, golem) {
		t.Errorf("candidates %v, want both graveyards' cards", pick.ChooseCards)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Fatalf("before the exile the price is %d, want 4", got)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, bolt) {
		t.Fatal("the picked card is exiled")
	}
	if got := priceInHand(t, g, me, "Shock", "Instant", "{2}{R}"); got != 2 {
		t.Errorf("an instant after exiling an instant: price %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("a creature shares no type yet: price %d, want 4", got)
	}

	_ = prowler
}

// An attack exiles again, and a type is counted once per spell however
// many exiled cards have it.
func TestHWCemeteryProwlerAttackExilesAndTypesCountOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: golem, Name: "Old Golem", TypeLine: "Artifact Creature — Golem", Owner: opp.ID, Controller: opp.ID})
	prowler := pushDiesCreatureForTest(g, me.ID, "Cemetery Prowler", hwCemeteryProwler, "Creature — Wolf", 3, 4)
	hwDeclare(t, g, opp.ID, prowler)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the attack trigger queued no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{golem}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 3 {
		t.Errorf("creature spell: price %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Thopter", "Artifact Creature — Thopter", "{3}{U}"); got != 2 {
		t.Errorf("artifact creature spell: price %d, want 2 ({1} per shared type)", got)
	}
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{2}{R}"); got != 3 {
		t.Errorf("instant shares no type: price %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{G}"); got != 1 {
		t.Errorf("a one-pip spell is never below its coloured mana: price %d, want 1", got)
	}
}
