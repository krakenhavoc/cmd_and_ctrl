package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB276NamedInGraveyard is "the number of cards named <name> in your
// graveyard" (Galvanic Bombardment, Take Inventory), counted as the spell
// resolves, so the resolving spell itself (on the stack) is not counted.
func S82OldB276NamedInGraveyard(g *game.Game, player uuid.UUID, name string) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

// S82OldB276DevilToken is Make Mischief's Devil: a 1/1 red Devil creature
// with "When this creature dies, it deals 1 damage to any target." The
// trigger is the token's own (ADR 0083): its controller is "you", and the
// damage is dealt by the token as it last existed (CR 608.2h).
func S82OldB276DevilToken() game.Card { return TokenFromCatalog(printedS82OldB276DevilToken) }

func printedS82OldB276DevilToken() TokenTemplate {
	return TokenTemplate{
		Slug: "s82-old-b276-devil",
		Card: game.Card{
			Name:      "Devil",
			TypeLine:  "Token Creature — Devil",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Devil — it deals 1 damage to any target", SourceDealsDamageToEachLegalTarget(1)), TargetAny()),
		},
		Text: "When this creature dies, it deals 1 damage to any target.",
	}
}
