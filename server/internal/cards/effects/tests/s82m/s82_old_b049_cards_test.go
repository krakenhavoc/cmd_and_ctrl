package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b049_cards_test.go — Strixhaven: School of Mages, batch 49 (#5502).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	oldB49Effigy    = "37d48f9b-0286-4c5b-b549-baccb473aea3"
	oldB49Celebrant = "5f1d3d22-18f7-42b7-8039-ef7aa5d29856"
	oldB49MageDuel  = "763882f9-a7ad-4fb1-bdf4-c9a158d4b68e"
	oldB49Welcome   = "ac27ccb8-81e2-42e2-a340-0b0ce9609688"
	oldB49Tendrils  = "c0112586-3826-452e-9e44-7e6edd197f78"
	oldB49Outburst  = "ffc79d71-5a1f-40a8-80af-84c85e12ba6e"
	oldB49Redcap    = "b14bd895-1d4e-4bd5-9ac1-fed56afb6c90"
)

func TestOldB49AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB49Effigy:           "Fuming Effigy",
		oldB49Celebrant:        "Flamescroll Celebrant",
		oldB49Celebrant + "#1": "Revel in Silence",
		oldB49MageDuel:         "Mage Duel",
		oldB49Welcome:          "Explosive Welcome",
		oldB49Tendrils:         "Devouring Tendrils",
		oldB49Outburst:         "Creative Outburst",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB49Move(t *testing.T, g *game.Game, owner, card uuid.UUID, from, to game.ZoneKind) {
	t.Helper()
	toOwner := owner
	if to == game.ZoneExile {
		toOwner = uuid.Nil
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: from, Owner: owner}, game.ZoneRef{Kind: to, Owner: toOwner}, card); err != nil {
		t.Fatalf("move: %v", err)
	}
}

func TestOldB49FumingEffigyPingsEachOpponentWhenCardsLeaveYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Fuming Effigy", "Creature — Spirit", oldB49Effigy, false)
	before := effectstest.LifeOf(g, opp.ID)
	// A card entering the graveyard is not one leaving it.
	effectstest.SocB11Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before {
		t.Fatalf("opponent at %d after a card entered my graveyard, want %d", got, before)
	}
	mine := effectstest.SocB11Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	oldB49Move(t, g, me.ID, mine, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Fatalf("opponent at %d after a card left my graveyard, want %d", got, before-1)
	}
	// A card leaving an opponent's graveyard does not trigger it.
	theirs := effectstest.SocB11Graveyard(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	oldB49Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("opponent at %d after a card left their own graveyard, want %d", got, before-1)
	}
}

func TestOldB49FlamescrollCelebrantPingsAnOpponentWhoActivatesNonManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	var me *game.Player
	for _, s := range g.Seats {
		if s.ID != opp.ID {
			me = s
			break
		}
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Flamescroll Celebrant", "Creature — Human Shaman", oldB49Celebrant, false)
	redcap := effectstest.B12Push(g, opp.ID, "Weaselback Redcap", "Creature — Goblin Knight", oldB49Redcap, 1, 1)
	before := effectstest.LifeOf(g, opp.ID)
	if err := effectstest.HobB4Activate(t, g, opp.ID, redcap, 0, "{R}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-1 {
		t.Errorf("the activating opponent is at %d, want %d", got, before-1)
	}
}

func TestOldB49FlamescrollCelebrantDoesNotPingItsOwnControllerOrManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	celebrant := effectstest.PushCatalogPermanent(g, me.ID, "Flamescroll Celebrant", "Creature — Human Shaman", oldB49Celebrant, false)
	before := effectstest.LifeOf(g, me.ID)
	power := effectstest.EffectivePower(t, g, celebrant)
	if err := effectstest.HobB4Activate(t, g, me.ID, celebrant, 0, "{R}{C}", nil); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != before {
		t.Errorf("the controller is at %d after activating, want %d", got, before)
	}
	if p := effectstest.EffectivePower(t, g, celebrant); p != power+2 {
		t.Errorf("the Celebrant has power %d after {1}{R}, want %d", p, power+2)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, celebrant, 0, "", nil); err == nil {
		t.Error("activated the pump with no mana")
	}
}

func TestOldB49RevelInSilenceStopsOpponentsCastingAndLoyaltyAbilitiesThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	walker := effectstest.B12Push(g, g.Seats[1].ID, "Test Walker", "Legendary Planeswalker — Test", "", 0, 0)
	gate := func(seat uuid.UUID, ability game.ActivationAbility) error {
		c, ok := g.LookupCardForEffect(walker)
		if !ok {
			t.Fatal("no walker")
		}
		var err error
		g.ReadSnapshot(func() { err = g.ActivationGateLocked(seat, c, game.ZoneBattlefield, ability) })
		return err
	}
	if err := gate(g.Seats[1].ID, game.ActivationAbility{Label: "loyalty", Loyalty: true}); err != nil {
		t.Fatalf("before Revel in Silence the loyalty gate says %v", err)
	}
	id := effectstest.CastInPlace(t, g, me.ID, "Revel in Silence", oldB49Celebrant+"#1")
	effectstest.PassPriorityAroundTable(t, g)
	spell := effectstest.InstantCardFor(t, g)
	for _, opp := range g.Seats[1:] {
		if err := effectstest.CastGateFor(g, opp.ID, spell, game.ZoneHand); err == nil {
			t.Errorf("seat %s cast a spell under Revel in Silence", opp.ID)
		}
		if err := gate(opp.ID, game.ActivationAbility{Label: "loyalty", Loyalty: true}); !errors.Is(err, game.ErrCantActivate) {
			t.Errorf("seat %s: loyalty ability gate = %v, want ErrCantActivate", opp.ID, err)
		}
		// Other activated abilities are untouched.
		if err := gate(opp.ID, game.ActivationAbility{Label: "other"}); err != nil {
			t.Errorf("seat %s: a non-loyalty ability is refused: %v", opp.ID, err)
		}
	}
	if err := effectstest.CastGateFor(g, me.ID, spell, game.ZoneHand); err != nil {
		t.Errorf("the caster cannot cast spells: %v", err)
	}
	if err := gate(me.ID, game.ActivationAbility{Label: "loyalty", Loyalty: true}); err != nil {
		t.Errorf("the caster's own loyalty abilities are refused: %v", err)
	}
	if got := effectstest.ZoneOfCard(g, id); got != game.ZoneExile {
		t.Errorf("Revel in Silence is in %v, want exile", got)
	}
}

func TestOldB49MageDuelGivesPlusOnePlusTwoThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Mage Duel", "Sorcery", oldB49MageDuel, effectstest.CardTargets(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tough != 4 {
		t.Errorf("my creature is %d/%d, want 3/4", p, tough)
	}
	if g.FindCardZoneForEffect(mine) == nil || g.FindCardZoneForEffect(mine).Kind != game.ZoneBattlefield {
		t.Error("my creature did not survive the fight (the +1/+2 should come first)")
	}
	if z := effectstest.ZoneOfCard(g, theirs); z != game.ZoneGraveyard {
		t.Errorf("their creature is in %v, want the graveyard", z)
	}
	// A creature you control is not a legal second target.
	other := effectstest.B12Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mage Duel", "Sorcery", oldB49MageDuel, effectstest.CardTargets(mine, other)); err == nil {
		t.Error("Mage Duel accepted a creature you control as the fight target")
	}
}

func TestOldB49MageDuelCostsTwoLessAfterAnotherInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	duel := effectstest.SocB17Hand(me, "Mage Duel", "Sorcery", "{2}{G}", oldB49MageDuel)
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, duel, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(mine, theirs)}); err == nil {
		t.Fatal("cast Mage Duel for {G} with no other instant or sorcery cast this turn")
	}
	// Casting an instant first makes the cost {G}.
	effectstest.CastInPlace(t, g, me.ID, "Revel in Silence", oldB49Celebrant+"#1")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, duel, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(mine, theirs)}); err != nil {
		t.Fatalf("cast Mage Duel for {G} after another instant: %v", err)
	}
}

func TestOldB49ExplosiveWelcomeDealsFiveAndThreeAndAddsRRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	before := effectstest.LifeOf(g, opp.ID)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear}}
	effectstest.CastCatalogSpell(t, g, "Explosive Welcome", "Instant", oldB49Welcome, targets)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-5 {
		t.Errorf("opponent at %d, want %d", got, before-5)
	}
	if z := effectstest.ZoneOfCard(g, bear); z != game.ZoneGraveyard {
		t.Errorf("the 3-damage target is in %v, want the graveyard", z)
	}
	pool := effectstest.PoolColors(me)
	if len(pool) != 3 || pool[0] != "R" || pool[1] != "R" || pool[2] != "R" {
		t.Errorf("mana pool %v, want {R}{R}{R}", pool)
	}
}

func TestOldB49ExplosiveWelcomeNeedsTwoDifferentTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	same := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: opp.ID}}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Explosive Welcome", "Instant", oldB49Welcome, same); err == nil {
		t.Error("Explosive Welcome accepted one target twice")
	}
}

func TestOldB49DevouringTendrilsBitesThenGainsTwoWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	theirs := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 3, 3)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Devouring Tendrils", "Sorcery", oldB49Tendrils, effectstest.CardTargets(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, theirs); z != game.ZoneGraveyard {
		t.Fatalf("their creature is in %v, want the graveyard", z)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want %d", got, life+2)
	}
	if got := effectstest.EffectiveToughness(t, g, mine); got != 3 {
		t.Errorf("the biter took damage: toughness %d", got)
	}
}

func TestOldB49DevouringTendrilsGainsWhenItDiesLaterThisTurnAndNotAfterwards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Ogre", "Creature — Ogre", "", 4, 4)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Devouring Tendrils", "Sorcery", oldB49Tendrils, effectstest.CardTargets(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Fatalf("life %d after a 1-damage bite, want %d", got, life)
	}
	// It dies later this turn: the delayed trigger fires.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d after it died this turn, want %d", got, life+2)
	}
}

func TestOldB49DevouringTendrilsRefusesYourOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Devouring Tendrils", "Sorcery", oldB49Tendrils, effectstest.CardTargets(a, b)); err == nil {
		t.Error("Devouring Tendrils accepted a creature you control as the second target")
	}
}

func TestOldB49CreativeOutburstDealsFiveAndTakesOneOfTheTopFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	before := effectstest.LifeOf(g, opp.ID)
	size := len(me.Library.Cards)
	effectstest.CastCatalogSpell(t, g, "Creative Outburst", "Instant", oldB49Outburst,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-5 {
		t.Errorf("opponent at %d, want %d", got, before-5)
	}
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top five")
	}
	if len(pick.ChooseCards) != 5 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards (max %d), want the top five and one", len(pick.ChooseCards), pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == ids[5] {
			t.Error("the sixth card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card is not in hand")
	}
	if len(me.Library.Cards) != size-1 || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != ids[5] {
		t.Errorf("library has %d cards with %v on top; want %d with the sixth card on top", len(me.Library.Cards), me.Library.Cards[len(me.Library.Cards)-1].Name, size-1)
	}
}

func TestOldB49CreativeOutburstChannelsFromHandForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Creative Outburst", "Instant", oldB49Outburst)
	// {U/R}{U/R}: one blue and one red both pay it.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB17Names(g, me.ID, "Treasure")); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if !effectstest.InGraveyard(me, id) {
		t.Error("the discarded card is not in the graveyard")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated again from the graveyard")
	}
}
