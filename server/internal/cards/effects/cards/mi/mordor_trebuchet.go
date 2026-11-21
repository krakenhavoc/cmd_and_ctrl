package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mordor Trebuchet — Artifact Creature — Wall {2}{B}, 1/4 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"Defender
//	 Whenever you attack with one or more Goblins and/or Orcs, create a 2/1
//	 colorless Construct artifact creature token with flying named
//	 Ballistic Boulder that's tapped and attacking. Sacrifice that token at
//	 end of combat."
//
// Fire Navy Trebuchet's trigger (one per declaration of attackers,
// OncePerBatch) narrowed to a declaration that includes a Goblin or Orc,
// and Kavaron Harrier's tapped-and-attacking token (the controller chooses
// what it attacks; it was never declared, CR 508.4) with the delayed
// sacrifice at the beginning of the end of combat step, for exactly the
// tokens that landed. The creature types are read through the source's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a8e0d0a-1101-45b6-8f70-1de2c07bc369",
		Name:            "Mordor Trebuchet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && (c.HasSubtype(source.TextSubtype("Goblin")) || c.HasSubtype(source.TextSubtype("Orc")))
			}, "Mordor Trebuchet — create a tapped and attacking Ballistic Boulder",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokensAttackingYourChoice{
						Template: effects.TokenCard("2/1 colorless Construct artifact with flying named Ballistic Boulder"),
						N:        1,
						Tapped:   true,
						Label:    "Mordor Trebuchet",
						Tail:     game.KeyedTail{Ref: mordorTrebuchetSacrifice},
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// mordorTrebuchetSacrifice is the Boulder's tail: the end of combat trigger
// that sacrifices it. An on-disk key: never renamed, never reused.
var mordorTrebuchetSacrifice = effects.AttackingTokensThen("attacking-tokens/mordor-trebuchet-sacrifice-the-boulder",
	func(ctx *effects.Context, created []uuid.UUID, _ game.KeyedTail) error {
		return effects.SacrificeAtEndOfCombat(ctx, created, "Mordor Trebuchet — sacrifice Ballistic Boulder")
	})
