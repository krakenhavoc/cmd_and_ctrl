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
	// EntryCounters describes fixed or X own-entry clauses. Other
	// replacements and conditional/board-dependent counts are not predicted.
	EntryCounters *[]EntryCounterView `json:"entry_counters,omitempty"`

	// LandEntry describes the land's own unconditional entry, not other
	// permanents' replacements. Empty for unknown/conditional entry.
	LandEntry          string `json:"land_entry,omitempty"`
	ReturnsLandOnEntry bool   `json:"returns_land_on_entry,omitempty"`

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
	// CreatureTokens is how many of Tokens are creature tokens (#3165).
	CreatureTokens int `json:"creature_tokens,omitempty"`
	// TokensPerControlled is the lowercase creature subtype whose count
	// under its controller's control is the number of creature tokens it
	// creates, counted at resolution (#3165).
	TokensPerControlled string `json:"tokens_per_controlled,omitempty"`
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
	// OnOpponentDiscard is present on a triggered row of a card that
	// pays out when an opponent's spell or ability makes its controller
	// discard that card, in discard_payoff's units (ADR 0153 §5).
	OnOpponentDiscard *DiscardPayoffView `json:"on_opponent_discard,omitempty"`
	// LandfallPayoff is present on a triggered row that pays out
	// whenever a land enters under its controller's control, and says
	// what it pays for each land (#3072).
	LandfallPayoff *LandfallPayoffView `json:"landfall_payoff,omitempty"`
	// CreatureEnterPayoff is present on a triggered row that pays out
	// whenever a creature enters under its controller's control, and
	// says what it pays for each creature (#3165).
	CreatureEnterPayoff *CreatureEnterPayoffView `json:"creature_enter_payoff,omitempty"`
	// Pump is what the row gives its own source until end of turn
	// (ADR 0130's amendment of 2026-10-07).
	Pump *PumpView `json:"pump,omitempty"`
	// ExtraCombat is the additional combat phases it adds.
	ExtraCombat int `json:"extra_combat,omitempty"`
	// ExtraTurns is the extra turns its controller takes after this one
	// (CR 500.7, ADR 0188 Decision 10).
	ExtraTurns int `json:"extra_turns,omitempty"`
	// SkipsTurns is the next turns its controller skips (CR 614.10).
	SkipsTurns int `json:"skips_turns,omitempty"`
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
	// ExtraLandDrops is the additional lands its controller may play:
	// each turn on a permanent ("you may play an additional land on
	// each of your turns"), this turn on an instant or sorcery (#2678).
	ExtraLandDrops int `json:"extra_land_drops,omitempty"`
	// LandfallExtraTriggers is how many additional times a landfall
	// trigger of a permanent its controller controls triggers, on the
	// card of a permanent that doubles them (#3146): Ancient
	// Greenwarden 1.
	LandfallExtraTriggers int `json:"landfall_extra_triggers,omitempty"`
	// ExtraLoyaltyActivations is how many more times this turn each
	// planeswalker its controller controls may activate a loyalty
	// ability (#4191): The Chain Veil's activated row 1.
	ExtraLoyaltyActivations int `json:"extra_loyalty_activations,omitempty"`
	// Impulse is the cards it exiles from the top of its controller's
	// library that its controller may play this turn (#2480).
	Impulse int `json:"impulse,omitempty"`
	// Targets is what happens to each target the statement names, one
	// entry per target clause, keyed by the clause's slot: the slot a
	// move's `targets[].slot` names, within the statement this purpose
	// rides on (ADR 0126's amendment of 2026-10-08). The amounts above
	// are the controller's; an entry's are its target's. Absent when
	// empty. A pointer so PurposeView stays comparable.
	Targets *[]TargetPurposeView `json:"targets,omitempty"`
	// Answers is what the row can do in response (ADR 0142): the wire
	// names of game.Answers, in the vocabulary's order ("protect", "survive",
	// "pump", "prevent", "remove", "sac_outlet", "restrict",
	// "combat_grant", "animate", "makes_blocker", "value"). Absent when
	// the row declares none; ["value"] is a declared "answers nothing".
	// On an activated row, a triggered row (only "remove"), and since
	// #2877 a card, a spell's mode and an alternative cost, which a
	// cast's value_only reads. It is not an amount: a purpose that
	// declares only answers is not Priced. A pointer so PurposeView
	// stays comparable.
	Answers *[]string `json:"answers,omitempty"`
	// Saves is which of its controller's other permanents a `survive` or
	// `protect` row covers (#2939 gap 3): Selfless Spirit's "creatures you
	// control gain indestructible" is {matches: "creatures"}. Absent for a
	// row that covers only its own source. Not an amount, like Answers. A
	// pointer so PurposeView stays comparable.
	Saves *SavesView `json:"saves,omitempty"`
	// Aura is "benefit" or "harm" on an Aura card (#3309): what it does
	// to the permanent it enchants. Absent when the card declares none.
	// Not an amount, like Answers.
	Aura string `json:"aura,omitempty"`
	// FlashCleanupSacrifice is true on a card that may be cast as though
	// it had flash and is sacrificed at the next cleanup step when it was
	// (Necromancy, Lightning Reflexes, #3712). Not an amount, like Aura.
	FlashCleanupSacrifice bool `json:"flash_cleanup_sacrifice,omitempty"`
	// Kicked is what the card does only when its kicker is paid, on top
	// of the fields above (ADR 0126's amendment of 2026-10-15, #3762).
	// A reader prices it on a cast that pays the kicker (the move's
	// optional_costs is non-empty), and on no other. Not an amount of
	// this purpose, so Priced does not count it.
	Kicked *PurposeView `json:"kicked,omitempty"`
	// TurnedFaceUp is what turning the permanent face up does (CR
	// 708.8): the morph trigger's effect. On a face-down permanent its
	// controller may look at, it is the only field sent (a face-down
	// object has no text of its own, CR 708.2a). Read by the price of
	// the turn_face_up special action. Not an amount of this purpose, so
	// Priced does not count it.
	TurnedFaceUp *PurposeView `json:"turned_face_up,omitempty"`
}

// Priced reports whether the purpose declares anything the bot prices:
// an amount, a sweep, a payoff, a pump or a target entry. Answers alone
// is not one (ADR 0142 §7). A reader that drops its proxy price for a
// declared purpose asks this, not "is purpose non-nil", so a row that
// declares only what it answers keeps the price it had. Nil-safe.
func (p *PurposeView) Priced() bool {
	if p == nil {
		return false
	}
	q := *p
	q.Answers = nil
	q.Saves = nil
	q.Aura = ""
	q.FlashCleanupSacrifice = false
	// What it does only when kicked or turned face up is read by the
	// cast that pays the kicker and the special action that turns it up
	// (#3762). On the unkicked cast it is nothing it does.
	q.Kicked, q.TurnedFaceUp = nil, nil
	return q != PurposeView{}
}

// Declares reports whether the purpose declares answer a, a wire name
// from Answers ("protect", "pump", …). Nil-safe.
func (p *PurposeView) Declares(a string) bool {
	if p == nil || p.Answers == nil {
		return false
	}
	for _, x := range *p.Answers {
		if x == a {
			return true
		}
	}
	return false
}

// AnswersNothing reports whether the purpose is declared "value" (ADR
// 0142 decision 1): the row answers nothing on the stack or in combat.
// False for a row that declares no answers at all. Nil-safe.
func (p *PurposeView) AnswersNothing() bool {
	return p != nil && p.Answers != nil && len(*p.Answers) == 1 && (*p.Answers)[0] == "value"
}

// TargetPurposeView is game.TargetPurpose on the wire: what the spell
// or ability does to the target picked for clause `slot`. Every amount
// is omitted when zero; `slot` is always sent.
type TargetPurposeView struct {
	Slot     int `json:"slot"`
	Draws    int `json:"draws,omitempty"`
	Discards int `json:"discards,omitempty"`
	// ExtraTurns: the target player takes this many extra turns (Time
	// Warp 1, Time Stretch 2; ADR 0188 Decision 10).
	ExtraTurns int `json:"extra_turns,omitempty"`
	Tokens     int `json:"tokens,omitempty"`
	LifeGain   int `json:"life_gain,omitempty"`
	LifeLoss   int `json:"life_loss,omitempty"`
	Damage     int `json:"damage,omitempty"`
	// DamageIsX: the damage is the spell's or ability's X (#1944);
	// Damage is then absent.
	DamageIsX bool `json:"damage_is_x,omitempty"`
	// DamagePerTapped: the damage is this many times the number of
	// permanents the cast's "tap any number" cost taps (#3961); Damage is
	// then absent.
	DamagePerTapped int `json:"damage_per_tapped,omitempty"`
	// Returns is what the target's controller is given when the target
	// is removed (#2679). Absent when nothing comes back.
	Returns *TargetReturnView `json:"returns,omitempty"`
	// Reanimates: the spell puts the picked graveyard card onto the
	// battlefield under its caster's control (#3161).
	Reanimates bool `json:"reanimates,omitempty"`
	// UntilEndOfTurn is "benefit" or "harm": which way a change that ends
	// with the turn cuts for the picked permanent's controller (#4302).
	// Absent when the entry declares none.
	UntilEndOfTurn string `json:"until_end_of_turn,omitempty"`
	// Pump is that change in printed amounts: the power and toughness it
	// adds (negative for a harm) and the keywords a benefit grants.
	// Absent for an amount counted at resolution or a change that is not
	// a pump. A pointer so TargetPurposeView stays comparable.
	Pump *PumpView `json:"pump,omitempty"`
}

// TargetReturnView is game.TargetReturn on the wire: what a removal
// spell hands the controller of the permanent it removes (#2679). Every
// field is omitted when zero.
type TargetReturnView struct {
	CreatureTokens   int  `json:"creature_tokens,omitempty"`
	TokenPower       int  `json:"token_power,omitempty"`
	TokenToughness   int  `json:"token_toughness,omitempty"`
	LifeEqualToPower bool `json:"life_equal_to_power,omitempty"`
	Lands            int  `json:"lands,omitempty"`
	LandsUntapped    int  `json:"lands_untapped,omitempty"`
}

// SavesView is game.Saves on the wire: the class of its controller's
// permanents a survive or protect row covers, named as a sweep names its
// class (SweepView.Matches).
type SavesView struct {
	Matches string `json:"matches"`
	Partial bool   `json:"partial,omitempty"`
}

// PumpView is game.Pump on the wire: a self pump until end of turn.
type PumpView struct {
	Power     int      `json:"power,omitempty"`
	Toughness int      `json:"toughness,omitempty"`
	Keywords  []string `json:"keywords,omitempty"`
	// CantBeBlocked: the change makes the creature unblockable this turn
	// (#4312).
	CantBeBlocked bool `json:"cant_be_blocked,omitempty"`
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

// LandfallPayoffView is game.LandfallPayoff on the wire: what a landfall
// row pays its controller for each land that enters under their control
// (#3072). Every field is omitted when zero.
type LandfallPayoffView struct {
	// Draws is the cards its controller draws per land.
	Draws int `json:"draws,omitempty"`
	// LifeGain is the life its controller gains per land.
	LifeGain int `json:"life_gain,omitempty"`
	// Tokens is the tokens it creates for its controller per land.
	Tokens int `json:"tokens,omitempty"`
	// Mana is the mana it adds per land.
	Mana int `json:"mana,omitempty"`
	// Counters is the +1/+1 counters it puts on its source per land.
	Counters int `json:"counters,omitempty"`
	// Energy is the energy counters its controller gets per land.
	Energy int `json:"energy,omitempty"`
	// DamageEachOpponent is the damage its source deals to each
	// opponent per land.
	DamageEachOpponent int `json:"damage_each_opponent,omitempty"`
}

// CreatureEnterPayoffView is game.CreatureEnterPayoff on the wire: what a
// row pays its controller for each creature that enters under their
// control (#3165), in LandfallPayoffView's fields.
type CreatureEnterPayoffView LandfallPayoffView

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

// EntryCounterView is one reliable own-entry counter amount.
type EntryCounterView struct {
	Kind      string `json:"kind"`
	Amount    int    `json:"amount,omitempty"`
	AmountIsX bool   `json:"amount_is_x,omitempty"`
}

// viewOfPurpose projects a declared purpose, nil when none is declared.
func viewOfPurpose(p game.Purpose) *PurposeView {
	if p.IsZero() {
		return nil
	}
	v := &PurposeView{
		LandEntry:                 p.LandEntry,
		ReturnsLandOnEntry:        p.ReturnsLandOnEntry,
		Kicked:                    viewOfPurpose(derefPurpose(p.Kicked)),
		TurnedFaceUp:              viewOfPurpose(derefPurpose(p.FaceUp)),
		Draws:                     p.Draws,
		ControllerLosesLife:       p.ControllerLosesLife,
		Discards:                  p.Discards,
		Lands:                     p.Lands,
		LandsUntapped:             p.LandsUntapped,
		Tutors:                    p.Tutors,
		SelfMillTutor:             p.SelfMillTutor,
		Tokens:                    p.Tokens,
		CreatureTokens:            p.CreatureTokens,
		TokensPerControlled:       p.TokensPerControlled,
		Energy:                    p.Energy,
		DeathPayoff:               p.DeathPayoff,
		ExtraCombat:               p.ExtraCombat,
		ExtraTurns:                p.ExtraTurns,
		SkipsTurns:                p.SkipsTurns,
		PreventCombatDamageToSelf: p.PreventCombatDamageToSelf,
		DamageToCreature:          p.DamageToCreature,
		DamageEachOpponent:        p.DamageEachOpponent,
		LifeGain:                  p.LifeGain,
		AwakenLand:                p.AwakenLand,
		ExtraLandDrops:            p.ExtraLandDrops,
		LandfallExtraTriggers:     p.LandfallExtraTriggers,
		ExtraLoyaltyActivations:   p.ExtraLoyaltyActivations,
		Impulse:                   p.Impulse,
		Aura:                      string(p.Aura),
		FlashCleanupSacrifice:     p.FlashCleanupSacrifice,
	}
	if p.EntryCounters != nil {
		entries := make([]EntryCounterView, 0, len(*p.EntryCounters))
		for _, e := range *p.EntryCounters {
			entries = append(entries, EntryCounterView{Kind: e.Kind, Amount: e.Amount, AmountIsX: e.AmountIsX})
		}
		v.EntryCounters = &entries
	}
	if pm := p.Pump; pm != nil {
		v.Pump = &PumpView{Power: pm.Power, Toughness: pm.Toughness, Keywords: append([]string(nil), pm.Keywords...), CantBeBlocked: pm.CantBeBlocked}
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
	if d := p.OnOpponentDiscard; d != nil {
		v.OnOpponentDiscard = &DiscardPayoffView{
			Any:                d.Any,
			Types:              append([]string(nil), d.Types...),
			Tokens:             d.Tokens,
			Counters:           d.Counters,
			DamageEachOpponent: d.DamageEachOpponent,
		}
	}
	if l := p.LandfallPayoff; l != nil {
		v.LandfallPayoff = &LandfallPayoffView{
			Draws: l.Draws, LifeGain: l.LifeGain, Tokens: l.Tokens, Mana: l.Mana,
			Counters: l.Counters, Energy: l.Energy, DamageEachOpponent: l.DamageEachOpponent,
		}
	}
	if c := p.CreatureEnterPayoff; c != nil {
		pay := CreatureEnterPayoffView(*c)
		v.CreatureEnterPayoff = &pay
	}
	if ts := p.Targets.List(); len(ts) > 0 {
		out := make([]TargetPurposeView, len(ts))
		for i, t := range ts {
			out[i] = TargetPurposeView{
				Slot: t.Slot, Draws: t.Draws, Discards: t.Discards, ExtraTurns: t.ExtraTurns, Tokens: t.Tokens,
				LifeGain: t.LifeGain, LifeLoss: t.LifeLoss, Damage: t.Damage, DamageIsX: t.DamageIsX, DamagePerTapped: t.DamagePerTapped, Reanimates: t.Reanimates,
				UntilEndOfTurn: string(t.UntilEndOfTurn),
			}
			if pm := t.Pump; pm != nil {
				out[i].Pump = &PumpView{Power: pm.Power, Toughness: pm.Toughness, Keywords: append([]string(nil), pm.Keywords...), CantBeBlocked: pm.CantBeBlocked}
			}
			if r := t.Returns; !r.IsZero() {
				out[i].Returns = &TargetReturnView{
					CreatureTokens: r.CreatureTokens, TokenPower: r.TokenPower, TokenToughness: r.TokenToughness,
					LifeEqualToPower: r.LifeEqualToPower, Lands: r.Lands, LandsUntapped: r.LandsUntapped,
				}
			}
		}
		v.Targets = &out
	}
	if w := p.Answers.Wire(); len(w) > 0 {
		v.Answers = &w
	}
	if s := p.Saves; s != nil {
		v.Saves = &SavesView{Matches: string(s.Matches), Partial: s.Partial}
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

// derefPurpose is *p, the zero Purpose for nil.
func derefPurpose(p *game.Purpose) game.Purpose {
	if p == nil {
		return game.Purpose{}
	}
	return *p
}
