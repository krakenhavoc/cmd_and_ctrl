package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// scoped_effects.go is ADR 0041 phase 3's data record for a continuous
// effect created by a resolving spell or ability (CR 611.2) — see the
// 2026-09-24 amendment to ADR 0041, Decision P1, and #1497.
//
// WHY DATA. The registry this replaced, `ScopedStatic`, held a
// `StaticAbility`, which is two closures, so a game holding one was not
// a restore point (`ContinuationCensus.ScopedStatics`). An earthbent
// land or an Agent of Treachery's theft kept one alive for the rest
// of the game, and a Giant Growth or a crewed Vehicle until cleanup. A
// `ScopedEffect` says the same thing as DATA: the objects it affects, a
// list of operations from a closed vocabulary, one timestamp and an
// ADR 0063 `Duration`. The snapshot carries it verbatim, and the
// running binary interprets it.
//
// The measurement behind the vocabulary is in the ADR: every scoped
// static in the tree — 115 card files and every engine site — is one
// of twelve operations over a pinned object set. Nothing needs a
// bespoke closure, so there is no per-card registry here; the
// `ModKind` string is the key, and the engine owns its namespace.
//
// Tier 3a (ADR 0041 P5) moved the last writer and deleted the closure
// registry, so this is the only place a duration-scoped continuous
// effect lives. It is swept by `durationExpiredLocked`, like every
// other duration in the game.

// ModKind names one operation a ScopedEffect performs. It is a STRING
// on disk, never an int, so renumbering constants can never change
// what an old restore point means — and it is an on-disk identity like
// a token slug or a grant key: never renamed, never reused. A mod
// whose meaning has to change gets a new kind.
//
// A restore point naming a kind this binary does not know is refused
// with ErrUnknownEffectKey rather than guessed at (ADR 0041 P4), which
// is what lets the vocabulary grow within a schema version.
type ModKind string

// The kinds. The layer each belongs to is fixed by modKinds below, so
// a record cannot put an operation in the wrong layer.
//
// `grantAbilities` (ADR 0093's ScopedGrant) is the last block below:
// ADR 0093 PR 4, #1584.
const (
	ModSetController    ModKind = "setController"    // layer 2
	ModAddTypes         ModKind = "addTypes"         // layer 4
	ModRemoveTypes      ModKind = "removeTypes"      // layer 4
	ModAddSubtypes      ModKind = "addSubtypes"      // layer 4
	ModAllCreatureTypes ModKind = "allCreatureTypes" // layer 4
	ModSetColors        ModKind = "setColors"        // layer 5
	ModAddKeywords      ModKind = "addKeywords"      // layer 6
	ModRemoveKeywords   ModKind = "removeKeywords"   // layer 6
	ModLoseAllAbilities ModKind = "loseAllAbilities" // layer 6
	ModAddRestrictions  ModKind = "addRestrictions"  // layer 6
	ModSetBasePower     ModKind = "setBasePower"     // layer 7b
	ModSetBaseToughness ModKind = "setBaseToughness" // layer 7b
	ModModifyPT         ModKind = "modifyPT"         // layer 7c
	// ModAddAttackRequirement is a CR 508.1d attack requirement
	// (#1571): "attacks … if able", or — with Player set — "attacks a
	// player other than Player if able". Layer 6 for the reason
	// ModAddRestrictions is: not a characteristic, written where the
	// text sits, only ever appended to (attack_requirements.go).
	ModAddAttackRequirement ModKind = "addAttackRequirement" // layer 6
	// ModAddBlockRequirement is a CR 509.1c block requirement (#1597):
	// Text names which one (BlockRequirementKind — "blocks", "lure",
	// "mustBeBlocked", "exactlyOne", "blocksAttacker"). Layer 6 for
	// ModAddAttackRequirement's reason (block_requirements.go).
	// Irresistible Prey's "target creature must be blocked this turn if
	// able", Taunting Challenge's "all creatures able to block target
	// creature this turn do so", and — with Objects naming the attacker
	// (#1684) — Provoke's "block it if able".
	ModAddBlockRequirement ModKind = "addBlockRequirement" // layer 6
	// ModAddBlockCapacity is "can block an additional creature this
	// turn" from a resolving spell or ability (#1715): Coastline
	// Chimera's and Mounted Archers' activations, Act of Heroism,
	// Yare's "up to two additional creatures". Reads Amount, the number
	// of ADDITIONAL attackers (at least 1), added to
	// Characteristic.AdditionalBlocks exactly as CanBlockAdditional's
	// static adds it, so several add up. Layer 6 for the reason the
	// static is: "can block" is an ability the effect grants.
	ModAddBlockCapacity ModKind = "addBlockCapacity" // layer 6
	// ModBlockAnyNumber is "can block any number of creatures this
	// turn" (#1715): Give No Ground, Valor Made Real, Blaze of Glory.
	// Reads nothing; sets Characteristic.BlocksAnyNumber, which beats
	// any count. Layer 6, like ModAddBlockCapacity.
	ModBlockAnyNumber ModKind = "blockAnyNumber" // layer 6
)

// ModGrantAbilities gives each affected object the named catalog
// ability BUNDLES (effects.AbilityGrant) for the record's duration —
// ADR 0093 Decision 7's ScopedGrant, as a mod (#1584; ADR 0041 phase 3
// owner decision 1). "Until end of turn, target creature gains 'When
// this creature dies, return it to the battlefield tapped …'" (Feign
// Death), "gains '{T}: Return target nonland permanent …'" (Retraction
// Helix), Urza's Saga's chapters with no duration at all.
//
// Layer 6, in the record's timestamp slot: the adapter builds the very
// StaticAbility.GrantAbilities declaration a granting permanent's
// static makes, so the recipient's Characteristic.GrantedAbilities,
// the `grant:` refs, the auto-tapper's tiering, CR 613.6's "a later
// removal takes it, an earlier one does not", and CR 707.2's "a copy
// does not copy it" are all the SAME code a static grant runs. The
// grantor recorded on the recipient is the record's source.
//
// Reads Grants. Each entry is a bundle key (either spelling of
// GrantKey); a restore point naming a bundle this binary's catalog does
// not register is refused (ErrUnknownEffectKey), exactly as an unknown
// kind is — the bundle is as much a part of the effect's meaning as the
// kind is.
const ModGrantAbilities ModKind = "grantAbilities" // layer 6

// ModCantAttackUnlessDefenderControls is "<affected> can't attack unless
// defending player controls <a permanent>" given by a resolved effect
// (#1879, ADR 0107 §2): Veiled Serpent's trigger makes it "a 4/4 Serpent
// creature with 'This creature can't attack unless defending player
// controls an Island.'" Reads Queries, any of them, at least one.
//
// Layer 6, in the record's timestamp slot: the quoted text is an ability
// the permanent gains (CR 613.1f). It appends the same
// AttackTargetRestriction the printed static writes
// (effects.CantAttackUnlessDefendingPlayerControls), attributed to the
// record's source, so every reader — both declaration verbs, the CR
// 508.1d search, the enumerator, the chip — is the printed card's. When
// the source is the affected permanent itself, the restriction is that
// permanent's own ability, so a later "loses all abilities" takes it
// (dropSelfAttackTargetRestrictions) and an earlier one does not.
const ModCantAttackUnlessDefenderControls ModKind = "cantAttackUnlessDefenderControls" // layer 6

// The replacement kinds (ADR 0041 P8, tier 3b, #1497). These are not
// layer operations: each one is a CR 614 replacement effect a
// resolving spell or ability created (CR 611.2), read by the
// replacement gather rather than by the layer pass. See
// scoped_replacements.go for what each one does and the registration
// function that writes it.
const (
	// ModPreventCombatDamage is "prevent all combat damage that would
	// be dealt this turn" (Fog, CR 615.1) — or, with Player set, "…
	// that would be dealt to <Player>" (Druid's Deliverance). Reads
	// Player. Scope ScopeGame.
	ModPreventCombatDamage ModKind = "preventCombatDamage"
	// ModPreventDamage is "prevent the next N damage that would be
	// dealt to <target> this turn" (Mending Hands, CR 615.7). Reads
	// Amount, the charge LEFT (at least 1), and CombatOnly. A pinned
	// object, or ScopeGame plus Player for a player.
	ModPreventDamage ModKind = "preventDamage"
	// ModExileInsteadOfLeaving is "if it would leave the battlefield,
	// exile it instead of putting it anywhere else" (Whip of Erebos,
	// unearth). Reads nothing. A pinned object.
	ModExileInsteadOfLeaving ModKind = "exileInsteadOfLeaving"
	// ModExileInsteadOfGraveyard is "if a permanent you control would
	// be put into a graveyard from the battlefield this turn, exile it
	// instead" plus a delayed trigger per redirected card (Cosmic
	// Intervention). Reads Then, a registered delayed-trigger body key.
	// Scope ScopeYourPermanents.
	ModExileInsteadOfGraveyard ModKind = "exileInsteadOfGraveyard"
	// ModGainNoLife is "if <Player> would gain life this turn, that
	// player gains no life instead" (Flames of the Blood Hand, ADR 0107
	// §5): a CR 614 replacement on the life window, NOT CR 119.7's
	// "can't gain life" (ModCantGainLife). The difference is CR 616:
	// another "if you would gain life" replacement may be ordered
	// before this one, where "can't" stops the gain before any
	// replacement sees it. Reads Player. Scope ScopeGame.
	ModGainNoLife ModKind = "gainNoLife"
	// ModExileInsteadOfYourGraveyard is "if a card would be put into
	// your graveyard from anywhere this turn, exile that card instead"
	// (Yawgmoth's Will, ADR 0108 §4, CR 614.1a): a CR 614 replacement on
	// every zone move and discard into Player's graveyard, a token
	// excepted (a token is not a card). Reads Player. Scope ScopeGame.
	ModExileInsteadOfYourGraveyard ModKind = "exileInsteadOfYourGraveyard"
)

// The rules kinds (ADR 0107 §5, #1853, #1880). Not layer operations and
// not replacements: each is a CR 613.11 rule-modifying effect a
// resolving spell or ability created, read at exactly one gate —
// damageUnpreventableLocked, damageCantBeRedirectedLocked
// (unpreventable_damage.go) or playerCantGainLifeLocked
// (cant_gain_life.go). The ADR 0106 §4 "can't be countered" shape, with
// a duration.
const (
	// ModDamageCantBePrevented is "damage can't be prevented this turn"
	// (Skullcrack, CR 615.12). Scope ScopeGame: all damage. Pinned to
	// one permanent: "damage that would be dealt to that creature this
	// turn can't be prevented" (Whippoorwill). Reads nothing.
	ModDamageCantBePrevented ModKind = "damageCantBePrevented"
	// ModDamageCantBeRedirected is "… can't be dealt instead to another
	// permanent or player" (Whippoorwill): pinned to the permanent the
	// damage would be dealt to. Reads nothing.
	ModDamageCantBeRedirected ModKind = "damageCantBeRedirected"
	// ModCantGainLife is "<players> can't gain life" for a duration
	// (CR 119.7): Skullcrack's "players … this turn", Atarka's
	// Command's "your opponents … this turn", Screaming Nemesis's "they
	// … for the rest of the game". Scope ScopeGame with Player set is
	// that one player, and with Player zero every player;
	// ScopeOpponentsAndTheirCreatures is the record Controller's
	// opponents (scopeCoversPlayer). Reads Player.
	ModCantGainLife ModKind = "cantGainLife"
)

// The block-rule kinds (ADR 0041 P8, tier 3b, #1497). These are not
// layer operations either: each one is a CR 509.1b block restriction a
// resolving spell or ability created, read by the block-rule walk
// (forEachBlockRuleLocked's third pass) rather than by the layer pass.
// See scoped_block_rules.go for what each one does and the
// registration functions that write them.
const (
	// ModCantBeBlockedExceptBy is "<creature> can't be blocked this
	// turn except by <keyword-or-subtype>" (Gingerbrute's activated
	// ability, Departed Deckhand's granted evasion). Reads Keywords and
	// Subtypes, each an any-of — a blocker matching either is allowed —
	// and Text, the allowed set as the card prints it, which the
	// refusal sentence reads. The pinned attacker(s).
	ModCantBeBlockedExceptBy ModKind = "cantBeBlockedExceptBy"
	// ModLimitBlockersPerDefender is "each opponent can't block with
	// more than N creatures this combat" (Mirri, Weatherlight Duelist's
	// attack trigger). Reads Amount. Scope ScopeOpponentsCreatures,
	// read live against the record's Controller (#1571's style: the
	// rule is about players, not a characteristic, so CR 611.2c does
	// not lock it).
	ModLimitBlockersPerDefender ModKind = "limitBlockersPerDefender"
)

// The hexproof kinds (#1651, ADR 0038's amendment of 2026-09-28). See
// cant_have.go and hexproof_bypass.go for what each one does.
const (
	// ModCantHaveKeywords is "loses <keywords> and can't have
	// <keywords>" (Arcane Lighthouse, CR 101.2). Reads Keywords. Layer
	// 6: its Apply records the tokens on Characteristic.CantHave, and
	// the strip after the layer-6 bucket removes them whatever granted
	// them, so a later grant cannot put one back.
	ModCantHaveKeywords ModKind = "cantHaveKeywords" // layer 6
	// ModWaiveHexproof is "<affected> can be the targets of spells and
	// abilities you control as though they didn't have hexproof"
	// (Detection Tower, CR 702.11). Reads nothing; the beneficiary is
	// the record's Controller. Not a layer operation: the targeting
	// choke point reads it (hexproof_bypass.go).
	ModWaiveHexproof ModKind = "waiveHexproof"
)

// ModBecomeCopy is "<affected> becomes a copy of <object> [until end of
// turn]" (#1593, ADR 0043's amendment of 2026-09-28): a CR 707.2 copy
// effect with a duration and a timestamp of its own, applied to a
// permanent that is already on the battlefield. Mirage Mirror,
// Cytoshape, Mirrorweave, Shifting Woodland's delirium, Unstable
// Shapeshifter, Lazav.
//
// Reads Copy, which holds exactly ONE PrintedValues: the copied
// object's copiable values as they were when the effect began, with the
// card's "except" clause already applied. Values, never a reference —
// the copied card may be in a graveyard (Shifting Woodland, Lazav) and
// may leave it, and the copy does not end when it does.
//
// Not a layer-pass operation. A copy has to change the oracle ID every
// catalog hook keys on, which no Characteristic field carries (ADR 0043
// Decision 1), so it is MATERIALISED onto the permanent's flat printed
// fields before the layer pass runs, by materialiseDurationCopiesLocked
// (duration_copy.go). That is layer 1 by construction: layers 2-7 then
// run on its result (CR 613.1a).
const ModBecomeCopy ModKind = "becomeCopy"

// AffectedScope is a ScopedEffect's affected set as a RULE read live at
// every layer pass, instead of a set of objects locked when the effect
// began (#1571). CR 611.2c locks the set only for an effect that
// changes characteristics or control; an effect that does neither
// "modifies the rules of the game, so it can affect objects that
// weren't affected when that continuous effect began" — Bident of
// Thassa's "creatures your opponents control attack this turn if able"
// reaches a creature an opponent casts after it resolved.
//
// A closed vocabulary like ModKind, and an on-disk identity for the
// same reason: never renamed, never reused. A restore point naming a
// scope this binary does not know is refused (ErrUnknownEffectKey).
type AffectedScope string

const (
	// ScopeNone is the ordinary record: Affected is the set.
	ScopeNone AffectedScope = ""
	// ScopeOpponentsCreatures is "creatures your opponents control",
	// the "you" being the record's Controller.
	ScopeOpponentsCreatures AffectedScope = "opponentsCreatures"
	// ScopeGame is an effect that names no object at all (tier 3b,
	// #1497): Fog's "all combat damage", or a prevention shield on a
	// PLAYER, whose Mod names the player. It matches no permanent.
	ScopeGame AffectedScope = "game"
	// ScopeYourPermanents is "permanents you control", read live — the
	// "you" being the record's Controller (tier 3b, Cosmic
	// Intervention). A replacement effect is not a characteristic, so
	// CR 611.2c does not lock its set.
	ScopeYourPermanents AffectedScope = "yourPermanents"
	// ScopeYourCreatures is "creatures you control", read live — the
	// "you" being the record's Controller (#1650, Glaring Spotlight's
	// "creatures you control … can't be blocked this turn").
	ScopeYourCreatures AffectedScope = "yourCreatures"
	// ScopeCreaturesWithoutFlying is "creatures without flying", every
	// player's, read live (#1650, Falter). Whether a creature has
	// flying is a layer-6 result, so a record with this scope may only
	// carry mods that are read after the layer pass is finished — see
	// foldRuleScopedRestrictionsLocked. Registration enforces it.
	ScopeCreaturesWithoutFlying AffectedScope = "creaturesWithoutFlying"
	// ScopeOpponentsAndTheirCreatures is "your opponents and creatures
	// your opponents control", read live — the "you" being the record's
	// Controller (#1651, Detection Tower). The one scope that also
	// names PLAYERS (scopeCoversPlayer); as a permanent predicate it is
	// ScopeOpponentsCreatures.
	ScopeOpponentsAndTheirCreatures AffectedScope = "opponentsAndTheirCreatures"
)

// KnownAffectedScope reports whether this binary can interpret s.
func KnownAffectedScope(s AffectedScope) bool {
	switch s {
	case ScopeNone, ScopeOpponentsCreatures, ScopeGame, ScopeYourPermanents,
		ScopeYourCreatures, ScopeCreaturesWithoutFlying, ScopeOpponentsAndTheirCreatures:
		return true
	}
	return false
}

// Mod is one operation of a ScopedEffect, with its plain parameters.
// Which fields a kind reads is documented on its constructor; the rest
// stay zero and are omitted on disk.
//
// PARAMETER TYPES ARE CLOSED (ADR 0041 P1): bool, ints, strings,
// uuid.UUID, the engine's string and bit enums, and slices of those.
// Never a func, an interface, a pointer, a Card or a CardPredicate —
// TestScopedEffectIsPureData holds the line.
type Mod struct {
	Kind         ModKind     `json:"kind"`
	Types        []string    `json:"types,omitempty"`
	Subtypes     []string    `json:"subtypes,omitempty"`
	Colors       []string    `json:"colors,omitempty"`
	Keywords     []string    `json:"keywords,omitempty"`
	Restrictions Restriction `json:"restrictions,omitempty"`
	Power        int         `json:"power,omitempty"`
	Toughness    int         `json:"toughness,omitempty"`
	Player       uuid.UUID   `json:"player,omitempty"`
	// Grants is ModGrantAbilities' bundle keys, in GrantKey form.
	Grants []string `json:"grants,omitempty"`
	// Amount is ModPreventDamage's charge LEFT (tier 3b). A shield that
	// absorbs damage is replaced by a record with a lower Amount, never
	// mutated in place (see ScopedEffect's contract).
	Amount int `json:"amount,omitempty"`
	// CombatOnly narrows ModPreventDamage to combat damage.
	CombatOnly bool `json:"combatOnly,omitempty"`
	// Then is ModExileInsteadOfGraveyard's delayed-trigger body key: a
	// registered BodyRef's key, scheduled at the next end step for each
	// card the replacement redirects. On ModPreventNextFromSource it is
	// the CR 615.5 follow-up body, run with the damage prevented. A
	// restore point naming a body this binary has not registered is
	// refused (ErrUnknownEffectKey).
	Then string `json:"then,omitempty"`
	// Text is ModCantBeBlockedExceptBy's printed parameter — "creatures
	// with haste", "Spirits" — read by the refusal sentence
	// (BlockRule.Label).
	Text string `json:"text,omitempty"`
	// Objects are the objects a mod names — the one attacking object a
	// "blocksAttacker" block requirement names (#1684;
	// BlocksAttackerMod), Provoke's "block IT", and the chosen source of
	// a ModPreventNextFromSource shield (ADR 0107 §6). Required (one
	// entry) on blocksAttacker, at most one on the shield, and refused
	// on every other mod. A slice, so
	// every other mod writes nothing and the record stays plain data
	// (ADR 0041 P1); an older binary refuses a file carrying one
	// (ADR 0041 P4), which the unknown kind already guarantees.
	Objects []ObjectRef `json:"objects,omitempty"`
	// SourceZone, Queries and SpentBatch are ModPreventNextFromSource's
	// (ADR 0107 §6, #1860; prevent_next_from_source.go), refused on every
	// other kind. SourceZone is the zone the chosen source (Objects[0])
	// was in when it was chosen, so a permanent spell's shield follows it
	// onto the battlefield and a permanent's follows it as it last existed
	// (CR 609.7a). Queries is the CR 615.9 recheck, any of them. SpentBatch
	// is the event batch the shield prevented its instance of damage in
	// (CR 615.8); zero is unspent. SpentInstance (ADR 0108 PR 0) is that
	// instance itself (damage_instance.go): a spent shield keeps applying
	// to the rest of its instance and to nothing else. Both are written. A
	// record from before SpentInstance existed has SpentBatch alone and
	// reads as it did, applying for the rest of that batch.
	SourceZone    ZoneKind         `json:"sourceZone,omitempty"`
	Queries       []PermanentQuery `json:"queries,omitempty"`
	SpentBatch    uint64           `json:"spentBatch,omitempty"`
	SpentInstance DamageInstance   `json:"spentInstance,omitempty"`
	// Copy is ModBecomeCopy's copied values (#1593): exactly one entry,
	// required on that kind and refused on every other. A slice for the
	// reason Objects is one — every other mod writes nothing, and the
	// record stays plain data with no pointer in it (ADR 0041 P1).
	Copy []PrintedValues `json:"copy,omitempty"`
}

// AffectedObject is one member of a ScopedEffect's affected set: the
// key every CR 611.2c set in the engine uses — an instance plus the
// battlefield-entry stamp it carried when the effect began, so a
// permanent that left and came back is a new object the effect no
// longer follows (CR 400.7).
//
// EnteredAt 0 means "this instance, whatever its entry stamp", the
// convention `sameObjectOnBattlefieldLocked` already uses. It is how
// suspend's haste follows its card from the stack onto the battlefield
// (CR 702.62a), and it is safe only because that record's Duration
// ends the effect the moment the card stops being the object it named.
// Nothing else may write it: pin a permanent with PinObject.
//
// Unstamped is the other side of that line (#1558): a permanent that
// carries no entry stamp — a fixture's seeded card, one put onto the
// battlefield without the zone-move event — pinned EXACTLY. It matches
// that instance only while it is still unstamped, so a flicker, which
// stamps the new object, ends it (CR 400.7), the way the legacy
// closure's `== stamp` comparison always did. Before #1558 such a
// permanent was written as the wildcard and stayed matched. A new
// field rather than a sentinel stamp, so an older v7 binary refuses a
// file carrying one (ADR 0041 P4) instead of reading it as a stamp no
// object has.
//
// Controller, when set, narrows the member further: the object is
// affected only while that player controls it. Suspend's haste again —
// "it has haste until that player loses control of it".
//
// OnStack and Epoch are the stack's pin (ADR 0104, #1745): the member
// is a SPELL, named by its instance ID and the Card.ObjectEpoch it has
// on the stack. MoveCard bumps the epoch on every zone change, so the
// same card cast again is a new object the pin does not match
// (CR 400.7). A stack member never matches a permanent: the battlefield
// pass skips it (affectedPredicate), and only the stack step of the
// layer pass reads it (stackControlPassLocked). Built with
// PinStackObject, never by hand.
//
// Epoch on a BATTLEFIELD member (OnStack false, EnteredAt 0, Unstamped
// false, Epoch above zero) pins the permanent by its ObjectEpoch
// instead of its entry stamp: PinObjectByEpoch. It exists for the one
// pin that has to be taken before the entry stamp is — the permanent a
// stolen permanent spell becomes, re-pinned as it lands and before the
// zone-move event stamps it (ADR 0104, CR 400.7a). An epoch is final
// the moment MoveCard lands the card, and a card that moved from the
// stack has an epoch of at least 1, so zero stays the wildcard.
type AffectedObject struct {
	ID         uuid.UUID `json:"id"`
	EnteredAt  int64     `json:"enteredAt,omitempty"`
	Unstamped  bool      `json:"unstamped,omitempty"`
	Controller uuid.UUID `json:"controller,omitempty"`
	OnStack    bool      `json:"onStack,omitempty"`
	Epoch      int       `json:"epoch,omitempty"`
}

// PinStackObject is the affected-set member for the spell `id` on the
// stack, as the object it is now: its instance plus its ObjectEpoch
// (ADR 0104).
func PinStackObject(id uuid.UUID, epoch int) AffectedObject {
	return AffectedObject{ID: id, OnStack: true, Epoch: epoch}
}

// PinObjectByEpoch is the battlefield member for the permanent `id` as
// the object with ObjectEpoch `epoch` (ADR 0104). `epoch` must be above
// zero; zero would be the wildcard.
func PinObjectByEpoch(id uuid.UUID, epoch int) AffectedObject {
	return AffectedObject{ID: id, Epoch: epoch}
}

// memberMatches reports whether the permanent `target` is the object
// the battlefield member `a` names — by epoch for an epoch pin, by
// entry stamp otherwise. A stack member never matches a permanent.
func memberMatches(a AffectedObject, target *Card) bool {
	if a.OnStack || a.ID != target.InstanceID {
		return false
	}
	if a.Epoch > 0 && a.EnteredAt == 0 && !a.Unstamped {
		return target.ObjectEpoch == a.Epoch
	}
	return entryMatches(a.EnteredAt, a.Unstamped, target.EnteredBattlefieldAt)
}

// PinObject is the affected-set member for the permanent `id` that
// entered the battlefield at `stamp`: pinned to that stamp, or — for an
// unstamped permanent — to being unstamped. Never the wildcard.
func PinObject(id uuid.UUID, stamp int64) AffectedObject {
	return AffectedObject{ID: id, EnteredAt: stamp, Unstamped: stamp == 0}
}

// entryMatches reports whether a permanent whose entry stamp is now
// `stamp` is the object an (enteredAt, unstamped) pin names: an exact
// "still unstamped" when unstamped is set, otherwise the 0 wildcard or
// an exact stamp. The one reading of a pin, shared by the affected set
// and Duration.Pinned.
func entryMatches(enteredAt int64, unstamped bool, stamp int64) bool {
	if unstamped {
		return stamp == 0
	}
	return enteredAt == 0 || stamp == enteredAt
}

// ScopedEffect is a continuous effect created by a resolving spell or
// ability (CR 611.2), as data. See the file comment.
//
// IMMUTABILITY CONTRACT: every field
// is written once at registration and never mutated. Clone copies the
// slice into a fresh backing array and shares the inner slices.
//
// One clarification, for the one record that changes (ADR 0041 P8): a
// ModPreventDamage shield that absorbs damage is REPLACED, never
// edited — the registry slice is rebuilt with a new record at that
// position (fresh Mods, lower Amount), and a spent shield is removed.
// A clone taken before the damage keeps the old record, so an undo
// rewinds the charge.
type ScopedEffect struct {
	// Affected is the CR 611.2c set, locked when the effect began.
	Affected []AffectedObject `json:"affected"`

	// Scope, when set, replaces Affected with a rule read live at
	// every pass (#1571): the record affects whatever the rule matches
	// now, including objects that did not exist when it began. Only
	// for an effect that changes neither characteristics nor control
	// (CR 611.2c): an attack requirement, a block rule, a replacement,
	// or (#1650) a restriction such as Falter's "can't block".
	Scope AffectedScope `json:"scope,omitempty"`

	// Mods are applied each in its own layer, all at Timestamp. One
	// record may span several layers — earthbend is layer 4, 6 and 7b
	// — and that is what makes it ONE effect for CR 613.7.
	Mods []Mod `json:"mods"`

	// Source is the spell or ability's source object, as an identity:
	// what `ControlSource` names for a control change (#930), and what
	// an operator reads. Last-known information is enough.
	Source     ObjectRef `json:"source"`
	SourceName string    `json:"sourceName,omitempty"`

	// Controller is who controlled the source as the effect began —
	// the "you" of the effect, read once (CR 611.2c).
	Controller uuid.UUID `json:"controller,omitempty"`

	// Timestamp is the CR 613.7 timestamp. The two halves of an
	// exchange of control share one (CR 701.12).
	Timestamp int64 `json:"timestamp"`

	// Duration is how long the effect lasts (ADR 0063), swept by
	// `durationExpiredLocked` like every other duration in the game.
	Duration Duration `json:"duration"`

	// Label is human-readable attribution for logs and tests.
	Label string `json:"label,omitempty"`

	// Seq is the record's identity for a reader that has to name it
	// across a pause (tier 3b, ADR 0041 P8): a replacement effect's
	// ReplacementEffectID is minted from it, because a CR 616 prompt
	// holds that ID while the registry shrinks under it (a sweep, a
	// pin collected, a spent shield) and a position would then name a
	// different record. Taken from Game.scopedEffectSeq at registration,
	// and only for a record with a non-layer mod: a layer-only record
	// is never named, and leaving it zero keeps every such record the
	// bytes an earlier v7 binary wrote and reads.
	Seq int64 `json:"seq,omitempty"`
}

// modReader is which part of the engine interprets a kind (ADR 0041
// P8): the layer pass, or — for the tier 3b kinds — the replacement
// gather or the block-rule walk. The layer adapter skips every kind
// that is not its own.
type modReader uint8

const (
	readerLayer modReader = iota
	readerReplacement
	readerBlockRule
	// readerTargeting is the targeting choke point (#1651).
	readerTargeting
	// readerCopy is the layer-1 materialiser (#1593,
	// duration_copy.go): it runs before the layer pass, not in it.
	readerCopy
	// readerRule is a CR 613.11 rules gate (ADR 0107 §5): the damage
	// prevention and redirection gates and the life-gain gate.
	readerRule
)

// modKindSpec is where a kind lives: its reader, and for a layer kind
// its place in the layer system.
type modKindSpec struct {
	reader   modReader
	layer    Layer
	subLayer SubLayer
	removes  bool // CR 613.1f ability removal (ADR 0046)
}

// modKinds is the closed vocabulary. A kind missing here is unknown:
// registration refuses it and restore refuses a file naming it.
var modKinds = map[ModKind]modKindSpec{
	ModSetController:    {layer: Layer2Control},
	ModAddTypes:         {layer: Layer4Type},
	ModRemoveTypes:      {layer: Layer4Type},
	ModAddSubtypes:      {layer: Layer4Type},
	ModAllCreatureTypes: {layer: Layer4Type},
	ModSetColors:        {layer: Layer5Color},
	ModAddKeywords:      {layer: Layer6Ability},
	ModRemoveKeywords:   {layer: Layer6Ability},
	ModLoseAllAbilities: {layer: Layer6Ability, removes: true},
	ModAddRestrictions:  {layer: Layer6Ability},
	ModSetBasePower:     {layer: Layer7PT, subLayer: SubLayer7B_Set},
	ModSetBaseToughness: {layer: Layer7PT, subLayer: SubLayer7B_Set},
	ModModifyPT:         {layer: Layer7PT, subLayer: SubLayer7C_Modify},
	// #1571
	ModAddAttackRequirement: {layer: Layer6Ability},
	// #1597
	ModAddBlockRequirement: {layer: Layer6Ability},
	// #1715
	ModAddBlockCapacity: {layer: Layer6Ability},
	ModBlockAnyNumber:   {layer: Layer6Ability},
	// #1584, ADR 0093 PR 4
	ModGrantAbilities: {layer: Layer6Ability},
	// #1879: Veiled Serpent's granted attack restriction.
	ModCantAttackUnlessDefenderControls: {layer: Layer6Ability},
	// Tier 3b (ADR 0041 P8): replacement effects, not layer operations.
	ModPreventCombatDamage:     {reader: readerReplacement},
	ModPreventDamage:           {reader: readerReplacement},
	ModExileInsteadOfLeaving:   {reader: readerReplacement},
	ModExileInsteadOfGraveyard: {reader: readerReplacement},
	// ADR 0107 §5 (#1880): Flames of the Blood Hand's replacement.
	ModGainNoLife: {reader: readerReplacement},
	// ADR 0108 §4 (#1823): Yawgmoth's Will's replacement.
	ModExileInsteadOfYourGraveyard: {reader: readerReplacement},
	// ADR 0107 §6 (#1860): the next damage from a source.
	ModPreventNextFromSource: {reader: readerReplacement},
	// ADR 0107 §5 (#1853, #1880): rules gates.
	ModDamageCantBePrevented:  {reader: readerRule},
	ModDamageCantBeRedirected: {reader: readerRule},
	ModCantGainLife:           {reader: readerRule},
	// Tier 3b (ADR 0041 P8): block-rule effects, not layer operations.
	ModCantBeBlockedExceptBy:    {reader: readerBlockRule},
	ModLimitBlockersPerDefender: {reader: readerBlockRule},
	// #1651: "can't have" is a layer-6 record; the waiver is read by
	// targeting.
	ModCantHaveKeywords: {layer: Layer6Ability},
	ModWaiveHexproof:    {reader: readerTargeting},
	// #1593: layer 1, applied to the printed baseline before the pass.
	ModBecomeCopy: {reader: readerCopy, layer: Layer1Copy},
}

// KnownModKind reports whether this binary can interpret k.
func KnownModKind(k ModKind) bool {
	_, ok := modKinds[k]
	return ok
}

// ModKinds is the vocabulary, sorted — for tests and diagnostics.
func ModKinds() []ModKind {
	out := make([]ModKind, 0, len(modKinds))
	for k := range modKinds {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// ---------------------------------------------------------------
// Constructors — what effect code writes
// ---------------------------------------------------------------

// SetControllerMod is "that permanent is controlled by <player>"
// (layer 2, CR 613.1b). Reads Player.
func SetControllerMod(player uuid.UUID) Mod { return Mod{Kind: ModSetController, Player: player} }

// AddTypesMod adds card types in addition to the object's others
// (layer 4). Reads Types.
func AddTypesMod(types ...string) Mod { return Mod{Kind: ModAddTypes, Types: copyStrings(types)} }

// RemoveTypesMod removes card types (layer 4) — Enduring Curiosity's
// "it's not a creature". Reads Types.
func RemoveTypesMod(types ...string) Mod {
	return Mod{Kind: ModRemoveTypes, Types: copyStrings(types)}
}

// AddSubtypesMod adds subtypes in addition to the object's others
// (layer 4) — amass's "it's also an Orc", Kyoshi's Island. Reads
// Subtypes.
func AddSubtypesMod(subtypes ...string) Mod {
	return Mod{Kind: ModAddSubtypes, Subtypes: copyStrings(subtypes)}
}

// AllCreatureTypesMod is "is every creature type" (layer 4,
// Characteristic.AllCreatureTypes).
func AllCreatureTypesMod() Mod { return Mod{Kind: ModAllCreatureTypes} }

// SetColorsMod sets the object's colours (layer 5). Reads Colors;
// an empty list makes it colourless.
func SetColorsMod(colors ...string) Mod { return Mod{Kind: ModSetColors, Colors: copyStrings(colors)} }

// AddKeywordsMod grants keyword abilities (layer 6) through
// AppendKeywordAbility, so a cumulative keyword stays cumulative and
// any other is not duplicated. Reads Keywords.
func AddKeywordsMod(keywords ...string) Mod {
	return Mod{Kind: ModAddKeywords, Keywords: copyStrings(keywords)}
}

// RemoveKeywordsMod removes keyword abilities (layer 6). Reads
// Keywords.
func RemoveKeywordsMod(keywords ...string) Mod {
	return Mod{Kind: ModRemoveKeywords, Keywords: copyStrings(keywords)}
}

// LoseAllAbilitiesMod is "loses all abilities" (layer 6, a CR 613.1f
// removal — ADR 0046), appending `keep` back as the same effect's
// grant. Reads Keywords.
func LoseAllAbilitiesMod(keep ...string) Mod {
	return Mod{Kind: ModLoseAllAbilities, Keywords: copyStrings(keep)}
}

// AddRestrictionsMod adds restriction bits (layer 6, ADR 0045). Reads
// Restrictions.
func AddRestrictionsMod(r Restriction) Mod { return Mod{Kind: ModAddRestrictions, Restrictions: r} }

// SetBasePowerMod sets base power (layer 7b). Reads Power.
func SetBasePowerMod(n int) Mod { return Mod{Kind: ModSetBasePower, Power: n} }

// SetBaseToughnessMod sets base toughness (layer 7b). Reads Toughness.
func SetBaseToughnessMod(n int) Mod { return Mod{Kind: ModSetBaseToughness, Toughness: n} }

// AddAttackRequirementMod is a CR 508.1d attack requirement (#1571):
// "attacks … if able" with otherThan uuid.Nil, "attacks a player other
// than otherThan if able" with it set. The requirement is attributed to
// the record's source, so a refusal names the card.
func AddAttackRequirementMod(otherThan uuid.UUID) Mod {
	return Mod{Kind: ModAddAttackRequirement, Player: otherThan}
}

// AddBlockRequirementMod is a CR 509.1c block requirement (#1597) of
// the given kind. The requirement is attributed to the record's source,
// so a refusal names the card. Registration refuses a kind this binary
// does not know, and so does restore (ErrUnknownEffectKey).
func AddBlockRequirementMod(kind BlockRequirementKind) Mod {
	return Mod{Kind: ModAddBlockRequirement, Text: string(kind)}
}

// BlocksAttackerMod is "<affected creature> blocks <attacker> this turn
// if able" (#1684): Provoke, Grappling Hook, Turntimber Basilisk. The
// requirement names the attacking OBJECT, so it asks nothing once that
// creature has left the battlefield, even if its card comes back.
func BlocksAttackerMod(attacker ObjectRef) Mod {
	return Mod{Kind: ModAddBlockRequirement, Text: string(BlockRequirementBlocksAttacker), Objects: []ObjectRef{attacker}}
}

// AddBlockCapacityMod is "can block N additional creatures this turn"
// (#1715): N more attackers on top of the one every creature may
// block. Reads Amount; registration refuses N < 1.
func AddBlockCapacityMod(n int) Mod { return Mod{Kind: ModAddBlockCapacity, Amount: n} }

// BlockAnyNumberMod is "can block any number of creatures this turn"
// (#1715).
func BlockAnyNumberMod() Mod { return Mod{Kind: ModBlockAnyNumber} }

// blockCapacityModProblem is the registration and restore check for
// ModAddBlockCapacity: a positive count. "" when the mod is sound. A
// record with none would be an effect that grants nothing, and one
// with a negative count would TAKE capacity, which no card prints.
func blockCapacityModProblem(m Mod) string {
	if m.Kind == ModAddBlockCapacity && m.Amount < 1 {
		return fmt.Sprintf("an addBlockCapacity mod needs an amount of at least 1, got %d", m.Amount)
	}
	return ""
}

// blockRequirementModProblem is the registration check for a
// block-requirement mod: a known kind, and an attacker object exactly
// when the kind names one. "" when the mod is sound.
func blockRequirementModProblem(m Mod) string {
	if m.Kind != ModAddBlockRequirement {
		// ADR 0107 §6: the next-damage shield names its chosen source
		// here (nextFromSourceModProblem checks it).
		if len(m.Objects) != 0 && m.Kind != ModPreventNextFromSource {
			return fmt.Sprintf("mod %q names objects, which only a blocksAttacker requirement reads", m.Kind)
		}
		return ""
	}
	kind := BlockRequirementKind(m.Text)
	if !KnownBlockRequirementKind(kind) {
		return fmt.Sprintf("names unknown block requirement %q", m.Text)
	}
	names := len(m.Objects) == 1 && m.Objects[0].ID != uuid.Nil
	if (kind == BlockRequirementBlocksAttacker) != names {
		return fmt.Sprintf("block requirement %q needs one attacking object exactly when it is blocksAttacker", m.Text)
	}
	if kind != BlockRequirementBlocksAttacker && len(m.Objects) != 0 {
		return fmt.Sprintf("block requirement %q names objects it does not read", m.Text)
	}
	return ""
}

// GrantAbilitiesMod is "gains '<ability>'" (layer 6, ADR 0093 PR 4):
// each affected object gets the named catalog bundles, with the
// record's source as the grantor. Reads Grants. Keys are stored in
// GrantKey form; an empty key is dropped, and a mod with no key left
// is refused at registration.
func GrantAbilitiesMod(keys ...string) Mod {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if k = GrantKey(k); k != "" {
			out = append(out, k)
		}
	}
	return Mod{Kind: ModGrantAbilities, Grants: out}
}

// CantAttackUnlessDefenderControlsMod is "can't attack unless defending
// player controls <any of qs>" from a resolved effect (#1879). Reads
// Queries; registration refuses a mod with none.
func CantAttackUnlessDefenderControlsMod(qs ...PermanentQuery) Mod {
	return Mod{Kind: ModCantAttackUnlessDefenderControls, Queries: clonePermanentQueries(qs)}
}

// defenderControlsModProblem is the registration and restore check for
// ModCantAttackUnlessDefenderControls: at least one query, since a
// restriction that asks for nothing would refuse every target. "" when
// the mod is sound.
func defenderControlsModProblem(m Mod) string {
	if m.Kind == ModCantAttackUnlessDefenderControls && len(m.Queries) == 0 {
		return "a cantAttackUnlessDefenderControls mod names no permanent"
	}
	return ""
}

// SetBasePTMods is "has base power and toughness P/T" — both 7b
// halves.
func SetBasePTMods(power, toughness int) []Mod {
	return []Mod{SetBasePowerMod(power), SetBaseToughnessMod(toughness)}
}

// ModifyPTMod is "gets +P/+T" (layer 7c). Reads Power and Toughness;
// negative values shrink.
func ModifyPTMod(power, toughness int) Mod {
	return Mod{Kind: ModModifyPT, Power: power, Toughness: toughness}
}

// ---------------------------------------------------------------
// Registration
// ---------------------------------------------------------------

// PinnedObjectsLocked returns the affected set for the given
// battlefield permanents, each keyed on the entry stamp it carries
// now (PinObject: an unstamped permanent is pinned exactly, not as the
// wildcard — #1558). IDs not on the battlefield are skipped, so the
// result may be empty. Caller must hold g.mu.
func (g *Game) PinnedObjectsLocked(ids ...uuid.UUID) []AffectedObject {
	out := make([]AffectedObject, 0, len(ids))
	for _, id := range ids {
		if c, ok := g.battlefieldCardLocked(id); ok {
			out = append(out, PinObject(id, c.EnteredBattlefieldAt))
		}
	}
	return out
}

// RegisterScopedEffectForEffect installs a data-backed continuous
// effect and bumps the layer version so the next recompute applies it.
// It reports false, having registered nothing, when there is nothing
// to affect or nothing to do.
//
// `sourceID` names the spell or ability's source; it is looked up for
// its identity, name and controller, and a miss stores the bare ID.
//
// It PANICS on a mod kind this binary does not know: that is a
// programming error in the caller, caught by the first test that
// exercises it, and never a state a player can reach.
//
// Caller must hold g.mu (write). Effects call this from inside the
// resolution frame, which already holds it.
func (g *Game) RegisterScopedEffectForEffect(sourceID uuid.UUID, affected []AffectedObject, mods []Mod, d Duration, label string) bool {
	return g.registerScopedEffectLocked(sourceID, affected, mods, d, label, timeNowUnixNano())
}

// registerScopedEffectLocked is the shared body with the CR 613.7
// timestamp passed in — chosen rather than read only for the two
// halves of an exchange (CR 701.12). Caller must hold g.mu (write).
func (g *Game) registerScopedEffectLocked(sourceID uuid.UUID, affected []AffectedObject, mods []Mod, d Duration, label string, ts int64) bool {
	if len(affected) == 0 || len(mods) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil, mods, d, label, ts)
}

// RegisterScopedRuleEffectForEffect installs a record whose affected
// set is a live RULE (AffectedScope) rather than a locked set of
// objects — CR 611.2c's "modifies the rules of the game" case, #1571.
// `controller` is the effect's "you" (the controller of the resolving
// spell or ability), which the scope reads. Reports false, registering
// nothing, for an unknown scope or no mods.
//
// Caller must hold g.mu (write).
func (g *Game) RegisterScopedRuleEffectForEffect(sourceID uuid.UUID, scope AffectedScope, controller uuid.UUID, mods []Mod, d Duration, label string) bool {
	if scope == ScopeNone || !KnownAffectedScope(scope) || len(mods) == 0 {
		return false
	}
	if scope == ScopeCreaturesWithoutFlying {
		// #1650: this scope reads a layer-6 result, so only a mod the
		// post-layer fold applies may use it. A layer mod would read
		// the keyword half-way through the pass.
		for _, m := range mods {
			if !foldedAfterLayers(scope, m.Kind) {
				panic(fmt.Sprintf("game: scoped effect %q: scope %q carries only addRestrictions, got %q", label, scope, m.Kind))
			}
		}
	}
	return g.appendScopedEffectLocked(sourceID, nil, scope, controller, mods, d, label, timeNowUnixNano())
}

// appendScopedEffectLocked is the one body both registrations share.
// A non-nil `controller` overrides the one read off the source.
//
// Caller must hold g.mu (write).
func (g *Game) appendScopedEffectLocked(sourceID uuid.UUID, affected []AffectedObject, scope AffectedScope, controller uuid.UUID, mods []Mod, d Duration, label string, ts int64) bool {
	named := false
	for _, m := range mods {
		if !KnownModKind(m.Kind) {
			panic(fmt.Sprintf("game: scoped effect %q uses unknown mod kind %q", label, m.Kind))
		}
		if m.Kind == ModGrantAbilities && len(m.Grants) == 0 {
			panic(fmt.Sprintf("game: scoped effect %q grants no ability bundle", label))
		}
		if problem := blockRequirementModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q %s", label, problem))
		}
		if problem := replacementModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := nextFromSourceModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := blockRuleModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := hexproofModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := copyModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := blockCapacityModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if problem := defenderControlsModProblem(m); problem != "" {
			panic(fmt.Sprintf("game: scoped effect %q: %s", label, problem))
		}
		if r := modKinds[m.Kind].reader; r != readerLayer && r != readerCopy {
			named = true
		}
	}
	if problem := stackPinProblem(affected, mods); problem != "" {
		panic(fmt.Sprintf("game: scoped effect %q %s", label, problem))
	}
	e := ScopedEffect{
		Affected:  append([]AffectedObject(nil), affected...),
		Scope:     scope,
		Mods:      cloneMods(mods),
		Source:    ObjectRef{ID: sourceID},
		Timestamp: ts,
		Duration:  d,
		Label:     label,
	}
	if src, ok := g.LookupCardForEffect(sourceID); ok {
		e.Source.Epoch = src.ObjectEpoch
		e.SourceName = src.Name
		e.Controller = src.Controller
	}
	if controller != uuid.Nil {
		e.Controller = controller
	}
	if named {
		g.scopedEffectSeq++
		e.Seq = g.scopedEffectSeq
	}
	g.ScopedEffects = append(g.ScopedEffects, e)
	g.layerVersion.Add(1)
	return true
}

func cloneMods(mods []Mod) []Mod {
	if len(mods) == 0 {
		return nil
	}
	out := make([]Mod, len(mods))
	for i, m := range mods {
		m.Types = copyStrings(m.Types)
		m.Subtypes = copyStrings(m.Subtypes)
		m.Colors = copyStrings(m.Colors)
		m.Keywords = copyStrings(m.Keywords)
		m.Grants = copyStrings(m.Grants)
		m.Objects = append([]ObjectRef(nil), m.Objects...)
		m.Queries = clonePermanentQueries(m.Queries)
		m.Copy = clonePrintedValuesSlice(m.Copy)
		out[i] = m
	}
	return out
}

// cloneScopedEffects copies the registry into a fresh backing array.
// Entries are immutable, so their inner slices are shared, exactly as
// every other duration registry clone does.
func cloneScopedEffects(in []ScopedEffect) []ScopedEffect {
	if len(in) == 0 {
		return nil
	}
	out := make([]ScopedEffect, len(in))
	copy(out, in)
	return out
}

// ---------------------------------------------------------------
// Sweep
// ---------------------------------------------------------------

// sweepScopedEffectsLocked drops records whose duration has run out,
// through the one function that decides what a duration means. Same
// fresh-slice rule as every duration sweep, for the same reason:
// the backing array is shared with every undo snapshot.
//
// Returns whether it dropped anything; the caller bumps the layer
// version. Caller must hold g.mu.
func (g *Game) sweepScopedEffectsLocked(endOfTurn bool) bool {
	if len(g.ScopedEffects) == 0 {
		return false
	}
	kept := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if !g.durationExpiredLocked(e.Duration, endOfTurn) {
			kept = append(kept, e)
		}
	}
	if len(kept) == len(g.ScopedEffects) {
		return false
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.ScopedEffects = kept
	return true
}

// ---------------------------------------------------------------
// The layer-pass adapter
// ---------------------------------------------------------------

// scopedEffectContinuousEffectsLocked adapts every live record into
// one ContinuousEffect per mod. The closures are built here, by the
// running binary, from the record's data — so they are REBUILT on
// every pass and never persisted, which is the whole point.
//
// MEMOISED (#1558). Rebuilding a stub source, a predicate and one
// closure per mod on every recompute cost 100 records on one creature
// about 300 allocations per pass over the legacy closure statics they
// replace — before tier 3a moves Giant Growth and crew onto this path.
// The adaptation depends on nothing but the records (every closure
// reads the board it is handed at call time), so the output is reused
// for as long as the registry holds the same records in the same
// order: see scopedEffectAdapterMemo for what "the same" means and why
// it is record identity rather than the layer version.
//
// The slice returned is capped at its length, so a caller that appends
// to it — activeStaticAbilitiesLocked does — copies rather than
// writing into the memo's backing array.
//
// Caller must hold g.mu in write mode (the recompute pass does).
func (g *Game) scopedEffectContinuousEffectsLocked() []ContinuousEffect {
	if len(g.ScopedEffects) == 0 {
		g.scopedEffectMemo = scopedEffectAdapterMemo{}
		return nil
	}
	if !g.scopedEffectMemo.matches(g.ScopedEffects) {
		g.scopedEffectMemo = scopedEffectAdapterMemo{
			keys:    scopedEffectKeys(g.ScopedEffects),
			effects: adaptScopedEffects(g.ScopedEffects),
		}
	}
	out := g.scopedEffectMemo.effects
	return out[:len(out):len(out)]
}

// scopedEffectAdapterMemo is the adapter's output and the identity of
// every record it was built from, position by position.
//
// WHY IDENTITY AND NOT layerVersion. The layer version is bumped by
// every registration and sweep, but it is not monotone: an undo stores
// the snapshot's version plus one (restoreFromLocked), so the same
// number can name two different registries, and a memo keyed on it
// could hand back a record the undo took away. A record's identity
// cannot collide that way. Records are immutable once registered and
// every registration and restore allocates its Mods afresh
// (cloneMods, deepCopyScopedEffects), so &Mods[0] names one record for
// as long as anything points at it — and the memo's own key keeps it
// alive, so the address cannot be reused while the memo holds it.
// Clone shares the inner slices, which is fine: the same address is
// the same immutable record, in either game.
//
// A cloned Game does not copy the memo (cloneLocked lists its fields
// and this is not one of them), and nothing mutates a memo in place —
// a rebuild assigns new slices — so two games can never write through
// a shared one.
type scopedEffectAdapterMemo struct {
	keys    []scopedEffectMemoKey
	effects []ContinuousEffect
}

// scopedEffectMemoKey is one record's identity. The pointers do the
// work; the scalars are the fields the adapter reads that are not
// behind them, compared so a key is never trusted on the pointers
// alone.
type scopedEffectMemoKey struct {
	mods       *Mod
	affected   *AffectedObject
	nMods      int
	nAffected  int
	scope      AffectedScope
	controller uuid.UUID
	source     uuid.UUID
	sourceName string
	timestamp  int64
}

func scopedEffectKeyOf(e *ScopedEffect) scopedEffectMemoKey {
	k := scopedEffectMemoKey{
		nMods:      len(e.Mods),
		nAffected:  len(e.Affected),
		scope:      e.Scope,
		controller: e.Controller,
		source:     e.Source.ID,
		sourceName: e.SourceName,
		timestamp:  e.Timestamp,
	}
	if len(e.Mods) > 0 {
		k.mods = &e.Mods[0]
	}
	if len(e.Affected) > 0 {
		k.affected = &e.Affected[0]
	}
	return k
}

func scopedEffectKeys(records []ScopedEffect) []scopedEffectMemoKey {
	out := make([]scopedEffectMemoKey, len(records))
	for i := range records {
		out[i] = scopedEffectKeyOf(&records[i])
	}
	return out
}

// matches reports whether the memo was built from exactly these
// records, in this order.
func (m *scopedEffectAdapterMemo) matches(records []ScopedEffect) bool {
	if m.keys == nil || len(m.keys) != len(records) {
		return false
	}
	for i := range records {
		if m.keys[i] != scopedEffectKeyOf(&records[i]) {
			return false
		}
	}
	return true
}

// adaptScopedEffects is the adaptation itself: one ContinuousEffect
// per mod, every closure built from the record's data.
func adaptScopedEffects(records []ScopedEffect) []ContinuousEffect {
	n := 0
	for i := range records {
		n += len(records[i].Mods)
	}
	out := make([]ContinuousEffect, 0, n)
	for i := range records {
		e := records[i]
		// A stub source: the one reader is setController's
		// ControlSource, which wants the instance ID.
		src := &Card{InstanceID: e.Source.ID, Name: e.SourceName, Controller: e.Controller}
		applies := affectedPredicate(e.Affected)
		if e.Scope != ScopeNone {
			applies = scopePredicate(e.Scope, e.Controller)
		}
		for _, m := range e.Mods {
			spec, ok := modKinds[m.Kind]
			if !ok || spec.reader != readerLayer {
				// Unknown: unreachable, since registration and restore
				// both refuse it. Not a layer kind: the replacement
				// gather reads it (scoped_replacements.go).
				continue
			}
			if foldedAfterLayers(e.Scope, m.Kind) {
				// #1650: a live-rule restriction is applied once the
				// pass is over (foldRuleScopedRestrictionsLocked).
				continue
			}
			out = append(out, staticContinuousEffect{
				ability: StaticAbility{
					Layer:            spec.layer,
					SubLayer:         spec.subLayer,
					RemovesAbilities: spec.removes,
					AppliesTo:        applies,
					Apply:            modApply(m),
					// ModGrantAbilities (ADR 0093 PR 4): the same
					// declaration a granting static makes, so the
					// engine writes the grant in this slot with the
					// record's source as the grantor. Nil for every
					// other kind.
					GrantAbilities: m.Grants,
				},
				source:    src,
				timestamp: e.Timestamp,
			})
		}
	}
	return out
}

// affectedPredicate is the AppliesTo for a pinned set.
func affectedPredicate(set []AffectedObject) func(*Card, *Game, *Card) bool {
	return func(target *Card, _ *Game, _ *Card) bool {
		if target == nil {
			return false
		}
		for _, a := range set {
			// A stack member names a spell, and this pass is asking
			// about a permanent (ADR 0104).
			if !memberMatches(a, target) {
				continue
			}
			if a.Controller != uuid.Nil && target.Controller != a.Controller {
				continue
			}
			return true
		}
		return false
	}
}

// scopePredicate is the AppliesTo for a live-rule record (#1571). An
// unknown scope matches nothing — unreachable, since registration and
// restore both refuse one.
func scopePredicate(scope AffectedScope, controller uuid.UUID) func(*Card, *Game, *Card) bool {
	switch scope {
	case ScopeOpponentsCreatures, ScopeOpponentsAndTheirCreatures:
		return func(target *Card, _ *Game, _ *Card) bool {
			return target != nil && target.IsCreature() && target.Controller != controller
		}
	case ScopeYourPermanents:
		return func(target *Card, _ *Game, _ *Card) bool {
			return target != nil && target.Controller == controller
		}
	case ScopeYourCreatures:
		return func(target *Card, _ *Game, _ *Card) bool {
			return target != nil && target.IsCreature() && target.Controller == controller
		}
	case ScopeCreaturesWithoutFlying:
		return func(target *Card, _ *Game, _ *Card) bool {
			return target != nil && target.IsCreature() && !HasKeyword(target, "flying")
		}
	}
	// ScopeGame names no object, so it matches none.
	return func(*Card, *Game, *Card) bool { return false }
}

// foldedAfterLayers reports whether a record's mod is applied by
// foldRuleScopedRestrictionsLocked rather than in its layer bucket: a
// restriction whose affected set is a live rule (#1650).
func foldedAfterLayers(scope AffectedScope, kind ModKind) bool {
	return scope != ScopeNone && kind == ModAddRestrictions
}

// foldRuleScopedRestrictionsLocked applies every live-rule restriction
// record once the layer pass is over (#1650). Two examples are Falter's
// "creatures without flying can't block this turn" and Glaring
// Spotlight's "creatures you control … can't be blocked this turn".
//
// WHY THE SET IS LIVE. CR 611.2c locks the affected set of an effect
// from a resolving spell or ability only when the effect changes
// characteristics or control. "Can't block" and "can't be blocked"
// change neither, and CR 613 gives a restriction no layer at all
// (restrictions.go). So the effect reaches a creature that arrives,
// becomes a creature, or loses flying after the spell resolved, and
// stops reaching one that gains flying. A pinned record (`Affected`,
// the single-target RestrictUntilEOT) is a different effect, "target
// creature can't block", and stays in its layer.
//
// WHY AFTER THE PASS. The scope reads the object's FINISHED
// characteristics. "Without flying" is a layer-6 result, and a layer-6
// bucket sorted by timestamp would ask the question before a
// later-timestamped flying grant had applied. Restriction bits are only
// ever OR'd in and no layer reads them, so applying them after layer 7
// changes nothing else. The records are read afresh on every pass and
// hold no closure, so undo and a restore point carry them as data.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) foldRuleScopedRestrictionsLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope == ScopeNone {
			continue
		}
		var bits Restriction
		for _, m := range e.Mods {
			if foldedAfterLayers(e.Scope, m.Kind) {
				bits |= m.Restrictions
			}
		}
		if bits == 0 {
			continue
		}
		applies := scopePredicate(e.Scope, e.Controller)
		for j := range g.Battlefield.Cards {
			target := &g.Battlefield.Cards[j]
			if target.effective == nil || !applies(target, g, nil) {
				continue
			}
			target.effective.Restrictions |= bits
		}
	}
}

// modApply is the interpreter: what each kind does to a
// Characteristic. The only code in the engine that gives a ModKind its
// meaning.
func modApply(m Mod) func(*Characteristic, *Card, *Game, *Card) {
	switch m.Kind {
	case ModSetController:
		player := m.Player
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			ch.Controller = player
			// #930: whoever wrote Controller last in this bucket won
			// CR 613.7, and the control-change event names it.
			if src != nil {
				ch.ControlSource = src.InstanceID
			}
		}
	case ModAddTypes:
		types := m.Types
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, t := range types {
				if !typeListHas(ch.Types, t) {
					ch.Types = append(ch.Types, t)
				}
			}
		}
	case ModRemoveTypes:
		types := m.Types
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			kept := make([]string, 0, len(ch.Types))
			for _, t := range ch.Types {
				if !typeListHas(types, t) {
					kept = append(kept, t)
				}
			}
			ch.Types = kept
		}
	case ModAddSubtypes:
		subtypes := m.Subtypes
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, s := range subtypes {
				if !typeListHas(ch.Subtypes, s) {
					ch.Subtypes = append(ch.Subtypes, s)
				}
			}
		}
	case ModAllCreatureTypes:
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.AllCreatureTypes = true
		}
	case ModSetColors:
		colors := m.Colors
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Colors = copyStrings(colors)
		}
	case ModAddKeywords, ModLoseAllAbilities:
		// loseAllAbilities: the engine empties the list for a
		// RemovesAbilities effect before Apply runs (ADR 0046), so
		// Apply only appends what the same effect grants back.
		keywords := m.Keywords
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			for _, kw := range keywords {
				ch.Abilities = AppendKeywordAbility(ch.Abilities, kw)
			}
		}
	case ModRemoveKeywords:
		keywords := m.Keywords
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			kept := make([]string, 0, len(ch.Abilities))
			for _, a := range ch.Abilities {
				drop := false
				for _, kw := range keywords {
					if strings.EqualFold(a, kw) {
						drop = true
						break
					}
				}
				if !drop {
					kept = append(kept, a)
				}
			}
			ch.Abilities = kept
		}
	case ModAddRestrictions:
		bits := m.Restrictions
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Restrictions |= bits
		}
	case ModCantHaveKeywords:
		return cantHaveApply(m.Keywords)
	case ModSetBasePower:
		n := m.Power
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) { ch.Power = n }
	case ModSetBaseToughness:
		n := m.Toughness
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) { ch.Toughness = n }
	case ModAddAttackRequirement:
		otherThan := m.Player
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			r := AttackRequirement{OtherThan: otherThan}
			if src != nil {
				r.Source, r.SourceName = src.InstanceID, src.Name
			}
			ch.AttackRequirements = append(ch.AttackRequirements, r)
		}
	case ModCantAttackUnlessDefenderControls:
		qs := clonePermanentQueries(m.Queries)
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			r := AttackTargetRestriction{DefenderMustControl: qs}
			if src != nil {
				r.Source, r.SourceName = src.InstanceID, src.Name
			}
			ch.AttackTargetRestrictions = append(ch.AttackTargetRestrictions, r)
		}
	case ModAddBlockRequirement:
		kind := BlockRequirementKind(m.Text)
		var attacker ObjectRef
		if len(m.Objects) > 0 {
			attacker = m.Objects[0]
		}
		return func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			r := BlockRequirement{Kind: kind, Attacker: attacker}
			if src != nil {
				r.Source, r.SourceName = src.InstanceID, src.Name
			}
			ch.BlockRequirements = append(ch.BlockRequirements, r)
		}
	case ModAddBlockCapacity:
		n := m.Amount
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.AdditionalBlocks += n
		}
	case ModBlockAnyNumber:
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.BlocksAnyNumber = true
		}
	case ModModifyPT:
		p, t := m.Power, m.Toughness
		return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			ch.Power += p
			ch.Toughness += t
		}
	}
	return nil
}
