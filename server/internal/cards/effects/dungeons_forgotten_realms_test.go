package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dungeons_forgotten_realms_test.go pins the three Adventures in the
// Forgotten Realms dungeons (ADR 0203 amendment of 2026-10-19, #4100):
// a plain venture offers all three and never the Undercity, and one walk
// per dungeon checks every room's ability against its printed text.

// frVenture has `player` venture into the dungeon (the card-side
// primitive) and settles, answering the dungeon and room prompts with
// `picks` in order: a dungeon key for "choose a dungeon", a room key for
// "choose the next room", a game.TargetRef for a room's target. It
// fails the test if a pick was never asked for.
func frVenture(t *testing.T, g *game.Game, player uuid.UUID, picks ...any) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (VentureIntoTheDungeon{Player: player}).Apply(&Context{Game: g}); err != nil {
			t.Errorf("VentureIntoTheDungeon: %v", err)
		}
	})
	if left := frSettle(t, g, picks...); len(left) > 0 {
		t.Fatalf("picks %v were never asked for", left)
	}
}

// frSettle resolves the stack, answering trigger orders, scries (keep
// the order) and searches (find nothing) itself and the venture and
// target prompts from `picks`. It stops at any other prompt and returns
// the picks it never used.
func frSettle(t *testing.T, g *game.Game, picks ...any) []any {
	t.Helper()
	for i := 0; i < 128; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		if open != nil {
			var err error
			switch open.Kind {
			case game.PendingChoiceTriggerOrder:
				err = g.ResolveTriggerOrder(open.ID, open.Chooser, append([]uuid.UUID(nil), open.TriggerOrderIDs...))
			case game.PendingChoiceSearchLibrary:
				err = g.ResolveSearchLibrary(open.ID, open.Chooser, nil)
			case game.PendingChoiceScry:
				err = g.ResolveScry(open.ID, open.Chooser, nil, append([]uuid.UUID(nil), open.ScryCards...))
			case game.PendingChoiceOptionPick:
				if len(picks) == 0 {
					return picks
				}
				key, ok := picks[0].(string)
				if !ok {
					return picks
				}
				idx := -1
				for j, o := range open.PickOptions {
					if (o.Room != "" && o.Room == key) || (o.Room == "" && o.Dungeon == key) {
						idx = j
					}
				}
				if idx < 0 {
					t.Fatalf("%q not offered: %+v", key, open.PickOptions)
				}
				err = g.ResolveOptionPick(open.ID, open.Chooser, idx)
				picks = picks[1:]
			case game.PendingChoicePickTarget:
				if len(picks) == 0 {
					return picks
				}
				ref, ok := picks[0].(game.TargetRef)
				if !ok {
					return picks
				}
				err = g.ResolvePickTargets(open.ID, open.Chooser, []game.TargetRef{ref})
				picks = picks[1:]
			default:
				return picks
			}
			if err != nil {
				t.Fatalf("answering %s: %v", open.Kind, err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return picks
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatal("the stack did not settle")
	return picks
}

// frTokens counts `controller`'s tokens named `name`.
func frTokens(g *game.Game, controller uuid.UUID, name string) int {
	return monTokens(g, controller, name)
}

// frCard targets a card.
func frCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// frPushTop puts a card on top of `p`'s library.
func frPushTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: manaCost, Owner: p.ID, Controller: p.ID}
	p.Library.PushTop(c)
	return c.InstanceID
}

// TestAPlainVentureOffersTheThreeForgottenRealmsDungeons is CR 309.2a
// and the Undercity's "You can't enter this dungeon unless you 'venture
// into Undercity'": a plain venture asks among exactly Lost Mine of
// Phandelver, Dungeon of the Mad Mage and Tomb of Annihilation.
func TestAPlainVentureOffersTheThreeForgottenRealmsDungeons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := (VentureIntoTheDungeon{Player: me.ID}).Apply(&Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	c := latestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatalf("no dungeon choice")
	}
	got := map[string]bool{}
	for _, o := range c.PickOptions {
		if o.Room != "" {
			t.Errorf("a dungeon option names a room: %+v", o)
		}
		got[o.Dungeon] = true
	}
	want := []string{LostMineOfPhandelverKey, DungeonOfTheMadMageKey, TombOfAnnihilationKey}
	if len(got) != len(want) {
		t.Fatalf("offered %v, want %v", got, want)
	}
	for _, k := range want {
		if !got[k] {
			t.Errorf("%s not offered", k)
		}
	}
	if got[game.UndercityKey] {
		t.Errorf("a plain venture offered the Undercity")
	}
}

// TestLostMineOfPhandelverWalk: Cave Entrance, Goblin Lair (a 1/1 red
// Goblin), Dark Pool (each opponent loses 1, you gain 1) and the Temple
// of Dumathoin (draw); the Temple's ability leaves the stack and the
// dungeon is completed (CR 704.5t), so the next venture asks for a
// dungeon again.
func TestLostMineOfPhandelverWalk(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	frVenture(t, g, me.ID, LostMineOfPhandelverKey)
	if room := initRoom(g, me.ID); room != "cave-entrance" {
		t.Fatalf("room = %q, want cave-entrance", room)
	}
	frVenture(t, g, me.ID, "goblin-lair")
	if n := frTokens(g, me.ID, "Goblin"); n != 1 {
		t.Fatalf("Goblin Lair made %d Goblins, want 1", n)
	}
	myLife, theirLife := me.Life, them.Life
	frVenture(t, g, me.ID, "dark-pool")
	if me.Life != myLife+1 || them.Life != theirLife-1 {
		t.Fatalf("Dark Pool: my life %d->%d, theirs %d->%d", myLife, me.Life, theirLife, them.Life)
	}
	hand := me.Hand.Size()
	frVenture(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the Temple drew %d, want 1", me.Hand.Size()-hand)
	}
	if room := initRoom(g, me.ID); room != "" {
		t.Fatalf("not completed: marker in %q", room)
	}
	if got := initCompleted(g, me.ID); len(got) != 1 || got[0] != LostMineOfPhandelverKey {
		t.Fatalf("completed = %v", got)
	}
	frVenture(t, g, me.ID, TombOfAnnihilationKey)
	if room := initRoom(g, me.ID); room != "trapped-entry" {
		t.Fatalf("after completing, room = %q, want a new dungeon's first room", room)
	}
}

// TestLostMineStoreroomAndFungiCavern: Mine Tunnels makes a Treasure,
// Fungi Cavern gives the targeted creature -4/-0 until your next turn,
// and (in a second dungeon) Storeroom puts a +1/+1 counter on its target.
func TestLostMineStoreroomAndFungiCavern(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, them.ID, "Their Giant", 5, 5)
	frVenture(t, g, me.ID, LostMineOfPhandelverKey)
	frVenture(t, g, me.ID, "mine-tunnels")
	if n := frTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Mine Tunnels made %d Treasures", n)
	}
	frVenture(t, g, me.ID, "fungi-cavern", frCard(theirs))
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Fatalf("Fungi Cavern: their Giant has power %d, want 1", p)
	}
	if tough := effectiveToughness(t, g, theirs); tough != 5 {
		t.Fatalf("Fungi Cavern changed toughness to %d", tough)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	bear := pushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	frVenture(t, g2, me2.ID, LostMineOfPhandelverKey)
	frVenture(t, g2, me2.ID, "goblin-lair")
	frVenture(t, g2, me2.ID, "storeroom", frCard(bear))
	if got := findBattlefieldCardForTest(g2, bear).Counters["+1/+1"]; got != 1 {
		t.Fatalf("Storeroom put %d counters, want 1", got)
	}
}

// TestDungeonOfTheMadMageWalk walks Yawning Portal, Dungeon Level,
// Twisted Caverns, Lost Level, Runestone Caverns, Deep Mines and the Mad
// Wizard's Lair, then completes.
func TestDungeonOfTheMadMageWalk(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	life := me.Life
	frVenture(t, g, me.ID, DungeonOfTheMadMageKey)
	if me.Life != life+1 {
		t.Fatalf("Yawning Portal: life %d, want %d", me.Life, life+1)
	}
	frVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "dungeon-level" {
		t.Fatalf("room = %q", room)
	}
	frVenture(t, g, me.ID, "twisted-caverns", frCard(theirs))
	g.ReadSnapshot(func() {
		c := findBattlefieldCardForTest(g, theirs)
		if c.Effective().Restrictions&game.CantAttack == 0 {
			t.Errorf("Twisted Caverns: their Bear can still attack")
		}
	})
	frVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "lost-level" {
		t.Fatalf("room = %q", room)
	}

	// Runestone Caverns: the top two cards go to exile, playable by me
	// for as long as they stay there.
	a := frPushTop(me, "Exiled Land", "Land — Forest", "")
	b := frPushTop(me, "Exiled Spell", "Sorcery", "{1}{R}")
	frVenture(t, g, me.ID, "runestone-caverns")
	for _, id := range []uuid.UUID{a, b} {
		if z := initZone(g, id); z != game.ZoneExile {
			t.Fatalf("Runestone Caverns left %v in %v, want exile", id, z)
		}
		g.ReadSnapshot(func() {
			p := g.CastPermissionOnCardByIDForEffect(id)
			if p == nil || p.Player != me.ID || p.CastOnly {
				t.Errorf("Runestone Caverns: %v has no play permission for me: %+v", id, p)
			}
		})
	}
	frVenture(t, g, me.ID)
	if room := initRoom(g, me.ID); room != "deep-mines" {
		t.Fatalf("room = %q", room)
	}

	// The Mad Wizard's Lair: draw three, reveal them, offer a free cast
	// of one of them (never the land).
	land := frPushTop(me, "Drawn Land", "Land — Island", "")
	spell := frPushTop(me, "Drawn Spell", "Sorcery", "{5}{U}{U}")
	creature := frPushTop(me, "Drawn Creature", "Creature — Giant", "{4}{G}")
	hand := me.Hand.Size()
	frVenture(t, g, me.ID)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("the Lair drew %d, want 3", me.Hand.Size()-hand)
	}
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("the Lair offered no free cast")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[spell] || !offered[creature] || offered[land] || len(offered) != 2 {
		t.Fatalf("the Lair offered %v, want the spell and the creature", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	g.ReadSnapshot(func() {
		p := g.CastPermissionOnCardByIDForEffect(creature)
		if p == nil || p.Player != me.ID || p.Cost != "{0}" {
			t.Errorf("the Lair opened no free cast of the creature: %+v", p)
		}
	})
	frSettle(t, g)
	if got := initCompleted(g, me.ID); len(got) != 1 || got[0] != DungeonOfTheMadMageKey {
		t.Fatalf("completed = %v", got)
	}
}

// TestDungeonOfTheMadMageBazaarAndGraveyard: Goblin Bazaar makes a
// Treasure and Muiral's Graveyard two 1/1 black Skeletons.
func TestDungeonOfTheMadMageBazaarAndGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frVenture(t, g, me.ID, DungeonOfTheMadMageKey)
	frVenture(t, g, me.ID)
	frVenture(t, g, me.ID, "goblin-bazaar")
	if n := frTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Goblin Bazaar made %d Treasures", n)
	}
	frVenture(t, g, me.ID)
	frVenture(t, g, me.ID, "muirals-graveyard")
	if n := frTokens(g, me.ID, "Skeleton"); n != 2 {
		t.Fatalf("Muiral's Graveyard made %d Skeletons, want 2", n)
	}
}

// TestTombOfAnnihilationWalk: Trapped Entry costs every player 1; Veils
// of Fear asks each player in turn order whether to lose 2 or discard;
// Sandfall Cell whether to lose 2 or sacrifice a creature, artifact or
// land, and a player with nothing to give is not asked; the Cradle makes
// The Atropal; and the Tomb is completed.
func TestTombOfAnnihilationWalk(t *testing.T) {
	g := newCatalogGame(t)
	me, second := g.Seats[0], g.Seats[1]
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	frVenture(t, g, me.ID, TombOfAnnihilationKey)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-1 {
			t.Fatalf("Trapped Entry: %s lost %d, want 1", p.Name, lives[p.ID]-p.Life)
		}
		lives[p.ID] = p.Life
	}

	// Veils of Fear. Seats 2 and 3 have empty hands: they lose 2 without
	// being asked.
	g.Seats[2].Hand.Cards = nil
	g.Seats[3].Hand.Cards = nil
	myHand := me.Hand.Size()
	if left := frSettle(t, g); len(left) != 0 {
		t.Fatal("unexpected picks")
	}
	g.WithWriteLock(func() {
		if err := (VentureIntoTheDungeon{Player: me.ID}).Apply(&Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	frSettle(t, g, "veils-of-fear")
	if latestOptionPickFor(g, second.ID) != nil {
		t.Fatal("the second player was asked before the first answered")
	}
	answerOptionPick(t, g, me.ID, 1)
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	answerOptionPick(t, g, second.ID, 0)
	frSettle(t, g)
	if me.Life != lives[me.ID] || me.Hand.Size() != myHand-1 {
		t.Errorf("Veils of Fear: I discarded %d and lost %d, want 1 and 0", myHand-me.Hand.Size(), lives[me.ID]-me.Life)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("Veils of Fear: %s lost %d, want 2", p.Name, lives[p.ID]-p.Life)
		}
		lives[p.ID] = p.Life
	}

	// Sandfall Cell. I have a land to sacrifice; nobody else has
	// anything, so only I am asked.
	forest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := (VentureIntoTheDungeon{Player: me.ID}).Apply(&Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	frSettle(t, g)
	answerOptionPick(t, g, me.ID, 1)
	answerSacrifice(t, g, me.ID, forest)
	frSettle(t, g)
	if g.Battlefield.Contains(forest) {
		t.Errorf("Sandfall Cell: the Forest was not sacrificed")
	}
	if me.Life != lives[me.ID] {
		t.Errorf("Sandfall Cell: I lost life after sacrificing")
	}
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("Sandfall Cell: %s lost %d, want 2", p.Name, lives[p.ID]-p.Life)
		}
	}

	frVenture(t, g, me.ID)
	atropal := findBattlefieldByName(g, "The Atropal")
	if atropal == uuid.Nil {
		t.Fatalf("the Cradle made no Atropal")
	}
	c := findBattlefieldCardForTest(g, atropal)
	if c.TypeLine != "Token Legendary Creature — God Horror" || c.Power != 4 || c.Toughness != 4 {
		t.Errorf("The Atropal is %q %d/%d", c.TypeLine, c.Power, c.Toughness)
	}
	if !hasAbility(effectiveAbilities(t, g, atropal), "deathtouch") {
		t.Errorf("The Atropal has no deathtouch")
	}
	if got := initCompleted(g, me.ID); len(got) != 1 || got[0] != TombOfAnnihilationKey {
		t.Fatalf("completed = %v", got)
	}
}

// TestTombOfAnnihilationOubliette: discard a card and sacrifice a
// creature, an artifact and a land — three different permanents.
func TestTombOfAnnihilationOubliette(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := pushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	forest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	keep := pushPermanentForTest(g, me.ID, "Island", "", "Basic Land — Island")
	frVenture(t, g, me.ID, TombOfAnnihilationKey)
	hand := me.Hand.Size()
	frVenture(t, g, me.ID, "oubliette")
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	answerSacrifice(t, g, me.ID, bear)
	answerSacrifice(t, g, me.ID, rock)
	answerSacrifice(t, g, me.ID, forest)
	frSettle(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("Oubliette discarded %d, want 1", hand-me.Hand.Size())
	}
	for _, id := range []uuid.UUID{bear, rock, forest} {
		if g.Battlefield.Contains(id) {
			t.Errorf("Oubliette left %v on the battlefield", id)
		}
	}
	if !g.Battlefield.Contains(keep) {
		t.Errorf("Oubliette took a second land")
	}
	if room := initRoom(g, me.ID); room != "oubliette" {
		t.Fatalf("room = %q", room)
	}
}
