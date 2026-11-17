package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// no_untapped_lands.go — "if you control no untapped lands" and "as long
// as you control no untapped lands" (the Prophecy condition, #5794).
//
// One fact about the battlefield: the player controls no permanent that is
// both a land and untapped. It is read through the layers (a land creature
// is a land, an animated land still is one, a phased-out land is not on
// the battlefield, CR 702.26b), with the same "you control" walk the other
// board conditions use (permanentsControlledByMatching) and the same
// untapped-land test as Lotus Vale's cost (untappedLand). A player who
// controls no lands at all controls no untapped lands.
//
// Readers, one per place the printed text puts the condition:
//
//   - NoUntappedLands(g, you) for a spell's or ability's own resolution.
//   - SelfPumpWhileNoUntappedLands for "this creature gets +p/+t as long
//     as …" (Scoria Cat, Spur Grappler), a layer 7c static (CR 613.4c).
//   - SelfKeywordWhileNoUntappedLands for "this creature has <keyword>
//     as long as …" (Fen Stalker, Vintara Snapper), a layer 6 grant
//     (CR 613.1f).
//   - WheneverThisAttacksIfNoUntappedLands and
//     AtYourEndStepIfNoUntappedLands are the trigger shapes: an
//     intervening "if" (CR 603.4), checked as the trigger fires and
//     again as it resolves (InterveningIf with CondNoUntappedLands).
//
// # Why the statics stay current
//
// A static's condition is not locked in (CR 611.3a): it applies at any
// moment to whatever its text says. The layer cache drops on every tap
// and untap (layerVersionBump's EventTapCard / EventUntapCard arm), on
// every land entering or leaving, and on a change of control, so the
// statics need no hint for those. The one tap that emits no
// EventTapCard is an attacker's own (CR 508.1f: declaring a creature as
// an attacker taps it), and an animated land can attack; the statics
// declare DependsOnEvents{EventAttack}, so the cache drops as the
// declaration is locked in and announced.

// NoUntappedLands reports whether `you` controls no untapped lands.
// Caller must hold g.mu (every effect body and every static's AppliesTo
// does).
func NoUntappedLands(g *game.Game, you uuid.UUID) bool {
	if you == uuid.Nil {
		return false
	}
	return len(permanentsControlledByMatching(g, you, untappedLandPredicate)) == 0
}

// untappedLandPredicate is untappedLand as a CardPredicate.
func untappedLandPredicate(_ *game.Game, _ uuid.UUID, c game.Card) bool { return untappedLand(c) }

// CondNoUntappedLands is "if you control no untapped lands" as an
// intervening-if Condition, for InterveningIf and If: read for the item's
// controller as it resolves. It declares a read of the battlefield
// (game.FootprintCondBoard): only a tap, an untap, a permanent entering
// or leaving, a change of control or a change of type can change the
// answer, and no footprinted step does any of those without the check
// seeing it (a tap or an untap is no footprinted step at all, so an item
// that does one already conflicts with everything).
type CondNoUntappedLands struct{}

func (CondNoUntappedLands) Holds(ctx *Context) bool {
	return NoUntappedLands(ctx.Game, ctx.Controller())
}

func (CondNoUntappedLands) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondBoard, 0), true
}

// YouControlNoUntappedLands is the When for "if you control no untapped
// lands" as the trigger fires, for composing with AllOf. "You" is the
// source's controller.
func YouControlNoUntappedLands(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return NoUntappedLands(g, source.Controller)
}

// selfWhileNoUntappedLands is "this creature …, as long as you control no
// untapped lands": the source, while it is a creature, and while its
// controller controls no untapped lands.
func selfWhileNoUntappedLands(target *game.Card, g *game.Game, source *game.Card) bool {
	return selfCreature(target, g, source) && NoUntappedLands(g, source.Controller)
}

// SelfPumpWhileNoUntappedLands is "This creature gets +p/+t as long as you
// control no untapped lands" (Scoria Cat, Spur Grappler): a layer 7c
// modification (CR 613.4c) that is on while the source's controller
// controls no untapped lands.
func SelfPumpWhileNoUntappedLands(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:           "This creature gets " + signedPT(power, toughness) + " as long as you control no untapped lands.",
		Layer:           game.Layer7PT,
		SubLayer:        game.SubLayer7C_Modify,
		AppliesTo:       selfWhileNoUntappedLands,
		Apply:           pumpBy(power, toughness),
		DependsOnEvents: []game.EventKind{game.EventAttack},
	}
}

// SelfKeywordWhileNoUntappedLands is "This creature has <keyword> as long
// as you control no untapped lands" (Fen Stalker's fear, Vintara
// Snapper's shroud): a layer 6 grant (CR 613.1f) that is on while the
// source's controller controls no untapped lands.
func SelfKeywordWhileNoUntappedLands(keyword string) game.StaticAbility {
	ab := KeywordGrant(selfWhileNoUntappedLands, keyword)
	ab.Label = "This creature has " + keyword + " as long as you control no untapped lands."
	ab.DependsOnEvents = []game.EventKind{game.EventAttack}
	return ab
}

// WheneverThisAttacksIfNoUntappedLands is "Whenever this creature attacks,
// if you control no untapped lands, …" (Keldon Berserker). The attack
// taps the attackers first (CR 508.1f), so an animated land that attacks
// beside it is already tapped when the condition is first read
// (CR 508.1m).
func WheneverThisAttacksIfNoUntappedLands(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondNoUntappedLands{}, On(game.EventAttack, AllOf(ThisAttacked, YouControlNoUntappedLands), label, effect))
}

// AtYourEndStepIfNoUntappedLands is "At the beginning of your end step, if
// you control no untapped lands, …" (Well of Life, Well of Discovery):
// the intervening "if" is checked as the end step begins and again as the
// trigger resolves (CR 603.4), so a land untapped in response leaves the
// ability doing nothing.
func AtYourEndStepIfNoUntappedLands(label string, effect Effect) game.TriggeredAbility {
	return InterveningIf(CondNoUntappedLands{}, On(game.EventBeginEndStep, AllOf(ByYou, YouControlNoUntappedLands), label, effect))
}
