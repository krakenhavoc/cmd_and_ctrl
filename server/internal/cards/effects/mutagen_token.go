package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// MutagenToken — CR 111.10v: "A Mutagen token is a colorless Mutagen
// artifact token with “{1}, {T}, Sacrifice this token: Put a +1/+1
// counter on target creature. Activate only as a sorcery.”" Mutagen is
// an artifact type (CR 205.3g). The Teenage Mutant Ninja Turtles cards
// make it (#4676).
//
// The target is any creature, whoever controls it. It is chosen as the
// ability is activated (CR 602.2b) and checked again as it resolves, so
// a creature that left in response gets nothing (CR 608.2b); the token
// is gone either way, since sacrificing it is the cost.
func MutagenToken() game.Card { return tokenFromCatalog(printedMutagenToken) }

// printedMutagenToken is the Mutagen as PRINTED — the abilities
// included. It is the catalog's entry for this token (token_catalog.go).
// Its slug is an on-disk identity and is never renamed.
//
// No Purpose.Answers: the row is sorcery speed and always names a
// target, which is the shape ADR 0142 leaves undeclared.
func printedMutagenToken() tokenTemplate {
	return tokenTemplate{
		Slug: "mutagen",
		Card: game.Card{
			Name:     "Mutagen",
			TypeLine: "Token Artifact — Mutagen",
		},
		Text: "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target creature. Activate only as a sorcery.",
		Activated: []game.ActivatedAbilityShape{{
			Label: "{1}, {T}, Sacrifice this artifact: Put a +1/+1 counter on target creature",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{1}",
			},
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       PlusOneCounterOnTheTarget,
		}},
	}
}
