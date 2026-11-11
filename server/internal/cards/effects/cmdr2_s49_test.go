package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s49TemptBunnies   = "6874be51-623d-4eaf-a67a-2fcbb9a8ceca"
	s49TemptVengeance = "8e356df5-ca92-4be2-871e-8965c2510fbe"
)

func s49Library(g *game.Game, n int) {
	for _, p := range g.Seats {
		for i := 0; i < n; i++ {
			pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
	}
}

func TestS49TemptWithBunnies(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 5)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	hand := func(p *game.Player) int { return len(p.Hand.Cards) }
	before := map[*game.Player]int{me: hand(me), o1: hand(o1), o2: hand(o2), o3: hand(o3)}

	castCatalogSpell(t, g, "Tempt with Bunnies", "Sorcery", s49TemptBunnies, nil)
	passPriorityAroundTable(t, g)
	// The caster's own half is done before any opponent is asked.
	if n := countOnBattlefield(g, "Rabbit", me.ID); n != 1 {
		t.Fatalf("caster rabbits before offers = %d, want 1", n)
	}
	answerMayChoice(t, g, o1.ID, true)
	answerMayChoice(t, g, o2.ID, false)
	answerMayChoice(t, g, o3.ID, true)

	if n := countOnBattlefield(g, "Rabbit", me.ID); n != 3 {
		t.Errorf("caster rabbits = %d, want 3", n)
	}
	for _, tc := range []struct {
		p    *game.Player
		want int
	}{{o1, 1}, {o2, 0}, {o3, 1}} {
		if n := countOnBattlefield(g, "Rabbit", tc.p.ID); n != tc.want {
			t.Errorf("%s rabbits = %d, want %d", tc.p.Name, n, tc.want)
		}
	}
	// The spell left the hand: three draws net +2 would be wrong only if a draw is missing.
	if hand(me) != before[me]+2 {
		t.Errorf("caster hand = %d, want %d (three draws, the spell left it)", hand(me), before[me]+2)
	}
	if hand(o1) != before[o1]+1 || hand(o2) != before[o2] || hand(o3) != before[o3]+1 {
		t.Errorf("opponent hands wrong: %d %d %d", hand(o1)-before[o1], hand(o2)-before[o2], hand(o3)-before[o3])
	}
}

func TestS49TemptWithVengeance(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	pr8CastX(t, g, "Tempt with Vengeance", "Sorcery", s49TemptVengeance, 2, nil)
	if n := countOnBattlefield(g, "Elemental", me.ID); n != 2 {
		t.Fatalf("caster elementals before offers = %d, want 2", n)
	}
	answerMayChoice(t, g, o1.ID, false)
	answerMayChoice(t, g, o2.ID, true)
	answerMayChoice(t, g, o3.ID, true)
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"caster", countOnBattlefield(g, "Elemental", me.ID), 6},
		{"o1", countOnBattlefield(g, "Elemental", o1.ID), 0},
		{"o2", countOnBattlefield(g, "Elemental", o2.ID), 2},
		{"o3", countOnBattlefield(g, "Elemental", o3.ID), 2},
	} {
		if tc.got != tc.want {
			t.Errorf("%s elementals = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Elemental" && !game.HasKeyword(&c, "haste") {
			t.Errorf("an Elemental without haste")
		}
	}
}

const (
	s49Selvala      = "b628a39b-c923-473d-900a-1617f47ecb76"
	s49Negotiator   = "60cb885c-5cd3-4963-8a6a-cebf07cc61b5"
	s49Researcher   = "ef764497-800e-4cbd-805e-7635919a983c"
	s49ShellGame    = "88c3fc46-f2f5-440c-80a6-fed5b42cd340"
	s49Disorienting = "7aa34ef4-5e11-491b-8dbe-6ef80e0ad26c"
	s49Offering     = "73ec7a0b-274a-4a10-a173-ff1625d62fce"
	s49Leonin       = "6ec98a34-5c04-423c-98e5-8cf3047135c1"
	s49Dilemma      = "fecbbc79-2f74-46b3-a049-45736f904f58"
)

// s49TopCards puts one known card on top of each library: nonland for
// the seats in `nonland`, a land for the others.
func s49TopCards(g *game.Game, nonland ...int) {
	for i, p := range g.Seats {
		c := game.Card{Name: "Top Land", TypeLine: "Basic Land — Forest"}
		for _, n := range nonland {
			if n == i {
				c = game.Card{Name: "Top Spell", TypeLine: "Creature — Bear"}
			}
		}
		pushLibraryCardForTest(p, c)
	}
}

func s49Hands(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = len(p.Hand.Cards)
	}
	return out
}

func TestS49SelvalaParleyIsAManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 4)
	s49TopCards(g, 0, 2)
	me := g.Seats[0]
	selvala := pushCatalogPermanent(g, me.ID, "Selvala, Explorer Returned", "Legendary Creature — Elf Scout", s49Selvala, false)
	hands, life := s49Hands(g), me.Life
	if err := g.ActivateManaAbility(me.ID, selvala, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool = %v, want two {G} (two nonland cards revealed)", got)
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	for i, h := range s49Hands(g) {
		if h != hands[i]+1 {
			t.Errorf("seat %d hand %d, want %d (each player draws)", i, h, hands[i]+1)
		}
	}
	if !stackFullyEmpty(g) {
		t.Error("a mana ability must not use the stack")
	}
}

func TestS49SelvalaWithOnlyLandsAddsNothing(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 4)
	s49TopCards(g)
	me := g.Seats[0]
	selvala := pushCatalogPermanent(g, me.ID, "Selvala, Explorer Returned", "Legendary Creature — Elf Scout", s49Selvala, false)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, selvala, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 0 || me.Life != life {
		t.Errorf("pool %v life %d→%d, want nothing", got, life, me.Life)
	}
}

func s49Attack(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

func TestS49StormFleetNegotiatorMakesAMapPerNonland(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 4)
	s49TopCards(g, 0, 1, 3)
	me := g.Seats[0]
	neg := pushCatalogPermanent(g, me.ID, "Storm Fleet Negotiator", "Creature — Siren Pirate", s49Negotiator, false)
	hands := s49Hands(g)
	s49Attack(t, g, neg)
	if n := countOnBattlefield(g, "Map", me.ID); n != 3 {
		t.Errorf("Maps = %d, want 3", n)
	}
	for i, h := range s49Hands(g) {
		if h != hands[i]+1 {
			t.Errorf("seat %d hand %d, want %d", i, h, hands[i]+1)
		}
	}
}

func TestS49InnocuousResearcherInvestigatesPerNonland(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 4)
	s49TopCards(g, 2)
	me := g.Seats[0]
	res := pushCatalogPermanent(g, me.ID, "Innocuous Researcher", "Creature — Centaur Detective", s49Researcher, false)
	s49Attack(t, g, res)
	if n := countOnBattlefield(g, "Clue", me.ID); n != 1 {
		t.Errorf("Clues = %d, want 1", n)
	}
}

func TestS49InnocuousResearcherEndStepUntapsLandsAndBansSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Innocuous Researcher", "Creature — Centaur Detective", s49Researcher, false)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == land && c.Tapped {
			t.Error("the land stayed tapped")
		}
	}
	bolt := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(bolt)
	err := g.CastSpell(me.ID, bolt.InstanceID, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
	if err == nil {
		t.Error("casting after accepting should be refused until your next turn")
	}
}

func TestS49SadisticShellGameEachPlayerChoosesACreatureYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	a := pushCatalogPermanent(g, o1.ID, "A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, o2.ID, "B", "Creature — Bear", "", false)
	c := pushCatalogPermanent(g, o3.ID, "C", "Creature — Bear", "", false)
	castCatalogSpell(t, g, "Sadistic Shell Game", "Sorcery", s49ShellGame, nil)
	passPriorityAroundTable(t, g)
	// Next opponent in turn order first, the caster last.
	for _, step := range []struct {
		who  *game.Player
		pick uuid.UUID
	}{{o1, b}, {o2, b}, {o3, c}, {me, a}} {
		ch := chooseCardsChoiceFor(g, step.who.ID)
		if ch == nil {
			t.Fatalf("no choice open for %s", step.who.Name)
		}
		for _, id := range ch.ChooseCards {
			if id == mine {
				t.Errorf("%s was offered a creature the caster controls", step.who.Name)
			}
		}
		answerChooseCards(t, g, step.who.ID, step.pick)
	}
	passPriorityAroundTable(t, g)
	onBF := map[uuid.UUID]bool{}
	for _, card := range g.BattlefieldCardsForEffect() {
		onBF[card.InstanceID] = true
	}
	if !onBF[mine] {
		t.Error("the caster's own creature was destroyed")
	}
	if onBF[a] || onBF[b] || onBF[c] {
		t.Errorf("chosen creatures must die: a=%v b=%v c=%v", onBF[a], onBF[b], onBF[c])
	}
}

func TestS49IntellectualOfferingTwoChoices(t *testing.T) {
	g := newCatalogGame(t)
	s49Library(g, 6)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	mineTapped := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, o2.ID, "Bear2", "Creature — Bear", "", false)
	other := pushCatalogPermanent(g, o1.ID, "Bear3", "Creature — Bear", "", false)
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case mineTapped, theirs, other, land:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hands := s49Hands(g)
	castCatalogSpell(t, g, "Intellectual Offering", "Instant", s49Offering, nil)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0) // first opponent: draws
	answerOptionPick(t, g, me.ID, 1) // second choice names another opponent
	passPriorityAroundTable(t, g)
	hs := s49Hands(g)
	if hs[0] != hands[0]-1+3 || hs[1] != hands[1]+3 || hs[2] != hands[2] || hs[3] != hands[3] {
		t.Errorf("hands %v from %v, want caster -1+3 and the first opponent +3", hs, hands)
	}
	tapped := map[uuid.UUID]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		tapped[c.InstanceID] = c.Tapped
	}
	if tapped[mineTapped] || tapped[theirs] {
		t.Error("your nonland permanents and the second chosen opponent's should untap")
	}
	if !tapped[land] {
		t.Error("lands do not untap")
	}
	if !tapped[other] {
		t.Error("an opponent who was not chosen keeps their permanents tapped")
	}
}

func s49Tap(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		}
	})
}

func TestS49DisorientingChoiceExilesOrFetchesLands(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	for i := 0; i < 3; i++ {
		b07LibraryLand(me, "Forest", "Basic Land — Forest")
	}
	rock := pushCatalogPermanent(g, o1.ID, "Rock", "Artifact", "", false)
	ench := pushCatalogPermanent(g, o2.ID, "Ench", "Enchantment", "", false)
	castCatalogSpell(t, g, "Disorienting Choice", "Sorcery", s49Disorienting, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: ench},
	})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, o1.ID, true)  // exiles the rock
	answerMayChoice(t, g, o2.ID, false) // keeps the enchantment
	// One chosen permanent is still on the battlefield: up to one land.
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no land search for the caster")
	}
	answerSearchByID(t, g, me.ID, c.SearchCards[0])
	lands, tapped := 0, 0
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.Controller == me.ID && card.IsLand() {
			lands++
			if card.Tapped {
				tapped++
			}
		}
	}
	if lands != 1 || tapped != 1 {
		t.Errorf("caster lands %d (tapped %d), want exactly 1 tapped land", lands, tapped)
	}
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.InstanceID == rock {
			t.Error("the exiled artifact is still on the battlefield")
		}
	}
}

func TestS49DisorientingChoiceAllExiledFindsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	b07LibraryLand(me, "Forest", "Basic Land — Forest")
	rock := pushCatalogPermanent(g, o1.ID, "Rock", "Artifact", "", false)
	castCatalogSpell(t, g, "Disorienting Choice", "Sorcery", s49Disorienting, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, o1.ID, true)
	if searchChoiceFor(g, me.ID) != nil {
		t.Error("nothing chosen survived, so no search")
	}
}

func TestS49StalkingLeoninExilesOnlyTheChosenPlayersAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	g.Turn.ActiveSeat = 0
	leo := pushCatalogPermanent(g, me.ID, "Stalking Leonin", "Creature — Cat Archer", s49Leonin, false)
	// The enters trigger is what asks; run it by hand through the prompt.
	g.WithWriteLock(func() {
		g.QueueSecretPlayerChoiceForEffect(me.ID, leo, "choose", []uuid.UUID{o1.ID, o2.ID, g.Seats[3].ID})
	})
	answerOptionPick(t, g, me.ID, 1) // o2
	if w, by := g.SecretChosenPlayerOf(leo); w != o2.ID || by != me.ID {
		t.Fatalf("secret choice = %v by %v, want %v", w, by, o2.ID)
	}
	// Only the chooser may reveal.
	if g.CanRevealSecretChoiceForEffect(o1.ID, leo) {
		t.Error("an opponent could reveal the choice")
	}
	a1 := pushCatalogPermanent(g, o1.ID, "Raider1", "Creature — Bear", "", false)
	a2 := pushCatalogPermanent(g, o2.ID, "Raider2", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case a1, a2:
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	// A creature of a player who was not chosen: the activation resolves and does nothing.
	pr7Activate(t, g, me.ID, leo, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a1}}})
	if countOnBattlefield(g, "Raider1", o1.ID) != 1 {
		t.Error("a creature not controlled by the chosen player was exiled")
	}
	// "Activate only once": spent.
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a2}}}); err == nil {
		t.Error("the ability activated a second time")
	}
}

func TestS49StalkingLeoninExilesTheChosenPlayersAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	leo := pushCatalogPermanent(g, me.ID, "Stalking Leonin", "Creature — Cat Archer", s49Leonin, false)
	g.WithWriteLock(func() {
		g.QueueSecretPlayerChoiceForEffect(me.ID, leo, "choose", []uuid.UUID{o1.ID})
	})
	answerOptionPick(t, g, me.ID, 0)
	a1 := pushCatalogPermanent(g, o1.ID, "Raider1", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a1 {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	pr7Activate(t, g, me.ID, leo, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a1}}})
	if countOnBattlefield(g, "Raider1", o1.ID) != 0 {
		t.Error("the chosen player's attacker was not exiled")
	}
}

func TestS49PrisonersDilemma(t *testing.T) {
	cases := []struct {
		name    string
		choices [3]string
		damage  [3]int
	}{
		{"all silence", [3]string{"silence", "silence", "silence"}, [3]int{4, 4, 4}},
		{"all snitch", [3]string{"snitch", "snitch", "snitch"}, [3]int{8, 8, 8}},
		{"mixed", [3]string{"silence", "snitch", "silence"}, [3]int{12, 0, 12}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			opps := g.Seats[1:]
			before := [3]int{opps[0].Life, opps[1].Life, opps[2].Life}
			meLife := me.Life
			castCatalogSpell(t, g, "Prisoner's Dilemma", "Sorcery", s49Dilemma, nil)
			passPriorityAroundTable(t, g)
			for i, o := range opps {
				if i > 0 {
					// Nobody else sees a ballot before the last is in.
					if ballotFor(g, o.ID) != nil {
						t.Fatalf("ballot for %s opened early", o.Name)
					}
				}
				voteFor(t, g, o.ID, tc.choices[i])
			}
			passPriorityAroundTable(t, g)
			for i, o := range opps {
				if o.Life != before[i]-tc.damage[i] {
					t.Errorf("%s life %d→%d, want -%d", o.Name, before[i], o.Life, tc.damage[i])
				}
			}
			if me.Life != meLife {
				t.Error("the caster is not asked and takes no damage")
			}
		})
	}
}
