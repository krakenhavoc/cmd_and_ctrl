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
// half's when it has one, else the second's.
func (p Purpose) plus(o Purpose) Purpose {
	out := Purpose{
		Draws:               p.Draws + o.Draws,
		ControllerLosesLife: p.ControllerLosesLife + o.ControllerLosesLife,
		Discards:            p.Discards + o.Discards,
		Lands:               p.Lands + o.Lands,
		Tutors:              p.Tutors + o.Tutors,
		SelfMillTutor:       p.SelfMillTutor + o.SelfMillTutor,
		Tokens:              p.Tokens + o.Tokens,
		Energy:              p.Energy + o.Energy,
		Sweep:               p.Sweep,
		DeathPayoff:         p.DeathPayoff || o.DeathPayoff,
		DiscardPayoff:       p.DiscardPayoff,
	}
	if out.Sweep.IsZero() {
		out.Sweep = o.Sweep
	}
	if out.DiscardPayoff == nil {
		out.DiscardPayoff = o.DiscardPayoff
	}
	return out
}
