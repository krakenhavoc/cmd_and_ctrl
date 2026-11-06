package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// avatar_b11_helpers.go — shared pieces of Avatar: The Last Airbender
// batch 11 (#4459).

// avatarB11CondYouAttacked is raid's intervening "if you attacked this
// turn" as an If condition (CR 603.4), re-checked as the item resolves.
// It declares FootprintCondAttacked: the attack tally only rises, and
// only by declaring attackers, which no item on the stack does.
type avatarB11CondYouAttacked struct{}

func (avatarB11CondYouAttacked) Holds(ctx *Context) bool {
	return b18AttackedThisTurn(ctx.Game, ctx.Controller())
}

func (avatarB11CondYouAttacked) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondAttacked, 0), true
}

// avatarB11LessonInYourGraveyard is "as long as there's a Lesson card in
// your graveyard" (Fire Nation Cadets), read as the creature attacks: the
// firebending it grants triggers only if it has the ability then.
func avatarB11LessonInYourGraveyard(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return len(b42CardsInGraveyardOf(g, source.Controller, isLessonCard)) > 0
}

// avatarB11FirebendingSoldierToken is Fire Nation Attacks' 2/2 red
// Soldier with firebending 1.
func avatarB11FirebendingSoldierToken() game.Card {
	return tokenFromCatalog(printedFirebendingSoldierToken)
}

// printedFirebendingSoldierToken is that Soldier as PRINTED: firebending
// is a triggered ability, so it rides the catalog slot (ADR 0083).
func printedFirebendingSoldierToken() tokenTemplate {
	return tokenTemplate{
		Slug: "soldier-firebending",
		Card: game.Card{
			Name:      "Soldier",
			TypeLine:  "Token Creature — Soldier",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Soldier — firebending 1", Do(firebending(1))),
		},
		Text: "Firebending 1 (Whenever this token attacks, add {R}. " +
			"This mana lasts until end of combat.)",
	}
}
