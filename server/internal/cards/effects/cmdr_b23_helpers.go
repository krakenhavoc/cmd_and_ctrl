package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b23_helpers.go — helpers for the Commander deck cards, batch 23
// (#5368). Every name carries the cmdrB23 prefix.

// cmdrB23WithCardTypesInAddition is a type line with the card types added
// "in addition to its other types" (CR 707.9b), keeping its supertypes and
// subtypes; a type it already has is not repeated.
func cmdrB23WithCardTypesInAddition(printed string, add ...string) string {
	super, types, subs := game.ParseTypeLine(printed)
	for _, a := range add {
		have := false
		for _, t := range types {
			if strings.EqualFold(t, a) {
				have = true
			}
		}
		if !have {
			types = append(types, a)
		}
	}
	head := make([]string, 0, len(super)+len(types))
	head = append(head, super...)
	head = append(head, types...)
	line := strings.Join(head, " ")
	if len(subs) > 0 {
		line += " — " + strings.Join(subs, " ")
	}
	return line
}

// cmdrB23ZombiesYouControl counts the Zombies controller controls, read
// through the layered view.
func cmdrB23ZombiesYouControl(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Zombie") {
			n++
		}
	}
	return n
}

// printedCmdrB23NalaarAetherjetToken is Nalaar Aetherjet as PRINTED: a
// colorless Vehicle artifact token with flying and crew 2 whose power and
// toughness are the X the creating ability gives (the template's 1/1 is
// replaced at creation, see CmdrB23NalaarAetherjetToken). Its slug is an
// on-disk identity and is never renamed.
func printedCmdrB23NalaarAetherjetToken() tokenTemplate {
	return tokenTemplate{
		Slug: "nalaar-aetherjet",
		Card: game.Card{
			Name:      "Nalaar Aetherjet",
			TypeLine:  "Token Artifact — Vehicle",
			Power:     1,
			Toughness: 1,
			Keywords:  []string{"flying"},
		},
		Activated: []game.ActivatedAbilityShape{crewTokenAbility("Nalaar Aetherjet", 2)},
		Text:      "Flying\nCrew 2",
	}
}

// CmdrB23NalaarAetherjetToken is the Aetherjet at x/x.
func CmdrB23NalaarAetherjetToken(x int) game.Card {
	c := tokenFromCatalog(printedCmdrB23NalaarAetherjetToken)
	c.Power, c.Toughness = x, x
	return c
}
