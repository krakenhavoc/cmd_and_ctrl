package game

import (
	"fmt"

	"github.com/google/uuid"
)

// scoped_permanents_enter_tapped.go — "<permanents> enter tapped this
// turn" (#6147): Due Respect's "Permanents enter tapped this turn" and
// Nahiri's Lithoforming's "Lands you control enter tapped this turn".
//
// A CR 614.1d replacement effect ("[Objects] enter . . .") a resolving
// spell creates (CR 611.2), held as a scoped record so it outlives the
// spell and ends with its duration (cleanup, CR 514.2), the general
// sibling of #6093's
// creaturesEnterTappedForOpponents (scoped_enters_tapped.go). The set is
// read live, as each permanent enters: a replacement effect is not a
// characteristic, so CR 611.2c does not lock it to the objects that were
// there when it began. The mod kind is an on-disk identity: never
// renamed, never reused.

// ModPermanentsEnterTapped is "<permanents> enter tapped" for as long as
// the record lasts. Mod.Types narrows it to permanents with at least one
// of those card types (lower case, "land"); empty is every permanent.
// Mod.Player narrows it to permanents entering under that player's
// control; uuid.Nil is any player's. A replacement, read by the
// replacement gather rather than the layer pass.
const ModPermanentsEnterTapped ModKind = "permanentsEnterTapped"

// permanentCardTypes are the card types a permanent can have (CR 110.4),
// the only ones a ModPermanentsEnterTapped filter may name.
var permanentCardTypes = map[string]bool{
	"artifact": true, "battle": true, "creature": true, "enchantment": true,
	"land": true, "planeswalker": true,
}

// permanentsEnterTappedModProblem is the registration and restore check
// for ModPermanentsEnterTapped. "" when the mod is sound or is another
// kind.
func permanentsEnterTappedModProblem(m Mod) string {
	if m.Kind != ModPermanentsEnterTapped {
		return ""
	}
	for _, t := range m.Types {
		if !permanentCardTypes[t] {
			return fmt.Sprintf("a permanentsEnterTapped mod names %q, not a permanent card type", t)
		}
	}
	return ""
}

// PermanentsEnterTappedThisTurnForEffect registers "<permanents> enter
// tapped this turn": every permanent with one of `types` (lower-case
// card types; none is every permanent) entering under `controller`'s
// control (uuid.Nil: anyone's) enters tapped until cleanup. "Permanents
// enter tapped this turn" (Due Respect) is (uuid.Nil, nil); "Lands you
// control enter tapped this turn" (Nahiri's Lithoforming) is (you,
// "land"). Returns false, registering nothing, for a controller who is
// not in the game.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) PermanentsEnterTappedThisTurnForEffect(sourceID, controller uuid.UUID, types []string, label string) bool {
	if controller != uuid.Nil && g.playerByIDLocked(controller) == nil {
		return false
	}
	m := Mod{Kind: ModPermanentsEnterTapped, Player: controller, Types: append([]string(nil), types...)}
	if problem := permanentsEnterTappedModProblem(m); problem != "" {
		panic(fmt.Sprintf("game: %s: %s", label, problem))
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{m}, g.UntilEndOfTurnDuration(), label)
}

// permanentsEnterTappedAppliesLocked is the kind's test: a permanent
// moving onto the battlefield, under the mod's player's control when it
// names one, with one of the mod's card types when it names any. The
// controller it enters under is read as the printed forms read it
// (LookupCardForEffect on the entering card). Caller must hold g.mu.
func permanentsEnterTappedAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled || ev.EntersTapped {
		return false
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	if m.Player != uuid.Nil && entering.Controller != m.Player {
		return false
	}
	if len(m.Types) == 0 {
		return true
	}
	for _, t := range m.Types {
		if entering.HasCardType(t) {
			return true
		}
	}
	return false
}
