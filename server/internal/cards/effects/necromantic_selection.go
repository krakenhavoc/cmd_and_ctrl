package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necromantic Selection — Sorcery {4}{B}{B}{B} (Tarkir: Dragonstorm Commander):
//
//	"Destroy all creatures, then return a creature card put into a
//	 graveyard this way to the battlefield under your control. It's a
//	 black Zombie in addition to its other colors and types. Exile
//	 Necromantic Selection."
//
// Deadly Tempest's wipe with a keyed tail. The tail is told the creatures
// that were destroyed; the candidates are the CARDS among them that are in
// a graveyard now (a token is not a card, a creature the CR 614 window
// saved or sent elsewhere did not land in one). One candidate is returned
// without asking; with several, the caster picks one (a choose_cards
// prompt). The card enters under the caster's control as a black Zombie in
// addition to its other colors and types (Liliana, Death's Majesty's
// pinned effect). The spell then exiles itself, whatever was returned:
// the exile is the last instruction, run from the return's own tail.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b48a149c-aead-43b2-a99e-c964d6d28aa0",
		Name:         "Necromantic Selection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Creature(), ThenTail: ctx.Tail(necromanticSelectionReturn)}.Apply(ctx)
		},
	})
}

// necromanticSelectionReturn is the wipe's tail: pick or take the one
// creature card put into a graveyard this way and return it, or just
// exile the spell when there is none.
var necromanticSelectionReturn = MassTail("mass/necromantic-selection-return", func(ctx *Context, swept []game.Card, _ int) error {
	var candidates []uuid.UUID
	for _, c := range swept {
		if c.IsToken() || !c.IsCreature() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z != nil && z.Kind == game.ZoneGraveyard {
			candidates = append(candidates, c.InstanceID)
		}
	}
	switch len(candidates) {
	case 0:
		return necromanticSelectionFinish(ctx, nil)
	case 1:
		return necromanticSelectionFinish(ctx, candidates)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Necromantic Selection — return a creature card put into a graveyard this way to the battlefield under your control",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		// No LiveZone: the candidates sit in several players' graveyards,
		// and the live re-check reads one owner's. Nothing can move them
		// before the answer, as it is part of this resolution.
		ThenKey: necromanticSelectionPick,
		Asked:   game.NewAskedItem(ctx.Item),
	})
	return nil
})

// necromanticSelectionPick is the caster's answer.
var necromanticSelectionPick = askedPickRule("choose-cards/necromantic-selection-pick", nil,
	func(ctx *Context, r game.CardsChosen) error {
		return necromanticSelectionFinish(ctx, r.Picked)
	})

// necromanticSelectionFinish returns the chosen card as a black Zombie,
// then exiles the spell; with nothing chosen it only exiles the spell.
func necromanticSelectionFinish(ctx *Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return necromanticSelectionExile(ctx, nil, 0)
	}
	return ReturnFromGraveyardTogether{
		Targets:    ids,
		Controller: ctx.Controller(),
		ThenTail:   necromanticSelectionZombie,
	}.Apply(ctx)
}

// necromanticSelectionZombie is "It's a black Zombie in addition to its
// other colors and types. Exile Necromantic Selection." after the return.
var necromanticSelectionZombie = TogetherTail("necromantic-selection/black-zombie-then-exile",
	func(ctx *Context, entered []uuid.UUID, n int) error {
		for _, id := range entered {
			if err := (ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...)},
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
				Label:    "Necromantic Selection — it is a black Zombie in addition to its other colors and types",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return necromanticSelectionExile(ctx, entered, n)
	})

// necromanticSelectionExile is the last sentence: the spell exiles itself
// (it moves itself, so the resolution does not put it into the graveyard).
func necromanticSelectionExile(ctx *Context, _ []uuid.UUID, _ int) error {
	return ctx.Game.ExileCardForEffect(ctx.Source())
}
