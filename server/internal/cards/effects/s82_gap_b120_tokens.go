package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b120_tokens.go — the token template of S82 set GAP, batch 120
// (#5948). Its slug is an on-disk identity: never renamed, never reused.

// S82GapB120DragonToken is Wake the Dragon's Dragon for CreateToken.
func S82GapB120DragonToken() game.Card { return TokenFromCatalog(PrintedS82GapB120Dragon) }

// PrintedS82GapB120Dragon is the Dragon as PRINTED (Tales of Middle-earth
// Commander token): a 6/6 black and red Dragon with flying, menace and
// "Whenever this token deals combat damage to a player, gain control of
// target artifact that player controls." The trigger's target is an artifact
// controlled by the player the event names, chosen as it goes on the stack;
// the control lasts indefinitely (CR 611.2a), as Hellkite Tyrant's does.
func PrintedS82GapB120Dragon() TokenTemplate {
	return TokenTemplate{
		Slug: "dragon-steals-artifact",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     6,
			Toughness: 6,
			Colors:    []string{"B", "R"},
			Keywords:  []string{"flying", "menace"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisDealsCombatDamageToAPlayer("Dragon — gain control of target artifact that player controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 || ts[0].Kind != game.TargetCard {
						return nil
					}
					return GainControl{
						Target:     ts[0].ID,
						Controller: item.Controller,
						Duration:   game.IndefiniteDuration(),
						Label:      "Dragon — gain control of target artifact",
					}.Apply(ctx)
				}),
				ControlledByThatPlayer(TargetPermanent("target artifact that player controls", Artifact()))),
		},
		Text: "Flying, menace\nWhenever this token deals combat damage to a player, gain control of target artifact that player controls.",
	}
}
