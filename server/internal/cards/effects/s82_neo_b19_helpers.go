package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b19_helpers.go — shared pieces of Kamigawa: Neon Dynasty batch 19
// (the double-faced Sagas of S82).

// counterHatching is the hatching counter (The Dragon-Kami Reborn): put on a
// card exiled face down, it marks the cards Dragon-Kami's Egg may cast.
const CounterHatching = "hatching"

// neoB19HatchingExileCards are the cards a player owns in exile with a
// hatching counter on them.
func NeoB19HatchingExileCards(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	if g.Exile == nil {
		return nil
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == owner && c.Counters[CounterHatching] > 0 {
			out = append(out, c)
		}
	}
	return out
}

// neoB19ExileIfDamagedByYourSourcesWouldDie is Etching of Kumano's "If a
// creature dealt damage this turn by a source you controlled would die,
// exile it instead." The damage is read off the turn's damage record by
// the controller of the dealing source as it dealt the damage
// (ObjectDamageRecord.SourceController), so it reaches the creatures any of
// the controller's sources damaged, and a source that has since left or
// changed hands still counts (CR 400.7, CR 608.2h).
func NeoB19ExileIfDamagedByYourSourcesWouldDie(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || !c.IsCreature() {
				return false
			}
			return g.DamageDealtToObjectThisTurnByControlledForEffect(ev.CardID, src.Controller) > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: replacementSourceController,
		Label:      label,
		RowLabel:   "If a creature dealt damage this turn by a source you controlled would die, exile it instead.",
	}
}

// printedNeoHumanMonkToken is Jugan Defends the Temple's 1/1 green Human
// Monk creature token with "{T}: Add {G}." The mana ability rides the
// template (token_catalog.go); its slug is an on-disk identity.
func printedNeoHumanMonkToken() TokenTemplate {
	return TokenTemplate{
		Slug: "human-monk",
		Card: game.Card{
			Name:      "Human Monk",
			TypeLine:  "Token Creature — Human Monk",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Text: "{T}: Add {G}.",
	}
}

// neoB19HumanMonkToken is the token Jugan Defends the Temple makes.
func NeoB19HumanMonkToken() game.Card { return TokenFromCatalog(printedNeoHumanMonkToken) }

// neoB19FiveModifiedCreatures is "you control five or more modified
// creatures" (CR 700.9), read live on every layer pass.
func NeoB19FiveModifiedCreatures(g *game.Game, you uuid.UUID) bool {
	return CountControlled(g, you, func(c game.Card) bool { return c.IsCreature() && B07IsModified(g, c) }) >= 5
}
