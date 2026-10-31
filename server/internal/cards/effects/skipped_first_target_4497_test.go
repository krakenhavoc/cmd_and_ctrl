package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// skipped_first_target_4497_test.go — #4497: a spell or ability whose
// first target clause is "up to one", cast with that clause left empty.
//
// CR 601.2c: a player announces how many targets a variable count gets
// before choosing them, and "up to one" allows none. Each target chosen
// answers one instance of the word "target", so the pick for the next
// clause belongs to that clause.
// Meager Meal ("Put a +1/+1 counter on up to one target creature.
// Target player gains 2 life.") cast with no creature was refused
// unless its player ref said slot 1: the server's fill for a list that
// names no clause filed the player under the creature clause.
//
// The browser client stamps every pick with its clause (targeting.ts
// `stamped`), and so do the bot enumerator and the MCP seat (omitting a
// zero). The tests here pin the wire shape a client builds from the
// view, a stamped pick read as given, and an unstamped list filed by
// what each pick can target.

// meagerMealInHand seats Gollum, Silent Slinker // Meager Meal in the
// active player's hand, known to them so the view projects its faces.
func meagerMealInHand(t *testing.T, g *game.Game) (*game.Player, uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: gollumSilentSlinkerOracleID, Layout: game.LayoutAdventure,
		Owner: me.ID, Controller: me.ID, Faces: []game.Face{hobA6GollumCreature, hobA6MeagerMeal},
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	}
	c.SetFace(0)
	me.Hand.PushTop(c)
	return me, c.InstanceID
}

// dispatchCast sends a cast_spell exactly as a client would, through the
// wire decoder.
func dispatchCast(g *game.Game, me *game.Player, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return actions.Dispatch(g, actions.Action{Type: actions.TypeCastSpell, Player: me.ID, Caller: me.ID, Params: raw})
}

// stackSlots is the slot each target of a stack item was filed under, by
// target ID.
func stackSlots(t *testing.T, g *game.Game, itemID uuid.UUID) map[uuid.UUID]int {
	t.Helper()
	item := g.StackMeta[itemID]
	if item == nil {
		t.Fatalf("no stack item %s", itemID)
	}
	out := make(map[uuid.UUID]int, len(item.Targets))
	for _, r := range item.Targets {
		out[r.ID] = r.Slot
	}
	return out
}

func listHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// The protocol half. The view names Meager Meal's two clauses, the first
// one optional; a client that skips it sends the player stamped with the
// clause it answered, and the cast is accepted and resolves.
func TestMeagerMealCastOverTheWireWithItsCreatureSkipped(t *testing.T) {
	g := newCatalogGame(t)
	me, id := meagerMealInHand(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := seedCreature(g, "Bear", me.ID)
	before := opp.Life

	view := protocol.ViewOfGameFor(g, me.ID.String())
	var card *protocol.CardView
	for i := range view.Seats {
		for j := range view.Seats[i].Hand.Cards {
			if c := &view.Seats[i].Hand.Cards[j]; c.InstanceID == id.String() {
				card = c
			}
		}
	}
	if card == nil || len(card.Faces) != 2 {
		t.Fatalf("the hand view does not carry Meager Meal's faces: %+v", card)
	}
	clauses := card.Faces[1].Clauses
	if len(clauses) != 2 {
		t.Fatalf("Meager Meal's face ships %d clauses, want 2", len(clauses))
	}
	if clauses[0].Min != 0 || clauses[0].Max != 1 || !listHas(clauses[0].Cards, bear.String()) {
		t.Errorf("clause 0 = %+v, want up to one creature with the bear in it", clauses[0])
	}
	if clauses[1].Min != 1 || !listHas(clauses[1].Players, opp.ID.String()) {
		t.Errorf("clause 1 = %+v, want one player with the opponent in it", clauses[1])
	}

	// What the client's walk sends with clause 0 left empty: the one pick,
	// stamped with clause 1's coordinates (targeting.ts `stamped`).
	err := dispatchCast(g, me, map[string]any{
		"instance_id": id.String(),
		"face":        1,
		"targets":     []map[string]any{{"kind": "player", "id": opp.ID.String(), "mode": 0, "slot": 1}},
	})
	if err != nil {
		t.Fatalf("cast with the creature skipped: %v", err)
	}
	if got := stackSlots(t, g, id); len(got) != 1 || got[opp.ID] != 1 {
		t.Fatalf("the player was filed under %v, want slot 1", got)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before+2 {
		t.Errorf("target player's life %d -> %d, want +2", before, opp.Life)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the unchosen creature has %d +1/+1 counters, want 0", n)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// A payload that names no clause (an older client, a hand-written test)
// is filed by what each pick can target: the player can only answer
// "target player". Both clauses answered stay in printed order, and a
// pick no clause takes is refused as before.
func TestMeagerMealUnstampedPicksAreFiledByWhatTheyCanTarget(t *testing.T) {
	t.Run("the creature skipped", func(t *testing.T) {
		g := newCatalogGame(t)
		me, id := meagerMealInHand(t, g)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		before := opp.Life
		err := dispatchCast(g, me, map[string]any{
			"instance_id": id.String(), "face": 1,
			"targets": []map[string]any{{"kind": "player", "id": opp.ID.String()}},
		})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		if got := stackSlots(t, g, id); got[opp.ID] != 1 {
			t.Fatalf("the player was filed under slot %d, want 1", got[opp.ID])
		}
		passPriorityAroundTable(t, g)
		if opp.Life != before+2 {
			t.Errorf("life %d -> %d, want +2", before, opp.Life)
		}
	})
	t.Run("both answered", func(t *testing.T) {
		g := newCatalogGame(t)
		me, id := meagerMealInHand(t, g)
		bear := seedCreature(g, "Bear", me.ID)
		err := dispatchCast(g, me, map[string]any{
			"instance_id": id.String(), "face": 1,
			"targets": []map[string]any{{"kind": "card", "id": bear.String()}, {"kind": "player", "id": me.ID.String()}},
		})
		if err != nil {
			t.Fatalf("cast: %v", err)
		}
		if got := stackSlots(t, g, id); got[bear] != 0 || got[me.ID] != 1 {
			t.Errorf("filed %v, want the bear at 0 and the player at 1", got)
		}
	})
	t.Run("a pick no clause takes", func(t *testing.T) {
		g := newCatalogGame(t)
		me, id := meagerMealInHand(t, g)
		rock := b12Permanent(g, me.ID, "Rock", "Artifact")
		err := dispatchCast(g, me, map[string]any{
			"instance_id": id.String(), "face": 1,
			"targets": []map[string]any{{"kind": "card", "id": rock.String()}},
		})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("an artifact for Meager Meal: err %v, want ErrIllegalTarget", err)
		}
		if !me.Hand.Contains(id) {
			t.Error("a refused cast left the hand")
		}
	})
}

// A pick that names its clause is filed there and nowhere else. The
// player sent with `slot: 0` is the caller's word, so the creature
// clause refuses it; the server does not move it to the player clause.
// The same pick sent with no slot is filed by what it can target.
func TestMeagerMealAStampedPickIsNeverMoved(t *testing.T) {
	g := newCatalogGame(t)
	me, id := meagerMealInHand(t, g)
	err := dispatchCast(g, me, map[string]any{
		"instance_id": id.String(), "face": 1,
		"targets": []map[string]any{{"kind": "player", "id": me.ID.String(), "mode": 0, "slot": 0}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a player stamped slot 0: err %v, want ErrIllegalTarget", err)
	}
	if !me.Hand.Contains(id) {
		t.Fatal("a refused cast left the hand")
	}
	if err := dispatchCast(g, me, map[string]any{
		"instance_id": id.String(), "face": 1,
		"targets": []map[string]any{{"kind": "player", "id": me.ID.String()}},
	}); err != nil {
		t.Fatalf("the same player with no slot: %v", err)
	}
	if got := stackSlots(t, g, id); got[me.ID] != 1 {
		t.Errorf("filed under slot %d, want 1", got[me.ID])
	}
}

// Every catalog statement whose clause before the last can take fewer
// than its maximum has the same shape, so the fix is the server's and
// not Meager Meal's. One case per kind: an "up to one" first clause
// skipped (River's Grasp, Stunning Shot), the first of the per-opponent
// clauses skipped (Windgrace's Judgment), an "up to two" first clause
// half filled (Cease) and a "one or two" one (Terrific Team-Up). No pick
// names its clause.
func TestAnUnstampedPickPassesOverAnOptionalClause(t *testing.T) {
	type built struct {
		card    uuid.UUID
		targets []game.TargetRef
		want    map[uuid.UUID]int
		mana    string // paid from the pool, for a card that reads it
	}
	player := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }
	card := func(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
	for _, tc := range []struct {
		name  string
		build func(g *game.Game, me, opp, opp2 *game.Player) built
	}{
		{"River's Grasp, the creature skipped", func(g *game.Game, me, opp, _ *game.Player) built {
			id := handCardOf(g, me, "River's Grasp", "Sorcery", "{3}{U/B}", "f671583f-bd22-4d3e-bf19-fb740c607f8f")
			seedCreature(g, "Bear", opp.ID)
			return built{id, []game.TargetRef{player(opp.ID)}, map[uuid.UUID]int{opp.ID: 1}, "{U}{B}{C}{C}"}
		}},
		{"Stunning Shot, your creature skipped", func(g *game.Game, me, opp, _ *game.Player) built {
			id := handCardOf(g, me, "Stunning Shot", "Instant", "{1}{W}", "0c0bb9b0-e7e0-465e-826b-2e56032aed1b")
			seedCreature(g, "Mine", me.ID)
			theirs := seedCreature(g, "Theirs", opp.ID)
			return built{id, []game.TargetRef{card(theirs)}, map[uuid.UUID]int{theirs: 1}, ""}
		}},
		{"Windgrace's Judgment, the first opponent skipped", func(g *game.Game, me, _, opp2 *game.Player) built {
			id := handCardOf(g, me, "Windgrace's Judgment", "Instant", "{3}{B}{G}", "de1ca6ed-b275-4f62-ba05-f31b3659b352")
			rock := b12Permanent(g, opp2.ID, "Their Rock", "Artifact")
			return built{id, []game.TargetRef{card(rock)}, map[uuid.UUID]int{rock: 1}, ""}
		}},
		{"Cease, one card of up to two", func(g *game.Game, me, opp, _ *game.Player) built {
			id := handCardOf(g, me, "Cease", "Instant", "{B/G}", "ba4d644f-1931-4fc4-aed5-681a476a5a58")
			dead := pushGraveyardCard(g, opp, "Their Card")
			return built{id, []game.TargetRef{card(dead), player(me.ID)}, map[uuid.UUID]int{dead: 0, me.ID: 1}, ""}
		}},
		{"Terrific Team-Up, one creature of one or two", func(g *game.Game, me, opp, _ *game.Player) built {
			id := handCardOf(g, me, "Terrific Team-Up", "Instant", "{3}{G}", "431c5b04-7387-4724-adfa-d1eb7874e79e")
			mine := seedCreature(g, "Mine", me.ID)
			theirs := seedCreature(g, "Theirs", opp.ID)
			return built{id, []game.TargetRef{card(mine), card(theirs)}, map[uuid.UUID]int{mine: 0, theirs: 1}, ""}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			n := len(g.Seats)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%n]
			opp2 := g.Seats[(g.Turn.ActiveSeat+2)%n]
			b := tc.build(g, me, opp, opp2)
			params := game.CastSpellParams{Targets: b.targets}
			if b.mana != "" {
				if err := g.AddManaForEffect(me.ID, uuid.Nil, b.mana); err != nil {
					t.Fatalf("AddManaForEffect: %v", err)
				}
				params.Strict = true
			}
			if err := g.CastSpell(me.ID, b.card, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			got := stackSlots(t, g, b.card)
			for id, slot := range b.want {
				if s, ok := got[id]; !ok || s != slot {
					t.Errorf("%s filed under slot %d (present %v), want %d", id, s, ok, slot)
				}
			}
		})
	}
}

// The activation path files the same way: Teferi, Who Slows the
// Sunset's +1 with only its third clause, "up to one target land",
// answered.
func TestAnUnstampedActivationPickPassesOverOptionalClauses(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	teferi := pushCatalogWalker(g, me.ID, "Teferi, Who Slows the Sunset", teferiWhoSlowsTheSunsetOracle, 4)
	land := b12Permanent(g, me.ID, "My Forest", "Land")
	tapForTest(t, g, land)

	if err := g.ActivateCatalogAbility(me.ID, teferi, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate the +1 with only a land: %v", err)
	}
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it.SourceCardID == teferi {
			item = it
		}
	}
	if item == nil || len(item.Targets) != 1 || item.Targets[0].Slot != 2 {
		t.Fatalf("the land was filed as %+v, want slot 2", item)
	}
	settleAnsweringProliferate(t, g)
	if b16Tapped(t, g, land) {
		t.Error("the chosen land should have untapped")
	}
}
