package game

// melee.go — CR 702.121, melee as a keyword the engine derives from an
// object's ability list rather than from its catalog entry (#4722).
// Exalted (exalted.go) and decayed (decayed.go) are the models.
//
//	702.121a Melee is a triggered ability. "Melee" means "Whenever this
//	         creature attacks, it gets +1/+1 until end of turn for each
//	         opponent you attacked with a creature this combat."
//	702.121b If a creature has multiple instances of melee, each
//	         triggers separately.
//
// WHY A TOKEN. A token in Characteristic.Abilities reaches a printed
// melee (the deck importer stamps it from Scryfall's keywords array
// with no catalog entry), a layer-6 grant ("Other creatures you control
// have melee", Titania, Proud Pummeler) and a token creature through
// one walk, keywordTriggersFor. A creature that loses its abilities
// loses the trigger with them.
//
// MULTIPLE INSTANCES. CR 702.121b: each instance triggers separately.
// Melee is CUMULATIVE (KeywordIsCumulative), so AppendKeywordAbility
// keeps a granted melee beside a printed one, and each is its own
// trigger and its own +1/+1 per opponent.
//
// THE TRIGGER. "Whenever this creature attacks" watches EventAttack for
// the source itself, which is emitted only for a declared attacker (CR
// 508.3a), so a creature put onto the battlefield attacking never
// triggers it.
//
// THE COUNT. The number of opponents is read as the ability RESOLVES,
// not as it triggers: melee has no "as it triggers" wording, so the
// count is the opponents attacked "this combat" at that moment. It is
// read from the turn's attack records (TurnTally.Attacks) for the
// current combat phase, so:
//
//   - it counts every creature the controller attacked with, not only
//     the melee creature ("with a creature this combat");
//   - a creature that attacked and has since left combat or the
//     battlefield still counts, because the records are history;
//   - an additional combat phase is a combat of its own (PhaseID);
//   - a creature put onto the battlefield attacking was never declared
//     and is no record;
//   - an attack on a planeswalker or a battle is an attack on that
//     permanent, not on its controller or protector, so it counts no
//     player (CR 506.3; the same reading AttackedPlayersThisTurn
//     makes). Each opponent is counted once, however many creatures
//     attacked them.
//
// Only the active player declares attackers (CR 508.1), so a melee
// trigger whose controller is not the active player when it resolves
// (control changed with the trigger on the stack) attacked no one.
//
// THE PUMP. The creature is fixed as the ability triggers, by instance
// ID and object epoch (Params.Object): one that left the battlefield
// and came back is a new object and gets nothing (CR 400.7). The
// +N/+N is a layer-7c scoped effect until end of turn.
//
// ORDERING (CR 603.3b). Melee triggers commute: each reads the
// attack history, which nothing on the stack changes, and writes only
// a +N/+N to its own pinned creature, so a batch of nothing but melee
// triggers needs no ordering prompt.

import "github.com/google/uuid"

// KeywordMelee is CR 702.121's canonical token.
const KeywordMelee = "melee"

// meleeLabel is the stack label every melee trigger carries, and the
// label of its +N/+N.
const meleeLabel = "Melee — +1/+1 until end of turn for each opponent attacked"

// meleePumpKey is the body every melee trigger names (ADR 0041 P9): an
// on-disk identity, never renamed or reused. Params.Object is the
// attacking creature the ability is on.
const meleePumpKey = "melee/pump"

// meleePumpBody is the registered body's reference, built from the key
// and registered in init, as exalted's is.
var meleePumpBody = BodyRef{key: meleePumpKey}

func init() {
	DelayedBody(meleePumpKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.resolveMeleeLocked(item, p.Object)
	})
}

// meleeTrigger is the one TriggeredAbility a single instance of melee
// is. A package-level value, never mutated: keywordTriggersFor hands
// out copies of it.
var meleeTrigger = TriggeredAbility{
	Keyword: KeywordMelee,
	Watches: []EventKind{EventAttack},
	AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
		return ev.CardID != uuid.Nil && ev.CardID == source.InstanceID
	},
	Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
		self := ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}
		item := NewKeyedTriggeredItem(source, meleeLabel, meleePumpBody, EffectParams{Object: self})
		// Melee instances commute with each other: see the file comment.
		item.Commutes = true
		return item
	},
}

// OpponentsAttackedThisCombat is how many opponents of `you` they
// attacked with a creature in the current combat phase: CR 702.121a's
// count. Each opponent counts once, an attack on a planeswalker or a
// battle counts no player, and a creature that attacked and has since
// left combat still counts (the attack records are history). Zero when
// `you` is not the active player, who alone declares attackers (CR
// 508.1).
//
// Caller must hold g.mu.
func (g *Game) OpponentsAttackedThisCombat(you uuid.UUID) int {
	if you == uuid.Nil || g.activePlayerIDLocked() != you {
		return 0
	}
	seen := map[uuid.UUID]bool{}
	for _, a := range g.TurnTally.Attacks {
		if a.PhaseID != g.Turn.PhaseID || a.Defender == you || seen[a.Defender] {
			continue
		}
		if g.playerByIDLocked(a.Defender) == nil {
			continue
		}
		seen[a.Defender] = true
	}
	return len(seen)
}

// resolveMeleeLocked is "it gets +1/+1 until end of turn for each
// opponent you attacked with a creature this combat" for the creature
// the trigger pinned, counting as the ability resolves. A creature no
// longer on the battlefield as the same object gets nothing (CR
// 400.7), and no opponent attacked is a +0/+0 that is not registered.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) resolveMeleeLocked(item *StackItem, self ObjectRef) error {
	if item == nil || self.ID == uuid.Nil {
		return nil
	}
	c := findBattlefieldCard(g, self.ID)
	if c == nil || c.ObjectEpoch != self.Epoch {
		return nil
	}
	n := g.OpponentsAttackedThisCombat(item.Controller)
	if n == 0 {
		return nil
	}
	g.RegisterScopedEffectForEffect(item.SourceCardID,
		[]AffectedObject{PinObject(c.InstanceID, c.EnteredBattlefieldAt)},
		[]Mod{ModifyPTMod(n, n)}, g.UntilEndOfTurnDuration(), meleeLabel)
	return nil
}

// MeleeCount reports how many instances of melee the card has — a read
// for tests, the bot and the view, through the same walk the harvester
// uses.
func MeleeCount(c *Card) int {
	return countAbilityTokens(c, KeywordMelee)
}
