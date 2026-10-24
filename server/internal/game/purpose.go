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
	// LandsUntapped is how many of those Lands enter UNTAPPED, so their
	// mana can be spent the turn the spell resolves (ADR 0136 §2, owner
	// answer 4): Harrow 2, Nature's Lore 1, Three Visits 1; Rampant
	// Growth's "onto the battlefield tapped" is 0. Never more than
	// Lands; effects.Register refuses one that is. A land's OWN
	// enters-tapped replacement is not counted against it: the printed
	// spell puts the land onto the battlefield untapped, and a tapland
	// it fetches still enters tapped, which the reader cannot know
	// before the search.
	LandsUntapped int
	// Tutors is how many cards it searches out to its controller's
	// hand or to the top of their library: Demonic Tutor 1, Vampiric
	// Tutor 1, Cultivate's land to hand 1.
	Tutors int
	// SelfMillTutor is how many cards it searches out into its
	// controller's graveyard: Entomb 1, Buried Alive 3.
	SelfMillTutor int
	// Tokens is how many tokens it creates for its controller: Big
	// Score's 2 Treasures. Tokens it gives to another player (Generous
	// Gift's Elephant) are not counted here; a removal's gift to its
	// target's controller is that entry's TargetPurpose.Returns.
	Tokens int
	// CreatureTokens is how many of Tokens are creature tokens: Dragon
	// Fodder's 2 Goblins, not Big Score's 2 Treasures. A creature
	// entering is what a CreatureEnterPayoff pays on, so the heuristic
	// prices only these by it (#3165). Never more than Tokens.
	CreatureTokens int
	// TokensPerControlled makes the creature tokens an amount counted at
	// resolution: the lowercase creature subtype whose count under its
	// controller's control is how many it creates ("goblin" on Krenko,
	// Mob Boss: X Goblins, X the Goblins you control). Tokens stays 0,
	// since no printed number would be true; the reader counts from the
	// view (#3165). Refused off a triggered or activated row.
	TokensPerControlled string
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
	// OnOpponentDiscard is set on a TRIGGERED row of a card that pays
	// out when an opponent's spell or ability makes its controller
	// discard THAT card ("if a spell or ability an opponent controls
	// causes you to discard this card, create four Gnome tokens"), in
	// DiscardPayoff's units and per printed amount (ADR 0153 §5). The
	// heuristic prices that discard of the card in hand that much
	// cheaper, and only when an opponent caused it. Any is set, because
	// the card is the discarded card. A row that sits on the battlefield
	// never pays out for other discards by it, which is why this is not
	// DiscardPayoff. Nil is "declares nothing". Refused on every other
	// slot. A pointer so Purpose stays comparable.
	OnOpponentDiscard *DiscardPayoff
	// LandfallPayoff is set on a TRIGGERED row that pays out whenever a
	// land enters under its controller's control ("Landfall — Whenever a
	// land you control enters"), and says what it pays for EACH land:
	// Tatyova's card and life, Lotus Cobra's mana, Tireless Tracker's
	// Clue (#3072). Nil is "not a landfall payoff". Refused on every
	// other slot. A pointer so Purpose stays comparable.
	LandfallPayoff *LandfallPayoff
	// CreatureEnterPayoff is set on a TRIGGERED row that pays out
	// whenever a creature enters under its controller's control
	// (Purphoros, God of the Forge's 2 damage; Impact Tremors' 1), and
	// says what it pays for EACH creature (#3165). Nil is "not a
	// creature-enter payoff". Refused on every other slot. A pointer so
	// Purpose stays comparable.
	CreatureEnterPayoff *CreatureEnterPayoff

	// The fields below were added for exert (ADR 0130 §9 and its
	// amendment of 2026-10-07). The heuristic reads them on a row the
	// catalog stamps as an exert row (TriggeredAbility.Exert); they are
	// printed amounts like the rest, and true wherever they are declared.

	// Pump is what the row gives its own source until end of turn:
	// "this creature gets +2/+0 and gains trample until end of turn".
	// Nil is "no pump". Refused off a triggered or activated row, except
	// as the inner purpose of FaceUp (a turn-face-up trigger's pump:
	// Unstable Hulk, #3776). A pointer so Purpose stays comparable.
	Pump *Pump
	// ExtraCombat is the additional combat phases it adds ("after this
	// phase, there is an additional combat phase"): Combat Celebrant 1.
	ExtraCombat int
	// ExtraTurns is the extra turns its controller takes after this one
	// (CR 500.7): Capture of Jingzhou 1, Magosi's fourth ability 1. The
	// CASTER's own turns, as Draws is the caster's own cards; "target
	// player takes an extra turn" (Time Warp) is a TargetPurpose entry,
	// because who takes it is whoever the move aims it at. Left off a
	// card that makes its controller lose the game in that turn (Final
	// Fortune, Last Chance, Warrior's Oath): the turn is the price, and a
	// bot that priced it as a gain would cast them. Refused on an
	// any-player activated row, where the activator is not the
	// controller. ADR 0188 Decision 10, #3566.
	ExtraTurns int
	// SkipsTurns is the next turns its controller skips (CR 614.10):
	// Magosi's third ability 1, Eater of Days 2. The caster's own turns,
	// like ExtraTurns. Priced negatively (Weights.Turn). Refused on an
	// any-player activated row: Lethal Vapors' skip falls on whoever
	// activates it, a bot never activates another player's copy, and a
	// declared cost would only hide that rule (ADR 0106 decision 2,
	// ADR 0188 Decision 10, #3566).
	SkipsTurns int
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

	// LandfallExtraTriggers is how many additional times a triggered
	// ability of a permanent its controller controls triggers when a
	// land entering causes it to trigger (#3146): Ancient Greenwarden
	// 1, Elesh Norn, Mother of Machines 1 (its doubler covers every
	// permanent, lands among them). Declared on the card slot only, of
	// a card whose Spec has a TriggerDoubler, which is the engine's own
	// field for the static; the heuristic multiplies the bot's landfall
	// payoffs by it.
	LandfallExtraTriggers int

	// ExtraLoyaltyActivations is how many more times this turn each
	// planeswalker its controller controls may activate a loyalty
	// ability (#4191): The Chain Veil's "{4}, {T}: For each planeswalker
	// you control, you may activate one of its loyalty abilities once
	// this turn as though none of its loyalty abilities have been
	// activated this turn" 1. The engine's own record is
	// Game.LoyaltyExtraActivationsThisTurn; the heuristic activates such
	// a row only in its own main phase, with planeswalkers to use it.
	ExtraLoyaltyActivations int

	// Impulse is how many cards it exiles from the top of its
	// controller's library that its controller may play this turn
	// (#2480): Professional Face-Breaker's "Exile the top card of your
	// library. You may play that card this turn." 1. Not a draw (CR
	// 121.1): the card is lost unless it is played before the turn
	// ends, so a reader prices it by whether it can be.
	Impulse int

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

	// Answers is what this ability can do in response: the reasons a
	// player would hold priority for it (ADR 0142). Declared by hand,
	// like every Purpose field. The zero value is "not declared", which
	// the reader (internal/legal's untargetedFlags) counts as
	// interacting (ADR 0142 owner answer 3), and every catalog row it
	// reads declares. AnswerValue declares "answers nothing".
	// Declared on an activated row only, for now: effects.Register
	// refuses it on a spell, a mode, an alternative cost and a
	// triggered row, which nothing reads it on yet (ADR 0142 §2).
	Answers Answers

	// Saves says which of its controller's permanents a `survive` or
	// `protect` row covers, when it covers more than its own source
	// (#2939 gap 3): Selfless Spirit's "creatures you control gain
	// indestructible". Nil is "its own source only", which every other
	// survive and protect row is. A pointer so Purpose stays comparable.
	// Refused off an activated row, and on one that declares neither
	// answer.
	Saves *Saves

	// Aura says which way an Aura cuts for the permanent it enchants
	// (#3309): a buff helps its host (Rancor), a curse hurts it
	// (Pacifism). Nothing on the wire tells the two apart, so the bot
	// priced every Aura's host as a target to hit. Declared on the card
	// of an Aura only; the zero value is "not declared" and keeps that
	// price. effects.Register refuses it off the card.
	Aura AuraPurpose

	// FlashCleanupSacrifice is set on the wire for a card that prints
	// the Visions-era "you may cast this as though it had flash; cast at
	// a time a sorcery couldn't be, it is sacrificed at the next cleanup
	// step" rider (#3712). It is read off Spec.FlashWithCleanupSacrifice
	// by CardPurposeOfPtr, never declared on a Purpose by hand: the bot
	// holds an instant-speed cast of such a card unless it matters before
	// cleanup.
	FlashCleanupSacrifice bool

	// Kicked is what the card does ONLY when its kicker is paid (ADR
	// 0126's amendment of 2026-10-15, #3762): Waterspout Elemental's
	// "if it was kicked, return all other creatures to their owners'
	// hands and you skip your next turn". The fields above are what the
	// card does however it is cast; these are added to them on a cast
	// that pays the kicker, and only then, so an unkicked Elemental is
	// priced as a Merfolk body and not as a mass bounce. Declared on the
	// card slot only, of a card with exactly one optional cost, which is
	// a kicker; effects.Register refuses it elsewhere. The inner purpose
	// may not itself declare Kicked or FaceUp, nor Targets (its clauses
	// would index a statement the cast announces without them). A
	// pointer so Purpose stays comparable.
	Kicked *Purpose
	// FaceUp is what turning the permanent face up does (CR 708.8): the
	// triggered ability of a morph, megamorph or disguise creature
	// ("When this creature is turned face up, ... You skip your next
	// turn" is Unstable Hulk). A face-down permanent has no text (CR
	// 708.2a), so nothing else on the wire says what is under it; the
	// view projects this, and only this, for its controller (see
	// CardPurposeOfPtr), and the heuristic prices the turn-face-up
	// special action by it. Declared on the card slot only, of a card
	// with a face-down cast offer. Same restrictions as Kicked. A pointer
	// so Purpose stays comparable.
	FaceUp *Purpose
}

// DeclaresSweep reports whether the purpose declares a board wipe in
// any condition it is read under: its own, or the one it adds when
// kicked or turned face up. The registry audit asks this, because a
// card that wipes the board only when kicked still wipes it (ADR 0126's
// amendment of 2026-10-15).
func (p Purpose) DeclaresSweep() bool {
	if !p.Sweep.IsZero() {
		return true
	}
	return (p.Kicked != nil && p.Kicked.DeclaresSweep()) || (p.FaceUp != nil && p.FaceUp.DeclaresSweep())
}

// AuraPurpose is Purpose.Aura: what an Aura does to what it enchants.
type AuraPurpose string

const (
	// AuraBenefit helps its host: a pump, a keyword, protection. A bot
	// enchants its own permanent with it.
	AuraBenefit AuraPurpose = "benefit"
	// AuraHarm hurts its host: Pacifism, Arrest. A bot enchants an
	// opponent's permanent with it.
	AuraHarm AuraPurpose = "harm"
)

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
	// ExtraTurns is the extra turns the target player takes after this
	// one (CR 500.7): Time Warp 1, Time Stretch 2. Whoever the move
	// aims it at takes them, so a bot aims it at itself (ADR 0188
	// Decision 10, #3566).
	ExtraTurns int
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
	// DamageIsX is set when the damage dealt to the target is the
	// spell's or ability's X (#1944: Chandra, Awakened Inferno's −X).
	// Damage is then zero, and a reader takes the X the move names, as
	// Sweep.AmountIsX does.
	DamageIsX bool
	// DamagePerTapped is the damage dealt to the target for each
	// permanent the spell's "tap any number of untapped creatures" cost
	// tapped (#3961): Burn at the Stake's "three times the number of
	// creatures tapped this way" is 3. Damage is then zero, and a reader
	// multiplies by the count the move names on teamwork_ids, as it
	// reads DamageIsX's X.
	DamagePerTapped int
	// Returns is what the target's CONTROLLER is given when the spell
	// or ability removes the target (#2679): Rapid Hybridization's 3/3
	// Frog Lizard, Swords to Plowshares' life, Path to Exile's basic
	// land. Declared on a clause that targets a permanent; the zero
	// value is "nothing comes back".
	Returns TargetReturn
	// Reanimates is set when the spell puts the picked graveyard card
	// onto the battlefield under the caster's control, whoever owns the
	// graveyard (#3161): Reanimate, Animate Dead, Zombify. The bot prices
	// the pick as the permanent it becomes, not as a card it would cast.
	Reanimates bool
	// UntilEndOfTurn says which way a change the clause makes to the
	// target PERMANENT cuts for that permanent's controller, when the
	// whole change ends with the turn (#4302). "Another target creature
	// gains lifelink until end of turn" (Heliod, Sun-Crowned) and "target
	// creature gets +2/+2 until end of turn" are TargetBenefit; "target
	// creature gets -2/-2 until end of turn" is TargetHarm. Nothing else
	// on the wire tells the two apart, so the bot priced every opposing
	// pick as removal and gave an opponent's creature lifelink. The zero
	// value is "not declared", which keeps that price.
	UntilEndOfTurn TargetEffect
	// Pump is what the change is, when it is printed in amounts: the
	// power and toughness it adds (negative for a harm's -N/-N) and the
	// keywords a benefit grants, lowercase as the wire's `abilities`
	// spells them. Nil for an amount counted at resolution (+X/+X) or a
	// change that is not a pump ("can't be blocked this turn"). Declared
	// only beside UntilEndOfTurn. A pointer so TargetPurpose stays
	// comparable.
	Pump *Pump
}

// TargetEffect is TargetPurpose.UntilEndOfTurn: which way a change that
// ends with the turn cuts for the target's controller (#4302).
type TargetEffect string

const (
	// TargetBenefit helps the target: a pump, a keyword, protection,
	// "can't be blocked". A bot aims it at its own permanent.
	TargetBenefit TargetEffect = "benefit"
	// TargetHarm hurts the target: -N/-N, "can't block", "loses all
	// abilities". A bot aims it at an opponent's permanent.
	TargetHarm TargetEffect = "harm"
)

// TargetReturn is what a removal spell hands the controller of the
// permanent it removes, as printed amounts (#2679). A removal that
// gives back a creature token, life or a land is worth less than the
// same removal without it, and the bot cannot read which is which
// without it: the gift is a closure in the card file.
type TargetReturn struct {
	// CreatureTokens is the creature tokens the target's controller
	// creates: Rapid Hybridization, Pongify, Beast Within and Generous
	// Gift 1, Stroke of Midnight 1.
	CreatureTokens int
	// TokenPower and TokenToughness are each such token's printed
	// power and toughness: 3/3 for the four above, 1/1 for Stroke of
	// Midnight.
	TokenPower     int
	TokenToughness int
	// LifeEqualToPower is set when the target's controller gains life
	// equal to the target's power: Swords to Plowshares. The amount is
	// counted at resolution, so it is a flag rather than a number; the
	// reader takes the power the target shows.
	LifeEqualToPower bool
	// Lands is the land cards the target's controller may put onto the
	// battlefield: Path to Exile and Assassin's Trophy 1.
	Lands int
	// LandsUntapped is how many of those Lands enter untapped:
	// Assassin's Trophy 1, Path to Exile 0. Never more than Lands.
	LandsUntapped int
}

// IsZero reports whether nothing comes back.
func (r TargetReturn) IsZero() bool { return r == TargetReturn{} }

// IsZero reports whether the entry says nothing about its target.
func (t TargetPurpose) IsZero() bool {
	return !t.HasPlayerAmount() && t.Damage == 0 && !t.DamageIsX && t.DamagePerTapped == 0 && t.Returns.IsZero() && !t.Reanimates &&
		t.UntilEndOfTurn == "" && t.Pump == nil
}

// HasPlayerAmount reports whether the entry names an amount only a
// player can be given: a draw, a discard, a token or a life change.
func (t TargetPurpose) HasPlayerAmount() bool {
	return t.Draws != 0 || t.Discards != 0 || t.Tokens != 0 || t.LifeGain != 0 || t.LifeLoss != 0 || t.ExtraTurns != 0
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

// Saves is the permanents a `survive` or `protect` row covers besides, or
// instead of, its own source (Purpose.Saves). It names a class the way a
// Sweep does, so a reader asks the same question of a sweep and of the
// answer to it: is this permanent in the class?
type Saves struct {
	// Matches is the class of its controller's permanents the row covers
	// ("creatures" for Selfless Spirit). It is one of SweepMatches.
	Matches SweepMatch
	// Partial is set when the row covers only some of the Matches class
	// by a condition this type does not name: creature tokens, creatures
	// with counters, nontoken Vampires. Matches is then an upper bound.
	Partial bool
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

// LandfallPayoff is what a landfall row pays its controller for each
// land that enters under their control, as printed amounts (#3072). The
// heuristic prices a land the bot puts onto the battlefield that much
// higher. An amount counted at resolution ("a +1/+1 counter on each
// Plant you control") or one the land's own name or count decides
// ("if you control seven or more differently named lands") is not
// declared, as everywhere in Purpose.
type LandfallPayoff struct {
	// Draws is the cards its controller draws: Tatyova 1.
	Draws int
	// LifeGain is the life its controller gains: Tatyova and Courser of
	// Kruphix 1, Primeval Bounty 3.
	LifeGain int
	// Tokens is the tokens it creates for its controller: Tireless
	// Tracker's Clue, Tireless Provisioner's Treasure or Food,
	// Sporemound's Saproling, Rampaging Baloths' Beast 1.
	Tokens int
	// Mana is the mana it adds: Lotus Cobra 1, Omnath, Locus of the
	// Void 2. It empties at the end of the step (CR 106.4).
	Mana int
	// Counters is the +1/+1 counters it puts on its source: Undergrowth
	// Champion 1.
	Counters int
	// Energy is the energy counters its controller gets: Architect of
	// the Untamed 1, Nissa, Worldsoul Speaker 2 (ADR 0129).
	Energy int
	// DamageEachOpponent is the damage its source deals to each
	// opponent: Sabotender 1.
	DamageEachOpponent int
}

// CreatureEnterPayoff is what a row pays its controller for each
// creature that enters under their control, in LandfallPayoff's units
// and with the same rules (#3165): printed amounts only, none counted at
// resolution.
type CreatureEnterPayoff LandfallPayoff

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
func CardPurposeOf(c Card) Purpose { return CardPurposeOfPtr(&c) }

// CardPurposeOfPtr is CardPurposeOf without the receiver copy (#2113).
func CardPurposeOfPtr(c *Card) Purpose {
	d := catalogDef(catalogKeyOf(c))
	if d == nil {
		// CR 708.2a: a face-down permanent has no text, so no purpose
		// of its own. What turning it face up does is not text it has;
		// it is the card underneath's, which only the viewer allowed to
		// look at it is sent (redactCardForViewer clears Purpose
		// otherwise). Nothing else of the card's purpose crosses.
		if c.faceDownPermanent() {
			if fd := catalogDef(underlyingTextKeyOf(c)); fd != nil && fd.Purpose.FaceUp != nil {
				return Purpose{FaceUp: fd.Purpose.FaceUp}
			}
		}
		return Purpose{}
	}
	p := d.Purpose
	p.FlashCleanupSacrifice = d.FlashWithCleanupSacrifice
	return p
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
		CreatureTokens:            p.CreatureTokens + o.CreatureTokens,
		TokensPerControlled:       p.TokensPerControlled,
		Energy:                    p.Energy + o.Energy,
		Sweep:                     p.Sweep,
		DeathPayoff:               p.DeathPayoff || o.DeathPayoff,
		DiscardPayoff:             p.DiscardPayoff,
		OnOpponentDiscard:         p.OnOpponentDiscard,
		LandfallPayoff:            p.LandfallPayoff,
		CreatureEnterPayoff:       p.CreatureEnterPayoff,
		Pump:                      p.Pump,
		ExtraCombat:               p.ExtraCombat + o.ExtraCombat,
		ExtraTurns:                p.ExtraTurns + o.ExtraTurns,
		SkipsTurns:                p.SkipsTurns + o.SkipsTurns,
		PreventCombatDamageToSelf: p.PreventCombatDamageToSelf || o.PreventCombatDamageToSelf,
		DamageToCreature:          p.DamageToCreature + o.DamageToCreature,
		DamageEachOpponent:        p.DamageEachOpponent + o.DamageEachOpponent,
		LifeGain:                  p.LifeGain + o.LifeGain,
		AwakenLand:                p.AwakenLand + o.AwakenLand,
		ExtraLandDrops:            p.ExtraLandDrops + o.ExtraLandDrops,
		LandfallExtraTriggers:     p.LandfallExtraTriggers + o.LandfallExtraTriggers,
		ExtraLoyaltyActivations:   p.ExtraLoyaltyActivations + o.ExtraLoyaltyActivations,
		Impulse:                   p.Impulse + o.Impulse,
		Answers:                   p.Answers | o.Answers,
	}
	if out.Pump == nil {
		out.Pump = o.Pump
	}
	if out.Sweep.IsZero() {
		out.Sweep = o.Sweep
	}
	out.Kicked, out.FaceUp = p.Kicked, p.FaceUp
	if out.Kicked == nil {
		out.Kicked = o.Kicked
	}
	if out.FaceUp == nil {
		out.FaceUp = o.FaceUp
	}
	if out.DiscardPayoff == nil {
		out.DiscardPayoff = o.DiscardPayoff
	}
	if out.OnOpponentDiscard == nil {
		out.OnOpponentDiscard = o.OnOpponentDiscard
	}
	if out.LandfallPayoff == nil {
		out.LandfallPayoff = o.LandfallPayoff
	}
	if out.TokensPerControlled == "" {
		out.TokensPerControlled = o.TokensPerControlled
	}
	if out.CreatureEnterPayoff == nil {
		out.CreatureEnterPayoff = o.CreatureEnterPayoff
	}
	targets := append([]TargetPurpose(nil), p.Targets.List()...)
	for _, t := range o.Targets.List() {
		t.Slot += leftClauses
		targets = append(targets, t)
	}
	out.Targets = ForTargets(targets...)
	return out
}
