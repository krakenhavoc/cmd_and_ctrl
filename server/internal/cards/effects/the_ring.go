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
// count. This file has the first line only (ADR 0114 PR 2):
//
//	"Your Ring-bearer is legendary and can't be blocked by creatures
//	 with greater power."
//
// The other three lines — the loot at two temptations, the end-of-
// combat sacrifice at three and the life loss at four — land with
// ADR 0114 PR 3, together with the first cards that tempt. No card
// tempts until all four are here, because a card that tempted while
// the emblem had only its first line would be weaker than printed for
// everyone who reached the second.
//
// # The card side
//
// TheRingTemptsYou is the primitive, and WheneverTheRingTemptsYou,
// WheneverYouChooseARingBearer and IfYouChoseAnotherRingBearer the
// trigger shapes the Ring cards print. See docs/adding-cards.md, "The
// Ring tempts you".

// theRingLine1 is the Ring's first line (CR 701.54c), gained at the
// first temptation.
const theRingLine1 = "Your Ring-bearer is legendary and can't be blocked by creatures with greater power."

func init() {
	registerRulesEmblem(game.RingEmblemKey, EmblemSpec{
		Label: "The Ring",
		Text:  theRingLine1,
		Lines: []game.EmblemLine{{Text: theRingLine1, At: 1}},
		Static: []game.StaticAbility{
			yourRingBearerIsLegendary(),
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
