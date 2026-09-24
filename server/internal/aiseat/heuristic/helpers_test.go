package heuristic_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// helpers_test.go builds protocol.GameViews and legal.Move lists by
// hand.
//
// That this is possible at all is the point of ADR 0033 §3: a Policy's
// entire input is a wire view and a move list, so the policy can be
// tested against a board that never existed in an engine. No game, no
// room, no goroutines — which means these tests run in microseconds
// and say exactly what went wrong when they fail. The engine-level
// tests live next door in aiseat, where they belong.

func seatID(i int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", i))
}

func cardID(i int) string {
	return fmt.Sprintf("10000000-0000-0000-0000-%012d", i)
}

type seatOpt func(*protocol.PlayerView)

func withLife(n int) seatOpt { return func(p *protocol.PlayerView) { p.Life = n } }

func withHand(cards ...protocol.CardView) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Hand = protocol.ZoneView{Kind: "hand", Owner: p.ID, Count: len(cards), Cards: cards}
	}
}

func withHandCount(n int) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Hand = protocol.ZoneView{Kind: "hand", Owner: p.ID, Count: n}
	}
}

func withCommanderCasts(n int) seatOpt {
	return func(p *protocol.PlayerView) {
		p.CommanderCasts = map[string]int{cardID(999): n}
	}
}

func withEliminated() seatOpt { return func(p *protocol.PlayerView) { p.Eliminated = true } }

func newSeat(i int, opts ...seatOpt) protocol.PlayerView {
	p := protocol.PlayerView{
		ID:          seatID(i).String(),
		Name:        fmt.Sprintf("Seat%d", i),
		Seat:        i,
		Life:        40,
		Library:     protocol.ZoneView{Kind: "library", Count: 60},
		Hand:        protocol.ZoneView{Kind: "hand"},
		Graveyard:   protocol.ZoneView{Kind: "graveyard"},
		Command:     protocol.ZoneView{Kind: "command"},
		MaxHandSize: 7,
	}
	for _, o := range opts {
		o(&p)
	}
	return p
}

type cardOpt func(*protocol.CardView)

func tapped() cardOpt    { return func(c *protocol.CardView) { c.Tapped = true } }
func sick() cardOpt      { return func(c *protocol.CardView) { c.SummoningSick = true } }
func commander() cardOpt { return func(c *protocol.CardView) { c.IsCommander = true } }

func keywords(kw ...string) cardOpt {
	return func(c *protocol.CardView) { c.Abilities = append(c.Abilities, kw...) }
}

// protection sets the PARSED protection the server projects onto a
// permanent (#662). It is deliberately separate from keywords(): the
// raw "protection from red" token rides Abilities like every other
// keyword, but nothing in a policy package may parse it — a policy
// package cannot import internal/game at all (ADR 0033 §3) — so the
// bot reads this list and a fixture has to set it the way the wire
// would.
func protection(qualities ...protocol.ProtectionView) cardOpt {
	return func(c *protocol.CardView) { c.Protection = append(c.Protection, qualities...) }
}

// proColor is the projected "protection from <colour>" for a wire
// colour letter, for fixtures.
func proColor(letter, printed string) protocol.ProtectionView {
	return protocol.ProtectionView{Printed: printed, Kind: "color", Value: letter}
}

func attacking(seat int) cardOpt {
	return func(c *protocol.CardView) { c.AttackingTarget = seatID(seat).String() }
}

func blocking(id string) cardOpt {
	return func(c *protocol.CardView) { c.BlockingTarget = id }
}

func creature(id string, controller int, name string, power, tough int, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(controller).String(),
		Controller: seatID(controller).String(),
		TypeLine:   "Creature — Bear",
		ManaCost:   "{1}{R}",
		Power:      power,
		Toughness:  tough,
		KnownByYou: true,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func land(id string, controller int, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID:    id,
		Name:          "Mountain",
		Owner:         seatID(controller).String(),
		Controller:    seatID(controller).String(),
		TypeLine:      "Basic Land — Mountain",
		KnownByYou:    true,
		ManaAbilities: []protocol.ManaAbilityView{{Index: 0, TapCost: true, Produced: "{R}"}},
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func spell(id string, controller int, name, cost string) protocol.CardView {
	return protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(controller).String(),
		Controller: seatID(controller).String(),
		TypeLine:   "Instant",
		ManaCost:   cost,
		KnownByYou: true,
	}
}

type viewOpt func(*protocol.GameView)

func withBattlefield(cards ...protocol.CardView) viewOpt {
	return func(v *protocol.GameView) {
		v.Battlefield = protocol.ZoneView{Kind: "battlefield", Count: len(cards), Cards: cards}
	}
}

func withTurn(number, active int, step string) viewOpt {
	return func(v *protocol.GameView) {
		v.Turn = protocol.TurnView{Seq: number, Number: number, ActiveSeat: active, PriorityHolder: active, Step: step}
	}
}

func withStack(cards ...protocol.CardView) viewOpt {
	return func(v *protocol.GameView) {
		v.Stack = protocol.ZoneView{Kind: "stack", Count: len(cards), Cards: cards}
	}
}

func withChoice(c protocol.PendingChoiceView) viewOpt {
	return func(v *protocol.GameView) { v.PendingChoices = append(v.PendingChoices, c) }
}

func newView(seats []protocol.PlayerView, opts ...viewOpt) protocol.GameView {
	v := protocol.GameView{
		ID:          uuid.Nil.String(),
		State:       "active",
		Seats:       seats,
		Battlefield: protocol.ZoneView{Kind: "battlefield"},
		Stack:       protocol.ZoneView{Kind: "stack"},
		Exile:       protocol.ZoneView{Kind: "exile"},
		Turn:        protocol.TurnView{Seq: 1, Number: 1, Step: "precombat_main"},
	}
	for _, o := range opts {
		o(&v)
	}
	return v
}

// --- moves ---------------------------------------------------------

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return b
}

func passMove(seat int) legal.Move {
	return legal.Move{Type: legal.TypePassPriority, Player: seatID(seat), Kind: legal.KindPass, Label: "Pass priority"}
}

func landMove(t *testing.T, seat int, id string) legal.Move {
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindLand,
		Label: "Play Mountain", Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand"}),
	}
}

func castMove(t *testing.T, seat int, id, label string, targets ...map[string]string) legal.Move {
	params := map[string]any{"instance_id": id, "from_zone": "hand"}
	if len(targets) > 0 {
		params["targets"] = targets
	}
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id), Params: mustJSON(t, params),
	}
}

func cardTarget(id string) map[string]string { return map[string]string{"kind": "card", "id": id} }
func playerTarget(i int) map[string]string {
	return map[string]string{"kind": "player", "id": seatID(i).String()}
}

func attackMove(t *testing.T, seat int, attacker string, target int) legal.Move {
	return legal.Move{
		Type: legal.TypeDeclareAttacker, Player: seatID(seat), Kind: legal.KindAttack,
		Label: fmt.Sprintf("Attack Seat%d with %s", target, attacker), Source: uuid.MustParse(attacker),
		Params: mustJSON(t, map[string]string{"attacker": attacker, "target": seatID(target).String()}),
	}
}

func blockMove(t *testing.T, seat int, blocker, attacker string) legal.Move {
	return legal.Move{
		Type: legal.TypeDeclareBlocker, Player: seatID(seat), Kind: legal.KindBlock,
		Label: "Block with " + blocker, Source: uuid.MustParse(blocker),
		Params: mustJSON(t, map[string]string{"blocker": blocker, "attacker": attacker}),
	}
}

func manaMove(t *testing.T, seat int, id string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateManaAbility, Player: seatID(seat), Kind: legal.KindMana,
		Label: "Mountain: add {R}", Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"card_id": id, "ability_index": 0}),
	}
}

func choiceMove(t *testing.T, seat int, choiceID, label string, params map[string]any) legal.Move {
	params["choice_id"] = choiceID
	return legal.Move{
		Type: legal.TypeResolveChoice, Player: seatID(seat), Kind: legal.KindChoice,
		Label: label, Params: mustJSON(t, params),
	}
}

func input(seat int, v protocol.GameView, moves ...legal.Move) aiseat.Input {
	return aiseat.Input{View: v, Seat: seatID(seat), Moves: moves}
}
