package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conditional_exile_more_cards_test.go — #3192's second slice: the cards
// #3211 left for a follow-up, one rules test (or more) per card.

const (
	gandalfGoblinsBaneOracle = "f48f2a9b-4ba5-4984-8180-3950889bba65"
	intetTheDreamerOracle    = "9518ffbc-bb3d-4f48-a900-f5bad8302a87"
	summonBrynhildrOracle    = "bab5e9ce-6d55-4d1e-a9ac-c2b954191be9"
	vivienChampionOracle     = "146eb8c8-93d5-4dd5-af7a-46d5acd4a89d"
	dayOfTheDoctorOracle     = "c2ee17b2-6119-4ade-b806-1add64513c3a"
	tasterOfWaresOracle      = "020aae46-cdba-4c37-9c54-283727c09994"
)

func TestConditionalExileMoreCardsAreFull(t *testing.T) {
	for _, id := range []string{
		gandalfGoblinsBaneOracle, gandalfGoblinsBaneOracle + "#1", intetTheDreamerOracle,
		summonBrynhildrOracle, vivienChampionOracle, dayOfTheDoctorOracle, tasterOfWaresOracle,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// ceOnBattlefield reads a permanent back by ID.
func ceOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%v is not on the battlefield", id)
	return game.Card{}
}

// --- Gandalf, Goblins' Bane // Flameshape -------------------------------

func gandalfCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   gandalfGoblinsBaneOracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Gandalf, Goblins' Bane", TypeLine: "Legendary Creature — Avatar Wizard", ManaCost: "{2}{R}", Power: 2, Toughness: 3},
			{Name: "Flameshape", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{R}"},
		},
	}
	c.SetFace(0)
	return c
}

// Flameshape exiles the top two face down, seen only by their owner, and
// they are playable for as long as they stay exiled but only while their
// holder controls a Wizard.
func TestFlameshapeCardsArePlayableOnlyWhileYouControlAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	topLibraryCard(me, "Deep Card", "Sorcery")
	second := topLibraryCard(me, "Second Card", "Sorcery")
	first := topLibraryCard(me, "First Card", "Instant")
	card := gandalfCard(me.ID)
	me.Hand.PushTop(card)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	b06AddMana(me, "R", "C")

	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Flameshape: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{first, second} {
		if !g.Exile.Contains(id) {
			t.Fatal("Flameshape did not exile the top two cards")
		}
		c, _ := g.LookupCardForEffect(id)
		if !c.FaceDown || !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
			t.Errorf("%s: face down = %v, known to me = %v, to an opponent = %v; want face down and seen only by me",
				c.Name, c.FaceDown, c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
		}
		perm := exiledPermission(g, id)
		if perm.Condition == nil || perm.Condition.Key != game.CondControlSubtype || perm.Condition.Subtype != "Wizard" {
			t.Fatalf("%+v: the grant is conditional on controlling a Wizard", perm)
		}
		if permissionLive(g, perm, me.ID) {
			t.Error("playable with no Wizard on the battlefield")
		}
	}

	wizard := b12Creature(g, me.ID, "Apprentice", "Creature — Human Wizard", 1, 1)
	if !permissionLive(g, exiledPermission(g, first), me.ID) {
		t.Fatal("a Wizard arrived and the card stayed shut")
	}
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, first), me.ID) {
		t.Error("the grant lapsed with the turn; it lasts for as long as the card remains exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wizard) })
	if permissionLive(g, exiledPermission(g, first), me.ID) {
		t.Error("the last Wizard left and the card stayed playable")
	}
}

// Gandalf's trigger: a noncreature spell pumps him and pings each
// opponent; a creature spell does nothing.
func TestGandalfPingsEachOpponentForANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gandalf := gandalfCard(me.ID)
	gandalfID := pushBattlefieldCardWithTimestamp(g, gandalf)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID] {
			t.Fatalf("a creature spell triggered Gandalf: %s at %d", p.Name, p.Life)
		}
	}

	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s: life %d, want %d", p.Name, p.Life, want)
		}
	}
	c := ceOnBattlefield(t, g, gandalfID)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 4 {
		t.Errorf("Gandalf is %d/%d, want 3/4 until end of turn", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Intet, the Dreamer --------------------------------------------------

func TestIntetPaysToExileFaceDownAndPlaysFreeWhileHeRemains(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "decline"
		if pay {
			name = "pay"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			intet := b12Push(g, me.ID, "Intet, the Dreamer", "Legendary Creature — Dragon", intetTheDreamerOracle, 6, 6)
			top := topLibraryCard(me, "Held Card", "Sorcery")

			attackWith(t, g, opp.ID, intet)
			passPriorityAroundTable(t, g)
			if pay {
				b06AddMana(me, "U", "C", "C")
			}
			prompt := answerPayUnless(t, g, me.ID, pay)
			if prompt.PayCost != "{2}{U}" {
				t.Errorf("optional cost = %q, want {2}{U}", prompt.PayCost)
			}
			if !pay {
				if g.Exile.Contains(top) {
					t.Fatal("declining the payment exiled a card")
				}
				return
			}
			if !g.Exile.Contains(top) {
				t.Fatal("paying {2}{U} exiled nothing")
			}
			c, _ := g.LookupCardForEffect(top)
			if !c.FaceDown || !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
				t.Errorf("face down = %v, known to me = %v, to the opponent = %v", c.FaceDown, c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
			}
			perm := exiledPermission(g, top)
			if perm.Cost != "{0}" || !permissionLive(g, perm, me.ID) {
				t.Fatalf("%+v: a free play while Intet remains", perm)
			}
			nextTurnOfMine(t, g)
			if !permissionLive(g, exiledPermission(g, top), me.ID) {
				t.Error("the grant lapsed with the turn")
			}
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(intet) })
			if permissionLive(g, exiledPermission(g, top), me.ID) {
				t.Error("playable after Intet left the battlefield")
			}
		})
	}
}

// --- Summon: Brynhildr ---------------------------------------------------

func TestBrynhildrChapterICardIsPlayableOnLoreCounterTurns(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	topLibraryCard(me, "Filler", "Sorcery")
	topLibraryCard(me, "Filler", "Sorcery")
	held := topLibraryCard(me, "Held Card", "Sorcery")

	tdCastFromHand(t, g, game.Card{Name: "Summon: Brynhildr", TypeLine: "Enchantment Creature — Saga Knight",
		OracleID: summonBrynhildrOracle, Power: 2, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g) // the creature, then chapter I
	if !g.Exile.Contains(held) {
		t.Fatal("chapter I exiled nothing")
	}
	perm := exiledPermission(g, held)
	if perm.Condition == nil || perm.Condition.Key != game.CondLoreCounterOnObject {
		t.Fatalf("%+v: the grant is conditional on a lore counter on this Saga", perm)
	}
	if !permissionLive(g, perm, me.ID) {
		t.Fatal("the Saga got its lore counter this turn and the card is shut")
	}

	// An opponent's turn: no lore counter, dormant but granted.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if perm := exiledPermission(g, held); perm.Player != me.ID || permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: dormant on a turn no lore counter went on the Saga", perm)
	}

	// My next turn: the draw step's lore counter opens it again.
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g) // chapter II
	if !permissionLive(g, exiledPermission(g, held), me.ID) {
		t.Error("a lore counter went on the Saga this turn and the card is shut")
	}
}

// Chapter II: the next creature spell this turn gains haste, and a
// noncreature spell does not spend it.
func TestBrynhildrChapterIIGivesTheNextCreatureSpellHaste(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Summon: Brynhildr", TypeLine: "Enchantment Creature — Saga Knight",
		OracleID: summonBrynhildrOracle, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 1},
	})
	advanceToPrecombatMainOf(t, g, seat) // the draw step adds lore counter II
	passPriorityAroundTable(t, g)
	if got := ceOnBattlefield(t, g, saga).Counters[game.CounterLore]; got != 2 {
		t.Fatalf("lore counters = %d, want 2", got)
	}

	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	first := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	second := castCatalogSpell(t, g, "Second Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)

	if c := ceOnBattlefield(t, g, first); !game.HasKeyword(&c, "haste") {
		t.Error("the first creature spell after chapter II has no haste")
	}
	if c := ceOnBattlefield(t, g, second); game.HasKeyword(&c, "haste") {
		t.Error("the promise is one use and a second creature got haste too")
	}
}

// --- Vivien, Champion of the Wilds ----------------------------------------

func TestVivienLetsCreaturesBeCastAtFlashSpeed(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	advanceToMainOf(t, g, seat)
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err == nil {
		t.Fatal("cast a creature in combat with no flash")
	}
	pushWalkerForTest(g, me.ID, "Vivien, Champion of the Wilds", vivienChampionOracle, 4)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("Vivien's static did not let a creature be cast in combat: %v", err)
	}
	sorcery := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Plain Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err == nil {
		t.Error("Vivien gave a sorcery flash")
	}
}

func TestVivienPlusOneGivesReachAndVigilanceUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	vivien := pushWalkerForTest(g, me.ID, "Vivien, Champion of the Wilds", vivienChampionOracle, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(me.ID, vivien, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, vivien); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
	c := ceOnBattlefield(t, g, bear)
	if !game.HasKeyword(&c, "reach") || !game.HasKeyword(&c, "vigilance") {
		t.Fatalf("keywords = %v, want reach and vigilance", c.Keywords)
	}
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if c := ceOnBattlefield(t, g, bear); !game.HasKeyword(&c, "reach") {
		t.Error("reach ended before Vivien's controller's next turn")
	}
	advanceToMainOf(t, g, seat)
	if c := ceOnBattlefield(t, g, bear); game.HasKeyword(&c, "reach") || game.HasKeyword(&c, "vigilance") {
		t.Error("the keywords outlasted the start of Vivien's controller's next turn")
	}
}

// −2: look at three, exile the chosen one face down, the rest on the
// bottom in the chosen order; the exiled card is castable only if it is
// a creature.
func TestVivienMinusTwoExilesOneFaceDownCastableIfACreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		castable       bool
	}{
		{"Held Bear", "Creature — Bear", true},
		{"Held Bolt", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			opp := g.Seats[(seat+1)%len(g.Seats)]
			deep := topLibraryCard(me, "Deep Card", "Sorcery")
			a := topLibraryCard(me, "Third", "Sorcery")
			b := topLibraryCard(me, "Second", "Sorcery")
			held := topLibraryCard(me, tc.name, tc.typeLine)
			vivien := pushWalkerForTest(g, me.ID, "Vivien, Champion of the Wilds", vivienChampionOracle, 4)
			advanceToMainOf(t, g, seat)

			if err := g.ActivateCatalogAbility(me.ID, vivien, 1, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("−2: %v", err)
			}
			passPriorityAroundTable(t, g)
			pick := chooseCardsChoiceFor(g, me.ID)
			if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
				t.Fatalf("want a pick of exactly one of three cards, got %+v", pick)
			}
			answerChooseCards(t, g, me.ID, held)
			order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
			if order == nil {
				t.Fatal("the rest are put on the bottom in an order the player chooses")
			}
			if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{b, a}, nil); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}

			if !g.Exile.Contains(held) {
				t.Fatal("the chosen card was not exiled")
			}
			c, _ := g.LookupCardForEffect(held)
			if !c.FaceDown || !c.IsKnownTo(me.ID) || c.IsKnownTo(opp.ID) {
				t.Errorf("face down = %v, known to me = %v, to an opponent = %v", c.FaceDown, c.IsKnownTo(me.ID), c.IsKnownTo(opp.ID))
			}
			if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != deep {
				t.Error("the two cards left were not put on the bottom")
			}
			if live := permissionLive(g, exiledPermission(g, held), me.ID); live != tc.castable {
				t.Errorf("%s: castable = %v, want %v", tc.typeLine, live, tc.castable)
			}
			if tc.castable {
				nextTurnOfMine(t, g)
				if !permissionLive(g, exiledPermission(g, held), me.ID) {
					t.Error("the grant lapsed; it lasts for as long as the card remains exiled")
				}
			}
		})
	}
}

// --- The Day of the Doctor -----------------------------------------------

func TestDayOfTheDoctorDigsToALegendaryCardPlayableWhileTheSagaRemains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deep := topLibraryCard(me, "Deep Card", "Sorcery")
	legend := topLibraryCard(me, "A Legend", "Legendary Creature — Human")
	land := topLibraryCard(me, "A Land", "Land")
	spell := topLibraryCard(me, "A Spell", "Sorcery")

	saga := castCatalogSpell(t, g, "The Day of the Doctor", "Enchantment — Saga", dayOfTheDoctorOracle, nil)
	passPriorityAroundTable(t, g) // the Saga, then chapter I
	if !g.Exile.Contains(legend) {
		t.Fatal("chapter I did not exile the legendary card")
	}
	if g.Exile.Contains(land) || g.Exile.Contains(spell) || !me.Library.Contains(land) || !me.Library.Contains(spell) {
		t.Fatal("the cards exiled on the way did not go back to the library")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != deep {
		t.Error("the rest went somewhere other than the bottom")
	}
	if !permissionLive(g, exiledPermission(g, legend), me.ID) {
		t.Fatal("the legendary card is not playable while the Saga remains")
	}
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, legend), me.ID) {
		t.Error("the grant lapsed with the turn")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(saga) })
	if permissionLive(g, exiledPermission(g, legend), me.ID) {
		t.Error("playable after the Saga left the battlefield")
	}
}

func TestDayOfTheDoctorChapterIVKeepsTheChosenDoctorsAndCosts13(t *testing.T) {
	for _, yes := range []bool{true, false} {
		name := "decline"
		if yes {
			name = "exile"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			opp := g.Seats[(seat+1)%len(g.Seats)]
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "The Day of the Doctor", TypeLine: "Enchantment — Saga",
				OracleID: dayOfTheDoctorOracle, Owner: me.ID, Controller: me.ID,
				Counters: map[string]int{game.CounterLore: 3},
			})
			kept := b12Creature(g, opp.ID, "Their Doctor", "Legendary Creature — Time Lord Doctor", 2, 2)
			otherDoctor := b12Creature(g, me.ID, "Unchosen Doctor", "Legendary Creature — Time Lord Doctor", 2, 2)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			life := me.Life

			advanceToPrecombatMainOf(t, g, seat) // lore counter IV
			passPriorityAroundTable(t, g)
			pick := chooseCardsChoiceFor(g, me.ID)
			if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 2 {
				t.Fatalf("want a pick of up to three among the two Doctors, got %+v", pick)
			}
			answerChooseCards(t, g, me.ID, kept)
			answerMayChoice(t, g, me.ID, yes)

			if !g.Battlefield.Contains(kept) {
				t.Error("the chosen Doctor was exiled")
			}
			if yes {
				if g.Battlefield.Contains(otherDoctor) || g.Battlefield.Contains(bear) {
					t.Error("a creature that was not chosen stayed")
				}
				if me.Life != life-13 {
					t.Errorf("life = %d, want %d", me.Life, life-13)
				}
				return
			}
			if !g.Battlefield.Contains(otherDoctor) || !g.Battlefield.Contains(bear) || me.Life != life {
				t.Error("declining exiled creatures or dealt damage")
			}
		})
	}
}

// --- Taster of Wares ------------------------------------------------------

// tasterSetUp seats a Taster's controller with a second Goblin (X = 2)
// and an opponent holding three cards.
func tasterSetUp(t *testing.T) (*game.Game, *game.Player, *game.Player, map[string]uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	b12Creature(g, me.ID, "Goblin Friend", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	cards := map[string]uuid.UUID{}
	for _, c := range []struct{ name, typeLine string }{
		{"Their Bolt", "Instant"}, {"Their Bear", "Creature — Bear"}, {"Their Rite", "Sorcery"},
	} {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, ManaCost: "{U}", Owner: opp.ID, Controller: opp.ID})
		cards[c.name] = id
	}
	return g, me, opp, cards
}

func tasterReveal(t *testing.T, g *game.Game, me, opp *game.Player, reveal []uuid.UUID, choose uuid.UUID) uuid.UUID {
	t.Helper()
	taster := tdCastFromHand(t, g, game.Card{Name: "Taster of Wares", TypeLine: "Creature — Goblin Warlock",
		OracleID: tasterOfWaresOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	reveal2 := chooseCardsChoiceFor(g, opp.ID)
	if reveal2 == nil || reveal2.ChooseMin != 2 || reveal2.ChooseMax != 2 {
		t.Fatalf("the opponent chooses exactly X = 2 cards to reveal, got %+v", reveal2)
	}
	answerChooseCards(t, g, opp.ID, reveal...)
	var rp *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRevealPick && c.Chooser == me.ID {
			rp = c
		}
	}
	if rp == nil {
		t.Fatal("no reveal pick for the Taster's controller")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{choose}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	return taster
}

func TestTasterOfWaresCastsAnExiledInstantWhileYouControlIt(t *testing.T) {
	g, me, opp, cards := tasterSetUp(t)
	taster := tasterReveal(t, g, me, opp, []uuid.UUID{cards["Their Bolt"], cards["Their Bear"]}, cards["Their Bolt"])

	bolt, rite := cards["Their Bolt"], cards["Their Rite"]
	if !g.Exile.Contains(bolt) {
		t.Fatal("the chosen card was not exiled")
	}
	if c, _ := g.LookupCardForEffect(rite); c.IsKnownTo(me.ID) {
		t.Error("a card the opponent did not reveal was shown to the Taster's controller")
	}
	perm := exiledPermission(g, bolt)
	if perm.Player != me.ID || !perm.CastOnly || !perm.AnyType || !permissionLive(g, perm, me.ID) {
		t.Fatalf("%+v: its controller may cast it with mana of any type", perm)
	}
	nextTurnOfMine(t, g)
	if !permissionLive(g, exiledPermission(g, bolt), me.ID) {
		t.Error("the grant lapsed with the turn")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(taster) })
	if permissionLive(g, exiledPermission(g, bolt), me.ID) {
		t.Error("castable after the Taster left the battlefield")
	}
}

func TestTasterOfWaresExilesACreatureWithNoGrant(t *testing.T) {
	g, me, opp, cards := tasterSetUp(t)
	tasterReveal(t, g, me, opp, []uuid.UUID{cards["Their Bolt"], cards["Their Bear"]}, cards["Their Bear"])
	bear := cards["Their Bear"]
	if !g.Exile.Contains(bear) {
		t.Fatal("the chosen card was not exiled")
	}
	if perm := g.CastPermissionOnCardByIDForEffect(bear); perm != nil {
		t.Errorf("a creature card was granted a cast: %+v", perm)
	}
	if !opp.Hand.Contains(cards["Their Bolt"]) {
		t.Error("the revealed card that was not chosen left the hand")
	}
}
