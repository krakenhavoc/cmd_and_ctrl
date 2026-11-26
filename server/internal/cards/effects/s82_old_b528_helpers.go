package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b528_helpers.go — Shadowmoor, S82 older sets batch 528 (#5591): the
// Duos' two cast triggers and Isleback Spawn's library-size static.

// S82OldB528Duo is a Duo's two triggered abilities (Emberstrike, Safehold,
// Tattermunge, Gravelgill and Thistledown Duo): "Whenever you cast a <pump
// colour> spell, this creature gets +1/+1 until end of turn" and "Whenever you
// cast a <grant colour> spell, this creature gains <keyword> until end of
// turn". They are two abilities (CR 603.2), so a spell of both colours
// triggers both, and each reads the spell's colours as it was cast (CR 601.2i).
// The effects are pinned to this creature as the object that was on the
// battlefield when the trigger resolved (CR 400.7).
func S82OldB528Duo(name, pumpColor, grantColor, keyword string) []game.TriggeredAbility {
	return []game.TriggeredAbility{
		WheneverYouCast(OfColor(pumpColor), name+" — this creature gets +1/+1 until end of turn",
			ThisGetsUntilEndOfTurn(1, 1, name+" — +1/+1 until end of turn")),
		WheneverYouCast(OfColor(grantColor), name+" — this creature gains "+keyword+" until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{keyword},
					Label: name + " — gains " + keyword + " until end of turn"}.Apply(NewContext(g, item))
			}),
	}
}

// ALibraryHasAtMost is "a library has N or fewer cards in it": any seated
// player's library, an eliminated player's having left the game.
func ALibraryHasAtMost(g *game.Game, n int) bool {
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Library != nil && len(p.Library.Cards) <= n {
			return true
		}
	}
	return false
}

// SelfPumpWhileALibraryHasAtMost is "This creature gets +p/+t as long as a
// library has N or fewer cards in it" (Isleback Spawn): a layer 7c
// modification (CR 613.4c) of the source alone, read on every recompute and
// invalidated by every change to a library's size (DependsOnLibrarySize).
func SelfPumpWhileALibraryHasAtMost(n, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:                "Gets " + SignedPT(power, toughness) + " as long as a library has " + NumberWord(n) + " or fewer cards in it.",
		Layer:                game.Layer7PT,
		SubLayer:             game.SubLayer7C_Modify,
		DependsOnLibrarySize: true,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return SelfCreature(target, g, source) && ALibraryHasAtMost(g, n)
		},
		Apply: PumpBy(power, toughness),
	}
}
