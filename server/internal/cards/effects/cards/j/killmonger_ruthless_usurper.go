package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Killmonger, Ruthless Usurper — {2}{R} Legendary Creature — Human Villain
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Trample
//	 Whenever Killmonger attacks, he gets +1/+0 until end of turn for each
//	 artifact defending player controls.
//	 Whenever Killmonger deals combat damage to a player, that player
//	 sacrifices an artifact of their choice and you create a Treasure
//	 token."
//
// "Defending player" is the player behind the attack target (a
// planeswalker or battle counts as its controller) and the artifacts are
// counted as the trigger resolves. The combat-damage trigger names the
// damaged player off the damage event; they pick the artifact (none, none
// sacrificed), and the Treasure is made either way, since the two halves
// are joined by "and", not "if you do".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aff21bd1-87b0-409c-bfad-08f3690ef220",
		Name:            "Killmonger, Ruthless Usurper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Killmonger — he gets +1/+0 until end of turn for each artifact defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !effects.B09SourceStillOnBattlefield(g, item) {
						return nil
					}
					defender := effects.B17DefendingPlayer(g, item.Trigger.Event)
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == defender && c.IsArtifact() {
							n++
						}
					}
					return effects.BoostUntilEOT{
						Target: item.SourceCardID, Power: n, Label: "Killmonger — +1/+0 for each artifact",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverThisDealsCombatDamageToAPlayer("Killmonger — that player sacrifices an artifact and you create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					victim := item.Trigger.Event.Target
					var artifacts []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == victim && c.IsArtifact() {
							artifacts = append(artifacts, c.InstanceID)
						}
					}
					if len(artifacts) > 0 {
						if err := (effects.SacrificeChoice{
							Player:     victim,
							Candidates: artifacts,
							Question:   "Killmonger — sacrifice an artifact",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
