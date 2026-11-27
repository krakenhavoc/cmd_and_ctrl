package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b566_helpers.go — shared pieces of Coldsnap batch 566 (#5597, S82).

// S82OldB566Ripple is the keyword "ripple N" on the spell that has it
// (CR 702.60a): "When you cast this spell, you may reveal the top N cards of
// your library. You may cast spells with the same name as this spell from
// among the revealed cards without paying their mana costs. Put the rest on
// the bottom of your library in any order." It triggers while the spell is on
// the stack (Cascade's FromStack shape); the reveal, the free casts (each its
// own "may", and each cast triggers ripple again) and the bottoming are
// Thrumming Stone's, which gives the same keyword to every spell.
func S82OldB566Ripple(name string, n int) game.TriggeredAbility {
	t := GapB200RippleTrigger(name, n)
	t.AppliesTo = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.CardID == source.InstanceID
	}
	t.FromStack = true
	return t
}

// S82OldB566SnowTapland is a snow land with no land types that "enters tapped"
// and "{T}: Add {a} or {b}" (Coldsnap's cycle): the entry self-replacement
// (CR 614.1c) and the two-colour mana ability the type line would not give it.
func S82OldB566SnowTapland(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{DualManaAbility(a, b)},
	}
}
