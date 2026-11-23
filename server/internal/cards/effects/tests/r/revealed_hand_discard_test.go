package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rhFireIce is Fire // Ice in a hand: both halves combined (CR 709.4),
// mana value 4.
func rhFireIce() game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		Layout:     game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}", Colors: []string{"R"}},
			{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}", Colors: []string{"U"}},
		},
	}
	c.SettleImported()
	return c
}

// The filter offers only matching cards. Thoughtseize's "nonland card"
// takes the MDFC by its front face (CR 712.8a) and leaves the Forest.
func TestThoughtseizeOffersOnlyNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhPelakka(), effectstest.RhBolt())
	life := caster.Life

	effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.OpenPick(t, g)
	if pick.Chooser != caster.ID || pick.FromPlayer != victim.ID || pick.Count != 1 {
		t.Fatalf("prompt = chooser %v from %v count %d", pick.Chooser, pick.FromPlayer, pick.Count)
	}
	if want := ids[1:]; !effectstest.SameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want the MDFC and the Bolt %v", pick.DiscardOptions, want)
	}
	if pick.DiscardLabel != "nonland card" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
	// The life loss is the instruction after the choice, so it waits for the
	// answer (CR 608.2c, #4155).
	if caster.Life != life {
		t.Errorf("caster life = %d with the pick open, want %d", caster.Life, life)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("taking the Bolt: %v", err)
	}
	g.SettleResolution()
	if caster.Life != life-2 {
		t.Errorf("caster life = %d after the pick, want %d", caster.Life, life-2)
	}
}

// ADR 0116 §2: the reveal reaches every player (CR 701.20a), with one
// grouped reveal run, not only the caster.
func TestThoughtseizeRevealsToTheWholeTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
	before := len(g.Events)

	effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range victim.Hand.Cards {
		for _, seat := range g.Seats {
			if !c.IsKnownTo(seat.ID) {
				t.Errorf("seat %s is not a knower of revealed %s", seat.Name, c.Name)
			}
		}
	}
	var revealed []uuid.UUID
	seqs := map[uint64]bool{}
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventRevealCards && ev.Actor == victim.ID {
			revealed = append(revealed, ev.CardID)
			seqs[ev.RevealSeq] = true
		}
	}
	if !effectstest.SameIDs(revealed, ids) || len(seqs) != 1 {
		t.Errorf("reveal run = %v in %d groups, want %v in one", revealed, len(seqs), ids)
	}
}

// The server refuses a non-matching pick sent anyway: the Forest is
// refused, the prompt stays open and nothing moves; the Bolt is then
// taken.
func TestThoughtseizeRefusesALandPickedAnyway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())

	effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)

	err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[0]})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("taking the Forest: err = %v, want ErrInvalidParam", err)
	}
	if victim.Hand.Size() != 2 || effectstest.OpenPick(t, g).ID != pick.ID {
		t.Fatalf("a refused pick changed the table: hand %d", victim.Hand.Size())
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("taking the Bolt: %v", err)
	}
	if !victim.Graveyard.Contains(ids[1]) || !victim.Hand.Contains(ids[0]) {
		t.Error("the Bolt was not discarded, or the Forest left the hand")
	}
}

// A hand with no match reveals and discards nothing (CR 609.3), and
// Thoughtseize still costs 2 life (its 2020-08-07 ruling).
func TestThoughtseizeOnAnAllLandHandRevealsAndDiscardsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhForest(), effectstest.RhForest())
	life := caster.Life
	graveyard := victim.Graveyard.Size()

	effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			t.Fatalf("a prompt went up with nothing to choose: %+v", c)
		}
	}
	if victim.Hand.Size() != 3 || victim.Graveyard.Size() != graveyard {
		t.Errorf("hand %d, graveyard +%d: nothing should have moved", victim.Hand.Size(), victim.Graveyard.Size()-graveyard)
	}
	for _, c := range victim.Hand.Cards {
		if !c.IsKnownTo(caster.ID) || !c.IsKnownTo(g.Seats[2].ID) {
			t.Errorf("%s was not revealed", c.Name)
		}
	}
	if caster.Life != life-2 {
		t.Errorf("caster life = %d, want %d", caster.Life, life-2)
	}
}

// Pelakka Predation reads mana value as it is in a hand: Fire // Ice is
// 4 (CR 709.4b), Fireball's X is 0 (CR 202.3e), a land is 0, and a
// copy of Pelakka Predation itself is 3 by its front face (CR 712.8a).
func TestPelakkaPredationChoosesManaValueThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), rhFireIce(), effectstest.RhFireball(), effectstest.RhPelakka(), effectstest.RhBolt())

	effectstest.CastCatalogSpell(t, g, "Pelakka Predation", "Sorcery", effectstest.PelakkaPredationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.OpenPick(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !effectstest.SameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want Fire // Ice and Pelakka Predation %v", pick.DiscardOptions, want)
	}
	if pick.DiscardLabel != "card with mana value 3 or greater" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
}

// Pelakka Predation targets an opponent, never its caster.
func TestPelakkaPredationTargetsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	err := effectstest.CastCatalogSpellErr(t, g, "Pelakka Predation", "Sorcery", effectstest.PelakkaPredationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Error("Pelakka Predation targeted its own caster")
	}
}

// Unmask cast through its pitch: the black card is exiled as the cost
// (CR 118.9, 601.2h) and the pick is Thoughtseize's.
func TestUnmaskPitchesABlackCardAndTakesANonlandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt())
	black := effectstest.HandCardFull(me, "Dark Ritual", "Instant", "{B}", "", []string{"B"})
	unmask := effectstest.HandCardFull(me, "Unmask", "Sorcery", "{3}{B}", effectstest.UnmaskOracle, []string{"B"})

	if err := g.CastSpell(me.ID, unmask, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "pitch",
		AltCostIDs:      []uuid.UUID{black},
		Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}},
	}); err != nil {
		t.Fatalf("Unmask on a pitch: %v", err)
	}
	if !g.Exile.Contains(black) {
		t.Error("the pitched card did not reach exile")
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want only the Bolt", pick.DiscardOptions)
	}
}

// Unmask cannot pitch itself (CR 601.2a), nor a card that isn't black.
func TestUnmaskRefusesABadPitch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	red := effectstest.HandCardFull(me, "Lightning Bolt", "Instant", "{R}", "", []string{"R"})
	unmask := effectstest.HandCardFull(me, "Unmask", "Sorcery", "{3}{B}", effectstest.UnmaskOracle, []string{"B"})
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}}

	for name, pitch := range map[string]uuid.UUID{"a red card": red, "itself": unmask} {
		if err := g.CastSpell(me.ID, unmask, game.CastSpellParams{
			Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}, Targets: target,
		}); err == nil {
			t.Errorf("Unmask pitched %s", name)
		}
	}
	if !me.Hand.Contains(unmask) {
		t.Error("a refused cast took Unmask out of hand")
	}
}

// Snapshot restore mid-prompt: the open pick is a restore point, and
// the restored game still refuses the land and accepts the spell.
func TestRevealedHandPickSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := g.Seats[1]
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhPelakka(), effectstest.RhBolt())

	effectstest.CastCatalogSpell(t, g, "Thoughtseize", "Sorcery", effectstest.ThoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.OpenPick(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	after := effectstest.OpenPick(t, restored)
	if after.ID != before.ID || !effectstest.SameIDs(after.DiscardOptions, before.DiscardOptions) || after.DiscardLabel != before.DiscardLabel {
		t.Fatalf("restored prompt %+v, want %+v", after, before)
	}
	chooser := after.Chooser
	if err := restored.ResolvePendingChoice(after.ID, chooser, []uuid.UUID{ids[0]}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("restored game took the Forest: %v", err)
	}
	if err := restored.ResolvePendingChoice(after.ID, chooser, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("restored game refused Pelakka Predation: %v", err)
	}
	var rv *game.Player
	for _, p := range restored.Seats {
		if p.ID == victim.ID {
			rv = p
		}
	}
	if rv == nil || !rv.Graveyard.Contains(ids[1]) {
		t.Error("the restored discard did not land")
	}
}
