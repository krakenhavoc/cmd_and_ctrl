package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dathon, His Heart Open — Legendary Creature — Tamarian Officer
// {G}{U}{R}, 3/3 (Star Trek, #4973):
//
//	"Whenever you cast an enchantment spell, draw a card. If it's a Saga
//	 spell, Dathon deals 2 damage to each opponent.
//	 {2}: Choose a counter on target enchantment you control. Put another
//	 counter of that kind on it. Activate only as a sorcery."
//
// The cast trigger is Enchantress's Presence's, with a rider read off the
// spell as it was cast (its snapshot on the trigger, else the card): a
// Saga spell also has Dathon deal 2 damage to each opponent, as Dathon
// last existed if it has left (CR 608.2h).
//
// The activated ability targets an enchantment its controller controls
// (CR 115.1). As it resolves its controller chooses one kind of counter
// on it (CR 608.2c) and it gets one more of that kind; with one kind
// there is nothing to ask, and with none nothing happens. It is not
// proliferate (CR 701.34): a "proliferate" replacement does not see it.
// A lore counter put on a Saga this way triggers its chapter (CR 714.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8de259a9-99d0-4f07-8dbc-f065e3af352b",
		Name:         "Dathon, His Heart Open",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventCast, effects.EnchantmentSpellCastByYou,
				"Dathon, His Heart Open — draw a card; if it's a Saga spell, 2 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					if !effects.TrekB24TriggeringSpellIsSaga(g, item) {
						return nil
					}
					return effects.DamageEachOpponent{N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}: Choose a counter on target enchantment you control. Put another counter of that kind on it. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{2}"),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target enchantment you control", effects.Enchantment(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return dathonChooseAKindOfCounter(ctx, t.ID)
					}
				}
				return nil
			},
		}},
	})
}

// dathonChooseAKindOfCounter asks for one kind of counter on perm and
// puts another of that kind on it.
func dathonChooseAKindOfCounter(ctx *effects.Context, perm uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(perm)
	if !ok {
		return nil
	}
	kinds := effects.MshA6KindsWithCounters(c.Counters)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return effects.AddCounter{Target: perm, Kind: kinds[0], N: 1}.Apply(ctx)
	}
	opts := make([]game.ChoiceOption, len(kinds))
	for i, k := range kinds {
		opts[i] = game.ChoiceOption{Label: "A " + k + " counter"}
	}
	return effects.PickOption{
		Question: "Dathon, His Heart Open — choose a kind of counter to put another of",
		Options:  opts,
		ThenKey:  dathonCounterKindPicked,
		Carry:    []uuid.UUID{perm},
		Words:    kinds,
	}.Apply(ctx)
}

var dathonCounterKindPicked = effects.OptionPickThen("option-pick/dathon-his-heart-open-counter-kind", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(r.Words) {
		return nil
	}
	return effects.AddCounter{Target: r.Carry[0], Kind: r.Words[r.Index], N: 1}.Apply(ctx)
})
