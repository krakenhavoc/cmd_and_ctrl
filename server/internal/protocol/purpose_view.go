package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// purpose_view.go — ADR 0126 §6: game.Purpose on the wire.
//
// A purpose is catalog data, declared on the card file, never derived:
// the printed amounts a spell, a mode, an alternative cost or an
// ability declares it does. The bot reads it from here and nowhere else
// (ADR 0033 §3: a policy reads the view, not the catalog).
//
// Where it rides, and who sees it:
//
//   - CardView.purpose — the spell, or a permanent's enters effect. On
//     the hand, the battlefield and the stack, the zones ability_rows
//     and the cast surface use. Cleared with ability_rows for a viewer
//     who cannot see the card (redactCardForViewer), so an opponent's
//     hand card and a face-down object never carry one.
//   - ModeOptionView.purpose and AlternativeCostView.purpose — beside
//     the bullet's label and the offer's label, which say the same
//     thing in words, and hidden exactly when they are.
//   - ActivatedAbilityView.purpose and AbilityRowView.purpose — the
//     row's own, public with the row.

// PurposeView is game.Purpose on the wire. Every field is omitted when
// zero, and a purpose with nothing declared is not sent at all.
type PurposeView struct {
	// Draws is the cards its controller draws (on an any-player row,
	// the activator).
	Draws int `json:"draws,omitempty"`
	// ControllerLosesLife is the life the SOURCE'S controller loses.
	// Any-player activated rows only (ADR 0106 §1 decision 8).
	ControllerLosesLife int `json:"controller_loses_life,omitempty"`
	// Discards is the cards its controller discards on resolution.
	Discards int `json:"discards,omitempty"`
	// Lands is the land cards it puts onto the battlefield.
	Lands int `json:"lands,omitempty"`
	// LandsUntapped is how many of those lands enter untapped, so their
	// mana is there the turn it resolves (ADR 0136 §2). Never more than
	// Lands.
	LandsUntapped int `json:"lands_untapped,omitempty"`
	// Tutors is the cards it searches out to hand or to the top of the
	// library.
	Tutors int `json:"tutors,omitempty"`
	// SelfMillTutor is the cards it searches out into the graveyard.
	SelfMillTutor int `json:"self_mill_tutor,omitempty"`
	// Tokens is the tokens it creates for its controller.
	Tokens int `json:"tokens,omitempty"`
	// Energy is the energy counters it gives its controller (ADR 0129
	// §7).
	Energy int `json:"energy,omitempty"`
	// Sweep is present on a board wipe.
	Sweep *SweepView `json:"sweep,omitempty"`
	// DeathPayoff is set on a triggered row that pays out whenever a
	// creature its controller controls dies.
	DeathPayoff bool `json:"death_payoff,omitempty"`
	// DiscardPayoff is present on a triggered row that pays out
	// whenever its controller discards a card it matches (ADR 0126's
	// amendment of 2026-10-06).
	DiscardPayoff *DiscardPayoffView `json:"discard_payoff,omitempty"`
	// Pump is what the row gives its own source until end of turn
	// (ADR 0130's amendment of 2026-10-07).
	Pump *PumpView `json:"pump,omitempty"`
	// ExtraCombat is the additional combat phases it adds.
	ExtraCombat int `json:"extra_combat,omitempty"`
	// PreventCombatDamageToSelf: it prevents all combat damage that
	// would be dealt to its source this turn.
	PreventCombatDamageToSelf bool `json:"prevent_combat_damage_to_self,omitempty"`
	// DamageToCreature is the damage it deals to one target creature.
	DamageToCreature int `json:"damage_to_creature,omitempty"`
	// DamageEachOpponent is the damage it deals to each opponent.
	DamageEachOpponent int `json:"damage_each_opponent,omitempty"`
	// LifeGain is the life its controller gains.
	LifeGain int `json:"life_gain,omitempty"`
	// AwakenLand is the N of "Awaken N—[cost]": the +1/+1 counters put
	// on a land its controller controls as it becomes a 0/0 Elemental
	// creature with haste (ADR 0135 §3). On an awaken offer only.
	AwakenLand int `json:"awaken_land,omitempty"`
	// Targets is what happens to each target the statement names, one
	// entry per target clause, keyed by the clause's slot: the slot a
	// move's `targets[].slot` names, within the statement this purpose
	// rides on (ADR 0126's amendment of 2026-10-08). The amounts above
	// are the controller's; an entry's are its target's. Absent when
	// empty. A pointer so PurposeView stays comparable.
	Targets *[]TargetPurposeView `json:"targets,omitempty"`
}

// TargetPurposeView is game.TargetPurpose on the wire: what the spell
// or ability does to the target picked for clause `slot`. Every amount
// is omitted when zero; `slot` is always sent.
type TargetPurposeView struct {
	Slot     int `json:"slot"`
	Draws    int `json:"draws,omitempty"`
	Discards int `json:"discards,omitempty"`
	Tokens   int `json:"tokens,omitempty"`
	LifeGain int `json:"life_gain,omitempty"`
	LifeLoss int `json:"life_loss,omitempty"`
	Damage   int `json:"damage,omitempty"`
}

// PumpView is game.Pump on the wire: a self pump until end of turn.
type PumpView struct {
	Power     int      `json:"power,omitempty"`
	Toughness int      `json:"toughness,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
}

// DiscardPayoffView is game.DiscardPayoff on the wire: which discarded
// cards the row pays on, and what it pays for each one.
type DiscardPayoffView struct {
	// Any: every card its controller discards pays.
	Any bool `json:"any,omitempty"`
	// Types: otherwise, the lowercase card types and subtypes it pays
	// on; a card with any one of them on its type line matches.
	Types []string `json:"types,omitempty"`
	// Tokens is the tokens it creates for its controller per card.
	Tokens int `json:"tokens,omitempty"`
	// Counters is the +1/+1 counters it puts on its source per card.
	Counters int `json:"counters,omitempty"`
	// DamageEachOpponent is the damage its source deals to each
	// opponent per card.
	DamageEachOpponent int `json:"damage_each_opponent,omitempty"`
}

// ActivationPurposeView is ADR 0106's name for PurposeView, from when a
// purpose was declared on any-player rows alone. Kept so a reader of
// that ADR finds the type.
type ActivationPurposeView = PurposeView

// SweepView is game.Sweep on the wire.
type SweepView struct {
	// Matches is the class of permanents it removes (game.SweepMatches).
	Matches string `json:"matches"`
	// How is destroy, exile, bounce, damage, minus or sacrifice.
	How string `json:"how"`
	// Amount is the damage or the N of -N/-N; absent otherwise and
	// when AmountIsX.
	Amount int `json:"amount,omitempty"`
	// AmountIsX: the amount is the spell's or ability's X.
	AmountIsX bool `json:"amount_is_x,omitempty"`
	// OpponentsOnly: only permanents its controller's opponents
	// control are removed.
	OpponentsOnly bool `json:"opponents_only,omitempty"`
	// Partial: some permanents of the class are spared by a condition
	// the view does not name, so Matches is an upper bound.
	Partial bool `json:"partial,omitempty"`
}

// viewOfPurpose projects a declared purpose, nil when none is declared.
func viewOfPurpose(p game.Purpose) *PurposeView {
	if p.IsZero() {
		return nil
	}
	v := &PurposeView{
		Draws:                     p.Draws,
		ControllerLosesLife:       p.ControllerLosesLife,
		Discards:                  p.Discards,
		Lands:                     p.Lands,
		LandsUntapped:             p.LandsUntapped,
		Tutors:                    p.Tutors,
		SelfMillTutor:             p.SelfMillTutor,
		Tokens:                    p.Tokens,
		Energy:                    p.Energy,
		DeathPayoff:               p.DeathPayoff,
		ExtraCombat:               p.ExtraCombat,
		PreventCombatDamageToSelf: p.PreventCombatDamageToSelf,
		DamageToCreature:          p.DamageToCreature,
		DamageEachOpponent:        p.DamageEachOpponent,
		LifeGain:                  p.LifeGain,
		AwakenLand:                p.AwakenLand,
	}
	if pm := p.Pump; pm != nil {
		v.Pump = &PumpView{Power: pm.Power, Toughness: pm.Toughness, Keywords: append([]string(nil), pm.Keywords...)}
	}
	if d := p.DiscardPayoff; d != nil {
		v.DiscardPayoff = &DiscardPayoffView{
			Any:                d.Any,
			Types:              append([]string(nil), d.Types...),
			Tokens:             d.Tokens,
			Counters:           d.Counters,
			DamageEachOpponent: d.DamageEachOpponent,
		}
	}
	if ts := p.Targets.List(); len(ts) > 0 {
		out := make([]TargetPurposeView, len(ts))
		for i, t := range ts {
			out[i] = TargetPurposeView{
				Slot: t.Slot, Draws: t.Draws, Discards: t.Discards, Tokens: t.Tokens,
				LifeGain: t.LifeGain, LifeLoss: t.LifeLoss, Damage: t.Damage,
			}
		}
		v.Targets = &out
	}
	if s := p.Sweep; !s.IsZero() {
		v.Sweep = &SweepView{
			Matches:       string(s.Matches),
			How:           string(s.How),
			Amount:        s.Amount,
			AmountIsX:     s.AmountIsX,
			OpponentsOnly: s.OpponentsOnly,
			Partial:       s.Partial,
		}
	}
	return v
}
