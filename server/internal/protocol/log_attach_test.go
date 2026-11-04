package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_attach_test.go — #4444, ADR 0209 §3: the client's attach_equip and
// attach_aura cues read the `attach` log line, which this issue added
// (EventAttach was a deliberate silence until then). It pins that every
// viewer, an opponent included, gets one `attach` entry each time an Aura,
// Equipment or Fortification becomes attached (CR 701.3a), with:
//
//   - `card_id` the attachment, `target` the host card (or `target_seat`
//     the host player, for a Curse), and `label` what the attachment is:
//     "Equipment", "Aura" or "Fortification";
//   - a `batch`, which is how the cue plays its sound once, and which a
//     token created attached shares with its `token` entry;
//
// for an equip-style attach (AttachForEffect is the one writer: equip,
// reconfigure and "attach it to" all end there), an effect moving it to
// another host, an Aura spell resolving onto its target (CR 303.4a,
// 608.3), a token copy of an Aura spell (CR 707.10) and a Role token
// created attached (CR 303.7). An attach onto the host it is already on
// does nothing (CR 701.3b) and writes no line.

// attachTable is an active two-seat game in a main phase of seat 0's turn
// order, with every opening hand kept.
func attachTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

// pushPermanent puts a public permanent onto the battlefield under `owner`.
func pushPermanent(g *game.Game, owner *game.Player, opp *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id,
			Name:       name,
			TypeLine:   typeLine,
			Power:      2,
			Toughness:  2,
			Owner:      owner.ID,
			Controller: owner.ID,
			KnownBy:    map[uuid.UUID]bool{owner.ID: true, opp.ID: true},
		})
	})
	return id
}

// attachEntries is the `attach` entries `viewer` sees for attachment id.
func attachEntries(g *game.Game, viewer *game.Player, id uuid.UUID) []LogEvent {
	var out []LogEvent
	for _, e := range entriesNaming(g, viewer, id) {
		if e.Kind == LogAttach {
			out = append(out, e)
		}
	}
	return out
}

func wantAttachWire(t *testing.T, e LogEvent, fields map[string]any) {
	t.Helper()
	if e.Batch == 0 {
		t.Errorf("the attach entry has no batch: %+v", e)
	}
	raw := wireFields(t, e)
	for key, want := range fields {
		if raw[key] != want {
			t.Errorf("wire field %q = %v, want %v (entry %v)", key, raw[key], want, raw)
		}
	}
}

func TestAnEquipmentBecomingAttachedIsLoggedAndAReattachIsNot(t *testing.T) {
	g, me, opp := attachTable(t)
	bear := pushPermanent(g, me, opp, "Grizzly Bears", "Creature — Bear")
	other := pushPermanent(g, me, opp, "Runeclaw Bear", "Creature — Bear")
	sword := pushPermanent(g, me, opp, "Test Sword", "Artifact — Equipment")

	attach := func(host uuid.UUID) {
		t.Helper()
		var err error
		g.WithWriteLock(func() {
			err = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: host})
		})
		if err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	}

	attach(bear)
	got := attachEntries(g, opp, sword)
	if len(got) != 1 {
		t.Fatalf("the opponent sees %d attach entries, want 1: %+v", len(got), got)
	}
	wantAttachWire(t, got[0], map[string]any{
		"kind": "attach", "card_id": sword.String(), "target": bear.String(), "label": "Equipment",
	})
	if got[0].Text != "Test Sword equips Grizzly Bears" {
		t.Errorf("text = %q", got[0].Text)
	}
	if got[0].Seat != g.Turn.ActiveSeat {
		t.Errorf("seat = %d, want the Equipment's controller's, %d", got[0].Seat, g.Turn.ActiveSeat)
	}

	// CR 701.3b: onto the host it is already on, the effect does nothing.
	attach(bear)
	if n := len(attachEntries(g, opp, sword)); n != 1 {
		t.Errorf("a re-attach to the same host wrote a line: %d entries", n)
	}

	// Moved to another creature: a line of its own, naming the new host.
	attach(other)
	got = attachEntries(g, opp, sword)
	if len(got) != 2 {
		t.Fatalf("after the move: %d attach entries, want 2", len(got))
	}
	wantAttachWire(t, got[1], map[string]any{"target": other.String(), "label": "Equipment"})
}

// The equip ability itself (CR 702.6a), end to end: activated, then
// resolved, with its line in the resolution's batch.
func TestAnEquipAbilityResolvingIsLoggedAsAnAttach(t *testing.T) {
	g, me, opp := attachTable(t)
	bear := pushPermanent(g, me, opp, "Grizzly Bears", "Creature — Bear")
	sword := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: sword, Name: "Bonesplitter", TypeLine: "Artifact — Equipment",
			OracleID: "452e3f5f-ce17-4682-966b-5cc100210aee",
			Owner:    me.ID, Controller: me.ID,
			KnownBy: map[uuid.UUID]bool{me.ID: true, opp.ID: true},
		})
	})
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	if n := len(attachEntries(g, opp, sword)); n != 0 {
		t.Fatalf("the activation already wrote %d attach entries", n)
	}
	for i := 0; len(g.StackMeta) > 0; i++ {
		if i > 8 {
			t.Fatal("the equip ability never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	got := attachEntries(g, opp, sword)
	if len(got) != 1 {
		t.Fatalf("the opponent sees %d attach entries, want 1: %+v", len(got), entriesNaming(g, opp, sword))
	}
	wantAttachWire(t, got[0], map[string]any{
		"kind": "attach", "card_id": sword.String(), "target": bear.String(), "label": "Equipment",
	})
	if got[0].Text != "Bonesplitter equips Grizzly Bears" {
		t.Errorf("text = %q", got[0].Text)
	}
}

func TestACurseIsLoggedAsEnchantingItsPlayer(t *testing.T) {
	g, me, opp := attachTable(t)
	curse := pushPermanent(g, me, opp, "Test Curse", "Enchantment — Aura Curse")
	var err error
	g.WithWriteLock(func() {
		err = g.AttachForEffect(curse, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	})
	if err != nil {
		t.Fatalf("AttachForEffect: %v", err)
	}
	got := attachEntries(g, me, curse)
	if len(got) != 1 {
		t.Fatalf("%d attach entries, want 1", len(got))
	}
	if got[0].Target != "" || got[0].TargetSeat == nil {
		t.Fatalf("a player host rides target_seat: %+v", got[0])
	}
	wantAttachWire(t, got[0], map[string]any{"label": "Aura"})
}

func TestAnAuraSpellResolvingOntoItsTargetIsLogged(t *testing.T) {
	g, me, opp := attachTable(t)
	bear := pushPermanent(g, opp, me, "Grizzly Bears", "Creature — Bear")
	aura := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: aura, Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if n := len(attachEntries(g, opp, aura)); n != 0 {
		t.Fatalf("the cast already wrote %d attach entries", n)
	}
	for i := 0; g.Stack.Contains(aura); i++ {
		if i > 8 {
			t.Fatal("the Aura spell never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	got := attachEntries(g, opp, aura)
	if len(got) != 1 {
		t.Fatalf("the opponent sees %d attach entries, want 1: %+v", len(got), entriesNaming(g, opp, aura))
	}
	wantAttachWire(t, got[0], map[string]any{
		"kind": "attach", "card_id": aura.String(), "target": bear.String(), "label": "Aura",
	})
	if got[0].Text != "Test Aura enchants Grizzly Bears" {
		t.Errorf("text = %q", got[0].Text)
	}
}

// tokenAndAttach is the `token` and `attach` entries for one created
// token that entered attached.
func tokenAndAttach(t *testing.T, g *game.Game, viewer *game.Player, name string) (token, attach LogEvent) {
	t.Helper()
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.IsToken() {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatalf("no %s token on the battlefield", name)
	}
	var tokens, attaches []LogEvent
	for _, e := range entriesNaming(g, viewer, id) {
		switch e.Kind {
		case LogToken:
			tokens = append(tokens, e)
		case LogAttach:
			attaches = append(attaches, e)
		}
	}
	if len(tokens) != 1 || len(attaches) != 1 {
		t.Fatalf("token %s: %d token and %d attach entries, want one each: %+v",
			name, len(tokens), len(attaches), entriesNaming(g, viewer, id))
	}
	if tokens[0].Batch == 0 || tokens[0].Batch != attaches[0].Batch {
		t.Errorf("a token created attached: token batch %d, attach batch %d, want the same",
			tokens[0].Batch, attaches[0].Batch)
	}
	return tokens[0], attaches[0]
}

func TestATokenCopyOfAnAuraSpellIsLoggedAsCreatedAndAttached(t *testing.T) {
	g, me, opp := attachTable(t)
	bear := pushPermanent(g, opp, me, "Grizzly Bears", "Creature — Bear")
	aura := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: aura, Name: "Test Aura", TypeLine: "Enchantment — Aura",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	var err error
	g.WithWriteLock(func() { err = g.CopySpellForEffect(aura, me.ID, false, nil) })
	if err != nil {
		t.Fatalf("CopySpellForEffect: %v", err)
	}
	for i := 0; g.Stack.Contains(aura); i++ {
		if i > 12 {
			t.Fatal("the Aura spell never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	_, attach := tokenAndAttach(t, g, opp, "Test Aura")
	wantAttachWire(t, attach, map[string]any{"target": bear.String(), "label": "Aura"})
	// The original resolved onto the same creature: its own line.
	if n := len(attachEntries(g, opp, aura)); n != 1 {
		t.Errorf("the Aura card itself: %d attach entries, want 1", n)
	}
}

func TestARoleTokenCreatedAttachedIsLoggedAsCreatedAndAttached(t *testing.T) {
	g, me, opp := attachTable(t)
	bear := pushPermanent(g, me, opp, "Grizzly Bears", "Creature — Bear")
	var err error
	g.WithWriteLock(func() {
		_, err = g.CreateTokensForEffect(me.ID, effects.RoleToken(effects.RoleMonster), 1,
			game.TokenEntryOptions{AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
	})
	if err != nil {
		t.Fatalf("CreateTokensForEffect: %v", err)
	}
	role := effects.RoleToken(effects.RoleMonster).Name
	_, attach := tokenAndAttach(t, g, opp, role)
	wantAttachWire(t, attach, map[string]any{"target": bear.String(), "label": "Aura"})
}
