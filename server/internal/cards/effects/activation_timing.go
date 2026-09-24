package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_timing.go — #1208, ADR 0066's and ADR 0073's amendments
// of 2026-09-23: constructors for Spec.ActivationTimings, the
// per-player "you may activate … any time you could cast an instant"
// statics (CR 602.5d, CR 606.3, CR 101.1).
//
// The twin of cast_timing.go, and written the same way: one
// constructor per printed SHAPE rather than a generic builder,
// because the shape carries the clause a reader matches against the
// oracle text along with the predicate, and a card file that wrote
// the predicate and forgot the label ships a rule nobody can explain.
// Register refuses both halves apart.
//
// CR 101.1 IS WHY THESE CARDS WORK AT ALL. CR 602.5d is a printed
// clause and CR 606.3 is a RULE — "a loyalty ability may be activated
// only any time its controller could cast a sorcery" is nowhere on
// the card. A card that contradicts a rule wins, which is the whole
// of what The Wandering Emperor and Teferi, Master of Time do, and it
// is why the read reaches loyalty abilities rather than exempting
// them: refusing to would make the only two printed users of this
// seam unwritable.
//
// The predicate runs under g.mu inside the activation path, the bot
// enumerator and the view. Read-only: *ForEffect accessors and plain
// field reads, never a locking mutator.

// ThisSourcesLoyaltyAbilitiesAtInstantSpeed is the planeswalker
// shape: "you may activate [this permanent's] loyalty abilities any
// time you could cast an instant".
//
// Teferi, Master of Time prints it flat ("You may activate loyalty
// abilities of Teferi on any player's turn any time you could cast an
// instant"); The Wandering Emperor prints it with a condition ("As
// long as The Wandering Emperor entered this turn, …"). `when` is
// that condition and nil is "always", which is the difference between
// the two cards and nothing else.
//
// THREE NARROWINGS, and each is a clause on the card:
//
//   - LOYALTY abilities. The clause says so, and a statement that
//     did not ask would also open an equip ability the same
//     permanent happened to have.
//   - THIS permanent. Both cards name themselves (CR 201.5), so the
//     statement is about the object that declares it and no other —
//     a second copy of Teferi opens its own abilities, not yours.
//   - ITS CONTROLLER. "You" is the permanent's controller, so a
//     stolen Teferi opens his abilities for the thief, which is
//     also who may activate them (CR 606.3's control clause).
//
// "On any player's turn" is NOT a fourth narrowing and is not
// modelled: it is what "any time you could cast an instant" already
// means, spelled out for players. The read has no turn test on the
// grant side at all.
func ThisSourcesLoyaltyAbilitiesAtInstantSpeed(label string, when func(g *game.Game, source game.Card) bool) game.ActivationTiming {
	return game.ActivationTiming{
		Label:  label,
		Timing: game.TimingFlash,
		Covers: func(q game.ActivationQuery) bool {
			if !q.Ability.Loyalty {
				return false
			}
			if q.Card.InstanceID != q.Source.InstanceID {
				return false
			}
			if q.Controller != q.Source.Controller {
				return false
			}
			if when != nil && !when(q.Game, q.Source) {
				return false
			}
			return true
		},
	}
}

// LoyaltyAbilitiesOfYourPlaneswalkersAtInstantSpeed is the EMBLEM
// shape (#1275): "You may activate loyalty abilities of planeswalkers
// you control on any player's turn any time you could cast an
// instant" — Teferi, Temporal Archmage's −10, and the −12 Teferi's
// Talent grants.
//
// ThisSourcesLoyaltyAbilitiesAtInstantSpeed with the self-reference
// widened to a control test, because that is the whole difference in
// the text: Teferi, Master of Time names HIMSELF, the emblem names
// every planeswalker its owner controls. Declared on
// `EmblemSpec.ActivationTimings`, so `q.Source` is the emblem and
// `q.Source.Controller` is its owner (CR 114.2 — an emblem's
// controller is the player who has it, and it never changes).
//
// THREE NARROWINGS again, each a clause on the card:
//
//   - LOYALTY abilities — an equip ability on a planeswalker-turned-
//     creature stays at its printed window.
//   - PLANESWALKERS — the object activating must be one right now.
//   - YOU CONTROL — the planeswalker's controller is the emblem's
//     owner, and so is the activator. A planeswalker you steal is
//     opened for you; one an opponent steals from you is not opened
//     for them, because the emblem is still yours.
//
// "Only one loyalty ability per planeswalker per turn" (CR 606.3's
// other half) is untouched: this opens the WINDOW, and the tally in
// `Game.LoyaltyActivatedThisTurn` is read after it.
func LoyaltyAbilitiesOfYourPlaneswalkersAtInstantSpeed(label string) game.ActivationTiming {
	return game.ActivationTiming{
		Label:  label,
		Timing: game.TimingFlash,
		Covers: func(q game.ActivationQuery) bool {
			if !q.Ability.Loyalty {
				return false
			}
			if !q.Card.IsPlaneswalker() {
				return false
			}
			if q.Card.Controller != q.Source.Controller {
				return false
			}
			return q.Controller == q.Source.Controller
		},
	}
}

// EquipAbilitiesAtInstantSpeed is Leonin Shikari: "You may activate
// equip abilities any time you could cast an instant."
//
// Narrowed on the ABILITY rather than on its source, which is the
// half of this seam `game.PermissionFilter` could not have said: the
// clause is about a KEYWORD (CR 702.6), and the Equipment it is
// printed on could be any artifact on the board. That is why
// `ActivationAbility` grew `Equip` and why a filter over card types
// would have been the wrong tool — see activation_timing.go in the
// game package.
//
// "You" is the SHIKARI's controller and the activator both: the
// clause says "you may activate", so it opens nothing for an
// opponent, and CR 702.6a's window is the activator's either way.
//
// `onlyDuringYourTurn` is Forge Anew's extra clause ("DURING YOUR
// TURN, you may activate equip abilities any time you could cast an
// instant") measured against the source's controller, the same seat
// `OpponentsSourcesCantActivateDuringYourTurn` measures its "your
// turn" against. False is Leonin Shikari, who says it always.
func EquipAbilitiesAtInstantSpeed(label string, onlyDuringYourTurn bool) game.ActivationTiming {
	return game.ActivationTiming{
		Label:  label,
		Timing: game.TimingFlash,
		Covers: func(q game.ActivationQuery) bool {
			if !q.Ability.Equip {
				return false
			}
			if q.Controller != q.Source.Controller {
				return false
			}
			if onlyDuringYourTurn && !isActivePlayer(q.Game, q.Source.Controller) {
				return false
			}
			return true
		},
	}
}

// SourceEnteredThisTurn is The Wandering Emperor's "as long as [this]
// entered this turn" condition, in the shape
// ThisSourcesLoyaltyAbilitiesAtInstantSpeed takes.
//
// game.EnteredThisTurn and NOT Card.SummonedThisTurn, and the
// difference is the whole of the clause: the summoning-sickness
// marker is cleared at its controller's untap step, so a permanent
// that entered during an opponent's turn still carries it on yours
// (CR 302.6) and DID NOT enter this turn. Using it would have given
// the Emperor instant-speed loyalty abilities for a whole turn cycle
// instead of for the turn she landed on — which, since she has flash
// and lands on somebody else's turn, is every time.
func SourceEnteredThisTurn(g *game.Game, source game.Card) bool {
	return g != nil && g.EnteredThisTurn(source.InstanceID)
}
