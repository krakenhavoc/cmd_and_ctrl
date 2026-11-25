package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selfless Glyphweaver // Deadly Vanity — modal double-faced card (Strixhaven:
// School of Mages, #5502):
//
//	Selfless Glyphweaver — Creature — Human Cleric {2}{W}, 2/3
//	 "Exile this creature: Creatures you control gain indestructible until end
//	  of turn."
//	Deadly Vanity — Sorcery {5}{B}{B}{B}
//	 "Choose a creature or planeswalker, then destroy all other creatures and
//	  planeswalkers."
//
// # Two keys, one card
//
// The creature keeps the bare oracle ID and the sorcery takes
// "<oracle_id>#1" (Wandering Archaic's keyspace).
//
// # The front face
//
// Hanged Executioner's exile-this cost. The creatures that gain indestructible
// are those the controller controls as the ability resolves (CR 611.2c fixes
// the set then); the Glyphweaver itself is already in exile, so it is not one.
//
// # The back face
//
// Mount Doom's keep-and-destroy-the-rest: the caster is asked as the spell
// resolves (it does not target, so hexproof or shroud keeps nothing safe from
// being chosen) to pick one creature or planeswalker among every one on the
// battlefield, and every other creature and planeswalker is destroyed. With
// nothing to choose, nothing is destroyed. The wipe is destruction, so an
// indestructible permanent and a regeneration shield do what they always do.
//
// No simplification on either face.
func init() {
	const id = "4507d35a-f7f3-4abe-b06b-90c5331b533a"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Selfless Glyphweaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile this creature: Creatures you control gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ExileThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"indestructible"},
					Label:    "Selfless Glyphweaver — indestructible",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Deadly Vanity",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true},
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var candidates []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() || c.IsPlaneswalker() {
					candidates = append(candidates, c.InstanceID)
				}
			}
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Deadly Vanity — choose a creature or planeswalker; destroy all others",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				LiveZone: game.ZoneBattlefield,
				ThenKey:  deadlyVanityKeep,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// deadlyVanityKeep is the chosen permanent, keyed (#4253): everything else
// that is a creature or a planeswalker is destroyed at once.
var deadlyVanityKeep = effects.AskedPick("choose-cards/deadly-vanity-keep", func(ctx *effects.Context, r game.CardsChosen) error {
	return effects.DestroyAllMatching{Match: effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.Not(effects.B10OneOf(r.Picked)))}.Apply(ctx)
})
