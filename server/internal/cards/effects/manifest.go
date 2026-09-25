package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest.go — #1194 / ADR 0082: manifest (CR 701.34), cloak
// (CR 701.58), and the ward {2} disguise and cloak give the object
// they make (CR 702.168a, CR 701.58a).
//
// Manifest is the keyword action side of the same seam morph is the
// cast side of, and the engine has carried it since ADR 0069:
// `ManifestForEffect` puts the top card of a library onto the
// battlefield face down as a 2/2, through the CR 614 entry pipeline,
// with the controller as its only knower (CR 708.5). What was missing
// was a card able to call it, and the answer to "may this be turned
// face up" — which is `game.TurnFaceUpOffer`'s manifested row: a
// CREATURE card for its mana cost (CR 701.40b), and nothing at all for
// anything else, which stays face down for as long as it is on the
// battlefield.

// Manifest is CR 701.34a: "manifest the top card of your library" —
// put it onto the battlefield face down as a 2/2 creature under
// `Player`'s control, or the resolving item's controller when that is
// zero, as every other primitive in this package spells "you".
//
// `N` is how many cards, taken one at a time from the top, which is
// what "manifest the top two cards of your library" means: CR 701.34b
// makes each its own manifest, so each one re-reads the top and each
// one opens its own CR 614 entry window. Zero or fewer manifests
// nothing.
//
// An empty library manifests nothing, and that is not an error and not
// a draw — the posture ManifestForEffect and ExileTopFaceDownForEffect
// both take.
//
// Nothing is REVEALED on the way. The controller becomes the card's
// only knower (CR 708.5), which is the whole difference between
// manifesting the top card of your library and putting it onto the
// battlefield.
type Manifest struct {
	Player uuid.UUID
	N      int
}

func (m Manifest) Apply(ctx *Context) error {
	if m.N <= 0 {
		return nil
	}
	player := m.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	for i := 0; i < m.N; i++ {
		if _, err := ctx.Game.ManifestForEffect(player); err != nil {
			return fmt.Errorf("manifest: %w", err)
		}
	}
	return nil
}

// faceDownWardCost is the ward disguise and cloak give the face-down
// object — {2} on every card that prints either (CR 702.168a,
// CR 701.58a). Like morph's {3}, it is the KEYWORD's number and not
// any card's, so it is written here once.
const (
	faceDownWardCost = "{2}"
	// faceDownWardLabel is the stack line the table reads when the
	// ward fires. It names the WARD and not the card, because the
	// card is face down and naming it would be the leak the whole
	// face-down model exists to stop.
	faceDownWardLabel = "Ward " + faceDownWardCost
)

// init wires the engine's face-down ward hook.
//
// The boundary is the whole reason this is a hook. Ward is a
// CR 702.21a TRIGGERED ability rather than a keyword token —
// `canonicalKeywords` is a closed set with nowhere to put the cost,
// which ADR 0069 decision 3 settled — so the ward a disguised or
// cloaked object has is a `game.TriggeredAbility` built by
// `effects.Ward`, and `game` cannot import this package. One function
// variable, set once, read by `TriggersForCard` before the catalog
// read that CR 708.2a silences.
//
// It is the OBJECT's ability, not the card's: it comes from being
// disguised or cloaked, so a disguised Sheoldred has ward {2} and
// nothing else, and the moment the permanent is turned face up the
// ward goes and the real card's abilities answer again.
func init() {
	ward := Ward(WardMana(faceDownWardCost), faceDownWardLabel)
	// The Key is the stack label, and Ward() leaves it to the card
	// files that stamp their own. This one has no card file to stamp
	// it — the ability belongs to a STATE — so it is named here, where
	// the state is.
	ward.Key = faceDownWardLabel
	// ADR 0041 P9 (#1497, tier 4): unlike every other Ward/WardGranted
	// card, this state has no catalog row at all, so its item cannot
	// wait on a later slice's AbilityRef — it is keyed directly here.
	// Watches / AppliesTo are unchanged (Ward's own, never serialised,
	// re-registered by this same init on every boot); only Build
	// changes, to a body that reads the stack item ID and the payer
	// off item.Trigger.Event rather than off captured closure
	// variables — stampTriggerContext has already put the same event
	// there by the time the body runs.
	ward.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		if ev.StackItemID == uuid.Nil || ev.Actor == uuid.Nil {
			return nil
		}
		return game.NewKeyedTriggeredItem(source, faceDownWardLabel, facedownWardBody, game.EffectParams{})
	}
	game.CatalogFaceDownWard = func() []game.TriggeredAbility {
		return []game.TriggeredAbility{ward}
	}
}

// facedownWardBody is "facedown/ward" (ADR 0041 P9, #1497, tier 4):
// wardPayOrCounter's targetingItem/payer pair, read back off
// item.Trigger.Event — the harvester's own record of the
// EventBecomesTarget this trigger fired on — instead of a captured
// closure. The mana cost is the state's own constant, never a card's.
var facedownWardBody = game.SimpleDelayedBody("facedown/ward", func(g *game.Game, item *game.StackItem) error {
	// An item with no trigger record (a copy, or a hand-built item) names
	// no targeting spell and no payer, so there is nothing to tax.
	if item.Trigger == nil {
		return nil
	}
	return wardPayOrCounter(g, item, item.Trigger.Event.StackItemID, item.Trigger.Event.Actor, WardMana(faceDownWardCost))
})
