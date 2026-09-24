package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shocklands.go — the Ravnica "shockland" cycle, all ten:
//
//	"({T}: Add {X} or {Y}.)"
//	"As this land enters, you may pay 2 life. If you don't, it
//	 enters tapped."
//
// The mana ability is declared rather than left to the synthetic
// basic-land shape: a shockland is a NONBASIC land that happens to
// carry two basic land types, and the reminder-text ability comes
// from those types. Declaring it is also what lets the pipe syntax
// offer one picker instead of two menu entries.
//
// # The entry is replaced, not patched up afterwards
//
// This cycle first shipped (PR #258) as a declared simplification:
// the land entered tapped and an optional ETB trigger untapped it
// for 2 life. The choice was right and everything around it was
// wrong — a tapped window between entering and the trigger
// resolving, an untap event nothing should have seen, and a trip
// through the stack that handed opponents priority in the middle of
// something that is not a trigger at all.
//
// It is now a real CR 614 self-replacement with the decision inside
// it. EntryLifeCost tells the apply-loop to stop and ask the
// controller before anything moves (see entry_choice.go): paying
// means the replacement never fires and the land enters untapped;
// declining — or being unable to pay, CR 119.4 — fires it and the
// land ENTERS tapped. No tapped window, no untap event, no priority
// pass. The tests count tap events rather than reading Tapped,
// because that is the only thing that tells the two implementations
// apart.
//
// # A FETCHED shockland gets the choice too (#478)
//
// The prompt was once offered only on the land-play path (from hand,
// or from an impulse exile). A shockland put onto the battlefield by
// an EFFECT — a fetchland cracking for it, a Farseek, a reanimation,
// a blink — ran this same replacement (#263 routed those paths
// through the CR 614 pipeline) but could not PAUSE to ask, so it took
// the un-paid branch and entered tapped.
//
// Those entry sites are entryResumable now: what the effect still
// owed (the search's shuffle, EventSearchLibrary and its `Then`; the
// exile return's new object identity) rides across the pause on
// ReplacementEvent.entryTail, so the payment is offered and the fetch
// finishes when it is answered. See entry_tail.go.
//
// # And so does one a spell PUTS onto the battlefield (#1322)
//
// The last entry site without a resume was the hand / library "put"
// batch — Warp World, Genesis Wave, Coiling Oracle, Arboreal Grazer —
// which runs every card's window against the pre-entry board and lands
// them together, so a per-card resume would have broken the
// simultaneity it exists for. The caveat said so ("a shockland a spell
// PUTS onto the battlefield always enters tapped"). The batch is now a
// value that asks one card's question at a time and lands everything
// once the last answer is in (server/internal/game/entry_batch.go), so
// the caveat is gone and the cycle is complete.
//
// TestEveryShocklandEntrySiteOffersThePayment (shocklands_test.go)
// holds every entry site against the engine: the land play, the
// search, the library put and the hand put all ask.

// EntersTappedUnlessYouPayLife is "as this permanent enters, you may
// pay N life. If you don't, it enters tapped." The prompt is queued
// by the replacement pipeline before the permanent moves, so the
// answer decides how it ENTERS.
//
// The Controller hook names the payer — EnteringPermanentChooser
// (helpers.go), shared with the copy selector and the reveal-lands:
// the player performing the play (ev.Actor, which the land path
// stamps) is the shockland case; the card's own controller / owner is
// the fallback for an entry driven by something else, such as an
// effect putting the land onto the battlefield.
func EntersTappedUnlessYouPayLife(name string, life int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		EntryLifeCost:   life,
		Label:           name,
		PromptQuestion:  fmt.Sprintf("%s — pay %d life so it enters untapped?", name, life),
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
	}
}

// shocklandLifeCost is the payment every member of the cycle asks
// for. Named rather than inlined so the prompt copy and the cost can
// never drift apart.
const shocklandLifeCost = 2

func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"43985bbc-a0f6-4812-984e-392bc8562633", "Blood Crypt", "B", "R"},
		{"20283c4a-f1f0-42f0-bc08-6da87474426b", "Breeding Pool", "G", "U"},
		{"73864fcc-1bde-4bc0-831e-2b93e546e417", "Godless Shrine", "W", "B"},
		{"f1750962-a87c-49f6-b731-02ae971ac6ea", "Hallowed Fountain", "W", "U"},
		{"975ec9a3-6f20-4177-8211-82526e092538", "Overgrown Tomb", "B", "G"},
		{"45181cb8-2090-4471-ba90-e5a8f04d525f", "Sacred Foundry", "R", "W"},
		{"17039058-822d-409f-938c-b727a366ba63", "Steam Vents", "U", "R"},
		{"16052b52-ade1-406f-a06b-ce7ea607fb63", "Stomping Ground", "R", "G"},
		{"f413a83d-a40d-434c-b20a-4c707c0527fa", "Temple Garden", "G", "W"},
		{"fc9ec820-4245-4a96-b009-5308a818ca58", "Watery Grave", "U", "B"},
	} {
		// Capture per iteration: the closures below outlive the loop
		// body, and a shared loop variable would give every shockland
		// the last entry's name.
		name := t.name
		Register(Spec{
			OracleID:     t.oracleID,
			Name:         name,
			Completeness: CompletenessFull,
			Replacements: []game.ReplacementEffect{
				EntersTappedUnlessYouPayLife(name, shocklandLifeCost),
			},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
