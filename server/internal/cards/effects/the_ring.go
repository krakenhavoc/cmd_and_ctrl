package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring.go is the catalog half of CR 701.54, "the Ring tempts you"
// (ADR 0114, #2076). The engine half — the tempt, the count, the
// Ring-bearer designation and the ring_bearer prompt — is
// game/ring.go.
//
// # The emblem
//
// The Ring is an ordinary ADR 0064 emblem whose key is fixed
// (game.RingEmblemKey), because no card creates it: CR 701.54c makes
// it a rules object every player gets at their first temptation. It is
// filed once, below, through registerRulesEmblem — into `defs` and not
// into `registry`, so the card census never counts it.
//
// Its abilities are written in the vocabulary every emblem uses, each
// behind game.RingTempted(n), the gate that reads the emblem's own
// count. CR 701.54c gives it four, gained in order and kept for the
// rest of the game (2023-06-16 ruling):
//
//	1  "Your Ring-bearer is legendary and can't be blocked by
//	    creatures with greater power."              (ADR 0114 PR 2)
//	2  "Whenever your Ring-bearer attacks, draw a card, then discard
//	    a card."                                    (PR 3)
//	3  "Whenever your Ring-bearer becomes blocked by a creature, the
//	    blocking creature's controller sacrifices it at end of
//	    combat."                                    (PR 3)
//	4  "Whenever your Ring-bearer deals combat damage to a player,
//	    each opponent loses 3 life."                (PR 3)
//
// Lines 2 to 4 are triggered abilities of the emblem (CR 113.7: the
// emblem is their source), so they are harvested, ordered and put on
// the stack like any emblem trigger, and controlled by the emblem's
// owner. Each one reads "your Ring-bearer" as the event happens (CR
// 701.54e), with the trigger harvested at the event, so a Ring-bearer
// that dies to the damage it is dealt in the same step still made line
// 4 trigger: the damage and its event come before state-based actions.
//
// # The card side
//
// TheRingTemptsYou is the primitive, and WheneverTheRingTemptsYou,
// WheneverYouChooseARingBearer and IfYouChoseAnotherRingBearer the
// trigger shapes the Ring cards print. See docs/adding-cards.md, "The
// Ring tempts you".

// The Ring's four lines (CR 701.54c), each gained at the temptation its
// number names.
const (
	theRingLine1 = "Your Ring-bearer is legendary and can't be blocked by creatures with greater power."
	theRingLine2 = "Whenever your Ring-bearer attacks, draw a card, then discard a card."
	theRingLine3 = "Whenever your Ring-bearer becomes blocked by a creature, the blocking creature's controller sacrifices it at end of combat."
	theRingLine4 = "Whenever your Ring-bearer deals combat damage to a player, each opponent loses 3 life."
)

// The stack labels of the Ring's three triggered lines. Each is also
// its row's catalog name (TriggeredAbility.Key), which a waiting
// trigger's restore looks the row up by — an on-disk identity, so
// never reworded.
const (
	theRingLootLabel      = "The Ring — your Ring-bearer attacked: draw a card, then discard a card"
	theRingSacrificeLabel = "The Ring — your Ring-bearer became blocked: its blocker is sacrificed at end of combat"
	theRingLoseLifeLabel  = "The Ring — your Ring-bearer dealt combat damage to a player: each opponent loses 3 life"
)

// theRingSacrificeBlockerBody is the delayed trigger line 3 creates:
// at the beginning of the end of combat step (CR 511.2, 603.7), the
// blocking creature's controller sacrifices it. A registered body (ADR
// 0041 phase 3), so a table with one waiting is a restore point.
var theRingSacrificeBlockerBody game.BodyRef

func init() {
	theRingSacrificeBlockerBody = game.DelayedBody("the-ring/sacrifice-blocker-at-end-of-combat", theRingSacrificeBlocker)

	registerRulesEmblem(game.RingEmblemKey, EmblemSpec{
		Label: "The Ring",
		Text:  strings.Join([]string{theRingLine1, theRingLine2, theRingLine3, theRingLine4}, "\n"),
		Lines: []game.EmblemLine{
			{Text: theRingLine1, At: 1},
			{Text: theRingLine2, At: 2},
			{Text: theRingLine3, At: 3},
			{Text: theRingLine4, At: 4},
		},
		Static: []game.StaticAbility{
			yourRingBearerIsLegendary(),
		},
		Triggered: []game.TriggeredAbility{
			theRingLine(2, On(game.EventAttack, yourRingBearerAttacked, theRingLootLabel, theRingLoot)),
			theRingLine(3, theRingBlockerTrigger()),
			theRingLine(4, On(game.EventDealDamage, yourRingBearerDealtCombatDamageToAPlayer, theRingLoseLifeLabel, theRingEachOpponentLosesThree)),
		},
		// The evasion is a CR 509.1b restriction, checked as blockers
		// are declared (ADR 0114 §5). The emblem exists only once the
		// Ring has tempted its owner, so this line needs no gate.
		BlockRules: []game.BlockRule{
			CantBeBlockedByComparing(YourRingBearer(), GreaterPowerThanAttacker(), "creatures with greater power"),
		},
	})
}

// yourRingBearerIsLegendary is "Your Ring-bearer is legendary": a layer
// 4 effect (CR 613.1d) adding the Legendary supertype to the emblem
// owner's Ring-bearer, with the emblem's timestamp (CR 613.7a). The
// legend rule (CR 704.5j) and every "legendary creature" reader see it
// through the effective supertypes.
func yourRingBearerIsLegendary() game.StaticAbility {
	return game.StaticAbility{
		Layer:      game.Layer4Type,
		ActiveWhen: game.RingTempted(1),
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source != nil && game.IsRingBearerOf(*target, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, s := range c.Supertypes {
				if strings.EqualFold(s, "Legendary") {
					return
				}
			}
			c.Supertypes = append(c.Supertypes, "Legendary")
		},
	}
}

// theRingLine gates one of the Ring's triggered lines on the count it
// is gained at: the ability does not exist until the Ring has tempted
// its owner n times (CR 701.54c), so it is not harvested before then.
func theRingLine(n int, t game.TriggeredAbility) game.TriggeredAbility {
	t.ActiveWhen = game.RingTempted(n)
	return t
}

// isYourRingBearer reports whether `id` is the Ring-bearer of the
// emblem's owner right now (CR 701.54e): on the battlefield, under
// their control, with the designation.
func isYourRingBearer(g *game.Game, emblem *game.Card, id uuid.UUID) bool {
	if emblem == nil || id == uuid.Nil || !onBattlefield(g, id) {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && game.IsRingBearerOf(c, emblem.Controller)
}

// yourRingBearerAttacked is line 2's condition: "whenever your
// Ring-bearer attacks" — EventAttack names each attacker as the
// declaration locks in (CR 508.1).
func yourRingBearerAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventAttack && isYourRingBearer(g, source, ev.CardID)
}

// theRingLoot is line 2's effect: draw a card, then discard a card. The
// discard is asked from the hand after the draw.
func theRingLoot(g *game.Game, item *game.StackItem) error {
	return lootOne(g, item, 1)
}

// yourRingBearerBecameBlocked is line 3's condition: "whenever your
// Ring-bearer becomes blocked by a creature". EventBlock is one event
// per blocker–attacker pair (CR 509.3d), CardID the blocker and Target
// the attacker, so two blockers are two triggers. The Ring-bearer
// blocking is not it becoming blocked, and does not match.
func yourRingBearerBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventBlock && ev.CardID != uuid.Nil && isYourRingBearer(g, source, ev.Target)
}

// theRingBlockerTrigger is line 3. Its Build records the blocker as the
// object it was when it blocked (ADR 0041 P9's fill-in Build, Ares's
// shape): "the blocking creature" is that object, so one that leaves
// the battlefield and comes back is a new object (CR 400.7) and is not
// sacrificed.
func theRingBlockerTrigger() game.TriggeredAbility {
	t := On(game.EventBlock, yourRingBearerBecameBlocked, theRingSacrificeLabel, theRingScheduleBlockerSacrifice)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, theRingSacrificeLabel)
		epoch := -1
		if c, ok := g.LookupCardForEffect(ev.CardID); ok {
			epoch = c.ObjectEpoch
		}
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: epoch}
		return item
	}
	return t
}

// theRingScheduleBlockerSacrifice is line 3's effect: a delayed
// trigger at the beginning of this combat's end of combat step (CR
// 511.2, 603.7) that has the blocker's controller sacrifice it.
func theRingScheduleBlockerSacrifice(g *game.Game, item *game.StackItem) error {
	blocker := item.Params.Object
	if blocker.ID == uuid.Nil {
		return nil
	}
	label := "The Ring — sacrifice the creature that blocked your Ring-bearer"
	if c, ok := g.LookupCardForEffect(blocker.ID); ok {
		label = "The Ring — sacrifice " + c.Name + ", which blocked your Ring-bearer"
	}
	return ScheduleDelayedTrigger{
		At:     game.StepEndCombat,
		Label:  label,
		Cards:  []uuid.UUID{blocker.ID},
		Body:   theRingSacrificeBlockerBody,
		Params: game.EffectParams{Object: blocker},
	}.Apply(NewContext(g, item))
}

// theRingSacrificeBlocker is the delayed trigger's body: "the blocking
// creature's controller sacrifices it". Nothing happens if it is no
// longer the object that blocked (CR 400.7) or no longer on the
// battlefield. Its controller is read now, as the trigger resolves: a
// blocker that changed hands is sacrificed by whoever controls it.
func theRingSacrificeBlocker(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !onBattlefield(g, p.Object.ID) {
		return nil
	}
	return SacrificePermanent{Target: p.Object.ID}.Apply(NewContext(g, item))
}

// yourRingBearerDealtCombatDamageToAPlayer is line 4's condition:
// "whenever your Ring-bearer deals combat damage to a player". The
// harvest reads it as the damage event is emitted, before state-based
// actions, so a Ring-bearer that the same combat damage kills still
// counts.
func yourRingBearerDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && combatDamageToPlayerBy(ev, source.Controller, g) && isYourRingBearer(g, source, ev.Source)
}

// theRingEachOpponentLosesThree is line 4's effect: each opponent of
// the emblem's owner loses 3 life — every opponent still in the game,
// not only the player who was dealt the damage.
func theRingEachOpponentLosesThree(g *game.Game, item *game.StackItem) error {
	return eachOpponentLosesLife(g, item, 3)
}

// registerRulesEmblem files an emblem no card creates, under a fixed
// key — a rules object such as the Ring (CR 701.54c). The emblem
// Register files beside a card's Spec, without the card: the same
// guards, the same def builder, and the same `fileDef`, which stamps
// its triggered rows with their catalog identity
// (game.IdentifyCatalogRows) so a waiting trigger restores like a
// card's.
//
// Into `defs` and never `registry`: an emblem is not a card (CR
// 114.5) and the census must not count it.
func registerRulesEmblem(key string, e EmblemSpec) {
	if !strings.HasPrefix(key, game.EmblemKeyPrefix) {
		panic(fmt.Sprintf("effects.registerRulesEmblem: key %q is not an emblem key", key))
	}
	if _, ok := defs[key]; ok {
		panic(fmt.Sprintf("effects.registerRulesEmblem: duplicate emblem key %q", key))
	}
	checkStaticGrants(e.Label, "emblem static", e.Static)
	checkTriggerZones(e.Label, "emblem trigger", e.Triggered)
	checkEmblemSpec(e.Label, &e)
	fileDef(key, buildEmblemDef(e))
}

// TheRingTemptsYou is the keyword action "the Ring tempts you"
// (CR 701.54a): the emblem if the player has none, its count up by
// one, a Ring-bearer chosen from the creatures they control (asked
// only when there are two or more), and the event every "whenever the
// Ring tempts you" ability watches. See game.RingTemptsForEffect.
type TheRingTemptsYou struct {
	// Player is who the Ring tempts. Zero means the effect's
	// controller, which is what every printed Ring card says.
	Player uuid.UUID

	// Then is the rest of the sentence, run once the tempt is complete
	// and handed the creature chosen as the Ring-bearer — uuid.Nil
	// when the player controlled no creature (the tempt still happened,
	// CR 701.54d). Ringsight's search and One Ring to Rule Them All's
	// mill come after the tempt; nil for a card whose sentence ends
	// there.
	Then func(ctx *Context, ringBearer uuid.UUID) error
}

func (r TheRingTemptsYou) Apply(ctx *Context) error {
	player := r.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	item := ctx.Item
	then := r.Then
	var tail func(g *game.Game, ringBearer uuid.UUID) error
	if then != nil {
		tail = func(g *game.Game, ringBearer uuid.UUID) error {
			return then(NewContext(g, item), ringBearer)
		}
	}
	return ctx.Game.RingTemptsForEffect(player, ctx.Source(), tail)
}

// RingTempted is "whenever the Ring tempts you" as a trigger condition:
// the Ring tempted this ability's controller (CR 701.54d), whether or
// not they could choose a creature.
func RingTempted(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && ev.Kind == game.EventRingTempted && ev.Actor == source.Controller
}

// WheneverTheRingTemptsYou — "Whenever the Ring tempts you, …"
// (Nazgûl, Sauron, the Dark Lord). It triggers on every temptation,
// including one in which no creature could be chosen (CR 701.54d).
func WheneverTheRingTemptsYou(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventRingTempted, RingTempted, label, effect)
}

// WheneverYouChooseARingBearer — "Whenever you choose a creature as
// your Ring-bearer, …" (Call of the Ring). The same event, only when a
// creature was chosen: "If the Ring tempts you but you can't choose a
// creature, Call of the Ring's last ability won't trigger", while
// re-choosing the creature that already is your Ring-bearer "still
// counts as choosing that creature" (2023-06-16 rulings).
func WheneverYouChooseARingBearer(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventRingTempted, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return RingTempted(ev, source, ch, g) && ev.CardID != uuid.Nil
	}, label, effect)
}

// IfYouChoseAnotherRingBearer narrows a "whenever the Ring tempts you"
// trigger to its intervening "if" (CR 603.4): "Whenever the Ring
// tempts you, if you chose a creature other than ~ as your
// Ring-bearer, …" (Aragorn, Company Leader; Faramir, Field Commander;
// Galadriel of Lothlórien; Gandalf, Friend of the Shire). The condition
// is a fact about the temptation that triggered it, carried on the
// event, so it reads the same as the trigger resolves.
func IfYouChoseAnotherRingBearer(t game.TriggeredAbility) game.TriggeredAbility {
	inner := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		if inner != nil && !inner(ev, source, ch, g) {
			return false
		}
		return source != nil && ev.CardID != uuid.Nil && ev.CardID != source.InstanceID
	}
	return t
}

// legendaryCreaturesYouControl is every legendary creature `player`
// controls, read with the layers caught up — after a tempt, the new
// Ring-bearer is legendary through the Ring's layer 4 effect
// (Ringsight's "a legendary creature you control").
func legendaryCreaturesYouControl(g *game.Game, player uuid.UUID) []game.Card {
	g.RecomputeLayersIfStaleLocked()
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.IsLegendary() {
			out = append(out, c)
		}
	}
	return out
}
