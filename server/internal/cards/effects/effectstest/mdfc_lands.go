package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mdfc_lands_test.go — the three bugs ADR 0034 set out to close,
// written as the reproduction first and the fix second.
//
//  1. Sea Gate Restoration cast as a SORCERY costs {4}{U}{U}{U}.
//     Before the face model its imported ManaCost was "" (Scryfall
//     nulls the top-level cost on every modal DFC), ParseCost("")
//     succeeded with the zero cost, and the spell was FREE.
//  2. Sea Gate, Reborn PLAYED as a land enters tapped unless you pay
//     3 life, with a real prompt and a real entry replacement.
//  3. A creature-front MDFC — Kazandu Mammoth — is no longer a free
//     land drop. Its top-level type line is "Creature — Elephant //
//     Land", Card.IsLand() is a substring scan, and CastSpell's land
//     branch RETURNS BEFORE THE COST GATE. Every ZNR MDFC in the
//     game was a one-mana-free Elephant that skipped the stack.
//
// Each fixture is imported the way a real card is — through the
// per-face data — rather than assembled by hand, so what is being
// tested is the whole chain from Scryfall shape to battlefield.

const (
	// Oracle IDs from the Scryfall default-cards dump.
	SeaGateRestorationOracle = "4a8d41fe-e04d-484b-a7d1-19be311e6ca7"
	KazanduMammothOracle     = "2ac1c95c-2a9d-40bc-9cad-9cadfa3f19f7"
)

// mdfcCard builds a two-faced game.Card the way deck.toGameCard
// does: per-face printed data, then SetFace(0) to materialise the
// front. The flat fields are deliberately NOT set here — if the face
// model is not carrying them, the card arrives blank and every
// assertion below fails loudly rather than silently passing on a
// hand-stamped value.
func MdfcCard(owner uuid.UUID, oracleID, layout string, faces ...game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracleID,
		Layout:     layout,
		Faces:      faces,
		Owner:      owner,
		Controller: owner,
	}
	c.SetFace(0)
	return c
}

// seaGateRestoration is the ZNR mythic: a seven-mana sorcery whose
// back is a pay-3-life land.
func SeaGateRestoration(owner uuid.UUID) game.Card {
	return MdfcCard(owner, SeaGateRestorationOracle, game.LayoutModalDFC,
		game.Face{
			Name:     "Sea Gate Restoration",
			TypeLine: "Sorcery",
			ManaCost: "{4}{U}{U}{U}",
			Colors:   []string{"U"},
		},
		game.Face{
			Name:     "Sea Gate, Reborn",
			TypeLine: "Land",
			ManaCost: "",
		},
	)
}

// handWithMDFC seats a game in a main phase with `card` in the
// active player's hand, and returns the game, the player and the
// card's instance ID.
func HandWithMDFC(t *testing.T, build func(uuid.UUID) game.Card) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := build(me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return g, me, c.InstanceID
}
