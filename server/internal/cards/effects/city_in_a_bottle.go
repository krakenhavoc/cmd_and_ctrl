package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// City in a Bottle — Artifact {2}:
//
//	"Whenever one or more other nontoken permanents with a name
//	 originally printed in the Arabian Nights expansion are on the
//	 battlefield, their controllers sacrifice them.
//	 Players can't cast spells or play lands with a name originally
//	 printed in the Arabian Nights expansion."
//
// ADR 0109 §4 (#1895), the card the land-play seam was filed for. Every
// "name originally printed in Arabian Nights" is CR 206.3a's list
// (game.ArabianNightsNames), read by all three clauses:
//
//   - The first sentence is a CR 603.8 state trigger (ADR 0107 §1): it
//     triggers whenever another nontoken permanent with such a name is on
//     the battlefield and is not already waiting or on the stack. It is
//     one ability for the whole state, so as it resolves every such
//     permanent is sacrificed by its own controller ("their controllers
//     sacrifice them"). The City itself is excluded ("other"), and so is
//     a token (CR 111.4: a token's name is its subtype text, and the
//     clause says nontoken).
//   - "Can't cast spells" is a CastRestriction and "can't play lands" is
//     a LandPlayRestriction (ADR 0109 §4): the same list, asked by the
//     same gates the enumerator and the view ask, so neither a bot nor
//     the client's highlight offers a play the server would refuse.
//
// No simplification.
func init() {
	const castLabel = "Players can't cast spells with a name originally printed in the Arabian Nights expansion."
	const landLabel = "Players can't play lands with a name originally printed in the Arabian Nights expansion."
	const sacLabel = "City in a Bottle — their controllers sacrifice them"
	Register(Spec{
		OracleID:     "a83f25e3-4d84-4c9b-ab12-19b8d326e459",
		Name:         "City in a Bottle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenState(sacLabel, func(g *game.Game, source *game.Card, _ uuid.UUID) bool {
				return len(cityInABottleVictims(g, source.InstanceID)) > 0
			}, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range cityInABottleVictims(g, item.SourceCardID) {
					if err := (SacrificePermanent{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
		CastRestrictions: []game.CastRestriction{
			PlayersCantCast(castLabel, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return game.IsArabianNightsName(c.Name)
			}),
		},
		LandPlayRestrictions: []game.LandPlayRestriction{
			CantPlayLandsNamed(landLabel, game.IsArabianNightsName),
		},
	})
}

// cityInABottleVictims is every nontoken permanent on the battlefield,
// other than the City, with a name originally printed in Arabian Nights.
func cityInABottleVictims(g *game.Game, cityID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != cityID && !c.IsToken() && game.IsArabianNightsName(c.Name) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
