package heuristic_test

import (
	"math"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attachment_test.go is #727: an attached permanent is priced ONCE, by
// what it does.
//
// Every test here is written as a pair of boards that differ only in
// the attachment relation, because that is the only way to see the
// double count: an Equipment's +2/+2 is already on its host's wire P/T,
// so the question is never "is the creature worth more" but "is the
// Equipment worth a second line as well".
//
// Each assertion fails if the role split is backed out of score.go —
// with `boardValue` gone, both halves of every pair price identically.

// --- fixtures --------------------------------------------------------

func owner(i int) cardOpt {
	return func(c *protocol.CardView) { c.Owner = seatID(i).String() }
}

func restricted(tokens ...string) cardOpt {
	return func(c *protocol.CardView) { c.Restrictions = append(c.Restrictions, tokens...) }
}

func attachedToCard(id string) cardOpt {
	return func(c *protocol.CardView) {
		c.AttachedTo = &protocol.TargetRefView{Kind: "card", ID: id}
	}
}

func attachedToPlayer(i int) cardOpt {
	return func(c *protocol.CardView) {
		c.AttachedTo = &protocol.TargetRefView{Kind: "player", ID: seatID(i).String()}
	}
}

// equipment is an "Artifact — Equipment" controlled by `controller`.
func equipment(id string, controller int, name string, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(controller).String(),
		Controller: seatID(controller).String(),
		TypeLine:   "Artifact — Equipment",
		ManaCost:   "{3}",
		KnownByYou: true,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// aura is an "Enchantment — Aura" controlled by `controller`.
func aura(id string, controller int, name string, opts ...cardOpt) protocol.CardView {
	c := protocol.CardView{
		InstanceID: id,
		Name:       name,
		Owner:      seatID(controller).String(),
		Controller: seatID(controller).String(),
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{1}{W}",
		KnownByYou: true,
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

func boardOf(t *testing.T, v protocol.GameView, seat int) float64 {
	t.Helper()
	e := heuristic.DefaultWeights().Evaluate(v)[seatID(seat).String()]
	if e == nil {
		t.Fatalf("no evaluation for seat %d", seat)
	}
	return e.Board
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// --- buff: the value rides the host ----------------------------------

func TestAnAttachedEquipmentIsNotASecondPermanent(t *testing.T) {
	w := heuristic.DefaultWeights()
	// One 4/4 — already carrying the Sword's +2/+2, because the wire's
	// power and toughness are post-layer — and one Sword. The only
	// difference between the boards is whether the Sword is attached.
	host := creature(cardID(1), 0, "Equipped Bear", 4, 4)
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}

	attached := newView(seats, withBattlefield(host,
		equipment(cardID(2), 0, "Sword", attachedToCard(cardID(1)))))
	loose := newView(seats, withBattlefield(host,
		equipment(cardID(2), 0, "Sword")))

	got, want := boardOf(t, attached, 0), boardOf(t, loose, 0)
	if got >= want {
		t.Errorf("an equipped board scored %.3f, the same creature with the Sword unattached %.3f — "+
			"the boost is already on the host's P/T, so the attached Sword must not also cost a full permanent", got, want)
	}
	// And it is priced at the reattach residual, not at nothing: the
	// Sword survives its host (CR 704.5n) and can be moved.
	if exp := w.CreatureValue(&host) + w.AttachedEquipment; !near(got, exp) {
		t.Errorf("equipped board = %.3f, want creature + AttachedEquipment = %.3f", got, exp)
	}
	if w.AttachedEquipment <= 0 || w.AttachedEquipment >= w.Permanent {
		t.Errorf("AttachedEquipment = %.3f: a reattachable Equipment is worth something, but less than a free-standing permanent",
			w.AttachedEquipment)
	}
}

func TestABuffAuraRidesItsHostAndIsWorthLessThanAnEquipment(t *testing.T) {
	w := heuristic.DefaultWeights()
	host := creature(cardID(1), 0, "Enchanted Bear", 4, 4)
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}

	attached := newView(seats, withBattlefield(host,
		aura(cardID(2), 0, "Holy Strength", attachedToCard(cardID(1)))))
	loose := newView(seats, withBattlefield(host,
		aura(cardID(2), 0, "Holy Strength")))

	got := boardOf(t, attached, 0)
	if want := boardOf(t, loose, 0); got >= want {
		t.Errorf("an enchanted board scored %.3f against %.3f with the Aura unattached — the +N/+N is already on the host", got, want)
	}
	if exp := w.CreatureValue(&host) + w.AttachedAura; !near(got, exp) {
		t.Errorf("enchanted board = %.3f, want creature + AttachedAura = %.3f", got, exp)
	}
	// An Aura dies with its host (CR 704.5m); an Equipment falls off and
	// waits for the next creature. The residuals have to say so.
	if w.AttachedAura >= w.AttachedEquipment {
		t.Errorf("AttachedAura %.3f >= AttachedEquipment %.3f: the one that survives its host is the one worth more",
			w.AttachedAura, w.AttachedEquipment)
	}
}

// --- restriction: the value is the host it neutralises ---------------

func TestRestrictionsDiscountTheCreatureTheyHold(t *testing.T) {
	w := heuristic.DefaultWeights()
	free := creature(cardID(1), 1, "Threat", 5, 5)
	pacified := creature(cardID(1), 1, "Threat", 5, 5, restricted("cant_attack", "cant_block"))
	arrested := creature(cardID(1), 1, "Threat", 5, 5,
		restricted("cant_attack", "cant_block", "cant_activate", "cant_activate_mana"))
	unblockable := creature(cardID(1), 1, "Threat", 5, 5, restricted("cant_be_blocked"))

	if w.CreatureValue(&pacified) >= w.CreatureValue(&free) {
		t.Errorf("a creature that can neither attack nor block (%.3f) must be worth less than the same body free (%.3f)",
			w.CreatureValue(&pacified), w.CreatureValue(&free))
	}
	if w.CreatureValue(&arrested) >= w.CreatureValue(&pacified) {
		t.Errorf("Arrest (%.3f) takes the abilities as well as the combat, so it must beat Pacifism (%.3f)",
			w.CreatureValue(&arrested), w.CreatureValue(&pacified))
	}
	// "cant_be_blocked" restricts the DEFENDER, not the host.
	if !near(w.CreatureValue(&unblockable), w.CreatureValue(&free)) {
		t.Errorf("cant_be_blocked discounted its own creature (%.3f vs %.3f) — the sign is backwards",
			w.CreatureValue(&unblockable), w.CreatureValue(&free))
	}
}

func TestPacifismOnAnOpponentsThreatIsAGainForTheCaster(t *testing.T) {
	w := heuristic.DefaultWeights()
	me := seatID(0).String()
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}

	// The same card in the same seats, doing nothing and doing its job.
	// Nothing else moves: the Aura is on my battlefield either way, so
	// what is being measured is only what attaching it BOUGHT.
	idle := newView(seats, withBattlefield(
		creature(cardID(1), 1, "Threat", 5, 5),
		aura(cardID(2), 0, "Pacifism"),
	))
	cast := newView(seats, withBattlefield(
		creature(cardID(1), 1, "Threat", 5, 5, restricted("cant_attack", "cant_block")),
		aura(cardID(2), 0, "Pacifism", attachedToCard(cardID(1))),
	))

	if heuristic.Score(cast, me) <= heuristic.Score(idle, me) {
		t.Errorf("Pacifism on an opponent's 5/5 scored %.3f against %.3f doing nothing — a removal Aura has to be worth casting",
			heuristic.Score(cast, me), heuristic.Score(idle, me))
	}

	// The ledger balances: what the host's controller lost is what the
	// Aura's controller gained, to the point.
	before := w.Evaluate(idle)
	after := w.Evaluate(cast)
	lost := before[seatID(1).String()].Board - after[seatID(1).String()].Board
	// The Aura's own line stops being a flat permanent and becomes the
	// neutralised value, so back the flat permanent out of the baseline.
	gained := after[me].Board - (before[me].Board - w.Permanent)
	if lost <= 0 {
		t.Fatalf("the pacified seat lost %.3f — the restriction discount is not reaching the host", lost)
	}
	if !near(lost, gained) {
		t.Errorf("the host's controller lost %.3f and the Aura's controller gained %.3f — a restriction Aura is worth exactly what it neutralises",
			lost, gained)
	}
}

// --- control: the layer-2 effect already moved the creature ----------

func TestAControlAuraIsCountedOnce(t *testing.T) {
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}
	// A 5/5 owned by seat 1 that seat 0 now controls — which is what
	// layer 2 writes when Control Magic resolves.
	stolen := creature(cardID(1), 0, "Stolen Threat", 5, 5, owner(1))

	withAura := newView(seats, withBattlefield(stolen,
		aura(cardID(2), 0, "Control Magic", attachedToCard(cardID(1)))))
	bare := newView(seats, withBattlefield(stolen))

	got, want := boardOf(t, withAura, 0), boardOf(t, bare, 0)
	if !near(got, want) {
		t.Errorf("a stolen creature with its Control Magic scored %.3f against %.3f for the creature alone — "+
			"the theft is already on the creature's line and must not be paid for twice", got, want)
	}
	// And the creature really is on the thief's ledger, not its owner's.
	if b := boardOf(t, withAura, 1); b != 0 {
		t.Errorf("the original owner still scores %.3f for a creature they do not control", b)
	}
}

// --- curse: its own permanent ----------------------------------------

func TestACurseOnAPlayerKeepsItsOwnLineWhileAnAuraOnACreatureDoesNot(t *testing.T) {
	w := heuristic.DefaultWeights()
	seats := []protocol.PlayerView{newSeat(0), newSeat(1)}
	host := creature(cardID(1), 0, "Bear", 4, 4)

	v := newView(seats, withBattlefield(
		host,
		aura(cardID(2), 0, "Holy Strength", attachedToCard(cardID(1))),
		aura(cardID(3), 0, "Curse of Opulence", attachedToPlayer(1)),
	))

	// There is no host permanent under a Curse for its value to ride,
	// so it is priced as the permanent it is — while the Aura beside it
	// on a creature is not.
	exp := w.CreatureValue(&host) + w.AttachedAura + w.Permanent
	if got := boardOf(t, v, 0); !near(got, exp) {
		t.Errorf("board = %.3f, want creature + AttachedAura + Permanent = %.3f "+
			"(the Curse keeps a full line, the buff Aura does not)", got, exp)
	}
}

// --- the classification itself ---------------------------------------

func TestAttachmentRoleReadsTheHostRatherThanTheName(t *testing.T) {
	buffHost := creature(cardID(1), 0, "Mine", 2, 2)
	stolenHost := creature(cardID(1), 0, "Stolen", 2, 2, owner(1))
	pacifiedHost := creature(cardID(1), 1, "Theirs", 2, 2, restricted("cant_attack", "cant_block"))
	cloakedHost := creature(cardID(1), 1, "Theirs", 2, 2, restricted("cant_be_blocked"))

	for _, tc := range []struct {
		name string
		card protocol.CardView
		host *protocol.CardView
		want heuristic.AttachRole
	}{
		{"unattached", equipment(cardID(2), 0, "Sword"), nil, heuristic.AttachNone},
		{"dangling", equipment(cardID(2), 0, "Sword", attachedToCard(cardID(9))), nil, heuristic.AttachNone},
		{"curse", aura(cardID(2), 0, "Curse", attachedToPlayer(1)), nil, heuristic.AttachCurse},
		{"equipment", equipment(cardID(2), 0, "Sword", attachedToCard(cardID(1))), &buffHost, heuristic.AttachBuff},
		{"buff aura", aura(cardID(2), 0, "Holy Strength", attachedToCard(cardID(1))), &buffHost, heuristic.AttachBuff},
		{"control aura", aura(cardID(2), 0, "Control Magic", attachedToCard(cardID(1))), &stolenHost, heuristic.AttachControl},
		{"restriction aura", aura(cardID(2), 0, "Pacifism", attachedToCard(cardID(1))), &pacifiedHost, heuristic.AttachRestriction},
		// The host can't be blocked, which helps it — an Aether Tunnel
		// on somebody else's creature is a gift, not an answer.
		{"evasion aura", aura(cardID(2), 0, "Aether Tunnel", attachedToCard(cardID(1))), &cloakedHost, heuristic.AttachBuff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := heuristic.AttachmentRole(&tc.card, tc.host); got != tc.want {
				t.Errorf("role = %v, want %v", got, tc.want)
			}
		})
	}
}
