package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_until_another_test.go — #2539, ADR 0066 amendment 2026-10-08:
// Unstable Amulet, Furious Rise and Superior Foes of Spider-Man, whose
// exiled card stays playable "until you exile another card with this".
// The engine half is game/exile_until_another_test.go.

const (
	EuaUnstableAmulet = "f3d4f5ab-0ff4-48e1-8cf0-163b290804a5"
	EuaFuriousRise    = "52b3a979-3b93-4441-9237-0c8a754e5523"
	EuaSuperiorFoes   = "a6a7af21-e52c-4f7f-a57f-bd365d075966"
)

// euaTop puts a fresh card on top of p's library and returns it.
func EuaTop(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// euaPlayable asks the one predicate the cast path, the view and the
// bot all ask: may p play card `id` from exile right now?
func EuaPlayable(g *game.Game, p *game.Player, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				ok = g.CastPermissionForLocked(p.ID, c, game.ZoneExile) != nil
			}
		}
	})
	return ok
}

// euaAmulet is a catalog table with an untapped Unstable Amulet for the
// active seat, and enough energy for `activations` activations.
func EuaAmulet(t *testing.T, activations int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := P7Table(t)
	amulet := EbPush(g, me, "Unstable Amulet", EuaUnstableAmulet, "Artifact", 0, 0)
	SetEnergy(t, g, me, 2*activations)
	return g, me, opp, amulet
}

// euaActivate untaps the Amulet, activates it and resolves the ability.
func EuaActivate(t *testing.T, g *game.Game, me *game.Player, amulet uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { FindBattlefieldCardForTest(g, amulet).Tapped = false })
	P7Activate(t, g, me, amulet, 0, game.ActivateAbilityParams{})
}
