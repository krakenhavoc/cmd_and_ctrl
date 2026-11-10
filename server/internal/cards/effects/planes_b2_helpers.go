package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b2_helpers.go holds the shapes the second and third batches of plane
// cards (ADR 0216 slice 4, #4870) share. Append-only.

// planesB2LoseLife is "you lose N life" for the ability's controller (the
// planar controller on a plane, CR 901.6).
type planesB2LoseLife struct {
	Player uuid.UUID
	Amount int
}

func (l planesB2LoseLife) Apply(ctx *Context) error {
	if l.Amount <= 0 {
		return nil
	}
	player := l.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, -l.Amount)
}

// planesB2Alien is the 2/2 white Alien creature token Doctor Who's planes make.
func planesB2Alien() game.Card { return TokenCard("2/2 white Alien") }

// planesB2CrewGrant is the granted bundle "Crew N": the ability a plane gives
// to the artifacts it makes into Vehicles (New New York, Ghirapur). Keyed
// by the granting plane, an on-disk key.
func planesB2CrewGrant(key string, n int) AbilityGrant {
	label := fmt.Sprintf("Crew %d", n)
	return AbilityGrant{
		Key: key,
		Activated: []ActivatedAbility{{
			Label:   label,
			Cost:    CrewCost(n),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("This Vehicle"),
		}},
		Text: label,
	}
}

// planesB2BecomeVehicles is "until end of turn, <matching artifacts you
// control> become <p>/<t> Vehicles in addition to their other types and gain
// <keywords> and crew <n>" (New New York, Ghirapur). The set is locked as the
// ability resolves (CR 611.2c, the Ghirapur ruling), and the Vehicle type, the
// base power and toughness, the keywords and the crew ability are ONE
// effect at one timestamp (CR 613.7). A Vehicle is a creature only while
// crewed (CR 301.7), so no creature type is granted.
func planesB2BecomeVehicles(label, crewKey string, match CardPredicate, power, toughness int, keywords ...string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		mods := []game.Mod{game.AddSubtypesMod("Vehicle")}
		mods = append(mods, game.SetBasePTMods(power, toughness)...)
		mods = append(mods, game.AddKeywordsMod(keywords...))
		return GrantAbilitiesFor{
			Match:    And(Artifact(), YouControl(), match),
			Keys:     []string{crewKey},
			Also:     mods,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label,
		}.Apply(ctx)
	}
}

// planesB2OwnedByNthOpponent passes for a card owned by the n-th seat (1-based)
// after the chooser's, in turn order: controlledByNthOpponent for a card in a
// graveyard, where an owner holds it.
func planesB2OwnedByNthOpponent(n int) CardPredicate {
	return func(g *game.Game, chooser uuid.UUID, c game.Card) bool {
		seats := len(g.Seats)
		if n < 1 || n >= seats {
			return false
		}
		for i, p := range g.Seats {
			if p != nil && p.ID == chooser {
				o := g.Seats[(i+n)%seats]
				return o != nil && c.Owner == o.ID
			}
		}
		return false
	}
}

// planesB2PerOpponentGraveyardClauses is "for each opponent, choose up to one
// target <noun> in that player's graveyard": one "up to one" clause per
// possible opponent (at most MaxPlayers-1), the n-th bound to the n-th seat
// after the chooser's, the shape perOpponentClauses uses for permanents.
func planesB2PerOpponentGraveyardClauses(noun string, what ...CardPredicate) *game.TargetSpec {
	var out []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		out = append(out, TargetCardInGraveyard("target "+noun+" in the "+nthOpponentWords[n]+" opponent after you in turn order's graveyard",
			append([]CardPredicate{planesB2OwnedByNthOpponent(n)}, what...)...).WithCount(0, 1))
	}
	return Clauses(out[0], out[1:]...)
}
