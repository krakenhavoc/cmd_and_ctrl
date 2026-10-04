package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0116 (#2078): the revealed-hand pick's card filter, through the
// three cards that ship on it. One test per rule interaction the ADR
// lists under Tests.

const (
	thoughtseizeOracle     = "edd8d1e8-be43-4c38-bb3a-83081fbaf0b5"
	unmaskOracle           = "1a3a274d-41da-47db-ad41-5720f36a7963"
	pelakkaPredationOracle = "b0fd6889-20b4-439b-aa97-2e90aca1675a"
)

// revealHand replaces p's hand with cards and returns their IDs in
// hand order.
func revealHand(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Hand.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = p.ID, p.ID
		p.Hand.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func rhForest() game.Card {
	return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}
}

func rhBolt() game.Card {
	return game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
}

// rhPelakka is Pelakka Predation in a hand, built the way the importer
// builds it: two faces, face 0 materialised (CR 712.8a).
func rhPelakka() game.Card {
	return mdfcCard(uuid.Nil, pelakkaPredationOracle, game.LayoutModalDFC,
		game.Face{Name: "Pelakka Predation", TypeLine: "Sorcery", ManaCost: "{2}{B}", Colors: []string{"B"}},
		game.Face{Name: "Pelakka Caverns", TypeLine: "Land", Colors: nil},
	)
}

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

func rhFireball() game.Card {
	return game.Card{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}", Colors: []string{"R"}}
}

// openPick returns the one revealed-hand prompt, or fails.
func openPick(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	var found *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDiscardFromHand {
			if found != nil {
				t.Fatal("more than one revealed-hand prompt is open")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no revealed-hand prompt is open")
	}
	return found
}

// The filter offers only matching cards. Thoughtseize's "nonland card"
// takes the MDFC by its front face (CR 712.8a) and leaves the Forest.
func TestThoughtseizeOffersOnlyNonlandCards(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), rhPelakka(), rhBolt())
	life := caster.Life

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	pick := openPick(t, g)
	if pick.Chooser != caster.ID || pick.FromPlayer != victim.ID || pick.Count != 1 {
		t.Fatalf("prompt = chooser %v from %v count %d", pick.Chooser, pick.FromPlayer, pick.Count)
	}
	if want := ids[1:]; !sameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want the MDFC and the Bolt %v", pick.DiscardOptions, want)
	}
	if pick.DiscardLabel != "nonland card" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
	if caster.Life != life-2 {
		t.Errorf("caster life = %d, want %d", caster.Life, life-2)
	}
}

// ADR 0116 §2: the reveal reaches every player (CR 701.20a), with one
// grouped reveal run, not only the caster.
func TestThoughtseizeRevealsToTheWholeTable(t *testing.T) {
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt())
	before := len(g.Events)

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

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
	if !sameIDs(revealed, ids) || len(seqs) != 1 {
		t.Errorf("reveal run = %v in %d groups, want %v in one", revealed, len(seqs), ids)
	}
}

// The server refuses a non-matching pick sent anyway: the Forest is
// refused, the prompt stays open and nothing moves; the Bolt is then
// taken.
func TestThoughtseizeRefusesALandPickedAnyway(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rhForest(), rhBolt())

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)

	err := g.ResolvePendingChoice(pick.ID, caster.ID, []uuid.UUID{ids[0]})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("taking the Forest: err = %v, want ErrInvalidParam", err)
	}
	if victim.Hand.Size() != 2 || openPick(t, g).ID != pick.ID {
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
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	revealHand(victim, rhForest(), rhForest(), rhForest())
	life := caster.Life
	graveyard := victim.Graveyard.Size()

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhFireIce(), rhFireball(), rhPelakka(), rhBolt())

	castCatalogSpell(t, g, "Pelakka Predation", "Sorcery", pelakkaPredationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)

	pick := openPick(t, g)
	if want := []uuid.UUID{ids[1], ids[3]}; !sameIDs(pick.DiscardOptions, want) {
		t.Errorf("options = %v, want Fire // Ice and Pelakka Predation %v", pick.DiscardOptions, want)
	}
	if pick.DiscardLabel != "card with mana value 3 or greater" {
		t.Errorf("label = %q", pick.DiscardLabel)
	}
}

// Pelakka Predation targets an opponent, never its caster.
func TestPelakkaPredationTargetsAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	err := castCatalogSpellErr(t, g, "Pelakka Predation", "Sorcery", pelakkaPredationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Error("Pelakka Predation targeted its own caster")
	}
}

// Unmask cast through its pitch: the black card is exiled as the cost
// (CR 118.9, 601.2h) and the pick is Thoughtseize's.
func TestUnmaskPitchesABlackCardAndTakesANonlandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	ids := revealHand(victim, rhForest(), rhBolt())
	black := handCardFull(me, "Dark Ritual", "Instant", "{B}", "", []string{"B"})
	unmask := handCardFull(me, "Unmask", "Sorcery", "{3}{B}", unmaskOracle, []string{"B"})

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
	passPriorityAroundTable(t, g)
	pick := openPick(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:]) {
		t.Errorf("options = %v, want only the Bolt", pick.DiscardOptions)
	}
}

// Unmask cannot pitch itself (CR 601.2a), nor a card that isn't black.
func TestUnmaskRefusesABadPitch(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	red := handCardFull(me, "Lightning Bolt", "Instant", "{R}", "", []string{"R"})
	unmask := handCardFull(me, "Unmask", "Sorcery", "{3}{B}", unmaskOracle, []string{"B"})
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
	g := newCatalogGame(t)
	victim := g.Seats[1]
	ids := revealHand(victim, rhForest(), rhPelakka(), rhBolt())

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery", thoughtseizeOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
	passPriorityAroundTable(t, g)
	before := openPick(t, g)

	restored := restoreRoundTrip(t, g, true)
	after := openPick(t, restored)
	if after.ID != before.ID || !sameIDs(after.DiscardOptions, before.DiscardOptions) || after.DiscardLabel != before.DiscardLabel {
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
