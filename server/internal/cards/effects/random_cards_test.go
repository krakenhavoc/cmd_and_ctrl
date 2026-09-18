package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func randomEvents(g *game.Game, kind game.EventKind) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}
func randomTokenCount(g *game.Game, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name {
			n++
		}
	}
	return n
}
func TestRandomDragonsAndOgre(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, token string
		attack              bool
	}{
		{"Ancient Copper Dragon", "48daee9d-ddaf-410f-8c3a-12fa1064ab56", "Treasure", false},
		{"Ancient Gold Dragon", "44cb725a-72fc-4e0d-b966-f0b6d33f6b79", "Faerie Dragon", false},
		{"Hoarding Ogre", "faeee3a7-c66c-48d8-a994-12377d3bb729", "Treasure", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushCatalogPermanent(g, me.ID, tc.name, "Creature", tc.oracle, false)
			ev := game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: id, Target: g.Seats[1].ID, Amount: 1, Combat: true}
			if tc.attack {
				ev = game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Target: g.Seats[1].ID}
			}
			g.WithWriteLock(func() { g.EmitEvent(ev) })
			if len(randomEvents(g, game.EventRollDie)) != 0 {
				t.Fatal("rolled before the trigger resolved")
			}
			passPriorityAroundTable(t, g)
			rolls := randomEvents(g, game.EventRollDie)
			if len(rolls) != 1 || rolls[0].Sides != 20 {
				t.Fatalf("rolls: %+v", rolls)
			}
			want := rolls[0].Amount
			if tc.attack {
				want = 1
				if rolls[0].Amount >= 10 {
					want = 2
				}
				if rolls[0].Amount == 20 {
					want = 3
				}
			}
			if got := randomTokenCount(g, tc.token); got != want {
				t.Fatalf("tokens=%d want %d", got, want)
			}
			if tc.token == "Faerie Dragon" {
				for _, c := range g.Battlefield.Cards {
					if c.IsToken() && (!hasEffectiveKeyword(t, g, c.InstanceID, "flying") || effectivePower(t, g, c.InstanceID) != 1) {
						t.Fatal("incorrect Faerie Dragon")
					}
				}
			}
		})
	}
}
func TestDeadbridgeChantEntryAndRandomReturn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Graveyard.Size()
	castCatalogSpell(t, g, "Deadbridge Chant", "Enchantment", "506667b1-7922-4959-a5b3-0f8abe8c3615", nil)
	for i := 0; i < 4; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if me.Graveyard.Size() != before || len(g.StackMeta) == 0 {
		t.Fatal("mill must wait on its ETB trigger")
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != before+10 {
		t.Fatalf("milled %d", me.Graveyard.Size()-before)
	}
	for _, creature := range []bool{false, true} {
		t.Run(map[bool]string{false: "hand", true: "battlefield"}[creature], func(t *testing.T) {
			me.Graveyard.Cards = nil
			c := game.NewCard("return me", me.ID)
			c.TypeLine = "Instant"
			if creature {
				c.TypeLine = "Creature"
				c.Power = 2
				c.Toughness = 2
			}
			me.Graveyard.PushTop(c)
			g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
			passPriorityAroundTable(t, g)
			if creature && !g.Battlefield.Contains(c.InstanceID) || !creature && !me.Hand.Contains(c.InstanceID) {
				t.Fatal("wrong return destination")
			}
		})
	}
}
func TestExaltedFlamerBothAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Exalted Flamer of Tzeentch", "Creature", "6591a589-deb4-462f-a4f1-f6261f3433c3", false)
	creature := b17GraveyardCard(me, "excluded", "Creature", "")
	spell := b17GraveyardCard(me, "spell", "Instant", "")
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID}) })
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) || !me.Graveyard.Contains(creature) {
		t.Fatal("return must filter to instants and sorceries")
	}
	lives := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	castCatalogSpell(t, g, "uncataloged instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	for i, want := range lives {
		if g.Seats[i+1].Life != want-1 {
			t.Fatal("cast did not damage each opponent")
		}
	}
}
func TestGoldSaucerCoinAndSacrificeAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "The Gold Saucer", "Land — Town", "93e38650-ce22-4ab9-b79d-cc7b6477c075", false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Fatal("missing colorless mana")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceCoinCall {
		t.Fatal("missing coin call")
	}
	if err := g.ResolveCoinCall(g.PendingChoices[0].ID, me.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	flips := randomEvents(g, game.EventFlipCoin)
	if len(flips) != 1 {
		t.Fatalf("flips=%v", flips)
	}
	want := 0
	if flips[0].Won {
		want = 1
	}
	if randomTokenCount(g, "Treasure") != want {
		t.Fatal("incorrect win reward")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
	a := pushToken(g, me.ID, TreasureToken())
	b := pushToken(g, me.ID, TreasureToken())
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Fatal("sacrifice/draw failed")
	}
}
func TestPuzzleboxDistinctBatchesManaAndTutor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Vexing Puzzlebox", "Artifact", "7267ab16-2157-4b86-93ea-ca2c0fee064a", false)
	total := 0
	g.WithWriteLock(func() {
		for _, n := range []int{2, 1} {
			rolls, err := g.RollDiceForEffect(game.RandomDraw{Player: me.ID, Source: id}, 6, n)
			if err != nil {
				t.Fatal(err)
			}
			for _, v := range rolls {
				total += v
			}
		}
	})
	if len(g.PendingTriggers) != 2 {
		t.Fatalf("two instructions must queue two triggers; got %d", len(g.PendingTriggers))
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, id, "charge") != total {
		t.Fatal("incorrect batch total")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("missing mana color choice")
	}
	if err := g.ResolveManaChoice(g.PendingChoices[0].ID, me.ID, "U"); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	rolls := randomEvents(g, game.EventRollDie)
	if len(rolls) != 4 || rolls[3].Sides != 20 || len(me.ManaPool) != 1 {
		t.Fatal("mana rider failed")
	}
	if counterCount(g, id, "charge") != total+rolls[3].Amount {
		t.Fatal("mana rider charge missing")
	}
	needed := 100 - counterCount(g, id, "charge")
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id); _ = g.AddCounterForEffect(id, "charge", needed) })
	artifact := game.NewCard("tutor me", me.ID)
	artifact.TypeLine = "Artifact"
	me.Library.PushBottom(artifact)
	otherArtifact := game.NewCard("other artifact", me.ID)
	otherArtifact.TypeLine = "Artifact"
	me.Library.PushBottom(otherArtifact)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	if counterCount(g, id, "charge") != 0 {
		t.Fatal("counters not paid")
	}
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatal("missing artifact search")
	}
	if err := g.ResolveSearchLibrary(g.PendingChoices[0].ID, me.ID, []uuid.UUID{artifact.InstanceID}); err != nil {
		t.Fatal(err)
	}
	if !g.Battlefield.Contains(artifact.InstanceID) {
		t.Fatal("artifact not on battlefield")
	}
}
func TestRecklessEndeavorBothResultAssignments(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(map[bool]string{false: "second", true: "first"}[accept], func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			creature := b12Push(g, g.Seats[1].ID, "survivor", "Creature", "", 1, 30)
			castCatalogSpell(t, g, "Reckless Endeavor", "Sorcery", "928bf0d9-89be-462d-aeb7-74a65e95535c", nil)
			passPriorityAroundTable(t, g)
			rolls := randomEvents(g, game.EventRollDie)
			if len(rolls) != 2 || rolls[0].Sides != 12 || rolls[0].BatchSeq != rolls[1].BatchSeq {
				t.Fatalf("rolls: %+v", rolls)
			}
			damage, tokens := rolls[0].Amount, rolls[1].Amount
			if len(g.PendingChoices) > 0 {
				if err := g.ResolveConfirm(g.PendingChoices[0].ID, me.ID, accept); err != nil {
					t.Fatal(err)
				}
				if !accept {
					damage, tokens = tokens, damage
				}
			}
			c, _ := g.LookupCardForEffect(creature)
			if c.DamageMarked != damage || randomTokenCount(g, "Treasure") != tokens {
				t.Fatalf("damage=%d tokens=%d, want %d/%d", c.DamageMarked, randomTokenCount(g, "Treasure"), damage, tokens)
			}
		})
	}
}
