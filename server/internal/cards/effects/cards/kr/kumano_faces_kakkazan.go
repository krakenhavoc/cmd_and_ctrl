package cardskr

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kumano Faces Kakkazan // Etching of Kumano — an Enchantment — Saga {R}
// (Kamigawa: Neon Dynasty, S82):
//
//	I   — This Saga deals 1 damage to each opponent and each planeswalker
//	      they control.
//	II  — When you next cast a creature spell this turn, that creature
//	      enters with an additional +1/+1 counter on it.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// Back face (registered under "<oracle_id>#1"), Etching of Kumano, an
// Enchantment Creature — Human Shaman 2/2: haste, and "If a creature dealt
// damage this turn by a source you controlled would die, exile it instead."
//
// II is a NextSpellPromise (Summon: Brynhildr's shape) carrying the entry
// counter, spent as the next creature spell becomes cast and lapsing at the
// end of the turn (CR 603.7). The back face's replacement is read off the
// turn's damage record by the dealing source's controller, so it covers any
// source its controller had, not one named object.
//
// No simplification.
func init() {
	const name = "Kumano Faces Kakkazan"
	effects.Register(effects.Spec{
		OracleID:     KumanoFacesKakkazanOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "1 damage to each opponent and each planeswalker they control"),
				kumanoFacesKakkazanDamage),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "when you next cast a creature spell this turn, it enters with an additional +1/+1 counter"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantNextSpellPromise{
						From: name,
						Promise: game.NextSpellPromise{
							Filter:      game.PermissionFilter{CreatureOnly: true},
							CounterKind: game.CounterPlusOne,
							Counters:    1,
							Text:        "When you next cast a creature spell this turn, that creature enters with an additional +1/+1 counter on it.",
						},
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        KumanoFacesKakkazanOracleID + "#1",
		Name:            "Etching of Kumano",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Replacements: []game.ReplacementEffect{
			effects.NeoB19ExileIfDamagedByYourSourcesWouldDie("Etching of Kumano: exile it instead of dying"),
		},
	})
}

// kumanoFacesKakkazanOracleID is shared with the back face.
const KumanoFacesKakkazanOracleID = "535916f8-b51a-4414-8ab0-fdccdefd9433"

// kumanoFacesKakkazanDamage is chapter I: the Saga deals 1 damage to each
// opponent, then to each planeswalker an opponent controls.
func kumanoFacesKakkazanDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opponents := ctx.Opponents()
	var walkers []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsPlaneswalker() && slices.Contains(opponents, c.Controller) {
			walkers = append(walkers, c)
		}
	}
	// One instruction deals all of it: one instance of damage (CR 615.8).
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range opponents {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: opp, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, c := range walkers {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: c.InstanceID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
