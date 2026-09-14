package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archetype of Aggression — Enchantment Creature — Human Warrior
// {1}{R}{R}, 3/2 (EDHREC rank 2850):
//
//	"Creatures you control have trample.
//	 Creatures your opponents control lose trample and can't have or
//	 gain trample."
//
// Two layer-6 statics: a keyword grant to the controller's creatures
// (Lord of Atlantis's shape) and a keyword REMOVAL from everyone
// else's — the first "lose <keyword>" static in the catalog
// (b27LoseKeyword), which strips trample from whatever the layers
// have granted a creature so far. A printed trample is in the
// layer-0 baseline, so an opponent's trampler that enters after the
// Archetype still loses it.
//
// Sandbox simplification, WEAKER than printed: "can't have or gain"
// is a CR 101.2 "can't" — it beats every grant regardless of
// timestamp — and the layer engine has no such gate. So a trample
// grant from a source with a NEWER timestamp than the Archetype
// applies after the removal and sticks: an opponent's Garruk's
// Uprising cast after it, and — because a catalog card's
// PrintedKeywords are re-applied by a self-only layer-6 static with
// the card's own timestamp — a catalog trampler that enters after
// it. (A trampler imported from a decklist carries the keyword on
// Card.Keywords, which is layer-0 baseline and is stripped.)
// Opponents keep a little trample they should not; the Archetype's
// controller is never given anything extra. Pinned by
// TestB27ArchetypeOfAggressionGivesAndTakesTrample.
func init() {
	Register(Spec{
		OracleID:     "263408e6-b315-4af5-8cb8-3fd1aa88e48c",
		Name:         "Archetype of Aggression",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Trample an opponent's creature gets from anything that entered the battlefield after the Archetype isn't taken away."},
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "trample"),
			b27LoseKeyword(b27CreaturesOpponentsControl, "trample"),
		},
	})
}
