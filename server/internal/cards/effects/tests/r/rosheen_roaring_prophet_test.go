package testsr

import (
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// Rosheen, Roaring Prophet (#6085): the mill-then-take on entering, and
// the mana ability that reveals any number of cards with {X} in their
// mana cost (CR 605.3b, CR 701.20a/b) for {C}{C} each, spendable only on
// costs that contain {X} (CR 106.6).

const rosheenOracle = "3108a4fe-b18a-4d02-9545-b19812454362"

func TestRosheenIsFull(t *testing.T) {
	spec, ok := effects.Lookup(rosheenOracle)
	if !ok || spec.Name != "Rosheen, Roaring Prophet" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered=%v spec=%q/%s/%v", ok, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func rosheenOnBattlefield(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := effectstest.FlipMain(t)
	effectstest.EmptyHand(g, me)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Rosheen, Roaring Prophet", "Legendary Creature — Giant Shaman", rosheenOracle, false)
	effectstest.Unsick(g, id)
	return g, me, opp, id
}

func rosheenHand(me *game.Player, name, cost string) uuid.UUID {
	return effectstest.AvB1Hand(me, game.Card{Name: name, TypeLine: "Sorcery", ManaCost: cost})
}

func rosheenPoolC(me *game.Player) (n int, restricted bool) {
	restricted = true
	for _, tok := range me.ManaPool {
		if tok.Color != "C" {
			continue
		}
		n += tok.N()
		if !slices.Contains(tok.Restrictions, game.ManaRestrictXCost) {
			restricted = false
		}
	}
	return n, restricted
}

// Revealing two of the three {X} cards adds {C}{C}{C}{C}, each spendable
// only on a cost with {X}; the cards stay in the hand, now known to the
// table (CR 701.20b).
func TestRosheenAddsTwoColorlessForEachCardRevealed(t *testing.T) {
	g, me, opp, id := rosheenOnBattlefield(t)
	fireball := rosheenHand(me, "Fireball", "{X}{R}")
	blaze := rosheenHand(me, "Blaze", "{X}{R}")
	rosheenHand(me, "Stream", "{X}{U}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{fireball, blaze}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	n, restricted := rosheenPoolC(me)
	if n != 4 || !restricted {
		t.Fatalf("pool has %d {C} (restricted=%v), want 4 restricted to costs with {X}", n, restricted)
	}
	if me.Hand.Size() != 3 {
		t.Fatalf("hand %d, want the three cards still there", me.Hand.Size())
	}
	g.ReadSnapshot(func() {
		for _, c := range me.Hand.Cards {
			known := c.IsKnownTo(opp.ID)
			if (c.InstanceID == fireball || c.InstanceID == blaze) != known {
				t.Errorf("%s known to the opponent = %v", c.Name, known)
			}
		}
	})
}

// Revealing nothing is a legal answer and adds no mana; the source still
// taps (the cost is paid).
func TestRosheenRevealingNothingAddsNothing(t *testing.T) {
	g, me, _, id := rosheenOnBattlefield(t)
	rosheenHand(me, "Fireball", "{X}{R}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n, _ := rosheenPoolC(me); n != 0 {
		t.Fatalf("pool has %d {C}, want none", n)
	}
	if !effectstest.BattlefieldCardPtr(t, g, id).Tapped {
		t.Fatal("Rosheen did not tap")
	}
}

// A card without {X} in its mana cost cannot be revealed, nor can one card
// be named twice; each refusal taps nothing.
func TestRosheenRefusesCardsWithoutXAndRepeats(t *testing.T) {
	g, me, _, id := rosheenOnBattlefield(t)
	fireball := rosheenHand(me, "Fireball", "{X}{R}")
	shock := rosheenHand(me, "Shock", "{R}")
	for name, ids := range map[string][]uuid.UUID{
		"a card without {X}":  {shock},
		"the same card twice": {fireball, fireball},
	} {
		err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: ids})
		if !errors.Is(err, game.ErrInvalidParam) {
			t.Errorf("%s: err %v, want ErrInvalidParam", name, err)
		}
		if effectstest.BattlefieldCardPtr(t, g, id).Tapped {
			t.Fatalf("%s: Rosheen tapped on a refused activation", name)
		}
	}
}

// The mana is not spent on a cost without {X} (CR 106.6).
func TestRosheenManaCannotPayACostWithoutX(t *testing.T) {
	g, me, _, id := rosheenOnBattlefield(t)
	fireball := rosheenHand(me, "Fireball", "{X}{R}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{fireball}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	cost, err := game.ParseCost("{2}")
	if err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendContext{Purpose: game.SpendPurposeCast}) {
		t.Fatal("Rosheen's mana paid a cost with no {X}")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendContext{Purpose: game.SpendPurposeCast, CostHasX: true}) {
		t.Fatal("Rosheen's mana did not pay a cost with {X}")
	}
}

// When Rosheen enters, six cards are milled and the player may take one
// with {X} in its mana cost from among them.
func TestRosheenEntersMillsSixAndMayTakeAnXCard(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.PlTop(me, "Deep", "Instant", "{X}{U}")
	var milled []uuid.UUID
	for range 4 {
		milled = append(milled, effectstest.PlTop(me, "Filler", "Instant", "{U}"))
	}
	xSpell := effectstest.PlTop(me, "Milled X", "Sorcery", "{X}{G}")
	land := effectstest.PlTop(me, "Milled Land", "Basic Land — Forest", "")
	lib := me.Library.Size()
	effectstest.CastFromHandForTest(t, g, me, "Rosheen, Roaring Prophet", "Legendary Creature — Giant Shaman", "{0}", rosheenOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-6 {
		t.Fatalf("library %d -> %d, want six milled", lib, me.Library.Size())
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card with {X} was offered")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != xSpell {
		t.Fatalf("offered %v, want only the milled {X} card (not the land %v nor the fillers %v)", pick.ChooseCards, land, milled)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, xSpell)
	if !effectstest.MiInHand(me, xSpell) {
		t.Fatal("the {X} card did not reach the hand")
	}
}

// The legal-move enumerator offers Rosheen's ability once, revealing
// every matching card (the most mana it can make), and the move it offers
// is one the engine accepts. The auto-tapper never plans it: which cards
// to show is the player's decision.
func TestRosheenMoveRevealsEveryXCardAndIsNotAutoTapped(t *testing.T) {
	g, me, _, id := rosheenOnBattlefield(t)
	fireball := rosheenHand(me, "Fireball", "{X}{R}")
	blaze := rosheenHand(me, "Blaze", "{X}{R}")
	rosheenHand(me, "Shock", "{R}")
	var params struct {
		RevealIDs []string `json:"reveal_ids"`
	}
	found := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeActivateManaAbility || m.Source != id {
			continue
		}
		found++
		if err := json.Unmarshal(m.Params, &params); err != nil {
			t.Fatal(err)
		}
	}
	if found != 1 {
		t.Fatalf("%d Rosheen mana moves, want 1", found)
	}
	want := []string{fireball.String(), blaze.String()}
	if !slices.Equal(slices.Sorted(slices.Values(params.RevealIDs)), slices.Sorted(slices.Values(want))) {
		t.Fatalf("move reveals %v, want %v", params.RevealIDs, want)
	}
	cost, _ := game.ParseCost("{2}")
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok || slices.Contains(plan, id) {
		t.Fatalf("the auto-tapper planned Rosheen: %v", plan)
	}
}
