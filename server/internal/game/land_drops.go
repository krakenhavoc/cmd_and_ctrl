package game

import "github.com/google/uuid"

// land_drops.go — how many lands a player may play this turn, and
// how many they have played (#500).
//
// CR 305.2: "A player can normally play one land during each of
// their turns." Until #500 the engine counted land plays
// (Game.LandsPlayedThisTurn) but refused nothing: the legal-move
// enumerator in internal/legal was the only thing that consulted the
// tally, so a client that did not ask the enumerator — or a player
// hand-sending an action — could play the whole hand as lands.
//
// The owner's decision on #500 is to ENFORCE it: "We are looking to
// get away from the cockatrice style sandbox. That is only supposed
// to be as a fallback method when the engine is not working." The
// gate lives in CastSpell's land branch (mutations.go), which is the
// one place a land is PLAYED; effects that PUT a land onto the
// battlefield (CR 305.4) are not land plays and are untouched.
//
// "One" is not hardcoded. The effective allowance is a sum of three
// sources, each with a distinct job:
//
//   - Player.LandDropsPerTurn — the BASE, DefaultLandDropsPerTurn on
//     a freshly seated player. The sandbox / admin knob, and the
//     field a future set_land_drops action would write.
//   - CatalogAdditionalLandPlays — static abilities of permanents
//     the player controls. Exploration (+1) and Azusa, Lost but
//     Seeking (+2) are this shape, and the reason the hook is
//     DERIVED rather than written into the player: two Explorations
//     have to compose, and one of them leaving must not restore an
//     allowance the other is still granting. Same argument, verbatim,
//     as CatalogNoMaxHandSize (#338) — see effect_hooks.go.
//   - Game.ExtraLandDropsThisTurn — one-shot grants that expire with
//     the turn ("you may play an additional land this turn"), written
//     by GrantAdditionalLandPlayForEffect and cleared in
//     onTurnBeganLocked with the other per-turn state.
//
// No catalog card declares AdditionalLandPlays yet. The hook exists
// so that when Exploration and Azusa are written they are a one-line
// Spec field and not an engine change.

// DefaultLandDropsPerTurn is the land plays a player gets in a normal
// turn (CR 305.2). Raised per player by Player.LandDropsPerTurn, by a
// controlled permanent's AdditionalLandPlays, or for one turn by
// Game.ExtraLandDropsThisTurn.
const DefaultLandDropsPerTurn = 1

// CatalogAdditionalLandPlays returns how many EXTRA lands per turn a
// battlefield permanent with the given catalog key lets its
// controller play — 1 for Exploration, 2 for Azusa, 0 (or a nil
// hook) for everything else. Populated at init time by the
// cards/effects package from `effects.Spec.AdditionalLandPlays`.
//
// Player-scoped continuous effect, so it is a derived hook rather
// than a CR 613 layer, for the reasons CatalogNoMaxHandSize spells
// out: the layer engine models characteristics of objects and "you
// may play an additional land" is not one.
var CatalogAdditionalLandPlays func(oracleID string) int

// EffectiveLandDropsLocked returns how many lands p may play in
// total this turn: their base allowance, plus every controlled
// permanent's static grant, plus any one-turn grant. Never negative —
// a nonsense base of -3 reads as zero rather than as a credit against
// the statics.
//
// Caller must hold g.mu (read or write).
func (g *Game) EffectiveLandDropsLocked(p *Player) int {
	if p == nil {
		return 0
	}
	n := p.LandDropsPerTurn
	if CatalogAdditionalLandPlays != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != p.ID {
				continue
			}
			// CatalogAbilityKey, not CatalogKey: "you may play an
			// additional land on each of your turns" is a static
			// ability, and an Exploration that has lost its abilities
			// grants nothing. The empty key — not an empty oracle ID
			// — is the skip, so a TOKEN with a catalog key of its own
			// is walked like any other permanent (ADR 0083 decision 3).
			key := CatalogAbilityKey(*c)
			if key == "" {
				continue
			}
			n += CatalogAdditionalLandPlays(key)
		}
	}
	n += g.ExtraLandDropsThisTurn[p.ID]
	if n < 0 {
		return 0
	}
	return n
}

// LandDropsRemainingLocked returns how many more lands playerID may
// play this turn — the effective allowance minus what they have
// already played, floored at zero. Zero for a player who is not
// seated.
//
// Caller must hold g.mu (read or write).
func (g *Game) LandDropsRemainingLocked(playerID uuid.UUID) int {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return 0
	}
	n := g.EffectiveLandDropsLocked(p) - g.LandsPlayedThisTurnFor(playerID)
	if n < 0 {
		return 0
	}
	return n
}

// LandDropsRemainingFor is LandDropsRemainingLocked for callers
// outside a locked frame — the legal-move enumerator and the
// snapshot view builder both ask it whether to offer a land play.
func (g *Game) LandDropsRemainingFor(playerID uuid.UUID) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LandDropsRemainingLocked(playerID)
}

// GrantAdditionalLandPlayForEffect gives playerID n more land plays
// for the REST OF THIS TURN — "you may play an additional land this
// turn" (Explore, Azusa's one-shot cousins, Ancient Greenwarden's
// neighbours). The grant expires at the next turn boundary with the
// rest of the per-turn state; a permanent's standing "on each of your
// turns" grant belongs in the catalog's AdditionalLandPlays instead,
// which is re-derived every turn and needs no expiry.
//
// n <= 0 is a no-op rather than a way to take land plays away: no
// card in the catalog removes one, and a negative grant that outlived
// the effect that made it would be invisible and unexplainable at the
// table.
//
// Caller must hold g.mu (write).
func (g *Game) GrantAdditionalLandPlayForEffect(playerID uuid.UUID, n int) {
	if playerID == uuid.Nil || n <= 0 {
		return
	}
	if g.ExtraLandDropsThisTurn == nil {
		g.ExtraLandDropsThisTurn = make(map[uuid.UUID]int)
	}
	g.ExtraLandDropsThisTurn[playerID] += n
}
