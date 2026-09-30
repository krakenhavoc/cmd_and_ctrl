package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ondu Spiritdancer — Creature — Kor Cleric {4}{W}:
//
//	"Whenever an enchantment you control enters, you may create a
//	 token that's a copy of it. Do this only once each turn."
//
// The "you may" is the CR 603.5 prompt asked as the enchantment
// enters. "Only once each turn" limits the DOING, not the asking: a
// trigger you declined does not use it up, and when two enchantments
// enter together, both triggers are offered but only the first to
// resolve makes a copy — the resolution tally decides, per object, so
// an Ondu that left and came back has a fresh allowance.
//
// The token is a real copy with the enchantment's oracle ID, so it
// brings every ability the catalog has for it, and its own entry does
// not copy again (a token copy's entry is an enchantment entering
// under your control too, but the once-per-turn tally stops it).
//
// Declared weaker than printed for Auras: a token copy of an Aura is
// not attached to anything and is put into the graveyard at once,
// because the copy is not given a host to enchant.
func init() {
	Register(Spec{
		OracleID:     "d2ae893f-52af-4645-8712-d62fb78f901e",
		Name:         "Ondu Spiritdancer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A token copy of an Aura isn't attached to anything, so it doesn't stay on the battlefield."},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, anEnchantmentEnteredUnderYourControl,
				onduSpiritdancerLabel, ondusCopyTheEnchantment),
				"Ondu Spiritdancer — create a token that's a copy of the enchantment?"),
		},
	})
}

const onduSpiritdancerLabel = "Ondu Spiritdancer — create a token that's a copy of that enchantment"

// anEnchantmentEnteredUnderYourControl is "whenever an enchantment you
// control enters" — any enchantment, the source included if it is one —
// and this Ondu has not made its copy yet this turn, so a player is not
// asked about one they could not make (the token's own entry, say).
func anEnchantmentEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsEnchantment() && g.ResolvedThisTurn(source.InstanceID, onduSpiritdancerLabel) == 0
}

// ondusCopyTheEnchantment makes the copy, unless this Ondu has already
// made one this turn. The tally counts the resolution in progress, so
// the first one sees 1.
func ondusCopyTheEnchantment(g *game.Game, item *game.StackItem) error {
	if g.ResolvedThisTurn(item.SourceCardID, onduSpiritdancerLabel) > 1 {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.Trigger.Event.CardID,
		N:          1,
	}.Apply(NewContext(g, item))
}
