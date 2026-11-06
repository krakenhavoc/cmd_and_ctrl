package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b26_part27_helpers.go — shared pieces of Avatar: The Last
// Airbender Eternal batch 27 (#5075).

// avB27SwampsOnTheBattlefield is "the number of Swamps on the
// battlefield" (Swampbenders): every land with the subtype under any
// controller, read through the layered type line. The caller reads the
// subtype through its own text (source.TextSubtype).
func avB27SwampsOnTheBattlefield(g *game.Game, swamp string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.HasSubtype(swamp) {
			n++
		}
	}
	return n
}

// avB27LandsYouControlAreAlso is "Lands you control are <subtype>s in
// addition to their other types" (Swampbenders): EachLandIsAlso narrowed
// to the source's controller's lands. Layer 4, an idempotent append.
func avB27LandsYouControlAreAlso(subtype string) game.StaticAbility {
	return game.StaticAbility{
		Label: "Lands you control are " + subtype + "s in addition to their other types.",
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsLand() && target.Controller == source.Controller
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			now := source.TextSubtype(subtype)
			for _, st := range c.Subtypes {
				if strings.EqualFold(st, now) {
					return
				}
			}
			c.Subtypes = append(c.Subtypes, now)
		},
	}
}

// avB27AlliesYouControlHaveDuringYourTurn is "During your turn, Allies
// you control have <keyword>" (Sokka's Charge): a layer-6 grant over the
// Ally creatures the source's controller controls while that player is
// the active player.
func avB27AlliesYouControlHaveDuringYourTurn(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() &&
			target.HasSubtype(source.TextSubtype("Ally")) && isActivePlayer(g, source.Controller)
	}, keyword)
	ab.Label = "During your turn, Allies you control have " + keyword + "."
	return ab
}

// avB27TappedArtifactOrCreature is "tapped artifacts and/or creatures"
// (Aang and Katara): a tapped permanent that is an artifact, a creature
// or both, counted once.
func avB27TappedArtifactOrCreature(c game.Card) bool {
	return c.Tapped && (c.IsArtifact() || c.IsCreature())
}

// avB27CondThisDidntAttack is "if this creature didn't attack this
// turn" (Air Nomad Student) as an intervening-if condition, re-checked
// as the item resolves (CR 603.4). It declares FootprintCondAttacked:
// the attack record only changes by declaring attackers, which no item
// on the stack does.
type avB27CondThisDidntAttack struct{}

func (avB27CondThisDidntAttack) Holds(ctx *Context) bool {
	return !ctx.Game.AttackedThisTurn(ctx.Source())
}

func (avB27CondThisDidntAttack) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondAttacked, 0), true
}

// avB27LandCreaturesYouControlHave is "Land creatures you control have
// <keyword>" (Toph, Greatest Earthbender): a layer-6 grant over each
// permanent the source's controller controls that is both a land and a
// creature, re-read every layer pass.
func avB27LandCreaturesYouControlHave(keyword string) game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsLand() && target.IsCreature()
	}, keyword)
	ab.Label = "Land creatures you control have " + keyword + "."
	return ab
}
