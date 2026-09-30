package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mechanized Production — Enchantment — Aura {2}{U}{U}:
//
//	"Enchant artifact you control
//	 At the beginning of your upkeep, create a token that's a copy of
//	 enchanted artifact. Then if you control eight or more artifacts
//	 with the same name as one another, you win the game."
//
// The token is created through the creation event, so a doubler
// doubles it, and the win check is that creation's continuation: the
// token can stop to ask (an entry replacement), and the count has to
// include it. The check runs whether or not a token was made — "Then
// if" is a separate sentence. Names are read from effective
// characteristics, so a copy effect changes what counts; a token
// copy and the artifact it copies share a name.
//
// The win is the ordinary win-the-game primitive, so a player who
// can't win (Platinum Angel on the other side) is stopped by the
// engine's gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39d8406a-90c3-460c-b4e8-0f590573db51",
		Name:         "Mechanized Production",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact you control", Artifact(), YouControl()),
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Mechanized Production — create a token that's a copy of enchanted artifact",
				mechanizedProductionUpkeep),
		},
	})
}

// mechanizedProductionUpkeep copies the enchanted artifact, then
// checks the win condition once the token is in play.
func mechanizedProductionUpkeep(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return mechanizedProductionCheckWin(g, item)
	}
	tmpl, ok := TokenCopyTemplate(g, host.InstanceID)
	if !ok {
		return mechanizedProductionCheckWin(g, item)
	}
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: tmpl, Count: 1}},
	}, func(g *game.Game, _ []uuid.UUID) error {
		return mechanizedProductionCheckWin(g, item)
	})
}

// mechanizedProductionCheckWin is "then if you control eight or more
// artifacts with the same name as one another, you win the game".
func mechanizedProductionCheckWin(g *game.Game, item *game.StackItem) error {
	if mostCommonArtifactNameCount(g, item.Controller) < 8 {
		return nil
	}
	err := WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	if errors.Is(err, game.ErrStopResolution) {
		return nil
	}
	return err
}

// mostCommonArtifactNameCount is the size of the largest set of
// artifacts `player` controls that share one name.
func mostCommonArtifactNameCount(g *game.Game, player uuid.UUID) int {
	counts := map[string]int{}
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsArtifact() {
			continue
		}
		name := c.Effective().Name
		counts[name]++
		if counts[name] > best {
			best = counts[name]
		}
	}
	return best
}
