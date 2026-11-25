package game

import "github.com/google/uuid"

// scoped_enters_under_control.go — #6260, Gather Specimens: "If a
// creature would enter the battlefield under an opponent's control this
// turn, it enters under your control instead."
//
// A CR 614.12 replacement that modifies how a permanent enters, created
// by a resolving spell (CR 611.2) and held as a scoped record for the
// turn. It changes under whose control the creature enters, so it is in
// CR 616.1b's tier (ReplacementEffect.ChangesEntryController): applied
// before every other effect in the window, and the rest are gathered
// against the new controller. Any creature entering counts: a spell
// resolving, a reanimation, a flicker, a token (whose creation an
// opponent's control has already named, CR 111.2 — the token enters as
// any permanent does). The would-be controller is read off the entering
// card the way the entry-controller choice writes it
// (inZoneControllerLocked), so a Captive Audience-style choice made
// earlier in the window is what this sees. The mod kind is an on-disk
// identity: never renamed, never reused.

// ModCreaturesEnterUnderYourControl is Gather Specimens' replacement for
// the record's Controller ("you"). Reads nothing else.
const ModCreaturesEnterUnderYourControl ModKind = "creaturesEnterUnderYourControl"

// CreaturesEnterUnderYourControlThisTurnForEffect registers the
// replacement for `controller` until end of turn. Returns false,
// registering nothing, for a player who is not in the game.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) CreaturesEnterUnderYourControlThisTurnForEffect(sourceID, controller uuid.UUID, label string) bool {
	if g.playerByIDLocked(controller) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModCreaturesEnterUnderYourControl}}, g.UntilEndOfTurnDuration(), label)
}

// creaturesEnterUnderYourControlAppliesLocked is the kind's test: a
// creature moving onto the battlefield under the control of a player
// other than the record's controller (every other player is an opponent
// at a free-for-all table). Caller must hold g.mu.
func creaturesEnterUnderYourControlAppliesLocked(g *Game, e ScopedEffect, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled || e.Controller == uuid.Nil {
		return false
	}
	if !g.playerInGameLocked(e.Controller) {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		if entering, ok = g.enteringTokenLocked(ev.CardID); !ok {
			return false
		}
	}
	who := g.inZoneControllerLocked(ev)
	if who == uuid.Nil {
		who = entering.Controller
	}
	return who != e.Controller && entering.IsCreature()
}
