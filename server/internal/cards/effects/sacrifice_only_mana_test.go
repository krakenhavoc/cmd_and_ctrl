package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_only_mana_test.go — the proof cards for #1242 and #1283.
//
//   - Gold, Eldrazi Spawn and Eldrazi Scion: the three catalog tokens
//     whose mana ability is a sacrifice with NO {T}. Until #1242 the
//     auto-tapper refused all three, so a board of them read as
//     unpayable to the preview, the strict gate and every bot.
//   - Deadly Dispute: the cast that names a Spawn to its additional
//     cost is not allowed to crack that same Spawn for its mana.
//   - Cadaverous Bloom: "Exile a card from your hand" — the auto-tapper
//     never plans it, the hand-clicked activation pays it, and the
//     card it spends is EXILED, not discarded.

const cadaverousBloomOracle = "fbb0f73b-5e30-4632-99c1-e49582e41f8d"

// seatToken puts a catalog token under `owner` on the battlefield.
func seatToken(g *game.Game, owner uuid.UUID, tok game.Card) uuid.UUID {
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = owner, owner
	return pushBattlefieldCardWithTimestamp(g, tok)
}

func tapEventFor(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventTapCard && ev.CardID == id {
			return true
		}
	}
	return false
}

// The three real tokens pay a {3} cast through the strict auto-tap
// path, and none of them is TAPPED on the way — each is cracked.
func TestGoldSpawnAndScionPayACastThroughAutoTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	tokens := []uuid.UUID{
		seatToken(g, me.ID, GoldToken()),
		seatToken(g, me.ID, EldraziSpawnToken()),
		seatToken(g, me.ID, b28EldraziScionToken()),
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Three-Drop Golem", TypeLine: "Artifact Creature — Golem",
		ManaCost: "{3}", Owner: me.ID, Controller: me.ID,
	})

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast off Gold + Spawn + Scion: %v", err)
	}
	for _, id := range tokens {
		if _, ok := battlefieldCard(g, id); ok {
			t.Errorf("token %v survived a cast it paid for", id)
		}
		if tapEventFor(g, id) {
			t.Errorf("token %v was TAPPED — its cost prints no {T}", id)
		}
	}
}

// Deadly Dispute names the only Spawn as its sacrifice, and the Spawn
// is also the only source for the {1}. The engine will not spend it
// twice: the cast is refused as unpayable BEFORE anything moves, and
// the Spawn, the Swamp and the spell are all where they were. Before
// the exclusion the plan cracked the Spawn for mana and the sacrifice
// then found nothing to eat.
func TestDeadlyDisputeDoesNotCrackTheSpawnItSacrifices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	swamp := seatToken(g, me.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	spawn := seatToken(g, me.ID, EldraziSpawnToken())
	dispute := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: dispute, Name: "Deadly Dispute", TypeLine: "Instant",
		ManaCost: "{1}{B}", OracleID: deadlyDisputeOracle, Owner: me.ID, Controller: me.ID,
	})

	err := g.CastSpell(me.ID, dispute, game.CastSpellParams{
		Strict: true, AutoTap: true, SacrificeIDs: []uuid.UUID{spawn},
	})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("CastSpell = %v, want an InsufficientManaError", err)
	}
	if _, ok := battlefieldCard(g, spawn); !ok {
		t.Error("the refused cast cracked the Spawn anyway")
	}
	if c, _ := battlefieldCard(g, swamp); c.Tapped {
		t.Error("the refused cast tapped the Swamp")
	}
	if !me.Hand.Contains(dispute) {
		t.Error("the refused spell left the hand")
	}

	// With a second Spawn the same cast goes through: one is the
	// sacrifice, the other the {1}.
	other := seatToken(g, me.ID, EldraziSpawnToken())
	if err := g.CastSpell(me.ID, dispute, game.CastSpellParams{
		Strict: true, AutoTap: true, SacrificeIDs: []uuid.UUID{spawn},
	}); err != nil {
		t.Fatalf("Deadly Dispute with a second Spawn: %v", err)
	}
	for _, id := range []uuid.UUID{spawn, other} {
		if _, ok := battlefieldCard(g, id); ok {
			t.Errorf("Spawn %v is still on the battlefield", id)
		}
	}
}

// --- Cadaverous Bloom ------------------------------------------------

func seatBloomWithHand(t *testing.T, n int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	bloom := seatToken(g, me.ID, game.Card{Name: "Cadaverous Bloom", TypeLine: "Enchantment", OracleID: cadaverousBloomOracle})
	hand := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare Spell", TypeLine: "Sorcery", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
		hand = append(hand, id)
	}
	return g, me, bloom, hand
}

// The hand-clicked activation: the named card goes to EXILE — not the
// graveyard, and with no discard event, so nothing that watches
// discards (madness, Marauding Mako) sees it — and the ability queues
// its one "{B}{B} or {G}{G}" pick, which adds two of the chosen colour.
func TestCadaverousBloomExilesTheNamedCardForTwoMana(t *testing.T) {
	g, me, bloom, hand := seatBloomWithHand(t, 2)

	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{hand[0]}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Hand.Contains(hand[0]) {
		t.Fatal("the exiled card is still in hand")
	}
	if !g.Exile.Contains(hand[0]) {
		t.Fatal("the named card did not reach exile")
	}
	if me.Graveyard.Contains(hand[0]) {
		t.Error("the named card went to the graveyard — exiling is not discarding")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard {
			t.Fatalf("an EventDiscardCard fired (%+v) — the Bloom's cost exiles, it does not discard", ev)
		}
	}
	if !me.Hand.Contains(hand[1]) {
		t.Error("the card that was not named left the hand")
	}
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Source == bloom {
			pick = c
		}
	}
	if pick == nil {
		t.Fatal("no {B}{B}-or-{G}{G} pick was queued")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	greens := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "G" {
			greens++
		}
	}
	if greens != 2 || len(me.ManaPool) != 2 {
		t.Errorf("pool = %+v, want {G}{G}", me.ManaPool)
	}
}

// The cost has to be paid, and paid on its own field: no exile_ids is a
// refusal, and so is naming the card as a DISCARD instead — the two are
// separate components with separate exits.
func TestCadaverousBloomRefusesAnUnpaidOrMisfiledCost(t *testing.T) {
	g, me, bloom, hand := seatBloomWithHand(t, 1)

	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the Bloom activated with no card exiled")
	}
	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{DiscardIDs: []uuid.UUID{hand[0]}}); err == nil {
		t.Error("the Bloom accepted a DISCARD as its exile payment")
	}
	if err := g.ActivateManaAbility(me.ID, bloom, 0, game.ManaAbilityParams{ExileIDs: []uuid.UUID{bloom}}); err == nil {
		t.Error("the Bloom accepted a card that is not in hand")
	}
	if !me.Hand.Contains(hand[0]) {
		t.Error("a refused activation spent the card anyway")
	}
	if len(me.ManaPool) != 0 || len(g.PendingChoices) != 0 {
		t.Errorf("a refused activation produced mana: pool %+v, choices %d", me.ManaPool, len(g.PendingChoices))
	}
}

// The auto-tapper never plans it: which card leaves the hand is a
// decision, and the planner makes none — even with a full hand and a
// cost the Bloom alone could pay.
func TestCadaverousBloomIsNotAnAutoTapSource(t *testing.T) {
	g, me, _, _ := seatBloomWithHand(t, 3)
	cost, err := game.ParseCost("{B}")
	if err != nil {
		t.Fatal(err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Fatalf("the auto-tapper planned %v — it would have picked a card to exile", plan)
	}
}
