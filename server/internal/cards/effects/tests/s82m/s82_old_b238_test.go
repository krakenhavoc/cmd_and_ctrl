package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB238Resilient = "eb23f999-e7f9-4082-b322-5b00ba64b458"
	oldB238Earth     = "df660a8e-39c2-455d-8860-522e2998a95e"
	oldB238Liliana   = "7c8b38b3-b808-4ad8-b4e5-39e216063ff6"
	oldB238Sunscourg = "f5e9c69f-b949-4e9a-91d8-e0d52f8d8d3e"
	oldB238Wasp      = "9331cfd7-a818-45ed-9e4a-4d3c381052e0"
	oldB238Gideon    = "d9d76570-d951-4c09-aa0e-c204c3407aaf"
	oldB238Dream     = "bf7201e6-1b3f-40f1-9b79-44921b00b0de"
	oldB238Graven    = "2590801e-0de0-4933-be2d-96a1d3cb4585"
	oldB238Striker   = "b4ed24b2-ba4f-492d-8b8d-44a5b21e4b33"
	oldB238Glory     = "d97fd43b-146f-4d13-9c45-51bceaff742e"
)

func TestOldB238CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB238Resilient: "Resilient Khenra", oldB238Earth: "Earthshaker Khenra", oldB238Liliana: "Liliana's Defeat",
		oldB238Sunscourg: "Sunscourge Champion", oldB238Wasp: "Wasp of the Bitter End", oldB238Gideon: "Gideon's Defeat",
		oldB238Dream: "Dreamstealer", oldB238Graven: "Graven Abomination", oldB238Striker: "Sinuous Striker",
		oldB238Glory: "Hour of Glory",
	}
	for oracle, name := range want {
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

// oldB238Creature puts a creature card with this catalog entry in hand and
// casts it, leaving its enters trigger for the caller.
func oldB238Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB238Graveyard(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	return id
}

func oldB238Token(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestOldB238SunscourgeChampionGainsItsPowerAndEternalizesWithADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	oldB238Cast(t, g, me, "Sunscourge Champion", "Creature — Human Wizard", oldB238Sunscourg, 2, 3)
	if me.Life != life+2 {
		t.Fatalf("life %d, want %d", me.Life, life+2)
	}
	gy := oldB238Graveyard(t, g, me, "Sunscourge Champion", "Creature — Human Wizard", "{2}{W}", oldB238Sunscourg, 2, 3)
	spare := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, gy, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("eternalize was accepted without discarding a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, gy, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("eternalize: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(gy) || !me.Graveyard.Contains(spare) {
		t.Error("the card was not exiled, or the discard did not happen")
	}
	tok, ok := oldB238Token(g, "Sunscourge Champion")
	if !ok {
		t.Fatal("no eternalized token")
	}
	if p, tough := effectstest.B38PT(t, g, tok.InstanceID); p != 4 || tough != 4 {
		t.Errorf("token %d/%d, want 4/4", p, tough)
	}
	if me.Life != life+2+4 {
		t.Errorf("life %d, want %d after the token gained 4", me.Life, life+6)
	}
}

func TestOldB238SinuousStrikerPumpsAndNeedsADiscardToEternalize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	st := effectstest.B12Push(g, me.ID, "Sinuous Striker", "Creature — Snake Warrior", oldB238Striker, 2, 2)
	effectstest.ApaMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, st, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.B38PT(t, g, st); p != 3 || tough != 1 {
		t.Errorf("%d/%d, want 3/1", p, tough)
	}
	gy := oldB238Graveyard(t, g, me, "Sinuous Striker", "Creature — Snake Warrior", "{2}{U}", oldB238Striker, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, gy, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("eternalize was accepted with no card to discard")
	}
	spare := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "C", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, gy, 1, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("eternalize: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB238Token(g, "Sinuous Striker"); !ok {
		t.Error("no eternalized token")
	}
}

func TestOldB238DreamstealerDiscardsThatManyCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d := effectstest.B12Push(g, me.ID, "Dreamstealer", "Creature — Human Wizard", oldB238Dream, 2, 2)
	effectstest.B64ReadyToAttack(g, d)
	for i := 0; i < 5; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	hand := opp.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, d, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
			break
		}
		effectstest.DiscardFromHand(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if opp.Hand.Size() != hand-2 {
		t.Errorf("hand %d, want %d after discarding two cards", opp.Hand.Size(), hand-2)
	}
}

func TestOldB238ResilientKhenraMayGiveItsPowerAsAPump(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Resilient Khenra", TypeLine: "Creature — Jackal Wizard", OracleID: oldB238Resilient, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, bear)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.B38PT(t, g, bear)
		want := 1
		if yes {
			want = 3
		}
		if p != want || tough != want {
			t.Errorf("yes=%v: %d/%d, want %d/%d", yes, p, tough, want, want)
		}
	}
}

func TestOldB238EarthshakerKhenraTargetsOnlyCreaturesNoStrongerThanItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	weak := effectstest.B12Creature(g, opp.ID, "Weak", "Creature — Bear", 2, 2)
	strong := effectstest.B12Creature(g, opp.ID, "Strong", "Creature — Ogre", 3, 3)
	oldB238Cast(t, g, me, "Earthshaker Khenra", "Creature — Jackal Warrior", oldB238Earth, 2, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, strong) {
		t.Error("a creature with greater power was offered")
	}
	if !effectstest.HasID(p.PickTargetCards, weak) {
		t.Error("an eligible creature was not offered")
	}
	effectstest.PickCard(t, g, me.ID, weak)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.SbCantBlock(t, g, weak) {
		t.Error("the target can still block")
	}
	if effectstest.SbCantBlock(t, g, strong) {
		t.Error("another creature can't block")
	}
}

func TestOldB238LilianasDefeatDestroysABlackPermanentAndPunishesALiliana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	black := effectstest.B12Push(g, opp.ID, "Black Bear", "Creature — Bear", "", 2, 2)
	green := effectstest.B12Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case black:
				g.Battlefield.Cards[i].Colors = []string{"B"}
			case green:
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Liliana, Test",
		TypeLine: "Legendary Planeswalker — Liliana", Colors: []string{"B"}, Counters: map[string]int{"loyalty": 4}, Owner: opp.ID, Controller: opp.ID})
	card := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Liliana's Defeat", TypeLine: "Sorcery", ManaCost: "{B}", OracleID: oldB238Liliana, Owner: me.ID, Controller: me.ID})
		return id
	}
	effectstest.ApaMana(me, "B")
	if err := g.CastSpell(me.ID, card(), game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(green)}); err == nil {
		t.Fatal("accepted a green creature")
	}
	if err := g.CastSpell(me.ID, card(), game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(black)}); err != nil {
		t.Fatalf("cast at the black creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(black) {
		t.Error("the black creature survived")
	}
	life := opp.Life
	effectstest.ApaMana(me, "B")
	if err := g.CastSpell(me.ID, card(), game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(walker)}); err != nil {
		t.Fatalf("cast at the planeswalker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the Liliana survived")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestOldB238GideonsDefeatExilesAnAttackingWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	white := effectstest.B12Creature(g, me.ID, "White Knight", "Creature — Knight", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Knight", "Creature — Knight", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == white || id == idle {
				g.Battlefield.Cards[i].Colors = []string{"W"}
			}
		}
	})
	effectstest.B64ReadyToAttack(g, white)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, white, opp.ID)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Gideon's Defeat", TypeLine: "Instant", ManaCost: "{W}", OracleID: oldB238Gideon, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(opp, "W")
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("accepted a white creature that is not attacking or blocking")
	}
	life := opp.Life
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(white)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || !g.Exile.Contains(white) {
		t.Error("the attacking white creature was not exiled")
	}
	if opp.Life != life {
		t.Errorf("life changed to %d for a non-Gideon", opp.Life)
	}
}

func TestOldB238WaspSacrificesToDestroyOnABolasWalkerCast(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		wasp := effectstest.B12Push(g, me.ID, "Wasp of the Bitter End", "Creature — Insect Horror", oldB238Wasp, 2, 1)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		// A non-Bolas spell does nothing.
		effectstest.CastCatalogSpell(t, g, "Plain Spell", "Sorcery", "", nil)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("a non-Bolas spell triggered the Wasp")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.CastCatalogSpell(t, g, "Nicol Bolas, Test", "Legendary Planeswalker — Bolas", "", nil)
		effectstest.PickCard(t, g, me.ID, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			if g.Battlefield.Contains(wasp) || g.Battlefield.Contains(victim) {
				t.Error("yes: the Wasp or the victim survived")
			}
			continue
		}
		if !g.Battlefield.Contains(wasp) || !g.Battlefield.Contains(victim) {
			t.Error("no: something left the battlefield")
		}
	}
}

func TestOldB238GravenAbominationExilesFromTheDefendersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ab := effectstest.B12Push(g, me.ID, "Graven Abomination", "Artifact Creature — Horror", oldB238Graven, 3, 1)
	effectstest.B64ReadyToAttack(g, ab)
	theirs := oldB238Graveyard(t, g, opp, "Their Card", "Sorcery", "{1}", "", 0, 0)
	mine := oldB238Graveyard(t, g, me, "My Card", "Sorcery", "{1}", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, ab, opp.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt on the attack")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("a card in your own graveyard was offered")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || !me.Graveyard.Contains(mine) {
		t.Error("the wrong card was exiled")
	}
}

func TestOldB238HourOfGloryExilesAndHitsTheHandOnlyForAGod(t *testing.T) {
	for _, god := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.AdvanceToMain(t, g)
		typeLine := "Creature — Bear"
		if god {
			typeLine = "Legendary Enchantment Creature — God"
		}
		victim := effectstest.B12Creature(g, opp.ID, "Test Deity", typeLine, 4, 4)
		copyA, copyB := uuid.New(), uuid.New()
		other := uuid.New()
		for _, id := range []uuid.UUID{copyA, copyB} {
			opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Deity", TypeLine: typeLine, Owner: opp.ID, Controller: opp.ID})
		}
		opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Other", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		effectstest.CastCatalogSpell(t, g, "Hour of Glory", "Instant", oldB238Glory, effectstest.CardRefs(victim))
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(victim) {
			t.Fatalf("god=%v: the creature was not exiled", god)
		}
		exiled := g.Exile.Contains(copyA) && g.Exile.Contains(copyB)
		if exiled != god {
			t.Errorf("god=%v: hand copies exiled = %v", god, exiled)
		}
		if !opp.Hand.Contains(other) {
			t.Errorf("god=%v: a card with another name left the hand", god)
		}
		_ = me
	}
}

func TestOldB238GideonsDefeatGainsFiveForAGideonPlaneswalkerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gid := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gideon, Test",
		TypeLine: "Legendary Planeswalker Creature — Gideon", Colors: []string{"W"}, Power: 4, Toughness: 4,
		Counters: map[string]int{"loyalty": 4}, Owner: me.ID, Controller: me.ID})
	effectstest.B64ReadyToAttack(g, gid)
	effectstest.AdvanceToMain(t, g)
	effectstest.LciB1Attack(t, g, gid, opp.ID)
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Gideon's Defeat", TypeLine: "Instant", ManaCost: "{W}", OracleID: oldB238Gideon, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(opp, "W")
	life := opp.Life
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(gid)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(gid) {
		t.Error("the Gideon was not exiled")
	}
	if opp.Life != life+5 {
		t.Errorf("life %d, want %d", opp.Life, life+5)
	}
}
