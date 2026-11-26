package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b482_helpers.go — shared pieces of Rise of the Eldrazi batch 482
// (S82, #5581): the level-counter cards.

// OldB482HasLevelUp is "a creature with level up": its layered view still
// lists the "Level up" activated ability (CR 702.87a). A creature that lost
// its abilities does not have it.
func OldB482HasLevelUp(c game.Card) bool {
	for _, ab := range game.ActivatedAbilitiesForCard(c) {
		if strings.HasPrefix(ab.Label, "Level up") {
			return true
		}
	}
	return false
}

// OldB482YouControlACreatureWithLevelCounters is "you control a creature
// with n or more level counters on it", read live.
func OldB482YouControlACreatureWithLevelCounters(n int) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, you uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == you && c.IsCreature() && c.Counters[game.CounterLevel] >= n {
				return true
			}
		}
		return false
	}
}

// OldB482SelfPumpWhileYouControlLevelCreature is "This creature gets +p/+t
// as long as you control a creature with n or more level counters on it"
// (Champion's Drake): a layer 7c modification (CR 613.4c) of the source
// alone, read on every recompute (counters and permanents already
// invalidate it).
func OldB482SelfPumpWhileYouControlLevelCreature(n, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Gets " + SignedPT(power, toughness) + " as long as you control a creature with " + NumberWord(n) + " or more level counters on it.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: WhileYouHave(OldB482YouControlACreatureWithLevelCounters(n), SelfCreature),
		Apply:     PumpBy(power, toughness),
	}
}
