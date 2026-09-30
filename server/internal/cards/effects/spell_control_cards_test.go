package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spell_control_cards_test.go — the ADR 0104 cards: Invert Polarity,
// Aethersnatch, Commandeer, Sudden Substitution and Perplexing Chimera.

const (
	invertPolarityOracle     = "69e12b1f-0fd9-43a9-b3db-fe08290442c6"
	aethersnatchOracle       = "45892ec2-8996-444f-a61c-5151926baa0b"
	commandeerOracle         = "ad7d854b-d303-40eb-acd6-03a8023e05e7"
	suddenSubstitutionOracle = "e5ddacfb-e5a8-4885-a4c9-fa117f321293"
	perplexingChimeraOracle  = "7d075b8a-a606-4590-b52b-b4ef3a9e342f"
)

// respondWith hands priority to `p` (passing from the active seat) and
// casts `name` from p's hand in response to what is on the stack.
func respondWith(t *testing.T, g *game.Game, p *game.Player, card game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	if card.InstanceID == uuid.Nil {
		card.InstanceID = uuid.New()
	}
	card.Owner, card.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(card) })
	for i := 0; i < 8 && g.Seats[g.Turn.PriorityHolder].ID != p.ID; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority toward %s: %v", p.Name, err)
		}
	}
	if err := g.CastSpell(p.ID, card.InstanceID, params); err != nil {
		t.Fatalf("CastSpell %s: %v", card.Name, err)
	}
	return card.InstanceID
}

func stackItemControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.ReadSnapshot(func() {
		if it := g.StackMeta[id]; it != nil {
			out = it.Controller
		}
	})
	return out
}

func permanentOf(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

// invertPolarityGame casts a Divination for seat 0, answers it with
// seat 1's Invert Polarity, and calls the flip `call`. Returns the
// game, the Divination and whether the flip was won.
func invertPolarityGame(t *testing.T, call string) (*game.Game, uuid.UUID, bool) {
	t.Helper()
	g := newCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	div := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	respondWith(t, g, thief, game.Card{Name: "Invert Polarity", TypeLine: "Instant", OracleID: invertPolarityOracle},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: div}}})
	passPriorityAroundTable(t, g)
	flip := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if flip == nil {
		t.Fatal("Invert Polarity asked for no coin call")
	}
	if flip.Chooser != thief.ID {
		t.Fatalf("the coin call went to %s, want Invert Polarity's controller %s (CR 705.2)", flip.Chooser, thief.ID)
	}
	if err := g.ResolveCoinCall(flip.ID, thief.ID, call); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	flips := randomEvents(g, game.EventFlipCoin)
	won := len(flips) > 0 && flips[len(flips)-1].Won
	_ = caster
	return g, div, won
}

// TestInvertPolarityStealsOnAWinAndCountersOnALoss plays fresh games
// until it has seen the flip won and lost — each game mints its own RNG
// key, and the engine draws won or lost rather than a face (ADR 0054),
// so the call does not decide it. The won game's Divination draws for
// the thief; the lost game's is countered. 64 games leave a 2^-63
// chance of never seeing one branch.
func TestInvertPolarityStealsOnAWinAndCountersOnALoss(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g, div, won := invertPolarityGame(t, "heads")
		caster, thief := g.Seats[0], g.Seats[1]
		if won {
			sawWin = true
			if got := stackItemControllerOf(g, div); got != thief.ID {
				t.Fatalf("won flip: Divination's controller = %s, want the thief %s", got, thief.ID)
			}
			thiefHand := len(thief.Hand.Cards)
			passPriorityAroundTable(t, g)
			if got := len(thief.Hand.Cards); got != thiefHand+2 {
				t.Errorf("won flip: the thief's hand = %d, want %d", got, thiefHand+2)
			}
			continue
		}
		sawLoss = true
		if g.Stack.Contains(div) {
			t.Error("lost flip: Divination is still on the stack — it should have been countered")
		}
		if !caster.Graveyard.Contains(div) {
			t.Error("lost flip: the countered Divination is not in its owner's graveyard")
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("opposite calls on one seed did not cover both branches (win %v, loss %v)", sawWin, sawLoss)
	}
}

// TestAethersnatchStealsAPermanentSpell — CR 110.2b: the permanent is
// the thief's, and its default controller is the caster.
func TestAethersnatchStealsAPermanentSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	giant := castCatalogSpell(t, g, "Hill Giant", "Creature — Giant", "", nil)
	respondWith(t, g, thief, game.Card{Name: "Aethersnatch", TypeLine: "Instant", OracleID: aethersnatchOracle},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: giant}}})
	passPriorityAroundTable(t, g)
	perm, ok := permanentOf(g, giant)
	if !ok {
		t.Fatal("the stolen creature spell did not resolve")
	}
	if perm.Controller != thief.ID {
		t.Errorf("controller = %s, want the thief %s", perm.Controller, thief.ID)
	}
	if perm.BaseController != caster.ID {
		t.Errorf("default controller = %s, want the caster %s", perm.BaseController, caster.ID)
	}
}

// TestCommandeerPitchesTwoBlueCards — the alternative cost exiles both
// named cards, and the steal still happens.
func TestCommandeerPitchesTwoBlueCards(t *testing.T) {
	g := newCatalogGame(t)
	caster, thief := g.Seats[0], g.Seats[1]
	div := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	blue1 := game.Card{InstanceID: uuid.New(), Name: "Blue One", TypeLine: "Instant", ManaCost: "{U}",
		Colors: []string{"U"}, Owner: thief.ID, Controller: thief.ID}
	blue2 := game.Card{InstanceID: uuid.New(), Name: "Blue Two", TypeLine: "Instant", ManaCost: "{U}",
		Colors: []string{"U"}, Owner: thief.ID, Controller: thief.ID}
	g.WithWriteLock(func() {
		thief.Hand.PushTop(blue1)
		thief.Hand.PushTop(blue2)
	})
	respondWith(t, g, thief, game.Card{Name: "Commandeer", TypeLine: "Instant", ManaCost: "{5}{U}{U}",
		OracleID: commandeerOracle},
		game.CastSpellParams{
			AlternativeCost: "pitch",
			AltCostIDs:      []uuid.UUID{blue1.InstanceID, blue2.InstanceID},
			Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: div}},
		})
	if !g.Exile.Contains(blue1.InstanceID) || !g.Exile.Contains(blue2.InstanceID) {
		t.Fatal("the two pitched blue cards are not in exile")
	}
	passPriorityAroundTable(t, g)
	if got := stackItemControllerOf(g, div); got != uuid.Nil {
		t.Fatalf("Divination is still on the stack under %s", got)
	}
	_ = caster
}

// TestCommandeerRefusesAPitchOfOneCard — the count is exactly two.
func TestCommandeerRefusesAPitchOfOneCard(t *testing.T) {
	g := newCatalogGame(t)
	thief := g.Seats[1]
	div := castCatalogSpell(t, g, "Divination", "Sorcery", spellControlDivinationOracle, nil)
	blue := game.Card{InstanceID: uuid.New(), Name: "Blue One", TypeLine: "Instant", ManaCost: "{U}",
		Colors: []string{"U"}, Owner: thief.ID, Controller: thief.ID}
	g.WithWriteLock(func() { thief.Hand.PushTop(blue) })
	cmd := game.Card{InstanceID: uuid.New(), Name: "Commandeer", TypeLine: "Instant", ManaCost: "{5}{U}{U}",
		OracleID: commandeerOracle, Owner: thief.ID, Controller: thief.ID}
	g.WithWriteLock(func() { thief.Hand.PushTop(cmd) })
	for i := 0; i < 8 && g.Seats[g.Turn.PriorityHolder].ID != thief.ID; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	err := g.CastSpell(thief.ID, cmd.InstanceID, game.CastSpellParams{
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{blue.InstanceID},
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: div}},
	})
	if err == nil {
		t.Error("Commandeer was cast by pitching ONE blue card")
	}
}

// TestSuddenSubstitutionSwapsASpellAndACreature — both halves, and the
// spell's new controller is offered new targets.
func TestSuddenSubstitutionSwapsASpellAndACreature(t *testing.T) {
	g := newCatalogGame(t)
	caster, other := g.Seats[0], g.Seats[1]
	creature := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hill Giant",
		TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: other.ID, Controller: other.ID})
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: other.ID}})
	respondWith(t, g, other, game.Card{Name: "Sudden Substitution", TypeLine: "Instant",
		OracleID: suddenSubstitutionOracle},
		game.CastSpellParams{Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: bolt, Slot: 0},
			{Kind: game.TargetCard, ID: creature, Slot: 1},
		}})
	passPriorityAroundTable(t, g)
	if got := stackItemControllerOf(g, bolt); got != other.ID {
		t.Errorf("Bolt controller = %s, want %s", got, other.ID)
	}
	if perm, _ := permanentOf(g, creature); perm.Controller != caster.ID {
		t.Errorf("creature controller = %s, want %s", perm.Controller, caster.ID)
	}
	if latestRetarget(g, other.ID) == nil {
		t.Error("the spell's new controller was not offered new targets")
	}
}

// TestPerplexingChimeraTradesItselfForASpell — the "you may" declares a
// trade and names the spell; answering yes exchanges the two, and the
// Chimera's former controller aims the stolen Bolt.
func TestPerplexingChimeraTradesItselfForASpell(t *testing.T) {
	g := newCatalogGame(t)
	caster, chimeraOwner := g.Seats[0], g.Seats[1]
	chimera := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Perplexing Chimera",
		TypeLine: "Enchantment Creature — Chimera", OracleID: perplexingChimeraOracle, Power: 3, Toughness: 3,
		Owner: chimeraOwner.ID, Controller: chimeraOwner.ID})
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: chimeraOwner.ID}})
	passPriorityAroundTable(t, g)

	prompt := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, chimeraOwner.ID)
	if prompt == nil {
		t.Fatal("the Chimera's controller was not asked")
	}
	if got := prompt.TradeSubject(); got != bolt {
		t.Errorf("the prompt's trade subject = %s, want the Bolt %s", got, bolt)
	}
	onWire := false
	for _, ch := range protocol.ViewOfGame(g).PendingChoices {
		if ch.ID == prompt.ID.String() && ch.TradeFor == bolt.String() {
			onWire = true
		}
	}
	if !onWire {
		t.Error("the trigger prompt does not carry trade_for on the wire")
	}
	answerLatestTriggerPrompt(t, g, chimeraOwner.ID, true)
	passPriorityAroundTable(t, g)

	if got := stackItemControllerOf(g, bolt); got != chimeraOwner.ID {
		t.Fatalf("Bolt controller = %s, want the Chimera's former controller %s", got, chimeraOwner.ID)
	}
	if perm, _ := permanentOf(g, chimera); perm.Controller != caster.ID {
		t.Errorf("Chimera controller = %s, want the Bolt's caster %s", perm.Controller, caster.ID)
	}
	retarget := latestRetarget(g, chimeraOwner.ID)
	if retarget == nil {
		t.Fatal("no new targets offered for the stolen Bolt")
	}
	if err := g.ResolveRetarget(retarget.ID, chimeraOwner.ID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, caster.ID); got != 37 {
		t.Errorf("the Bolt's caster is at %d, want 37", got)
	}
}
