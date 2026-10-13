package game

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// trigger_order_why.go — the sentence under "Order your triggers"
// (#3015, ADR 0018's #3015 amendment).
//
// triggerOrderAskLocked already knows WHY it asks: the first conflict
// the independence check found, counted by TriggerOrderReason. This
// file carries the detail the sentence needs beyond that label (the
// card, the event, the player) and turns it into one line in the
// player's terms, stored on the prompt (PendingChoice.TriggerOrderWhy)
// and sent to the client as trigger_order_why.
//
// The text is for a person. Nothing reads it back: not the enumerator,
// not a bot, not a rule. It never changes whether the seat is asked.

// triggerOrderWhy is a TriggerOrderReason with the facts its sentence
// names. The zero value (no reason) means "no conflict found".
type triggerOrderWhy struct {
	reason TriggerOrderReason
	// card is the name of the card the sentence is about ("" when the
	// cause has no card: a player static, a delayed trigger).
	card string
	// event is the event kind a reactor watches (named as a family,
	// never a direction), and life the way the item that changes a
	// characteristic moves life totals.
	event EventKind
	life  lifeDirection
	// player is the player a sentence about leaving the game names.
	player uuid.UUID
	// both is set when two items both write the object an object tie
	// names; identical when the batch is copies of one trigger; token
	// when the reactor is a token the batch creates.
	both, identical, token bool
}

// possessive is "Riders'" for a name ending in s and "Marwyn's" for
// any other.
func possessive(name string) string {
	if strings.HasSuffix(name, "s") {
		return name + "'"
	}
	return name + "'s"
}

// lifeWords names what the batch does to life totals.
func lifeWords(d lifeDirection) string {
	switch d {
	case lifeUp:
		return "life gain"
	case lifeDown:
		return "life loss"
	}
	return "life changes"
}

// eventFamily names what a watcher of the event kind watches, as the
// family of event and never a direction in it. A card's declared watch
// is an event KIND (change_life), which does not say whether it cares
// about a gain or a loss, so a sentence that said "reacts to life gain"
// could be false for a card that watches loss. The family is true of
// every watcher of the kind.
func eventFamily(kind EventKind) string {
	switch kind {
	case EventChangeLife:
		return "life totals"
	case EventDealDamage:
		return "damage"
	case EventCounterPlaced:
		return "counters being put on permanents"
	case EventPlayerCounterPlaced:
		return "players getting counters"
	case EventZoneMove:
		return "cards changing zones"
	case EventLTB:
		return "permanents leaving the battlefield"
	case EventUnattach:
		return "permanents coming unattached"
	case EventDrawCard:
		return "cards being drawn"
	case EventMill:
		return "cards being milled"
	case EventScry:
		return "scrying"
	case EventSurveil:
		return "surveilling"
	case EventTokenCreated:
		return "tokens being created"
	case EventETB:
		return "permanents entering the battlefield"
	case EventMonarchChanged:
		return "the monarch changing"
	}
	return "what these triggers do"
}

// triggerOrderSentenceLocked is the line under the heading. It is empty
// for a batch that was not asked about; every reason has a sentence,
// and an unknown one falls back to a generic line rather than none.
//
// Caller must hold g.mu.
func (g *Game) triggerOrderSentenceLocked(w triggerOrderWhy, chooser uuid.UUID) string {
	switch w.reason {
	case "":
		return ""
	case TriggerOrderReasonAlways:
		return "You've set your triggers to ask every time."
	case TriggerOrderReasonUndescribed:
		if w.card != "" {
			return possessive(w.card) + " trigger is handled by the players, so the game can't tell whether the order matters."
		}
		return "The game can't tell whether the order of these triggers matters."
	case TriggerOrderReasonBoardOpaque:
		if w.card != "" {
			return w.card + " could change how these triggers resolve."
		}
		return "Something in play could change how these triggers resolve."
	case TriggerOrderReasonReactor:
		what := eventFamily(w.event)
		switch {
		case w.token:
			return "A token these triggers create watches " + what + "."
		case w.card != "":
			return w.card + " watches " + what + "."
		}
		return "Something in play watches " + what + "."
	case TriggerOrderReasonRipple:
		if w.card != "" {
			return "Something else in play is keeping track of " + w.card + "."
		}
		return "Something else in play is keeping track of what these triggers change."
	case TriggerOrderReasonElimination:
		who := "A player"
		if w.player == chooser {
			who = "You"
		} else if p := g.playerByIDLocked(w.player); p != nil && p.Name != "" {
			who = p.Name
		}
		return who + " could leave the game partway through these triggers, which changes what the rest do."
	case TriggerOrderReasonObjectTie:
		name := w.card
		if name == "" {
			name = "the same permanent"
		}
		if w.both {
			return "These triggers both change " + name + "."
		}
		return "One of these triggers changes " + name + ", which another looks at."
	case TriggerOrderReasonRules:
		if w.card != "" {
			return possessive(w.card) + " trigger changes the rules the other triggers resolve under."
		}
		return "One of these triggers changes the rules the others resolve under."
	case TriggerOrderReasonLibrary:
		return "These triggers both work on the same library or graveyard."
	case TriggerOrderReasonMonarch:
		return "These triggers both make you the monarch."
	case TriggerOrderReasonCharacteristic:
		what := "What one of these triggers does"
		if w.life != 0 {
			what = "The " + lifeWords(w.life)
			if w.identical {
				what = "The first " + lifeWords(w.life)
			} else {
				what += " from one trigger"
			}
		}
		if w.card == "" {
			return what + " can change a bonus that another trigger counts."
		}
		return what + " can change " + possessive(w.card) + " bonus, which another trigger counts."
	}
	return "The order of these triggers can change how they resolve."
}

// whyCardNameLocked is the name of any card the game knows, or "".
//
// Caller must hold g.mu.
func (g *Game) whyCardNameLocked(id uuid.UUID) string {
	if id == (uuid.UUID{}) {
		return ""
	}
	if c, ok := g.LookupCardForEffect(id); ok {
		return whyCardName(c)
	}
	return ""
}

// sortedIDs lists a set of IDs in a fixed order, so a sentence that
// names "the first" of a set names the same one every time.
func sortedIDs(set map[uuid.UUID]bool) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}

// whyCardName is how a sentence names a card: a legend by the part of
// its name before the comma ("Marwyn" for "Marwyn, the Nurturer"), and
// a face-down permanent by nothing but what the table can see, because
// a sentence the whole table reads must not give its name away
// (CR 708.2).
func whyCardName(c Card) string {
	if c.FaceDownIsPermanent() {
		return "a face-down permanent"
	}
	if i := strings.Index(c.Name, ", "); i > 0 && c.HasSupertype("legendary") {
		return c.Name[:i]
	}
	return c.Name
}
