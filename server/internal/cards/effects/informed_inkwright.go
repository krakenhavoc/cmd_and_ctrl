package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Informed Inkwright — Creature — Human Wizard {1}{W}, 2/2 (Secrets of
// Strixhaven promo, #4834):
//
//	"Vigilance
//	 Repartee — Whenever you cast an instant or sorcery spell that targets
//	 a creature, create a 1/1 white and black Inkling creature token with
//	 flying."
//
// Repartee is an ability word; the ability is one trigger on your instant
// and sorcery casts. Whether the spell "targets a creature" is read from
// the targets it was cast with (heroic's reading): a target that is a
// creature on the battlefield as the spell is cast. The rulings: it
// resolves before the spell and even if that spell is countered, and it
// does not matter if the creature leaves or stops being a creature
// before the spell resolves. A spell that targets only players, or a
// creature spell, does not count. Several creature targets still make one
// Inkling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2a372d8-3eb2-4ba4-a75f-7951654de383",
		Name:            "Informed Inkwright",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Informed Inkwright — Repartee: create a 1/1 Inkling with flying",
				Do(CreateToken{Template: TokenCard("1/1 white and black Inkling with flying"), N: 1})),
		},
	})
}

// informedInkwrightCast is "you cast an instant or sorcery spell that
// targets a creature": the cast spell's own targets, as announced.
func informedInkwrightCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() {
			return true
		}
	}
	return false
}
