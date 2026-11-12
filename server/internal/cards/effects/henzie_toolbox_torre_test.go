package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// henzie_toolbox_torre_test.go — Henzie "Toolbox" Torre and the granted
// blitz seam (#5667).

const (
	henzieOracle   = "5320f5a6-8066-47d4-aa0a-4c879f39d413"
	henzieTypeLine = "Legendary Creature — Devil Rogue"
	henzieBigName  = "Henzie Test Ogre"
)

func TestHenzieIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(henzieOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
	rows := map[string]bool{}
	for _, r := range game.AbilityRowsOf(game.Card{Name: spec.Name, OracleID: henzieOracle, TypeLine: henzieTypeLine}) {
		rows[r.Label] = true
	}
	for _, want := range []string{
		"Each creature spell you cast with mana value 4 or greater has blitz. The blitz cost is equal to its mana cost.",
		"Blitz costs you pay cost {1} less for each time you've cast your commander from the command zone this game.",
	} {
		if !rows[want] {
			t.Errorf("no ability row %q in %v", want, rows)
		}
	}
}

// henzieOnBattlefield walks the active seat to its main phase with a
// Henzie it controls.
func henzieOnBattlefield(t *testing.T, g *game.Game) (me *game.Player, henzie uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	henzie = pushCatalogPermanent(g, me.ID, "Henzie \"Toolbox\" Torre", henzieTypeLine, henzieOracle, false)
	return me, henzie
}

// henzieCardInHand puts a 4/4 with this type line and mana cost in `p`'s
// hand.
func henzieCardInHand(g *game.Game, p *game.Player, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: henzieBigName, TypeLine: typeLine, ManaCost: manaCost,
			Power: 4, Toughness: 4, Colors: []string{"G"}, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// henzieSettle passes priority until the stack is empty, answering a
// trigger-order prompt in the order offered.
func henzieSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}

// henzieCommanderCasts records that `p` has cast a commander from the
// command zone `n` times this game.
func henzieCommanderCasts(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		if p.CommanderCasts == nil {
			p.CommanderCasts = map[uuid.UUID]int{}
		}
		p.CommanderCasts[uuid.New()] = n
	})
}

// henzieCard reads a card wherever it is.
func henzieCard(g *game.Game, id uuid.UUID) game.Card {
	var card game.Card
	g.WithWriteLock(func() { card, _ = g.LookupCardForEffect(id) })
	return card
}

var henzieBlitz = game.CastSpellParams{AlternativeCost: GrantedAltCostBlitz, Strict: true}

// henzieBlitzed asserts the three things blitz gives the permanent, then
// walks to the end step and asserts the sacrifice and the draw.
func henzieBlitzed(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if !g.Battlefield.Contains(id) {
		t.Fatal("the blitzed creature did not enter")
	}
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("the blitzed creature has no haste")
	}
	if n := grantedTriggerCount(g, id); n != 1 {
		t.Errorf("%d granted triggers, want the one \"When this creature dies, draw a card.\"", n)
	}
	hand := me.Hand.Size()
	endTheTurn(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the blitzed creature was not sacrificed at the beginning of the end step")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d: one card for the creature dying", got, hand+1)
	}
}

// The whole first ability: a creature with mana value 4 is cast for a
// blitz cost equal to its mana cost, has haste and the draw, and is
// sacrificed at the next end step.
func TestHenzieGivesABigCreatureBlitzAtItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")

	// The blitz cost is the mana cost: {2}{G} does not pay it.
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, henzieBlitz); err == nil {
		t.Fatal("blitzed a {3}{G} creature with three mana")
	}
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, id, henzieBlitz); err != nil {
		t.Fatalf("cast for the granted blitz: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want the whole {3}{G} spent", me.ManaPool.Total())
	}
	henzieSettle(t, g)
	henzieBlitzed(t, g, me, id)
}

// Cast for its mana cost, the same creature is an ordinary creature.
func TestHenzieCreatureCastForItsManaCostStays(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	henzieSettle(t, g)
	if hasEffectiveKeyword(t, g, id, "haste") || grantedTriggerCount(g, id) != 0 {
		t.Error("a creature cast for its mana cost got blitz's haste or draw")
	}
	endTheTurn(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("a creature cast for its mana cost was sacrificed")
	}
}

// "Each creature spell you cast with mana value 4 or greater": not a
// creature of mana value 3, not a noncreature spell, not an opponent's
// creature.
func TestHenzieBlitzReachesOnlyYourCreatureSpellsOfManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	offers := func(p *game.Player, id uuid.UUID) bool {
		found := false
		g.WithWriteLock(func() {
			card, _ := g.LookupCardForEffect(id)
			for _, ac := range g.CastOffersForLocked(p.ID, card, game.ZoneHand, nil) {
				if ac != nil && ac.Key == GrantedAltCostBlitz {
					found = true
				}
			}
		})
		return found
	}
	small := henzieCardInHand(g, me, "Creature — Ogre", "{2}{G}")
	artifact := henzieCardInHand(g, me, "Artifact", "{5}")
	// {X}{G}{G}: mana value 2 until a value is announced for X.
	hydra := henzieCardInHand(g, me, "Creature — Hydra", "{X}{G}{G}")
	big := henzieCardInHand(g, me, "Creature — Ogre", "{2}{G}{G}")
	if offers(me, small) || offers(me, artifact) || offers(me, hydra) {
		t.Errorf("granted blitz offered to a three-drop (%v), a noncreature (%v) or an {X}{G}{G} creature (%v)",
			offers(me, small), offers(me, artifact), offers(me, hydra))
	}
	if !offers(me, big) {
		t.Fatal("no granted blitz for a four-drop creature")
	}
	floatMana(t, g, me, "{G}{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, small, henzieBlitz); err == nil {
		t.Error("blitzed a creature of mana value 3")
	}
	if err := g.CastSpell(me.ID, artifact, henzieBlitz); err == nil {
		t.Error("blitzed a noncreature spell")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := henzieCardInHand(g, opp, "Creature — Ogre", "{3}{G}")
	if offers(opp, theirs) {
		t.Error("an opponent's creature spell was offered Henzie's blitz")
	}
}

// "Blitz costs you pay cost {1} less for each time you've cast your
// commander from the command zone this game": generic mana only, off the
// granted cost, and never off the same spell cast for its mana cost.
func TestHenzieReducesBlitzCostsByCommanderCasts(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	henzieCommanderCasts(g, me, 2)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")

	card := henzieCard(g, id)
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{AlternativeCost: GrantedAltCostBlitz})
	if err != nil || price.Total.String() != "{1}{G}" {
		t.Fatalf("blitz price = %q (%v), want {1}{G}", price.Total.String(), err)
	}
	price, err = g.PriceCast(me.ID, card, game.CastSpellParams{})
	if err != nil || price.Total.String() != "{3}{G}" {
		t.Fatalf("mana-cost price = %q (%v), want {3}{G}: the reduction is of blitz costs", price.Total.String(), err)
	}

	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a {3}{G} creature for its mana cost with two mana")
	}
	if err := g.CastSpell(me.ID, id, henzieBlitz); err != nil {
		t.Fatalf("blitz for {1}{G}: %v", err)
	}
	henzieSettle(t, g)
	henzieBlitzed(t, g, me, id)
}

// The reduction stops at the coloured mana (CR 118.7a).
func TestHenzieReductionNeverTouchesColouredMana(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	henzieCommanderCasts(g, me, 7)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{2}{G}{G}")
	card := henzieCard(g, id)
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{AlternativeCost: GrantedAltCostBlitz})
	if err != nil || price.Total.String() != "{G}{G}" {
		t.Fatalf("blitz price = %q (%v), want {G}{G}", price.Total.String(), err)
	}
}

// A printed blitz cost is a blitz cost too: Riveteers Requisitioner's
// {2}{R} is {R} after two commander casts, though its mana value of 2
// gets no blitz from Henzie.
func TestHenzieReducesAPrintedBlitzCost(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	henzieCommanderCasts(g, me, 2)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Riveteers Requisitioner", TypeLine: "Creature — Lizard Rogue",
			ManaCost: "{1}{R}", OracleID: requisitionerOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, henzieBlitz); err == nil {
		t.Fatal("a two-drop was given blitz")
	}
	floatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "blitz", Strict: true}); err != nil {
		t.Fatalf("printed blitz for {R}: %v", err)
	}
}

// An opponent's Henzie reduces nothing I pay.
func TestHenzieReductionIsItsControllersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushCatalogPermanent(g, opp.ID, "Henzie \"Toolbox\" Torre", henzieTypeLine, henzieOracle, false)
	henzieCommanderCasts(g, me, 2)
	henzieCommanderCasts(g, opp, 2)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Riveteers Requisitioner", TypeLine: "Creature — Lizard Rogue",
			ManaCost: "{1}{R}", OracleID: requisitionerOracle, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	card := henzieCard(g, id)
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{AlternativeCost: "blitz"})
	if err != nil || price.Total.String() != "{2}{R}" {
		t.Fatalf("blitz price under an opponent's Henzie = %q (%v), want {2}{R}", price.Total.String(), err)
	}
}

// A commander of mana value 4 in the command zone has blitz too (CR
// 118.9a: its mana cost may be paid there). The tax is added to the blitz
// cost (CR 118.9d), the reduction counts the casts made BEFORE this one
// (CR 601.2f comes before the spell becomes cast), and the cast is then
// counted.
func TestHenzieBlitzesACommanderFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	cmd := uuid.New()
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: cmd, Name: "Henzie Test Commander", TypeLine: "Legendary Creature — Ogre",
			ManaCost: "{2}{R}{R}", Power: 4, Toughness: 4, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID, IsCommander: true})
		if me.CommanderCasts == nil {
			me.CommanderCasts = map[uuid.UUID]int{}
		}
		me.CommanderCasts[cmd] = 1
	})
	card := henzieCard(g, cmd)
	params := game.CastSpellParams{AlternativeCost: GrantedAltCostBlitz, FromZone: "command", Strict: true}
	// {2}{R}{R}, plus {2} tax for one earlier cast, less {1} for that cast.
	price, err := g.PriceCast(me.ID, card, params)
	if err != nil || price.Total.String() != "{3}{R}{R}" {
		t.Fatalf("blitz price from the command zone = %q (%v), want {3}{R}{R}", price.Total.String(), err)
	}
	floatMana(t, g, me, "{R}{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, cmd, params); err != nil {
		t.Fatalf("blitz the commander: %v", err)
	}
	if got := me.CommanderCasts[cmd]; got != 2 {
		t.Errorf("commander casts = %d, want 2", got)
	}
	henzieSettle(t, g)
	if !g.Battlefield.Contains(cmd) || !hasEffectiveKeyword(t, g, cmd, "haste") {
		t.Error("the blitzed commander did not enter with haste")
	}
}

// CR 702.152a hangs the riders on "this spell's blitz cost was paid", not
// on Henzie: the creature is blitzed though Henzie left before it
// resolved.
func TestHenzieBlitzedSpellKeepsBlitzAfterHenzieLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, henzie := henzieOnBattlefield(t, g)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, henzieBlitz); err != nil {
		t.Fatalf("cast for the granted blitz: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, henzie); err != nil {
		t.Fatalf("exile Henzie: %v", err)
	}
	henzieSettle(t, g)
	henzieBlitzed(t, g, me, id)
}

// A creature that prints blitz keeps its own cost beside Henzie's (CR
// 702.152b): both offers are made, at their own prices.
func TestHenzieBlitzBesideAPrintedBlitz(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Workshop Warchief", TypeLine: "Creature — Rhino Warrior",
			ManaCost: "{3}{G}{G}", OracleID: workshopWarchiefOracle, Power: 5, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	prices := map[string]string{}
	g.WithWriteLock(func() {
		card, _ := g.LookupCardForEffect(id)
		for _, ac := range g.CastOffersForLocked(me.ID, card, game.ZoneHand, nil) {
			if ac != nil {
				prices[ac.Key] = ac.ManaCost
			}
		}
	})
	if prices["blitz"] != "{4}{G}{G}" || prices[GrantedAltCostBlitz] != "{3}{G}{G}" {
		t.Fatalf("offers = %v, want the printed blitz {4}{G}{G} and the granted one {3}{G}{G}", prices)
	}
}

// CR 707.10: a copy of a spell cast for the granted blitz cost copied
// that cost, so the token it becomes has haste and the draw and is
// sacrificed too.
func TestACopyOfASpellCastForGrantedBlitzIsBlitzedToo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, henzieBlitz); err != nil {
		t.Fatalf("cast for the granted blitz: %v", err)
	}
	copyTheSpell(t, g, me, id)
	henzieSettle(t, g)
	toks, _ := copiesOnBattlefield(g, henzieBigName)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	tok := toks[0].InstanceID
	if !hasEffectiveKeyword(t, g, tok, "haste") || grantedTriggerCount(g, tok) != 1 {
		t.Error("the copy's token lacks blitz's haste or draw")
	}
	hand := me.Hand.Size()
	endTheTurn(t, g)
	if g.Battlefield.Contains(tok) || !me.Graveyard.Contains(id) {
		t.Error("the token and the original are both sacrificed at the end step")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand is %d, want %d: one card each for the original and the token", got, hand+2)
	}
}

// The claim is carried on the stack item, so a game restored with the
// spell on the stack still blitzes it; restored again with the creature
// on the battlefield, the sacrifice and the draw still happen.
func TestGrantedBlitzSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := henzieOnBattlefield(t, g)
	seat := g.Turn.ActiveSeat
	id := henzieCardInHand(g, me, "Creature — Ogre", "{3}{G}")
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, id, henzieBlitz); err != nil {
		t.Fatalf("cast for the granted blitz: %v", err)
	}

	onStack := restoreThroughJSON(t, g)
	henzieSettle(t, onStack)
	if !onStack.Battlefield.Contains(id) || !hasEffectiveKeyword(t, onStack, id, "haste") {
		t.Fatal("restored with the spell on the stack: the creature did not enter with haste")
	}

	entered := restoreThroughJSON(t, onStack)
	henzieBlitzed(t, entered, entered.Seats[seat], id)
}
