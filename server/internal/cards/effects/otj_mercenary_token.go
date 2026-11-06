package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// otj_mercenary_token.go — Outlaws of Thunder Junction's Mercenary token
// (#4468): "a 1/1 red Mercenary creature token with '{T}: Target
// creature you control gets +1/+0 until end of turn. Activate only as a
// sorcery.'" Nezumi Linkbreaker, Prickly Pair, Prosperity Tycoon and
// Mourner's Surprise make it.
//
// The ability is printed on the token, so it is a token template (ADR
// 0083) and not a tokens_table.go row. The pump is the before-attackers
// cards' body (boostTargetUntilEOT), read at resolution (CR 608.2b), and
// "Activate only as a sorcery" is the row's SorcerySpeed (CR 602.5d).
// The {T} is a creature's tap cost, so a Mercenary that came under its
// controller's control this turn can't pay it.

// MercenaryToken is the Outlaws of Thunder Junction Mercenary.
func MercenaryToken() game.Card { return tokenFromCatalog(printedMercenaryToken) }

// printedMercenaryToken is the Mercenary as PRINTED.
//
// No Purpose.Answers: the row is sorcery speed and always names a
// target, the shape ADR 0142 leaves undeclared.
func printedMercenaryToken() tokenTemplate {
	return tokenTemplate{
		Slug: "mercenary",
		Card: game.Card{
			Name:      "Mercenary",
			TypeLine:  "Token Creature — Mercenary",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Text: "{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery.",
		Activated: []game.ActivatedAbilityShape{{
			Label:        "{T}: Target creature you control gets +1/+0 until end of turn. Activate only as a sorcery.",
			Cost:         game.AbilityCost{Tap: true},
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Mercenary — +1/+0", 1, 0),
		}},
	}
}

// createAMercenary is "create a 1/1 red Mercenary creature token with …"
// as an ability's whole effect, under the ability's controller.
func createAMercenary() Effect { return Do(CreateToken{Template: MercenaryToken(), N: 1}) }
