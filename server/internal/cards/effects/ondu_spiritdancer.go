package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ondu Spiritdancer — Creature — Kor Cleric {4}{W}:
//
//	"Whenever an enchantment you control enters, you may create a
//	 token that's a copy of it. Do this only once each turn."
//
// The "you may" is the CR 603.5 prompt asked as the trigger resolves.
// "Only once each turn" limits the DOING, not the asking: a trigger you
// declined does not use it up, and when two enchantments enter
// together, both triggers are offered but only the first "yes" makes a
// copy — the done tally decides (b28DoneThisTurn), per object, so an
// Ondu that left and came back has a fresh allowance.
//
// The token is a real copy with the enchantment's oracle ID, so it
// brings every ability the catalog has for it, and its own entry does
// not copy again (a token copy's entry is an enchantment entering
// under your control too, but the once-per-turn tally stops it).
//
// A token copy of an Aura chooses what it enchants as it is created
// (CR 303.4f, #3158); with nothing it could enchant, it isn't created
// (CR 303.4g).
func init() {
	Register(Spec{
		OracleID:     "d2ae893f-52af-4645-8712-d62fb78f901e",
		Name:         "Ondu Spiritdancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, anEnchantmentEnteredUnderYourControl,
				onduSpiritdancerLabel, ondusCopyTheEnchantment),
				"Ondu Spiritdancer — create a token that's a copy of the enchantment?"),
		},
	})
}

const onduSpiritdancerLabel = "Ondu Spiritdancer — create a token that's a copy of that enchantment"

// onduSpiritdancerDone is the once-each-turn tally's label: the copy
// was made this turn.
const onduSpiritdancerDone = "Ondu Spiritdancer — copied an enchantment this turn"

// anEnchantmentEnteredUnderYourControl is "whenever an enchantment you
// control enters" — any enchantment, the source included if it is one —
// and this Ondu has not made its copy yet this turn, so a player is not
// asked about one they could not make (the token's own entry, say).
func anEnchantmentEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsEnchantment() && !b28DoneThisTurn(g, source.InstanceID, onduSpiritdancerDone)
}

// ondusCopyTheEnchantment makes the copy, unless this Ondu has already
// made one this turn.
func ondusCopyTheEnchantment(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, onduSpiritdancerDone) {
		return nil
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, onduSpiritdancerDone)
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.Trigger.Event.CardID,
		N:          1,
	}.Apply(NewContext(g, item))
}
