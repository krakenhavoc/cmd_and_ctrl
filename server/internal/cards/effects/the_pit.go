package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Pit — Plane — The Abyss (#4870):
//
//	"When you planeswalk to The Pit, each player creates their choice of a
//	 3/3 white Angel creature token with flying or a 6/6 black Demon creature
//	 token with flying, trample, and "At the beginning of your upkeep,
//	 sacrifice another creature. If you can't, this token deals 6 damage to
//	 you."
//	 Whenever chaos ensues, each player sacrifices a nonartifact creature of
//	 their choice."
//
// Per the ruling, starting with the planar controller and proceeding in turn
// order each player chooses which token (a keyed option pick that carries the
// players still to choose and the choices made), then every token is created
// at the same time (game.CreateTokensForPlayersTogetherForEffect). The Demon's
// upkeep ability lives on the token (token_catalog.go, "demon-the-pit"): it
// can't sacrifice itself ("another"), and with no other creature it deals the
// 6 damage to its controller. The chaos edict is EachPlayerSacrifices: the
// choices are made in turn order and the creatures leave together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b60e8617-fc35-4c8a-beb0-4688b88484c4",
		Name:         "The Pit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("The Pit — each player creates their choice of a 3/3 Angel or a 6/6 Demon",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return thePitAsk(ctx, apnapPlayers(g), nil)
				}),
			WheneverChaosEnsues("The Pit — each player sacrifices a nonartifact creature of their choice",
				Do(EachPlayerSacrifices{Match: And(Creature(), Not(Artifact())), Label: "a nonartifact creature"})),
		},
	})
}

// thePitAsk asks the first of `left` which token they want, carrying the
// others and the answers (0 Angel, 1 Demon) already given; with nobody left
// it creates every token together.
func thePitAsk(ctx *Context, left []uuid.UUID, picks []int) error {
	if len(left) == 0 {
		return thePitCreate(ctx, picks)
	}
	return PickOption{
		Player:   left[0],
		Question: "The Pit — create a 3/3 Angel with flying or a 6/6 Demon with flying, trample and an upkeep sacrifice",
		Options: []game.ChoiceOption{
			{Label: "3/3 white Angel creature token with flying"},
			{Label: "6/6 black Demon creature token with flying, trample, and \"At the beginning of your upkeep, sacrifice another creature. If you can't, this token deals 6 damage to you.\""},
		},
		ThenKey:  thePitChose,
		Carry:    append([]uuid.UUID(nil), left[1:]...),
		Nums:     append([]int(nil), picks...),
		WithItem: true,
	}.Apply(ctx)
}

// thePitChose records one answer and asks the next player. An on-disk key:
// never renamed, never reused.
// Registered in init: the answer asks the next player, which names this key.
var thePitChose game.OptionPickThen

func init() {
	thePitChose = OptionPickThen("option-pick/the-pit-token-choice", func(ctx *Context, r game.OptionPicked) error {
		picks := append(append([]int(nil), r.Nums...), max(r.Index, 0))
		return thePitAsk(ctx, r.Carry, picks)
	})
}

// thePitCreate creates every player's chosen token in one entry. The players
// are the living seats in turn order from the planar controller, the order
// the answers were given in.
func thePitCreate(ctx *Context, picks []int) error {
	players := apnapPlayers(ctx.Game)
	var specs []game.TokenCreation
	for i, p := range players {
		if i >= len(picks) {
			break
		}
		tmpl := TokenCard("3/3 white Angel with flying")
		if picks[i] == 1 {
			tmpl = ThePitDemonToken()
		}
		specs = append(specs, game.TokenCreation{
			Controller: p, Source: ctx.Source(),
			Groups: []game.TokenGroup{{Template: tmpl, Count: 1}},
		})
	}
	return ctx.Game.CreateTokensForPlayersTogetherForEffect(specs)
}

// ThePitDemonToken is The Pit's 6/6 black Demon creature token with flying,
// trample and the upkeep drawback.
func ThePitDemonToken() game.Card { return tokenFromCatalog(printedThePitDemonToken) }

// printedThePitDemonToken is that Demon as PRINTED. The trigger is the
// token's own, controlled by whoever controls the token (CR 603.3a): "another
// creature" excludes the token itself, and "if you can't" is the case of no
// other creature to sacrifice (or none the controller may sacrifice), where
// the token deals 6 damage to its controller.
func printedThePitDemonToken() tokenTemplate {
	return tokenTemplate{
		Slug: "demon-the-pit",
		Card: game.Card{
			Name:      "Demon",
			TypeLine:  "Token Creature — Demon",
			Power:     6,
			Toughness: 6,
			Colors:    []string{"B"},
			Keywords:  []string{"flying", "trample"},
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Demon — sacrifice another creature; if you can't, this token deals 6 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					var others []uuid.UUID
					if g.CanEffectCauseSacrificeForEffect(item.Controller) {
						for _, id := range b04CreatureIDsControlledBy(g, item.Controller) {
							if id != item.SourceCardID {
								others = append(others, id)
							}
						}
					}
					ctx := NewContext(g, item)
					if len(others) == 0 {
						return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 6}.Apply(ctx)
					}
					return SacrificeChoice{
						Player: item.Controller, Candidates: others,
						Question: "Demon — sacrifice another creature",
					}.Apply(ctx)
				}),
		},
		Text: "Flying, trample. At the beginning of your upkeep, sacrifice another creature. If you can't, this token deals 6 damage to you.",
	}
}
