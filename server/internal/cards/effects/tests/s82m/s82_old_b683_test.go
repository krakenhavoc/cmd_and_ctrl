package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b683_test.go — Mercadian Masques, "Creatures with keywords", batch
// 683 (#5619). Every test plays the real catalog entry through the engine.

const (
	o683Guard     = "0c3b3c04-017d-4741-90c8-4b6c0bd46e49"
	o683Fiend     = "10736ea3-6253-44c3-8a1e-3b9a7f38f3cf"
	o683Badger    = "1f5ac983-a7a5-4b9d-a25c-727a7de6c465"
	o683Sprite    = "283433cb-9466-4b40-8af9-30934654d43b"
	o683Delraich  = "38f9e766-a2ad-437b-a38c-d6f3ffd261fb"
	o683Runner    = "48a77c0c-a179-47f2-83b2-da80ba5448f9"
	o683Smugglers = "5041dbc4-d01d-4f0a-b5c5-6e1df0976340"
	o683Irregs    = "595cbe23-4969-4fc2-8386-e5f3e6adfd5f"
	o683Spirit    = "6470afc0-2f67-454a-8a5b-d8dfea4b84ce"
	o683Bolt      = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestS82OldB683EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o683Guard, o683Fiend, o683Badger, o683Sprite, o683Delraich, o683Runner, o683Smugglers, o683Irregs, o683Spirit} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestS82OldB683PrintedKeywordsAreOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		keywords               []string
	}{
		{"Steadfast Guard", "Creature — Human Rebel", o683Guard, 2, 2, []string{"vigilance"}},
		{"Misshapen Fiend", "Creature — Horror Mercenary", o683Fiend, 1, 1, []string{"flying"}},
		{"Rock Badger", "Creature — Badger Beast", o683Badger, 3, 3, []string{"mountainwalk"}},
		{"Cloud Sprite", "Creature — Faerie", o683Sprite, 1, 1, []string{"flying"}},
		{"Delraich", "Creature — Horror", o683Delraich, 6, 6, []string{"trample"}},
		{"Lava Runner", "Creature — Lizard", o683Runner, 2, 2, []string{"haste"}},
		{"Bog Smugglers", "Creature — Human Mercenary", o683Smugglers, 2, 2, []string{"swampwalk"}},
		{"Gerrard's Irregulars", "Creature — Human Soldier", o683Irregs, 4, 2, []string{"trample", "haste"}},
		{"Extravagant Spirit", "Creature — Spirit", o683Spirit, 4, 4, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tough)
		for _, kw := range tc.keywords {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

func TestS82OldB683SteadfastGuardAttacksWithoutTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Steadfast Guard", "Creature — Human Rebel", o683Guard, 2, 2)
	effectstest.AttackWith(t, g, opp.ID, guard)
	if c, _ := effectstest.BattlefieldCard(g, guard); c.Tapped {
		t.Error("the attacking Guard tapped")
	}
}

func TestS82OldB683LandwalkersCantBeBlockedByADefenderWithTheLand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, land, landType string
	}{
		{"Rock Badger", "Creature — Badger Beast", o683Badger, "Mountain", "Basic Land — Mountain"},
		{"Bog Smugglers", "Creature — Human Mercenary", o683Smugglers, "Swamp", "Basic Land — Swamp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			walker := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
			blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
			effectstest.AttackIntoBlocks(t, g, opp.ID, walker)
			if reason := effectstest.BlockRefusal(t, g, blocker, walker); reason != "" {
				t.Fatalf("blocked refused without the land: %v", reason)
			}
			effectstest.SeedLand(g, opp.ID, tc.land, tc.landType, "")
			if reason := effectstest.BlockRefusal(t, g, blocker, walker); reason == "" {
				t.Errorf("a defender with a %s blocked %s", tc.land, tc.name)
			}
		})
	}
}

func TestS82OldB683CloudSpriteBlocksOnlyFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sprite := effectstest.B12Push(g, opp.ID, "Cloud Sprite", "Creature — Faerie", o683Sprite, 1, 1)
	ground := effectstest.B12Creature(g, me.ID, "Ground", "Creature — Bear", 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	effectstest.AttackIntoBlocks(t, g, opp.ID, ground, flier)
	if reason := effectstest.BlockRefusal(t, g, sprite, ground); reason == "" {
		t.Error("Cloud Sprite blocked a creature without flying")
	}
	if reason := effectstest.BlockRefusal(t, g, sprite, flier); reason != "" {
		t.Errorf("Cloud Sprite could not block a flier: %v", reason)
	}
}

func TestS82OldB683DelraichSacrificesThreeBlackCreaturesInsteadOfItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	del := effectstest.HandCardAtMain(t, g, "Delraich", "Creature — Horror", o683Delraich)
	black := []uuid.UUID{
		effectstest.SeedColoredCreature(g, me.ID, "Zombie A", []string{"B"}, 0),
		effectstest.SeedColoredCreature(g, me.ID, "Zombie B", []string{"B"}, 0),
	}
	green := effectstest.SeedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	if err := effectstest.CastForAltCost(g, del, "", "sacrifice", append(append([]uuid.UUID(nil), black...), green), nil); err == nil {
		t.Fatal("a green creature paid \"three black creatures\"")
	}
	for _, id := range append(append([]uuid.UUID(nil), black...), green) {
		if !g.Battlefield.Contains(id) {
			t.Fatal("the refused cast sacrificed a creature anyway")
		}
	}
	black = append(black, effectstest.SeedColoredCreature(g, me.ID, "Zombie C", []string{"B"}, 0))
	life := me.Life
	if err := effectstest.CastForAltCost(g, del, "", "sacrifice", black, nil); err != nil {
		t.Fatalf("Delraich for three black creatures: %v", err)
	}
	if me.Life != life {
		t.Errorf("life %d → %d, the cost has no life payment", life, me.Life)
	}
	for _, id := range black {
		if g.Battlefield.Contains(id) {
			t.Error("a black creature was not sacrificed")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, del) == nil {
		t.Error("Delraich did not resolve onto the battlefield")
	}
}

func TestS82OldB683LavaRunnerMakesTheTargetingPlayerSacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, opp.ID, "Lava Runner", "Creature — Lizard", o683Runner, 2, 2)
	keep := effectstest.SeedLand(g, me.ID, "Forest", "Basic Land — Forest", "")
	lose := effectstest.SeedLand(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	theirs := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", o683Bolt, effectstest.CardRefs(runner))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) {
		t.Error("the chosen land was not sacrificed")
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(theirs) {
		t.Error("a land that was not chosen left the battlefield")
	}
}

func o683SpiritUpkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Extravagant Spirit", "Creature — Spirit", o683Spirit, 4, 4)
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			me.Hand.Cards = me.Hand.Cards[:0]
		}
	})
	for i := 0; i < 2; i++ {
		effectstest.PushHandCardOfType(me, "Card", "Sorcery")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, spirit
}

func TestS82OldB683ExtravagantSpiritStaysWhenYouPayOnePerCardInHand(t *testing.T) {
	g, me, spirit := o683SpiritUpkeep(t)
	effectstest.B06AddMana(me, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spirit) {
		t.Error("the Spirit should stay after paying {2} for two cards")
	}
}

func TestS82OldB683ExtravagantSpiritIsSacrificedWhenYouDontPay(t *testing.T) {
	g, me, spirit := o683SpiritUpkeep(t)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spirit) {
		t.Error("the Spirit stayed without payment")
	}
}

func TestS82OldB683ExtravagantSpiritWithAnEmptyHandIsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spirit := effectstest.B12Push(g, me.ID, "Extravagant Spirit", "Creature — Spirit", o683Spirit, 4, 4)
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:0] })
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spirit) {
		t.Error("the Spirit was sacrificed with no cards in hand")
	}
}
