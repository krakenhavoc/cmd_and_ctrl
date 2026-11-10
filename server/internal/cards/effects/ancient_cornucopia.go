package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Cornucopia — Artifact {2}{G} (The Big Score, #4468):
//
//	"Whenever you cast a spell that's one or more colors, you may gain 1
//	 life for each of that spell's colors. Do this only once each turn.
//	 {T}: Add one mana of any color."
//
// The number of colors is counted when the trigger is built, so a spell
// countered in response is read as it last was (CR 608.2h, per the printed
// ruling). The "may" is asked as the ability resolves (CR 603.5); a "yes"
// is what uses up the turn, recorded on the object's per-turn mark
// (ReplacementAppliedThisTurnForEffect), not the cast: a "no" leaves the
// next colored spell free to trigger again, and a second trigger already on
// the stack does nothing once the life has been gained.
//
// No simplification.
func init() {
	const label = "Ancient Cornucopia — gain 1 life for each of that spell's colors"
	Register(Spec{
		OracleID:     "f0950c95-a297-489d-b166-74bba6af5376",
		Name:         "Ancient Cornucopia",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Triggered: []game.TriggeredAbility{Optional(game.TriggeredAbility{
			Watches: []game.EventKind{game.EventCast},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && ev.Actor == source.Controller &&
					spellColorCount(g, ev.CardID) > 0 &&
					!g.ReplacementAppliedThisTurnForEffect(source, label)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = spellColorCount(g, ev.CardID)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || item.Params.Amount <= 0 || g.ReplacementAppliedThisTurnForEffect(&src, label) {
					return nil
				}
				g.MarkReplacementAppliedThisTurnForEffect(&src, label)
				return GainLife{Amount: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}, "Ancient Cornucopia — gain life for each of that spell's colors?")},
	})
}
