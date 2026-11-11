package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s32_cards_test.go — Commander deck cards, batch 32: Clues as
// objects, detectives and exile-until-leaves (#5368).

const (
	cmdr2S32RansomNoteOracle   = "49c0aac5-d650-4ada-a194-3bf77f104bc1"
	cmdr2S32KelpOracle         = "1d7d9404-cad6-4c1a-bd47-7a31ebb2791c"
	cmdr2S32ArmedOracle        = "b401a490-d25f-4029-bb91-8999e140c6a1"
	cmdr2S32SophiaOracle       = "eaf8027a-017a-4471-ae7b-a372fd8e0ff3"
	cmdr2S32SereneSleuthOracle = "b0d639cb-71d4-454f-9dd6-89cd7a59b702"
	cmdr2S32SteamboatOracle    = "d97a0e41-9dd7-4466-afb7-3d8bd6cd0eed"
	cmdr2S32EscortOracle       = "c0ebc1df-51ea-413b-bb88-7666a53ad783"
	cmdr2S32ForagersOracle     = "0f5f6f3e-1b3c-43e2-b11d-bd454366845c"
)

func TestCmdr2S32EveryBuiltCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S32RansomNoteOracle: "Ransom Note",
		cmdr2S32KelpOracle:       "Tangletrove Kelp",
		cmdr2S32ArmedOracle:      "Armed with Proof",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Ransom Note --------------------------------------------------------

func TestCmdr2S32RansomNoteSurveilsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "One", "Two", "Three")
	castHoldCreature(t, g, "Ransom Note", "Artifact — Clue", cmdr2S32RansomNoteOracle, 0, 0)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
}

func cmdr2S32Note(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	note := b12Push(g, me.ID, "Ransom Note", "Artifact — Clue", cmdr2S32RansomNoteOracle, 0, 0)
	seedLibrary(me, "One", "Two", "Three")
	mkmAC2AddMana(t, g, me, "{C}{C}")
	return g, me, opp, note
}

func TestCmdr2S32RansomNoteDrawMode(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	hand := me.Hand.Size()
	p7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{2}})
	if g.Battlefield.Contains(note) {
		t.Error("Ransom Note was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestCmdr2S32RansomNoteCloakMode(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	lib, bf := me.Library.Size(), len(g.Battlefield.Cards)
	p7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("library lost %d cards, want 1", got)
	}
	// The Note left and the cloaked card entered.
	if got := len(g.Battlefield.Cards); got != bf {
		t.Errorf("battlefield has %d cards, want %d", got, bf)
	}
}

func TestCmdr2S32RansomNoteGoadMode(t *testing.T) {
	g, me, opp, note := cmdr2S32Note(t)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	p7Activate(t, g, me, note, 0, game.ActivateAbilityParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if c, ok := g.LookupCardForEffect(bear); !ok || c.LatestGoader() != me.ID {
		t.Fatal("the bear was not goaded")
	}
	if g.Battlefield.Contains(note) {
		t.Error("Ransom Note was not sacrificed")
	}
}

func TestCmdr2S32RansomNoteGoadNeedsACreature(t *testing.T) {
	g, me, _, note := cmdr2S32Note(t)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, note, 0, game.ActivateAbilityParams{Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("a non-creature was accepted as the goad target")
	}
	if !g.Battlefield.Contains(note) {
		t.Error("a refused activation sacrificed the Note")
	}
}

// --- Tangletrove Kelp ---------------------------------------------------

func TestCmdr2S32KelpAnimatesOtherCluesAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kelp := b12Push(g, me.ID, "Tangletrove Kelp", "Artifact Creature — Clue Plant", cmdr2S32KelpOracle, 6, 6)
	mine := pushToken(g, me.ID, ClueToken())
	theirs := pushToken(g, opp.ID, ClueToken())
	if c := tmtB1Card(t, g, mine); c.IsCreature() {
		t.Fatal("a Clue is a creature before combat")
	}
	advanceToStepInTurn(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, mine)
	if !c.IsCreature() || !c.HasSubtype("Plant") || !c.HasSubtype("Clue") {
		t.Errorf("my Clue: creature %v, types %v", c.IsCreature(), c.TypeLine)
	}
	if p, tough := mkmB1PT(t, g, mine); p != 6 || tough != 6 {
		t.Errorf("my Clue is %d/%d, want 6/6", p, tough)
	}
	if tmtB1Card(t, g, theirs).IsCreature() {
		t.Error("an opponent's Clue became a creature")
	}
	if p, _ := mkmB1PT(t, g, kelp); p != 6 {
		t.Errorf("the Kelp is %d power, want 6", p)
	}
	advanceToStepInTurn(t, g, game.StepCleanup)
	if tmtB1Card(t, g, mine).IsCreature() {
		t.Error("the Clue is still a creature after the turn ended")
	}
}

func TestCmdr2S32KelpSacrificeDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	auraSeedLibrary(me, 3)
	kelp := b12Push(g, me.ID, "Tangletrove Kelp", "Artifact Creature — Clue Plant", cmdr2S32KelpOracle, 6, 6)
	mkmAC2AddMana(t, g, me, "{C}{C}")
	hand := me.Hand.Size()
	p7Activate(t, g, me, kelp, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(kelp) {
		t.Error("the Kelp was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Armed with Proof ---------------------------------------------------

func TestCmdr2S32ArmedWithProofInvestigatesTwiceAndEquips(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Armed with Proof", "Enchantment", cmdr2S32ArmedOracle, 0, 0)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 2 {
		t.Fatalf("%d Clues on entering, want 2", got)
	}
	var clue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Clue" {
			clue = c.InstanceID
			break
		}
	}
	if !tmtB1Card(t, g, clue).HasSubtype("Equipment") {
		t.Fatal("a Clue is not an Equipment")
	}
	bear := seedBear(g, me.ID)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	beEquip(t, g, me.ID, clue, bear, "Equip {2}")
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped bear's power is %d, want 4", p)
	}
	// The enchantment leaves: the Clue stops being Equipment and falls off.
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == cmdr2S32ArmedOracle {
				if err := g.SacrificePermanentForEffect(c.InstanceID); err != nil {
					t.Fatalf("sacrifice: %v", err)
				}
				break
			}
		}
	})
	g.RunStateChecksForTest()
	if tmtB1Card(t, g, clue).HasSubtype("Equipment") {
		t.Error("the Clue is still Equipment without Armed with Proof")
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the bear's power is %d, want 2", p)
	}
}
