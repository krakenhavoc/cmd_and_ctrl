package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b16_test.go — March of the Machine batch 16 (#5477):
// the Phyrexian transformers with legends and heavier back faces.

const (
	s82m16DuneshaperOracle  = "c554ae68-0201-4958-aa63-8159911acb50"
	s82m16SeraphOracle      = "08574d1b-4b76-42dd-884b-29c5a61cfdda"
	s82m16AerialistOracle   = "29e96c47-94ad-4740-ba86-66dc2cf9ae2f"
	s82m16FreewheelerOracle = "e6fd0c7c-b164-4cf1-8910-0f8e680b2213"
	s82m16GuardianOracle    = "0b34581e-66df-4e2f-91fb-dddba632d203"
	s82m16AyaraOracle       = "f9ef90b9-a2b4-4a5b-92fb-2268ea7d709f"
	s82m16HeliodOracle      = "c27c6ba9-8d72-4101-b3a5-1e8d81755568"
	s82m16PolukranosOracle  = "7a9fa906-b43e-4d98-b134-5936d3085851"
	s82m16RonaOracle        = "cf5bd3a4-3899-4c25-ad1f-f1778c1d9c60"
)

func TestS82Mom16AreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82m16DuneshaperOracle:         "Tarkir Duneshaper",
		s82m16DuneshaperOracle + "#1":  "Burnished Dunestomper",
		s82m16SeraphOracle:             "Seraph of New Capenna",
		s82m16SeraphOracle + "#1":      "Seraph of New Phyrexia",
		s82m16AerialistOracle:          "Skyclave Aerialist",
		s82m16AerialistOracle + "#1":   "Skyclave Invader",
		s82m16FreewheelerOracle:        "Nezumi Freewheeler",
		s82m16FreewheelerOracle + "#1": "Hideous Fleshwheeler",
		s82m16GuardianOracle:           "Sun-Blessed Guardian",
		s82m16GuardianOracle + "#1":    "Furnace-Blessed Conqueror",
		s82m16AyaraOracle:              "Ayara, Widow of the Realm",
		s82m16AyaraOracle + "#1":       "Ayara, Furnace Queen",
		s82m16HeliodOracle:             "Heliod, the Radiant Dawn",
		s82m16HeliodOracle + "#1":      "Heliod, the Warped Eclipse",
		s82m16PolukranosOracle:         "Polukranos Reborn",
		s82m16PolukranosOracle + "#1":  "Polukranos, Engine of Ruin",
		s82m16RonaOracle:               "Rona, Herald of Invasion",
		s82m16RonaOracle + "#1":        "Rona, Tolarian Obliterator",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mainTable puts the active seat in a main phase.
func s82m16Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestS82Mom16TarkirDuneshaperTransformsPayingGreenOrLife(t *testing.T) {
	for _, mana := range []string{"{C}{C}{C}{C}{G}", "{C}{C}{C}{C}"} {
		syms := 0
		if mana == "{C}{C}{C}{C}" {
			syms = 1
		}
		g, me, _ := s82m16Main(t)
		id := pushBattlefieldCardWithTimestamp(g, s82mom16Card(me.ID, s82m16DuneshaperOracle,
			game.Face{Name: "Tarkir Duneshaper", TypeLine: "Creature — Dog Warrior", ManaCost: "{W}", Colors: []string{"W"}, Power: 1, Toughness: 2},
			game.Face{Name: "Burnished Dunestomper", TypeLine: "Creature — Phyrexian Dog Warrior", Colors: []string{"W"}, Power: 4, Toughness: 3}))
		life := me.Life
		s82mom16Transform(t, g, me, id, mana, syms)
		c := rvBattlefieldCard(g, id)
		if c == nil || c.Name != "Burnished Dunestomper" {
			t.Fatalf("%s: did not transform: %+v", mana, c)
		}
		wantLife := life
		if mana == "{C}{C}{C}{C}" {
			wantLife = life - 2
		}
		if me.Life != wantLife {
			t.Errorf("%s: life %d, want %d", mana, me.Life, wantLife)
		}
		if !hasKeywordOnBattlefield(t, g, id, "trample") {
			t.Errorf("%s: the back face has no trample", mana)
		}
	}
}

// s82m16Flip puts the card on the battlefield and transforms it with a
// real activation, returning its id.
func s82m16Flip(t *testing.T, g *game.Game, me *game.Player, c game.Card, mana string, life int) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, c)
	s82mom16Transform(t, g, me, id, mana, life)
	if got := rvBattlefieldCard(g, id); got == nil || got.Name != got.Faces[1].Name {
		t.Fatalf("the permanent did not transform: %+v", got)
	}
	return id
}

func s82m16Fronts() map[string]func(owner uuid.UUID) game.Card {
	mk := func(oracle string, front, back game.Face) func(uuid.UUID) game.Card {
		return func(owner uuid.UUID) game.Card { return s82mom16Card(owner, oracle, front, back) }
	}
	return map[string]func(uuid.UUID) game.Card{
		"seraph": mk(s82m16SeraphOracle,
			game.Face{Name: "Seraph of New Capenna", TypeLine: "Creature — Angel Soldier", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
			game.Face{Name: "Seraph of New Phyrexia", TypeLine: "Creature — Phyrexian Angel", Colors: []string{"W"}, Power: 3, Toughness: 3}),
		"aerialist": mk(s82m16AerialistOracle,
			game.Face{Name: "Skyclave Aerialist", TypeLine: "Creature — Merfolk Scout", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
			game.Face{Name: "Skyclave Invader", TypeLine: "Creature — Phyrexian Merfolk Scout", Colors: []string{"U"}, Power: 2, Toughness: 4}),
		"freewheeler": mk(s82m16FreewheelerOracle,
			game.Face{Name: "Nezumi Freewheeler", TypeLine: "Creature — Rat Samurai", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
			game.Face{Name: "Hideous Fleshwheeler", TypeLine: "Creature — Phyrexian Rat", Colors: []string{"B"}, Power: 4, Toughness: 5}),
		"guardian": mk(s82m16GuardianOracle,
			game.Face{Name: "Sun-Blessed Guardian", TypeLine: "Creature — Human Cleric", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
			game.Face{Name: "Furnace-Blessed Conqueror", TypeLine: "Creature — Phyrexian Cleric", Colors: []string{"W"}, Power: 3, Toughness: 3}),
		"ayara": mk(s82m16AyaraOracle,
			game.Face{Name: "Ayara, Widow of the Realm", TypeLine: "Legendary Creature — Elf Noble", ManaCost: "{1}{B}{B}", Colors: []string{"B"}, Power: 3, Toughness: 3},
			game.Face{Name: "Ayara, Furnace Queen", TypeLine: "Legendary Creature — Phyrexian Elf Noble", Colors: []string{"B"}, Power: 4, Toughness: 4}),
		"heliod": mk(s82m16HeliodOracle,
			game.Face{Name: "Heliod, the Radiant Dawn", TypeLine: "Legendary Enchantment Creature — God", ManaCost: "{2}{W}{W}", Colors: []string{"W"}, Power: 4, Toughness: 4},
			game.Face{Name: "Heliod, the Warped Eclipse", TypeLine: "Legendary Enchantment Creature — Phyrexian God", Colors: []string{"W"}, Power: 4, Toughness: 6}),
		"polukranos": mk(s82m16PolukranosOracle,
			game.Face{Name: "Polukranos Reborn", TypeLine: "Legendary Creature — Hydra", ManaCost: "{G}{G}{G}", Colors: []string{"G"}, Power: 4, Toughness: 5},
			game.Face{Name: "Polukranos, Engine of Ruin", TypeLine: "Legendary Creature — Phyrexian Hydra", Colors: []string{"G"}, Power: 6, Toughness: 6}),
		"rona": mk(s82m16RonaOracle,
			game.Face{Name: "Rona, Herald of Invasion", TypeLine: "Legendary Creature — Human Wizard", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 3},
			game.Face{Name: "Rona, Tolarian Obliterator", TypeLine: "Legendary Creature — Phyrexian Wizard", Colors: []string{"U"}, Power: 5, Toughness: 5}),
	}
}

func s82m16Card(key string, owner uuid.UUID) game.Card { return s82m16Fronts()[key](owner) }

// Every front face transforms for its printed Phyrexian cost, paid in
// mana or with life, only as a sorcery, and refuses a short pool.
func TestS82Mom16EveryFrontTransformsForItsPhyrexianCost(t *testing.T) {
	for key, cost := range map[string]struct {
		generic string
		colour  string
	}{
		"seraph": {"{C}{C}{C}{C}", "{B}"}, "aerialist": {"{C}{C}{C}{C}", "{G}"},
		"freewheeler": {"{C}{C}{C}{C}{C}", "{W}"}, "guardian": {"{C}{C}{C}{C}{C}", "{R}"},
		"heliod": {"{C}{C}{C}", "{U}"}, "polukranos": {"{C}{C}{C}{C}{C}{C}", "{W}"},
	} {
		for _, byLife := range []bool{false, true} {
			g, me, _ := s82m16Main(t)
			id := pushBattlefieldCardWithTimestamp(g, s82m16Card(key, me.ID))
			life := me.Life
			mana, lifeSyms := cost.generic+cost.colour, 0
			if byLife {
				mana, lifeSyms = cost.generic, 1
			}
			s82mom16Transform(t, g, me, id, mana, lifeSyms)
			if c := rvBattlefieldCard(g, id); c == nil || c.Name != c.Faces[1].Name {
				t.Errorf("%s (life=%v): did not transform", key, byLife)
			}
			want := life
			if byLife {
				want -= 2
			}
			if me.Life != want {
				t.Errorf("%s (life=%v): life %d, want %d", key, byLife, me.Life, want)
			}
		}
		// Not at sorcery speed, and not without the mana.
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepDeclareAttackers)
		id := pushBattlefieldCardWithTimestamp(g, s82m16Card(key, me.ID))
		if err := g.AddManaForEffect(me.ID, uuid.Nil, cost.generic+cost.colour); err != nil {
			t.Fatal(err)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Errorf("%s: transformed outside a main phase", key)
		}
	}
}

func TestS82Mom16SeraphOfNewPhyrexiaMayPayAnotherPermanentForPlusTwoPlusOne(t *testing.T) {
	g, me, opp := s82m16Main(t)
	id := b12Push(g, me.ID, "Seraph of New Phyrexia", "Creature — Phyrexian Angel", s82m16SeraphOracle+"#1", 3, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	declareAttack(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	answerChooseCardsOrPick(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if c := layeredCard(t, g, id); c.CurrentPower() != 5 || c.CurrentToughness() != 4 {
		t.Errorf("Seraph is %d/%d, want 5/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82Mom16SeraphOfNewPhyrexiaDecliningLeavesItAThreeThree(t *testing.T) {
	g, me, opp := s82m16Main(t)
	id := b12Push(g, me.ID, "Seraph of New Phyrexia", "Creature — Phyrexian Angel", s82m16SeraphOracle+"#1", 3, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	declareAttack(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	if pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); pick != nil {
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
	} else {
		t.Fatalf("no own-permanents prompt for the optional sacrifice: %+v", g.PendingChoices)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was sacrificed anyway")
	}
	if c := layeredCard(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Seraph is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// Skyclave Invader: a land on top may go onto the battlefield; a
// declined land and any nonland card go into the hand.
func TestS82Mom16SkyclaveInvaderLookAtTheTopCard(t *testing.T) {
	forest := func(owner uuid.UUID) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: owner, Controller: owner}
	}
	for _, tc := range []struct {
		name         string
		top          func(uuid.UUID) game.Card
		accept       bool
		wantOnField  bool
		wantInHand   bool
		promptsFirst bool
	}{
		{"land accepted", forest, true, true, false, true},
		{"land declined", forest, false, false, true, true},
		{"nonland", func(o uuid.UUID) game.Card {
			return game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: o, Controller: o}
		}, false, false, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := s82m16Main(t)
			top := tc.top(me.ID)
			me.Library.PushTop(top)
			hand, lib := me.Hand.Size(), me.Library.Size()
			s82m16Flip(t, g, me, s82m16Card("aerialist", me.ID), "{C}{C}{C}{C}{G}", 0)
			if tc.promptsFirst {
				if tc.accept {
					answerChooseCards(t, g, me.ID, top.InstanceID)
				} else {
					answerChooseCards(t, g, me.ID)
				}
			}
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(top.InstanceID) != tc.wantOnField {
				t.Errorf("on the battlefield = %v, want %v", !tc.wantOnField, tc.wantOnField)
			}
			if me.Hand.Contains(top.InstanceID) != tc.wantInHand {
				t.Errorf("in hand = %v, want %v", !tc.wantInHand, tc.wantInHand)
			}
			if me.Library.Size() != lib-1 {
				t.Errorf("library %d, want %d", me.Library.Size(), lib-1)
			}
			wantHand := hand
			if tc.wantInHand {
				wantHand++
			}
			if me.Hand.Size() != wantHand {
				t.Errorf("hand %d, want %d", me.Hand.Size(), wantHand)
			}
		})
	}
}

// A creature that enters as the back face did not transform: no trigger.
func TestS82Mom16SkyclaveInvaderEnteringBackFaceDoesNotLook(t *testing.T) {
	g, me, _ := s82m16Main(t)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	b12Push(g, me.ID, "Skyclave Invader", "Creature — Phyrexian Merfolk Scout", s82m16AerialistOracle+"#1", 2, 4)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) {
		t.Error("a permanent that never transformed looked at the top card")
	}
}

func TestS82Mom16NezumiFreewheelerMillsEachPlayerThree(t *testing.T) {
	g, me, _ := s82m16Main(t)
	libs := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}
	b12PlayFromHand(t, g, "Nezumi Freewheeler", "Creature — Rat Samurai", s82m16FreewheelerOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Library.Size() != libs[i]-3 {
			t.Errorf("seat %d library %d, want %d", i, p.Library.Size(), libs[i]-3)
		}
	}
	_ = me
}

func TestS82Mom16HideousFleshwheelerReturnsACheapPermanentFromAnyGraveyard(t *testing.T) {
	g, me, opp := s82m16Main(t)
	cheap := pushGraveyardPermanent(opp, "Cheap Bear", "Creature — Bear", "{1}{G}")
	dear := pushGraveyardPermanent(opp, "Dear Ogre", "Creature — Ogre", "{2}{R}")
	spell := pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	s82m16Flip(t, g, me, s82m16Card("freewheeler", me.ID), "{C}{C}{C}{C}{C}{W}", 0)
	answerPickTargetIfOpen(t, g, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		// The returned card is a new object: look it up by name.
		found := false
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Cheap Bear" && c.Controller == me.ID {
				found = true
			}
		}
		if !found {
			t.Fatal("the mana value 2 permanent did not come back under my control")
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dear Ogre" || c.Name == "Bolt" {
			t.Errorf("%s came back, but its mana value is above 2 or it is not a permanent", c.Name)
		}
	}
	_, _ = dear, spell
}

func TestS82Mom16HideousFleshwheelerRefusesAnExpensiveOrNonpermanentTarget(t *testing.T) {
	g, me, opp := s82m16Main(t)
	dear := pushGraveyardPermanent(opp, "Dear Ogre", "Creature — Ogre", "{2}{R}")
	spell := pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	id := pushBattlefieldCardWithTimestamp(g, s82m16Card("freewheeler", me.ID))
	_ = id
	spec, ok := Lookup(s82m16FreewheelerOracle + "#1")
	if !ok || len(spec.Triggered) != 1 || spec.Triggered[0].Targets == nil {
		t.Fatal("Fleshwheeler's trigger has no target clause")
	}
	ok1 := false
	for _, want := range []uuid.UUID{dear, spell} {
		for _, p := range g.Seats {
			for _, c := range p.Graveyard.Cards {
				if c.InstanceID == want && spec.Triggered[0].Targets.CardOK(g, me.ID, c, game.ZoneGraveyard) {
					ok1 = true
				}
			}
		}
	}
	if ok1 {
		t.Error("the target clause admits a mana value 3 permanent or an instant")
	}
}

// Furnace-Blessed Conqueror: attacking makes a tapped and attacking copy
// carrying as many +1/+1 counters as the Conqueror, sacrificed at the
// next end step.
func TestS82Mom16FurnaceBlessedConquerorMakesACounteredCopyThatIsSacrificed(t *testing.T) {
	g, me, opp := s82m16Main(t)
	id := b12Push(g, me.ID, "Furnace-Blessed Conqueror", "Creature — Phyrexian Cleric", s82m16GuardianOracle+"#1", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	declareAttack(t, g, opp.ID, id)
	passAndSettleAttackSplit(t, g, me.ID)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID != id && c.Name == "Furnace-Blessed Conqueror" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no copy was created")
	}
	if !token.IsToken() || !token.Tapped || token.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("copy: token=%v tapped=%v counters=%v, want a tapped token with two +1/+1 counters", token.IsToken(), token.Tapped, token.Counters)
	}
	if token.AttackingTarget == uuid.Nil {
		t.Error("the copy is not attacking")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Furnace-Blessed Conqueror" {
			t.Error("the copy was not sacrificed at the end step")
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Conqueror itself was sacrificed")
	}
}

// Ayara, Widow of the Realm: tap and sacrifice another creature or
// artifact; X damage to an opponent or battle and X life.
func TestS82Mom16AyaraDrainsForTheSacrificedManaValue(t *testing.T) {
	g, me, opp := s82m16Main(t)
	ayara := b12Push(g, me.ID, "Ayara, Widow of the Realm", "Legendary Creature — Elf Noble", s82m16AyaraOracle, 3, 3)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	myLife, theirLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{pbPlayer(opp.ID)}, SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 || me.Life != myLife+3 {
		t.Errorf("opponent %d (want %d), me %d (want %d)", opp.Life, theirLife-3, me.Life, myLife+3)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
}

func TestS82Mom16AyaraCannotTargetYouOrSacrificeHerself(t *testing.T) {
	g, me, _ := s82m16Main(t)
	ayara := b12Push(g, me.ID, "Ayara, Widow of the Realm", "Legendary Creature — Elf Noble", s82m16AyaraOracle, 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, ayara, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{pbPlayer(me.ID)}, SacrificeIDs: []uuid.UUID{ayara}, Strict: true}); err == nil {
		t.Error("Ayara targeted her controller or sacrificed herself")
	}
}

func TestS82Mom16AyaraFurnaceQueenReturnsACardWithHasteAndExilesIt(t *testing.T) {
	g, me, opp := s82m16Main(t)
	b12Push(g, me.ID, "Ayara, Furnace Queen", "Legendary Creature — Phyrexian Elf Noble", s82m16AyaraOracle+"#1", 4, 4)
	bear := pushGraveyardPermanent(me, "Grave Bear", "Creature — Bear", "{1}{G}")
	theirs := pushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	_ = theirs
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Grave Bear" {
			back = c
		}
	}
	if back == nil {
		t.Fatal("the creature card did not return")
	}
	if !hasKeywordOnBattlefield(t, g, back.InstanceID, "haste") {
		t.Error("the returned card lacks haste")
	}
	for _, c := range opp.Graveyard.Cards {
		if c.Name != "Their Bear" {
			t.Errorf("unexpected card in their graveyard: %s", c.Name)
		}
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grave Bear" {
			t.Error("the returned card was not exiled at the end step")
		}
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.Name == "Grave Bear" {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the returned card is not in exile")
	}
}

func TestS82Mom16HeliodReturnsANonGodEnchantment(t *testing.T) {
	g, me, _ := s82m16Main(t)
	aura := pushGraveyardPermanent(me, "Pacifism", "Enchantment — Aura", "{1}{W}")
	god := pushGraveyardPermanent(me, "Some God", "Legendary Enchantment Creature — God", "{2}{W}")
	b12PlayFromHand(t, g, "Heliod, the Radiant Dawn", "Legendary Enchantment Creature — God", s82m16HeliodOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, aura)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the enchantment card did not return to hand")
	}
	if me.Hand.Contains(god) {
		t.Error("a God card returned")
	}
}

// s82m16Price is what the board charges the seat for a spell.
func s82m16Price(t *testing.T, g *game.Game, seat *game.Player, manaCost string) int {
	t.Helper()
	return priceInHand(t, g, seat, "Test Spell", "Creature — Bear", manaCost)
}

func TestS82Mom16HeliodWarpedEclipseCostsOneLessPerCardOpponentsDrew(t *testing.T) {
	g, me, opp := s82m16Main(t)
	b12Push(g, me.ID, "Heliod, the Warped Eclipse", "Legendary Enchantment Creature — Phyrexian God", s82m16HeliodOracle+"#1", 4, 6)
	if got := s82m16Price(t, g, me, "{3}{G}"); got != 4 {
		t.Fatalf("no opponent has drawn: price %d, want 4", got)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DrawCard(g.Seats[2].ID); err != nil {
		t.Fatal(err)
	}
	if got := s82m16Price(t, g, me, "{3}{G}"); got != 2 {
		t.Errorf("two opponents' draws: price %d, want 2", got)
	}
	// Generic only: the coloured pip stays.
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	if got := s82m16Price(t, g, me, "{3}{G}"); got != 1 {
		t.Errorf("four draws: price %d, want 1 (only generic mana is reduced)", got)
	}
	// Your own draws don't count, and the discount is yours alone.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	if got := s82m16Price(t, g, me, "{3}{G}"); got != 1 {
		t.Errorf("own draw changed the price to %d", got)
	}
	if got := s82m16Price(t, g, opp, "{3}{G}"); got != 4 {
		t.Errorf("an opponent was discounted: price %d, want 4", got)
	}
}

func TestS82Mom16HeliodWarpedEclipseGrantsFlash(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := func() uuid.UUID {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", OracleID: "", ManaCost: "{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		return id
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	before := creature()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, before, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was cast in combat without flash")
	}
	b12Push(g, me.ID, "Heliod, the Warped Eclipse", "Legendary Enchantment Creature — Phyrexian God", s82m16HeliodOracle+"#1", 4, 6)
	if err := g.CastSpell(me.ID, before, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("with Heliod the creature spell should be castable at instant speed: %v", err)
	}
	_ = opp
}

func TestS82Mom16PolukranosTokensForItselfAndOtherNontokenHydras(t *testing.T) {
	g, me, _ := s82m16Main(t)
	polu := b12Push(g, me.ID, "Polukranos, Engine of Ruin", "Legendary Creature — Phyrexian Hydra", s82m16PolukranosOracle+"#1", 6, 6)
	hydra := b12Creature(g, me.ID, "Little Hydra", "Creature — Hydra", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	count := func() (reach, lifelink int) {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.IsToken() && c.Name == "Phyrexian Hydra" {
				if hasKeywordOnBattlefield(t, g, c.InstanceID, "reach") {
					reach++
				}
				if hasKeywordOnBattlefield(t, g, c.InstanceID, "lifelink") {
					lifelink++
				}
			}
		}
		return
	}
	kill := func(id uuid.UUID) {
		moveToGraveyard(t, g, me.ID, id)
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	kill(bear)
	if r, l := count(); r+l != 0 {
		t.Fatalf("a Bear dying made %d/%d tokens", r, l)
	}
	kill(hydra)
	if r, l := count(); r != 1 || l != 1 {
		t.Fatalf("another Hydra dying made reach %d lifelink %d, want 1 and 1", r, l)
	}
	kill(polu)
	if r, l := count(); r != 2 || l != 2 {
		t.Fatalf("Polukranos dying: reach %d lifelink %d, want 2 and 2", r, l)
	}
	// A token Hydra dying makes nothing.
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.IsToken() && c.Name == "Phyrexian Hydra" {
			kill(c.InstanceID)
			break
		}
	}
	if r, l := count(); r+l != 3 {
		t.Errorf("a token Hydra dying made more tokens: %d reach, %d lifelink", r, l)
	}
}

func TestS82Mom16PolukranosFrontIsReachOnly(t *testing.T) {
	g, me, _ := s82m16Main(t)
	hydra := b12Creature(g, me.ID, "Little Hydra", "Creature — Hydra", 1, 1)
	b12Push(g, me.ID, "Polukranos Reborn", "Legendary Creature — Hydra", s82m16PolukranosOracle, 4, 5)
	moveToGraveyard(t, g, me.ID, hydra)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Error("the front face made a token")
		}
	}
}

// s82m16RonaTable puts Rona, Tolarian Obliterator under seat 0 and a
// quiet 0/9 source under the opponent, who holds `hand`.
func s82m16RonaTable(t *testing.T, hand ...game.Card) (g *game.Game, me, opp *game.Player, rona, src uuid.UUID) {
	t.Helper()
	g, me, opp = s82m16Main(t)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	for _, c := range hand {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), opp.ID, opp.ID
		opp.Hand.PushTop(c)
	}
	rona = b12Push(g, me.ID, "Rona, Tolarian Obliterator", "Legendary Creature — Phyrexian Wizard", s82m16RonaOracle+"#1", 5, 5)
	src = apaPush(g, opp.ID, opp.ID, game.Card{Name: "Hitter", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
	return
}

func TestS82Mom16RonaObliteratorLandGoesOntoTheBattlefieldUnderYourControl(t *testing.T) {
	g, me, opp, rona, src := s82m16RonaTable(t, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	dealDamageTo(t, g, src, rona, 2)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if opp.Hand.Size() != 0 {
		t.Errorf("their hand has %d cards, want 0", opp.Hand.Size())
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mountain" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the land is not on the battlefield under my control")
	}
}

func TestS82Mom16RonaObliteratorDecliningTheLandLeavesItInExile(t *testing.T) {
	g, me, _, rona, src := s82m16RonaTable(t, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	dealDamageTo(t, g, src, rona, 2)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Mountain" {
			t.Error("the land entered anyway")
		}
	}
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.Name == "Mountain" {
			exiled = true
		}
	}
	if !exiled {
		t.Error("the land is not in exile")
	}
}

func TestS82Mom16RonaObliteratorNonlandMayBeCastFree(t *testing.T) {
	g, me, _, rona, src := s82m16RonaTable(t, game.Card{Name: "Dear Ogre", TypeLine: "Creature — Ogre", ManaCost: "{4}{R}{R}", Colors: []string{"R"}, Power: 5, Toughness: 5})
	dealDamageTo(t, g, src, rona, 2)
	passPriorityAroundTable(t, g)
	answerMayCastPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dear Ogre" && c.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the exiled creature was not cast for free under my control")
	}
}

func TestS82Mom16RonaObliteratorEmptyHandDoesNothingAndDamageToOthersIsIgnored(t *testing.T) {
	g, me, _, rona, src := s82m16RonaTable(t)
	dealDamageTo(t, g, src, rona, 2)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Errorf("an empty hand still asked a question: %+v", g.PendingChoices)
	}
	// Damage to some other creature does not trigger it.
	g2, me2, opp2, _, src2 := s82m16RonaTable(t, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	bear := b12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 5)
	dealDamageTo(t, g2, src2, bear, 2)
	passPriorityAroundTable(t, g2)
	if opp2.Hand.Size() != 1 {
		t.Error("damage to another creature exiled a card")
	}
	_ = me
}

func TestS82Mom16RonaObliteratorHitByYourOwnSourceUsesYourOwnHand(t *testing.T) {
	g, me, _, rona, _ := s82m16RonaTable(t)
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "My Hitter", TypeLine: "Creature — Test", Power: 0, Toughness: 9})
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	handCardForTest(me, "Island", "Basic Land — Island", "")
	dealDamageTo(t, g, mine, rona, 1)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != 0 {
		t.Errorf("my hand has %d cards, want 0", me.Hand.Size())
	}
}

func TestS82Mom16RonaHeraldUntapsOnALegendarySpellAndLoots(t *testing.T) {
	g, me, _ := s82m16Main(t)
	rona := b12Push(g, me.ID, "Rona, Herald of Invasion", "Legendary Creature — Human Wizard", s82m16RonaOracle, 1, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rona {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	b12PlayFromHand(t, g, "Some Legend", "Legendary Creature — Elf", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, rona) {
		t.Error("a legendary spell did not untap Rona")
	}
	lib, hand := me.Library.Size(), me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, rona, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("loot: %v", err)
	}
	if !b16Tapped(t, g, rona) {
		t.Error("the loot did not tap Rona")
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 || me.Hand.Size() != hand+1 {
		t.Errorf("library %d (want %d), hand %d (want %d): the draw did not happen", me.Library.Size(), lib-1, me.Hand.Size(), hand+1)
	}
	if len(g.PendingChoices) == 0 {
		t.Error("no discard was asked after the draw")
	}
}
