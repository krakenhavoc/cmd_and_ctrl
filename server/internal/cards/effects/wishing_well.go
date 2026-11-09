package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wishing Well — Artifact {3}{U} (Bloomburrow):
//
//	"{T}: Put a coin counter on this artifact. When you do, you may cast target
//	 instant or sorcery card with mana value equal to the number of coin counters
//	 on this artifact from your graveyard without paying its mana cost. If that
//	 spell would be put into your graveyard, exile it instead. Activate only as a
//	 sorcery."
//
// The counter is put on, and "when you do" is a reflexive triggered ability
// (CR 603.12) whose target is chosen as it goes on the stack, from instants and
// sorceries in your graveyard whose mana value equals the counters now on the
// artifact (stdB31WishingWellBody carries that number as Params.Amount). The cast
// is made as the trigger resolves (CR 608.2g, Daring Waverider's free cast): it
// can be declined, an X in the cost is zero, and the spell is exiled whether it
// resolves, is countered or fizzles. With no legal target the trigger is removed
// (CR 603.3d), and the counter stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78dbbc15-9304-4178-b9b2-1db6c64ca11a",
		Name:         "Wishing Well",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}: Put a coin counter on this artifact. When you do, you may cast target instant or sorcery card with mana value equal to the number of coin counters on this artifact from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead. Activate only as a sorcery.",
			Cost:         TapCost(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (AddCounter{Target: item.SourceCardID, Kind: stdB31CoinCounter, N: 1}).Apply(ctx); err != nil {
					return err
				}
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				return ReflexiveTrigger{
					Label:  "Wishing Well — you may cast target instant or sorcery card with mana value equal to the number of coin counters from your graveyard without paying its mana cost",
					Body:   stdB31WishingWellBody,
					Params: game.EffectParams{Amount: src.Counters[stdB31CoinCounter]},
				}.Apply(ctx)
			},
		}},
	})
}
