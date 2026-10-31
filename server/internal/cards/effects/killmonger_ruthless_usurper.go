package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "aff21bd1-87b0-409c-bfad-08f3690ef220",
		Name:            "Killmonger, Ruthless Usurper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Killmonger — he gets +1/+0 until end of turn for each artifact defending player controls",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					defender := b17DefendingPlayer(g, item.Trigger.Event)
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == defender && c.IsArtifact() {
							n++
						}
					}
					return BoostUntilEOT{
						Target: item.SourceCardID, Power: n, Label: "Killmonger — +1/+0 for each artifact",
					}.Apply(NewContext(g, item))
				}),
			WheneverThisDealsCombatDamageToAPlayer("Killmonger — that player sacrifices an artifact and you create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := NewContext(g, item)
					victim := item.Trigger.Event.Target
					var artifacts []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == victim && c.IsArtifact() {
							artifacts = append(artifacts, c.InstanceID)
						}
					}
					if len(artifacts) > 0 {
						if err := (SacrificeChoice{
							Player:     victim,
							Candidates: artifacts,
							Question:   "Killmonger — sacrifice an artifact",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
