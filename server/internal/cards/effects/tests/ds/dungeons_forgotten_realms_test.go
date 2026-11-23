package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestAPlainVentureOffersTheThreeForgottenRealmsDungeons is CR 309.2a
// and the Undercity's "You can't enter this dungeon unless you 'venture
// into Undercity'": a plain venture asks among exactly Lost Mine of
// Phandelver, Dungeon of the Mad Mage and Tomb of Annihilation.
func TestAPlainVentureOffersTheThreeForgottenRealmsDungeons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := (effects.VentureIntoTheDungeon{Player: me.ID}).Apply(&effects.Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	c := effectstest.LatestOptionPickFor(g, me.ID)
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
	want := []string{effects.LostMineOfPhandelverKey, effects.DungeonOfTheMadMageKey, effects.TombOfAnnihilationKey}
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.FrVenture(t, g, me.ID, effects.LostMineOfPhandelverKey)
	if room := effectstest.InitRoom(g, me.ID); room != "cave-entrance" {
		t.Fatalf("room = %q, want cave-entrance", room)
	}
	effectstest.FrVenture(t, g, me.ID, "goblin-lair")
	if n := effectstest.FrTokens(g, me.ID, "Goblin"); n != 1 {
		t.Fatalf("Goblin Lair made %d Goblins, want 1", n)
	}
	myLife, theirLife := me.Life, them.Life
	effectstest.FrVenture(t, g, me.ID, "dark-pool")
	if me.Life != myLife+1 || them.Life != theirLife-1 {
		t.Fatalf("Dark Pool: my life %d->%d, theirs %d->%d", myLife, me.Life, theirLife, them.Life)
	}
	hand := me.Hand.Size()
	effectstest.FrVenture(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("the Temple drew %d, want 1", me.Hand.Size()-hand)
	}
	if room := effectstest.InitRoom(g, me.ID); room != "" {
		t.Fatalf("not completed: marker in %q", room)
	}
	if got := effectstest.InitCompleted(g, me.ID); len(got) != 1 || got[0] != effects.LostMineOfPhandelverKey {
		t.Fatalf("completed = %v", got)
	}
	effectstest.FrVenture(t, g, me.ID, effects.TombOfAnnihilationKey)
	if room := effectstest.InitRoom(g, me.ID); room != "trapped-entry" {
		t.Fatalf("after completing, room = %q, want a new dungeon's first room", room)
	}
}

// TestLostMineStoreroomAndFungiCavern: Mine Tunnels makes a Treasure,
// Fungi Cavern gives the targeted creature -4/-0 until your next turn,
// and (in a second dungeon) Storeroom puts a +1/+1 counter on its target.
func TestLostMineStoreroomAndFungiCavern(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Giant", 5, 5)
	effectstest.FrVenture(t, g, me.ID, effects.LostMineOfPhandelverKey)
	effectstest.FrVenture(t, g, me.ID, "mine-tunnels")
	if n := effectstest.FrTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Mine Tunnels made %d Treasures", n)
	}
	effectstest.FrVenture(t, g, me.ID, "fungi-cavern", effectstest.FrCard(theirs))
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Fatalf("Fungi Cavern: their Giant has power %d, want 1", p)
	}
	if tough := effectstest.EffectiveToughness(t, g, theirs); tough != 5 {
		t.Fatalf("Fungi Cavern changed toughness to %d", tough)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	bear := effectstest.PushVanillaCreature(g2, me2.ID, "Bear", 2, 2)
	effectstest.FrVenture(t, g2, me2.ID, effects.LostMineOfPhandelverKey)
	effectstest.FrVenture(t, g2, me2.ID, "goblin-lair")
	effectstest.FrVenture(t, g2, me2.ID, "storeroom", effectstest.FrCard(bear))
	if got := effectstest.FindBattlefieldCardForTest(g2, bear).Counters["+1/+1"]; got != 1 {
		t.Fatalf("Storeroom put %d counters, want 1", got)
	}
}

// TestDungeonOfTheMadMageWalk walks Yawning Portal, Dungeon Level,
// Twisted Caverns, Lost Level, Runestone Caverns, Deep Mines and the Mad
// Wizard's Lair, then completes.
func TestDungeonOfTheMadMageWalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	life := me.Life
	effectstest.FrVenture(t, g, me.ID, effects.DungeonOfTheMadMageKey)
	if me.Life != life+1 {
		t.Fatalf("Yawning Portal: life %d, want %d", me.Life, life+1)
	}
	effectstest.FrVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "dungeon-level" {
		t.Fatalf("room = %q", room)
	}
	effectstest.FrVenture(t, g, me.ID, "twisted-caverns", effectstest.FrCard(theirs))
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, theirs)
		if c.Effective().Restrictions&game.CantAttack == 0 {
			t.Errorf("Twisted Caverns: their Bear can still attack")
		}
	})
	effectstest.FrVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "lost-level" {
		t.Fatalf("room = %q", room)
	}

	// Runestone Caverns: the top two cards go to exile, playable by me
	// for as long as they stay there.
	a := effectstest.FrPushTop(me, "Exiled Land", "Land — Forest", "")
	b := effectstest.FrPushTop(me, "Exiled Spell", "Sorcery", "{1}{R}")
	effectstest.FrVenture(t, g, me.ID, "runestone-caverns")
	for _, id := range []uuid.UUID{a, b} {
		if z := effectstest.InitZone(g, id); z != game.ZoneExile {
			t.Fatalf("Runestone Caverns left %v in %v, want exile", id, z)
		}
		g.ReadSnapshot(func() {
			p := g.CastPermissionOnCardByIDForEffect(id)
			if p == nil || p.Player != me.ID || p.CastOnly {
				t.Errorf("Runestone Caverns: %v has no play permission for me: %+v", id, p)
			}
		})
	}
	effectstest.FrVenture(t, g, me.ID)
	if room := effectstest.InitRoom(g, me.ID); room != "deep-mines" {
		t.Fatalf("room = %q", room)
	}

	// The Mad Wizard's Lair: draw three, reveal them, offer a free cast
	// of one of them (never the land).
	land := effectstest.FrPushTop(me, "Drawn Land", "Land — Island", "")
	spell := effectstest.FrPushTop(me, "Drawn Spell", "Sorcery", "{5}{U}{U}")
	creature := effectstest.FrPushTop(me, "Drawn Creature", "Creature — Giant", "{4}{G}")
	hand := me.Hand.Size()
	effectstest.FrVenture(t, g, me.ID)
	if me.Hand.Size() != hand+3 {
		t.Fatalf("the Lair drew %d, want 3", me.Hand.Size()-hand)
	}
	// The offer is a resolution_cast inside the room's resolution (#4250):
	// the table waiting on it is a restore point, and the rest of the walk
	// runs on the restored table.
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	offered := map[uuid.UUID]bool{}
	for _, id := range c.DiscardOptions {
		offered[id] = true
	}
	if !offered[spell] || !offered[creature] || offered[land] || len(offered) != 2 {
		t.Fatalf("the Lair offered %v, want the spell and the creature", c.DiscardOptions)
	}
	g.ReadSnapshot(func() {
		p := g.CastPermissionOnCardByIDForEffect(creature)
		if p == nil || p.Player != me.ID || p.Cost != "{0}" {
			t.Errorf("the Lair opened no free cast of the creature: %+v", p)
		}
	})
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the Lair's free cast: %v", err)
	}
	if !g.Stack.Contains(creature) || effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("the Lair's one free cast did not reach the stack and close the offer")
	}
	effectstest.FrSettle(t, g)
	if got := effectstest.InitCompleted(g, me.ID); len(got) != 1 || got[0] != effects.DungeonOfTheMadMageKey {
		t.Fatalf("completed = %v", got)
	}
}

// TestDungeonOfTheMadMageBazaarAndGraveyard: Goblin Bazaar makes a
// Treasure and Muiral's Graveyard two 1/1 black Skeletons.
func TestDungeonOfTheMadMageBazaarAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.FrVenture(t, g, me.ID, effects.DungeonOfTheMadMageKey)
	effectstest.FrVenture(t, g, me.ID)
	effectstest.FrVenture(t, g, me.ID, "goblin-bazaar")
	if n := effectstest.FrTokens(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Goblin Bazaar made %d Treasures", n)
	}
	effectstest.FrVenture(t, g, me.ID)
	effectstest.FrVenture(t, g, me.ID, "muirals-graveyard")
	if n := effectstest.FrTokens(g, me.ID, "Skeleton"); n != 2 {
		t.Fatalf("Muiral's Graveyard made %d Skeletons, want 2", n)
	}
}

// TestTombOfAnnihilationWalk: Trapped Entry costs every player 1; Veils
// of Fear asks each player in turn order whether to lose 2 or discard;
// Sandfall Cell whether to lose 2 or sacrifice a creature, artifact or
// land, and a player with nothing to give is not asked; the Cradle makes
// The Atropal; and the Tomb is completed.
func TestTombOfAnnihilationWalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, second := g.Seats[0], g.Seats[1]
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.FrVenture(t, g, me.ID, effects.TombOfAnnihilationKey)
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
	if left := effectstest.FrSettle(t, g); len(left) != 0 {
		t.Fatal("unexpected picks")
	}
	g.WithWriteLock(func() {
		if err := (effects.VentureIntoTheDungeon{Player: me.ID}).Apply(&effects.Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.FrSettle(t, g, "veils-of-fear")
	if effectstest.LatestOptionPickFor(g, second.ID) != nil {
		t.Fatal("the second player was asked before the first answered")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.AnswerOptionPick(t, g, second.ID, 0)
	effectstest.FrSettle(t, g)
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
	forest := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	g.WithWriteLock(func() {
		if err := (effects.VentureIntoTheDungeon{Player: me.ID}).Apply(&effects.Context{Game: g}); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.FrSettle(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerSacrifice(t, g, me.ID, forest)
	effectstest.FrSettle(t, g)
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

	effectstest.FrVenture(t, g, me.ID)
	atropal := effectstest.FindBattlefieldByName(g, "The Atropal")
	if atropal == uuid.Nil {
		t.Fatalf("the Cradle made no Atropal")
	}
	c := effectstest.FindBattlefieldCardForTest(g, atropal)
	if c.TypeLine != "Token Legendary Creature — God Horror" || c.Power != 4 || c.Toughness != 4 {
		t.Errorf("The Atropal is %q %d/%d", c.TypeLine, c.Power, c.Toughness)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, atropal), "deathtouch") {
		t.Errorf("The Atropal has no deathtouch")
	}
	if got := effectstest.InitCompleted(g, me.ID); len(got) != 1 || got[0] != effects.TombOfAnnihilationKey {
		t.Fatalf("completed = %v", got)
	}
}

// TestTombOfAnnihilationOubliette: discard a card and sacrifice a
// creature, an artifact and a land — three different permanents.
func TestTombOfAnnihilationOubliette(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	rock := effectstest.PushPermanentForTest(g, me.ID, "Mind Stone", "", "Artifact")
	forest := effectstest.PushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	keep := effectstest.PushPermanentForTest(g, me.ID, "Island", "", "Basic Land — Island")
	effectstest.FrVenture(t, g, me.ID, effects.TombOfAnnihilationKey)
	hand := me.Hand.Size()
	effectstest.FrVenture(t, g, me.ID, "oubliette")
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.AnswerSacrifice(t, g, me.ID, bear)
	effectstest.AnswerSacrifice(t, g, me.ID, rock)
	effectstest.AnswerSacrifice(t, g, me.ID, forest)
	effectstest.FrSettle(t, g)
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
	if room := effectstest.InitRoom(g, me.ID); room != "oubliette" {
		t.Fatalf("room = %q", room)
	}
}
