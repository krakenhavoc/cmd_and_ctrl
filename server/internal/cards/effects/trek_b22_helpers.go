package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b22_helpers.go — shared pieces for Star Trek batch 22 (#4973).

// trekB22CreateTappedNewPlanet is "create a tapped Planet land token named
// New Planet" (batch 16's template, trek_b16_helpers.go) as one Do step,
// so a trigger that does only this declares its footprint.
func trekB22CreateTappedNewPlanet() CreateTokenAdvanced {
	return CreateTokenAdvanced{Spec: Token(trekB16NewPlanetToken()).EntersTapped(), N: 1}
}

// trekB22AbilitiesFromOpponentsArtifactSourcesCantTargetThis is Species
// 8472's "This creature can't be the target of abilities from artifact
// sources your opponents control."
//
// It refuses only ABILITIES (TargetSource.IsSpell is false): an artifact
// spell is still free to target it, which is where the clause comes apart
// from "hexproof from artifacts" (CR 702.11e refuses spells too). The
// source is judged as it is now, or as it last existed when it has left
// the battlefield (CR 608.2h, stackItemSourceLocked). "Your opponents" is
// read off the creature's controller, as hexproof's is (CR 702.11b), so the
// creature's own controller may target it with an artifact's ability.
func trekB22AbilitiesFromOpponentsArtifactSourcesCantTargetThis(label string) game.TargetingRestriction {
	return game.TargetingRestriction{
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		Forbids: func(q game.TargetingQuery) bool {
			if q.Card.InstanceID != q.Source.InstanceID || q.Controller == q.Card.Controller {
				return false
			}
			if q.Controller == uuid.Nil || q.By.IsSpell() {
				return false
			}
			ch := q.By.Characteristics()
			if ch == nil {
				return false
			}
			for _, t := range ch.Types {
				if strings.EqualFold(t, "artifact") {
					return true
				}
			}
			return false
		},
	}
}

// trekB22AnyCard matches every card ("a card from your graveyard").
func trekB22AnyCard(game.Card) bool { return true }

// TwoPlusOneCountersOnTheTarget is "put two +1/+1 counters on (up to one)
// target creature" (Syndicate Liquidators, Oko, the Trickster's +1): two counters on the item's first card target still legal as
// it resolves (CR 608.2b), nothing when none is.
func TwoPlusOneCountersOnTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		}
	}
	return nil
}
