package game

// purpose.go — ADR 0126 §6: a declared purpose for what a spell or an
// ability does.
//
// The heuristic bot reads the seat's filtered view and nothing else
// (ADR 0033 §3). Wrath of God and Divination are both untargeted
// sorceries with a mana value, and a spell's effect is an OnResolve
// closure, so nothing on the wire told them apart. A Purpose is the
// catalog's answer: printed amounts, declared by hand on the card file
// like Completeness, never derived at run time.
//
// The engine never reads it. It is catalog data projected into the
// view (`purpose` on a card, a mode, an alternative cost, an activated
// row and an ability row), so it is not game state: nothing captures
// it and SnapshotSchemaVersion does not move for it.
//
// One type for every place it is declared. ADR 0106 §1 decision 8
// introduced it as ActivationPurpose, for "Any player may activate
// this ability" rows; ADR 0126 widened it. The fields that make no
// sense in a slot are refused there by effects.Register (see
// effects/purpose.go), so a declaration that nothing would read fails
// at boot rather than being silently ignored.

// Purpose is what a spell, a mode, an alternative cost or an ability
// does, as printed amounts. The zero value is "no declared purpose".
//
// Every amount is the PRINTED number: Night's Whisper draws 2, Harrow
// puts 2 lands onto the battlefield. A card whose amount is X, or is
// counted at resolution ("draw a card for each creature you control"),
// does not declare it, because no number here would be true.
type Purpose struct {
	// Draws is how many cards its controller draws ("you draw a
	// card", CR 109.5: "you" is the player who cast or activated it).
	// On an any-player row that is the activator (ADR 0106).
	Draws int
	// ControllerLosesLife is how much life the SOURCE'S CONTROLLER
	// loses ("Xantcha's controller loses 2 life"). Only on an
	// any-player activated row, where the activator and the
	// controller differ (ADR 0106 §1 decision 8).
	ControllerLosesLife int
	// Discards is how many cards its controller discards on
	// resolution: a loot's 1, Faithless Looting's 2. A discard that is
	// a COST is not here; it is the additional cost's own field.
	Discards int
	// Lands is how many land cards it puts onto the battlefield under
	// its controller's control: Rampant Growth 1, Harrow 2. A land put
	// into the hand (Cultivate's second) is a Tutor, not a Land.
	Lands int
	// Tutors is how many cards it searches out to its controller's
	// hand or to the top of their library: Demonic Tutor 1, Vampiric
	// Tutor 1, Cultivate's land to hand 1.
	Tutors int
	// SelfMillTutor is how many cards it searches out into its
	// controller's graveyard: Entomb 1, Buried Alive 3.
	SelfMillTutor int
	// Tokens is how many tokens it creates for its controller: Big
	// Score's 2 Treasures. Tokens it gives to another player (Generous
	// Gift's Elephant) are not counted.
	Tokens int
	// Energy is how many energy counters it gives its controller ("you
	// get {E}{E}" is 2; ADR 0129 §7). An amount counted at resolution
	// ("you get {E} for each creature you control") is not declared.
	Energy int
	// Sweep marks a board wipe and says which permanents it removes
	// (ADR 0126 §4). The zero Sweep is "not a wipe".
	Sweep Sweep
	// DeathPayoff is set on a TRIGGERED row that pays out whenever a
	// creature its controller controls dies: Blood Artist, Zulaport
	// Cutthroat (ADR 0126 §7). Refused on every other slot.
	DeathPayoff bool
	// DiscardPayoff is set on a TRIGGERED row that pays out whenever
	// its controller discards a card it matches: Mary Read and Anne
	// Bonny's Treasure for an Island, Pirate or Vehicle card, Marauding
	// Mako's +1/+1 counter for any card (ADR 0126's amendment of
	// 2026-10-06). Nil is "not a discard payoff". Refused on every
	// other slot. A pointer so Purpose stays comparable.
	DiscardPayoff *DiscardPayoff

	// The fields below were added for exert (ADR 0130 §9 and its
	// amendment of 2026-10-07). The heuristic reads them on a row the
	// catalog stamps as an exert row (TriggeredAbility.Exert); they are
	// printed amounts like the rest, and true wherever they are declared.

	// Pump is what the row gives its own source until end of turn:
	// "this creature gets +2/+0 and gains trample until end of turn".
	// Nil is "no pump". Refused off a triggered or activated row. A
	// pointer so Purpose stays comparable.
	Pump *Pump
	// ExtraCombat is the additional combat phases it adds ("after this
	// phase, there is an additional combat phase"): Combat Celebrant 1.
	ExtraCombat int
	// PreventCombatDamageToSelf is set when it prevents all combat
	// damage that would be dealt to its source this turn: Oketra's
	// Avenger. Refused off a triggered or activated row.
	PreventCombatDamageToSelf bool
	// DamageToCreature is the damage it deals to one target creature:
	// Glorybringer's 4.
	DamageToCreature int
	// DamageEachOpponent is the damage it deals to each opponent:
	// Resolute Survivors' 1. (DiscardPayoff.DamageEachOpponent is the
	// same amount PER discarded card, on a discard payoff.)
	DamageEachOpponent int
	// LifeGain is the life its controller gains: Resolute Survivors' 1.
	LifeGain int

	// AwakenLand is the N of "Awaken N—[cost]" (CR 702.113a, ADR 0135
	// §3, owner decision 6): the +1/+1 counters the cast puts on a land
	// its controller controls as it makes it a 0/0 Elemental creature
	// with haste. Declared only on an awaken offer, by effects.Awaken;
	// the heuristic prices it as a hasty N/N body that is also a land,
	// beside the spell's own purpose, which awaken leaves alone.
	AwakenLand int

	// ExtraLandDrops is the additional lands its controller may play
	// (#2678, CR 305.2): on a permanent, "you may play an additional
	// land on each of your turns" (Oracle of Mul Daya, Exploration 1,
	// Azusa 2); on an instant or sorcery, "you may play an additional
	// land this turn" (Explore 1). Declared on the card slot only. On a
	// permanent it must agree with the Spec's AdditionalLandPlays, the
	// engine's own field for the static, and effects.Register refuses
	// one that does not.
	ExtraLandDrops int

	// Targets is what happens TO each target the statement names, one
	// entry per target clause (ADR 0126's amendment of 2026-10-08, owner
	// answer 1). Every amount above is its controller's; an entry's are
	// the target's. "Target player draws two cards" is {Slot: 0, Draws:
	// 2} here, never Draws 2 above, because who draws is whoever the
	// move aims it at. The statement is the one the purpose is declared
	// on: the card's own clause list, a mode's, an alternative cost's or
	// an ability row's, and Slot indexes it as TargetRef.Slot does.
	//
	// Nil is "no target entries"; effects.Register refuses an empty
	// list. A pointer so Purpose stays comparable.
	Targets *TargetPurposes
}

// TargetPurposes is a statement's target entries (Purpose.Targets), at
// most one per clause.
type TargetPurposes []TargetPurpose

// TargetPurpose is what a spell or an ability does to the target chosen
// for one of its clauses, as printed amounts. A player amount (draws,
// discards, tokens, life gained or lost) is that player's; Damage is
// dealt to whatever the clause's pick is, player or permanent.
type TargetPurpose struct {
	// Slot is the target clause's index in its statement (CR 601.2c:
	// one clause per instance of the word "target"), the slot a move's
	// target names.
	Slot int
	// Draws is the cards the target player draws: Sign in Blood 2.
	Draws int
	// Discards is the cards the target player discards on resolution:
	// Prismari Command's loot 2.
	Discards int
	// Tokens is the tokens the target player creates: Prismari
	// Command's Treasure 1.
	Tokens int
	// LifeGain is the life the target player gains.
	LifeGain int
	// LifeLoss is the life the target player loses: Sign in Blood 2.
	// Damage is not here; it is Damage.
	LifeLoss int
	// Damage is the damage dealt to the target: Lightning Bolt 3.
	Damage int
}

// IsZero reports whether the entry says nothing about its target.
func (t TargetPurpose) IsZero() bool {
	return !t.HasPlayerAmount() && t.Damage == 0
}

// HasPlayerAmount reports whether the entry names an amount only a
// player can be given: a draw, a discard, a token or a life change.
func (t TargetPurpose) HasPlayerAmount() bool {
	return t.Draws != 0 || t.Discards != 0 || t.Tokens != 0 || t.LifeGain != 0 || t.LifeLoss != 0
}

// ForTargets is the Purpose.Targets of these entries, nil for none.
// The card files' constructor:
//
//	Purpose: game.Purpose{Targets: game.ForTargets(
//		game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2})}
func ForTargets(ts ...TargetPurpose) *TargetPurposes {
	if len(ts) == 0 {
		return nil
	}
	out := TargetPurposes(append([]TargetPurpose(nil), ts...))
	return &out
}

// List is the entries, nil-safe.
func (tp *TargetPurposes) List() []TargetPurpose {
	if tp == nil {
		return nil
	}
	return *tp
}

// Pump is a self pump until end of turn (Purpose.Pump): the power and
// toughness it adds and the keywords it grants, lowercase as the wire's
// `abilities` spells them.
type Pump struct {
	Power     int
	Toughness int
	Keywords  []string
}

// DiscardPayoff says which discarded cards a triggered row pays on, and
// what it pays for EACH one, as printed amounts. The heuristic prices a
// discard of a matching card that much cheaper.
type DiscardPayoff struct {
	// Any is set when every card its controller discards pays ("a
	// card", "one or more cards").
	Any bool
	// Types, when Any is not set, are the card types and subtypes it
	// pays on, lowercase, as printed: Mary Read's "island", "pirate",
	// "vehicle". A card with any one of them on its type line matches.
	Types []string
	// Tokens is the tokens it creates for its controller per card:
	// Mary Read's tapped Treasure.
	Tokens int
	// Counters is the +1/+1 counters it puts on its source per card:
	// Marauding Mako's "that many".
	Counters int
	// DamageEachOpponent is the damage its source deals to each
	// opponent per card: Glint-Horn Buccaneer's 1.
	DamageEachOpponent int
}

// IsZero reports whether nothing is declared.
func (p Purpose) IsZero() bool {
	return p == Purpose{}
}

// Sweep describes a board wipe: which permanents it matches, how it
// removes them, and for damage or -N/-N, how much.
type Sweep struct {
	// Matches is the class of permanents it removes.
	Matches SweepMatch
	// How is how it removes them.
	How SweepHow
	// Amount is the damage dealt or the N of -N/-N, for SweepDamage and
	// SweepMinus. Zero with AmountIsX.
	Amount int
	// AmountIsX is set when the amount is the spell's or ability's X
	// (Earthquake, Toxic Deluge, Crypt Rats). Amount is then zero, and
	// a reader takes the X the move names.
	AmountIsX bool
	// OpponentsOnly is set when only permanents its controller's
	// opponents control are removed: "creatures your opponents
	// control", "you don't control", and "target player controls",
	// which is cast at an opponent.
	OpponentsOnly bool
	// Partial is set when the printed sweep spares some permanents of
	// the Matches class by a condition this type does not name:
	// nonwhite, nontoken, without flying, power 4 or greater, with no
	// counters, of a chosen type. Matches is then an upper bound.
	Partial bool
}

// IsZero reports whether the sweep is unset ("not a wipe").
func (s Sweep) IsZero() bool { return s == Sweep{} }

// SweepMatch is the class of permanents a sweep removes. ADR 0126 §4
// names six, and a value is added only when a curated card needs one.
//
// A card declares the smallest listed class that holds everything it
// removes, leaving out the kinds no class names (planeswalkers,
// battles, lands, which no sweep the bot prices is about), with Partial
// set when that class holds permanents the card spares. So Nevinyrral's
// Disk ("artifacts, creatures, and enchantments") is
// nonland_permanents, partial; In Garruk's Wake ("creatures you don't
// control and planeswalkers you don't control") is creatures; and a
// bullet that destroys only planeswalkers declares no sweep.
type SweepMatch string

const (
	SweepCreatures                SweepMatch = "creatures"
	SweepNonlandPermanents        SweepMatch = "nonland_permanents"
	SweepArtifacts                SweepMatch = "artifacts"
	SweepEnchantments             SweepMatch = "enchantments"
	SweepArtifactsAndEnchantments SweepMatch = "artifacts_and_enchantments"
	SweepAllPermanents            SweepMatch = "all_permanents"
	// Austere Command's last two modes (and Ritual of Soot): a
	// creature sweep split by mana value. Added for the curated
	// esper-control deck.
	SweepCreaturesManaValue3OrLess SweepMatch = "creatures_mana_value_3_or_less"
	SweepCreaturesManaValue4OrMore SweepMatch = "creatures_mana_value_4_or_greater"
)

// SweepMatches lists every SweepMatch, in declaration order. The
// registration guard and the wire documentation read it.
var SweepMatches = []SweepMatch{
	SweepCreatures, SweepNonlandPermanents, SweepArtifacts, SweepEnchantments,
	SweepArtifactsAndEnchantments, SweepAllPermanents,
	SweepCreaturesManaValue3OrLess, SweepCreaturesManaValue4OrMore,
}

// SweepHow is how a sweep removes what it matches.
type SweepHow string

const (
	SweepDestroy SweepHow = "destroy"
	SweepExile   SweepHow = "exile"
	// SweepBounce returns the permanents to their owners' hands.
	SweepBounce SweepHow = "bounce"
	// SweepDamage deals Amount damage to each matched creature (or
	// planeswalker); one whose toughness is above it survives.
	SweepDamage SweepHow = "damage"
	// SweepMinus gives each matched creature -Amount/-Amount.
	SweepMinus SweepHow = "minus"
	// SweepSacrifice makes each affected player sacrifice what it
	// matches. Indestructible does not save a sacrificed permanent.
	// Added for the curated mono-black deck's Living Death.
	SweepSacrifice SweepHow = "sacrifice"
)

// SweepHows lists every SweepHow, in declaration order.
var SweepHows = []SweepHow{SweepDestroy, SweepExile, SweepBounce, SweepDamage, SweepMinus, SweepSacrifice}

// CardPurposeOf is the purpose the catalog declares for the card
// itself: what the spell does on resolution, or a permanent's
// enters-the-battlefield effect. Zero for an uncatalogued card, a card
// that declares none, and a face-down object (CR 708.2a: it has no
// text, and catalogKeyOf already answers the empty key for it).
//
// The card's OWN key, not CatalogAbilityKey's merged one: a layer-6
// grant gives a permanent abilities, not a different resolution.
func CardPurposeOf(c Card) Purpose {
	d := catalogDef(catalogKeyOf(&c))
	if d == nil {
		return Purpose{}
	}
	return d.Purpose
}

// plus is what two effects declare together: a fused split spell
// does both halves (fuse). Amounts add; a sweep is the first
// half's when it has one, else the second's. The target entries are
// the first half's followed by the second's, whose slots move past the
// first half's `leftClauses` clauses, as the fused statement numbers
// them (split_fuse.go).
func (p Purpose) plus(o Purpose, leftClauses int) Purpose {
	out := Purpose{
		Draws:                     p.Draws + o.Draws,
		ControllerLosesLife:       p.ControllerLosesLife + o.ControllerLosesLife,
		Discards:                  p.Discards + o.Discards,
		Lands:                     p.Lands + o.Lands,
		Tutors:                    p.Tutors + o.Tutors,
		SelfMillTutor:             p.SelfMillTutor + o.SelfMillTutor,
		Tokens:                    p.Tokens + o.Tokens,
		Energy:                    p.Energy + o.Energy,
		Sweep:                     p.Sweep,
		DeathPayoff:               p.DeathPayoff || o.DeathPayoff,
		DiscardPayoff:             p.DiscardPayoff,
		Pump:                      p.Pump,
		ExtraCombat:               p.ExtraCombat + o.ExtraCombat,
		PreventCombatDamageToSelf: p.PreventCombatDamageToSelf || o.PreventCombatDamageToSelf,
		DamageToCreature:          p.DamageToCreature + o.DamageToCreature,
		DamageEachOpponent:        p.DamageEachOpponent + o.DamageEachOpponent,
		LifeGain:                  p.LifeGain + o.LifeGain,
		AwakenLand:                p.AwakenLand + o.AwakenLand,
		ExtraLandDrops:            p.ExtraLandDrops + o.ExtraLandDrops,
	}
	if out.Pump == nil {
		out.Pump = o.Pump
	}
	if out.Sweep.IsZero() {
		out.Sweep = o.Sweep
	}
	if out.DiscardPayoff == nil {
		out.DiscardPayoff = o.DiscardPayoff
	}
	targets := append([]TargetPurpose(nil), p.Targets.List()...)
	for _, t := range o.Targets.List() {
		t.Slot += leftClauses
		targets = append(targets, t)
	}
	out.Targets = ForTargets(targets...)
	return out
}
