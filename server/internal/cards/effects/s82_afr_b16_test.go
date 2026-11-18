package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const truepolymorphOracle = "8d24e181-74c1-4913-a18a-59378ec7d9e7"

func TestS82AfrB16TruePolymorphMakesTheFirstTargetACopyForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := b12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 5, 5)
	castCatalogSpell(t, g, "True Polymorph", "Instant", truepolymorphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: dragon}})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if c := findBattlefieldCardByID(g, bear); c.Name != "Dragon" {
		t.Errorf("name %q, want Dragon", c.Name)
	}
	// Indefinite: still a Dragon after cleanup.
	dcCleanup(g)
	if p, _ := ptOf(t, g, bear); p != 5 {
		t.Errorf("power %d after cleanup, want 5 (no duration)", p)
	}
}

func TestS82AfrB16TruePolymorphNeedsTwoDifferentTargets(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	err := castCatalogSpellErrStrict(t, g, "True Polymorph", "Instant", truepolymorphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: bear}})
	if err == nil {
		t.Fatal("one permanent filled both targets")
	}
}

const voloOracle = "f105960e-6609-4192-afc0-e243013b7cc1"

func s82B16CastCreature(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

func TestS82AfrB16VoloCopiesACreatureSpellThatSharesNoType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Volo, Guide to Monsters", "Legendary Creature — Human Wizard", voloOracle, 3, 2)
	s82B16CastCreature(t, g, "Ogre", "Creature — Ogre")
	if n := onBattlefieldNamed(g, "Ogre"); n != 2 {
		t.Fatalf("%d Ogres on the battlefield, want 2 (the spell and its token copy)", n)
	}
}

func TestS82AfrB16VoloDoesNotCopyWhenATypeIsSharedWithAYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Volo, Guide to Monsters", "Legendary Creature — Human Wizard", voloOracle, 3, 2)
	s82B16CastCreature(t, g, "Wizard Apprentice", "Creature — Elf Wizard")
	if n := onBattlefieldNamed(g, "Wizard Apprentice"); n != 1 {
		t.Fatalf("%d on the battlefield, want 1 (shares Wizard with Volo)", n)
	}
}

func TestS82AfrB16VoloDoesNotCopyWhenATypeIsSharedWithAGraveyardCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Volo, Guide to Monsters", "Legendary Creature — Human Wizard", voloOracle, 3, 2)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Ogre", TypeLine: "Creature — Ogre", Owner: me.ID, Controller: me.ID})
	s82B16CastCreature(t, g, "Ogre", "Creature — Ogre")
	if n := onBattlefieldNamed(g, "Ogre"); n != 1 {
		t.Fatalf("%d Ogres, want 1 (an Ogre card is in the graveyard)", n)
	}
}

func TestS82AfrB16VoloIgnoresAnOpponentsCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Volo, Guide to Monsters", "Legendary Creature — Human Wizard", voloOracle, 3, 2)
	b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 2, 2)
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Dead Ogre", TypeLine: "Creature — Ogre", Owner: opp.ID, Controller: opp.ID})
	s82B16CastCreature(t, g, "Ogre", "Creature — Ogre")
	if n := onBattlefieldNamed(g, "Ogre"); n != 2 {
		t.Fatalf("%d Ogres, want 2 (only your creatures count)", n)
	}
}

const (
	ochreJellyOracle   = "2b757d32-09a3-49cd-ada4-84b5335178ce"
	wildMagicOracle    = "a450ef56-8511-43e0-bdca-0c25f54ba338"
	rodOfAbsorbOracle  = "d32345c1-ed20-4bfc-a5fa-6ac7b99542ec"
	phantomSteedOracle = "22c643b1-62dd-477c-8944-97b3becdfdbe"
	demilichOracle     = "eeb50bfc-8443-482b-bfc1-ea96b37809cd"
	xanatharOracle     = "1a168ba5-8e11-41dd-8f8a-8cb23bbde056"
)

func s82B16Jelly(t *testing.T, g *game.Game, counters int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Ochre Jelly", "Creature — Ooze", ochreJellyOracle, 0, 0)
	addCountersForTest(g, id, game.CounterPlusOne, counters)
	return id
}

func TestS82AfrB16OchreJellySplitsWithHalfTheCountersAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	id := s82B16Jelly(t, g, 5)
	killCreature(t, g, g.Seats[0].ID, id)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Ochre Jelly"); n != 0 {
		t.Fatalf("%d Jellies at once, want 0 (the token waits for the end step)", n)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Ochre Jelly"); n != 1 {
		t.Fatalf("%d Jellies after the end step began, want 1", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Ochre Jelly" {
			if !c.IsToken() {
				t.Error("the copy must be a token")
			}
			if got := c.Counters[game.CounterPlusOne]; got != 2 {
				t.Errorf("the token has %d +1/+1 counters, want 2 (half of 5, rounded down)", got)
			}
		}
	}
}

func TestS82AfrB16OchreJellyWithOneCounterLeavesNothing(t *testing.T) {
	g := newCatalogGame(t)
	id := s82B16Jelly(t, g, 1)
	killCreature(t, g, g.Seats[0].ID, id)
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Ochre Jelly"); n != 0 {
		t.Fatalf("%d Jellies, want 0 (it had fewer than two counters)", n)
	}
}

func s82B16CastFromExile(t *testing.T, g *game.Game, name string) {
	t.Helper()
	me := g.Seats[0]
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID})
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player: me.ID, CastOnly: true, Duration: game.WhileInZoneDuration(), Cost: "{0}",
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast %s from exile: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82AfrB16WildMagicSorcererGivesCascadeToTheFirstSpellFromExileOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedCheapLibrary(g, me, 6)
	pushCatalogPermanent(g, me.ID, "Wild-Magic Sorcerer", "Creature — Orc Shaman Sorcerer", wildMagicOracle, false)

	// A spell from the hand does not use the clock up and has no cascade.
	castWithCost(t, g, "Hand Spell", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Fatalf("a spell from hand got %d cascade offers, want 0", got)
	}
	s82B16CastFromExile(t, g, "First Exile Spell")
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("the first spell from exile got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	s82B16CastFromExile(t, g, "Second Exile Spell")
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("the second spell from exile got %d cascade offers, want 0", got)
	}
}

// --- Rod of Absorption ---------------------------------------------------------

func s82B16RodExiledWith(g *game.Game, me *game.Player, rod uuid.UUID, name, cost string) uuid.UUID {
	id := uuid.New()
	c, _ := battlefieldCard(g, rod)
	g.Exile.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: cost, Owner: me.ID, Controller: me.ID,
		ExiledWith: game.PermissionCardRef{ID: rod, Epoch: c.ObjectEpoch},
	})
	return id
}

func TestS82AfrB16RodOfAbsorptionExilesEveryInstantAndSorcerySpellInsteadOfTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rod := pushCatalogPermanent(g, me.ID, "Rod of Absorption", "Artifact", rodOfAbsorbOracle, false)
	spell := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)
	c := cmdr2S41InExile(g, spell)
	if c == nil {
		t.Fatal("the resolved instant was not exiled")
	}
	if c.ExiledWith.ID != rod {
		t.Errorf("the exiled spell is linked to %v, want the Rod", c.ExiledWith.ID)
	}
	if inGraveyard(me, spell) {
		t.Error("the spell is also in the graveyard")
	}
}

func TestS82AfrB16RodOfAbsorptionCastsExiledSpellsUpToTotalManaValueX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rod := pushCatalogPermanent(g, me.ID, "Rod of Absorption", "Artifact", rodOfAbsorbOracle, false)
	a := s82B16RodExiledWith(g, me, rod, "Spell A", "{1}{R}")
	b := s82B16RodExiledWith(g, me, rod, "Spell B", "{1}{R}")
	floatForTest(g, me, "CCC")
	b16Activate(t, g, me.ID, rod, 0, game.ActivateAbilityParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, rod); ok {
		t.Fatal("the Rod must be sacrificed as a cost")
	}
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the first spell (mana value 2 of X = 3): %v", err)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the second spell fit under the budget (total 4 > X = 3)")
	}
	left := 0
	for _, id := range []uuid.UUID{a, b} {
		if cmdr2S41InExile(g, id) != nil {
			left++
		}
	}
	if left != 1 {
		t.Errorf("%d of the two mana value 2 spells remain exiled, want exactly 1", left)
	}
}

func TestS82AfrB16RodOfAbsorptionWithTooSmallAnXCastsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rod := pushCatalogPermanent(g, me.ID, "Rod of Absorption", "Artifact", rodOfAbsorbOracle, false)
	a := s82B16RodExiledWith(g, me, rod, "Spell A", "{1}{R}")
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, rod, 0, game.ActivateAbilityParams{XValue: 1})
	passPriorityAroundTable(t, g)
	answerAllMayCast(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if cmdr2S41InExile(g, a) == nil {
		t.Error("a mana value 2 spell was cast with X = 1")
	}
}

// --- Phantom Steed -------------------------------------------------------------

func s82B16CastSteed(t *testing.T, g *game.Game, bear uuid.UUID) uuid.UUID {
	t.Helper()
	id := blbGLCastCreature(t, g, "Phantom Steed", "Creature — Horse Illusion", phantomSteedOracle, 4, 3, nil)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	return id
}

func TestS82AfrB16PhantomSteedExilesUntilItLeavesAndReturnsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	steed := s82B16CastSteed(t, g, bear)
	if _, ok := battlefieldCard(g, bear); ok {
		t.Fatal("the Bear should be exiled")
	}
	if !hasEffectiveKeyword(t, g, steed, "flash") {
		t.Error("the Steed has flash")
	}
	stdB9Kill(t, g, steed)
	if got := len(socB8Named(g, me.ID, "Bear")); got != 1 {
		t.Errorf("%d Bears, want the exiled one back", got)
	}
}

func TestS82AfrB16PhantomSteedAttackMakesATappedAttackingIllusionCopyThenSacrificesItAtEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	addCountersForTest(g, bear, game.CounterPlusOne, 2)
	steed := s82B16CastSteed(t, g, bear)
	unsick(g, steed)
	declareAttack(t, g, opp.ID, steed)
	passAndSettleAttackSplit(t, g, me.ID)
	copies := socB8Named(g, me.ID, "Bear")
	if len(copies) != 1 {
		t.Fatalf("%d Bear tokens, want 1", len(copies))
	}
	c := copies[0]
	if !c.IsToken() || !c.Tapped || c.Power != 2 || c.Toughness != 2 {
		t.Errorf("token %+v: want a tapped 2/2 token copy (no counters)", c)
	}
	if !c.HasSubtype("Illusion") || !c.HasSubtype("Bear") {
		t.Errorf("the token should be an Illusion in addition to its other types: %q", c.TypeLine)
	}
	var defender uuid.UUID
	g.ReadSnapshot(func() { defender = g.DefendingPlayerForAttackerForEffect(c.InstanceID) })
	if defender != opp.ID {
		t.Errorf("the token attacks %v, want the player the Steed attacks", defender)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := len(socB8Named(g, me.ID, "Bear")); got != 0 {
		t.Errorf("%d token copies remain after end of combat, want 0", got)
	}
}

// --- Demilich ------------------------------------------------------------------

func TestS82AfrB16DemilichCostsOneBlueLessForEachInstantAndSorceryCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	card := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Demilich", TypeLine: "Creature — Skeleton Wizard", ManaCost: "{U}{U}{U}{U}",
			OracleID: demilichOracle, Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID})
		return id
	}
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, card(), game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("Demilich cast for three blue with no spell cast first")
	}
	// Two instants/sorceries cast first make it {U}{U}.
	castCatalogSpell(t, g, "Spell One", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Spell Two", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, card(), game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("Demilich for {U}{U} after two instants and sorceries: %v", err)
	}
}

func TestS82AfrB16DemilichCastsFromTheGraveyardByExilingFourInstantsOrSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	demi := graveyardCardOf(g, me, "Demilich", "Creature — Skeleton Wizard", "{U}{U}{U}{U}", demilichOracle)
	var pay []uuid.UUID
	for i := 0; i < 4; i++ {
		typ := "Instant"
		if i%2 == 1 {
			typ = "Sorcery"
		}
		pay = append(pay, graveyardCardOf(g, me, "Spent", typ, "{R}", ""))
	}
	creature := graveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	payMana(t, g, me, "{U}{U}{U}{U}")
	cast := func(ids []uuid.UUID) error {
		return g.CastSpell(me.ID, demi, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: "exile_from_graveyard_in_addition", AltCostIDs: ids, Strict: true,
		})
	}
	// A creature card is not a legal payment, and three are too few.
	if err := cast(append(append([]uuid.UUID(nil), pay[:3]...), creature)); err == nil {
		t.Fatal("exiled a creature card to pay")
	}
	if err := cast(pay[:3]); err == nil {
		t.Fatal("cast with only three cards exiled")
	}
	if err := cast(pay); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	for _, id := range pay {
		if !g.Exile.Contains(id) {
			t.Errorf("paid card %v was not exiled", id)
		}
	}
	if !g.Stack.Contains(demi) {
		t.Error("Demilich did not reach the stack")
	}
}

func TestS82AfrB16DemilichAttackExilesAndCopiesAnInstantOrSorceryFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demi := b12Push(g, me.ID, "Demilich", "Creature — Skeleton Wizard", demilichOracle, 4, 3)
	spell := graveyardCardOf(g, me, "Old Bolt", "Instant", "{R}", "")
	declareAttack(t, g, opp.ID, demi)
	answerPickTargetIfOpen(t, g, spell)
	passPriorityAroundTable(t, g)
	if inGraveyard(me, spell) {
		t.Fatal("the card is still in the graveyard")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the original must be exiled (and stay there)")
	}
	cp := copyOf(t, g, "Old Bolt")
	if cp.InstanceID == spell || cp.Owner != me.ID {
		t.Errorf("copy = %+v: a new object owned by Demilich's controller", cp)
	}
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardForEffect(cp, game.ZoneExile)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("the controller has no permission over the copy: %+v", perm)
		}
		if perm.Cost == "{0}" {
			t.Errorf("the copy must be cast paying its mana cost, not for free")
		}
	})
	declineResolutionCasts(t, g, me.ID)
}

// --- Xanathar, Guild Kingpin ---------------------------------------------------

func TestS82AfrB16XanatharBansTheOpponentsCastsAndLetsYouPlayTheirTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	b12Push(g, me.ID, "Xanathar, Guild Kingpin", "Legendary Creature — Beholder", xanatharOracle, 5, 6)
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Their Top", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	advanceToUpkeepOf(t, g, 1)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	hand := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: hand, Name: "Their Spell", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	var cantCast *game.CantCastError
	if err := g.CastSpell(opp.ID, hand, game.CastSpellParams{Strict: true}); !errors.As(err, &cantCast) {
		t.Errorf("the chosen opponent cast a spell: err = %v, want a CantCastError", err)
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("playing the top card of their library: %v", err)
	}
	if !g.Stack.Contains(top) {
		t.Error("the card from their library did not reach the stack")
	}
}

func TestS82AfrB16XanatharLeavesOtherOpponentsFree(t *testing.T) {
	g := newCatalogGame(t)
	me, chosen, other := g.Seats[1], g.Seats[0], g.Seats[2]
	b12Push(g, me.ID, "Xanathar, Guild Kingpin", "Legendary Creature — Beholder", xanatharOracle, 5, 6)
	advanceToUpkeepOf(t, g, 1)
	answerPickTargetPlayer(t, g, chosen.ID)
	passPriorityAroundTable(t, g)
	id := castInPlace(t, g, other.ID, "Other Spell", "")
	if id == uuid.Nil {
		t.Error("an opponent that was not chosen could not cast")
	}
}
