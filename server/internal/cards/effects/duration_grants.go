package effects

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// duration_grants.go — ADR 0093 PR 4 (#1584): an ability a RESOLVING
// spell or ability gives a permanent for a while.
//
//	"Until end of turn, target creature gains 'When this creature
//	 dies, return it to the battlefield tapped under its owner's
//	 control.'"                                        (Malakir Rebirth)
//	"Until end of turn, target creature gains '{T}: Return target
//	 nonland permanent to its owner's hand.'"         (Retraction Helix)
//	"I — This Saga gains '{T}: Add {C}.'"                  (Urza's Saga)
//
// The ability is a catalog BUNDLE the card declares in Spec.Grants,
// exactly as a Cryptolith Rite declares its own; what differs is only
// who says "has". A granting permanent's static says it for as long as
// the permanent is there (GrantAbilities). A resolving spell says it
// once, with a duration, and that is ADR 0041 phase 3's ScopedEffect
// record with a `grantAbilities` mod (game.GrantAbilitiesMod) — DATA,
// so a table holding one is still a restore point, and the layer pass
// turns it into the same layer-6 declaration a static makes. Every
// reader downstream — the `grant:` refs, the ability menus, the
// auto-tapper, the trigger harvest, CR 613.6 and CR 707.2 — is the
// static grant's code, unchanged.
//
// The affected set is locked at resolution (CR 611.2c): a creature
// that leaves and comes back is a new object without the ability
// (CR 400.7). That is Feign Death's whole design — the returned
// creature does not return again.

// GrantAbilitiesFor is "<affected> gains '<ability>' <duration>" from
// a resolving spell or ability:
//
//	return GrantAbilitiesFor{
//	    Target: t,
//	    Keys:   []string{feignDeathReturn},
//	    Label:  "Feign Death — the creature gains the return",
//	}.Apply(ctx)
//
// Keys name bundles some card declares in Spec.Grants; either
// spelling of game.GrantKey works. TestEveryDurationGrantKeyResolves
// holds every key written in the catalog to a registered bundle at
// build time, and Apply refuses an unregistered one at resolution.
type GrantAbilitiesFor struct {
	// Target pins the grant to one permanent. Ignored when Match is
	// set.
	Target uuid.UUID

	// Match selects the affected permanents, evaluated ONCE, now
	// (CR 611.2c) — "creatures you control gain … until end of turn".
	Match CardPredicate

	// Keys are the bundles granted.
	Keys []string

	// Also is the rest of the SAME effect, applied at the same
	// timestamp: Fake Your Own Death's "gets +2/+0 and gains …" is
	// Also: []game.Mod{game.ModifyPTMod(2, 0)}. One sentence, one
	// effect, one CR 613.7 timestamp. The grant is applied after
	// these, so a "loses all abilities and has '…'" in Also empties
	// the list before the grant is written (ADR 0046 §2).
	Also []game.Mod

	// Duration is how long the grant lasts (CR 611.2). Leave it zero
	// for "until end of turn"; otherwise build it with one of the
	// Duration* builders (durations.go) or game.IndefiniteDuration
	// pinned to its object for a grant with no stated duration
	// (CR 611.2a).
	Duration game.Duration

	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the grant. Nothing matched registers nothing; a key
// no card registered is an error, because a grant of an unknown bundle
// would be a creature that silently has no new ability.
//
// Caller must be inside the resolution frame (holds g.mu write).
func (s GrantAbilitiesFor) Apply(ctx *Context) error {
	if len(s.Keys) == 0 {
		return fmt.Errorf("effects: GrantAbilitiesFor %q names no ability bundle", s.Label)
	}
	d := s.Duration
	if d == (game.Duration{}) {
		d = DurationUntilEndOfTurn(ctx)
	}
	mods := make([]game.Mod, 0, len(s.Also)+1)
	mods = append(mods, s.Also...)
	mods = append(mods, game.GrantAbilitiesMod(s.Keys...))
	return ScopedEffectFor{
		Target:   s.Target,
		Match:    s.Match,
		Mods:     mods,
		Duration: d,
		Label:    eotLabel(s.Label, "ability grant"),
	}.Apply(ctx)
}

// checkGrantMods refuses a grantAbilities mod naming a bundle the
// catalog does not register, or one with a Static slot — ADR 0093
// Decision 10: the layer pass gathers every static before layer 1, so
// a static that only exists after layer 6 would never be gathered. The
// same refusal TestEveryGrantKeyResolves makes of a granting static.
func checkGrantMods(mods []game.Mod) error {
	for _, m := range mods {
		if m.Kind != game.ModGrantAbilities {
			continue
		}
		if len(m.Grants) == 0 {
			return fmt.Errorf("effects: a grantAbilities mod names no ability bundle")
		}
		for _, k := range m.Grants {
			if err := durationGrantKeyProblem(k); err != "" {
				return fmt.Errorf("effects: %s", err)
			}
		}
	}
	return nil
}

// durationGrantKeyProblem is the one reading of "may a resolving
// effect grant this bundle": registered, and without a Static slot.
// Empty when it may. Shared by the resolution-time check and the
// build-time scan.
func durationGrantKeyProblem(key string) string {
	def := defs[game.GrantKey(key)]
	if def == nil {
		return fmt.Sprintf("ability grant %q names no bundle any card registers in Spec.Grants", key)
	}
	if len(def.Static) > 0 {
		return fmt.Sprintf("ability grant %q names a bundle with a Static slot — a layer-6 grant cannot give a static ability (ADR 0093 Decision 10)", key)
	}
	return ""
}

// returnThisCreatureFromGraveyard is the body of the granted "When
// this creature dies, return it to the battlefield tapped under its
// owner's control [with <counters> on it]" — Feign Death, Fake Your
// Own Death and Malakir Rebirth all grant it. The trigger is the
// HOST's (ADR 0093 Decision 4), so "it" is the trigger's source: the
// card that died, now in its owner's graveyard. A token is gone
// (CR 111.7), and a card that has already left the graveyard is a new
// object the ability cannot find (CR 400.7), so both return nothing —
// ErrCardNotFound is swallowed, the CR 608.2b posture.
//
// `then` is the rest of the sentence ("and you create a Treasure
// token") and runs whether or not the card came back: it is joined by
// "and", not "if you do".
func returnThisCreatureFromGraveyard(counters map[string]int, then func(ctx *Context) error) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneGraveyard {
			if _, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, uuid.Nil, true, counters); err != nil && !errors.Is(err, game.ErrCardNotFound) {
				return err
			}
		}
		if then != nil {
			return then(ctx)
		}
		return nil
	}
}
