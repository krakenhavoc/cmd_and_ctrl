package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ur_sphinx_test.go — ADR 0140: eminence, the first command-zone static,
// proved on The Ur-Sphinx (#2797). "As long as The Ur-Sphinx is in the command
// zone or on the battlefield, other Sphinx spells you cast cost {1} less."

const urSphinxOracle = "4a3fdb8e-4699-4bd9-84e6-3cc7fea0e1ef"

// urSphinxCard is The Ur-Sphinx as a card object owned by `p`.
func urSphinxCard(p *game.Player) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "The Ur-Sphinx", TypeLine: "Legendary Creature — Sphinx Avatar",
		ManaCost: "{6}{W}{U}{B}", OracleID: urSphinxOracle, IsCommander: true,
		Owner: p.ID, Controller: p.ID,
	}
}

// priceOf is priceInHand for a spell cast from `zone`, with a caller-chosen
// object (so The Ur-Sphinx can price itself).
func priceOf(t *testing.T, g *game.Game, seat *game.Player, card game.Card, zone game.ZoneKind) int {
	t.Helper()
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", card.ManaCost, err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: seat.ID, FromZone: zone})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

func sphinxSpell(owner *game.Player) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Test Sphinx", TypeLine: "Creature — Sphinx",
		ManaCost: "{3}{U}{U}", Owner: owner.ID, Controller: owner.ID,
	}
}

func TestUrSphinxIsRegisteredWithAnEminenceModifier(t *testing.T) {
	spec, ok := Lookup(urSphinxOracle)
	if !ok {
		t.Fatal("The Ur-Sphinx is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("Completeness = %v, want Full", spec.Completeness)
	}
	mods := game.CostModifiersFor(urSphinxOracle)
	if len(mods) != 1 || !mods[0].Eminence {
		t.Fatalf("cost modifiers = %+v, want one Eminence modifier", mods)
	}
}

// The discount applies from the command zone, and only to OTHER Sphinx spells
// its owner casts.
func TestEminenceDiscountsSphinxSpellsFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ur := urSphinxCard(me)
	me.Command.PushTop(ur)

	if got := priceOf(t, g, me, sphinxSpell(me), game.ZoneHand); got != 4 {
		t.Errorf("a Sphinx spell with The Ur-Sphinx in my command zone costs %d, want 4", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{3}{G}{G}"); got != 5 {
		t.Errorf("a non-Sphinx spell costs %d, want it untouched at 5", got)
	}
	if got := priceOf(t, g, opp, sphinxSpell(opp), game.ZoneHand); got != 5 {
		t.Errorf("an opponent's Sphinx spell costs %d, want 5 — \"you cast\"", got)
	}
	if got := priceOf(t, g, me, ur, game.ZoneCommand); got != 9 {
		t.Errorf("The Ur-Sphinx itself costs %d from the command zone, want 9 — \"other\"", got)
	}
	// The reduction spends against generic mana only: {W}{U} has none to spend.
	colored := sphinxSpell(me)
	colored.ManaCost = "{W}{U}"
	if got := priceOf(t, g, me, colored, game.ZoneHand); got != 2 {
		t.Errorf("a {W}{U} Sphinx costs %d, want 2 — no generic mana to reduce", got)
	}
}

// Eminence is for the command zone and the battlefield and nowhere else.
func TestEminenceWorksOnlyInTheCommandZoneAndOnTheBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name string
		put  func(g *game.Game, p *game.Player, c game.Card)
		want int
	}{
		{"command zone", func(_ *game.Game, p *game.Player, c game.Card) { p.Command.PushTop(c) }, 4},
		{"battlefield", func(g *game.Game, p *game.Player, c game.Card) {
			pushCatalogPermanentWithID(g, c.InstanceID, p.ID, c.Name, c.TypeLine, urSphinxOracle, false)
		}, 4},
		{"hand", func(_ *game.Game, p *game.Player, c game.Card) { p.Hand.PushTop(c) }, 5},
		{"graveyard", func(_ *game.Game, p *game.Player, c game.Card) { p.Graveyard.PushTop(c) }, 5},
		{"library", func(_ *game.Game, p *game.Player, c game.Card) { p.Library.PushTop(c) }, 5},
		{"exile", func(g *game.Game, _ *game.Player, c game.Card) { g.Exile.PushTop(c) }, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			tc.put(g, me, urSphinxCard(me))
			if got := priceOf(t, g, me, sphinxSpell(me), game.ZoneHand); got != tc.want {
				t.Errorf("The Ur-Sphinx in the %s: a Sphinx spell costs %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

// Two eminences stack (each is a reduction of {1}), one in each zone.
func TestEminenceStacksAcrossZonesAndBelongsToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Command.PushTop(urSphinxCard(me))
	pushCatalogPermanent(g, me.ID, "Another Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	opp.Command.PushTop(urSphinxCard(opp))

	if got := priceOf(t, g, me, sphinxSpell(me), game.ZoneHand); got != 3 {
		t.Errorf("two Ur-Sphinxes of mine: a Sphinx costs %d, want 3", got)
	}
	if got := priceOf(t, g, opp, sphinxSpell(opp), game.ZoneHand); got != 4 {
		t.Errorf("the opponent's own command-zone Ur-Sphinx: their Sphinx costs %d, want 4 — only theirs counts for them", got)
	}
}

// A command-zone card's ORDINARY statics do nothing there (CR 113.6): only a
// modifier that says Eminence works from the zone.
func TestACommandZoneCardsOrdinaryCostModifierDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Command.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Sphere of Resistance", TypeLine: "Artifact",
		OracleID: sphereOfResistanceOracle, Owner: me.ID, Controller: me.ID,
	})
	if got := priceInHand(t, g, me, "Swords to Plowshares", "Instant", "{W}"); got != 1 {
		t.Errorf("a Sphere of Resistance in the command zone taxed a spell: price %d, want 1", got)
	}
}

// --- the attack trigger -------------------------------------------------

// The grant is real: the card milled from an OPPONENT's library is cast from
// that opponent's graveyard for no mana, and comes out under my control.
func TestUrSphinxFreeCastFromAnOpponentsGraveyardResolves(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ur := pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	prize := uuid.New()
	opp.Library.PushTop(game.Card{
		InstanceID: prize, Name: "Free Hulk", TypeLine: "Creature — Giant", ManaCost: "{5}{G}{G}",
		Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID,
	})

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ur, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for i := 0; i < 8; i++ {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			break
		}
		if hasID(pick.ChooseCards, prize) {
			answerChooseCards(t, g, me.ID, prize)
		} else {
			answerChooseCards(t, g, me.ID)
		}
	}
	if !opp.Graveyard.Contains(prize) {
		t.Fatal("the prize was not milled into the opponent's graveyard")
	}
	if err := g.CastSpell(me.ID, prize, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("casting the milled card for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, prize)
	if !ok {
		t.Fatal("the cast card did not reach the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("the card is controlled by %s, want me — I cast it", c.Controller)
	}
}

// "Whenever one or more Sphinxes you control attack, each player mills that
// many cards. For each player, you may cast a card that player milled this way
// without paying its mana cost."
func TestUrSphinxAttackMillsThatManyAndOffersAFreeCastFromEachPile(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	other := pushVanillaCreature(g, me.ID, "Other Sphinx", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == other {
				g.Battlefield.Cards[i].TypeLine = "Creature — Sphinx"
			}
		}
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	// The opponent's pile is two lands: nothing there can be cast.
	for i := 0; i < 2; i++ {
		opp.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID,
		})
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Graveyard.Size()
	}

	var ursphinx uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == urSphinxOracle {
			ursphinx = c.InstanceID
		}
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{ursphinx, other, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if got := p.Graveyard.Size() - before[p.ID]; got != 2 {
			t.Errorf("%s milled %d cards, want 2 — two Sphinxes attacked (a Bear is not one)", p.Name, got)
		}
	}

	// One prompt per player whose pile has a castable card: three (not the
	// opponent with two lands), each from that player's own graveyard.
	var granted []uuid.UUID
	for i := 0; i < 8; i++ {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			break
		}
		if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
			t.Errorf("bounds %d..%d, want 0..1 — \"you MAY cast a card\"", pick.ChooseMin, pick.ChooseMax)
		}
		for _, id := range pick.ChooseCards {
			if c, ok := g.LookupCardForEffect(id); !ok || c.IsLand() {
				t.Errorf("a land (or a missing card) was offered as castable: %v", c.Name)
			}
		}
		first := pick.ChooseCards[0]
		answerChooseCards(t, g, me.ID, first)
		granted = append(granted, first)
	}
	if len(granted) != len(g.Seats)-1 {
		t.Fatalf("%d free-cast prompts, want %d (every pile but the lands)", len(granted), len(g.Seats)-1)
	}
	for _, id := range granted {
		perm := g.CastPermissionOnCardByIDForEffect(id)
		if perm == nil {
			t.Errorf("no cast permission was granted over %s", id)
			continue
		}
		if perm.Cost != "{0}" || perm.Player != me.ID {
			t.Errorf("permission = cost %q player %s, want {0} for me", perm.Cost, perm.Player)
		}
	}
}

// Declining every offer grants nothing, and a single Sphinx mills one card each.
func TestUrSphinxAttackMillsOnePerSphinxAndDeclineGrantsNothing(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ur := pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	before := opp.Graveyard.Size()

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ur, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - before; got != 1 {
		t.Errorf("one Sphinx attacked and the opponent milled %d, want 1", got)
	}
	var milled []uuid.UUID
	for i := 0; i < 8; i++ {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			break
		}
		milled = append(milled, pick.ChooseCards...)
		answerChooseCards(t, g, me.ID) // decline
	}
	if len(milled) == 0 {
		t.Fatal("no free-cast prompt was offered")
	}
	for _, id := range milled {
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
			t.Errorf("declining still granted a cast permission over %s", id)
		}
	}
}

// Sphinxes that attack are counted, not Sphinxes that exist: no attacking
// Sphinx, no trigger.
func TestUrSphinxDoesNotTriggerWhenNoSphinxAttacks(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "The Ur-Sphinx", "Legendary Creature — Sphinx Avatar", urSphinxOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := opp.Graveyard.Size()

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - before; got != 0 {
		t.Errorf("a Bear attacked alone and the opponent milled %d", got)
	}
}
