package protocol

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// legal_actions_test.go — ADR 0105 §9, sub-PR 1 (#1789): the
// legal_actions digest. Privacy first, then agreement with the
// uncapped list, then the ref join the client depends on, then size.

const oracleViviOrnitier = "0452be2a-e97a-4269-a9e3-b616265ecb2e"

// legalActionsFrameBudget is the ceiling ADR 0105 §5 commits to for
// the legal_actions field alone, on the same worst boards
// TestLegalMovesFrameBudget measures. If this fails, find out which
// entry grew; do not raise the constant.
const legalActionsFrameBudget = 4 * 1024

// --- privacy -------------------------------------------------------

// ownerOf maps every card instance on the table to the seat that owns
// or controls it, from the authoritative game rather than any view.
func ownersOf(g *game.Game) map[string][]uuid.UUID {
	out := map[string][]uuid.UUID{}
	add := func(z *game.Zone) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			out[c.InstanceID.String()] = []uuid.UUID{c.Owner, c.Controller}
		}
	}
	add(g.Battlefield)
	add(g.Stack)
	add(g.Exile)
	for _, p := range g.Seats {
		add(p.Hand)
		add(p.Library)
		add(p.Graveyard)
		add(p.Command)
	}
	return out
}

// blockersTable is busyTable in declare_blockers with one attacker
// pointed at the next seat: the one frame shape on which TWO seats owe
// a decision at once (the active seat holds priority; the defender
// owes a block declaration, #328).
func blockersTable(t *testing.T) (g *game.Game, active, defender *game.Player, attacker uuid.UUID) {
	t.Helper()
	g = busyTable(t, 3)
	active = g.Seats[g.Turn.ActiveSeat]
	defender = g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceTo(t, g, game.StepDeclareAttackers)
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == active.ID && c.IsCreature() && !c.Tapped && !game.HasSummoningSickness(c) {
			attacker = c.InstanceID
			break
		}
	}
	if attacker == uuid.Nil {
		t.Skip("no legal attacker in the fixture; combat shape covered by internal/legal")
	}
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	return g, active, defender, attacker
}

// TestLegalActionsOwnSeatOnly is the hidden-information gate. A digest
// names the cards in a hand that its seat could cast, exactly as the
// list does, so a seat must only ever receive its own — and every
// source in it must be one of that seat's own cards.
func TestLegalActionsOwnSeatOnly(t *testing.T) {
	t.Run("main phase", func(t *testing.T) {
		g := busyTable(t, 3)
		active := g.Seats[g.Turn.ActiveSeat]
		owners := ownersOf(g)
		for _, p := range g.Seats {
			v := ViewOfGameFor(g, p.ID.String())
			if p.ID == active.ID {
				if v.LegalActions == nil || len(v.LegalActions.Sources) == 0 {
					t.Fatal("active seat holding priority in its main phase got no digest")
				}
			} else if v.LegalActions != nil {
				t.Errorf("seat %s owes no decision but received a digest with %d sources", p.Name, len(v.LegalActions.Sources))
			}
			assertDigestIsOwn(t, v.LegalActions, p.ID, owners)
		}
	})
	t.Run("two seats deciding at once", func(t *testing.T) {
		g, active, defender, _ := blockersTable(t)
		owners := ownersOf(g)
		for _, p := range []*game.Player{active, defender} {
			v := ViewOfGameFor(g, p.ID.String())
			if v.LegalActions == nil {
				t.Fatalf("seat %s owes a decision in declare_blockers but got no digest", p.Name)
			}
			assertDigestIsOwn(t, v.LegalActions, p.ID, owners)
		}
		dv := ViewOfGameFor(g, defender.ID.String())
		if !slices.ContainsFunc(sourcesOf(dv.LegalActions), func(s *LegalSourceView) bool { return len(s.Blocks) > 0 }) {
			t.Error("the defender's digest names no block candidate")
		}
	})
}

func sourcesOf(d *LegalActionsView) []*LegalSourceView {
	if d == nil {
		return nil
	}
	out := make([]*LegalSourceView, 0, len(d.Sources))
	for _, s := range d.Sources {
		out = append(out, s)
	}
	return out
}

func assertDigestIsOwn(t *testing.T, d *LegalActionsView, seat uuid.UUID, owners map[string][]uuid.UUID) {
	t.Helper()
	if d == nil {
		return
	}
	for id := range d.Sources {
		who, ok := owners[id]
		if !ok {
			t.Errorf("digest for %s names %s, which is no card on the table", seat, id)
			continue
		}
		if !slices.Contains(who, seat) {
			t.Errorf("digest for %s names %s, a card owned by %s and controlled by %s", seat, id, who[0], who[1])
		}
	}
}

// TestLegalActionsNeverInTheRawView: the crash dump and the replay log
// marshal the UNFILTERED view, so it must carry no seat's digest.
func TestLegalActionsNeverInTheRawView(t *testing.T) {
	g, _, _, _ := blockersTable(t)
	raw := ViewOfGame(g)
	if raw.LegalActions != nil {
		t.Error("unfiltered view carries a digest; it must carry none")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "legal_actions") {
		t.Error("unfiltered view marshalled a legal_actions key")
	}
}

// TestLegalActionsSpectatorGetsNone: the admin view (""), a spectator
// (SpectatorViewerID, #1791) and a viewer who is no seat have no own
// seat, so there is no digest they are entitled to — on a frame where
// two seats have one.
func TestLegalActionsSpectatorGetsNone(t *testing.T) {
	g, _, _, _ := blockersTable(t)
	for name, viewer := range map[string]string{
		"admin":     "",
		"spectator": SpectatorViewerID,
		"unknown":   uuid.New().String(),
	} {
		v := ViewOfGameFor(g, viewer)
		if v.LegalActions != nil {
			t.Errorf("%s got a digest with %d sources", name, len(v.LegalActions.Sources))
		}
		if len(v.LegalMoves) != 0 {
			t.Errorf("%s got %d moves", name, len(v.LegalMoves))
		}
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(b), "legal_actions") {
			t.Errorf("%s's frame marshalled a legal_actions key", name)
		}
	}
}

// TestFilterViewForIsIdempotentOnLegalActions: re-filtering a filtered
// view for another seat must not hand that seat the first one's digest.
func TestFilterViewForIsIdempotentOnLegalActions(t *testing.T) {
	g := busyTable(t, 3)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]

	once := FilterViewFor(ViewOfGame(g), active.ID.String())
	if once.LegalActions == nil {
		t.Fatal("active seat got no digest on the first filter")
	}
	for _, viewer := range []string{other.ID.String(), active.ID.String(), SpectatorViewerID, ""} {
		if twice := FilterViewFor(once, viewer); twice.LegalActions != nil {
			t.Errorf("re-filtering for %q produced a digest", viewer)
		}
	}
}

// --- agreement with the uncapped list -------------------------------

// pathologicalTable is TestLegalMovesPathologicalBoardStaysCapped's
// board: enough crossed-product sources that the wire list degrades.
func pathologicalTable(t *testing.T) *game.Game {
	t.Helper()
	g := busyTable(t, 4)
	active := g.Seats[g.Turn.ActiveSeat]
	for range 3 {
		put(g.Battlefield, active, game.Card{
			Name: "Goblin Bombardment", TypeLine: "Enchantment",
			ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
		})
		put(active.Hand, active, game.Card{
			Name: "Lightning Bolt", TypeLine: "Instant",
			ManaCost: "{R}", OracleID: oracleLightningBolt,
		})
	}
	return g
}

// TestLegalActionsAgreeWithTheList holds the digest to the UNCAPPED
// enumeration it was folded from, in both directions:
//
//   - every source with a digested move has an entry, carrying that
//     move's kind, and its `moves` is the uncapped count;
//   - every ref, zone, face, attack target and blocked attacker in an
//     entry names at least one enumerated move of that source.
//
// On the pathological board it also shows the digest surviving the
// wire cap: some source has more moves in the digest than in the
// capped legal_moves the same frame carries.
func TestLegalActionsAgreeWithTheList(t *testing.T) {
	boards := map[string]func(*testing.T) *game.Game{
		"busy 3":       func(t *testing.T) *game.Game { return busyTable(t, 3) },
		"busy 10":      func(t *testing.T) *game.Game { return busyTable(t, 10) },
		"pathological": pathologicalTable,
	}
	for name, build := range boards {
		t.Run(name, func(t *testing.T) {
			g := build(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			v := ViewOfGameFor(g, seat.ID.String())
			all := legal.EnumerateFor(g, seat.ID)
			assertDigestAgrees(t, v.LegalActions, all)

			if name != "pathological" {
				return
			}
			if len(all) <= legalMovesWireCap {
				t.Fatalf("pathological board enumerates %d moves, not past the %d cap; the test proves nothing", len(all), legalMovesWireCap)
			}
			capped := map[string]int{}
			for _, m := range v.LegalMoves {
				capped[m.Source.String()]++
			}
			survived := false
			for id, e := range v.LegalActions.Sources {
				if e.Moves > capped[id] {
					survived = true
				}
			}
			if !survived {
				t.Error("no source has more moves in the digest than in the capped list; it was not built from the uncapped enumeration")
			}
		})
	}
	t.Run("declare blockers", func(t *testing.T) {
		g, active, defender, _ := blockersTable(t)
		for _, p := range []*game.Player{active, defender} {
			v := ViewOfGameFor(g, p.ID.String())
			assertDigestAgrees(t, v.LegalActions, legal.EnumerateFor(g, p.ID))
		}
	})
}

// assertDigestAgrees is TestLegalActionsAgreeWithTheList's check, one
// seat at a time. It reads each move's params generically, so it does
// not share digestLegalMoves's decoding.
func assertDigestAgrees(t *testing.T, d *LegalActionsView, all []legal.Move) {
	t.Helper()
	if d == nil {
		t.Fatal("seat owes a decision but got no digest")
	}
	type involvement struct {
		kinds []legal.Kind
		moves int
		// every string/number value in the params of this source's
		// moves, keyed by param name
		values map[string][]string
		blocks []string
	}
	want := map[string]*involvement{}
	get := func(id string) *involvement {
		if want[id] == nil {
			want[id] = &involvement{values: map[string][]string{}}
		}
		return want[id]
	}
	pass := false
	for _, m := range all {
		switch m.Kind {
		case legal.KindPass:
			pass = true
			continue
		case legal.KindChoice, legal.KindMulligan:
			continue
		}
		if m.Source == uuid.Nil {
			continue
		}
		var params map[string]any
		if len(m.Params) > 0 {
			if err := json.Unmarshal(m.Params, &params); err != nil {
				t.Fatalf("move %q has unreadable params: %v", m.Label, err)
			}
		}
		src := m.Source.String()
		inv := get(src)
		inv.moves++
		inv.kinds = append(inv.kinds, m.Kind)
		for k, val := range params {
			inv.values[k] = append(inv.values[k], fmt.Sprint(val))
		}
		if m.Type == legal.TypeDeclareBlockers {
			blocks, _ := params["blocks"].([]any)
			for _, raw := range blocks {
				b, _ := raw.(map[string]any)
				blocker, _ := b["blocker"].(string)
				attacker, _ := b["attacker"].(string)
				bi := get(blocker)
				if blocker != src {
					bi.moves++
					bi.kinds = append(bi.kinds, legal.KindBlock)
				}
				bi.blocks = append(bi.blocks, attacker)
			}
		}
		if m.Type == legal.TypeDeclareBlocker {
			inv.blocks = append(inv.blocks, fmt.Sprint(params["attacker"]))
		}
	}

	if d.Pass != pass {
		t.Errorf("digest pass = %v, the list's = %v", d.Pass, pass)
	}
	if len(d.Sources) != len(want) {
		t.Errorf("digest has %d sources, the uncapped list %d", len(d.Sources), len(want))
	}
	for id, inv := range want {
		e := d.Sources[id]
		if e == nil {
			t.Errorf("source %s has %d moves in the uncapped list and no digest entry", id, inv.moves)
			continue
		}
		if e.Moves != inv.moves {
			t.Errorf("source %s: digest says %d moves, the uncapped list has %d", id, e.Moves, inv.moves)
		}
		for _, k := range inv.kinds {
			if !slices.Contains(e.Kinds, k) {
				t.Errorf("source %s has a %s move the digest's kinds %v omit", id, k, e.Kinds)
			}
		}
		for _, k := range e.Kinds {
			if !slices.Contains(inv.kinds, k) {
				t.Errorf("source %s: digest kind %s names no enumerated move", id, k)
			}
		}
		named := func(field, param string, vals []string) {
			for _, s := range vals {
				if !slices.Contains(inv.values[param], s) {
					t.Errorf("source %s: %s %q names no enumerated move (params %s = %v)", id, field, s, param, inv.values[param])
				}
			}
		}
		named("abilities", "ref", e.Abilities)
		named("mana_abilities", "ref", e.ManaAbilities)
		named("special_actions", "kind", e.SpecialActions)
		named("zones", "from_zone", e.Zones)
		named("attack_targets", "target", e.AttackTargets)
		for _, f := range e.Faces {
			// face is omitempty on the wire: the front face is absent.
			if f == 0 {
				if len(inv.values["face"]) < inv.moves {
					continue
				}
			}
			if !slices.Contains(inv.values["face"], fmt.Sprint(f)) {
				t.Errorf("source %s: face %d names no enumerated move", id, f)
			}
		}
		for _, a := range e.Blocks {
			if !slices.Contains(inv.blocks, a) {
				t.Errorf("source %s: blocks %s names no enumerated block", id, a)
			}
		}
		for _, a := range inv.blocks {
			if !slices.Contains(e.Blocks, a) {
				t.Errorf("source %s may block %s but the digest omits it", id, a)
			}
		}
	}
}

// --- the ref join --------------------------------------------------

// knowTheTable marks what a real game would have marked as the cards
// arrived: every public-zone card known to every seat, and each hidden
// card known to its owner. busyTable's put() pushes cards straight
// into zones and skips that, so without it the fixture's cards reach
// every frame redacted, rows and all.
func knowTheTable(g *game.Game) {
	seats := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		seats = append(seats, p.ID)
	}
	for _, z := range []*game.Zone{g.Battlefield, g.Stack, g.Exile} {
		for i := range z.Cards {
			z.Cards[i].AddKnowersAll(seats)
		}
	}
	for _, p := range g.Seats {
		for _, z := range []*game.Zone{p.Graveyard, p.Command} {
			for i := range z.Cards {
				z.Cards[i].AddKnowersAll(seats)
			}
		}
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(p.ID)
		}
	}
}

// cardInFrame finds a card in any zone of a filtered frame.
func cardInFrame(v GameView, id string) *CardView {
	zones := []*ZoneView{&v.Battlefield, &v.Stack, &v.Exile, &v.PhasedOut}
	for i := range v.Seats {
		s := &v.Seats[i]
		zones = append(zones, &s.Hand, &s.Library, &s.Graveyard, &s.Command)
	}
	for _, z := range zones {
		for i := range z.Cards {
			if z.Cards[i].InstanceID == id {
				return &z.Cards[i]
			}
		}
	}
	return nil
}

// TestLegalActionRefsMatchTheRows: the client joins an ability row to
// the digest by `ref`, so every ref the digest names must be the ref of
// a row on that card in the same frame — an activated ref on
// activated_abilities or zone_abilities, a mana ref on mana_abilities
// or zone_mana_abilities.
func TestLegalActionRefsMatchTheRows(t *testing.T) {
	g := pathologicalTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	put(g.Battlefield, seat, game.Card{
		Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
		ManaCost: "{1}{U}{R}", Power: 2, Toughness: 3, OracleID: oracleViviOrnitier,
	})
	knowTheTable(g)
	v := ViewOfGameFor(g, seat.ID.String())
	if v.LegalActions == nil {
		t.Fatal("no digest")
	}
	var activated, mana int
	for id, e := range v.LegalActions.Sources {
		if len(e.Abilities) == 0 && len(e.ManaAbilities) == 0 {
			continue
		}
		c := cardInFrame(v, id)
		if c == nil {
			t.Errorf("digest source %s is not in the seat's own frame", id)
			continue
		}
		var rows, manaRows []string
		for _, r := range slices.Concat(c.ActivatedAbilities, c.ZoneAbilities) {
			rows = append(rows, r.Ref)
		}
		for _, r := range slices.Concat(c.ManaAbilities, c.ZoneManaAbilities) {
			manaRows = append(manaRows, r.Ref)
		}
		for _, ref := range e.Abilities {
			activated++
			if !slices.Contains(rows, ref) {
				t.Errorf("%s: ability ref %q is on no activated row %v", c.Name, ref, rows)
			}
		}
		for _, ref := range e.ManaAbilities {
			mana++
			if !slices.Contains(manaRows, ref) {
				t.Errorf("%s: mana ref %q is on no mana row %v", c.Name, ref, manaRows)
			}
		}
	}
	if activated == 0 || mana == 0 {
		t.Errorf("the fixture exercised %d activated and %d mana refs; it needs both", activated, mana)
	}
}

// TestLegalActionsMarkVivisManaAbility is #1621's case, the one that
// started ADR 0105: Vivi's free, once-per-turn mana ability is in the
// controller's digest on their own turn with priority, and not on an
// opponent's turn, and not once it has been used this turn.
func TestLegalActionsMarkVivisManaAbility(t *testing.T) {
	viviRef := func(t *testing.T, g *game.Game, seat *game.Player, vivi uuid.UUID) (string, bool) {
		t.Helper()
		v := ViewOfGameFor(g, seat.ID.String())
		if v.LegalActions == nil {
			t.Fatalf("seat %s holds priority but got no digest", seat.Name)
		}
		c := cardInFrame(v, vivi.String())
		if c == nil || len(c.ManaAbilities) == 0 {
			t.Fatal("Vivi is not on the battlefield with a mana row")
		}
		e := v.LegalActions.Sources[vivi.String()]
		return c.ManaAbilities[0].Ref, e != nil && slices.Contains(e.ManaAbilities, c.ManaAbilities[0].Ref)
	}

	t.Run("own turn, then used", func(t *testing.T) {
		g := busyTable(t, 3)
		me := g.Seats[g.Turn.ActiveSeat]
		vivi := put(g.Battlefield, me, game.Card{
			Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
			ManaCost: "{1}{U}{R}", Power: 1, Toughness: 3, OracleID: oracleViviOrnitier,
		})
		knowTheTable(g)
		if ref, ok := viviRef(t, g, me, vivi); !ok {
			t.Fatalf("Vivi's mana ref %q is not in its controller's digest on their own turn", ref)
		}
		if err := g.ActivateManaAbility(me.ID, vivi, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
				if err := g.ResolveManaChoice(c.ID, me.ID, "R"); err != nil {
					t.Fatalf("ResolveManaChoice: %v", err)
				}
			}
		}
		if ref, ok := viviRef(t, g, me, vivi); ok {
			t.Errorf("Vivi's mana ref %q is still in the digest after it was used this turn", ref)
		}
	})

	t.Run("opponent's turn", func(t *testing.T) {
		g := busyTable(t, 3)
		them := g.Seats[(g.Turn.ActiveSeat+1)%4]
		vivi := put(g.Battlefield, them, game.Card{
			Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
			ManaCost: "{1}{U}{R}", Power: 1, Toughness: 3, OracleID: oracleViviOrnitier,
		})
		knowTheTable(g)
		// The active seat passes, so priority reaches Vivi's controller
		// on somebody else's turn.
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Seats[g.Turn.PriorityHolder].ID != them.ID {
			t.Skip("priority did not pass to the next seat in this fixture")
		}
		if ref, ok := viviRef(t, g, them, vivi); ok {
			t.Errorf("Vivi's mana ref %q is in the digest on an opponent's turn", ref)
		}
	})
}

// --- digestLegalMoves, directly -------------------------------------

// A grouped block (#750: the two creatures a menace attacker takes) is
// one move whose Source is its first blocker. Both creatures are block
// candidates, so both get an entry.
func TestDigestCountsEveryBlockerInAGroup(t *testing.T) {
	first, second, atk := uuid.New(), uuid.New(), uuid.New()
	params, err := json.Marshal(map[string]any{"blocks": []map[string]string{
		{"blocker": first.String(), "attacker": atk.String()},
		{"blocker": second.String(), "attacker": atk.String()},
	}})
	if err != nil {
		t.Fatal(err)
	}
	d := digestLegalMoves([]legal.Move{
		{Type: legal.TypePassPriority, Kind: legal.KindPass, Label: "Pass"},
		{Type: legal.TypeDeclareBlockers, Kind: legal.KindBlock, Label: "Block with both", Source: first, Params: params},
	})
	if d == nil || !d.Pass {
		t.Fatalf("digest = %+v, want pass and two sources", d)
	}
	for _, id := range []uuid.UUID{first, second} {
		e := d.Sources[id.String()]
		if e == nil {
			t.Fatalf("blocker %s has no entry", id)
		}
		if e.Moves != 1 || !slices.Equal(e.Kinds, []legal.Kind{legal.KindBlock}) || !slices.Equal(e.Blocks, []string{atk.String()}) {
			t.Errorf("blocker %s: %+v, want one block move against %s", id, e, atk)
		}
	}
}

// Choice and mulligan answers are not digested (their surfaces read
// pending_choices), so a seat that owes only a choice gets no digest.
func TestDigestOfOnlyChoicesIsNil(t *testing.T) {
	d := digestLegalMoves([]legal.Move{
		{Type: legal.TypeResolveChoice, Kind: legal.KindChoice, Label: "Pick", Source: uuid.New()},
		{Type: legal.TypeKeepHand, Kind: legal.KindMulligan, Label: "Keep"},
	})
	if d != nil {
		t.Errorf("digest = %+v, want nil", d)
	}
	if digestLegalMoves(nil) != nil {
		t.Error("an empty list produced a digest")
	}
}

// The digest marshals its sources in enumeration order, not sorted-key
// order, so two identical boards with different random instance IDs
// marshal to the same bytes once the IDs are replaced (the agreement
// fixtures rely on it). And it still decodes as an ordinary object.
func TestLegalActionsMarshalInEnumerationOrder(t *testing.T) {
	ids := []uuid.UUID{
		uuid.MustParse("ffffffff-0000-4000-8000-000000000000"),
		uuid.MustParse("00000000-0000-4000-8000-00000000000a"),
		uuid.MustParse("88888888-0000-4000-8000-000000000000"),
	}
	moves := []legal.Move{{Type: legal.TypePassPriority, Kind: legal.KindPass, Label: "Pass"}}
	for _, id := range ids {
		params, err := json.Marshal(map[string]any{"instance_id": id.String(), "from_zone": "hand"})
		if err != nil {
			t.Fatal(err)
		}
		moves = append(moves, legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindCast, Label: "Cast", Source: id, Params: params})
	}
	d := digestLegalMoves(moves)
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	last := -1
	for _, id := range ids {
		at := strings.Index(string(b), id.String())
		if at < last {
			t.Fatalf("sources are not in enumeration order: %s", b)
		}
		last = at
	}
	var back LegalActionsView
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !back.Pass || len(back.Sources) != len(ids) {
		t.Fatalf("round trip lost data: %s -> %+v", b, back)
	}
	// A decoded view has no recorded order and falls back to sorted
	// keys; it must still marshal every source.
	again, err := json.Marshal(back)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	for _, id := range ids {
		if !strings.Contains(string(again), id.String()) {
			t.Errorf("re-marshalled view dropped %s: %s", id, again)
		}
	}
}

// --- size ----------------------------------------------------------

// TestLegalActionsFrameBudget measures the digest's wire cost on the
// worst four-player frames and fails past ADR 0105 §5's 4 KiB. The
// numbers are logged (go test -v) so a change to the enumerator shows
// its trend here.
func TestLegalActionsFrameBudget(t *testing.T) {
	boards := []struct {
		name  string
		build func(*testing.T) *game.Game
	}{
		{"creatures=3", func(t *testing.T) *game.Game { return busyTable(t, 3) }},
		{"creatures=6", func(t *testing.T) *game.Game { return busyTable(t, 6) }},
		{"creatures=10", func(t *testing.T) *game.Game { return busyTable(t, 10) }},
		{"pathological", pathologicalTable},
	}
	for _, b := range boards {
		t.Run(b.name, func(t *testing.T) {
			g := b.build(t)
			active := g.Seats[g.Turn.ActiveSeat]
			with := ViewOfGameFor(g, active.ID.String())
			if with.LegalActions == nil {
				t.Fatal("no digest on a priority frame")
			}
			without := with
			without.LegalActions = nil
			full, err := json.Marshal(with)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			base, err := json.Marshal(without)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			growth := len(full) - len(base)
			t.Logf("%s | frame %d B -> %d B | legal_actions %d B across %d sources (legal_moves: %d moves)",
				b.name, len(base), len(full), growth, len(with.LegalActions.Sources), len(with.LegalMoves))
			if growth > legalActionsFrameBudget {
				t.Errorf("legal_actions cost %d B, over the %d B budget", growth, legalActionsFrameBudget)
			}
		})
	}
}

// TestLegalActionsQuietFrameIsFree: on a frame where the viewer owes no
// decision the field costs nothing.
func TestLegalActionsQuietFrameIsFree(t *testing.T) {
	g := busyTable(t, 6)
	idle := g.Seats[(g.Turn.ActiveSeat+1)%4]
	v := ViewOfGameFor(g, idle.ID.String())
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "legal_actions") {
		t.Error("an idle seat's frame carries a legal_actions key; omitempty should have dropped it")
	}
}
